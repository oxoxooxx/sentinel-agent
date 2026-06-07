# Sentinel

輕量資安監控平台，跨來源收事件（FortiGate syslog、NAS log、自訂 webhook），LINE / Telegram / Teams 告警 + 一鍵處理。

> **狀態：** 開發中 / MVP

## 快速啟動

```bash
# 1. 複製設定範例並填入 token
cp sentinel/config.example.yaml sentinel/config.yaml
# 編輯 config.yaml，填入 LINE token 或 Telegram bot_token

# 2. 啟動
cd sentinel
docker-compose up -d

# 3. 確認健康狀態
curl http://localhost:8080/health
```

## 支援的事件來源

| 來源 | 協定 | 說明 |
|------|------|------|
| FortiGate | UDP Syslog :514 | 防火牆事件，自動解析嚴重等級 |
| NAS（Synology / QNAP）| HTTP Webhook | POST `/webhook` JSON 格式 |
| 自訂系統 | HTTP Webhook | 任何可發 HTTP POST 的系統 |

## 支援的通知頻道

| 頻道 | 設定 |
|------|------|
| LINE Notify | `alert.channels[].type: line` |
| Telegram Bot | `alert.channels[].type: telegram` |
| Microsoft Teams | `alert.channels[].type: teams` |

## API 端點

所有端點需在 Header 帶入 `Authorization: Bearer <secret>`（對應 `api.secret`）。

| 方法 | 路徑 | 說明 |
|------|------|------|
| GET | `/api/events` | 查詢事件列表 |
| GET | `/api/alerts` | 查詢告警列表 |
| PUT | `/api/alerts/:id/ack` | 確認告警 |
| PUT | `/api/alerts/:id/resolve` | 解決告警 |
| GET | `/api/rules` | 查詢告警規則 |
| POST | `/api/rules` | 新增告警規則 |
| GET | `/api/sources` | 查詢事件來源 |
| POST | `/api/sources` | 新增事件來源 |

## Webhook 格式

```bash
curl -X POST http://localhost:8080/webhook \
  -H "Authorization: Bearer change-me" \
  -H "Content-Type: application/json" \
  -d '{
    "source": "synology-nas",
    "severity": "warning",
    "message": "異常登入嘗試偵測"
  }'
```

## 架構

```
外部來源
  ├─ FortiGate → UDP Syslog :514
  └─ NAS / 自訂 → HTTP Webhook :8080/webhook

Sentinel Server
  ├─ ingestion：syslog + webhook 接收
  ├─ rules：規則引擎（關鍵字 / 閾值 / 白名單）
  ├─ storage：SQLite（預設）或 PostgreSQL
  ├─ alert：LINE / Telegram / Teams 派送
  └─ api：REST API :8080
```

## 技術棧

- **語言：** Go 1.22
- **框架：** Gin
- **儲存：** SQLite（預設）/ PostgreSQL（選用）
- **部署：** Docker + docker-compose

## License

AGPLv3 — 個人與開源免費使用。
商業或企業授權：oxoxooxx@gmail.com

## Author

Kevin Tu / [KTLabs](https://github.com/oxoxooxx)
