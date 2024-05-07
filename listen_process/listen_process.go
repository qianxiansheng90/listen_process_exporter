// Package listen_process
// @Description: listen process
package listen_process

import (
	"context"
	"errors"
	"sync"
	"time"
)

const (
	DefaultInterval = time.Second * 60
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
	if intervalSecond < 1 {
		return errors.New("invalid interval second")
	}
	t.Reset(time.Duration(intervalSecond) * time.Second)
	refreshIntervalSecond = intervalSecond
	return nil
}

func GetListenPortPid(listenPort uint32) (ListenProcess, error) {
	if v, exist := getListenProcessPid(listenPort); exist {
		return v, nil
	}
	// if not exist then refresh
	refreshListenProcess(context.Background())
	if v, exist := listenProcessCache[listenPort]; exist {
		return v, nil
	}
	return ListenProcess{}, errors.New("listen_process not found")
}

/*
 *  @Description: refresh listen process interval
 */
func RefreshListenProcessGoroutine() {
	for {
		select {
		case <-t.C:
			refreshListenProcess(context.Background())
		}
	}
}

func refreshListenProcess(ctx context.Context) {
	if time.Now().Sub(refreshTime).Seconds() < float64(refreshIntervalSecond) {
		return
	}
	if listenProcess, err := collectListenProcess(ctx); err == nil {
		resetListenProcessCache(listenProcess)
	}
	return
}

func resetListenProcessCache(cache map[uint32]ListenProcess) {
	lock.Lock()
	defer lock.Unlock()
	listenProcessCache = cache
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
