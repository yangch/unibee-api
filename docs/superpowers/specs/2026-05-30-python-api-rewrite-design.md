# Python API Rewrite Design

## Purpose

Rewrite the UniBee Go API as a Python codebase while preserving the public API contract. The first milestone is API skeleton parity, not business-logic parity: every public endpoint should exist in FastAPI with matching request and response schemas, route metadata, and controlled stub behavior.

## Decisions

- Rewrite type: full Python rewrite.
- Compatibility contract: API-compatible only.
- Python stack: FastAPI, Pydantic, SQLAlchemy-ready architecture.
- Business logic strategy: manual domain rewrite.
- First production milestone: public API skeleton parity.

## Architecture

The Python service should live beside the Go codebase at first, under a top-level `python-api/` directory or a separate repository later.

Proposed structure:

```text
python-api/
  app/
    main.py
    core/
    api/
    schemas/
    services/
    repositories/
    integrations/
    workers/
  tests/
    contract/
    unit/
    integration/
```

Responsibilities:

- `core/`: settings, logging, errors, auth context, startup, OpenAPI customization.
- `api/`: FastAPI routers grouped like the existing API/controller domains.
- `schemas/`: Pydantic request and response models matching the Go API structs.
- `services/`: domain service interfaces and manual rewrites.
- `repositories/`: persistence abstractions, initially thin or stubbed.
- `integrations/`: Stripe, PayPal, SendGrid, VAT, Nacos/config compatibility adapters.
- `workers/`: later Redis stream consumers and cron/batch jobs.

Milestone 1 should make `api/` and `schemas/` real. Services, repositories, integrations, and workers can mostly be explicit stubs with documented boundaries. Endpoints should return controlled compatibility stubs until their domains are implemented.

## API Contract And Data Flow

The source of truth for milestone 1 is the existing Go API surface:

- API definitions under `api/`.
- Controller routing behavior under `internal/controller/`.
- Current OpenAPI JSON from the running Go service, when available.
- Existing request, response, and error envelope conventions.

Build flow:

```text
Go API definitions / OpenAPI
  -> Python schema inventory
  -> Pydantic models
  -> FastAPI routers
  -> Contract tests compare Python OpenAPI to Go OpenAPI
```

Runtime flow:

```text
request
  -> FastAPI route
  -> auth/context dependency
  -> Pydantic request model
  -> domain service method
  -> response model / error envelope
```

For milestone 1, service methods can raise a standard not-implemented application error. The route, request parsing, auth requirement, response model, tags, and OpenAPI metadata should already match the Go API as closely as possible.

## Components And Migration Order

1. Contract inventory: extract paths, methods, request structs, response structs, tags, auth expectations, and error envelope conventions from the Go project and OpenAPI baseline.
2. FastAPI foundation: app startup, settings, logging, error handling, request context, auth dependency stubs, health checks, and OpenAPI customization.
3. Schema and router parity: create Pydantic models and routers for all public endpoints. Handlers call service interfaces; most services are initially stubs.
4. Contract test suite: compare Python OpenAPI against the Go OpenAPI baseline and smoke test route loading.
5. Domain rewrites: port business logic manually by domain.
6. Operational parity: add real repositories, Redis streams, cron jobs, Nacos/config compatibility, deployment manifests, and observability after the API skeleton is stable.

Recommended domain rewrite order:

```text
auth/session
  -> merchant/config
  -> product/plan
  -> subscription
  -> invoice
  -> payment/gateway/webhook
  -> credit
  -> metrics/reporting
  -> batch/export/workers
```

## Error Handling And Compatibility

Define one application error model early and use it everywhere, including stubs.

The rewrite should preserve:

- HTTP status codes used by the Go API.
- Response envelope shape.
- Validation error behavior where clients depend on it.
- Auth and session failure behavior.
- Webhook failure and retry semantics.
- Idempotency behavior for payment and subscription actions.

Python error flow:

```text
domain/service error
  -> typed AppError
  -> FastAPI exception handler
  -> Go-compatible JSON response
```

Unimplemented endpoints should return deliberate compatibility responses, probably a consistent `501 Not Implemented` envelope, instead of framework-default crashes.

Payment and webhook domains require stricter error compatibility later because vendors may depend on exact status behavior for retries.

## Testing And Success Criteria

Milestone 1 tests:

- OpenAPI diff test comparing Python OpenAPI to a saved Go OpenAPI baseline, with an allowlist for intentional differences.
- Route smoke tests proving every route accepts syntactically valid requests and returns a controlled response.
- Schema round-trip tests for Go-compatible field names and optional/null behavior.
- Error envelope tests for validation, auth stub failures, and `501` stubs.
- Import/startup test proving the FastAPI app boots without external services.

Milestone 1 is successful when:

- All public Go API endpoints are represented in FastAPI.
- Generated Python OpenAPI is reviewed against the Go baseline.
- Endpoint grouping and tags are understandable.
- No endpoint contains real business logic except health/docs behavior.
- Stubs are explicit, searchable, and tracked by domain.
- The next implementation plan can assign domains one by one.

## Non-Goals For Milestone 1

- Reusing or migrating the exact database layer.
- Rewriting all business logic.
- Replacing Redis stream consumers or cron jobs.
- Matching GoFrame internals.
- Shipping the Python service as the production runtime.
