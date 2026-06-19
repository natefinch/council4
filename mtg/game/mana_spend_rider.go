package game

import (
	"github.com/natefinch/council4/mtg/game/id"
	"github.com/natefinch/council4/mtg/game/mana"
)

// ManaSpendConditionKind identifies the exact condition under which a mana
// spend rider fires. The set is closed: only fully modeled spend conditions are
// represented, so any other "when that mana is spent" wording fails closed in
// the parser rather than mapping to a generic rider here.
type ManaSpendConditionKind int

// Mana spend condition kinds.
const (
	ManaSpendConditionUnknown ManaSpendConditionKind = iota
	// ManaSpendCastCommanderCreatureType is "spent to cast a creature spell that
	// shares a creature type with your commander" (Path of Ancestry). The rider
	// fires once for each unit of tagged mana spent to cast such a spell.
	ManaSpendCastCommanderCreatureType
)

// ManaSpendRider is a one-shot delayed triggered ability associated with a unit
// of produced mana. When that mana is spent to cast a spell that satisfies
// Condition, Effect resolves for the mana's controller using the stack. Mana
// produced by a mana ability stays a mana ability (CR 605); the rider is a
// separate triggered ability.
type ManaSpendRider struct {
	// Condition is the closed, fully modeled spend condition that must hold for
	// the rider to fire.
	Condition ManaSpendConditionKind
	// Effect is the rider's resolving content (for Path of Ancestry, "scry 1").
	Effect Mode
}

// ManaRiderInstance is a runtime record that a unit of mana carrying a spend
// rider sits in a player's mana pool. One instance is created for each unit of
// tagged mana produced; spending or emptying the mana removes the instance. The
// instance is the rider's individual identity: it is fired or dropped on the
// exact payment that consumes its backing mana unit, never reattached to a later
// unit of the same color.
type ManaRiderInstance struct {
	// Unit is the exact mana unit (color and snow provenance) that carries the
	// rider. Tracking the full unit, not just its color, keeps rider provenance
	// distinct from same-color mana of a different snow provenance.
	Unit mana.Unit
	// Controller is the player whose pool holds the tagged mana and who controls
	// the rider if it fires.
	Controller PlayerID
	// SourceID is the CardInstance ID of the permanent whose mana ability
	// produced the tagged mana; it is the rider's source card identity.
	SourceID id.ID
	// SourceObjectID is the producing permanent's ObjectID, used as the rider
	// stack object's source when it fires (mirroring an ability put on the
	// stack from a battlefield permanent).
	SourceObjectID id.ID
	// Rider is the spend rider to evaluate when the tagged mana is spent.
	Rider ManaSpendRider
}

// Ability returns the rider's effect content as a triggered ability for putting
// on the stack when the rider fires.
func (r ManaSpendRider) Ability() TriggeredAbility {
	return TriggeredAbility{Content: r.Effect.Ability()}
}
