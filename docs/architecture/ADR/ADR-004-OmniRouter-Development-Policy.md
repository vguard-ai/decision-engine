# ADR-004: Omni Router Development Policy

| Field | Value |
|---|---|
| Title | Omni Router Development Policy |
| Version | 1.0 |
| Status | Draft — pending CTO/Founder approval |
| Owner | Erwin Sinaga (Founder & CEO) |
| Maintainer | V-Guard AI Engineering Team |
| Architecture Author | ChatGPT (AI Assistant) |
| Technical Review Support | Claude (AI Assistant) |
| QA Review Support | Gemini (AI Assistant) |
| Last Updated | 14 Juli 2026 |
| Effective Date | — |
| Related ADR | ADR-002 |
| Related Sprint | — |

## Context

Blueprint BAB VIII menetapkan Omni Router sebagai tool khusus development, digunakan lintas AI assistant (ChatGPT, Claude, Manus, Gemini) untuk riset dan eksperimen multi-model, dengan tujuan mengurangi vendor lock-in dan mengoptimalkan biaya R&D.

## Decision

- **Omni Router** hanya boleh digunakan untuk: Development, Research, Experiment, Testing.
- **Production** wajib tetap menggunakan **V-Guard AI Gateway** (lihat ADR-002) — tidak ada jalur produksi yang melewati Omni Router.
- Endpoint Omni Router bersifat satu pintu untuk kebutuhan eksperimen multi-model selama fase development.

## Consequences

- Tim dapat bereksperimen dengan berbagai model AI tanpa risiko terhadap sistem production.
- Perlu kontrol/monitoring terpisah agar penggunaan Omni Router pada environment development tidak secara tidak sengaja bocor ke jalur production.
