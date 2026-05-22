[toc]

### 场景：没有 redis/mq，点击短链接时同步写入日志时

```
环境：
- CPU: Intel i7-14700HX
- RAM: 64GB DDR5 5200MT/s
- OS: Windows 11
- 工具: wrk，参数 -t8 -c200 -d60s (并发数已通过二分，得出峰值并发为 220)。其中 P50/P95/P99 通过 latency + 自定义 lua 实现
- db: PostgreSQL 18，数据库仅有一个链接与一个用户

总结：
- RPS: 940
- P50/P95/P99: 205.27ms/267.82ms/309.41ms


详细结果：
Running 1m test @ http://172.28.192.1:58000/api/v1/s/V1T1YV
  8 threads and 200 connections
  Thread Stats   Avg      Stdev     Max   +/- Stdev
    Latency   212.21ms   28.72ms 488.52ms   77.98%
    Req/Sec   117.95     24.65   210.00     69.12%
  56452 requests in 1.00m, 10.12MB read
Requests/sec:    940.09
Transfer/sec:    172.60KB
P50: 205.27 ms
P95: 267.82 ms
P99: 309.41 ms

SELECT calls, rows, round(mean_exec_time::numeric, 3) AS mean_ms,
       round(total_exec_time::numeric, 2) AS total_ms,
       query
FROM pg_stat_statements
WHERE query ILIKE '%links%'
   OR query ILIKE '%click_log%'
ORDER BY calls DESC
LIMIT 20;
 calls | rows  | mean_ms |  total_ms  |  query
-------+-------+---------+------------+--------------------------------------------------------------------------------------------------------------------------------------
 55847 | 55847 |   0.024 |    1325.59 | SELECT original_url FROM links WHERE code = $1 AND deleted_at IS NULL AND status = $2 AND (expires_at IS NULL OR expires_at > NOW())
 55699 | 55699 |  28.625 | 1594363.57 | UPDATE links
                         +
       |       |         |            | SET view_count = view_count + $2, updated_at = NOW()                                 
                         +
       |       |         |            | WHERE code = $1                                                                       
                         +
       |       |         |            |         AND deleted_at IS NULL                                                       
                         +
       |       |         |            |         AND status = $3
 55697 | 55697 |   0.020 |    1139.92 | INSERT INTO click_log(code, ip, user_agent, referer, clicked_at)                     
                         +
       |       |         |            | VALUES ($1, $2, $3, $4, $5)
(3 rows)
```



### 场景：引入 mq 异步写入数据库日志时

```
环境：
- CPU: Intel i7-14700HX
- RAM: 64GB DDR5 5200MT/s
- OS: Windows 11
- 工具: wrk，参数 -t8 -c200 -d60s (并发数已通过二分，得出峰值并发为 220)。其中 P50/P95/P99 通过 latency + 自定义 lua 实现
- db: PostgreSQL 18，数据库仅有一个链接与一个用户

结果：
- RPS: 31859
- P50/P95/P99: 5.88ms/10.52ms/13.26ms

详细结果：
Running 1m test @ http://172.28.192.1:58000/api/v1/s/y6p7O0
  8 threads and 200 connections
  Thread Stats   Avg      Stdev     Max   +/- Stdev
    Latency     6.34ms    2.92ms 108.39ms   85.68%
    Req/Sec     4.00k   413.90     5.40k    70.54%
  1912885 requests in 1.00m, 342.96MB read
Requests/sec:  31859.27
Transfer/sec:      5.71MB
P50: 5.88 ms
P95: 10.52 ms
P99: 13.26 ms
```



### 场景：通过加上 Header 构造浏览器预获取, 测试不调用 event_id + publisher 时

