package main

// essa API foi feita para testa o banco de dados distribuidor

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var collection *mongo.Collection

func main() {
	// Função que tenta conectar até dar certo (retry infinito com backoff)
	client := connectWithRetry()
	defer client.Disconnect(context.Background())

	collection = client.Database("trabalho").Collection("produtos")

	r := gin.Default()

	// Rota para adicionar produto
	r.POST("/produtos", func(c *gin.Context) {
		var p struct {
			Nome  string  `json:"nome" binding:"required"`
			Preco float64 `json:"preco,omitempty"`
		}

		if err := c.ShouldBindJSON(&p); err != nil {
			c.JSON(400, gin.H{"erro": "JSON inválido: " + err.Error()})
			return
		}

		res, err := collection.InsertOne(context.Background(), p)
		if err != nil {
			c.JSON(500, gin.H{"erro": "falha ao inserir: " + err.Error()})
			return
		}

		c.JSON(201, gin.H{
			"mensagem": "produto inserido com sucesso",
			"id":       res.InsertedID,
		})
	})

	// Rota para listar todos
	r.GET("/produtos", func(c *gin.Context) {
		cursor, err := collection.Find(context.Background(), bson.M{})
		if err != nil {
			c.JSON(500, gin.H{"erro": err.Error()})
			return
		}
		defer cursor.Close(context.Background())

		var resultados []bson.M
		if err = cursor.All(context.Background(), &resultados); err != nil {
			c.JSON(500, gin.H{"erro": err.Error()})
			return
		}

		c.JSON(200, resultados)
	})

	// Health check
	r.GET("/health", func(c *gin.Context) {
		c.String(200, "API + MongoDB Replica Set rodando perfeitamente!")
	})

	fmt.Println("API Go rodando na porta 8080")
	r.Run(":8080")
}

// Conecta com retry até o Replica Set ficar saudável
func connectWithRetry() *mongo.Client {
	uri := "mongodb://172.24.95.101:27017/?replicaSet=rs0"

	for {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		client, err := mongo.Connect(ctx, options.Client().ApplyURI(uri))
		if err != nil {
			log.Println("MongoDB ainda não está pronto... tentando novamente em 3s")
			cancel()
			time.Sleep(3 * time.Second)
			continue
		}

		// Testa ping real
		err = client.Ping(ctx, nil)
		cancel()
		if err == nil {
			log.Println("Conectado ao MongoDB Replica Set com sucesso!")
			return client
		}

		log.Println("Ping falhou, tentando novamente em 3s...", err)
		time.Sleep(3 * time.Second)
	}
}
