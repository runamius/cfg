FROM golang:1.25-alpine AS builder
WORKDIR /app

RUN apk add --no-cache git
COPY go.mod go.sum ./
ENV GOPROXY=https://proxy.golang.org,direct
ENV GOSUMDB=sum.golang.org
RUN go mod download

COPY . .
RUN go build -o server ./cmd/main.go

FROM debian:bookworm-slim
WORKDIR /app

COPY --from=builder /app/server .
COPY --from=builder /app/schema ./schema

EXPOSE 8080
CMD ["./server"]