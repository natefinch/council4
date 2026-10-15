package q

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// QuickDrawKatana is the card definition for Quick-Draw Katana.
//
// Type: Artifact — Equipment
// Cost: {2}
//
// Oracle text:
//
//	During your turn, equipped creature gets +2/+0 and has first strike. (It deals combat damage before creatures without first strike.)
//	Equip {2} ({2}: Attach to target creature you control. Equip only as a sorcery.)
var QuickDrawKatana = newQuickDrawKatana

func newQuickDrawKatana() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name: "Quick-Draw Katana",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
			}),
			Types:    []types.Card{types.Artifact},
			Subtypes: []types.Sub{types.Equipment},
			StaticAbilities: []game.StaticAbility{
				game.StaticAbility{
					Text: "During your turn, equipped creature gets +2/+0 and has first strike. (It deals combat damage before creatures without first strike.)",
					Condition: opt.Val(game.Condition{
						Text:                 "During your turn",
						SourceCounterKind:    counter.PlusOnePlusOne,
						SourceControllerTurn: true,
					}),
					ContinuousEffects: []game.ContinuousEffect{
						game.ContinuousEffect{
							Duration:   game.DurationPermanent,
							Group:      game.AttachedObjectGroup(game.SourcePermanentReference()),
							Layer:      game.LayerPowerToughnessModify,
							PowerDelta: 2,
						},
						game.ContinuousEffect{
							Duration:    game.DurationPermanent,
							Group:       game.AttachedObjectGroup(game.SourcePermanentReference()),
							Layer:       game.LayerAbility,
							AddKeywords: []game.Keyword{game.FirstStrike},
						},
					},
				},
			},
			ActivatedAbilities: []game.ActivatedAbility{
				game.EquipActivatedAbility(cost.Mana{cost.O(2)}),
			},
			OracleText: `
			During your turn, equipped creature gets +2/+0 and has first strike. (It deals combat damage before creatures without first strike.)
			Equip {2} ({2}: Attach to target creature you control. Equip only as a sorcery.)
		`,
		},
	}
}
