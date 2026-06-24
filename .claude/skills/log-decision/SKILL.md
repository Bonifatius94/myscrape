---
name: log-decision
description: Record a non-trivial technical decision for myscrape-go into specs/DEVLOG.md. Use whenever you pick a library, pattern, or approach where a reasonable alternative existed — especially after researching options.
---

# Log a decision

Append to the **Decisions** section of `specs/DEVLOG.md`, newest on top. Use the
next `G-NNN` id.

## Format
```
### G-0NN — <short imperative title>
<1–4 sentences: the Context (what forced the choice), the Decision, and the Why
(what alternative was rejected and the trade-off). Reference evidence in
specs/EXPERIMENTS.md or a commit when relevant.>
```

## When to log
- Picking a library or pattern where a real alternative existed.
- Changing an operating point, an interface seam, or the error taxonomy.
- Anything a future contributor would otherwise have to reverse-engineer from the
  diff.

Keep it tight. One decision per entry. If a decision is later reversed, add a new
entry that supersedes it rather than editing history.
