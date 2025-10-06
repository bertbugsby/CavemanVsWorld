package main

import (
	"fmt"
	_ "image/png"
	"log"
	"math/rand"
	"os"
	"slices"
	"strconv"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/audio"
	"github.com/hajimehoshi/ebiten/v2/audio/wav"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/solarlune/resolv"
)

type Game struct {
	playerColl *resolv.ConvexPolygon

	background      *ebiten.Image
	backgroundYView int

	player *ebiten.Image
	hurt   bool
	rock   *ebiten.Image
	enemy  *ebiten.Image

	SpriteLoc string

	xloc        int
	yloc        int
	speed       int
	rockSpeed   int
	allRocks    []Rock
	allEnemies  []Enemy
	revEnemies  []RevEnemy
	Health      int
	rockQ       int
	enemyQ      int
	spriteCount int
	timer       int
	spriteLen   int
	keyPressed  bool
	tSurvived   int
	maxY        int
	heart       *ebiten.Image
	sound       soundDemo
}

type RevEnemy struct {
	pict          *ebiten.Image
	collisionRect *resolv.ConvexPolygon
	xLoc          float64
	yLoc          float64
	SpriteLoc     string
	spriteLen     int
	spriteCount   int
	timer         int
	speed         int
}

type Enemy struct {
	pict          *ebiten.Image
	collisionRect *resolv.ConvexPolygon
	xLoc          float64
	yLoc          float64
	SpriteLoc     string
	spriteLen     int
	spriteCount   int
	timer         int
	speed         int
	hit           bool
}

type Rock struct {
	rockColl    *resolv.ConvexPolygon
	pict        *ebiten.Image
	xLoc        float64
	yLoc        float64
	SpriteLoc   string
	spriteCount int
	spriteLen   int
	timer       int
}

type soundDemo struct {
	audioContext *audio.Context
	Shoot        *audio.Player
	Hurt         *audio.Player
	Die          *audio.Player
	Kill         *audio.Player
	counter      int
}

