package u

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// UndercityInformer is the card definition for Undercity Informer.
//
// Type: Creature — Human Rogue
// Cost: {2}{B}
//
// Oracle text:
//
//	{1}, Sacrifice a creature: Target player reveals cards from the top of their library until they reveal a land card, then puts those cards into their graveyard.
var UndercityInformer = &game.CardDef{
	ColorIdentity: color.NewIdentity(color.Black),
	CardFace: game.CardFace{
		Name: "Undercity Informer",
		ManaCost: opt.Val(cost.Mana{
			cost.O(2),
			cost.B,
		}),
		Colors:    []color.Color{color.Black},
		Types:     []types.Card{types.Creature},
		Subtypes:  []types.Sub{types.Human, types.Rogue},
		Power:     opt.Val(game.PT{Value: 2}),
		Toughness: opt.Val(game.PT{Value: 3}),
		ActivatedAbilities: []game.ActivatedAbility{
			game.ActivatedAbility{
				ManaCost: opt.Val(cost.Mana{cost.O(1)}),
				AdditionalCosts: []cost.Additional{
					{
						Kind:               cost.AdditionalSacrifice,
						Text:               "Sacrifice a creature",
						Amount:             1,
						MatchPermanentType: true,
						PermanentType:      types.Creature,
					},
				},
				ZoneOfFunction: zone.Battlefield,
				Content: game.Mode{
					Targets: []game.TargetSpec{
						game.TargetSpec{
							MinTargets: 1,
							MaxTargets: 1,
							Constraint: "Target player",
							Allow:      game.TargetAllowPlayer,
						},
					},
					Sequence: []game.Instruction{
						{
							Primitive: game.RevealUntil{
								Player:      game.TargetPlayerReference(0),
								Until:       game.Selection{RequiredTypes: []types.Card{types.Land}},
								Destination: zone.Graveyard,
							},
						},
					},
				}.Ability(),
			},
		},
		OracleText: `
			{1}, Sacrifice a creature: Target player reveals cards from the top of their library until they reveal a land card, then puts those cards into their graveyard.
		`,
	},
}
