# ADR-002: AI Gateway

| Field | Value |
|---|---|
| Title | AI Gateway |
| Version | 1.0 |
| Status | Draft — pending CTO/Founder approval |
| Owner | Erwin Sinaga (Founder & CEO) |
| Maintainer | V-Guard AI Engineering Team |
| Architecture Author | ChatGPT (AI Assistant) |
| Technical Review Support | Claude (AI Assistant) |
| QA Review Support | Gemini (AI Assistant) |
| Last Updated | 14 Juli 2026 |
| Effective Date | — |
| Related ADR | ADR-004 |
| Related Sprint | — |

## Context

Blueprint BAB V menetapkan prinsip **LLM Agnostic**: sistem tidak boleh bergantung pada satu provider AI, dan production wajib melalui V-Guard AI Gateway, bukan langsung ke vendor AI.

## Decision

- Seluruh pemanggilan AI Provider dari **production** wajib melalui **V-Guard AI Gateway**.
- Alur arsitektur mengikuti pola decoupled:

  ```
  Business Logic → Decision Engine → AI Gateway → AI Provider
  ```

- Tidak ada modul production yang memanggil vendor AI secara langsung.
- Omni Router (lihat ADR-004) hanya untuk development/eksperimen, tidak untuk production.

## Consequences

- Memudahkan penggantian/penambahan AI provider tanpa mengubah business logic.
- Menambah satu lapisan (Gateway) yang perlu dijaga availability dan latency-nya, mengingat SLO Fraud Detection Blueprint BAB IX (< 2 detik).
