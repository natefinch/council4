package shared

// SourceOrder is a node's position in its ability's (or mode's) dense
// source-order ranking. The parser assigns these monotonic ranks so downstream
// stages can reason about source order and structural nesting without
// inspecting raw byte offsets. Start and End are the ranks of the node span's
// start and end boundaries within the ability-wide union of every boundary
// offset; ranks preserve the relative order of those boundaries while
// discarding absolute positions.
type SourceOrder struct {
	Start int
	End   int
}

// Contains reports whether the outer source-order range encloses the inner one.
// It is the rank-space equivalent of testing that one span structurally
// contains another, with no byte-offset arithmetic.
func (outer SourceOrder) Contains(inner SourceOrder) bool {
	return outer.Start <= inner.Start && outer.End >= inner.End
}
