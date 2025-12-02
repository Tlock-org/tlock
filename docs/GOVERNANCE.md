# TLOCK Governance

## Overview

TLOCK uses the standard Cosmos SDK governance module for on-chain governance. This document describes the actual governance features implemented in the codebase, not theoretical or planned features.

## Governance Module

### Standard Cosmos SDK Governance

TLOCK includes the standard Cosmos SDK governance module (`x/gov`) which provides:

- **Proposal submission and voting**
- **Parameter change proposals**
- **Community spend proposals**
- **Software upgrade proposals**

### Module Configuration

From the codebase (`app/app.go`):

```go
govConfig := govtypes.DefaultConfig()
govConfig.MaxMetadataLen = 20000
govKeeper := govkeeper.NewKeeper(
    appCodec,
    runtime.NewKVStoreService(keys[govtypes.StoreKey]),
    app.AccountKeeper,
    app.BankKeeper,
    app.StakingKeeper,
    app.DistrKeeper,
    app.MsgServiceRouter(),
    govConfig,
    authtypes.NewModuleAddress(govtypes.ModuleName).String(),
)
```

**Key Configuration**:
- Maximum metadata length: 20,000 characters
- Integrated with staking, bank, and distribution modules
- Uses standard Cosmos SDK governance authority

## Available Governance Features

### 1. Parameter Change Proposals

Modify blockchain parameters for any module:

```bash
# Submit parameter change proposal
tlockd tx gov submit-proposal param-change proposal.json --from mykey
```

**Example proposal.json**:
```json
{
  "title": "Update Block Time",
  "description": "Proposal to update consensus block time",
  "changes": [
    {
      "subspace": "consensus",
      "key": "timeout_commit",
      "value": "5s"
    }
  ]
}
```

### 2. Community Pool Spend Proposals

Allocate funds from the community pool:

```bash
# Submit community spend proposal
tlockd tx gov submit-proposal community-pool-spend proposal.json --from mykey
```

### 3. Software Upgrade Proposals

Coordinate chain upgrades:

```bash
# Submit upgrade proposal
tlockd tx gov submit-proposal software-upgrade v2.0.0 --upgrade-height 1000000 --from mykey
```

### 4. Text Proposals

General governance discussions:

```bash
# Submit text proposal
tlockd tx gov submit-proposal --type="Text" --title="Community Discussion" --description="..." --from mykey
```

## Governance Process

### 1. Proposal Submission

**Requirements**:
- Minimum deposit (configurable via governance)
- Valid proposal format
- Proper authorization

**Command**:
```bash
tlockd tx gov submit-proposal [proposal-file] --from [key] --deposit [amount]
```

### 2. Deposit Period

- Proposals must reach minimum deposit to enter voting
- Community can contribute to deposits
- Failed to reach minimum deposit = proposal rejected

### 3. Voting Period

**Vote Options**:
- `Yes`: Support the proposal
- `No`: Oppose the proposal  
- `Abstain`: Participate in quorum without position
- `NoWithVeto`: Strong opposition (can veto if >33.4%)

**Voting Command**:
```bash
tlockd tx gov vote [proposal-id] [vote-option] --from [key]
```

### 4. Tallying and Execution

**Requirements for Passing**:
- Quorum: Minimum participation threshold
- Threshold: Majority of participating votes
- Veto: Less than 33.4% NoWithVeto votes

## CLI Commands

### Query Commands

```bash
# List all proposals
tlockd query gov proposals

# Get specific proposal
tlockd query gov proposal [proposal-id]

# Check vote
tlockd query gov vote [proposal-id] [voter-address]

# Get voting parameters
tlockd query gov params voting

# Get deposit parameters  
tlockd query gov params deposit

# Get tallying parameters
tlockd query gov params tallying
```

### Transaction Commands

