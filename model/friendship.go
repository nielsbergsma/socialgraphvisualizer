package model

type Friendship struct {
	From *Actor `json:"from"`
	To *Actor `json:"to"`
}