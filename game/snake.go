package game

type Snake struct {
	Body []Coord
	Dir  Coord
}

func newSnake() *Snake {
	return &Snake{
		Body: []Coord{
			{X: 30, Y: 12},
			{X: 29, Y: 12},
			{X: 28, Y: 12},
		},
		Dir: Coord{X: 1, Y: 0},
	}
}

func (s *Snake) Head() Coord { return s.Body[0] }

func (s *Snake) Move() {
	head := s.Head()
	newHead := Coord{X: head.X + s.Dir.X, Y: head.Y + s.Dir.Y}
	s.Body = append([]Coord{newHead}, s.Body...)
	s.Body = s.Body[:len(s.Body)-1]
}

func (s *Snake) Grow() {
	tail := s.Body[len(s.Body)-1]
	s.Body = append(s.Body, tail)
}

func (s *Snake) ChangeDir(dx, dy int) {
	newDir := Coord{X: dx, Y: dy}

	// nao permite movimento oposto ao atual
	if len(s.Body) > 1 {
		currentDir := Coord{
			X: s.Body[0].X - s.Body[1].X,
			Y: s.Body[0].Y - s.Body[1].Y,
		}

		// se a nova direção for oposta a atual, ele ignora e retorna
		if newDir.X == -currentDir.X && newDir.Y == -currentDir.Y {
			return
		}
	}

	s.Dir = newDir
}

func (s *Snake) IsOnPosition(c Coord) bool {
	for _, seg := range s.Body {
		if seg.X == c.X && seg.Y == c.Y {
			return true
		}
	}
	return false
}

func (s *Snake) SelfCollision() bool {
	head := s.Head()
	for i, seg := range s.Body {
		if i > 0 && seg.X == head.X && seg.Y == head.Y {
			return true
		}
	}
	return false
}
