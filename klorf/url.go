package klorf

import (
	"fmt"
	"time"
)

type Url struct {
	User string    `json:"user"`
	Date time.Time `json:"date"`
	Url  string    `json:"url"`
}

func (u *Url) IsRobot(url, user string) bool {
	if u.Url == url && u.User != user {
		return true
	}

	return false
}

func (u *Url) IsSimilar(url string) bool {
	if u.Url == url {
		return true
	}
	return false
}

func (u *Url) GetRobotString() string {
	return fmt.Sprintf("ROBOT: (%v ago by %s)", time.Now().Sub(u.Date), u.User)
}

func NewUrl(user, url string) *Url {
	u := new(Url)
	u.User = user
	u.Url = url
	u.Date = time.Now()

	return u
}
