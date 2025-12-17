# go-musthave-metrics-tpl

В ходе профилирования было выявлено, что существенная доля аллокаций приходится на gzip middleware
Для оптимизации был реализован sync.Pool для переиспользования gzip.Writer.
При сравнении профилей по alloc_space абсолютные значения увеличились, что объясняется накопительным характером метрики и различиями в длительности и объёме нагрузки.
При этом профиль подтверждает использование sync.Pool (sync.(*Pool).Get), а оптимизация снижает частоту повторных аллокаций при стабильной нагрузке.

# Результат первого профилирования
PS D:\prog\YaPracticum\Go dev\go-metrics-agent> go tool pprof .\base.pprof  
File: server.exe
Build ID: C:\Users\dimag\AppData\Local\Temp\go-build2616133121\b001\exe\server.exe2025-12-17 13:21:08.2381047 +0300 MSK
Type: inuse_space
Time: 2025-12-17 13:21:37 MSK
Entering interactive mode (type "help" for commands, "o" for options)
(pprof) top
Showing nodes accounting for 5911.76kB, 100% of 5911.76kB total
Showing top 10 nodes out of 41
      flat  flat%   sum%        cum   cum%
 1805.17kB 30.54% 30.54%  1805.17kB 30.54%  compress/flate.NewWriter (inline)
    1539kB 26.03% 56.57%     1539kB 26.03%  runtime.allocm
  518.65kB  8.77% 65.34%   518.65kB  8.77%  time.map.init.1
  512.56kB  8.67% 74.01%   512.56kB  8.67%  html/template.map.init.3
  512.31kB  8.67% 82.68%   512.31kB  8.67%  net.newFD
  512.05kB  8.66% 91.34%   512.05kB  8.66%  os/signal.NotifyContext.func1
  512.02kB  8.66%   100%   512.02kB  8.66%  runtime.gcBgMarkWorker
         0     0%   100%  1805.17kB 30.54%  compress/gzip.(*Writer).Close
         0     0%   100%  1805.17kB 30.54%  compress/gzip.(*Writer).Write
         0     0%   100%  1805.17kB 30.54%  github.com/go-chi/chi/v5.(*Mux).Mount.func1

# Результат профилирования после оптимизации
go tool pprof .\result.pprof
File: server.exe
Build ID: C:\Users\dimag\AppData\Local\go-build\35\354544259272550212bff69910e0d679af6474d6dbd55e7448451e6106d40169-d\server.exe2025-12-17 12:16:15.5139309 +0300 MSK
Type: inuse_space
Time: 2025-12-17 13:20:37 MSK
Entering interactive mode (type "help" for commands, "o" for options)
(pprof) top
Showing nodes accounting for 7235.26kB, 100% of 7235.26kB total
Showing top 10 nodes out of 49
      flat  flat%   sum%        cum   cum%
 3610.34kB 49.90% 49.90%  4155.01kB 57.43%  compress/flate.NewWriter (inline)
  544.67kB  7.53% 57.43%   544.67kB  7.53%  compress/flate.(*compressor).init
  518.65kB  7.17% 64.60%   518.65kB  7.17%  time.map.init.1
     513kB  7.09% 71.69%      513kB  7.09%  runtime.allocm
  512.31kB  7.08% 78.77%   512.31kB  7.08%  net.newFD
  512.22kB  7.08% 85.85%   512.22kB  7.08%  runtime.malg
  512.05kB  7.08% 92.92%   512.05kB  7.08%  runtime.acquireSudog
  512.02kB  7.08%   100%   512.02kB  7.08%  net/textproto.NewReader
         0     0%   100%  4155.01kB 57.43%  compress/gzip.(*Writer).Close
         0     0%   100%  4155.01kB 57.43%  compress/gzip.(*Writer).Write

