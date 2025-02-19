# Use Golang base image
FROM golang:1.23 as builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go build -o chatserver goServer/server.go

FROM debian:bookworm-slim 

WORKDIR /root/

COPY --from=builder /app/chatserver .

EXPOSE 9090

CMD ["./chatserver"]

# docker-compose up --build            
