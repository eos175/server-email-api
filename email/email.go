package email

import (
	"errors"
	"fmt"
	"net"
	"net/smtp"
	"net/textproto"
	"strings"
	"time"

	"github.com/eos175/email"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
	"gitlab.com/jeosgram-go/server-email/model"
)

var (
	ErrMail = errors.New("invalid mail")
)

func checkMail(s string) bool {
	// a@b.io
	return len(s) > 5 && strings.IndexByte(s, '@') != -1
}

func generateMessageID() string {
	return fmt.Sprintf("<%s@jeosgram.io>", uuid.NewString())
}

func newEmail(e *model.Email) (*email.Email, string) {
	msgID := generateMessageID()
	return &email.Email{
		ReplyTo: e.ReplyTo,
		From:    e.From,
		To:      e.To,
		Bcc:     e.Bcc,
		Cc:      e.Cc,
		Subject: e.Subject,
		Text:    []byte(e.Text),
		HTML:    []byte(e.Html),
		Headers: textproto.MIMEHeader{"Message-ID": []string{msgID}},
	}, msgID
}

// ========================================================

func Queue(addr, user, pass string) func(e *model.Email) (emailID string, err error) {
	const (
		timeout    = 8 * time.Second
		maxRetries = 2
		maxConn    = 4
	)

	host, _, _ := net.SplitHostPort(addr)
	pool, err := email.NewPool(
		addr,
		maxConn,
		smtp.PlainAuth("", user, pass, host),
	)
	if err != nil {
		log.Fatal().Err(err).Send()
	}

	ch := make(chan *email.Email, 1)
	for i := 0; i < maxConn; i++ {
		go func(pid int) {
			for e := range ch {
				for i := 0; i < maxRetries; i++ {
					log.Info().Int("pid", pid).Msg("send mail")
					err := pool.Send(e, timeout)
					if err == nil {
						break
					}
					log.Error().Int("pid", pid).Int("retry", i).Err(err).Msg("no send mail")
				}
			}
		}(i + 1)
	}

	return func(e *model.Email) (string, error) {
		if len(e.To) == 0 {
			return "", ErrMail
		}

		if !checkMail(e.From) {
			return "", ErrMail
		}

		for _, v := range e.To {
			if !checkMail(v) {
				return "", ErrMail
			}
		}

		email, msgID := newEmail(e)
		ch <- email
		return msgID, nil
	}
}
