# MCP Server (playground)

This branch includes a **playground MCP server** to allow LLMs/clients to call the LIA toolchain and (in the future) pause for elicitation.

> This is **not in develop** and is intentionally experimental.

## How to run

```bash
go run ./cmd/lia-mcp
```

The server reads JSON-RPC messages from stdin and writes responses to stdout.

## Tools (current MVP)

- `lia.parse` — `.lia` -> `.liao`
- `lia.check` — validate `.liao`/`.lial`
- `lia.link` — link inputs into `.lial` + decision log
- `lia.lower` — `.lial` -> target language (python/java)
- `lia.explain` — hashes + summary
- `lia.build` — parse + check + link + optional lower

## Elicitation (future)

The design goal is to allow the server to **pause and ask questions** (business decisions) during build/link. In this MVP, tools return diagnostics only; elicitation hooks are planned.

## Notes
- Default pack directories: `./docs/spec/packs` and `./packs`
- This server is a lightweight JSON-RPC implementation intended for experimentation only.
