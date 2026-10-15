package r

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// RoamingThrone is the card definition for Roaming Throne.
//
// Type: Artifact Creature — Golem
// Cost: {4}
//
// Oracle text:
//
//	Ward {2}
//	As this creature enters, choose a creature type.
//	This creature is the chosen type in addition to its other types.
//	If a triggered ability of another creature you control of the chosen type triggers, it triggers an additional time.
var RoamingThrone = newRoamingThrone

func newRoamingThrone() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name: "Roaming Throne",
			ManaCost: opt.Val(cost.Mana{
				cost.O(4),
			}),
			Types:     []types.Card{types.Artifact, types.Creature},
			Subtypes:  []types.Sub{types.Golem},
			Power:     opt.Val(game.PT{Value: 4}),
			Toughness: opt.Val(game.PT{Value: 4}),
			StaticAbilities: []game.StaticAbility{
				game.WardStaticAbility(cost.Mana{cost.O(2)}),
				game.StaticAbility{
					Text: "This creature is the chosen type in addition to its other types.",
					ContinuousEffects: []game.ContinuousEffect{
						game.ContinuousEffect{
							Duration:                  game.DurationPermanent,
							AffectedSource:            true,
							Layer:                     game.LayerType,
							AddSubtypeFromEntryChoice: game.EntryTypeChoiceKey,
						},
					},
				},
				game.StaticAbility{
					Text: "If a triggered ability of another creature you control of the chosen type triggers, it triggers an additional time.",
					RuleEffects: []game.RuleEffect{
						game.RuleEffect{
							Kind:     game.RuleEffectAdditionalTriggerForChosenCreatureType,
							Duration: game.DurationPermanent,
						},
					},
				},
			},
			ReplacementAbilities: []game.ReplacementAbility{
				game.EntryTypeChoiceReplacement("As this creature enters, choose a creature type."),
			},
			OracleText: `
			Ward {2}
			As this creature enters, choose a creature type.
			This creature is the chosen type in addition to its other types.
			If a triggered ability of another creature you control of the chosen type triggers, it triggers an additional time.
		`,
		},
	}
}
