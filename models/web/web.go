package models

type User struct {
	Id       int64  `json:"id"`
	Email    string `json:"email"`
	Username string `json:"username"`
	Password string `json:"password"`
}

// initializeWebsocket  { -> for feeding/routing the moves into the chess engine/client game

// }
