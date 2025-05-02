package models

import "time"

type User struct {
	ID       int       `json:"id"`
	Name     string    `json:"name"`
	Email    string    `json:"email"`
	Phone    string    `json:"phone"`
	Password string    `json:"password"`
	Created  time.Time `json:"created"`
	Updated  time.Time `json:"updated"`
}

// var Users = []User{
// 	{ID: 1, Name: "Alice", Email: "alice@example.com"},
// 	{ID: 2, Name: "Bob", Email: "bob@example.com"},
// }
