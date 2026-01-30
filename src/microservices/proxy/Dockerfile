
FROM alpine:latest
RUN apk --no-cache add ca-certificates

WORKDIR /root/
COPY --from=builder /app/proxy .

EXPOSE 8000
CMD ["./proxy"]