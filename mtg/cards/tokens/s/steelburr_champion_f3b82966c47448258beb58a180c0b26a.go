package s

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// SteelburrChampion is the card definition for Steelburr Champion.
//
// Type: Token Creature — Mouse Soldier
// Cost: {2}{W}
//
// Oracle text:
//   Vigilance
//   Whenever an opponent casts a noncreature spell, put a +1/+1 counter on this creature. (This token's mana cost is {2}{W}.)

// SteelburrChampionTokenf3b82966c47448258beb58a180c0b26a is the card definition for Steelburr Champion.
var SteelburrChampionTokenf3b82966c47448258beb58a180c0b26a = newSteelburrChampionTokenf3b82966c47448258beb58a180c0b26a()

func newSteelburrChampionTokenf3b82966c47448258beb58a180c0b26a() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White),
		CardFace: game.CardFace{
			Name: "Steelburr Champion",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.W,
			}),
			Colors:    []color.Color{color.White},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Mouse, types.Soldier},
			Power:     opt.Val(game.PT{Value: 1}),
			Toughness: opt.Val(game.PT{Value: 1}),
			StaticAbilities: []game.StaticAbility{
				game.VigilanceStaticBody,
			},
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Text: "Whenever an opponent casts a noncreature spell, put a +1/+1 counter on this creature. (This token's mana cost is {2}{W}.)",
					Trigger: game.TriggerCondition{
						Type: game.TriggerWhenever,
						Pattern: game.TriggerPattern{
							Event:           game.EventSpellCast,
							Controller:      game.TriggerControllerOpponent,
							CardSelection:   game.Selection{ExcludedTypes: []types.Card{types.Creature}},
							StackObjectKind: game.StackSpell,
							CounterKind:     counter.PlusOnePlusOne,
						},
					},
					Content: game.AbilityContent{
						Modes: []game.Mode{
							game.Mode{
								Sequence: []game.Instruction{
									game.Instruction{
										Primitive: game.AddCounter{
											Amount:      game.Fixed(1),
											Object:      game.SourcePermanentReference(),
											CounterKind: counter.PlusOnePlusOne,
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
			Vigilance
			Whenever an opponent casts a noncreature spell, put a +1/+1 counter on this creature. (This token's mana cost is {2}{W}.)
		`,
		},
	}
}
