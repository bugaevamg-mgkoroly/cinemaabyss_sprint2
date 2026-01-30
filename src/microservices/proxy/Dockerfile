FROM golang:1.21 AS builder 
WORKDIR /app
COPY . .
RUN go build -o /app/proxy


# Этап 2: финальный образ
FROM alpine:latest
WORKDIR /root
COPY --from=builder /app/proxy .  
EXPOSE 8000
CMD ["./proxy"]