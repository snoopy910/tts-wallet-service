# TSS Wallet Service

A Go-based service for generating and managing Ethereum TSS wallets using MPC.

## Prerequisites

- Go 1.21 or later
- Git

## Installation

1. Clone the repository:
```
git clone https://github.com/snoopy910/tss-wallet-service
cd tss-wallet-service
```

2. Install dependencies:
```
go mod download
```

## Running the Service

Start the server:
```
go run cmd/server/main.go
```

The server will start on port 8080.

## API Endpoints

### Create Wallet
```
curl -X POST http://localhost:8080/wallet
```

### List Wallets
```
curl http://localhost:8080/wallets
```

### Sign Data
```
curl "http://localhost:8080/sign?wallet={wallet_address}&data=hello"
```

## Design Choices

- **Gin Framework**: Used for its simplicity, performance, and extensive middleware ecosystem
- **TSS-lib**: Binance's implementation of threshold signatures
- **In-memory Storage**: Simple storage solution for demonstration purposes
- **Modular Architecture**: Separated into api, service, and storage layers for better maintainability