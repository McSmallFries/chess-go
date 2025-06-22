package routes

import (
	"chess-go/database"
	"chess-go/models"
	"fmt"
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"
	"github.com/notnil/chess"
	"github.com/robfig/cron/v3"
)

func Initialize(e *echo.Echo) {
	initializeRoutes(e)
	models.GetSingleton().MakeLobbyWarehouse()
	models.GetSingleton().MakeAllGames()
	params := models.GetConnectionParams()
	params.ConnectionString = `root:@(localhost:3306)/chess-db`
	database.Initialize(params)
	c := cron.New()
	c.AddFunc("@every 5s", func() {
		models.GetSingleton().GetLobbyWarehouse().UpdateLobbyWarehouse(
			database.GetConnection().GetDB().MustBegin())
	})
}

func HelloWorld(ctx echo.Context) error {
	fmt.Println("Hello World")
	return nil
}

func GetNewGame(ctx echo.Context) (err error) {
	var id int
	i := ctx.QueryParam("idUser")
	id, err = strconv.Atoi(i)
	if err != nil {
		return ctx.JSON(http.StatusBadRequest, err)
	}
	var player models.Player = *new(models.Player)
	player.UserID = int32(id)
	warehouse := models.GetSingleton().GetLobbyWarehouse()
	lobby, err := warehouse.FindAndJoinLobby(player)
	if err != nil {
		return ctx.JSON(http.StatusBadRequest, err)
	}
	game := models.OnlineGame{
		Lobby: lobby,
		Game:  chess.NewGame(),
	}
	gameJson := models.OnlineGame{
		LobbyID: lobby.LobbyID,
		GameID:  game.GameID,
	}
	gameJson.Insert(database.GetConnection().GetDB())
	game.GameID = gameJson.GameID
	game.LobbyID = gameJson.LobbyID
	models.GetSingleton().AddNewGame(game)
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

func wsHandler(ctx echo.Context) (err error) {
	var id int
	i := ctx.QueryParam("gameId")
	id, err = strconv.Atoi(i)
	if err != nil {
		return ctx.JSON(http.StatusBadRequest, err)
	}
	rw := ctx.Response().Writer
	req := ctx.Request()
	if err = models.WsHandler(rw, req, int64(id)); err != nil {
		fmt.Println(err)
	}
	return ctx.JSON(http.StatusOK, id)
}

func initializeRoutes(e *echo.Echo) {
	addPlayerToLobbyAndStartIfFull := echo.HandlerFunc(GetNewGame)
	helloWorld := echo.HandlerFunc(HelloWorld)
	login := echo.HandlerFunc(Login)
	register := echo.HandlerFunc(Register)
	e.GET("ws/game/:id", wsHandler)
	e.GET("/game/new", addPlayerToLobbyAndStartIfFull)
	e.GET("/hello", helloWorld)

	e.POST("/user/login", login)
	e.POST("/user/register", register)

}

// func GetPlayer(idUser int64) models.Player {
// 	db := database.GetConnection().GetDB()

// }
