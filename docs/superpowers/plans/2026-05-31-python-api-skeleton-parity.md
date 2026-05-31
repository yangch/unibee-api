# Python API Skeleton Parity Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Build the first Python/FastAPI milestone for UniBee: a side-by-side `python-api/` service with Go API route inventory, generated schema/route stubs, compatibility error envelopes, and OpenAPI contract checks.

**Architecture:** Keep the Python rewrite beside the Go service under `python-api/`. Use small handwritten FastAPI foundation modules, then generate the large repetitive API skeleton from Go `api/**` request/response structs and `g.Meta` metadata. All generated endpoints return explicit compatibility stubs until business logic is rewritten by domain.

**Tech Stack:** Python 3.11, FastAPI, Pydantic v2, pytest, httpx, ruff, mypy, standard-library Go source parsing with regex-backed inventory extraction.

---

## Source Spec

Use [docs/superpowers/specs/2026-05-30-python-api-rewrite-design.md](/home/yangch/unibee-api/docs/superpowers/specs/2026-05-30-python-api-rewrite-design.md) as the approved design.

## File Structure

- Create `python-api/pyproject.toml`: Python package metadata, runtime dependencies, test/lint/type-check commands.
- Create `python-api/README.md`: local development commands and milestone behavior.
- Create `python-api/app/main.py`: FastAPI application factory, router wiring, OpenAPI metadata.
- Create `python-api/app/core/config.py`: settings object with environment-driven service metadata.
- Create `python-api/app/core/errors.py`: Go-compatible JSON envelope, `AppError`, and exception handlers.
- Create `python-api/app/core/context.py`: request ID dependency and request-scoped helpers.
- Create `python-api/app/core/auth.py`: explicit auth-context stubs for protected route groups.
- Create `python-api/app/api/health.py`: health/version endpoints with real responses.
- Create `python-api/app/api/generated.py`: generated route stubs from the Go inventory.
- Create `python-api/app/schemas/generated.py`: generated Pydantic request/response models from Go structs.
- Create `python-api/tools/extract_go_contract.py`: inventory extractor for Go `g.Meta` route metadata and struct fields.
- Create `python-api/tools/generate_stub_api.py`: deterministic generator for `app/api/generated.py` and `app/schemas/generated.py`.
- Create `python-api/contracts/go_route_inventory.json`: generated inventory committed for review.
- Create `python-api/contracts/openapi_allowlist.json`: intentional OpenAPI differences.
- Create `python-api/tests/`: startup, envelope, extractor, generator, route smoke, and OpenAPI contract tests.

## Task 1: Python Project Skeleton

**Files:**
- Create: `python-api/pyproject.toml`
- Create: `python-api/README.md`
- Create: `python-api/app/__init__.py`
- Create: `python-api/app/api/__init__.py`
- Create: `python-api/app/core/__init__.py`
- Create: `python-api/app/schemas/__init__.py`
- Create: `python-api/tests/__init__.py`

- [ ] **Step 1: Write the project metadata**

Create `python-api/pyproject.toml`:

```toml
[project]
name = "unibee-python-api"
version = "0.1.0"
description = "FastAPI skeleton rewrite of the UniBee Go API"
requires-python = ">=3.11"
dependencies = [
  "fastapi>=0.111,<1.0",
  "pydantic>=2.7,<3.0",
  "pydantic-settings>=2.2,<3.0",
  "uvicorn[standard]>=0.29,<1.0"
]

[project.optional-dependencies]
dev = [
  "httpx>=0.27,<1.0",
  "pytest>=8.2,<9.0",
  "ruff>=0.4,<1.0",
  "mypy>=1.10,<2.0"
]

[tool.pytest.ini_options]
testpaths = ["tests"]
pythonpath = ["."]

[tool.ruff]
line-length = 100
target-version = "py311"

[tool.ruff.lint]
select = ["E", "F", "I", "UP", "B"]

[tool.mypy]
python_version = "3.11"
strict = true
plugins = []
```

- [ ] **Step 2: Add local usage notes**

Create `python-api/README.md`:

```markdown
# UniBee Python API

This directory contains the FastAPI rewrite skeleton for the UniBee Go API.

Milestone 1 preserves the public API surface only:

- Health/version endpoints return real responses.
- Generated API endpoints expose the Go paths, methods, tags, summaries, request models, and response models.
- Generated API endpoints return a controlled `501` compatibility envelope until each domain is manually rewritten.

## Commands

```bash
python -m pip install -e ".[dev]"
python tools/extract_go_contract.py --repo-root .. --output contracts/go_route_inventory.json
python tools/generate_stub_api.py --inventory contracts/go_route_inventory.json --api-output app/api/generated.py --schema-output app/schemas/generated.py
pytest -q
uvicorn app.main:app --reload
```
```

