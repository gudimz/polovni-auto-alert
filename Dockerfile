# syntax=docker/dockerfile:1

FROM golang:1.22-alpine AS builder

RUN apk add --no-cache make

WORKDIR /app

ARG BINARY

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN make build-${BINARY}

FROM alpine:latest

WORKDIR /app

COPY --from=builder /app/bin/${BINARY} .

CMD ["/app/${BINARY}"]
