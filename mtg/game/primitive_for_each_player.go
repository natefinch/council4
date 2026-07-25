package game

// DistributiveRemoval names the removal applied to each player's chosen
// permanent by ForEachPlayer. Its zero value is invalid so that a lowering that
// forgets to set it fails loudly rather than silently exiling.
type DistributiveRemoval int

// Distributive removal verbs.
const (
	// DistributiveRemovalUnknown is the invalid zero value.
	DistributiveRemovalUnknown DistributiveRemoval = iota
	// DistributiveRemovalExile moves each chosen permanent to exile.
	DistributiveRemovalExile
	// DistributiveRemovalDestroy destroys each chosen permanent.
	DistributiveRemovalDestroy
)

// ForEachPlayer walks each player in Scope in APNAP order and, for each, has
// Chooser select from that player's own candidate pool - the permanents that
// player controls matching Selection - and applies Removal to the choice. It
// models the distributive removal template
//
//	"For each player, <verb> up to one target permanent that player controls."
//
// across every combination of scope and verb: "For each player, destroy up to
// one target creature that player controls." (The Curse of Fenric), "For each
// player, exile up to one target creature that player controls." (Unexplained
// Absence), and "For each opponent, exile up to one target permanent that
// player controls with mana value 3 or greater." (King Solomon's Frogs).
//
// Each player's permanents form an independent pool, so the effect removes at
// most one permanent per player. Selection's ExcludeSource models an "other"
// qualifier. Every removed permanent is remembered under LinkedKey, keyed by the
// source, so a paired payoff (a linked return, a token per destroyed permanent,
// a draw per exiled permanent) can consume the set; LinkedKey must be set,
// because the removed permanents are otherwise unrecoverable.
//
// It replaced three primitives that differed only in scope and verb
// (ExileForEachPlayer, DestroyForEachPlayer, ExileForEachOpponent). Adding a new
// scope/verb combination is now a lowering change with no new runtime.
type ForEachPlayer struct {
	// Scope selects the players walked. It is the full reference type rather
	// than a two-valued enum so "each other player" and similar groups need no
	// further primitive.
	Scope PlayerGroupReference
	// Chooser selects who makes each choice. It is resolved once per member of
	// Scope with that member bound as the group-offer member, so a
	// GroupMember-relative Chooser gives each player their own choice.
	Chooser   PlayerReference
	Selection Selection
	Removal   DistributiveRemoval
	LinkedKey LinkedKey
	// ReplaceLink clears LinkedKey's existing set before the walk, so a
	// re-resolution does not inherit the previous resolution's permanents. It is
	// unset for the Saga chapters, whose links are consumed by a later chapter.
	ReplaceLink bool
	// Required makes each chooser select exactly one permanent when their pool is
	// nonempty. The zero value is the "up to one" wording.
	Required bool
	// Extremum narrows each player's independent pool to the permanents tied for
	// the requested greatest characteristic before the choice is made.
	Extremum PermanentChoiceExtremum
	// Simultaneous collects every choice in APNAP order before moving all chosen
	// permanents in one zone-change batch.
	Simultaneous bool
}

// Kind implements Primitive for ForEachPlayer.
func (ForEachPlayer) Kind() PrimitiveKind { return PrimitiveForEachPlayer }
