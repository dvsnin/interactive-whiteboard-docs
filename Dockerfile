FROM golang:1.23 AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY software_engineering/ ./software_engineering/

WORKDIR /app/software_engineering
RUN CGO_ENABLED=0 GOOS=linux go build -o /server main.go

FROM scratch

WORKDIR /root
COPY --from=builder /server .

EXPOSE 8080
CMD ["./server"]
