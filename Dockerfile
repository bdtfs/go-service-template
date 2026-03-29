FROM golang:1.26-alpine AS builder

WORKDIR /build

COPY go.mod go.sum ./
RUN go mod download

COPY . .

ARG RELEASE
ARG GIT_HASH
RUN CGO_ENABLED=0 GOOS=linux go build \
    -ldflags "-X main.Release=${RELEASE} -X main.GitHash=${GIT_HASH}" \
    -o /build/bin/service ./cmd/service

FROM alpine:3.21

RUN apk --no-cache add ca-certificates

COPY --from=builder /build/bin/service /app/service
COPY --from=builder /build/config.yaml /app/config.yaml

WORKDIR /app

RUN addgroup -S appgroup && adduser -S appuser -G appgroup
USER appuser

ENTRYPOINT ["/app/service"]
