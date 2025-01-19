package routes

import (
	"fmt"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/notnil/chess"
)

func HelloWorld(ctx echo.Context) error {
	fmt.Println("Hello World")
	return nil
}

func GetNewGame(ctx echo.Context /* player1 Player, player2 Player */) error {
	var player Player
	err := ctx.Bind(&player)
	if err != nil {
		return ctx.String(http.StatusBadRequest, "bad request")
	}
	warehouse := Singleton.GetLobbyWarehouse()
	lobby, err := warehouse.FindAndJoinLobby(player)
	if err != nil {
		return err
	}
	if !lobby.InProgress {
		ctx.String(http.StatusBadRequest, "lobby not started") // not a bad request - find better http code.
	}
	game := OnlineGame{
		Lobby: lobby,
		Game:  chess.NewGame(),
	}
	fmt.Println("player joining lobby")
	return ctx.JSON(http.StatusOK, game)

}

func Register(ctx echo.Context) error {
	var user User
	err := ctx.Bind(&user)
	if err != nil {
		return ctx.JSON(http.StatusBadRequest, err)
	}
	// register

	// login

	return ctx.JSON(http.StatusOK, user)
}

func Login(ctx echo.Context) error {
	var user User
	err := ctx.Bind(&user)
	if err != nil {
		return ctx.JSON(http.StatusBadRequest, err)
	}
	// login

	return ctx.JSON(http.StatusOK, user)
}

func HookupRoutes(e *echo.Echo) {
	addPlayerToLobbyAndStartIfFull := echo.HandlerFunc(GetNewGame)
	helloWorld := echo.HandlerFunc(HelloWorld)
	login := echo.HandlerFunc(Login)
	register := echo.HandlerFunc(Register)
	e.GET("/game/new", addPlayerToLobbyAndStartIfFull)
	e.GET("/hello", helloWorld)

	e.POST("/user/login", login)
	e.POST("/user/register", register)

}
