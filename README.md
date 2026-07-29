# national-defense-knowledge-quiz

High performance backend for national defense knowledge quiz consent.

## Deploy

First copy the env file

```bash
cp .env.production .env # or .env.example for local test
```

then generate problem and prize into database

```bash
docker compose run --rm app /app/seed
```

then run the docker container

```bash
docker compose up -d
```

The backend exposed on port 9090, postgres on port 5432 as default.

## Benchmark

There's a benchmark script under `./cmd/bench`.

Make sure the backend is up, then

```bash
go run ./cmd/bench   -scenario workflow -concurrency 10000 -requests 100000 --url http://localhost:9000
```

There's some options for the bench, to list them:

```bash
go run ./cmd/bench/ --help
```

On a laptop with `Intel(R) Core(TM) Ultra 9 185H (12+8+2) @ 5.10 GHz` CPU under `Fedora Linux 44 (KDE Plasma Desktop Edition) x86_64` system, a typical benchmark result is

```text
$ go run ./cmd/bench   -scenario workflow -concurrency 10000 -requests 100000 --url http://localhost:9000
2026/07/30 00:25:57 Benchmark config: scenario=workflow concurrency=10000 requests=100000
2026/07/30 00:25:57 Resetting session tables (type=sqlite url=./dev.db)...
2026/07/30 00:25:57 DB reset done.
2026/07/30 00:25:57 Starting load test...
2026/07/30 00:26:25 Load test completed in 27.47s

Endpoint            Count   OK      Errors  Timeouts  Min(ms)  P50(ms)  P95(ms)  P99(ms)  Max(ms)  RPS   SvP50  SvP99
--------            -----   --      ------  --------  -------  -------  -------  -------  -------  ---   -----  -----
GET /exam/info/     100000  100000  0       0         0.1      3.3      276.5    462.3    620.6    3641  0.0    0.0
POST /exam/start/   200000  200000  0       0         0.9      628.9    2146.4   3098.4   7363.2   7282  0.0    0.0
POST /exam/submit/  100000  100000  0       0         2.2      876.1    2428.8   3406.0   7624.6   3641  0.0    0.0

2026/07/30 00:26:25 Results saved to bench_results/results.json
```

The backend takes about 2GB memory.
