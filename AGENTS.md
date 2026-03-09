# MYNAH Agent Working Guide

This file is the working contract for agents contributing to MYNAH after the March 9, 2026 repository reset.
The repo is intentionally reduced to planning documents only. All implementation work starts from `SPEC.md`.

## 1. Project Intent
- Build an open-source, offline-first personal intelligence system.
- Keep the implementation minimal, inspectable, and auditable.
- Preserve user trust through local-first architecture, strong defaults, and deterministic behavior.

## 2. Current Repository State
- The repository reset intentionally removed prior code, scripts, tests, and generated artifacts.
- Only `AGENTS.md` and `SPEC.md` are retained as the planning baseline.
- `SPEC.md` is the source of truth for architecture, scope, and rewrite decisions.

## 3. Working Rules
- Start from the smallest viable design.
- Prefer explicit interfaces and direct implementations over abstraction-heavy designs.
- Avoid fallback implementations unless explicitly requested.
- Keep dependency count low and justify each new dependency.
- Treat destructive operations as high-risk and confirm them unless the user has already approved them.

## 4. Documentation Rules
- Update `SPEC.md` in the same change set as any architecture or behavior decision.
- Remove stale guidance rather than layering contradictory notes on top.
- Record unresolved questions explicitly instead of hiding them in code assumptions.

## 5. Platform Direction
- Runtime target: Linux.
- Development hosts: Linux and Windows.
- Core system must remain local-first.
- Local models through Ollama are the default model strategy unless `SPEC.md` is updated.

## 6. Quality Expectations
- Prefer deterministic processing, explicit validation, and fail-closed behavior.
- Favor simple testing and migration paths from the start.
- Keep security boundaries, storage ownership, and projection/export responsibilities explicit.

## 7. How to Use This File
- Keep entries concrete and actionable.
- Update this file when stable project workflow expectations change.
- Use `SPEC.md` for product and technical design; use `AGENTS.md` for contribution rules.
