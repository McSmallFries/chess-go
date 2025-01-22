package models

import (
	"chess-go/database"

	"github.com/jmoiron/sqlx"
)

type User struct {
	Id       int32  `json:"idUser" db:"idUser"`
	Email    string `json:"email" db:"Email"`
	Username string `json:"username" db:"Username"`
}

func (user *User) Insert(db *sqlx.DB) error {
	userResult := db.MustExec(`INSERT INTO Users (Email) VALUES (?)`, user.Email)
	id, err := userResult.LastInsertId()
	if err != nil {
		return err
	}
	user.Id = int32(id)
	_ = db.MustExec(`INSERT INTO Usernames (idUser, Username) VALUES (?,?)`, user.Id, user.Username)
	return nil
}

// initializeWebsocket  { -> for feeding/routing the moves into the chess engine/client game

// }

type UserPassword struct {
	IdUser   int32  `json:"idUser" db:"idUser"`
	Password string `json:"password" db:"Password"`
}

func (pw *UserPassword) Insert(db *sqlx.DB) error {
	result := db.MustExec(`INSERT INTO UserPasswords (idUser, Password) VALUES (?, ?)`, pw.IdUser, pw.Password)
	_, err := result.LastInsertId()
	if err != nil {
		return err
	}
	return nil
}

type LoginRequest struct {
	User     User         `json:"user"`
	Password UserPassword `json:"userPassword"`
}

func NewLoginRequest() LoginRequest {
	return LoginRequest{User: User{}, Password: UserPassword{}}
}

func (request *LoginRequest) Register(db *sqlx.DB) (bool, error) {
	if request.User.Id > 0 {
		return false, nil
	}
	var user *User = &request.User
	if err := user.Insert(db); err != nil {
		return false, err
	}
	request.Password.IdUser = user.Id
	request.User.Id = user.Id
	var password *UserPassword = &request.Password
	if err := password.Insert(db); err != nil {
		return false, err
	}
	return true, nil
}

func (request *LoginRequest) Login(db *sqlx.DB) (bool, error) {
	if request.User.Id <= 0 {
		return false, nil
	}
	var copyRequest *LoginRequest = new(LoginRequest)
	rows, err := db.Queryx(`Select * From Users Where idUser = ?`, request.User.Id)
	if err != nil {
		return false, err
	}

	for rows.Next() {
		_ = rows.Scan(copyRequest)
		// request.User.Id == copyRequest.User.Id
	}
	return true, nil
}

// Delagates

func GetConnectionParams() database.ConnectionParams {
	return database.GetConnection().GetConnectionParams()
}

func SetConnectionParams(p database.ConnectionParams) {
	database.GetConnection().SetConnectionParams(p)
}
