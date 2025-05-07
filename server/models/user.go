package models

import "time"

type User struct {
	ID       int       `json:"id"`
	Name     string    `json:"name"`
	Email    string    `json:"email"`
	Phone    string    `json:"phone"`
	Password string    `json:"password"`
	Role     string    `json:"role"` // Must be: admin, sub_admin, client
	Created  time.Time `json:"created"`
	Updated  time.Time `json:"updated"`
}
