package r

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// RuricTharBiomagus is the card definition for Ruric Thar, Biomagus.
//
// Type: Legendary Creature — Ogre Crab Wizard
// Cost: {4}{U}{U}
//
// Oracle text:
//
//	Flying
//	Prowess, prowess (Whenever you cast a noncreature spell, this creature gets +1/+1 until end of turn twice.)
//	Whenever Ruric Thar becomes the target of a spell or ability an opponent controls, draw a card.
var RuricTharBiomagus = newRuricTharBiomagus

func newRuricTharBiomagus() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue),
		CardFace: game.CardFace{
			Name: "Ruric Thar, Biomagus",
			ManaCost: opt.Val(cost.Mana{
				cost.O(4),
				cost.U,
				cost.U,
			}),
			Colors:     []color.Color{color.Blue},
			Supertypes: []types.Super{types.Legendary},
			Types:      []types.Card{types.Creature},
			Subtypes:   []types.Sub{types.Ogre, types.Crab, types.Wizard},
			Power:      opt.Val(game.PT{Value: 4}),
			Toughness:  opt.Val(game.PT{Value: 6}),
			StaticAbilities: []game.StaticAbility{
				game.FlyingStaticBody,
				game.ProwessStaticBody,
				game.ProwessStaticBody,
			},
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Text: "Whenever Ruric Thar becomes the target of a spell or ability an opponent controls, draw a card.",
					Trigger: game.TriggerCondition{
						Type: game.TriggerWhenever,
						Pattern: game.TriggerPattern{
							Event:           game.EventObjectBecameTarget,
							CauseController: game.TriggerControllerOpponent,
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
										Primitive: game.Draw{
											Amount: game.Fixed(1),
											Player: game.ControllerReference(),
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
			Flying
			Prowess, prowess (Whenever you cast a noncreature spell, this creature gets +1/+1 until end of turn twice.)
			Whenever Ruric Thar becomes the target of a spell or ability an opponent controls, draw a card.
		`,
		},
	}
}
