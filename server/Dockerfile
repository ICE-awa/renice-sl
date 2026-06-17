FROM golang:1.26-alpine AS builder

WORKDIR /app

ENV GOPROXY=https://goproxy.cn,direct
ENV GOSUMDB=sum.golang.google.cn

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN go build -o api ./cmd/api
RUN go build -o worker ./cmd/worker
RUN go build -o migrate ./cmd/migrate

FROM alpine:3.23.3

WORKDIR /app
RUN apk add --no-cache ca-certificates
COPY --from=builder /app/api .
COPY --from=builder /app/worker .
COPY --from=builder /app/migrate .

EXPOSE 8080
CMD ["./api"]