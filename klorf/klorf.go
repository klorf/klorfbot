package klorf

import (
	"errors"
	"fmt"
	"log"
	"math/rand"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	irc "github.com/klorf/goirc/client"
)

type Klorf struct {
	Logger   string     `json:"logger"`
	Channels []*Channel `json:"channels"`
}

func NewKlorf() *Klorf {
	k := new(Klorf)

	return k
}

func New(log string) *Klorf {
	k := new(Klorf)
	k.Logger = log

	return k
}

func (k *Klorf) Roll(conn *irc.Conn, line *irc.Line) {
	c := line.Args[0]

	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	m := tokens(line.Args[1])
	d := regexp.MustCompile(`\b(\d+)d(\d+)([\-\+\*\/]?)(\d*)`)
	msg := ""
	matched := false

	for _, y := range m {
		if d.MatchString(y) {
			matched = true

			roll, err := k.runRoll(d.FindStringSubmatch(y), r)
			if err != nil {
				conn.Privmsg(c, fmt.Sprintf("%s: %s", line.Nick, err.Error()))
			}
			if roll != "" {
				msg = fmt.Sprintf("%s %s", msg, roll)
			} else {
				msg = fmt.Sprintf("%s %s", msg, y)
			}
		} else {
			msg = fmt.Sprintf("%s %s", msg, y)
		}
	}

	if matched {
		msg = fmt.Sprintf("%s:%s", line.Nick, msg)

		chanLog := string(line.Args[0][1:])
		k.logToFile(chanLog, msg, line.Time)

		conn.Privmsg(c, msg)
	}
}

func (k *Klorf) Log(conn *irc.Conn, line *irc.Line) {
	entry := fmt.Sprintf("%s: %s", line.Nick, line.Args[1])

	c := string(line.Args[0][1:])
	k.logToFile(c, entry, line.Time)
}

func (k *Klorf) Joined(conn *irc.Conn, line *irc.Line) {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	if line.Args[0] == "#klorf" && (line.Nick == "debnath" || line.Nick == "debtNath" || line.Nick == "debnath1") {
		if r.Intn(101)+1 > 89 {
			conn.Privmsg(line.Args[0], fmt.Sprintf("%s: Roll for Perception!", line.Nick))
		}
	}

	for _, i := range k.Channels {
		if i.Channel == line.Args[0] {
			i.Add(line.Nick)
		}
	}

	c := string(line.Args[0][1:])
	k.logToFile(c, fmt.Sprintf("%s joined %s", line.Nick, line.Args[0]), line.Time)
}

func (k *Klorf) Parted(conn *irc.Conn, line *irc.Line) {
	for _, i := range k.Channels {
		if i.Channel == line.Args[0] {
			i.Remove(line.Nick)
		}
	}

	k.logToFile(line.Args[0][1:], fmt.Sprintf("%s %s %s", line.Nick, strings.ToLower(line.Cmd), line.Args[0]), line.Time)
}

func (k *Klorf) Quit(conn *irc.Conn, line *irc.Line) {
	for _, i := range k.Channels {
		if i.HasUser(line.Nick) {
			i.Remove(line.Nick)

			k.logToFile(i.Channel, fmt.Sprintf("%s has quit.", line.Nick), line.Time)
		}
	}
}

func (k *Klorf) List(conn *irc.Conn, line *irc.Line) {
	var c *Channel
	for _, i := range k.Channels {
		if i.Channel == line.Args[2] {
			c = i
			break
		}
	}

	for _, nick := range line.Args[3:] {
		c.Add(nick)
	}
}

func (k *Klorf) runRoll(in []string, r *rand.Rand) (string, error) {
	var err error
	msg := "["
	total := 0

	diceCount, _ := strconv.Atoi(in[1])
	if diceCount < 1 {
		return "", errors.New("Too little die")
	} else if diceCount > 30 {
		diceCount = 30
	}

	diceType, _ := strconv.Atoi(in[2])
	if (diceType != 2 && diceType != 100) && (diceType < 4 || diceType > 20 || diceType%2 != 0) {
		return "", nil
	}

	for i := 0; i < diceCount; i++ {
		roll := r.Intn(diceType) + 1

		total += roll
		msg = fmt.Sprintf("%s %d", msg, roll)
	}
	msg = fmt.Sprintf("%s ]", msg)

	t, _ := strconv.Atoi(in[4])
	if in[3] == "+" {
		msg = fmt.Sprintf("%s + %s", msg, in[4])
		total += t
	} else if in[3] == "-" {
		msg = fmt.Sprintf("%s - %s", msg, in[4])
		total -= t
	}
	msg = fmt.Sprintf("(%s = %d)", msg, total)

	return msg, err
}

func (k *Klorf) logToFile(channel, message string, t time.Time) {
	for {
		if fmt.Sprintf("%q", channel[0]) == "'#'" {
			channel = channel[1:]
		} else {
			break
		}
	}

	f := fmt.Sprintf("%s%s_%d-%d-%d.txt", k.Logger, channel, t.Year(), t.Month(), t.Day())
	fh, _ := os.OpenFile(f, os.O_CREATE|os.O_RDWR|os.O_APPEND, 0775)
	defer fh.Close()

	lfile := log.New(fh, "", log.LstdFlags)
	lfile.Println(message)
}

func tokens(m string) []string {
	return strings.Split(m, " ")
}
