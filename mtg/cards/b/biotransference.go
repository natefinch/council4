package b

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// Biotransference is the card definition for Biotransference.
//
// Type: Enchantment
// Cost: {2}{B}{B}
//
// Oracle text:
//
//	Creatures you control are artifacts in addition to their other types. The same is true for creature spells you control and creature cards you own that aren't on the battlefield.
//	Whenever you cast an artifact spell, you lose 1 life and create a 2/2 black Necron Warrior artifact creature token.
var Biotransference = newBiotransference

func newBiotransference() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Black),
		CardFace: game.CardFace{
			Name: "Biotransference",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.B,
				cost.B,
			}),
			Colors: []color.Color{color.Black},
			Types:  []types.Card{types.Enchantment},
			StaticAbilities: []game.StaticAbility{
				game.StaticAbility{
					Text: "Creatures you control are artifacts in addition to their other types. The same is true for creature spells you control and creature cards you own that aren't on the battlefield.",
					ContinuousEffects: []game.ContinuousEffect{
						game.ContinuousEffect{
							Duration:       game.DurationPermanent,
							Group:          game.ObjectControlledGroup(game.SourcePermanentReference(), game.Selection{RequiredTypes: []types.Card{types.Creature}}),
							Layer:          game.LayerType,
							SpellSelection: &game.Selection{RequiredTypes: []types.Card{types.Creature}},
							AddTypes:       []types.Card{types.Artifact},
						},
					},
				},
			},
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Text: "Whenever you cast an artifact spell, you lose 1 life and create a 2/2 black Necron Warrior artifact creature token.",
					Trigger: game.TriggerCondition{
						Type: game.TriggerWhenever,
						Pattern: game.TriggerPattern{
							Event:           game.EventSpellCast,
							Controller:      game.TriggerControllerYou,
							CardSelection:   game.Selection{RequiredTypes: []types.Card{types.Artifact}},
							StackObjectKind: game.StackSpell,
							CounterKind:     counter.PlusOnePlusOne,
						},
					},
					Content: game.AbilityContent{
						Modes: []game.Mode{
							game.Mode{
								Sequence: []game.Instruction{
									game.Instruction{
										Primitive: game.LoseLife{
											Amount: game.Fixed(1),
											Player: game.ControllerReference(),
										},
									},
									game.Instruction{
										Primitive: game.CreateToken{
											Amount: game.Fixed(1),
											Source: game.TokenDef(biotransferenceToken),
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
			Creatures you control are artifacts in addition to their other types. The same is true for creature spells you control and creature cards you own that aren't on the battlefield.
			Whenever you cast an artifact spell, you lose 1 life and create a 2/2 black Necron Warrior artifact creature token.
		`,
		},
	}
}

var biotransferenceToken = newBiotransferenceToken()

func newBiotransferenceToken() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name:      "Necron Warrior",
			Colors:    []color.Color{color.Black},
			Types:     []types.Card{types.Artifact, types.Creature},
			Subtypes:  []types.Sub{types.Necron, types.Warrior},
			Power:     opt.Val(game.PT{Value: 2}),
			Toughness: opt.Val(game.PT{Value: 2}),
		},
	}
}
