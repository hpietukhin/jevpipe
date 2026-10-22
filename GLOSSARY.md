# jevPipe glossary

- **Stdin state** — Content from standard input, parsed by `routing.StdinState` (JSON with `json.Number` for integers, or a plain string).
- **Question routing** — The `routing` module: `Question(state, prompt)` returns a `typesafe.Question` (`Noul` or `Choice`).
- **Scalar candidate** — A leaf value and JSON path collected for `Choice` criteria (`routing.scalarCandidate`, internal).
- **System One request** — Typed `typesafe.Request` sent to `POST /v1/systemone`.
- **CLI output** — JSON on stdout after a successful call (`output` in `main`: model, usage, and the answer's fields flattened).
