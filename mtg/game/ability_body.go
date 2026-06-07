package game

import (
	"fmt"

	"github.com/natefinch/council4/mtg/game/zone"

	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/opt"
)

// AbilityBody is a sealed data-only variant for how an ability functions.
type AbilityBody interface {
	isAbilityBody()
}

// AbilityContent is a sealed data-only variant for an ability's instructions.
type AbilityContent interface {
	isAbilityContent()
}

// PlainAbilityContent is a non-modal target/effect sequence.
type PlainAbilityContent struct {
	Targets  []TargetSpec
	Sequence []Instruction
}

// ModalAbilityContent is a mode-choice ability body.
type ModalAbilityContent struct {
	SharedTargets       []TargetSpec
	Modes               []Mode
	MinModes            int
	MaxModes            int
	AllowDuplicateModes bool
}

// SpellAbilityBody is an instruction on an instant or sorcery spell.
type SpellAbilityBody struct {
	Text             string
	Content          AbilityContent
	AdditionalCosts  []cost.Additional
	AlternativeCosts []cost.Alternative
}

// ActivatedAbilityBody is a non-mana, non-loyalty activated ability.
type ActivatedAbilityBody struct {
	Text                string
	ManaCost            opt.V[cost.Mana]
	AdditionalCosts     []cost.Additional
	AlternativeCosts    []cost.Alternative
	ZoneOfFunction      zone.Type
	Timing              TimingRestriction
	ActivationCondition opt.V[Condition]
	Content             AbilityContent
	// KeywordAbilities lists keyword abilities carried by this activation, e.g.
	// EquipKeyword for equip activations. Rules use it for keyword dispatch and
	// cost routing without inspecting Content.
	KeywordAbilities []KeywordAbility
}

// ManaAbilityBody is an activated mana ability.
type ManaAbilityBody struct {
	Text                string
	ManaCost            opt.V[cost.Mana]
	AdditionalCosts     []cost.Additional
	ZoneOfFunction      zone.Type
	Timing              TimingRestriction
	ActivationCondition opt.V[Condition]
	// Content is the mana output, supporting both plain sequences and modal
	// mana (ModalAbilityContent).
	Content AbilityContent
}

// LoyaltyAbilityBody is a planeswalker loyalty ability.
type LoyaltyAbilityBody struct {
	Text                string
	LoyaltyCost         int
	ActivationCondition opt.V[Condition]
	Content             AbilityContent
}

// TriggeredAbilityBody is an ability that triggers from a game event or state.
type TriggeredAbilityBody struct {
	Text               string
	Trigger            TriggerCondition
	Optional           bool
	MaxTriggersPerTurn int
	Content            AbilityContent
}

// StaticAbilityBody is a static ability that functions from a zone.
type StaticAbilityBody struct {
	Text              string
	Condition         opt.V[Condition]
	ZoneOfFunction    zone.Type
	KeywordAbilities  []KeywordAbility
	ContinuousEffects []ContinuousEffect
	RuleEffects       []RuleEffect
}

// ReplacementAbilityDef is a replacement/prevention ability on a printed face.
type ReplacementAbilityDef struct {
	Text        string
	Replacement ReplacementEffect
	UnlessPaid  opt.V[ResolutionPayment]
}

// EntersTappedReplacement creates a replacement ability for "enters tapped".
func EntersTappedReplacement(text string) ReplacementAbilityDef {
	replacement := etbReplacement(text)
	replacement.EntersTapped = true
	return ReplacementAbilityDef{Text: text, Replacement: replacement}
}

// EntersTappedIfReplacement creates a conditional "enters tapped" replacement.
func EntersTappedIfReplacement(text string, condition *Condition) ReplacementAbilityDef {
	replacement := etbReplacement(text)
	replacement.Condition = opt.Val(*condition)
	replacement.EntersTapped = true
	return ReplacementAbilityDef{Text: text, Replacement: replacement}
}

// EntersTappedUnlessPaidReplacement creates an ETB payment replacement. If the
// payment is not paid, the permanent enters tapped.
func EntersTappedUnlessPaidReplacement(text string, payment ResolutionPayment) ReplacementAbilityDef {
	replacement := etbReplacement(text)
	replacement.EntersTapped = true
	return ReplacementAbilityDef{
		Text:        text,
		Replacement: replacement,
		UnlessPaid:  opt.Val(payment),
	}
}

