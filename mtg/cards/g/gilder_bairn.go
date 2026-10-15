package g

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// GilderBairn is the card definition for Gilder Bairn.
//
// Type: Creature — Ouphe
// Cost: {1}{G/U}{G/U}
//
// Oracle text:
//
//	{2}{G/U}, {Q}: Double the number of each kind of counter on target permanent. ({Q} is the untap symbol.)
var GilderBairn = newGilderBairn

func newGilderBairn() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue, color.Green),
		CardFace: game.CardFace{
			Name: "Gilder Bairn",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
				cost.HybridMana(mana.G, mana.U),
				cost.HybridMana(mana.G, mana.U),
			}),
			Colors:    []color.Color{color.Green, color.Blue},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Ouphe},
			Power:     opt.Val(game.PT{Value: 1}),
			Toughness: opt.Val(game.PT{Value: 3}),
			ActivatedAbilities: []game.ActivatedAbility{
				game.ActivatedAbility{
					Text: "{2}{G/U}, {Q}: Double the number of each kind of counter on target permanent. ({Q} is the untap symbol.)",
					ManaCost: opt.Val(cost.Mana{
						cost.O(2),
						cost.HybridMana(mana.G, mana.U),
					}),
					AdditionalCosts: []cost.Additional{
						cost.Additional{
							Kind:        cost.AdditionalUntap,
							Text:        "{Q}",
							CounterKind: counter.PlusOnePlusOne,
						},
					},
					ZoneOfFunction: zone.Battlefield,
					Content: game.AbilityContent{
						Modes: []game.Mode{
							game.Mode{
								Targets: []game.TargetSpec{
									game.TargetSpec{
										MinTargets: 1,
										MaxTargets: 1,
										Constraint: "target permanent",
										Allow:      game.TargetAllowPermanent,
									},
								},
								Sequence: []game.Instruction{
									game.Instruction{
										Primitive: game.AddCounter{
											Object:      game.TargetPermanentReference(0),
											CounterKind: counter.PlusOnePlusOne,
											AllKinds:    true,
										},
									},
								},
							},
						},
						MinModes: 1,
						MaxModes: 1,
					},
				},
			},
			OracleText: `
			{2}{G/U}, {Q}: Double the number of each kind of counter on target permanent. ({Q} is the untap symbol.)
		`,
		},
	}
}
