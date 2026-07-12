---
name: kore-dev-workflow
description: >-
  Workflow développement Kore (Go + Nuxt). Couvre make up, structure modules,
  BFF Nitro, auth cookies, migrations. Use when implementing features,
  debugging local stack, or onboarding on the Kore monorepo.
---

# Kore — workflow dev

## Démarrage

```bash
make up && make migrate && make seed
```

| Service | URL |
|---------|-----|
| Frontend | http://localhost:3001 |
| API | http://localhost:8081 |
| Admin | ADM_admin / Admin123! |

## Feature full-stack

1. Go : ports → app → handler
2. BFF : `frontend/server/api/` + `apiAuthHeaders`
3. UI : page + i18n FR/EN
4. Mobile : skill `kore-responsive-check`
5. Build : `go build ./...` && `npm run build`

## Références

- `.cursor/AGENTS.md`
- `documentation/CHARTE_GRAPHIQUE.md`
