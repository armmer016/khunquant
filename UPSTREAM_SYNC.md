# Upstream Sync: sipeed/picoclaw → armmer016/khunquant

## Overview

- **Upstream:** https://github.com/sipeed/picoclaw
- **Fork:** https://github.com/armmer016/khunquant
- **Diverged at:** `96fd4e0` (common ancestor)
- **Sync branch:** `sync/upstream-system-upgrades`
- **Strategy:** Cherry-pick only bug fixes, security fixes, performance, and dependency upgrades. Features are decided separately.
- **Last updated:** 2026-04-11

---

## How to Apply

```bash
# Setup (one-time)
git remote add upstream https://github.com/sipeed/picoclaw.git
git fetch upstream main
git checkout -b sync/upstream-system-upgrades

# Cherry-pick a commit
git cherry-pick <sha>

# If conflict occurs
# 1. Resolve conflict manually
# 2. git add <resolved-files>
# 3. git cherry-pick --continue

# Skip a commit that can't be applied cleanly
git cherry-pick --skip

# When done — merge into main
git checkout main
git merge --no-ff sync/upstream-system-upgrades -m "chore(sync): apply upstream system upgrades from sipeed/picoclaw"
```

---

## Progress

Legend: `[ ]` pending · `[x]` done · `[~]` skipped · `[!]` conflict resolved

---

### 🔴 Security (Apply First)

| Status | SHA | Description |
|--------|-----|-------------|
| `[x]` | `71e2b63` | fix: Use secure defaults for Pico channel, stop leaking token in URL |
| `[x]` | `c806598` | chore(web): upgrade eslint deps to resolve flatted vulnerability |
| `[x]` | `bda18f5` | chore(deps): upgrade eslint chain to resolve flatted vulnerability |
| `[x]` | `68d182a` | chore(deps): bump Go toolchain to 1.25.8 for stdlib security fixes |
| `[x]` | `fe87376` | chore(deps): upgrade modelcontextprotocol go-sdk v1.4.1 for security fixes |
| `[x]` | `f07a8a8` | chore(web): patch vulnerable frontend tooling dependencies |
| `[x]` | `ae021ef` | fix: more accurate deny pattern for disk wiping |
| `[x]` | `187189a` | fix(seahorse): sanitize user input for FTS5 MATCH queries (injection) |
| `[x]` | `8d5fc73` | security: add open-by-default warning and `*` allow_from support |

---

### 🟠 Agent Core Fixes

| Status | SHA | Description |
|--------|-----|-------------|
| `[x]` | `9c65d78` | fix(agent): forceCompression must not assume history[0] is system prompt |
| `[x]` | `d5fdd5e` | fix(agent): include ReasoningContent and Media in token estimation |
| `[x]` | `8034ee7` | fix(agent): correct media token arithmetic and tool call double-counting |
| `[x]` | `edbdc3b` | fix(agent): findSafeBoundary returns 0 for single-Turn history |
| `[x]` | `c63c644` | fix(agent): forceCompression recovers from single oversized Turn |
| `[x]` | `6b5d7e3` | fix(agent): resolve critical race conditions and resource leaks in SubTurn |
| `[x]` | `3c2d373` | fix(agent): resolve race conditions and resource leaks in SubTurn |
| `[x]` | `672d11c` | fix(agent): prevent double result delivery and panic bypass in SubTurn |
| `[x]` | `12a8590` | fix(agent): enhance SubTurn robustness and fix race conditions |
| `[x]` | `c7ea018` | fix(agent): prevent duplicate history during subturn context recoveries |
| `[x]` | `e20ff43` | fix(agent): resolve subturn deadlocks, panics and context retry state |
| `[x]` | `7868c58` | fix(agent): fix subturn panic result, hard abort rollback, and drain bus exit |
| `[x]` | `3611034` | fix(agent): implement Critical flag, complete tools.SubTurnConfig |
| `[x]` | `82d574e` | fix(agent): separate empty-response and tool-limit fallbacks |
| `[x]` | `276a0cb` | fix(agent): rebind provider after /switch model to |
| `[x]` | `844a4ee` | fix(agent): avoid process exit on exec init failure + regression test |
| `[x]` | `f93d2b4` | fix: Avoid failure of main agent process due to tool call failures |
| `[x]` | `85dfb34` | fix(agent): suppress heartbeat tool feedback |
| `[ ]` | `1c65866` | fix(agent): scope steering |
| `[x]` | `336d5d4` | fix(agent): route reasoning_content to reasoning channel |
| `[x]` | `93f391a` | fix(agent): include SystemParts in token estimation and add reasoning guards |
| `[x]` | `1a44752` | fix(agent): prevent double-counting system message tokens in estimator |
| `[x]` | `e011284` | fix(agent): use light provider for routed model calls |
| `[x]` | `bd88385` | fix(agent): gate pico interim publish for internal turns |
| `[x]` | `9ac21c5` | fix: add missing recover panic in subturn.go |
| `[x]` | `9c31b0c` | fix: bus closed with consumers having unfinished messages |
| `[x]` | `fcf406b` | fix(config): start model round robin from the first match |

