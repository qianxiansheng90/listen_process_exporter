// Package exporter
// @Description:exporter
package exporter

import (
	"context"
	"log"
	"strconv"

	"github.com/prometheus/client_golang/prometheus"
	"listen_process_exporter/listen_process"
)

const (
	listenPort     = "listen_port"
	listProcessPID = "pid"
	// See https://github.com/prometheus/procfs/blob/master/proc_stat.go for details on userHZ.
	userHZ = 100
)

type Exporter struct {
	listenPort          int
	collectChildProcess bool
	listenProcess       map[uint32]listen_process.ListenProcess
	debug               bool
}

var (
	numThreadDesc = prometheus.NewDesc(
		"listen_port_process_thread_count",
		"number of threads in listen process",
		[]string{listenPort, listProcessPID}, nil)

	cpuSecsDesc = prometheus.NewDesc(
		"listen_port_process_cpu_seconds_total",
		"Cpu user usage in seconds",
		[]string{listenPort, listProcessPID, "mode"}, nil)

	readBytesDesc = prometheus.NewDesc(
		"listen_port_process_read_bytes_total",
		"number of bytes read by this process",
		[]string{listenPort, listProcessPID}, nil)

	readCallsDesc = prometheus.NewDesc(
		"listen_port_process_read_calls_total",
		"number of calls read by this process",
		[]string{listenPort, listProcessPID}, nil)

	writeBytesDesc = prometheus.NewDesc(
		"listen_port_process_write_bytes_total",
		"number of bytes written by this process",
		[]string{listenPort, listProcessPID}, nil)

	writeCallsDesc = prometheus.NewDesc(
		"listen_port_process_write_calls_total",
		"number of calls written by this process",
		[]string{listenPort, listProcessPID}, nil)

	majorPageFaultsDesc = prometheus.NewDesc(
		"listen_port_process_major_page_faults_total",
		"Major page faults",
		[]string{listenPort, listProcessPID}, nil)

	minorPageFaultsDesc = prometheus.NewDesc(
		"listen_port_process_minor_page_faults_total",
		"Minor page faults",
		[]string{listenPort, listProcessPID}, nil)

	contextSwitchesDesc = prometheus.NewDesc(
		"listen_port_process_context_switches_total",
		"Context switches",
		[]string{listenPort, listProcessPID, "ctx_switch_type"}, nil)

	memBytesDesc = prometheus.NewDesc(
		"listen_port_process_memory_bytes",
		"number of bytes of memory in use",
		[]string{listenPort, listProcessPID, "memory_type"}, nil)

	openFDsDesc = prometheus.NewDesc(
		"listen_port_process_open_file_desc",
		"number of open file descriptors for this group",
		[]string{listenPort, listProcessPID}, nil)

	startTimeDesc = prometheus.NewDesc(
		"listen_port_process_oldest_start_time_seconds",
		"start time in seconds since 1970/01/01 of listen process",
		[]string{listenPort, listProcessPID}, nil)

	scrapeErrorsDesc = prometheus.NewDesc(
		"listen_port_process_scrape_errors",
		"general scrape errors: no proc metrics collected during a cycle",
		nil, nil)

	scrapeProcReadErrorsDesc = prometheus.NewDesc(
		"listen_port_process_scrape_procread_errors",
		"incremented each time a proc's metrics collection fails",
		nil, nil)

	scrapePartialErrorsDesc = prometheus.NewDesc(
		"listen_port_process_scrape_partial_errors",
		"incremented each time a tracked proc's metrics collection fails partially, e.g. unreadable I/O stats",
		nil, nil)
)

func NewExporter(collectChildProcess bool, listenPort int) *Exporter {
	return &Exporter{
		listenPort:          listenPort,
		collectChildProcess: collectChildProcess,
		listenProcess:       make(map[uint32]listen_process.ListenProcess),
	}
}

func (e *Exporter) Describe(ch chan<- *prometheus.Desc) {
	ch <- cpuSecsDesc
	ch <- numThreadDesc
	ch <- readBytesDesc
	ch <- readCallsDesc
	ch <- writeBytesDesc
	ch <- writeCallsDesc
	ch <- memBytesDesc
	ch <- openFDsDesc
	ch <- startTimeDesc
	ch <- majorPageFaultsDesc
	ch <- minorPageFaultsDesc
	ch <- contextSwitchesDesc
	ch <- scrapeErrorsDesc
	ch <- scrapeProcReadErrorsDesc
	ch <- scrapePartialErrorsDesc
}

