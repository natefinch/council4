package n

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// NetworkTerminal is the card definition for Network Terminal.
//
// Type: Artifact
// Cost: {3}
//
// Oracle text:
//
//	{T}: Add one mana of any color.
//	{1}, {T}, Tap another untapped artifact you control: Draw a card, then discard a card.
var NetworkTerminal = newNetworkTerminal()

func newNetworkTerminal() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name: "Network Terminal",
			ManaCost: opt.Val(cost.Mana{
				cost.O(3),
			}),
			Types: []types.Card{types.Artifact},
			ActivatedAbilities: []game.ActivatedAbility{
				game.ActivatedAbility{
					Text:     "{1}, {T}, Tap another untapped artifact you control: Draw a card, then discard a card.",
					ManaCost: opt.Val(cost.Mana{cost.O(1)}),
					AdditionalCosts: []cost.Additional{
						{
							Kind: cost.AdditionalTap,
						},
						{
							Kind:               cost.AdditionalTapPermanents,
							Text:               "Tap another untapped artifact you control",
							Amount:             1,
							MatchPermanentType: true,
							PermanentType:      types.Artifact,
							ExcludeSource:      true,
						},
					},
					ZoneOfFunction: zone.Battlefield,
					Content: game.Mode{
						Sequence: []game.Instruction{
							{
								Primitive: game.Draw{
									Amount: game.Fixed(1),
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
				game.TapManaChoiceAbility(mana.W, mana.U, mana.B, mana.R, mana.G),
			},
			OracleText: `
			{T}: Add one mana of any color.
			{1}, {T}, Tap another untapped artifact you control: Draw a card, then discard a card.
		`,
		},
	}
}
