FROM golang:1.23-alpine AS builder

WORKDIR /app 
COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-w -s" -o pending-watcher

FROM scratch

WORKDIR /app
COPY --from=builder /app/pending-watcher /usr/bin/

LABEL org.opencontainers.image.source=https://github.com/shadi/pending-watcher

ENTRYPOINT ["pending-watcher"]
