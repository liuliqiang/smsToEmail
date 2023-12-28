package main

import (
	"context"
	"crypto/tls"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"

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

	mux := http.NewServeMux()
	mux.HandleFunc("/sms", smsHTTPHandler)
	server := &http.Server{
		Addr:    addr,
		Handler: mux,
	}

	ctx, cancel := context.WithCancel(context.Background())
	signalHandler(ctx, cancel)

	log4go.Info("Sms to email server listen at: %s", addr)
	go func() {
		defer func() {
			if err := recover(); err != nil {
				log4go.Error("http server panic: %v", err)
			}
		}()
		if err := server.ListenAndServe(); err != nil {
			log4go.Error("http server listen and serve: %v", err)
		}
		cancel()
	}()

	<-ctx.Done()
	log4go.Info("Ready to shutdown http server")
	_ = server.Shutdown(context.Background())
	log4go.Info("Shutdown http server success")
}

// add signal handler for graceful shutdown
func signalHandler(ctx context.Context, cancelFunc context.CancelFunc) {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGTERM, syscall.SIGINT)

	go func() {
		select {
		case <-sigChan:
			log4go.Info("Receive signal, cancel context")
			cancelFunc()
		case <-ctx.Done():
			log4go.Info("Context canceled, signal handler exit")
			return
		}
	}()
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
	srcEmail = os.Getenv("SENDER_EMAIL_ADDR")
	emailPass = os.Getenv("SENDER_EMAIL_PASS")
	dstEmail = os.Getenv("RECEIVER_EMAIL_ADDR")
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
