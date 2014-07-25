package testpod

import (
	"bytes"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"
)

type Plan struct {
	l           sync.Mutex
	Clients     map[*Client]bool
	Outstanding map[string]map[int]time.Time

	ConnectThrottle int

	NumChannels int
	NumClients  int
	URLs        []string
}

func NewPlan() *Plan {
	return &Plan{
		Clients:     make(map[*Client]bool),
		Outstanding: make(map[string]map[int]time.Time),
	}
}

func (p *Plan) AddClient(c *Client) {
	p.l.Lock()
	defer p.l.Unlock()
	p.Clients[c] = true
}

func (p *Plan) DeleteClient(c *Client) {
	p.l.Lock()
	defer p.l.Unlock()

	delete(p.Clients, c)
}

func (p *Plan) Go() {
	for i := 0; i < p.NumClients; i++ {
		go func(url string) {
			for {
				c := NewClient()
				err := p.Loop(c, url)
				log.Print("Error: ", err)
				time.Sleep(5 * time.Second)
			}
		}(p.URLs[i%len(p.URLs)])
	}
	for {
		time.Sleep(5 * time.Second)

	}

}

func (p *Plan) Loop(c *Client, url string) error {
	err := c.Dial(url)
	if err != nil {
		return err
	}
	err = c.Handshake()
	if err != nil {
		return err
	}
	for {
		msg, err := c.ReadMsg()
		if err != nil {
			return err
		}

		switch msg.MessageType {
		case "":
			err := c.Ping()
			if err != nil {
				return err
			}
		case "hello":
			c.UAID = msg.UAID
			for i := 0; i < p.NumChannels; i++ {
				err := c.Register()
				if err != nil {
					return err
				}
			}
		case "register":
			c.ChannelIDs = append(c.ChannelIDs, msg.ChannelID)
			go p.StartSender(msg.PushEndpoint, msg.ChannelID)
		case "notification":
			err := c.Ack(msg.Updates)
			if err != nil {
				return err
			}
			p.l.Lock()
			for _, u := range msg.Updates {
				when := p.Outstanding[u.ChannelID][u.Version]
				fmt.Println("Took:", time.Since(when))
				delete(p.Outstanding[u.ChannelID], u.Version)
			}
			p.l.Unlock()
		}
	}
	return nil
}

func (p *Plan) StartSender(url, channelID string) {
	c := &http.Client{}
	p.l.Lock()
	p.Outstanding[channelID] = make(map[int]time.Time)
	p.l.Unlock()

	version := 1
	for {
		body := bytes.NewBufferString(fmt.Sprintf("version=%d", version))
		req, err := http.NewRequest("PUT", url, body)
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		if err != nil {
			fmt.Print("StartSender:", err)
			return
		}

		p.l.Lock()
		p.Outstanding[channelID][version] = time.Now()
		p.l.Unlock()

		resp, err := c.Do(req)
		if err != nil {
			fmt.Print("StartSender:", err)
			return
		}
		resp.Body.Close()

		version++
		time.Sleep(5 * time.Second)
	}
}
