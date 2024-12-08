FROM golang:1.23.3-alpine3.20 AS build

WORKDIR /app

COPY . .

RUN go mod download

RUN CGO_ENABLED=0 GOOS=linux go build -o chat-reader cmd/chat-reader/main.go
RUN CGO_ENABLED=0 GOOS=linux go build -o create-topics cmd/create-topics/main.go

# ----

FROM alpine:3.20.3

WORKDIR /app

RUN addgroup -S nonroot && adduser -S nonroot -G nonroot

USER nonroot

COPY --from=build /app/chat-reader .
COPY --from=build /app/create-topics .

ENTRYPOINT ["/app/chat-reader"]