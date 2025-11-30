// db.go
package game

import (
	"context"
	"log"
	"os"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Score struct {
	Nome   string    `bson:"nome"`
	Pontos int       `bson:"pontos"`
	Data   time.Time `bson:"data"`
}

var (
	scoresCollection *mongo.Collection
	isDocker         bool
	client           *mongo.Client // mantido para desconexão futura se precisar
)

func initDB() {
	if os.Getenv("MONGO_URI") != "" || os.Getenv("DOCKER_ENV") != "" {
		isDocker = true
		uri := os.Getenv("MONGO_URI")
		if uri == "" {
			uri = "mongodb://mongo1:27017,mongo2:27017,mongo3:27017/trabalho?replicaSet=rs0&connect=direct"
		}
		connectWithRetry(uri)
	} else {
		log.Println("Modo local detectado — scores serão salvos apenas na sessão (sem MongoDB)")
	}
}

func connectWithRetry(uri string) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	var err error
	client, err = mongo.Connect(ctx, options.Client().ApplyURI(uri))
	if err != nil {
		log.Fatalf("Falha ao criar cliente MongoDB: %v", err)
	}

	for i := 0; i < 30; i++ {
		if err = client.Ping(ctx, nil); err == nil {
			scoresCollection = client.Database("trabalho").Collection("snake_scores")
			log.Println("MongoDB Replica Set conectado com sucesso!")
			return
		}
		log.Printf("Aguardando MongoDB... (tentativa %d/30) - erro: %v", i+1, err)
		time.Sleep(2 * time.Second)
	}

	log.Println("MongoDB não disponível após 30 tentativas — scores não serão persistidos")
	scoresCollection = nil
}

func SaveScore(name string, points int) {
	if scoresCollection == nil || !isDocker {
		log.Printf("Score local: %s — %d pontos", name, points)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := scoresCollection.InsertOne(ctx, bson.M{
		"nome":   name,
		"pontos": points,
		"data":   time.Now(),
	})

	if err != nil {
		log.Printf("Erro ao salvar score no MongoDB: %v", err)
	} else {
		log.Printf("Score salvo com sucesso: %s — %d pontos", name, points)
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
