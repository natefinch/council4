package l

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// LockjawSnapper is the card definition for Lockjaw Snapper.
//
// Type: Artifact Creature — Scarecrow
// Cost: {4}
//
// Oracle text:
//
//	Wither (This deals damage to creatures in the form of -1/-1 counters.)
//	When this creature dies, put a -1/-1 counter on each creature with a -1/-1 counter on it.
var LockjawSnapper = newLockjawSnapper

func newLockjawSnapper() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name: "Lockjaw Snapper",
			ManaCost: opt.Val(cost.Mana{
				cost.O(4),
			}),
			Types:     []types.Card{types.Artifact, types.Creature},
			Subtypes:  []types.Sub{types.Scarecrow},
			Power:     opt.Val(game.PT{Value: 2}),
			Toughness: opt.Val(game.PT{Value: 2}),
			StaticAbilities: []game.StaticAbility{
				game.WitherStaticBody,
			},
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Text: "When this creature dies, put a -1/-1 counter on each creature with a -1/-1 counter on it.",
					Trigger: game.TriggerCondition{
						Type: game.TriggerWhen,
						Pattern: game.TriggerPattern{
							Event:            game.EventPermanentDied,
							Source:           game.TriggerSourceSelf,
							SubjectSelection: game.Selection{RequiredTypes: []types.Card{types.Creature}},
							StackObjectKind:  game.StackSpell,
							CounterKind:      counter.PlusOnePlusOne,
						},
					},
					Content: game.AbilityContent{
						Modes: []game.Mode{
							game.Mode{
								Sequence: []game.Instruction{
									game.Instruction{
										Primitive: game.AddCounter{
											Amount:      game.Fixed(1),
											Group:       game.BattlefieldGroup(game.Selection{RequiredTypes: []types.Card{types.Creature}, MatchCounter: true, RequiredCounter: counter.MinusOneMinusOne}),
											CounterKind: counter.MinusOneMinusOne,
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
			Wither (This deals damage to creatures in the form of -1/-1 counters.)
			When this creature dies, put a -1/-1 counter on each creature with a -1/-1 counter on it.
		`,
		},
	}
}
