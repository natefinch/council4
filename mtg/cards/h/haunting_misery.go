package h

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// HauntingMisery is the card definition for Haunting Misery.
//
// Type: Sorcery
// Cost: {1}{B}{B}
//
// Oracle text:
//
//	As an additional cost to cast this spell, exile X creature cards from your graveyard.
//	Haunting Misery deals X damage to target player or planeswalker.
var HauntingMisery = newHauntingMisery

func newHauntingMisery() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Black),
		CardFace: game.CardFace{
			Name: "Haunting Misery",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
				cost.B,
				cost.B,
			}),
			Colors: []color.Color{color.Black},
			Types:  []types.Card{types.Sorcery},
			AdditionalCosts: []cost.Additional{
				{
					Kind:          cost.AdditionalExile,
					Text:          "exile X creature cards from your graveyard",
					AmountFromX:   true,
					Source:        zone.Graveyard,
					MatchCardType: true,
					CardType:      types.Creature,
				},
			},
			SpellAbility: opt.Val(game.Mode{
				Targets: []game.TargetSpec{
					game.TargetSpec{
						MinTargets: 1,
						MaxTargets: 1,
						Constraint: "target player or planeswalker",
						Allow:      game.TargetAllowPermanent | game.TargetAllowPlayer,
						Selection:  opt.Val(game.Selection{RequiredTypesAny: []types.Card{types.Planeswalker}}),
					},
				},
				Sequence: []game.Instruction{
					{
						Primitive: game.Damage{
							Amount: game.Dynamic(game.DynamicAmount{
								Kind: game.DynamicAmountX,
							}),
							Recipient: game.AnyTargetDamageRecipient(0),
						},
					},
				},
			}.Ability()),
			OracleText: `
			As an additional cost to cast this spell, exile X creature cards from your graveyard.
			Haunting Misery deals X damage to target player or planeswalker.
		`,
		},
	}
}