```
环境：
- CPU: Intel i7-14700HX
- RAM: 64GB DDR5 5200MT/s
- OS: Windows 11
- 工具: wrk，参数 -t8 -c200 -d60s -H "Sec-Purpose: prefetch"
- db: PostgreSQL 18，数据库仅有一个链接与一个用户

结果：
- RPS: 37083
- P50/P95/P99: 5.09ms/8.89ms/11.49ms

结论：
event_id + publisher 的调用成本大约会使 RPS 减少 5000 左右

详细结果：
Running 1m test @ http://172.28.192.1:58000/api/v1/s/MCGlpg
  8 threads and 200 connections
  Thread Stats   Avg      Stdev     Max   +/- Stdev
    Latency     5.44ms    2.26ms  89.33ms   83.32%
    Req/Sec     4.66k   460.98     5.87k    70.69%
  2225649 requests in 1.00m, 399.04MB read
Requests/sec:  37083.53
Transfer/sec:      6.65MB
P50: 5.09 ms
P95: 8.89 ms
P99: 11.49 ms
```



### 场景：未加入空值缓存时，访问固定不存在短链接时

```
环境：
- CPU: Intel i7-14700HX
- RAM: 64GB DDR5 5200MT/s
- OS: Windows 11
- 工具: wrk，参数 -t8 -c200 -d60s
- db: PostgreSQL 18，数据库为空

结果：
- RPS: 33906
- P50/P95/P99: 5.58ms/9.68ms/12.34ms
- PG Calls: 2035521

详细结果：
Running 1m test @ http://172.28.192.1:58000/api/v1/s/123456
  8 threads and 200 connections
  Thread Stats   Avg      Stdev     Max   +/- Stdev
    Latency     5.98ms    2.98ms 112.81ms   89.91%
    Req/Sec     4.26k   414.85     5.53k    75.44%
  2035331 requests in 1.00m, 333.86MB read
  Non-2xx or 3xx responses: 2035331
Requests/sec:  33906.64
Transfer/sec:      5.56MB
P50: 5.58 ms
P95: 9.68 ms
P99: 12.34 ms

  calls  | rows | mean_ms | total_ms |                                           query
---------+------+---------+----------+-------------------------------------------------------------------------------------------
 2035521 |    0 |   0.026 | 53770.88 | SELECT original_url, status, expires_at FROM links WHERE code = $1 AND deleted_at IS NULL

```



### 场景：未加入空值缓存时，访问随机不存在短链接时

```
环境：
- CPU: Intel i7-14700HX
- RAM: 64GB DDR5 5200MT/s
- OS: Windows 11
- 工具: wrk，参数 -t8 -c200 -d60s
- db: PostgreSQL 18，数据库为空

结果：
- RPS: 32246
- P50/P95/P99: 5.86ms/10.18ms/12.98ms
- PG Calls: 1935700

详细结果：
Running 1m test @ http://172.28.192.1:58000
  8 threads and 200 connections
  Thread Stats   Avg      Stdev     Max   +/- Stdev
    Latency     6.22ms    2.30ms 125.98ms   77.51%
    Req/Sec     4.05k   465.10     5.55k    68.33%
  1935507 requests in 1.00m, 317.49MB read
  Non-2xx or 3xx responses: 1935507
Requests/sec:  32246.70
Transfer/sec:      5.29MB
P50: 5.86 ms
P95: 10.18 ms
P99: 12.98 ms

  calls  | rows | mean_ms | total_ms |                                           query
---------+------+---------+----------+-------------------------------------------------------------------------------------------
 1935700 |    0 |   0.027 | 52151.49 | SELECT original_url, status, expires_at FROM links WHERE code = $1 AND deleted_at IS NULL
(1 row)
```



### 场景：加入空值缓存时，访问单一不存在短链接时

```
环境：
- CPU: Intel i7-14700HX
- RAM: 64GB DDR5 5200MT/s
- OS: Windows 11
- 工具: wrk，参数 -t8 -c200 -d60s
- db: PostgreSQL 18，数据库为空

结果：
- RPS: 38026
- P50/P95/P99: 4.99ms/8.55ms/11.22ms
- PG Calls: 23
- Redis Keyspace Hits: 2282479
- Redis Keyspace Misses: 23
- Redis 缓存命中率: 99.99%

详细结果：
Running 1m test @ http://172.28.192.1:58000/api/v1/s/123456
  8 threads and 200 connections
  Thread Stats   Avg      Stdev     Max   +/- Stdev
    Latency     5.31ms    2.16ms  84.16ms   83.70%
    Req/Sec     4.78k   435.91     5.81k    72.58%
  2282310 requests in 1.00m, 374.37MB read
  Non-2xx or 3xx responses: 2282310
Requests/sec:  38026.70
Transfer/sec:      6.24MB
P50: 4.99 ms
P95: 8.55 ms
P99: 11.22 ms

 calls | rows | mean_ms | total_ms |                                           query
-------+------+---------+----------+-------------------------------------------------------------------------------------------
    23 |    0 |   0.018 |     0.42 | SELECT original_url, status, expires_at FROM links WHERE code = $1 AND deleted_at IS NULL
    
keyspace_hits:2282479
keyspace_misses:23
```