func (e *Exporter) Collect(ch chan<- prometheus.Metric) {
	listenProcess, err := listen_process.GetListenPortPid(uint32(e.listenPort))
	if err != nil {
		log.Printf("query listen port %d error: %v", e.listenPort, err)
		return
	}
	if listenProcess.Pid == 0 {
		log.Printf("not found listen port %d pid", e.listenPort)
		return
	}
	processStats, err := collectProcessStat(context.Background(), listenProcess.Pid)
	if err != nil {
		log.Printf("query listen port %d pid %d error: %v", e.listenPort, listenProcess.Pid, err)
		return
	}
	ch <- prometheus.MustNewConstMetric(startTimeDesc,
		prometheus.GaugeValue, float64(processStats.Stat.Starttime),
		listenPortToString(listenProcess.Port), listenProcessPIDToString(listenProcess.Pid))

	ch <- prometheus.MustNewConstMetric(numThreadDesc,
		prometheus.GaugeValue, float64(processStats.Stat.NumThreads),
		listenPortToString(listenProcess.Port), listenProcessPIDToString(listenProcess.Pid))

	ch <- prometheus.MustNewConstMetric(cpuSecsDesc,
		prometheus.CounterValue, float64(processStats.Stat.Utime)/userHZ,
		listenPortToString(listenProcess.Port), listenProcessPIDToString(listenProcess.Pid), "user")
	ch <- prometheus.MustNewConstMetric(cpuSecsDesc,
		prometheus.CounterValue, float64(processStats.Stat.Stime)/userHZ,
		listenPortToString(listenProcess.Port), listenProcessPIDToString(listenProcess.Pid), "system")

	ch <- prometheus.MustNewConstMetric(memBytesDesc,
		prometheus.GaugeValue, float64(processStats.Status.VmRSS),
		listenPortToString(listenProcess.Port), listenProcessPIDToString(listenProcess.Pid), "resident")
	ch <- prometheus.MustNewConstMetric(memBytesDesc,
		prometheus.GaugeValue, float64(processStats.Status.VmSize),
		listenPortToString(listenProcess.Port), listenProcessPIDToString(listenProcess.Pid), "virtual")
	ch <- prometheus.MustNewConstMetric(memBytesDesc,
		prometheus.GaugeValue, float64(processStats.Status.VmSwap),
		listenPortToString(listenProcess.Port), listenProcessPIDToString(listenProcess.Pid), "swapped")

	ch <- prometheus.MustNewConstMetric(readBytesDesc,
		prometheus.CounterValue, float64(processStats.IO.ReadBytes),
		listenPortToString(listenProcess.Port), listenProcessPIDToString(listenProcess.Pid))
	ch <- prometheus.MustNewConstMetric(readCallsDesc,
		prometheus.CounterValue, float64(processStats.IO.Syscr),
		listenPortToString(listenProcess.Port), listenProcessPIDToString(listenProcess.Pid))

	ch <- prometheus.MustNewConstMetric(writeBytesDesc,
		prometheus.CounterValue, float64(processStats.IO.WriteBytes),
		listenPortToString(listenProcess.Port), listenProcessPIDToString(listenProcess.Pid))
	ch <- prometheus.MustNewConstMetric(writeCallsDesc,
		prometheus.CounterValue, float64(processStats.IO.Syscw),
		listenPortToString(listenProcess.Port), listenProcessPIDToString(listenProcess.Pid))

	ch <- prometheus.MustNewConstMetric(majorPageFaultsDesc,
		prometheus.CounterValue, float64(processStats.Stat.Majflt),
		listenPortToString(listenProcess.Port), listenProcessPIDToString(listenProcess.Pid))
	ch <- prometheus.MustNewConstMetric(minorPageFaultsDesc,
		prometheus.CounterValue, float64(processStats.Stat.Minflt),
		listenPortToString(listenProcess.Port), listenProcessPIDToString(listenProcess.Pid))

	ch <- prometheus.MustNewConstMetric(contextSwitchesDesc,
		prometheus.CounterValue, float64(processStats.Status.NonvoluntaryCtxtSwitches),
		listenPortToString(listenProcess.Port), listenProcessPIDToString(listenProcess.Pid), "voluntary")
	ch <- prometheus.MustNewConstMetric(contextSwitchesDesc,
		prometheus.CounterValue, float64(processStats.Status.VoluntaryCtxtSwitches),
		listenPortToString(listenProcess.Port), listenProcessPIDToString(listenProcess.Pid), "nonvoluntary")

	ch <- prometheus.MustNewConstMetric(openFDsDesc,
		prometheus.GaugeValue, float64(processStats.Status.FDSize),
		listenPortToString(listenProcess.Port), listenProcessPIDToString(listenProcess.Pid))

}

func listenPortToString(p uint32) string {
	return strconv.FormatInt(int64(p), 10)
}

func listenProcessPIDToString(p int32) string {
	return strconv.FormatInt(int64(p), 10)
}
