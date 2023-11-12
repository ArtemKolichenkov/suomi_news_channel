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
HTTP_HOST=127.0.0.1
# Note - name must be PORT for render deploy https://render.com/docs/web-services#deploying-from-a-git-repository
# In prod don't even specify it, let render handle it themselves
PORT=8080
```

Run the app

```bash
go run main.go
```

## Running in render.com

0. Create managed Redis in render.com (all default settings)
1. Create "Web Service"
2. Choose `Go` as runtime (use default build settings)
3. Under `Advanced` menu - click `Add Secret File`, name the file `.env` and put this config in there:

```
TELEGRAM_BOT_TOKEN=<YOUR_BOT_TOKEN>
CHANNEL_ID=<PUBLIC_CHANNEL_ID_FOR_NEWS>
ADMIN_CHANNEL_ID=<PRIVATE_ADMIN_CHANNEL_ID_FOR_APPROVALS>

# Get this url from render.com managed Redis
REDIS_URL=red-blablabla:6379
HTTP_HOST=0.0.0.0
```

Note that `HTTP_HOST` must be `0.0.0.0` and `PORT` should not be there at all (render handles it themselves)

4. Click `Create Web Service`
