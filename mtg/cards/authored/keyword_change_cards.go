package authored

import (
	"github.com/natefinch/council4/mtg/cards/cardset"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

const (
	fearOfFallingOracleID = "00ba96e0-42a7-432a-a500-5e607d75a358"
	shortCircuitOracleID  = "5ea0e3a5-9c38-4da4-98a6-41c14be512d9"
	canopyDragonOracleID  = "a9e59ee4-02a6-498c-82f0-f4f2ff334d65"

	fearOfFallingAttackText = "Whenever this creature attacks, target creature defending player controls gets -2/-0 and loses flying until your next turn."
	shortCircuitText        = "Flash\nEnchant artifact or creature\nAs long as enchanted permanent is a creature, it gets -3/-0 and loses flying."
	canopyDragonAbilityText = "{1}{G}: This creature gains flying and loses trample until end of turn."
)

var (
	FearOfFalling = cardset.Entry{
		OracleID: fearOfFallingOracleID,
		Name:     "Fear of Falling",
		New:      newFearOfFalling,
		Verification: "mtg/rules/keyword_change_cards_test.go: " +
			"TestFearOfFallingAttacksDefendingPlayerAndRestoresFlyingNextTurn",
	}

	ShortCircuit = cardset.Entry{
		OracleID: shortCircuitOracleID,
		Name:     "Short Circuit",
		New:      newShortCircuit,
		Verification: "mtg/rules/keyword_change_cards_test.go: " +
			"TestShortCircuitEnchantsArtifactsAndCreaturesConditionally",
	}

	CanopyDragon = cardset.Entry{
		OracleID: canopyDragonOracleID,
		Name:     "Canopy Dragon",
		New:      newCanopyDragon,
		Verification: "mtg/rules/keyword_change_cards_test.go: " +
			"TestCanopyDragonGainsFlyingAndLosesTrampleUntilEndOfTurn",
	}
)

func newFearOfFalling() *game.CardDef {
	return &game.CardDef{CardFace: game.CardFace{
		Name:       "Fear of Falling",
		ManaCost:   opt.Val(cost.Mana{cost.O(3), cost.U, cost.U}),
		Types:      []types.Card{types.Enchantment, types.Creature},
		Subtypes:   []types.Sub{types.Nightmare},
		Colors:     []color.Color{color.Blue},
		Power:      opt.Val(game.PT{Value: 4}),
		Toughness:  opt.Val(game.PT{Value: 4}),
		OracleText: "Flying\n" + fearOfFallingAttackText,
		StaticAbilities: []game.StaticAbility{
			game.FlyingStaticBody,
		},
		TriggeredAbilities: []game.TriggeredAbility{{
			Text: fearOfFallingAttackText,
			Trigger: game.TriggerCondition{
				Type: game.TriggerWhenever,
				Pattern: game.TriggerPattern{
					Event:  game.EventAttackerDeclared,
					Source: game.TriggerSourceSelf,
				},
			},
			Content: game.Mode{
				Targets: []game.TargetSpec{{
					MinTargets: 1,
					MaxTargets: 1,
					Constraint: "target creature defending player controls",
					Allow:      game.TargetAllowPermanent,
					Selection: opt.Val(game.Selection{
						RequiredTypes:               []types.Card{types.Creature},
						ControlledByDefendingPlayer: true,
					}),
				}},
				Sequence: []game.Instruction{{
					Primitive: game.ApplyContinuous{
						Object: opt.Val(game.TargetPermanentReference(0)),
						ContinuousEffects: []game.ContinuousEffect{
							{Layer: game.LayerPowerToughnessModify, PowerDelta: -2},
							{Layer: game.LayerAbility, RemoveKeywords: []game.Keyword{game.Flying}},
						},
						Duration: game.DurationUntilYourNextTurn,
					},
				}},
			}.Ability(),
		}},
	}}
}

func newShortCircuit() *game.CardDef {
	attached := game.AttachedObjectGroup(game.SourcePermanentReference())
	enchantedCreature := game.Condition{
		Object:        opt.Val(game.SourceAttachedPermanentReference()),
		ObjectMatches: opt.Val(game.Selection{RequiredTypes: []types.Card{types.Creature}}),
	}
	return &game.CardDef{CardFace: game.CardFace{
		Name:       "Short Circuit",
		ManaCost:   opt.Val(cost.Mana{cost.O(1), cost.U}),
		Types:      []types.Card{types.Enchantment},
		Subtypes:   []types.Sub{types.Aura},
		Colors:     []color.Color{color.Blue},
		OracleText: shortCircuitText,
		StaticAbilities: []game.StaticAbility{
			game.FlashStaticBody,
			{
				Text: "Enchant artifact or creature",
				KeywordAbilities: []game.KeywordAbility{game.EnchantKeyword{
					Target: game.TargetSpec{
						MinTargets: 1,
						MaxTargets: 1,
						Constraint: "target artifact or creature",
						Allow:      game.TargetAllowPermanent,
						Selection: opt.Val(game.Selection{
							RequiredTypesAny: []types.Card{types.Artifact, types.Creature},
						}),
					},
				}},
			},
			{
				Text:      "As long as enchanted permanent is a creature, it gets -3/-0 and loses flying.",
				Condition: opt.Val(enchantedCreature),
				ContinuousEffects: []game.ContinuousEffect{
					{Layer: game.LayerPowerToughnessModify, Group: attached, PowerDelta: -3},
					{Layer: game.LayerAbility, Group: attached, RemoveKeywords: []game.Keyword{game.Flying}},
				},
			},
		},
	}}
}

func newCanopyDragon() *game.CardDef {
	return &game.CardDef{CardFace: game.CardFace{
		Name:       "Canopy Dragon",
		ManaCost:   opt.Val(cost.Mana{cost.O(4), cost.G, cost.G}),
		Types:      []types.Card{types.Creature},
		Subtypes:   []types.Sub{types.Dragon},
		Colors:     []color.Color{color.Green},
		Power:      opt.Val(game.PT{Value: 4}),
		Toughness:  opt.Val(game.PT{Value: 4}),
		OracleText: "Trample\n" + canopyDragonAbilityText,
		StaticAbilities: []game.StaticAbility{
			game.TrampleStaticBody,
		},
		ActivatedAbilities: []game.ActivatedAbility{{
			Text:           canopyDragonAbilityText,
			ManaCost:       opt.Val(cost.Mana{cost.O(1), cost.G}),
			ZoneOfFunction: zone.Battlefield,
			Content: game.Mode{Sequence: []game.Instruction{{
				Primitive: game.ApplyContinuous{
					Object: opt.Val(game.SourcePermanentReference()),
					ContinuousEffects: []game.ContinuousEffect{
						{Layer: game.LayerAbility, AddKeywords: []game.Keyword{game.Flying}},
						{Layer: game.LayerAbility, RemoveKeywords: []game.Keyword{game.Trample}},
					},
					Duration: game.DurationUntilEndOfTurn,
				},
			}}}.Ability(),
		}},
	}}
}
