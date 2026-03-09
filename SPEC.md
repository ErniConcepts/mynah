# MYNAH Specification

Version: 0.0-reset-baseline
Status: Active
Last updated: 2026-03-09

## 1. Purpose
MYNAH is being restarted from ground zero.

This document preserves the state of the repository before the reset so the next architecture can be designed intentionally rather than from memory. It is both:
- a record of what existed,
- and a basis for deciding what the new system should keep, discard, or redesign.

## 2. Reset Decision
On March 9, 2026, the repository was intentionally reduced to planning documents only.

Reason:
- the prior repo had accumulated multiple overlapping architecture directions,
- the current implementation mixed unrelated concerns in large services and experimental branches,
- and the project needs a fresh foundation built around a clearer three-pillar model.

After the reset, only these files remain:
- `AGENTS.md`
- `SPEC.md`

## 3. Current Rewrite Direction
The intended future architecture is centered on three pillars:

1. Orchestration pillar
   - Uses a local LLM.
   - Ingests source data.
   - Validates and transforms data.
   - Persists structured results into the database pillar.
   - Produces deterministic exports into the `/me` git pillar.

2. Database pillar
   - Holds canonical system state.
   - Stores ingest metadata, structured memory data, audit trails, and other internal records.

3. `/me` git pillar
   - Materializes human-readable output from canonical database state.
   - Acts as an inspectable personal data projection, not as the primary source of truth.

These pillars are directional goals only. The concrete language, schema, API surface, and deployment shape are still open for design.

## 4. What Existed Before Reset

### 4.1 Project Goal
The pre-reset project described MYNAH as an offline-first personal intelligence system.

### 4.2 Major Branches Present
Before cleanup, the repo had these local branches:
- `main`
- `experiment/postgres-write-plan-pipeline`
- `experiment/memory-e2e-datasets-and-writeplan`
- `feature/zip-ingest-me-repo`

The active worktree at reset time was on:
- `feature/zip-ingest-me-repo`

### 4.3 Documented Architecture Before Reset
The previous docs described a system that:
- ingested GPT export zip files,
- chunked conversations with wider context windows,
- extracted structured folder-aligned JSON with a local Ollama model,
- stored outputs in relational tables,
- and wrote run outputs into a `/ME` git repository.

### 4.4 Human Data Taxonomy in Prior Design
The branch documentation organized data under:
- `human/perception`
- `human/memory`
- `human/decision`
- `human/self_model`
- `human/meta`

With subfolders including:
- `sensory_streams`
- `body_state`
- `episodic`
- `semantic`
- `procedural`
- `emotional`
- `working`
- `goals`
- `policies`
- `simulations`
- `reward_history`
- `identity`
- `traits`
- `preferences`
- `social_models`
- `attention`
- `confidence`
- `uncertainty`

This taxonomy may still be useful, but it should be re-evaluated during the rewrite instead of being treated as automatically correct.

## 5. State of the Previous Implementation

### 5.1 Strengths Worth Remembering
The old repo had several ideas worth preserving conceptually:
- local-first operation,
- Ollama-based structured extraction,
- deterministic `/me` git exports,
- auditable ingest runs and chunk-level records,
- and a realistic memory E2E dataset/testing effort.

### 5.2 Problems With the Previous State
The previous repo had several structural issues:
- architecture drift between documents and code,
- multiple experimental directions in the same repository history,
- a large monolithic service file that mixed HTTP routes, orchestration, model calls, storage writes, and git projection logic,
- unclear boundaries between canonical storage and exported artifacts,
- and accumulated unrelated artifacts and generated files inside the repo.

### 5.3 Explicit Mismatch Observed at Reset Time
A critical inconsistency existed in the written architecture:
- `AGENTS.md` specified SQLite with direct SQL and no ORM for v0.x.
- `spec.md` and the branch implementation were centered on PostgreSQL.

This mismatch is one of the main reasons the rewrite should begin from a clean specification.

## 6. State of the Memory E2E Dataset Branch
The `experiment/memory-e2e-datasets-and-writeplan` branch represented the clearest snapshot of the last major direction.

It included:
- a dataset generation harness,
- an ingest-and-report harness,
- partial reporting on realistic transcript and Codex-history ingestion,
- and a branch direction focused on structured memory extraction and `/ME` outputs.

However, it was not a clean rewrite base because it also sat on top of:
- earlier retrieval and pipeline experiments,
- other runtime assumptions,
- and unrelated branch baggage.

Conclusion:
- the ideas from that branch are worth mining,
- but the code itself should not be treated as the new foundation.

## 7. Pre-Reset Runtime Snapshot
The branch-level runtime direction immediately before reset can be summarized as:
- API server in Python with FastAPI,
- local Ollama model calls for structured extraction,
- relational schema for ingest runs, chunks, and per-folder outputs,
- `/ME` repository bootstrap and commit logic,
- and a test harness for ingesting transcripts, Codex history, and health data.

This is a historical snapshot only, not a commitment for the rewrite.

## 8. Language Direction
The next implementation language has not been finalized.

Current preference under discussion:
- Go as the strongest candidate for the core rewrite,
- TypeScript as a possible language for UI or tooling,
- Rust as a possible later choice only if there is a proven need for lower-level performance or stronger compile-time guarantees in a critical subsystem.

Rationale for Go as the leading candidate:
- simple deployable binaries,
- good fit for local services and pipelines,
- straightforward concurrency,
- direct SQLite support,
- and lower implementation overhead than Rust for a ground-up v0 reset.

This remains a design decision, not a locked requirement.

## 9. Constraints to Revisit During Rewrite
The next design pass should explicitly decide, not assume, the following:
- canonical storage engine,
- runtime language,
- API shape,
- whether the `/me` repo is a projection or a primary record,
- data model granularity,
- ingest source types for v0,
- and how much of the prior human-folder taxonomy is actually necessary.

## 10. Requirements That Still Seem Valid
These principles still appear aligned with the project intent:
- offline-first by default,
- local-only core paths,
- least privilege,
- deterministic and auditable processing,
- explicit failure over silent fallback,
- minimal dependency footprint,
- and inspectable user-owned outputs.

## 11. Open Questions for the New Design
- What is the canonical v0 data model?
- Which sources are truly in scope for v0 ingest?
- What should the first `/me` projection format look like?
- Should the system be a single binary/service or a small set of tightly scoped components?
- What is the minimum viable API surface?
- How much LLM involvement should exist in the first version?
- What should be stored as raw evidence versus derived memory?

## 12. Immediate Next Step
Use this document to design the new MYNAH from scratch.

The next meaningful change should answer:
- the language choice,
- the three-pillar boundaries,
- and the minimum viable v0 architecture.
