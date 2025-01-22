package routes

import (
	"chess-go/database"
	"chess-go/models"
	"fmt"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/notnil/chess"
)

func Initialize(e *echo.Echo) {
	initializeRoutes(e)
	models.GetSingleton().MakeLobbyWarehouse()
	models.GetSingleton().MakeAllGames()
	params := models.GetConnectionParams()
	params.ConnectionString = `root:@(localhost:3306)/chess-db`
	database.Initialize(params)
}

func HelloWorld(ctx echo.Context) error {
	fmt.Println("Hello World")
	return nil
}

func GetNewGame(ctx echo.Context /* player1 Player, player2 Player */) error {
	var player models.Player = *new(models.Player)
	err := ctx.Bind(&player)
	if err != nil {
		return ctx.String(http.StatusBadRequest, "bad request")
	}
	warehouse := models.GetSingleton().GetLobbyWarehouse()
	lobby, err := warehouse.FindAndJoinLobby(player)
	if err != nil {
		return err
	}
	if !lobby.InProgress {
		ctx.String(http.StatusBadRequest, "lobby not started") // not a bad request - find better http code.
	}
	game := models.OnlineGame{
		Lobby: lobby,
		Game:  chess.NewGame(),
	}
	fmt.Println("player joining lobby")
	return ctx.JSON(http.StatusOK, game)
}

func Register(ctx echo.Context) error {
	request := new(models.LoginRequest)
	err := ctx.Bind(request)
	if err != nil {
		return ctx.JSON(http.StatusBadRequest, err)
	}
	// register
	db := database.GetConnection().GetDB()
	if result, err := request.Register(db); err != nil {
		return ctx.JSON(http.StatusInternalServerError, err)
	} else if !result {
		fmt.Println("Unable to register. Bad gateway.")
		return ctx.JSON(http.StatusBadGateway, "user id not zero")
	}
	return ctx.JSON(http.StatusOK, request.User)
}

func Login(ctx echo.Context) error {
	request := *new(models.LoginRequest)
	err := ctx.Bind(&request)
	if err != nil {
		return ctx.JSON(http.StatusBadRequest, err)
	}
	db := database.GetConnection().GetDB()

	if dbUser, err := request.User.FindDBUser(db); err != nil {
		return ctx.JSON(http.StatusBadGateway, "cannot find user by this email/username in database.")
	} else {
		request.User.Id = dbUser.Id
		request.Password.IdUser = dbUser.Id
	}

	if result, err := request.Login(db); err != nil {
		return ctx.JSON(http.StatusInternalServerError, err)
	} else if !result {
		fmt.Println("Unable to login. Bad gateway.")
		return ctx.JSON(http.StatusBadGateway, "user id cannot be zero")
	}
	return ctx.JSON(http.StatusOK, request.User)
}

func initializeRoutes(e *echo.Echo) {
	addPlayerToLobbyAndStartIfFull := echo.HandlerFunc(GetNewGame)
	helloWorld := echo.HandlerFunc(HelloWorld)
	login := echo.HandlerFunc(Login)
	register := echo.HandlerFunc(Register)
	e.GET("/game/new", addPlayerToLobbyAndStartIfFull)
	e.GET("/hello", helloWorld)

	e.POST("/user/login", login)
	e.POST("/user/register", register)

}
