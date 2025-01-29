FROM golang:1.20-alpine AS builder
WORKDIR /app
COPY . .
RUN go build -o go-simple-http-proxy go-simple-http-proxy.go

FROM alpine:latest
WORKDIR /app
COPY --from=builder /app/go-simple-http-proxy .
EXPOSE 80
CMD ["./go-simple-proxy-server"]
