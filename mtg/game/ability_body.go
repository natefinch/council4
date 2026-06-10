package game

import (
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// Ability is a sealed data-only variant for how an ability functions.
type Ability interface {
	isAbility()
}

// AbilityContent is an ability's target and instruction content. Ordinary
// non-modal abilities contain one required mode; modal abilities contain a
// choice among multiple modes or use a mode range other than exactly one.
type AbilityContent struct {
	SharedTargets       []TargetSpec
	Modes               []Mode
	MinModes            int
	MaxModes            int
	AllowDuplicateModes bool
}

// IsModal reports whether the content requires a mode choice. Exactly one mode
// with a minimum and maximum of one is ordinary non-modal content.
func (m AbilityContent) IsModal() bool {
	return len(m.Modes) != 1 || m.MinModes != 1 || m.MaxModes != 1
}

// ActivatedAbility is a non-mana, non-loyalty activated ability.
type ActivatedAbility struct {
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

// ManaAbility is an activated mana ability.
type ManaAbility struct {
	Text                string
	ManaCost            opt.V[cost.Mana]
	AdditionalCosts     []cost.Additional
	ZoneOfFunction      zone.Type
	Timing              TimingRestriction
	ActivationCondition opt.V[Condition]
	// Content is the mana output.
	Content AbilityContent
}

// LoyaltyAbility is a planeswalker loyalty ability.
type LoyaltyAbility struct {
	Text                string
	LoyaltyCost         int
	ActivationCondition opt.V[Condition]
	Content             AbilityContent
}

// TriggeredAbility is an ability that triggers from a game event or state.
type TriggeredAbility struct {
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

// ChapterAbility is a Saga chapter ability associated with one or more lore
// counter numbers.
type ChapterAbility struct {
	Text     string
	Chapters []int
	Content  AbilityContent
}

// StaticAbility is a static ability that functions from a zone.
type StaticAbility struct {
	Text              string
	Condition         opt.V[Condition]
	ZoneOfFunction    zone.Type
	KeywordAbilities  []KeywordAbility
	ContinuousEffects []ContinuousEffect
	RuleEffects       []RuleEffect
}

// ReplacementAbility is a replacement/prevention ability on a printed face.
type ReplacementAbility struct {
	Text        string
	Replacement ReplacementEffect
	UnlessPaid  opt.V[ResolutionPayment]
}

// EntersTappedReplacement creates a replacement ability for "enters tapped".
func EntersTappedReplacement(text string) ReplacementAbility {
	replacement := etbReplacement(text)
	replacement.EntersTapped = true
	return ReplacementAbility{Text: text, Replacement: replacement}
}

// EntersTappedIfReplacement creates a conditional "enters tapped" replacement.
func EntersTappedIfReplacement(text string, condition *Condition) ReplacementAbility {
	replacement := etbReplacement(text)
	replacement.Condition = opt.Val(*condition)
	replacement.EntersTapped = true
	return ReplacementAbility{Text: text, Replacement: replacement}
}

// EntersTappedUnlessPaidReplacement creates an ETB payment replacement. If the
// payment is not paid, the permanent enters tapped.
func EntersTappedUnlessPaidReplacement(text string, payment ResolutionPayment) ReplacementAbility {
	replacement := etbReplacement(text)
	replacement.EntersTapped = true
	return ReplacementAbility{
		Text:        text,
		Replacement: replacement,
		UnlessPaid:  opt.Val(payment),
	}
}

// EntersWithCountersReplacement creates an ETB counter-placement replacement.
func EntersWithCountersReplacement(text string, placements ...CounterPlacement) ReplacementAbility {
	replacement := etbReplacement(text)
	replacement.EntersWithCounters = append([]CounterPlacement(nil), placements...)
	return ReplacementAbility{Text: text, Replacement: replacement}
}

// TokenCreationReplacement creates a persistent replacement that multiplies
// token creation events matching controller.
func TokenCreationReplacement(text string, multiplier int, filter TriggerControllerFilter) ReplacementAbility {
	return ReplacementAbility{
		Text: text,
		Replacement: ReplacementEffect{
			Description:      text,
			MatchEvent:       EventTokenCreated,
			ControllerFilter: filter,
			TokenMultiplier:  multiplier,
			Duration:         DurationPermanent,
		},
	}
}

// CounterPlacementReplacement creates a persistent replacement that multiplies
// placement of one specific counter kind.
func CounterPlacementReplacement(text string, multiplier int, kindFilter counter.Kind, filter TriggerControllerFilter) ReplacementAbility {
	replacement := AnyCounterPlacementReplacement(text, multiplier, filter)
	replacement.Replacement.MatchCounterKind = true
	replacement.Replacement.CounterKindFilter = kindFilter
	replacement.Replacement.CounterRecipientTypes = []types.Card{types.Creature}
	replacement.Replacement.CounterUseRecipientController = true
	return replacement
}

// AnyCounterPlacementReplacement creates a persistent replacement that
// multiplies placement of any counter kind.
func AnyCounterPlacementReplacement(text string, multiplier int, filter TriggerControllerFilter) ReplacementAbility {
	return ReplacementAbility{
		Text: text,
		Replacement: ReplacementEffect{
			Description:       text,
			MatchEvent:        EventCountersAdded,
			ControllerFilter:  filter,
			CounterMultiplier: multiplier,
			Duration:          DurationPermanent,
		},
	}
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

func (AbilityContent) isAbility()     {}
func (ActivatedAbility) isAbility()   {}
func (ManaAbility) isAbility()        {}
func (LoyaltyAbility) isAbility()     {}
func (TriggeredAbility) isAbility()   {}
func (ChapterAbility) isAbility()     {}
func (ReplacementAbility) isAbility() {}
func (StaticAbility) isAbility()      {}

// BodyContent returns the content of a sealed ability body.
func BodyContent(body Ability) AbilityContent {
	switch b := body.(type) {
	case AbilityContent:
		return b
	case *AbilityContent:
		if b == nil {
			return AbilityContent{}
		}
		return *b
	case ActivatedAbility:
		return b.Content
	case *ActivatedAbility:
		if b == nil {
			return AbilityContent{}
		}
		return b.Content
	case ManaAbility:
		return b.Content
	case *ManaAbility:
		if b == nil {
			return AbilityContent{}
		}
		return b.Content
	case LoyaltyAbility:
		return b.Content
	case *LoyaltyAbility:
		if b == nil {
			return AbilityContent{}
		}
		return b.Content
	case TriggeredAbility:
		return b.Content
	case *TriggeredAbility:
		if b == nil {
			return AbilityContent{}
		}
		return b.Content
	case ChapterAbility:
		return b.Content
	case *ChapterAbility:
		if b == nil {
			return AbilityContent{}
		}
		return b.Content
	default:
		return AbilityContent{}
	}
}

// BodyTargets returns the target specs for a sealed ability body's content.
// Non-modal content uses its sole mode's targets; modal content uses shared targets.
func BodyTargets(body Ability) []TargetSpec {
	content := BodyContent(body)
	if len(content.Modes) == 1 && !content.IsModal() {
		targets := append([]TargetSpec(nil), content.SharedTargets...)
		return append(targets, content.Modes[0].Targets...)
	}
	return content.SharedTargets
}

// BodyFunctionZone returns the zone where the body functions, if it has one.
func BodyFunctionZone(body Ability) zone.Type {
	switch b := body.(type) {
	case StaticAbility:
		return b.ZoneOfFunction
	case *StaticAbility:
		if b == nil {
			return zone.None
		}
		return b.ZoneOfFunction
	case ActivatedAbility:
		return b.ZoneOfFunction
	case *ActivatedAbility:
		if b == nil {
			return zone.None
		}
		return b.ZoneOfFunction
	case ManaAbility:
		return b.ZoneOfFunction
	case *ManaAbility:
		if b == nil {
			return zone.None
		}
		return b.ZoneOfFunction
	default:
		return zone.None
	}
}

// BodyTimingRestriction returns the timing restriction for the body, if any.
func BodyTimingRestriction(body Ability) TimingRestriction {
	switch b := body.(type) {
	case ActivatedAbility:
		return b.Timing
	case *ActivatedAbility:
		if b == nil {
			return NoTimingRestriction
		}
		return b.Timing
	case ManaAbility:
		return b.Timing
	case *ManaAbility:
		if b == nil {
			return NoTimingRestriction
		}
		return b.Timing
	default:
		return NoTimingRestriction
	}
}

// BodyActivationCondition returns the activation condition for the body, if any.
func BodyActivationCondition(body Ability) opt.V[Condition] {
	switch b := body.(type) {
	case ActivatedAbility:
		return b.ActivationCondition
	case *ActivatedAbility:
		if b == nil {
			return opt.V[Condition]{}
		}
		return b.ActivationCondition
	case ManaAbility:
		return b.ActivationCondition
	case *ManaAbility:
		if b == nil {
			return opt.V[Condition]{}
		}
		return b.ActivationCondition
	case LoyaltyAbility:
		return b.ActivationCondition
	case *LoyaltyAbility:
		if b == nil {
			return opt.V[Condition]{}
		}
		return b.ActivationCondition
	default:
		return opt.V[Condition]{}
	}
}

// BodyLoyaltyCost returns the loyalty cost for the body, if any.
func BodyLoyaltyCost(body Ability) int {
	switch loyalty := body.(type) {
	case LoyaltyAbility:
		return loyalty.LoyaltyCost
	case *LoyaltyAbility:
		if loyalty == nil {
			return 0
		}
		return loyalty.LoyaltyCost
	}
	return 0
}
