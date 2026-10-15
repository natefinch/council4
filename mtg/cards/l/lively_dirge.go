package l

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// LivelyDirge is the card definition for Lively Dirge.
//
// Type: Sorcery
// Cost: {1}{B}
//
// Oracle text:
//
//	Spree (Choose one or more additional costs.)
//	+ {1} — Search your library for a card, put it into your graveyard, then shuffle.
//	+ {2} — Return up to two creature cards with total mana value 4 or less from your graveyard to the battlefield.
var LivelyDirge = newLivelyDirge

func newLivelyDirge() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Black),
		CardFace: game.CardFace{
			Name: "Lively Dirge",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
				cost.B,
			}),
			Colors: []color.Color{color.Black},
			Types:  []types.Card{types.Sorcery},
			SpellAbility: opt.Val(game.AbilityContent{
				Modes: []game.Mode{
					game.Mode{
						Text: "{1} — Search your library for a card, put it into your graveyard, then shuffle.",
						Cost: opt.Val(cost.Mana{cost.O(1)}),
						Sequence: []game.Instruction{
							{
								Primitive: game.Search{
									Player: game.ControllerReference(),
									Spec: game.SearchSpec{
										SourceZone:       zone.Library,
										Destination:      zone.Graveyard,
										FailToFindPolicy: game.SearchMustFindIfAvailable,
									},
									Amount: game.Fixed(1),
								},
							},
						},
					},
					game.Mode{
						Text: "{2} — Return up to two creature cards with total mana value 4 or less from your graveyard to the battlefield.",
						Cost: opt.Val(cost.Mana{cost.O(2)}),
						Sequence: []game.Instruction{
							{
								Primitive: game.ChooseFromZone{
									Player:     game.ControllerReference(),
									SourceZone: zone.Graveyard,
									Filter:     game.Selection{RequiredTypes: []types.Card{types.Creature}, Controller: game.ControllerYou},
									Quantity:   game.Fixed(2),
									Destination: game.ChooseDestination{
										Zone: zone.Battlefield,
									},
									Riders: game.ChooseRiders{
										MaxTotalManaValue: opt.Val(4),
									},
									Prompt: "Choose a card to return to the battlefield",
								},
							},
						},
					},
				},
				MinModes: 1,
				MaxModes: 2,
			}),
			OracleText: `
			Spree (Choose one or more additional costs.)
			+ {1} — Search your library for a card, put it into your graveyard, then shuffle.
			+ {2} — Return up to two creature cards with total mana value 4 or less from your graveyard to the battlefield.
		`,
		},
	}
}
