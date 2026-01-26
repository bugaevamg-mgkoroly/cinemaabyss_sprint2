FROM golang


RUN apk --no-cache add git ca-certificates


WORKDIR /app
COPY . .

RUN go mod tidy
RUN go build -o proxy


EXPOSE 8000
CMD ["./proxy"]
