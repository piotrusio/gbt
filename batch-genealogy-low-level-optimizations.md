
# High-Performance Batch Genealogy: Low-Level Optimizations (Go + CSR + NATS JetStream)

This is a practical checklist of **low-level optimizations** for building an ultra-fast batch-genealogy traversal service.
Assumes:
- **Go** backend
- **CSR (Compressed Sparse Row)** graph representation (forward + reverse)
- **Event-sourced ingestion** via **NATS JetStream**
- **Frontends** (e.g., Vue Flow) only fetch **incremental subgraphs** and enrich after traversal

---

## 1) Data Layout & Types

**Core principles**
- Prefer **SoA (struct-of-arrays)**, contiguous memory; avoid pointer-heavy structures.
- Use **`uint32` IDs** for nodes/edges when possible (≤4,294,967,295).
- Keep **graph payload-free**; attach metadata only **after** traversal (via master-data services).

**Arrays (all contiguous):**
- `offsetF []uint32` — CSR offsets for **forward** neighbors; length = V+1
- `nbrF    []uint32` — concatenated neighbor node IDs (forward); length = E
- `offsetR []uint32` — CSR offsets for **reverse** neighbors; length = V+1
- `nbrR    []uint32` — concatenated neighbor node IDs (reverse); length = E

**Edge metadata (single copy, by edge index):**
- `edgeTs   []uint32` — event timestamp (epoch-seconds/ms; choose width)
- `edgeType []uint8`  — small enum for consumption/production/move/etc.

**Optional convenience arrays:**
- `degreeF []uint32`, `degreeR []uint32` — fast degree access
- Node relabeling map(s) if you renumber nodes for locality

**Locality / relabeling**
- Relabel nodes to improve cache locality: **BFS order**, **Reverse Cuthill–McKee (RCM)**, or **time-bucket then degree**.
- Within each node’s adjacency, **sort by `edgeTs`** to enable fast time-window slicing via binary search.

---

## 2) Traversal Core (Alloc-Free)

**Visited bookkeeping (epoch trick)**
- Use a generation counter to avoid clearing bitsets between queries.

```go
type Graph struct {
    offsetF, nbrF []uint32
    offsetR, nbrR []uint32
    edgeTs        []uint32
    // visited-epoch trick
    visitEpoch []uint32 // len = V
    curEpoch   uint32
}

func (g *Graph) startQuery() uint32 {
    g.curEpoch++
    if g.curEpoch == 0 { // overflow safety
        for i := range g.visitEpoch { g.visitEpoch[i] = 0 }
        g.curEpoch = 1
    }
    return g.curEpoch
}

func (g *Graph) markVisited(v uint32, epoch uint32) { g.visitEpoch[v] = epoch }
func (g *Graph) isVisited(v uint32, epoch uint32) bool { return g.visitEpoch[v] == epoch }
```

**Alloc-free BFS skeleton**
```go
func (g *Graph) BFS(start uint32, maxHops int, forward bool, tsFrom, tsTo uint32, out func(u, v uint32, eidx int)) {
    epoch := g.startQuery()
    // Preallocate two frontiers we reuse by reslicing
    frontier := make([]uint32, 0, 1<<12)
    next     := make([]uint32, 0, 1<<12)

    g.markVisited(start, epoch)
    frontier = append(frontier, start)

    for hop := 0; hop < maxHops && len(frontier) > 0; hop++ {
        next = next[:0]

        for _, u := range frontier {
            off := g.offsetF; nbr := g.nbrF
            if !forward { off = g.offsetR; nbr = g.nbrR }

            // Adjacency range
            lo, hi := off[u], off[u+1]

            // Optional: time-window slice via binary search over edgeTs (see below)
            // lo, hi = g.sliceByTime(u, lo, hi, forward, tsFrom, tsTo)

            for i := lo; i < hi; i++ {
                v := nbr[i]
                if !g.isVisited(v, epoch) {
                    g.markVisited(v, epoch)
                    next = append(next, v)
                }
                if out != nil {
                    out(u, v, int(i))
                }
            }
        }
        // swap
        frontier, next = next, frontier
    }
}
```

**Time-window slicing of adjacency**
```go
// Assume neighbors per node are sorted by edge timestamp
func (g *Graph) sliceByTime(u, lo, hi uint32, forward bool, from, to uint32) (uint32, uint32) {
    // binary search edgeTs for [from, to] window (left/right bounds)
    // Implement using sort.Search with index mapping u’s edge range.
    return lo, hi // if no filtering
}
```

**Other micro-optimizations**
- **No channels in hot loops**; if you need queues, use a **lock-free ring buffer** with atomics.
- **Avoid maps** on the hot path.
- Use `[:0]` re-slicing, not `make` inside loops.
- Consider `runtime.KeepAlive` and avoid capturing large slices in closures on critical paths.

---

## 3) Pruning & Deep Trees (e.g., depth up to 70)

- **Hop limits** and **domain filters** (material, plant, direction) applied before neighbor expansion.
- **Time-window pruning**: stick to **recent slices first** (e.g., “last 90 days”) and expand older neighbors on demand.
- **Hub control** (for extremely high-degree nodes): cap expansion per hop, expose **“+N more edges”** for UI to request later.
- **Bidirectional BFS** for reachability/path-length queries across very deep trees.

---

## 4) Concurrency & Throughput

