# Fiberlane — Photonics Sourcing MVP

## Quick start

A fresh clone has no `.env` files (they are gitignored). Bootstrap them from the
committed `*.env.example` templates, then fill in secrets:

```sh
make setup        # copies .env.example -> .env at root and per service (idempotent, never clobbers)
$EDITOR .env      # fill in secrets: JWT_SECRET, *_PASSWORD, *_API_KEY, ...
make up-d         # start the full stack
```

`make setup` uses `cp -n`, so it never overwrites existing local `.env` files.
Root `.env` supplies the shared secrets Compose injects; the per-service
`services/<svc>/.env` files back the `env_file:` lines and support local
`go run .`. See `CLAUDE.md` for the full architecture and commands.