// EntersWithCountersReplacement creates an ETB counter-placement replacement.
func EntersWithCountersReplacement(text string, placements ...CounterPlacement) ReplacementAbilityDef {
	replacement := etbReplacement(text)
	replacement.EntersWithCounters = append([]CounterPlacement(nil), placements...)
	return ReplacementAbilityDef{Text: text, Replacement: replacement}
}

func etbReplacement(text string) ReplacementEffect {
	return ReplacementEffect{
		Description: text,
		MatchEvent:  EventPermanentEnteredBattlefield,
		MatchToZone: true,
		ToZone:      zone.Battlefield,
		Duration:    DurationPermanent,
	}
}

func (PlainAbilityContent) isAbilityContent() {}
func (ModalAbilityContent) isAbilityContent() {}

func (SpellAbilityBody) isAbilityBody()     {}
func (ActivatedAbilityBody) isAbilityBody() {}
func (ManaAbilityBody) isAbilityBody()      {}
func (LoyaltyAbilityBody) isAbilityBody()   {}
func (TriggeredAbilityBody) isAbilityBody() {}
func (StaticAbilityBody) isAbilityBody()    {}

// AbilityBodyKind returns the legacy AbilityKind represented by a body variant.
func AbilityBodyKind(body AbilityBody) AbilityKind {
	switch body.(type) {
	case SpellAbilityBody:
		return SpellAbility
	case ActivatedAbilityBody:
		return ActivatedAbility
	case ManaAbilityBody:
		return ActivatedAbility
	case LoyaltyAbilityBody:
		return ActivatedAbility
	case TriggeredAbilityBody:
		return TriggeredAbility
	case StaticAbilityBody:
		return StaticAbility
	case nil:
		panic("game: nil AbilityBody")
	default:
		panic(fmt.Sprintf("game: unsupported AbilityBody %T", body))
	}
}

// EffectiveKind returns the ability kind, preferring the sealed body when present.
func (ability *AbilityDef) EffectiveKind() AbilityKind {
	if ability.Body != nil {
		return AbilityBodyKind(ability.Body)
	}
	return ability.Kind
}

// IsSpell reports whether this ability is a spell ability.
func (ability *AbilityDef) IsSpell() bool {
	if ability.Body != nil {
		_, ok := ability.Body.(SpellAbilityBody)
		return ok
	}
	return ability.Kind == SpellAbility
}

// IsActivated reports whether this ability is a non-mana, non-loyalty activated ability.
func (ability *AbilityDef) IsActivated() bool {
	if ability.Body != nil {
		_, ok := ability.Body.(ActivatedAbilityBody)
		return ok
	}
	return ability.Kind == ActivatedAbility && !ability.IsManaAbility && !ability.IsLoyaltyAbility
}

// IsMana reports whether this ability is an activated mana ability.
func (ability *AbilityDef) IsMana() bool {
	if ability.Body != nil {
		_, ok := ability.Body.(ManaAbilityBody)
		return ok
	}
	return ability.Kind == ActivatedAbility && ability.IsManaAbility && !ability.IsLoyaltyAbility
}

// IsLoyalty reports whether this ability is a loyalty ability.
func (ability *AbilityDef) IsLoyalty() bool {
	if ability.Body != nil {
		_, ok := ability.Body.(LoyaltyAbilityBody)
		return ok
	}
	return ability.Kind == ActivatedAbility && ability.IsLoyaltyAbility
}

// IsTriggered reports whether this ability has a triggered ability body.
func (ability *AbilityDef) IsTriggered() bool {
	if ability.Body != nil {
		_, ok := ability.Body.(TriggeredAbilityBody)
		return ok
	}
	return ability.Kind == TriggeredAbility
}

// IsStatic reports whether this ability is a static ability.
func (ability *AbilityDef) IsStatic() bool {
	if ability.Body != nil {
		_, ok := ability.Body.(StaticAbilityBody)
		return ok
	}
	return ability.Kind == StaticAbility
}

