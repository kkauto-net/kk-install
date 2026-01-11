---
title: "License Verification for kk init"
description: "Add license validation step before Docker check in kk init command"
status: in-progress
priority: P1
effort: 2h
branch: main
tags: [license, security, init, api]
created: 2026-01-11
---

# License Verification for kk init

## Overview

Add license verification as Step 0 in `kk init` command flow. License must be validated against `https://kkauto.net/api/license/config` before allowing stack initialization.

## Problem Statement

Currently `kk init` allows anyone to initialize a stack. Need to:
- Validate license key against remote API
- Retrieve `public_key` from server response
- Store both in `.env` file
- Block initialization if license invalid

## Implementation Phases

| Phase | Description | Status | Effort |
|-------|-------------|--------|--------|
| [Phase 01](phase-01-license-module.md) | Create `pkg/license/` module | done | 45m |
| [Phase 02](phase-02-init-integration.md) | Integrate license step into init flow | pending | 45m |
| [Phase 03](phase-03-tests-i18n.md) | Add tests and i18n messages | pending | 30m |

## Architecture

```
pkg/
├── license/           [NEW]
│   ├── license.go     # LicenseClient, Validate(), ValidateFormat()
│   └── license_test.go
├── templates/
│   └── embed.go       # Add LicenseKey, ServerPublicKey to Config
└── ui/
    ├── lang_en.go     # Add license messages
    └── lang_vi.go     # Add license messages
```

## API Specification

**Endpoint:** `POST https://kkauto.net/api/license/config`

**Request:**
```json
{"license": "LICENSE-64ABBE22C2134D1D"}
```

**Success Response:**
```json
{
  "status": "success",
  "public_key": "<encrypted_key>",
  "message": "License configuration retrieved successfully"
}
```

## Key Decisions

| Aspect | Decision |
|--------|----------|
| Step position | Step 0 (before Docker check) |
| Error behavior | Block completely |
| Force mode | Still requires license |
| Storage | Project-level (.env) |
| Validation | Format regex + Server API |

## Success Criteria

- [ ] License prompt appears first in `kk init`
- [ ] Invalid format rejected before API call
- [ ] Invalid license blocks init with clear error
- [ ] Valid license saves LICENSE_KEY and SERVER_PUBLIC_KEY_ENCRYPTED to .env
- [ ] Force mode still requires license
- [ ] Tests pass with mocked API

## Related Files

- [cmd/init.go](../../cmd/init.go) - Main init command
- [pkg/templates/embed.go](../../pkg/templates/embed.go) - Config struct
- [pkg/templates/env.tmpl](../../pkg/templates/env.tmpl) - Env template
- [Brainstorm Report](../reports/brainstorm-260111-1138-license-verification.md)
