package klorf

import (
	"fmt"

	irc "github.com/klorf/goirc/client"
)

type Channel struct {
	Channel string   `json:"channel"`
	Members []string `json:"members"`
}

func (k *Klorf) Join(conn *irc.Conn, channel string) {
	conn.Join(channel)
	conn.Privmsg(channel, "klorf klorf klorf")

	c := new(Channel)
	c.Channel = channel

	k.Channels = append(k.Channels, c)
}

func (c *Channel) Add(nick string) {
	c.Members = append(c.Members, nick)
}

func (c *Channel) HasUser(nick string) bool {
	fmt.Println(c.Members)
	for _, n := range c.Members {
		if nick == n {
			return true
		}
	}
	return false
}

func (c *Channel) Remove(nick string) {
	var mem []string
	for _, n := range c.Members {
		if nick != n {
			mem = append(mem, n)
		}
	}
	c.Members = mem
}
