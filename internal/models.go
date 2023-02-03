package internal

import "time"

type Post struct {
	ID        int64     `db:"id"`
	Npub      string    `db:"npub"`
	Body      string    `db:"body"`
	RelayList string    `db:"relaylist"`
	CreatedAt time.Time `db:"created_at"`
}
