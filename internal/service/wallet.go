package service

import (
	"crypto/ecdsa"
	"fmt"
	"math/big"
	"sync"
	"time"

	"github.com/bnb-chain/tss-lib/common"
	"github.com/bnb-chain/tss-lib/ecdsa/keygen"
	"github.com/bnb-chain/tss-lib/ecdsa/signing"
	"github.com/bnb-chain/tss-lib/tss"
	"github.com/ethereum/go-ethereum/crypto"
)

type WalletService struct {
	storage Storage
}

type Storage interface {
	SaveWallet(address string, wallet *Wallet) error
	GetWallet(address string) (*Wallet, error)
	ListWallets() ([]string, error)
}

type Wallet struct {
	Address   string
	PublicKey *ecdsa.PublicKey
	KeyShare  []keygen.LocalPartySaveData
}

func NewWalletService(storage Storage) *WalletService {
	return &WalletService{
		storage: storage,
	}
}

func (s *WalletService) CreateWallet() (*Wallet, error) {
	// Initialize TSS parameters
	threshold := 2
	partyCount := 3

	// Create mock party IDs
	partyIDs := make([]*tss.PartyID, partyCount)
	for i := 1; i <= partyCount; i++ {
		id := tss.NewPartyID(fmt.Sprintf("%d", i), fmt.Sprintf("party-%d", i), new(big.Int).SetInt64(int64(i)))
		partyIDs[i-1] = id
	}

	// Sort party IDs
	sortedPartyIDs := tss.SortPartyIDs(partyIDs)

	// Create peer context
	ctx := tss.NewPeerContext(sortedPartyIDs)

	// Channels for each party
	outChs := make([]chan tss.Message, partyCount)
	endChs := make([]chan keygen.LocalPartySaveData, partyCount)
	parties := make([]tss.Party, partyCount)

	// Initialize channels and parties
	for i := 0; i < partyCount; i++ {
		outChs[i] = make(chan tss.Message, partyCount*3)
		endChs[i] = make(chan keygen.LocalPartySaveData, 1)

		params := tss.NewParameters(tss.S256(), ctx, partyIDs[i], partyCount, threshold)
		parties[i] = keygen.NewLocalParty(params, outChs[i], endChs[i])
	}

	// Start all parties
	for i := 0; i < partyCount; i++ {
		go func(party tss.Party) {
			if err := party.Start(); err != nil {
				fmt.Printf("Failed to start party: %v\n", err)
			}
		}(parties[i])
	}

	// Handle message routing between parties
	var wg sync.WaitGroup
	wg.Add(partyCount)

	for i := 0; i < partyCount; i++ {
		go func(idx int) {
			defer wg.Done()
			for {
				select {
				case msg := <-outChs[idx]:
					// Route the message to all parties (including self for broadcast)
					dest := msg.GetTo()
					if dest == nil { // broadcast
						for j := 0; j < partyCount; j++ {
							fmt.Printf("Broadcasting to %d\n", j)
							wireBytes, _, err := msg.WireBytes()
							if err != nil {
								fmt.Printf("Failed to get wire bytes for party %d: %v\n", j, err)
								continue
							}
							if _, err := parties[j].UpdateFromBytes(wireBytes, msg.GetFrom(), true); err != nil {
								fmt.Printf("Failed to update party %d: %v\n", j, err)
							}
						}
					} else { // point to point
						destID := dest[0].Index
						fmt.Printf("Sending to %d\n", destID)
						wireBytes, _, err := msg.WireBytes()
						if err != nil {
							fmt.Printf("Failed to get wire bytes for party %d: %v\n", destID, err)
							continue
						}
						if _, err := parties[destID].UpdateFromBytes(wireBytes, msg.GetFrom(), false); err != nil {
							fmt.Printf("Failed to update party %d: %v\n", destID, err)
						}
					}
				case <-time.After(60 * time.Second):
					return
				}
			}
		}(i)
	}

	// collect key shares
	keyShares := make([]keygen.LocalPartySaveData, partyCount)
	for i := 0; i < partyCount; i++ {
		select {
		case keyShare := <-endChs[i]:
			keyShares[i] = keyShare
			fmt.Printf("Received key share from party %d\n", i)
		case <-time.After(120 * time.Second):
			return nil, fmt.Errorf("key generation timed out")
		}
	}

	// Use the first key share for the wallet
	keyShare := keyShares[0]
	pubKey := &ecdsa.PublicKey{
		Curve: tss.S256(),
		X:     keyShare.ECDSAPub.X(),
		Y:     keyShare.ECDSAPub.Y(),
	}
	address := crypto.PubkeyToAddress(*pubKey)

	// Create wallet instance
	wallet := &Wallet{
		Address:   address.Hex(),
		PublicKey: pubKey,
		KeyShare:  keyShares,
	}

	// Save wallet
	if err := s.storage.SaveWallet(address.Hex(), wallet); err != nil {
		return nil, err
	}

	return wallet, nil
}

