/*
 * Copyright (c) 2011 Qunar.com. All Rights Reserved.
 * @Author: fangyuan.qian
 * @Create: 2024-04-30 10:38:58
 * @Description: desc
 */
package listen_process

import (
	"encoding/hex"
	"fmt"
	"net"
	"strconv"
	"strings"
	"syscall"
)

// Addr is implemented compatibility to psutil
type Addr struct {
	IP   string `json:"ip"`
	Port uint32 `json:"port"`
}

// decodeAddress decode addresse represents addr in proc/net/*
// ex:
// "0500000A:0016" -> "10.0.0.5", 22
// "0085002452100113070057A13F025401:0035" -> "2400:8500:1301:1052:a157:7:154:23f", 53
func decodeAddress(family uint32, src string) (Addr, error) {
	t := strings.Split(src, ":")
	if len(t) != 2 {
		return Addr{}, fmt.Errorf("does not contain port, %s", src)
	}
	addr := t[0]
	port, err := strconv.ParseInt("0x"+t[1], 0, 64)
	if err != nil {
		return Addr{}, fmt.Errorf("invalid port, %s", src)
	}
	decoded, err := hex.DecodeString(addr)
	if err != nil {
		return Addr{}, fmt.Errorf("decode error, %s", err)
	}
	var ip net.IP
	// Assumes this is little_endian
	if family == syscall.AF_INET {
		ip = net.IP(Reverse(decoded))
	} else { // IPv6
		ip, err = parseIPv6HexString(decoded)
		if err != nil {
			return Addr{}, err
		}
	}
	return Addr{
		IP:   ip.String(),
		Port: uint32(port),
	}, nil
}

// Reverse reverses array of tool_bytes.
func Reverse(s []byte) []byte {
	for i, j := 0, len(s)-1; i < j; i, j = i+1, j-1 {
		s[i], s[j] = s[j], s[i]
	}
	return s
}

// parseIPv6HexString parse array of tool_bytes to IPv6 string
func parseIPv6HexString(src []byte) (net.IP, error) {
	if len(src) != 16 {
		return nil, fmt.Errorf("invalid IPv6 string")
	}

	buf := make([]byte, 0, 16)
	for i := 0; i < len(src); i += 4 {
		r := Reverse(src[i : i+4])
		buf = append(buf, r...)
	}
	return net.IP(buf), nil
}
