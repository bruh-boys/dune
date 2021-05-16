
```
$ go test -bench .
goos: linux
goarch: amd64
pkg: github.com/dunelang/dune/bench
cpu: Intel(R) Core(TM) i7-8750H CPU @ 2.20GHz
BenchmarkClassProperty-12                5363065               219.7 ns/op            24 B/op          2 allocs/op
BenchmarkClassField-12                   6582004               174.9 ns/op             8 B/op          1 allocs/op
BenchmarkObjectField-12                  6686628               190.4 ns/op             8 B/op          1 allocs/op
BenchmarkGoField-12                    275856254               4.164 ns/op             0 B/op          0 allocs/op
BenchmarkClassMethod-12                  3841426               302.9 ns/op            56 B/op          3 allocs/op
BenchmarkObjectMethod-12                 4201852               272.5 ns/op            24 B/op          2 allocs/op
BenchmarkParallelHTTP_Go-12                93436               20570 ns/op          9975 B/op         94 allocs/op
BenchmarkParallelHTTP-12                   83007               20994 ns/op         10977 B/op        106 allocs/op
BenchmarkHTTP_Go-12                        20677               61458 ns/op          5579 B/op         66 allocs/op
BenchmarkHTTP-12                           17989               86281 ns/op          6495 B/op         78 allocs/op
BenchmarkCall-12                         3970718               295.9 ns/op            32 B/op          2 allocs/op
BenchmarkGoCall-12                      23561005               49.87 ns/op             0 B/op          0 allocs/op
BenchmarkLoop-12                           37652               30726 ns/op          3008 B/op        176 allocs/op
BenchmarkGoLoop-12                        232718                4952 ns/op             0 B/op          0 allocs/op
BenchmarkNoTailCall-12                    129741                9240 ns/op          1952 B/op         63 allocs/op
BenchmarkTailCall-12                      279379                4356 ns/op           184 B/op         21 allocs/op
PASS
ok      github.com/dunelang/dune/bench  24.050s
```