package game

import "time"

// Coord representa uma coordenada na arena
type Coord struct {
	X, Y int
}

// Tipos de comida
const (
	FOOD_NORMAL = iota
	FOOD_BONUS
	FOOD_PENALTY
)

// Tipos de power-ups
const (
	POWERUP_SHIELD = iota
	POWERUP_GHOST
	POWERUP_MAGNET
)

// Tipos de obstáculos
const (
	OBSTACLE_WALL = iota
	OBSTACLE_MOVING
)

// Food representa a comida na arena
type Food struct {
	Coord
	Points    int
	FoodType  int
	SpawnTime time.Time
	Lifetime  time.Duration
}

// PowerUp representa um power-up na arena
type PowerUp struct {
	Coord
	PowerType int
	SpawnTime time.Time
	Lifetime  time.Duration
}

// Obstacle representa um obstáculo na arena
type Obstacle struct {
	Coord
	ObstacleType int
	IsTemporary  bool
	SpawnTime    time.Time
	Lifetime     time.Duration
}

// Boss representa uma cobra inimiga
type Boss struct {
	Body     []Coord
	Dir      Coord
	Speed    time.Duration
	LastMove time.Time
	Points   int
	IsAlive  bool
	Health   int // agora é 3 hits para matar
}

// ComboSystem controla o sistema de combos
type ComboSystem struct {
	CurrentCombo int
	LastFoodTime time.Time
	ComboTimeout time.Duration
	MaxCombo     int
}

// GameMessage representa mensagens na tela
type GameMessage struct {
	Text      string
	CreatedAt time.Time
	Duration  time.Duration
}
