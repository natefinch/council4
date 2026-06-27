package c

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// CroakingCounterpart is the card definition for Croaking Counterpart.
//
// Type: Sorcery
// Cost: {1}{G}{U}
//
// Oracle text:
//
//	Create a token that's a copy of target non-Frog creature, except it's a 1/1 green Frog.
//	Flashback {3}{G}{U} (You may cast this card from your graveyard for its flashback cost. Then exile it.)
var CroakingCounterpart = newCroakingCounterpart()

func newCroakingCounterpart() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue, color.Green),
		CardFace: game.CardFace{
			Name: "Croaking Counterpart",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
				cost.G,
				cost.U,
			}),
			Colors: []color.Color{color.Green, color.Blue},
			Types:  []types.Card{types.Sorcery},
			StaticAbilities: []game.StaticAbility{
				game.StaticAbility{
					KeywordAbilities: []game.KeywordAbility{
						game.FlashbackKeyword{Cost: cost.Mana{cost.O(3), cost.G, cost.U}},
					},
				},
			},
			SpellAbility: opt.Val(game.Mode{
				Targets: []game.TargetSpec{
					game.TargetSpec{
						MinTargets: 1,
						MaxTargets: 1,
						Constraint: "target non-Frog creature",
						Allow:      game.TargetAllowPermanent,
						Selection:  opt.Val(game.Selection{RequiredTypesAny: []types.Card{types.Creature}, ExcludedSubtype: types.Sub("Frog")}),
					},
				},
				Sequence: []game.Instruction{
					{
						Primitive: game.CreateToken{
							Amount: game.Fixed(1),
							Source: game.TokenCopyOf(game.TokenCopySpec{
								Source:       game.TokenCopySourceObject,
								Object:       game.TargetPermanentReference(0),
								SetPower:     opt.Val(game.PT{Value: 1}),
								SetToughness: opt.Val(game.PT{Value: 1}),
								SetColors:    []color.Color{color.Green},
								SetSubtypes:  []types.Sub{types.Frog},
							}),
						},
					},
				},
			}.Ability()),
			OracleText: `
			Create a token that's a copy of target non-Frog creature, except it's a 1/1 green Frog.
			Flashback {3}{G}{U} (You may cast this card from your graveyard for its flashback cost. Then exile it.)
		`,
		},
	}
}
