package main

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/liuliqiang/log4go"
)

func smsHTTPHandler(w http.ResponseWriter, r *http.Request) {
	defer func() {
		if err := recover(); err != nil {
			log4go.Error("process http request panic: %v", err)
		}
	}()
	reqData, err := io.ReadAll(r.Body)
	if err != nil {
		log4go.Error("Failed to get request data: %v", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	var sms smsInfo
	if err = json.Unmarshal(reqData, &sms); err != nil {
		log4go.Error("Failed to parse request data: %v", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	log4go.Info("sms from: %s", sms.GetFrom())
	log4go.Info("sms msg: %s", sms.GetSMS())
	go func() {
		defer func() {
			if err := recover(); err != nil {
				log4go.Error("send email panic: %v", err)
			}
		}()
		for i := 0; i < retryCount; i++ {
			if err = sendEmail(&sms); err != nil {
				log4go.Error("Failed to send email(%d/%d): %v", i+1, retryCount, err)
			} else {
				log4go.Info("Send sms success")
				return
			}
		}
	}()
	w.WriteHeader(http.StatusOK)
}
