# jevPipe

Unix pipe to [TypeSafe](https://typesafe.ai) **Jev** (`/v1/systemone`). Reads **stdin**, writes **JSON** to **stdout** (see [unix-piping-course](https://github.com/gerbenjacobs/unix-piping-course)).

Go **1.27**.

| Package | Role |
| --- | --- |
| `routing` | `StdinState` (JSON + `UseNumber`), `Question` — stdin → TypeSafe question |
| `typesafe` | System One client (`encoding/json/v2` on the wire), primitives, CLI output |

## Setup

```bash
cd ~/Developer/jevPipe
cp mise.jev.toml.example mise.jev.toml   # or copy from review-mine
mise exec -E jev -- go build -o jev .   # loads mise.jev.toml
# optional: ln -sf "$(pwd)/jev" ~/bin/jev
```

`mise.jev.toml` supplies `JEV_KEY` (same as review-mine). `TYPESAFE_API_KEY` also works.

## Usage

```bash
echo '{"project_id":31935,"name":"asistent-theseus"}' | jev "what is the project id?"
cat record.json | jev "does this mention billing?"
cat thread.json | jev "is this comment about code style?" | jq .
```

- **JSON on stdin** → parsed as `state`; for open questions, leaf scalar values become **choice** candidates (Jev picks one).
- **Plain text on stdin** → `state` is the string; **noul** (yes/no probability) unless the question looks like yes/no already.
- Questions starting with *is / does / has / …* → **noul** even when JSON has many fields.

## Environment

| Variable | Default |
| --- | --- |
| `JEV_KEY` / `TYPESAFE_API_KEY` | required |
| `TYPESAFE_BASE_URL` | `https://api.typesafe.ai` |
| `JEV_MODEL` / `TYPESAFE_DEFAULT_MODEL` | `jev-latest` |

## Output

Compact JSON, for example:

```json
{"model":"jev-latest","type":"choice","choice":"31935","confidence":0.98,"usage":{...}}
```

```json
{"model":"jev-latest","type":"noul","noul":0.91,"usage":{...}}
```

Errors go to **stderr**; exit code `1` on failure, `2` on usage.

## Tests

```bash
go test ./...
```

Property tests use [Hegel](https://hegel.dev/go/hegel) (`state_hegel_test.go`, `typesafe/truncate_hegel_test.go`). Hegel requires the native `libhegel` runtime (installed via `go get` / FFI).
