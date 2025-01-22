package models

import (
	"errors"

	"github.com/notnil/chess"
)

var singleton = &GrandMasterSingleton{}

func GetSingleton() *GrandMasterSingleton {
	return singleton
}

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
		GetSingleton().AddLobby(*l)
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
