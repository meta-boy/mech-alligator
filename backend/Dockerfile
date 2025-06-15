FROM golang:1.24-alpine AS builder

WORKDIR /app
# RUN go mod download

COPY . .
RUN go build -o main .

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/

COPY --from=builder /app/main .
EXPOSE 8080
CMD ["./main"]