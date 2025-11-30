package game

import (
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/nsf/termbox-go"
)

type Game struct {
	arena         *Arena
	isRunning     bool
	score         int
	userID        string
	CurrentPlayer *Player
	bonusActive   bool
	bonusType     string
	bonusTimer    *time.Timer
	speed         time.Duration
	menuSnake     []Coord
	menuDir       Coord
	menuTicker    *time.Ticker
	menuMutex     sync.Mutex
	stopChan      chan bool
}

func NewGame() *Game {
	return &Game{
		arena:         newArena(60, 25),
		userID:        "visitante",
		speed:         120 * time.Millisecond,
		menuSnake:     []Coord{{X: 5, Y: 5}, {X: 4, Y: 5}, {X: 3, Y: 5}},
		menuDir:       Coord{X: 1, Y: 0},
		stopChan:      make(chan bool),
		CurrentPlayer: nil,
	}
}

func (g *Game) Start() {
	err := termbox.Init()
	if err != nil {
		panic(err)
	}
	defer termbox.Close()
	defer g.cleanup()

	termbox.SetInputMode(termbox.InputEsc)
	termbox.HideCursor()

	g.showAuthScreen()

	g.showMainMenu()
}

func (g *Game) cleanup() {
	if g.menuTicker != nil {
		g.menuTicker.Stop()
	}
	if g.bonusTimer != nil {
		g.bonusTimer.Stop()
	}
	close(g.stopChan)
}

func (g *Game) showAuthScreen() {
	options := []string{"Fazer Login", "Criar Conta", "Sair do Jogo"}
	selected := 0

	for {
		termbox.Clear(termbox.ColorDefault, termbox.ColorDefault)
		width, height := termbox.Size()

		statusText := "OFFLINE"
		statusColor := termbox.ColorRed | termbox.AttrBold

		if DBConnected {
			statusText = "ONLINE"
			statusColor = termbox.ColorGreen | termbox.AttrBold
		}

		drawText(g.arena.X+g.arena.Width-10, g.arena.Y-3, statusColor, termbox.ColorDefault, statusText)

		title := "SNAKE GO - UFPI 2025"
		subtitle := "Ally • Vini • Kleber"
		drawText((width-len(title))/2, height/2-10, termbox.ColorGreen|termbox.AttrBold, termbox.ColorDefault, title)
		drawText((width-len(subtitle))/2, height/2-8, termbox.ColorCyan, termbox.ColorDefault, subtitle)

		for i, option := range options {
			x := (width - 15) / 2
			y := height/2 - 1 + i*2

			fgColor := termbox.ColorWhite
			if i == selected {
				fgColor = termbox.ColorYellow | termbox.AttrBold
				drawText(x-3, y, fgColor, termbox.ColorDefault, ">")
			}

			drawText(x, y, fgColor, termbox.ColorDefault, option)
		}

		controls := "Use ↑↓ para navegar, ENTER para selecionar, ESC para sair"
		drawText((width-len(controls))/2, height-4, termbox.ColorDarkGray, termbox.ColorDefault, controls)

		termbox.Flush()

		ev := termbox.PollEvent()
		if ev.Type == termbox.EventKey {
			switch ev.Key {
			case termbox.KeyArrowUp:
				selected = (selected - 1 + len(options)) % len(options)
			case termbox.KeyArrowDown:
				selected = (selected + 1) % len(options)
			case termbox.KeyEnter:
				switch selected {
				case 0:
					if g.loginScreen() {
						return
					}
				case 1:
					g.registerScreen()

				case 2:
					termbox.Close()
					os.Exit(0)
				}
			case termbox.KeyEsc:
				termbox.Close()
				os.Exit(0)
			}
		}
	}
}

