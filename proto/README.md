# go-job-queue Proto Definitions

Protocol buffer definitions for the go-job-queue project, compiled for Go and TypeScript/JavaScript.

## Prerequisites

- [Buf CLI](https://buf.build/docs/installation) - Protocol buffer tooling
- Node.js (for TypeScript/JavaScript builds)

## Project Structure

```
proto/
├── src/proto/           # Source .proto files
├── gen/                 # Generated TypeScript (intermediate)
│   └── mpataki/jobqueue/v1/
│       ├── job_service_pb.ts
│       └── job_service_connect.ts
├── dist/                # Compiled JavaScript (published)
│   └── mpataki/jobqueue/v1/
│       ├── job_service_pb.js
│       ├── job_service_pb.d.ts
│       ├── job_service_connect.js
│       └── job_service_connect.d.ts
└── buf.gen.yaml         # Buf generation config
```

## Build Commands

### Generate Code from Proto Files

```bash
# Generate TypeScript and Go code
buf generate

# Or use npm script (runs buf generate + tsc)
npm run build
```

This generates:
- **Go code**: `gen/go/` - Go protobuf and Connect-RPC code
- **TypeScript**: `gen/` - TypeScript source files
- **JavaScript**: `dist/` - Compiled JS + type definitions (ready for npm publish)

### Clean Generated Files

```bash
npm run clean
```

Removes both `gen/` and `dist/` directories.

## Usage

### TypeScript/JavaScript

```bash
npm install @mpataki/go-job-queue-proto
```

```typescript
import { Job, JobStatus } from '@mpataki/go-job-queue-proto';
import { JobServiceConnect } from '@mpataki/go-job-queue-proto';
```

### Go

```go
import jobv1 "github.com/mpataki/go-job-queue/proto/gen/go/mpataki/jobqueue/v1"
```

## Development

### Adding New Proto Definitions

1. Add/modify `.proto` files in `src/proto/`
2. Run `npm run build` to regenerate code
3. Commit both source and generated files

### Configuration Files

- `buf.yaml` - Buf module configuration
- `buf.gen.yaml` - Code generation plugins and output paths
- `tsconfig.json` - TypeScript compiler settings
- `package.json` - npm package configuration
