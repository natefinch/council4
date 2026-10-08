package game

import "slices"

// LocalProducts declares parser-owned products whose lifetime is one instruction
// sequence invocation. Other links and results, including paid costs and CR 607
// ability links, retain their existing persistent lifetime.
type LocalProducts struct {
	Results []ResultKey
	Links   []LinkedKey
}

// HasResult reports whether this invocation owns the named result cell.
func (p LocalProducts) HasResult(key ResultKey) bool {
	return slices.Contains(p.Results, key)
}

// HasLink reports whether this invocation owns the named linked-object cell.
func (p LocalProducts) HasLink(key LinkedKey) bool {
	return slices.Contains(p.Links, key)
}
