package main

import (
    _ "embed"
    "image"
    "image/color"
    "math"

    "github.com/hajimehoshi/ebiten/v2"
    "github.com/hajimehoshi/ebiten/v2/ebitenutil"
    "github.com/hajimehoshi/ebiten/v2/inpututil"
)

const (
    canvasSize   = 280
    windowScale  = 2
    initialBrush = 16
)

type Game struct {
    started       bool
    canvas        *ebiten.Image
    brush         *ebiten.Image
    brushBaseSize int
    brushRadius   float64
    prevX, prevY  float64
    prevValid     bool
    tmp28         *ebiten.Image
    sendCh        chan []float32
    recvCh        chan []float32

    pred []float32
}

func NewGame() *Game {
    g := &Game{
        canvas:        ebiten.NewImage(canvasSize, canvasSize),
        brushBaseSize: 32,
        brushRadius:   initialBrush,
        sendCh:        make(chan []float32),
        recvCh:        make(chan []float32),
    }
    g.canvas.Fill(color.RGBA{})
    g.brush = makeFeatheredBrush(g.brushBaseSize, 0.01)

    return g
}

func makeFeatheredBrush(size int, hardness float64) *ebiten.Image {
    rgba := image.NewRGBA(image.Rect(0, 0, size, size))
    cx, cy := float64(size-1)/2, float64(size-1)/2
    r := math.Min(cx, cy)
    for y := 0; y < size; y++ {
        for x := 0; x < size; x++ {
            dx, dy := float64(x)-cx, float64(y)-cy
            t := math.Hypot(dx, dy) / (r - 10)
            if t <= 1 {

                alpha := math.Pow(1-t, hardness)
                a := uint8(math.Round(alpha * 255))
                rgba.SetRGBA(x, y, color.RGBA{R: 255, G: 255, B: 255, A: a})
            } else {
                rgba.SetRGBA(x, y, color.RGBA{})
            }
        }
    }
    return ebiten.NewImageFromImage(rgba)
}

func (g *Game) stamp(x, y float64, erasing bool) {
    scale := (g.brushRadius * 2) / float64(g.brushBaseSize)
    op := &ebiten.DrawImageOptions{}
    op.Filter = ebiten.FilterLinear
    op.GeoM.Translate(-float64(g.brushBaseSize)/2, -float64(g.brushBaseSize)/2)
    op.GeoM.Scale(scale, scale)
    op.GeoM.Translate(x, y)

    if erasing {
        op.Blend = ebiten.BlendDestinationOut
    } else {

        op.Blend = ebiten.BlendSourceOver
    }
    g.canvas.DrawImage(g.brush, op)
}

func (g *Game) Update() error {
    if inpututil.IsKeyJustPressed(ebiten.KeyEscape) {
        return ebiten.Termination
    }

    if !g.started {
        g.sendCh <- g.Vector28x28()
        g.started = true
        return nil
    }

    if inpututil.IsKeyJustPressed(ebiten.KeySpace) {
        g.canvas.Clear()
        g.prevValid = false
    }

    lmb := ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft)
    mx, my := ebiten.CursorPosition()
    x := float64(mx) / windowScale
    y := float64(my) / windowScale
    erasing := ebiten.IsKeyPressed(ebiten.KeyShift)

    if lmb {
        if !g.prevValid {
            g.prevX, g.prevY = x, y
            g.prevValid = true
            g.stamp(x, y, erasing)
        } else {

            dx, dy := x-g.prevX, y-g.prevY
            dist := math.Hypot(dx, dy)
            step := math.Max(1, g.brushRadius*0.4)
            steps := int(math.Ceil(dist / step))
            for i := 1; i <= steps; i++ {
                t := float64(i) / float64(steps)
                g.stamp(g.prevX+dx*t, g.prevY+dy*t, erasing)
            }
            g.prevX, g.prevY = x, y
        }
    } else {
        g.prevValid = false
    }

    // keep ping-ponging between the two channels
    select {
    case pred := <-g.recvCh:
        g.pred = pred
        g.sendCh <- g.Vector28x28()
    }

    return nil
}

func (g *Game) Draw(screen *ebiten.Image) {
    op := &ebiten.DrawImageOptions{}
    op.Filter = ebiten.FilterLinear
    op.GeoM.Scale(windowScale, windowScale)
    screen.Fill(color.Black)
    screen.DrawImage(g.canvas, op)

    if g.pred == nil {
        return
    }

    out := formatPredictions(g.pred)

    ebitenutil.DebugPrint(screen, out)
}

func (g *Game) Layout(_, _ int) (int, int) {
    return canvasSize * windowScale, canvasSize * windowScale
}

func (g *Game) Vector28x28() []float32 {
    if g.tmp28 == nil {
        g.tmp28 = ebiten.NewImage(28, 28)
    }

    g.tmp28.Clear()
    op := &ebiten.DrawImageOptions{}
    op.Filter = ebiten.FilterLinear
    scale := 28.0 / float64(canvasSize)
    op.GeoM.Scale(scale, scale)
    g.tmp28.DrawImage(g.canvas, op)

    buf := make([]byte, 28*28*4)
    g.tmp28.ReadPixels(buf)

    out := make([]float32, 28*28)
    for i := 0; i < 28*28; i++ {
        a := buf[i*4+3]
        out[i] = float32(a) / 255.0
    }
    return out
}
