package storage

import (
	"fmt"
	"sync"
)

type MemoryStorage struct {
	wallets map[string]*Wallet
	mu      sync.RWMutex
}

func NewMemoryStorage() *MemoryStorage {
	return &MemoryStorage{
		wallets: make(map[string]*Wallet),
	}
}

func (s *MemoryStorage) SaveWallet(address string, wallet *Wallet) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.wallets[address] = wallet
	return nil
}

func (s *MemoryStorage) GetWallet(address string) (*Wallet, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	wallet, exists := s.wallets[address]
	if !exists {
		return nil, fmt.Errorf("wallet not found")
	}

	return wallet, nil
}

func (s *MemoryStorage) ListWallets() []string {
	s.mu.RLock()
	defer s.mu.RUnlock()

	addresses := make([]string, 0, len(s.wallets))
	for addr := range s.wallets {
		addresses = append(addresses, addr)
	}

	return addresses
}
