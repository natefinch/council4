package i

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// InsatiableAvarice is the card definition for Insatiable Avarice.
//
// Type: Sorcery
// Cost: {B}
//
// Oracle text:
//
//	Spree (Choose one or more additional costs.)
//	+ {2} — Search your library for a card, then shuffle and put that card on top.
//	+ {B}{B} — Target player draws three cards and loses 3 life.
var InsatiableAvarice = newInsatiableAvarice

func newInsatiableAvarice() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Black),
		CardFace: game.CardFace{
			Name: "Insatiable Avarice",
			ManaCost: opt.Val(cost.Mana{
				cost.B,
			}),
			Colors: []color.Color{color.Black},
			Types:  []types.Card{types.Sorcery},
			SpellAbility: opt.Val(game.AbilityContent{
				Modes: []game.Mode{
					game.Mode{
						Text: "{2} — Search your library for a card, then shuffle and put that card on top.",
						Cost: opt.Val(cost.Mana{cost.O(2)}),
						Sequence: []game.Instruction{
							{
								Primitive: game.Search{
									Player: game.ControllerReference(),
									Spec: game.SearchSpec{
										SourceZone:          zone.Library,
										Destination:         zone.Library,
										DestinationPosition: game.SearchPositionTop,
										FailToFindPolicy:    game.SearchMustFindIfAvailable,
									},
									Amount: game.Fixed(1),
								},
							},
						},
					},
					game.Mode{
						Text: "{B}{B} — Target player draws three cards and loses 3 life.",
						Cost: opt.Val(cost.Mana{cost.B, cost.B}),
						Targets: []game.TargetSpec{
							game.TargetSpec{
								MinTargets: 1,
								MaxTargets: 1,
								Constraint: "target player",
								Allow:      game.TargetAllowPlayer,
							},
						},
						Sequence: []game.Instruction{
							{
								Primitive: game.Draw{
									Amount: game.Fixed(3),
									Player: game.TargetPlayerReference(0),
								},
							},
							{
								Primitive: game.LoseLife{
									Amount: game.Fixed(3),
									Player: game.TargetPlayerReference(0),
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
			+ {2} — Search your library for a card, then shuffle and put that card on top.
			+ {B}{B} — Target player draws three cards and loses 3 life.
		`,
		},
	}
}
