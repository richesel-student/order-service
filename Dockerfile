FROM golang:1.23-alpine AS build
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /app/order-service ./cmd/service

FROM alpine:3.20
WORKDIR /app

COPY --from=build /app/order-service /app/order-service
COPY --from=build /app/web /app/web
EXPOSE 8082
ENTRYPOINT ["/app/order-service"]
