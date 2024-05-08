# listen_process_exporter-exporter
Prometheus exporter that mines /proc to report on selected processes.
Support collect listen port process ,such as mysql、postgresql、nginx、ssh and so on.

## Attention
**Only support Linux System**


## Running

Usage:

```
  listen_process_exporter -collector.port=3306
```

## Configure
Refresh listen process through  cron(default 60s) or http request. restart process will refresh listen process too.

#### http request
```http request
curl 'http://127.0.0.1:9911/refresh_listen_process'
```

#### cron
```shell
listen_process_exporter -collector.refresh=30
```

## Metrics


All these metrics start with `listen_port_process_` and have at minimum
the label `listen_port` and `pid`.

### cpu_seconds_total counter

CPU usage based on /proc/[pid]/stat fields utime(14) and stime(15) i.e. user and system time. This is similar to the node\_exporter's `node_cpu_seconds_total`.

### read_bytes_total counter

Bytes read based on /proc/[pid]/io field read_bytes.  The man page
says

> Attempt to count the number of bytes which this process really did cause to be fetched from the storage layer.  This is accurate for block-backed filesystems.

but I would take it with a grain of salt.

As `/proc/[pid]/io` are set by the kernel as read only to the process' user (see #137), to get these values you should run `process-exporter` either as that user or as `root`. Otherwise, we can't read these values and you'll get a constant 0 in the metric.

### write_bytes_total counter

Bytes written based on /proc/[pid]/io field write_bytes.  As with
read_bytes, somewhat dubious.  May be useful for isolating which processes
are doing the most I/O, but probably not measuring just how much I/O is happening.

### major_page_faults_total counter

Number of major page faults based on /proc/[pid]/stat field majflt(12).

### minor_page_faults_total counter

Number of minor page faults based on /proc/[pid]/stat field minflt(10).

### context_switches_total counter

Number of context switches based on /proc/[pid]/status fields voluntary_ctxt_switches
and nonvoluntary_ctxt_switches.  The extra label `ctxswitchtype` can have two values:
`voluntary` and `nonvoluntary`.

### memory_bytes gauge

Number of bytes of memory used.  The extra label `memtype` can have three values:

*resident*: Field rss(24) from /proc/[pid]/stat, whose doc says:

> This is just the pages which count toward text, data, or stack space.  This does not include pages which have not been demand-loaded in, or which are swapped out.

*virtual*: Field vsize(23) from /proc/[pid]/stat, virtual memory size.

*swapped*: Field VmSwap from /proc/[pid]/status, translated from KB to bytes.

If gathering smaps file is enabled, two additional values for `memtype` are added:

*proportionalResident*: Sum of "Pss" fields from /proc/[pid]/smaps, whose doc says:

> The "proportional set size" (PSS) of a process is the count of pages it has
> in memory, where each page is divided by the number of processes sharing it.

*proportionalSwapped*: Sum of "SwapPss" fields from /proc/[pid]/smaps

### open_filedesc gauge

Number of file descriptors, based on counting how many entries are in the directory
/proc/[pid]/fd.


### num_threads gauge

Sum of number of threads of all process in the group.  Based on field num_threads(20)
from /proc/[pid]/stat.



## Building

Requires Go 1.13 installed.
```
go build
```

## Exposing metrics through HTTP
Running
```
$ ./listen_process_exporter -collector.refresh=60 &
$ curl http://localhost:9911/metrics | grep listen_port_process

# TYPE listen_port_process_context_switches_total counter
listen_port_process_context_switches_total{ctx_switch_type="nonvoluntary",listen_port="3306",pid="438332"} 134
listen_port_process_context_switches_total{ctx_switch_type="voluntary",listen_port="3306",pid="438332"} 54
# HELP listen_port_process_cpu_seconds_total Cpu user usage in seconds
# TYPE listen_port_process_cpu_seconds_total counter
listen_port_process_cpu_seconds_total{listen_port="3306",mode="system",pid="438332"} 311.72
listen_port_process_cpu_seconds_total{listen_port="3306",mode="user",pid="438332"} 309.01
# HELP listen_port_process_major_page_faults_total Major page faults
# TYPE listen_port_process_major_page_faults_total counter
listen_port_process_major_page_faults_total{listen_port="3306",pid="438332"} 542
# HELP listen_port_process_memory_bytes number of bytes of memory in use
# TYPE listen_port_process_memory_bytes gauge
listen_port_process_memory_bytes{listen_port="3306",memory_type="resident",pid="438332"} 240848
listen_port_process_memory_bytes{listen_port="3306",memory_type="swapped",pid="438332"} 146172
listen_port_process_memory_bytes{listen_port="3306",memory_type="virtual",pid="438332"} 1.8172e+06
# HELP listen_port_process_minor_page_faults_total Minor page faults
# TYPE listen_port_process_minor_page_faults_total counter
listen_port_process_minor_page_faults_total{listen_port="3306",pid="438332"} 108105
# HELP listen_port_process_oldest_start_time_seconds start time in seconds since 1970/01/01 of listen process
# TYPE listen_port_process_oldest_start_time_seconds gauge
listen_port_process_oldest_start_time_seconds{listen_port="3306",pid="438332"} 3.21104591e+08
# HELP listen_port_process_open_file_desc number of open file descriptors for this group
# TYPE listen_port_process_open_file_desc gauge
listen_port_process_open_file_desc{listen_port="3306",pid="438332"} 256
# HELP listen_port_process_read_bytes_total number of bytes read by this process
# TYPE listen_port_process_read_bytes_total counter
listen_port_process_read_bytes_total{listen_port="3306",pid="438332"} 5.2322304e+07
# HELP listen_port_process_read_calls_total number of calls read by this process
# TYPE listen_port_process_read_calls_total counter
listen_port_process_read_calls_total{listen_port="3306",pid="438332"} 982
# HELP listen_port_process_thread_count number of threads in listen process
# TYPE listen_port_process_thread_count gauge
listen_port_process_thread_count{listen_port="3306",pid="438332"} 39
# HELP listen_port_process_write_bytes_total number of bytes written by this process
# TYPE listen_port_process_write_bytes_total counter
listen_port_process_write_bytes_total{listen_port="3306",pid="438332"} 1.9570688e+07
# HELP listen_port_process_write_calls_total number of calls written by this process
# TYPE listen_port_process_write_calls_total counter
listen_port_process_write_calls_total{listen_port="3306",pid="438332"} 158
```

## Thanks
This project is inspired by [process-exporter](https://github.com/ncabatoff/process-exporter).