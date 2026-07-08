package d

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// DredgerSInsight is the card definition for Dredger's Insight.
//
// Type: Enchantment
// Cost: {1}{G}
//
// Oracle text:
//
//	Whenever one or more artifact and/or creature cards leave your graveyard, you gain 1 life.
//	When this enchantment enters, mill four cards. You may put an artifact, creature, or land card from among the milled cards into your hand. (To mill four cards, put the top four cards of your library into your graveyard.)
var DredgerSInsight = newDredgerSInsight

func newDredgerSInsight() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Green),
		CardFace: game.CardFace{
			Name: "Dredger's Insight",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
				cost.G,
			}),
			Colors: []color.Color{color.Green},
			Types:  []types.Card{types.Enchantment},
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Trigger: game.TriggerCondition{
						Type: game.TriggerWhenever,
						Pattern: game.TriggerPattern{
							Event:            game.EventZoneChanged,
							Player:           game.TriggerPlayerYou,
							MatchFromZone:    true,
							FromZone:         zone.Graveyard,
							OneOrMore:        true,
							SubjectSelection: game.Selection{RequiredTypesAny: []types.Card{types.Artifact, types.Creature}},
						},
					},
					Content: game.Mode{
						Sequence: []game.Instruction{
							{
								Primitive: game.GainLife{
									Amount: game.Fixed(1),
									Player: game.ControllerReference(),
								},
							},
						},
					}.Ability(),
				},
				game.TriggeredAbility{
					Trigger: game.TriggerCondition{
						Type: game.TriggerWhen,
						Pattern: game.TriggerPattern{
							Event:  game.EventPermanentEnteredBattlefield,
							Source: game.TriggerSourceSelf,
						},
					},
					Content: game.Mode{
						Sequence: []game.Instruction{
							{
								Primitive: game.Dig{
									Player:   game.ControllerReference(),
									Look:     game.Fixed(4),
									Take:     game.Fixed(1),
									Filter:   opt.Val(game.Selection{RequiredTypesAny: []types.Card{types.Artifact, types.Creature, types.Land}}),
									TakeUpTo: true,
									Reveal:   true,
								},
							},
						},
					}.Ability(),
				},
			},
			OracleText: `
			Whenever one or more artifact and/or creature cards leave your graveyard, you gain 1 life.
			When this enchantment enters, mill four cards. You may put an artifact, creature, or land card from among the milled cards into your hand. (To mill four cards, put the top four cards of your library into your graveyard.)
		`,
		},
	}
}
