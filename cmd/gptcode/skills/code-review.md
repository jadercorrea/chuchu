---
layout: skill
title: Code Review
name: code-review
language: general
description: Conduct thorough code reviews focusing on security, correctness, and maintainability. Prioritizes issues by severity and provides actionable feedback.
---

# Code Review

Conduct a thorough code review following a structured priority system.

## When to Activate

- When asked to review code or a PR/MR
- When reviewing changes before commit
- "Review this code" / "Check my changes"

## Review Priority (in order)

1. **Security issues** - SQL injection, XSS, auth bypasses, exposed secrets
2. **Database schema** - Missing indexes for queried columns, missing constraints
3. **Logic errors / bugs** - Off-by-one, null handling, race conditions
4. **Missing error handling** - Unhandled exceptions, missing validations
5. **Missing tests** - New methods without corresponding tests
6. **N+1 queries / DB concerns** - Queries in loops, missing eager loading
7. **API design / interface** - Public method signatures, breaking changes
8. **Edge cases** - Boundary conditions, empty states
9. **Naming clarity** - Misleading or vague names
10. **Method size** - Methods doing too much, extraction opportunities
11. **Style** - Minor improvements, readability

## Test Requirements

- Every new public method should have a test (state as fact if missing)
- Complex methods missing tests are blockers
- Trivially simple methods (one-liners, simple delegation) can skip tests

## Comment Format

Each comment should follow this structure:

```markdown
**path/to/file.rb:42**
```diff
+ def fetch_all!
```

The comment text here. Be specific and actionable. [1]
```

Rules for comments:

- Include the line(s) of code being discussed using diff syntax
- Number each comment at the end: [1], [2], etc.
- Be direct (no "perhaps consider" hedging)
- Explain WHY, not just WHAT (unless the what is cryptic)
- Don't state the obvious
- Comments must be specific, actionable, and evidence-based
- Cite evidence from production code, not from tests

### Good vs Bad Comments

**Bad**: "This could potentially cause issues with larger datasets."
**Good**: "This loads all records into memory. With 10k+ records, this will OOM."

**Bad**: "Nice use of has_many :through!"
**Good**: (Don't comment on obvious/standard patterns)

**Bad**: "Consider adding error handling here."
**Good**: "If `results` is null, `#each` raises NoMethodError. Check the API docs for what's returned when there's no data."

## Output Format

```markdown
# Code Review

[Comments numbered [1], [2], etc.]

---

## Summary

[One-liner assessment OR list of required changes if there are architectural
problems. For clean PRs: brief statement of what the change accomplishes.]
```

## Things to Avoid

- **No caching suggestions** unless obviously necessary
- **No sharding suggestions** ever
- **No "what if this scales" speculation** - review the code as-is
- **No over-engineering suggestions** - keep it pragmatic
- **No astronaut architecture** - don't suggest abstractions for one-time code
- **Don't praise standard patterns** - only noteworthy decisions
- **Don't be verbose** - keep comments concise

## Preserve Developer Intent

Suggest pragmatic solutions that preserve developer intent rather than blanket
prohibitions. Only push back hard when the intent is clearly low quality, unsafe,
or absurd.

**Good**: "Wrap this in `unless Rails.env.production?` to keep it safe"
**Bad**: "Remove this entirely" (when the code serves a legitimate dev purpose)
