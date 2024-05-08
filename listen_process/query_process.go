// Package util
// @Description: 获取
package listen_process

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"strconv"
	"strings"
	"syscall"
)

const (
	LinuxProcDir         = "/proc"
	LinuxProcNetTcpFile  = "/proc/net/tcp"
	LinuxProcNetTcp6File = "/proc/net/tcp6"
)

/*
 *  @Description: struct
 */
type ListenProcess struct {
	Pid  int32  `json:"pid"`
	Port uint32 `json:"port"`
}

// PS:github.com/shirou/gopsutil
type inodeMap struct {
	pid int32
	fd  uint32
}

/*
 *  @Description: collect listen port
 */
func collectListenProcess(ctx context.Context) (processList map[uint32]ListenProcess, err error) {
	var lpArr map[uint32]ListenProcess
	processList = make(map[uint32]ListenProcess)
	// ipv4
	lpArr, err = getListenIPVxService(ctx, uint32(syscall.AF_INET), LinuxProcNetTcpFile, true)
	if err != nil {
		return
	}
	for k, v := range lpArr {
		processList[k] = v
	}
	// ipv6
	lpArr6, err := getListenIPVxService(ctx, uint32(syscall.AF_INET6), LinuxProcNetTcp6File, true)
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
	inodes, err := getProcInodesAll(ctx, LinuxProcDir, 0)
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
		inode := l[9]
		if la, err = decodeAddress(family, lAddr); err != nil {
			continue
		}

		if ra, err = decodeAddress(family, rAddr); err != nil {
			continue
		}
		i, exists := inodes[inode]
		if exists {
			pid = i[0].pid
		}

		if listen {
			// 0.0.0.0 or :: or ::1
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

func getProcInodesAll(ctx context.Context, root string, max int) (map[string][]inodeMap, error) {
	pids, err := PidsWithContext(ctx)
	if err != nil {
		return nil, err
	}
	ret := make(map[string][]inodeMap)

	for _, pid := range pids {
		t, err := getProcInodes(root, pid, max)
		if err != nil {
			// skip if permission error or no longer exists
			if os.IsPermission(err) || os.IsNotExist(err) || err == io.EOF {
				continue
			}
			return ret, err
		}
		if len(t) == 0 {
			continue
		}
		// TODO: update ret.
		ret = combineMap(ret, t)
	}
	return ret, nil
}

/*
 * @Description: 获取进程相关的inode信息
 * @Param root:
 * @Param pid:进程的pid
 * @Param max:
 * @Return map[string][]inodeMap:
 * @Return error:
 */
func getProcInodes(root string, pid int32, max int) (map[string][]inodeMap, error) {
	ret := make(map[string][]inodeMap)

	dir := fmt.Sprintf("%s/%d/fd", root, pid)
	f, err := os.Open(dir)
	if err != nil {
		return ret, err
	}
	defer f.Close()
	files, err := f.Readdir(max)
	if err != nil {
		return ret, err
	}
	for _, fd := range files {
		inodePath := fmt.Sprintf("%s/%d/fd/%s", root, pid, fd.Name())

		inode, err := os.Readlink(inodePath)
		if err != nil {
			continue
		}
		if !strings.HasPrefix(inode, "socket:[") {
			continue
		}
		// the process is using a socket
		l := len(inode)
		inode = inode[8 : l-1]
		_, ok := ret[inode]
		if !ok {
			ret[inode] = make([]inodeMap, 0)
		}
		fd, err := strconv.Atoi(fd.Name())
		if err != nil {
			continue
		}

		i := inodeMap{
			pid: pid,
			fd:  uint32(fd),
		}
		ret[inode] = append(ret[inode], i)
	}
	return ret, nil
}

/*
  - @Description:
    Pids retunres all pids.
    Note: this is a copy of process_linux.Pids()
    FIXME: Import process occures import cycle.
    move to common made other platform breaking. Need consider.
  - @Param ctx:
  - @Return []int32: 所有进程的pid列表
  - @Return error:
*/
func PidsWithContext(ctx context.Context) ([]int32, error) {
	var ret []int32

	d, err := os.Open(LinuxProcDir)
	if err != nil {
		return nil, err
	}
	defer d.Close()

	fnames, err := d.Readdirnames(-1)
	if err != nil {
		return nil, err
	}
	for _, fname := range fnames {
		pid, err := strconv.ParseInt(fname, 10, 32)
		if err != nil {
			// if not numeric name, just skip
			continue
		}
		ret = append(ret, int32(pid))
	}

	return ret, nil
}

/*
 * @Description: 合并map信息
 * @Param src: 源inode信息
 * @Param add: 需要增加的inode信息
 * @Return map[string][]inodeMap: 结果信息
 */
func combineMap(src map[string][]inodeMap, add map[string][]inodeMap) map[string][]inodeMap {
	for key, value := range add {
		a, exists := src[key]
		if !exists {
			src[key] = value
			continue
		}
		src[key] = append(a, value...)
	}
	return src
}
