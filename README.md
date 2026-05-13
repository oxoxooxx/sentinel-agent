# Sentinel Agent

Lightweight endpoint security monitoring agent that detects data exfiltration behaviors — including git repository cloning and uploads to AI services (ChatGPT, NotebookLM, Claude, Gemini, etc.).

> **Status:** Early development / MVP

## Features (v1 MVP)

- **Git activity monitoring** — detects `git clone` / `git pull` on endpoints, logs repo URL, user, and timestamp
- **AI service upload detection** — monitors outbound connections to known AI services, alerts on abnormal data volume
- **Centralized log collection** — agents report to an on-premise server
- **Telegram alerting** — notifies admin when thresholds are exceeded
- **Cross-platform** — supports Windows and macOS

## Roadmap

- [ ] v1: Agent (git + network monitoring) + Server (log collection + alerting)
- [ ] v2: Web Dashboard
- [ ] v2: Content scanning (keyword-based DLP)
- [ ] v3: Auto-block / quarantine
- [ ] v3: AI-powered behavior analysis

## Architecture

```
Endpoints (Win/Mac)
  └─ sentinel-agent
       ├─ Git operation monitor
       └─ AI service network monitor
            ↓ HTTP (internal network)
On-premise Server
  └─ sentinel-server
       ├─ Log aggregation (SQLite)
       └─ Alerting (Telegram)
```

## Tech Stack

- **Language:** Go (cross-platform, lightweight)
- **Storage:** SQLite
- **Alerting:** Telegram Bot API
- **License:** AGPLv3 (see LICENSE)

## License

AGPLv3 — free for personal and open-source use.  
For commercial or enterprise licensing: oxoxooxx@gmail.com

## Author

Kevin Tu / [KTLabs](https://github.com/oxoxooxx)
