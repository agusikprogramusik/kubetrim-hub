# Changelog

All notable changes to the **kubetrim-hub** project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [0.2.0-dev] - 2026-10-22

### Added
- **Multi-Cluster Time-Series Storage Engine (`pkg/storage`, `pkg/domain`):** Implemented thread-safe time-series snapshot store, `ExecutiveKPIs` & `ClusterOverview` aggregators, and PostgreSQL 18 + TimescaleDB + `pgvector(384)` + RLS DDL schema (`PostgresSchemaSQL`).
- **Hub Ingestion & REST API (`pkg/api`, `apps/api`):** Added Gin v1.12+ HTTP server with API Key authentication (`Authorization: Bearer` / `X-API-Key`), non-blocking `202 Accepted` snapshot ingestion (`POST /api/v1/clusters/:cluster_id/snapshots`), executive KPIs (`GET /api/v1/executive/kpis`), attribution (`GET /api/v1/clusters/:cluster_id/attribution`), anomalies (`GET /api/v1/clusters/:cluster_id/anomalies`), GitOps PR endpoints (`POST /api/v1/gitops/pr`), and FinOps Copilot (`POST /api/v1/copilot/chat`).
- **Before-and-After Attribution & SLA Regression Guard (`pkg/diff`):** Implemented $T_0 \to T_1$ `ConfigHash` ROI attribution (`VERIFIED_SAVINGS`, `REGRESSION_ALERT`, `UNCHANGED`) with automatic rollback recommendation when OOMKills or CPU throttling (>15% delta) spike.
- **Sawtooth Memory Leak & Idle GPU Detector (`pkg/anomaly`, `apps/worker`):** Added 14-day monotonic memory slope + OOM reset sawtooth detector (`LEAK_SUSPECTED` with `BlockMemoryUpsizing=true`) and Idle GPU detector (`IDLE_GPU` below 20% utilization).
- **1-Click GitOps Pull Request Automation (`pkg/gitops`):** Added Kustomize v5 strategic-merge patch synthesizer, Executive Financial Impact PR description generator, and GitHub REST API branch/PR automation.
- **Slack & MS Teams Executive Digest & Alert Bots (`pkg/notifications`):** Implemented webhook dispatcher for Weekly Executive FinOps Digests, `IDLE_GPU` alerts, and `REGRESSION_ALERT` notifications.
- **FinOps RAG & Copilot Engine (`pkg/copilot`):** Implemented grounded natural-language telemetry Q&A with citations, supporting local `llama-server` (`http://localhost:8080/v1`, `unsloth/Qwen3.5-9B-GGUF:Q8_0`) and deterministic fallback.
- **Containerized Stack (`deploy/docker-compose.yaml`):** Added Dockerfiles for `hub-api` and `hub-worker` plus `docker-compose.yaml` with `pgvector/pgvector:pg18`.

## [0.1.0-dev] - 2026-10-22

### Added
- Initialized `kubetrim-hub` repository for the KubeTrim Enterprise FinOps Control Plane (Sprint 2).
- Added `DESIGN_SYSTEM.md`, `AGENTS.md`, and `docs/architecture/design-philosophy.md` codifying the **October 2026 Tactile Editorial & Precision UI** design system, strict Anti-AI Slop constraints, color palettes (`Obsidian Industrial Dark` & `Warm Analog Light`), and the 2-step Visual Architecture $\to$ Implementation workflow.
