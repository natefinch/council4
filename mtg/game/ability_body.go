package game

import (
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/mana"
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
	ModeChoiceBonus     ModeChoiceBonus
	AllowDuplicateModes bool
}

// ModeChoiceCondition identifies a cast-time condition that expands the
// available modal choice range.
type ModeChoiceCondition int

const (
	// ModeChoiceConditionNone marks content without a modal bonus.
	ModeChoiceConditionNone ModeChoiceCondition = iota
	// ModeChoiceConditionControlsCommander requires controlling a commander.
	ModeChoiceConditionControlsCommander
)

// ModeChoiceBonus adds optional mode choices when its cast-time condition is
// true. Chosen modes remain locked into the stack object after announcement.
type ModeChoiceBonus struct {
	Condition          ModeChoiceCondition
	AdditionalMaxModes int
}

// IsModal reports whether the content requires a mode choice. Exactly one mode
// with a minimum and maximum of one is ordinary non-modal content.
//
// The receiver is a pointer for consistency with isAbility (see the receiver
// rationale in this file); AbilityContent values are always addressable at call
// sites, so this does not change how IsModal is called.
func (m *AbilityContent) IsModal() bool {
	return len(m.Modes) != 1 || m.MinModes != 1 || m.MaxModes != 1
}

// ActivatedAbility is a non-mana, non-loyalty activated ability.
type ActivatedAbility struct {
	Text                string
	ManaCost            opt.V[cost.Mana]
	AdditionalCosts     []cost.Additional
	AlternativeCosts    []cost.Alternative
	CostModifiers       []CostModifier
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

// EntersTappedGroupReplacement creates a continuous static replacement that taps
// a group of OTHER permanents as they enter, as in "Creatures your opponents
// control enter tapped." (Authority of the Consuls). The controller filter is
// evaluated relative to the source's controller and cardTypes restricts the
// affected permanents (empty taps every entering permanent).
func EntersTappedGroupReplacement(text string, controller TriggerControllerFilter, cardTypes ...types.Card) ReplacementAbility {
	replacement := etbReplacement(text)
	replacement.EntersTapped = true
	replacement.EntersTappedOthers = true
	replacement.ControllerFilter = controller
	replacement.EntersTappedTypes = append([]types.Card(nil), cardTypes...)
	return ReplacementAbility{Text: text, Replacement: replacement}
}

// GraveyardRedirectReplacement creates a continuous static replacement that
// exiles a card (or permanent) that would be put into a watched graveyard
// instead (CR 614), as in "If a card would be put into a graveyard from
// anywhere, exile it instead." (Leyline of the Void). ownerFilter selects whose
// graveyard is watched relative to the source's controller; cardTypes restricts
// the redirected cards to any of the listed types (empty redirects every card);
// fromBattlefieldOnly limits the redirect to cards leaving the battlefield ("a
// permanent").
func GraveyardRedirectReplacement(text string, ownerFilter TriggerControllerFilter, fromBattlefieldOnly bool, cardTypes ...types.Card) ReplacementAbility {
	replacement := ReplacementEffect{
		Description:            text,
		MatchEvent:             EventZoneChanged,
		MatchToZone:            true,
		ToZone:                 zone.Graveyard,
		ReplaceToZone:          zone.Exile,
		Duration:               DurationPermanent,
		ContinuousZoneRedirect: true,
		RedirectOwnerFilter:    ownerFilter,
		RedirectTypeFilter:     append([]types.Card(nil), cardTypes...),
	}
	if fromBattlefieldOnly {
		replacement.MatchFromZone = true
		replacement.FromZone = zone.Battlefield
	}
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

// EntersUnlessPaidElseZoneReplacement creates an optional self enters-the-
// battlefield replacement for "If this permanent would enter, you may <pay an
// alternative cost> instead. If you do, put it onto the battlefield. If you
// don't, put it into <zone>." (Mox Diamond). As the permanent would enter, its
// controller may pay the alternative cost to keep it on the battlefield; if the
// cost is not paid, the permanent is put into the destination zone instead of
// entering.
func EntersUnlessPaidElseZoneReplacement(text string, payment ResolutionPayment, destination zone.Type) ReplacementAbility {
	replacement := etbReplacement(text)
	replacement.ReplaceToZone = destination
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

// EntersWithCountersIfReplacement creates a conditional ETB counter-placement
// replacement, as in "This creature enters with a +1/+1 counter on it if you
// attacked this turn." (Raid) or "... if a creature died this turn." (Morbid).
// The counters are placed only when the condition is satisfied as the permanent
// enters (CR 614).
func EntersWithCountersIfReplacement(text string, condition *Condition, placements ...CounterPlacement) ReplacementAbility {
	replacement := etbReplacement(text)
	replacement.EntersWithCounters = append([]CounterPlacement(nil), placements...)
	replacement.Condition = opt.Val(*condition)
	return ReplacementAbility{Text: text, Replacement: replacement}
}

// BloodthirstReplacement creates the conditional ETB counter-placement
// replacement for the Bloodthirst N keyword (CR 702.54): "If an opponent was
// dealt damage this turn, this creature enters with N +1/+1 counters on it."
// The intervening-if is an event-history condition matching any damage dealt to
// an opponent during the current turn.
func BloodthirstReplacement(text string, n int) ReplacementAbility {
	condition := Condition{
		EventHistory: opt.Val(EventHistoryCondition{
			Pattern: TriggerPattern{
				Event:           EventDamageDealt,
				DamageRecipient: DamageRecipientPlayer,
				Player:          TriggerPlayerOpponent,
			},
			Window: EventHistoryCurrentTurn,
		}),
	}
	return EntersWithCountersIfReplacement(text, &condition, CounterPlacement{
		Kind:   counter.PlusOnePlusOne,
		Amount: n,
	})
}

// EntersTappedWithCountersReplacement creates a combined ETB replacement for
// "This permanent enters tapped with N <kind> counters on it." (the Vivid land
// cycle). The permanent enters tapped and with the listed counters.
func EntersTappedWithCountersReplacement(text string, placements ...CounterPlacement) ReplacementAbility {
	replacement := etbReplacement(text)
	replacement.EntersTapped = true
	replacement.EntersWithCounters = append([]CounterPlacement(nil), placements...)
	return ReplacementAbility{Text: text, Replacement: replacement}
}

// EntryColorChoiceReplacement creates an entry-time replacement for "As this
// permanent enters, choose a color." The controller chooses a color as the
// permanent enters and the result is stored on the permanent under
// EntryColorChoiceKey for later abilities to read (CR 614.12).
func EntryColorChoiceReplacement(text string) ReplacementAbility {
	replacement := etbReplacement(text)
	replacement.EntryColorChoice = true
	return ReplacementAbility{Text: text, Replacement: replacement}
}

// EntersTappedColorChoiceReplacement creates an entry-time replacement for the
// combined "This permanent enters tapped. As it enters, choose a color." The
// permanent enters tapped and the controller chooses a color as it enters; the
// result is stored on the permanent under EntryColorChoiceKey (CR 614.12).
func EntersTappedColorChoiceReplacement(text string) ReplacementAbility {
	replacement := etbReplacement(text)
	replacement.EntersTapped = true
	replacement.EntryColorChoice = true
	return ReplacementAbility{Text: text, Replacement: replacement}
}

// EntryColorChoiceExcludingReplacement creates an entry-time replacement for "As
// this permanent enters, choose a color other than <color>." The controller
// chooses a color other than the excluded one as the permanent enters; the
// result is stored on the permanent under EntryColorChoiceKey (CR 614.12).
func EntryColorChoiceExcludingReplacement(text string, exclude mana.Color) ReplacementAbility {
	replacement := etbReplacement(text)
	replacement.EntryColorChoice = true
	replacement.EntryColorChoiceExclude = exclude
	return ReplacementAbility{Text: text, Replacement: replacement}
}

// EntersTappedColorChoiceExcludingReplacement creates the combined "This land
// enters tapped. As it enters, choose a color other than <color>." entry
// replacement of the Gate/Thriving land cycle. The permanent enters tapped and
// the controller chooses a color other than the excluded one; the result is
// stored on the permanent under EntryColorChoiceKey (CR 614.12).
func EntersTappedColorChoiceExcludingReplacement(text string, exclude mana.Color) ReplacementAbility {
	replacement := etbReplacement(text)
	replacement.EntersTapped = true
	replacement.EntryColorChoice = true
	replacement.EntryColorChoiceExclude = exclude
	return ReplacementAbility{Text: text, Replacement: replacement}
}

// EntryTypeChoiceReplacement creates an entry-time replacement for "As this
// permanent enters, choose a creature type." The controller chooses a creature
// type as the permanent enters and the result is stored on the permanent under
// EntryTypeChoiceKey for later abilities to read (CR 614.12).
func EntryTypeChoiceReplacement(text string) ReplacementAbility {
	replacement := etbReplacement(text)
	replacement.EntryTypeChoice = true
	return ReplacementAbility{Text: text, Replacement: replacement}
}

// EntersAsCopyReplacement creates a self enters-the-battlefield replacement for
// "You may have this creature enter the battlefield as a copy of <filter>[,
// except <rider>]." (Clone, Clever Impersonator, Phyrexian Metamorph). As the
// permanent enters its controller chooses one permanent matching selection to
// copy; optional marks the "You may ..." form, notLegendary applies the "except
// it isn't legendary" rider, and addTypes applies the "except it's an <type> in
// addition to its other types" rider (CR 706, CR 614).
func EntersAsCopyReplacement(text string, selection *Selection, optional, notLegendary bool, conditionalCounters []ConditionalCounterPlacement, untilEndOfTurn bool, addKeywords []Keyword, addTypes ...types.Card) ReplacementAbility {
	replacement := etbReplacement(text)
	replacement.EntersAsCopy = true
	replacement.EntersAsCopyOptional = optional
	replacement.EntersAsCopySelection = selection
	replacement.EntersAsCopyNotLegendary = notLegendary
	replacement.EntersAsCopyAddTypes = append([]types.Card(nil), addTypes...)
	replacement.EntersAsCopyConditionalCounters = append([]ConditionalCounterPlacement(nil), conditionalCounters...)
	replacement.EntersAsCopyUntilEndOfTurn = untilEndOfTurn
	replacement.EntersAsCopyAddKeywords = append([]Keyword(nil), addKeywords...)
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

// TokenCreationReplacementSpec parameterizes a token-creation value-modifying
// replacement (CR 614). Multiplier multiplies the created token count (1 leaves
// it unchanged) and Addend then adds a fixed number of extra tokens. Subtypes,
// when non-empty, restricts the replacement to tokens carrying all listed
// subtypes. Filter scopes which player's creations are affected.
type TokenCreationReplacementSpec struct {
	Multiplier int
	Addend     int
	Subtypes   []types.Sub
	Filter     TriggerControllerFilter
}

// TokenCreationReplacementFiltered creates a persistent replacement that
// multiplies and/or augments token creation events, optionally restricted to a
// token subtype and scoped by player. It backs the token-doubling family
// (Doubling Season, Primal Vigor) and the additive variant (Xorn).
func TokenCreationReplacementFiltered(text string, spec *TokenCreationReplacementSpec) ReplacementAbility {
	return ReplacementAbility{
		Text: text,
		Replacement: ReplacementEffect{
			Description:           text,
			MatchEvent:            EventTokenCreated,
			ControllerFilter:      spec.Filter,
			TokenMultiplier:       spec.Multiplier,
			TokenAddend:           spec.Addend,
			TokenRequiredSubtypes: append([]types.Sub(nil), spec.Subtypes...),
			Duration:              DurationPermanent,
		},
	}
}

// NamedTokenSetReplacement creates a persistent replacement that, when the
// controller would create a token whose name matches one of defs, instead
// creates one of each token in defs (Academy Manufactor). The defs double as
// both the trigger set (matched by name) and the tokens created instead.
func NamedTokenSetReplacement(text string, defs []*CardDef, filter TriggerControllerFilter) ReplacementAbility {
	return ReplacementAbility{
		Text: text,
		Replacement: ReplacementEffect{
			Description:           text,
			MatchEvent:            EventTokenCreated,
			ControllerFilter:      filter,
			CreateOneOfEachTokens: defs,
			Duration:              DurationPermanent,
		},
	}
}

// DrawFromEmptyLibraryWinReplacement creates a persistent replacement that, when
// the controller would draw a card from an empty library, instead causes that
// controller to win the game (Laboratory Maniac, Jace, Wielder of Mysteries). It
// is registered while its source is on the battlefield.
func DrawFromEmptyLibraryWinReplacement(text string) ReplacementAbility {
	return ReplacementAbility{
		Text: text,
		Replacement: ReplacementEffect{
			Description:              text,
			MatchEvent:               EventCardDrawn,
			ControllerFilter:         TriggerControllerYou,
			DrawFromEmptyLibraryWins: true,
			Duration:                 DurationPermanent,
		},
	}
}

// DrawCardMultiplierReplacement builds the draw-doubling replacement "If you
// would draw a card[ except the first one you draw in each of your draw steps],
// draw <multiplier> cards instead." (Teferi's Ageless Insight, Thought
// Reflection). When exceptFirstInDrawStep is set, the controller's first draw in
// each of their own draw steps is not multiplied.
func DrawCardMultiplierReplacement(text string, multiplier int, exceptFirstInDrawStep bool) ReplacementAbility {
	return ReplacementAbility{
		Text: text,
		Replacement: ReplacementEffect{
			Description:                   text,
			MatchEvent:                    EventCardDrawn,
			ControllerFilter:              TriggerControllerYou,
			DrawCardMultiplier:            multiplier,
			DrawCardExceptFirstInDrawStep: exceptFirstInDrawStep,
			Duration:                      DurationPermanent,
		},
	}
}

// CounterPlacementReplacement creates a persistent replacement that modifies
// placement of one specific counter kind by multiplying the count and then
// adding a fixed amount (CR 614).
func CounterPlacementReplacement(text string, multiplier, addend int, kindFilter counter.Kind, filter TriggerControllerFilter) ReplacementAbility {
	replacement := AnyCounterPlacementReplacement(text, multiplier, addend, filter)
	replacement.Replacement.MatchCounterKind = true
	replacement.Replacement.CounterKindFilter = kindFilter
	replacement.Replacement.CounterRecipientTypes = []types.Card{types.Creature}
	replacement.Replacement.CounterUseRecipientController = true
	return replacement
}

// AnyCounterPlacementReplacement creates a persistent replacement that modifies
// placement of any counter kind by multiplying the count and then adding a
// fixed amount (CR 614).
func AnyCounterPlacementReplacement(text string, multiplier, addend int, filter TriggerControllerFilter) ReplacementAbility {
	return ReplacementAbility{
		Text: text,
		Replacement: ReplacementEffect{
			Description:       text,
			MatchEvent:        EventCountersAdded,
			ControllerFilter:  filter,
			CounterMultiplier: multiplier,
			CounterAddend:     addend,
			Duration:          DurationPermanent,
		},
	}
}

// ControlledPermanentCounterPlacementReplacement creates a persistent
// replacement that modifies placement of any counter kind on a permanent the
// controller controls, as in Doubling Season (CR 614).
func ControlledPermanentCounterPlacementReplacement(text string, multiplier, addend int, filter TriggerControllerFilter) ReplacementAbility {
	replacement := AnyCounterPlacementReplacement(text, multiplier, addend, filter)
	replacement.Replacement.CounterUseRecipientController = true
	replacement.Replacement.CounterRecipientAnyPermanent = true
	return replacement
}

// ControlledPermanentCounterKindPlacementReplacement creates a persistent
// replacement that modifies placement of one specific counter kind on a
// permanent the controller controls, as in Kami of Whispered Hopes (CR 614).
func ControlledPermanentCounterKindPlacementReplacement(text string, multiplier, addend int, kindFilter counter.Kind, filter TriggerControllerFilter) ReplacementAbility {
	replacement := ControlledPermanentCounterPlacementReplacement(text, multiplier, addend, filter)
	replacement.Replacement.MatchCounterKind = true
	replacement.Replacement.CounterKindFilter = kindFilter
	return replacement
}

// DamageReplacement creates a persistent replacement that modifies damage from
// matching sources before it is dealt.
func DamageReplacement(text string, multiplier, addend int, sourceColors []color.Color, filter TriggerControllerFilter) ReplacementAbility {
	return ReplacementAbility{
		Text: text,
		Replacement: ReplacementEffect{
			Description:        text,
			MatchEvent:         EventDamageDealt,
			ControllerFilter:   filter,
			DamageMultiplier:   multiplier,
			DamageAddend:       addend,
			DamageSourceColors: append([]color.Color(nil), sourceColors...),
			Duration:           DurationPermanent,
		},
	}
}

// DamageReplacementExcludingSource creates a damage replacement that does not
// apply to damage from the permanent carrying the replacement ability.
func DamageReplacementExcludingSource(text string, multiplier, addend int, sourceColors []color.Color, filter TriggerControllerFilter) ReplacementAbility {
	replacement := DamageReplacement(text, multiplier, addend, sourceColors, filter)
	replacement.Replacement.DamageExcludeSource = true
	return replacement
}

// DamageReplacementSpec parameterizes a filtered damage replacement: the
// multiplier and additive bonus applied to the damage, the source's color and
// card-type filters, whether the source belongs to the replacement's controller
// (Controller) or any player, whether the replacement excludes its own source,
// whether it applies only to opponent recipients, and whether it applies only
// to noncombat damage.
type DamageReplacementSpec struct {
	Multiplier        int
	Addend            int
	SourceColors      []color.Color
	SourceTypes       []types.Card
	ExcludeSource     bool
	RecipientOpponent bool
	NoncombatOnly     bool
	Controller        TriggerControllerFilter
}

// DamageReplacementFiltered creates a persistent damage replacement (CR 614)
// from a spec, supporting opponent-only recipients, noncombat-only damage, and
// card-type source filters in addition to the color and source-exclusion
// filters of DamageReplacement.
func DamageReplacementFiltered(text string, spec *DamageReplacementSpec) ReplacementAbility {
	return ReplacementAbility{
		Text: text,
		Replacement: ReplacementEffect{
			Description:             text,
			MatchEvent:              EventDamageDealt,
			ControllerFilter:        spec.Controller,
			DamageMultiplier:        spec.Multiplier,
			DamageAddend:            spec.Addend,
			DamageSourceColors:      append([]color.Color(nil), spec.SourceColors...),
			DamageSourceTypes:       append([]types.Card(nil), spec.SourceTypes...),
			DamageExcludeSource:     spec.ExcludeSource,
			DamageRecipientOpponent: spec.RecipientOpponent,
			DamageNoncombatOnly:     spec.NoncombatOnly,
			Duration:                DurationPermanent,
		},
	}
}

// LifeGainReplacement creates a persistent replacement that modifies life the
// controller would gain by multiplying it and then adding a fixed amount (CR
// 614), backing wordings such as "you gain twice that much life instead" and
// "you gain that much life plus N instead".
func LifeGainReplacement(text string, multiplier, addend int) ReplacementAbility {
	return ReplacementAbility{
		Text: text,
		Replacement: ReplacementEffect{
			Description:        text,
			MatchEvent:         EventLifeGained,
			ControllerFilter:   TriggerControllerYou,
			LifeGainMultiplier: multiplier,
			LifeGainAddend:     addend,
			Duration:           DurationPermanent,
		},
	}
}

// LifeLossReplacement creates a persistent replacement that modifies life a
// player would lose by multiplying it and then adding a fixed amount (CR 614),
// backing "they lose twice that much life instead." (Bloodletter of Aclazotz).
// recipientOpponent restricts it to opponents of the controller (false matches
// any player); duringControllerTurn restricts it to the controller's own turn.
func LifeLossReplacement(text string, multiplier, addend int, recipientOpponent, duringControllerTurn bool) ReplacementAbility {
	return ReplacementAbility{
		Text: text,
		Replacement: ReplacementEffect{
			Description:                  text,
			MatchEvent:                   EventLifeLost,
			ControllerFilter:             TriggerControllerYou,
			LifeLossMultiplier:           multiplier,
			LifeLossAddend:               addend,
			LifeLossRecipientOpponent:    recipientOpponent,
			LifeLossDuringControllerTurn: duringControllerTurn,
			Duration:                     DurationPermanent,
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

// The Ability marker is implemented on POINTER receivers, and the ability
// accessors (CardFace.BodyAt and friends) hand back the address of the slice
// element rather than the element itself.
//
// Why pointers: an Ability is an interface, so returning a concrete ability
// VALUE as an Ability boxes a heap copy of the (large) ability struct on every
// call. CardFace.BodyAt is called for every ability of every permanent on every
// effective-value computation, so in long games that boxing dominated
// allocation (billions of allocations / hundreds of GB in a profiled game).
// With pointer receivers, BodyAt returns &face.Slice[i]: the interface wraps the
// existing, addressable element and allocates nothing.
//
// This is safe only because a *CardDef is treated as immutable once a card
// instance references it (all modifications go through copyCardDef on a fresh
// copy; Game.Clone deliberately shares *CardDef). The returned pointer aliases
// into the def's ability slice, so callers must treat it as read-only and must
// not retain it across a mutation of that def. See docs/adr/0010 (abilities
// addressed, not copied).
func (*AbilityContent) isAbility()     {}
func (*ActivatedAbility) isAbility()   {}
func (*ManaAbility) isAbility()        {}
func (*LoyaltyAbility) isAbility()     {}
func (*TriggeredAbility) isAbility()   {}
func (*ChapterAbility) isAbility()     {}
func (*ReplacementAbility) isAbility() {}
func (*StaticAbility) isAbility()      {}

// BodyContent returns the content of a sealed ability body.
func BodyContent(body Ability) AbilityContent {
	switch b := body.(type) {
	case *AbilityContent:
		if b == nil {
			return AbilityContent{}
		}
		return *b
	case *ActivatedAbility:
		if b == nil {
			return AbilityContent{}
		}
		return b.Content
	case *ManaAbility:
		if b == nil {
			return AbilityContent{}
		}
		return b.Content
	case *LoyaltyAbility:
		if b == nil {
			return AbilityContent{}
		}
		return b.Content
	case *TriggeredAbility:
		if b == nil {
			return AbilityContent{}
		}
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
	case *StaticAbility:
		if b == nil {
			return zone.None
		}
		return b.ZoneOfFunction
	case *ActivatedAbility:
		if b == nil {
			return zone.None
		}
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
	case *ActivatedAbility:
		if b == nil {
			return NoTimingRestriction
		}
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
	case *ActivatedAbility:
		if b == nil {
			return opt.V[Condition]{}
		}
		return b.ActivationCondition
	case *ManaAbility:
		if b == nil {
			return opt.V[Condition]{}
		}
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
	if loyalty, ok := body.(*LoyaltyAbility); ok && loyalty != nil {
		return loyalty.LoyaltyCost
	}
	return 0
}
