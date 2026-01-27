FROM golang



WORKDIR /app
COPY . .

RUN go mod tidy
RUN go build -o events

EXPOSE 8082
CMD ["./events"]
