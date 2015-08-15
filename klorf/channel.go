package klorf

import (
	"encoding/json"
	"fmt"
	"os"
	"regexp"

	irc "github.com/klorf/goirc/client"
)

type Channel struct {
	Channel string   `json:"channel"`
	Members []string `json:"members"`
	Urls    []*Url   `json:"urls"`
	Logger  string   `json:"logger"`
}

func (k *Klorf) Join(conn *irc.Conn, channel string) {
	conn.Join(channel)
	conn.Privmsg(channel, "klorf klorf klorf")

	c := new(Channel)
	c.Channel = channel
	c.Logger = k.Logger
	c = c.loadFromFile()

	k.Channels = append(k.Channels, c)
}

func (c *Channel) Add(nick string) {
	c.Members = append(c.Members, nick)
	c.dumpToFile()
}

func (c *Channel) HasUser(nick string) bool {
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
	c.dumpToFile()
}

func (c *Channel) Robot(line *irc.Line) string {
	message := tokens(line.Args[1])
	regex := regexp.MustCompile(`\b(([\w-]+://?|www[.])[^\s()<>]+(?:\([\w\d]+\)|([^[:punct:]\s]|/)))`)

	for _, y := range message {
		if regex.MatchString(y) {
			if a, b := c.checkRobot(regex, y, line.Nick); a == true {
				return b.GetRobotString()
			} else if b != nil {
				c.Urls = append(c.Urls, b)
			}
		}
	}

	c.dumpToFile()
	return ""
}

func (c *Channel) checkRobot(reg *regexp.Regexp, msg string, user string) (bool, *Url) {
	for _, y := range c.Urls {
		if y.IsRobot(msg, user) {
			return true, y
		} else if y.IsSimilar(msg) {
			return false, nil
		}
	}

	u := NewUrl(user, msg)
	return false, u
}

func (c *Channel) dumpToFile() {
	// store the data
	filename := c.getFilename()
	var stream *os.File
	var err error

	stream, err = os.OpenFile(filename, os.O_CREATE|os.O_RDWR, 0775)
	defer stream.Close()
	if err != nil {
		fmt.Println("Couldn't dump: " + err.Error())
	}

	stream.Truncate(0)
	pusher := json.NewEncoder(stream)
	pusher.Encode(&c)
}

func (c *Channel) loadFromFile() *Channel {
	// retrieve previous state
	filename := c.getFilename()
	_, err := os.Stat(filename)
	if err != nil || os.IsNotExist(err) {
		return c
	}

	// load the file and parse it
	stream, err := os.OpenFile(filename, os.O_RDWR, 0775)
	if err != nil {
		fmt.Println("Awww shit: " + err.Error())
	}

	channel := new(Channel)
	parser := json.NewDecoder(stream)
	parser.Decode(&channel)

	// members is populated on join
	channel.Members = []string{}
	return channel
}

func (c *Channel) getFilename() string {
	channel := c.Channel
	for {
		if fmt.Sprintf("%q", channel[0]) == "'#'" {
			channel = channel[1:]
		} else {
			break
		}
	}

	return fmt.Sprintf("%s%s.channel", c.Logger, channel)
}
