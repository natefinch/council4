package b

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// BrigidSCommand is the card definition for Brigid's Command.
//
// Type: Kindred Sorcery — Kithkin
// Cost: {1}{G}{W}
//
// Oracle text:
//
//	Choose two —
//	• Create a token that's a copy of target Kithkin you control.
//	• Target player creates a 1/1 green and white Kithkin creature token.
//	• Target creature you control gets +3/+3 until end of turn.
//	• Target creature you control fights target creature an opponent controls.
var BrigidSCommand = newBrigidSCommand

func newBrigidSCommand() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White, color.Green),
		CardFace: game.CardFace{
			Name: "Brigid's Command",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
				cost.G,
				cost.W,
			}),
			Colors:   []color.Color{color.Green, color.White},
			Types:    []types.Card{types.Kindred, types.Sorcery},
			Subtypes: []types.Sub{types.Kithkin},
			SpellAbility: opt.Val(game.AbilityContent{
				Modes: []game.Mode{
					game.Mode{
						Text: "Create a token that's a copy of target Kithkin you control.",
						Targets: []game.TargetSpec{
							game.TargetSpec{
								MinTargets: 1,
								MaxTargets: 1,
								Constraint: "target Kithkin you control",
								Allow:      game.TargetAllowPermanent,
								Selection:  opt.Val(game.Selection{SubtypesAny: []types.Sub{types.Sub("Kithkin")}, Controller: game.ControllerYou}),
							},
						},
						Sequence: []game.Instruction{
							{
								Primitive: game.CreateToken{
									Amount: game.Fixed(1),
									Source: game.TokenCopyOf(game.TokenCopySpec{
										Source: game.TokenCopySourceObject,
										Object: game.TargetPermanentReference(0),
									}),
								},
							},
						},
					},
					game.Mode{
						Text: "Target player creates a 1/1 green and white Kithkin creature token.",
						Targets: []game.TargetSpec{
							game.TargetSpec{
								MinTargets: 1,
								MaxTargets: 1,
								Constraint: "Target player",
								Allow:      game.TargetAllowPlayer,
							},
						},
						Sequence: []game.Instruction{
							{
								Primitive: game.CreateToken{
									Amount:    game.Fixed(1),
									Source:    game.TokenDef(brigidSCommandToken),
									Recipient: opt.Val(game.TargetPlayerReference(0)),
								},
							},
						},
					},
					game.Mode{
						Text: "Target creature you control gets +3/+3 until end of turn.",
						Targets: []game.TargetSpec{
							game.TargetSpec{
								MinTargets: 1,
								MaxTargets: 1,
								Constraint: "target creature you control",
								Allow:      game.TargetAllowPermanent,
								Selection:  opt.Val(game.Selection{RequiredTypesAny: []types.Card{types.Creature}, Controller: game.ControllerYou}),
							},
						},
						Sequence: []game.Instruction{
							{
								Primitive: game.ModifyPT{
									Object:         game.TargetPermanentReference(0),
									PowerDelta:     game.Fixed(3),
									ToughnessDelta: game.Fixed(3),
									Duration:       game.DurationUntilEndOfTurn,
								},
							},
						},
					},
					game.Mode{
						Text: "Target creature you control fights target creature an opponent controls.",
						Targets: []game.TargetSpec{
							game.TargetSpec{
								MinTargets: 1,
								MaxTargets: 1,
								Constraint: "Target creature you control",
								Allow:      game.TargetAllowPermanent,
								Selection:  opt.Val(game.Selection{RequiredTypesAny: []types.Card{types.Creature}, Controller: game.ControllerYou}),
							},
							game.TargetSpec{
								MinTargets: 1,
								MaxTargets: 1,
								Constraint: "target creature an opponent controls",
								Allow:      game.TargetAllowPermanent,
								Selection:  opt.Val(game.Selection{RequiredTypesAny: []types.Card{types.Creature}, Controller: game.ControllerOpponent}),
							},
						},
						Sequence: []game.Instruction{
							{
								Primitive: game.Fight{
									Object:        game.TargetPermanentReference(0),
									RelatedObject: game.TargetPermanentReference(1),
								},
							},
						},
					},
				},
				MinModes: 2,
				MaxModes: 2,
			}),
			OracleText: `
			Choose two —
			• Create a token that's a copy of target Kithkin you control.
			• Target player creates a 1/1 green and white Kithkin creature token.
			• Target creature you control gets +3/+3 until end of turn.
			• Target creature you control fights target creature an opponent controls.
		`,
		},
	}
}

var brigidSCommandToken = newBrigidSCommandToken()

func newBrigidSCommandToken() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name:      "Kithkin",
			Colors:    []color.Color{color.Green, color.White},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Kithkin},
			Power:     opt.Val(game.PT{Value: 1}),
			Toughness: opt.Val(game.PT{Value: 1}),
		},
	}
}
