package game

import (
	"context"
	"errors"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"golang.org/x/crypto/bcrypt"
)

type Player struct {
	ID           primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Username     string             `bson:"username" json:"username"`
	PasswordHash string             `bson:"password_hash" json:"-"`
	CreatedAt    time.Time          `bson:"created_at" json:"created_at"`
	BestScore    int                `bson:"best_score" json:"best_score"`
	LastLogin    time.Time          `bson:"last_login,omitempty" json:"last_login,omitempty"`
}

func RegisterPlayer(username string, password string) (*Player, error) {
	if playersCollection == nil {
		return nil, errors.New("banco de dados nao disponivel")
	}

	var existing Player
	err := playersCollection.FindOne(context.TODO(), bson.M{"username": username}).Decode(&existing)
	if err == nil {
		return nil, errors.New("nome de usuario ja existe")
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	newPlayer := &Player{
		Username:     username,
		PasswordHash: string(hash),
		CreatedAt:    time.Now(),
		BestScore:    0,
	}
	_, err = playersCollection.InsertOne(context.TODO(), newPlayer)
	if err != nil {
		return nil, err
	}
	return newPlayer, nil

}

// func para login do jogador
func AuthenticatePlayer(username string, password string) (*Player, error) {
	if playersCollection == nil {
		return nil, errors.New("banco de dados nao disponivel")
	}

	var player Player
	err := playersCollection.FindOne(context.TODO(), bson.M{"username": username}).Decode(&player)
	if err != nil {
		return nil, errors.New("nome de usuário ou senha inválidos")
	}

	err = bcrypt.CompareHashAndPassword([]byte(player.PasswordHash), []byte(password))
	if err != nil {
		return nil, errors.New("nome de usuário ou senha inválidos")
	}

	playersCollection.UpdateOne(context.TODO(),
		bson.M{"_id": player.ID},
		bson.M{"$set": bson.M{"last_login": time.Now()}},
	)

	return &player, nil
}

// atualiza o score do jogador
func (p *Player) UpdateBestScore(score int) error {
	if score <= p.BestScore {
		return nil // não melhorou
	}

	_, err := playersCollection.UpdateOne(context.TODO(),
		bson.M{"_id": p.ID},
		bson.M{"$set": bson.M{"best_score": score}},
	)
	if err == nil {
		p.BestScore = score
	}
	return err
}
