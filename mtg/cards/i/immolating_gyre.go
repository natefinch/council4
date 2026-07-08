package i

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// ImmolatingGyre is the card definition for Immolating Gyre.
//
// Type: Sorcery
// Cost: {4}{R}{R}
//
// Oracle text:
//
//	Immolating Gyre deals X damage to each creature and planeswalker you don't control, where X is the number of instant and sorcery cards in your graveyard.
var ImmolatingGyre = newImmolatingGyre

func newImmolatingGyre() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Red),
		CardFace: game.CardFace{
			Name: "Immolating Gyre",
			ManaCost: opt.Val(cost.Mana{
				cost.O(4),
				cost.R,
				cost.R,
			}),
			Colors: []color.Color{color.Red},
			Types:  []types.Card{types.Sorcery},
			SpellAbility: opt.Val(game.Mode{
				Sequence: []game.Instruction{
					{
						Primitive: game.Damage{
							Amount: game.Dynamic(game.DynamicAmount{
								Kind:       game.DynamicAmountCountCardsInZone,
								Multiplier: 1,
								Player:     func() *game.PlayerReference { ref := game.ControllerReference(); return &ref }(),
								CardZone:   zone.Graveyard,
								Selection:  &game.Selection{RequiredTypesAny: []types.Card{types.Instant, types.Sorcery}},
							}),
							Recipient: game.GroupDamageRecipient(game.BattlefieldGroup(game.Selection{RequiredTypesAny: []types.Card{types.Creature, types.Planeswalker}, Controller: game.ControllerNotYou})),
						},
					},
				},
			}.Ability()),
			OracleText: `
			Immolating Gyre deals X damage to each creature and planeswalker you don't control, where X is the number of instant and sorcery cards in your graveyard.
		`,
		},
	}
}