---

### 🟠 Provider / API Fixes

| Status | SHA | Description |
|--------|-----|-------------|
| `[x]` | `54654d2` | fix(anthropic): skip tool calls with empty names to prevent API errors |
| `[x]` | `05c65d2` | fix(provider): skip empty anthropic tool names |
| `[x]` | `f81b44b` | fix(provider): deduplicate tool results and merge consecutive tool_result blocks for Anthropic |
| `[ ]` | `8d97896` | fix(providers): handle nil input in GLM series tool_use blocks |
| `[x]` | `97dec16` | fix(providers): improve context overflow detection and classification |
| `[x]` | `d014f3e` | fix(api): include auth header in local model probe |
| `[x]` | `38e1fe4` | fix(config): model_list inherits api_key/api_base from providers |
| `[x]` | `6ce0306` | fix: use per candidate provider for model_fallbacks |
| `[x]` | `f327859` | fix(api): enhance model availability probing with backoff and caching |
| `[x]` | `cd3f660` | fix(utils): honor Retry-After for 429 retries |

---

### 🟡 Config Fixes

| Status | SHA | Description |
|--------|-----|-------------|
| `[x]` | `5660b8f` | fix(heartbeat): ignore untouched default template |
| `[ ]` | `5c210e6` | fix(config): disable tool feedback by default |
| `[ ]` | `0c9e4f0` | fix: FlexibleStringSlice cause picoclaw start crash issue |
| `[ ]` | `f1cb7cc` | fix: gateway reload will cause pico stop working issue |
| `[ ]` | `9fb01bc` | fix(config): persist disabled placeholder settings |
| `[ ]` | `d23c24c` | fix(config): normalize empty security config before save/load |
| `[ ]` | `1f9d390` | fix: apply security credentials before config validation in web handlers |
| `[ ]` | `b17cbe5` | fix: apply security credentials before config validation in web handlers |
| `[ ]` | `6bd8fec` | fix: security config precedence during migration |
| `[ ]` | `8b6cbd9` | fix: Prevent security.yml from being overwritten during config migration |
| `[ ]` | `d385491` | fix(config): array placeholder |

---

### 🟡 Tools Fixes

