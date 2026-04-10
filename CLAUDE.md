# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

LightAC is a mobile lightweight agent client for observing and continuing coding agent sessions (Codex, Claude Code) from a phone. The product is not a full mobile IDE, terminal, or secondary agent; it's a structured agent work information flow with low-friction observation and handoff.

## Architecture

The system follows a layered architecture with three main runtime components:

1. **Mobile App** (React Native + TypeScript): Observes sessions, displays agent events, sends messages, and handles confirmations.
2. **Signaling Service** (TypeScript + Node.js): Cloud control plane for user authentication, device pairing, WebRTC signaling, push notification routing.
3. **Remote Daemon / Session Gateway** (Go): Runs on the user's server/development machine, manages sessions, normalizes agent events, handles WebRTC DataChannel connections.

Additional architectural layers:
- **Provider Adapter**: Translates provider-specific output (Codex, Claude Code) to unified Agent Work Protocol.
- **Session Runtime Backend**: Abstracts session persistence and execution (CLI/tmux, provider-native server, cloud agent).
- **Agent Work Protocol**: Structured JSON messages over WebRTC DataChannel (not raw terminal streams).

## Directory Structure (Planned)

The repository is organized as a monorepo:

```
apps/
  mobile/                 # React Native + TypeScript app
  control-plane/         # Node.js signaling service
daemon/                  # Go daemon (remote session gateway)
  cmd/lightac-daemon/
  internal/...
packages/
  protocol/              # JSON Schema / TypeSpec definitions
  shared/                # Shared TypeScript utilities
docs/                    # Product and architecture documentation
infra/                   # TURN server, Docker, deployment configs
```

## Technology Stack

- **Mobile App**: React Native (CLI or Expo custom dev client), TypeScript, react-native-webrtc
- **Control Plane**: Node.js with Fastify/NestJS, PostgreSQL, Redis, push notification services
- **Daemon**: Go with Pion WebRTC, SQLite event store, pty/tmux integration
- **Protocol**: JSON Schema with generated TypeScript types and Go structs
- **TURN**: coturn (external service)

## Common Development Tasks

### Mobile App
```bash
cd apps/mobile
npm install                 # Install dependencies
npm run ios                # Run iOS simulator
npm run android            # Run Android emulator
npm run test               # Run tests
npm run build:ios          # Build for iOS
```

### Control Plane
```bash
cd apps/control-plane
npm install
npm run dev                # Development server
npm run test
npm run build
npm run migrate            # Run database migrations
```

### Daemon
```bash
cd daemon
go mod download            # Download dependencies
go build ./cmd/lightac-daemon  # Build binary
go test ./...              # Run tests
go run ./cmd/lightac-daemon --config=config.yaml  # Run locally
```

### Protocol Generation
```bash
cd packages/protocol
npm install
npm run generate           # Generate TypeScript and Go types
```

### Monorepo Management
If using npm workspaces or similar:
```bash
npm install                # Install all workspace dependencies
npm run build --workspaces # Build all packages
```

## Important Notes

- This is a new project; many of the above directories may not yet exist.
- The current implementation focus is on the daemon and protocol first (M1 milestone).
- The architecture prioritizes structured agent events over terminal text streams.
- WebRTC DataChannel is used for real-time communication, not for remote terminal emulation.
- Provider support begins with Codex CLI via tmux/pty backend; Claude Code and other providers are future work.

## Key Files to Understand

- `docs/prd.md`: Product requirements and user scenarios
- `docs/b-plus-architecture-design.md`: Detailed architecture and protocol design
- `packages/protocol/schemas/`: Protocol schema definitions (once created)
- `daemon/internal/session/`: Session management logic (once created)
- `apps/mobile/src/screens/SessionDetail.tsx`: Main session observation UI (once created)

## Development Workflow

1. **Protocol first**: Update schema definitions and regenerate types before implementing features.
2. **Daemon-driven**: Much of the business logic lives in the Go daemon; mobile app is primarily a UI layer.
3. **Event normalization**: Agent output parsing happens in the daemon's provider adapter, not in the mobile app.
4. **Cursor-based recovery**: All agent events have monotonic cursors for reliable reconnection and resume.

## Testing Strategy

- Unit tests for daemon internal packages and control plane services.
- Integration tests for WebRTC connectivity and session lifecycle.
- End-to-end tests using a mock provider adapter.
- Mobile UI tests with Detox or similar framework.

## Deployment

- Daemon: Single static binary distributed via GitHub Releases.
- Control Plane: Containerized deployment to cloud platform (AWS, GCP, etc.).
- Mobile App: App Store and Play Store distribution.
- TURN: Managed coturn instance or cloud service (Twilio, etc.).

## Contributing

Refer to the architecture documents before making significant changes to ensure consistency with the layered design. All protocol changes must be backward compatible or versioned.