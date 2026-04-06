FROM golang:1.22-bookworm AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
# Building with CGO enabled for go-sqlite3
RUN CGO_ENABLED=1 GOOS=linux go build -o sysagent ./cmd/sysagent

FROM ubuntu:22.04
WORKDIR /app
# Install CA certs and any needed libs
RUN apt-get update && apt-get install -y ca-certificates sqlite3 && rm -rf /var/lib/apt/lists/*
COPY --from=builder /app/sysagent .

EXPOSE 8080
CMD ["./sysagent"]
