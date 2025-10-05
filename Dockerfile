FROM golang:1.23 AS builder

WORKDIR /app

COPY software_engineering/go.mod software_engineering/go.sum ./
RUN go mod download

COPY software_engineering/ ./

RUN CGO_ENABLED=0 GOOS=linux go build -o server main.go

FROM scratch

WORKDIR /root
COPY --from=builder /app/server .

EXPOSE 8080
CMD ["./server"]
