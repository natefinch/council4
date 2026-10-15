package t

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// TreasonousOgre is the card definition for Treasonous Ogre.
//
// Type: Creature — Ogre Shaman
// Cost: {3}{R}
//
// Oracle text:
//
//	Dethrone (Whenever this creature attacks the player with the most life or tied for most life, put a +1/+1 counter on it.)
//	Pay 3 life: Add {R}.
var TreasonousOgre = newTreasonousOgre

func newTreasonousOgre() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Red),
		CardFace: game.CardFace{
			Name: "Treasonous Ogre",
			ManaCost: opt.Val(cost.Mana{
				cost.O(3),
				cost.R,
			}),
			Colors:    []color.Color{color.Red},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Ogre, types.Shaman},
			Power:     opt.Val(game.PT{Value: 2}),
			Toughness: opt.Val(game.PT{Value: 3}),
			ManaAbilities: []game.ManaAbility{
				game.ManaAbility{
					Text: "Pay 3 life: Add {R}.",
					AdditionalCosts: []cost.Additional{
						cost.Additional{
							Kind:        cost.AdditionalPayLife,
							Text:        "Pay 3 life",
							Amount:      3,
							CounterKind: counter.PlusOnePlusOne,
						},
					},
					Content: game.AbilityContent{
						Modes: []game.Mode{
							game.Mode{
								Sequence: []game.Instruction{
									game.Instruction{
										Primitive: game.AddMana{
											Amount:    game.Fixed(1),
											ManaColor: mana.R,
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
			TriggeredAbilities: []game.TriggeredAbility{
				game.DethroneTriggeredBody,
			},
			OracleText: `
			Dethrone (Whenever this creature attacks the player with the most life or tied for most life, put a +1/+1 counter on it.)
			Pay 3 life: Add {R}.
		`,
		},
	}
}
