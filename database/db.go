package database

import (
	"fmt"
	"log"

	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
)

type ConnectionParams struct {
	ConnectionString string
}

type Connection struct {
	db     *sqlx.DB
	params ConnectionParams
}

func (c *Connection) SetDB(db *sqlx.DB) {
	c.db = db
}

func (c *Connection) GetDB() *sqlx.DB {
	return c.db
}

func (c *Connection) SetConnectionParams(params ConnectionParams) {
	c.params = params
}

func (c *Connection) GetConnectionParams() ConnectionParams {
	return c.params
}

var connection = Connection{}

func Initialize(params ConnectionParams) {
	conn := GetConnection()
	conn.SetConnectionParams(params)
	conn.Connect()

}

func GetConnection() *Connection {
	return &connection
}

func (c *Connection) Connect() {
	db, err := sqlx.Connect("mysql", c.params.ConnectionString)
	if err != nil {
		log.Fatalln(err)
	}
	fmt.Println("Connected to DB")
	GetConnection().SetDB(db)
}
