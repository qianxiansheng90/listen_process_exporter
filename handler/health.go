/*
 * Copyright (c) 2011 Qunar.com. All Rights Reserved.
 * @Author: fangyuan.qian
 * @Create: 2024-07-02 15:01:04
 * @Description: desc
 */
package handler

import (
	"crypto/md5"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"

	"listen_process_exporter/comm"
)

type HTTPResponse struct {
	Code string `json:"code"`
	Msg  string `json:"msg"`
	Data any    `json:"data"`
}

var (
	hostName, _  = os.Hostname()
	fileChecksum = getFileChecksum()
)

func getFileChecksum() string {
	path, _ := exec.LookPath(os.Args[0])
	abs, _ := filepath.Abs(path)
	checksum, _, _ := checksumFile(abs)
	return checksum
}

func checksumFile(filePath string) (fileMD5 string, size int64, err error) {
	var ioErr error
	f, err := os.OpenFile(filePath, os.O_RDONLY, 0600)
	if err != nil {
		return
	}

	defer f.Close()

	md5hash := md5.New()
	if size, ioErr = io.Copy(md5hash, f); ioErr != nil {
		err = ioErr
		return
	}

	fileMD5 = fmt.Sprintf("%x", md5hash.Sum(nil))
	err = nil
	return
}
func Health() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var data, _ = json.Marshal(HTTPResponse{
			Code: "success",
			Msg:  comm.Version,
			Data: struct {
				Hostname string `json:"hostname"`
				CheckSum string `json:"checksum"`
			}{
				Hostname: hostName,
				CheckSum: fileChecksum,
			},
		})
		w.Write(data)
	}
}
