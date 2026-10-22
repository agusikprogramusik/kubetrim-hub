# KubeTrim Hub

> **Enterprise FinOps Control Plane, Multi-Cluster Attribution Engine & GitOps Automation Hub for Kubernetes**
> *Current Version: `0.1.0-dev` | Backend: `Go 1.27.1` | Frontend: `Next.js 16 / React 19 / TailwindCSS v4`*

KubeTrim Hub is the central control plane for the [`kubetrim`](../kubetrim) ecosystem. It ingests read-only cluster audit snapshots (`AuditReport` v1), tracks realized before-and-after savings ($T_0 \to T_1$ via deterministic `ConfigHash`), detects 14-day sawtooth memory leaks, synthesizes 1-Click GitOps Pull Requests, and delivers an **Avant-Garde Tactile Editorial & Precision UI** Executive Dashboard.

---

## 🏛️ Architecture Overview (Modular Monolith)

```text
kubetrim-hub/
├── apps/
│   ├── api/                     # Go 1.27.1 REST & SSE server (hub-api)
│   ├── worker/                  # Go 1.27.1 async processor (hub-worker)
│   └── web/                     # Next.js 16 + React 19 Frontend (hub-web)
├── pkg/                         # Shared Hub Domain Packages
│   ├── domain/                  # Core entities (Cluster, Snapshot, Attribution, PR)
│   ├── ingestion/               # Snapshot receiver & schema v1 validation
│   ├── diff/                    # T0 -> T1 ROI attribution & SLA regression guard
│   ├── anomaly/                 # 14-day sawtooth memory leak detector
│   ├── gitops/                  # 1-Click PR automation (GitHub / GitLab API)
│   ├── copilot/                 # FinOps RAG pipeline over PostgreSQL 18 pgvector
│   ├── notifications/           # Slack & MS Teams executive digest bots
│   └── storage/                 # PostgreSQL 18 + TimescaleDB + pgvector repository
├── docs/
│   └── architecture/
│       └── design-philosophy.md # October 2026 Tactile Editorial & Precision UI spec
├── DESIGN_SYSTEM.md             # Canonical UI/UX Philosophy & 2-Step Design Prompt
├── AGENTS.md                    # Multi-agent engineering & design invariants
└── CHANGELOG.md                 # Semantic versioning release history
```

---

## 🎨 UI/UX Design Standard (October 2026)

All frontend interfaces in `kubetrim-hub` follow the **Tactile Editorial & Precision UI** philosophy documented in [`DESIGN_SYSTEM.md`](./DESIGN_SYSTEM.md) and [`AGENTS.md`](./AGENTS.md):
* **Zero AI/SaaS Clichés:** No indigo/purple gradients, no 3-column centered icon cards, no blurry drop shadows.
* **Typography Trio:** Editorial Serif (`Instrument Serif` / `Newsreader`) + Geometric Sans (`Plus Jakarta Sans` / `Inter`) + Technical Monospace (`JetBrains Mono`).
* **Swiss Blueprint Grid:** Exposed `1px solid` structural hairlines, SVG `fractalNoise` texture (`0.035` opacity), and mechanical `cubic-bezier(0.22, 1, 0.36, 1)` micro-interactions.
