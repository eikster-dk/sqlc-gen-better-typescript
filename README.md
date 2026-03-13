# sqlc-effect

sqlc.dev plugin that generates Effect.ts code

## Monorepo Structure

This is a Turborepo monorepo containing:

- **packages/plugin** - The main sqlc WASM plugin for generating Effect.ts code
- **packages/examples** - Example usage and demonstrations

## Getting Started

Install dependencies:

```bash
npm install
```

Build all packages:

```bash
npm run build
```

Run in development mode:

```bash
npm run dev
```

## Generate protobuf code
```shell
buf generate --template buf.gen.yaml buf.build/sqlc/sqlc --path plugin/
```


## Development

This monorepo uses:

- **Turborepo** - For fast, efficient builds
- **npm workspaces** - For package management
- **TypeScript** - For type safety
- **Effect.ts** - For functional programming patterns
