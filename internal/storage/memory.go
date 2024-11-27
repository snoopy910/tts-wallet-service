package storage

import (
	"fmt"
	"sync"

	"github.com/snoopy910/tss-wallet-service/internal/service"
)

type MemoryStorage struct {
	wallets map[string]*service.Wallet
	mu      sync.RWMutex
}

func NewMemoryStorage() *MemoryStorage {
	return &MemoryStorage{
		wallets: make(map[string]*service.Wallet),
	}
}

func (s *MemoryStorage) SaveWallet(address string, wallet *service.Wallet) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.wallets[address] = wallet
	return nil
}

func (s *MemoryStorage) GetWallet(address string) (*service.Wallet, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	wallet, exists := s.wallets[address]
	if !exists {
		return nil, fmt.Errorf("wallet not found")
	}

	return wallet, nil
}

func (s *MemoryStorage) ListWallets() ([]string, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	addresses := make([]string, 0, len(s.wallets))
	for addr := range s.wallets {
		addresses = append(addresses, addr)
	}

	return addresses, nil
}
