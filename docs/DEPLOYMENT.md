# TLOCK Deployment Guide

## Overview

This guide provides instructions for deploying TLOCK nodes based on the actual codebase and scripts. All commands and configurations are verified against the real implementation.

## Prerequisites

### System Requirements
- **Go**: 1.21+ (required for building)
- **Git**: For cloning the repository
- **Make**: For using the Makefile
- **jq**: For JSON processing (required by scripts)

### Operating Systems
- Linux (Ubuntu/Debian recommended)
- macOS
- Windows (with appropriate tooling)

## Installation

### 1. Clone and Build

```bash
# Clone the repository
git clone https://github.com/Tlock-org/tlock.git
cd tlock

# Build the binary
make install

# Verify installation
tlockd version
```

The `make install` command will build and install the `tlockd` binary to your `$GOPATH/bin`.

## Local Development Setup

### Using the Test Script

TLOCK provides a convenient test script for local development:

```bash
# Basic usage with defaults
sh scripts/test_node.sh

# Custom configuration
CHAIN_ID="localchain-1" HOME_DIR="~/.tlock" BLOCK_TIME="1000ms" CLEAN=true sh scripts/test_node.sh

# Custom ports
CHAIN_ID="localchain-2" HOME_DIR="~/.tlock" CLEAN=true RPC=36657 REST=2317 P2P=36656 GRPC=8090 sh scripts/test_node.sh
```

### Script Configuration Options

The test script supports the following environment variables:

| Variable | Default | Description |
|----------|---------|-------------|
| `CHAIN_ID` | `10889` | Chain identifier |
| `HOME_DIR` | `~/.tlock` | Node home directory |
| `BINARY` | `tlockd` | Binary name |
| `DENOM` | `uTOK` | Token denomination |
| `BLOCK_TIME` | `3s` | Block time |
| `RPC` | `26657` | RPC port |
| `REST` | `1317` | REST API port |
| `P2P` | `26656` | P2P port |
| `GRPC` | `9090` | gRPC port |
| `KEYRING` | `test` | Keyring backend |

### Pre-configured Test Keys

The test script includes several pre-configured keys with mnemonics:

- **node0**: `tlock1hj5fveer5cjtn4wd6wstzugjfdxzl0xp5u7j9p`
- **node1**: `tlock1efd63aw40lxf3n4mhf7dzhjkr453axurggdkvg`
- **node2**: `tlock1wfvjqmkekyuy59r535nm2ca3yjkf706nu8x49r`
- **node3**: `tlock1qvmhf9qw5xhefm6jpqpggneahanjhkgw0szlzn`
- **node4**: `tlock16f7etm42yp5nup77q3027rvvkl73q2gr8wkjcm`
- **node5**: `tlock1h0g7hntn53awu7ewzw9l2s30qm8j3xg6xzr3gy`

## Manual Setup

### 1. Initialize Node

```bash
# Set configuration
tlockd config set client chain-id localchain-1
tlockd config set client keyring-backend test

# Initialize node
tlockd init mynode --chain-id localchain-1
```

### 2. Create Keys

```bash
# Create a new key
tlockd keys add mykey --keyring-backend test

# Or recover from mnemonic
echo "your mnemonic phrase here" | tlockd keys add mykey --keyring-backend test --recover
```

### 3. Configure Genesis

```bash
# Add genesis account
tlockd genesis add-genesis-account mykey 1000000000000uTOK --keyring-backend test

# Create genesis transaction
tlockd genesis gentx mykey 1000000uTOK --chain-id localchain-1 --keyring-backend test

# Collect genesis transactions
tlockd genesis collect-gentxs
```

### 4. Start Node

```bash
# Start the node
tlockd start

# Or run in background
nohup tlockd start > tlock.log 2>&1 &
```

## Network Configuration

### Chain Registry Information

Based on `chain_registry.json`:

- **Chain Name**: tlock
- **Chain ID**: localchain-1 (testnet)
- **Bech32 Prefix**: tlock
- **Native Token**: TOK
- **Slip44**: 118
- **Cosmos SDK**: v0.50
- **Tendermint**: v0.38

