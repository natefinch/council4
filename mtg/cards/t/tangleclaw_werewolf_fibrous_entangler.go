package t

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// TangleclawWerewolf is the card definition for Tangleclaw Werewolf // Fibrous Entangler.
//
// Type: Creature — Werewolf Horror // Creature — Eldrazi Werewolf
// Face: Fibrous Entangler — Creature — Eldrazi Werewolf
//
// Oracle text:
//
//	This creature can block an additional creature each combat.
//	{6}{G}: Transform this creature.
var TangleclawWerewolf = newTangleclawWerewolf

func newTangleclawWerewolf() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Green),
		CardFace: game.CardFace{
			Name: "Tangleclaw Werewolf",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.G,
				cost.G,
			}),
			Colors:    []color.Color{color.Green},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Werewolf, types.Horror},
			Power:     opt.Val(game.PT{Value: 2}),
			Toughness: opt.Val(game.PT{Value: 4}),
			StaticAbilities: []game.StaticAbility{
				game.StaticAbility{
					Text: "This creature can block an additional creature each combat.",
					RuleEffects: []game.RuleEffect{
						game.RuleEffect{
							Kind:                 game.RuleEffectCanBlockAdditional,
							AffectedSource:       true,
							Duration:             game.DurationPermanent,
							AdditionalBlockCount: 1,
						},
					},
				},
			},
			ActivatedAbilities: []game.ActivatedAbility{
				game.ActivatedAbility{
					Text: "{6}{G}: Transform this creature.",
					ManaCost: opt.Val(cost.Mana{
						cost.O(6),
						cost.G,
					}),
					ZoneOfFunction: zone.Battlefield,
					Content: game.AbilityContent{
						Modes: []game.Mode{
							game.Mode{
								Sequence: []game.Instruction{
									game.Instruction{
										Primitive: game.Transform{
											Object: game.SourcePermanentReference(),
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
			This creature can block an additional creature each combat.
			{6}{G}: Transform this creature.
		`,
		},
		Layout: game.LayoutTransform,
		Back: opt.Val(game.CardFace{
			Name:      "Fibrous Entangler",
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Eldrazi, types.Werewolf},
			Power:     opt.Val(game.PT{Value: 4}),
			Toughness: opt.Val(game.PT{Value: 6}),
			StaticAbilities: []game.StaticAbility{
				game.VigilanceStaticBody,
				game.MustBeBlockedStaticBody,
				game.StaticAbility{
					Text: "This creature can block an additional creature each combat.",
					RuleEffects: []game.RuleEffect{
						game.RuleEffect{
							Kind:                 game.RuleEffectCanBlockAdditional,
							AffectedSource:       true,
							Duration:             game.DurationPermanent,
							AdditionalBlockCount: 1,
						},
					},
				},
			},
			OracleText: `
			Vigilance
			This creature must be blocked if able.
			This creature can block an additional creature each combat.
		`,
		}),
	}
}
