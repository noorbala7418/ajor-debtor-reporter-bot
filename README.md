# Alive Bot

Simple Telegram Bot for Get Status of your X-UI users with Golang.

## Environment Variables

- `TELEGRAM_BOT_DEBUG_MODE`: Enable or Disable Debug mode. Possible values: `true` or `false`.
- `TELEGRAM_BOT_ADMIN_ID`: Telegram ID of admin user.
- `TELEGRAM_BOT_TOKEN`: API Token.
- `XPANEL_URL`: X-UI panel address. like `http://localhost:54321`.
- `XPANEL_USERNAME`: X-UI Username like `admin`.
- `XPANEL_PASSWORD`: X-UI Password like `admin`.
- `APP_LOG_MODE`: Log level mode. options `info` or `debug`. Default is `info`

## Build docker image

```bash
docker build -t ajor-debtor-bot:latest .
```

## Run docker container

```bash
docker run -d --name ajor-debtor-bot --restart always -e TELEGRAM_BOT_DEBUG_MODE=false -e TELEGRAM_BOT_ADMIN_ID=YOUR_TELEGRAM_USER_ID -e TELEGRAM_BOT_TOKEN="YOUR_BOT_TOKEN" -e XPANEL_URL="http://localhost:54321" -e XPANEL_USERNAME=admin -e XPANEL_PASSWORD=admin ajor-debtor-bot:latest
```
