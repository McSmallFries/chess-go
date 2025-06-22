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

// You can edit this code!
// Click here and start typing.
package main

import (
	"fmt"
	"strconv"
)

type ListNode struct {
	Val  int
	Next *ListNode
}

func addTwoNumbers(l1 *ListNode, l2 *ListNode) *ListNode {
	var L1AsString = string(l1.Val)
	var L2AsString = string(l2.Val)
	for {
		if l1.Next == nil && l2.Next == nil {
			break
		}
		if l1.Next != nil {
			var curr = *l1
			L1AsString += string(curr.Val)
			l1 = curr.Next
		}
		if l2.Next != nil {
			var curr = *l2
			L2AsString += string(curr.Val)
			l2 = curr.Next
		}
	}
	L1StrReversed := ""
	L2StrReversed := ""
	SumAsString := ""
	for i := len(L1AsString) - 1; i > 0; i -= 1 {
		L1StrReversed += L1AsString[i]
	}
	for i := len(L2AsString) - 1; i > 0; i -= 1 {
		L2StrReversed += L2AsString[i]
	}
	// 342 + 465 => 807
	SumAsString = string(strconv.Atoi(L1StrReversed) + strconv.Atoi(L2StrReversed))
	var listToReturn = *ListNode{}
	var curr = *ListNode{}
	var next = *ListNode{}
	for i := len(SumAsString) - 1; i > 0; i -= 1 {
		curr.Val = strconv.Atoi(SumAsString[i])
		if i-1 > 0 {
			next.Val = strconv.Atoi(SumAsString[i-1])
			curr.Next = next
			listToReturn.Next = curr
		}
	}
	return listToReturn
}

func main() {
	fmt.Println("Hello, 世界")
}


func main() {
	// ServerMain()
	// GameMain()
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
