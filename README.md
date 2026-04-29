# websearch-mcp-server

![Status](https://img.shields.io/badge/status-alpha-orange)
![WIP](https://img.shields.io/badge/🚧-WIP-yellow)

Advanced web search tools and CLI for agents via MCP.

## Features

- **Multi-engine web search** — DuckDuckGo and Tavily, selectable per query
- **Content fetching** — fetch full page content by URL (browser-style)

## Quick Start

```bash
go build -o websearch-mcp-server .
./websearch-mcp-server
```

Server listens on port `8848` (configurable via `PORT` env var).

## Environment Variables

| Variable | Default | Description |
|----------|---------|-------------|
| `PORT` | `8848` | Server listen port |
| `SEARCH_ENGINE` | `duckduckgo` | Default search engine (`duckduckgo` or `tavily`) |
| `TAVILY_API_KEY` | — | API key for Tavily search |

## MCP Tools

### `search`

Search the web and return results with URL, title, and snippet.

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `query` | string | yes | Search query |
| `engine` | string | no | `duckduckgo` or `tavily` (default: `SEARCH_ENGINE` env or `duckduckgo`) |

### `fetch_content`

Fetch the full text content of a web page.

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `url` | string | yes | URL of the page to fetch |

## CLI

For AI agents that prefer direct invocation without MCP protocol overhead:

```bash
# Search
websearch-mcp-server search --query "golang best practices" --engine duckduckgo

# Fetch
websearch-mcp-server fetch --url "https://example.com"
```

Outputs JSON to stdout. Exit code 0 on success, non-zero on failure.