- [ ] **Step 3: Add package marker files**

Create empty files:

```text
python-api/app/__init__.py
python-api/app/api/__init__.py
python-api/app/core/__init__.py
python-api/app/schemas/__init__.py
python-api/tests/__init__.py
```

- [ ] **Step 4: Run packaging checks**

Run:

```bash
cd python-api
python -m pip install -e ".[dev]"
python -m pytest -q
```

Expected: dependency installation succeeds and pytest reports `no tests ran`.

- [ ] **Step 5: Commit**

```bash
git add python-api/pyproject.toml python-api/README.md python-api/app python-api/tests
git commit -m "chore(python-api): scaffold FastAPI project"
```

## Task 2: FastAPI Foundation And Compatibility Envelope

**Files:**
- Create: `python-api/app/core/config.py`
- Create: `python-api/app/core/context.py`
- Create: `python-api/app/core/errors.py`
- Create: `python-api/app/core/auth.py`
- Create: `python-api/app/api/health.py`
- Create: `python-api/app/main.py`
- Create: `python-api/tests/test_foundation.py`

- [ ] **Step 1: Write failing foundation tests**

Create `python-api/tests/test_foundation.py`:

```python
from fastapi.testclient import TestClient

from app.main import app


client = TestClient(app)


def test_health_uses_go_compatible_envelope() -> None:
    response = client.get("/health")

    assert response.status_code == 200
    body = response.json()
    assert body["code"] == 0
    assert body["message"] == ""
    assert body["data"] == {"status": "ok"}
    assert isinstance(body["requestId"], str)


def test_version_uses_go_compatible_envelope() -> None:
    response = client.get("/version")

    assert response.status_code == 200
    body = response.json()
    assert body["code"] == 0
    assert body["data"]["service"] == "unibee-python-api"


def test_app_error_is_go_compatible() -> None:
    response = client.get("/__test__/not-implemented")

    assert response.status_code == 501
    assert response.json()["code"] == 501
    assert response.json()["message"] == "Endpoint skeleton is present but domain logic is not implemented"
```

- [ ] **Step 2: Run tests to verify failure**

Run:

```bash
cd python-api
pytest tests/test_foundation.py -q
```

Expected: FAIL because `app.main` does not exist.

- [ ] **Step 3: Implement settings and request context**

Create `python-api/app/core/config.py`:

```python
from functools import lru_cache

from pydantic_settings import BaseSettings, SettingsConfigDict


class Settings(BaseSettings):
    service_name: str = "unibee-python-api"
    version: str = "0.1.0"
    openapi_title: str = "OpenAPI UniBee"
    openapi_description: str = "UniBee Api Server"

    model_config = SettingsConfigDict(env_prefix="UNIBEE_", extra="ignore")


@lru_cache
def get_settings() -> Settings:
    return Settings()
```

Create `python-api/app/core/context.py`:

```python
from uuid import uuid4

from fastapi import Request


REQUEST_ID_HEADER = "X-Request-Id"


def get_request_id(request: Request) -> str:
    incoming = request.headers.get(REQUEST_ID_HEADER)
    if incoming:
        return incoming
    existing = getattr(request.state, "request_id", None)
    if isinstance(existing, str):
        return existing
    request_id = str(uuid4())
    request.state.request_id = request_id
    return request_id
```

- [ ] **Step 4: Implement errors and envelope helpers**

Create `python-api/app/core/errors.py`:

```python
from typing import Any

from fastapi import FastAPI, Request
from fastapi.exceptions import RequestValidationError
from fastapi.responses import JSONResponse
from pydantic import BaseModel, Field

from app.core.context import get_request_id


class JsonEnvelope(BaseModel):
    code: int = 0
    message: str = ""
    data: Any = Field(default_factory=dict)
    redirect: str = ""
    requestId: str = ""


class AppError(Exception):
    def __init__(self, code: int, message: str, status_code: int = 400, data: Any | None = None):
        self.code = code
        self.message = message
        self.status_code = status_code
        self.data = {} if data is None else data


def envelope(request: Request, data: Any | None = None, code: int = 0, message: str = "") -> dict[str, Any]:
    return JsonEnvelope(
        code=code,
        message=message,
        data={} if data is None else data,
        requestId=get_request_id(request),
    ).model_dump()


def register_exception_handlers(app: FastAPI) -> None:
    @app.exception_handler(AppError)
    async def app_error_handler(request: Request, exc: AppError) -> JSONResponse:
        return JSONResponse(
            status_code=exc.status_code,
            content=envelope(request, data=exc.data, code=exc.code, message=exc.message),
        )

    @app.exception_handler(RequestValidationError)
    async def validation_error_handler(request: Request, exc: RequestValidationError) -> JSONResponse:
        return JSONResponse(
            status_code=422,
            content=envelope(request, data={"errors": exc.errors()}, code=51, message="Validation Failed"),
        )
```

