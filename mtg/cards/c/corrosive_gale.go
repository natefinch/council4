package c

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// CorrosiveGale is the card definition for Corrosive Gale.
//
// Type: Sorcery
// Cost: {X}{G/P}
//
// Oracle text:
//
//	({G/P} can be paid with either {G} or 2 life.)
//	Corrosive Gale deals X damage to each creature with flying.
var CorrosiveGale = newCorrosiveGale

func newCorrosiveGale() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Green),
		CardFace: game.CardFace{
			Name: "Corrosive Gale",
			ManaCost: opt.Val(cost.Mana{
				cost.X,
				cost.PhyrexianMana(mana.G),
			}),
			Colors: []color.Color{color.Green},
			Types:  []types.Card{types.Sorcery},
			SpellAbility: opt.Val(game.Mode{
				Sequence: []game.Instruction{
					{
						Primitive: game.Damage{
							Amount: game.Dynamic(game.DynamicAmount{
								Kind: game.DynamicAmountX,
							}),
							Recipient: game.GroupDamageRecipient(game.BattlefieldGroup(game.Selection{RequiredTypes: []types.Card{types.Creature}, Keyword: game.Flying})),
						},
					},
				},
			}.Ability()),
			OracleText: `
			({G/P} can be paid with either {G} or 2 life.)
			Corrosive Gale deals X damage to each creature with flying.
		`,
		},
	}
}
