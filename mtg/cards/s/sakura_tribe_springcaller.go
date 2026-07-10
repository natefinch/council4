package s

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// SakuraTribeSpringcaller is the card definition for Sakura-Tribe Springcaller.
//
// Type: Creature — Snake Shaman
// Cost: {3}{G}
//
// Oracle text:
//
//	At the beginning of your upkeep, add {G}. Until end of turn, you don't lose this mana as steps and phases end.
var SakuraTribeSpringcaller = newSakuraTribeSpringcaller

func newSakuraTribeSpringcaller() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Green),
		CardFace: game.CardFace{
			Name: "Sakura-Tribe Springcaller",
			ManaCost: opt.Val(cost.Mana{
				cost.O(3),
				cost.G,
			}),
			Colors:    []color.Color{color.Green},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Snake, types.Shaman},
			Power:     opt.Val(game.PT{Value: 2}),
			Toughness: opt.Val(game.PT{Value: 4}),
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Trigger: game.TriggerCondition{
						Type: game.TriggerAt,
						Pattern: game.TriggerPattern{
							Event:      game.EventBeginningOfStep,
							Controller: game.TriggerControllerYou,
							Step:       game.StepUpkeep,
						},
					},
					Content: game.Mode{
						Sequence: []game.Instruction{
							{
								Primitive: game.AddMana{
									Amount:                game.Fixed(1),
									ManaColor:             mana.G,
									PersistUntilEndOfTurn: true,
								},
							},
						},
					}.Ability(),
				},
			},
			OracleText: `
			At the beginning of your upkeep, add {G}. Until end of turn, you don't lose this mana as steps and phases end.
		`,
		},
	}
}
