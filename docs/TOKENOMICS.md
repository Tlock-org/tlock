# TLOCK Tokenomics

## Overview

The TLOCK ecosystem is powered by the native token **TOK**, a versatile asset designed to facilitate all interactions within our decentralized social media platform. TOK serves multiple critical functions while maintaining a sustainable economic model through innovative mechanisms.

## Token Utility

### Primary Functions

**Security Incentives**
- Rewards block validators and delegators
- Ensures blockchain security and integrity
- Staking rewards for network participation

**Content Incentives**
- Rewards users for creating posts
- Incentivizes likes and comments
- Encourages high-quality content generation

### Secondary Functions

**Advertising Payments**
- Advertisers pay in TOK for maximum exposure
- Ad revenue flows back to reward pool
- Creates sustainable revenue cycle

**Moderation Salaries**
- Moderators receive monthly salaries in TOK
- Incentivizes quality content maintenance
- Supports decentralized governance

**Governance Participation**
- Proposal creation requires TOK staking
- Voting conducted using TOK
- Enables transparent decision-making

**Fee Grant Operations**
- Module holds TOK for periodic allowances
- Reduces onboarding barriers for new users
- Promotes seamless ecosystem entry

## Token Distribution

| Allocation | Amount | Percentage | Description |
|------------|--------|------------|-------------|
| **Content Reward Pool** | 100 billion TOK | 66.7% | Dynamic halving mechanism |
| **Community Airdrop** | 30 billion TOK | 20.0% | Community building and adoption |
| **Private/Public Sale** | 7 billion TOK | 4.7% | Development funding |
| **Team Allocation** | 7 billion TOK | 4.7% | Team incentives and retention |
| **Community Reserve** | 6 billion TOK | 4.0% | Future community initiatives |
| **Total Supply** | **150 billion TOK** | **100%** | Fixed maximum supply |

### Team Allocation Vesting
- **3 billion TOK**: Unlock upon mainnet launch
- **4 billion TOK**: Unlock after 1 year cliff

## Content-to-Earn Mechanism

TLOCK emits two concurrent reward streams for post interactions:

### 1. Instant Engagement Rewards

**Immediate Transfer**: Rewards transferred instantly to users who like or comment

**Reward Formula**:
```
Rate_n = 2^(n-1) / 100
Reward = Rate_n × RewardBase
```

**Base Amounts**:
- **Comments**: 100 TOK per comment
- **Likes**: 1 TOK per like

**Daily Limits**:
- First **5 comments** per wallet per 24 hours are reward-bearing
- First **20 likes** per wallet per 24 hours are reward-bearing

**Reputation Multiplier Examples**:
- **Level 1**: 1 TOK per comment, 0.01 TOK per like
- **Level 7**: 64 TOK per comment, 0.64 TOK per like
- **Level 7+**: Rate capped at 100%

**Stake Bypass Option**:
- Stake 100,000 TOK to receive full 100% payout
- Bypasses reputation multiplier system
- Reverts to reputation-based rate when stake withdrawn

### 2. Accruing Creator Rewards

**Post Author Credits**: Every rewarded interaction credits the post's author

**Unlock Period**: 7 days after accumulation

**Scaling**: No upper bound - viral posts scale linearly

**Verification Requirement**: Interactions from level 0 (unverified) wallets ignored

## Dynamic Halving Mechanism

### Overview

Advanced adaptive system that adjusts user rewards based on reward pool health, ensuring long-term sustainability and network growth alignment.

### Core Components

**Reward Pool**: 100 billion TOK held in secure module account

**Initial Rewards**: Users earn base amounts (e.g., 100 TOK per comment)

**Dynamic Adjustment**: Automatic halving when pool balance drops below thresholds

**Revenue Integration**: Advertising fees replenish pool and can reverse halvings

### Halving Threshold Formula

```
T_n = S × (1/2)^n
```

Where:
- **S** = Initial reward pool amount (100 billion TOK)
- **n** = Number of halving events occurred

**Examples**:
- **T1** = 50 billion TOK (first halving threshold)
- **T2** = 25 billion TOK (second halving threshold)
- **T3** = 12.5 billion TOK (third halving threshold)

### Reward Adjustment Formula

```
R_n = R_0 × (1/2)^n
```

Where:
- **R_0** = Initial reward per action
- **n** = Number of halving events

**Example**: After first halving, comment reward = 100 × 0.5 = 50 TOK

### Cooling Mechanism

**Purpose**: Prevents rapid oscillatory halving from short-term fluctuations

**Duration**: 14-day cooling period after each halving event

**Benefit**: Allows time for reward pool recovery through advertising income

### Adaptive Reversible Halving

**Halving Events**:
- Automatic when pool balance falls below thresholds
- Moderates token outflow during heavy distribution periods

**Reversal Capability**:
- Advertising fees can restore pool balance above previous thresholds
- System can reverse halved state and restore higher reward rates
- Creates flexible, self-balancing mechanism

**Sustainability**: Continuously balances reward outflows with advertising income

## Economic Model Analysis

### Revenue Streams

1. **Advertising Revenue**: Paid in TOK, flows back to reward pool
2. **Premium Features**: Paid image storage, enhanced visibility
3. **Governance Fees**: Proposal submission costs
4. **Staking Rewards**: Network security incentives

### Cost Structure

1. **Content Rewards**: Largest expense, controlled by dynamic halving
2. **Validator Rewards**: Network security costs
3. **Moderation Costs**: Decentralized team salaries
4. **Development Costs**: Ongoing platform improvements

### Sustainability Mechanisms

**Dynamic Halving**: Prevents reward pool depletion while maintaining incentives

**Revenue Recycling**: Ad revenue directly supports reward pool

**Reputation Gating**: Prevents Sybil attacks and reward farming

**Stake Requirements**: High-value actions require skin in the game

## Token Metrics

### Supply Dynamics

- **Fixed Maximum Supply**: 150 billion TOK
- **Circulating Supply**: Increases as rewards are distributed
- **Burn Mechanisms**: None currently implemented
- **Deflationary Pressure**: Staking and governance lock-ups

### Velocity Considerations

**High Velocity Factors**:
- Daily reward distributions
- Active trading for content rewards
- Advertising payment flows

**Low Velocity Factors**:
- Staking for governance participation
- Long-term holding for reputation benefits
- Creator reward 7-day lock periods

## Governance Integration

### Proposal System

**Proposal Types**:
- Parameter changes 
- Chief Moderator elections
- Community fund allocations
- Protocol upgrades

**Voting Power**: Proportional to TOK holdings

**Quorum Requirements**: Minimum participation thresholds

**Execution**: Automatic implementation of passed proposals

### Decentralized Decision Making

**Community Ownership**: Token holders control platform direction

**Transparent Process**: All governance actions on-chain

**Stakeholder Alignment**: Decisions benefit token holders and users



## Conclusion

TLOCK's tokenomics design creates a sustainable, self-regulating economic system that aligns incentives across all participants. The dynamic halving mechanism ensures long-term viability while the multi-utility token design creates strong network effects.

By combining innovative reward mechanisms with proven economic principles, TLOCK establishes a foundation for the next generation of decentralized social media platforms.

---

*This tokenomics model is subject to governance decisions and may evolve based on community feedback and network requirements.*
