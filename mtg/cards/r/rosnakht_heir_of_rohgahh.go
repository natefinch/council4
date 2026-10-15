package r

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// RosnakhtHeirOfRohgahh is the card definition for Rosnakht, Heir of Rohgahh.
//
// Type: Legendary Creature — Kobold Warrior
// Cost: {R}
//
// Oracle text:
//
//	Battle cry (Whenever this creature attacks, each other attacking creature gets +1/+0 until end of turn.)
//	Heroic — Whenever you cast a spell that targets Rosnakht, create a 0/1 red Kobold creature token named Kobolds of Kher Keep.
var RosnakhtHeirOfRohgahh = newRosnakhtHeirOfRohgahh

func newRosnakhtHeirOfRohgahh() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Red),
		CardFace: game.CardFace{
			Name: "Rosnakht, Heir of Rohgahh",
			ManaCost: opt.Val(cost.Mana{
				cost.R,
			}),
			Colors:     []color.Color{color.Red},
			Supertypes: []types.Super{types.Legendary},
			Types:      []types.Card{types.Creature},
			Subtypes:   []types.Sub{types.Kobold, types.Warrior},
			Power:      opt.Val(game.PT{Value: 0}),
			Toughness:  opt.Val(game.PT{Value: 1}),
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Text: "Whenever this creature attacks, each other attacking creature gets +1/+0 until end of turn.",
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
										Primitive: game.ApplyContinuous{
											ContinuousEffects: []game.ContinuousEffect{
												game.ContinuousEffect{
													Duration:   game.DurationPermanent,
													Group:      game.BattlefieldGroupExcluding(game.Selection{RequiredTypes: []types.Card{types.Creature}, CombatState: game.CombatStateAttacking}, game.SourcePermanentReference()),
													Layer:      game.LayerPowerToughnessModify,
													PowerDelta: 1,
												},
											},
											Duration: game.DurationUntilEndOfTurn,
										},
									},
								},
							},
						},
						MinModes: 1,
						MaxModes: 1,
					},
				},
				game.TriggeredAbility{
					Text: "Heroic — Whenever you cast a spell that targets Rosnakht, create a 0/1 red Kobold creature token named Kobolds of Kher Keep.",
					Trigger: game.TriggerCondition{
						Type: game.TriggerWhenever,
						Pattern: game.TriggerPattern{
							Event:              game.EventSpellCast,
							Controller:         game.TriggerControllerYou,
							StackObjectKind:    game.StackSpell,
							SpellTargetsSource: true,
							CounterKind:        counter.PlusOnePlusOne,
						},
					},
					Content: game.AbilityContent{
						Modes: []game.Mode{
							game.Mode{
								Sequence: []game.Instruction{
									game.Instruction{
										Primitive: game.CreateToken{
											Amount: game.Fixed(1),
											Source: game.TokenDef(rosnakhtHeirOfRohgahhToken),
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
			Battle cry (Whenever this creature attacks, each other attacking creature gets +1/+0 until end of turn.)
			Heroic — Whenever you cast a spell that targets Rosnakht, create a 0/1 red Kobold creature token named Kobolds of Kher Keep.
		`,
		},
	}
}

var rosnakhtHeirOfRohgahhToken = newRosnakhtHeirOfRohgahhToken()

func newRosnakhtHeirOfRohgahhToken() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name:      "Kobolds of Kher Keep",
			Colors:    []color.Color{color.Red},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Kobold},
			Power:     opt.Val(game.PT{Value: 0}),
			Toughness: opt.Val(game.PT{Value: 1}),
		},
	}
}
