package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"golang.org/x/net/websocket"
	"io"
	"net/http"
	"strconv"
	"time"
)

type player struct {
	Id int
	ch chan interface{}
}

type idAssignment struct {
	id int
}

type playerDisconnect struct {
	id int
}

// Find a unique ID for a player
func getNewPlayerId(players map[int]player) int {
	for i := 0; i < 10000000; i += 1 {
		if _, ok := players[i]; ok == false {
			return i
		}
	}
	panic("Unique ID not found") // Server should have exploded by now
}

// Send data to a particular player
func sendTo(p player, data interface{}) {
	p.ch <- data
}

// Send data to all players
func sendToAll(players map[int]player, data interface{}) {
	for _, p := range players {
		p.ch <- data
	}
}

// Send data to all players except a certain player
func sendToAllExcept(players map[int]player, data interface{}, except int) {
	for _, p := range players {
		if p.Id != except {
			p.ch <- data
		}
	}
}

func server(incoming chan interface{}) {
	fmt.Printf("Server started at %s\n", time.Now())

	// Map of players from ID to player object
	var players = make(map[int]player)

	// Count of online players
	var playerCount = 0

	for {
		var msg interface{}
		select {
		case msg = <-incoming:
			switch msg.(type) {

			// A new player has joined the server
			case player:

				// Increment the player count
				playerCount += 1

				// Pull player object from channel
				newPlayer := msg.(player)

				// Generate new ID for the player
				newPlayer.Id = getNewPlayerId(players)

				// Tell the player their new ID
				newPlayer.ch <- idAssignment{id: newPlayer.Id}

				// Add the player to the player list
				players[newPlayer.Id] = newPlayer

				fmt.Println("Server got a new player!", newPlayer)
				fmt.Println("Total number of players now:", len(players))

				// Notify all players about new player count
				sendToAll(players, playerCount)

			// A player has disconnected from the server
			case playerDisconnect:

				// Decrement the player count
				playerCount -= 1

				// Get the players ID from the channel
				id := msg.(playerDisconnect).id
				fmt.Println("Player:", id, "has disconnected")

				// Remove the player from the player list
				delete(players, id)
				fmt.Println("Total number of players now:", len(players))

				// Notify all players about new player count
				sendToAll(players, playerCount)

			default:
				fmt.Println("Unindentified type", msg)
			}
		}
	}
}

// Return a handler which will take a channel to listen on to send messages back to the client
func handlerGen(incoming chan interface{}) func(*websocket.Conn) {
	return func(ws *websocket.Conn) {
		fmt.Printf(
			"New connection from %s %s\n", ws.RemoteAddr().Network(),
			ws.RemoteAddr().String())

		// Socket to listen on for errors
		var socketErrorChan = make(chan error)

		go func() {
			dec := json.NewDecoder(ws)

			// Decode json while the socket is alive - Not currently used
			for {
				var m map[string]interface{}
				if err := dec.Decode(&m); err == io.EOF {
					break
				} else if err != nil {
					socketErrorChan <- err
				}
			}

			// Socket has closed
			socketErrorChan <- errors.New("EOF")
		}()

		// Player for this closure
		var thisPlayer player

		thisPlayer = player{
			Id: 0, // server will assign a unique id
			ch: make(chan interface{}, 10),
		}

		// Tell server about this player
		incoming <- thisPlayer

		var fromServer interface{}
		var fromSocketError error
		for {
			select {

			// Messages coming from the server
			case fromServer = <-thisPlayer.ch:
				switch fromServer.(type) {

				case int:
					var countStr = strconv.Itoa(fromServer.(int))
					ws.Write([]byte("{\"count\":" + countStr + "}"))

				// This player has been assigned an ID
				case idAssignment:
					thisPlayer.Id = fromServer.(idAssignment).id

				default:
					fmt.Println("Player", thisPlayer.Id, "got unidentified message from server", fromServer)
				}
			case fromSocketError = <-socketErrorChan:
				switch fromSocketError.Error() {

				// Player has quit/disconnected
				case "EOF":
					incoming <- playerDisconnect{id: thisPlayer.Id}

				default:
					fmt.Println("Socket error:", fromSocketError, "for player", thisPlayer)
				}
			}
		}
	}
}

func main() {
	// Channel to coordniate all users with the server
	var incoming = make(chan interface{}, 100)

	// Server
	go server(incoming)

	// Websocket listener
	http.Handle("/", websocket.Handler(handlerGen(incoming)))
	err := http.ListenAndServe(":12345", nil)
	if err != nil {
		panic("ListenAndServe: " + err.Error())
	}
}
