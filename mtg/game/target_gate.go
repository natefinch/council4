package game

// TargetGate names the cast-time branch a target spec depends on. A spell may
// have a target only if a condition is met, and that target is chosen only if
// the condition is met (CR 601.2c). A gated spec participates in target
// announcement, validation, and counting only when its gate is active for the
// branch the spell is being cast on; otherwise it contributes no targets and
// reserves no announced slot. The gate is keyed to cast-time state (the gift
// promise, the kicker payment), not to Oracle wording, so the compiler assigns
// it structurally from how a spec's referencing instructions are already gated.
//
// The default TargetGateAlways keeps ordinary, unconditional targets required on
// every cast, so existing specs that omit the field are unaffected. New
// cast-branch conditions extend the enum without adding a boolean per mechanic to
// TargetSpec or to the targeting, validation, and copy paths.
type TargetGate uint8

// Target gate values identify the cast branch under which a target spec is
// active.
const (
	// TargetGateAlways marks an ordinary target required on every cast branch.
	TargetGateAlways TargetGate = iota
	// TargetGateGiftPromised marks a target required only when the spell's Gift
	// keyword action promised a gift as it was cast (CR 702.171), such as an
	// additional or alternative "if the gift was promised" target.
	TargetGateGiftPromised
	// TargetGateGiftNotPromised marks a target required only when the gift was
	// not promised, such as the base "instead" target a promised cast replaces.
	TargetGateGiftNotPromised
	// TargetGateSpellKicked marks a target required only when the spell's kicker
	// cost was paid ("if this spell was kicked, ..."; CR 702.32).
	TargetGateSpellKicked
	// TargetGateSpellNotKicked marks a target required only when the kicker cost
	// was not paid, such as the base "instead" target a kicked cast replaces.
	TargetGateSpellNotKicked
	// TargetGateSpellBargained marks a target required only when the spell's
	// Bargain additional cost was paid ("if this spell was bargained, ...";
	// CR 702.166c).
	TargetGateSpellBargained
	// TargetGateSpellNotBargained marks a target required only when the Bargain
	// cost was not paid, such as a base target a bargained cast replaces.
	TargetGateSpellNotBargained
	// TargetGateBestowed marks a target required only when the spell was cast for
	// its Bestow alternative cost (CR 702.103): a bestowed cast is an Aura spell
	// that must choose a legal creature target, while an ordinary creature cast
	// of the same card requires no target.
	TargetGateBestowed
	// TargetGateSpellOffspring marks a target required only when the spell's
	// Offspring additional mana cost was paid ("if the offspring cost was paid,
	// ..."; CR 702.171b).
	TargetGateSpellOffspring
	// TargetGateSpellNotOffspring marks a target required only when the Offspring
	// cost was not paid.
	TargetGateSpellNotOffspring
)

// CastBranch captures the cast-time choices that decide which gated target specs
// are active: whether the gift was promised and whether the spell was kicked. It
// is derived from the cast action while announcing, and from the resolving stack
// object's captured state when copying or retargeting. New cast-branch mechanics
// add a field here rather than a boolean to every targeting entry point.
type CastBranch struct {
	// GiftPromised reports that the Gift keyword action promised a gift as the
	// spell was cast.
	GiftPromised bool
	// Kicked reports that the spell's kicker cost was paid at least once.
	Kicked bool
	// Bargained reports that the spell's Bargain additional cost was paid as it
	// was cast (CR 702.166b).
	Bargained bool
	// Bestowed reports that the spell was cast for its Bestow alternative cost
	// (CR 702.103), making it a bestowed Aura spell that requires an enchant
	// target.
	Bestowed bool
	// Offspring reports that the spell's Offspring additional mana cost was paid
	// as it was cast (CR 702.171b).
	Offspring bool
}

// ActiveIn reports whether a spec carrying this gate participates in targeting on
// the given cast branch.
func (gate TargetGate) ActiveIn(branch CastBranch) bool {
	switch gate {
	case TargetGateGiftPromised:
		return branch.GiftPromised
	case TargetGateGiftNotPromised:
		return !branch.GiftPromised
	case TargetGateSpellKicked:
		return branch.Kicked
	case TargetGateSpellNotKicked:
		return !branch.Kicked
	case TargetGateSpellBargained:
		return branch.Bargained
	case TargetGateSpellNotBargained:
		return !branch.Bargained
	case TargetGateBestowed:
		return branch.Bestowed
	case TargetGateSpellOffspring:
		return branch.Offspring
	case TargetGateSpellNotOffspring:
		return !branch.Offspring
	default:
		return true
	}
}

// Valid reports whether the gate is a known enum value.
func (gate TargetGate) Valid() bool {
	return gate <= TargetGateSpellNotOffspring
}
