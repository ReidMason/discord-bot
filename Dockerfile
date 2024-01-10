FROM golang:alpine AS builder

# We need SSL certs to connect to Discord later
RUN apk update && apk add git && apk add ca-certificates

WORKDIR /app

COPY . .

RUN go build -o ./discord-bot

FROM scratch

# We need SSL certs from the build server to connect to Discord
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

COPY --from=builder /app/discord-bot ./discord-bot

ENTRYPOINT ["./discord-bot"]
