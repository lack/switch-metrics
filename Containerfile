FROM golang:latest AS builder
WORKDIR /build
COPY ./ ./
RUN go mod download
RUN go build ./cmd/switch-metrics

FROM scratch
WORKDIR /config
WORKDIR /app
COPY --from=builder /build/switch-metrics ./
ENV XDG_CONFIG_HOME=/config
ENV LISTEN_PORT=2121
EXPOSE ${LISTEN_PORT}
ENTRYPOINT ["./switch-metrics"]
