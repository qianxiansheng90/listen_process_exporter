// Package util
// @Description: 获取
package listen_process

import (
	"bytes"
	"context"
	"io/ioutil"
	"strings"
	"syscall"
)

const (
	LinuxProcNetTcpFile  = "/proc/net/tcp"
	LinuxProcNetTcp6File = "/proc/net/tcp6"
)

/*
 *  @Description: struct
 */
type ListenProcess struct {
	Pid  int32
	Port uint32
}

/*
 *  @Description: collect listen port
 */
func collectListenProcess(ctx context.Context) (processList map[uint32]ListenProcess, err error) {
	var lpArr map[uint32]ListenProcess
	processList = make(map[uint32]ListenProcess)
	// ipv4
	lpArr, err = getListenIPVxService(ctx, uint32(syscall.AF_INET), LinuxProcNetTcpFile, false)
	if err != nil {
		return
	}
	for k, v := range lpArr {
		processList[k] = v
	}
	// ipv6
	lpArr6, err := getListenIPVxService(ctx, uint32(syscall.AF_INET6), LinuxProcNetTcp6File, false)
	if err != nil {
		return lpArr, err
	}
	for k, v := range lpArr6 {
		processList[k] = v
	}
	return lpArr, err
}

/*
 *  @Description: read file and find listen port
 */
func getListenIPVxService(ctx context.Context, family uint32, file string, listen bool) (map[uint32]ListenProcess, error) {
	var lpArr = map[uint32]ListenProcess{}

	// Read the contents of the /proc file with a single read sys call.
	// This minimizes duplicates in the returned connections
	// For more info:
	// https://github.com/shirou/gopsutil/pull/361
	contents, err := ioutil.ReadFile(file)
	if err != nil {
		return lpArr, err
	}

	lines := bytes.Split(contents, []byte("\n"))
	// skip first line
	for _, line := range lines[1:] {
		var la, ra Addr
		l := strings.Fields(string(line))
		if len(l) < 10 {
			continue
		}
		lAddr := l[1]
		rAddr := l[2]
		pid := int32(0)

		if la, err = decodeAddress(family, lAddr); err != nil {
			continue
		}

		if ra, err = decodeAddress(family, rAddr); err != nil {
			continue
		}
		if listen {
			// 0.0.0.0 或者 :: 或者 ::1
			if ra.IP != "0.0.0.0" && strings.Trim(ra.IP, ":") != "" && strings.Trim(ra.IP, ":") != "1" {
				// 是连接不是监听端口
				continue
			}
		}
		lpArr[la.Port] = ListenProcess{
			Pid:  pid,
			Port: la.Port,
		}
	}

	return lpArr, nil

}
