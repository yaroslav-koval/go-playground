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

Both are off by default. Set them near application start, for example inside `main`. The
two settings are explained below; each of them can add latency, so be careful with the
values. The starting numbers come from the
[Datadog Go profiling guide](https://github.com/DataDog/go-profiler-notes/blob/main/guide/README.md).
If production is under high load, avoid profiling there: use these settings
locally, where aggressive values are affordable and give more precise profiles.

```go
// /metrics/pprof/block.
// 0 = off (default).
// Prod: start with 100_000_000 (100 ms), go lower if needed.
// Too low: higher CPU/latency; precise profile, rare waits visible.
// Too high: no CPU/latency overhead; short and rare waits lost in sampling.
runtime.SetBlockProfileRate(100_000)

// /metrics/pprof/mutex.
// 0 = off (default).
// Prod: start with 100, go lower if needed.
// Too low: higher CPU/latency on contended mutexes; precise profile, rarely contended mutexes visible.
// Too high: no CPU/latency overhead; rarely contended mutexes lost in sampling.
runtime.SetMutexProfileFraction(100)

```

### Block profile (SetBlockProfileRate)

**What it is.** For every function that got parked, it keeps two totals: how many times it
waited and for how long. More specifically, the sample names are `contentions` and `delay`,
both visible in pprof. Parked means waiting on:

- a channel blocking operation (send, receive, `select`)
- a mutex lock (`Mutex.Lock`, `RWMutex.Lock`, `RWMutex.RLock`)
- a wait primitive (`WaitGroup.Wait`, `Cond.Wait`)

Not sleep, not IO.

**How it is measured.** At the moment the goroutine is unblocked, the runtime calculates
`waitedFor` (duration from parked to now) and compares it with the SetBlockProfileRate
argument (`rate`), both in nanoseconds:

```go
if waitedFor >= rate {
contentions += 1
delay += waitedFor
} else if randN(rate) < waitedFor {
contentions += rate / waitedFor
delay += rate
}
```

Each capture walks the goroutine's stack, which is the cost.

**Example.** `rate` = 100 ns.

- `waitedFor` is 3 ms, so `waitedFor >= rate` is true. Captured every time:
  `contentions += 1`, `delay += 3 ms`.
- `waitedFor` is 20 ns, so `waitedFor >= rate` is false. `randN(100)` gives a random value
  from 0 to 99; it is below 20 in 20 cases out of 100, so the wait is captured 1 time in 5.
  When captured: `contentions += 5` (100 / 20), `delay += 100 ns` (`rate`). Five such waits
  produce on average one capture worth 5 contentions and 100 ns, the same totals as
  counting all five.

**Noise.** That total is correct only with many captures: with rate 100 ms and 1 ms waits
you capture 1 in 100, and a lock that waited 50 times most likely shows nothing, or one
record of 100 ms where the truth is 50 ms (50 waits of 1 ms). Noise shrinks roughly as 1/sqrt(captures): 100
captures is about 10%, 10 000 about 1%. So **one rate does not fit every lock**: a hot lock
with thousands of waits a second is measured fine at a high rate, and the high rate is
what keeps its overhead low; a rare lock needs a low rate, or 1 for a minute of test, to
be seen at all.

**Prod values.** Datadog used to call 10_000 safe, then saw up to 4% CPU overhead in
production and now says start at 100_000_000 (100 ms) and lower only if needed. My vision:
both Datadog's values are correct, but for different scenarios.
It's better to leave production without it, reproduce the load locally or on staging.
If needed permanent on prod, start with higher number (100ms) and lower carefully.

**Trick for prod.** To temporarily observe block profile, inside /block endpoint add `SetBlockProfileRate(N)`,
then `defer SetBlockProfileRate(0)`, wait for X seconds and return profile. N and X may be in HTTP query.
The same works for /mutex with SetMutexProfileFraction.

### Mutex profile (SetMutexProfileFraction)

**What it is.** For every function that released a mutex while other goroutines were
waiting for it, it keeps two totals: how many times that happened and how long **the waiters**
had waited. More specifically, the sample names are `contentions` and `delay`, both
visible in pprof. The profile reads "the code holding this lock made others wait 40 s in
total". It covers only `sync.Mutex` and `sync.RWMutex`. Not channels, not `WaitGroup`,
not `Cond`: those are block profile only.

**How it is measured.** At the moment `Unlock()` runs and finds goroutines waiting, the
runtime calculates `waitedFor` (time **all the waiters** sat parked, summed over all of them, not
only the one being woken) and rolls against the SetMutexProfileFraction argument (`N`):

```go
if randN(N) == 0 {
contentions += N
delay += N * waitedFor
}
```

The stack recorded is the `Unlock()` call with its callers. Each capture walks that
stack, which is the cost.

**Example.** `N` = 100. One mutex with 5 000 contended unlocks (with 1+ waiters), each
releasing waiters that sat 20 us in total. `randN(100) == 0` is true with probability
1/100, which gives about 50 captures out of 5 000 unlocks. Per capture:
`contentions += 100`, `delay += 100 * 20 us` = 2 ms. Shown: 5 000 contentions, 100 ms.
Truth: 5 000 contentions, 100 ms.

**Noise.** Same rule as for the block profile: totals are right only with many captures,
noise shrinks roughly as 1/sqrt(captures). A mutex contended fewer than N times in the
window is missing or misreported; rarely contended mutexes need a small N.

**Prod values.** Same rule as for the block profile, with 100 as the starting value: one
capture per 100 contended unlocks, and a contended unlock is already a rare event next to
a wait. Lower it carefully, and only when the mutex you hunt does not show up. The same
"Trick for prod" applies here, with `SetMutexProfileFraction` inside the /mutex endpoint.

### Capturing profiles

```bash
# block
curl -H "Authorization: prod-secret" -o block-t1.prof "http://localhost:8080/metrics/pprof/block"
# mutex
curl -H "Authorization: prod-secret" -o mutex-t1.prof "http://localhost:8080/metrics/pprof/mutex"
```

Both profiles are cumulative, like `allocs`: the counters start when the rate is set and
only grow, there is no reset. A single fetch therefore covers everything since then, quiet
periods included. To look at one window, fetch twice under the same load and diff them:

```bash
go tool pprof -http=:8000 -base block-t1.prof block-t2.prof

go tool pprof -http=:8000 -base mutex-t1.prof mutex-t2.prof
```

### Reading them

Both profiles have two sample types, switched in the `SAMPLE` menu: `delay` (default) and
`contentions`. Divide delay by contentions to get the average wait per event, e.g.
361 s / 5676 = 64 ms. The views
below are in the `VIEW` menu.

Block profile answers "who waits, and on what".
- `Flame Graph`: the general picture, the widest stacks are the biggest waits. Skip system
  noise: `net/http.(*persistConn).writeLoop` and `readLoop`, worker pools, tickers, those
  are idle waits.
- `Source`: filter by your module path from go.mod and look for the lines with the most
  time, or the most time per event (`delay` / `contentions`).

Mutex profile answers "who held the lock while others waited".
- `Flame Graph`: same picture from the holder side; the function above
  `sync.(*Mutex).Unlock` is the holder, the frames below are the request path.
- `Source`: same filter; with `defer mu.Unlock()` the time sits on the `return` line the
  holder took, so it shows which branch of the function held the lock.

How they complete each other: block names the line that waited but not the holder, who
was running, not waiting. Mutex names the holder but not who suffered. Read them together:
block gives the waiting line, mutex gives the holding function. The code between that
function's `Lock` and `Unlock` is what to shorten or move out of the lock.

Remember that the mutex profile does not explain blocks on channels, `select`, `WaitGroup`
or `Cond`: for those the block profile is the only source.

### Trace for lock contention

Use `go tool trace` when you know which lock it is but need the timing: how long one
request holds it, whether holders queue up one after another, what the other goroutines do
meanwhile. Fetch `/metrics/pprof/trace?seconds=5` under load, run `go tool trace trace.out`.
"Goroutine analysis" groups goroutines by the function they started in; click the
`net/http.(*conn).serve` row, every request goroutine starts there, and read per goroutine
how much was execution, sync block (lock or channel), network wait, scheduler wait.
"Synchronization blocking profile" is the same kind of data as the block profile, taken
from the trace: only these 5 s, every wait recorded, no rate needed. "View trace" rows are
processors, not goroutines; under lock contention you see few goroutines running at a time
with gaps between them while the Goroutines counter at the top stays high.

## Trace

```bash
curl -H "Authorization: prod-secret" -o trace.out "http://localhost:8080/metrics/pprof/trace?seconds=5"
go tool trace trace.out
```

Scheduler view, GC pauses, goroutine analysis, blocking profiles per goroutine. Use when the
question is latency or stalls rather than CPU share. Keep it to a few seconds, files get big.
