go tool pprof http://localhost:6060/debug/pprof/profile?seconds=30
# (pprof) top
# (pprof) list functionName
# (pprof) web
go tool pprof http://localhost:6060/debug/pprof/heap
go tool pprof http://localhost:6060/debug/pprof/goroutine

# block works with runtime.SetBlockProfileRate(1)
go tool pprof http://localhost:6060/debug/pprof/block

# Each of the functions above produce dump fine that can be used here
go tool pprof <file>
# or here
go tool pprof -http=:8080 <file>
