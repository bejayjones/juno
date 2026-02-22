// Package email provides EmailSender implementations for the reporting context.
package email

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"net"
	"net/smtp"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

// ErrDeliveryQueued is returned by QueueOnlySender to signal that the delivery
// was accepted for later send, not that an error occurred. The reporting
// application service treats this as "leave delivery in pending state".
var ErrDeliveryQueued = errors.New("delivery queued for later send")

// QueueOnlySender satisfies EmailSender but never transmits email.
// Deliveries are persisted as pending records and sent by an external process
// (or a manual retry call). Suitable for fully offline mode.
type QueueOnlySender struct{}

func NewQueueOnlySender() *QueueOnlySender { return &QueueOnlySender{} }

func (s *QueueOnlySender) SendReport(_ context.Context, _, _ string) error {
	return ErrDeliveryQueued
}

// SMTPSender sends the report PDF as an email attachment via SMTP.
type SMTPSender struct {
	host string
	port int
	user string
	pass string
}

func NewSMTPSender(host string, port int, user, pass string) *SMTPSender {
	return &SMTPSender{host: host, port: port, user: user, pass: pass}
}

func (s *SMTPSender) SendReport(ctx context.Context, toEmail, pdfPath string) error {
	pdfData, err := os.ReadFile(pdfPath)
	if err != nil {
		return fmt.Errorf("read pdf: %w", err)
	}

	subject := "Home Inspection Report"
	fileName := filepath.Base(pdfPath)

	// Build a MIME multipart message with the PDF attachment.
	boundary := "==JUNO_BOUNDARY=="
	var body strings.Builder
	body.WriteString("MIME-Version: 1.0\r\n")
	body.WriteString(fmt.Sprintf("To: %s\r\n", toEmail))
	body.WriteString(fmt.Sprintf("Subject: %s\r\n", subject))
	body.WriteString(fmt.Sprintf("Content-Type: multipart/mixed; boundary=%q\r\n\r\n", boundary))

	body.WriteString(fmt.Sprintf("--%s\r\n", boundary))
	body.WriteString("Content-Type: text/plain; charset=utf-8\r\n\r\n")
	body.WriteString("Please find your home inspection report attached.\r\n\r\n")

	body.WriteString(fmt.Sprintf("--%s\r\n", boundary))
	body.WriteString(fmt.Sprintf("Content-Type: application/pdf; name=%q\r\n", fileName))
	body.WriteString("Content-Transfer-Encoding: base64\r\n")
	body.WriteString(fmt.Sprintf("Content-Disposition: attachment; filename=%q\r\n\r\n", fileName))
	body.WriteString(base64Encode(pdfData))
	body.WriteString(fmt.Sprintf("\r\n--%s--\r\n", boundary))

	addr := net.JoinHostPort(s.host, strconv.Itoa(s.port))
	auth := smtp.PlainAuth("", s.user, s.pass, s.host)

	// Use TLS for port 465, STARTTLS negotiation for 587.
	if s.port == 465 {
		tlsCfg := &tls.Config{ServerName: s.host}
		conn, err := tls.Dial("tcp", addr, tlsCfg)
		if err != nil {
			return fmt.Errorf("smtp tls dial: %w", err)
		}
		client, err := smtp.NewClient(conn, s.host)
		if err != nil {
			return fmt.Errorf("smtp client: %w", err)
		}
		defer client.Close()
		if err := client.Auth(auth); err != nil {
			return fmt.Errorf("smtp auth: %w", err)
		}
		if err := client.Mail(s.user); err != nil {
			return fmt.Errorf("smtp mail from: %w", err)
		}
		if err := client.Rcpt(toEmail); err != nil {
			return fmt.Errorf("smtp rcpt: %w", err)
		}
		w, err := client.Data()
		if err != nil {
			return fmt.Errorf("smtp data: %w", err)
		}
		defer w.Close()
		_, err = fmt.Fprint(w, body.String())
		return err
	}

	return smtp.SendMail(addr, auth, s.user, []string{toEmail}, []byte(body.String()))
}

// base64Encode encodes data in base64 with line wrapping at 76 chars (RFC 2045).
func base64Encode(data []byte) string {
	const chars = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789+/"
	encoded := make([]byte, 0, (len(data)+2)/3*4)
	for i := 0; i < len(data); i += 3 {
		var b [3]byte
		n := copy(b[:], data[i:])
		encoded = append(encoded,
			chars[b[0]>>2],
			chars[(b[0]&0x3)<<4|b[1]>>4],
		)
		if n > 1 {
			encoded = append(encoded, chars[(b[1]&0xf)<<2|b[2]>>6])
		} else {
			encoded = append(encoded, '=')
		}
		if n > 2 {
			encoded = append(encoded, chars[b[2]&0x3f])
		} else {
			encoded = append(encoded, '=')
		}
	}
	// Wrap at 76 characters.
	var out strings.Builder
	for i := 0; i < len(encoded); i += 76 {
		end := i + 76
		if end > len(encoded) {
			end = len(encoded)
		}
		out.Write(encoded[i:end])
		out.WriteString("\r\n")
	}
	return out.String()
}
