package game

import "time"

// representa uma coordenada na arena
type Coord struct {
	X, Y int
}

// de comida
const (
	FOOD_NORMAL = iota
	FOOD_BONUS
	FOOD_PENALTY
)

// power-ups
const (
	POWERUP_SHIELD = iota
	POWERUP_GHOST
	POWERUP_MAGNET
)

// obstaculos
const (
	OBSTACLE_WALL = iota
	OBSTACLE_MOVING
)

// comida na arena
type Food struct {
	Coord
	Points    int
	FoodType  int
	SpawnTime time.Time
	Lifetime  time.Duration
}

// power-up na arena
type PowerUp struct {
	Coord
	PowerType int
	SpawnTime time.Time
	Lifetime  time.Duration
}

// obstaculo na arena
type Obstacle struct {
	Coord
	ObstacleType int
	IsTemporary  bool
	SpawnTime    time.Time
	Lifetime     time.Duration
}

// estrangeiro inimigo
type Boss struct {
	Body     []Coord
	Dir      Coord
	Speed    time.Duration
	LastMove time.Time
	Points   int
	IsAlive  bool
	Health   int // verificar la no boos.go
}

// controla o sistema de combos
type ComboSystem struct {
	CurrentCombo int
	LastFoodTime time.Time
	ComboTimeout time.Duration
	MaxCombo     int
}

// mensagens na tela
type GameMessage struct {
	Text      string
	CreatedAt time.Time
	Duration  time.Duration
}
