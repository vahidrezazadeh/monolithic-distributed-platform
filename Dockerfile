FROM golang:1.24-alpine AS builder

WORKDIR /app

RUN apk add --no-cache git make gcc musl-dev

# تنظیم GOPROXY
ENV GOPROXY=https://goproxy.io,direct

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -ldflags "-X main.buildTime=$(date -u +'%Y-%m-%dT%H:%M:%SZ')" -o app .

FROM alpine:latest

WORKDIR /app

COPY --from=builder /app/app .

EXPOSE 8080 50051 7946

CMD ["./app"]