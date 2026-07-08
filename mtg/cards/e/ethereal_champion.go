package e

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// EtherealChampion is the card definition for Ethereal Champion.
//
// Type: Creature — Avatar
// Cost: {2}{W}{W}{W}
//
// Oracle text:
//
//	Pay 1 life: Prevent the next 1 damage that would be dealt to this creature this turn.
var EtherealChampion = newEtherealChampion

func newEtherealChampion() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White),
		CardFace: game.CardFace{
			Name: "Ethereal Champion",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.W,
				cost.W,
				cost.W,
			}),
			Colors:    []color.Color{color.White},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Avatar},
			Power:     opt.Val(game.PT{Value: 3}),
			Toughness: opt.Val(game.PT{Value: 4}),
			ActivatedAbilities: []game.ActivatedAbility{
				game.ActivatedAbility{
					Text: "Pay 1 life: Prevent the next 1 damage that would be dealt to this creature this turn.",
					AdditionalCosts: []cost.Additional{
						{
							Kind:   cost.AdditionalPayLife,
							Text:   "Pay 1 life",
							Amount: 1,
						},
					},
					ZoneOfFunction: zone.Battlefield,
					Content: game.Mode{
						Sequence: []game.Instruction{
							{
								Primitive: game.PreventDamage{
									Object: game.SourcePermanentReference(),
									Amount: game.Fixed(1),
								},
							},
						},
					}.Ability(),
				},
			},
			OracleText: `
			Pay 1 life: Prevent the next 1 damage that would be dealt to this creature this turn.
		`,
		},
	}
}
