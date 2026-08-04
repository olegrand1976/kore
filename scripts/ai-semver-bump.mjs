import { resolve } from 'node:path'
import { fileURLToPath } from 'node:url'

const bumpTypes = ['major', 'minor', 'patch']

function fail(msg) {
  process.stderr.write(`${msg}\n`)
  process.exit(1)
}

function parseArgs(argv) {
  const out = {}
  for (let i = 0; i < argv.length; i += 1) {
    const a = argv[i]
    if (!a.startsWith('--')) continue
    const k = a.slice(2)
    const v = argv[i + 1]
    out[k] = v
    i += 1
  }
  return out
}

function parseSemver(v) {
  const m = String(v).trim().replace(/^v/, '').match(/^(\d+)\.(\d+)\.(\d+)$/)
  if (!m) return null
  return { major: Number(m[1]), minor: Number(m[2]), patch: Number(m[3]) }
}

function nextVersion(current, bump) {
  const s = parseSemver(current)
  if (!s) fail(`Invalid current version: ${current}`)
  switch (bump) {
    case 'major':
      return `${s.major + 1}.0.0`
    case 'minor':
      return `${s.major}.${s.minor + 1}.0`
    case 'patch':
      return `${s.major}.${s.minor}.${s.patch + 1}`
    default:
      return `${s.major}.${s.minor}.${s.patch + 1}`
  }
}

// Le workflow fournit les commits sous forme de liste (« - <sujet> », cf.
// git log --pretty=format:'- %s') : le marqueur de puce doit être toléré, sinon
// aucun type conventionnel n'est reconnu et tout devient un patch.
const BULLET = String.raw`[-*+]\s*`

function conventionalType(text, type) {
  return new RegExp(String.raw`(^|\n)\s*(?:${BULLET})?${type}(\(.+?\))?!?:`, 'i').test(text)
}

function fallbackFromCommits(text) {
  const lower = text.toLowerCase()
  if (lower.includes('breaking change') || lower.includes('!:')) return 'major'
  if (conventionalType(text, 'feat')) return 'minor'
  if (conventionalType(text, 'fix')) return 'patch'
  return 'patch'
}

// Gemini est le fournisseur LLM du projet (cf. internal/modules/ai/adapters/gemini).
export const DEFAULT_GEMINI_MODEL = 'gemini-3.6-flash'
const GEMINI_BASE_URL = 'https://generativelanguage.googleapis.com/v1beta'

const SYSTEM_PROMPT = 'Return only valid JSON, no markdown.'

function buildPrompt(commits) {
  return [
    'You are a release automation assistant.',
    'Task: decide SemVer bump type for the next git tag based on commit messages since last tag.',
    'Output STRICT JSON ONLY with keys: bump (major|minor|patch), reason (string).',
    'Rules:',
    '- major when backward-incompatible changes are introduced (BREAKING CHANGE, or obvious API/behavior break).',
    '- minor when new features are added in a backward compatible way.',
    '- patch for bug fixes, chores, refactors, docs, tests, internal changes.',
    '',
    'Commit messages:',
    commits
  ].join('\n')
}

// Le modèle peut encadrer sa réponse d'une clôture markdown malgré la consigne.
export function stripJsonFence(text) {
  const trimmed = String(text).trim()
  const fenced = trimmed.match(/^```(?:json)?\s*\n?([\s\S]*?)\n?```$/i)
  return (fenced ? fenced[1] : trimmed).trim()
}

export function parseBumpPayload(raw) {
  let parsed
  try {
    parsed = JSON.parse(stripJsonFence(raw))
  } catch {
    throw new Error(`Gemini: non-JSON response: ${raw}`)
  }
  const bump = parsed?.bump
  if (!bumpTypes.includes(bump)) throw new Error(`Gemini: invalid bump: ${String(bump)}`)
  return { bump, reason: String(parsed?.reason || '') }
}

async function geminiClassify({ apiKey, model, commits, baseURL = GEMINI_BASE_URL }) {
  const url = `${baseURL.replace(/\/+$/, '')}/models/${model}:generateContent`
  const res = await fetch(url, {
    method: 'POST',
    headers: {
      'x-goog-api-key': apiKey,
      'Content-Type': 'application/json'
    },
    body: JSON.stringify({
      contents: [
        { role: 'user', parts: [{ text: `Instructions système:\n${SYSTEM_PROMPT}` }] },
        { role: 'model', parts: [{ text: 'Compris.' }] },
        { role: 'user', parts: [{ text: buildPrompt(commits) }] }
      ],
      generationConfig: { temperature: 0, responseMimeType: 'application/json' }
    })
  })

  const raw = await res.text()
  let data
  try {
    data = JSON.parse(raw)
  } catch {
    throw new Error(`Gemini decode: ${res.status}: ${raw.slice(0, 300)}`)
  }
  if (data?.error) throw new Error(`Gemini api: ${data.error.message || res.status}`)
  if (!res.ok) throw new Error(`Gemini http ${res.status}: ${raw.slice(0, 300)}`)

  const text = (data?.candidates ?? [])
    .flatMap((c) => c?.content?.parts ?? [])
    .map((p) => p?.text)
    .filter(Boolean)
    .join('\n')
    .trim()
  if (!text) throw new Error('Gemini: empty response')
  return parseBumpPayload(text)
}

async function main() {
  const args = parseArgs(process.argv.slice(2))
  const current = args.current
  const commits = args.commits || ''
  if (!current) fail('Missing --current')

  const apiKey = (process.env.GEMINI_API_KEY || '').trim()
  const model = (process.env.GEMINI_MODEL || '').trim() || DEFAULT_GEMINI_MODEL

  const fallback = fallbackFromCommits(commits)
  let decision = { bump: fallback, reason: 'fallback' }

  if (apiKey) {
    try {
      decision = await geminiClassify({ apiKey, model, commits })
    } catch (e) {
      process.stderr.write(`AI eval failed, using fallback: ${String(e?.message || e)}\n`)
      decision = { bump: fallback, reason: 'fallback' }
    }
  }

  const next = nextVersion(current, decision.bump)
  const out = {
    bump: decision.bump,
    reason: decision.reason,
    current,
    next,
    nextTag: `v${next}`
  }
  process.stdout.write(`${JSON.stringify(out)}\n`)
}

// N'exécute la CLI que lors d'un appel direct, pour que les helpers exportés
// puissent être importés (et testés) sans déclencher main().
const invokedDirectly =
  process.argv[1] !== undefined && resolve(process.argv[1]) === fileURLToPath(import.meta.url)

if (invokedDirectly) {
  main().catch((e) => fail(String(e?.message || e)))
}

