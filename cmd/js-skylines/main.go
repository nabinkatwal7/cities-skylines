package main

import (
	"time"

	rl "github.com/gen2brain/raylib-go/raylib"

	"github.com/katwate/js-skylines/internal/terrain"
)

const (
	screenWidth  = 1280
	screenHeight = 720
)

type Game struct {
	camera   rl.Camera3D
	gridSize int32
	terrain  *terrain.Manager
}

func NewGame() *Game {
	cam := rl.Camera3D{
		Position:   rl.NewVector3(0, 80, 80),
		Target:     rl.NewVector3(0, 0, 0),
		Up:         rl.NewVector3(0, 1, 0),
		Fovy:       60,
		Projection: rl.CameraPerspective,
	}

	t := terrain.NewManager(time.Now().UnixNano())
	t.Generate()

	return &Game{
		camera:   cam,
		gridSize: 40,
		terrain:  t,
	}
}

func (g *Game) Update() {
	speed := float32(20.0)
	if rl.IsKeyDown(rl.KeyW) {
		g.camera.Position.Z -= speed
		g.camera.Target.Z -= speed
	}
	if rl.IsKeyDown(rl.KeyS) {
		g.camera.Position.Z += speed
		g.camera.Target.Z += speed
	}
	if rl.IsKeyDown(rl.KeyA) {
		g.camera.Position.X -= speed
		g.camera.Target.X -= speed
	}
	if rl.IsKeyDown(rl.KeyD) {
		g.camera.Position.X += speed
		g.camera.Target.X += speed
	}
}

func (g *Game) Draw() {
	rl.BeginDrawing()
	rl.ClearBackground(rl.NewColor(135, 206, 235, 255))

	rl.BeginMode3D(g.camera)

	g.terrain.Draw()
	rl.DrawGrid(g.gridSize, 2.0)

	rl.EndMode3D()

	rl.DrawFPS(10, 10)
	rl.DrawText("Ciities Skylines", 10, 30, 20, rl.Black)

	rl.EndDrawing()
}

func main() {
	rl.SetConfigFlags(rl.FlagMsaa4xHint)
	rl.InitWindow(screenWidth, screenHeight, "JS Skylines - Go Edition")
	defer rl.CloseWindow()

	rl.SetTargetFPS(60)

	game := NewGame()

	for !rl.WindowShouldClose() {
		game.terrain.Update(float64(rl.GetFrameTime()))
		game.Update()
		game.Draw()
	}

	game.terrain.Unload()
}
