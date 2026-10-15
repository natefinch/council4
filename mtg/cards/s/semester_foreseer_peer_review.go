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

// SemesterForeseer is the card definition for Semester Foreseer // Peer Review.
//
// Type: Creature — Human Wizard // Sorcery
// Cost: {3}{U} // {2}{W/U}
// Face: Peer Review — Sorcery ({2}{W/U})
//
// Oracle text:
//
//	This creature enters prepared. (While it's prepared, you may cast a copy of its spell. Doing so unprepares it.)
//	When this creature enters, surveil 1.
var SemesterForeseer = newSemesterForeseer

func newSemesterForeseer() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White, color.Blue),
		CardFace: game.CardFace{
			Name: "Semester Foreseer",
			ManaCost: opt.Val(cost.Mana{
				cost.O(3),
				cost.U,
			}),
			Colors:         []color.Color{color.Blue},
			EntersPrepared: true,
			Types:          []types.Card{types.Creature},
			Subtypes:       []types.Sub{types.Human, types.Wizard},
			Power:          opt.Val(game.PT{Value: 3}),
			Toughness:      opt.Val(game.PT{Value: 4}),
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Text: "When this creature enters, surveil 1.",
					Trigger: game.TriggerCondition{
						Type: game.TriggerWhen,
						Pattern: game.TriggerPattern{
							Event:           game.EventPermanentEnteredBattlefield,
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
										Primitive: game.Surveil{
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
			This creature enters prepared. (While it's prepared, you may cast a copy of its spell. Doing so unprepares it.)
			When this creature enters, surveil 1.
		`,
		},
		Layout: game.LayoutPrepare,
		Alternate: opt.Val(game.CardFace{
			Name: "Peer Review",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.HybridMana(mana.W, mana.U),
			}),
			Colors: []color.Color{color.White, color.Blue},
			Types:  []types.Card{types.Sorcery},
			SpellAbility: opt.Val(game.Mode{
				Sequence: []game.Instruction{
					{
						Primitive: game.CreateToken{
							Amount: game.Fixed(1),
							Source: game.TokenDef(semesterForeseerToken),
						},
					},
					{
						Primitive: game.Surveil{
							Amount: game.Fixed(1),
							Player: game.ControllerReference(),
						},
					},
				},
			}.Ability()),
			OracleText: `
			Create a 2/2 colorless Wizard Soldier creature token named Cadet. Surveil 1.
		`,
		}),
	}
}

var semesterForeseerToken = newSemesterForeseerToken()

func newSemesterForeseerToken() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name:      "Cadet",
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Wizard, types.Soldier},
			Power:     opt.Val(game.PT{Value: 2}),
			Toughness: opt.Val(game.PT{Value: 2}),
		},
	}
}
