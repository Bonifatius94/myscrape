# myscrape-go — multi-stage build to a tiny static binary image.
# No interpreter, no browser, no GPU in the image itself: web_search/web_fetch and
# the default extractive web_research are all GPU-free. LLM synthesis (opt-in) talks
# to an external Ollama (see docker-compose.yml).

FROM golang:1.25-alpine AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
# Static, stripped binary (go-trafilatura is pure Go, so CGO is off).
RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /out/myscrape ./cmd/myscrape

FROM gcr.io/distroless/static-debian12:nonroot
COPY --from=build /out/myscrape /myscrape
# Serve over HTTP in a container by default; GPU-free synthesis.
ENV MYSCRAPE_MCP_TRANSPORT=streamable-http \
    MYSCRAPE_MCP_HOST=0.0.0.0 \
    MYSCRAPE_MCP_PORT=8000 \
    MYSCRAPE_RESEARCH_SYNTHESIS=simple
EXPOSE 8000
ENTRYPOINT ["/myscrape"]
