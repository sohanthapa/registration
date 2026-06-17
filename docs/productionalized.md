# Productionalization

A checklist of items to consider when taking this service to production.

*NOTE*: The sections/steps below are to the best of my knowledge. I'm open to more ideas or feedback from the team if something doesn't look right, something is missing, or there are better approaches.

## Git Branch Setup

- Enforce a feature-branch workflow: all work happens on a branch off `main`, and direct pushes/commits to `main` are blocked via branch protection. Changes land through a pull request.
- Require at least one approving review on every PR before it can be merged into `main` (stale approvals dismissed on new commits).
- Require all status checks (lint, unit, and integration tests) to pass, and the branch to be up to date with `main`, before merging.

## Development and Testing

- Seeding test data, in lower env, for easier testing.
- Add integration tests: Our unit tests rely on mocks, so they check pieces in isolation but never exercise the real flow. Integration tests run the actual pieces together - a real request through the API down to a real database - e.g. duplicate emails are rejected, passwords are stored hashed, wrong passwords fail).
- Add more unit tests if necessary.

## Database

- add migration up and down and run it via go migrate tool.
- Always do any database upgrade/maintenance or big migration during off hours. Always do it in lower env first to ensure it does not break the existing config.
  - Learned from my past mistake: I have learned this from my past experience, once i gave approval to do database config change in prd during low peak traffic hours and we saw a huge error/latency for few mins (error not connecting to database)

## Deployment

- Configure CI/CD pipelines
  - Every PR gets checked with lint, unit, and integration tests. When code merges to main, we build one Docker image tagged with the git SHA and push it. We deploy that same image to staging first, then to production via Helm.
  - Make prod deployment manual approval (after testing in lower envs).
  - Also, have a rollback pipeline setup to quickly revert if something breaks in production environment
- Configure slack notifiers after each deployment for each service.

## Observability and Traceability

- enable GC profiling in Datadog (or any other observability tool) so we can monitor OOM, CPU usage/time, deployment comparisons.
- enable Database monitoring for tracking availability, query performance, and resource usage.
  - we can monitor p99, p95 latencies to ensure we are not seeing abnormal spikes.
    - From my personal experience: if using aws rds, we can also use CloudWatch to monitor CPU usage for each database connection/replicas.
- Setup Database monitors, synthetic tests.

## Scalability and Reliability

- For production-level registration bursts, we could add rate limiting, request timeouts, metrics, logs.
  - example: I have added rate limiting when creating thousands of coupons.
- Tune the pgx pool: set `max_open_conn`/`max_idle_conns`.
  - max_open_conn (max open connections) protects our database from being overwhelmed and prevents service opening unlimited connections.
  - max_idle_conn (max idle connections) helps reuse warm connection instead of spinning a new connection for every request.
    - From my development experience, sweet configuration has been to keep the same value for both i.e., max_open_conn = max_idle_conn
- Add per-IP and per-email throttling/rate-limiting on `/signup` and `/login`, with load shedding (return HTTP 429) when bursts exceed capacity.
  - example: max 5 login attempts per minute for `user@example.com`, max 10 signups per minute from one IP.
- Horizontal scaling via Kubernetes - Run more copies/replicas of the service when traffic is high. Kubernetes can add or remove copies automatically based on CPU usage.
  - We can use Helm-style configuration that controls **autoscaling (via using variables like**  minReplicas/maxReplicas) — how Kubernetes automatically adds or removes copies (pods) of our service based on load.
    - NOTE: we can use other names besides minReplicas/maxReplicas - these are the names I am familiar with and haved used in microservices.
  - We can use `kubectl get pods -o wide` command to check how many pods are currently deployed for our service.
- Adding secrets folder to store password and other credentials (we can use sops tool for encrypt and decrypting the secret file)

## Caching and Performance

- If implementing cache (such as redis) we need to consider to prevent Cache stampede (example: implement SingleFlight logic) 
  - I have experienced this in our prod env in the past.
- bcrypt cost/CPU: revisit the existing note in `internal/app/`GC profiling`accounts.go` about checking email existence before hashing, and tune the bcrypt cost vs. throughput.

## Internal Tools

- We can build service to service testing tool for engineers to test in lower env.
  - If we need to test in prd env, we can add `dry_run` feature in our gRPC. When `dry_run` is enabled, we need to ensure we do not have any side effects (like manipulating database, triggering any events/metrics etc)
- Admin tool for non-eng stakeholders or operations team to manage customer issues.
  - example: lookup customer's account via email, force password reset, soft-delete an account. (some of these features are not implemented but just noting down some ideas on how we can leverage the admin tool)

## Authentication and Security

Please refer to `registration/docs/cursorlog/auth_security_section_e2c3d8ee.plan.md` doc for more info on this. 



## On-call incidents & Public Facing API

- Setup on-call escaltion policies via Pagerduty so we get alerts for any critical issue and eng can act upon it fast.
  - Always do a thorough post-mortem and act upon the action items.
- A status page to indicate if something failed in our service and root cause analysis so our customers are aware of any on-going issue.

## AI Agents Setup

- Add an `AGENTS.md` file at the repo root that tells AI agents how to work in this repo: how to build, run, and test the service, the project layout, and any conventions to follow. This keeps agents consistent and saves us from re-explaining the same things.
- Add an ignore file (like `.cursorignore`) to list files and folders agents should leave alone, such as secrets, generated code and database credentials. This keeps agents away from sensitive or risky files.
- Create an AI agent to help during on-call issues
  - helping writing/updating the runbooks.
  - triage any on-call issue by helping figure out root cause and suggest possible solution.

