# ADR Index

| Field | Value |
|---|---|
| Title | Architecture Decision Records — Index |
| Version | 1.0 |
| Status | Draft — pending CTO/Founder approval |
| Owner | Erwin Sinaga (Founder & CEO) |
| Maintainer | V-Guard AI Engineering Team |
| Architecture Author | ChatGPT (AI Assistant) |
| Technical Review Support | Claude (AI Assistant) |
| QA Review Support | Gemini (AI Assistant) |
| Last Updated | 14 Juli 2026 |
| Effective Date | — |
| Repository | vguard-ai/decision-engine |
| Related ADR | — |
| Related Sprint | ZIP 9 |

## Index

| ADR ID | Title | Status | Related Sprint | Description |
|---|---|---|---|---|
| [ADR-001](ADR-001-Repository-Standard.md) | Repository Standard | Draft | — | Struktur folder `contracts/`, `internal/`, `docs/` dan aturan penambahan modul baru. |
| [ADR-002](ADR-002-AI-Gateway.md) | AI Gateway | Draft | — | Production wajib melalui V-Guard AI Gateway; prinsip LLM Agnostic dan decoupled architecture. |
| [ADR-003](ADR-003-Policy-Resolver.md) | Policy Resolver | Draft | ZIP 9 (B3-005) | Resolver deterministik (ALLOW/REVIEW/BLOCK) tanpa AI inference, di `internal/policy/`. |
| [ADR-004](ADR-004-OmniRouter-Development-Policy.md) | Omni Router Development Policy | Draft | — | Omni Router hanya untuk development/riset; production tetap lewat AI Gateway. |

## Cara Menambah ADR Baru

1. Salin format header standar (lihat dokumen ADR mana pun sebagai template).
2. Beri nomor urut berikutnya, mis. `ADR-005-<Judul-Singkat>.md`.
3. Ikuti prosedur Change Management di Blueprint BAB X sebelum status diubah dari Draft ke Approved.
4. Tambahkan baris baru ke tabel index di atas.
