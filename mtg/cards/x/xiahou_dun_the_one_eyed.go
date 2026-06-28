package x

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// XiahouDunTheOneEyed is the card definition for Xiahou Dun, the One-Eyed.
//
// Type: Legendary Creature — Human Soldier
// Cost: {2}{B}{B}
//
// Oracle text:
//
//	Horsemanship (This creature can't be blocked except by creatures with horsemanship.)
//	Sacrifice Xiahou Dun: Return target black card from your graveyard to your hand. Activate only during your turn, before attackers are declared.
var XiahouDunTheOneEyed = newXiahouDunTheOneEyed()

func newXiahouDunTheOneEyed() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Black),
		CardFace: game.CardFace{
			Name: "Xiahou Dun, the One-Eyed",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.B,
				cost.B,
			}),
			Colors:     []color.Color{color.Black},
			Supertypes: []types.Super{types.Legendary},
			Types:      []types.Card{types.Creature},
			Subtypes:   []types.Sub{types.Human, types.Soldier},
			Power:      opt.Val(game.PT{Value: 3}),
			Toughness:  opt.Val(game.PT{Value: 2}),
			StaticAbilities: []game.StaticAbility{
				game.HorsemanshipStaticBody,
			},
			ActivatedAbilities: []game.ActivatedAbility{
				game.ActivatedAbility{
					Text: "Sacrifice Xiahou Dun: Return target black card from your graveyard to your hand. Activate only during your turn, before attackers are declared.",
					AdditionalCosts: []cost.Additional{
						{
							Kind:   cost.AdditionalSacrificeSource,
							Text:   "Sacrifice Xiahou Dun",
							Amount: 1,
						},
					},
					ZoneOfFunction: zone.Battlefield,
					Timing:         game.DuringYourTurnBeforeAttackers,
					Content: game.Mode{
						Targets: []game.TargetSpec{
							game.TargetSpec{
								MinTargets: 1,
								MaxTargets: 1,
								Constraint: "target black card from your graveyard",
								Allow:      game.TargetAllowCard,
								TargetZone: zone.Graveyard,
								Selection:  opt.Val(game.Selection{ColorsAny: []color.Color{color.Black}, Controller: game.ControllerYou}),
							},
						},
						Sequence: []game.Instruction{
							{
								Primitive: game.MoveCard{
									Card:        game.CardReference{Kind: game.CardReferenceTarget},
									FromZone:    zone.Graveyard,
									Destination: zone.Hand,
								},
							},
						},
					}.Ability(),
				},
			},
			OracleText: `
			Horsemanship (This creature can't be blocked except by creatures with horsemanship.)
			Sacrifice Xiahou Dun: Return target black card from your graveyard to your hand. Activate only during your turn, before attackers are declared.
		`,
		},
	}
}
