package queue

import (
	"fmt"
	"log"
	"strings"

	"github.com/mitchellh/mapstructure"
	sendgrid "github.com/sendgrid/sendgrid-go"
	"github.com/sendgrid/sendgrid-go/helpers/mail"
)

// EmailPayload Data for sending email stored in and pulled from queue
type EmailPayload struct {
	Attachments  []Attachment `json:"attachments"`
	FromEmail    string       `json:"from_email"`
	FromName     string       `json:"from_name"`
	HTMLBody     string       `json:"html_body"`
	ReplyToEmail string       `json:"reply_to_email"`
	ReplyToName  string       `json:"reply_to_name"`
	Subject      string       `json:"subject"`
	TextBody     string       `json:"text_body"`
	ToEmail      string       `json:"to_email"`
	ToName       string       `json:"to_name"`
}

// Attachment For email, data should be base64 encoded
type Attachment struct {
	Data     string
	Filename string
	Type     string
}

// EmailSendgridAction Handler for emails in message queue
type EmailSendgridAction struct {
	BccEmail    *string
	SendgridAPI string
}

// NewEmailSendgridAction Returns an email action for esync
func NewEmailSendgridAction(sendgridAPI string, bccEmail *string) EmailSendgridAction {
	return EmailSendgridAction{SendgridAPI: sendgridAPI, BccEmail: bccEmail}
}

// Do Handle a particular email task
func (a EmailSendgridAction) Do(task Task) (TaskResult, string) {
	retry := TaskResultRetryFailure

	var p EmailPayload

	err := mapstructure.Decode(task.Data, &p)

	if err != nil {
		return retry, err.Error()
	}

	from := mail.NewEmail(p.FromName, p.FromEmail)
	to := mail.NewEmail(p.ToName, p.ToEmail)

	m := mail.NewSingleEmail(from, p.Subject, to, p.TextBody, p.HTMLBody)

	if a.BccEmail != nil && len(*a.BccEmail) > 0 && len(m.Personalizations) > 0 && strings.TrimSpace(*a.BccEmail) != strings.TrimSpace(to.Address) {
		bccTo := mail.NewEmail("", *a.BccEmail)
		m.Personalizations[0].AddBCCs(bccTo)
	}

	for _, a := range p.Attachments {

		m.AddAttachment(&mail.Attachment{
			Content:  a.Data,
			Filename: a.Filename,
			Type:     a.Type,
		})
	}

	// Add reply to email address if provided
	if len(p.ReplyToEmail) > 0 {
		replyTo := mail.Email{
			Name:    p.ReplyToName,
			Address: p.ReplyToEmail,
		}
		m.ReplyTo = &replyTo
	}

	client := sendgrid.NewSendClient(a.SendgridAPI)
	response, err := client.Send(m)

	if err != nil {
		return retry, err.Error()
	}

	if response.StatusCode < 200 || response.StatusCode >= 300 {
		fmt.Println(response.StatusCode)
		fmt.Println(response.Body)
		fmt.Println(response.Headers)
		msg := fmt.Sprintf("Unexpected status code %d when sending email from %s to %s: %s", response.StatusCode, p.FromEmail, p.ToEmail, response.Body)
		return retry, msg
	}

	log.Printf("Email sent to %s.  Response status: %d.  Body: %s", p.ToEmail, response.StatusCode, response.Body)

	return TaskResultSuccess, "Success"
}
