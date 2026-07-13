package t

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// TemptWithImmortality is the card definition for Tempt with Immortality.
//
// Type: Sorcery
// Cost: {4}{B}
//
// Oracle text:
//
//	Tempting offer — Return a creature card from your graveyard to the battlefield. Each opponent may return a creature card from their graveyard to the battlefield. For each opponent who does, return a creature card from your graveyard to the battlefield.
var TemptWithImmortality = newTemptWithImmortality

func newTemptWithImmortality() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Black),
		CardFace: game.CardFace{
			Name: "Tempt with Immortality",
			ManaCost: opt.Val(cost.Mana{
				cost.O(4),
				cost.B,
			}),
			Colors: []color.Color{color.Black},
			Types:  []types.Card{types.Sorcery},
			SpellAbility: opt.Val(game.Mode{
				Sequence: []game.Instruction{
					{
						Primitive: game.ChooseFromZone{
							Player:     game.GroupOfferMemberReference(),
							SourceZone: zone.Graveyard,
							Filter:     game.Selection{RequiredTypes: []types.Card{types.Creature}},
							Quantity:   game.Fixed(1),
							Destination: game.ChooseDestination{
								Zone: zone.Battlefield,
							},
							Prompt: "Choose a card to return to the battlefield",
						},
						Optional:           true,
						OptionalActorGroup: opt.Val(game.OpponentsReference()),
						TemptingOffer:      true,
					},
				},
			}.Ability()),
			OracleText: `
			Tempting offer — Return a creature card from your graveyard to the battlefield. Each opponent may return a creature card from their graveyard to the battlefield. For each opponent who does, return a creature card from your graveyard to the battlefield.
		`,
		},
	}
}