// FunctionZone returns the zone where this ability functions.
func (ability *AbilityDef) FunctionZone() zone.Type {
	switch body := ability.Body.(type) {
	case StaticAbilityBody:
		return body.ZoneOfFunction
	case ActivatedAbilityBody:
		return body.ZoneOfFunction
	case ManaAbilityBody:
		return body.ZoneOfFunction
	}
	return ability.ZoneOfFunction
}

// TimingRestriction returns the timing restriction for this ability.
func (ability *AbilityDef) TimingRestriction() TimingRestriction {
	switch body := ability.Body.(type) {
	case ActivatedAbilityBody:
		return body.Timing
	case ManaAbilityBody:
		return body.Timing
	}
	return ability.Timing
}

// ActivationConditionValue returns the activation condition for this ability.
func (ability *AbilityDef) ActivationConditionValue() opt.V[Condition] {
	switch body := ability.Body.(type) {
	case ActivatedAbilityBody:
		return body.ActivationCondition
	case ManaAbilityBody:
		return body.ActivationCondition
	case LoyaltyAbilityBody:
		return body.ActivationCondition
	}
	return ability.ActivationCondition
}

// LoyaltyCostValue returns the loyalty cost for this ability.
func (ability *AbilityDef) LoyaltyCostValue() int {
	if body, ok := ability.Body.(LoyaltyAbilityBody); ok {
		return body.LoyaltyCost
	}
	return ability.LoyaltyCost
}

// WithBody returns a copy of this ability with both Body and flat compatibility
// fields fully populated. When Body is already set, flat fields are lowered from
// the body so rules consumers see consistent values. When Body is nil, it is
// synthesised from the legacy flat fields, then the flat fields are refreshed
// from that body.
func (ability *AbilityDef) WithBody() AbilityDef {
	normalized := *ability
	if normalized.Body != nil {
		lowerBodyToFlat(&normalized)
		return normalized
	}
	switch normalized.Kind {
	case SpellAbility:
		if body, ok := normalized.SpellBody(); ok {
			normalized.Body = body
		}
	case ActivatedAbility:
		if normalized.IsManaAbility {
			if body, ok := normalized.ManaBody(); ok {
				normalized.Body = body
			}
			return normalized
		}
		if normalized.IsLoyaltyAbility {
			if body, ok := normalized.LoyaltyBody(); ok {
				normalized.Body = body
			}
			return normalized
		}
		if body, ok := normalized.ActivatedBody(); ok {
			normalized.Body = body
		}
	case TriggeredAbility:
		if body, ok := normalized.TriggeredBody(); ok {
			normalized.Body = body
		}
	case StaticAbility:
		if body, ok := normalized.StaticBody(); ok {
			normalized.Body = body
		}
	default:
	}
	return normalized
}

// SpellBody returns this ability's spell body, including a legacy view.
func (ability *AbilityDef) SpellBody() (SpellAbilityBody, bool) {
	if body, ok := ability.Body.(SpellAbilityBody); ok {
		return body, true
	}
	if ability.Body != nil {
		return SpellAbilityBody{}, false
	}
	if ability.Kind != SpellAbility {
		return SpellAbilityBody{}, false
	}
	return SpellAbilityBody{
		Text:             ability.Text,
		Content:          ability.legacyContent(),
		AdditionalCosts:  append([]cost.Additional(nil), ability.AdditionalCosts...),
		AlternativeCosts: append([]cost.Alternative(nil), ability.AlternativeCosts...),
	}, true
}

// ActivatedBody returns this ability's activated body, including a legacy view.
func (ability *AbilityDef) ActivatedBody() (ActivatedAbilityBody, bool) {
	if body, ok := ability.Body.(ActivatedAbilityBody); ok {
		return body, true
	}
	if ability.Body != nil {
		return ActivatedAbilityBody{}, false
	}
	if ability.Kind != ActivatedAbility || ability.IsManaAbility || ability.IsLoyaltyAbility {
		return ActivatedAbilityBody{}, false
	}
	return ActivatedAbilityBody{
		Text:                ability.Text,
		ManaCost:            ability.ManaCost,
		AdditionalCosts:     append([]cost.Additional(nil), ability.AdditionalCosts...),
		AlternativeCosts:    append([]cost.Alternative(nil), ability.AlternativeCosts...),
		ZoneOfFunction:      ability.ZoneOfFunction,
		Timing:              ability.Timing,
		ActivationCondition: ability.ActivationCondition,
		Content:             ability.legacyContent(),
		KeywordAbilities:    append([]KeywordAbility(nil), ability.KeywordAbilities...),
	}, true
}

