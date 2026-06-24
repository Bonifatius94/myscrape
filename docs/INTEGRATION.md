# Integrating myscrape into another product

**No special sauce.** myscrape is a standard MCP server exposing `web_search`,
`web_fetch`, `web_research`. Any MCP-capable agent consumes it by registering the
server — no SDK, no per-product glue. The Go build ships as a compiled binary (no
interpreter), so registration is just a path or a URL.

## stdio — agent spawns the binary
Generic `mcpServers` config (Claude Code/Desktop and most clients):
```json
{
  "mcpServers": {
    "myscrape": {
      "command": "/path/to/myscrape",
      "env": { "MYSCRAPE_TAVILY_API_KEY": "..." }
    }
  }
}
```
Defaults to stdio. Provider keys come from the environment or a local `.env` in the
working directory.

## HTTP — point at the shared service
Run one instance (streamable-HTTP on `:8000/mcp`):
```bash
docker compose up -d --build          # GPU-free (extractive)
```
Then every consumer registers the URL:
```json
{ "mcpServers": { "myscrape": { "type": "http", "url": "http://YOUR_HOST:8000/mcp" } } }
```

## Choosing
| | stdio | HTTP service |
|--|-------|--------------|
| Run | agent spawns the binary | one long-running instance |
| Best for | a single local agent | several agents/products sharing one box |
| Keys/LLM | per agent (`.env`) | once, on the host |

## Operational notes
- **Keys/quota:** set provider keys in the environment; the round-robin spreads load
  and fails over (see `specs/PROVIDERS.md`).
- **LLM/GPU:** default synthesis is extractive (no GPU). Set
  `MYSCRAPE_RESEARCH_SYNTHESIS=llm` + `MYSCRAPE_LLM_BASE_URL` to opt in; the
  server-mode cap (`MYSCRAPE_MAX_CONCURRENT_RESEARCH`, default 2) bounds GPU
  contention.
- **Auth:** the MCP endpoint has no built-in auth — put a proxy/token in front of
  `:8000` if it's reachable beyond localhost.
- Tool contracts and error codes: `specs/SPEC.md`.