func (g *Game) inputString(prompt string, y int) string {
	buf := []rune{}
	width, _ := termbox.Size()
	x := (width - len(prompt) - 30) / 2

	for {
		termbox.Clear(termbox.ColorDefault, termbox.ColorDefault)
		drawText(x, y-2, termbox.ColorWhite, termbox.ColorDefault, prompt)
		drawText(x+len(prompt)+2, y-2, termbox.ColorCyan, termbox.ColorDefault, string(buf))
		drawText(x, y, termbox.ColorDarkGray, termbox.ColorDefault, "ENTER → confirmar | BACKSPACE → apagar | ESC → cancelar")
		termbox.Flush()

		ev := termbox.PollEvent()
		if ev.Type == termbox.EventKey {
			if ev.Key == termbox.KeyEnter && len(buf) > 0 {
				return string(buf)
			}
			if ev.Key == termbox.KeyEsc {
				return ""
			}
			if ev.Key == termbox.KeyBackspace || ev.Key == termbox.KeyBackspace2 {
				if len(buf) > 0 {
					buf = buf[:len(buf)-1]
				}
			} else if ev.Ch != 0 {
				buf = append(buf, ev.Ch)
			}
		}
	}
}

func (g *Game) loginScreen() bool {
	termbox.Clear(termbox.ColorDefault, termbox.ColorDefault)
	w, h := termbox.Size()
	drawText(w/2-12, h/2-4, termbox.ColorCyan|termbox.AttrBold, termbox.ColorDefault, "LOGIN")

	username := g.inputString("Usuario:", h/2)
	if username == "" {
		return false
	}

	password := g.inputString("Senha:", h/2+2)
	if password == "" {
		return false
	}

	player, err := AuthenticatePlayer(username, password)
	if err != nil {
		drawText(w/2-20, h/2+6, termbox.ColorRed, termbox.ColorDefault, "Erro: "+err.Error())
		drawText(w/2-15, h/2+8, termbox.ColorWhite, termbox.ColorDefault, "Pressione qualquer tecla...")
		termbox.Flush()
		termbox.PollEvent()
		return false
	}

	g.CurrentPlayer = player
	g.userID = player.Username

	drawText(w/2-20, h/2+6, termbox.ColorGreen|termbox.AttrBold, termbox.ColorDefault, "Login realizado com sucesso!")
	drawText(w/2-20, h/2+8, termbox.ColorYellow, termbox.ColorDefault, "Bem-vindo, "+player.Username+"!")
	termbox.Flush()
	time.Sleep(1500 * time.Millisecond)
	return true
}

func (g *Game) registerScreen() {
	termbox.Clear(termbox.ColorDefault, termbox.ColorDefault)
	w, h := termbox.Size()
	drawText(w/2-12, h/2-4, termbox.ColorCyan|termbox.AttrBold, termbox.ColorDefault, "CRIAR CONTA")

	username := g.inputString("Novo usuario:", h/2)
	if len(username) < 3 {
		g.showMessage("Usuario deve ter pelo menos 3 caracteres", termbox.ColorRed)
		return
	}

	password := g.inputString("Nova senha:", h/2+2)
	if len(password) < 4 {
		g.showMessage("Senha deve ter pelo menos 4 caracteres", termbox.ColorRed)
		return
	}

	_, err := RegisterPlayer(username, password)
	if err != nil {
		g.showMessage("Erro: "+err.Error(), termbox.ColorRed)
		return
	}
	g.showMessage("Conta criada com sucesso! Faca login agora.", termbox.ColorGreen)

}

func (g *Game) showMessage(text string, color termbox.Attribute) {
	w, h := termbox.Size()
	drawText(w/2-len(text)/2, h/2+6, color, termbox.ColorDefault, text)
	drawText(w/2-20, h/2+8, termbox.ColorWhite, termbox.ColorDefault, "Pressione qualquer tecla para continuar...")
	termbox.Flush()
	termbox.PollEvent()
}

