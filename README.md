## Nakaflow — Nakamoto Coefficient Calculator

Live Nakamoto coefficient data for 30+ proof-of-stake blockchains, plus an RSS feed, embeddable badges, and a public JSON API.

- Live site: [nakaflow.io](https://nakaflow.io)
- API docs: [nakaflow.io/docs](https://nakaflow.io/docs)

### What is the Nakamoto coefficient?

The [Nakamoto coefficient](https://news.earn.com/quantifying-decentralization-e39db233c28e) measures how decentralized a chain is: the minimum number of validators that would need to collude to disrupt consensus. Nakaflow computes it consistently across chains as:

```
nakamoto-coefficient = number of validators controlling 33% of total network stake
```

A goroutine refreshes all chains every 6 hours and persists snapshots to a local SQLite history DB.

See also: this [Messari report](https://messari.io/report/evaluating-validator-decentralization-geographic-and-infrastructure-distribution-in-proof-of-stake-networks) on operational decentralization in PoS networks.

#### Disclaimer

The same 33% threshold is applied to every chain. Some chains use different consensus-critical thresholds (e.g. 50%). Interpret the numbers with that context in mind and cross-verify before drawing conclusions. Feedback welcome on [Discord](https://discord.gg/Una8qmFg).

### Supported chains

Agoric · Algorand · Aptos · Avail · Avalanche · Base · BNB Smart Chain · Cardano · Celestia · Cosmos · Ethereum · Graph Protocol · Hedera · Hype · Juno · Mina · Monad · MultiversX · Namada · Nano · Near · Osmosis · Plume · Polkadot · Polygon · Pulsechain · Regen · Sei · Solana · Stargaze · Story · Sui · Thorchain

### Running locally

Requires Go 1.25+.

```shell
go run .
```

The server listens on `:8080`. On first run the history DB path defaults to `/app/data/nc_history.db`; if that isn't writable, history is disabled but live endpoints still work.

#### Environment variables

| Variable | Purpose |
|---|---|
| `SOLANA_API_KEY` | validators.app API key for Solana. [Sign up](https://www.validators.app/users/sign_up?locale=en&network=mainnet). |
| `RATED_API_KEY` | rated.network API key. |
| `SUBSCAN_API_KEY` | Subscan API key (Polkadot, Avail). |
| `NAKAFLOW_BASE_URL` | Public base URL used in RSS links and embed backlinks. Defaults to `https://nakaflow.io`. |

### Running with Docker

```shell
docker build . --platform=linux/amd64 -t nakaflow/nc-calc:latest

docker run --rm \
  -e SOLANA_API_KEY=... \
  -e RATED_API_KEY=... \
  -e SUBSCAN_API_KEY=... \
  -p 8080:8080 nakaflow/nc-calc:latest
```

Or bring up server + frontend together:

```shell
docker compose up
```

### API & Embeds

All endpoints are public, CORS-enabled (`*`), and served from `https://nakaflow.io`. Interactive reference: [`/docs`](https://nakaflow.io/docs). Machine-readable spec: [`/openapi.json`](https://nakaflow.io/openapi.json).

#### JSON API

| Endpoint | Description |
|---|---|
| `GET /naka-coeffs` | Current Nakamoto coefficient for every supported chain. |
| `GET /nc-history?chain={token}&days={n}` | Historical snapshots. `chain` optional; `days` defaults to 30 when omitted. |
| `GET /solana-details` | Per-entity breakdown of Solana's coefficient. |

```bash
curl https://nakaflow.io/naka-coeffs
curl "https://nakaflow.io/nc-history?chain=SOL&days=90"
```

#### RSS feed

`GET /feed.xml` — RSS 2.0 feed, one item per chain. Chains with recent changes are prioritized first. Wire it into any RSS reader, Zapier, or newsletter tool.

```
https://nakaflow.io/feed.xml
```

#### Badge (SVG)

`GET /embed/badge/{TOKEN}` returns a shields.io-style SVG. Color-coded: green (≥20), lime (≥10), yellow (≥5), red (<5). The `{TOKEN}` is case-insensitive — e.g. `SOL`, `ETH`, `ATOM`.

Markdown:
```markdown
[![Nakamoto Coefficient](https://nakaflow.io/embed/badge/SOL)](https://nakaflow.io/?chain=SOL)
```

HTML:
```html
<a href="https://nakaflow.io/?chain=SOL">
  <img src="https://nakaflow.io/embed/badge/SOL" alt="Solana Nakamoto Coefficient">
</a>
```

#### Widget (iframe)

`GET /embed/widget/{TOKEN}` returns a self-contained, dark-mode-aware HTML card suitable for embedding via `<iframe>`.

```html
<iframe
  src="https://nakaflow.io/embed/widget/SOL"
  width="360" height="140"
  frameborder="0" loading="lazy"
  title="Solana Nakamoto Coefficient"></iframe>
```

### Project layout

```
core/chains/    per-chain stake queries and coefficient math
main.go         HTTP server and route wiring (Gin)
db.go           SQLite history store
feed.go         RSS rendering
embed.go        SVG badge + iframe widget rendering
docs.go         OpenAPI spec + Swagger UI
```

### Contributing

To add a new chain:

1. Add a file in `core/chains/` that implements a function returning `(int, error)`.
2. Register the token in `core/chains/chain.go` — add to the `Token` constants, `ChainName()`, `Tokens`, and the `switch` in `newValues`.
3. Add the chain name to the list above.

PRs welcome.
