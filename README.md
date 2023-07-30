![GitHub Workflow Status (with event)](https://img.shields.io/github/actions/workflow/status/ReidMason/discord-bot/docker-publish.yml?logo=Docker)

# Discord bot

Simple Discord bot for messing around with build in Golang

## Deployment

Deploy using Docker

#### Docker compose

```yaml
version: "3.8"
services:
  discord-bot:
    container_name: discord-bot
    image: skippythesnake/discord-bot:latest
    environment:
      - TOKEN=DISCORD_BOT_TOKEN
    restart: unless-stopped
```
