# ultrawhale — multi-arch Docker image
# Build: docker buildx build --platform linux/amd64,linux/arm64 -t ghcr.io/peterlodri-sec/ultrawhale:v10.1.1 .
FROM golang:1.24-alpine AS builder
WORKDIR /app
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o ultrawhale ./cmd/whale
RUN CGO_ENABLED=0 GOOS=linux go build -o ultrawhale-shell ./cmd/ultrawhale-shell/

FROM alpine:3.20
RUN apk add --no-cache ca-certificates curl
COPY --from=builder /app/ultrawhale /usr/local/bin/ultrawhale
COPY --from=builder /app/ultrawhale-shell /usr/local/bin/ultrawhale-shell
EXPOSE 8585 9797
ENV DEEPSEEK_API_KEY=""
ENTRYPOINT ["ultrawhale"]
CMD ["--help"]
