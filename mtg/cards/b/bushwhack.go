package b

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/zone"

	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// Bushwhack is the card definition for Bushwhack.
//
// Type: Sorcery
// Cost: {G}
//
// Oracle text:
//
//	Choose one —
//	• Search your library for a basic land card, reveal it, put it into your hand, then shuffle.
//	• Target creature you control fights target creature you don't control. (Each deals damage equal to its power to the other.)
var Bushwhack = &game.CardDef{
	ColorIdentity: color.NewIdentity(color.Green),
	CardFace: game.CardFace{
		Name: "Bushwhack",
		ManaCost: opt.Val(cost.Mana{
			cost.G,
		}),
		Colors: []color.Color{color.Green},
		Types:  []types.Card{types.Sorcery},
		OracleText: `
			Choose one —
			• Search your library for a basic land card, reveal it, put it into your hand, then shuffle.
			• Target creature you control fights target creature you don't control. (Each deals damage equal to its power to the other.)
		`,
		SpellAbility: opt.Val(
			game.SpellAbilityBody{
				Text: `
					Choose one —
					• Search your library for a basic land card, reveal it, put it into your hand, then shuffle.
					• Target creature you control fights target creature you don't control.
				`,
				Content: game.ModalAbilityContent{
					Modes: []game.Mode{
						{
							Text: "Search your library for a basic land card, reveal it, put it into your hand, then shuffle.",
							Sequence: []game.Instruction{
								{
									Primitive: game.Search{
										TargetIndex: game.TargetIndexController,
										Spec: game.SearchSpec{
											SourceZone:  zone.Library,
											Destination: zone.Hand,
											CardType:    opt.Val(types.Land),
											Supertype:   opt.Val(types.Basic),
											Reveal:      true,
										},
									},
								},
							},
						},
						{
							Text: "Target creature you control fights target creature you don't control.",
							Sequence: []game.Instruction{
								{
									Primitive: game.Fight{},
								},
							},
							Targets: []game.TargetSpec{
								{
									MinTargets: 1,
									MaxTargets: 1,
									Constraint: "creature you control",
									Allow:      game.TargetAllowPermanent,
									Predicate: game.TargetPredicate{
										PermanentTypes: []types.Card{types.Creature},
										Controller:     game.ControllerYou,
									},
								},
								{
									MinTargets: 1,
									MaxTargets: 1,
									Constraint: "creature you don't control",
									Allow:      game.TargetAllowPermanent,
									Predicate: game.TargetPredicate{
										PermanentTypes: []types.Card{types.Creature},
										Controller:     game.ControllerNotYou,
									},
								},
							},
						},
					},
				},
			},
		),
	},
}
