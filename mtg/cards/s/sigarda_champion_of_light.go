package s

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/compare"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// SigardaChampionOfLight is the card definition for Sigarda, Champion of Light.
//
// Type: Legendary Creature — Angel
// Cost: {1}{G}{W}{W}
//
// Oracle text:
//
//	Flying, trample
//	Humans you control get +1/+1.
//	Coven — Whenever Sigarda attacks, if you control three or more creatures with different powers, look at the top five cards of your library. You may reveal a Human creature card from among them and put it into your hand. Put the rest on the bottom of your library in a random order.
var SigardaChampionOfLight = newSigardaChampionOfLight

func newSigardaChampionOfLight() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White, color.Green),
		CardFace: game.CardFace{
			Name: "Sigarda, Champion of Light",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
				cost.G,
				cost.W,
				cost.W,
			}),
			Colors:     []color.Color{color.Green, color.White},
			Supertypes: []types.Super{types.Legendary},
			Types:      []types.Card{types.Creature},
			Subtypes:   []types.Sub{types.Angel},
			Power:      opt.Val(game.PT{Value: 4}),
			Toughness:  opt.Val(game.PT{Value: 4}),
			StaticAbilities: []game.StaticAbility{
				game.FlyingStaticBody,
				game.TrampleStaticBody,
				game.StaticAbility{
					ContinuousEffects: []game.ContinuousEffect{
						game.ContinuousEffect{
							Layer:          game.LayerPowerToughnessModify,
							Group:          game.ObjectControlledGroup(game.SourcePermanentReference(), game.Selection{SubtypesAny: []types.Sub{types.Sub("Human")}}),
							PowerDelta:     1,
							ToughnessDelta: 1,
						},
					},
				},
			},
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Trigger: game.TriggerCondition{
						Type: game.TriggerWhenever,
						Pattern: game.TriggerPattern{
							Event:  game.EventAttackerDeclared,
							Source: game.TriggerSourceSelf,
						},
						InterveningIf: "if you control three or more creatures with different powers",
						InterveningCondition: opt.Val(game.Condition{
							Aggregates: []game.AggregateComparison{{Aggregate: game.AggregateControllerCreaturePowerDiversity, Op: compare.GreaterOrEqual, Value: 3}},
						}),
					},
					Content: game.Mode{
						Sequence: []game.Instruction{
							{
								Primitive: game.Dig{
									Player:    game.ControllerReference(),
									Look:      game.Fixed(5),
									Take:      game.Fixed(1),
									Remainder: game.DigRemainderLibraryBottom,
									Filter:    opt.Val(game.Selection{RequiredTypes: []types.Card{types.Creature}, SubtypesAny: []types.Sub{types.Sub("Human")}}),
									TakeUpTo:  true,
									Reveal:    true,
								},
							},
						},
					}.Ability(),
				},
			},
			OracleText: `
			Flying, trample
			Humans you control get +1/+1.
			Coven — Whenever Sigarda attacks, if you control three or more creatures with different powers, look at the top five cards of your library. You may reveal a Human creature card from among them and put it into your hand. Put the rest on the bottom of your library in a random order.
		`,
		},
	}
}
