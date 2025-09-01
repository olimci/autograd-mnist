package main

import (
    _ "embed"
    "encoding/json"

    "github.com/hajimehoshi/ebiten/v2"
    "github.com/olimci/autograd"
    "github.com/olimci/autograd/module"
    "github.com/olimci/autograd/tensor"
)

var (
    //go:embed mnist_state.json
    modelData []byte
)

func loadModel(model module.Module) error {
    state := make(map[string]*tensor.Tensor[float32])

    err := json.Unmarshal(modelData, &state)
    if err != nil {
        return err
    }

    model.SetState(state)

    return nil
}

func readVal(t *tensor.Tensor[float32]) []float32 {
    out := make([]float32, t.Size())
    for i := t.IterStride(); i.Done(); i.Next() {
        out[i.Count()] = t.Data[i.Offset()]
    }

    return out
}

func modelThread(model module.Module, sendCh chan []float32, recvCh chan []float32) {
    for {
        select {
        case img := <-sendCh:
            data := autograd.Const(tensor.New(img, 28, 28, 1))
            result := model.Forward(data)
            recvCh <- readVal(result.Val)
        }
    }
}

func main() {
    eval := true
    model := module.Sequential{
        module.Conv2D(1, 32, 3, 3, 1, 1, 1, 1).InitHe(),
        module.ReLU,
        module.Conv2D(32, 64, 3, 3, 1, 1, 1, 1).InitHe(),
        module.ReLU,
        module.MaxPool2D(2, 2, 2, 2),

        module.Conv2D(64, 128, 3, 3, 1, 1, 1, 1).InitHe(),
        module.ReLU,
        module.MaxPool2D(2, 2, 2, 2),

        module.Flatten,
        module.Affine(128*7*7, 256).InitHe(),
        module.ReLU,

        module.Affine(256, 10).InitXavier(),
        module.DoWhen(&eval, module.Softmax),
    }

    err := loadModel(&model)
    if err != nil {
        panic(err)
    }

    g := NewGame()

    go modelThread(&model, g.sendCh, g.recvCh)

    ebiten.SetWindowSize(640, 640)
    ebiten.SetWindowTitle("MNIST test")

    if err := ebiten.RunGame(g); err != nil {
        panic(err)
    }
}
