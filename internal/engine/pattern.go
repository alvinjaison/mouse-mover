package engine

import "fmt"

// Pattern identifies a small cursor movement shape.
type Pattern int

const (
	// PatternNudge moves the cursor by +delta on both axes, then back.
	PatternNudge Pattern = iota
	// PatternJiggle moves +delta, then -delta, then back to origin.
	PatternJiggle
	// PatternSquare traces a tiny square and returns to origin.
	PatternSquare
)

// String returns the canonical lowercase name of the pattern.
func (p Pattern) String() string {
	switch p {
	case PatternNudge:
		return "nudge"
	case PatternJiggle:
		return "jiggle"
	case PatternSquare:
		return "square"
	default:
		return fmt.Sprintf("pattern(%d)", int(p))
	}
}

// ParsePattern maps a lowercase name to a Pattern. Unknown names return an error.
func ParsePattern(name string) (Pattern, error) {
	switch name {
	case "nudge":
		return PatternNudge, nil
	case "jiggle":
		return PatternJiggle, nil
	case "square":
		return PatternSquare, nil
	default:
		return 0, fmt.Errorf("unknown movement pattern %q", name)
	}
}

// offsets returns the sequence of relative (dx, dy) offsets from the origin that
// the pattern visits, in order. The engine applies each offset as an absolute
// MoveTo (origin + offset, clamped) and always restores to origin afterwards.
func (p Pattern) offsets(delta int32) []Point {
	if delta < 1 {
		delta = 1
	}
	switch p {
	case PatternNudge:
		return []Point{{X: delta, Y: delta}}
	case PatternJiggle:
		return []Point{{X: delta, Y: delta}, {X: -delta, Y: -delta}}
	case PatternSquare:
		return []Point{
			{X: delta, Y: 0},
			{X: delta, Y: delta},
			{X: 0, Y: delta},
			{X: 0, Y: 0},
		}
	default:
		return []Point{{X: delta, Y: delta}}
	}
}
