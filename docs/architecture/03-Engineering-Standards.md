# Engineering Standards

| Field | Value |
|---|---|
| Title | Engineering Standards |
| Version | 1.1 |
| Status | Draft — Derived from Blueprint v1.0 BAB VII, pending Founder/CTO sign-off |
| Owner | Erwin Sinaga (Founder & CEO) |
| Maintainer | V-Guard AI Engineering Team |
| Architecture Author | ChatGPT (AI Assistant) |
| Technical Review Support | Claude (AI Assistant) |
| QA Review Support | Gemini (AI Assistant) |
| Last Updated | 14 Juli 2026 |
| Effective Date | — |
| Repository | vguard-ai/decision-engine |
| Related ADR | ADR-001 |
| Related Sprint | ZIP 9 |

> **Catatan:** Dokumen ini mengembangkan poin-poin BAB VII Master Blueprint ("Repository Hygiene wajib dipatuhi, semua commit mengikuti standar, semua perubahan Architecture menggunakan ADR, tidak boleh hardcode secret, semua perubahan besar wajib melalui Blueprint") menjadi standar teknis operasional. Versi 1.1 hanya mengorganisasi ulang isi versi 1.0 ke dalam chapter dan menambahkan chapter baru (Semantic Versioning, Code Review Checklist, Security Checklist, Release Checklist, Documentation Rules) — tidak ada kebijakan yang sudah ada di versi 1.0 yang diubah maknanya. Detail spesifik (nama tool CI, format commit-lint, dsb.) tetap perlu dikonfirmasi CTO sebelum berlaku efektif.

## Daftar Isi

1. [Repository Standards](#1-repository-standards)
2. [Git Branch Strategy](#2-git-branch-strategy)
3. [Commit Convention](#3-commit-convention)
4. [Semantic Versioning](#4-semantic-versioning)
5. [Pull Request Rules](#5-pull-request-rules)
6. [Code Review Checklist](#6-code-review-checklist)
7. [Security Checklist](#7-security-checklist)
8. [Release Checklist](#8-release-checklist)
9. [Documentation Rules](#9-documentation-rules)

---

## 1. Repository Standards

- Struktur folder mengikuti konvensi yang sudah ditetapkan (`contracts/`, `internal/`, `docs/`) — lihat [ADR-001](ADR/ADR-001-Repository-Standard.md).
- Tidak ada file sisa (build artifact, file sementara, `.DS_Store`, dsb.) yang ikut ter-commit.
- README wajib ada di root dan di setiap modul non-trivial.

## 2. Git Branch Strategy

- `main` — production-ready, selalu dalam keadaan deployable. Tidak ada commit langsung tanpa review.
- `feature/<nama-singkat>` — pengembangan fitur baru.
- `fix/<nama-singkat>` — perbaikan bug.
- `docs/<nama-singkat>` — perubahan dokumentasi murni.
- Semua pekerjaan digabungkan ke `main` lewat Pull Request.

## 3. Commit Convention

Format: `[<Work Item ID>] <Deskripsi singkat>`

Contoh: `[B3-005] Policy Resolver (FRD-001 Vertical Slice)`

## 4. Semantic Versioning

- Blueprint dan dokumen governance menggunakan versi `MAJOR.MINOR` (contoh: Blueprint tetap `v1.0` sampai ada perubahan yang disetujui lewat ADR + persetujuan Founder — lihat BAB X).
- Untuk komponen kode (bila/ketika dirilis sebagai package/API), disarankan mengikuti [SemVer](https://semver.org) standar: `MAJOR.MINOR.PATCH`.
  - `MAJOR` — perubahan yang memutus backward compatibility.
  - `MINOR` — penambahan fungsi yang backward-compatible (mis. rule baru di Policy Resolver).
  - `PATCH` — perbaikan bug tanpa perubahan perilaku kontrak.
- Perubahan pada `contracts/` mengikuti aturan Backward Compatibility (lihat §7 Security Checklist & ADR terkait).

## 5. Pull Request Rules

- Wajib deskripsi jelas: tujuan, perubahan, cara test.
- Wajib lolos test otomatis sebelum merge.
- Perubahan arsitektur wajib menyertakan referensi ADR.

## 6. Code Review Checklist

- [ ] Kode mengikuti idiom bahasa yang digunakan (untuk Go: `gofmt`, `go vet` bersih).
- [ ] Tidak ada perubahan struktur/paket di luar scope task tanpa ADR.
- [ ] Nama package dan fungsi deskriptif, konsisten dengan struktur `internal/` yang sudah ada.
- [ ] Unit test mencakup kondisi normal (ALLOW/REVIEW/BLOCK untuk Decision Engine), boundary condition, dan input tidak valid.
- [ ] Test deterministik — tidak bergantung pada waktu eksekusi atau random tanpa seed.
- [ ] Tidak ada perubahan pada `contracts/` tanpa ADR eksplisit.

## 7. Security Checklist

- [ ] Tidak ada secret, API key, atau credential di dalam kode maupun riwayat commit (Zero Secret Policy).
- [ ] Kredensial disimpan di secret manager, bukan di kode atau file konfigurasi yang di-commit.
- [ ] File `.env` dan sejenisnya masuk `.gitignore`.
- [ ] Data sensitif (PII) melalui masking/hashing/encryption sesuai Blueprint BAB VI sebelum keluar ke AI Provider.
- [ ] Dependency dicek terhadap known vulnerability.
- [ ] Panggilan ke AI Provider dari production hanya lewat V-Guard AI Gateway (lihat [ADR-002](ADR/ADR-002-AI-Gateway.md)), tidak langsung ke vendor.

## 8. Release Checklist

- [ ] Semua item Code Review Checklist dan Security Checklist terpenuhi.
- [ ] Versi didokumentasikan sesuai §4 Semantic Versioning.
- [ ] `git log --oneline` dan `git diff --stat` direview sebelum merge ke `main`.
- [ ] Perubahan besar (arsitektur/roadmap/governance) sudah melalui ADR dan persetujuan Founder sesuai Blueprint BAB X.
- [ ] Tidak ada push ke `main`/production tanpa persetujuan eksplisit sesuai directive yang berlaku.

## 9. Documentation Rules

- Setiap dokumen di `docs/` wajib memakai header metadata standar (Title, Version, Status, Owner, Maintainer, Architecture Author, Technical Review Support, QA Review Support, Last Updated, Effective Date, Repository, Related ADR, Related Sprint).
- Perubahan pada Master Blueprint (isi/makna) hanya lewat ADR + persetujuan Founder; perubahan formatting murni tidak memerlukan ADR.
- Dokumen turunan (bukan Blueprint) berstatus **Draft** sampai direview CTO dan disetujui Founder.
- Backward Compatibility: perubahan pada `contracts/` (mis. `decision_request.go`) tidak boleh memutus konsumer yang ada, kecuali lewat ADR dan versi baru; field yang sudah publish tidak dihapus/diganti tipe tanpa deprecation path.
