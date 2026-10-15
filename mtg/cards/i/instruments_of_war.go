package i

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// InstrumentsOfWar is the card definition for Instruments of War.
//
// Type: Artifact
// Cost: {4}
//
// Oracle text:
//
//	Flash
//	As this artifact enters, choose a creature type.
//	Creatures you control of the chosen type get +1/+1.
var InstrumentsOfWar = newInstrumentsOfWar

func newInstrumentsOfWar() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name: "Instruments of War",
			ManaCost: opt.Val(cost.Mana{
				cost.O(4),
			}),
			Types: []types.Card{types.Artifact},
			StaticAbilities: []game.StaticAbility{
				game.FlashStaticBody,
				game.StaticAbility{
					Text: "Creatures you control of the chosen type get +1/+1.",
					ContinuousEffects: []game.ContinuousEffect{
						game.ContinuousEffect{
							Duration:       game.DurationPermanent,
							Group:          game.ObjectControlledGroup(game.SourcePermanentReference(), game.Selection{RequiredTypes: []types.Card{types.Creature}, SubtypeChoice: game.SubtypeChoiceSourceEntry}),
							Layer:          game.LayerPowerToughnessModify,
							PowerDelta:     1,
							ToughnessDelta: 1,
						},
					},
				},
			},
			ReplacementAbilities: []game.ReplacementAbility{
				game.EntryTypeChoiceReplacement("As this artifact enters, choose a creature type."),
			},
			OracleText: `
			Flash
			As this artifact enters, choose a creature type.
			Creatures you control of the chosen type get +1/+1.
		`,
		},
	}
}
