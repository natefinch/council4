package game

import (
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// PrimitiveKind identifies the variant of a Primitive.
type PrimitiveKind int

// PrimitiveKind values identify each supported primitive variant.
const (
	PrimitiveUnknown PrimitiveKind = iota
	PrimitiveDamage
	PrimitiveDraw
	PrimitiveDiscard
	PrimitiveDestroy
	PrimitiveAddMana
	PrimitiveAddCounter
	PrimitiveAddPlayerCounter
	PrimitiveMoveCounters
	PrimitiveApplyContinuous
	PrimitiveApplyRule
	PrimitiveModifyPT
	PrimitiveFight
	PrimitiveTap
	PrimitiveSearch
	PrimitiveReveal
	PrimitivePutOnBattlefield
	PrimitiveCreateToken
	PrimitiveShufflePermanentIntoLibrary
	PrimitiveStartEngines
	PrimitiveSetClassLevel
	PrimitiveMonstrosity
	PrimitiveDiscoverCards
	PrimitivePay
	PrimitiveChoose
	PrimitiveGainLife
	PrimitiveLoseLife
	PrimitiveExile
	PrimitiveBounce
	PrimitiveSacrifice
	PrimitiveUntap
	PrimitiveCounterObject
	PrimitiveMill
	PrimitiveScry
	PrimitiveSurveil
	PrimitiveInvestigate
	PrimitiveProliferate
	PrimitiveGoad
	PrimitiveRemoveCounter
	PrimitiveTransform
	PrimitivePhaseOut
	PrimitiveRegenerate
	PrimitiveSkipStep
	PrimitiveCreateEmblem
	PrimitiveCreateDelayedTrigger
	PrimitiveCreateReplacement
	PrimitivePreventDamage
	PrimitiveMoveCard
	PrimitiveGrantCastPermission
)

// primitiveKindCount is the number of supported primitive kinds.
const primitiveKindCount = int(PrimitiveGrantCastPermission) + 1

// PrimitiveKindCount exposes primitiveKindCount to packages that need fixed-size tables.
const PrimitiveKindCount = primitiveKindCount

// Primitive is a sealed data-only interface for a single effect building block.
// Only types in this package may implement it.
type Primitive interface {
	Kind() PrimitiveKind
	isPrimitive()
	instructionRefs() primitiveRefs
	validatePrimitive([]TargetSpec, bool) error
}

// primitiveRefs describes what keys a Primitive consumes and publishes
// (distinct from the Instruction envelope's PublishResult).
type primitiveRefs struct {
	consumesResults []ResultKey
	consumesChoices []ChoiceKey
	consumesLinked  []LinkedKey
	publishesChoice ChoiceKey
	publishesLinked LinkedKey
}

type damageRecipientKind int

const (
	damageRecipientObject damageRecipientKind = iota
	damageRecipientPlayer
	damageRecipientAnyTarget
	damageRecipientGroup
	damageRecipientPlayerGroup
)

// DamageRecipient is a typed union identifying who receives damage.
// Use ObjectDamageRecipient, PlayerDamageRecipient, AnyTargetDamageRecipient,
// GroupDamageRecipient, or PlayerGroupDamageRecipient to construct.
type DamageRecipient struct {
	set         bool
	kind        damageRecipientKind
	object      ObjectReference
	player      PlayerReference
	group       *GroupReference
	playerGroup PlayerGroupReference
}

// ObjectDamageRecipient creates a recipient for a single permanent.
func ObjectDamageRecipient(object ObjectReference) DamageRecipient {
	return DamageRecipient{set: true, kind: damageRecipientObject, object: object}
}

// PlayerDamageRecipient creates a recipient for a single player.
func PlayerDamageRecipient(player PlayerReference) DamageRecipient {
	return DamageRecipient{set: true, kind: damageRecipientPlayer, player: player}
}

// AnyTargetDamageRecipient creates a recipient for a target slot that may name a
// player or permanent.
func AnyTargetDamageRecipient(targetIndex int) DamageRecipient {
	return DamageRecipient{
		set:    true,
		kind:   damageRecipientAnyTarget,
		object: TargetPermanentReference(targetIndex),
		player: TargetPlayerReference(targetIndex),
	}
}

// GroupDamageRecipient creates a recipient for a group of permanents.
func GroupDamageRecipient(group GroupReference) DamageRecipient {
	g := group
	return DamageRecipient{set: true, kind: damageRecipientGroup, group: &g}
}

// PlayerGroupDamageRecipient creates a recipient for a group of players.
func PlayerGroupDamageRecipient(group PlayerGroupReference) DamageRecipient {
	return DamageRecipient{set: true, kind: damageRecipientPlayerGroup, playerGroup: group}
}

// Valid reports whether the recipient identifies a supported target set.
func (r DamageRecipient) Valid() bool {
	if !r.set {
		return false
	}
	switch r.kind {
	case damageRecipientObject:
		return r.object.Kind() != ObjectReferenceNone && len(r.object.Validate()) == 0
	case damageRecipientPlayer:
		return r.player.Kind() != PlayerReferenceNone && len(r.player.Validate()) == 0
	case damageRecipientAnyTarget:
		return r.object.Kind() == ObjectReferenceTargetPermanent &&
			r.player.Kind() == PlayerReferenceTargetPlayer &&
			len(r.object.Validate()) == 0 &&
			len(r.player.Validate()) == 0
	case damageRecipientGroup:
		return r.group != nil && r.group.Valid()
	case damageRecipientPlayerGroup:
		return len(r.playerGroup.Validate()) == 0
	default:
		return false
	}
}

// ObjectReference returns the permanent reference when this recipient addresses one permanent.
func (r DamageRecipient) ObjectReference() (ObjectReference, bool) {
	if !r.Valid() || r.kind != damageRecipientObject {
		return ObjectReference{}, false
	}
	return r.object, true
}

// PlayerReference returns the player reference when this recipient addresses one player.
func (r DamageRecipient) PlayerReference() (PlayerReference, bool) {
	if !r.Valid() || r.kind != damageRecipientPlayer {
		return PlayerReference{}, false
	}
	return r.player, true
}

// GroupReference returns the group reference when this recipient addresses a permanent group.
func (r DamageRecipient) GroupReference() (GroupReference, bool) {
	if !r.Valid() || r.kind != damageRecipientGroup {
		return GroupReference{}, false
	}
	return *r.group, true
}

// PlayerGroupReference returns the player-group reference when this recipient addresses a player group.
func (r DamageRecipient) PlayerGroupReference() (PlayerGroupReference, bool) {
	if !r.Valid() || r.kind != damageRecipientPlayerGroup {
		return PlayerGroupReference{}, false
	}
	return r.playerGroup, true
}

// AnyTargetObjectReference returns the permanent reference when this recipient addresses any target.
func (r DamageRecipient) AnyTargetObjectReference() (ObjectReference, bool) {
	if !r.Valid() || r.kind != damageRecipientAnyTarget {
		return ObjectReference{}, false
	}
	return r.object, true
}

// AnyTargetPlayerReference returns the player reference when this recipient addresses any target.
func (r DamageRecipient) AnyTargetPlayerReference() (PlayerReference, bool) {
	if !r.Valid() || r.kind != damageRecipientAnyTarget {
		return PlayerReference{}, false
	}
	return r.player, true
}

type tokenSourceKind int

const (
	tokenSourceDef tokenSourceKind = iota
	tokenSourceCopy
)

// TokenSource is a mutually-exclusive union for the token definition.
// Use TokenDef or TokenCopyOf to construct.
type TokenSource struct {
	set  bool
	kind tokenSourceKind
	def  *CardDef
	copy TokenCopySpec
}

// TokenDef creates a TokenSource using an explicit CardDef.
func TokenDef(def *CardDef) TokenSource {
	return TokenSource{set: true, kind: tokenSourceDef, def: def}
}

// TokenCopyOf creates a TokenSource using a TokenCopySpec (copy-of-something).
func TokenCopyOf(spec TokenCopySpec) TokenSource {
	return TokenSource{set: true, kind: tokenSourceCopy, copy: spec}
}

// Valid reports whether the source identifies a concrete token definition.
func (s TokenSource) Valid() bool {
	if !s.set {
		return false
	}
	switch s.kind {
	case tokenSourceDef:
		return s.def != nil
	case tokenSourceCopy:
		return s.copy.Source != TokenCopySourceNone
	default:
		return false
	}
}

// TokenDefRef returns the token CardDef when this source uses an explicit definition.
func (s TokenSource) TokenDefRef() (*CardDef, bool) {
	if !s.Valid() || s.kind != tokenSourceDef {
		return nil, false
	}
	return s.def, true
}

// TokenCopy returns the TokenCopySpec when this source copies another object/card.
func (s TokenSource) TokenCopy() (TokenCopySpec, bool) {
	if !s.Valid() || s.kind != tokenSourceCopy {
		return TokenCopySpec{}, false
	}
	return s.copy, true
}

type battlefieldSourceKind int

const (
	battlefieldSourceCard battlefieldSourceKind = iota
	battlefieldSourceLinked
)

// BattlefieldSource identifies what card or object to put on the battlefield.
// Use CardBattlefieldSource or LinkedBattlefieldSource to construct.
type BattlefieldSource struct {
	set    bool
	kind   battlefieldSourceKind
	card   CardReference
	linked LinkedKey
}

// CardBattlefieldSource creates a source referencing a specific card.
func CardBattlefieldSource(ref CardReference) BattlefieldSource {
	return BattlefieldSource{set: true, kind: battlefieldSourceCard, card: ref}
}

// LinkedBattlefieldSource creates a source referencing an object linked by key.
func LinkedBattlefieldSource(key LinkedKey) BattlefieldSource {
	return BattlefieldSource{set: true, kind: battlefieldSourceLinked, linked: key}
}

// Valid reports whether the source identifies a concrete card or linked object.
func (s BattlefieldSource) Valid() bool {
	if !s.set {
		return false
	}
	switch s.kind {
	case battlefieldSourceCard:
		return s.card.Kind != CardReferenceNone
	case battlefieldSourceLinked:
		return s.linked != ""
	default:
		return false
	}
}

// CardRef returns the direct card reference when this source names a specific card.
func (s BattlefieldSource) CardRef() (CardReference, bool) {
	if !s.Valid() || s.kind != battlefieldSourceCard {
		return CardReference{}, false
	}
	return s.card, true
}

// LinkedKey returns the linked-object key when this source uses linked objects.
func (s BattlefieldSource) LinkedKey() (LinkedKey, bool) {
	if !s.Valid() || s.kind != battlefieldSourceLinked {
		return "", false
	}
	return s.linked, true
}

// sourceLinkedKey returns the LinkedKey if this is a linked source; otherwise empty.
func (s BattlefieldSource) sourceLinkedKey() LinkedKey {
	key, ok := s.LinkedKey()
	if !ok {
		return ""
	}
	return key
}

// Damage deals an amount of damage to a target.
type Damage struct {
	Amount           Quantity
	Recipient        DamageRecipient
	DamageSource     opt.V[ObjectReference]
	ResultAmountKind EffectResultAmountKind
}

// Draw draws cards for a referenced player.
type Draw struct {
	Amount Quantity
	Player PlayerReference
}

// Discard causes a referenced player to discard cards.
type Discard struct {
	Amount Quantity
	Player PlayerReference
}

// Destroy destroys one referenced permanent or every permanent in a referenced group.
type Destroy struct {
	Object ObjectReference
	Group  GroupReference
}

// AddMana adds mana to the controller's pool.
type AddMana struct {
	Amount Quantity
	// ManaColor is the color of mana produced.
	ManaColor mana.Color
	// ChoiceFrom links a prior Choose{Choice: ResolutionChoiceMana} result
	// to determine the mana color dynamically.
	ChoiceFrom ChoiceKey
}

// AddCounter places counters on a referenced permanent.
type AddCounter struct {
	Amount      Quantity
	Object      ObjectReference
	CounterKind counter.Kind
}

// AddPlayerCounter places counters on a referenced player.
type AddPlayerCounter struct {
	Amount      Quantity
	Player      PlayerReference
	CounterKind counter.Kind
}

// MoveCounters moves counters from a source to a target permanent.
type MoveCounters struct {
	Amount      Quantity
	Object      ObjectReference
	CounterKind counter.Kind
	Source      CounterSourceSpec
}

// ApplyContinuous applies continuous effects to a target (or globally).
type ApplyContinuous struct {
	Object            opt.V[ObjectReference]
	ContinuousEffects []ContinuousEffect
	Duration          EffectDuration
}

// ApplyRule creates rule effects for a target (or globally).
type ApplyRule struct {
	Object      opt.V[ObjectReference]
	RuleEffects []RuleEffect
	Duration    EffectDuration
}

// ModifyPT modifies a permanent's power and/or toughness.
type ModifyPT struct {
	Object         ObjectReference
	PowerDelta     Quantity
	ToughnessDelta Quantity
	Duration       EffectDuration
}

// Fight makes two permanents fight each other.
type Fight struct {
	Object        ObjectReference
	RelatedObject ObjectReference
}

// Tap taps the referenced permanent.
type Tap struct {
	Object ObjectReference
}

// Search searches a player's library for cards matching spec.
type Search struct {
	Player PlayerReference
	Spec   SearchSpec
	Amount Quantity
}

// Reveal reveals cards from a player's zone and optionally links them.
type Reveal struct {
	Amount        Quantity
	Player        PlayerReference
	Recipient     opt.V[PlayerReference]
	PublishLinked LinkedKey
}

// PutOnBattlefield puts a card or linked object onto the battlefield.
type PutOnBattlefield struct {
	Source            BattlefieldSource
	Recipient         opt.V[PlayerReference]
	ContinuousEffects []ContinuousEffect
	EntryTapped       bool
	EntryCounters     []CounterPlacement
}

// CreateToken creates one or more tokens.
type CreateToken struct {
	Amount    Quantity
	Source    TokenSource
	Recipient opt.V[PlayerReference]
}

// ShufflePermanentIntoLibrary shuffles the referenced permanent into its owner's library.
type ShufflePermanentIntoLibrary struct {
	Object ObjectReference
}

// StartEngines starts engine effects for a player.
type StartEngines struct {
	Player PlayerReference
}

// SetClassLevel sets the class level of a referenced Class permanent.
type SetClassLevel struct {
	Object ObjectReference
	Amount Quantity
}

// Monstrosity makes a referenced creature monstrous.
type Monstrosity struct {
	Object ObjectReference
	Amount Quantity
}

// DiscoverCards performs a discover for N.
type DiscoverCards struct {
	Amount Quantity
}

// Pay prompts the controller to pay an optional cost during resolution.
// The instruction's Optional field controls whether declining is allowed.
// Results are published via the Instruction.PublishResult for downstream ResultGate checks.
type Pay struct {
	Payment ResolutionPayment
	Prompt  string
}

// Choose makes a resolution-time choice and publishes it via PublishChoice.
type Choose struct {
	Choice        ResolutionChoice
	PublishChoice ChoiceKey
}

// GainLife causes a referenced player to gain life.
type GainLife struct {
	Amount Quantity
	Player PlayerReference
}

// LoseLife causes a referenced player to lose life.
type LoseLife struct {
	Amount Quantity
	Player PlayerReference
}

// Exile exiles one referenced permanent or every permanent in a referenced group.
// ExileLinkedKey remembers the exiled object for later "exile it, then return it" patterns.
type Exile struct {
	Object         ObjectReference
	Group          GroupReference
	ExileLinkedKey LinkedKey
}

// Bounce returns one referenced permanent or every permanent in a referenced group to hand.
type Bounce struct {
	Object ObjectReference
	Group  GroupReference
}

// MoveCard moves a referenced card between two non-battlefield zones.
type MoveCard struct {
	Card              CardReference
	FromZone          zone.Type
	Destination       zone.Type
	DestinationBottom bool
}

// GrantCastPermission allows a referenced card to be cast from a specific zone
// using a specific face for a bounded duration.
type GrantCastPermission struct {
	Card     CardReference
	FromZone zone.Type
	Face     FaceIndex
	Duration EffectDuration
}

// Sacrifice sacrifices the referenced permanent. When no object is set, the
// controller's first permanent is used.
type Sacrifice struct {
	Object ObjectReference
}

// Untap untaps one referenced permanent or every permanent in a referenced group.
type Untap struct {
	Object ObjectReference
	Group  GroupReference
}

// CounterObject counters a referenced spell or ability on the stack.
type CounterObject struct {
	Object ObjectReference
}

// Mill puts cards from the top of a referenced player's library into their graveyard.
type Mill struct {
	Amount Quantity
	Player PlayerReference
}

// Scry looks at and reorders the top cards of a referenced player's library.
type Scry struct {
	Amount Quantity
	Player PlayerReference
}

// Surveil looks at the top cards of a referenced player's library, putting any into the
// graveyard.
type Surveil struct {
	Amount Quantity
	Player PlayerReference
}

// Investigate creates Clue tokens for the recipient (controller by default).
type Investigate struct {
	Amount    Quantity
	Recipient opt.V[PlayerReference]
}

// Proliferate lets the controller add a counter of an existing kind to each
// chosen permanent or player.
type Proliferate struct{}

// Goad goads the referenced creature.
type Goad struct {
	Object ObjectReference
}

// RemoveCounter removes counters from one referenced permanent or every permanent in a referenced group.
type RemoveCounter struct {
	Amount      Quantity
	Object      ObjectReference
	Group       GroupReference
	CounterKind counter.Kind
}

// Transform transforms the referenced permanent.
type Transform struct {
	Object ObjectReference
}

// PhaseOut phases out the referenced permanent.
type PhaseOut struct {
	Object ObjectReference
}

// Regenerate sets up a regeneration shield on the referenced permanent.
type Regenerate struct {
	Object ObjectReference
}

// SkipStep schedules a referenced player to skip a step.
type SkipStep struct {
	Player PlayerReference
	Step   Step
}

// CreateEmblem creates an emblem owned by the controller with the given abilities.
type CreateEmblem struct {
	EmblemAbilities []Ability
}

// CreateDelayedTrigger schedules a delayed triggered ability.
type CreateDelayedTrigger struct {
	Trigger DelayedTriggerDef
}

// CreateReplacement creates a replacement effect that applies to a future event.
type CreateReplacement struct {
	Replacement *ReplacementEffect
	Duration    EffectDuration
}

// PreventDamage creates a damage-prevention shield for exactly one referenced
// player or permanent.
type PreventDamage struct {
	Amount Quantity
	Object ObjectReference
	Player PlayerReference
}

// Kind implements Primitive for Damage.
func (Damage) Kind() PrimitiveKind { return PrimitiveDamage }

// Kind implements Primitive for Draw.
func (Draw) Kind() PrimitiveKind { return PrimitiveDraw }

// Kind implements Primitive for Discard.
func (Discard) Kind() PrimitiveKind { return PrimitiveDiscard }

// Kind implements Primitive for Destroy.
func (Destroy) Kind() PrimitiveKind { return PrimitiveDestroy }

// Kind implements Primitive for AddMana.
func (AddMana) Kind() PrimitiveKind { return PrimitiveAddMana }

// Kind implements Primitive for AddCounter.
func (AddCounter) Kind() PrimitiveKind { return PrimitiveAddCounter }

// Kind implements Primitive for AddPlayerCounter.
func (AddPlayerCounter) Kind() PrimitiveKind { return PrimitiveAddPlayerCounter }

// Kind implements Primitive for MoveCounters.
func (MoveCounters) Kind() PrimitiveKind { return PrimitiveMoveCounters }

// Kind implements Primitive for ApplyContinuous.
func (ApplyContinuous) Kind() PrimitiveKind { return PrimitiveApplyContinuous }

// Kind implements Primitive for ApplyRule.
func (ApplyRule) Kind() PrimitiveKind { return PrimitiveApplyRule }

// Kind implements Primitive for ModifyPT.
func (ModifyPT) Kind() PrimitiveKind { return PrimitiveModifyPT }

// Kind implements Primitive for Fight.
func (Fight) Kind() PrimitiveKind { return PrimitiveFight }

// Kind implements Primitive for Tap.
func (Tap) Kind() PrimitiveKind { return PrimitiveTap }

// Kind implements Primitive for Search.
func (Search) Kind() PrimitiveKind { return PrimitiveSearch }

// Kind implements Primitive for Reveal.
func (Reveal) Kind() PrimitiveKind { return PrimitiveReveal }

// Kind implements Primitive for PutOnBattlefield.
func (PutOnBattlefield) Kind() PrimitiveKind { return PrimitivePutOnBattlefield }

// Kind implements Primitive for CreateToken.
func (CreateToken) Kind() PrimitiveKind { return PrimitiveCreateToken }

// Kind implements Primitive for ShufflePermanentIntoLibrary.
func (ShufflePermanentIntoLibrary) Kind() PrimitiveKind { return PrimitiveShufflePermanentIntoLibrary }

// Kind implements Primitive for StartEngines.
func (StartEngines) Kind() PrimitiveKind { return PrimitiveStartEngines }

// Kind implements Primitive for SetClassLevel.
func (SetClassLevel) Kind() PrimitiveKind { return PrimitiveSetClassLevel }

// Kind implements Primitive for Monstrosity.
func (Monstrosity) Kind() PrimitiveKind { return PrimitiveMonstrosity }

// Kind implements Primitive for DiscoverCards.
func (DiscoverCards) Kind() PrimitiveKind { return PrimitiveDiscoverCards }

// Kind implements Primitive for Pay.
func (Pay) Kind() PrimitiveKind { return PrimitivePay }

// Kind implements Primitive for Choose.
func (Choose) Kind() PrimitiveKind { return PrimitiveChoose }

// Kind implements Primitive for GainLife.
func (GainLife) Kind() PrimitiveKind { return PrimitiveGainLife }

// Kind implements Primitive for LoseLife.
func (LoseLife) Kind() PrimitiveKind { return PrimitiveLoseLife }

// Kind implements Primitive for Exile.
func (Exile) Kind() PrimitiveKind { return PrimitiveExile }

// Kind implements Primitive for Bounce.
func (Bounce) Kind() PrimitiveKind { return PrimitiveBounce }

// Kind implements Primitive for Sacrifice.
func (Sacrifice) Kind() PrimitiveKind { return PrimitiveSacrifice }

// Kind implements Primitive for Untap.
func (Untap) Kind() PrimitiveKind { return PrimitiveUntap }

// Kind implements Primitive for CounterObject.
func (CounterObject) Kind() PrimitiveKind { return PrimitiveCounterObject }

// Kind implements Primitive for Mill.
func (Mill) Kind() PrimitiveKind { return PrimitiveMill }

// Kind implements Primitive for Scry.
func (Scry) Kind() PrimitiveKind { return PrimitiveScry }

// Kind implements Primitive for Surveil.
func (Surveil) Kind() PrimitiveKind { return PrimitiveSurveil }

// Kind implements Primitive for Investigate.
func (Investigate) Kind() PrimitiveKind { return PrimitiveInvestigate }

// Kind implements Primitive for Proliferate.
func (Proliferate) Kind() PrimitiveKind { return PrimitiveProliferate }

// Kind implements Primitive for Goad.
func (Goad) Kind() PrimitiveKind { return PrimitiveGoad }

// Kind implements Primitive for RemoveCounter.
func (RemoveCounter) Kind() PrimitiveKind { return PrimitiveRemoveCounter }

// Kind implements Primitive for Transform.
func (Transform) Kind() PrimitiveKind { return PrimitiveTransform }

// Kind implements Primitive for PhaseOut.
func (PhaseOut) Kind() PrimitiveKind { return PrimitivePhaseOut }

// Kind implements Primitive for Regenerate.
func (Regenerate) Kind() PrimitiveKind { return PrimitiveRegenerate }

// Kind implements Primitive for SkipStep.
func (SkipStep) Kind() PrimitiveKind { return PrimitiveSkipStep }

// Kind implements Primitive for CreateEmblem.
func (CreateEmblem) Kind() PrimitiveKind { return PrimitiveCreateEmblem }

// Kind implements Primitive for CreateDelayedTrigger.
func (CreateDelayedTrigger) Kind() PrimitiveKind { return PrimitiveCreateDelayedTrigger }

// Kind implements Primitive for CreateReplacement.
func (CreateReplacement) Kind() PrimitiveKind { return PrimitiveCreateReplacement }

// Kind implements Primitive for PreventDamage.
func (PreventDamage) Kind() PrimitiveKind { return PrimitivePreventDamage }

// Kind implements Primitive for MoveCard.
func (MoveCard) Kind() PrimitiveKind { return PrimitiveMoveCard }

// Kind implements Primitive for GrantCastPermission.
func (GrantCastPermission) Kind() PrimitiveKind { return PrimitiveGrantCastPermission }

func (Damage) isPrimitive()                      {}
func (Draw) isPrimitive()                        {}
func (Discard) isPrimitive()                     {}
func (Destroy) isPrimitive()                     {}
func (AddMana) isPrimitive()                     {}
func (AddCounter) isPrimitive()                  {}
func (AddPlayerCounter) isPrimitive()            {}
func (MoveCounters) isPrimitive()                {}
func (ApplyContinuous) isPrimitive()             {}
func (ApplyRule) isPrimitive()                   {}
func (ModifyPT) isPrimitive()                    {}
func (Fight) isPrimitive()                       {}
func (Tap) isPrimitive()                         {}
func (Search) isPrimitive()                      {}
func (Reveal) isPrimitive()                      {}
func (PutOnBattlefield) isPrimitive()            {}
func (CreateToken) isPrimitive()                 {}
func (ShufflePermanentIntoLibrary) isPrimitive() {}
func (StartEngines) isPrimitive()                {}
func (SetClassLevel) isPrimitive()               {}
func (Monstrosity) isPrimitive()                 {}
func (DiscoverCards) isPrimitive()               {}
func (Pay) isPrimitive()                         {}
func (Choose) isPrimitive()                      {}
func (GainLife) isPrimitive()                    {}
func (LoseLife) isPrimitive()                    {}
func (Exile) isPrimitive()                       {}
func (Bounce) isPrimitive()                      {}
func (Sacrifice) isPrimitive()                   {}
func (Untap) isPrimitive()                       {}
func (CounterObject) isPrimitive()               {}
func (Mill) isPrimitive()                        {}
func (Scry) isPrimitive()                        {}
func (Surveil) isPrimitive()                     {}
func (Investigate) isPrimitive()                 {}
func (Proliferate) isPrimitive()                 {}
func (Goad) isPrimitive()                        {}
func (RemoveCounter) isPrimitive()               {}
func (Transform) isPrimitive()                   {}
func (PhaseOut) isPrimitive()                    {}
func (Regenerate) isPrimitive()                  {}
func (SkipStep) isPrimitive()                    {}
func (CreateEmblem) isPrimitive()                {}
func (CreateDelayedTrigger) isPrimitive()        {}
func (CreateReplacement) isPrimitive()           {}
func (PreventDamage) isPrimitive()               {}
func (MoveCard) isPrimitive()                    {}
func (GrantCastPermission) isPrimitive()         {}

func (p Damage) instructionRefs() primitiveRefs     { return quantityRefs(p.Amount) }
func (p Draw) instructionRefs() primitiveRefs       { return quantityRefs(p.Amount) }
func (p Discard) instructionRefs() primitiveRefs    { return quantityRefs(p.Amount) }
func (Destroy) instructionRefs() primitiveRefs      { return primitiveRefs{} }
func (p AddCounter) instructionRefs() primitiveRefs { return quantityRefs(p.Amount) }
func (p AddPlayerCounter) instructionRefs() primitiveRefs {
	return quantityRefs(p.Amount)
}
func (p MoveCounters) instructionRefs() primitiveRefs  { return quantityRefs(p.Amount) }
func (ApplyContinuous) instructionRefs() primitiveRefs { return primitiveRefs{} }
func (ApplyRule) instructionRefs() primitiveRefs       { return primitiveRefs{} }

func (p ModifyPT) instructionRefs() primitiveRefs {
	return mergePrimitiveRefs(quantityRefs(p.PowerDelta), quantityRefs(p.ToughnessDelta))
}
func (Fight) instructionRefs() primitiveRefs    { return primitiveRefs{} }
func (Tap) instructionRefs() primitiveRefs      { return primitiveRefs{} }
func (p Search) instructionRefs() primitiveRefs { return quantityRefs(p.Amount) }

func (p CreateToken) instructionRefs() primitiveRefs               { return quantityRefs(p.Amount) }
func (ShufflePermanentIntoLibrary) instructionRefs() primitiveRefs { return primitiveRefs{} }
func (StartEngines) instructionRefs() primitiveRefs                { return primitiveRefs{} }
func (p SetClassLevel) instructionRefs() primitiveRefs             { return quantityRefs(p.Amount) }
func (p Monstrosity) instructionRefs() primitiveRefs               { return quantityRefs(p.Amount) }
func (p DiscoverCards) instructionRefs() primitiveRefs             { return quantityRefs(p.Amount) }
func (Pay) instructionRefs() primitiveRefs                         { return primitiveRefs{} }

func (p AddMana) instructionRefs() primitiveRefs {
	refs := quantityRefs(p.Amount)
	if p.ChoiceFrom != "" {
		refs.consumesChoices = append(refs.consumesChoices, p.ChoiceFrom)
	}
	return refs
}

func (p Reveal) instructionRefs() primitiveRefs {
	refs := quantityRefs(p.Amount)
	refs.publishesLinked = p.PublishLinked
	return refs
}

func (p PutOnBattlefield) instructionRefs() primitiveRefs {
	if key := p.Source.sourceLinkedKey(); key != "" {
		return primitiveRefs{consumesLinked: []LinkedKey{key}}
	}
	return primitiveRefs{}
}

func (p Choose) instructionRefs() primitiveRefs {
	return primitiveRefs{publishesChoice: p.PublishChoice}
}

func (p GainLife) instructionRefs() primitiveRefs { return quantityRefs(p.Amount) }
func (p LoseLife) instructionRefs() primitiveRefs { return quantityRefs(p.Amount) }

func (p Exile) instructionRefs() primitiveRefs {
	return primitiveRefs{publishesLinked: p.ExileLinkedKey}
}
func (Bounce) instructionRefs() primitiveRefs        { return primitiveRefs{} }
func (Sacrifice) instructionRefs() primitiveRefs     { return primitiveRefs{} }
func (Untap) instructionRefs() primitiveRefs         { return primitiveRefs{} }
func (CounterObject) instructionRefs() primitiveRefs { return primitiveRefs{} }
func (p Mill) instructionRefs() primitiveRefs        { return quantityRefs(p.Amount) }
func (p Scry) instructionRefs() primitiveRefs        { return quantityRefs(p.Amount) }
func (p Surveil) instructionRefs() primitiveRefs     { return quantityRefs(p.Amount) }
func (p Investigate) instructionRefs() primitiveRefs { return quantityRefs(p.Amount) }
func (Proliferate) instructionRefs() primitiveRefs   { return primitiveRefs{} }
func (Goad) instructionRefs() primitiveRefs          { return primitiveRefs{} }

func (p RemoveCounter) instructionRefs() primitiveRefs      { return quantityRefs(p.Amount) }
func (Transform) instructionRefs() primitiveRefs            { return primitiveRefs{} }
func (PhaseOut) instructionRefs() primitiveRefs             { return primitiveRefs{} }
func (Regenerate) instructionRefs() primitiveRefs           { return primitiveRefs{} }
func (SkipStep) instructionRefs() primitiveRefs             { return primitiveRefs{} }
func (CreateEmblem) instructionRefs() primitiveRefs         { return primitiveRefs{} }
func (CreateDelayedTrigger) instructionRefs() primitiveRefs { return primitiveRefs{} }
func (CreateReplacement) instructionRefs() primitiveRefs    { return primitiveRefs{} }
func (p PreventDamage) instructionRefs() primitiveRefs      { return quantityRefs(p.Amount) }
func (p MoveCard) instructionRefs() primitiveRefs           { return cardReferenceRefs(p.Card) }
func (p GrantCastPermission) instructionRefs() primitiveRefs {
	return cardReferenceRefs(p.Card)
}

func cardReferenceRefs(reference CardReference) primitiveRefs {
	if reference.Kind != CardReferenceLinked || reference.LinkID == "" {
		return primitiveRefs{}
	}
	return primitiveRefs{consumesLinked: []LinkedKey{LinkedKey(reference.LinkID)}}
}

func quantityRefs(quantity Quantity) primitiveRefs {
	if !quantity.IsDynamic() {
		return primitiveRefs{}
	}
	dynamic := quantity.DynamicAmount().Val
	switch dynamic.Kind {
	case DynamicAmountPreviousEffectResult, DynamicAmountPreviousEffectExcessDamage:
		if dynamic.ResultKey != "" {
			return primitiveRefs{consumesResults: []ResultKey{dynamic.ResultKey}}
		}
	default:
	}
	return primitiveRefs{}
}

func mergePrimitiveRefs(left, right primitiveRefs) primitiveRefs {
	left.consumesResults = append(left.consumesResults, right.consumesResults...)
	left.consumesChoices = append(left.consumesChoices, right.consumesChoices...)
	left.consumesLinked = append(left.consumesLinked, right.consumesLinked...)
	return left
}
