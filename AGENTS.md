# KubeTrim Hub — AI Agent Governance & Design Invariants

Every AI agent working within `kubetrim-hub` MUST strictly enforce both the **2026 Technical Stack Standards** and the **Avant-Garde Tactile Editorial & Precision UI Mandate** defined below.

---

## 1. Frontend & Design System Mandate (October 2026)

### 1.1 Role & Design Philosophy
Act as an Avant-Garde Staff Product Designer and Creative Technologist (Awwwards / Linear / Teenage Engineering / Stripe Press caliber). Design and code a high-craft UI that feels unmistakably human-designed in late 2026, completely devoid of generic "AI-generated template" aesthetics.

### 1.2 Strict Negative Constraints ("Anti-AI Slop" Rules)
1. **NO generic SaaS gradients** (strictly ban indigo-to-purple, blue-to-cyan, or neon glowing blobs).
2. **NO symmetrical 3-column feature cards** with a centered icon in a colored circle.
3. **NO centered Hero section** with a generic pill-shaped badge above an H1 and two centered rounded-full buttons.
4. **NO default Tailwind drop shadows** (`shadow-md`, `shadow-lg`, `shadow-xl`) on cards.
5. **NO pure `#FFFFFF` or `#000000` backgrounds**, and **NO standard Tailwind `blue-500` or `indigo-600` primary accents**.
6. **NO generic placeholder text** like "Supercharge your workflow" or "Unlock the power of...". Write realistic, sharp, domain-specific Kubernetes FinOps microcopy (e.g., `p95 CPU headroom +25%`, `QoS Guaranteed oom_score_adj=-997`, `HPA anti-flapping lock`, `T0 -> T1 ConfigHash delta`).

### 1.3 2026 Visual & Technical Direction ("Tactile Editorial & Precision UI")
* **Typography-First Hierarchy:** Pair 3 distinct Google Fonts:
  * High-contrast Editorial Serif (`Instrument Serif` or `Newsreader`) for dramatic headlines and executive callouts.
  * Crisp geometric Sans (`Plus Jakarta Sans` or `Inter`) for dense operational UI.
  * Monospace (`JetBrains Mono`) for indices (`01 //`), SHA-256 `ConfigHash` fingerprints, PromQL metrics, timestamps, and USD waste figures.
  * Use fluid display typography (`clamp()`) and tight heading tracking (`-0.03em`).
* **Layout & Spatial Architecture:** Asymmetric Swiss-style grid / asymmetric Bento architecture with exposed structural `1px solid` hairlines (`#22242A` in Dark / `#E6E4DD` in Light) and intentional negative space.
* **Tactile Materiality:**
  * Fixed SVG `fractalNoise` overlay (`opacity: 0.035`, `pointer-events: none`) across the viewport.
  * Ultra-thin `1px` borders with subtle top-edge inner highlights (`box-shadow: inset 0 1px 0 0 rgba(255,255,255,0.1)`) instead of blurry outer drop shadows.
* **Approved Color Palettes:**
  * **OBSIDIAN INDUSTRIAL DARK:** Deep carbon `#0B0C0E` background, elevated surface `#121316`, hairline border `#22242A`, crisp off-white `#ECECEE` text, and raw amber `#F59E0B` or acid lime `#D4FF00` accent.
  * **WARM ANALOG LIGHT:** Warm alabaster/paper `#F6F5F0` background, deep charcoal `#141413` ink, muted stone `#E6E4DD` borders, and international orange `#FF4F00` or forest green `#1B4332` accent.
* **Micro-interactions:** Custom `cubic-bezier(0.22, 1, 0.36, 1)` transitions with crisp mechanical responses (border brightening, monospace `[ -> ]` shift, background fill wipe).
* **Two-Step Execution Protocol:**
  * **Step 1 (Visual Architecture):** Propose the exact design system (HEX palette, font trio, structural grid, and 3 bespoke interaction details) before writing UI code.
  * **Step 2 (Implementation):** Implement the approved design with full CSS, typography, and micro-interaction fidelity.

---

## 2. Technical Stack & Engineering Invariants

* **Backend (`apps/api`, `apps/worker`, `pkg/`):** Go 1.27.1+, Gin v1.12+, Clean Architecture, TDD with `-race`.
* **Frontend (`apps/web/`):** Node 26 LTS, Next.js 16 (App Router), React 19, TypeScript 5.9+, TailwindCSS v4+.
* **Database:** PostgreSQL 18+ with `TimescaleDB` (hypertables for snapshots) and `pgvector` (384-dim `BAAI/bge-small-en-v1.5` embeddings for RAG).
* **Frontier AI & Local Sovereign LLM:** Support modern 2026 frontier models (`gemini-3.8-flash`, `claude-sonnet-5`, `gpt-5.6`, `deepseek-v4-pro`) and local `llama-server` (`Qwen3.5-9B-GGUF:Q8_0`).
* **Language & Git Workflow:** Strictly English in all code, commits, PRs, and documentation. Branching via `feat/...`, `fix/...`, `chore/...` with Pull Requests and SemVer updates in `CHANGELOG.md`.
