package game

import (
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/zone"
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
// Casting costs (AdditionalCosts, AlternativeCosts) live on the enclosing
// CardFace, not here.
type SpellAbilityBody struct {
	Text    string
	Content AbilityContent
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
	// KeywordAbilities lists keyword abilities carried by this triggered ability,
	// e.g. WardKeyword for ward triggers. Rules use it for keyword dispatch without
	// inspecting Content.
	KeywordAbilities []KeywordAbility
	Content          AbilityContent
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

// ReplacementAbilityBody is a replacement/prevention ability on a printed face.
type ReplacementAbilityBody struct {
	Text        string
	Replacement ReplacementEffect
	UnlessPaid  opt.V[ResolutionPayment]
}

// EntersTappedReplacement creates a replacement ability for "enters tapped".
func EntersTappedReplacement(text string) ReplacementAbilityBody {
	replacement := etbReplacement(text)
	replacement.EntersTapped = true
	return ReplacementAbilityBody{Text: text, Replacement: replacement}
}

// EntersTappedIfReplacement creates a conditional "enters tapped" replacement.
func EntersTappedIfReplacement(text string, condition *Condition) ReplacementAbilityBody {
	replacement := etbReplacement(text)
	replacement.Condition = opt.Val(*condition)
	replacement.EntersTapped = true
	return ReplacementAbilityBody{Text: text, Replacement: replacement}
}

// EntersTappedUnlessPaidReplacement creates an ETB payment replacement. If the
// payment is not paid, the permanent enters tapped.
func EntersTappedUnlessPaidReplacement(text string, payment ResolutionPayment) ReplacementAbilityBody {
	replacement := etbReplacement(text)
	replacement.EntersTapped = true
	return ReplacementAbilityBody{
		Text:        text,
		Replacement: replacement,
		UnlessPaid:  opt.Val(payment),
	}
}

// EntersWithCountersReplacement creates an ETB counter-placement replacement.
func EntersWithCountersReplacement(text string, placements ...CounterPlacement) ReplacementAbilityBody {
	replacement := etbReplacement(text)
	replacement.EntersWithCounters = append([]CounterPlacement(nil), placements...)
	return ReplacementAbilityBody{Text: text, Replacement: replacement}
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

func (SpellAbilityBody) isAbilityBody()       {}
func (ActivatedAbilityBody) isAbilityBody()   {}
func (ManaAbilityBody) isAbilityBody()        {}
func (LoyaltyAbilityBody) isAbilityBody()     {}
func (TriggeredAbilityBody) isAbilityBody()   {}
func (ReplacementAbilityBody) isAbilityBody() {}
func (StaticAbilityBody) isAbilityBody()      {}

// BodyContent returns the AbilityContent of a sealed ability body.
func BodyContent(body AbilityBody) AbilityContent {
	switch b := body.(type) {
	case SpellAbilityBody:
		return b.Content
	case *SpellAbilityBody:
		if b == nil {
			return nil
		}
		return b.Content
	case ActivatedAbilityBody:
		return b.Content
	case *ActivatedAbilityBody:
		if b == nil {
			return nil
		}
		return b.Content
	case ManaAbilityBody:
		return b.Content
	case *ManaAbilityBody:
		if b == nil {
			return nil
		}
		return b.Content
	case LoyaltyAbilityBody:
		return b.Content
	case *LoyaltyAbilityBody:
		if b == nil {
			return nil
		}
		return b.Content
	case TriggeredAbilityBody:
		return b.Content
	case *TriggeredAbilityBody:
		if b == nil {
			return nil
		}
		return b.Content
	default:
		return nil
	}
}

// BodyTargets returns the target specs for a sealed ability body's content.
// For plain content, returns the content's targets. For modal content, returns
// shared targets.
func BodyTargets(body AbilityBody) []TargetSpec {
	switch content := BodyContent(body).(type) {
	case PlainAbilityContent:
		return content.Targets
	case ModalAbilityContent:
		return content.SharedTargets
	default:
		return nil
	}
}

// BodyFunctionZone returns the zone where the body functions, if it has one.
func BodyFunctionZone(body AbilityBody) zone.Type {
	switch b := body.(type) {
	case StaticAbilityBody:
		return b.ZoneOfFunction
	case *StaticAbilityBody:
		if b == nil {
			return zone.None
		}
		return b.ZoneOfFunction
	case ActivatedAbilityBody:
		return b.ZoneOfFunction
	case *ActivatedAbilityBody:
		if b == nil {
			return zone.None
		}
		return b.ZoneOfFunction
	case ManaAbilityBody:
		return b.ZoneOfFunction
	case *ManaAbilityBody:
		if b == nil {
			return zone.None
		}
		return b.ZoneOfFunction
	default:
		return zone.None
	}
}

// BodyTimingRestriction returns the timing restriction for the body, if any.
func BodyTimingRestriction(body AbilityBody) TimingRestriction {
	switch b := body.(type) {
	case ActivatedAbilityBody:
		return b.Timing
	case *ActivatedAbilityBody:
		if b == nil {
			return NoTimingRestriction
		}
		return b.Timing
	case ManaAbilityBody:
		return b.Timing
	case *ManaAbilityBody:
		if b == nil {
			return NoTimingRestriction
		}
		return b.Timing
	default:
		return NoTimingRestriction
	}
}

// BodyActivationCondition returns the activation condition for the body, if any.
func BodyActivationCondition(body AbilityBody) opt.V[Condition] {
	switch b := body.(type) {
	case ActivatedAbilityBody:
		return b.ActivationCondition
	case *ActivatedAbilityBody:
		if b == nil {
			return opt.V[Condition]{}
		}
		return b.ActivationCondition
	case ManaAbilityBody:
		return b.ActivationCondition
	case *ManaAbilityBody:
		if b == nil {
			return opt.V[Condition]{}
		}
		return b.ActivationCondition
	case LoyaltyAbilityBody:
		return b.ActivationCondition
	case *LoyaltyAbilityBody:
		if b == nil {
			return opt.V[Condition]{}
		}
		return b.ActivationCondition
	default:
		return opt.V[Condition]{}
	}
}

// BodyLoyaltyCost returns the loyalty cost for the body, if any.
func BodyLoyaltyCost(body AbilityBody) int {
	switch loyalty := body.(type) {
	case LoyaltyAbilityBody:
		return loyalty.LoyaltyCost
	case *LoyaltyAbilityBody:
		if loyalty == nil {
			return 0
		}
		return loyalty.LoyaltyCost
	}
	return 0
}
