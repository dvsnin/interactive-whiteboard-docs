FROM golang:1.23 AS builder
WORKDIR /app
COPY . .
RUN cd software_engineering && go build -o server main.go

FROM debian:stable-slim
WORKDIR /root/
COPY --from=builder /app/software_engineering/server .
EXPOSE 8080
CMD ["./server"]
