package k

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// KarnGildedGuardian is the card definition for Karn, Gilded Guardian.
//
// Type: Legendary Artifact Creature — Golem
// Cost: {2/W}{2/U}{2/B}{2/R}{2/G}
//
// Oracle text:
//
//	Vigilance, trample
//	When Karn enters, draw a card for each color among other artifacts you control.
var KarnGildedGuardian = newKarnGildedGuardian

func newKarnGildedGuardian() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White, color.Blue, color.Black, color.Red, color.Green),
		CardFace: game.CardFace{
			Name: "Karn, Gilded Guardian",
			ManaCost: opt.Val(cost.Mana{
				cost.Twobrid(mana.W),
				cost.Twobrid(mana.U),
				cost.Twobrid(mana.B),
				cost.Twobrid(mana.R),
				cost.Twobrid(mana.G),
			}),
			Colors:     []color.Color{color.Black, color.Green, color.Red, color.Blue, color.White},
			Supertypes: []types.Super{types.Legendary},
			Types:      []types.Card{types.Artifact, types.Creature},
			Subtypes:   []types.Sub{types.Golem},
			Power:      opt.Val(game.PT{Value: 5}),
			Toughness:  opt.Val(game.PT{Value: 5}),
			StaticAbilities: []game.StaticAbility{
				game.VigilanceStaticBody,
				game.TrampleStaticBody,
			},
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Text: "When Karn enters, draw a card for each color among other artifacts you control.",
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
										Primitive: game.Draw{
											Amount: game.Dynamic(game.DynamicAmount{
												Kind:       game.DynamicAmountColorCountInGroup,
												Multiplier: 1,
												Group:      game.BattlefieldGroup(game.Selection{RequiredTypes: []types.Card{types.Artifact}, Controller: game.ControllerYou, ExcludeSource: true}),
											}),
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
			Vigilance, trample
			When Karn enters, draw a card for each color among other artifacts you control.
		`,
		},
	}
}
