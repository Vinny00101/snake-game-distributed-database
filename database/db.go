package database

import (
	"context"
	"fmt"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// Variáveis exportadas (letra maiúscula) para serem usadas pelo pacote game
var (
	Client      *mongo.Client
	DBConnected bool = false
)

// Endereços corrigidos para o Replica Set (SUBSTITUA se os IPs mudaram novamente!)
const ConnectionString = "mongodb://172.19.71.63:27017,172.19.66.213:27017,172.19.68.155:27017/?replicaSet=rs0"

// ConnectDB tenta se conectar ao MongoDB e retorna o cliente.
func ConnectDB() (*mongo.Client, error) {
	// Contexto Principal para a criação do cliente (30s)
	ctxConnect, cancelConnect := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancelConnect()

	client, err := mongo.Connect(ctxConnect, options.Client().ApplyURI(ConnectionString))
	if err != nil {
		return nil, fmt.Errorf("falha ao criar cliente MongoDB: %w", err)
	}

	// Loop de Ping com Retry
	for i := 0; i < 30; i++ {
		// Contexto de curta duração APENAS para o PING (5s por tentativa)
		ctxPing, cancelPing := context.WithTimeout(context.Background(), 5*time.Second)

		if err = client.Ping(ctxPing, nil); err == nil {
			cancelPing()    // Ping BEM-SUCEDIDO
			Client = client // Salva o cliente na variável global exportada
			DBConnected = true
			log.Println("✅ MongoDB conectado e Replica Set ativo!")
			return client, nil
		}
		cancelPing() // Limpa o contexto se o ping falhar

		log.Printf("Aguardando MongoDB... (tentativa %d/30) - erro: %v", i+1, err)
		time.Sleep(2 * time.Second)
	}

	log.Println("❌ MongoDB não disponível após 30 tentativas.")
	if err := client.Disconnect(context.Background()); err != nil {
		log.Printf("Aviso: Erro ao desconectar cliente após falha no retry: %v", err)
	}
	return nil, fmt.Errorf("MongoDB indisponível")
}
