FROM golang:1.22-bookworm AS builder

WORKDIR /app

COPY go.mod ./
RUN go mod download

COPY cmd ./cmd
COPY internal ./internal

RUN go build -o /app/codeact-server ./cmd/server

FROM golang:1.22-bookworm

WORKDIR /app

COPY --from=builder /app/codeact-server ./codeact-server
COPY workspace ./workspace

ENV CODEACT_ADDR=:8080

EXPOSE 8080

CMD ["./codeact-server"]
