FROM golang:1.26-alpine AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN go build -o main ./cmd/api

FROM alpine:3.23.3

WORKDIR /app
COPY --from=builder /app/main .

EXPOSE 8080
ENTRYPOINT ["./main"]