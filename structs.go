package spacetraders

import (
	"fmt"
	"time"
)

type Client struct {
	username string
	token    string
	server   string
}

type StatusRes struct {
	Status string `json:"status"`
}

type ClaimRes struct {
	Token string `json:"token"`
	User  User   `json:"user"`
}

type AccountRes struct {
	User User `json:"user"`
}

type User struct {
	Credits        int       `json:"credits"`
	JoinedAt       time.Time `json:"joinedAt"`
	ShipCount      int       `json:"shipCount"`
	StructureCount int       `json:"structureCount"`
	Username       string    `json:"username"`
}

func (u *User) String() string {
	return fmt.Sprintf("%s: Credits: %d, Ships: %d, Structures: %d, Joined: %s",
		u.Username, u.Credits, u.ShipCount, u.StructureCount, u.JoinedAt)
}