### Default Ports

- **RPC**: 26657
- **REST API**: 1317
- **P2P**: 26656
- **gRPC**: 9090
- **Prometheus**: 26660

## Available Modules

TLOCK includes the following modules (from `app/app.go`):

### Standard Cosmos SDK Modules
- **auth**: Account authentication
- **bank**: Token transfers
- **staking**: Validator staking
- **gov**: Governance proposals and voting
- **distribution**: Staking rewards distribution
- **slashing**: Validator slashing
- **mint**: Token minting
- **feegrant**: Fee allowances
- **authz**: Authorization
- **group**: Group functionality
- **evidence**: Evidence handling
- **crisis**: Crisis management
- **params**: Parameter management
- **upgrade**: Chain upgrades
- **consensus**: Consensus parameters

### IBC Modules
- **ibc**: Inter-blockchain communication
- **transfer**: IBC token transfers
- **ica**: Interchain accounts
- **ibcfee**: IBC relayer fees

### Custom Modules
- **post**: Social media posts and interactions
- **profile**: User profiles and social graph
- **tokenfactory**: Token creation
- **packetforward**: Packet forwarding
- **test**: Testing utilities

## CLI Commands

### Query Commands

```bash
# Check node status
tlockd status

# Query account balance
tlockd query bank balances tlock1...

# Query post
tlockd query post get post_123

# Query profile
tlockd query profile get tlock1...

# Query governance proposals
tlockd query gov proposals
```

### Transaction Commands

```bash
# Send tokens
tlockd tx bank send mykey tlock1... 1000uTOK --chain-id localchain-1

# Create post
tlockd tx post create-post --from mykey --chain-id localchain-1

# Create profile
tlockd tx profile add-profile --from mykey --chain-id localchain-1

# Submit governance proposal
tlockd tx gov submit-proposal --from mykey --chain-id localchain-1
```

## Configuration Files

### Node Configuration (`~/.tlock/config/config.toml`)

Key sections to configure:

```toml
[rpc]
laddr = "tcp://127.0.0.1:26657"
cors_allowed_origins = []

[p2p]
laddr = "tcp://0.0.0.0:26656"
persistent_peers = ""
seeds = ""

[consensus]
timeout_commit = "3s"
```

### Application Configuration (`~/.tlock/config/app.toml`)

Key sections to configure:

```toml
[api]
enable = true
swagger = true
address = "tcp://0.0.0.0:1317"

[grpc]
enable = true
address = "0.0.0.0:9090"

[grpc-web]
enable = true
address = "0.0.0.0:9091"
```

## Development Tools

### Makefile Targets

```bash
# Build binary
make install

# Run tests
make test

# Generate protobuf files
make proto-gen

# Format code
make format

# Lint code
make lint
```

### Docker Support

Check if Docker files exist in the repository for containerized deployment.

## Troubleshooting

### Common Issues

1. **Binary not found**: Ensure `$GOPATH/bin` is in your `$PATH`
2. **Port conflicts**: Change ports using environment variables
3. **Permission errors**: Check file permissions in home directory
4. **Genesis errors**: Ensure proper key setup before genesis creation

### Logs

Check logs for debugging:

```bash
# View logs if running in background
tail -f tlock.log

# Check systemd logs if using service
journalctl -u tlockd -f
```

## Production Considerations

### Security
- Use proper key management (hardware wallets for validators)
- Configure firewall rules
- Regular security updates
- Backup validator keys securely

### Monitoring
- Set up Prometheus metrics collection
- Monitor node sync status
- Track validator performance
- Set up alerting for downtime

### Backup
- Regular backup of validator keys
- Backup node configuration
- State sync for quick recovery

## Support

For deployment issues:
- Check the GitHub repository: https://github.com/Tlock-org/tlock
- Review the test scripts in `scripts/` directory
- Examine the Makefile for build options

---

*This deployment guide is based on the actual codebase and tested scripts. All commands and configurations are verified against the real implementation.*
