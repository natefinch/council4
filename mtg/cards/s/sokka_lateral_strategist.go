package s

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// SokkaLateralStrategist is the card definition for Sokka, Lateral Strategist.
//
// Type: Legendary Creature — Human Warrior Ally
// Cost: {1}{W/U}{W/U}
//
// Oracle text:
//
//	Vigilance
//	Whenever Sokka and at least one other creature attack, draw a card.
var SokkaLateralStrategist = newSokkaLateralStrategist

func newSokkaLateralStrategist() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White, color.Blue),
		CardFace: game.CardFace{
			Name: "Sokka, Lateral Strategist",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
				cost.HybridMana(mana.W, mana.U),
				cost.HybridMana(mana.W, mana.U),
			}),
			Colors:     []color.Color{color.Blue, color.White},
			Supertypes: []types.Super{types.Legendary},
			Types:      []types.Card{types.Creature},
			Subtypes:   []types.Sub{types.Human, types.Warrior, types.Ally},
			Power:      opt.Val(game.PT{Value: 2}),
			Toughness:  opt.Val(game.PT{Value: 4}),
			StaticAbilities: []game.StaticAbility{
				game.VigilanceStaticBody,
			},
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Trigger: game.TriggerCondition{
						Type: game.TriggerWhenever,
						Pattern: game.TriggerPattern{
							Event:                     game.EventAttackerDeclared,
							Source:                    game.TriggerSourceSelf,
							AttacksAlongsideCount:     1,
							AttacksAlongsideSelection: game.Selection{RequiredTypes: []types.Card{types.Creature}},
						},
					},
					Content: game.Mode{
						Sequence: []game.Instruction{
							{
								Primitive: game.Draw{
									Amount: game.Fixed(1),
									Player: game.ControllerReference(),
								},
							},
						},
					}.Ability(),
				},
			},
			OracleText: `
			Vigilance
			Whenever Sokka and at least one other creature attack, draw a card.
		`,
		},
	}
}
