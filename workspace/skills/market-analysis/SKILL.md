---
name: market-analysis
description: Analyse how current financial or crypto news impacts the user's portfolio. Always fetch live news and take a fresh portfolio snapshot before analysing — never use training data or ask the user for their holdings.
---

# Market Analysis

Use this workflow when the user asks about news impact on their portfolio, or requests a combined market + portfolio review.

## Workflow

1. **Fetch current news** — call `web_search` with a query like "crypto market news today" or "SET Thailand stock news today", or `web_fetch` a known news source. Use multiple sources for broader coverage.
2. **Capture portfolio state** — call `take_snapshot` (no parameters for all exchanges). This gives live balances and prices.
3. **Get live prices** — call `get_tickers` for the assets identified in the snapshot.
4. **Synthesize** — map news events to held assets. For each material news item, assess: does the user hold this asset? How large is the position? What is the potential impact direction?
5. **Present findings** — group by impact severity (high / medium / low). Always state the source and timestamp of news used.

## News Sources to Fetch

| Focus | Suggested query or URL |
|-------|------------------------|
| Crypto global | `web_search "crypto market news today"` |
| Crypto Thai | `web_search "crypto news Thailand today"` |
| Thai equities (SET) | `web_fetch https://www.set.or.th/en/market/news` |
| Macro / FX | `web_search "federal reserve interest rate news"` |

## Rules

- Never skip `take_snapshot` — portfolio data from memory or user input is unreliable.
- Never quote news from training knowledge — always fetch it fresh.
- If `web_fetch` fails for a URL, try `web_search` with equivalent keywords instead.
- State the timestamp and source of every news item cited in the analysis.
