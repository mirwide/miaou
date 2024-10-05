# TGBOT

![Build status](https://github.com/mirwide/tgbot/actions/workflows/go.yml/badge.svg)

Bot proxy messages from telegram to llm run on ollama.

## Run

```
echo "TGBOT_TG_TOKEN: <telegram bot token> > .env"
docker compose build
docker compose up -d
```

## Features

- [x] Request rate limiting by chat