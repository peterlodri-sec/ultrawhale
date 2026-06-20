# ultrawhale Docker

```sh
docker pull ghcr.io/peterlodri-sec/ultrawhale:v10.1.1
docker run -e DEEPSEEK_API_KEY=sk-... ghcr.io/peterlodri-sec/ultrawhale:v10.1.1

# Multi-arch build
docker buildx build --platform linux/amd64,linux/arm64 \
  -t ghcr.io/peterlodri-sec/ultrawhale:v10.1.1 \
  -t ghcr.io/peterlodri-sec/ultrawhale:latest \
  --push .
```
