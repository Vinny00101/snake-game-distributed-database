// falta configurar o banco de dados distribuído MongoDB Replica Set completamente
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

var scoresCollection *mongo.Collection
var isDocker = false

func initDB() {
	if os.Getenv("MONGO_URI") != "" || os.Getenv("DOCKER_ENV") != "" {
		isDocker = true
		uri := os.Getenv("MONGO_URI")
		if uri == "" {
			uri = "mongodb://mongo1:27017,mongo2:27017,mongo3:27017/trabalho?replicaSet=rs0"
		}
		connectWithRetry(uri)
	} else {
		log.Println("Modo local detectado — scores serão salvos apenas na sessão (sem MongoDB)")
	}
}

func connectWithRetry(uri string) {
	var client *mongo.Client
	var err error

	for i := 0; i < 30; i++ {
		client, err = mongo.Connect(context.TODO(), options.Client().ApplyURI(uri))
		if err == nil {
			if err = client.Ping(context.TODO(), nil); err == nil {
				scoresCollection = client.Database("trabalho").Collection("snake_scores")
				log.Println("MongoDB Replica Set conectado com sucesso!")
				return
			}
		}
		log.Printf("Aguardando MongoDB... (tentativa %d) - erro: %v", i+1, err)
		time.Sleep(2 * time.Second)
	}
	log.Println("MongoDB não disponível — scores não serão salvos")
}

func SaveScore(name string, points int) {
	if scoresCollection == nil || !isDocker {
		log.Printf("Score local: %s — %d pontos", name, points)
		return
	}

	_, err := scoresCollection.InsertOne(context.TODO(), bson.M{
		"nome":   name,
		"pontos": points,
		"data":   time.Now(),
	})

	if err != nil {
		log.Println("Erro ao salvar no MongoDB:", err)
	} else {
		log.Printf("Score salvo no MongoDB distribuído: %s — %d", name, points)
	}
}

func GetTop10() []Score {
	if scoresCollection == nil || !isDocker {
		// fallback para dados mock para teste local
		return []Score{
			{Nome: "JOGADOR01", Pontos: 150, Data: time.Now()},
			{Nome: "JOGADOR02", Pontos: 100, Data: time.Now()},
			{Nome: generateUserID(), Pontos: 80, Data: time.Now()},
		}
	}

	cursor, err := scoresCollection.Find(context.TODO(),
		bson.M{},
		options.Find().SetSort(bson.D{{"pontos", -1}}).SetLimit(10))

	if err != nil {
		log.Println("Erro ao buscar scores:", err)
		return nil
	}
	defer cursor.Close(context.TODO())

	var results []Score
	if err := cursor.All(context.TODO(), &results); err != nil {
		log.Println("Erro ao decodificar scores:", err)
		return nil
	}

	return results
}
