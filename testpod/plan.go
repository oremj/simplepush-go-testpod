package testpod

import (
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"
)

type Plan struct {
	Clients         map[*Client]bool
	ConnectThrottle int

	Failed int

	l sync.Mutex

	NumChannels int
	NumClients  int
	URLs        []string
}

func NewPlan() *Plan {
	return &Plan{
		Clients: make(map[*Client]bool),
	}
}

type ClientStats struct {
	waitConnect   int
	waitHandshake int
	connected     int
}

func (p *Plan) AddClient(c *Client) {
	p.l.Lock()
	defer p.l.Unlock()
	p.Clients[c] = true
}

func (p *Plan) ClientStats() *ClientStats {
	p.l.Lock()
	defer p.l.Unlock()
	stats := new(ClientStats)
	for c := range p.Clients {
		switch c.State {
		case stateWaitConnect:
			stats.waitConnect++
		case stateWaitHandshake:
			stats.waitHandshake++
		case stateConnected:
			stats.connected++
		}
	}
	return stats
}

func (p *Plan) DeleteClient(c *Client) {
	p.l.Lock()
	defer p.l.Unlock()

	p.Failed++
	delete(p.Clients, c)
}

func (p *Plan) Go() {
	for {
		stats := p.ClientStats()
		fmt.Println(stats, p.Failed)

		for i := 0; i < p.NumClients-len(p.Clients); i++ {
			c := NewClient()
			p.AddClient(c)
			go func(c *Client, url string) {
				defer p.DeleteClient(c)
				err := c.Dial(url)
				if err != nil {
					log.Print("Error connecting: ", err)
					return
				}
				err = p.Loop(c)
				log.Print("Error: ", err)
			}(c, p.URLs[i%len(p.URLs)])
		}

		time.Sleep(1 * time.Second)
	}
}

func (p *Plan) Loop(c *Client) error {
	err := c.Handshake()
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
			c.State = stateConnected
			for i := 0; i < p.NumChannels; i++ {
				err := c.Register()
				if err != nil {
					return err
				}
			}
		case "register":
			c.ChannelIDs = append(c.ChannelIDs, msg.ChannelID)
			go p.StartSender(msg.PushEndpoint)
		case "notification":
			err := c.Ack(msg.Updates)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (p *Plan) StartSender(url string) {
	c := &http.Client{}
	for {
		req, err := http.NewRequest("PUT", url, nil)
		if err != nil {
			fmt.Print("StartSender:", err)
			return
		}
		resp, err := c.Do(req)
		if err != nil {
			fmt.Print("StartSender:", err)
			return
		}
		resp.Body.Close()
	}
}
