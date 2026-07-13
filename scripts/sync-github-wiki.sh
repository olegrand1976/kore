#!/usr/bin/env bash
# Sync project documentation to the GitHub wiki repository.
# Requires: GITHUB_TOKEN (contents:write), GITHUB_REPOSITORY (owner/repo).
set -euo pipefail

REPO="${GITHUB_REPOSITORY:?GITHUB_REPOSITORY is required}"
TOKEN="${WIKI_SYNC_TOKEN:-${GITHUB_TOKEN:-}}"
if [[ -z "${TOKEN}" ]]; then
  echo "Error: set WIKI_SYNC_TOKEN (recommended) or GITHUB_TOKEN." >&2
  exit 1
fi
SOURCE_SHA="${GITHUB_SHA:-local}"
WIKI_URL="https://x-access-token:${TOKEN}@github.com/${REPO}.wiki.git"

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
WORK_DIR="$(mktemp -d)"
trap 'rm -rf "${WORK_DIR}"' EXIT

git config --global user.name "github-actions[bot]"
git config --global user.email "41898282+github-actions[bot]@users.noreply.github.com"

echo "Cloning wiki repository for ${REPO}..."
if git clone "${WIKI_URL}" "${WORK_DIR}/wiki" 2>/dev/null; then
  :
else
  echo "Wiki repository empty or missing — initializing..."
  mkdir -p "${WORK_DIR}/wiki"
  git -C "${WORK_DIR}/wiki" init -b master
  git -C "${WORK_DIR}/wiki" remote add origin "${WIKI_URL}"
fi

WIKI="${WORK_DIR}/wiki"
cd "${WIKI}"

# Remove previously synced trees (keep .git only).
find . -mindepth 1 -maxdepth 1 ! -name '.git' -exec rm -rf {} +

echo "Copying documentation/ → Documentation/"
mkdir -p Documentation
rsync -a --delete "${ROOT}/documentation/" Documentation/

echo "Copying technical/ → Technical/"
mkdir -p Technical
rsync -a --delete "${ROOT}/technical/" Technical/

echo "Copying db/migrations/README.md → Database-Migrations.md"
cp "${ROOT}/db/migrations/README.md" Database-Migrations.md

SYNC_DATE="$(date -u +"%Y-%m-%d %H:%M UTC")"
cat > Home.md <<EOF
# Kore — Wiki projet

Documentation synchronisée automatiquement depuis la branche \`main\`.

| | |
| --- | --- |
| Dernière sync | ${SYNC_DATE} |
| Commit source | \`${SOURCE_SHA:0:7}\` |
| Dépôt | [github.com/${REPO}](https://github.com/${REPO}) |

---

## Documentation fonctionnelle

| Page | Description |
| --- | --- |
| [SPECIFICATION_FONCTIONNELLE](Documentation/SPECIFICATION_FONCTIONNELLE) | Spécification fonctionnelle détaillée (SFD) |
| [SCHEMA_DB](Documentation/SCHEMA_DB) | Schéma PostgreSQL actuel (tables, relations) |
| [CHARTE_GRAPHIQUE](Documentation/CHARTE_GRAPHIQUE) | Charte visuelle UI |
| [ANALYSE_COMMERCIALE](Documentation/ANALYSE_COMMERCIALE) | Analyse commerciale |

## Spécifications techniques

| Page | Description |
| --- | --- |
| [Technical/README](Technical/README) | Index des specs techniques |
| [Technical/ROADMAP](Technical/ROADMAP) | Roadmap par phases |
| [Technical/foundation/03-database](Technical/foundation/03-database) | Principes base de données |
| [Database-Migrations](Database-Migrations) | Ordre des migrations SQL |

> Les liens pointent vers les pages wiki générées depuis \`documentation/\`, \`technical/\` et \`db/migrations/\`.
EOF

git add -A
if git diff --staged --quiet; then
  echo "Wiki already up to date — nothing to commit."
  exit 0
fi

git commit -m "docs(wiki): sync from ${SOURCE_SHA:0:7} (${SYNC_DATE})"

WIKI_BRANCH="$(git branch --show-current)"
if [[ -z "${WIKI_BRANCH}" ]]; then
  WIKI_BRANCH=master
fi
if ! git push origin "HEAD:${WIKI_BRANCH}"; then
  echo "Warning: wiki push failed (wiki disabled or token lacks wiki scope) — skipping." >&2
  exit 0
fi

echo "Wiki synced successfully."
