package game

import (
	"math"
	"math/rand"
	"time"
)

func newBoss(arenaWidth, arenaHeight int, playerSnake *Snake) *Boss {
	side := rand.Intn(4)
	var head Coord

	switch side {
	case 0: // esquerda
		head = Coord{X: 3, Y: rand.Intn(arenaHeight-8) + 5}
	case 1: // direita
		head = Coord{X: arenaWidth - 4, Y: rand.Intn(arenaHeight-8) + 5}
	case 2: // cima
		head = Coord{X: rand.Intn(arenaWidth-8) + 5, Y: 4}
	default: // baixo
		head = Coord{X: rand.Intn(arenaWidth-8) + 5, Y: arenaHeight - 5}
	}

	// prevencao para nao nascer em cima do jogador
	for i := 0; i < 30; i++ {
		if !playerSnake.IsOnPosition(head) && !playerSnake.IsOnPosition(Coord{head.X + 1, head.Y}) {
			break
		}
		// tenta outra posição na mesma borda
		switch side {
		case 0:
			head.Y = rand.Intn(arenaHeight-8) + 5
		case 1:
			head.Y = rand.Intn(arenaHeight-8) + 5
		case 2:
			head.X = rand.Intn(arenaWidth-8) + 5
		case 3:
			head.X = rand.Intn(arenaWidth-8) + 5
		}
	}

	// sempre para DENTRO da arena
	var dir Coord
	switch side {
	case 0:
		dir = Coord{1, 0} // da esquerda → direita
	case 1:
		dir = Coord{-1, 0} // da direita → esquerda
	case 2:
		dir = Coord{0, 1} // de cima → baixo
	case 3:
		dir = Coord{0, -1} // de baixo → cima
	}

	// body inicial
	body := []Coord{head}
	for i := 1; i < 9; i++ {
		body = append(body, Coord{
			X: head.X - dir.X*i,
			Y: head.Y - dir.Y*i,
		})
	}

	return &Boss{
		Body:     body,
		Dir:      dir,
		Speed:    160 * time.Millisecond,
		LastMove: time.Now(),
		Points:   250,
		IsAlive:  true,
		Health:   1, // vida inicial
	}
}

// IA do estrangeiro
func (b *Boss) calculateDirection(playerHead Coord, foods []*Food, arenaWidth, arenaHeight int) Coord {
	head := b.Body[0]

	// TODO: 1. PRIORIDADE MAXIMA: ir atras da fruta mais proxima
	var closestFood *Food
	bestDist := 999.0
	for _, f := range foods {
		dx := float64(f.X - head.X)
		dy := float64(f.Y - head.Y)
		dist := math.Sqrt(dx*dx + dy*dy)
		if dist < bestDist {
			bestDist = dist
			closestFood = f
		}
	}

	// vai atras da fruta se estiver a ate 20 blocos de distancia
	if closestFood != nil && bestDist < 20 {
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

	// 2. so persegue jogador se estiver MUITO PERTO, menos de 6 blocos
	dxP := playerHead.X - head.X
	dyP := playerHead.Y - head.Y
	distToPlayer := math.Abs(float64(dxP)) + math.Abs(float64(dyP))
	if distToPlayer < 6 {
		if math.Abs(float64(dxP)) > math.Abs(float64(dyP)) {
			if dxP > 0 {
				return Coord{1, 0}
			}
			return Coord{-1, 0}
		} else {
			if dyP > 0 {
				return Coord{0, 1}
			}
			return Coord{0, -1}
		}
	}

	// 3. random moviment se estiver longe
	directions := []Coord{{1, 0}, {-1, 0}, {0, 1}, {0, -1}}
	rand.Shuffle(len(directions), func(i, j int) {
		directions[i], directions[j] = directions[j], directions[i]
	})

	for _, d := range directions {
		if d == (Coord{-b.Dir.X, -b.Dir.Y}) {
			continue
		}
		nx := head.X + d.X
		ny := head.Y + d.Y
		if nx >= 3 && nx <= arenaWidth-4 && ny >= 3 && ny <= arenaHeight-4 {
			return d
		}
	}

	return b.Dir // fica parado se encurralado (raro)
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
