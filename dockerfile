FROM golang:1.21-alpine AS builder
WORKDIR /app
COPY . .
RUN go mod download
RUN GOOS=linux GOARCH=amd64 go build -o node cmd/chat/main.go

FROM alpine:latest
WORKDIR /root/
COPY --from=builder /app/node .
EXPOSE 4001 
CMD ["./node"]