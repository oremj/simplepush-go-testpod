package testpod

import (
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
)

const (
	stateWaitConnect = iota
	stateWaitHandshake
	stateConnected
)

type Client struct {
	WS *websocket.Conn

	UAID       string
	ChannelIDs []string

	rl sync.Mutex
	wl sync.Mutex
}

func NewClient() *Client {
	return &Client{
		ChannelIDs: []string{},
	}
}

func (c *Client) Dial(url string) error {
	dialer := new(websocket.Dialer)
	headers := http.Header{
		"Origin": []string{"http://localhost"},
	}
	ws, _, err := dialer.Dial(url, headers)
	if err != nil {
		return err
	}

	c.WS = ws
	c.ChannelIDs = []string{}

	return nil
}

func (c *Client) Ack(updates []Update) error {
	msg := AckClient{
		MessageType: "ack",
		Updates:     updates,
	}
	return c.Send(msg)
}

func (c *Client) Handshake() error {
	msg := HandshakeClient{
		MessageType: "hello",
		UAID:        c.UAID,
		ChannelIDs:  c.ChannelIDs,
	}
	return c.Send(msg)
}

func (c *Client) Ping() error {
	return c.Send(struct{}{})
}

func (c *Client) ReadMsg() (*ServerMessage, error) {
	c.rl.Lock()
	defer c.rl.Unlock()

	msg := new(ServerMessage)
	err := c.WS.ReadJSON(msg)
	return msg, err
}

func (c *Client) Register() error {
	cid := "d9b74644-4f97-46aa-b8fa-9393985cd6cd"
	msg := RegisterClient{
		MessageType: "register",
		ChannelID:   cid,
	}
	return c.Send(msg)
}

func (c *Client) Send(msg interface{}) error {
	c.wl.Lock()
	defer c.wl.Unlock()
	return c.WS.WriteJSON(msg)
}
