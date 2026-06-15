---
name: Auth security section
overview: Capture how the serious consequences of authentication failures (for users, the business, and trust) change the way we build and operate this auth service compared to a typical CRUD API, and the concrete hardening items that follow for this repo.
todos: []
isProject: false
---

## Goal

Answer the question: "Authentication failures can have serious consequences for users, for the business, and for trust. What does that change about how you'd build or operate this service compared to a typical CRUD API?" The deliverable is a clear, code-grounded write-up of the mindset shift and the concrete hardening items that follow for this repo.

## Why auth differs from a typical CRUD API

- A CRUD API optimizes the happy path; for auth the failure paths *are* the product.
- The failures are asymmetric and expensive: a false accept is a breach/fraud, a false reject locks out a legitimate (paying) user, and a silent failure is the worst of both.
- So the principles invert: fail closed by default, never leak information, treat every failure as a security signal, and assume adversaries rather than just load.

## What that changes (concrete items for this repo)

- Fail closed everywhere: any error on the login/verify path must result in "denied", never a token. We already do this in `SignIn` and the handler's `default` case, but it needs explicit tests on every error branch so a refactor can't accidentally fail open.
- Don't leak which accounts exist (enumeration): login already returns a uniform "invalid email or password", but we skip bcrypt when the user is missing (a timing leak) and signup's `409` reveals existence. Compare against a dummy hash on the missing-user path; prefer email-verification flows over a direct conflict message.
- Secrets / JWT signing key (highest blast radius): remove the hardcoded `defaultJWTSecret` in `cmd/app/main.go`, refuse to boot if missing/placeholder, load from a secret manager, support rotation via `kid`.
- Token lifecycle: 24h HS256 tokens have no `jti`, no refresh, no revocation. Move to short-lived access tokens + refresh/sessions and a revocation denylist by `jti`.
- Verification path: when we add token-validating middleware, pin the algorithm (reject `alg=none` / HS-vs-RS confusion) and validate `iss`/`aud`/expiry strictly.
- Abuse protection: add per-IP and per-account throttling plus lockout/backoff, placed *in front of* the bcrypt concurrency cap so flooding `/signup` can't starve the hashing slots and cause a false-reject DoS.
- Audit logging + detection: structured auth events (outcome, source IP, user-agent, request id; never passwords/tokens) and alerts on failure-rate anomalies.
- Transport + availability: serve only over TLS (we currently run plain `ListenAndServe`); treat auth as product-critical with a higher SLO, and ensure any degradation fails closed.
- Least-privilege data access: `email` + `password_hash` are credentials/PII - least-privilege DB user, encryption at rest, restricted backups, and watch rollbacks/migrations for security regressions.

## Notes / considerations

- Analysis is grounded in the current code (onion architecture: `internal/app`, `internal/adapters/security`, `internal/transport/http`).
- The verification path does not exist yet (we only issue tokens), which is where most auth-failure handling will eventually live.
