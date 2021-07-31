package main

import (
	"github.com/hajimehoshi/ebiten"
	"github.com/hajimehoshi/ebiten/inpututil"
	"runtime"
)

// Game is the structure of the game state
type Game struct {
	state    GameState
	aiMode   bool
	ball     *Ball
	player1  *Paddle
	player2  *Paddle
	rally    int
	level    int
	maxScore int
}

const (
	initBallVelocity = 5.0
	initPaddleSpeed  = 10.0
	speedUpdateCount = 6
	speedIncrement   = 0.5
)

const (
	windowWidth  = 800
	windowHeight = 600
)

// NewGame creates an initializes a new game
func NewGame(aiMode bool) *Game {
	g := &Game{}
	g.init(aiMode)
	return g
}

func (g *Game) init(aiMode bool) {
	g.state = StartState
	g.aiMode = aiMode
	if aiMode {
		g.maxScore = 100
	} else {
		g.maxScore = 11
	}

	g.player1 = &Paddle{
		Position: Position{
			X: InitPaddleShift,
			Y: float32(windowHeight / 2)},
		Score:  0,
		Speed:  initPaddleSpeed,
		Width:  InitPaddleWidth,
		Height: InitPaddleHeight,
		Color:  ObjColor,
		Up:     ebiten.KeyW,
		Down:   ebiten.KeyS,
	}
	g.player2 = &Paddle{
		Position: Position{
			X: windowWidth - InitPaddleShift - InitPaddleWidth,
			Y: float32(windowHeight / 2)},
		Score:  0,
		Speed:  initPaddleSpeed,
		Width:  InitPaddleWidth,
		Height: InitPaddleHeight,
		Color:  ObjColor,
		Up:     ebiten.KeyO,
		Down:   ebiten.KeyK,
	}
	g.ball = &Ball{
		Position: Position{
			X: float32(windowWidth / 2),
			Y: float32(windowHeight / 2)},
		Radius:    InitBallRadius,
		Color:     ObjColor,
		XVelocity: initBallVelocity,
		YVelocity: initBallVelocity,
	}
	g.level = 0
	g.ball.Img, _ = ebiten.NewImage(int(g.ball.Radius*2), int(g.ball.Radius*2), ebiten.FilterDefault)
	g.player1.Img, _ = ebiten.NewImage(g.player1.Width, g.player1.Height, ebiten.FilterDefault)
	g.player2.Img, _ = ebiten.NewImage(g.player2.Width, g.player2.Height, ebiten.FilterDefault)

	InitFonts()
}

func (g *Game) reset(screen *ebiten.Image, state GameState) {
	w, _ := screen.Size()
	g.state = state
	g.rally = 0
	g.level = 0
	if state == StartState {
		g.player1.Score = 0
		g.player2.Score = 0
	}
	g.player1.Position = Position{
		X: InitPaddleShift, Y: GetCenter(screen).Y}
	g.player2.Position = Position{
		X: float32(w - InitPaddleShift - InitPaddleWidth), Y: GetCenter(screen).Y}
	g.ball.Position = GetCenter(screen)
	g.ball.XVelocity = initBallVelocity
	g.ball.YVelocity = initBallVelocity
}

// Update updates the game state
func (g *Game) Update(screen *ebiten.Image) error {
	switch g.state {
	case StartState:
		if inpututil.IsKeyJustPressed(ebiten.KeyC) {
			g.state = ControlsState
		} else if inpututil.IsKeyJustPressed(ebiten.KeyA) {
			g.aiMode = true
			g.state = PlayState
		} else if inpututil.IsKeyJustPressed(ebiten.KeyV) {
			g.aiMode = false
			g.state = PlayState
		}

	case ControlsState:
		if inpututil.IsKeyJustPressed(ebiten.KeySpace) {
			g.state = StartState
		}
	case PlayState:
		w, _ := screen.Size()

		if inpututil.IsKeyJustPressed(ebiten.KeySpace) {
			g.state = PauseState
			break
		}

		g.player1.Update(screen)
		if g.aiMode {
			g.player2.AiUpdate(g.ball)
		} else {
			g.player2.Update(screen)
		}

		xV := g.ball.XVelocity
		g.ball.Update(g.player1, g.player2, screen)
		// rally count
		if xV*g.ball.XVelocity < 0 {
			// score up when ball touches human player's paddle
			if g.aiMode && g.ball.X < float32(w/2) {
				g.player1.Score++
			}

			g.rally++

			// spice things up
			if (g.rally)%speedUpdateCount == 0 {
				g.level++
				g.ball.XVelocity += speedIncrement
				g.ball.YVelocity += speedIncrement
				g.player1.Speed += speedIncrement
				g.player2.Speed += speedIncrement
			}
		}

		if g.ball.X < 0 {
			g.player2.Score++
			if g.aiMode {
				g.state = GameOverState
				break
			}
			g.reset(screen, InterState)
		} else if g.ball.X > float32(w) {
			g.player1.Score++
			if g.aiMode {
				g.state = GameOverState
				break
			}
			g.reset(screen, InterState)
		}

		if g.player1.Score == g.maxScore || g.player2.Score == g.maxScore {
			g.state = GameOverState
		}

	case InterState, PauseState:
		if inpututil.IsKeyJustPressed(ebiten.KeySpace) {
			g.state = PlayState
		} else if inpututil.IsKeyJustPressed(ebiten.KeyR) {
			g.reset(screen, StartState)
		}

	case GameOverState:
		if inpututil.IsKeyJustPressed(ebiten.KeySpace) {
			g.reset(screen, StartState)
		}
	}

	g.Draw(screen)

	return nil
}

// Draw updates the game screen elements drawn
func (g *Game) Draw(screen *ebiten.Image) error {
	screen.Fill(BgColor)

	DrawCaption(g.state, ObjColor, screen)
	DrawBigText(g.state, ObjColor, screen)

	if g.state != ControlsState {
		g.player1.Draw(screen, ArcadeFont, false)
		g.player2.Draw(screen, ArcadeFont, g.aiMode)
		g.ball.Draw(screen)
	}

	//ebitenutil.DebugPrint(screen, fmt.Sprintf("TPS: %0.2f", ebiten.CurrentTPS()))

	return nil
}

// Layout sets the screen layout
func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	return windowWidth, windowHeight
}

func main() {
	// On browsers, let's use fullscreen so that this is playable on any browsers.
	// It is planned to ignore the given 'scale' apply fullscreen automatically on browsers (#571).
	if runtime.GOARCH == "js" || runtime.GOOS == "js" {
		ebiten.SetFullscreen(true)
	}
	ai := true
	g := NewGame(ai)
	if err := ebiten.RunGame(g); err != nil {
		panic(err)
	}
}
