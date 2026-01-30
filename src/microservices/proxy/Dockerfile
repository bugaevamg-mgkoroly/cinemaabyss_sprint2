
FROM golang:1.23-alpine AS builder
RUN apk --no-cache add ca-certificates

WORKDIR /root/
COPY --from=builder /app/proxy .

EXPOSE 8000
CMD ["./proxy"]