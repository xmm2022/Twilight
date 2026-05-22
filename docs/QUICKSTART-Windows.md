# Windows Quickstart

## Requirements

- Go 1.23+
- Node.js 22+
- pnpm

## Backend

```powershell
go test ./...
go run ./cmd/twilight api --host 127.0.0.1 --port 5000 --config config.toml --debug
```

To build a binary:

```powershell
go build -o bin/twilight.exe ./cmd/twilight
.\bin\twilight.exe api --host 127.0.0.1 --port 5000 --config config.toml
```

## Frontend

```powershell
cd webui
pnpm install --frozen-lockfile
pnpm dev
```

Open `http://localhost:3000`.
