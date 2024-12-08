FROM node:22-alpine3.20 AS node-build

WORKDIR /app/frontend

COPY ./frontend .

ARG VITE_WEBSOCKET_URL

RUN npm ci && npm run build

# ----

FROM golang:1.23.3-alpine3.20 AS go-build

WORKDIR /app/backend

COPY ./backend .

RUN go mod download

RUN CGO_ENABLED=0 GOOS=linux go build -o website cmd/website/main.go

# ----

FROM alpine:3.20.3

WORKDIR /app

RUN addgroup -S nonroot && adduser -S nonroot -G nonroot

USER nonroot

COPY --from=go-build /app/backend/website /app/website
COPY --from=node-build /app/frontend/dist /app/public

ENTRYPOINT ["/app/website"]