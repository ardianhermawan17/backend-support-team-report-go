# syntax=docker/dockerfile:1.7
FROM golang:1.25 AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN --mount=type=cache,target=/go/pkg/mod go mod download
COPY . .
RUN --mount=type=cache,target=/root/.cache/go-build \
	go build -o /bin/api ./cmd/api

FROM gcr.io/distroless/base-debian12
WORKDIR /app
COPY --from=builder /bin/api /bin/api
COPY --from=builder /app/configs /app/configs
ENTRYPOINT ["/bin/api"]
