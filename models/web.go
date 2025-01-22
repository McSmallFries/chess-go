package models

import (
	"chess-go/database"
	"errors"

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

func (user *User) FindDBUser(db *sqlx.DB) (dbUser User, err error) {
	username := user.Username
	email := user.Email
	var dbID int32
	var dbUsername string
	var dbEmail string
	row := db.QueryRow(`SELECT u.idUser, usn.Username, e.Email FROM Users u JOIN Usernames usn ON 
	  u.idUser = usn.idUser WHERE u.Email = ? OR usn.Username = ?`,
		email, username)
	if err = row.Scan(&dbID); err != nil {
		return dbUser, err
	}
	if err = row.Scan(&dbUsername); err != nil {
		return dbUser, err
	}
	if err = row.Scan(&dbEmail); err != nil {
		return dbUser, err
	}
	dbUser.Id = dbID
	dbUser.Email = dbEmail
	dbUser.Username = dbUsername
	return dbUser, nil
}

// initializeWebsocket  { -> for feeding/routing the moves into the chess engine/client game

// }

type UserPassword struct {
	IdUser   int32  `json:"idUser" db:"idUser"`
	Password string `json:"password" db:"Password"`
}

func (password *UserPassword) FindDBUserPassword(db *sqlx.DB) (dbPassword string, err error) {
	id := password.IdUser
	row := db.QueryRow(`SELECT Password FROM UserPasswords WHERE idUser = ?`, id)
	if err = row.Scan(&dbPassword); err != nil {
		return dbPassword, err
	}
	return dbPassword, nil
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
	if dbUser, err := request.User.FindDBUser(db); err != nil {
		return false, err
	} else if !compareUserCreds(dbUser, request.User) {
		return false, errors.New("user credentials mismatch")
	}

	if dbPassword, err := request.Password.FindDBUserPassword(db); err != nil {
		return false, err
	} else if !compareUserPassword(dbPassword, request.Password.Password) {
		return false, nil
	}

	return true, nil
}

// Helpers (Private)
func compareUserCreds(c1 User, c2 User) bool {
	return c1.Email == c2.Email && c1.Username == c2.Username
}

func compareUserPassword(p1 string, p2 string) bool {
	return p1 == p2
}

// Delagates

func GetConnectionParams() database.ConnectionParams {
	return database.GetConnection().GetConnectionParams()
}

func SetConnectionParams(p database.ConnectionParams) {
	database.GetConnection().SetConnectionParams(p)
}
