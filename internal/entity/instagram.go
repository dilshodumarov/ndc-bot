package entity

type InstagramWebhookPayload struct {
	Object string  `json:"object"`
	Entry  []Entry `json:"entry"`
}

type Entry struct {
	ID        string      `json:"id"`
	Time      float64     `json:"time"`
	Messaging []Messaging `json:"messaging"`
}

type Messaging struct {
	Sender      UserInsta         `json:"sender"`
	Recipient   UserInsta         `json:"recipient"`
	Timestamp   float64           `json:"timestamp"`
	Message     *InstagramMessage `json:"message,omitempty"`
	AccessToken string            `json:"access_token"`
}

type UserInsta struct {
	ID string `json:"id"`
}

type InstagramMessage struct {
	MID         string                `json:"mid"`
	Text        string                `json:"text,omitempty"`
	IsEcho      bool                  `json:"is_echo,omitempty"`
	Attachments []InstagramAttachment `json:"attachments,omitempty"`
}

type InstagramAttachment struct {
	Type    string                     `json:"type"` // image, video, audio, file, location
	Payload InstagramAttachmentPayload `json:"payload"`
}

type InstagramAttachmentPayload struct {
	URL string `json:"url"` // Fayl manzili
}

