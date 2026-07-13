# ADR-001: Repository Standard

| Field | Value |
|---|---|
| Title | Repository Standard |
| Version | 1.0 |
| Status | Draft — pending CTO/Founder approval |
| Owner | Erwin Sinaga (Founder & CEO) |
| Maintainer | V-Guard AI Engineering Team |
| Architecture Author | ChatGPT (AI Assistant) |
| Technical Review Support | Claude (AI Assistant) |
| QA Review Support | Gemini (AI Assistant) |
| Last Updated | 14 Juli 2026 |
| Effective Date | — |
| Related ADR | — |
| Related Sprint | ZIP 9 |

## Context

Blueprint BAB VII mensyaratkan Repository Hygiene wajib dipatuhi, seluruh commit mengikuti standar, dan tidak boleh ada hardcoded secret. Repository `vguard-ai/decision-engine` saat ini menggunakan struktur `contracts/`, `internal/runtime/`, `internal/config/`, `internal/validator/`, `internal/health/`.

## Decision

Struktur repository mengikuti pola berikut:

- `contracts/` — kontrak data lintas modul (mis. `decision_request.go`), tidak diubah tanpa ADR terpisah.
- `internal/` — implementasi internal, dipecah per domain (`runtime`, `config`, `validator`, `health`, dan modul baru seperti `policy`).
- `docs/` — dokumentasi arsitektur dan governance.

Setiap penambahan folder domain baru di `internal/` mengikuti pola paket domain yang sudah ada, tanpa mengubah paket yang sudah berjalan.

## Consequences

- Konsistensi struktur memudahkan onboarding dan review.
- Perubahan struktur besar (menambah/menghapus top-level folder) wajib melalui ADR baru.
