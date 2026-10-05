package game

// LocalProducts declares parser-owned products whose lifetime is one instruction
// sequence invocation. Other links and results, including paid costs and CR 607
// ability links, retain their existing persistent lifetime.
type LocalProducts struct {
	Results []ResultKey
	Links   []LinkedKey
}

func (p LocalProducts) HasResult(key ResultKey) bool {
	for _, declared := range p.Results {
		if declared == key {
			return true
		}
	}
	return false
}

func (p LocalProducts) HasLink(key LinkedKey) bool {
	for _, declared := range p.Links {
		if declared == key {
			return true
		}
	}
	return false
}
