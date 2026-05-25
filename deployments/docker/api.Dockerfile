# STAGE 1: Build
FROM golang:1.25-alpine AS builder

RUN apk add --no-cache git build-base

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -installsuffix cgo -o main ./cmd/api/main.go

# STAGE 2: Final Image
FROM alpine:latest
RUN apk --no-cache add ca-certificates

WORKDIR /root/
COPY --from=builder /app/main .
COPY --from=builder /app/.env .

RUN chmod +x ./main

EXPOSE 8081

ENTRYPOINT ["./main"]
