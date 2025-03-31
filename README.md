# Ajor Debtor Reporter Bot

Simple Telegram Bot for Get Status of your X-UI users in Golang.

## Environment Variables

- `TELEGRAM_BOT_DEBUG_MODE`: Enable or Disable Debug mode. Possible values: `true` or `false`. Default is `false`.
- `TELEGRAM_BOT_ADMIN_ID`: Telegram ID of admin user. You can add more admins using comma. exmaple: `12233,121121`
- `TELEGRAM_BOT_TOKEN`: Telegram Bot API Token.
- `XPANEL_URL`: X-UI panel address. like `http://localhost:54321`.
- `XPANEL_USERNAME`: X-UI Username like `admin`.
- `XPANEL_PASSWORD`: X-UI Password like `admin`.
- `APP_LOG_MODE`: Log level mode. options `info` or `debug`. Default is `info`
- `APP_VERSION`: I don't think you need it. The default value is empty. You can see it in the `/report` command.

## Build docker image

```bash
docker build -t ajor-debtor-bot:latest .
```

## Run docker container

```bash
docker run -d --name ajor-debtor-bot \
    --restart always \
    -e TELEGRAM_BOT_DEBUG_MODE=false \
    -e APP_LOG_MODE=info \
    -e TELEGRAM_BOT_ADMIN_ID=YOUR_TELEGRAM_USER_IDs \
    -e TELEGRAM_BOT_TOKEN="YOUR_BOT_TOKEN" \
    -e XPANEL_URL="http://localhost:54321" \
    -e XPANEL_USERNAME=admin \
    -e XPANEL_PASSWORD=admin \
    ghcr.io/noorbala7418/ajor-debtor-reporter-bot:latest
```
