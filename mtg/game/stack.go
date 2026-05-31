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

	// SourceID is the CardInstance ID for spells and hand-zone activated
	// abilities, or the Permanent's ObjectID for battlefield activated/triggered
	// abilities.
	SourceID id.ID

	// Face is the selected spell face or the source face captured when an
	// ability was put on the stack.
	Face FaceIndex

	// SourceCardID is the source CardInstance ID for activated/triggered
	// abilities when it is known. It lets ability resolution preserve source
	// identity even if the source permanent has left the battlefield.
	SourceCardID id.ID

	// SourceTokenDef is the source definition for token abilities, which have
	// no source CardInstance ID.
	SourceTokenDef *CardDef

	// AbilityIndex identifies the source ability for activated/triggered
	// abilities. It is ignored for spells.
	AbilityIndex int

	// InlineAbility stores generated abilities, such as delayed triggers, that
	// are not addressable by AbilityIndex on the source definition.
	InlineAbility *AbilityDef

	// TriggerEvent is the event that caused this triggered ability to trigger.
	// HasTriggerEvent distinguishes a real zero-valued event from no event.
	TriggerEvent    GameEvent
	HasTriggerEvent bool

	// WardTargetStackObjectID is the spell or ability a Ward trigger may
	// counter unless its controller pays the Ward cost.
	WardTargetStackObjectID id.ID

	// Controller is the player who controls this spell or ability.
	Controller PlayerID

	// Targets are the chosen runtime targets (permanents, players, etc.).
	// Targets are locked in when the spell/ability is put on the stack
	// (CR 601.2c, 603.3d).
	Targets []Target

	// ChosenModes are the indices of chosen modes for modal spells/abilities.
	ChosenModes []int

	// XValue is the chosen value of X for spells with {X} in their cost.
	XValue int

	// KickerPaid is true if the kicker cost was paid.
	KickerPaid bool

	// Flashback is true if this spell was cast from a graveyard using
	// flashback; it is exiled if it would leave the stack (CR 702.34).
	Flashback bool

	// Suspend is true if this spell was cast from exile by suspend.
	Suspend bool

	// FaceDown is true if this spell was cast face-down via Morph or Disguise.
	FaceDown bool

	// FaceDownFace records the printed face hidden under a face-down spell.
	// It is ignored unless FaceDown is true.
	FaceDownFace FaceIndex

	// FaceDownKind records whether this spell was cast face-down by Morph,
	// Disguise, or a future face-down mechanic.
	FaceDownKind FaceDownKind

	// Copy is true if this stack object is a copy of a spell rather than the
	// physical source card.
	Copy bool

	// AdditionalCostsPaid describes any additional costs that were paid
	// (e.g., "sacrificed a creature", "discarded a card").
	AdditionalCostsPaid []string

	// ResolvedAmounts stores named numeric results from earlier effects on this
	// stack object for "that much" style follow-up effects.
	ResolvedAmounts map[string]int

	// ResolutionResults stores named success/choice results from earlier
	// effects on this stack object for "if you do" / "if you don't" branches.
	ResolutionResults map[string]EffectResolutionResult

	// ResolutionChoices stores named values chosen while resolving this stack
	// object, for later instructions such as "of the chosen color" (CR 608.2c).
	ResolutionChoices map[string]ResolutionChoiceResult
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
func (s *Stack) Pop() (*StackObject, bool) {
	if len(s.objects) == 0 {
		return nil, false
	}
	top := s.objects[len(s.objects)-1]
	s.objects = s.objects[:len(s.objects)-1]
	return top, true
}

// Peek returns the top object without removing it.
func (s *Stack) Peek() (*StackObject, bool) {
	if len(s.objects) == 0 {
		return nil, false
	}
	top := s.objects[len(s.objects)-1]
	return top, true
}

// RemoveByID removes and returns the stack object with the given ID.
func (s *Stack) RemoveByID(objectID id.ID) (*StackObject, bool) {
	for i, obj := range s.objects {
		if obj.ID != objectID {
			continue
		}
		s.objects = append(s.objects[:i], s.objects[i+1:]...)
		return obj, true
	}
	return nil, false
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

// RemoveControlledBy removes stack objects controlled by playerID.
func (s *Stack) RemoveControlledBy(playerID PlayerID) {
	kept := s.objects[:0]
	for _, obj := range s.objects {
		if obj.Controller == playerID {
			continue
		}
		kept = append(kept, obj)
	}
	s.objects = kept
}
