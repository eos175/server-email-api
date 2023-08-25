package grpc

import (
	"server-email/model"

	"gitlab.com/jeosgram-go/qrpc"
)

type Client2 struct {
	*qrpc.Client
}

func (c *Client2) SendEmail(e *model.Email) (msgID string, err error) {
	var ret string
	err = c.Send(e, &ret)
	if err != nil {
		return "", err
	}
	return ret, nil
}
