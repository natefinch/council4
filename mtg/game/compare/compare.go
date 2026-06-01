// Package compare provides reusable comparison predicates for game data.
package compare

// Op identifies an integer comparison operation.
type Op int

const (
	Any Op = iota
	Equal
	LessOrEqual
	GreaterOrEqual
	LessThan
	GreaterThan
)

// Int is a simple comparison against a fixed integer value.
type Int struct {
	Op    Op
	Value int
}

// Matches reports whether value satisfies the comparison.
func (c Int) Matches(value int) bool {
	switch c.Op {
	case Equal:
		return value == c.Value
	case LessOrEqual:
		return value <= c.Value
	case GreaterOrEqual:
		return value >= c.Value
	case LessThan:
		return value < c.Value
	case GreaterThan:
		return value > c.Value
	default:
		return true
	}
}
