# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

LightAC is a mobile lightweight agent client for observing and continuing coding agent sessions (Codex, Claude Code) from a phone. The product is not a full mobile IDE, terminal, or secondary agent; it's a structured agent work information flow with low-friction observation and handoff.

## Architecture

The system follows a layered architecture with three main runtime components:

1. **Mobile App** (Flutter + Dart): Observes sessions, displays agent events, sends messages, and handles confirmations.
2. **Cloud Control Plane** (Go): User authentication, device pairing, LiveKit room/token issuance, daemon registry, push notification routing.
3. **Remote Daemon / Session Gateway** (Go): Runs on the user's server/development machine, manages sessions, normalizes agent events, handles WebRTC DataChannel connections.

Additional architectural layers:
- **Provider Adapter**: Translates provider-specific output (Codex, Claude Code) to unified Agent Work Protocol.
- **Session Runtime Backend**: Abstracts session persistence and execution (CLI/tmux, provider-native server, cloud agent).
- **Agent Work Protocol**: Structured JSON messages over WebRTC DataChannel (not raw terminal streams).

## Directory Structure (Planned)

The repository is organized as a monorepo:

```
apps/
  mobile/                 # Flutter + Dart app
  control-plane/          # Go control plane
daemon/                   # Go daemon (remote session gateway)
  cmd/lightac-daemon/
  internal/...
packages/
  protocol/               # JSON Schema / protobuf definitions and generated Dart/Go types
docs/                     # Product and architecture documentation
infra/                    # LiveKit/coturn, Docker, deployment configs
```

## Technology Stack

- **Mobile App**: Flutter + Dart, LiveKit Flutter SDK, flutter_secure_storage, push notification integration
- **Control Plane**: Go, PostgreSQL, Redis, LiveKit room/token issuance, push notification services
- **Daemon**: Go, SQLite event store, pty/tmux integration, local WebSocket transport for M0, LiveKit/Pion transport spikes
- **Protocol**: JSON Schema or protobuf schema with generated Dart models and Go structs
- **Transport**: M0 local WebSocket; M1/M2 LiveKit Cloud; production may use LiveKit Cloud, self-hosted LiveKit, or custom signaling + coturn

## Common Development Tasks

### Mobile App
```bash
cd apps/mobile
flutter pub get            # Install dependencies
flutter run                # Run app
flutter test               # Run tests
```

### Control Plane
```bash
cd apps/control-plane
go mod download
go run ./cmd/lightac-control-plane
go test ./...
```

### Daemon
```bash
cd daemon
go mod download            # Download dependencies
go build ./cmd/lightac-daemon  # Build binary
go test ./...              # Run tests
go run ./cmd/lightac-daemon --addr=127.0.0.1:8765 --db=lightac-m0.db  # Run M0 locally
```

### Protocol Generation
```bash
cd packages/protocol
make generate              # Generate Dart and Go types once tooling exists
```

### M0 Client
```bash
cd tools/m0-client
go run .                   # Exercise hello/heartbeat/session.list/event.subscribe
```

## Important Notes

- This is a new project; many of the above directories may not yet exist.
- The current implementation focus is M0 technical spike: daemon skeleton, local WebSocket transport, SQLite Event Store, TmuxCliBackend spike, and Codex CLI coverage report.
- The architecture prioritizes structured agent events over terminal text streams.
- Real-time transport is pluggable. M0 uses local WebSocket; M1/M2 may use LiveKit Cloud data APIs; self-hosted/custom WebRTC is a later production decision.
- Provider support begins with Codex CLI via tmux/pty backend; Claude Code and other providers are future work.

## Key Files to Understand

- `docs/prd.md`: Product requirements and user scenarios
- `docs/b-plus-architecture-design.md`: Detailed architecture and protocol design
- `docs/m0-implementation-plan.md`: M0 technical spike plan
- `docs/m0-codex-cli-coverage-template.md`: Template for Codex CLI parseability report
- `daemon/internal/protocol/`: M0 protocol envelope implementation
- `daemon/internal/store/`: M0 SQLite event store
- `daemon/internal/transport/ws/`: M0 local WebSocket transport
- `tools/m0-client/`: M0 local protocol client
- `packages/protocol/schemas/`: Protocol schema definitions (once created)
- `daemon/internal/session/`: Session management logic (once created)
- `apps/mobile/src/screens/SessionDetail.tsx`: Main session observation UI (once created)

## Development Workflow

1. **Protocol first**: Keep Agent Work Protocol changes explicit and versioned.
2. **Daemon-driven**: Much of the business logic lives in the Go daemon; mobile app is primarily a UI layer.
3. **Event normalization**: Agent output parsing happens in the daemon's provider adapter, not in the mobile app.
4. **Cursor-based recovery**: All agent events have monotonic cursors for reliable reconnection and resume.

## Testing Strategy

- Unit tests for daemon internal packages and control plane services.
- Integration tests for local WebSocket/LiveKit connectivity and session lifecycle.
- End-to-end tests using a mock provider adapter.
- Mobile UI tests with Flutter test/integration_test once the app exists.

## Deployment

- Daemon: Single static binary distributed via GitHub Releases.
- Control Plane: Go service deployed to a cloud platform.
- Mobile App: App Store and Play Store distribution.
- Transport: LiveKit Cloud initially; self-hosted LiveKit or coturn only if production constraints require it.

## Contributing

Refer to the architecture documents before making significant changes to ensure consistency with the layered design. All protocol changes must be backward compatible or versioned.


# Compact instructions

Context may be auto-compacted near its limit.
Do not stop early because of token budget.

Before context refresh or compaction:
- save the current plan
- save changed files
- save failing test names
- save the exact next step

When exploring code:
- prefer rg / symbol search / line ranges
- do not read huge files in full unless necessary
- summarize long logs instead of pasting them verbatim
- keep tool output concise

When using compact, preserve:
- current task goal
- files already changed
- exact test command and latest result
- blocker and next edit target