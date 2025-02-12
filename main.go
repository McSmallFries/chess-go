package main

import (
	"chess-go/routes"
	"fmt"
	"math/rand"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/notnil/chess"
)

var port_http = ":8081"
var port_ws = ":8080"
var port_client = ":4200"
var domain = "http://localhost"
var version = "v0.0"
var title = "Chess !"

func ServerMain() {
	fmt.Println(title + " -- " + version)
	fmt.Println("wss " + " -- " + port_ws)
	e := echo.New()
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins: []string{domain + port_client},
		AllowHeaders: []string{echo.HeaderOrigin, echo.HeaderContentType, echo.HeaderAccept},
	}))
	routes.Initialize(e)
	e.GET("/", func(c echo.Context) error {
		return c.String(http.StatusOK, title)
	})
	e.Logger.Fatal(e.Start(port_http))

	
}

func GameMain() {
	game := chess.NewGame()
	// generate moves until game is over, demostrate a no outcome game, randomly.
	for game.Outcome() == chess.NoOutcome {
		// select a random move
		moves := game.ValidMoves()
		move := moves[rand.Intn(len(moves))]
		game.Move(move)
	}
	// print outcome and game PGN
	fmt.Println(game.Position().Board().Draw())
	fmt.Printf("Game completed. %s by %s.\n", game.Outcome(), game.Method())
	fmt.Println(game.String())
}

func main() {
	ServerMain()
	GameMain()
}

/*
	Chess Output:

	 A B C D E F G H
	8- - - - - - - -
	7- - - - - - ♚ -
	6- - - - ♗ - - -
	5- - - - - - - -
	4- - - - - - - -
	3♔ - - - - - - -
	2- - - - - - - -
	1- - - - - - - -

	Game completed. 1/2-1/2 by InsufficientMaterial.

	1.Nc3 b6 2.a4 e6 3.d4 Bb7 ...
*/
