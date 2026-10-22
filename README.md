# KubeTrim Hub

> **Enterprise FinOps Control Plane, Multi-Cluster Attribution Engine & GitOps Automation Hub for Kubernetes**
> *Current Version: `0.2.2-dev` | Backend: `Go 1.27.1` | Frontend Repo: [`kubetrim-web`](../kubetrim-web) (`Next.js 16 / React 19 / TailwindCSS v4`)*

KubeTrim Hub is the central control plane for the [`kubetrim`](../kubetrim) ecosystem. It ingests read-only cluster audit snapshots (`AuditReport` v1), tracks realized before-and-after savings ($T_0 \to T_1$ via deterministic `ConfigHash`), detects 14-day sawtooth memory leaks, synthesizes 1-Click GitOps Pull Requests, and delivers an **Avant-Garde Tactile Editorial & Precision UI** Executive Dashboard.

---

## 🏛️ Architecture Overview (Modular Monolith)

```text
kubetrim-hub/
├── apps/
│   ├── api/                     # Go 1.27.1 REST & SSE server (hub-api)
│   ├── worker/                  # Go 1.27.1 async processor (hub-worker)
│   └── web/                     # Next.js 16 + React 19 Frontend (hub-web)
├── pkg/                         # Shared Hub Domain Packages (>85% TDD Coverage)
│   ├── domain/                  # Core entities (AuditReport, Attribution, Anomaly, PR)
│   ├── api/                     # Gin v1.12+ HTTP router, auth middleware & handlers
│   ├── diff/                    # T0 -> T1 ROI attribution & SLA regression guard
│   ├── anomaly/                 # 14-day sawtooth memory leak & Idle GPU detector
│   ├── gitops/                  # 1-Click PR automation (Kustomize v5 + GitHub API)
│   ├── copilot/                 # FinOps RAG & Qwen3.5-9B / frontier LLM Copilot
│   ├── notifications/           # Slack & MS Teams executive digest & alert bots
│   └── storage/                 # Multi-cluster time-series store + PostgreSQL 18 DDL
├── deploy/
│   └── docker-compose.yaml      # Local stack (pgvector:pg18 + hub-api + hub-worker)
├── docs/
│   └── architecture/
│       └── design-philosophy.md # October 2026 Tactile Editorial & Precision UI spec
├── DESIGN_SYSTEM.md             # Canonical UI/UX Philosophy & 2-Step Design Prompt
├── AGENTS.md                    # Multi-agent engineering & design invariants
└── CHANGELOG.md                 # Semantic versioning release history
```

---

## 🔌 REST API Endpoints (`v1`)

| Method | Endpoint | Description |
| :--- | :--- | :--- |
| `GET` | `/healthz`, `/ready` | Liveness & readiness probes |
| `POST` | `/api/v1/clusters/:cluster_id/snapshots` | Non-blocking (`202 Accepted`) `AuditReport` v1 ingestion |
| `GET` | `/api/v1/executive/kpis` | Fleet-wide FinOps KPIs (Spend, Waste, Realized Savings, Alerts) |
| `GET` | `/api/v1/clusters` | Multi-cluster overview and efficiency scores |
| `GET` | `/api/v1/clusters/:cluster_id/attribution` | $T_0 \to T_1$ `ConfigHash` ROI & SLA Regression verification |
| `GET` | `/api/v1/clusters/:cluster_id/anomalies` | Sawtooth Memory Leaks (`LEAK_SUSPECTED`) & Idle GPU alerts |
| `POST` | `/api/v1/gitops/pr` | 1-Click Kustomize v5 Patch & GitHub Pull Request synthesizer |
| `POST` | `/api/v1/copilot/chat` | Grounded FinOps RAG & Copilot Q&A (`Qwen3.5-9B-GGUF:Q8_0`) |
| `POST` | `/api/v1/pricing/resolve` | Resolve dynamic cloud provider & per-VM SKU pricing for in-cluster agents |
| `GET` | `/api/v1/pricing/catalog` | Query cloud provider & VM instance pricing catalog (`aws`, `gcp`, `azure`, `on-prem`) |

---

## 🎨 UI/UX Design Standard (October 2026)

All frontend interfaces in `kubetrim-hub` follow the **Tactile Editorial & Precision UI** philosophy documented in [`DESIGN_SYSTEM.md`](./DESIGN_SYSTEM.md) and [`AGENTS.md`](./AGENTS.md):
* **Zero AI/SaaS Clichés:** No indigo/purple gradients, no 3-column centered icon cards, no blurry drop shadows.
* **Typography Trio:** Editorial Serif (`Instrument Serif` / `Newsreader`) + Geometric Sans (`Plus Jakarta Sans` / `Inter`) + Technical Monospace (`JetBrains Mono`).
* **Swiss Blueprint Grid:** Exposed `1px solid` structural hairlines, SVG `fractalNoise` texture (`0.035` opacity), and mechanical `cubic-bezier(0.22, 1, 0.36, 1)` micro-interactions.
