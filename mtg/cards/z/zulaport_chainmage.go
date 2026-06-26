package z

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// ZulaportChainmage is the card definition for Zulaport Chainmage.
//
// Type: Creature — Human Shaman Ally
// Cost: {3}{B}
//
// Oracle text:
//
//	Cohort — {T}, Tap an untapped Ally you control: Target opponent loses 2 life.
var ZulaportChainmage = newZulaportChainmage()

func newZulaportChainmage() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Black),
		CardFace: game.CardFace{
			Name: "Zulaport Chainmage",
			ManaCost: opt.Val(cost.Mana{
				cost.O(3),
				cost.B,
			}),
			Colors:    []color.Color{color.Black},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Human, types.Shaman, types.Ally},
			Power:     opt.Val(game.PT{Value: 4}),
			Toughness: opt.Val(game.PT{Value: 2}),
			ActivatedAbilities: []game.ActivatedAbility{
				game.ActivatedAbility{
					Text: "Cohort — {T}, Tap an untapped Ally you control: Target opponent loses 2 life.",
					AdditionalCosts: []cost.Additional{
						{
							Kind: cost.AdditionalTap,
						},
						{
							Kind:        cost.AdditionalTapPermanents,
							Text:        "Tap an untapped Ally you control",
							Amount:      1,
							SubtypesAny: cost.SubtypeSet{types.Ally},
						},
					},
					ZoneOfFunction: zone.Battlefield,
					Content: game.Mode{
						Targets: []game.TargetSpec{
							game.TargetSpec{
								MinTargets: 1,
								MaxTargets: 1,
								Constraint: "Target opponent",
								Allow:      game.TargetAllowPlayer,
								Selection:  opt.Val(game.Selection{Player: game.PlayerOpponent}),
							},
						},
						Sequence: []game.Instruction{
							{
								Primitive: game.LoseLife{
									Amount: game.Fixed(2),
									Player: game.TargetPlayerReference(0),
								},
							},
						},
					}.Ability(),
				},
			},
			OracleText: `
			Cohort — {T}, Tap an untapped Ally you control: Target opponent loses 2 life.
		`,
		},
	}
}
