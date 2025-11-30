// db.go
package game

import (
	"context"
	"log"
	"snake-game-distributed/database"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Score struct {
	PlayerID primitive.ObjectID `bson:"_id,omitempty"`
	Nome     string             `bson:"nome"`
	Pontos   int                `bson:"pontos"`
	Data     time.Time          `bson:"data"`
}

var (
	scoresCollection  *mongo.Collection
	playersCollection *mongo.Collection
	isDocker          bool
	DBConnected       bool = false // AQUI! indica se o MongoDB ta online
)

func InitDB() {

	if database.DBConnected {
		db := database.Client.Database("trabalho")

		scoresCollection = db.Collection("snake_scores")
		playersCollection = db.Collection("players")

		DBConnected = true
		log.Println("Collections prontas: snake_scores e players")
	} else {
		log.Println("Modo local detectado — scores serão salvos apenas na sessão (sem MongoDB)")
	}
}

// db.go  ← substitua a função antiga por esta
func SaveScore(player *Player, points int) {
	if player == nil {
		log.Println("SaveScore: player é nil, ignorando")
		return
	}

	// 1. Atualiza o melhor score do jogador (em memória + banco)
	player.UpdateBestScore(points)

	// 2. Se não tem MongoDB (modo local), só loga e sai
	if playersCollection == nil || scoresCollection == nil || !isDocker {
		log.Printf("[LOCAL] Score registrado: %s — %d pontos (melhor: %d)",
			player.Username, points, player.BestScore)
		return
	}

	// 3. Salva no ranking geral
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	scoreDoc := bson.M{
		"player_id": player.ID,       // importante para consultas futuras
		"nome":      player.Username, // continua mostrando no ranking
		"pontos":    points,
		"data":      time.Now(),
	}

	_, err := scoresCollection.InsertOne(ctx, scoreDoc)
	if err != nil {
		log.Printf("Erro ao salvar score no ranking: %v", err)
	} else {
		log.Printf("Score salvo no ranking: %s — %d pontos (melhor: %d)",
			player.Username, points, player.BestScore)
	}
}

func GetTop10() []Score {
	// modo local ou MongoDB indisponível → retorna mock
	if scoresCollection == nil || !isDocker {
		return []Score{
			{Nome: "JOGADOR01", Pontos: 250, Data: time.Now().Add(-time.Hour)},
			{Nome: "JOGADOR02", Pontos: 180, Data: time.Now().Add(-2 * time.Hour)},
			{Nome: generateUserID(), Pontos: 120, Data: time.Now()},
		}
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	cursor, err := scoresCollection.Find(ctx,
		bson.M{},
		options.Find().SetSort(bson.D{{Key: "pontos", Value: -1}}).SetLimit(10),
	)
	if err != nil {
		log.Printf("Erro ao buscar ranking: %v", err)
		return nil
	}
	defer cursor.Close(ctx)

	var results []Score
	if err := cursor.All(ctx, &results); err != nil {
		log.Printf("Erro ao decodificar scores: %v", err)
		return nil
	}

	return results
}