func (g *Game) Update() error {
	err := g.checkPlayerCollision()
	if err != nil {
		return err
	}
	backgroundHeight := g.background.Bounds().Dy()
	maxY := backgroundHeight * 2
	g.backgroundYView -= 4
	g.backgroundYView %= maxY

	g.SpriteLoc = "walkUp/"
	frames, _ := os.ReadDir("walkUp")
	g.spriteLen = len(frames)
	g.keyPressed = true

	if ebiten.IsKeyPressed(ebiten.KeyW) && g.yloc > 0 {
		g.SpriteLoc = "walkUp/"
		frames, _ := os.ReadDir("walkUp")
		g.spriteLen = len(frames)
		g.keyPressed = true
		g.yloc -= g.speed
		g.playerColl.SetY(float64(g.yloc))
	}
	if ebiten.IsKeyPressed(ebiten.KeyS) && g.yloc < (g.background.Bounds().Dy()-g.player.Bounds().Dy()) {
		g.SpriteLoc = "walkDown/"
		frames, _ := os.ReadDir("walkDown")
		g.spriteLen = len(frames)
		g.keyPressed = true
		g.yloc += g.speed
		g.playerColl.SetY(float64(g.yloc))
	}

	if ebiten.IsKeyPressed(ebiten.KeyD) && g.xloc < (g.background.Bounds().Dx()-g.player.Bounds().Dx()) {
		g.SpriteLoc = "walkRight/"
		frames, _ := os.ReadDir("walkRight")
		g.spriteLen = len(frames)
		g.keyPressed = true
		g.xloc += g.speed
		g.playerColl.SetX(float64(g.xloc))
	}
	if ebiten.IsKeyPressed(ebiten.KeyA) && g.xloc > 0 {
		g.SpriteLoc = "walkLeft/"
		frames, _ := os.ReadDir("walkLeft")
		g.spriteLen = len(frames)
		g.keyPressed = true
		g.xloc -= g.speed
		g.playerColl.SetX(float64(g.xloc))
	}

	if inpututil.IsKeyJustPressed(ebiten.KeyT) && g.rockQ < 5 {
		g.allRocks = append(g.allRocks, NewRocks(g.xloc, g.yloc, g.rock))
		g.rockQ++
		err := g.sound.Shoot.Rewind()
		if err != nil {
			return err
		}
		g.sound.Shoot.Play()
	}

	g.spawnEnemy(rand.Intn(g.maxY))

	if g.allRocks != nil {
		ammoPassed := false
		for i := range g.allRocks {
			g.allRocks[i].yLoc -= float64(g.rockSpeed)
			g.allRocks[i].rockColl.SetY(g.allRocks[i].yLoc)
			if g.allRocks[i].yLoc < -100 {
				ammoPassed = true
			}
		}
		if ammoPassed {
			g.allRocks = slices.Delete(g.allRocks, 0, 1)
			g.rockQ -= 1
		}
	}

	if g.allEnemies != nil {
		enemyPassed := false
		for i := range g.allEnemies {
			g.allEnemies[i].yLoc += float64(g.allEnemies[i].speed)
			g.allEnemies[i].collisionRect.SetY(g.allEnemies[i].yLoc)
			if g.allEnemies[i].yLoc > float64(g.background.Bounds().Dy()) {

				enemyPassed = true
				if rand.Intn(2) == 1 && g.maxY < 280 {
					g.spawnRevEnemy(int(g.allEnemies[i].xLoc), int(g.allEnemies[i].yLoc))
				}
			}
		}

		if enemyPassed {
			g.allEnemies = slices.Delete(g.allEnemies, 0, 1)
			g.enemyQ -= 1
		}
	}

	if g.revEnemies != nil {
		enemyRevPassed := false
		for i := range g.revEnemies {
			g.revEnemies[i].yLoc -= float64(g.revEnemies[i].speed)
			g.revEnemies[i].collisionRect.SetY(g.revEnemies[i].yLoc)
			if g.revEnemies[i].yLoc < float64(-100) {
				enemyRevPassed = true
			}

		}

		if enemyRevPassed {
			g.revEnemies = slices.Delete(g.revEnemies, 0, 1)
		}
	}

	return nil
}

func (g *Game) Draw(screen *ebiten.Image) {
	drawOps := ebiten.DrawImageOptions{}
	//vWidth := float64(g.background.Bounds().Dx())
	vHeight := float64(g.background.Bounds().Dy())
	g.Animation(g.SpriteLoc)
	hx := 380
	hy := 460
	rx := 20
	ry := 460

	const repeat = 3
	for count := 0; count < repeat; count += 1 {
		drawOps.GeoM.Reset()
		drawOps.GeoM.Translate(0, -float64(int(vHeight)*count))
		drawOps.GeoM.Translate(0, -float64(g.backgroundYView))
		screen.DrawImage(g.background, &drawOps)

		if g.allRocks != nil {
			for i := range g.allRocks {
				drawOps.GeoM.Reset()
				drawOps.GeoM.Translate(g.allRocks[i].xLoc, g.allRocks[i].yLoc)
				g.rockAnimation(g.allRocks[i].SpriteLoc, i)
				screen.DrawImage(g.allRocks[i].pict, &drawOps)
			}
		}

		if g.allEnemies != nil {
			for i := range g.allEnemies {
				drawOps.GeoM.Reset()
				drawOps.GeoM.Translate(g.allEnemies[i].xLoc, g.allEnemies[i].yLoc)
				g.enemyAnimation(g.allEnemies[i].SpriteLoc, i)
				screen.DrawImage(g.allEnemies[i].pict, &drawOps)
			}
		}

		if g.revEnemies != nil {
			for i := range g.revEnemies {
				drawOps.GeoM.Reset()
				drawOps.GeoM.Translate(g.revEnemies[i].xLoc, g.revEnemies[i].yLoc)
				g.revEnemyAnimation(g.revEnemies[i].SpriteLoc, i)
				screen.DrawImage(g.revEnemies[i].pict, &drawOps)
			}
		}

		drawOps.GeoM.Reset()
		drawOps.GeoM.Translate(float64(g.xloc), float64(g.yloc))
		if g.hurt {
			time.Sleep(150 * time.Millisecond)
			g.hurt = false
		} else {
			screen.DrawImage(g.player, &drawOps)
		}

		for heart := range g.Health {
			drawOps.GeoM.Reset()
			drawOps.GeoM.Translate(float64(hx+(42*heart)), float64(hy))
			screen.DrawImage(g.heart, &drawOps)
		}
		rockImage, _, _ := ebitenutil.NewImageFromFile("rock.png")
		for rock := range 5 - (len(g.allRocks)) {
			drawOps.GeoM.Reset()
			drawOps.GeoM.Translate(float64(rx+(42*rock)), float64(ry))
			screen.DrawImage(rockImage, &drawOps)
		}
	}
}

