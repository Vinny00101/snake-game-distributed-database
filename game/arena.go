package game

import (
	"math/rand"
	"time"
)

type Arena struct {
	X, Y            int
	Width, Height   int
	Snake           *Snake
	Foods           []*Food
	PowerUps        []*PowerUp
	Obstacles       []*Obstacle
	Bosses          []*Boss
	Points          int
	Level           int
	ComboSystem     *ComboSystem
	Messages        []GameMessage
	lastFoodTime    time.Time
	foodCooldown    time.Duration
	maxFoods        int
	speedMultiplier float64
	lastBossSpawn   time.Time
	bossCooldown    time.Duration
}

func newArena(width, height int) *Arena {
	rand.Seed(time.Now().UnixNano())
	a := &Arena{
		X:         2,
		Y:         3,
		Width:     width,
		Height:    height,
		Snake:     newSnake(),
		Foods:     make([]*Food, 0),
		PowerUps:  make([]*PowerUp, 0),
		Obstacles: make([]*Obstacle, 0),
		Bosses:    make([]*Boss, 0),
		Messages:  make([]GameMessage, 0),
		ComboSystem: &ComboSystem{
			ComboTimeout: 3 * time.Second,
			MaxCombo:     0,
		},
		foodCooldown:    2 * time.Second,
		maxFoods:        3,
		speedMultiplier: 1.0,
		Level:           1,
		bossCooldown:    30 * time.Second, // Boss a cada 30 segundos
	}
	a.placeFood()
	return a
}

func (a *Arena) AddMessage(text string, duration time.Duration) {
	a.Messages = append(a.Messages, GameMessage{
		Text:      text,
		CreatedAt: time.Now(),
		Duration:  duration,
	})
}

func (a *Arena) RemoveExpiredMessages() {
	now := time.Now()
	validMessages := make([]GameMessage, 0)
	for _, msg := range a.Messages {
		if now.Sub(msg.CreatedAt) < msg.Duration {
			validMessages = append(validMessages, msg)
		}
	}
	a.Messages = validMessages
}

func (a *Arena) increaseDifficulty() {
	a.Level++
	a.speedMultiplier = 1.0 + (float64(a.Level) * 0.1)
	a.maxFoods = 3 + a.Level/3

	// agora vai adicionar obstaculos baseado no nivel
	if a.Level%2 == 0 && len(a.Obstacles) < 5+a.Level/2 {
		a.placeObstacle()
	}
}

func (a *Arena) placeFood() {
	if time.Since(a.lastFoodTime) < a.foodCooldown || len(a.Foods) >= a.maxFoods {
		return
	}

	for attempts := 0; attempts < 50; attempts++ {
		x := rand.Intn(a.Width-4) + a.X + 2
		y := rand.Intn(a.Height-4) + a.Y + 2
		c := Coord{X: x, Y: y}

		if a.isPositionValid(c) {
			randType := rand.Float32()
			var foodType int
			var points int
			var lifetime time.Duration

			if randType < 0.50 {
				foodType = FOOD_NORMAL
				points = 10
				lifetime = 8 * time.Second
			} else if randType < 0.80 {
				foodType = FOOD_BONUS
				points = 25
				lifetime = 6 * time.Second
			} else {
				foodType = FOOD_PENALTY
				points = -20
				lifetime = 10 * time.Second
			}

			newFood := &Food{
				Coord:     c,
				Points:    points,
				FoodType:  foodType,
				SpawnTime: time.Now(),
				Lifetime:  lifetime,
			}

			a.Foods = append(a.Foods, newFood)
			a.lastFoodTime = time.Now()
			a.foodCooldown = time.Duration(2+rand.Intn(3)) * time.Second
			return
		}
	}
}

func (a *Arena) placeObstacle() {
	for attempts := 0; attempts < 30; attempts++ {
		x := rand.Intn(a.Width-4) + a.X + 2
		y := rand.Intn(a.Height-4) + a.Y + 2
		c := Coord{X: x, Y: y}

		if a.isPositionValid(c) {
			obstacle := &Obstacle{
				Coord:        c,
				ObstacleType: OBSTACLE_WALL,
				IsTemporary:  rand.Float32() < 0.3,
				SpawnTime:    time.Now(),
				Lifetime:     time.Duration(10+rand.Intn(20)) * time.Second,
			}
			a.Obstacles = append(a.Obstacles, obstacle)
			return
		}
	}
}

func (a *Arena) trySpawnBoss() {
	// defini para o nível 2
	if a.Level >= 2 && len(a.Bosses) == 0 && time.Since(a.lastBossSpawn) > a.bossCooldown {
		boss := newBoss(a.Width, a.Height, a.Snake)
		a.Bosses = append(a.Bosses, boss)
		a.lastBossSpawn = time.Now()
		a.AddMessage("Um estrangeiro invadiu seu mundo!", 4*time.Second)
	}
}

func (a *Arena) isPositionValid(c Coord) bool {
	// check cobra
	if a.Snake.IsOnPosition(c) {
		return false
	}
	// check comidas
	for _, food := range a.Foods {
		if food.X == c.X && food.Y == c.Y {
			return false
		}
	}
	// check obstaculos
	for _, obs := range a.Obstacles {
		if obs.X == c.X && obs.Y == c.Y {
			return false
		}
	}
	// check bosses
	for _, boss := range a.Bosses {
		if boss.IsOnPosition(c) {
			return false
		}
	}
	// check powerups
	for _, powerup := range a.PowerUps {
		if powerup.X == c.X && powerup.Y == c.Y {
			return false
		}
	}
	return true
}

