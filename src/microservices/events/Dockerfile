FROM golang:1.20-alpine


RUN apk --no-cache add git ca-certificates


WORKDIR /app
COPY . .

RUN go mod tidy
RUN go build -o events

EXPOSE 8082
CMD ["./events"]
