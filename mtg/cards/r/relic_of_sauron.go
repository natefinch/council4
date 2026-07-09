package r

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// RelicOfSauron is the card definition for Relic of Sauron.
//
// Type: Artifact
// Cost: {4}
//
// Oracle text:
//
//	{T}: Add two mana in any combination of {U}, {B}, and/or {R}.
//	{3}, {T}: Draw two cards, then discard a card.
var RelicOfSauron = newRelicOfSauron

func newRelicOfSauron() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue, color.Black, color.Red),
		CardFace: game.CardFace{
			Name: "Relic of Sauron",
			ManaCost: opt.Val(cost.Mana{
				cost.O(4),
			}),
			Types: []types.Card{types.Artifact},
			ActivatedAbilities: []game.ActivatedAbility{
				game.ActivatedAbility{
					Text:            "{3}, {T}: Draw two cards, then discard a card.",
					ManaCost:        opt.Val(cost.Mana{cost.O(3)}),
					AdditionalCosts: cost.Tap,
					ZoneOfFunction:  zone.Battlefield,
					Content: game.Mode{
						Sequence: []game.Instruction{
							{
								Primitive: game.Draw{
									Amount: game.Fixed(2),
									Player: game.ControllerReference(),
								},
							},
							{
								Primitive: game.Discard{
									Amount: game.Fixed(1),
									Player: game.ControllerReference(),
								},
							},
						},
					}.Ability(),
				},
			},
			ManaAbilities: []game.ManaAbility{
				game.ManaAbility{
					AdditionalCosts: cost.Tap,
					Content: game.Mode{
						Sequence: []game.Instruction{
							{
								Primitive: game.AddMana{
									Amount:            game.Fixed(2),
									CombinationColors: []mana.Color{mana.U, mana.B, mana.R},
								},
							},
						},
					}.Ability(),
				},
			},
			OracleText: `
			{T}: Add two mana in any combination of {U}, {B}, and/or {R}.
			{3}, {T}: Draw two cards, then discard a card.
		`,
		},
	}
}
