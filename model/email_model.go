package model

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

func (Email) CRC() uint32 {
	return 0xeda364fd
}
