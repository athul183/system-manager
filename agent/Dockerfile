FROM rust:1.77-slim AS builder
WORKDIR /app

# Step 1: Create a dummy project to pre-compile and cache dependencies
COPY Cargo.toml Cargo.lock* ./
RUN mkdir src && echo "fn main() {}" > src/main.rs
RUN cargo build --release
RUN rm -rf src

# Step 2: Build actual source code
COPY src src
# "touch" main.rs to ensure Cargo knows the dummy compiled object is stale
RUN touch src/main.rs
RUN cargo build --release

# Step 3: Minimal runner image
FROM ubuntu:22.04
WORKDIR /app
RUN apt-get update && apt-get install -y ca-certificates sqlite3 libsqlite3-0 && rm -rf /var/lib/apt/lists/*
COPY --from=builder /app/target/release/sysagent .

EXPOSE 8080
CMD ["./sysagent"]
