package c

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// CaveIn is the card definition for Cave-In.
//
// Type: Sorcery
// Cost: {3}{R}{R}
//
// Oracle text:
//
//	You may exile a red card from your hand rather than pay this spell's mana cost.
//	Cave-In deals 2 damage to each creature and each player.
var CaveIn = newCaveIn

func newCaveIn() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Red),
		CardFace: game.CardFace{
			Name: "Cave-In",
			ManaCost: opt.Val(cost.Mana{
				cost.O(3),
				cost.R,
				cost.R,
			}),
			Colors: []color.Color{color.Red},
			Types:  []types.Card{types.Sorcery},
			AlternativeCosts: []cost.Alternative{
				cost.Alternative{
					Label: "Exile a red card",
					AdditionalCosts: []cost.Additional{
						{
							Kind:           cost.AdditionalExile,
							Amount:         1,
							Source:         zone.Hand,
							MatchCardColor: true,
							CardColor:      color.Red,
						},
					},
				},
			},
			SpellAbility: opt.Val(game.Mode{
				Sequence: []game.Instruction{
					{
						Primitive: game.Damage{
							Amount:    game.Fixed(2),
							Recipient: game.GroupDamageRecipient(game.BattlefieldGroup(game.Selection{RequiredTypes: []types.Card{types.Creature}})),
						},
					},
					{
						Primitive: game.Damage{
							Amount:    game.Fixed(2),
							Recipient: game.PlayerGroupDamageRecipient(game.AllPlayersReference()),
						},
					},
				},
			}.Ability()),
			OracleText: `
			You may exile a red card from your hand rather than pay this spell's mana cost.
			Cave-In deals 2 damage to each creature and each player.
		`,
		},
	}
}
