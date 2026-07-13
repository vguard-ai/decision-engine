# ADR-003: Policy Resolver

| Field | Value |
|---|---|
| Title | Policy Resolver |
| Version | 1.0 |
| Status | Draft — pending CTO/Founder approval |
| Owner | Erwin Sinaga (Founder & CEO) |
| Maintainer | V-Guard AI Engineering Team |
| Architecture Author | ChatGPT (AI Assistant) |
| Technical Review Support | Claude (AI Assistant) |
| QA Review Support | Gemini (AI Assistant) |
| Last Updated | 14 Juli 2026 |
| Effective Date | — |
| Related ADR | ADR-001 |
| Related Sprint | ZIP 9 (B3-005) |

## Context

Blueprint BAB III (Fraud Detection) dan BAB XI Phase 2 menempatkan Policy Resolver sebagai bagian dari Decision Engine, dievaluasi di sprint ZIP 9 (Work Item B3-005). Resolver ini bertugas mengevaluasi request transaksi terhadap kumpulan rule fraud dan mengembalikan keputusan.

## Decision

- Policy Resolver bersifat **deterministic rule evaluation** — tidak melibatkan AI inference atau LLM (selaras dengan BAB V: AI hanya dipanggil lewat AI Gateway, dan Policy Resolver bukan bagian dari alur AI).
- Output keputusan terbatas pada tiga nilai: `ALLOW`, `REVIEW`, `BLOCK`.
- Diletakkan di `internal/policy/`, mengikuti pola paket domain pada ADR-001, tanpa mengubah `contracts/` maupun `internal/runtime/` yang sudah ada.
- Resolver dirancang dapat menerima rule tambahan di versi berikutnya (extensible rule set), tanpa mengubah kontrak keputusan yang sudah ada.

## Consequences

- Karena deterministik, hasil resolver dapat diuji secara penuh dengan unit test tanpa mocking model AI.
- Penambahan rule baru di masa depan berpotensi memerlukan ADR tambahan bila mengubah bentuk kontrak (`decision_request.go`) atau alur runtime.
