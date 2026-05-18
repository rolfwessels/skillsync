---
description: General Go coding practices
alwaysApply: true
paths:
  - "*.go"
---

# General Go Practices

- Follow SOLID principles; keep functions under 15 lines
- Wrap errors with context: `fmt.Errorf("doing X: %w", err)`
- Sentinel errors use `Err` prefix and `errors.Is` checks
- Use interface-based dependency injection for testability
- No comments that narrate what the code does
