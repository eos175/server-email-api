package qrpc

import (
	"gitlab.com/jeosgram-go/qrpc"
	"gitlab.com/jeosgram-go/server-email/model"
)

type Client struct {
	*qrpc.Client
}

func NewClient(addr string) (*Client, error) {
	clientTmp, err := qrpc.NewClient(addr)
	if err != nil {
		return nil, err
	}
	return &Client{
		Client: clientTmp,
	}, nil
}

func (c *Client) SendEmail(e *model.Email) (msgID string, err error) {
	var ret string
	err = c.Send(e, &ret)
	if err != nil {
		return "", err
	}
	return ret, nil
}
