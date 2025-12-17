curl -o cpu.prof http://localhost:6061/debug/pprof/profile?seconds=30
go tool pprof -http=:8000 cpu.prof

curl -o heap.prof http://localhost:6061/debug/pprof/heap
go tool pprof -http=:8000 heap.prof

# GOROUTINE
curl -o gr.prof http://localhost:6061/debug/pprof/goroutine && go tool pprof -http=:8000 gr.prof
