# Server Email API

The Email API is a service that allows you to send emails through a simple RESTful interface. It provides endpoints to send emails and upload attachments. The API is designed to be easy to use and can be integrated into various applications.

## API Endpoints

### Send Email

Send an email by making a POST request to the `/v1/smtp/email` endpoint.

- Endpoint: `/v1/smtp/email`
- Method: `POST`
- Request Body: JSON object representing the email to be sent.

The email struct should have the following format:

```go
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
```

### Upload Attachment

Upload an attachment by making a POST request to the `/v1/files` endpoint.

- Endpoint: `/v1/files`
- Method: `POST`
- Request Body: The attachment file to be uploaded.

### Error Handling

The Email API handles the following errors:

- `ErrBigFile`: Indicates that the file size is too large for upload.
- `ErrMail`: Indicates that the email is invalid.

## TODO

There are some pending tasks to be addressed in the Email API:

- Add support for sending attachments along with emails.
- Improve the overall structure and organization of the codebase.
- Consider implementing token-based authentication for enhanced security, although the current implementation is restricted to internal network usage.
- Expand the functionality to allow for more than just the REST API.
- Implement a mechanism to verify email delivery status using the message ID.

Your contributions are welcome to help improve and enhance the Email API!

## License

This project is licensed under the [MIT License](LICENSE).
