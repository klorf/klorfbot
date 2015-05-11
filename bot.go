package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/CerealBoy/klorfbot/klorf"
	irc "github.com/klorf/goirc/client"
)

func main() {
	cfg := irc.NewConfig("klorfbot")
	cfg.SSL = true
	cfg.Server = "morgan.freenode.net:6697"
	cfg.NewNick = func(n string) string { return n + "^" }
	cfg.Me.Ident = "klorfbot"
	cfg.Me.Name = "klorfbot"

	log := os.Getenv("KLORF_LOGFILE")
	k := klorf.New(log)

	c := irc.Client(cfg)
	c.HandleFunc("connected", connect)

	quit := make(chan bool)
	c.HandleFunc("disconnected", func(conn *irc.Conn, line *irc.Line) { quit <- true })

	c.HandleFunc("privmsg", k.Log)
	c.HandleFunc("privmsg", k.Roll)

	if err := c.Connect(); err != nil {
		fmt.Printf("Connection error: %s\n", err)
	}

	<-quit
}

func connect(conn *irc.Conn, line *irc.Line) {
	p := os.Getenv("KLORF_PASS")
	conn.Privmsg("NickServ", fmt.Sprintf("IDENTIFY %s", p))

	c := os.Getenv("KLORF_CHANS")
	for _, x := range strings.Split(c, ":") {
		conn.Join(x)
		conn.Privmsg(x, "klorf klorf klorf")
	}
}
