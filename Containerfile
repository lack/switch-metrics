FROM golang:alpine AS builder
WORKDIR /build
COPY ./ ./
RUN go mod download
RUN go build -ldflags '-extldflags "-static"' ./cmd/switch-metrics

FROM golang:alpine
WORKDIR /config
WORKDIR /app
COPY --from=builder /build/switch-metrics /app/
ENV XDG_CONFIG_HOME=/config
ENV LISTEN_PORT=2121
EXPOSE ${LISTEN_PORT}
ENTRYPOINT ["/app/switch-metrics"]
