package email

import (
	"bytes"
	"crypto/tls"
	"encoding/base64"
	"fmt"
	"mime/multipart"
	"net"
	"net/smtp"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

// Sender is the sender of the email message
type Sender struct {
	smtp.Auth
	client *smtp.Client
	Host   string
	Port   int
	From   string
}

// NewMessageSender creates a new email sender
func NewMessageSender(host string, port int, from, accessToken string) (*Sender, error) {
	conn, err := net.Dial("tcp", net.JoinHostPort(host, strconv.Itoa(port)))
	if err != nil {
		println(err)
	}

	c, err := smtp.NewClient(conn, host)
	if err != nil {
		println(err)
	}

	tlsconfig := &tls.Config{
		ServerName: host,
	}

	// send STARTTLS command to server.
	if err = c.StartTLS(tlsconfig); err != nil {
		return nil, fmt.Errorf("StartTLS err: %v", err)
	}

	token := fmt.Sprintf("user=%s\001auth=Bearer %s\001\001", from, accessToken)
	auth := Auth(token)
	if err := c.Auth(auth); err != nil {
		return nil, fmt.Errorf("Auth err: %v", err)
	}

	return &Sender{Auth: auth, From: from, Host: host, Port: port, client: c}, nil
}

// Send sends the email
func (s *Sender) Send(m *Message) error {
	host := net.JoinHostPort(s.Host, strconv.Itoa(s.Port))
	return smtp.SendMail(host, s.Auth, s.From, m.to, m.Bytes())
}

type attachment struct {
	Content     []byte
	ContentType string
}

// Message of email
type Message struct {
	header  map[string][]string
	to      []string
	cc      []string
	bcc     []string
	subject string
	body    struct {
		content     []byte
		contentType string
	}
	attachments []*file
	embedded    []*file
}

type file struct {
	Name    string
	Header  map[string][]string
	Content []byte
}

// NewMessage creates a new email message
func NewMessage() *Message {
	return &Message{header: make(map[string][]string), attachments: make([]*file, 0), embedded: make([]*file, 0)}
}

func (m *Message) setHeader(header string, content []string) {
	m.header[header] = content
}

// Subject sets subject of the email
func (m *Message) Subject(content string) {
	m.setHeader("Subject", []string{content})
}

// To sets the recipients of the email
func (m *Message) To(to []string) {
	m.to = to
}

// CC sets the carbon copy recipients of the email
func (m *Message) CC(cc []string) {
	m.cc = cc
}

// BCC sets the blind carbon copy recipients of the email
func (m *Message) BCC(bcc []string) {
	m.bcc = bcc
}

// Body sets body of the email along with content type
func (m *Message) Body(content, contentType string) {
	m.body.content = []byte(content)
	m.body.contentType = contentType
}

// Attach attaches the files to the email.
func (m *Message) Attach(file []byte, filename string) {
	m.attachments = m.appendFile(m.attachments, file, filename)
}

// Embed embeds the images to the email.
func (m *Message) Embed(file []byte, filename string) {
	m.embedded = m.appendFile(m.embedded, file, filename)
}

func (m *Message) appendFile(list []*file, content []byte, name string) []*file {
	f := &file{
		Name:    filepath.Base(name),
		Header:  make(map[string][]string),
		Content: content,
	}
	if list == nil {
		return []*file{f}
	}

	return append(list, f)
}

// Bytes convert message to bytes with multipart
func (m *Message) Bytes() []byte {
	buf := bytes.NewBuffer(nil)

	m.setHeader("MIME-Version", []string{"1.0"})
	m.setHeader("Date", []string{time.Now().Format(time.RFC1123Z)})
	m.setHeader("To", m.to)
	if len(m.cc) > 0 {
		m.setHeader("Cc", m.cc)
	}
	if len(m.bcc) > 0 {
		m.setHeader("Bcc", m.bcc)
	}
	for k, v := range m.header {
		buf.WriteString(fmt.Sprintf("%s: %s\n", k, strings.Join(v, ",")))
	}

	writer := multipart.NewWriter(buf)
	boundary := writer.Boundary()
	if len(m.attachments) > 0 {
		buf.WriteString(fmt.Sprintf("Content-Type: multipart/mixed; boundary=%s\r\n", boundary))
	}
	if len(m.body.content) > 0 {
		if m.body.contentType == "" {
			m.body.contentType = "text/plain"
		}
		buf.WriteString(fmt.Sprintf("--%s\r\n", boundary))
		buf.WriteString(fmt.Sprintf("Content-Type: %s; charset=utf-8\r\n", m.body.contentType))
		buf.WriteString("Content-Transfer-Encoding: base64\r\n\r\n")

		bodyBytes := make([]byte, base64.StdEncoding.EncodedLen(len(m.body.content)))
		base64.StdEncoding.Encode(bodyBytes, m.body.content)
		buf.WriteString(string(bodyBytes))
		buf.WriteString("\r\n")
	}
	if len(m.attachments) > 0 {
		for _, v := range m.attachments {
			buf.WriteString(fmt.Sprintf("--%s\r\n", boundary))
			buf.WriteString(fmt.Sprintf("Content-Type: %s\r\n", v.Header["Content-Type"][0]))
			buf.WriteString("Content-Transfer-Encoding: base64\r\n")
			buf.WriteString(fmt.Sprintf("Content-Disposition: attachment; filename=%s\r\n\r\n", v.Name))

			b := make([]byte, base64.StdEncoding.EncodedLen(len(v.Content)))
			base64.StdEncoding.Encode(b, v.Content)
			buf.Write(b)
		}
	}

	return buf.Bytes()
}

func encodeBase64(bytes []byte) []byte {
	b := make([]byte, base64.StdEncoding.EncodedLen(len(bytes)))
	base64.StdEncoding.Encode(b, bytes)
	return b
}
