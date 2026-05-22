# Install

## Requirements

- Go 1.23+
- Node.js 22+
- pnpm
- Emby or Jellyfin server
- Redis for production sessions and rate limits is recommended

## Backend

```bash
go build -o bin/twilight ./cmd/twilight
cp config.production.toml config.toml
bash start_backend_prod.sh
```

The API listens on `127.0.0.1:5000` or the host/port configured through `TWILIGHT_API_HOST` and `TWILIGHT_API_PORT`.

## Frontend

```bash
cd webui
pnpm install --frozen-lockfile
pnpm build
pnpm start -p 3000
```

## systemd

Use the Go service files in `deploy/`. They expect the backend binary at `/root/Twilight/bin/twilight`.

```bash
go build -o bin/twilight ./cmd/twilight
sudo cp deploy/twilight.service /etc/systemd/system/
sudo systemctl daemon-reload
sudo systemctl enable --now twilight
```
