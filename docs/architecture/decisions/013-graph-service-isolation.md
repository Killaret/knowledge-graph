# ADR 013: Graph Service Isolation

## Status
Accepted

## Context
Knowledge Graph is a graph-based note-taking system where relationships between notes are as important as the notes themselves. As the system evolved, several challenges emerged:

- **Complex graph operations** (traversal, neighbor search, pathfinding) are computationally expensive
- **Main backend became bloated** with graph-specific logic mixed with CRUD operations
- **Scaling concerns**: Graph operations have different performance characteristics than typical CRUD
- **Caching complexity**: Graph data changes infrequently but is read frequently, requiring specialized caching
- **Technology constraints**: Some graph algorithms are better suited for different programming languages/stacks

### Current State Analysis
The monolithic backend handles:
- CRUD operations for notes
- Link management
- Graph traversal and recommendation algorithms
- Embedding similarity search
- User authentication and authorization
- All other business logic

This creates several issues:
- Graph operations can slow down the entire backend
- Hard to scale graph operations independently
- Difficult to optimize graph algorithms without affecting other features
- Testing graph logic requires spinning up the entire backend

## Problem Statement
How do we isolate graph operations to improve performance, maintainability, and scalability while maintaining data consistency with the main system?

## Decision Drivers
- **Performance**: Graph operations should not impact CRUD operations
- **Scalability**: Ability to scale graph service independently
- **Maintainability**: Clear separation of concerns between graph and non-graph logic
- **Data Consistency**: Graph service must have access to consistent data
- **Development Velocity**: Changes to graph algorithms should not require redeploying entire backend
- **Technology Flexibility**: Ability to use optimal technologies for graph operations

## Considered Options

### Option 1: Keep Graph in Main Backend (Status Quo)
Continue handling all graph operations within the monolithic backend.

**Pros:**
- ✅ Simpler deployment (single service)
- ✅ Shared database access (no consistency concerns)
- ✅ Easier development for small team
- ✅ No network overhead for graph queries

**Cons:**
- ❌ Graph operations can slow down entire backend
- ❌ Cannot scale independently
- ❌ Harder to optimize graph algorithms
- ❌ Testing requires full backend stack
- ❌ Technology constraints (locked into Go stack)

### Option 2: Graph Service with API Calls (HTTP REST)
Extract graph logic into separate service, communicate via REST API.

**Pros:**
- ✅ Clear separation of concerns
- ✅ Independent scaling
- ✅ Can use different technology stack
- ✅ Easier to test in isolation
- ✅ Independent deployment

**Cons:**
- ❌ HTTP overhead for each graph query
- ❌ Network latency
- ❌ Serialization/deserialization cost
- ❌ Need for service discovery
- ❌ More complex deployment

### Option 3: Graph Service with gRPC + REST Hybrid
Extract graph logic into separate service with gRPC for internal communication and REST for external/monitoring.

**Pros:**
- ✅ gRPC is more efficient than REST (binary, HTTP/2)
- ✅ Strong typing with Protocol Buffers
- ✅ Bidirectional streaming support
- ✅ Independent scaling
- ✅ Technology flexibility
- ✅ REST for debugging/monitoring
- ✅ Clear service boundaries

**Cons:**
- ❌ More complex than monolithic approach
- ❌ Need to manage multiple services
- ❌ Network latency (though less than REST)
- ❌ Learning curve for gRPC

### Option 4: Graph Service with Direct Database Access
Graph service connects directly to PostgreSQL database for reading data, events for updates.

**Pros:**
- ✅ No API call overhead for data access
- ✅ Direct SQL allows query optimization
- ✅ Independent scaling
- ✅ Technology flexibility
- ✅ Can use database-specific features

**Cons:**
- ❌ Database becomes shared dependency
- ❌ Need to handle schema changes carefully
- ❌ Potential for inconsistent reads
- ❌ Tight coupling to database schema
- ❌ Bypasses business logic layer

## Decision
**Chosen Approach: Option 3 + Option 4 Hybrid**
- Separate Graph Service with gRPC for internal communication
- Direct database access for reads (performance optimization)
- REST endpoint for health checks and debugging
- Event-driven updates via Redis Pub/Sub for cache invalidation

### Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                     Main Backend (Go)                       │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐        │
│  │ Note CRUD   │  │ Auth/User   │  │  Events     │        │
│  └──────┬──────┘  └──────┬──────┘  └──────┬──────┘        │
│         │                │                │                 │
│         │                │                │ Redis Pub/Sub  │
│         │                │                │                 │
└─────────┼────────────────┼────────────────┼─────────────────┘
          │                │                │
          │ PostgreSQL     │                │
          ▼                ▼                ▼
┌─────────────────────────────────────────────────────────────┐
│                  PostgreSQL Database                         │
│  ┌─────────────┐  ┌─────────────┐                           │
│  │ notes table │  │ links table │                           │
│  └─────────────┘  └─────────────┘                           │
└─────────────────────────────────────────────────────────────┘
          │
          │ Direct Connection (Read-only)
          │
┌─────────▼─────────────────────────────────────────────────────┐
│                   Graph Service (Go)                         │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐        │
│  │ gRPC Server │  │ Graph Algos │  │  Cache      │        │
│  └─────────────┘  └─────────────┘  └─────────────┘        │
│         │                │                │                 │
│         │                │                │                 │
└─────────┼────────────────┼────────────────┼─────────────────┘
          │                │                │
          │ gRPC           │                │
          │                │                │
┌─────────▼────────────────┼────────────────┼─────────────────┐
│                   Frontend                               │
└─────────────────────────────────────────────────────────────┘
```

### Implementation Details

#### 1. gRPC Protocol Definition
```protobuf
service GraphService {
  rpc GetGraph(GraphRequest) returns (GraphResponse);
  rpc GetNeighbors(NeighborRequest) returns (NeighborResponse);
  rpc GetPath(PathRequest) returns (PathResponse);
  rpc GetRecommendations(RecommendationRequest) returns (RecommendationResponse);
}
```

#### 2. Direct Database Access
Graph service connects to PostgreSQL with read-only credentials:
- Optimized queries for graph operations
- Materialized views for common graph queries
- Connection pooling for performance

#### 3. Cache Invalidation
- Main backend publishes events on note/link changes
- Graph service subscribes to Redis Pub/Sub
- On receiving event, invalidates relevant cache entries
- Ensures cache consistency without polling

#### 4. REST Endpoint for Monitoring
- `/health` - health check
- `/metrics` - Prometheus metrics
- `/debug/graph/{noteId}` - debug graph data

## Consequences

### Positive Consequences
- ✅ **Performance**: Graph operations don't affect main backend performance
- ✅ **Scalability**: Can scale graph service independently based on load
- ✅ **Maintainability**: Clear separation of graph vs non-graph logic
- ✅ **Technology Flexibility**: Can use specialized graph libraries/algorithms
- ✅ **Development Velocity**: Graph algorithm changes don't require full backend redeployment
- ✅ **Testing**: Easier to test graph logic in isolation
- ✅ **Monitoring**: Dedicated metrics for graph operations

### Negative Consequences
- ❌ **Complexity**: Additional service to deploy and monitor
- ❌ **Network Latency**: gRPC calls have latency (though minimal)
- ❌ **Data Consistency**: Need event-driven cache invalidation
- ❌ **Operational Overhead**: Multiple services to manage
- ❌ **Development Setup**: Need to run multiple services locally
- ❌ **Database Coupling**: Graph service directly coupled to database schema

### Mitigation Strategies
- **Complexity**: Use Docker Compose for local development, Kubernetes for production
- **Network Latency**: Deploy services in same network/region, use connection pooling
- **Data Consistency**: Redis Pub/Sub for real-time cache invalidation
- **Operational Overhead**: Centralized logging (Loki), monitoring (Prometheus), tracing (Jaeger)
- **Database Coupling**: Version database schema carefully, use migration tools
- **Development Setup**: Provide `docker-compose.dev.yml` for local development

## References
- [ADR 001: Layered Architecture](./001-layered-architecture.md)
- [ADR 014: Event-Driven Cache Invalidation](./014-event-driven-cache-invalidation.md)
- [RECOMMENDATION_ARCHITECTURE.md](../../RECOMMENDATION_ARCHITECTURE.md)
