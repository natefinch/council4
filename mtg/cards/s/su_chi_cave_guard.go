package s

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// SuChiCaveGuard is the card definition for Su-Chi Cave Guard.
//
// Type: Artifact Creature — Construct
// Cost: {8}
//
// Oracle text:
//
//	Vigilance
//	Ward {4} (Whenever this creature becomes the target of a spell or ability an opponent controls, counter it unless that player pays {4}.)
//	When this creature dies, add eight {C}. Until end of turn, you don't lose this mana as steps and phases end.
var SuChiCaveGuard = newSuChiCaveGuard

func newSuChiCaveGuard() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name: "Su-Chi Cave Guard",
			ManaCost: opt.Val(cost.Mana{
				cost.O(8),
			}),
			Types:     []types.Card{types.Artifact, types.Creature},
			Subtypes:  []types.Sub{types.Construct},
			Power:     opt.Val(game.PT{Value: 8}),
			Toughness: opt.Val(game.PT{Value: 8}),
			StaticAbilities: []game.StaticAbility{
				game.VigilanceStaticBody,
				game.WardStaticAbility(cost.Mana{cost.O(4)}),
			},
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
								Primitive: game.AddMana{
									Amount:                game.Fixed(1),
									ManaColor:             mana.C,
									PersistUntilEndOfTurn: true,
								},
							},
							{
								Primitive: game.AddMana{
									Amount:                game.Fixed(1),
									ManaColor:             mana.C,
									PersistUntilEndOfTurn: true,
								},
							},
							{
								Primitive: game.AddMana{
									Amount:                game.Fixed(1),
									ManaColor:             mana.C,
									PersistUntilEndOfTurn: true,
								},
							},
							{
								Primitive: game.AddMana{
									Amount:                game.Fixed(1),
									ManaColor:             mana.C,
									PersistUntilEndOfTurn: true,
								},
							},
							{
								Primitive: game.AddMana{
									Amount:                game.Fixed(1),
									ManaColor:             mana.C,
									PersistUntilEndOfTurn: true,
								},
							},
							{
								Primitive: game.AddMana{
									Amount:                game.Fixed(1),
									ManaColor:             mana.C,
									PersistUntilEndOfTurn: true,
								},
							},
							{
								Primitive: game.AddMana{
									Amount:                game.Fixed(1),
									ManaColor:             mana.C,
									PersistUntilEndOfTurn: true,
								},
							},
							{
								Primitive: game.AddMana{
									Amount:                game.Fixed(1),
									ManaColor:             mana.C,
									PersistUntilEndOfTurn: true,
								},
							},
						},
					}.Ability(),
				},
			},
			OracleText: `
			Vigilance
			Ward {4} (Whenever this creature becomes the target of a spell or ability an opponent controls, counter it unless that player pays {4}.)
			When this creature dies, add eight {C}. Until end of turn, you don't lose this mana as steps and phases end.
		`,
		},
	}
}
