package game

import (
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/mana"
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
)

// primitiveKindCount is the number of supported primitive kinds.
const primitiveKindCount = int(PrimitiveChoose) + 1

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
	damageRecipientTarget damageRecipientKind = iota
	damageRecipientSelector
	damageRecipientPlayerSelector
)

// DamageRecipient is a typed union identifying who receives damage.
// Use TargetRecipient, SelectorRecipient, or PlayerSelectorRecipient to construct.
type DamageRecipient struct {
	set   bool
	kind  damageRecipientKind
	index int
	sel   EffectSelector
	psel  PlayerSelector
}

// TargetRecipient creates a recipient for a single chosen target at index.
// Use TargetIndexController (-1) for the controller, TargetIndexSourcePermanent (-2) for the source.
func TargetRecipient(index int) DamageRecipient {
	return DamageRecipient{set: true, kind: damageRecipientTarget, index: index}
}

// SelectorRecipient creates a recipient for a mass-damage selector covering permanents.
func SelectorRecipient(sel EffectSelector) DamageRecipient {
	return DamageRecipient{set: true, kind: damageRecipientSelector, sel: sel}
}

// PlayerSelectorRecipient creates a recipient for a mass-damage player selector.
func PlayerSelectorRecipient(psel PlayerSelector) DamageRecipient {
	return DamageRecipient{set: true, kind: damageRecipientPlayerSelector, psel: psel}
}

// Valid reports whether the recipient identifies a supported target set.
func (r DamageRecipient) Valid() bool {
	if !r.set {
		return false
	}
	switch r.kind {
	case damageRecipientTarget:
		return true
	case damageRecipientSelector:
		return r.sel != EffectSelectorNone
	case damageRecipientPlayerSelector:
		return r.psel != PlayerSelectorNone
	default:
		return false
	}
}

// TargetIndex returns the target index when this recipient addresses one target.
func (r DamageRecipient) TargetIndex() (int, bool) {
	if !r.Valid() || r.kind != damageRecipientTarget {
		return 0, false
	}
	return r.index, true
}

// Selector returns the permanent selector when this recipient is a mass selector.
func (r DamageRecipient) Selector() (EffectSelector, bool) {
	if !r.Valid() || r.kind != damageRecipientSelector {
		return EffectSelectorNone, false
	}
	return r.sel, true
}

// PlayerSelector returns the player selector when this recipient is a mass player selector.
func (r DamageRecipient) PlayerSelector() (PlayerSelector, bool) {
	if !r.Valid() || r.kind != damageRecipientPlayerSelector {
		return PlayerSelectorNone, false
	}
	return r.psel, true
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

// Draw draws cards for a player identified by TargetIndex (-1 = controller).
type Draw struct {
	Amount      Quantity
	TargetIndex int
}

// Discard causes a player to discard cards.
type Discard struct {
	Amount      Quantity
	TargetIndex int
}

// Destroy destroys the permanent identified by TargetIndex.
type Destroy struct {
	TargetIndex int
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

// AddCounter places counters on a permanent.
type AddCounter struct {
	Amount      Quantity
	TargetIndex int
	CounterKind counter.Kind
}

// MoveCounters moves counters from a source to a target permanent.
type MoveCounters struct {
	Amount      Quantity
	TargetIndex int
	CounterKind counter.Kind
	Source      CounterSourceSpec
}

// ApplyContinuous applies continuous effects to a target (or globally).
type ApplyContinuous struct {
	TargetIndex       int
	Object            opt.V[ObjectReference]
	ContinuousEffects []ContinuousEffect
	Duration          EffectDuration
}

// ApplyRule creates rule effects for a target (or globally).
type ApplyRule struct {
	TargetIndex int
	RuleEffects []RuleEffect
	Duration    EffectDuration
}

// ModifyPT modifies a permanent's power and/or toughness.
type ModifyPT struct {
	TargetIndex    int
	Object         opt.V[ObjectReference]
	PowerDelta     Quantity
	ToughnessDelta Quantity
	Duration       EffectDuration
}

// Fight makes two permanents fight each other.
type Fight struct {
	TargetIndex        int
	RelatedTargetIndex opt.V[int]
}

// Tap taps the permanent identified by TargetIndex.
type Tap struct {
	TargetIndex int
}

// Search searches a library for cards matching spec.
type Search struct {
	TargetIndex int
	Spec        SearchSpec
	Amount      Quantity
}

// Reveal reveals cards from a zone and optionally links them.
type Reveal struct {
	Amount        Quantity
	TargetIndex   int
	Recipient     opt.V[PlayerReference]
	PublishLinked LinkedKey
}

// PutOnBattlefield puts a card or linked object onto the battlefield.
type PutOnBattlefield struct {
	TargetIndex       int
	Source            BattlefieldSource
	Recipient         opt.V[PlayerReference]
	ContinuousEffects []ContinuousEffect
}

// CreateToken creates one or more tokens.
type CreateToken struct {
	Amount    Quantity
	Source    TokenSource
	Recipient opt.V[PlayerReference]
}

// ShufflePermanentIntoLibrary shuffles the target permanent into its owner's library.
type ShufflePermanentIntoLibrary struct {
	TargetIndex int
}

// StartEngines starts engine effects for a player.
type StartEngines struct {
	TargetIndex int
}

// SetClassLevel sets the class level of the source Class permanent.
type SetClassLevel struct {
	TargetIndex int
	Amount      Quantity
}

// Monstrosity makes the source creature monstrous.
type Monstrosity struct {
	TargetIndex int
	Amount      Quantity
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

func (Damage) isPrimitive()                      {}
func (Draw) isPrimitive()                        {}
func (Discard) isPrimitive()                     {}
func (Destroy) isPrimitive()                     {}
func (AddMana) isPrimitive()                     {}
func (AddCounter) isPrimitive()                  {}
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

func (p Damage) instructionRefs() primitiveRefs        { return quantityRefs(p.Amount) }
func (p Draw) instructionRefs() primitiveRefs          { return quantityRefs(p.Amount) }
func (p Discard) instructionRefs() primitiveRefs       { return quantityRefs(p.Amount) }
func (Destroy) instructionRefs() primitiveRefs         { return primitiveRefs{} }
func (p AddCounter) instructionRefs() primitiveRefs    { return quantityRefs(p.Amount) }
func (p MoveCounters) instructionRefs() primitiveRefs  { return quantityRefs(p.Amount) }
func (ApplyContinuous) instructionRefs() primitiveRefs { return primitiveRefs{} }
func (ApplyRule) instructionRefs() primitiveRefs       { return primitiveRefs{} }

//nolint:gocritic // Value receivers are required by the sealed Primitive interface.
func (p ModifyPT) instructionRefs() primitiveRefs {
	return mergePrimitiveRefs(quantityRefs(p.PowerDelta), quantityRefs(p.ToughnessDelta))
}
func (Fight) instructionRefs() primitiveRefs    { return primitiveRefs{} }
func (Tap) instructionRefs() primitiveRefs      { return primitiveRefs{} }
func (p Search) instructionRefs() primitiveRefs { return quantityRefs(p.Amount) }

//nolint:gocritic // Value receivers are required by the sealed Primitive interface.
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
