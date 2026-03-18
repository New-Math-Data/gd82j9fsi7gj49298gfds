# AI Collaboration Notes

## Overall Approach

I used Claude as a pair-programmer throughout this assessment. My role was to drive requirements, constraints, and spec interpretation — then review, annotate, and take ownership of the generated code. Claude's role was to handle structure, boilerplate, and implementation of well-understood patterns (BFS, JSON marshalling, CSV parsing).

The workflow: I read the README carefully, identified ambiguities in the spec (e.g. the bracket-array token behavior in `Z1=[0.141506, 1.399508]`), and gave Claude explicit constraints before any code was written. This produced a detailed plan that I reviewed before implementation began.

## How to Read `claude.transcript`

The transcript is a running log of every exchange, formatted as `[User]` / `[Claude]` turns. It shows:

- How requirements were refined iteratively (e.g. the decision to use unweighted BFS without a visited map — which Claude then flagged as incorrect for undirected graphs and fixed)
- Where I gave explicit constraints (no comments in generated code, bracket-array token behavior, test strategy)
- Where Claude caught issues proactively (the infinite-loop bug in BFS for disconnected nodes)

Reading the transcript alongside the code gives a clear picture of what was human-directed vs. AI-generated.

## Where to Look in the Code

All implementation is in `main.go`. The comments throughout are my own additions, written after reviewing the generated code. They reflect my understanding of the logic, not Claude's explanation of it.

Key design decisions I drove:
- `normalizeBusID` handles both `.` and `_` separators, with the special case that only numeric-prefix `_` splits are applied (e.g. `tyn201_feeder` is a line ID, not a bus)
- `parseSpecs` accumulates multi-token quoted values (for `wires="..."`) and bracket-array values (for `Z1=[0.141506, 1.399508]`) — the latter deviates from the README's expected output, which shows a spec error: the README naively splits on whitespace and produces `"Z1": "[0.141506,"` and `"1.399508]": "true"`, which is clearly wrong. The implementation instead produces `"Z1": "[0.141506, 1.399508]"`.
- BFS uses a visited set — necessary for undirected graphs even when the underlying data is acyclic
