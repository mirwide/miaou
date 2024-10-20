# Miaou

![Build status](https://github.com/mirwide/tgbot/actions/workflows/go.yml/badge.svg)

AI Assistant. Communication from telegram chat to on-prem LLM run on [ollama](https://github.com/ollama/ollama).

## Run

```bash
echo "MIAOU_TG_TOKEN: <telegram bot token>" > .env.local
docker compose build
docker compose up -d
```

## Features

- [x] Text communication from telegram
- [x] Request rate limiting by chat
- [x] Store chat context with TTL and clear command
- [x] Image to text for multimodal model
- [x] Support external tool
- [x] Get weather from open-meteo(non-commercial use only)
- [x] Search content on wiki
- [ ] Generate image
- [ ] Voice communication
- [ ] Integration with Xiaomi smart home(?)
