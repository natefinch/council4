package game

import "github.com/natefinch/council4/mtg/game/id"

// StackObjectKind classifies what kind of object is on the stack.
type StackObjectKind int

const (
	// StackSpell is a spell being cast (instant, sorcery, or permanent spell).
	StackSpell StackObjectKind = iota

	// StackActivatedAbility is an activated ability on the stack.
	StackActivatedAbility

	// StackTriggeredAbility is a triggered ability on the stack.
	StackTriggeredAbility
)

// StackObject represents a spell or ability on the stack waiting to resolve.
// The stack resolves last-in, first-out (LIFO).
type StackObject struct {
	// ID is the unique identifier for this stack object.
	ID id.ID

	// Kind classifies this as a spell, activated ability, or triggered ability.
	Kind StackObjectKind

	// SourceID is the CardInstance ID for spells, or the Permanent's ObjectID
	// for activated/triggered abilities.
	SourceID id.ID

	// Controller is the player who controls this spell or ability.
	Controller PlayerID

	// Targets are the chosen target IDs (permanents, players, etc.).
	// Targets are locked in when the spell/ability is put on the stack
	// (CR 601.2c, 603.3d).
	Targets []id.ID

	// ChosenModes are the indices of chosen modes for modal spells/abilities.
	ChosenModes []int

	// XValue is the chosen value of X for spells with {X} in their cost.
	XValue int

	// KickerPaid is true if the kicker cost was paid.
	KickerPaid bool

	// AdditionalCostsPaid describes any additional costs that were paid
	// (e.g., "sacrificed a creature", "discarded a card").
	AdditionalCostsPaid []string
}

// Stack represents the game stack — the zone where spells and abilities
// wait to resolve in last-in, first-out (LIFO) order (CR 405).
type Stack struct {
	objects []*StackObject
}

// Push adds an object to the top of the stack.
func (s *Stack) Push(obj *StackObject) {
	s.objects = append(s.objects, obj)
}

// Pop removes and returns the top object from the stack.
// Returns nil if the stack is empty.
func (s *Stack) Pop() *StackObject {
	if len(s.objects) == 0 {
		return nil
	}
	top := s.objects[len(s.objects)-1]
	s.objects = s.objects[:len(s.objects)-1]
	return top
}

// Peek returns the top object without removing it.
// Returns nil if the stack is empty.
func (s *Stack) Peek() *StackObject {
	if len(s.objects) == 0 {
		return nil
	}
	return s.objects[len(s.objects)-1]
}

// IsEmpty reports whether the stack has no objects.
func (s *Stack) IsEmpty() bool {
	return len(s.objects) == 0
}

// Size returns the number of objects on the stack.
func (s *Stack) Size() int {
	return len(s.objects)
}

// Objects returns a copy of all objects on the stack, from bottom to top.
func (s *Stack) Objects() []*StackObject {
	result := make([]*StackObject, len(s.objects))
	copy(result, s.objects)
	return result
}