| Status | SHA | Description |
|--------|-----|-------------|
| `[x]` | `1bc05e8` | fix(tools): allow sandbox access to temp media files |
| `[x]` | `bb1a414` | fix(tools): harden whitelist path resolution |
| `[x]` | `cef0f28` | fix(tools): normalize whitelist path checks for symlinked allowed roots |
| `[ ]` | `eb86e10` | fix(tools): propagate tool registry to subagents |
| `[ ]` | `29a161e` | fix(tools): prevent nil pointer dereference in spawn tools |
| `[ ]` | `89af3b2` | fix(tools): message tool no longer suppresses reply to originating chat |
| `[ ]` | `3e3b6ae` | fix(tools): message tool no longer suppresses reply to originating chat (#2180) |
| `[ ]` | `df4f322` | fix(tool): route binary outputs through the media pipeline |
| `[ ]` | `71337b6` | fix(tool): clarify write_file nested-JSON escape semantics + tests |

---

### 🟡 Gateway / Web Fixes

| Status | SHA | Description |
|--------|-----|-------------|
| `[ ]` | `1120718` | fix: proxy WebSocket through web server port |
| `[ ]` | `afe22c5` | bug fix: gateway should not start when gateway server is not running |
| `[ ]` | `c513ad2` | fix(web): refactor pico chat flow and fix proxied websocket URLs |
| `[ ]` | `778f939` | fix [BUG] WebUI cannot connect to gateway started by WebUI |
| `[ ]` | `7bf6cbe` | fix(gateway): harden PID liveness handling and websocket proxy state |
| `[ ]` | `6a8552a` | fix(web): derive WebSocket URL from browser location instead of backend |
| `[ ]` | `7d16764` | fix(gateway): validate PID ownership and clean stale pid files |
| `[ ]` | `4914187` | fix(gateway): log startup errors before exit |
| `[ ]` | `6aff5b7` | fix(pico): use O(1) session indexing and harden websocket concurrency |
| `[ ]` | `a9c76ec` | bug: fix picoToken is empty when gateway started by launcher |
| `[ ]` | `795ec9a` | fix(launcher): fall back to token auth on unsupported platforms |
| `[ ]` | `d997771` | fix(launcher): align react and react-dom versions |
| `[ ]` | `dd54601` | fix(web): hydrate cached Pico token for websocket proxy |
| `[ ]` | `cf9e049` | fix: launcher can't save model api_key issue |
| `[ ]` | `ffbcbea` | fix(web): persist api_key when adding models |
| `[ ]` | `257aa0f` | fix(channels): fail fast when all channel startups fail |
| `[ ]` | `5b596ed` | fix(chat): keep tool summaries and assistant output together |
| `[ ]` | `2aeed8f` | fix(pico): stream assistant text between tool calls |
| `[ ]` | `9982ee2` | fix(pico): avoid duplicate final websocket message |
| `[ ]` | `bd13092` | fix(review): align tool feedback reconstruction with runtime behavior |
| `[ ]` | `748ac58` | fix(chat): keep tool-call summary and assistant output in sync |

---

### 🟡 Logger / Secret Masking

| Status | SHA | Description |
|--------|-----|-------------|
| `[ ]` | `8fc36a4` | fix(logger): mask bot tokens in 3rd-party logger output |
| `[ ]` | `64ceb5a` | fix(logger): show first/last 4 chars of bot token for identification |
| `[ ]` | `ce16190` | fix(chat): avoid full secret exposure for 7-char secrets |
| `[ ]` | `f1ac1a1` | fix(web): ensure at least 40% of characters are masked for api key |
| `[ ]` | `1ace296` | fix: use fileEvent instead of event when appending fields for file logger |

---

### 🟡 Telegram / Channel Fixes

| Status | SHA | Description |
|--------|-----|-------------|
| `[x]` | `dc037f0` | fix(telegram): stop typing indicator when LLM fails or hangs |
| `[ ]` | `a1e8ee5` | fix(telegram): improve HTML chunking and preserve word boundaries |
| `[ ]` | `08fa9bb` | fix: agent triggered on empty message in telegram |
| `[ ]` | `bc0be17` | fix(identity): support negative integers in isNumeric for Telegram group IDs |
| `[ ]` | `f12c09b` | fix: retry on dimension failure for tg media upload |

---

### 🟡 Cron Fixes

| Status | SHA | Description |
|--------|-----|-------------|
| `[x]` | `f71eaaf` | fix(cron): default scheduled jobs to agent execution |
| `[ ]` | `e414b82` | fix(cron): publish agent response to outbound bus for cron-triggered jobs |
| `[ ]` | `61a899c` | fix(cron): update test to use OutboundChan instead of removed SubscribeOutbound |

---

### 🟡 Build / Platform Fixes

| Status | SHA | Description |
|--------|-----|-------------|
| `[ ]` | `2ccac18` | fix(build): exclude matrix on unsupported mipsle and netbsd targets |
| `[ ]` | `51f8285` | fix(build): disable Matrix gateway import on freebsd/arm |
| `[ ]` | `661ce5e` | fix(build): gate seahorse context manager on unsupported platforms |
| `[ ]` | `330de0c` | fix(agent): disable seahorse context manager on freebsd/arm |
| `[ ]` | `0499cda` | build: use WEB_GO for web targets and preserve backend dist directory |
| `[ ]` | `170ae09` | fix: windows make build error and support custom build env |
| `[ ]` | `a9720da` | fix(test): skip TestPrepareCommand_AppliesUserEnv on unsupported OS |
| `[ ]` | `5e44a99` | fix(docker): run self-built images as root for parity with release |
| `[ ]` | `9ec2783` | fix(docker): add -console flag and open network for launcher |
| `[ ]` | `e4f4afc` | fix(release): ignore nightly tags in goreleaser changelog |

---

### 🟢 Performance

| Status | SHA | Description |
|--------|-----|-------------|
| `[ ]` | `5e7545a` | perf: precompute BM25 index for repeated searches |
| `[ ]` | `230942d` | fix(loop): polling improvement |

---

### 🔵 Go Module Dependency Upgrades

| Status | SHA | Description |
|--------|-----|-------------|
| `[x]` | `45c01f4` | golang.org/x/oauth2: 0.35.0 → 0.36.0 |
| `[x]` | `dd93630` | github.com/mymmrac/telego: 1.6.0 → 1.7.0 |
| `[x]` | `e9d240d` | github.com/caarlos0/env/v11: 11.3.1 → 11.4.0 |
| `[x]` | `2f40a8c` | github.com/anthropics/anthropic-sdk-go (latest) |
| `[x]` | `43eb6fe` | github.com/github/copilot-sdk/go: 0.1.23 → 0.1.32 |
| `[ ]` | `80d9a90` | github.com/ergochat/irc-go: 0.5.0 → 0.6.0 |
| `[ ]` | `c9ac19c` | maunium.net/go/mautrix: 0.26.3 → 0.26.4 |
| `[ ]` | `82c78e8` | build(deps): upgrade pty + reorganize sqlite |
| `[ ]` | `d844bf3` | github.com/github/copilot-sdk/go: 0.1.32 → 0.2.0 |
| `[ ]` | `74dfd93` | golang.org/x/time: 0.14.0 → 0.15.0 |
| `[ ]` | `5c6e13e` | modernc.org/sqlite: 1.46.1 → 1.47.0 |
| `[ ]` | `b732abf` | github.com/rs/zerolog: 1.34.0 → 1.35.0 |
| `[ ]` | `c3e7396` | github.com/pion/rtp: 1.8.7 → 1.10.1 |
| `[ ]` | `29277d4` | modernc.org/sqlite: 1.47.0 → 1.48.0 |
| `[ ]` | `c71cd1e` | github.com/aws/aws-sdk-go-v2/config (latest) |
| `[ ]` | `01a33bb` | github.com/mymmrac/telego: 1.7.0 → 1.8.0 |
| `[ ]` | `919e9eb` | modernc.org/sqlite: 1.48.0 → 1.48.2 |
| `[ ]` | `c6d15da` | golang.org/x/sys: 0.42.0 → 0.43.0 |
| `[ ]` | `7788ed4` | github.com/modelcontextprotocol/go-sdk (latest) |

---

### 🔵 Frontend Dependency Upgrades

| Status | SHA | Description |
|--------|-----|-------------|
| `[x]` | `b8dfd0b` | jotai: 2.18.0 → 2.18.1 |
| `[x]` | `3bf8a27` | react-i18next: 16.5.4 → 16.5.8 |
| `[x]` | `99304d1` | dayjs: 1.11.19 → 1.11.20 |
| `[x]` | `4178b2c` | @tanstack/react-router (latest) |
| `[ ]` | `1fd6dd1` | shadcn: 4.0.5 → 4.0.8 |
| `[ ]` | `cff85cf` | tailwindcss: 4.2.1 → 4.2.2 |
| `[ ]` | `77d0c67` | @tabler/icons-react (latest) |
| `[ ]` | `7dc0d02` | i18next: 25.8.20 → 25.10.10 |
| `[ ]` | `465baba` | i18next: 26.0.1 → 26.0.3 |
| `[ ]` | `4169eb3` | react-i18next: 16.6.6 → 17.0.2 |
| `[ ]` | `7fd6772` | @tanstack/react-query (latest) |
| `[ ]` | `8aa110c` | shadcn: 4.1.1 → 4.1.2 |
| `[ ]` | `4840707` | jotai: 2.19.0 → 2.19.1 |
| `[ ]` | `1949314` | react: 19.2.4 → 19.2.5 |
| `[ ]` | `f1fe2db` | @tanstack/react-query (latest) |
| `[ ]` | `e58f00b` | shadcn: 4.1.2 → 4.2.0 |

---

## Conflict Log

Record conflicts and how they were resolved here.

| SHA | File(s) | Resolution |
|-----|---------|------------|
| — | — | — |

---

## Skipped Commits

Commits intentionally skipped with reason.

| SHA | Description | Reason |
|-----|-------------|--------|
| — | — | — |

---

## Summary

- **Total to cherry-pick:** 153
- **Done:** 65
- **Skipped:** 0
- **Conflicts resolved:** 0

### Done by category
| Category | Done | Total |
|----------|------|-------|
| 🔴 Security | 9 | 9 |
| 🟠 Agent Core | 26 | 27 |
| 🟠 Provider/API | 9 | 10 |
| 🟡 Config | 1 | 11 |
| 🟡 Tools | 3 | 9 |
| 🟡 Gateway/Web | 0 | 21 |
| 🟡 Logger/Masking | 0 | 5 |
| 🟡 Telegram/Channel | 1 | 5 |
| 🟡 Cron | 1 | 3 |
| 🟡 Build/Platform | 0 | 10 |
| 🟢 Performance | 0 | 2 |
| 🔵 Go Module Deps | 5 | 19 |
| 🔵 Frontend Deps | 4 | 16 |

---

## Verification Checklist

- [x] `make check` passes (deps + fmt + vet + test)
- [x] `make build` produces binary
- [ ] Custom fork features still work: Bitkub, SettTrade, MLX LM
- [ ] Merge sync branch into main with `--no-ff`
