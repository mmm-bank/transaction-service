FROM golang:latest AS builder

WORKDIR /build

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go build -o transaction .

FROM ubuntu:latest

WORKDIR /app

COPY --from=builder /build/transaction .

ENTRYPOINT ["/app/transaction"]
