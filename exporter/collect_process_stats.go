// Package exporter
// @Description: collect process stat
package exporter

import (
	"context"
	"log"
	"os"
	"path/filepath"
	"strconv"

	linuxproc "github.com/c9s/goprocinfo/linux"
)

const (
	LinuxProcDir = "/proc"
)

type ProcessStats struct {
	ProcessID     int32                       `json:"pid"` // 进程id
	Status        *linuxproc.ProcessStatus    `json:"status"`
	Statm         *linuxproc.ProcessStatm     `json:"statm"`
	Stat          *linuxproc.ProcessStat      `json:"stat"`
	IO            *linuxproc.ProcessIO        `json:"io"`              // io 信息
	Schedule      *linuxproc.ProcessSchedStat `json:"schedule"`        // 调度信息
	FileDescCount int                         `json:"file_desc_count"` // 打开的文件列表
	Cmdline       string                      `json:"cmdline"`
}

/*
 *  @Description: collect process stats include stats
 */
func collectProcessStat(ctx context.Context, pid int32) (processStats ProcessStats, err error) {
	var (
		p        = filepath.Join(LinuxProcDir, strconv.FormatInt(int64(pid), 10))
		io       *linuxproc.ProcessIO
		stat     *linuxproc.ProcessStat
		statm    *linuxproc.ProcessStatm
		status   *linuxproc.ProcessStatus
		cmdline  string
		schedule *linuxproc.ProcessSchedStat
		names    []string
	)

	if _, err = os.Stat(p); err != nil {
		return
	}

	if io, err = linuxproc.ReadProcessIO(filepath.Join(p, "io")); err != nil {
		io = &linuxproc.ProcessIO{}
	}
	if stat, err = linuxproc.ReadProcessStat(filepath.Join(p, "stat")); err != nil {
		stat = &linuxproc.ProcessStat{}
	}
	if statm, err = linuxproc.ReadProcessStatm(filepath.Join(p, "statm")); err != nil {
		statm = &linuxproc.ProcessStatm{}
	}
	if status, err = linuxproc.ReadProcessStatus(filepath.Join(p, "status")); err != nil {
		status = &linuxproc.ProcessStatus{}
	}
	// not used
	//if schedule, err = linuxproc.ReadProcessSchedStat(filepath.Join(p, "schedstat")); err != nil {
	//	schedule = &linuxproc.ProcessSchedStat{}
	//}
	if names, err = fileDescriptors(filepath.Join(p, "fd")); err != nil {
		log.Printf("read fd: %v", err)
	}

	processStats = ProcessStats{
		ProcessID:     pid,
		Status:        status,
		Statm:         statm,
		Stat:          stat,
		IO:            io,
		Schedule:      schedule,
		Cmdline:       cmdline,
		FileDescCount: len(names),
	}
	return
}

func fileDescriptors(dirPath string) ([]string, error) {
	d, err := os.Open(dirPath)
	if err != nil {
		return nil, err
	}
	defer d.Close()

	names, err := d.Readdirnames(-1)
	if err != nil {
		return nil, err
	}

	return names, nil
}
