FROM golang:1.23-alpine AS build
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /order-service ./cmd/service

FROM alpine:3.20
WORKDIR /
COPY --from=build /order-service /order-service
EXPOSE 8082
ENTRYPOINT ["/order-service"]
