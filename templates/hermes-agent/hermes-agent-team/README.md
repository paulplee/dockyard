# Hermes Agent Team Setup

Four-profile agent team based on the NeoAIForecast article on Hermes multi-agent profiles.

## Role map

| Profile | Original name | Role |
|---|---|---|
| `orchestrator` | Hermes | Plans, decomposes, routes, synthesizes |
| `research` | Rebecca | Web research, source validation, fact-checking |
| `narrative` | Natelie | Writing, editing, structure, documentation |
| `debugger` | Debra | Code review, debugging, systems engineering |

## What Hermes profiles actually are

A profile is a **separate Hermes home directory** at `~/.hermes/profiles/<name>/`.
Each gets its own `SOUL.md`, `config.yaml`, `.env`, memories, skills, sessions, cron jobs, and state database.
When you create a profile called `research`, Hermes also registers a command alias `research`
so that `research chat`, `research setup`, `research gateway start` all work out of the box.

## Resulting directory layout

After completing all steps below:

```
~/.hermes/                        <- default profile = orchestrator
  SOUL.md                         <- orchestrator identity
  config.yaml
  .env
  memories/
  skills/

~/.hermes/profiles/
  research/
    SOUL.md                       <- research identity
    config.yaml
    .env
    memories/
    skills/
  narrative/
    SOUL.md
    config.yaml
    .env
    memories/
    skills/
  debugger/
    SOUL.md
    config.yaml
    .env
    memories/
    skills/

~/.hermes/kanban.db               <- shared task board (all profiles read/write this)
```

---

## Full setup instructions

### 1. Install Hermes (skip if already installed)

```bash
curl -fsSL https://raw.githubusercontent.com/NousResearch/hermes-agent/main/scripts/install.sh | bash
```

Requires Node.js, Python 3.11+, and a model with at least 64,000 tokens of context.
After install, run `hermes setup` to configure your API keys and model.

---

### 2. Configure the default profile as the orchestrator

The default install at `~/.hermes` becomes the orchestrator.

```bash
# Write the orchestrator SOUL.md from this repo
cp orchestrator_SOUL.md ~/.hermes/SOUL.md

# Set the model (strong reasoning for routing decisions)
hermes config set model.default anthropic/claude-sonnet-4-5

# Point to your working project directory
hermes config set terminal.cwd /absolute/path/to/your/project

# Install the orchestrator Kanban skill
# (encodes decomposition rules and anti-temptation guards)
hermes skills install devops/kanban-orchestrator
```

> SOUL.md changes take effect on the **next new session** only.
> Existing sessions keep the prompt that was active when they started.

---

### 3. Create the research profile

```bash
# --clone copies config.yaml, .env, and SOUL.md — fresh memory and sessions
hermes profile create research --clone

# Overwrite the cloned SOUL.md with the research identity
cp research_SOUL.md ~/.hermes/profiles/research/SOUL.md

# Model choice: factual, grounded outputs benefit from a strong reasoning model
research config set model.default anthropic/claude-sonnet-4-5

# Install the kanban-worker skill so the Kanban dispatcher can assign tasks to it
research skills install devops/kanban-worker
```

---

### 4. Create the narrative profile

```bash
hermes profile create narrative --clone

cp narrative_SOUL.md ~/.hermes/profiles/narrative/SOUL.md

# Higher-capability model for nuanced long-form writing
narrative config set model.default anthropic/claude-opus-4-5

narrative skills install devops/kanban-worker
```

---

### 5. Create the debugger profile

```bash
hermes profile create debugger --clone

cp debugger_SOUL.md ~/.hermes/profiles/debugger/SOUL.md

debugger config set model.default anthropic/claude-sonnet-4-5

debugger skills install devops/kanban-worker
```

---

### 6. Initialise the Kanban board

The Kanban board lives at `~/.hermes/kanban.db` and is shared across all profiles.
This is the coordination layer for tasks that cross agent boundaries.

```bash
hermes kanban init
```

---

### 7. Start the gateway

The gateway hosts the embedded Kanban dispatcher that automatically picks up ready tasks
and spawns the correct profile for each one.

```bash
# Orchestrator gateway runs the dispatcher
hermes gateway start

# Optionally start specialist gateways (each needs its own bot token in its .env)
research  gateway start
narrative gateway start
debugger  gateway start
```

If you want each gateway to persist across reboots, install it as a systemd service:

```bash
hermes    gateway install
research  gateway install
narrative gateway install
debugger  gateway install
```

---

### 8. Verify everything

```bash
# List all profiles with status
hermes profile list

# Check Kanban board is initialised
hermes kanban stats

# Sanity-check each agent knows its role
hermes -p research  chat -q "what is your role?"
hermes -p narrative chat -q "what is your role?"
hermes -p debugger  chat -q "what is your role?"
```

---

## Daily usage

### Chat directly with a profile

```bash
orchestrator chat    # or just: hermes chat
research chat
narrative chat
debugger chat
```

### Assign a Kanban task manually

```bash
# Create a research task
hermes kanban create "Research thermal dissipation properties of graphene nanoplatelets" \
  --assignee research

# Chain a writing task that depends on the research result
hermes kanban create "Write a two-page technical summary for CB Nano investors" \
  --assignee narrative \
  --parent <research-task-id>

# Assign a debug task
hermes kanban create "Debug vLLM memory leak on pegasus RTX 5090" \
  --assignee debugger

# Watch all tasks in real time
hermes kanban watch
```

### Orchestrator-led decomposition

The cleanest workflow: give the orchestrator a high-level goal and let it decompose and route.

```bash
hermes chat
> Research the latest Qwen3 benchmarks, then write a technical comparison against Llama-4 Scout for PPB.
```

The orchestrator will use `kanban_create` to fan out a research task and a narrative task,
link them (narrative depends on research), and synthesise the final output when both are done.

---

## delegate_task vs Kanban

| | `delegate_task` | Kanban |
|---|---|---|
| Shape | Blocking RPC call (fork → join) | Durable message queue + state machine |
| Child identity | Anonymous, no persistent memory | Named profile with its own memory |
| Resumable | No — failed = failed | Yes — block, unblock, retry |
| Human-in-the-loop | Not supported | Comment / unblock at any point |
| Use when | Short answer needed in current response | Work crosses profiles or must survive restarts |

---

## Key files per profile

| File | What it does |
|---|---|
| `SOUL.md` | Agent identity — injected as layer 1 of the system prompt every session |
| `config.yaml` | Model, provider, toolsets, memory limits, terminal backend |
| `.env` | API keys and bot tokens (secrets never go in config.yaml) |
| `memories/MEMORY.md` | What the agent has learned about your environment (auto-managed) |
| `memories/USER.md` | What the agent has learned about you (auto-managed) |

---

## Managing profiles

```bash
hermes profile list                     # show all profiles with status
hermes profile show research            # detailed info for one profile
hermes profile rename narrative writer  # rename (updates alias + service)
hermes profile export research          # export to research.tar.gz
hermes profile delete debugger          # remove profile and all its data
```
