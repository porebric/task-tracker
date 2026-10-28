FROM golang:1.26-alpine AS build

WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/task-tracker ./cmd/task-tracker

FROM alpine:3.21

RUN apk add --no-cache ca-certificates tzdata && adduser -D -u 10001 app
WORKDIR /app
COPY --from=build /out/task-tracker /app/task-tracker
COPY config /app/config
COPY migrations /app/migrations
USER app
EXPOSE 8080 8081 8082
ENTRYPOINT ["/app/task-tracker"]
