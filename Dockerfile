FROM golang:1.26-alpine AS builder

WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .

ARG BINARY=server
RUN go build -o /app ./cmd/${BINARY}

FROM alpine:3.19

RUN adduser -D -u 10001 appuser
USER appuser

COPY --from=builder /app /app

EXPOSE 8080

ENTRYPOINT ["/app"]