### 场景：加入空值缓存时，访问随机不存在短链接时

```
环境：
- CPU: Intel i7-14700HX
- RAM: 64GB DDR5 5200MT/s
- OS: Windows 11
- 工具: wrk，参数 -t8 -c200 -d60s
- db: PostgreSQL 18，数据库为空

结果：
- RPS: 24398
- P50/P95/P99: 8.07ms/11.52ms/13.60ms
- PG Calls: 1323000
- Redis Keyspace Hits: 141755
- Redis Keyspace Misses: 1323000
- Redis 缓存命中率：9.68%

结论：
加入空值缓存时，由于链路更长，如果是随机短链接访问时反而 RPS 会降低且拦截效果也减弱了
```

详细结果：

```bash
> wrk -t8 -c200 -d60s -s random-code.lua http://172.28.192.1:58000

Running 1m test @ http://172.28.192.1:58000
  8 threads and 200 connections
  Thread Stats   Avg      Stdev     Max   +/- Stdev
    Latency     8.23ms    2.62ms 133.27ms   83.42%
    Req/Sec     3.07k   224.22     5.14k    77.19%
  1464571 requests in 1.00m, 240.24MB read
  Non-2xx or 3xx responses: 1464571
Requests/sec:  24398.86
Transfer/sec:      4.00MB
P50: 8.07 ms
P95: 11.52 ms
P99: 13.60 ms
```

```postgresql
CREATE EXTENSION IF NOT EXISTS pg_stat_statements;
SELECT pg_stat_statements_reset();

SELECT calls, rows, round(mean_exec_time::numeric, 3) AS mean_ms,
       round(total_exec_time::numeric, 2) AS total_ms,
       query
FROM pg_stat_statements
WHERE query ILIKE '%links%'
   OR query ILIKE '%click_log%'
ORDER BY calls DESC
LIMIT 20;

  calls  | rows | mean_ms | total_ms |                                           query
---------+------+---------+----------+-------------------------------------------------------------------------------------------
 1323000 |    0 |   0.027 | 35080.54 | SELECT original_url, status, expires_at FROM links WHERE code = $1 AND deleted_at IS NULL
```

```redis
CONFIG RESETSTAT
INFO stats

keyspace_hits:141755
keyspace_misses:1323000
```



### 场景：加入布隆过滤器时，访问固定不存在短链接时

```
环境：
- CPU: Intel i7-14700HX
- RAM: 64GB DDR5 5200MT/s
- OS: Windows 11
- 工具: wrk，参数 -t8 -c200 -d60s
- db: PostgreSQL 18，数据库中有 100 个链接与一个用户
- 布隆过滤器配置: 容纳量 100'000，错误率 1%，占用大约 117KB 内存

结果：
- RPS: 39005
- P50/P95/P99: 4.86ms/8.92ms/12.12ms
- PG Calls: 0
- Redis Keyspace Hits: 0
- Redis Keyspace Misses: 0
```

详细结果：

```bash
> wrk -t8 -c200 -d60s -s done.lua http://172.28.192.1:58000/api/v1/s/123456

Running 1m test @ http://172.28.192.1:58000/api/v1/s/123456
  8 threads and 200 connections
  Thread Stats   Avg      Stdev     Max   +/- Stdev
    Latency     5.20ms    2.45ms 113.92ms   82.58%
    Req/Sec     4.90k   825.27     6.99k    68.15%
  2340878 requests in 1.00m, 383.98MB read
  Socket errors: connect 0, read 0, write 0, timeout 1
  Non-2xx or 3xx responses: 2340878
Requests/sec:  39005.20
Transfer/sec:      6.40MB
P50: 4.86 ms
P95: 8.92 ms
P99: 12.12 ms
```