func (g *Game) showMainMenu() {
	selected := 0
	options := []string{"Iniciar Jogo", "Ver Ranking", "Sair"}

	g.menuTicker = time.NewTicker(100 * time.Millisecond)
	defer g.menuTicker.Stop()

	go func() {
		for {
			select {
			case <-g.menuTicker.C:
				g.animateMenuSnake()
			case <-g.stopChan:
				return
			}
		}
	}()

	for {
		g.drawMainMenu(selected, options)
		termbox.Flush()

		ev := termbox.PollEvent()
		switch ev.Type {
		case termbox.EventKey:
			switch ev.Key {
			case termbox.KeyArrowUp:
				selected = (selected - 1 + len(options)) % len(options)
			case termbox.KeyArrowDown:
				selected = (selected + 1) % len(options)
			case termbox.KeyEnter:
				switch selected {
				case 0:
					g.startGame()
					return
				case 1:
					g.showLeaderboard()
					return
				case 2:
					return
				}
			case termbox.KeyEsc:
				return
			}
		}
	}
}

func (g *Game) animateMenuSnake() {
	width, height := termbox.Size()
	head := g.menuSnake[0]

	newHead := Coord{X: head.X + g.menuDir.X, Y: head.Y + g.menuDir.Y}

	// verifica colisao com bordas e mudar direcao
	if newHead.X <= 1 {
		g.menuDir = Coord{X: 1, Y: 0}
	} else if newHead.X >= width-2 {
		g.menuDir = Coord{X: -1, Y: 0}
	} else if newHead.Y <= 1 {
		g.menuDir = Coord{X: 0, Y: 1}
	} else if newHead.Y >= height-2 {
		g.menuDir = Coord{X: 0, Y: -1}
	} else {
		// verificar se esta na area do menu para evitar sobreposicao
		menuLeft := (width - 20) / 2
		menuRight := menuLeft + 20
		menuTop := height/2 - 2
		menuBottom := height/2 + 4

		if newHead.X >= menuLeft && newHead.X <= menuRight &&
			newHead.Y >= menuTop && newHead.Y <= menuBottom {
			if g.menuDir.X != 0 {
				g.menuDir = Coord{X: 0, Y: 1}
			} else {
				g.menuDir = Coord{X: 1, Y: 0}
			}
		}
	}

	// reaclcula nova posição com direção atualizada
	head = g.menuSnake[0]
	newHead = Coord{X: head.X + g.menuDir.X, Y: head.Y + g.menuDir.Y}

	g.menuSnake = append([]Coord{newHead}, g.menuSnake...)
	if len(g.menuSnake) > 10 { // tamanho da cobra do menu
		g.menuSnake = g.menuSnake[:10]
	}
}

func (g *Game) drawMainMenu(selected int, options []string) {
	termbox.Clear(termbox.ColorDefault, termbox.ColorDefault)
	width, height := termbox.Size()

	// desenhar cobrinha animada no fundo
	for i, seg := range g.menuSnake {
		color := termbox.ColorGreen
		if i == 0 {
			color = termbox.ColorGreen | termbox.AttrBold
		}
		termbox.SetCell(seg.X, seg.Y, '█', color, termbox.ColorDefault)
	}

	// game title
	title := "SNAKE GO - UFPI 2025"
	subtitle := "Ally,Vini, Kleber Versao.0.7"
	drawText((width-len(title))/2, height/2-5, termbox.ColorGreen|termbox.AttrBold, termbox.ColorDefault, title)
	drawText((width-len(subtitle))/2, height/2-4, termbox.ColorCyan, termbox.ColorDefault, subtitle)

	// op
	for i, option := range options {
		x := (width - 15) / 2
		y := height/2 - 1 + i*2

		fgColor := termbox.ColorWhite
		if i == selected {
			fgColor = termbox.ColorYellow | termbox.AttrBold
			drawText(x-2, y, fgColor, termbox.ColorDefault, ">")
		}

		drawText(x, y, fgColor, termbox.ColorDefault, option)
	}

	userInfo := fmt.Sprintf("Jogador: %s", g.userID)
	drawText(2, height-1, termbox.ColorBlue, termbox.ColorDefault, userInfo)

	controls := "Use ↑↓ para navegar, ENTER para selecionar, ESC para sair"
	drawText((width-len(controls))/2, height-2, termbox.ColorDarkGray, termbox.ColorDefault, controls)
}

