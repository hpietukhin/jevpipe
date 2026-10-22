# jevPipe

Unix pipe to [TypeSafe](https://typesafe.ai) **Jev** (`/v1/systemone`). Reads **stdin**, writes **JSON** to **stdout** (see [unix-piping-course](https://github.com/gerbenjacobs/unix-piping-course)).

Go **1.27**.

| Package | Role |
| --- | --- |
| `routing` | `StdinState` (JSON + `UseNumber`), `Question` — stdin → TypeSafe question |
| `typesafe` | System One client (`encoding/json/v2` on the wire) returning decoded answers, primitives |
| `main` | CLI: args, env config, stdout JSON |

## Setup

```bash
cd ~/Developer/jevPipe
cp mise.jev.toml.example mise.jev.toml   # or copy from review-mine
mise exec -E jev -- go build -o jev .   # loads mise.jev.toml
# optional: ln -sf "$(pwd)/jev" ~/bin/jev
```

`mise.jev.toml` supplies `JEV_KEY` (same as review-mine).

## Usage

```bash
echo '{"project_id":31935,"name":"asistent-theseus"}' | jev "what is the project id?"
cat record.json | jev "does this mention billing?"
cat thread.json | jev "is this comment about code style?" | jq .
```

- **JSON object or array on stdin** → parsed as `state`; for open questions, leaf scalar values become **choice** candidates (Jev picks one, or `none`).
- **Anything else on stdin** (plain text, a bare JSON scalar, nothing) → `state` is that value; always **noul** (yes/no probability).
- Questions starting with *is / does / has / …* → **noul** even when JSON has many fields.

## Environment

| Variable | Default |
| --- | --- |
| `JEV_KEY` | required |
| `JEV_BASE_URL` | `https://api.typesafe.ai` |
| `JEV_MODEL` | `jev-latest` |

## Output

Compact JSON: `model`, `usage`, then every field of the answer. For example:

```json
{"model":"jev-1.14.0","usage":{...},"type":"choice","choice":"31935","confidence":1,"probabilities":{"31935":1,"none":0}}
```

```json
{"model":"jev-1.14.0","usage":{...},"type":"noul","noul":0.98}
```

Errors go to **stderr**; exit code `1` on failure, `2` on usage.

## Tests

```bash
mise exec -E jev -- go test ./... -count=1
```

Client and CLI tests call the real System One API with `JEV_KEY` and skip when no key is set. `-count=1` bypasses Go's test cache so they actually hit the API.

Property tests use [Hegel](https://hegel.dev/go/hegel) (`state_hegel_test.go`, `typesafe/truncate_hegel_test.go`). Hegel requires the native `libhegel` runtime (installed via `go get` / FFI).
