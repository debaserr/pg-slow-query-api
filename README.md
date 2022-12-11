# PostgreSQL slow query API

## Get started

Export required env variables using direnv

```direnv allow```

or manually source `.envrc`

```source .envrc```

Start a local PostgreSQL db

```docker compose up -d```

Run API server locally

```go build```

```./pg-slow-query-api```

(This should be dockerized and added to docker-compose instead)

Send a GET request to `localhost:3000/demo/init` to seed data.

API is now accessible at `localhost:3000/slow-queries?page=1&page_size=2&query_type=select&order_by=asc`

Use `docker compose down -v` to take down database.
