package testpod

type AckClient struct {
	MessageType string   `json:"messageType"`
	Updates     []Update `json:"updates"`
}

type HandshakeClient struct {
	MessageType string   `json:"messageType"`
	UAID        string   `json:"uaid"`
	ChannelIDs  []string `json:"channelIDs"`
}

type RegisterClient struct {
	MessageType string `json:"messageType"`
	ChannelID   string `json:"channelID"`
}

type ServerMessage struct {
	MessageType string `json:"messageType"`

	ChannelID    string   `json:"channelID"`
	PushEndpoint string   `json:"pushEndpoint"`
	Status       int      `json:"status"`
	UAID         string   `json:"uaid"`
	Updates      []Update `json:"updates"`
}

type Update struct {
	ChannelID string `json:"channelID"`
	Version   int    `json:"version"`
}
