package u

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// UrzaSFactory is the card definition for Urza's Factory.
//
// Type: Land — Urza's
//
// Oracle text:
//
//	{T}: Add {C}.
//	{7}, {T}: Create a 2/2 colorless Assembly-Worker artifact creature token.
var UrzaSFactory = newUrzaSFactory

func newUrzaSFactory() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name:     "Urza's Factory",
			Types:    []types.Card{types.Land},
			Subtypes: []types.Sub{types.Sub("Urza's")},
			ActivatedAbilities: []game.ActivatedAbility{
				game.ActivatedAbility{
					Text: "{7}, {T}: Create a 2/2 colorless Assembly-Worker artifact creature token.",
					ManaCost: opt.Val(cost.Mana{
						cost.O(7),
					}),
					AdditionalCosts: []cost.Additional{
						cost.Additional{
							Kind:        cost.AdditionalTap,
							CounterKind: counter.PlusOnePlusOne,
						},
					},
					ZoneOfFunction: zone.Battlefield,
					Content: game.AbilityContent{
						Modes: []game.Mode{
							game.Mode{
								Sequence: []game.Instruction{
									game.Instruction{
										Primitive: game.CreateToken{
											Amount: game.Fixed(1),
											Source: game.TokenDef(urzaSFactoryToken),
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
			ManaAbilities: []game.ManaAbility{
				game.TapManaAbility(mana.C),
			},
			OracleText: `
			{T}: Add {C}.
			{7}, {T}: Create a 2/2 colorless Assembly-Worker artifact creature token.
		`,
		},
	}
}

var urzaSFactoryToken = newUrzaSFactoryToken()

func newUrzaSFactoryToken() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name:      "Assembly-Worker",
			Types:     []types.Card{types.Artifact, types.Creature},
			Subtypes:  []types.Sub{types.AssemblyWorker},
			Power:     opt.Val(game.PT{Value: 2}),
			Toughness: opt.Val(game.PT{Value: 2}),
		},
	}
}
