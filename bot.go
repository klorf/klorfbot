package main

import (
	"fmt"
	"math/rand"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	irc "github.com/klorf/goirc/client"
)

func main() {
	cfg := irc.NewConfig("klorfbot")
	cfg.SSL = true
	cfg.Server = "morgan.freenode.net:6697"
	cfg.NewNick = func(n string) string { return n + "^" }
	cfg.Me.Ident = "klorfbot"
	cfg.Me.Name = "klorfbot"

	c := irc.Client(cfg)
	c.HandleFunc("connected", connect)

	quit := make(chan bool)
	c.HandleFunc("disconnected", func(conn *irc.Conn, line *irc.Line) { quit <- true })

	// set up some channel managers
	c.HandleFunc("privmsg", channel)

	if err := c.Connect(); err != nil {
		fmt.Printf("Connection error: %s\n", err)
	}

	<-quit
}

func connect(conn *irc.Conn, line *irc.Line) {
	p := os.Getenv("KLORF_PASS")
	conn.Privmsg("NickServ", fmt.Sprintf("IDENTIFY %s", p))

	conn.Join("#klorf")
	conn.Privmsg("#klorf", "klorf klorf klorf")
}

func channel(conn *irc.Conn, line *irc.Line) {
	c := line.Args[0]
	fmt.Printf("[%s] %s: %s {%s}.\n", line.Time, line.Nick, line.Args[1], c)

	/* kept for safe-keeping
	if line.Nick == "Seguer" && r.Intn(100) >= 99 {
		conn.Privmsg(c, "Dammit Seguer!")
	}
	*/

	roll(conn, line)
}

func roll(conn *irc.Conn, line *irc.Line) {
	c := line.Args[0]
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	m := tokens(line.Args[1])
	msg := "["

	if strings.ToLower(m[0]) != "!roll" {
		return
	}

	// need to parse the dice rolling
	d := regexp.MustCompile(`^(\d+)d(\d+)([\-\+\*\/]?)(\d*)$`)
	if !d.MatchString(m[1]) {
		conn.Privmsg(c, fmt.Sprintf("%s: Malformed dice roll [%s]", line.Nick, m[1]))
		return
	}

	x := d.FindStringSubmatch(m[1])
	count, _ := strconv.Atoi(x[1])
	if count < 1 {
		conn.Privmsg(c, fmt.Sprintf("%s: You need at least 1 die to roll", line.Nick))
		return
	} else if count > 30 {
		conn.Privmsg(c, fmt.Sprintf("%s: Limiting you to 30 die at once", line.Nick))
		count = 30
	}

	dice, _ := strconv.Atoi(x[2])
	if dice < 4 || dice > 20 || dice%2 != 0 {
		conn.Privmsg(c, fmt.Sprintf("%s: Invalid dice type", line.Nick))
		return
	}

	total := 0
	for i := 0; i < count; i++ {
		roll := r.Intn(dice-1) + 1

		total += roll
		msg = fmt.Sprintf("%s %d", msg, roll)
	}
	msg = fmt.Sprintf("%s ]", msg)

	t, _ := strconv.Atoi(x[4])
	if x[3] == "+" {
		msg = fmt.Sprintf("%s + %s", msg, x[4])
		total += t
	} else if x[3] == "-" {
		msg = fmt.Sprintf("%s - %s", msg, x[4])
		total -= t
	}

	conn.Privmsg(c, fmt.Sprintf("%s: %s = %d", line.Nick, msg, total))
}

func tokens(m string) []string {
	return strings.Split(m, " ")
}
