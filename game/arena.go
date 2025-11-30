package game

import (
	"math/rand"
	"time"
)

type Arena struct {
	X, Y          int
	Width, Height int
	Snake         *Snake
	Foods         []*Food
	Points        int
	lastFoodTime  time.Time
	foodCooldown  time.Duration
	maxFoods      int
}

func newArena(width, height int) *Arena {
	rand.Seed(time.Now().UnixNano())
	a := &Arena{
		X:            2,
		Y:            3,
		Width:        width,
		Height:       height,
		Snake:        newSnake(),
		Foods:        make([]*Food, 0),
		foodCooldown: 2 * time.Second,
		maxFoods:     3, // max de foods
	}
	a.placeFood()
	return a
}

func (a *Arena) placeFood() {
	if time.Since(a.lastFoodTime) < a.foodCooldown || len(a.Foods) >= a.maxFoods {
		return
	}

	for attempts := 0; attempts < 50; attempts++ {
		// defini uma area segura para comida
		minX := a.X + 2
		maxX := a.X + a.Width - 3
		minY := a.Y + 2
		maxY := a.Y + a.Height - 3

		if maxX <= minX || maxY <= minY {
			minX = a.X + 1
			maxX = a.X + a.Width - 2
			minY = a.Y + 1
			maxY = a.Y + a.Height - 2
		}

		x := rand.Intn(maxX-minX) + minX
		y := rand.Intn(maxY-minY) + minY
		c := Coord{X: x, Y: y}

		// if para ver se não colide com a cobra nem com outras comidas
		if !a.Snake.IsOnPosition(c) && !a.isFoodOnPosition(c) {
			// probabilidades: 50% normal, 30% bônus, 20% penalidade
			randType := rand.Float32()
			var foodType int
			var points int
			var lifetime time.Duration

			if randType < 0.50 {
				foodType = FOOD_NORMAL
				points = 10
				lifetime = 8 * time.Second // 8 segundos
			} else if randType < 0.80 {
				foodType = FOOD_BONUS
				points = 25
				lifetime = 6 * time.Second // 6 segundos (mais rara)
			} else {
				foodType = FOOD_PENALTY
				points = -20
				lifetime = 10 * time.Second // 10 segundos (dá tempo de evitar)
			}

			newFood := &Food{
				Coord:     Coord{X: x, Y: y},
				Points:    points,
				FoodType:  foodType,
				SpawnTime: time.Now(),
				Lifetime:  lifetime,
			}

			a.Foods = append(a.Foods, newFood)
			a.lastFoodTime = time.Now()

			a.foodCooldown = time.Duration(3+rand.Intn(3)) * time.Second
			return
		}
	}
}

func (a *Arena) isFoodOnPosition(c Coord) bool {
	for _, food := range a.Foods {
		if food.X == c.X && food.Y == c.Y {
			return true
		}
	}
	return false
}

func (a *Arena) removeExpiredFoods() {
	now := time.Now()
	validFoods := make([]*Food, 0)

	for _, food := range a.Foods {
		// mantem a comida se ainda não expirou
		if now.Sub(food.SpawnTime) < food.Lifetime {
			validFoods = append(validFoods, food)
		}
	}

	a.Foods = validFoods
}

func (a *Arena) Tick(game *Game) bool {
	a.Snake.Move()

	head := a.Snake.Head()

	// colisao com parede
	if head.X <= a.X || head.X >= a.X+a.Width-1 ||
		head.Y <= a.Y || head.Y >= a.Y+a.Height-1 {
		return false
	}

	// colisao com corpo
	if a.Snake.SelfCollision() {
		return false
	}

	// verifica colisão com comidas
	eatenFoods := make([]*Food, 0)
	remainingFoods := make([]*Food, 0)

	for _, food := range a.Foods {
		if head.X == food.X && head.Y == food.Y {
			// comeu esta comida?
			a.Points += food.Points

			switch food.FoodType {
			case FOOD_BONUS:
				// if de só gera bônus se o jogador não tiver um ativo
				if !game.bonusActive {
					bonusTypes := []string{"VELOCIDADE", "CRESCIMENTO", "PONTOS"}
					bonusType := bonusTypes[rand.Intn(len(bonusTypes))]
					game.activateBonus(bonusType)
				}
				a.Snake.Grow()
			case FOOD_PENALTY:
				// penalidade: diminui tamanho (mínimo 3 segmentos)
				if len(a.Snake.Body) > 3 {
					a.Snake.Body = a.Snake.Body[:len(a.Snake.Body)-2] // remove 2 segmentos
					if len(a.Snake.Body) < 3 {
						a.Snake.Body = a.Snake.Body[:3] // garante mínimo de 3
					}
				}
			default: // FOOD_NORMAL
				a.Snake.Grow()
			}

			eatenFoods = append(eatenFoods, food)
		} else {
			// mantem comida não comida
			remainingFoods = append(remainingFoods, food)
		}
	}

	// atualizar lista de comidas
	a.Foods = remainingFoods

	// remover comidas expiradas
	a.removeExpiredFoods()

	// tentar colocar nova comida se alguma foi comida
	if len(eatenFoods) > 0 {
		a.placeFood()
	} else {
		// tentar colocar comida mesmo se não comeu (para manter o fluxo)
		a.placeFood()
	}

	return true
}
