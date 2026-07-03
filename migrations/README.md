# Database migrations (goose)

SQL migrations for PostgreSQL. Applied via [goose](https://github.com/pressly/goose).

## Commands

```bash
# Apply all pending migrations
make migrate-up

# Roll back last migration
make migrate-down

# Show migration status
make migrate-status

# Create new migration
make migrate-create NAME=add_users
```

Requires `DATABASE_URL` in environment, e.g.:

```
postgresql://anonimus:anonimus@localhost:5432/anonimus?sslmode=disable
```

Inside Docker Compose network use host `postgres` instead of `localhost`.

Real schema tables — task **006**.
