package game

import "fmt"

// Vote resolves the "Starting with you, each player votes for <A> or <B>."
// voting interaction (CR 701.32). Each non-eliminated player, starting with the
// resolving controller and proceeding in turn order, votes for one of the named
// options. The instruction publishes a result whose Amount is the signed vote
// margin Options[0] minus Options[1], so downstream arm instructions gate on the
// margin's sign through a result-gate amount range: a positive margin means the
// first option won, a negative margin the second, and zero a tie.
//
// Options carries the printed choice labels in their printed order. The labels
// are display text only: the voting logic keys on option index, never on the
// label spelling, so the runtime stays text-blind.
type Vote struct {
	Options []string
}

// Kind implements Primitive for Vote.
func (Vote) Kind() PrimitiveKind { return PrimitiveVote }

func (Vote) isPrimitive() {}

func (Vote) instructionRefs() primitiveRefs { return primitiveRefs{} }

// validatePrimitive requires exactly two named options, the binary vote shape
// whose signed margin the result-gate amount range encodes.
func (p Vote) validatePrimitive([]TargetSpec, bool) error {
	if len(p.Options) != 2 {
		return fmt.Errorf("Vote requires exactly two options, got %d", len(p.Options))
	}
	for i, option := range p.Options {
		if option == "" {
			return fmt.Errorf("Vote option %d is empty", i)
		}
	}
	return nil
}
