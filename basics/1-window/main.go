package main

import (
	_ "image/png"
	"log"

	"github.com/hajimehoshi/ebiten/v2"
)


type Game struct {
	count int
}

func (g *Game) Update() error {
	g.count++
	return nil
}

func (g *Game) Draw(screen *ebiten.Image) {
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	return 320, 240
}

func main() {

	ebiten.SetWindowSize(320, 240)
	ebiten.SetWindowTitle("Game Loop")
	if err := ebiten.RunGame(&Game{}); err != nil {
		log.Fatal(err)
	}
}
