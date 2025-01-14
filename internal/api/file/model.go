package file

import "time"

type File struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Data      []byte    `json:"data"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
