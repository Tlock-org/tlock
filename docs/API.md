# TLOCK API Documentation

## Overview

TLOCK provides comprehensive REST and gRPC APIs for interacting with the decentralized social media platform. All APIs are automatically generated from Protocol Buffer definitions and follow Cosmos SDK standards.

## Table of Contents

1. [Base Configuration](#base-configuration)
2. [Authentication](#authentication)
3. [Post Module APIs](#post-module-apis)
4. [Profile Module APIs](#profile-module-apis)
5. [Standard Cosmos SDK APIs](#standard-cosmos-sdk-apis)
6. [Data Structures](#data-structures)
7. [Error Handling](#error-handling)

## Base Configuration

### Network Information
- **Chain ID**: `localchain-1` (testnet) / `tlock-1` (mainnet)
- **Bech32 Prefix**: `tlock`
- **Native Token**: `TOK`
- **Cosmos SDK Version**: v0.50
- **Tendermint Version**: v0.38

### Endpoints
- **REST API**: `http://localhost:1317` (local) / `https://api.tlock.org` (mainnet)
- **RPC**: `http://localhost:26657` (local) / `https://rpc.tlock.org` (mainnet)
- **gRPC**: `localhost:9090` (local) / `grpc.tlock.org:443` (mainnet)

## Authentication

TLOCK uses Cosmos SDK standard authentication with cryptographic signatures. No API keys required.

### Transaction Signing
All write operations require proper transaction signing using your wallet's private key.

### Periodic Allowance
Users with granted periodic allowances can perform transactions without holding TOK tokens.

## Post Module APIs

### Query Endpoints (GET)

#### Get Module Parameters
```http
GET /post/v1/params
```
**Response**: Module configuration parameters

#### Resolve Name
```http
GET /post/v1/name/{address}
```
**Parameters**:
- `address`: User's wallet address

**Response**:
```json
{
  "name": "user_display_name"
}
```

#### Get Home Posts
```http
GET /post/v1/homePosts/{page_size}
```
**Parameters**:
- `page_size`: Number of posts per page

**Response**:
```json
{
  "page": 1,
  "posts": [PostResponse]
}
```

#### Get First Page Home Posts
```http
GET /post/v1/firstPageHomePosts/{page}/{page_size}
```
**Parameters**:
- `page`: Page number
- `page_size`: Number of posts per page

#### Search Topics
```http
GET /post/v1/topics/{matching}
```
**Parameters**:
- `matching`: Search term for topic matching

**Response**:
```json
{
  "topics": [TopicResponse]
}
```

#### Get Topic Posts
```http
GET /post/v1/topic/posts/{topic_id}/{page}
```
**Parameters**:
- `topic_id`: Topic identifier
- `page`: Page number

#### Get Specific Post
```http
GET /post/v1/get/{post_id}
```
**Parameters**:
- `post_id`: Post identifier

**Response**:
```json
{
  "post": PostResponse,
  "topics": [TopicResponse]
}
```

#### Get User Created Posts
```http
GET /post/v1/user/created/{address}/{page}
```
**Parameters**:
- `address`: User's wallet address
- `page`: Page number

#### Get Likes I Made
```http
GET /post/v1/likes/i/made/{address}/{page}
```

#### Get Saves I Made
```http
GET /post/v1/saves/i/made/{address}/{page}
```

#### Get Likes Received
```http
GET /post/v1/likes/received/{address}/{page}
```

#### Get Comments
```http
GET /post/v1/comments/{id}/{page}
```
**Parameters**:
- `id`: Post or comment ID
- `page`: Page number

#### Get Comments Received
```http
GET /post/v1/comments/received/{address}/{page}
```

#### Get Activities Received
```http
GET /post/v1/activities/received/{address}/{page}
```

#### Get Categories
```http
GET /post/v1/categories
```

**Response**:
```json
{
  "categories": [CategoryResponse]
}
```

#### Get Topics by Category
```http
GET /post/v1/category/topics/{category_id}/{page}
```

#### Get Category by Topic
```http
GET /post/v1/topic/category/{topic_id}
```

#### Get Category Posts
```http
GET /post/v1/category/posts/{category_id}/{page}
```

#### Get Following Posts
```http
GET /post/v1/following/posts/{address}/{page}
```

#### Get Following Topics
```http
GET /post/v1/following/topics/{address}/{page}
```

#### Check if Following Topic
```http
GET /post/v1/isFollowing/topic/{address}/{topic_id}
```

**Response**:
```json
{
  "is_following": true
}
```

#### Get Uncategorized Topics
```http
GET /post/v1/uncategorized/topics/{address}/{page}/{limit}
```

#### Get Vote Option
```http
GET /post/v1/vote/option/{address}/{post_id}
```

#### Get Topic Image
```http
GET /post/v1/topic/image/{topic_id}
```

#### Get Topic Details
```http
GET /post/v1/topic/{topic_id}
```

#### Get Trending Keywords
```http
GET /post/v1/trending/keywords/{page}
```

#### Get Trending Topics
```http
GET /post/v1/trending/topics/{page}
```

#### Get Paid Post Image
```http
GET /post/v1/paid/image/{image_id}
```

### Transaction Endpoints (POST)

All transaction endpoints require proper Cosmos SDK transaction formatting and signing.

#### Create Post
**Endpoint**: `/cosmos/tx/v1beta1/txs`
**Message Type**: `MsgCreatePost`

**Message Structure**:
```json
{
  "creator": "tlock1...",
  "post_detail": {
    "title": "Post Title",
    "content": "Post content",
    "imagesBase64": ["base64_encoded_image"],
    "imagesUrl": ["https://example.com/image.jpg"],
    "videosUrl": ["https://example.com/video.mp4"],
    "quote": "post_id_to_quote",
    "mention": ["tlock1address1", "tlock1address2"],
    "topic": ["topic1", "topic2"],
    "category": "category_name",
    "poll": {
      "totalVotes": 0,
      "votingStart": 1640995200,
      "votingEnd": 1641081600,
      "vote": [
        {"option": "Option 1", "count": 0, "id": 1},
        {"option": "Option 2", "count": 0, "id": 2}
      ]
    }
  }
}
```

#### Like Post
**Message Type**: `MsgLikeRequest`
```json
{
  "sender": "tlock1...",
  "id": "post_id"
}
```

#### Unlike Post
**Message Type**: `MsgUnlikeRequest`
```json
{
  "sender": "tlock1...",
  "id": "post_id"
}
```

#### Comment on Post
**Message Type**: `MsgCommentRequest`
```json
{
  "creator": "tlock1...",
  "parentId": "post_id",
  "comment": "Comment content",
  "mention": ["tlock1address1"]
}
```

#### Quote Post
**Message Type**: `MsgQuotePostRequest`
```json
{
  "creator": "tlock1...",
  "quote": "post_id_to_quote",
  "comment": "Quote comment",
  "mention": ["tlock1address1"],
  "topic": ["topic1"],
  "category": "category_name"
}
```

#### Repost
**Message Type**: `MsgRepostRequest`
```json
{
  "creator": "tlock1...",
  "quote": "post_id_to_repost",
  "mention": ["tlock1address1"],
  "topic": ["topic1"],
  "category": "category_name"
}
```

#### Save Post
**Message Type**: `MsgSaveRequest`
```json
{
  "sender": "tlock1...",
  "id": "post_id"
}
```

#### Cast Vote on Poll
**Message Type**: `CastVoteOnPollRequest`
```json
{
  "creator": "tlock1...",
  "id": "post_id",
  "optionId": 1
}
```

#### Follow Topic
**Message Type**: `MsgFollowTopicRequest`
```json
{
  "creator": "tlock1...",
  "topic_id": "topic_id"
}
```

#### Grant Allowance from Module
**Message Type**: `MsgGrantAllowanceFromModuleRequest`
```json
{
  "sender": "tlock1admin...",
  "userAddress": "tlock1user..."
}
```

## Profile Module APIs

### Query Endpoints (GET)

#### Get Module Parameters
```http
GET /profile/v1/params
```

#### Get Profile
```http
GET /profile/v1/get/{address}
```
**Response**:
```json
{
  "profile": {
    "wallet_address": "tlock1...",
    "user_handle": "username",
    "nickname": "Display Name",
    "avatar": "base64_or_url",
    "has_avatar": true,
    "bio": "User bio",
    "level": 5,
    "admin_level": 0,
    "following": 150,
    "followers": 300,
    "creation_time": 1640995200,
    "location": "Location",
    "website": "https://example.com",
    "idVerification_status": "ID_VERIFICATION_PERSONAL",
    "score": 15625,
    "line_manager": ""
  }
}
```

#### Get Profile Avatar
```http
GET /profile/v1/avatar/{address}
```

#### Get Following List
```http
GET /profile/v1/following/{address}/{page}/{limit}
```

#### Get Followers List
```http
GET /profile/v1/followers/{address}
```

#### Get Follow Relationship
```http
GET /profile/v1/follow/relationship/{addressA}/{addressB}
```
**Response**:
```json
{
  "relationship": 1
}
```
*Relationship codes: 0-No relationship; 1-A follows B; 2-B follows A; 3-Mutual follows*

#### Get Mention Suggestions
```http
GET /profile/v1/getMentionSuggestions/{address}/{matching}
```

#### Get Activities Received Count
```http
GET /profile/v1/activities/received/count/{address}
```

#### Search Users
```http
GET /profile/v1/users/search/{matching}
```

#### Check if Admin
```http
GET /profile/v1/isAdmin/{address}
```

### Transaction Endpoints (POST)

#### Add/Update Profile
**Message Type**: `MsgAddProfileRequest`
```json
{
  "creator": "tlock1...",
  "profile_json": {
    "nickname": "Display Name",
    "user_handle": "username",
    "avatar": "base64_encoded_image",
    "bio": "User biography",
    "location": "City, Country",
    "website": "https://example.com"
  }
}
```

#### Follow User
**Message Type**: `MsgFollowRequest`
```json
{
  "creator": "tlock1...",
  "targetAddr": "tlock1target..."
}
```

#### Unfollow User
**Message Type**: `MsgUnfollowRequest`
```json
{
  "creator": "tlock1...",
  "targetAddr": "tlock1target..."
}
```

#### Add Admin
**Message Type**: `MsgAddAdminRequest`
```json
{
  "creator": "tlock1...",
  "address": "tlock1newadmin..."
}
```

#### Manage Admin
**Message Type**: `MsgManageAdminRequest`
```json
{
  "creator": "tlock1...",
  "manage_json": {
    "line_manager": "tlock1manager...",
    "admin_address": "tlock1admin...",
    "admin_level": 2,
    "editable": true
  },
  "action": "update"
}
```

## Standard Cosmos SDK APIs

TLOCK includes all standard Cosmos SDK modules with their respective APIs:

### Bank Module
- **Base Path**: `/cosmos/bank/v1beta1/`
- **Endpoints**: balances, supply, metadata, etc.

### Auth Module
- **Base Path**: `/cosmos/auth/v1beta1/`
- **Endpoints**: accounts, params

### Staking Module
- **Base Path**: `/cosmos/staking/v1beta1/`
- **Endpoints**: validators, delegations, unbonding

### Governance Module
- **Base Path**: `/cosmos/gov/v1beta1/`
- **Endpoints**: proposals, votes, params

### Distribution Module
- **Base Path**: `/cosmos/distribution/v1beta1/`
- **Endpoints**: rewards, commission

### Fee Grant Module
- **Base Path**: `/cosmos/feegrant/v1beta1/`
- **Endpoints**: allowances

### IBC Module
- **Base Path**: `/ibc/`
- **Endpoints**: channels, connections, clients

### Transaction Broadcasting
```http
POST /cosmos/tx/v1beta1/txs
```

### Transaction Query
```http
GET /cosmos/tx/v1beta1/txs/{hash}
```

### Node Info
```http
GET /cosmos/base/tendermint/v1beta1/node_info
```

### Block Info
```http
GET /cosmos/base/tendermint/v1beta1/blocks/{height}
```

## Data Structures

### PostResponse
```json
{
  "post": {
    "id": "post_123",
    "post_type": "ORIGINAL",
    "parent_id": "",
    "title": "Post Title",
    "content": "Post content",
    "creator": "tlock1...",
    "timestamp": 1640995200,
    "image_ids": ["img_1", "img_2"],
    "images_url": ["https://example.com/img.jpg"],
    "videos_url": ["https://example.com/vid.mp4"],
    "quote": "quoted_post_id",
    "like_count": 42,
    "comment_count": 15,
    "repost_count": 8,
    "save_count": 23,
    "score": 1250,
    "homePostsUpdate": 1640995200,
    "poll": {
      "totalVotes": 100,
      "votingStart": 1640995200,
      "votingEnd": 1641081600,
      "vote": [
        {"option": "Yes", "count": 60, "id": 1},
        {"option": "No", "count": 40, "id": 2}
      ]
    }
  },
  "profile": Profile,
  "quote_post": Post,
  "quote_profile": Profile
}
```

### Profile
```json
{
  "wallet_address": "tlock1...",
  "user_handle": "username",
  "nickname": "Display Name",
  "avatar": "base64_or_url",
  "has_avatar": true,
  "bio": "User biography",
  "level": 5,
  "admin_level": 0,
  "following": 150,
  "followers": 300,
  "creation_time": 1640995200,
  "location": "City, Country",
  "website": "https://example.com",
  "idVerification_status": "ID_VERIFICATION_PERSONAL",
  "score": 15625,
  "line_manager": ""
}
```

### TopicResponse
```json
{
  "id": "topic_123",
  "name": "blockchain",
  "image": "base64_or_url",
  "title": "Blockchain Technology",
  "summary": "Discussion about blockchain",
  "score": 5000,
  "trending_keywords_score": 1200,
  "category_id": "tech",
  "creator": "tlock1..."
}
```

### CategoryResponse
```json
{
  "id": "tech",
  "name": "Technology",
  "avatar": "base64_or_url",
  "index": 1
}
```

### CommentResponse
```json
{
  "post": Post,
  "profile": Profile,
  "targetProfile": Profile
}
```

## Error Handling

### Standard HTTP Status Codes
- **200**: Success
- **400**: Bad Request - Invalid parameters
- **404**: Not Found - Resource doesn't exist
- **500**: Internal Server Error

### Cosmos SDK Error Format
```json
{
  "code": 2,
  "message": "account sequence mismatch",
  "details": []
}
```

### Common Errors

#### Transaction Errors
```json
{
  "code": 4,
  "message": "insufficient funds",
  "details": []
}
```

#### Validation Errors
```json
{
  "code": 400,
  "message": "post content exceeds maximum length",
  "details": []
}
```

#### Authentication Errors
```json
{
  "code": 401,
  "message": "signature verification failed",
  "details": []
}
```

## Rate Limiting

### Default Limits
- **Query Endpoints**: 1000 requests per hour per IP
- **Transaction Endpoints**: 100 transactions per hour per address
- **Search Endpoints**: 500 requests per hour per IP

### Headers
```http
X-RateLimit-Limit: 1000
X-RateLimit-Remaining: 950
X-RateLimit-Reset: 1640998800
```

## Development Tools

### CLI Commands
```bash
# Query examples
tlockd query post get post_123
tlockd query profile get tlock1...

# Transaction examples
tlockd tx post create-post --from mykey
tlockd tx profile add-profile --from mykey
```

### gRPC Clients
```bash
# Using grpcurl
grpcurl -plaintext localhost:9090 list
grpcurl -plaintext localhost:9090 post.v1.Query/QueryPost
```

### Swagger/OpenAPI
- **Local**: `http://localhost:1317/swagger/`
- **Mainnet**: `https://api.tlock.org/swagger/`

## Support and Resources

### Documentation
- **Cosmos SDK Docs**: https://docs.cosmos.network
- **Protocol Buffers**: https://developers.google.com/protocol-buffers

### Community
- **Discord**: https://discord.gg/tlock
- **GitHub**: https://github.com/Tlock-org/tlock
- **API Issues**: https://github.com/Tlock-org/tlock/issues

---

*This API documentation is automatically generated from Protocol Buffer definitions and reflects the actual implementation in the codebase.*
