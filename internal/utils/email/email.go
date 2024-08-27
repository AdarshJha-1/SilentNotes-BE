package email

import (
	"bytes"
	"html/template"
	"log"
	"os"

	"gopkg.in/gomail.v2"
)

var (
	rootEmail   = os.Getenv("ROOT_EMAIL")
	emailSecret = os.Getenv("EMAIL_SECRET")
	subject     = "AMA | Verification Code"
)

type EmailStrut struct {
	Name string
	Code int
}

func SendVerificationEmail(username, email string, otp int) error {

	var body bytes.Buffer
	t, err := template.ParseFiles("internal/utils/email/email.html")
	if err != nil {
		return err
	}
	if err := t.Execute(&body, EmailStrut{Name: username, Code: otp}); err != nil {
		return err
	}

	// send with go mail
	m := gomail.NewMessage()
	m.SetHeader("From", rootEmail)
	m.SetHeader("To", email)
	m.SetHeader("Subject", subject)
	m.SetBody("text/html", body.String())

	d := gomail.NewDialer("smtp.gmail.com", 587, rootEmail, emailSecret)

	if err := d.DialAndSend(m); err != nil {
		log.Printf("Error sending email: %v", err)
		return err
	}
	return nil
}
