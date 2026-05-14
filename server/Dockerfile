FROM golang:1.26-alpine AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN go build -o api ./cmd/api
RUN go build -o worker ./cmd/worker

FROM alpine:3.23.3

WORKDIR /app
COPY --from=builder /app/api .
COPY --from=builder /app/worker .

EXPOSE 8080
CMD ["./api"]