# Suomi News telegram channel

## Local development

Run redis in docker

```bash
docker run -p 6379:6379 redis:alpine
```

Create `.env` file, put these variables in there:

```
TELEGRAM_BOT_TOKEN=<YOUR_BOT_TOKEN>
CHANNEL_ID=<PUBLIC_CHANNEL_ID_FOR_NEWS>
ADMIN_CHANNEL_ID=<PRIVATE_ADMIN_CHANNEL_ID_FOR_APPROVALS>

REDIS_URL=127.0.0.1:6379
```

Run the app

```bash
go run main.go
```

## Running in Railway

0. Create managed Redis in Railway (all default settings)
1. Connect github repo to the project
2. Add environmental variables

```
TELEGRAM_BOT_TOKEN=<YOUR_BOT_TOKEN>
CHANNEL_ID=<PUBLIC_CHANNEL_ID_FOR_NEWS>
ADMIN_CHANNEL_ID=<PRIVATE_ADMIN_CHANNEL_ID_FOR_APPROVALS>

# Get this url from Railway managed Redis
REDIS_URL=redis.railway.internal:6379
REDIS_USERNAME=default
REDIS_PASSWORD=password
APP_ENV=PROD
```

## Tests

To run tests use:

```bash
go test ./...
```

The `./...` makes tests run in all subdirectories.

To see more verbose output of tests use:

```bash
go test ./... -v
```