func (g *Game) showLeaderboard() {
	inLeaderboard := true

	for inLeaderboard {
		termbox.Clear(termbox.ColorDefault, termbox.ColorDefault)

		width, height := termbox.Size()
		title := "RANKING - TOP 10"
		drawText((width-len(title))/2, 2, termbox.ColorYellow|termbox.AttrBold, termbox.ColorDefault, title)

		scores := GetTop10()
		if len(scores) == 0 {
			noScores := "Nenhum score registrado ainda!"
			drawText((width-len(noScores))/2, height/2, termbox.ColorWhite, termbox.ColorDefault, noScores)
		} else {
			// cabeçalho
			header := "Pos Jogador       Pontos Data"
			drawText((width-len(header))/2, 5, termbox.ColorCyan|termbox.AttrBold, termbox.ColorDefault, header)

			// separador
			separator := "-----------------------------"
			drawText((width-len(separator))/2, 6, termbox.ColorWhite, termbox.ColorDefault, separator)

			// pontuacoes
			for i, score := range scores {
				if i >= 10 {
					break
				}

				color := termbox.ColorWhite
				if i == 0 {
					color = termbox.ColorYellow | termbox.AttrBold
				} else if i == 1 {
					color = termbox.ColorWhite | termbox.AttrBold
				} else if i == 2 {
					color = termbox.ColorMagenta | termbox.AttrBold
				}

				playerDisplay := score.Nome
				if len(playerDisplay) > 12 {
					playerDisplay = playerDisplay[:12]
				}

				line := fmt.Sprintf("%2d. %-12s %6d %s",
					i+1, playerDisplay, score.Pontos, score.Data.Format("02/01"))

				drawText((width-len(line))/2, 7+i, color, termbox.ColorDefault, line)
			}
		}

		backMsg := "Pressione ESC para voltar ao menu"
		drawText((width-len(backMsg))/2, height-3, termbox.ColorGreen, termbox.ColorDefault, backMsg)

		termbox.Flush()

		ev := termbox.PollEvent()
		if ev.Type == termbox.EventKey && ev.Key == termbox.KeyEsc {
			inLeaderboard = false
		}
	}

	g.showMainMenu()
}

func (g *Game) startGame() {
	g.isRunning = true
	g.score = 0
	g.bonusActive = false
	g.bonusType = ""
	g.speed = 120 * time.Millisecond
	g.arena = newArena(60, 25)

	ticker := time.NewTicker(g.speed)
	defer ticker.Stop()

	eventQueue := make(chan termbox.Event)
	go func() {
		for g.isRunning {
			eventQueue <- termbox.PollEvent()
		}
	}()

	for g.isRunning {
		select {
		case ev := <-eventQueue:
			if ev.Type == termbox.EventKey {
				g.handleInput(ev)
			}
		case <-ticker.C:
			g.update()
			g.drawGame()
		}
	}

	g.gameOver()
}

