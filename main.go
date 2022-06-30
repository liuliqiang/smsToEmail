package main

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/liuliqiang/log4go"
	gomail "gopkg.in/mail.v2"
)

var (
	port       string = "8080"
	srcEmail   string
	emailPass  string
	dstEmail   string
	retryCount = 3
)

func main() {
	if portEnv := os.Getenv("SMS_PORT"); portEnv != "" {
		port = portEnv
	}
	addr := fmt.Sprintf("0.0.0.0:%s", port)

	http.HandleFunc("/sms", func(w http.ResponseWriter, r *http.Request) {
		reqData, err := ioutil.ReadAll(r.Body)
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
	})

	log4go.Info("Sms to email server listen at: %s", addr)
	http.ListenAndServe(addr, nil)
}

type smsInfo struct {
	From string `json:"from"`
	SMS  string `json:"sms"`
}

func (i *smsInfo) GetFrom() string {
	if i == nil {
		return ""
	}
	return i.From
}

func (i *smsInfo) GetSMS() string {
	if i == nil {
		return ""
	}
	return i.SMS
}

func sendEmail(sms *smsInfo) error {
	srcEmail = os.Getenv("SRC_EMAIL_ADDR")
	emailPass = os.Getenv("SRC_EMAIL_PASS")
	dstEmail = os.Getenv("DEST_EMAIL_ADDR")
	log4go.Info("email: '%s', password: '%s****'", srcEmail, emailPass[:4])

	// Message.
	m := gomail.NewMessage()
	smtpInfo := getSmtpInfo(srcEmail)

	m.SetHeader("From", srcEmail)
	m.SetHeader("To", dstEmail)
	m.SetHeader("Subject", "CN SMS From "+sms.GetFrom())
	m.SetBody("text/plain", sms.GetSMS())

	d := gomail.NewDialer(smtpInfo.GetHost(), smtpInfo.MustGetIntPort(), srcEmail, emailPass)
	d.TLSConfig = &tls.Config{InsecureSkipVerify: true}
	if err := d.DialAndSend(m); err != nil {
		return fmt.Errorf("send email: %w", err)
	}
	return nil
}

type smtpInfo struct {
	host string
	port string
}

func (i *smtpInfo) GetHost() string {
	if i == nil {
		return ""
	}
	return i.host
}

func (i *smtpInfo) GetPort() string {
	if i == nil {
		return ""
	}
	return i.port
}

func (i *smtpInfo) MustGetIntPort() int {
	portStr := i.GetPort()
	if port, err := strconv.Atoi(portStr); err != nil {
		panic("convert port " + portStr)
	} else {
		return port
	}
}

func (i *smtpInfo) GetAddr() string {
	if i == nil {
		return ""
	}
	return i.host + ":" + i.port
}

func getSmtpInfo(mail string) *smtpInfo {
	switch {
	case strings.Contains(mail, "gmail"):
		return &smtpInfo{
			host: "smtp.gmail.com",
			port: "587",
		}
	case strings.Contains(mail, "qq"):
		return &smtpInfo{
			host: "smtp.qq.com",
			port: "465",
		}
	default:
		return nil
	}
}
