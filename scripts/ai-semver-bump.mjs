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

function fallbackFromCommits(text) {
  const lower = text.toLowerCase()
  if (lower.includes('breaking change') || lower.includes('!:')) return 'major'
  if (/(^|\n)\s*feat(\(.+\))?:/i.test(text)) return 'minor'
  if (/(^|\n)\s*fix(\(.+\))?:/i.test(text)) return 'patch'
  return 'patch'
}

async function openAIClassify({ apiKey, model, commits }) {
  const prompt = [
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

  const res = await fetch('https://api.openai.com/v1/chat/completions', {
    method: 'POST',
    headers: {
      Authorization: `Bearer ${apiKey}`,
      'Content-Type': 'application/json'
    },
    body: JSON.stringify({
      model,
      temperature: 0,
      messages: [
        { role: 'system', content: 'Return only valid JSON, no markdown.' },
        { role: 'user', content: prompt }
      ]
    })
  })

  if (!res.ok) {
    const body = await res.text()
    throw new Error(`OpenAI error ${res.status}: ${body}`)
  }
  const data = await res.json()
  const content = data?.choices?.[0]?.message?.content
  if (!content) throw new Error('OpenAI: empty response')
  let parsed
  try {
    parsed = JSON.parse(content)
  } catch {
    throw new Error(`OpenAI: non-JSON response: ${content}`)
  }
  const bump = parsed?.bump
  if (!bumpTypes.includes(bump)) throw new Error(`OpenAI: invalid bump: ${String(bump)}`)
  return { bump, reason: String(parsed?.reason || '') }
}

async function main() {
  const args = parseArgs(process.argv.slice(2))
  const current = args.current
  const commits = args.commits || ''
  if (!current) fail('Missing --current')

  const apiKey = process.env.OPENAI_API_KEY || ''
  const model = process.env.OPENAI_MODEL || 'gpt-4o-mini'

  const fallback = fallbackFromCommits(commits)
  let decision = { bump: fallback, reason: 'fallback' }

  if (apiKey) {
    try {
      decision = await openAIClassify({ apiKey, model, commits })
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

main().catch((e) => fail(String(e?.message || e)))

