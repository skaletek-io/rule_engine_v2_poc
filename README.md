# Rule Engine V2 — POC

A small **Skaletek Rule V2** proof of concept in Go: **events** are evaluated against **boolean rules** compiled with [expr-lang/expr](https://github.com/expr-lang/expr). When a rule’s expression evaluates to `true`, an **alert** is emitted and optionally written to disk.

## What it demonstrates

- **Declarative rules** — Each rule is a string expression (e.g. `payment.amount > 10000 && sender.accountAge < 30`) plus metadata (name, severity, message).
- **Template-scoped rules** — `Rule.TemplateID` must match `Event.TemplateID` (unless the rule’s template is empty). That keeps banking rules off fintech events, etc.
- **Compile once, run many** — Rules are compiled to `*vm.Program` up front; each event runs the relevant programs with `expr.Run`.
- **Nested JSON-style payloads** — Event `Payload` is `map[string]any`. Before evaluation, `flattenPayload` merges one level of nested maps so expressions can use dotted names like `payment.amount` and top-level keys from nested objects (e.g. `sender` fields also appear at the top level of the env map).
- **Resilient evaluation** — If `expr.Run` fails (missing field, type issue), that rule is skipped for the event with a log line; other rules still run.

## Domains in the seed data

Sample **rules** and **events** cover six template types:

| Template ID           | Domain (illustrative)   |
|-----------------------|-------------------------|
| `bank_wire_transfer`  | Banking / wires         |
| `fintech_payment`     | Fintech / card payments |
| `crypto_withdrawal`   | Crypto withdrawals      |
| `hotel_reservation`   | Hospitality             |
| `insurance_claim`     | Insurance claims        |
| `ecommerce_order`     | E-commerce orders       |

These are synthetic scenarios (AML-style, fraud, sanctions, etc.) for exercising the engine, not production policy.

## Project layout

| File        | Role |
|-------------|------|
| `main.go`   | Orchestrates compile → per-event evaluation → console output → `results.json` |
| `engine.go` | `compileRules`, `evaluateRules`, `flattenPayload`, `CompiledRule` |
| `rules.go`  | `Rule` struct and `seedRules()` |
| `events.go` | `Event` struct and `seedEvents()` |
| `alert.go`  | `Alert`, `newAlert`, `persistAlerts` |

## Run

```bash
go run .
```

Dependencies are resolved via `go.mod` (Go **1.24.5**, `github.com/expr-lang/expr`).

After a run, matched alerts are appended in memory across all events and written to **`results.json`** as pretty-printed JSON (`id`, `event_id`, `rule_id`, `rule_name`, `severity`, `message`, `status`, `created_at`).

## Limits of this POC

- Rules and events are **in-code seeds**, not loaded from a database or API.
- No persistence layer, queues, or auth — single-process demo only.
- Expression environment is a **flattened map**; deeply nested paths beyond one level are not automatically bridged unless keys are duplicated at the top level by the flattener.

---

*Internal Skaletek experiment — not a production service.*
