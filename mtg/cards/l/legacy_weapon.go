package l

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// LegacyWeapon is the card definition for Legacy Weapon.
//
// Type: Legendary Artifact
// Cost: {7}
//
// Oracle text:
//
//	{W}{U}{B}{R}{G}: Exile target permanent.
//	If Legacy Weapon would be put into a graveyard from anywhere, reveal Legacy Weapon and shuffle it into its owner's library instead.
var LegacyWeapon = newLegacyWeapon

func newLegacyWeapon() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White, color.Blue, color.Black, color.Red, color.Green),
		CardFace: game.CardFace{
			Name: "Legacy Weapon",
			ManaCost: opt.Val(cost.Mana{
				cost.O(7),
			}),
			Supertypes: []types.Super{types.Legendary},
			Types:      []types.Card{types.Artifact},
			ActivatedAbilities: []game.ActivatedAbility{
				game.ActivatedAbility{
					Text: "{W}{U}{B}{R}{G}: Exile target permanent.",
					ManaCost: opt.Val(cost.Mana{
						cost.W,
						cost.U,
						cost.B,
						cost.R,
						cost.G,
					}),
					ZoneOfFunction: zone.Battlefield,
					Content: game.AbilityContent{
						Modes: []game.Mode{
							game.Mode{
								Targets: []game.TargetSpec{
									game.TargetSpec{
										MinTargets: 1,
										MaxTargets: 1,
										Constraint: "target permanent",
										Allow:      game.TargetAllowPermanent,
									},
								},
								Sequence: []game.Instruction{
									game.Instruction{
										Primitive: game.MovePermanent{
											Object:      game.TargetPermanentReference(0),
											Destination: zone.Exile,
										},
									},
								},
							},
						},
						MinModes: 1,
						MaxModes: 1,
					},
				},
			},
			ReplacementAbilities: []game.ReplacementAbility{
				game.ReplacementAbility{Text: "If Legacy Weapon would be put into a graveyard from anywhere, reveal Legacy Weapon and shuffle it into its owner's library instead.", Replacement: game.ReplacementEffect{
					MatchEvent:         game.EventZoneChanged,
					MatchToZone:        true,
					ToZone:             zone.Graveyard,
					ReplaceToZone:      zone.Library,
					Duration:           game.DurationPermanent,
					ShuffleIntoLibrary: true,
					RevealSource:       true,
				}},
			},
			OracleText: `
			{W}{U}{B}{R}{G}: Exile target permanent.
			If Legacy Weapon would be put into a graveyard from anywhere, reveal Legacy Weapon and shuffle it into its owner's library instead.
		`,
		},
	}
}