- [ ] **Step 5: Implement auth stubs**

Create `python-api/app/core/auth.py`:

```python
from dataclasses import dataclass

from fastapi import Header


@dataclass(frozen=True)
class AuthContext:
    token: str | None
    is_open_api_call: bool


async def optional_auth_context(authorization: str | None = Header(default=None)) -> AuthContext:
    return AuthContext(token=authorization, is_open_api_call=bool(authorization))


async def merchant_auth_context(authorization: str | None = Header(default=None)) -> AuthContext:
    return AuthContext(token=authorization, is_open_api_call=bool(authorization))


async def user_auth_context(authorization: str | None = Header(default=None)) -> AuthContext:
    return AuthContext(token=authorization, is_open_api_call=bool(authorization))
```

- [ ] **Step 6: Implement health routes and app factory**

Create `python-api/app/api/health.py`:

```python
from fastapi import APIRouter, Request

from app.core.config import get_settings
from app.core.errors import envelope

router = APIRouter(tags=["Health"])


@router.get("/health")
async def health(request: Request) -> dict[str, object]:
    return envelope(request, {"status": "ok"})


@router.get("/version")
async def version(request: Request) -> dict[str, object]:
    settings = get_settings()
    return envelope(request, {"service": settings.service_name, "version": settings.version})
```

Create `python-api/app/main.py`:

```python
from fastapi import FastAPI, Request

from app.api import health
from app.core.config import get_settings
from app.core.errors import AppError, register_exception_handlers


def create_app() -> FastAPI:
    settings = get_settings()
    fastapi_app = FastAPI(
        title=settings.openapi_title,
        description=settings.openapi_description,
        version=settings.version,
    )
    register_exception_handlers(fastapi_app)
    fastapi_app.include_router(health.router)

    @fastapi_app.get("/__test__/not-implemented", include_in_schema=False)
    async def test_not_implemented(_: Request) -> None:
        raise AppError(
            code=501,
            message="Endpoint skeleton is present but domain logic is not implemented",
            status_code=501,
        )

    return fastapi_app


app = create_app()
```

- [ ] **Step 7: Run tests to verify pass**

Run:

```bash
cd python-api
pytest tests/test_foundation.py -q
```

Expected: PASS.

- [ ] **Step 8: Commit**

```bash
git add python-api/app python-api/tests/test_foundation.py
git commit -m "feat(python-api): add FastAPI foundation"
```

## Task 3: Go Contract Inventory Extractor

**Files:**
- Create: `python-api/tools/extract_go_contract.py`
- Create: `python-api/tests/test_extract_go_contract.py`
- Create: `python-api/contracts/.gitkeep`

- [ ] **Step 1: Write failing extractor tests**

Create `python-api/tests/test_extract_go_contract.py`:

```python
from pathlib import Path

from tools.extract_go_contract import extract_contract


def test_extracts_meta_and_fields_from_go_api_file(tmp_path: Path) -> None:
    repo = tmp_path
    api_file = repo / "api" / "checkout" / "vat" / "vat.go"
    api_file.parent.mkdir(parents=True)
    api_file.write_text(
        '''
package vat

import "github.com/gogf/gf/v2/frame/g"

type CountryListReq struct {
    g.Meta `path:"/country_list" tags:"Checkout" method:"get,post" summary:"Vat Country List"`
    MerchantId uint64 `json:"merchantId" v:"required"`
}

type CountryListRes struct {
    VatCountryList []string `json:"vatCountryList"`
}
''',
        encoding="utf-8",
    )

    inventory = extract_contract(repo)

    assert inventory["endpoints"][0]["path"] == "/checkout/vat/country_list"
    assert inventory["endpoints"][0]["methods"] == ["GET", "POST"]
    assert inventory["endpoints"][0]["request_model"] == "CheckoutVatCountryListReq"
    assert inventory["models"]["CheckoutVatCountryListReq"]["fields"][0]["json_name"] == "merchantId"
    assert inventory["models"]["CheckoutVatCountryListReq"]["fields"][0]["required"] is True
```

- [ ] **Step 2: Run tests to verify failure**

Run:

```bash
cd python-api
pytest tests/test_extract_go_contract.py -q
```

Expected: FAIL because `tools.extract_go_contract` does not exist.

- [ ] **Step 3: Implement extractor**