func (g *Game) handleInput(ev termbox.Event) {
	// cheats: suposto a bugs
	if ev.Type == termbox.EventKey {
		switch ev.Ch {
		case 'g', 'G': // god mode
			g.arena.Snake.Body = append(g.arena.Snake.Body, g.arena.Snake.Body[len(g.arena.Snake.Body)-1])
			g.arena.Snake.Body = append(g.arena.Snake.Body, g.arena.Snake.Body[len(g.arena.Snake.Body)-1])
			g.arena.AddMessage("god mode, isso e uma maldicao", 3*time.Second)
		case 'p', 'P': // +1000 pontos instantaneos
			g.arena.Points += 1000
			g.arena.AddMessage("adm desligado, voce recebeu +1000 pts", 3*time.Second)
		case 'l', 'L': // subir de nível
			g.arena.Level += 5
			g.arena.increaseDifficulty()
			g.arena.AddMessage("voce recebeu uma dadiva! level +5", 3*time.Second)
		case 'b', 'B': // spawn boss instantâneo
			boss := newBoss(g.arena.Width, g.arena.Height, g.arena.Snake)
			g.arena.Bosses = append(g.arena.Bosses, boss)
			g.arena.AddMessage("um bug foi encontrado, um estrangeiro apareceu", 4*time.Second)
		case 'k', 'K': // matar todos os bosses
			for _, boss := range g.arena.Bosses {
				boss.IsAlive = false
			}
			g.arena.AddMessage("uma bencao divina extinguiu os estrangeiros", 4*time.Second)
		}
	}

	// mapeamento das teclas
	switch ev.Key {
	case termbox.KeyArrowUp:
		g.arena.Snake.ChangeDir(0, -1)
	case termbox.KeyArrowDown:
		g.arena.Snake.ChangeDir(0, 1)
	case termbox.KeyArrowLeft:
		g.arena.Snake.ChangeDir(-1, 0)
	case termbox.KeyArrowRight:
		g.arena.Snake.ChangeDir(1, 0)
	case termbox.KeyEsc:
		g.isRunning = false
	}
}

func (g *Game) update() {
	if !g.arena.Tick(g) {
		g.isRunning = false
	}
	g.score = g.arena.Points
}

func (g *Game) drawMessages() {
	now := time.Now()
	width, _ := termbox.Size()

	for i := len(g.arena.Messages) - 1; i >= 0; i-- {
		msg := g.arena.Messages[i]
		if now.Sub(msg.CreatedAt) < msg.Duration {
			x := (width - len(msg.Text)) / 2
			y := 2 + (len(g.arena.Messages)-1-i)*2
			drawText(x, y, termbox.ColorYellow|termbox.AttrBold, termbox.ColorDefault, msg.Text)
		}
	}
}

func (g *Game) drawGame() {
	termbox.Clear(termbox.ColorDefault, termbox.ColorDefault)

	// desenhar borda da arena
	g.drawArenaBorder()

	// desenhar obstáculos
	g.drawObstacles()

	// desenhar bosses
	g.drawBosses()

	// desenhar cobra
	g.drawSnake()

	// desenhar comida
	g.drawFood()

	// desenhar mensagens
	g.drawMessages()

	// desenhar HUD expandido
	g.drawHUD()

	termbox.Flush()
}

func (g *Game) drawArenaBorder() {
	// cantos
	termbox.SetCell(g.arena.X-1, g.arena.Y-1, '┌', termbox.ColorWhite, termbox.ColorDefault)
	termbox.SetCell(g.arena.X+g.arena.Width, g.arena.Y-1, '┐', termbox.ColorWhite, termbox.ColorDefault)
	termbox.SetCell(g.arena.X-1, g.arena.Y+g.arena.Height, '└', termbox.ColorWhite, termbox.ColorDefault)
	termbox.SetCell(g.arena.X+g.arena.Width, g.arena.Y+g.arena.Height, '┘', termbox.ColorWhite, termbox.ColorDefault)

	// bordas horizontais
	for x := g.arena.X; x < g.arena.X+g.arena.Width; x++ {
		termbox.SetCell(x, g.arena.Y-1, '─', termbox.ColorWhite, termbox.ColorDefault)
		termbox.SetCell(x, g.arena.Y+g.arena.Height, '─', termbox.ColorWhite, termbox.ColorDefault)
	}

	// bordas verticais
	for y := g.arena.Y; y < g.arena.Y+g.arena.Height; y++ {
		termbox.SetCell(g.arena.X-1, y, '│', termbox.ColorWhite, termbox.ColorDefault)
		termbox.SetCell(g.arena.X+g.arena.Width, y, '│', termbox.ColorWhite, termbox.ColorDefault)
	}
}

func (g *Game) drawObstacles() {
	for _, obs := range g.arena.Obstacles {
		char := '█'
		color := termbox.ColorMagenta
		if obs.IsTemporary {
			color = termbox.ColorMagenta | termbox.AttrBold
		}
		termbox.SetCell(obs.X, obs.Y, char, color, termbox.ColorDefault)
	}
}

