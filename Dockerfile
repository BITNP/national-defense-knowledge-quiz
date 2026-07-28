FROM golang:1.26-alpine AS builder

WORKDIR /build
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 go build -o /build/server ./cmd/server
RUN CGO_ENABLED=0 go build -o /build/seed ./cmd/seed

FROM alpine:3.21

RUN apk add --no-cache ca-certificates tzdata

WORKDIR /app
COPY --from=builder /build/server /app/server
COPY --from=builder /build/seed /app/seed
COPY --from=builder /build/static /app/static
COPY --from=builder /build/config /app/config

EXPOSE 9000

CMD ["/app/server"]
