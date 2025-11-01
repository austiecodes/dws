# Distributed RPC Integration Plan

## Goals

- Replace implicit DB-based coordination between the scheduler and workers with explicit RPC contracts that work when multiple workers run on different machines.
- Expose container lifecycle operations on each worker via RPC/HTTP so that the platform service can provision and tear down remote containers without direct Docker access.
- Reuse the existing `workers` registry in Postgres to discover endpoints, track capacity, and record heartbeat information.

## Current State Snapshot

- **Scheduler → Worker**: Scheduler marks tasks as `running` in Postgres; workers poll `tasks` table and execute anything in the `running` state.
- **Platform → Worker**: Platform service calls the Docker SDK directly; all container lifecycle happens on the same host.
- **Workers**: Only run the polling executor and timeout watcher; no network API surface.

## Target Architecture

```
┌──────────────┐          ┌────────────────┐          ┌────────────────────┐
│ Platform API │──RPC────▶│ Worker Control │          │                    │
│              │          │  (per worker)  │──Docker─▶│ Worker Host Docker │
└──────────────┘          └────────────────┘          └────────────────────┘
        │                            ▲
        │                            │ heartbeats, capacity reports
        │                            │ task results, metrics
        │                            │
        │                            ▼
        │                     ┌───────────────┐
        └───HTTP/DB──────────▶│ Scheduler RPC │
                              │   (singleton) │
                              └───────────────┘
                                ▲          │
                                │ RPC pull │ new tasks
                                │          ▼
                          ┌───────────────┐
                          │ Worker Client │
                          └───────────────┘
```

### Core Services

1. **Scheduler RPC server (`SchedulerService`)**
   - `PollTasks(stream WorkerPollRequest) returns stream TaskAssignment`
   - `ReportTaskResult(TaskResult)` ack
   - `SendHeartbeat(stream Heartbeat)` for liveness and dynamic capacity updates
2. **Worker RPC server (`WorkerControlService`)**
   - `CreateContainer/CreateContainerFromSpec`
   - `StartContainer`, `StopContainer`, `DeleteContainer`
   - `ExecuteCommand` (internal use for scheduler-assigned tasks)
   - Optional `StreamLogs` for user-facing log streaming

### Message Flow

1. Worker boots, loads config, and opens two bidirectional gRPC streams to the scheduler (`SendHeartbeat`, `PollTasks`).
2. Scheduler maintains an in-memory map of active workers (ID → capacity, free slots). When a new task is ready, it selects a worker and pushes an assignment down the `PollTasks` stream.
3. Worker acknowledges receipt, executes the task locally, and sends a `TaskResult` back to the scheduler.
4. Platform service takes the current `workers` table, selects a candidate (e.g., least containers), and calls that worker’s `WorkerControlService` RPCs to create/delete containers. Worker saves container metadata (Docker ID, SSH port) locally and reports the new state back via the heartbeat payload or a dedicated `ContainerSync` RPC.

## Required Schema & Config Updates

- Ensure `containers.worker_id` is populated for every container (backfill existing rows or default to a special `local` worker entry).
- Extend `configs/app.toml` with:
  ```toml
  [scheduler.rpc]
  listen_addr = "0.0.0.0:7001"

  [worker.rpc]
  scheduler_addr = "scheduler:7001"
  listen_addr    = "0.0.0.0:7002"
  ```

## Implementation Checklist

1. **API surface**
   - Add `api/v1/scheduler.proto` and `api/v1/worker.proto` with the contracts outlined above.
   - Generate Go stubs under `internal/lib/apis` (use `buf` or `protoc` with `--go_out --go-grpc_out`).
2. **Scheduler**
   - New package `internal/scheduler/grpcserver` for the gRPC server.
   - Refactor dispatcher to request available capacity from the in-memory worker registry instead of raw DB counts.
   - Persist worker heartbeats and capacities back into the existing `workers` table.
3. **Worker**
   - New package `internal/worker/grpcclient` to wrap scheduler RPCs with reconnect/backoff logic.
   - New package `internal/worker/grpcserver` exposing `WorkerControlService`.
   - Update `cmd/worker/main.go` to start both the RPC server and the background executor/timeout watcher.
4. **Platform**
   - Introduce `internal/platform/services/worker_api` that dials the per-worker RPC endpoint (pulled from `repository.Workers`).
   - Refactor `ContainerService` to call the worker RPC instead of `libdocker` directly. The worker returns the Docker container metadata which the platform persists.
5. **Shared**
   - Create helper in `internal/lib/apis/grpcutil` for dial options (timeouts, TLS placeholders).
   - Add integration tests that spin up an in-memory gRPC server (use `bufconn`) to verify scheduler ↔ worker flows.

## Deployment Notes

- Scheduler and worker binaries now require exposed RPC ports (document in `deployments/` when manifests are added).
- For staged rollout, keep the existing DB polling logic behind a feature flag so the old pathway can be re-enabled while RPC components are tested.
- Consider mTLS once basic functionality is stable; shape config now (`tls_cert`, `tls_key`) to avoid another config churn.

## Open Questions

- Do we need push-style task assignments (scheduler initiated) in addition to the worker pull stream? If so, we can add an optional server-initiated stream once the basic pull loop is stable.
- Should container log streaming be proxied through the platform or exposed directly from workers?
- How to handle scheduler failover (single instance today). A standby scheduler would need to recover worker streams; document this as a later milestone.

