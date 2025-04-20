# Stage 1 - build
FROM golang:latest as builder
WORKDIR /app
COPY . .
RUN go build .

# Stage 2 - run
FROM debian:bookworm-slim
WORKDIR /app
COPY --from=builder /app/users-api /app/
CMD ["/app/users-api"]
