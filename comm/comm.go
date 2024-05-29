/*
 * Copyright (c) 2011 Qunar.com. All Rights Reserved.
 * @Author: fangyuan.qian
 * @Create: 2024-04-29 18:55:53
 * @Description: desc
 */
package comm

import "log"

const (
	Version = "1.1.0"
)

var (
	debug = false
)

func SetDebug(de bool) {
	debug = de
	log.Printf("set debug = %v", de)
}

func Debug() bool {
	return debug
}
