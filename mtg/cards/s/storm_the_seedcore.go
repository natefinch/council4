package s

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// StormTheSeedcore is the card definition for Storm the Seedcore.
//
// Type: Sorcery
// Cost: {2}{G}{G}
//
// Oracle text:
//
//	Distribute four +1/+1 counters among up to four target creatures you control. Creatures you control gain vigilance and trample until end of turn.
var StormTheSeedcore = newStormTheSeedcore

func newStormTheSeedcore() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Green),
		CardFace: game.CardFace{
			Name: "Storm the Seedcore",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.G,
				cost.G,
			}),
			Colors: []color.Color{color.Green},
			Types:  []types.Card{types.Sorcery},
			SpellAbility: opt.Val(game.Mode{
				Targets: []game.TargetSpec{
					game.TargetSpec{
						MinTargets: 0,
						MaxTargets: 4,
						Constraint: "up to four target creatures you control",
						Allow:      game.TargetAllowPermanent,
						Selection:  opt.Val(game.Selection{RequiredTypesAny: []types.Card{types.Creature}, Controller: game.ControllerYou}),
					},
				},
				Sequence: []game.Instruction{
					{
						Primitive: game.AddCounter{
							Amount:      game.Fixed(4),
							Object:      game.AllTargetPermanentsReference(0),
							CounterKind: counter.PlusOnePlusOne,
							Distribute:  true,
						},
					},
					{
						Primitive: game.ApplyContinuous{
							ContinuousEffects: []game.ContinuousEffect{
								game.ContinuousEffect{
									Layer: game.LayerAbility,
									Group: game.BattlefieldGroup(game.Selection{RequiredTypes: []types.Card{types.Creature}, Controller: game.ControllerYou}),
									AddKeywords: []game.Keyword{
										game.Vigilance,
										game.Trample,
									},
								},
							},
							Duration: game.DurationUntilEndOfTurn,
						},
					},
				},
			}.Ability()),
			OracleText: `
			Distribute four +1/+1 counters among up to four target creatures you control. Creatures you control gain vigilance and trample until end of turn.
		`,
		},
	}
}