```postgresql
CREATE EXTENSION IF NOT EXISTS pg_stat_statements;
SELECT pg_stat_statements_reset();

SELECT calls, rows, round(mean_exec_time::numeric, 3) AS mean_ms,
       round(total_exec_time::numeric, 2) AS total_ms,
       query
FROM pg_stat_statements
WHERE query ILIKE '%links%'
   OR query ILIKE '%click_log%'
ORDER BY calls DESC
LIMIT 20;

 calls | rows | mean_ms | total_ms | query
-------+------+---------+----------+-------
(0 rows)
```

```redis
CONFIG RESETSTAT
INFO stats

keyspace_hits:0
keyspace_misses:0
```



### 场景：加入布隆过滤器时，访问随机不存在短链接时

```
环境：
- CPU: Intel i7-14700HX
- RAM: 64GB DDR5 5200MT/s
- OS: Windows 11
- 工具: wrk，参数 -t8 -c200 -d60s
- db: PostgreSQL 18，数据库中有 100 个链接与一个用户
- 布隆过滤器配置: 容纳量 100'000，错误率 1%，占用大约 117KB 内存

结果：
- RPS: 40903
- P50/P95/P99: 4.62ms/8.42ms/11.23ms
- PG Calls: 0
- Redis Keyspace Hits: 0
- Redis Keyspace Misses: 0

结论：
添加上布隆过滤器后，缓存穿透问题得到有效解决，且成本较低
```

详细结果：

```bash
> wrk -t8 -c200 -d60s -s random-code.lua http://172.28.192.1:58000

Running 1m test @ http://172.28.192.1:58000
  8 threads and 200 connections
  Thread Stats   Avg      Stdev     Max   +/- Stdev
    Latency     4.93ms    2.19ms 117.39ms   80.84%
    Req/Sec     5.14k   776.77     7.34k    67.12%
  2454833 requests in 1.00m, 402.67MB read
  Non-2xx or 3xx responses: 2454833
Requests/sec:  40903.59
Transfer/sec:      6.71MB
P50: 4.62 ms
P95: 8.42 ms
P99: 11.23 ms
```

```postgresql
CREATE EXTENSION IF NOT EXISTS pg_stat_statements;
SELECT pg_stat_statements_reset();

SELECT calls, rows, round(mean_exec_time::numeric, 3) AS mean_ms,
       round(total_exec_time::numeric, 2) AS total_ms,
       query
FROM pg_stat_statements
WHERE query ILIKE '%links%'
   OR query ILIKE '%click_log%'
ORDER BY calls DESC
LIMIT 20;

 calls | rows | mean_ms | total_ms | query
-------+------+---------+----------+-------
(0 rows)
```

```redis
CONFIG RESETSTAT
INFO stats

keyspace_hits:0
keyspace_misses:0
```



### 场景：将缓存 TTL 设为 1s 时，模拟缓存击穿，访问已存在链接

```
环境：
- CPU: Intel i7-14700HX
- RAM: 64GB DDR5 5200MT/s
- OS: Windows 11
- 工具: wrk，参数 -t8 -c200 -d60s，为了防止 mq 写数据库影响 calls 查询，构造了 prefetch 请求跳过 mq 写（mq 写所产生的性能开销上面已测试过）
- db: PostgreSQL 18，数据库中有 1 个链接与一个用户
- redis 链接缓存 TTL: 1s

结果：
- RPS: 35614
- P50/P95/P99: 5.32ms/9.28ms/12.29ms
- PG Calls: 5171
- Redis Keyspace Hits: 2132684
- Redis Keyspace Misses: 5171
- Redis 缓存命中率: 99.7%
```

详细结果：

