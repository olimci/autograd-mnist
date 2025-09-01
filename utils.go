package main

import (
    "fmt"
    "math"
)

// bar8 returns a bar of length `width` characters representing v in [0,1].
// It uses 1/8th-width block characters for a smooth edge.
func bar10(v float64, width int) string {
    if width <= 0 {
        return ""
    }
    if v < 0 {
        v = 0
    } else if v > 1 {
        v = 1
    }

    // round to nearest integer number of filled chars
    filled := int(math.Round(v * float64(width)))
    if filled > width {
        filled = width
    }

    bar := ""
    for i := 0; i < filled; i++ {
        bar += "#"
    }
    for i := filled; i < width; i++ {
        bar += " "
    }
    return bar
}

// formatPredictions prints "i: v  |bar|" lines.
// pred must already be normalized [0,1].
func formatPredictions(pred []float32) string {
    out := ""
    for i := 0; i < len(pred); i++ {
        out += fmt.Sprintf("%d: %.2f  %s\n", i, pred[i], bar10(float64(pred[i]), 10))
    }
    return out
}
