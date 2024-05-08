// Package listen_process
// @Description: listen process
package listen_process

import (
	"context"
	"errors"
	"log"
	"sync"
	"time"

	"listen_process_exporter/comm"
)

const (
	DefaultInterval = time.Second * 60
	ForbidInterval  = -1
)

var (
	refreshIntervalSecond = 60
	t                     = time.NewTicker(DefaultInterval)
	listenProcessCache    = map[uint32]ListenProcess{}
	lock                  = sync.RWMutex{}
	refreshTime           = time.Unix(0, 0)
)

/*
 *  @Description: set refresh ticker interval
 */
func SetTickerInterval(intervalSecond int) error {
	if intervalSecond == ForbidInterval {
		refreshIntervalSecond = ForbidInterval
		log.Printf("forbid refresh listen process interval second")
		return nil
	}
	if intervalSecond < 5 {
		return errors.New("invalid interval second")
	}
	t.Reset(time.Duration(intervalSecond) * time.Second)
	refreshIntervalSecond = intervalSecond
	log.Printf("set refresh interval second: %d", refreshIntervalSecond)
	return nil
}

func GetListenPortPid(listenPort uint32) (ListenProcess, error) {
	if v, exist := getListenProcessPid(listenPort); exist {
		return v, nil
	}
	// if not exist then refresh
	_, _ = RefreshListenProcess(context.Background())
	if v, exist := listenProcessCache[listenPort]; exist {
		return v, nil
	}
	return ListenProcess{}, errors.New("listen_process not found")
}

/*
 *  @Description: refresh listen process interval
 */
func RefreshListenProcessGoroutine() {
	if refreshIntervalSecond == ForbidInterval {
		log.Printf("will not start refresh listen process goroutine")
		log.Printf("refresh listen process by restart process or through http request")
		return
	}
	log.Printf("start refresh listen process goroutine")
	_, _ = RefreshListenProcess(context.Background())
	for {
		select {
		case <-t.C:
			_, _ = RefreshListenProcess(context.Background())
		}
	}
}

func RefreshListenProcess(ctx context.Context) (listenProcess map[uint32]ListenProcess, err error) {
	if comm.Debug() {
		log.Printf("check refresh listen process")
	}
	if time.Now().Sub(refreshTime).Seconds() < float64(refreshIntervalSecond) {
		return
	}
	if listenProcess, err = collectListenProcess(ctx); err == nil {
		resetListenProcessCache(listenProcess)
	}
	if comm.Debug() {
		log.Printf("refresh listen process success")
	}
	return
}

func resetListenProcessCache(cache map[uint32]ListenProcess) {
	lock.Lock()
	defer lock.Unlock()
	listenProcessCache = cache
	if comm.Debug() {
		for k, v := range cache {
			log.Printf("found listen port %d pid %d  ", k, v.Pid)
		}
	}
}

/*
 *  @Description: get listen process pid
 */
func getListenProcessPid(listenPort uint32) (p ListenProcess, exist bool) {
	lock.RLock()
	defer lock.RUnlock()
	p, exist = listenProcessCache[listenPort]
	return
}
