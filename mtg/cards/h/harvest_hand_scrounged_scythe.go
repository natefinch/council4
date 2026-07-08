package h

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// HarvestHand is the card definition for Harvest Hand // Scrounged Scythe.
//
// Type: Artifact Creature — Scarecrow // Artifact — Equipment
// Face: Scrounged Scythe — Artifact — Equipment
//
// Oracle text:
//
//	When this creature dies, return it to the battlefield transformed under your control.
var HarvestHand = newHarvestHand()

func newHarvestHand() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name: "Harvest Hand",
			ManaCost: opt.Val(cost.Mana{
				cost.O(3),
			}),
			Types:     []types.Card{types.Artifact, types.Creature},
			Subtypes:  []types.Sub{types.Scarecrow},
			Power:     opt.Val(game.PT{Value: 2}),
			Toughness: opt.Val(game.PT{Value: 2}),
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Trigger: game.TriggerCondition{
						Type: game.TriggerWhen,
						Pattern: game.TriggerPattern{
							Event:            game.EventPermanentDied,
							Source:           game.TriggerSourceSelf,
							SubjectSelection: game.Selection{RequiredTypes: []types.Card{types.Creature}},
						},
					},
					Content: game.Mode{
						Sequence: []game.Instruction{
							{
								Primitive: game.PutOnBattlefield{
									Source:           game.CardBattlefieldSource(game.CardReference{Kind: game.CardReferenceEvent}),
									EntryTransformed: true,
								},
							},
						},
					}.Ability(),
				},
			},
			OracleText: `
			When this creature dies, return it to the battlefield transformed under your control.
		`,
		},
		Layout: game.LayoutTransform,
		Back: opt.Val(game.CardFace{
			Name:     "Scrounged Scythe",
			Types:    []types.Card{types.Artifact},
			Subtypes: []types.Sub{types.Equipment},
			StaticAbilities: []game.StaticAbility{
				game.StaticAbility{
					ContinuousEffects: []game.ContinuousEffect{
						game.ContinuousEffect{
							Layer:          game.LayerPowerToughnessModify,
							Group:          game.AttachedObjectGroup(game.SourcePermanentReference()),
							PowerDelta:     1,
							ToughnessDelta: 1,
						},
					},
				},
				game.StaticAbility{
					Condition: opt.Val(game.Condition{
						Object:        opt.Val(game.SourceAttachedPermanentReference()),
						ObjectMatches: opt.Val(game.Selection{SubtypesAny: []types.Sub{types.Sub("Human")}}),
					}),
					ContinuousEffects: []game.ContinuousEffect{
						game.ContinuousEffect{
							Layer: game.LayerAbility,
							Group: game.AttachedObjectGroup(game.SourcePermanentReference()),
							AddKeywords: []game.Keyword{
								game.Menace,
							},
						},
					},
				},
			},
			ActivatedAbilities: []game.ActivatedAbility{
				game.EquipActivatedAbility(cost.Mana{cost.O(2)}),
			},
			OracleText: `
			Equipped creature gets +1/+1.
			As long as equipped creature is a Human, it has menace. (It can't be blocked except by two or more creatures.)
			Equip {2} ({2}: Attach to target creature you control. Equip only as a sorcery.)
		`,
		}),
	}
}
