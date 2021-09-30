package email

import "net/smtp"

// SMTPAuth SMTPAuth
type SMTPAuth struct {
	accessToken string
}

// Auth auth
func Auth(token string) smtp.Auth {
	return &SMTPAuth{token}
}

// Start start
func (a *SMTPAuth) Start(server *smtp.ServerInfo) (string, []byte, error) {
	return "xoauth2", []byte(""), nil
}

// Next next
func (a *SMTPAuth) Next(fromServer []byte, more bool) ([]byte, error) {
	if more {
		return []byte(a.accessToken), nil
	}
	return nil, nil
}
