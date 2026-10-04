package game

import (
	"errors"
	"fmt"

	"github.com/natefinch/council4/mtg/game/zone"
)

type resultObjectPublisher interface {
	publishesResultObjects() bool
	publishesResultCards() bool
}

func (Discard) publishesResultObjects() bool             { return true }
func (Destroy) publishesResultObjects() bool             { return true }
func (Sacrifice) publishesResultObjects() bool           { return true }
func (SacrificePermanents) publishesResultObjects() bool { return true }
func (MovePermanent) publishesResultObjects() bool       { return true }
func (MoveTopOfLibrary) publishesResultObjects() bool    { return true }
func (p MoveCard) publishesResultObjects() bool          { return p.Destination != zone.Battlefield }
func (p ChooseFromZone) publishesResultObjects() bool {
	return p.Destination.Zone != zone.Battlefield &&
		(!p.SplitSecondary.Exists || p.SplitSecondary.Val.Destination.Zone != zone.Battlefield)
}

func (Discard) publishesResultCards() bool             { return true }
func (Destroy) publishesResultCards() bool             { return false }
func (Sacrifice) publishesResultCards() bool           { return false }
func (SacrificePermanents) publishesResultCards() bool { return false }
func (MovePermanent) publishesResultCards() bool       { return false }
func (MoveTopOfLibrary) publishesResultCards() bool    { return true }
func (p MoveCard) publishesResultCards() bool          { return p.publishesResultObjects() }
func (p ChooseFromZone) publishesResultCards() bool    { return p.publishesResultObjects() }

// ValidateResultObjectSelection checks the characteristic atoms available for
// both card results and permanent departure snapshots.
func ValidateResultObjectSelection(selection Selection) []string {
	problems := selection.Validate()
	for _, alternative := range selection.AnyOf {
		problems = append(problems, ValidateResultObjectSelection(alternative)...)
	}
	selection.AnyOf = nil
	selection.RequiredTypes = nil
	selection.RequiredTypesAny = nil
	selection.ExcludedTypes = nil
	selection.Supertypes = nil
	selection.ExcludedSupertype = ""
	selection.SubtypesAny = nil
	selection.ExcludedSubtype = ""
	selection.ColorsAny = nil
	selection.ExcludedColors = nil
	selection.Colorless = false
	selection.Multicolored = false
	selection.Colored = false
	selection.Name = ""
	if !selection.Empty() {
		problems = append(problems, "result object selection requires unsupported contextual predicates")
	}
	return problems
}

func validateResultObjectGate(gate InstructionResultGate, sequence []Instruction, published map[ResultKey]int) error {
	if gate.Negate && !gate.ObjectSelection.Exists {
		return errors.New("negated result gate requires an object selection")
	}
	if gate.CardOnly && !gate.ObjectSelection.Exists {
		return errors.New("card-only result gate requires an object selection")
	}
	if !gate.ObjectSelection.Exists {
		return nil
	}
	if gate.Key == "" || gate.Succeeded != TriTrue {
		return errors.New("object selection requires a named successful result")
	}
	if problems := ValidateResultObjectSelection(gate.ObjectSelection.Val); len(problems) > 0 {
		return fmt.Errorf("object selection: %s", problems[0])
	}
	index, ok := published[gate.Key]
	if !ok {
		return fmt.Errorf("object selection references result %q not yet published", gate.Key)
	}
	publisher, ok := sequence[index].Primitive.(resultObjectPublisher)
	if !ok || !publisher.publishesResultObjects() {
		return fmt.Errorf("result %q does not publish actual result objects", gate.Key)
	}
	if gate.CardOnly && !publisher.publishesResultCards() {
		return fmt.Errorf("result %q does not publish post-move card objects", gate.Key)
	}
	return nil
}
