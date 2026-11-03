# pprof against a real service

Endpoints sit behind auth, so fetch with curl and open the file. `go tool pprof` cannot
send headers. Adjust host, path and header to the service.

## CPU

```bash
curl -H "Authorization: prod-secret" -o cpu.prof "http://localhost:8080/metrics/pprof/profile?seconds=30"
```

Returned in under the requested seconds = error body, not a profile. Check with `cat cpu.prof`
(a profile is binary, several KB). Always sample while the load is on.

```bash
go tool pprof -http=:8000 cpu.prof
```

Views: Flame Graph (read widths, fix the widest first), Source (per-line flat/cum for one
function), Peek (caller -> callee edges with the call-site line).

```bash
go tool pprof -top -cum cpu.prof | head -40
go tool pprof -top -cum -focus="project-name" cpu.prof | head -40
go tool pprof -top -cum -nodefraction=0.05 cpu.prof | head -40
```

`top` = flat (time in the function itself), `top -cum` = cumulative (function plus everything
it calls). Interactive: `go tool pprof cpu.prof`, then `top -cum 30`, `peek funcName`,
`list funcName`.

### Compare before and after

```bash
go tool pprof -http=:8000 -base cpu.prof cpu-after-fix.prof
go tool pprof -base cpu.prof cpu-after-fix.prof
```

Green / negative = removed, red / positive = grown. `-base` compares absolute seconds, so when
throughput went up, per-response frames (JSON write, syscalls) show red although they did not
get slower per request. Options:

```bash
go tool pprof -http=:8000 -base cpu.prof -normalize cpu-after-fix.prof
```

scales the base to the same total, or re-run the after load capped at the old rps and profile
that. `-diff_base` only changes how percentages are reported (relative to the base total), it
does not scale.

## Memory

```bash
curl -H "Authorization: prod-secret" -o heap.prof "http://localhost:8080/metrics/pprof/heap?gc=1"
curl -H "Authorization: prod-secret" -o allocs.prof "http://localhost:8080/metrics/pprof/allocs"
go tool pprof -http=:8000 heap.prof
```

A heap profile describes the heap as of the last completed GC. Under light load or a big
GOGC headroom that GC can be minutes old, so two snapshots look identical or show garbage
that is already dead. `?gc=1` runs a GC first, then samples: every snapshot is current and
holds only live objects. Only the `heap` path honors it (`allocs` ignores it; `net/http/pprof`
checks `name == "heap"`). Cost: one forced GC cycle, fine for a hand-taken snapshot, not for
a scraper hitting it every few seconds.

Sample types (SAMPLE menu, or `-sample_index=`):

- `inuse_space` - bytes live at the last GC. The default for heap. What holds memory now (OOM
  hunting).
- `alloc_space` - bytes allocated since process start, freed or not. Allocation churn, GC
  pressure.
- `inuse_objects` / `alloc_objects` - same, counted in objects.

```bash
go tool pprof -sample_index=inuse_space -top heap.prof | head -30
go tool pprof -sample_index=alloc_space -top allocs.prof | head -30
```

Growth over time: take two `?gc=1` snapshots minutes apart, under the same steady load, and
diff them.

```bash
go tool pprof -http=:8000 -base heap-t0.prof heap-t5.prof
```

The frames that grow between snapshots are the retainers. Heap profile is a snapshot of the
last GC, not live RSS; RSS also holds freed-but-not-returned spans and non-heap memory
(stacks, cgo). `GODEBUG=gctrace=1` on the process prints every GC with heap sizes.

## Goroutines

```bash
curl -H "Authorization: prod-secret" -o goroutines.prof "http://localhost:8080/metrics/pprof/goroutine"
go tool pprof -http=:8000 goroutines.prof
```

Same two-snapshot method as heap: take one profile early in the load and one later, under
the same steady load, and diff them.

```bash
curl -H "Authorization: prod-secret" -o goroutines-t0.prof "http://localhost:8080/metrics/pprof/goroutine"
curl -H "Authorization: prod-secret" -o goroutines-t1.prof "http://localhost:8080/metrics/pprof/goroutine"

go tool pprof -http=:8000 -base goroutines-t0.prof goroutines-t1.prof

go tool pprof -top -base goroutines-t0.prof -focus="mymodule" goroutines-t1.prof | head -30
```

`-focus` keeps only the stacks that contain the regexp.

Sample value is a goroutine count, so the flame graph widths and the `flat` column are
goroutines. The stack that is positive in the diff is the leak; its top frames (`chansend`,
`selectgo`, `semacquire`, `netpollblock`) say what it is parked on, and the Source view
filtered by your module path gives the line. Bounded groups (worker pools, idle keepalive
connections, tickers) come out near zero. Divide the growth by requests served in the same
window to get leaked goroutines per request, which does not depend on the load level.

```bash
curl -H "Authorization: prod-secret" "http://localhost:8080/metrics/pprof/goroutine?debug=1"
curl -H "Authorization: prod-secret" "http://localhost:8080/metrics/pprof/goroutine?debug=2"
```

- `debug=1` - text, goroutines grouped by identical stack with a count per group. First
  stop for a leak: the group with the big count is the leak, its top frame is where they park.
- `debug=2` - full dump, one entry per goroutine with state (`chan receive`, `select`,
  `IO wait`, `semacquire`) and how long it has been blocked (`, 5 minutes`). Big output.

No `gc=1` here and none needed: the goroutine dump walks the live stacks at request time,
so it is always current. Finished goroutines are gone from it immediately, nothing waits for
a GC.

Leak check: two dumps minutes apart, the group whose count grows is the leak. Compare counts
with the `go_goroutines` metric over time.

## Block and mutex

Both are off by default. The service must call `runtime.SetBlockProfileRate(1)` (or a
sampling rate like 10000 ns) and `runtime.SetMutexProfileFraction(5)` at startup, otherwise
the profiles are empty.

```bash
curl -H "Authorization: prod-secret" -o block.prof "http://localhost:8080/metrics/pprof/block"
curl -H "Authorization: prod-secret" -o mutex.prof "http://localhost:8080/metrics/pprof/mutex"
go tool pprof -http=:8000 block.prof
```

Block = time goroutines waited on channels, selects, sync primitives. Mutex = time holders of
contended mutexes made others wait (points at the holder, not the waiter).

## Trace

```bash
curl -H "Authorization: prod-secret" -o trace.out "http://localhost:8080/metrics/pprof/trace?seconds=5"
go tool trace trace.out
```

Scheduler view, GC pauses, goroutine analysis, blocking profiles per goroutine. Use when the
question is latency or stalls rather than CPU share. Keep it to a few seconds, files get big.