// ManaBody returns this ability's mana body, including a legacy view.
func (ability *AbilityDef) ManaBody() (ManaAbilityBody, bool) {
	if body, ok := ability.Body.(ManaAbilityBody); ok {
		return body, true
	}
	if ability.Body != nil {
		return ManaAbilityBody{}, false
	}
	if ability.Kind != ActivatedAbility || !ability.IsManaAbility || ability.IsLoyaltyAbility {
		return ManaAbilityBody{}, false
	}
	return ManaAbilityBody{
		Text:                ability.Text,
		ManaCost:            ability.ManaCost,
		AdditionalCosts:     append([]cost.Additional(nil), ability.AdditionalCosts...),
		ZoneOfFunction:      ability.ZoneOfFunction,
		Timing:              ability.Timing,
		ActivationCondition: ability.ActivationCondition,
		Content:             ability.legacyContent(),
	}, true
}

// LoyaltyBody returns this ability's loyalty body, including a legacy view.
func (ability *AbilityDef) LoyaltyBody() (LoyaltyAbilityBody, bool) {
	if body, ok := ability.Body.(LoyaltyAbilityBody); ok {
		return body, true
	}
	if ability.Body != nil {
		return LoyaltyAbilityBody{}, false
	}
	if ability.Kind != ActivatedAbility || !ability.IsLoyaltyAbility {
		return LoyaltyAbilityBody{}, false
	}
	return LoyaltyAbilityBody{
		Text:                ability.Text,
		LoyaltyCost:         ability.LoyaltyCost,
		ActivationCondition: ability.ActivationCondition,
		Content:             ability.legacyContent(),
	}, true
}

// TriggeredBody returns this ability's triggered body, including a legacy view.
func (ability *AbilityDef) TriggeredBody() (TriggeredAbilityBody, bool) {
	if body, ok := ability.Body.(TriggeredAbilityBody); ok {
		return body, true
	}
	if ability.Body != nil {
		return TriggeredAbilityBody{}, false
	}
	if ability.Kind != TriggeredAbility {
		return TriggeredAbilityBody{}, false
	}
	trigger := TriggerCondition{}
	if ability.Trigger.Exists {
		trigger = ability.Trigger.Val
	}
	return TriggeredAbilityBody{
		Text:               ability.Text,
		Trigger:            trigger,
		Optional:           ability.Optional,
		MaxTriggersPerTurn: ability.MaxTriggersPerTurn,
		Content:            ability.legacyContent(),
	}, true
}

// StaticBody returns this ability's static body, including a legacy view.
func (ability *AbilityDef) StaticBody() (StaticAbilityBody, bool) {
	if body, ok := ability.Body.(StaticAbilityBody); ok {
		return body, true
	}
	if ability.Body != nil {
		return StaticAbilityBody{}, false
	}
	if ability.Kind != StaticAbility {
		return StaticAbilityBody{}, false
	}
	return StaticAbilityBody{
		Text:             ability.Text,
		Condition:        ability.Condition,
		ZoneOfFunction:   ability.ZoneOfFunction,
		KeywordAbilities: append([]KeywordAbility(nil), ability.KeywordAbilities...),
	}, true
}

func (ability *AbilityDef) legacyContent() AbilityContent {
	if len(ability.Modes) != 0 || ability.MinModes != 0 || ability.MaxModes != 0 || ability.AllowDuplicateModes {
		return ModalAbilityContent{
			SharedTargets:       append([]TargetSpec(nil), ability.Targets...),
			Modes:               append([]Mode(nil), ability.Modes...),
			MinModes:            ability.MinModes,
			MaxModes:            ability.MaxModes,
			AllowDuplicateModes: ability.AllowDuplicateModes,
		}
	}
	return PlainAbilityContent{
		Targets: append([]TargetSpec(nil), ability.Targets...),
	}
}
