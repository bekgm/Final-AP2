package email

import (
	"fmt"
	"net/smtp"
)

type SMTPSender struct {
	host     string
	port     string
	username string
	password string
	from     string
}

// NewSMTPSender creates a sender compatible with Gmail and Microsoft SMTP.
//
//	Gmail:     host="smtp.gmail.com" port="587"  (use an App Password)
//	Microsoft: host="smtp.office365.com" port="587"
func NewSMTPSender(host, port, username, password, from string) *SMTPSender {
	return &SMTPSender{
		host:     host,
		port:     port,
		username: username,
		password: password,
		from:     from,
	}
}

func (s *SMTPSender) send(to, subject, body string) error {
	auth := smtp.PlainAuth("", s.username, s.password, s.host)

	msg := fmt.Sprintf(
		"From: %s\r\nTo: %s\r\nSubject: %s\r\nMIME-version: 1.0;\r\nContent-Type: text/plain; charset=\"UTF-8\";\r\n\r\n%s",
		s.from, to, subject, body,
	)

	addr := fmt.Sprintf("%s:%s", s.host, s.port)
	return smtp.SendMail(addr, auth, s.from, []string{to}, []byte(msg))
}

// SendApplicationReceived notifies the client that a freelancer applied.
func (s *SMTPSender) SendApplicationReceived(toEmail, jobTitle, freelancerName string) error {
	subject := fmt.Sprintf("New application for \"%s\"", jobTitle)
	body := fmt.Sprintf(
		"Hello,\n\nA freelancer (%s) has applied to your job posting \"%s\".\n\nLog in to review their application.\n\nFreelance Platform",
		freelancerName, jobTitle,
	)
	return s.send(toEmail, subject, body)
}

// SendFreelancerAccepted notifies the freelancer they were accepted.
func (s *SMTPSender) SendFreelancerAccepted(toEmail, jobTitle string) error {
	subject := fmt.Sprintf("You have been accepted for \"%s\"", jobTitle)
	body := fmt.Sprintf(
		"Congratulations!\n\nYou have been accepted for the job \"%s\".\n\nLog in to start communicating with the client.\n\nFreelance Platform",
		jobTitle,
	)
	return s.send(toEmail, subject, body)
}

// NoopSender is a no-op email sender for testing.
type NoopSender struct{}

func (n *NoopSender) SendApplicationReceived(_, _, _ string) error { return nil }
func (n *NoopSender) SendFreelancerAccepted(_, _ string) error     { return nil }
