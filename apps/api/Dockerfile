FROM golang:1.25-alpine AS builder
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /app/main .

FROM alpine:3.19
RUN apk add --no-cache ca-certificates
COPY --from=builder /app/main /app/main
WORKDIR /app
EXPOSE 5024
CMD ["/app/main"]
