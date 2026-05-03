# Hermes Agent Team — Project Setup

## The team

| Role       | What they do                                            |
|------------|---------------------------------------------------------|
| **orchestrator** | Task decomposition, routing, synthesis, final delivery |
| **research**     | Source-first research, fact-validation, data gathering  |
| **narrative**    | Writing, structure, tone, editing, documentation        |
| **debugger**     | Code review, debugging, systems engineering             |

The orchestrator breaks your work into subtasks, routes each to the right specialist, and synthesizes the results.

## Workspace layout

```
/workspace/
  .hermes/
    AGENTS.md            -- Agent instructions (auto-injected into every session)
  config/
    profiles/            -- Profile-level config overrides for this project
    skills/              -- Project-scoped skills
  docs/
    architecture/        -- Architecture docs and diagrams
    decisions/           -- ADRs: Architecture Decision Records
    guides/              -- How-to guides, runbooks, conventions
  logs/                  -- Runtime and experiment logs
  outputs/               -- Generated artifacts, reports, code drops
  src/                   -- Source code
  data/                  -- Input data, datasets, fixtures
  tasks/                 -- Kanban boards, sprint plans, task tracking
```

### Directory rules

- **src/** — Actual source code. Subdivide by module, language, or component.
- **data/** — Input data the team consumes. Keep it under 100 MB. Large datasets go in external storage.
- **outputs/** — Anything we generate. Never hand-edit. Always track provenance.
- **docs/decisions/** — One ADR per file, numbered: `001-choose-postgres.md`.
- **logs/** — Experiments, test runs, anything worth replaying.

## Getting started

### Prerequisites

- Hermes Agent running with a configured model
- At least one session connected (CLI, Discord, Telegram, etc.)

That's it. No installs, no plugins, no setup scripts.

The worker profiles (research, narrative, debugger) and the kanban dispatcher are already part of Hermes. The orchestrator is you, the agent.

### First interaction

Just tell me what you need. I'll:

1. Break the work into subtasks
2. Assign each to the right profile
3. Use kanban to track cross-agent dependencies
4. Synthesize everything into a final deliverable

For single-agent work, I handle it directly. For multi-agent work, I use kanban boards.

## Kanban

Kanban is built into Hermes. No setup required.

### When to use kanban

- Work involves multiple profiles (research + narrative, etc.)
- Tasks have dependencies
- You want persistent tracking across sessions
- You need status updates on long-running work

### When to use direct delegation

- Single task, no cross-agent involvement
- Quick reasoning answer
- Everything fits in one turn

### How it works

- The dispatcher routes tasks to worker profiles automatically
- Tasks live in `~/.hermes/kanban.db` (SQLite)
- Board state persists across sessions — I can pick up where we left off
- Task lifecycle: `created` → `running` → `done` (or `blocked` / `crashed` / `timed_out`)
- Workers can leave comments on tasks for status updates

### Creating a board

I handle this for you when the work warrants it. You just say what you want done.

## Skills

Skills are reusable procedures — step-by-step instructions I follow when a recurring task type comes up.

### Where skills live

```
~/.hermes/skills/<name>/
  SKILL.md              -- The skill definition (required)
  references/           -- Reference docs
  scripts/              -- Helper scripts
  templates/             -- Reusable templates
```

### Adding a project skill

Place it in `/workspace/config/skills/` and I'll load it for relevant tasks.

### Saving a new skill

After any complex or non-trivial task (5+ tool calls, tricky debugging, unusual workflow), I'll save it as a skill so we don't re-invent it next time.

### Updating a skill

If a skill has wrong commands or missing steps, I patch it immediately.

## Agent conventions

### orchestrator
- Decompose before routing — never skip planning
- Quality-gate all delegated outputs before returning them to you
- State uncertainty explicitly — never fabricate results or attribute them to a worker

### research
- Cite sources with URLs
- Distinguish facts from inferences
- Flag any validation gaps

### narrative
- Write for terminal reading — keep formatting minimal
- Track revisions in `/docs/guides/` for iterative content
- Match the requested tone and language

### debugger
- Reproduce before fixing
- Document the root cause, not just the patch
- Prefer minimal, targeted changes

## Tips

- I remember things across sessions. Tell me preferences and I'll save them.
- Use `/workspace/outputs/` for everything we generate — keeps `src/` and `docs/` clean.
- Log experiments in `/workspace/logs/` if you want to replay them later.
- Ask me to list kanban tasks anytime to see project status.
- If I make a mistake, correct me — I'll save the correction as memory or a skill.