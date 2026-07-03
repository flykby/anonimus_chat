# Database migrations (goose)

SQL migrations for PostgreSQL with **pgvector**.

## Commands

```bash
# Start postgres (pgvector image)
make compose-up-infra

# Apply migrations
make migrate-up

# Seed test persona + photos
make seed

# Status / rollback
make migrate-status
make migrate-down
```

`DATABASE_URL` example:

```
postgresql://anonimus:anonimus@localhost:5432/anonimus?sslmode=disable
```

Inside Docker Compose use host `postgres` instead of `localhost`.

## Files

| Migration | Description |
|-----------|-------------|
| `00001_extensions.sql` | `vector` extension (pgvector) |
| `00002_schema.sql` | Core tables, enums, indexes |

## Go layer

- Models: `internal/shared/models.go`
- DB pool: `internal/db/pool.go`
