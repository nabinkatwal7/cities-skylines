package main

import (
	"time"

	rl "github.com/gen2brain/raylib-go/raylib"
)

const (
	screenWidth  = 1280
	screenHeight = 720
)

type Game struct {
	camera rl.Camera3D
	gridSize int32
}

func NewGame() *Game {
	return &Game{
		camera: rl.Camera3D{
			Position:   rl.NewVector3(50, 50, 50),
			Target:     rl.NewVector3(0, 0, 0),
			Up:         rl.NewVector3(0, 1, 0),
			Fovy:       60,
			Projection: rl.CameraPerspective,
		},
		gridSize: 40,
	}
}

func (g *Game) Update() {
	rl.UpdateCamera(&g.camera, rl.CameraOrbital)
	g.camera.Up = rl.NewVector3(0, 1, 0)
}

func (g *Game) Draw() {
	rl.BeginDrawing()
	rl.ClearBackground(rl.NewColor(135, 206, 235, 255))

	rl.BeginMode3D(g.camera)

	rl.DrawGrid(g.gridSize, 2.0)

	rl.DrawCube(rl.NewVector3(0, 0.5, 0), 2, 1, 2, rl.Brown)
	rl.DrawCube(rl.NewVector3(3, 0.5, 0), 2, 1, 2, rl.Gray)
	rl.DrawCube(rl.NewVector3(-3, 0.5, 0), 2, 1, 2, rl.Red)
	rl.DrawCube(rl.NewVector3(0, 0.5, 3), 2, 1, 2, rl.Yellow)
	rl.DrawCube(rl.NewVector3(0, 0.5, -3), 1, 1, 1, rl.Green)

	rl.EndMode3D()

	rl.DrawFPS(10, 10)
	rl.DrawText("JS Skylines - Go Edition", 10, 30, 20, rl.Black)

	rl.EndDrawing()
}

func main() {
	rl.SetConfigFlags(rl.FlagMsaa4xHint)
	rl.InitWindow(screenWidth, screenHeight, "JS Skylines - Go Edition")
	defer rl.CloseWindow()

	rl.SetTargetFPS(60)

	game := NewGame()
	lastTime := time.Now()

	for !rl.WindowShouldClose() {
		dt := time.Since(lastTime).Seconds()
		lastTime = time.Now()

		_ = dt

		game.Update()
		game.Draw()
	}
}
