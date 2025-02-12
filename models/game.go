package models

import (
	"errors"
	"fmt"

	"github.com/jmoiron/sqlx"
	"github.com/notnil/chess"
	"golang.org/x/net/websocket"
)

var singleton = &GrandMasterSingleton{}

func GetSingleton() *GrandMasterSingleton {
	return singleton
}

type GrandMasterSingleton struct {
	LobbyWarehouse *LobbyWarehouse
	AllGames       *[]OnlineGame
	GameSockets    *[]ChessGameSocket
}

func (gm *GrandMasterSingleton) NewSocket() {
	// set up a new socket connection and add to the singleton
	// will have to use WaitGroup so that the server doesnt get stuck in
	// fucking for loops man.

	// there will be a nice video on that somewhere
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

type Player struct {
	PlayerID int64 `json:"playerID" db:"PlayerID"`
	UserID   int32 `json:"idUser" db:"idUser"`
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

type ChessGameSocket struct {
	Id int64 `json:"socketId" db:"SocketID"`
	Ws *websocket.Conn
}

type OnlineGame struct {
	GameID   int64       `db:"gameId" json:"gameId"`
	Game     *chess.Game `db:"game" json:"game"`
	SocketID *int64      `db:"SocketID" json:"socketId"`
	Lobby    *Lobby      `db:"lobby" json:"lobby"`
}

type OnlineGameFinder struct {
	GameId  int64 `db:"gameId" json:"gameId"`
	LobbyId int64 `db:"lobbyId" json:"lobbyId"`
}

func (*OnlineGameFinder) FindOnlineGameByLobbyID(lobbyId int64) (*OnlineGame, error) {
	var allGames []OnlineGame = *GetSingleton().GetAllGames()
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
	var allGames []OnlineGame = *GetSingleton().GetAllGames()
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

type Host struct {
	Id        int64  `db:"id" json:"id"`
	IpAddress string `db:"ipAddress" json:"ipAddress"`
}

func (host *Host) GetHostByID(id int64, db *sqlx.DB) (err error) {
	query := `SELECT * FROM Hosts WHERE HostID = ?`
	if err = db.Select(host, query, id); err != nil {
		return err
	}
	if !(host.Id > 0) {
		return errors.New("something went wrong obtaining host")
	}
	return nil
}

type LobbyWarehouse struct {
	Lobbies []Lobby `db:"lobbies" json:"lobbies"` // -> select idLobby from lobbies as 'lobbies'
}

func (warehouse *LobbyWarehouse) FixLobbyWarehouse(err error) {
	fmt.Println(err)
	// iterate over lobbys and check common failure points
	// ids, isCompletelyFucked
	// fix or just put player(s) in a new lobby, clear up the broken one and carry on
}

func (warehouse *LobbyWarehouse) UpdateLobbyWarehouse(db *sqlx.DB) (err error) {
	var lobbies []Lobby
	var lobbyIds []int64
	var rows *sqlx.Rows
	query := `SELECT LobbyID, Lobbies.* from Lobbies WHERE IsComplete`
	if rows, err = db.Queryx(query); err != nil {
		return err
	}
	if err = rows.Scan(&lobbies, &lobbyIds); err != nil {
		return err
	}
	warehouse.Lobbies = lobbies
	for _, lobby := range lobbies {
		if err = lobby.Host.GetHostByID(lobby.HostID, db); err != nil {
			warehouse.FixLobbyWarehouse(err)
		}
	}
	return nil
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
		GetSingleton().AddLobby(*l)
		return l, nil
	}
	return &Lobby{IsCompletelyFucked: true}, nil
}

type Lobby struct {
	LobbyID            int64  `db:"LobbyID" json:"lobbyId"`
	HostID             int64  `db:"HostID"`
	Host               Host   `json:"host"`
	IsFull             bool   `db:"IsFull" json:"isFull"`
	InProgress         bool   `json:"inProgress"` // periodically update and confirm health of lobby.
	Player1            Player `db:"IDPlayer1" json:"player1"`
	Player2            Player `db:"IDPlayer2" json:"player2"`
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
