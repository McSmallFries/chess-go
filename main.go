package main

import (
	"errors"
	"fmt"
	"math/rand"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/notnil/chess"
)

type GrandMasterSingleton struct {
	LobbyWarehouse *LobbyWarehouse
	AllGames       *[]OnlineGame
}

func (gm *GrandMasterSingleton) MakeLobbyWarehouse() *LobbyWarehouse {
	lobbies := make([]Lobby, 0)
	gm.LobbyWarehouse = &LobbyWarehouse{Lobbies: lobbies}
	return gm.GetLobbyWarehouse()
}

func (gm *GrandMasterSingleton) MakeAllGames() *[]OnlineGame {
	games := make([]OnlineGame, 0)
	gm.AllGames = &games
	return gm.GetAllGames()
}

func (gm *GrandMasterSingleton) GetAllGames() *[]OnlineGame {
	return gm.AllGames
}

func (gm *GrandMasterSingleton) GetLobbyWarehouse() *LobbyWarehouse {
	// fmt.Println(*gm)
	return gm.LobbyWarehouse
}

func (gm *GrandMasterSingleton) AddLobby(lobby Lobby) {
	gm.GetLobbyWarehouse().Lobbies = append(gm.GetLobbyWarehouse().Lobbies, lobby)
}

var Singleton = &GrandMasterSingleton{}

type Player struct {
	Name string `db:"name" json:"name"`
	Id   int64  `db:"id" json:"id"`
}

func (player *Player) PromoteToHost(ipAddr string) (*Host, error) {

	host := Host{
		Id:        1, //playerId (might need HostID too)
		IpAddress: ipAddr,
	}

	return &host, nil

	// use Echo to ping IP, if error -> return error.

	// return &Host{}, nil
}

type OnlineGame struct {
	GameID int64       `db:"gameId" json:"gameId"`
	Game   *chess.Game `db:"game" json:"game"`
	Lobby  *Lobby      `db:"lobby" json:"lobby"`
}

type OnlineGameFinder struct {
	GameId  int64 `db:"gameId" json:"gameId"`
	LobbyId int64 `db:"lobbyId" json:"lobbyId"`
}

func (*OnlineGameFinder) FindOnlineGameByLobbyID(lobbyId int64) (*OnlineGame, error) {
	var allGames []OnlineGame = *Singleton.GetAllGames()
	var game OnlineGame
	found := false
	for _, v := range allGames {
		if v.Lobby.LobbyID == lobbyId {
			game = v
			found = true
			break
		}
	}

	if !found {
		return &OnlineGame{}, errors.New("no game found")
	}
	return &game, nil
}

func (*OnlineGameFinder) FindOnlineGameByGameID(gameId int64) (*OnlineGame, error) {
	var allGames []OnlineGame = *Singleton.GetAllGames()
	var game OnlineGame
	found := false
	for _, v := range allGames {
		if v.GameID == gameId {
			game = v
			found = true
			break
		}
	}

	if !found {
		return &OnlineGame{}, errors.New("no game found")
	}
	return &game, nil
}

type User struct {
	Id       int64  `json:"id"`
	Email    string `json:"email"`
	Username string `json:"username"`
	Password string `json:"password"`
}

type Host struct {
	Id        int64  `db:"id" json:"id"`
	IpAddress string `db:"ipAddress" json:"ipAddress"`
}

type LobbyWarehouse struct {
	Lobbies []Lobby `db:"lobbies" json:"lobbies"` // -> select idLobby from lobbies as 'lobbies'
}

func (warehouse *LobbyWarehouse) FindAndJoinLobby(player Player) (*Lobby, error) {
	playerIsSorted := false
	for _, lobby := range warehouse.Lobbies {
		if !lobby.IsFull {
			playerIsSorted = true
			lobby, err := lobby.AddPlayer(player)
			if err != nil {
				return &Lobby{}, errors.New("could not join lobby")
			}
			lobby.Start()
			return lobby, nil
		}

		if !lobby.InProgress && lobby.IsFull {
			lobby.Start() // give it a kickstart.
		}
	}

	if !playerIsSorted {
		lobby := Lobby{}
		l, err := lobby.AddPlayer(player)
		if err != nil {
			l.IsCompletelyFucked = true
			return l, err
		}
		Singleton.AddLobby(*l)
		return l, nil
	}
	return &Lobby{IsCompletelyFucked: true}, nil
}

type Lobby struct {
	LobbyID            int64  `db:"lobbyId" json:"lobbyId"`
	Host               Host   `db:"host" json:"host"`
	IsFull             bool   `db:"isFull" json:"isFull"`
	InProgress         bool   `db:"inProgress" json:"inProgress"` // periodically update and confirm health of lobby.
	Player1            Player `db:"player1" json:"player1"`
	Player2            Player `db:"player2" json:"player2"`
	IsCompletelyFucked bool
}

func (lobby *Lobby) AddPlayer(player Player) (*Lobby, error) {
	if lobby.IsFull {
		host, err := player.PromoteToHost("get.ip.of.player")
		return &Lobby{
			Host:    *host,
			IsFull:  false,
			Player1: player,
			Player2: Player{},
		}, err
	}
	lobby.Player2 = player
	return lobby, nil
}

func (lobby *Lobby) Start() {
	// make a POST http request and put both players into a new game with their own angular interface
	//  if error, log, & put both players back in home and try again.
	didSuccessfullyStart := true
	lobby.InProgress = didSuccessfullyStart

}

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

func hookupRoutes(e *echo.Echo) {
	addPlayerToLobbyAndStartIfFull := echo.HandlerFunc(GetNewGame)
	helloWorld := echo.HandlerFunc(HelloWorld)
	login := echo.HandlerFunc(Login)
	register := echo.HandlerFunc(Register)
	e.GET("/game/new", addPlayerToLobbyAndStartIfFull)
	e.GET("/hello", helloWorld)

	e.POST("/user/login", login)
	e.POST("/user/register", register)

}

// initializeWebsocket  { -> for feeding/routing the moves into the chess engine/client game

// }

func ServerMain() {
	//startServer()
	e := echo.New()
	hookupRoutes(e)
	// Middleware
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins: []string{"http://localhost:4200"},
		AllowHeaders: []string{echo.HeaderOrigin, echo.HeaderContentType, echo.HeaderAccept},
	}))
	e.GET("/", func(c echo.Context) error {
		return c.String(http.StatusOK, "Hello, World!1111")
	})
	Singleton.MakeLobbyWarehouse()
	Singleton.MakeAllGames()

	e.Logger.Fatal(e.Start(":1323"))
}

func main() {
	ServerMain()
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
	/*
		Output:

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
}
