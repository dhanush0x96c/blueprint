# Blueprint Template Naming Convention

This document defines the official naming convention for all Blueprint templates. The goal is to ensure consistency,
predictability, discoverability, and long‑term scalability across builtin and community templates.

## Table of Contents

- [Core Principles](#core-principles)
- [General Format](#general-format)
- [Archetypes](#archetypes)
  - [Supported Archetypes](#supported-archetypes)
- [Variants](#variants)
  - [Subvariants](#subvariants)
- [Language Positioning](#language-positioning)
- [When to Omit Variants](#when-to-omit-variants)
- [Components and Features](#components-and-features)
- [Reserved Patterns](#reserved-patterns)
- [Stability Rules](#stability-rules)
- [Examples (Canonical v1 Set)](#examples-canonical-v1-set)
- [Design Philosophy](#design-philosophy)

---

## Core Principles

1. **Names must be predictable** – Users should be able to guess a template name.
2. **Names must be composable** – Variants should follow a clear structural pattern.
3. **Names must be CLI-friendly** – Short, lowercase, kebab-case only.
4. **Names must reflect architecture, not marketing terms.**

---

## General Format

Templates use kebab-case and follow this structure:

```
<language>-<archetype>[-<variant>]
```

Where:

- `language` – Primary implementation language (go, node, python, etc.)
- `archetype` – Architectural category (cli, api, lib, worker, etc.)
- `variant` – Optional framework or specialization (express, fastapi, grpc, etc.)

Examples:

```
go-cli
go-api
go-api-grpc
node-api-express
python-api-fastapi
```

---

## Archetypes

Archetypes describe architectural intent. They must remain stable and limited in number.

### Supported Archetypes

- `cli` – Command-line application
- `web` – Web Application
- `api` – HTTP API service (REST or similar)
- `lib` – Reusable library (no executable entrypoint)
- `worker` – Background job processor
- `service` – Long-running service (non-HTTP)

New archetypes should be introduced sparingly.

---

## Variants

Variants specify frameworks or implementation details.

Examples:

- `node-api-express`
- `node-api-fastify`
- `python-api-fastapi`
- `go-api-grpc`

Variants should not duplicate archetype meaning.

Incorrect:

```
node-web-api-express
```

Correct:

```
node-api-express
```

---

### Subvariants

Subvariants are additional refinement segments appended after a variant.

```
python-api-fastapi-sqlalchemy
```

Ordering must follow general → specific:

- `fastapi` → framework
- `sqlalchemy` → persistence integration

Each segment must narrow the implementation of the previous one. Reversing the order is not allowed.

## Language Positioning

Language always appears first.

Correct:

```
go-api
node-api-express
```

Incorrect:

```
api-go
express-node
```

---

## When to Omit Variants

If a language has a clear default implementation, the variant may be omitted.

Example:

```
go-api
```

This implies a canonical HTTP stack defined by Blueprint.

If multiple frameworks are supported, the variant must be explicit:

```
node-api-express
node-api-fastify
```

---

## Components and Features

Features and components follow the same naming rules when language-specific:

```
go-database-postgres
go-logging
node-auth-jwt
```

Language-agnostic components omit the language prefix:

```
docker
ci-github-actions
license-mit
```

---

## Reserved Patterns

The following patterns are not allowed:

- CamelCase names
- Underscores
- Framework-first naming (e.g., `express-api-node`)
- Redundant segments (e.g., `node-web-api-express`)

---

## Stability Rules

- Template names are part of the public API.
- Renaming a template is a breaking change.
- Archetype vocabulary must remain small and stable.

---

## Examples (Canonical v1 Set)

```
go-cli
go-api
node-api-express
python-api-fastapi
```

These serve as the baseline pattern for future additions.

---

## Design Philosophy

Blueprint prioritizes:

- Architectural clarity over brevity
- Consistency over clever inference
- Long-term scalability over short-term convenience

Template names should feel mechanical and predictable, not creative or descriptive.

If a name requires explanation, it is likely too complex.
