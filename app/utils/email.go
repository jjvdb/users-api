package utils

import (
	"versequick-users-api/app/appdata"

	gomail "gopkg.in/mail.v2"
)

func SendEmail(to string, subject string, body string) error {
	m := gomail.NewMessage()

	m.SetHeader("From", appdata.SmtpUsername)
	m.SetHeader("To", to)
	m.SetHeader("Subject", subject)
	m.SetBody("text/plain", "This is Gomail test body")
	d := gomail.NewDialer(appdata.SmtpServer, int(appdata.SmtpPort), appdata.SmtpUsername, appdata.SmtpPassword)

	// This is only needed when SSL/TLS certificate is not valid on server.
	// In production this should be set to false.

	// d.TLSConfig = &tls.Config{InsecureSkipVerify: true}

	if err := d.DialAndSend(m); err != nil {
		return err
	}

	return nil
}
