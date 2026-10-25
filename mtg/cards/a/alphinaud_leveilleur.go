package a

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// AlphinaudLeveilleur is the card definition for Alphinaud Leveilleur.
//
// Type: Legendary Creature — Elf Wizard
// Cost: {3}{U}
//
// Oracle text:
//
//	Partner with Alisaie Leveilleur (When this creature enters, target player may put Alisaie Leveilleur into their hand from their library, then shuffle.)
//	Vigilance
//	Eukrasia — Whenever you cast your second spell each turn, draw a card.
var AlphinaudLeveilleur = newAlphinaudLeveilleur

func newAlphinaudLeveilleur() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue),
		CardFace: game.CardFace{
			Name: "Alphinaud Leveilleur",
			ManaCost: opt.Val(cost.Mana{
				cost.O(3),
				cost.U,
			}),
			Colors:     []color.Color{color.Blue},
			Supertypes: []types.Super{types.Legendary},
			Types:      []types.Card{types.Creature},
			Subtypes:   []types.Sub{types.Elf, types.Wizard},
			Power:      opt.Val(game.PT{Value: 2}),
			Toughness:  opt.Val(game.PT{Value: 4}),
			StaticAbilities: []game.StaticAbility{
				game.PartnerWithStaticBody,
				game.VigilanceStaticBody,
			},
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Text: "Eukrasia — Whenever you cast your second spell each turn, draw a card.",
					Trigger: game.TriggerCondition{
						Type: game.TriggerWhenever,
						Pattern: game.TriggerPattern{
							Event:                      game.EventSpellCast,
							Controller:                 game.TriggerControllerYou,
							StackObjectKind:            game.StackSpell,
							PlayerEventOrdinalThisTurn: 2,
							CounterKind:                counter.PlusOnePlusOne,
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
				game.PartnerWithTriggeredAbility("Alisaie Leveilleur"),
			},
			OracleText: `
			Partner with Alisaie Leveilleur (When this creature enters, target player may put Alisaie Leveilleur into their hand from their library, then shuffle.)
			Vigilance
			Eukrasia — Whenever you cast your second spell each turn, draw a card.
		`,
		},
	}
}
