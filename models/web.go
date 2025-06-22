package models

import (
	"chess-go/database"
	"chess-go/utils"

	"github.com/jmoiron/sqlx"
)

type User struct {
	Id       int32  `json:"idUser" db:"IDUser"`
	Email    string `json:"email" db:"Email"`
	Username string `json:"username" db:"Username"`
}

// TODO - Use db.Begin() to start a Tx and use Commit() / Rollback()
// for database control here.
func (user *User) Insert(db *sqlx.DB) error {
	userResult := db.MustExec(`INSERT INTO Users (Email) VALUES (?)`, user.Email)
	id, err := userResult.LastInsertId()
	if err != nil {
		return err
	}
	user.Id = int32(id)
	_ = db.MustExec(`INSERT INTO Usernames (IDUser, Username) VALUES (?,?)`, user.Id, user.Username)
	return nil
}

// TODO - Use db.Begin() to start a Tx and use Commit() / Rollback()
// for database control here.
func (user *User) FindDBUser(db *sqlx.DB) (dbUser User, err error) {
	username := user.Username
	email := user.Email
	var dbID int32
	var dbUsername string
	var dbEmail string
	row := db.QueryRow(`SELECT u.IDUser as IDUser, usn.Username as Username, u.Email as Email FROM Users u JOIN Usernames usn ON 
	  u.IDUser = usn.IDUser WHERE u.Email = ? OR usn.Username = ?`,
		email, username)
	if err = row.Scan(&dbID, &dbUsername, &dbEmail); err != nil {
		return dbUser, err
	} else {
		dbUser.Id = dbID
		dbUser.Email = dbEmail
		dbUser.Username = dbUsername
	}

	return dbUser, nil
}

// initializeWebsocket  { -> for feeding/routing the moves into the chess engine/client game

// }

type UserPassword struct {
	IdUser   int32  `json:"idUser" db:"IDUser"`
	Password string `json:"password" db:"Password"`
}

// TODO - Use db.Begin() to start a Tx and use Commit() / Rollback()
// for database control here.
func (password *UserPassword) ValidateDBUserPassword(db *sqlx.DB) (match bool, err error) {
	var dbPassword string
	id := password.IdUser
	pwToVerify := password.Password
	row := db.QueryRow(`SELECT Password FROM UserPasswords WHERE IDUser = ?`, id)
	if err = row.Scan(&dbPassword); err != nil {
		return false, err
	}
	match = utils.VerifyPassword(pwToVerify, dbPassword)
	return match, nil
}

// TODO - Use db.Begin() to start a Tx and use Commit() / Rollback()
// for database control here.
func (pw *UserPassword) Insert(db *sqlx.DB) error {
	hash, err := utils.HashPassword(pw.Password)
	if err != nil {
		return err
	}
	result := db.MustExec(`INSERT INTO UserPasswords (IDUser, Password) VALUES (?, ?)`, pw.IdUser, hash)
	_, err = result.LastInsertId()
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

// TODO - Use db.Begin() to start a Tx and use Commit() / Rollback()
// for database control here.
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

// TODO - Use db.Begin() to start a Tx and use Commit() / Rollback()
// for database control here.
func (request *LoginRequest) Login(db *sqlx.DB) (bool, error) {
	if request.User.Id <= 0 {
		return false, nil
	}
	if _, err := request.User.FindDBUser(db); err != nil {
		return false, err
	}

	if dbPassword, err := request.Password.ValidateDBUserPassword(db); err != nil {
		return false, err
	} else {
		return dbPassword, nil
	}
}

// Helpers (Private)
// func compareUserCreds(c1 User, c2 User) bool {
// 	return c1.Email == c2.Email || c1.Username == c2.Username
// }

// func compareUserPassword(p1 string, p2 string) bool {
// 	return p1 == p2
// }

// Delagates

func GetConnectionParams() database.ConnectionParams {
	return database.GetConnection().GetConnectionParams()
}

func SetConnectionParams(p database.ConnectionParams) {
	database.GetConnection().SetConnectionParams(p)
}
