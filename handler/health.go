/*
 * Copyright (c) 2011 Qunar.com. All Rights Reserved.
 * @Author: fangyuan.qian
 * @Create: 2024-07-02 15:01:04
 * @Description: desc
 */
package handler

import (
	"encoding/json"
	"net/http"
	"os"
)

const (
	binVersion = "1.0"
)

type HTTPResponse struct {
	Code string `json:"code"`
	Msg  string `json:"msg"`
	Data any    `json:"data"`
}

var (
	hostName, _ = os.Hostname()
)

func Health() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var data, _ = json.Marshal(HTTPResponse{
			Code: "success",
			Msg:  binVersion,
			Data: struct {
				Hostname string `json:"hostname"`
			}{
				Hostname: hostName,
			},
		})
		w.Write(data)
	}
}
