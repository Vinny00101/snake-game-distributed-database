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

// Food representa a comida na arena
type Food struct {
	Coord
	Points    int
	FoodType  int
	SpawnTime time.Time
	Lifetime  time.Duration
}
