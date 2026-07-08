package r

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// RiteOfOblivion is the card definition for Rite of Oblivion.
//
// Type: Sorcery
// Cost: {W}{B}
//
// Oracle text:
//
//	As an additional cost to cast this spell, sacrifice a nonland permanent.
//	Exile target nonland permanent.
//	Flashback {2}{W}{B} (You may cast this card from your graveyard for its flashback cost and any additional costs. Then exile it.)
var RiteOfOblivion = newRiteOfOblivion

func newRiteOfOblivion() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White, color.Black),
		CardFace: game.CardFace{
			Name: "Rite of Oblivion",
			ManaCost: opt.Val(cost.Mana{
				cost.W,
				cost.B,
			}),
			Colors: []color.Color{color.Black, color.White},
			Types:  []types.Card{types.Sorcery},
			StaticAbilities: []game.StaticAbility{
				game.StaticAbility{
					KeywordAbilities: []game.KeywordAbility{
						game.FlashbackKeyword{Cost: cost.Mana{cost.O(2), cost.W, cost.B}},
					},
				},
			},
			AdditionalCosts: []cost.Additional{
				{
					Kind:                 cost.AdditionalSacrifice,
					Text:                 "sacrifice a nonland permanent",
					Amount:               1,
					ExcludePermanentType: types.Land,
				},
			},
			SpellAbility: opt.Val(game.Mode{
				Targets: []game.TargetSpec{
					game.TargetSpec{
						MinTargets: 1,
						MaxTargets: 1,
						Constraint: "target nonland permanent",
						Allow:      game.TargetAllowPermanent,
						Selection: opt.Val(game.Selection{
							ExcludedTypes: []types.Card{types.Land},
						}),
					},
				},
				Sequence: []game.Instruction{
					{
						Primitive: game.Exile{
							Object: game.TargetPermanentReference(0),
						},
					},
				},
			}.Ability()),
			OracleText: `
			As an additional cost to cast this spell, sacrifice a nonland permanent.
			Exile target nonland permanent.
			Flashback {2}{W}{B} (You may cast this card from your graveyard for its flashback cost and any additional costs. Then exile it.)
		`,
		},
	}
}
