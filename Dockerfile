FROM golang:1.23-alpine AS builder

ENV GOTOOLCHAIN=go1.25.5

WORKDIR /app
COPY . .
RUN go mod download
RUN GOOS=linux GOARCH=amd64 go build -o node cmd/node/main.go

FROM alpine:latest
RUN apk add --no-cache ca-certificates
WORKDIR /root/
COPY --from=builder /app/node .
EXPOSE 4001 8080
CMD ["./node"]