func (a *Arena) removeExpiredItems() {
	now := time.Now()

	// remove comidas expiradas
	validFoods := make([]*Food, 0)
	for _, food := range a.Foods {
		if now.Sub(food.SpawnTime) < food.Lifetime {
			validFoods = append(validFoods, food)
		}
	}
	a.Foods = validFoods

	// remove obstaculos expirados
	validObstacles := make([]*Obstacle, 0)
	for _, obs := range a.Obstacles {
		if !obs.IsTemporary || now.Sub(obs.SpawnTime) < obs.Lifetime {
			validObstacles = append(validObstacles, obs)
		}
	}
	a.Obstacles = validObstacles
}

func (a *Arena) updateCombo() {
	now := time.Now()
	if now.Sub(a.ComboSystem.LastFoodTime) > a.ComboSystem.ComboTimeout {
		a.ComboSystem.CurrentCombo = 0
	} else {
		a.ComboSystem.CurrentCombo++
		if a.ComboSystem.CurrentCombo > a.ComboSystem.MaxCombo {
			a.ComboSystem.MaxCombo = a.ComboSystem.CurrentCombo
		}
	}
	a.ComboSystem.LastFoodTime = now
}

func (a *Arena) Tick(game *Game) bool {
	a.Snake.Move()
	head := a.Snake.Head()

	// colisoes normais...
	if head.X <= a.X || head.X >= a.X+a.Width-1 ||
		head.Y <= a.Y || head.Y >= a.Y+a.Height-1 {
		return false
	}
	if a.Snake.SelfCollision() {
		return false
	}
	for _, obs := range a.Obstacles {
		if head.X == obs.X && head.Y == obs.Y {
			return false
		}
	}

	a.trySpawnBoss()

	for i := len(a.Bosses) - 1; i >= 0; i-- {
		boss := a.Bosses[i]
		if !boss.IsAlive {
			a.Bosses = append(a.Bosses[:i], a.Bosses[i+1:]...)
			continue
		}

		boss.Move(head, a.Foods, a.Width, a.Height)

		// TODO: boss é para comer fruta, talvez esteja bugado
		for j := len(a.Foods) - 1; j >= 0; j-- {
			food := a.Foods[j]
			if boss.Head().X == food.X && boss.Head().Y == food.Y {
				boss.Grow()
				a.Foods = append(a.Foods[:j], a.Foods[j+1:]...)
				a.AddMessage("O estrangeiro comeu uma fruta!", 2*time.Second)
				a.placeFood()
			}
		}

		// hitKill colide com qualquer parte do boss → morre
		if a.Snake.CollidesWith(&Snake{Body: boss.Body}) {
			return false
		}

		// jogador bate na CABECA do boss → dano
		if head.X == boss.Head().X && head.Y == boss.Head().Y {
			if boss.TakeDamage() {

				a.Points += boss.Points
				growAmount := len(boss.Body) - 5
				for k := 0; k < growAmount; k++ {
					a.Snake.Grow()
				}
				a.Bosses = append(a.Bosses[:i], a.Bosses[i+1:]...)
				a.AddMessage("ESTRANGEIRO DERROTADO! +"+string(rune(boss.Points))+" pontos e +"+string(rune(growAmount))+" tamanho!", 5*time.Second)
			} else {
				a.AddMessage("Estrangeiro ferido! ("+string(rune(boss.Health))+"/3)", 2*time.Second)
			}
		} else {

			a.Bosses = append(a.Bosses[:i], a.Bosses[i+1:]...)
			i--
		}
	}

	// verifica comidas
	eatenFoods := make([]*Food, 0)
	remainingFoods := make([]*Food, 0)

	for _, food := range a.Foods {
		if head.X == food.X && head.Y == food.Y {
			basePoints := food.Points

			a.updateCombo()
			comboMultiplier := 1 + (a.ComboSystem.CurrentCombo / 3)
			finalPoints := basePoints * comboMultiplier

			a.Points += finalPoints

			switch food.FoodType {
			case FOOD_BONUS:
				if !game.bonusActive {
					bonusTypes := []string{"VELOCIDADE", "CRESCIMENTO", "PONTOS"}
					bonusType := bonusTypes[rand.Intn(len(bonusTypes))]
					game.activateBonus(bonusType)
				}
				a.Snake.Grow()
			case FOOD_PENALTY:
				if len(a.Snake.Body) > 3 {
					a.Snake.Shrink()
				}
			default:
				a.Snake.Grow()
			}

			eatenFoods = append(eatenFoods, food)

			// aumentar dificuldade a cada 50 pontos
			if a.Points/50 > (a.Points-finalPoints)/50 {
				a.increaseDifficulty()
			}
		} else {
			remainingFoods = append(remainingFoods, food)
		}
	}
	a.Foods = remainingFoods

	a.removeExpiredItems()
	a.RemoveExpiredMessages()

	if len(eatenFoods) > 0 || len(a.Foods) < a.maxFoods/2 {
		a.placeFood()
	}

	return true
}
