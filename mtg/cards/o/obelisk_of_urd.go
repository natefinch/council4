package o

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// ObeliskOfUrd is the card definition for Obelisk of Urd.
//
// Type: Artifact
// Cost: {6}
//
// Oracle text:
//
//	Convoke (Your creatures can help cast this spell. Each creature you tap while casting this spell pays for {1} or one mana of that creature's color.)
//	As this artifact enters, choose a creature type.
//	Creatures you control of the chosen type get +2/+2.
var ObeliskOfUrd = newObeliskOfUrd

func newObeliskOfUrd() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name: "Obelisk of Urd",
			ManaCost: opt.Val(cost.Mana{
				cost.O(6),
			}),
			Types: []types.Card{types.Artifact},
			StaticAbilities: []game.StaticAbility{
				game.ConvokeStaticBody,
				game.StaticAbility{
					Text: "Creatures you control of the chosen type get +2/+2.",
					ContinuousEffects: []game.ContinuousEffect{
						game.ContinuousEffect{
							Duration:       game.DurationPermanent,
							Group:          game.ObjectControlledGroup(game.SourcePermanentReference(), game.Selection{RequiredTypes: []types.Card{types.Creature}, SubtypeChoice: game.SubtypeChoiceSourceEntry}),
							Layer:          game.LayerPowerToughnessModify,
							PowerDelta:     2,
							ToughnessDelta: 2,
						},
					},
				},
			},
			ReplacementAbilities: []game.ReplacementAbility{
				game.EntryTypeChoiceReplacement("As this artifact enters, choose a creature type."),
			},
			OracleText: `
			Convoke (Your creatures can help cast this spell. Each creature you tap while casting this spell pays for {1} or one mana of that creature's color.)
			As this artifact enters, choose a creature type.
			Creatures you control of the chosen type get +2/+2.
		`,
		},
	}
}
