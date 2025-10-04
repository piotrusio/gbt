# OGBT Project Summary

## Vision

Open infrastructure for supply chain safety. Manufacturing companies cannot trace materials through their supply chains because systems are disconnected. OGBT provides unified batch genealogy across internal systems (MES/ERP/WMS) and enables peer-to-peer supply chain communication via blockchain, eliminating centralized vendor intermediaries.

**Mission:** Safer medicines, safer food, safer products. Make critical traceability infrastructure open, shared, and permanent.

## Core Principles

- **Open Source First** - Apache 2.0, no feature gating
- **Mission Over Money** - Truly Open
- **Technical Excellence** - 1000x performance advantage is the moat
- **Long-term Build** - 5+ year timeline, sustainablee pace
- **No Vendor Lock-in** - Companies control their data and deployment

**Everything Critical = Free:**
- Core platform and graph engine
- All adapters (MES/ERP/WMS)
- Compliance modules (GxP, 21 CFR Part 11)
- Blockchain interface
- Documentation

## High Level Architecture

### Core Components

**1. Adapters** (Source System Integration)
- Pluggable binaries per system (SAP, Siemens, Rockwell, etc)
- Normalize source data → unified NATS schema
- Community extensible
- Deploy only what's needed

**2. Event Store** (NATS JetStream)
- Append-only event log
- Source of truth for all batch movements
- Replay capability for graph rebuilds
- Snapshots for fast recovery

**3. Master Data Service**
- KV storage (Valkey/Redis)
- REST API for CRUD operations
- Material/location/equipment mappings
- Enables cross-system correlation (MES ↔ WMS ↔ ERP)
- Live updates (not static config)

**4. Graph Engine** (Go, in-memory)
- **Data Structure:** CSR (Compressed Sparse Row) with int64 IDs
- **Multi-resolution:** ERP level → MES detail → WMS physical paths
- **Capacity:** 5B edges @ 256GB RAM
- **Performance:** 300k nodes traversal in 200ms (epoch trick + goroutines)
- **Scaling:**
  - Vertical: More RAM → more edges
  - Horizontal: Read replicas → more concurrent queries

**5. Query Engine** (Dual approach)
- **Graph traversal:** Topology queries (fast, integer operations)
- **DuckDB enrichment:** Join graph results with metadata (SQL flexibility)
- Combined flow: Graph IDs → DuckDB metadata join → Full results

**6. Watcher/Agent** (Intelligent Monitoring)
- OpenTelemetry ingestion (metrics, traces, logs)
- Quality checks (data completeness, anomalies)
- Auto-remediation triggers
- ML-enhanced anomaly detection (future)
- **"Intelligent Batch Tracing"** - learns, watches, acts

**7. Publishers** (Output)
- Broadcast OGBT events
- Webhooks, message queues
- Audit trails, external integrations

**8. APIs** (External Interface)
- REST, GraphQL, WebSocket
- Multi-protocol for different use cases
- Real-time updates via WebSocket

### Data Flow

```
External Systems → Adapters → NATS JetStream → Event Processing
                                    ↓
                            Graph Engine (RAM)
                                    ↓
                    Query Engine (Graph + DuckDB)
                                    ↓
                            APIs → Users
                                    ↓
                            Publishers → External Systems

Master Data Service ← Adapters (push mappings)
Master Data Service → All components (lookup)

Watcher/Agent monitors all components via OpenTelemetry
```

### Storage Strategy

**Event Store:**
- NATS JetStream (persistent, replay-able)
- Snapshots for fast graph rebuilds

**Graph:**
- In-memory CSR structure (integers only)
- Metadata separated (Parquet files)
- Hot/warm/cold tiering for historical data

**Metadata:**
- Parquet files (columnar, compressed)
- DuckDB for SQL queries
- Embedded, zero ops

**Master Data:**
- Valkey/Redis (KV store)
- Fast lookups, live updates

### Multi-Resolution Graph

**Key Insight:** Same flow, different granularities:

**ERP Level (coarse):**
```
BatchA --[Issue→Receipt]--> BatchB (1 edge)
```

**MES Level (fine):**
```
BatchA → WIP-1 → WIP-2 → WIP-3 → WIP-4 → WIP-5 → BatchB (6 nodes)
```

**WMS Level (fine):**
```
Location-A → Location-B → Location-C → Location-D (physical path)
```

Query resolution determines depth. Graph structure supports zoom in/out.

## Technical Decisions

**Language:** Go
- Performance (graph traversal)
- Goroutines (concurrent queries)
- Single binary deployment

**Graph Structure:** CSR with int64
- Tiny, cache-friendly
- 5B edges capacity
- Integer-only operations during traversal

**Event Sourcing:** NATS JetStream
- Already in stack for async
- Persistent streams
- Replay capability

**Metadata:** DuckDB + Parquet
- Embedded, zero DB ops
- SQL flexibility
- Columnar storage efficiency

**Master Data:** Valkey/Redis
- Fast KV lookups
- Live updates
- Standard tooling

**Observability:** OpenTelemetry
- Industry standard
- Extensible
- ML-ready

## Scale Targets

**Single Node Capacity:**
- 5 billion edges @ 256GB RAM
- Realistic pharma: 10-100M edges
- 300k traversals on 10M edges graph: 200ms (already tested on Mac 12 cores)
- Room for 50-500x growth

**Horizontal Scaling:**
- Read replicas (full graph copy)
- Independent queries
- Linear throughput scaling
- No distributed graph complexity

# OGBT Technical Stack

## Backend

```
Language:              Go 1.25+
EventStore/Messegaing: NATS JetStream
Graph Engine:          Custom CSR (Compressed Sparse Row) + DFS/BFS Algo
Metadata Storage:      Apache Pinot
Master Data:           Apache Pinot
Observability:         Op enTelemetry
Blockchain:            Hyperledger Fabric SDK
```

## Frontend

```
Framework:          React 18
Build Tool:         Vite
Graph Viz:          React Flow
Data Fetching:      TanStack Query (React Query)
State Management:   Zustand
Styling:            Tailwind CSS
Components:         shadcn/ui (optional)
```

## Infrastructure

```
Container:          Docker
Orchestration:      Kubernetes (optional, for scale)
CI/CD:              GitHub Actions
Package Manager:    Go modules (backend), npm/pnpm (frontend)
```

## Development

```
Version Control:    Git
Monorepo:           Single repo, separate dirs (cmd/, pkg/, web/)
Documentation:      Markdown
Testing:            Go test, Vitest (frontend)
```