const (
	SoundSampleRate = 12000
)

func (g *Game) Layout(int, int) (screenWidth int, screenHeight int) {
	return g.background.Bounds().Dx(), g.background.Bounds().Dy()
}

func main() {
	ebiten.SetWindowSize(512, 512)
	//ebiten.SetFullscreen(true)
	ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)
	ebiten.SetWindowTitle("CavemanVsWorld")

	soundContext := audio.NewContext(SoundSampleRate)

	cavemanColl := resolv.NewRectangle(256, 256, 5, 12)
	backgroundPict, _, err := ebitenutil.NewImageFromFile("background.png")
	if err != nil {
		fmt.Println("Unable to load background image:", err)
	}

	enemyImage, _, err := ebitenutil.NewImageFromFile("tigerRun/1-frame.png")
	if err != nil {
		fmt.Println("Unable to load player image:", err)
	}
	caveman, _, err := ebitenutil.NewImageFromFile("walkDown/1-frame.png")
	if err != nil {
		fmt.Println("Unable to load player image:", err)
	}
	rockImage, _, err := ebitenutil.NewImageFromFile("rockThrow/1-frame.png")
	if err != nil {
		fmt.Println("Unable to load bone:", err)
	}
	heartImage, _, err := ebitenutil.NewImageFromFile("heart.png")

	rockSlice := make([]Rock, 0)
	enemySlice := make([]Enemy, 0)
	enemySliceRev := make([]RevEnemy, 0)
	frames, _ := os.ReadDir("walkDown")
	soundGame := soundDemo{
		audioContext: soundContext,
		Hurt:         LoadWav("Sound/bat.wav", soundContext),
		Shoot:        LoadWav("Sound/bat.wav", soundContext),
		Die:          LoadWav("Sound/bat.wav", soundContext),
		Kill:         LoadWav("Sound/bat.wav", soundContext),
	}
	g := Game{
		playerColl:  cavemanColl,
		background:  backgroundPict,
		player:      caveman,
		hurt:        false,
		rock:        rockImage,
		enemy:       enemyImage,
		SpriteLoc:   "walkDown/",
		xloc:        256,
		yloc:        256,
		speed:       4,
		rockSpeed:   9,
		Health:      3,
		allRocks:    rockSlice,
		allEnemies:  enemySlice,
		revEnemies:  enemySliceRev,
		rockQ:       0,
		enemyQ:      0,
		spriteCount: 1,
		timer:       0,
		spriteLen:   len(frames),
		keyPressed:  false,
		tSurvived:   0,
		maxY:        350,
		heart:       heartImage,
		sound:       soundGame,
	}
	err = ebiten.RunGame(&g)
	if err := ebiten.RunGame(&g); err != nil {
		log.Fatal(err)
	}
}