Create `python-api/tools/extract_go_contract.py`:

```python
from __future__ import annotations

import argparse
import json
import re
from pathlib import Path
from typing import Any


META_RE = re.compile(r"g\.Meta\s+`(?P<tag>[^`]+)`")
STRUCT_RE = re.compile(r"type\s+(?P<name>\w+)\s+struct\s*\{(?P<body>.*?)\n\}", re.DOTALL)
FIELD_RE = re.compile(r"^\s*(?P<name>\w+)\s+(?P<type>[^\s`]+(?:\[[^\]]+\])?)\s+`(?P<tag>[^`]+)`", re.MULTILINE)
TAG_RE = re.compile(r'(?P<key>\w+):"(?P<value>[^"]*)"')


PREFIX_BY_API_DIR = {
    "api/checkout/checkout": "/checkout/merchant_checkout",
    "api/checkout/gateway": "/checkout/gateway",
    "api/checkout/ip": "/checkout/ip",
    "api/checkout/payment": "/checkout/payment",
    "api/checkout/plan": "/checkout/plan",
    "api/checkout/subscription": "/checkout/subscription",
    "api/checkout/translater": "/checkout",
    "api/checkout/vat": "/checkout/vat",
    "api/merchant/auth": "/merchant/auth",
    "api/merchant/checkout": "/merchant/checkout",
    "api/merchant/credit": "/merchant/credit",
    "api/merchant/discount": "/merchant/discount",
    "api/merchant/email": "/merchant/email",
    "api/merchant/gateway": "/merchant/gateway",
    "api/merchant/integration": "/merchant/integration",
    "api/merchant/invoice": "/merchant/invoice",
    "api/merchant/member": "/merchant/member",
    "api/merchant/merchant": "/merchant",
    "api/merchant/metric": "/merchant/metric",
    "api/merchant/oss": "/merchant/oss",
    "api/merchant/payment": "/merchant/payment",
    "api/merchant/plan": "/merchant/plan",
    "api/merchant/product": "/merchant/product",
    "api/merchant/profile": "/merchant",
    "api/merchant/role": "/merchant/role",
    "api/merchant/search": "/merchant/search",
    "api/merchant/session": "/merchant/session",
    "api/merchant/subscription": "/merchant/subscription",
    "api/merchant/task": "/merchant/task",
    "api/merchant/track": "/merchant/track",
    "api/merchant/user": "/merchant/user",
    "api/merchant/vat": "/merchant/vat",
    "api/merchant/webhook": "/merchant/webhook",
    "api/system/auth": "/system/auth",
    "api/system/information": "/system/information",
    "api/system/invoice": "/system/invoice",
    "api/system/payment": "/system/payment",
    "api/system/plan": "/system/plan",
    "api/system/refund": "/system/refund",
    "api/system/subscription": "/system/subscription",
    "api/system/user": "/system/user",
    "api/user/auth": "/user/auth",
    "api/user/gateway": "/user/gateway",
    "api/user/invoice": "/user/invoice",
    "api/user/merchant": "/user/merchant",
    "api/user/metric": "/user/metric",
    "api/user/payment": "/user/payment",
    "api/user/plan": "/user/plan",
    "api/user/product": "/user/product",
    "api/user/profile": "/user",
    "api/user/subscription": "/user/subscription",
    "api/user/vat": "/user/vat",
}


def parse_tags(raw: str) -> dict[str, str]:
    return {match.group("key"): match.group("value") for match in TAG_RE.finditer(raw)}


def model_name(relative_dir: str, go_name: str) -> str:
    parts = [part.title().replace("_", "") for part in relative_dir.split("/")[1:]]
    return "".join(parts) + go_name


def go_type_to_python(go_type: str) -> str:
    if go_type in {"string"}:
        return "str"
    if go_type in {"int", "int64", "uint", "uint64", "float64", "float32"}:
        return "float" if go_type.startswith("float") else "int"
    if go_type == "bool":
        return "bool"
    if go_type.startswith("[]"):
        return "list"
    if go_type.startswith("map["):
        return "dict"
    return "dict"


def extract_contract(repo_root: Path) -> dict[str, Any]:
    endpoints: list[dict[str, Any]] = []
    models: dict[str, Any] = {}
    api_root = repo_root / "api"

    for path in sorted(api_root.glob("**/*.go")):
        if "/bean/" in path.as_posix():
            continue
        relative_dir = path.parent.relative_to(repo_root).as_posix()
        prefix = PREFIX_BY_API_DIR.get(relative_dir)
        if prefix is None:
            continue
        text = path.read_text(encoding="utf-8")
        structs = {match.group("name"): match.group("body") for match in STRUCT_RE.finditer(text)}
        for go_name, body in structs.items():
            if not go_name.endswith("Req"):
                continue
            meta_match = META_RE.search(body)
            if meta_match is None:
                continue
            meta = parse_tags(meta_match.group("tag"))
            endpoint_path = f"{prefix.rstrip('/')}/{meta['path'].lstrip('/')}"
            request_model = model_name(relative_dir, go_name)
            response_model = model_name(relative_dir, go_name.removesuffix("Req") + "Res")
            endpoints.append(
                {
                    "path": endpoint_path,
                    "methods": [method.strip().upper() for method in meta.get("method", "GET").split(",")],
                    "tags": [meta.get("tags", relative_dir)],
                    "summary": meta.get("summary", ""),
                    "description": meta.get("dc", ""),
                    "request_model": request_model,
                    "response_model": response_model,
                    "source": path.relative_to(repo_root).as_posix(),
                }
            )

        for go_name, body in structs.items():
            fields: list[dict[str, Any]] = []
            for field in FIELD_RE.finditer(body):
                tags = parse_tags(field.group("tag"))
                json_name = tags.get("json", field.group("name"))
                if json_name == "-":
                    continue
                json_name = json_name.split(",")[0]
                fields.append(
                    {
                        "go_name": field.group("name"),
                        "json_name": json_name,
                        "go_type": field.group("type"),
                        "python_type": go_type_to_python(field.group("type")),
                        "required": "required" in tags.get("v", ""),
                        "description": tags.get("dc") or tags.get("description") or "",
                    }
                )
            models[model_name(relative_dir, go_name)] = {"source": path.relative_to(repo_root).as_posix(), "fields": fields}

    return {"endpoints": endpoints, "models": models}


def main() -> None:
    parser = argparse.ArgumentParser()
    parser.add_argument("--repo-root", type=Path, required=True)
    parser.add_argument("--output", type=Path, required=True)
    args = parser.parse_args()

    inventory = extract_contract(args.repo_root.resolve())
    args.output.parent.mkdir(parents=True, exist_ok=True)
    args.output.write_text(json.dumps(inventory, indent=2, sort_keys=True) + "\n", encoding="utf-8")


if __name__ == "__main__":
    main()
```

- [ ] **Step 4: Run extractor tests**

Run:

```bash
cd python-api
pytest tests/test_extract_go_contract.py -q
```

Expected: PASS.

- [ ] **Step 5: Generate and inspect inventory**

Run:

```bash
cd python-api
python tools/extract_go_contract.py --repo-root .. --output contracts/go_route_inventory.json
python -m json.tool contracts/go_route_inventory.json > /tmp/go_route_inventory.pretty.json
```

Expected: `contracts/go_route_inventory.json` exists and contains non-empty `endpoints` and `models`.

- [ ] **Step 6: Commit**

```bash
git add python-api/tools/extract_go_contract.py python-api/tests/test_extract_go_contract.py python-api/contracts
git commit -m "feat(python-api): extract Go API contract inventory"
```

## Task 4: Generated Pydantic Models And Route Stubs

**Files:**
- Create: `python-api/tools/generate_stub_api.py`
- Create: `python-api/tests/test_generate_stub_api.py`
- Generate: `python-api/app/schemas/generated.py`
- Generate: `python-api/app/api/generated.py`
- Modify: `python-api/app/main.py`

- [ ] **Step 1: Write failing generator tests**

Create `python-api/tests/test_generate_stub_api.py`:

```python
import json
from pathlib import Path

from tools.generate_stub_api import generate_files


def test_generates_importable_models_and_routes(tmp_path: Path) -> None:
    inventory = {
        "models": {
            "CheckoutVatCountryListReq": {
                "fields": [
                    {
                        "json_name": "merchantId",
                        "python_type": "int",
                        "required": True,
                        "description": "Merchant ID",
                    }
                ]
            },
            "CheckoutVatCountryListRes": {
                "fields": [
                    {
                        "json_name": "vatCountryList",
                        "python_type": "list",
                        "required": False,
                        "description": "VAT country list",
                    }
                ]
            },
        },
        "endpoints": [
            {
                "path": "/checkout/vat/country_list",
                "methods": ["GET", "POST"],
                "tags": ["Checkout"],
                "summary": "Vat Country List",
                "description": "",
                "request_model": "CheckoutVatCountryListReq",
                "response_model": "CheckoutVatCountryListRes",
                "source": "api/checkout/vat/vat.go",
            }
        ],
    }
    inventory_path = tmp_path / "inventory.json"
    schema_path = tmp_path / "generated_schema.py"
    api_path = tmp_path / "generated_api.py"
    inventory_path.write_text(json.dumps(inventory), encoding="utf-8")

    generate_files(inventory_path, api_path, schema_path)

    assert "class CheckoutVatCountryListReq" in schema_path.read_text(encoding="utf-8")
    assert '@router.api_route("/checkout/vat/country_list"' in api_path.read_text(encoding="utf-8")
    assert "Endpoint skeleton is present but domain logic is not implemented" in api_path.read_text(encoding="utf-8")
```

- [ ] **Step 2: Run tests to verify failure**

Run:

```bash
cd python-api
pytest tests/test_generate_stub_api.py -q
```

Expected: FAIL because `tools.generate_stub_api` does not exist.

- [ ] **Step 3: Implement generator**

Create `python-api/tools/generate_stub_api.py`:

```python
from __future__ import annotations

import argparse
import json
import re
from pathlib import Path
from typing import Any


PYTHON_TYPE_BY_INVENTORY_TYPE = {
    "str": "str",
    "int": "int",
    "float": "float",
    "bool": "bool",
    "list": "list[Any]",
    "dict": "dict[str, Any]",
}


def snake_case(name: str) -> str:
    value = re.sub("(.)([A-Z][a-z]+)", r"\1_\2", name)
    return re.sub("([a-z0-9])([A-Z])", r"\1_\2", value).lower()


def py_literal(value: object) -> str:
    return repr(value)


def generate_schema(inventory: dict[str, Any]) -> str:
    lines = [
        "from __future__ import annotations",
        "",
        "from typing import Any",
        "",
        "from pydantic import BaseModel, ConfigDict, Field",
        "",
        "",
        "class GoCompatibleModel(BaseModel):",
        "    model_config = ConfigDict(populate_by_name=True, extra=\"allow\")",
        "",
    ]
    for model_name, model in sorted(inventory["models"].items()):
        lines.append("")
        lines.append(f"class {model_name}(GoCompatibleModel):")
        fields = model.get("fields", [])
        if not fields:
            lines.append("    pass")
            continue
        used_names: set[str] = set()
        for field in fields:
            json_name = field["json_name"]
            attr_name = snake_case(json_name)
            if attr_name in used_names:
                attr_name = f"{attr_name}_field"
            used_names.add(attr_name)
            py_type = PYTHON_TYPE_BY_INVENTORY_TYPE.get(field["python_type"], "Any")
            default = "..." if field["required"] else "None"
            if default == "None":
                py_type = f"{py_type} | None"
            description = field.get("description", "")
            lines.append(
                f"    {attr_name}: {py_type} = Field({default}, alias={py_literal(json_name)}, description={py_literal(description)})"
            )
    lines.append("")
    return "\n".join(lines)


def generate_api(inventory: dict[str, Any]) -> str:
    lines = [
        "from __future__ import annotations",
        "",
        "from typing import Any",
        "",
        "from fastapi import APIRouter, Body, Request",
        "",
        "from app.core.errors import AppError",
        "from app.schemas import generated as schemas",
        "",
        "",
        "router = APIRouter()",
        "",
        "",
        "def not_implemented_error() -> AppError:",
        "    return AppError(",
        "        code=501,",
        "        message=\"Endpoint skeleton is present but domain logic is not implemented\",",
        "        status_code=501,",
        "    )",
        "",
    ]
    for index, endpoint in enumerate(inventory["endpoints"]):
        function_name = f"stub_{index}_{snake_case(endpoint['request_model'])}"
        methods = endpoint["methods"]
        path = endpoint["path"]
        tags = endpoint.get("tags", [])
        summary = endpoint.get("summary", "")
        description = endpoint.get("description", "")
        request_model = endpoint["request_model"]
        lines.append("")
        lines.append(
            f"@router.api_route({py_literal(path)}, methods={py_literal(methods)}, tags={py_literal(tags)}, summary={py_literal(summary)}, description={py_literal(description)})"
        )
        lines.append(
            f"async def {function_name}(request: Request, payload: schemas.{request_model} | None = Body(default=None)) -> dict[str, Any]:"
        )
        lines.append("    _ = request")
        lines.append("    _ = payload")
        lines.append("    raise not_implemented_error()")
    lines.append("")
    return "\n".join(lines)


def generate_files(inventory_path: Path, api_output: Path, schema_output: Path) -> None:
    inventory = json.loads(inventory_path.read_text(encoding="utf-8"))
    api_output.parent.mkdir(parents=True, exist_ok=True)
    schema_output.parent.mkdir(parents=True, exist_ok=True)
    schema_output.write_text(generate_schema(inventory), encoding="utf-8")
    api_output.write_text(generate_api(inventory), encoding="utf-8")


def main() -> None:
    parser = argparse.ArgumentParser()
    parser.add_argument("--inventory", type=Path, required=True)
    parser.add_argument("--api-output", type=Path, required=True)
    parser.add_argument("--schema-output", type=Path, required=True)
    args = parser.parse_args()
    generate_files(args.inventory, args.api_output, args.schema_output)


if __name__ == "__main__":
    main()
```

- [ ] **Step 4: Run generator tests**

Run:

```bash
cd python-api
pytest tests/test_generate_stub_api.py -q
```

Expected: PASS.

- [ ] **Step 5: Generate route and schema modules**

Run:

```bash
cd python-api
python tools/generate_stub_api.py --inventory contracts/go_route_inventory.json --api-output app/api/generated.py --schema-output app/schemas/generated.py
```

Expected: `app/api/generated.py` and `app/schemas/generated.py` are created.

- [ ] **Step 6: Wire generated router into the app**

Modify `python-api/app/main.py` to import and include generated routes:

```python
from fastapi import FastAPI, Request

from app.api import generated, health
from app.core.config import get_settings
from app.core.errors import AppError, register_exception_handlers


def create_app() -> FastAPI:
    settings = get_settings()
    fastapi_app = FastAPI(
        title=settings.openapi_title,
        description=settings.openapi_description,
        version=settings.version,
    )
    register_exception_handlers(fastapi_app)
    fastapi_app.include_router(health.router)
    fastapi_app.include_router(generated.router)

    @fastapi_app.get("/__test__/not-implemented", include_in_schema=False)
    async def test_not_implemented(_: Request) -> None:
        raise AppError(
            code=501,
            message="Endpoint skeleton is present but domain logic is not implemented",
            status_code=501,
        )

    return fastapi_app


app = create_app()
```

- [ ] **Step 7: Run focused tests**

Run:

```bash
cd python-api
pytest tests/test_generate_stub_api.py tests/test_foundation.py -q
python -m compileall app tools
```

Expected: PASS and compileall succeeds.

- [ ] **Step 8: Commit**

```bash
git add python-api/tools/generate_stub_api.py python-api/tests/test_generate_stub_api.py python-api/app python-api/contracts/go_route_inventory.json
git commit -m "feat(python-api): generate API skeleton stubs"
```

## Task 5: Route Smoke And OpenAPI Contract Tests

**Files:**
- Create: `python-api/contracts/openapi_allowlist.json`
- Create: `python-api/tests/test_route_smoke.py`
- Create: `python-api/tests/test_openapi_contract.py`

- [ ] **Step 1: Add OpenAPI allowlist**

Create `python-api/contracts/openapi_allowlist.json`:

```json
{
  "missing_from_python": [],
  "extra_in_python": [
    "GET /health",
    "GET /version"
  ],
  "notes": [
    "Generated stubs intentionally return 501 until each domain is manually rewritten.",
    "Health and version are real Python service endpoints."
  ]
}
```

- [ ] **Step 2: Write route smoke tests**

Create `python-api/tests/test_route_smoke.py`:

```python
import json
from pathlib import Path

from fastapi.testclient import TestClient

from app.main import app


client = TestClient(app)


def test_generated_routes_return_controlled_stub_response() -> None:
    inventory = json.loads(Path("contracts/go_route_inventory.json").read_text(encoding="utf-8"))
    endpoint = inventory["endpoints"][0]
    method = endpoint["methods"][0]

    response = client.request(method, endpoint["path"], json={})

    assert response.status_code in {422, 501}
    if response.status_code == 501:
        body = response.json()
        assert body["code"] == 501
        assert body["message"] == "Endpoint skeleton is present but domain logic is not implemented"
        assert isinstance(body["requestId"], str)


def test_every_inventory_route_is_registered() -> None:
    inventory = json.loads(Path("contracts/go_route_inventory.json").read_text(encoding="utf-8"))
    registered = {
        (method, route.path)
        for route in app.routes
        for method in getattr(route, "methods", set())
        if route.include_in_schema
    }

    missing = []
    for endpoint in inventory["endpoints"]:
        for method in endpoint["methods"]:
            if (method, endpoint["path"]) not in registered:
                missing.append(f"{method} {endpoint['path']}")

    assert missing == []
```

- [ ] **Step 3: Write OpenAPI contract tests**

Create `python-api/tests/test_openapi_contract.py`:

```python
import json
from pathlib import Path

from app.main import app


def test_openapi_contains_every_inventory_operation() -> None:
    inventory = json.loads(Path("contracts/go_route_inventory.json").read_text(encoding="utf-8"))
    schema = app.openapi()

    missing = []
    for endpoint in inventory["endpoints"]:
        path_item = schema["paths"].get(endpoint["path"])
        if path_item is None:
            missing.append(endpoint["path"])
            continue
        for method in endpoint["methods"]:
            if method.lower() not in path_item:
                missing.append(f"{method} {endpoint['path']}")

    assert missing == []


def test_openapi_has_go_style_bearer_security_scheme() -> None:
    schema = app.openapi()

    security_schemes = schema["components"]["securitySchemes"]
    assert security_schemes["Authorization"]["type"] == "http"
    assert security_schemes["Authorization"]["scheme"] == "bearer"
```

- [ ] **Step 4: Run tests to verify security failure**

Run:

```bash
cd python-api
pytest tests/test_route_smoke.py tests/test_openapi_contract.py -q
```

Expected: route registration test passes, security scheme test fails because OpenAPI security has not been customized.

- [ ] **Step 5: Add OpenAPI security customization**

Modify `python-api/app/main.py`:

```python
from fastapi import FastAPI, Request

from app.api import generated, health
from app.core.config import get_settings
from app.core.errors import AppError, register_exception_handlers


def install_openapi_security(fastapi_app: FastAPI) -> None:
    original_openapi = fastapi_app.openapi

    def custom_openapi() -> dict[str, object]:
        schema = original_openapi()
        components = schema.setdefault("components", {})
        security_schemes = components.setdefault("securitySchemes", {})
        security_schemes["Authorization"] = {
            "type": "http",
            "scheme": "bearer",
            "bearerFormat": "JWT",
        }
        schema["security"] = [{"Authorization": []}]
        return schema

    fastapi_app.openapi = custom_openapi


def create_app() -> FastAPI:
    settings = get_settings()
    fastapi_app = FastAPI(
        title=settings.openapi_title,
        description=settings.openapi_description,
        version=settings.version,
    )
    register_exception_handlers(fastapi_app)
    install_openapi_security(fastapi_app)
    fastapi_app.include_router(health.router)
    fastapi_app.include_router(generated.router)

    @fastapi_app.get("/__test__/not-implemented", include_in_schema=False)
    async def test_not_implemented(_: Request) -> None:
        raise AppError(
            code=501,
            message="Endpoint skeleton is present but domain logic is not implemented",
            status_code=501,
        )

    return fastapi_app


app = create_app()
```

- [ ] **Step 6: Run contract tests**

Run:

```bash
cd python-api
pytest tests/test_route_smoke.py tests/test_openapi_contract.py -q
```

Expected: PASS.

- [ ] **Step 7: Commit**

```bash
git add python-api/contracts/openapi_allowlist.json python-api/tests/test_route_smoke.py python-api/tests/test_openapi_contract.py python-api/app/main.py
git commit -m "test(python-api): cover route and OpenAPI skeleton parity"
```

## Task 6: Quality Gates And Graph Update

**Files:**
- Modify: `python-api/README.md`
- Modified by command: `graphify-out/**`

- [ ] **Step 1: Document verification commands**

Modify `python-api/README.md` so the command block includes:

```markdown
## Verification

```bash
ruff check .
mypy app tools
pytest -q
python -m compileall app tools
```
```

- [ ] **Step 2: Run full Python verification**

Run:

```bash
cd python-api
ruff check .
mypy app tools
pytest -q
python -m compileall app tools
```

Expected: all commands pass.

- [ ] **Step 3: Run graph update from repository root**

Run:

```bash
cd ..
graphify update .
```

Expected: graphify updates `graphify-out/**`.

- [ ] **Step 4: Review git status**

Run:

```bash
git status --short
```

Expected: only `python-api/**`, `docs/superpowers/plans/2026-05-31-python-api-skeleton-parity.md`, and graphify output changed by this work are staged or unstaged. Existing unrelated dirty files under `internal/logic/email/**` must remain untouched.

- [ ] **Step 5: Commit final verification docs and graph update**

```bash
git add python-api/README.md graphify-out docs/superpowers/plans/2026-05-31-python-api-skeleton-parity.md
git commit -m "docs(python-api): add skeleton parity implementation plan"
```

## Self-Review Notes

- Spec coverage: the plan covers a side-by-side `python-api/`, FastAPI foundation, Go-compatible envelope, route/schema inventory, generated routers, generated Pydantic models, OpenAPI security metadata, route smoke tests, OpenAPI contract tests, and graph update.
- Scope: this plan implements milestone 1 skeleton parity only. Business logic rewrites, repositories, Redis consumers, cron jobs, and production deployment remain outside this milestone, matching the approved non-goals.
- Ambiguity resolved: generated endpoints use a consistent `501` envelope for stubs; health and version are the only real runtime endpoints in this milestone.
