FROM golang:1.21-alpine as builder

WORKDIR /app

COPY go.* .

RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build ./cmd/main.go

FROM alpine:3.19

ENV TELEGRAM_BOT_DEBUG_MODE=false

COPY --from=builder /app/main /app/bot

CMD ["/app/bot"]