- Throughput scales until **memory bandwidth** saturates; run **many queries in parallel** (one goroutine per query) rather than over-optimizing a single traversal.
- On big NUMA boxes (optional): **shard by node ID range**, keep per-shard worker pools, **pin** goroutines to sockets, and **avoid cross-socket** data sharing.
- **False-sharing** prevention: pad shared counters/structs; avoid frequently-written globals.

---

## 5) Memory & GC Discipline

- **Preallocate** frontiers and scratch buffers; zero **allocs** in hot traversal path.
- Keep hot structures **flat slices** (SoA). No per-edge structs, no pointers.
- `sync.Pool` only for **cold-path** buffers; avoid creating GC pressure in tight loops.
- **Visited** via epoch array or a `[]uint64` bitset; both are cache-friendly.
- **Typical footprint** (uint32 IDs, forward+reverse CSR, edgeTs + type):
  - `nbrF + nbrR`: ~ `E * 4 * 2` bytes
  - `offsetF + offsetR`: ~ `(V+1) * 4 * 2` bytes
  - `edgeTs`: ~ `E * 4` bytes; `edgeType`: ~ `E * 1` byte
  - Example @ **4M nodes / 110M edges**: ≈ **1.4–1.7 GB** for the core structure (+ indices).
- **Visited bitset per query**: `V/8` bytes ≈ 0.5 MB for 4M nodes (if you use a bitset instead of epochs).

---

## 6) Persistence & Updates (Event Sourcing)

- Ingest immutable events to **NATS JetStream**.
- Maintain **immutable base CSR** + a small **delta layer** (COO / per-node append lists).
- Periodically **compact**: rebuild CSR snapshot from base + delta, then **pointer-swap** to activate.
- Use **monotonic IDs** for nodes/edges to keep snapshots simple and minimize remapping.
- Snapshot rebuild is deterministic (perfect for audits/validation).

---

## 7) API Shape for Traversals (Streaming)

**Endpoints (examples):**
- `GET /where-used/{node}?hops=3&from=2024-01-01&to=2025-09-01&material=...`
- `GET /where-from/{node}?hops=6&...`
- Both should **stream** results in **batches** (NDJSON or chunked JSON) with:
  - `nodes: []`, `edges: []` per batch
  - Optional **layout hints** (x/y) for client rendering stability
  - **Resume tokens** if you paginate large frontiers

**Result enrichment**
- Return only **IDs** on the traversal stream; the UI calls **master-data services** to hydrate labels, attributes, docs after the graph comes back.

---

## 8) UI Streaming Patterns (Vue Flow-friendly)

- **Incremental reveal**: expand **N hops** at a time and render progressively.
- **Path compaction**: compress long linear chains into a segment with an **expand-on-click** affordance.
- **Hub bundling**: show `+1,248 edges (older than YYYY-MM)` groups; expand lazily.
- **Server-side layout hints**: run **ELK.js / Dagre** (or precomputed coords) per returned subgraph to reduce client churn.

---

## 9) Benchmarking Checklist

Track these under realistic parallel load:
- **p50/p95 latency** for `where-used`/`where-from` at several hop limits.
- **QPS** at saturation on your target VM (parallel queries = vCPU count).
- **Snapshot rebuild time** from JetStream → CSR.
- **Memory footprint** of CSR + deltas + indices.
- **UI throughput**: nodes/edges/sec streamed & rendered without jank.

---

## 10) Minimal Hardware Targets

- Typical enterprise graphs (≤20–50M edges): **16–32 vCPU**, **64–128 GB RAM** VM is ample.
- Your larger case (110M edges / 4M nodes): core graph fits in **~2–3 GB** (+headroom).
- Scale up cores for **throughput**; you’ll hit memory bandwidth ceilings before CPU.

---

## Appendix: Tiny Patterns

**Lock-free ring buffer sketch (single-producer/single-consumer):**
```go
type Ring[T any] struct {
    buf  []T
    mask uint64
    head uint64 // producer
    tail uint64 // consumer
}

func NewRing[T any](pow2 int) *Ring[T] {
    return &Ring[T]{buf: make([]T, 1<<pow2), mask: uint64((1<<pow2)-1)}
}
func (r *Ring[T]) Enq(x T) bool {
    n := r.head - r.tail
    if n == uint64(len(r.buf)) { return false }
    r.buf[r.head & r.mask] = x
    r.head++
    return true
}
func (r *Ring[T]) Deq(out *T) bool {
    if r.tail == r.head { return false }
    *out = r.buf[r.tail & r.mask]
    r.tail++
    return true
}
```

**Binary search skeleton for time-window slice:**
```go
// Given per-node contiguous edge range [lo, hi), find subrange for [from, to].
func boundEdgeRange(ts []uint32, lo, hi uint32, from, to uint32) (uint32, uint32) {
    // implement lower_bound/upper_bound (sort.Search)
    return lo, hi
}
```

**Frontier reuse pattern (no allocs):**
```go
frontier := make([]uint32, 0, 8192)
next     := make([]uint32, 0, 8192)
// ... in loop
next = next[:0]
// fill next
frontier, next = next, frontier // swap
```

---

**Bottom line:** Keep the **graph in contiguous slices**, do **alloc-free traversals**, **prune early** (time, hop, domain), and **stream** subgraphs. Persist events in JetStream, rebuild CSR snapshots deterministically, and enrich **after** traversal. This is how you stay fast, simple, and GxP-friendly.
