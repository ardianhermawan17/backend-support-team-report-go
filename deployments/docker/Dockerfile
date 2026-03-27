FROM golang:1.25 AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -o /bin/api ./cmd/api

FROM gcr.io/distroless/base-debian12
WORKDIR /app
COPY --from=builder /bin/api /bin/api
COPY --from=builder /app/configs /app/configs
ENTRYPOINT ["/bin/api"]
