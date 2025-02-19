# Use Golang base image
FROM golang:1.23 as builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go build -o chatclient goClient/client.go

FROM debian:bookworm-slim 

WORKDIR /root/

COPY --from=builder /app/chatclient .

CMD ["./chatclient"]