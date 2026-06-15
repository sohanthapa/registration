# Productionalization

A checklist of items to consider when taking this service to production.

*NOTE*: The sections/steps below are to the best of my knowledge. I'm open to more ideas or feedback from the team if something doesn't look right, something is missing, or there are better approaches.

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
- Setup Database monitors, synthetic tests

## Scalability and Reliability

- For production-level registration bursts, we could add rate limiting, request timeouts, metrics, logs.
  - example: I have added rate limiting when creating thousands of coupons.
- Tune the pgx pool: set `max_open_conn`/`max_idle_conns`.
  - max_open_conn (max open connections) protects our database from being overwhelmed and prevents service opening unlimited connections.
  - max_idle_conn (max idle connections) helps reuse warm connection instead of spinning a new connection for every request.
  - From my development experience, sweet configuration has been to keep the same value for both i.e., max_open_conn = max_idle_conn
- Add per-IP and per-email throttling on `/signup` and `/login`, with load shedding (return HTTP 429) when bursts exceed capacity.
- Run more copies/replicas of the service when traffic is high. Kubernetes can add or remove copies automatically based on CPU usage.
  - Helm-style configuration block that controls **autoscaling** — how Kubernetes automatically adds or removes copies (pods) of our service based on load.
    - minReplicas/maxReplicas
  - We can use `kubectl get pods -o wide` command to check how many pods are currently deployed for our service.
- Adding secrets folder to store password and other credentials (we can use sops tool for encrypt and decrypting the secret file)

## Caching and Performance

- If implementing cache (such as redis) we need to consider to prevent Cache stampede (example: implement SingleFlight logic) 
  - I have experienced this in our prod env in the past.
- bcrypt cost/CPU: revisit the existing note in `internal/app/accounts.go` about checking email existence before hashing, and tune the bcrypt cost vs. throughput.



## Internal Tools

- We can build service to service testing tool for engineers to test in lower env.
  - If we need to test in prd env, we can add `dry_run` feature in our gRPC. When `dry_run` is enabled, we need to ensure we do not have any side effects (like manipulating database, triggering any events/metrics etc)
- Admin tool for non-eng stakeholders or operations team to manage customer issues.
  - example: lookup customer's account via email, force password reset, soft-delete an account. (some of these features are not implemented but just noting down some ideas on how we can leverage the admin tool)

### Admin panel features

### Authentication

## AI Agents Setup

- Add an `AGENTS.md` file at the repo root that tells AI agents how to work in this repo: how to build, run, and test the service, the project layout, and any conventions to follow. This keeps agents consistent and saves us from re-explaining the same things.
- Add an ignore file (like `.cursorignore`) to list files and folders agents should leave alone, such as secrets, generated code and database credentials. This keeps agents away from sensitive or risky files.
- Create an AI agent to help during on-call issues
  - helping writing/updating the runbooks.
  - triage any on-call issue by helping figure out root cause and suggest possible solution.