func NewRocks(rockX, rockY int, image *ebiten.Image) Rock {
	return Rock{
		pict:        image,
		rockColl:    resolv.NewRectangle(float64(rockX), float64(rockY), 16, 16),
		xLoc:        float64(rockX),
		yLoc:        float64(rockY),
		SpriteLoc:   "rockThrow/",
		spriteCount: 1,
		spriteLen:   4,
		timer:       0,
	}
}

func NewEnemy(maxX, maxY int, image *ebiten.Image) Enemy {
	maxX = rand.Intn(maxX)
	return Enemy{
		pict:          image,
		collisionRect: resolv.NewRectangle(float64(maxX+60), float64(maxY), 16, 16),
		xLoc:          float64(maxX + 60),
		yLoc:          0,
		SpriteLoc:     "tigerRun/",
		spriteCount:   1,
		spriteLen:     4,
		timer:         0,
		speed:         3,
		hit:           false,
	}
}

func NewRevEnemy(maxX, maxY int, image *ebiten.Image) RevEnemy {
	return RevEnemy{
		pict:          image,
		collisionRect: resolv.NewRectangle(float64(maxX), float64(maxY), 16, 16),
		xLoc:          float64(maxX),
		yLoc:          float64(maxY),
		SpriteLoc:     "tigerReverse/",
		spriteCount:   1,
		spriteLen:     4,
		timer:         0,
		speed:         5,
	}
}

func (g *Game) Animation(fLocation string) {
	if g.keyPressed {
		spriteLocation := fLocation + strconv.Itoa(g.spriteCount) + "-frame.png"
		g.player, _, _ = ebitenutil.NewImageFromFile(spriteLocation)

		if g.timer == 10 {
			g.spriteCount++
			g.timer = 0
		} else {
			g.timer++
		}
		if g.spriteCount > g.spriteLen {
			g.spriteCount = 1
		}
	} else {
		g.timer = 0
		g.spriteCount = 1
		spriteLocation := fLocation + strconv.Itoa(g.spriteCount) + "-frame.png"
		g.player, _, _ = ebitenutil.NewImageFromFile(spriteLocation)
		g.keyPressed = false
	}
	g.keyPressed = false
}

func (g *Game) rockAnimation(fLocation string, slice int) {
	spriteLocation := fLocation + strconv.Itoa(g.allRocks[slice].spriteCount) + "-frame.png"
	g.allRocks[slice].pict, _, _ = ebitenutil.NewImageFromFile(spriteLocation)
	if g.allRocks[slice].timer == 5 {
		g.allRocks[slice].spriteCount++
		g.allRocks[slice].timer = 0
	} else {
		g.allRocks[slice].timer++
	}
	if g.allRocks[slice].spriteCount > g.allRocks[slice].spriteLen {
		g.allRocks[slice].spriteCount = 1
	}

}

func (g *Game) enemyAnimation(fLocation string, slice int) {
	spriteLocation := fLocation + strconv.Itoa(g.allEnemies[slice].spriteCount) + "-frame.png"
	g.allEnemies[slice].pict, _, _ = ebitenutil.NewImageFromFile(spriteLocation)
	if g.allEnemies[slice].timer == 10 {
		g.allEnemies[slice].spriteCount++
		g.allEnemies[slice].timer = 0
	} else {
		g.allEnemies[slice].timer++
	}
	if g.allEnemies[slice].spriteCount > g.allEnemies[slice].spriteLen {
		g.allEnemies[slice].spriteCount = 1
	}

}

func (g *Game) revEnemyAnimation(fLocation string, slice int) {
	spriteLocation := fLocation + strconv.Itoa(g.revEnemies[slice].spriteCount) + "-frame.png"
	g.revEnemies[slice].pict, _, _ = ebitenutil.NewImageFromFile(spriteLocation)
	if g.revEnemies[slice].timer == 10 {
		g.revEnemies[slice].spriteCount++
		g.revEnemies[slice].timer = 0
	} else {
		g.revEnemies[slice].timer++
	}
	if g.revEnemies[slice].spriteCount > g.revEnemies[slice].spriteLen {
		g.revEnemies[slice].spriteCount = 1
	}

}

