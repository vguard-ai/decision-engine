# V-Guard AI — Documentation Hub

## Project Overview

V-Guard AI adalah platform AI-Integrated POS dan Business Intelligence untuk UMKM hingga Enterprise di Indonesia, dengan fokus pada keamanan transaksi (fraud detection), efisiensi operasional, dan pengambilan keputusan berbasis AI. Lihat [Master Blueprint](architecture/01-V-Guard-AI-Master-Blueprint-v1.0.md) untuk visi, misi, dan governance lengkap.

## Reading Order

Untuk memahami proyek ini dari awal, baca dengan urutan berikut:

1. [Master Blueprint v1.0](architecture/01-V-Guard-AI-Master-Blueprint-v1.0.md) — Source of Truth, wajib dibaca pertama.
2. [Architecture Decisions](architecture/02-Architecture-Decisions.md) — ringkasan ADR yang berlaku.
3. [Engineering Standards](architecture/03-Engineering-Standards.md) — standar teknis operasional.
4. [Roadmap](architecture/04-Roadmap.md) — tahapan pengembangan.
5. [ADR Index](architecture/ADR/README.md) — detail tiap keputusan arsitektur.

## Documentation Navigation

| Dokumen | Deskripsi |
|---|---|
| [Master Blueprint](architecture/01-V-Guard-AI-Master-Blueprint-v1.0.md) | Constitution / Source of Truth proyek |
| [Architecture Decisions](architecture/02-Architecture-Decisions.md) | Daftar & status ADR |
| [Engineering Standards](architecture/03-Engineering-Standards.md) | Standar repository, git, testing, security |
| [Roadmap](architecture/04-Roadmap.md) | Rencana pengembangan per fase |
| [ADR Index](architecture/ADR/README.md) | Index lengkap Architecture Decision Records |
| [Diagrams](diagrams/README.md) | Placeholder diagram arsitektur (belum dibuat) |

## Repository Documentation Map

```
docs/
├── README.md                                   ← kamu di sini
├── architecture/
│   ├── 01-V-Guard-AI-Master-Blueprint-v1.0.md  ← Source of Truth
│   ├── 02-Architecture-Decisions.md
│   ├── 03-Engineering-Standards.md
│   ├── 04-Roadmap.md
│   └── ADR/
│       ├── README.md                           ← ADR Index
│       ├── ADR-001-Repository-Standard.md
│       ├── ADR-002-AI-Gateway.md
│       ├── ADR-003-Policy-Resolver.md
│       └── ADR-004-OmniRouter-Development-Policy.md
├── diagrams/     ← placeholder, lihat diagrams/README.md
├── operations/   ← belum diisi
├── product/      ← belum diisi
└── api/          ← belum diisi
```

## Status Dokumentasi

Blueprint bersifat **Approved for Implementation**. Dokumen turunan (Architecture Decisions, Engineering Standards, Roadmap, ADR-001–004) berstatus **Draft**, menunggu review CTO dan persetujuan Founder sesuai prosedur Change Management di Blueprint BAB X.
