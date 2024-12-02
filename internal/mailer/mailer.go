package mailer

import (
	"bytes"
	"embed"
	"text/template"
	"time"

	"github.com/go-mail/mail/v2"
)

// this comment is directive for embeding the template folder into our binary to store templates

//go:embed "templates"
var templateFS embed.FS

// define mailer struct
type Mailer struct {
	dialer *mail.Dialer
	sender string
}

// return new instance of mailer
func New(host string, port int, username, password, sender string) Mailer {
	// initialize a new mail.Dialer instance
	dialer := mail.NewDialer(host, port, username, password)
	dialer.Timeout = 3 * time.Second

	// return mailer instance
	return Mailer{
		dialer: dialer,
		sender: sender,
	}
}

// send method takes mailer and  template string, data interface{}
func (m Mailer) Send(recipient, templateFile string, data interface{}) error {
	// User parseFS() for teplate file from the embeded file system
	templ, err := template.New("email").ParseFS(templateFS, "templates/"+templateFile)
	if err != nil {
		return err
	}

	// storing the data in bytes.buffer
	subject := new(bytes.Buffer)
	err = templ.ExecuteTemplate(subject, "subject", data)
	if err != nil {
		return err
	}

	// for plain body
	plainBody := new(bytes.Buffer)
	err = templ.ExecuteTemplate(plainBody, "plainBody", data)
	if err != nil {
		return err
	}

	// for html body
	htmlBody := new(bytes.Buffer)
	err = templ.ExecuteTemplate(htmlBody, "htmlBody", data)
	if err != nil {
		return err
	}

	// create mail.message instance and set the headers and body
	msg := mail.NewMessage()
	msg.SetHeader("To", recipient)
	msg.SetHeader("From", m.sender)
	msg.SetHeader("Subject", subject.String())
	msg.SetBody("text/plain", plainBody.String())
	msg.AddAlternative("text/html", htmlBody.String())

	for i := 1; i <= 3; i++ {

		// call dial and send method on the dialer
		err = m.dialer.DialAndSend(msg)
		if nil == err {
			return nil
		}
		time.Sleep(500 * time.Microsecond)
	}
	return err
}
