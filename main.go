package main

import (
	"log"
	"snake-game-distributed/database"
	"snake-game-distributed/game"
)

func main() {
	database.ConnectDB()

	game.InitDB()

	if !database.DBConnected {
		log.Println("Atenção: Jogo iniciando em modo local (sem persistência de scores).")
	} else {
		game.DBConnected = true
	}
	game.NewGame().Start()
}
