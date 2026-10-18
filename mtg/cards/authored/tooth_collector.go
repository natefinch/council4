package authored

import (
	"github.com/natefinch/council4/mtg/cardprogram"
	"github.com/natefinch/council4/mtg/cards/cardset"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/compare"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

const toothCollectorOracleID = "b7c165a4-53a2-4117-a452-b9a8fe309c09"

var toothCollectorEnter = game.ProgramReference{ID: toothCollectorOracleID + "/enter/0", Version: "1"}
var toothCollectorDelirium = game.ProgramReference{ID: toothCollectorOracleID + "/upkeep/0", Version: "1"}

// Source review: Oracle printing revision febc9f63-a225-438d-b835-07a561ba91f5
// (2016-04-08); its Scryfall rulings URI returned five WotC rulings, all dated
// 2016-04-08. They require counting distinct card types (not cards,
// supertypes, or subtypes), using only a double-faced card's front face, and
// excluding tokens and spell copies; delirium's intervening-if condition is
// checked both when the ability triggers and when it resolves.
// Rules read: verified official Comprehensive Rules effective 2026-09-25,
// sections 111.6, 112.1a, 113.7a, 205.2a, 207.2c, 400.7, 501.1, 601.2c,
// 603.2b, 603.3d, 603.4, 608.2b, 611.2a, 613.4c, 704.5f, 707.10, and 712.8.
// These govern card types and non-card tokens/copies, independent trigger
// resolution, upkeep timing and target announcement/revalidation, the two
// delirium checks, temporary layer-7c P/T changes, and zero-toughness state-
// based actions. The verified-copy path and hash are in the implementation
// handoff.
var ToothCollector = cardset.Entry{
	OracleID: toothCollectorOracleID, Name: "Tooth Collector", New: newToothCollector,
	Programs: []cardprogram.Program{
		{Reference: toothCollectorEnter, Resolve: resolveToothCollectorEnter},
		{Reference: toothCollectorDelirium, Resolve: resolveToothCollectorDelirium},
	},
	Verification: "mtg/rules/tooth_collector_test.go: TestToothCollectorEnterTriggerAppliesMinusOneMinusOneToOpponentCreature; TestToothCollectorDeliriumRequiresFourDistinctCardTypes; TestToothCollectorDeliriumRechecksAtResolution; TestToothCollectorDeliriumTargetsEventPlayerAndResolvesAfterSourceLeaves; TestToothCollectorDeliriumTargetMustRemainControlledByEventPlayer; TestToothCollectorBothAbilitiesExpireAtCleanup",
}

const toothCollectorEnterText = "When this creature enters, target creature an opponent controls gets -1/-1 until end of turn."
const toothCollectorDeliriumText = "Delirium — At the beginning of each opponent's upkeep, if there are four or more card types among cards in your graveyard, target creature that player controls gets -1/-1 until end of turn."

func newToothCollector() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Black),
		CardFace: game.CardFace{
			Name: "Tooth Collector", OracleText: toothCollectorEnterText + "\n" + toothCollectorDeliriumText,
			ManaCost: opt.Val(cost.Mana{cost.O(2), cost.B}),
			Types:    []types.Card{types.Creature}, Subtypes: []types.Sub{types.Human, types.Rogue},
			Colors: []color.Color{color.Black}, Power: opt.Val(game.PT{Value: 3}),
			Toughness: opt.Val(game.PT{Value: 2}),
			TriggeredAbilities: []game.TriggeredAbility{
				{
					Trigger: game.TriggerCondition{Type: game.TriggerWhen, Pattern: game.TriggerPattern{
						Event: game.EventPermanentEnteredBattlefield, Source: game.TriggerSourceSelf,
					}},
					Content: game.Mode{
						Text: toothCollectorEnterText,
						Targets: []game.TargetSpec{{
							MinTargets: 1, MaxTargets: 1, Constraint: "target creature an opponent controls",
							Allow: game.TargetAllowPermanent,
							Selection: opt.Val(game.Selection{
								RequiredTypes: []types.Card{types.Creature}, Controller: game.ControllerOpponent,
							}),
						}},
						Program: toothCollectorEnter,
					}.Ability(),
				},
				{
					Trigger: game.TriggerCondition{
						Type: game.TriggerAt,
						Pattern: game.TriggerPattern{
							Event: game.EventBeginningOfStep, Step: game.StepUpkeep,
							Player: game.TriggerPlayerOpponent,
						},
						InterveningIf: "if there are four or more card types among cards in your graveyard",
						InterveningCondition: opt.Val(game.Condition{
							Text: "if there are four or more card types among cards in your graveyard",
							Aggregates: []game.AggregateComparison{{
								Aggregate: game.AggregateControllerGraveyardCardTypeCount,
								Op:        compare.GreaterOrEqual,
								Value:     4,
							}},
						}),
					},
					Content: game.Mode{
						Text: toothCollectorDeliriumText,
						Targets: []game.TargetSpec{{
							MinTargets: 1, MaxTargets: 1, Constraint: "target creature that player controls",
							Allow: game.TargetAllowPermanent,
							Selection: opt.Val(game.Selection{
								RequiredTypes: []types.Card{types.Creature}, ControlledByEventPlayer: true,
							}),
						}},
						Program: toothCollectorDelirium,
					}.Ability(),
				},
			},
		},
	}
}

func resolveToothCollectorEnter(ctx cardprogram.Context) error {
	_, err := ctx.Execute(game.ModifyPT{
		Object:         game.TargetPermanentReference(0),
		PowerDelta:     game.Fixed(-1),
		ToughnessDelta: game.Fixed(-1),
		Duration:       game.DurationUntilEndOfTurn,
	})
	return err
}

func resolveToothCollectorDelirium(ctx cardprogram.Context) error {
	_, err := ctx.Execute(game.ModifyPT{
		Object:         game.TargetPermanentReference(0),
		PowerDelta:     game.Fixed(-1),
		ToughnessDelta: game.Fixed(-1),
		Duration:       game.DurationUntilEndOfTurn,
	})
	return err
}