func (g *Game) spawnEnemy(num int) {
	if num <= 10 && g.enemyQ < 10 {
		g.allEnemies = append(g.allEnemies, NewEnemy(390, 0, g.enemy))
		g.enemyQ++
		if g.maxY > 10 {
			g.tSurvived++
		}
	}
	if g.tSurvived == 10 {
		if g.maxY > 20 {
			g.maxY -= 10
			g.tSurvived = 0
		}
	}
}

func (g *Game) spawnRevEnemy(maxX, maxY int) {
	g.revEnemies = append(g.revEnemies, NewRevEnemy(maxX, maxY, g.enemy))
}

func (g *Game) checkPlayerCollision() error {
	c2 := 0
	for _, enemy := range g.allEnemies {
		c := 0
		rhit := false
		if hit :=
			g.playerColl.Intersection(
				enemy.collisionRect); !hit.IsEmpty() {
			enemy.collisionRect.SetX(-10)
			g.Health -= 1
			err := g.sound.Hurt.Rewind()
			if err != nil {
				return err
			}
			g.sound.Hurt.Play()
			g.hurt = true
			fmt.Println("Collision detected")
		} else {
			if g.allRocks != nil {
				for x := 0; x < len(g.allRocks); x++ {
					if hit :=
						g.allRocks[x].rockColl.Intersection(
							enemy.collisionRect); !hit.IsEmpty() {
						fmt.Println("Collision detected")
						rhit = true
						c = x
						break
					}
				}
			}
		}
		if g.Health == 0 {
			return ebiten.Termination
		}
		if rhit {
			g.allEnemies = slices.Delete(g.allEnemies, c2, c2+1)
			g.allRocks = slices.Delete(g.allRocks, c, c+1)
			g.rockQ -= 1
			g.enemyQ -= 1
			break
		}
		c2++
	}
	c2 = 0
	for _, rev := range g.revEnemies {
		c := 0
		rhit := false
		if hit :=
			g.playerColl.Intersection(
				rev.collisionRect); !hit.IsEmpty() {
			rev.collisionRect.SetX(-10)
			g.Health -= 1
			err := g.sound.Hurt.Rewind()
			if err != nil {
				return err
			}
			g.sound.Hurt.Play()

			g.hurt = true
			fmt.Println("Collision detected")
		} else {
			if g.allRocks != nil {
				for x := 0; x < len(g.allRocks); x++ {
					if hit :=
						g.allRocks[x].rockColl.Intersection(
							rev.collisionRect); !hit.IsEmpty() {
						fmt.Println("Collision detected")
						rhit = true
						c = x
						break
					}
				}
			}
		}
		if g.Health == 0 {
			return ebiten.Termination
		}
		if rhit {
			fmt.Println(c2, g.allEnemies[c2])
			g.revEnemies = slices.Delete(g.revEnemies, c2, c2+1)
			g.allRocks = slices.Delete(g.allRocks, c, c+1)
			g.rockQ -= 1
			g.enemyQ -= 1
			break
		}
		c2++
	}
	return nil
}

func LoadWav(name string, context *audio.Context) *audio.Player {
	thunderFile, err := os.Open(name)
	if err != nil {
		fmt.Println("Error Loading sound: ", err)
	}
	thunderSound, err := wav.DecodeWithoutResampling(thunderFile)
	if err != nil {
		fmt.Println("Error interpreting sound file: ", err)
	}
	soundPlayer, err := context.NewPlayer(thunderSound)
	if err != nil {
		fmt.Println("Couldn't create sound player: ", err)
	}
	return soundPlayer
}
