# 🌙 Lume

[![License: AGPL v3](https://img.shields.io/badge/License-AGPL%20v3-blue.svg)](https://www.gnu.org/licenses/agpl-3.0)
[![Go Version](https://img.shields.io/badge/Go-1.26+-00ADD8?logo=go)](https://go.dev/)
[![Docker](https://img.shields.io/badge/Docker-ready-2496ED?logo=docker)](https://www.docker.com/)

**Privacy-first, self-hosted menstrual cycle tracker**

No tracking. No ads. No data mining. Your cycle data stays on YOUR server.

---

## Why Lume?

Commercial period trackers like Flo and Clue have been caught:
- Selling data to Meta/Facebook
- Sharing sensitive health data with advertisers
- Creating legal risks in jurisdictions with restrictive laws

**Lume is different:**
- 🔒 **100% self-hosted** - data never leaves your server
- 🚫 **No external tracking** - no analytics, no third-party requests
- 🌍 **Open source** (AGPL v3) - auditable, trustworthy
- 🎨 **Beautiful UI** - warm, journal-like design
- 🌐 **Multi-language** - Russian + English (more coming)

---

## Features

✅ **Cycle tracking** - Mark period days, flow intensity, symptoms
✅ **Smart predictions** - Next period, ovulation, fertile window
✅ **Calendar view** - Visual monthly overview
✅ **Statistics** - Cycle length, regularity, symptom patterns
✅ **Partner view** (optional) - Share calendar without private details
✅ **Telegram alerts** (optional) - Period reminders
✅ **Export data (CSV/JSON)** - Your data, your backup

---

## Quick Start

### Docker (recommended)

\\\ash
git clone https://github.com/terraincognita07/lume.git
cd lume
docker-compose up -d
\\\

Open http://localhost:8080

### Manual

Requirements: Go 1.26+, Node.js 18+

\\\ash
# Clone
git clone https://github.com/terraincognita07/lume.git
cd lume

# Build frontend
npm install
npm run build:css

# Run backend
go mod tidy
go run cmd/lume/main.go
\\\

---

## Configuration

Edit `docker-compose.yml` or set environment variables:

```env
# Core
TZ=UTC
DEFAULT_LANGUAGE=ru
SECRET_KEY=change_me_in_production
DB_PATH=data/lume.db
PORT=8080

# Optional notifications
TELEGRAM_BOT_TOKEN=
TELEGRAM_CHAT_ID=

# Rate limits (self-host safe defaults)
RATE_LIMIT_LOGIN_MAX=8
RATE_LIMIT_LOGIN_WINDOW=15m
RATE_LIMIT_FORGOT_PASSWORD_MAX=8
RATE_LIMIT_FORGOT_PASSWORD_WINDOW=1h
RATE_LIMIT_API_MAX=300
RATE_LIMIT_API_WINDOW=1m

# Reverse proxy trust (enable only behind your own proxy)
TRUST_PROXY_ENABLED=false
PROXY_HEADER=X-Forwarded-For
TRUSTED_PROXIES=127.0.0.1,::1
```

---

## Screenshots

*(Coming soon)*

---

## License

**AGPL v3** - This ensures that:
- ✅ You can use Lume freely forever
- ✅ You can modify and improve it
- ✅ Anyone hosting a modified version MUST share their code
- ✅ Commercial SaaS competitors must open-source their changes

See [LICENSE](LICENSE) for full text.

**Why AGPL?** We chose AGPL to protect the privacy-first mission. If someone builds a hosted service based on Lume, they must contribute improvements back to the community.

---

## Contributing

Contributions welcome! Please:
1. Open an issue first to discuss changes
2. Follow existing code style
3. Add tests for new features
4. Update documentation

---

## Roadmap

- [x] MVP: tracking, predictions, calendar
- [x] Multi-language (RU/EN)
- [ ] Mobile PWA
- [ ] Import from Flo/Clue
- [ ] Export to PDF (for doctors)
- [ ] End-to-end encryption (sync between devices)
- [ ] iOS/Android apps

---

## Support

- 🐛 **Issues:** [GitHub Issues](https://github.com/terraincognita07/lume/issues)
- 💬 **Discussions:** [GitHub Discussions](https://github.com/terraincognita07/lume/discussions)
- 📧 **Email:** lume@yourdomain.com *(update this)*

---

## Privacy

Lume is designed for privacy:
- ❌ No analytics or tracking
- ❌ No external API calls (except optional Telegram)
- ❌ No cookies (except authentication session)
- ✅ All data stored locally in SQLite
- ✅ You control backups and exports

---

## Alternatives

If Lume doesn't fit your needs:

- **drip** - FOSS Android app (offline only)
- **perioden** - Simple CLI tracker
- **Clue** - Commercial app (data concerns)
- **Flo** - Commercial app (data concerns)

---

## Credits

Built with ❤️ by [terraincognita07](https://github.com/terraincognita07)

Inspired by the need for privacy-respecting health tools.

---

**Star ⭐ this repo if you believe in privacy-first software!**
