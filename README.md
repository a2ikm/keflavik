## Migrations

## Development tips

### How to run postgres

```
$ docker compose up -d
```

You can run psql command on the postgres container:

```
$ docker compose exec postgres psql --username postgres --password --dbname keflavik
```