func (s *WalletService) ListWallets() ([]string, error) {
	return s.storage.ListWallets()
}

func (s *WalletService) SignData(walletAddress string, data []byte) ([]byte, error) {
	wallet, err := s.storage.GetWallet(walletAddress)
	if err != nil {
		return nil, fmt.Errorf("wallet not found: %w", err)
	}

	// Initialize signing parameters
	partyCount := 3
	threshold := 2

	// Create party IDs (same as during key generation)
	partyIDs := make([]*tss.PartyID, partyCount)
	for i := 1; i <= partyCount; i++ {
		id := tss.NewPartyID(fmt.Sprintf("%d", i), fmt.Sprintf("party-%d", i), new(big.Int).SetInt64(int64(i)))
		partyIDs[i-1] = id
	}
	sortedPartyIDs := tss.SortPartyIDs(partyIDs)

	// Create peer context
	ctx := tss.NewPeerContext(sortedPartyIDs)

	// Create channels for each party
	outChs := make([]chan tss.Message, partyCount)
	endChs := make([]chan common.SignatureData, partyCount)
	parties := make([]tss.Party, partyCount)

	// Hash the message to be signed
	fmt.Printf("data is here: %s\n", data)
	msgHash := crypto.Keccak256(data)
	msg := new(big.Int).SetBytes(msgHash)
	fmt.Printf("message is here: %s\n", msg)

	// Initialize channels and parties
	for i := 0; i < partyCount; i++ {
		outChs[i] = make(chan tss.Message, partyCount*3)
		endChs[i] = make(chan common.SignatureData, 1)

		params := tss.NewParameters(tss.S256(), ctx, partyIDs[i], partyCount, threshold)

		// Create signing party using the key share from key generation
		parties[i] = signing.NewLocalParty(msg, params, wallet.KeyShare[i], outChs[i], endChs[i])
	}

	// Start all parties
	for i := 0; i < partyCount; i++ {
		go func(party tss.Party) {
			if err := party.Start(); err != nil {
				fmt.Printf("Failed to start signing party: %v\n", err)
			}
		}(parties[i])
	}

	// Handle message routing between parties
	var wg sync.WaitGroup
	wg.Add(partyCount)

	for i := 0; i < partyCount; i++ {
		go func(idx int) {
			defer wg.Done()
			for {
				select {
				case msg := <-outChs[idx]:
					// Route the message to all parties
					dest := msg.GetTo()
					if dest == nil { // broadcast
						for j := 0; j < partyCount; j++ {
							wireBytes, _, err := msg.WireBytes()
							if err != nil {
								fmt.Printf("Failed to get wire bytes for party %d: %v\n", j, err)
								continue
							}
							if _, err := parties[j].UpdateFromBytes(wireBytes, msg.GetFrom(), true); err != nil {
								fmt.Printf("Failed to update party %d: %v\n", j, err)
							}
						}
					} else { // point to point
						destID := dest[0].Index
						wireBytes, _, err := msg.WireBytes()
						if err != nil {
							fmt.Printf("Failed to get wire bytes for party %d: %v\n", destID, err)
							continue
						}
						if _, err := parties[destID].UpdateFromBytes(wireBytes, msg.GetFrom(), false); err != nil {
							fmt.Printf("Failed to update party %d: %v\n", destID, err)
						}
					}
				case <-time.After(30 * time.Second):
					return
				}
			}
		}(i)
	}

	// Collect signatures
	var signature *common.SignatureData
	select {
	case sigData := <-endChs[0]:
		signature = &sigData
		fmt.Println("Signature generated successfully")
	case <-time.After(30 * time.Second):
		return nil, fmt.Errorf("signature generation timed out")
	}

	// Convert signature to Ethereum format
	// R, S values for Ethereum signature
	rBytes := signature.R
	sBytes := signature.S

	// Ensure R and S are 32 bytes each
	R := make([]byte, 32)
	S := make([]byte, 32)
	copy(R[32-len(rBytes):], rBytes)
	copy(S[32-len(sBytes):], sBytes)

	// Combine R, S and V into a 65-byte signature
	sig := make([]byte, 65)
	copy(sig[0:32], R)
	copy(sig[32:64], S)
	sig[64] = 0x1b // Default recovery ID

	return sig, nil
}