```bash
> wrk -t8 -c200 -d60s -s done.lua -H "Sec-Purpose: prefetch" http://172.28.192.1:58000/api/v1/s/aDmOAA

Running 1m test @ http://172.28.192.1:58000/api/v1/s/aDmOAA
  8 threads and 200 connections
  Thread Stats   Avg      Stdev     Max   +/- Stdev
    Latency     5.67ms    2.42ms  88.93ms   83.86%
    Req/Sec     4.47k   612.59     6.24k    67.67%
  2137678 requests in 1.00m, 383.27MB read
Requests/sec:  35614.44
Transfer/sec:      6.39MB
P50: 5.32 ms
P95: 9.28 ms
P99: 12.29 ms
```

```postgresql
CREATE EXTENSION IF NOT EXISTS pg_stat_statements;
SELECT pg_stat_statements_reset();

SELECT calls, rows, round(mean_exec_time::numeric, 3) AS mean_ms,
       round(total_exec_time::numeric, 2) AS total_ms,
       query
FROM pg_stat_statements
WHERE query ILIKE '%links%'
   OR query ILIKE '%click_log%'
ORDER BY calls DESC
LIMIT 20;

 calls | rows | mean_ms | total_ms |                                           query                                    
-------+------+---------+----------+-------------------------------------------------------------------------------------------
  5171 | 5171 |   0.034 |   178.18 | SELECT original_url, status, expires_at FROM links WHERE code = $1 AND deleted_at IS NULL
```

```redis
CONFIG RESETSTAT
INFO stats

keyspace_hits:2132684
keyspace_misses:5171
```



### 场景：将缓存 TTL 设为 1s 时，模拟缓存击穿，加上 singleflight 访问已存在链接时

```
环境：
- CPU: Intel i7-14700HX
- RAM: 64GB DDR5 5200MT/s
- OS: Windows 11
- 工具: wrk，参数 -t8 -c200 -d60s，为了防止 mq 写数据库影响 calls 查询，构造了 prefetch 请求跳过 mq 写（mq 写所产生的性能开销上面已测试过）
- db: PostgreSQL 18，数据库中有 1 个链接与一个用户
- redis 链接缓存 TTL: 1s

结果：
- RPS: 30101
- P50/P95/P99: 6.24ms/11.41ms/15.19ms
- PG Calls: 60
- Redis Keyspace Hits: 1801113
- Redis Keyspace Misses: 6333
- Redis 缓存命中率: 99.6%

结论：
虽然 RPS 降低了 5000 左右，但这是在 TTL 1s 的刻意制造极端缓存击穿场景导致的影响，但换来了 PG 回源数的很大程度降低，当 TTL 拉到 30min 时影响可以忽略。
```

```bash
> wrk -t8 -c200 -d60s -s done.lua -H "Sec-Purpose: prefetch" http://172.28.192.1:58000/api/v1/s/aDmOAA

Running 1m test @ http://172.28.192.1:58000/api/v1/s/aDmOAA
  8 threads and 200 connections
  Thread Stats   Avg      Stdev     Max   +/- Stdev
    Latency     6.72ms    2.99ms 115.67ms   82.19%
    Req/Sec     3.78k   699.61     5.67k    68.85%
  1807129 requests in 1.00m, 324.00MB read
Requests/sec:  30101.20
Transfer/sec:      5.40MB
P50: 6.24 ms
P95: 11.41 ms
P99: 15.19 ms
```

```postgresql
CREATE EXTENSION IF NOT EXISTS pg_stat_statements;
SELECT pg_stat_statements_reset();

SELECT calls, rows, round(mean_exec_time::numeric, 3) AS mean_ms,
       round(total_exec_time::numeric, 2) AS total_ms,
       query
FROM pg_stat_statements
WHERE query ILIKE '%links%'
   OR query ILIKE '%click_log%'
ORDER BY calls DESC
LIMIT 20;

 calls | rows | mean_ms | total_ms |                                           query                                    
-------+------+---------+----------+-------------------------------------------------------------------------------------------
    60 |   60 |   0.060 |     3.62 | SELECT original_url, status, expires_at FROM links WHERE code = $1 AND deleted_at IS NULL
```

```redis
CONFIG RESETSTAT
INFO stats

keyspace_hits:1801113
keyspace_misses:6333
```

