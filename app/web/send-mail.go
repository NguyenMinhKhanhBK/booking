package main

import (
	"booking/models"
	"fmt"
	"io/ioutil"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
	mail "github.com/xhit/go-simple-mail/v2"
)

func listenForMail() {
	go func() {
		logrus.Info("listenForMail goroutine created")
		defer logrus.Info("listenForMail destroyed")
		for {
			msg := <-app.MailChan
			logrus.WithField("msg", msg).Info("listenForMail receives an email request")
			sendMsg(msg)
		}
	}()
}

func sendMsg(m models.MailData) {
	server := mail.NewSMTPClient()
	server.Host = "localhost"
	server.Port = 1025
	server.KeepAlive = false
	server.ConnectTimeout = 10 * time.Second
	server.SendTimeout = 10 * time.Second

	client, err := server.Connect()
	if err != nil {
		logrus.WithError(err).Error("cannot connect server")
		return
	}

	logrus.Info("Connected to email server")
	email := mail.NewMSG()
	email.SetFrom(m.From).AddTo(m.To).SetSubject(m.Subject)
	if m.Template == "" {
		email.SetBody(mail.TextHTML, m.Content)
	} else {
		data, err := ioutil.ReadFile(fmt.Sprintf("./email-templates/%s", m.Template))
		if err != nil {
			logrus.WithError(err).Error("cannot read file")
			return
		}
		mailTemplate := string(data)
		msgToSend := strings.Replace(mailTemplate, "[%body%]", m.Content, 1)
		email.SetBody(mail.TextHTML, msgToSend)
	}

	err = email.Send(client)
	if err != nil {
		logrus.WithError(err).Error("cannot send email")
	} else {
		logrus.Info("Email sent!")
	}
}
