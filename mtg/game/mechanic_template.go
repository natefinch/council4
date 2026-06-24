package game

import (
	"fmt"
	"slices"
	"strconv"
	"strings"

	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/compare"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

const tapManaChoiceKey = ChoiceKey("oracle-mana-color")

const tapManaCommanderColorKey = ChoiceKey("oracle-commander-color")

const (
	cumulativeUpkeepAgeCounterResult = ResultKey("cumulative-upkeep-age-counter")
	cumulativeUpkeepPaymentResult    = ResultKey("cumulative-upkeep-payment")
)

// tapManaLandsProduceKey publishes the color chosen for a "mana of any color
// that a land ... could produce" ability (Reflecting Pool, Exotic Orchard,
// Fellwar Stone; see TapManaLandsProduceAbility).
const tapManaLandsProduceKey = ChoiceKey("oracle-lands-produce-color")

// tapManaLinkedExileColorKey publishes the color chosen for a "mana of any of
// the exiled card's colors" ability (Chrome Mox; see
// TapLinkedExileColorManaAbility).
const tapManaLinkedExileColorKey = ChoiceKey("oracle-linked-exile-color")

// tapManaAmongControlledColorsKey publishes the color chosen for a "mana of any
// color among <permanents> you control" ability (Mox Amber, Plaza of Heroes;
// see TapManaAmongControlledColorsAbility).
const tapManaAmongControlledColorsKey = ChoiceKey("oracle-among-controlled-color")

// tapManaFilterFirstKey and tapManaFilterSecondKey publish the two independent
// color choices of a filter-land mana ability (see TwoColorFilterManaAbility).
// They are distinct so the instruction sequence publishes each choice under its
// own key (CR 608.2/duplicate-key validation).
const tapManaFilterFirstKey = ChoiceKey("oracle-filter-mana-first")

const tapManaFilterSecondKey = ChoiceKey("oracle-filter-mana-second")

// CantBlockStaticBody is the complete static ability for a creature that cannot block.
var CantBlockStaticBody = StaticAbility{
	Text: "This creature can't block.",
	RuleEffects: []RuleEffect{{
		Kind:           RuleEffectCantBlock,
		AffectedSource: true,
	}},
}

// CantAttackStaticBody is the complete static ability for a creature that cannot attack.
var CantAttackStaticBody = StaticAbility{
	Text: "This creature can't attack.",
	RuleEffects: []RuleEffect{{
		Kind:           RuleEffectCantAttack,
		AffectedSource: true,
	}},
}

// MustBeBlockedStaticBody is the complete static ability for a creature that
// must be blocked if able.
var MustBeBlockedStaticBody = StaticAbility{
	Text: "This creature must be blocked if able.",
	RuleEffects: []RuleEffect{{
		Kind:           RuleEffectMustBeBlocked,
		AffectedSource: true,
	}},
}

// CantBeBlockedStaticBody is the complete static ability for an unblockable creature.
var CantBeBlockedStaticBody = StaticAbility{
	Text: "This creature can't be blocked.",
	RuleEffects: []RuleEffect{{
		Kind:           RuleEffectCantBeBlocked,
		AffectedSource: true,
	}},
}

// MustAttackStaticBody is the complete static ability for a creature that must attack.
var MustAttackStaticBody = StaticAbility{
	Text: "This creature attacks each combat if able.",
	RuleEffects: []RuleEffect{{
		Kind:           RuleEffectMustAttack,
		AffectedSource: true,
	}},
}

// CantBeCounteredStaticBody is the complete static ability for an uncounterable spell.
var CantBeCounteredStaticBody = StaticAbility{
	Text:           "This spell can't be countered.",
	ZoneOfFunction: zone.Stack,
	RuleEffects: []RuleEffect{{
		Kind:           RuleEffectCantBeCountered,
		AffectedSource: true,
	}},
}

// DoesntUntapStaticBody is the complete static ability for a permanent that does
// not untap during its controller's untap step.
var DoesntUntapStaticBody = StaticAbility{
	Text: "This permanent doesn't untap during your untap step.",
	RuleEffects: []RuleEffect{{
		Kind:           RuleEffectDoesntUntap,
		AffectedSource: true,
	}},
}

// CantAttackOrBlockStaticBody is the complete static ability for a creature that
// can neither attack nor block.
var CantAttackOrBlockStaticBody = StaticAbility{
	Text: "This creature can't attack or block.",
	RuleEffects: []RuleEffect{
		{Kind: RuleEffectCantAttack, AffectedSource: true},
		{Kind: RuleEffectCantBlock, AffectedSource: true},
	},
}

// NoMaximumHandSizeStaticBody is the complete static ability for "You have no
// maximum hand size." The controller never discards down to a hand-size limit.
var NoMaximumHandSizeStaticBody = StaticAbility{
	Text: "You have no maximum hand size.",
	RuleEffects: []RuleEffect{{
		Kind:           RuleEffectNoMaximumHandSize,
		AffectedPlayer: PlayerYou,
	}},
}

// PlayLandsFromGraveyardStaticBody is the complete static ability for "You may
// play lands from your graveyard." The controller may play land cards from their
// graveyard, subject to the usual one-land-per-turn limit.
var PlayLandsFromGraveyardStaticBody = StaticAbility{
	Text: "You may play lands from your graveyard.",
	RuleEffects: []RuleEffect{{
		Kind:           RuleEffectPlayLandsFromZone,
		AffectedPlayer: PlayerYou,
		CastFromZone:   zone.Graveyard,
		PermanentTypes: []types.Card{types.Land},
	}},
}

// PlayLandsFromLibraryTopStaticBody is the complete static ability for "You may
// play lands from the top of your library." The controller may play the top card
// of their library if it is a land, subject to the usual one-land-per-turn limit.
var PlayLandsFromLibraryTopStaticBody = StaticAbility{
	Text: "You may play lands from the top of your library.",
	RuleEffects: []RuleEffect{{
		Kind:           RuleEffectPlayLandsFromZone,
		AffectedPlayer: PlayerYou,
		CastFromZone:   zone.Library,
		PermanentTypes: []types.Card{types.Land},
		TopCardOnly:    true,
	}},
}

// PlayWithTopCardRevealedStaticBody is the complete static ability for "Play with
// the top card of your library revealed." The controller's top library card is
// revealed to all players.
var PlayWithTopCardRevealedStaticBody = StaticAbility{
	Text: "Play with the top card of your library revealed.",
	RuleEffects: []RuleEffect{{
		Kind:           RuleEffectPlayWithTopCardRevealed,
		AffectedPlayer: PlayerYou,
	}},
}

// LookAtTopCardAnyTimeStaticBody is the complete static ability for "You may look
// at the top card of your library any time." The controller may privately look
// at the top card of their library at any time.
var LookAtTopCardAnyTimeStaticBody = StaticAbility{
	Text: "You may look at the top card of your library any time.",
	RuleEffects: []RuleEffect{{
		Kind:           RuleEffectLookAtTopCardAnyTime,
		AffectedPlayer: PlayerYou,
	}},
}

// WardStaticAbility builds the complete static ability for Ward with a mana cost.
func WardStaticAbility(manaCost cost.Mana) StaticAbility {
	keywordCost := append(cost.Mana(nil), manaCost...)
	return StaticAbility{
		Text: "Ward " + manaCost.String(),
		KeywordAbilities: []KeywordAbility{
			WardKeyword{Cost: keywordCost},
		},
	}
}

// WardStaticAbilityWithCosts builds the complete static ability for Ward with a
// composite or non-mana cost ("Ward—Pay 2 life.", "Ward—{2}, Pay 2 life.",
// "Ward—Sacrifice a creature."). manaCost may be empty when the ward cost has no
// mana component; additionalCosts carries the non-mana components an opponent
// must pay alongside the mana to avoid having their spell or ability countered.
func WardStaticAbilityWithCosts(manaCost cost.Mana, additionalCosts []cost.Additional) StaticAbility {
	return StaticAbility{
		Text: "Ward",
		KeywordAbilities: []KeywordAbility{
			WardKeyword{
				Cost:            append(cost.Mana(nil), manaCost...),
				AdditionalCosts: slices.Clone(additionalCosts),
			},
		},
	}
}

// DredgeStaticAbility builds the complete static ability for the Dredge N
// keyword (CR 702.52). It functions from its owner's graveyard, where it offers
// to replace one of that player's draws with milling n cards and returning this
// card to hand. n is the printed N and must be positive.
func DredgeStaticAbility(n int) StaticAbility {
	return StaticAbility{
		Text:             "Dredge " + strconv.Itoa(n),
		ZoneOfFunction:   zone.Graveyard,
		KeywordAbilities: []KeywordAbility{DredgeKeyword{Count: n}},
	}
}

// EnchantStaticAbility builds the complete static ability for Enchant.
func EnchantStaticAbility(target *TargetSpec) StaticAbility {
	targetCopy := cloneTargetSpec(target)
	return StaticAbility{
		Text: "Enchant " + targetCopy.Constraint,
		KeywordAbilities: []KeywordAbility{
			EnchantKeyword{Target: targetCopy},
		},
	}
}

func cloneTargetSpec(source *TargetSpec) TargetSpec {
	target := *source
	target.Predicate.PermanentTypes = append([]types.Card(nil), target.Predicate.PermanentTypes...)
	target.Predicate.ExcludedTypes = append([]types.Card(nil), target.Predicate.ExcludedTypes...)
	target.Predicate.Supertypes = append([]types.Super(nil), target.Predicate.Supertypes...)
	target.Predicate.Subtypes = append([]types.Sub(nil), target.Predicate.Subtypes...)
	target.Predicate.Colors = append([]color.Color(nil), target.Predicate.Colors...)
	target.Predicate.ExcludedColors = append([]color.Color(nil), target.Predicate.ExcludedColors...)
	if target.Selection.Exists {
		target.Selection = opt.Val(cloneSelection(target.Selection.Val))
	}
	return target
}

// ProtectionFromColorsStaticAbility builds the complete static ability for
// protection from one or more colors.
func ProtectionFromColorsStaticAbility(colors ...color.Color) StaticAbility {
	protectedColors := append([]color.Color(nil), colors...)
	validateProtectionColors(protectedColors)
	return StaticAbility{
		Text: protectionFromColorsText(protectedColors),
		KeywordAbilities: []KeywordAbility{
			ProtectionKeyword{FromColors: protectedColors},
		},
	}
}

// ProtectionFromTypesStaticAbility builds the static ability for protection
// from one or more card types.
func ProtectionFromTypesStaticAbility(cardTypes ...types.Card) StaticAbility {
	ts := append([]types.Card(nil), cardTypes...)
	if len(ts) == 0 {
		panic("game: protection from types requires at least one type")
	}
	return StaticAbility{
		Text:             protectionFromTypesText(ts),
		KeywordAbilities: []KeywordAbility{ProtectionKeyword{FromTypes: ts}},
	}
}

// ProtectionFromSubtypesStaticAbility builds the static ability for protection
// from one or more creature/land subtypes.
func ProtectionFromSubtypesStaticAbility(subtypes ...types.Sub) StaticAbility {
	ss := append([]types.Sub(nil), subtypes...)
	if len(ss) == 0 {
		panic("game: protection from subtypes requires at least one subtype")
	}
	return StaticAbility{
		Text:             protectionFromSubtypesText(ss),
		KeywordAbilities: []KeywordAbility{ProtectionKeyword{FromSubtypes: ss}},
	}
}

// ProtectionFromEverythingStaticAbility builds the static ability for
// protection from everything.
func ProtectionFromEverythingStaticAbility() StaticAbility {
	return StaticAbility{
		Text:             "Protection from everything",
		KeywordAbilities: []KeywordAbility{ProtectionKeyword{Everything: true}},
	}
}

// ProtectionFromEachColorStaticAbility builds the static ability for
// protection from each color.
func ProtectionFromEachColorStaticAbility() StaticAbility {
	return StaticAbility{
		Text:             "Protection from each color",
		KeywordAbilities: []KeywordAbility{ProtectionKeyword{EachColor: true}},
	}
}

// ProtectionFromChosenColorStaticAbility builds the static ability for
// protection from a single color chosen as the granting ability resolves. The
// rules rewrite the ChosenColor marker into a concrete FromColors entry before
// the continuous effect is stored.
func ProtectionFromChosenColorStaticAbility() StaticAbility {
	return StaticAbility{
		Text:             "Protection from the color of your choice",
		KeywordAbilities: []KeywordAbility{ProtectionKeyword{ChosenColor: true}},
	}
}

// ProtectionFromMulticoloredStaticAbility builds the static ability for
// protection from multicolored sources.
func ProtectionFromMulticoloredStaticAbility() StaticAbility {
	return StaticAbility{
		Text:             "Protection from multicolored",
		KeywordAbilities: []KeywordAbility{ProtectionKeyword{Multicolored: true}},
	}
}

// ProtectionFromMonocoloredStaticAbility builds the static ability for
// protection from monocolored sources.
func ProtectionFromMonocoloredStaticAbility() StaticAbility {
	return StaticAbility{
		Text:             "Protection from monocolored",
		KeywordAbilities: []KeywordAbility{ProtectionKeyword{Monocolored: true}},
	}
}

func validateProtectionColors(colors []color.Color) {
	if len(colors) == 0 {
		panic("game: protection requires at least one color")
	}
	seen := make(map[color.Color]struct{}, len(colors))
	for _, protectedColor := range colors {
		switch protectedColor {
		case color.White, color.Blue, color.Black, color.Red, color.Green:
		default:
			panic(fmt.Sprintf("game: invalid protection color %q", protectedColor))
		}
		if _, ok := seen[protectedColor]; ok {
			panic(fmt.Sprintf("game: duplicate protection color %q", protectedColor))
		}
		seen[protectedColor] = struct{}{}
	}
}

func protectionFromColorsText(colors []color.Color) string {
	phrases := make([]string, len(colors))
	for i, protectedColor := range colors {
		phrases[i] = "from " + strings.ToLower(string(protectedColor))
	}
	switch len(phrases) {
	case 1:
		return "Protection " + phrases[0]
	case 2:
		return "Protection " + phrases[0] + " and " + phrases[1]
	default:
		return "Protection " +
			strings.Join(phrases[:len(phrases)-1], ", ") +
			", and " +
			phrases[len(phrases)-1]
	}
}

func protectionFromTypesText(cardTypes []types.Card) string {
	phrases := make([]string, len(cardTypes))
	for i, t := range cardTypes {
		phrases[i] = "from " + strings.ToLower(string(t)) + "s"
	}
	return "Protection " + joinProtectionPhrases(phrases)
}

func protectionFromSubtypesText(subtypes []types.Sub) string {
	phrases := make([]string, len(subtypes))
	for i, s := range subtypes {
		phrases[i] = "from " + strings.ToLower(string(s)) + "s"
	}
	return "Protection " + joinProtectionPhrases(phrases)
}

func joinProtectionPhrases(phrases []string) string {
	switch len(phrases) {
	case 1:
		return phrases[0]
	case 2:
		return phrases[0] + " and " + phrases[1]
	default:
		return strings.Join(phrases[:len(phrases)-1], ", ") + ", and " + phrases[len(phrases)-1]
	}
}

// CyclingActivatedAbility builds the complete activated ability for Cycling
// with a mana cost.
func CyclingActivatedAbility(manaCost cost.Mana) ActivatedAbility {
	activationCost := append(cost.Mana(nil), manaCost...)
	keywordCost := append(cost.Mana(nil), manaCost...)
	return ActivatedAbility{
		Text:           "Cycling " + manaCost.String(),
		ManaCost:       opt.Val(activationCost),
		ZoneOfFunction: zone.Hand,
		AdditionalCosts: []cost.Additional{{
			Kind:   cost.AdditionalDiscard,
			Text:   "Discard this card",
			Amount: 1,
			Source: zone.Hand,
		}},
		KeywordAbilities: []KeywordAbility{
			CyclingKeyword{Cost: keywordCost},
		},
		Content: Mode{Sequence: []Instruction{{
			Primitive: Draw{
				Amount: Fixed(1),
				Player: ControllerReference(),
			},
		}}}.Ability(),
	}
}

// OutlastActivatedAbility builds the complete activated ability for Outlast
// with a mana cost (CR 702.105): "[cost], {T}: Put a +1/+1 counter on this
// creature. Activate only as a sorcery." The keyword carries no continuous
// effect; the activated ability is self-contained.
func OutlastActivatedAbility(manaCost cost.Mana) ActivatedAbility {
	activationCost := append(cost.Mana(nil), manaCost...)
	keywordCost := append(cost.Mana(nil), manaCost...)
	return ActivatedAbility{
		Text:            "Outlast " + manaCost.String(),
		ManaCost:        opt.Val(activationCost),
		AdditionalCosts: cost.Tap,
		ZoneOfFunction:  zone.Battlefield,
		Timing:          SorceryOnly,
		KeywordAbilities: []KeywordAbility{
			OutlastKeyword{Cost: keywordCost},
		},
		Content: Mode{Sequence: []Instruction{{
			Primitive: AddCounter{
				Amount:      Fixed(1),
				Object:      SourcePermanentReference(),
				CounterKind: counter.PlusOnePlusOne,
			},
		}}}.Ability(),
	}
}

// SaddleActivatedAbility builds the complete activated ability for Saddle N
// (CR 702.166): "Tap any number of other creatures you control with total power
// N or more: This Mount becomes saddled until end of turn. Saddle only as a
// sorcery." The ability has no mana cost; its additional cost taps other
// creatures the controller controls with total power at least n.
func SaddleActivatedAbility(n int) ActivatedAbility {
	return ActivatedAbility{
		Text: "Saddle " + strconv.Itoa(n),
		AdditionalCosts: []cost.Additional{{
			Kind:               cost.AdditionalTapPermanents,
			Text:               "Tap any number of other creatures you control with total power " + strconv.Itoa(n) + " or more",
			MatchPermanentType: true,
			PermanentType:      types.Creature,
			ExcludeSource:      true,
			TotalPowerAtLeast:  n,
		}},
		ZoneOfFunction: zone.Battlefield,
		Timing:         SorceryOnly,
		KeywordAbilities: []KeywordAbility{
			SaddleKeyword{Power: n},
		},
		Content: Mode{Sequence: []Instruction{{
			Primitive: BecomeSaddled{Object: SourcePermanentReference()},
		}}}.Ability(),
	}
}

// CrewActivatedAbility builds the complete activated ability for Crew N
// (CR 702.122): "Tap any number of creatures you control with total power N or
// more: This Vehicle becomes an artifact creature until end of turn." The
// ability has no mana cost and is activated at instant speed; its additional
// cost taps creatures the controller controls with total power at least n. The
// Vehicle keeps its printed power and toughness, which become relevant once it
// is a creature.
func CrewActivatedAbility(n int) ActivatedAbility {
	return ActivatedAbility{
		Text: "Crew " + strconv.Itoa(n),
		AdditionalCosts: []cost.Additional{{
			Kind:               cost.AdditionalTapPermanents,
			Text:               "Tap any number of creatures you control with total power " + strconv.Itoa(n) + " or more",
			MatchPermanentType: true,
			PermanentType:      types.Creature,
			TotalPowerAtLeast:  n,
		}},
		ZoneOfFunction: zone.Battlefield,
		KeywordAbilities: []KeywordAbility{
			CrewKeyword{Power: n},
		},
		Content: Mode{Sequence: []Instruction{{
			Primitive: ApplyContinuous{
				Object: opt.Val(SourcePermanentReference()),
				ContinuousEffects: []ContinuousEffect{{
					Layer:    LayerType,
					AddTypes: []types.Card{types.Creature},
				}},
				Duration: DurationUntilEndOfTurn,
			},
		}}}.Ability(),
	}
}

// LandcyclingActivatedAbility builds the complete activated ability for the
// typed landcycling family (Basic landcycling, Plainscycling, and so on). It is
// a cycling variant (CR 702.29): the discard-from-hand activation searches the
// library for a land matching spec instead of drawing a card. The caller
// supplies the land filter through spec; the source zone, hand destination, and
// reveal are fixed by the template.
func LandcyclingActivatedAbility(manaCost cost.Mana, spec SearchSpec) ActivatedAbility {
	activationCost := append(cost.Mana(nil), manaCost...)
	keywordCost := append(cost.Mana(nil), manaCost...)
	spec.SourceZone = zone.Library
	spec.Destination = zone.Hand
	spec.Reveal = true
	return ActivatedAbility{
		ManaCost:       opt.Val(activationCost),
		ZoneOfFunction: zone.Hand,
		AdditionalCosts: []cost.Additional{{
			Kind:   cost.AdditionalDiscard,
			Text:   "Discard this card",
			Amount: 1,
			Source: zone.Hand,
		}},
		KeywordAbilities: []KeywordAbility{
			CyclingKeyword{Cost: keywordCost},
		},
		Content: Mode{Sequence: []Instruction{{
			Primitive: Search{
				Player: ControllerReference(),
				Spec:   spec,
				Amount: Fixed(1),
			},
		}}}.Ability(),
	}
}

// CumulativeUpkeepTriggeredAbility builds the complete upkeep trigger for a
// fixed mana cost.
func CumulativeUpkeepTriggeredAbility(manaCost cost.Mana) TriggeredAbility {
	keywordCost := slices.Clone(manaCost)
	paymentCost := slices.Clone(manaCost)
	multiplier := DynamicAmount{
		Kind:        DynamicAmountObjectCounters,
		Object:      SourcePermanentReference(),
		CounterKind: counter.Age,
	}
	return TriggeredAbility{
		Text: "Cumulative upkeep " + manaCost.String(),
		Trigger: TriggerCondition{
			Pattern: TriggerPattern{
				Event:      EventBeginningOfStep,
				Controller: TriggerControllerYou,
				Step:       StepUpkeep,
			},
		},
		KeywordAbilities: []KeywordAbility{
			CumulativeUpkeepKeyword{Cost: keywordCost},
		},
		Content: Mode{Sequence: []Instruction{
			{
				Primitive: AddCounter{
					Amount:      Fixed(1),
					Object:      SourcePermanentReference(),
					CounterKind: counter.Age,
				},
				PublishResult: cumulativeUpkeepAgeCounterResult,
			},
			{
				Primitive: Pay{Payment: ResolutionPayment{
					ManaCost:           opt.Val(paymentCost),
					ManaCostMultiplier: opt.Val(&multiplier),
				}},
				ResultGate: opt.Val(InstructionResultGate{
					Key:       cumulativeUpkeepAgeCounterResult,
					Succeeded: TriTrue,
				}),
				PublishResult: cumulativeUpkeepPaymentResult,
			},
			{
				Primitive: Sacrifice{Object: SourcePermanentReference()},
				ResultGate: opt.Val(InstructionResultGate{
					Key:       cumulativeUpkeepPaymentResult,
					Succeeded: TriFalse,
				}),
			},
		}}.Ability(),
	}
}

// fabricateServoToken is the canonical 1/1 colorless Servo artifact creature
// token created by the Fabricate keyword.
var fabricateServoToken = &CardDef{
	CardFace: CardFace{
		Name:      "Servo",
		Types:     []types.Card{types.Artifact, types.Creature},
		Subtypes:  []types.Sub{types.Servo},
		Power:     opt.Val(PT{Value: 1}),
		Toughness: opt.Val(PT{Value: 1}),
	},
}

// FabricateTriggeredAbility builds the entry trigger for Fabricate N: a modal
// choice to put N +1/+1 counters on this creature or create N Servo tokens.
func FabricateTriggeredAbility(count int) TriggeredAbility {
	return TriggeredAbility{
		Text: fmt.Sprintf("Fabricate %d", count),
		Trigger: TriggerCondition{
			Type: TriggerWhen,
			Pattern: TriggerPattern{
				Event:  EventPermanentEnteredBattlefield,
				Source: TriggerSourceSelf,
			},
		},
		KeywordAbilities: []KeywordAbility{
			FabricateKeyword{Count: count},
		},
		Content: AbilityContent{
			Modes: []Mode{
				{
					Text: fmt.Sprintf("Put %d +1/+1 counters on it.", count),
					Sequence: []Instruction{{
						Primitive: AddCounter{
							Amount:      Fixed(count),
							Object:      EventPermanentReference(),
							CounterKind: counter.PlusOnePlusOne,
						},
					}},
				},
				{
					Text: fmt.Sprintf("Create %d 1/1 colorless Servo artifact creature tokens.", count),
					Sequence: []Instruction{{
						Primitive: CreateToken{
							Amount: Fixed(count),
							Source: TokenDef(fabricateServoToken),
						},
					}},
				},
			},
			MinModes: 1,
			MaxModes: 1,
		},
	}
}

// EvokeSacrificeTriggeredAbility builds the canonical Evoke sacrifice trigger
// (CR 702.74): "When this permanent enters, if its evoke cost was paid,
// sacrifice it." The intervening-if gates the sacrifice on the spell having been
// cast for its Evoke alternative cost, preserved on the entering permanent event
// for both trigger-time and resolution-time checks (CR 603.4).
func EvokeSacrificeTriggeredAbility() TriggeredAbility {
	return TriggeredAbility{
		Text: "When this permanent enters, if its evoke cost was paid, sacrifice it.",
		Trigger: TriggerCondition{
			Type: TriggerWhen,
			Pattern: TriggerPattern{
				Event:  EventPermanentEnteredBattlefield,
				Source: TriggerSourceSelf,
			},
			InterveningIfEventPermanentWasEvoked: true,
		},
		Content: Mode{
			Sequence: []Instruction{{
				Primitive: Sacrifice{Object: SourcePermanentReference()},
			}},
		}.Ability(),
	}
}

// RampageTriggeredAbility builds the canonical Rampage N triggered ability
// (CR 702.23): "Whenever this creature becomes blocked, it gets +N/+N until end
// of turn for each creature blocking it beyond the first." The +N/+N delta is a
// dynamic amount counting the source's blockers beyond the first as the ability
// resolves, scaled by N, then locked for the turn. Each printed instance is its
// own triggered ability, so multiple instances stack (CR 702.23c).
func RampageTriggeredAbility(n int) TriggeredAbility {
	delta := Dynamic(DynamicAmount{
		Kind:       DynamicAmountBlockingCreaturesBeyondFirst,
		Multiplier: n,
	})
	return TriggeredAbility{
		Text: fmt.Sprintf("Rampage %d", n),
		Trigger: TriggerCondition{
			Type: TriggerWhenever,
			Pattern: TriggerPattern{
				Event:  EventAttackerBecameBlocked,
				Source: TriggerSourceSelf,
			},
		},
		KeywordAbilities: []KeywordAbility{
			RampageKeyword{Count: n},
		},
		Content: Mode{Sequence: []Instruction{{
			Primitive: ModifyPT{
				Object:         SourcePermanentReference(),
				PowerDelta:     delta,
				ToughnessDelta: delta,
				Duration:       DurationUntilEndOfTurn,
			},
		}}}.Ability(),
	}
}

// NinjutsuActivatedAbility builds the complete hand-zone activation template
// for Ninjutsu with a mana cost.
func NinjutsuActivatedAbility(manaCost cost.Mana) ActivatedAbility {
	activationCost := append(cost.Mana(nil), manaCost...)
	keywordCost := append(cost.Mana(nil), manaCost...)
	return ActivatedAbility{
		Text:           "Ninjutsu " + manaCost.String(),
		ManaCost:       opt.Val(activationCost),
		ZoneOfFunction: zone.Hand,
		Timing:         DuringCombat,
		AdditionalCosts: []cost.Additional{{
			Kind:   cost.AdditionalReturnUnblockedAttacker,
			Text:   "Return an unblocked attacker you control to its owner's hand",
			Amount: 1,
		}},
		KeywordAbilities: []KeywordAbility{
			NinjutsuKeyword{Cost: keywordCost},
		},
		Content: Mode{}.Ability(),
	}
}

// MutateStaticAbility builds the hand-zone keyword ability for Mutate.
func MutateStaticAbility(manaCost cost.Mana) StaticAbility {
	keywordCost := append(cost.Mana(nil), manaCost...)
	return StaticAbility{
		Text:           "Mutate " + manaCost.String(),
		ZoneOfFunction: zone.Hand,
		KeywordAbilities: []KeywordAbility{
			MutateKeyword{Cost: keywordCost},
		},
	}
}

// EquipActivatedAbility builds the complete activated ability for Equip with a
// mana cost.
func EquipActivatedAbility(manaCost cost.Mana) ActivatedAbility {
	return EquipRestrictedActivatedAbility(manaCost, nil, nil)
}

// EquipRestrictedActivatedAbility builds the complete activated ability for a
// restricted Equip ("Equip legendary creature {3}", "Equip Knight {2}"): the
// Equipment may attach only to a creature you control that has every supertype
// and at least one of the subtypes. Nil supertypes and subtypes yield the
// unrestricted Equip.
func EquipRestrictedActivatedAbility(manaCost cost.Mana, supertypes []types.Super, subtypes []types.Sub) ActivatedAbility {
	activationCost := append(cost.Mana(nil), manaCost...)
	keywordCost := append(cost.Mana(nil), manaCost...)
	predicate := TargetPredicate{
		PermanentTypes: []types.Card{types.Creature},
		Controller:     ControllerYou,
		Supertypes:     append([]types.Super(nil), supertypes...),
	}
	predicate.Subtypes = append([]types.Sub(nil), subtypes...)
	return ActivatedAbility{
		Text:           "Equip " + manaCost.String(),
		ManaCost:       opt.Val(activationCost),
		ZoneOfFunction: zone.Battlefield,
		Timing:         SorceryOnly,
		KeywordAbilities: []KeywordAbility{
			EquipKeyword{Cost: keywordCost},
		},
		Content: Mode{Targets: []TargetSpec{{
			MinTargets: 1,
			MaxTargets: 1,
			Constraint: equipRestrictionConstraint(supertypes, subtypes),
			Allow:      TargetAllowPermanent,
			Predicate:  predicate,
		}}}.Ability(),
	}
}

// equipRestrictionConstraint renders the human-readable target constraint for a
// restricted Equip, defaulting to "creature you control".
func equipRestrictionConstraint(supertypes []types.Super, subtypes []types.Sub) string {
	words := make([]string, 0, len(supertypes)+2)
	for _, supertype := range supertypes {
		words = append(words, strings.ToLower(string(supertype)))
	}
	noun := "creature"
	if len(subtypes) > 0 {
		parts := make([]string, len(subtypes))
		for i, subtype := range subtypes {
			parts[i] = string(subtype)
		}
		noun = strings.Join(parts, " or ")
	}
	words = append(words, noun, "you control")
	return strings.Join(words, " ")
}

// TapManaAbility builds the complete "{T}: Add {X}." mana ability.
func TapManaAbility(manaColor mana.Color) ManaAbility {
	return ManaAbility{
		Text:            fmt.Sprintf("{T}: Add {%s}.", manaSymbol(manaColor)),
		AdditionalCosts: cost.Tap,
		Content: Mode{Sequence: []Instruction{{
			Primitive: AddMana{
				Amount:    Fixed(1),
				ManaColor: manaColor,
			},
		}}}.Ability(),
	}
}

func manaSymbol(manaColor mana.Color) string {
	switch manaColor {
	case mana.W, mana.U, mana.B, mana.R, mana.G:
		return string(manaColor)
	case mana.C:
		return "C"
	default:
		panic(fmt.Sprintf("game: invalid mana color %q", manaColor))
	}
}

// TapManaChoiceAbility builds the complete tap ability for adding one mana
// chosen from two through five colors.
func TapManaChoiceAbility(colors ...mana.Color) ManaAbility {
	manaColors := append([]mana.Color(nil), colors...)
	validateManaColorChoice(manaColors)
	prompt := "Choose a color"
	if containsManaColor(manaColors, mana.C) {
		prompt = "Choose a type of mana"
	}
	return ManaAbility{
		Text:            tapManaChoiceText(manaColors),
		AdditionalCosts: []cost.Additional{{Kind: cost.AdditionalTap}},
		Content: Mode{Sequence: []Instruction{
			{
				Primitive: Choose{
					Choice: ResolutionChoice{
						Kind:   ResolutionChoiceMana,
						Prompt: prompt,
						Colors: manaColors,
					},
					PublishChoice: tapManaChoiceKey,
				},
			},
			{
				Primitive: AddMana{
					Amount:     Fixed(1),
					ChoiceFrom: tapManaChoiceKey,
				},
			},
		}}.Ability(),
	}
}

// TapManaChoiceCountAbility builds the tap ability for "Add <count> mana of any
// one color." (Gilded Lotus: "Add three mana of any one color."), count >= 2.
// The controller chooses a single color from colors as the ability resolves and
// adds that many mana of the one chosen color. text is the ability's oracle
// text; the renderer passes it through so the rendered ability matches the
// lowered one regardless of the cardinal wording.
func TapManaChoiceCountAbility(text string, count int, colors ...mana.Color) ManaAbility {
	manaColors := append([]mana.Color(nil), colors...)
	validateManaColorChoice(manaColors)
	prompt := "Choose a color"
	if containsManaColor(manaColors, mana.C) {
		prompt = "Choose a type of mana"
	}
	return ManaAbility{
		Text:            text,
		AdditionalCosts: []cost.Additional{{Kind: cost.AdditionalTap}},
		Content: Mode{Sequence: []Instruction{
			{
				Primitive: Choose{
					Choice: ResolutionChoice{
						Kind:   ResolutionChoiceMana,
						Prompt: prompt,
						Colors: manaColors,
					},
					PublishChoice: tapManaChoiceKey,
				},
			},
			{
				Primitive: AddMana{
					Amount:     Fixed(count),
					ChoiceFrom: tapManaChoiceKey,
				},
			},
		}}.Ability(),
	}
}

// TapManaChosenColorDevotionAbility builds the resolving content for "Choose a
// color. Add an amount of mana of that color equal to your devotion to that
// color." (Nykthos, Shrine to Nyx). The controller chooses a color as the
// ability resolves; the produced mana is that color and its amount is the
// controller's devotion to the chosen color (CR 700.5), read from the published
// color choice so the devotion color tracks the choice rather than a fixed
// color.
func TapManaChosenColorDevotionAbility(text string) ManaAbility {
	return ManaAbility{
		Text:            text,
		AdditionalCosts: cost.Tap,
		Content: Mode{Sequence: []Instruction{
			{
				Primitive: Choose{
					Choice: ResolutionChoice{
						Kind:   ResolutionChoiceMana,
						Prompt: "Choose a color",
						Colors: []mana.Color{mana.W, mana.U, mana.B, mana.R, mana.G},
					},
					PublishChoice: tapManaChoiceKey,
				},
			},
			{
				Primitive: AddMana{
					Amount: Dynamic(DynamicAmount{
						Kind:      DynamicAmountDevotion,
						ColorFrom: tapManaChoiceKey,
					}),
					ChoiceFrom: tapManaChoiceKey,
				},
			},
		}}.Ability(),
	}
}

// TapManaChosenColorCountAbility builds the complete tap ability for "Choose a
// color. Add an amount of mana of that color equal to <dynamic count>." (Three
// Tree City: "...equal to the number of creatures you control of the chosen
// type."). The controller chooses a color as the ability resolves; the produced
// mana is that color and its amount is the count of battlefield permanents
// matching selection.
func TapManaChosenColorCountAbility(text string, selection Selection) ManaAbility {
	return ManaAbility{
		Text:            text,
		AdditionalCosts: cost.Tap,
		Content: Mode{Sequence: []Instruction{
			{
				Primitive: Choose{
					Choice: ResolutionChoice{
						Kind:   ResolutionChoiceMana,
						Prompt: "Choose a color",
						Colors: []mana.Color{mana.W, mana.U, mana.B, mana.R, mana.G},
					},
					PublishChoice: tapManaChoiceKey,
				},
			},
			{
				Primitive: AddMana{
					Amount: Dynamic(DynamicAmount{
						Kind:       DynamicAmountCountSelector,
						Multiplier: 1,
						Group:      BattlefieldGroup(selection),
					}),
					ChoiceFrom: tapManaChoiceKey,
				},
			},
		}}.Ability(),
	}
}

// TapManaChosenColorDynamicAbility builds the complete tap ability for "Add X
// mana of any one color, where X is <dynamic amount>." (Kami of Whispered Hopes:
// "...where X is this creature's power."). The controller chooses any one color
// as the ability resolves; the produced mana is that color and its amount is the
// supplied dynamic value.
func TapManaChosenColorDynamicAbility(text string, amount DynamicAmount) ManaAbility {
	return ManaAbility{
		Text:            text,
		AdditionalCosts: cost.Tap,
		Content: Mode{Sequence: []Instruction{
			{
				Primitive: Choose{
					Choice: ResolutionChoice{
						Kind:   ResolutionChoiceMana,
						Prompt: "Choose a color",
						Colors: []mana.Color{mana.W, mana.U, mana.B, mana.R, mana.G},
					},
					PublishChoice: tapManaChoiceKey,
				},
			},
			{
				Primitive: AddMana{
					Amount:     Dynamic(amount),
					ChoiceFrom: tapManaChoiceKey,
				},
			},
		}}.Ability(),
	}
}

// TapManaChoiceWithSpendRiderAbility builds a tap mana-choice ability whose
// produced unit carries the supplied spend restriction or rider.
func TapManaChoiceWithSpendRiderAbility(text string, rider ManaSpendRider, colors ...mana.Color) ManaAbility {
	ability := TapManaChoiceAbility(colors...)
	ability.Text = text
	add, ok := ability.Content.Modes[0].Sequence[1].Primitive.(AddMana)
	if !ok {
		panic("game: tap mana choice template has no add-mana instruction")
	}
	add.SpendRider = opt.Val(rider)
	ability.Content.Modes[0].Sequence[1].Primitive = add
	return ability
}

// TapChosenColorManaAbility builds the complete tap ability for "{T}: Add one
// mana of the chosen color." The color is read from the entry-time choice stored
// on the source permanent under EntryColorChoiceKey, so this ability prompts no
// choice of its own.
func TapChosenColorManaAbility(text string) ManaAbility {
	return ManaAbility{
		Text:            text,
		AdditionalCosts: cost.Tap,
		Content: Mode{Sequence: []Instruction{
			{
				Primitive: AddMana{
					Amount:          Fixed(1),
					EntryChoiceFrom: EntryColorChoiceKey,
				},
			},
		}}.Ability(),
	}
}

// TapFixedOrChosenColorManaAbility builds the complete tap ability for the
// composite "{T}: Add {C} or one mana of the chosen color." (the Gate/Thriving
// land cycle). On activation the controller chooses between the fixed color and
// the color chosen as the source permanent entered (read from EntryColorChoiceKey
// seeded on the resolving ability); one mana of the selected color is added.
func TapFixedOrChosenColorManaAbility(text string, fixed mana.Color) ManaAbility {
	return ManaAbility{
		Text:            text,
		AdditionalCosts: cost.Tap,
		Content: Mode{Sequence: []Instruction{
			{
				Primitive: Choose{
					Choice: ResolutionChoice{
						Kind:           ResolutionChoiceMana,
						Prompt:         "Choose a color",
						Colors:         []mana.Color{fixed},
						ColorSource:    ResolutionChoiceColorSourceFixedOrEntryChosen,
						EntryChoiceKey: EntryColorChoiceKey,
					},
					PublishChoice: tapManaChoiceKey,
				},
			},
			{
				Primitive: AddMana{
					Amount:     Fixed(1),
					ChoiceFrom: tapManaChoiceKey,
				},
			},
		}}.Ability(),
	}
}

// TapManaCommanderIdentityAbility builds the complete "{T}: Add one mana of any
// color in your commander's color identity." mana ability (CR 903.4). The
// choosable colors are resolved dynamically from the controller's commander
// color identity at activation; the ability is unactivatable when that identity
// is empty.
func TapManaCommanderIdentityAbility() ManaAbility {
	return ManaAbility{
		Text:            "{T}: Add one mana of any color in your commander's color identity.",
		AdditionalCosts: cost.Tap,
		Content: Mode{Sequence: []Instruction{
			{
				Primitive: Choose{
					Choice: ResolutionChoice{
						Kind:        ResolutionChoiceMana,
						Prompt:      "Choose a color in your commander's color identity",
						ColorSource: ResolutionChoiceColorSourceCommanderIdentity,
					},
					PublishChoice: tapManaCommanderColorKey,
				},
			},
			{
				Primitive: AddMana{
					Amount:     Fixed(1),
					ChoiceFrom: tapManaCommanderColorKey,
				},
			},
		}}.Ability(),
	}
}

// TapManaCommanderIdentityWithSpendRiderAbility builds the commander-identity
// mana ability of TapManaCommanderIdentityAbility with rider attached to the
// produced mana, modelling Path of Ancestry's "{T}: Add one mana of any color
// in your commander's color identity. When that mana is spent to cast a creature
// spell that shares a creature type with your commander, scry 1." The add-mana
// instruction tags each produced unit with rider so the rules engine fires it
// when that exact mana is spent on a qualifying spell. Producing the mana stays
// a mana ability (CR 605); the rider uses the stack when it fires.
func TapManaCommanderIdentityWithSpendRiderAbility(text string, rider ManaSpendRider) ManaAbility {
	return ManaAbility{
		Text:            text,
		AdditionalCosts: cost.Tap,
		Content: Mode{Sequence: []Instruction{
			{
				Primitive: Choose{
					Choice: ResolutionChoice{
						Kind:        ResolutionChoiceMana,
						Prompt:      "Choose a color in your commander's color identity",
						ColorSource: ResolutionChoiceColorSourceCommanderIdentity,
					},
					PublishChoice: tapManaCommanderColorKey,
				},
			},
			{
				Primitive: AddMana{
					Amount:     Fixed(1),
					ChoiceFrom: tapManaCommanderColorKey,
					SpendRider: opt.Val(rider),
				},
			},
		}}.Ability(),
	}
}

// TapLinkedExileColorManaAbility builds the complete "{T}: Add one mana of any
// of the exiled card's colors." mana ability (Chrome Mox). linkID names the
// object-scoped linked object (the imprinted exiled card) published by the
// source permanent's enter-the-battlefield exile. The choosable colors are
// recomputed from that linked card at resolution; a missing, declined, or
// colorless imprint leaves the choice empty and the ability unactivatable
// (CR 605.1a), while a multicolored imprint offers exactly its colors.
func TapLinkedExileColorManaAbility(linkID string) ManaAbility {
	return ManaAbility{
		Text:            "{T}: Add one mana of any of the exiled card's colors.",
		AdditionalCosts: cost.Tap,
		Content: Mode{Sequence: []Instruction{
			{
				Primitive: Choose{
					Choice: ResolutionChoice{
						Kind:        ResolutionChoiceMana,
						Prompt:      "Choose a color of the exiled card",
						ColorSource: ResolutionChoiceColorSourceLinkedExileColors,
						LinkID:      linkID,
					},
					PublishChoice: tapManaLinkedExileColorKey,
				},
			},
			{
				Primitive: AddMana{
					Amount:     Fixed(1),
					ChoiceFrom: tapManaLinkedExileColorKey,
				},
			},
		}}.Ability(),
	}
}

// TapManaLandsProduceAbility builds the complete mana ability that adds one mana
// of any color a land could produce, scoped to lands matching the given player
// relation (CR 106.7, CR 605.1a). PlayerYou models Reflecting Pool
// ("a land you control could produce."); PlayerOpponent models Exotic Orchard
// and Fellwar Stone ("a land an opponent controls could produce."). When
// includeColorless is true the ability uses the "any type" wording and also
// offers colorless ({C}) if a matching land could produce it (Reflecting Pool,
// Naga Vitalist); otherwise it uses "any color" and offers only colored mana.
// The choosable mana is recomputed from the battlefield at resolution: every
// color (and colorless, when included) any matching land's mana abilities could
// add. When no matching land could produce mana the choice is empty and the
// ability is unactivatable. Mana abilities that derive their color from this
// same source contribute nothing, matching the loop-avoidance ruling for two
// opposing Exotic Orchards.
func TapManaLandsProduceAbility(relation PlayerRelation, includeColorless bool) ManaAbility {
	text, prompt := landsProduceTexts(relation, includeColorless)
	return ManaAbility{
		Text:            text,
		AdditionalCosts: cost.Tap,
		Content: Mode{Sequence: []Instruction{
			{
				Primitive: Choose{
					Choice: ResolutionChoice{
						Kind:             ResolutionChoiceMana,
						Prompt:           prompt,
						ColorSource:      ResolutionChoiceColorSourceLandsProduce,
						PlayerRelation:   relation,
						IncludeColorless: includeColorless,
					},
					PublishChoice: tapManaLandsProduceKey,
				},
			},
			{
				Primitive: AddMana{
					Amount:     Fixed(1),
					ChoiceFrom: tapManaLandsProduceKey,
				},
			},
		}}.Ability(),
	}
}

// landsProduceTexts returns the exact oracle text and the choice prompt for a
// "mana of any color/type that a land ... could produce" ability of the given
// scope. It panics on any unsupported relation so an over-broad caller fails
// loudly rather than emitting a mislabeled ability.
func landsProduceTexts(relation PlayerRelation, includeColorless bool) (text, prompt string) {
	kind := "color"
	promptKind := "color"
	if includeColorless {
		kind = "type"
		promptKind = "type of mana"
	}
	switch relation {
	case PlayerYou:
		return fmt.Sprintf("{T}: Add one mana of any %s that a land you control could produce.", kind),
			fmt.Sprintf("Choose a %s a land you control could produce", promptKind)
	case PlayerOpponent:
		return fmt.Sprintf("{T}: Add one mana of any %s that a land an opponent controls could produce.", kind),
			fmt.Sprintf("Choose a %s a land an opponent controls could produce", promptKind)
	default:
		panic(fmt.Sprintf("game: unsupported lands-produce mana scope %d", relation))
	}
}

// triggerLandProducedManaKey publishes the type chosen for a "one mana of any
// type that land produced" mana-doubler trigger (Mirari's Wake, Zendikar
// Resurgent; see TriggerLandProducedManaContent).
const triggerLandProducedManaKey = ChoiceKey("oracle-trigger-land-produced-type")

// TriggerLandProducedManaContent builds the triggered-ability content of a "add
// one mana of any type that land produced" mana doubler (Mirari's Wake, Zendikar
// Resurgent). The controller chooses one of the types the land that fired the
// tapped-for-mana trigger produced on that tap and adds one mana of it. The
// candidate types are recomputed at resolution from the triggering tap's
// TriggerEvent.ProducedManaColors; an empty set produces no mana (CR 605.1a).
func TriggerLandProducedManaContent() AbilityContent {
	return Mode{Sequence: []Instruction{
		{
			Primitive: Choose{
				Choice: ResolutionChoice{
					Kind:        ResolutionChoiceMana,
					Prompt:      "Choose a type of mana that land produced",
					ColorSource: ResolutionChoiceColorSourceTriggerLandProduced,
				},
				PublishChoice: triggerLandProducedManaKey,
			},
		},
		{
			Primitive: AddMana{
				Amount:     Fixed(1),
				ChoiceFrom: triggerLandProducedManaKey,
			},
		},
	}}.Ability()
}

// TapManaAmongControlledColorsAbility builds the complete "{T}: Add one mana of
// any color among <permanents> you control." mana ability (Mox Amber, Plaza of
// Heroes). text is the exact oracle text. selection describes which permanents
// the controller controls contribute their colors; the choosable colors are
// recomputed at resolution as the union of the matching permanents' colors. When
// no matching permanent is colored the choice is empty and the ability is
// unactivatable (CR 605.1a).
func TapManaAmongControlledColorsAbility(text string, selection Selection) ManaAbility {
	return ManaAbility{
		Text:            text,
		AdditionalCosts: cost.Tap,
		Content: Mode{Sequence: []Instruction{
			{
				Primitive: Choose{
					Choice: ResolutionChoice{
						Kind:        ResolutionChoiceMana,
						Prompt:      "Choose a color among permanents you control",
						ColorSource: ResolutionChoiceColorSourceControlledPermanentColors,
						Selection:   &selection,
					},
					PublishChoice: tapManaAmongControlledColorsKey,
				},
			},
			{
				Primitive: AddMana{
					Amount:     Fixed(1),
					ChoiceFrom: tapManaAmongControlledColorsKey,
				},
			},
		}}.Ability(),
	}
}

// TapManaEachControlledColorAbility builds the complete "{T}: For each color
// among <permanents> you control, add one mana of that color." mana ability
// (Bloom Tender). text is the exact oracle text. selection describes which
// permanents the controller controls contribute their colors; one mana of each
// color in the union of the matching permanents' colors is produced at
// resolution. When no matching permanent is colored no mana is produced and the
// ability is unactivatable (CR 605.1a).
func TapManaEachControlledColorAbility(text string, selection Selection) ManaAbility {
	return ManaAbility{
		Text:            text,
		AdditionalCosts: cost.Tap,
		Content: Mode{Sequence: []Instruction{
			{
				Primitive: AddMana{
					Amount:              Fixed(1),
					EachControlledColor: &selection,
				},
			},
		}}.Ability(),
	}
}

// TwoColorFilterManaAbility builds the activated mana ability shared by the
// "filter land" cycle (Mystic Gate, Sunken Ruins, Fetid Heath, Cascade Bluffs,
// Rugged Prairie, Graven Cairns, Twilight Mire, Wooded Bastion, Fire-Lit
// Thicket, and Flooded Grove). Their second ability reads
// "{X/Y}, {T}: Add {X}{X}, {X}{Y}, or {Y}{Y}.": paying one hybrid {X/Y} mana and
// tapping the land adds two mana, each independently either color of the fixed
// pair. The three printed combinations {X}{X}, {X}{Y}, and {Y}{Y} are exactly
// the unordered two-mana multisets over {X, Y}, so two independent color choices
// over the pair reproduce the printed output faithfully. The two choices publish
// under distinct keys so the instruction sequence is valid.
func TwoColorFilterManaAbility(first, second mana.Color) ManaAbility {
	validateFilterManaPair(first, second)
	firstSymbol := manaSymbol(first)
	secondSymbol := manaSymbol(second)
	return ManaAbility{
		Text: fmt.Sprintf(
			"{%s/%s}, {T}: Add {%s}{%s}, {%s}{%s}, or {%s}{%s}.",
			firstSymbol, secondSymbol,
			firstSymbol, firstSymbol,
			firstSymbol, secondSymbol,
			secondSymbol, secondSymbol,
		),
		ManaCost:        opt.Val(cost.Mana{cost.HybridMana(first, second)}),
		AdditionalCosts: cost.Tap,
		Content: Mode{Sequence: []Instruction{
			{
				Primitive: Choose{
					Choice: ResolutionChoice{
						Kind:   ResolutionChoiceMana,
						Prompt: "Choose a color",
						Colors: []mana.Color{first, second},
					},
					PublishChoice: tapManaFilterFirstKey,
				},
			},
			{
				Primitive: AddMana{
					Amount:     Fixed(1),
					ChoiceFrom: tapManaFilterFirstKey,
				},
			},
			{
				Primitive: Choose{
					Choice: ResolutionChoice{
						Kind:   ResolutionChoiceMana,
						Prompt: "Choose a color",
						Colors: []mana.Color{first, second},
					},
					PublishChoice: tapManaFilterSecondKey,
				},
			},
			{
				Primitive: AddMana{
					Amount:     Fixed(1),
					ChoiceFrom: tapManaFilterSecondKey,
				},
			},
		}}.Ability(),
	}
}

// validateFilterManaPair panics unless first and second are two distinct basic
// colors (W, U, B, R, or G), the only inputs the filter-land output body admits.
func validateFilterManaPair(first, second mana.Color) {
	for _, manaColor := range []mana.Color{first, second} {
		switch manaColor {
		case mana.W, mana.U, mana.B, mana.R, mana.G:
		default:
			panic(fmt.Sprintf("game: invalid filter mana color %q", manaColor))
		}
	}
	if first == second {
		panic(fmt.Sprintf("game: filter mana pair requires two distinct colors, got %q twice", first))
	}
}

func validateManaColorChoice(colors []mana.Color) {
	if len(colors) < 2 || len(colors) > 6 {
		panic("game: tap mana choice requires two through six mana types")
	}
	seen := make(map[mana.Color]struct{}, len(colors))
	for _, manaColor := range colors {
		switch manaColor {
		case mana.W, mana.U, mana.B, mana.R, mana.G, mana.C:
		default:
			panic(fmt.Sprintf("game: invalid mana color choice %q", manaColor))
		}
		if _, ok := seen[manaColor]; ok {
			panic(fmt.Sprintf("game: duplicate mana color choice %q", manaColor))
		}
		seen[manaColor] = struct{}{}
	}
}

func tapManaChoiceText(colors []mana.Color) string {
	if len(colors) == 5 &&
		colors[0] == mana.W &&
		colors[1] == mana.U &&
		colors[2] == mana.B &&
		colors[3] == mana.R &&
		colors[4] == mana.G {
		return "{T}: Add one mana of any color."
	}
	symbols := make([]string, len(colors))
	for i, manaColor := range colors {
		symbols[i] = fmt.Sprintf("{%s}", manaSymbol(manaColor))
	}
	if len(symbols) == 2 {
		return fmt.Sprintf("{T}: Add %s or %s.", symbols[0], symbols[1])
	}
	return fmt.Sprintf(
		"{T}: Add %s, or %s.",
		strings.Join(symbols[:len(symbols)-1], ", "),
		symbols[len(symbols)-1],
	)
}

func containsManaColor(colors []mana.Color, want mana.Color) bool {
	return slices.Contains(colors, want)
}

// livingWeaponGermToken is the canonical 0/0 black Phyrexian Germ creature token
// created by the Living weapon keyword (CR 702.91).
var livingWeaponGermToken = &CardDef{
	CardFace: CardFace{
		Name:      "Germ",
		Types:     []types.Card{types.Creature},
		Subtypes:  []types.Sub{types.Phyrexian, types.Germ},
		Colors:    []color.Color{color.Black},
		Power:     opt.Val(PT{Value: 0}),
		Toughness: opt.Val(PT{Value: 0}),
	},
}

// livingWeaponGermLinkKey links the freshly created Germ token to the subsequent
// self-attach so Living weapon attaches the Equipment to that exact token.
const livingWeaponGermLinkKey = LinkedKey("living-weapon-germ")

// LivingWeaponTriggeredAbility builds the entry trigger for Living weapon
// (CR 702.91): when this Equipment enters, create a 0/0 black Phyrexian Germ
// creature token, then attach this Equipment to it.
func LivingWeaponTriggeredAbility() TriggeredAbility {
	return TriggeredAbility{
		Text: "Living weapon",
		Trigger: TriggerCondition{
			Type: TriggerWhen,
			Pattern: TriggerPattern{
				Event:  EventPermanentEnteredBattlefield,
				Source: TriggerSourceSelf,
			},
		},
		KeywordAbilities: []KeywordAbility{
			SimpleKeyword{Kind: LivingWeapon},
		},
		Content: Mode{Sequence: []Instruction{
			{
				Primitive: CreateToken{
					Amount:        Fixed(1),
					Source:        TokenDef(livingWeaponGermToken),
					PublishLinked: livingWeaponGermLinkKey,
				},
			},
			{
				Primitive: Attach{
					Attachment: SourcePermanentReference(),
					Target:     LinkedObjectReference(string(livingWeaponGermLinkKey)),
				},
			},
		}}.Ability(),
	}
}

// SoulshiftTriggeredAbility builds the canonical Soulshift N triggered ability
// (CR 702.46): when this creature dies, its controller may return a target
// Spirit card with mana value N or less from their graveyard to their hand.
func SoulshiftTriggeredAbility(n int) TriggeredAbility {
	return TriggeredAbility{
		Text: fmt.Sprintf("Soulshift %d", n),
		Trigger: TriggerCondition{
			Type: TriggerWhen,
			Pattern: TriggerPattern{
				Event:            EventPermanentDied,
				Source:           TriggerSourceSelf,
				SubjectSelection: Selection{RequiredTypes: []types.Card{types.Creature}},
			},
		},
		Optional: true,
		KeywordAbilities: []KeywordAbility{
			SoulshiftKeyword{Count: n},
		},
		Content: Mode{
			Targets: []TargetSpec{{
				MinTargets: 1,
				MaxTargets: 1,
				Constraint: fmt.Sprintf("target Spirit card with mana value %d or less from your graveyard", n),
				Allow:      TargetAllowCard,
				TargetZone: zone.Graveyard,
				Selection: opt.Val(Selection{
					SubtypesAny: []types.Sub{types.Sub("Spirit")},
					Controller:  ControllerYou,
					ManaValue:   opt.Val(compare.Int{Op: compare.LessOrEqual, Value: n}),
				}),
			}},
			Sequence: []Instruction{{
				Primitive: MoveCard{
					Card:        CardReference{Kind: CardReferenceTarget},
					FromZone:    zone.Graveyard,
					Destination: zone.Hand,
				},
			}},
		}.Ability(),
	}
}