func (g *Game) drawBosses() {
	for _, boss := range g.arena.Bosses {
		if !boss.IsAlive {
			continue
		}
		for i, seg := range boss.Body {
			color := termbox.ColorRed
			if i == 0 {
				color = termbox.ColorRed | termbox.AttrBold
			}
			char := '■'
			termbox.SetCell(seg.X, seg.Y, char, color, termbox.ColorDefault)
		}
	}
}

func (g *Game) drawSnake() {
	rainbowColors := []termbox.Attribute{
		termbox.ColorRed,
		termbox.ColorYellow,
		termbox.ColorGreen,
		termbox.ColorCyan,
		termbox.ColorBlue,
		termbox.ColorMagenta,
	}

	for i, seg := range g.arena.Snake.Body {
		var color termbox.Attribute

		if g.bonusActive {
			// rainbow se bônus ativo
			colorIdx := (i + int(time.Now().UnixNano()/100000000)) % len(rainbowColors)
			color = rainbowColors[colorIdx] | termbox.AttrBold
		} else {
			// color tradicional - green
			color = termbox.ColorGreen
			if i == 0 {
				color = termbox.ColorGreen | termbox.AttrBold
			}
		}

		char := '■'
		termbox.SetCell(seg.X, seg.Y, char, color, termbox.ColorDefault)
	}
}

func (g *Game) drawFood() {
	// desenhar todas as comidas
	for _, food := range g.arena.Foods {
		if food == nil {
			continue
		}

		var char rune
		var color termbox.Attribute

		// TODO: aqui vão os diferentes tipos de comida
		switch food.FoodType {
		case FOOD_NORMAL:
			char = '●' // fruta normal
			color = termbox.ColorRed | termbox.AttrBold
		case FOOD_BONUS:
			char = '★' // fruta para bônus
			color = termbox.ColorYellow | termbox.AttrBold
		case FOOD_PENALTY:
			char = '☠' // fruta para penalidade
			color = termbox.ColorGreen | termbox.AttrBold
		}

		// efetuar transparência baseada no tempo restante
		timeLeft := food.Lifetime - time.Since(food.SpawnTime)
		if timeLeft < 2*time.Second {
			if (time.Now().UnixNano()/500000000)%2 == 0 {
				color = color | termbox.AttrBlink
			}
		}

		termbox.SetCell(food.X, food.Y, char, color, termbox.ColorDefault)
	}
}

func (g *Game) drawHUD() {
	scoreText := fmt.Sprintf("Score: %d", g.score)
	drawText(g.arena.X+2, g.arena.Y-3, termbox.ColorYellow|termbox.AttrBold, termbox.ColorDefault, scoreText)

	levelText := fmt.Sprintf("Nivel: %d", g.arena.Level)
	drawText(g.arena.X+2, g.arena.Y-2, termbox.ColorCyan, termbox.ColorDefault, levelText)

	comboText := fmt.Sprintf("Combo: x%d", g.arena.ComboSystem.CurrentCombo+1)
	drawText(g.arena.X+25, g.arena.Y-2, termbox.ColorMagenta, termbox.ColorDefault, comboText)

	sizeText := fmt.Sprintf("Tamanho: %d", len(g.arena.Snake.Body))
	drawText(g.arena.X+45, g.arena.Y-2, termbox.ColorWhite, termbox.ColorDefault, sizeText)

	if g.bonusActive {
		bonusText := "BONUS: " + g.bonusType + "!"
		drawText(g.arena.X+g.arena.Width-len(bonusText)-4, g.arena.Y-3,
			termbox.ColorYellow|termbox.AttrBold|termbox.AttrBlink, termbox.ColorDefault, bonusText)
	}

	controls := "←↑→↓ mover • ESC sair"
	drawText(g.arena.X+2, g.arena.Y+g.arena.Height+1,
		termbox.ColorDarkGray, termbox.ColorDefault, controls)

	foodsText := fmt.Sprintf("Frutas: %d/%d", len(g.arena.Foods), g.arena.maxFoods)
	drawText(g.arena.X+g.arena.Width-len(foodsText)-4, g.arena.Y+g.arena.Height+1,
		termbox.ColorWhite, termbox.ColorDefault, foodsText)
}