# Сравнение двух профилей по памяти
go tool pprof -top -diff_base=./profiles/base.pprof ./profiles/result.pprof             
File: server.exe
Build ID: C:\Users\dimag\AppData\Local\go-build\e3\e3ff4dbd33bde64f30b0d902a01e841f1d0ca68fb01347e38c6575d32c57234a-d\server.exe2025-12-17 10:33:36.3706621 +0300 MSK
Type: inuse_space
Time: 2025-12-17 11:50:51 MSK
Showing nodes accounting for 14019.96kB, 100% of 14019.96kB total
      flat  flat%   sum%        cum   cum%
 6318.10kB 45.07% 45.07%  6318.10kB 45.07%  compress/flate.NewWriter (inline)
    3591kB 25.61% 70.68%     3591kB 25.61%  runtime.allocm
 1037.31kB  7.40% 78.08%  1037.31kB  7.40%  time.map.init.1
  513.12kB  3.66% 81.74%   513.12kB  3.66%  reflect.growslice
  512.31kB  3.65% 85.39%   512.31kB  3.65%  net.newFD (inline)
  512.05kB  3.65% 89.04%   512.05kB  3.65%  os/signal.NotifyContext.func1
  512.05kB  3.65% 92.70%   512.05kB  3.65%  runtime.acquireSudog
  512.01kB  3.65% 96.35%   512.01kB  3.65%  internal/syscall/windows.errnoErr (inline)
  512.01kB  3.65%   100%   512.01kB  3.65%  runtime.(*timers).addHeap
         0     0%   100%   512.01kB  3.65%  bufio.(*Reader).Peek
         0     0%   100%   512.01kB  3.65%  bufio.(*Reader).fill
         0     0%   100%  6318.10kB 45.07%  compress/gzip.(*Writer).Close
         0     0%   100%  6318.10kB 45.07%  compress/gzip.(*Writer).Write
         0     0%   100%   513.12kB  3.66%  encoding/json.(*decodeState).array
         0     0%   100%   513.12kB  3.66%  encoding/json.(*decodeState).unmarshal
         0     0%   100%   513.12kB  3.66%  encoding/json.(*decodeState).value
         0     0%   100%   513.12kB  3.66%  encoding/json.Unmarshal
         0     0%   100%  6831.23kB 48.73%  github.com/go-chi/chi/v5.(*Mux).Mount.func1
         0     0%   100%  6831.23kB 48.73%  github.com/go-chi/chi/v5.(*Mux).ServeHTTP
         0     0%   100%  6831.23kB 48.73%  github.com/go-chi/chi/v5.(*Mux).routeHTTP
         0     0%   100%  3610.34kB 25.75%  github.com/makimaki04/go-metrics-agent.git/internal/compress.ReleaseWriter
         0     0%   100%   513.12kB  3.66%  github.com/makimaki04/go-metrics-agent.git/internal/handler.(*Handler).UpdateMetricBatch
         0     0%   100%   512.01kB  3.65%  internal/poll.(*FD).Read
         0     0%   100%   512.01kB  3.65%  internal/poll.execIO
         0     0%   100%   512.01kB  3.65%  internal/syscall/windows.WSAGetOverlappedResult
         0     0%   100%  6831.23kB 48.73%  main.main.func1.4.GzipMiddleware.1
         0     0%   100%  6831.23kB 48.73%  main.main.func1.4.WithLogging.2
         0     0%   100%   512.31kB  3.65%  main.main.func3
         0     0%   100%   512.31kB  3.65%  net.(*ListenConfig).Listen
         0     0%   100%   512.01kB  3.65%  net.(*conn).Read
         0     0%   100%   512.01kB  3.65%  net.(*netFD).Read
         0     0%   100%   512.31kB  3.65%  net.(*sysListener).listenMPTCP
         0     0%   100%   512.31kB  3.65%  net.(*sysListener).listenTCP (inline)
         0     0%   100%   512.31kB  3.65%  net.(*sysListener).listenTCPProto
         0     0%   100%   512.31kB  3.65%  net.Listen
         0     0%   100%   512.31kB  3.65%  net.internetSocket
         0     0%   100%   512.31kB  3.65%  net.socket
         0     0%   100%   512.31kB  3.65%  net/http.(*Server).ListenAndServe
         0     0%   100%  7343.23kB 52.38%  net/http.(*conn).serve
         0     0%   100%   512.01kB  3.65%  net/http.(*connReader).Read
         0     0%   100%  6831.23kB 48.73%  net/http.HandlerFunc.ServeHTTP
         0     0%   100%  6831.23kB 48.73%  net/http.serverHandler.ServeHTTP
         0     0%   100%   513.12kB  3.66%  reflect.Value.Grow
         0     0%   100%   513.12kB  3.66%  reflect.Value.grow
         0     0%   100%   512.01kB  3.65%  runtime.(*scavengerState).sleep
         0     0%   100%   512.01kB  3.65%  runtime.(*timer).maybeAdd
         0     0%   100%   512.01kB  3.65%  runtime.(*timer).modify
         0     0%   100%   512.01kB  3.65%  runtime.(*timer).reset (inline)
         0     0%   100%   512.01kB  3.65%  runtime.bgscavenge
         0     0%   100%   512.05kB  3.65%  runtime.chanrecv
         0     0%   100%   512.05kB  3.65%  runtime.chanrecv1
         0     0%   100%  1037.31kB  7.40%  runtime.doInit (inline)
         0     0%   100%  1037.31kB  7.40%  runtime.doInit1
         0     0%   100%  1037.31kB  7.40%  runtime.main
         0     0%   100%     2565kB 18.30%  runtime.mcall
         0     0%   100%     1026kB  7.32%  runtime.mstart
         0     0%   100%     1026kB  7.32%  runtime.mstart0
         0     0%   100%     1026kB  7.32%  runtime.mstart1
         0     0%   100%     3591kB 25.61%  runtime.newm
         0     0%   100%     2565kB 18.30%  runtime.park_m
         0     0%   100%     3591kB 25.61%  runtime.resetspinning
         0     0%   100%     3591kB 25.61%  runtime.schedule
         0     0%   100%     3591kB 25.61%  runtime.startm
         0     0%   100%   512.05kB  3.65%  runtime.unique_runtime_registerUniqueMapCleanup.func2
         0     0%   100%     3591kB 25.61%  runtime.wakep
         0     0%   100%  1037.31kB  7.40%  time.init