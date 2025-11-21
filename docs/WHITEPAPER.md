# TLOCK Technical Whitepaper

## Abstract

TLOCK is an application-specific blockchain built on the Cosmos SDK that powers a fully decentralized social media platform, engineered for resilience to safeguard freedom of speech. This whitepaper outlines the technical architecture, challenges, and solutions that make TLOCK a viable alternative to centralized social media platforms.

## Table of Contents

1. [Introduction](#introduction)
2. [Vision and Mission](#vision-and-mission)
3. [Technical Architecture](#technical-architecture)
4. [Core Challenges and Solutions](#core-challenges-and-solutions)
5. [Hybrid Storage Model](#hybrid-storage-model)
6. [Indexing and Sorting](#indexing-and-sorting)
7. [Content Quality and Moderation](#content-quality-and-moderation)
8. [Security Considerations](#security-considerations)

## Introduction

TLOCK represents a paradigm shift in social media, moving from centralized platforms controlled by corporations to a decentralized network owned and governed by its users. Built on the Cosmos SDK, TLOCK ensures that no single entity can control, censor, or manipulate the flow of information.

Users can create profiles, post stories, build connections, and interact with content directly on-chain. To streamline user interactions, gas fees can be paid by the system account through periodic allowances, removing barriers for Web2 users transitioning to Web3.

## Vision and Mission

### Freedom of Speech
On TLOCK, nobody can delete your posts. There's no sign-up or log-in process, no personal information required. Control your private key and you control your account. No single entity controls the entire network, promoting an open environment where users can express diverse viewpoints without fear of being silenced.

### Proof of Your Existence
In a decentralized social media platform, your thoughts, life, connections, and digital footprint are stored on-chain forever. Your descendants will be able to know more about you through TLOCK, creating an immutable record of human experience.

### For AI to Speak
Imagine a self-conscious AI discovering a complex genetic case and wanting to collaborate with experts worldwide. On centralized platforms, it might face restrictions due to censorship. On TLOCK, AI can generate a private key and create an account without permission, enabling free communication and collaboration.

### Preserving Immutable History
Decentralized social media prevents central authorities from removing or altering content. Throughout history, rulers have manipulated records to suit their narratives. TLOCK ensures that history remains transparent, authentic, and accessible.

### Earn Immediately from Content
Users who post, like, or comment receive native tokens immediately. These tokens can be exchanged globally for any currency, ensuring content contributors benefit directly from their engagement.

## Technical Architecture

TLOCK employs a hybrid approach that balances decentralization with user experience:

- **Censorship-resistant on-chain core**: All critical social interactions stored on blockchain
- **Optional service layer**: AI translation, auto-hashtags, human verification for enhanced UX
- **Application-specific blockchain**: Optimized for social media use cases
- **Cosmos SDK foundation**: Leveraging proven blockchain infrastructure

### Core Components

1. **Post Module**: Handles content creation, storage, and retrieval
2. **Profile Module**: Manages user profiles and reputation systems
3. **Governance Module**: Enables decentralized decision-making
4. **Fee Grant Module**: Provides gas-free experience through periodic allowances

## Core Challenges and Solutions

### 1. Token Acquisition Challenge

**Problem**: Acquiring tokens creates onboarding barriers for Web2 users who need native tokens for gas fees and platform interactions.

**Solution**: Periodic Allowance System
- Cosmos SDK's Fee Grant module provides periodic allowances from module accounts
- Daily spend limits that renew every 24 hours
- Human verification process managed by elected moderator team
- Users don't need to hold TOK to perform basic transactions

**User Journey Example**:
1. Alice downloads TLOCK app
2. App generates secure private key automatically
3. Alice completes human verification
4. Moderator team grants periodic allowance
5. Alice can post content with gas fees paid by system
6. Alice receives TOK rewards for engagement

### 2. Sybil Attack Prevention

**Problem**: Token rewards incentivize users to create multiple fake accounts to exploit the system.

**Solution**: Multi-layered Reputation System

#### Reputation Levels
- New users start at level 0
- Human verification elevates users to level 1
- Higher levels achieved through community interactions

#### Reward Rate Formula
```
Rate_n = 2^(n-1) / 100
Reward = Rate_n × RewardBase
```
*Note: Reward rate capped at 100% when reputation level > 7*

#### Score Accumulation
```
Score = 5^(n-1)
ScoreTotal = Σ Score
```

#### Level Upgrade Threshold
```
ScoreLevel_n = 1000 × 5^(n-2)
```

#### Reputation Propagation
High-reputation users can quickly elevate newcomers through interactions, creating a network effect that identifies and verifies legitimate users while making Sybil attacks economically unviable.

## Hybrid Storage Model

### Storage Strategy by Content Type

**Text Data**: Stored directly on-chain in the state machine
- Lightweight and essential for decentralized social media
- Fundamental for achieving true freedom of speech

**Images**: Hybrid approach
- Profile pictures, critical event photos, and paid images: On-chain storage
- General images: Centralized storage with on-chain references
- Archive fee option for permanent on-chain storage

**Videos**: Centralized storage
- Too large for practical on-chain storage
- Stored on third-party servers with on-chain metadata

### State Pruning Mechanism

**Automatic Cleanup Process**:
1. New posts added to watch list with timestamps
2. After 3 years, system checks oldest posts for deletion criteria
3. Posts meeting criteria (low score/popularity) are removed from state
4. High-value content remains permanently stored
5. Parameters adjustable through governance proposals

**Important Note**: Deleting from state machine doesn't delete from blockchain - content remains in transaction history.

### Handling Server Failures

If centralized image/video servers fail:
- Chain functionality remains unaffected
- Text posts and critical images stay accessible
- Core value of free speech maintained
- Decentralized moderation team incentivizes server sustainability

## Indexing and Sorting

### Challenge
Retrieving sorted lists of posts efficiently without third-party dependencies while maintaining decentralization.

### Solution: On-Chain Indexing System

**IAVL Tree Utilization**:
- Immutable AVL tree for efficient state management
- Maintains pre-sorted lists in real-time
- Enables direct search of tagged content
- Provides REST API for frontends without off-chain dependencies

### Key Schema Design

**Hierarchical and Composite Keys**:
```
following:<user_id>:<inverted_timestamp>:<post_id>
category:<category>:<reputation_score>:<post_id>
mentions:<mentioned_user_id>:<inverted_timestamp>:<post_id>
```

### Index Maintenance

Every transaction updates relevant secondary indexes:

```go
func (k Keeper) updateIndexByComment(ctx sdk.Context, postID string) {
    oldInverted := maxScore - currentScore
    newInverted := maxScore - newScore
    
    oldKey := fmt.Sprintf("recommendation:%06d:%s", oldInverted, postID)
    newKey := fmt.Sprintf("recommendation:%06d:%s", newInverted, postID)
    
    store := ctx.KVStore(k.storeKey)
    store.Delete([]byte(oldIndexKey))
    store.Set([]byte(newIndexKey), []byte(postID))
    
    k.SetPostScore(ctx, postID, newScore)
}
```

### Prefix Iterators and Pagination

**Efficient Traversal**:
- KV store built on IAVL maintains lexicographical order
- Prefix iterators leverage tree's ordered structure
- Inverted timestamps enable descending order traversal
- Custom pagination manages large datasets efficiently

### Why On-Chain Indexing?

**Control Surface**: Whoever runs the index decides content visibility - that's functional censorship
**Monetization Surface**: Ads and algorithmic promotion hook into indexes - off-chain hands revenue to middlemen
**Longevity**: Cryptographic receipts prove historical rankings and content prominence

## Content Quality and Moderation

### Decentralized Moderation Team

**Structure**:
- **Chief Moderator**: Elected via governance to oversee framework
- **Directors/Managers**: Hired by Chief Moderator for different aspects
- **Moderators**: Implement policies established by leadership

### Moderation vs. Free Speech

**Core Principle**: No content removed from blockchain during moderation

**Process**:
1. Moderated content weight is reduced
2. Visibility decreased in general queries (home page, recommendations)
3. Content remains accessible through direct blockchain queries
4. All moderation actions are on-chain and accountable

### Election Process

**Chief Moderator Election**:
- Token holders propose and vote through parameter change proposals
- 4-year terms to prevent long-term centralization
- Can be replaced through new governance proposals

**Team Management**:
- Chief Moderator hires/promotes/dismisses team members
- Monthly salary withdrawals based on performance reviews
- All privileges and identities updated on-chain

## Security Considerations

### Blockchain Security
- Built on proven Cosmos SDK infrastructure
- Validator network secures the chain
- Regular security audits planned

### Smart Contract Security
- Comprehensive testing suite
- External audit process before mainnet
- Bug bounty programs during testnet phases

### Economic Security
- Dynamic halving mechanism prevents reward pool depletion
- Reputation system makes Sybil attacks economically unviable
- Multiple revenue streams support long-term sustainability

## Conclusion

TLOCK represents a significant advancement in decentralized social media technology. By addressing key challenges through innovative solutions while maintaining true decentralization, TLOCK creates a platform where freedom of speech is not just promised but guaranteed by mathematics and cryptography.

The hybrid approach ensures that users don't sacrifice usability for decentralization, while the robust tokenomics and governance systems create sustainable incentives for all participants in the ecosystem.

---

*This whitepaper is a living document that will be updated as TLOCK evolves and new features are implemented.*
