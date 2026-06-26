package o

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// OrcGeneral is the card definition for Orc General.
//
// Type: Creature — Orc Warrior
// Cost: {2}{R}
//
// Oracle text:
//
//	{T}, Sacrifice another Orc or Goblin: Other Orc creatures get +1/+1 until end of turn.
var OrcGeneral = newOrcGeneral()

func newOrcGeneral() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Red),
		CardFace: game.CardFace{
			Name: "Orc General",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.R,
			}),
			Colors:    []color.Color{color.Red},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Orc, types.Warrior},
			Power:     opt.Val(game.PT{Value: 2}),
			Toughness: opt.Val(game.PT{Value: 2}),
			ActivatedAbilities: []game.ActivatedAbility{
				game.ActivatedAbility{
					Text: "{T}, Sacrifice another Orc or Goblin: Other Orc creatures get +1/+1 until end of turn.",
					AdditionalCosts: []cost.Additional{
						{
							Kind: cost.AdditionalTap,
						},
						{
							Kind:          cost.AdditionalSacrifice,
							Text:          "Sacrifice another Orc or Goblin",
							Amount:        1,
							ExcludeSource: true,
							SubtypesAny:   cost.SubtypeSet{types.Orc, types.Goblin},
						},
					},
					ZoneOfFunction: zone.Battlefield,
					Content: game.Mode{
						Sequence: []game.Instruction{
							{
								Primitive: game.ApplyContinuous{
									ContinuousEffects: []game.ContinuousEffect{
										game.ContinuousEffect{
											Layer:          game.LayerPowerToughnessModify,
											Group:          game.BattlefieldGroupExcluding(game.Selection{SubtypesAny: []types.Sub{types.Sub("Orc")}}, game.SourcePermanentReference()),
											PowerDelta:     1,
											ToughnessDelta: 1,
										},
									},
									Duration: game.DurationUntilEndOfTurn,
								},
							},
						},
					}.Ability(),
				},
			},
			OracleText: `
			{T}, Sacrifice another Orc or Goblin: Other Orc creatures get +1/+1 until end of turn.
		`,
		},
	}
}
