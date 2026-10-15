package s

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// SukiKyoshiWarrior is the card definition for Suki, Kyoshi Warrior.
//
// Type: Legendary Creature — Human Warrior Ally
// Cost: {2}{G/W}{G/W}
//
// Oracle text:
//
//	Suki's power is equal to the number of creatures you control.
//	Whenever Suki attacks, create a 1/1 white Ally creature token that's tapped and attacking.
var SukiKyoshiWarrior = newSukiKyoshiWarrior

func newSukiKyoshiWarrior() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White, color.Green),
		CardFace: game.CardFace{
			Name: "Suki, Kyoshi Warrior",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.HybridMana(mana.G, mana.W),
				cost.HybridMana(mana.G, mana.W),
			}),
			Colors:       []color.Color{color.Green, color.White},
			Supertypes:   []types.Super{types.Legendary},
			Types:        []types.Card{types.Creature},
			Subtypes:     []types.Sub{types.Human, types.Warrior, types.Ally},
			Power:        opt.Val(game.PT{IsStar: true}),
			Toughness:    opt.Val(game.PT{Value: 4}),
			DynamicPower: opt.Val(game.DynamicValue{Kind: game.DynamicValueControllerCreatureCount}),
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Text: "Whenever Suki attacks, create a 1/1 white Ally creature token that's tapped and attacking.",
					Trigger: game.TriggerCondition{
						Type: game.TriggerWhenever,
						Pattern: game.TriggerPattern{
							Event:           game.EventAttackerDeclared,
							Source:          game.TriggerSourceSelf,
							StackObjectKind: game.StackSpell,
							CounterKind:     counter.PlusOnePlusOne,
						},
					},
					Content: game.AbilityContent{
						Modes: []game.Mode{
							game.Mode{
								Sequence: []game.Instruction{
									game.Instruction{
										Primitive: game.CreateToken{
											Amount:         game.Fixed(1),
											Source:         game.TokenDef(sukiKyoshiWarriorToken),
											EntryTapped:    true,
											EntryAttacking: true,
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
			Suki's power is equal to the number of creatures you control.
			Whenever Suki attacks, create a 1/1 white Ally creature token that's tapped and attacking.
		`,
		},
	}
}

var sukiKyoshiWarriorToken = newSukiKyoshiWarriorToken()

func newSukiKyoshiWarriorToken() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name:      "Ally",
			Colors:    []color.Color{color.White},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Ally},
			Power:     opt.Val(game.PT{Value: 1}),
			Toughness: opt.Val(game.PT{Value: 1}),
		},
	}
}
