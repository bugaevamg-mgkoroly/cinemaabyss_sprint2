FROM golang:1.25.5-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN go build -o events

FROM alpine:3.19

WORKDIR /app
COPY --from=builder /app/events ./events

EXPOSE 8082
CMD ["./events"]