func (g *Game) activateBonus(bonusType string) {
	// nao ativa novo bonus se ja estiver ativo
	if g.bonusActive {
		return
	}

	g.bonusActive = true
	g.bonusType = bonusType

	if g.bonusTimer != nil {
		g.bonusTimer.Stop()
	}

	switch bonusType {
	case "VELOCIDADE":
		g.speed = 60 * time.Millisecond // dobra velocidade
	case "CRESCIMENTO":
		// e para crescer instantaneamente
		for i := 0; i < 3; i++ {
			g.arena.Snake.Grow()
		}
	case "PONTOS":
		g.score += 50 // Bônus de pontos extra
	}

	g.bonusTimer = time.AfterFunc(5*time.Second, func() {
		g.bonusActive = false
		g.bonusType = ""
		if bonusType == "VELOCIDADE" {
			g.speed = 120 * time.Millisecond // return à velocidade normal
		}
	})
}

func (g *Game) gameOver() {
	// salva pontuacao
	if g.CurrentPlayer != nil {
		SaveScore(g.CurrentPlayer, g.score)
	} else {
		SaveScore(&Player{Username: g.userID}, g.score)
	}

	// limpar timer do bônus
	if g.bonusTimer != nil {
		g.bonusTimer.Stop()
	}

	selected := 0
	options := []string{"Jogar Novamente", "Ver Ranking", "Menu Principal"}

	for {
		termbox.Clear(termbox.ColorDefault, termbox.ColorDefault)

		width, height := termbox.Size()

		// game over
		gameOverText := "GAME OVER"
		drawText((width-len(gameOverText))/2, height/2-3, termbox.ColorRed|termbox.AttrBold, termbox.ColorDefault, gameOverText)

		// pontuacao final
		scoreText := fmt.Sprintf("Score Final: %d", g.score)
		drawText((width-len(scoreText))/2, height/2-1, termbox.ColorYellow, termbox.ColorDefault, scoreText)

		// nível alcançado
		levelText := fmt.Sprintf("Nivel Alcancado: %d", g.arena.Level)
		drawText((width-len(levelText))/2, height/2, termbox.ColorCyan, termbox.ColorDefault, levelText)

		// max combo
		comboText := fmt.Sprintf("Max Combo: x%d", g.arena.ComboSystem.MaxCombo+1)
		drawText((width-len(comboText))/2, height/2+1, termbox.ColorMagenta, termbox.ColorDefault, comboText)

		// op
		for i, option := range options {
			x := (width - 20) / 2
			y := height/2 + 3 + i*2

			fgColor := termbox.ColorWhite
			if i == selected {
				fgColor = termbox.ColorGreen | termbox.AttrBold
				drawText(x-2, y, fgColor, termbox.ColorDefault, ">")
			}

			drawText(x, y, fgColor, termbox.ColorDefault, option)
		}

		termbox.Flush()

		ev := termbox.PollEvent()
		switch ev.Type {
		case termbox.EventKey:
			switch ev.Key {
			case termbox.KeyArrowUp:
				selected = (selected - 1 + len(options)) % len(options)
			case termbox.KeyArrowDown:
				selected = (selected + 1) % len(options)
			case termbox.KeyEnter:
				switch selected {
				case 0:
					g.startGame()
					return
				case 1:
					g.showLeaderboard()
					return
				case 2:
					g.showMainMenu()
					return
				}
			case termbox.KeyEsc:
				g.showMainMenu()
				return
			}
		}
	}
}

func drawText(x, y int, fg, bg termbox.Attribute, text string) {
	for i, ch := range text {
		termbox.SetCell(x+i, y, ch, fg, bg)
	}
}
