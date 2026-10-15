package g

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// GlimpseTheCore is the card definition for Glimpse the Core.
//
// Type: Sorcery
// Cost: {1}{G}
//
// Oracle text:
//
//	Choose one —
//	• Search your library for a basic Forest card, put that card onto the battlefield tapped, then shuffle.
//	• Return target Cave card from your graveyard to the battlefield tapped.
var GlimpseTheCore = newGlimpseTheCore

func newGlimpseTheCore() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Green),
		CardFace: game.CardFace{
			Name: "Glimpse the Core",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
				cost.G,
			}),
			Colors: []color.Color{color.Green},
			Types:  []types.Card{types.Sorcery},
			SpellAbility: opt.Val(game.AbilityContent{
				Modes: []game.Mode{
					game.Mode{
						Text: "Search your library for a basic Forest card, put that card onto the battlefield tapped, then shuffle.",
						Sequence: []game.Instruction{
							{
								Primitive: game.Search{
									Player: game.ControllerReference(),
									Spec: game.SearchSpec{
										SourceZone:   zone.Library,
										Destination:  zone.Battlefield,
										Filter:       game.Selection{Supertypes: []types.Super{types.Basic}, SubtypesAny: []types.Sub{types.Forest}},
										EntersTapped: true,
									},
									Amount: game.Fixed(1),
								},
							},
						},
					},
					game.Mode{
						Text: "Return target Cave card from your graveyard to the battlefield tapped.",
						Targets: []game.TargetSpec{
							game.TargetSpec{
								MinTargets: 1,
								MaxTargets: 1,
								Constraint: "target Cave card from your graveyard",
								Allow:      game.TargetAllowCard,
								TargetZone: zone.Graveyard,
								Selection:  opt.Val(game.Selection{SubtypesAny: []types.Sub{types.Sub("Cave")}, Controller: game.ControllerYou}),
							},
						},
						Sequence: []game.Instruction{
							{
								Primitive: game.PutOnBattlefield{
									Source:      game.CardBattlefieldSource(game.CardReference{Kind: game.CardReferenceTarget}),
									EntryTapped: true,
								},
							},
						},
					},
				},
				MinModes: 1,
				MaxModes: 1,
			}),
			OracleText: `
			Choose one —
			• Search your library for a basic Forest card, put that card onto the battlefield tapped, then shuffle.
			• Return target Cave card from your graveyard to the battlefield tapped.
		`,
		},
	}
}
