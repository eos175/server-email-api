package main

import (
	"errors"
	"fmt"
	"net"
	"net/smtp"
	"net/textproto"
	"os"
	"strings"
	"time"

	"github.com/eos175/email"

	"github.com/gofiber/fiber/v2"
	fiberlog "github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/gofiber/fiber/v2/middleware/requestid"
	"github.com/google/uuid"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

/*

TODO()
	- agregar para enviar archivos
	- estructurarlo bien
	- quizas algun token quemado o algo, aunq solo va funcionar con la red interna
	- permitir mas cosas y no solo la `api rest`
	- permitir verificar por el msgID si el correo se envio

*/

type ByteSize uint

const (
	_           = iota
	KB ByteSize = 1 << (10 * iota)
	MB
	GB
	TB
)

type Email struct {
	From        string   `json:"from"`
	To          []string `json:"to"`
	ReplyTo     []string `json:"replyTo"`
	Bcc         []string `json:"bcc"`
	Cc          []string `json:"cc"`
	Subject     string   `json:"subject"`
	Text        string   `json:"text"`        // Plaintext message (optional)
	Html        string   `json:"html"`        // Html message (optional)
	Attachments []string `json:"attachments"` // file-ID (optional)
}

var (
	ErrBigFile = errors.New("big file")
	ErrMail    = errors.New("invalid mail")
)

func checkMail(s string) bool {
	// a@b.io
	return len(s) > 5 && strings.IndexByte(s, '@') != -1
}

func generateMessageID() string {
	return fmt.Sprintf("<%s@example.io>", uuid.NewString())
}

// ========================================================

func main() {

	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stdout})

	sender := Queue("smtp.example.io:587", "user@example.io", "secret++")

	// https://docs.gofiber.io/guide/error-handling
	app := fiber.New(fiber.Config{
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			// Status code defaults to 500
			code := fiber.StatusInternalServerError

			// Retrieve the custom status code if it's a *fiber.Error
			var e *fiber.Error
			if errors.As(err, &e) {
				code = e.Code
			}

			// Set Content-Type: text/plain; charset=utf-8
			c.Set(fiber.HeaderContentType, fiber.MIMETextPlainCharsetUTF8)

			// Return status code with error message
			return c.Status(code).SendString(err.Error())
		}})

	app.Use(fiberlog.New())
	app.Use(requestid.New())
	app.Use(recover.New())

	api := app.Group("/v1")
	api.Post("/smtp/email", func(c *fiber.Ctx) error {
		var e Email

		if err := c.BodyParser(&e); err != nil {
			return err
		}

		if len(e.To) == 0 {
			return ErrMail
		}

		if !checkMail(e.From) {
			return ErrMail
		}

		for _, v := range e.To {
			if !checkMail(v) {
				return ErrMail
			}
		}

		msgID := sender(&e)
		return c.JSON(fiber.Map{"ok": true, "message-id": msgID})
	})

	api.Post("/files", func(c *fiber.Ctx) error {
		file, err := c.FormFile("file")
		if err != nil {
			return err
		}

		// primero sube el archivo, luego se adjunta al correo

		if file.Size > int64(10*MB) {
			return ErrBigFile
		} else {
			return ErrBigFile

		}

		c.SaveFile(file, fmt.Sprintf("/tmp/uploads_files_email/%s", file.Filename))
		return nil
	})

	log.Fatal().Err(app.Listen(":8080")).Send()
}

// ========================================================

func newEmail(e *Email) (*email.Email, string) {
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

func Queue(addr, user, pass string) func(e *Email) (emailID string) {
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

	return func(e *Email) string {
		email, msgID := newEmail(e)
		ch <- email
		return msgID
	}
}