```bash
# Submit proposal
tlockd tx gov submit-proposal [proposal-file] --from [key]

# Deposit on proposal
tlockd tx gov deposit [proposal-id] [amount] --from [key]

# Vote on proposal
tlockd tx gov vote [proposal-id] [vote-option] --from [key]

# Weighted vote (for validators)
tlockd tx gov weighted-vote [proposal-id] [weighted-votes] --from [key]
```

## Governance Parameters

### Default Parameters

The governance module uses Cosmos SDK defaults, which can be modified through governance:

**Voting Parameters**:
- Voting period: 2 weeks (configurable)
- Quorum: 33.4% of bonded tokens
- Threshold: 50% of participating votes
- Veto threshold: 33.4% of participating votes

**Deposit Parameters**:
- Minimum deposit: Configurable amount in TOK
- Deposit period: 2 weeks (configurable)

**Tallying Parameters**:
- Quorum: 0.334 (33.4%)
- Threshold: 0.5 (50%)
- Veto threshold: 0.334 (33.4%)

## Validator Governance

### Validator Voting

Validators can vote on behalf of their delegators:

```bash
# Validator votes for delegators
tlockd tx gov vote [proposal-id] [vote-option] --from validator-key
```

### Delegator Override

Delegators can override their validator's vote:

```bash
# Delegator votes directly
tlockd tx gov vote [proposal-id] [vote-option] --from delegator-key
```

## Module Integration

### Governance Authority

The governance module has authority over:

- **Parameter changes** for all modules
- **Community pool** spending
- **Software upgrades**
- **Module-specific** governance actions

### Custom Module Governance

TLOCK's custom modules (post, profile) can be governed through:

1. **Parameter change proposals** for module parameters
2. **Custom message types** executed by governance authority

## REST API Endpoints

### Query Endpoints

```http
# Get all proposals
GET /cosmos/gov/v1beta1/proposals

# Get specific proposal
GET /cosmos/gov/v1beta1/proposals/{proposal-id}

# Get proposal votes
GET /cosmos/gov/v1beta1/proposals/{proposal-id}/votes

# Get governance parameters
GET /cosmos/gov/v1beta1/params/{param-type}
```

### Transaction Endpoints

All governance transactions use the standard Cosmos transaction endpoint:

```http
POST /cosmos/tx/v1beta1/txs
```

## Governance in Practice

### Current Status

Based on the codebase, TLOCK has:

✅ **Standard Cosmos SDK governance module**
✅ **Parameter change proposals**
✅ **Community spend proposals**
✅ **Software upgrade coordination**
✅ **Validator and delegator voting**

### Not Currently Implemented

❌ **Custom Chief Moderator elections** (would require custom module)
❌ **Special moderation governance** (would require custom implementation)
❌ **Custom voting mechanisms** (uses standard Cosmos SDK only)

## Development and Testing

### Local Testing

```bash
# Start local node with governance
sh scripts/test_node.sh

# Submit test proposal
tlockd tx gov submit-proposal --type="Text" --title="Test" --description="Test proposal" --from node0 --keyring-backend test

# Vote on proposal
tlockd tx gov vote 1 yes --from node0 --keyring-backend test
```

### Governance Module State

Check governance state:

```bash
# Check governance module account
tlockd query auth account cosmos10d07y265gmmuvt4z0w9aw880jnsr700j6zn9kn

# Check community pool
tlockd query distribution community-pool
```

## Future Governance Extensions

To implement custom governance features mentioned in other documentation, TLOCK would need:

1. **Custom governance modules** for specialized voting
2. **Extended proposal types** for platform-specific decisions
3. **Integration with custom modules** (post, profile) for content governance
4. **Custom voting mechanisms** beyond standard Cosmos SDK

## Conclusion

TLOCK currently implements standard Cosmos SDK governance, providing a solid foundation for decentralized decision-making. The governance system allows for parameter changes, community spending, and coordinated upgrades through a proven, battle-tested framework.

Any advanced governance features described in other documentation would require additional development and are not currently implemented in the codebase.

---

*This governance documentation reflects the actual implementation in the TLOCK codebase, not planned or theoretical features.*
