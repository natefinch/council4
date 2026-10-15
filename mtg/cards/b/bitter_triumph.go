package b

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// BitterTriumph is the card definition for Bitter Triumph.
//
// Type: Instant
// Cost: {1}{B}
//
// Oracle text:
//
//	As an additional cost to cast this spell, discard a card or pay 3 life.
//	Destroy target creature or planeswalker.
var BitterTriumph = newBitterTriumph

func newBitterTriumph() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Black),
		CardFace: game.CardFace{
			Name: "Bitter Triumph",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
				cost.B,
			}),
			Colors: []color.Color{color.Black},
			Types:  []types.Card{types.Instant},
			AdditionalCosts: []cost.Additional{
				{
					Kind:        cost.AdditionalDiscard,
					Text:        "discard a card",
					Amount:      1,
					Source:      zone.Hand,
					ChoiceGroup: 1,
				},
				{
					Kind:        cost.AdditionalPayLife,
					Text:        "pay 3 life",
					Amount:      3,
					ChoiceGroup: 1,
				},
			},
			SpellAbility: opt.Val(game.Mode{
				Targets: []game.TargetSpec{
					game.TargetSpec{
						MinTargets: 1,
						MaxTargets: 1,
						Constraint: "target creature or planeswalker",
						Allow:      game.TargetAllowPermanent,
						Selection:  opt.Val(game.Selection{RequiredTypesAny: []types.Card{types.Creature, types.Planeswalker}}),
					},
				},
				Sequence: []game.Instruction{
					{
						Primitive: game.Destroy{
							Object: game.TargetPermanentReference(0),
						},
					},
				},
			}.Ability()),
			OracleText: `
			As an additional cost to cast this spell, discard a card or pay 3 life.
			Destroy target creature or planeswalker.
		`,
		},
	}
}
