package game

import (
	"github.com/natefinch/council4/mtg/game/id"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
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
	// ManaSpendCastChosenCreatureType is "spent only to cast a creature spell of
	// the chosen type". The chosen subtype is captured from the producing
	// permanent when the mana is created.
	ManaSpendCastChosenCreatureType
	// ManaSpendCastLegendarySpell is "spent only to cast a legendary spell"
	// (Delighted Halfling). The restriction excludes nonlegendary spells from the
	// mana's reach; a qualifying legendary spell may additionally be made
	// uncounterable via SpellRuleEffect.
	ManaSpendCastLegendarySpell
	// ManaSpendCastOrActivateChosenCreatureType is "spent only to cast a creature
	// spell of the chosen type or activate an ability of a creature source of the
	// chosen type" (Secluded Courtyard). It extends ManaSpendCastChosenCreatureType
	// to also admit the tagged mana to the activated-ability costs of a creature
	// source of the chosen type. The chosen subtype is captured from the producing
	// permanent when the mana is created.
	ManaSpendCastOrActivateChosenCreatureType
	// ManaSpendCastCreatureSpell is "spent on a creature spell" (Arena of Glory,
	// Generator Servant). It is an unrestricted bonus rider: the tagged mana may
	// be spent on anything, but a creature spell paid for with it gains the
	// rider's SpellGainsKeywords until end of turn once it resolves.
	ManaSpendCastCreatureSpell
)

// ManaSpendRestrictionKind identifies whether a tagged mana unit may be spent
// only when its condition is satisfied.
type ManaSpendRestrictionKind int

const (
	// ManaSpendUnrestricted permits ordinary payments even when Condition is not
	// satisfied.
	ManaSpendUnrestricted ManaSpendRestrictionKind = iota
	// ManaSpendRestrictedToCondition permits only payments satisfying Condition.
	ManaSpendRestrictedToCondition
)

// ManaSpendRider is spend-linked semantics associated with a unit of produced
// mana. A qualifying spend may fire Effect as a triggered ability or apply
// SpellRuleEffect directly to the paid spell; Restriction may prohibit all
// nonqualifying payments.
type ManaSpendRider struct {
	// Condition is the closed, fully modeled spend condition that must hold for
	// the rider to fire.
	Condition ManaSpendConditionKind
	// Restriction makes the tagged mana unavailable to payments that do not
	// satisfy Condition.
	Restriction ManaSpendRestrictionKind
	// Effect is the rider's resolving content (for Path of Ancestry, "scry 1").
	Effect Mode
	// SpellRuleEffect is applied directly to a qualifying spell paid for with
	// this mana. RuleEffectCantBeCountered models Cavern of Souls.
	SpellRuleEffect RuleEffectKind
	// SpellGainsKeywords are the keyword abilities a qualifying creature spell
	// paid for with this mana gains until end of turn once it resolves into a
	// permanent (Arena of Glory, Generator Servant: "it gains haste until end of
	// turn"). The grant is a layer-ability continuous effect applied to the
	// resolved permanent, so it is empty for restriction-only and stack-rule
	// riders.
	SpellGainsKeywords []Keyword
	// ChosenSubtypeFrom names the source permanent's entry-time subtype choice
	// captured onto each produced mana unit.
	ChosenSubtypeFrom ChoiceKey
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
	// ChosenSubtype is the source permanent's captured entry-time subtype choice,
	// when the rider condition references one.
	ChosenSubtype types.Sub
	// Rider is the spend rider to evaluate when the tagged mana is spent.
	Rider ManaSpendRider
}

// Ability returns the rider's effect content as a triggered ability for putting
// on the stack when the rider fires.
func (r ManaSpendRider) Ability() TriggeredAbility {
	return TriggeredAbility{Content: r.Effect.Ability()}
}

// FiresOnSpend reports whether the rider has a triggered effect to put on the
// stack when its tagged mana is spent on a qualifying payment. A pure
// restriction rider (Unclaimed Territory, Secluded Courtyard) has none: its
// tagged mana is merely consumed, with no ability triggered.
func (r ManaSpendRider) FiresOnSpend() bool {
	return len(r.Effect.Sequence) > 0
}

// MatchesChosenCreatureType reports whether spell is a creature spell of the
// subtype captured on this mana unit.
func (r ManaRiderInstance) MatchesChosenCreatureType(spell *CardDef) bool {
	return spell != nil &&
		spell.HasType(types.Creature) &&
		types.KnownSubtypeForType(types.Creature, r.ChosenSubtype) &&
		spell.HasSubtype(r.ChosenSubtype)
}
