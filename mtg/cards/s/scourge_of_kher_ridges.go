package s

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// ScourgeOfKherRidges is the card definition for Scourge of Kher Ridges.
//
// Type: Creature — Dragon
// Cost: {6}{R}{R}
//
// Oracle text:
//
//	Flying
//	{1}{R}: This creature deals 2 damage to each creature without flying.
//	{5}{R}: This creature deals 6 damage to each other creature with flying.
var ScourgeOfKherRidges = newScourgeOfKherRidges

func newScourgeOfKherRidges() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Red),
		CardFace: game.CardFace{
			Name: "Scourge of Kher Ridges",
			ManaCost: opt.Val(cost.Mana{
				cost.O(6),
				cost.R,
				cost.R,
			}),
			Colors:    []color.Color{color.Red},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Dragon},
			Power:     opt.Val(game.PT{Value: 6}),
			Toughness: opt.Val(game.PT{Value: 6}),
			StaticAbilities: []game.StaticAbility{
				game.FlyingStaticBody,
			},
			ActivatedAbilities: []game.ActivatedAbility{
				game.ActivatedAbility{
					Text: "{1}{R}: This creature deals 2 damage to each creature without flying.",
					ManaCost: opt.Val(cost.Mana{
						cost.O(1),
						cost.R,
					}),
					ZoneOfFunction: zone.Battlefield,
					Content: game.AbilityContent{
						Modes: []game.Mode{
							game.Mode{
								Sequence: []game.Instruction{
									game.Instruction{
										Primitive: game.Damage{
											Amount:       game.Fixed(2),
											Recipient:    game.GroupDamageRecipient(game.BattlefieldGroup(game.Selection{RequiredTypes: []types.Card{types.Creature}, ExcludedKeyword: game.Flying})),
											DamageSource: opt.Val(game.SourcePermanentReference()),
										},
									},
								},
							},
						},
						MinModes: 1,
						MaxModes: 1,
					},
				},
				game.ActivatedAbility{
					Text: "{5}{R}: This creature deals 6 damage to each other creature with flying.",
					ManaCost: opt.Val(cost.Mana{
						cost.O(5),
						cost.R,
					}),
					ZoneOfFunction: zone.Battlefield,
					Content: game.AbilityContent{
						Modes: []game.Mode{
							game.Mode{
								Sequence: []game.Instruction{
									game.Instruction{
										Primitive: game.Damage{
											Amount:       game.Fixed(6),
											Recipient:    game.GroupDamageRecipient(game.BattlefieldGroupExcluding(game.Selection{RequiredTypes: []types.Card{types.Creature}, Keyword: game.Flying}, game.SourcePermanentReference())),
											DamageSource: opt.Val(game.SourcePermanentReference()),
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
			OracleText: `
			Flying
			{1}{R}: This creature deals 2 damage to each creature without flying.
			{5}{R}: This creature deals 6 damage to each other creature with flying.
		`,
		},
	}
}
