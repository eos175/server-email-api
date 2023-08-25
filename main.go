package main

import (
	"errors"
	"fmt"
	"os"
	"server-email/model"

	"server-email/email"

	"github.com/gofiber/fiber/v2"
	fiberlog "github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/gofiber/fiber/v2/middleware/requestid"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"gitlab.com/jeosgram-go/qrpc"
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

var (
	ErrBigFile = errors.New("big file")
)

// ========================================================

func main() {

	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stdout})

	sender := email.Queue("smtp.jeosgram.io:587", "user@jeosgram.io", "secret0")

	go func() {
		app := qrpc.NewServer()

		qrpc.RegisterRpcFuncResponse(app, func(c *qrpc.Context, data model.Email) error {
			msgID, err := sender(&data)
			if err != nil {
				return c.SendError(500, "ERROR_SENDING_EMAIL")
			}

			return c.SendResponse(msgID)
		})

		log.Error().Err(app.Listen(":8080")).Msg("error server")
	}()

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
		var e model.Email

		if err := c.BodyParser(&e); err != nil {
			return err
		}

		msgID, err := sender(&e)
		if err != nil {
			return err
		}

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
