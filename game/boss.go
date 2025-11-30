package game

import (
	"math"
	"math/rand"
	"time"
)

func newBoss(arenaWidth, arenaHeight int, playerSnake *Snake) *Boss {
	var start Coord
	playerHead := playerSnake.Head()

	// spawn no lado oposto ao jogador
	if playerHead.X < arenaWidth/2 {
		start.X = arenaWidth - 5
	} else {
		start.X = 4
	}
	if playerHead.Y < arenaHeight/2 {
		start.Y = arenaHeight - 5
	} else {
		start.Y = 4
	}

	// garante posição livre
	for i := 0; i < 20; i++ {
		candidate := Coord{
			X: rand.Intn(arenaWidth-8) + 4,
			Y: rand.Intn(arenaHeight-8) + 4,
		}
		if !playerSnake.IsOnPosition(candidate) {
			start = candidate
			break
		}
	}

	// direção inicial aleatória segura
	dir := Coord{X: 1, Y: 0}
	directions := []Coord{{1, 0}, {-1, 0}, {0, 1}, {0, -1}}
	rand.Shuffle(len(directions), func(i, j int) { directions[i], directions[j] = directions[j], directions[i] })

	body := []Coord{start}
	for i := 1; i < 8; i++ {
		body = append(body, Coord{X: start.X - dir.X*i, Y: start.Y - dir.Y*i})
	}

	return &Boss{
		Body:     body,
		Dir:      dir,
		Speed:    160 * time.Millisecond,
		LastMove: time.Now(),
		Points:   200,
		IsAlive:  true,
		Health:   3, // 3 hits para matar
	}
}

// Prioridade do boss: 1º fruta → 2º evitar paredes → 3º perseguir jogador só se muito perto
func (b *Boss) calculateDirection(playerHead Coord, foods []*Food, arenaWidth, arenaHeight int) Coord {
	head := b.Body[0]

	// 1. Procurar fruta mais próxima
	var closestFood *Food
	minDist := math.MaxFloat64
	for _, f := range foods {
		dx := float64(f.X - head.X)
		dy := float64(f.Y - head.Y)
		dist := dx*dx + dy*dy
		if dist < minDist {
			minDist = dist
			closestFood = f
		}
	}

	// Se tem fruta perto (distância < 15), vai nela
	if closestFood != nil && minDist < 225 { // 15²
		dx := closestFood.X - head.X
		dy := closestFood.Y - head.Y

		if math.Abs(float64(dx)) > math.Abs(float64(dy)) {
			if dx > 0 {
				return Coord{1, 0}
			}
			return Coord{-1, 0}
		} else {
			if dy > 0 {
				return Coord{0, 1}
			}
			return Coord{0, -1}
		}
	}

	// 2. Se jogador estiver MUITO perto (< 8 blocos), persegue agressivamente
	dxPlayer := playerHead.X - head.X
	dyPlayer := playerHead.Y - head.Y
	if math.Abs(float64(dxPlayer))+math.Abs(float64(dyPlayer)) < 8 {
		if math.Abs(float64(dxPlayer)) > math.Abs(float64(dyPlayer)) {
			if dxPlayer > 0 {
				return Coord{1, 0}
			}
			return Coord{-1, 0}
		} else {
			if dyPlayer > 0 {
				return Coord{0, 1}
			}
			return Coord{0, -1}
		}
	}

	// 3. caso contrário: movimento que evita paredes
	directions := []Coord{{1, 0}, {-1, 0}, {0, 1}, {0, -1}}
	rand.Shuffle(len(directions), func(i, j int) { directions[i], directions[j] = directions[j], directions[i] })

	for _, d := range directions {
		if d == (Coord{-b.Dir.X, -b.Dir.Y}) { // nao volta pra tras
			continue
		}
		nx := head.X + d.X
		ny := head.Y + d.Y
		if nx > 2 && nx < arenaWidth-3 && ny > 2 && ny < arenaHeight-3 {
			return d
		}
	}

	// ultimo recurso: direção atual
	return b.Dir
}

func (b *Boss) Move(playerHead Coord, foods []*Food, arenaWidth, arenaHeight int) {
	if time.Since(b.LastMove) < b.Speed || !b.IsAlive {
		return
	}

	b.Dir = b.calculateDirection(playerHead, foods, arenaWidth, arenaHeight)
	newHead := Coord{X: b.Head().X + b.Dir.X, Y: b.Head().Y + b.Dir.Y}

	// security contra parede
	if newHead.X <= 2 || newHead.X >= arenaWidth-3 || newHead.Y <= 2 || newHead.Y >= arenaHeight-3 {
		// tenta outra direcao
		b.Dir = b.calculateDirection(playerHead, foods, arenaWidth, arenaHeight)
		newHead = Coord{X: b.Head().X + b.Dir.X, Y: b.Head().Y + b.Dir.Y}
	}

	b.Body = append([]Coord{newHead}, b.Body...)
	b.Body = b.Body[:len(b.Body)-1]
	b.LastMove = time.Now()
}

func (b *Boss) Head() Coord {
	return b.Body[0]
}

func (b *Boss) Grow() {
	tail := b.Body[len(b.Body)-1]
	b.Body = append(b.Body, tail)
	b.Points += 20 // cresce = mais pontos ao morrer
}

func (b *Boss) TakeDamage() (died bool) {
	b.Health--
	if b.Health <= 0 {
		b.IsAlive = false
		return true
	}
	return false
}

func (b *Boss) IsOnPosition(c Coord) bool {
	for _, seg := range b.Body {
		if seg.X == c.X && seg.Y == c.Y {
			return true
		}
	}
	return false
}
