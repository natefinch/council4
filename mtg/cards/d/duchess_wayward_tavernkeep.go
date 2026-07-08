package d

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// DuchessWaywardTavernkeep is the card definition for Duchess, Wayward Tavernkeep.
//
// Type: Legendary Creature — Human Citizen
// Cost: {3}{R}
//
// Oracle text:
//
//	Hunters for Hire — Whenever a creature you control deals combat damage to a player, put a quest counter on it.
//	{1}, Remove a quest counter from a permanent you control: Create a Junk token. (It's an artifact with "{T}, Sacrifice this token: Exile the top card of your library. You may play that card this turn. Activate only as a sorcery.")
var DuchessWaywardTavernkeep = newDuchessWaywardTavernkeep

func newDuchessWaywardTavernkeep() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Red),
		CardFace: game.CardFace{
			Name: "Duchess, Wayward Tavernkeep",
			ManaCost: opt.Val(cost.Mana{
				cost.O(3),
				cost.R,
			}),
			Colors:     []color.Color{color.Red},
			Supertypes: []types.Super{types.Legendary},
			Types:      []types.Card{types.Creature},
			Subtypes:   []types.Sub{types.Human, types.Citizen},
			Power:      opt.Val(game.PT{Value: 4}),
			Toughness:  opt.Val(game.PT{Value: 3}),
			ActivatedAbilities: []game.ActivatedAbility{
				game.ActivatedAbility{
					Text:     "{1}, Remove a quest counter from a permanent you control: Create a Junk token. (It's an artifact with \"{T}, Sacrifice this token: Exile the top card of your library. You may play that card this turn. Activate only as a sorcery.\")",
					ManaCost: opt.Val(cost.Mana{cost.O(1)}),
					AdditionalCosts: []cost.Additional{
						{
							Kind:        cost.AdditionalRemoveCounterAmong,
							Text:        "Remove a quest counter from a permanent you control",
							Amount:      1,
							CounterKind: counter.Quest,
						},
					},
					ZoneOfFunction: zone.Battlefield,
					Content: game.Mode{
						Sequence: []game.Instruction{
							{
								Primitive: game.CreateToken{
									Amount: game.Fixed(1),
									Source: game.TokenDef(duchessWaywardTavernkeepToken),
								},
							},
						},
					}.Ability(),
				},
			},
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Trigger: game.TriggerCondition{
						Type: game.TriggerWhenever,
						Pattern: game.TriggerPattern{
							Event:                 game.EventDamageDealt,
							Controller:            game.TriggerControllerYou,
							Subject:               game.TriggerSubjectDamageSource,
							RequireCombatDamage:   true,
							DamageRecipient:       game.DamageRecipientPlayer,
							DamageSourceSelection: game.Selection{RequiredTypes: []types.Card{types.Creature}},
						},
					},
					Content: game.Mode{
						Sequence: []game.Instruction{
							{
								Primitive: game.AddCounter{
									Amount:      game.Fixed(1),
									Object:      game.EventPermanentReference(),
									CounterKind: counter.Quest,
								},
							},
						},
					}.Ability(),
				},
			},
			OracleText: `
			Hunters for Hire — Whenever a creature you control deals combat damage to a player, put a quest counter on it.
			{1}, Remove a quest counter from a permanent you control: Create a Junk token. (It's an artifact with "{T}, Sacrifice this token: Exile the top card of your library. You may play that card this turn. Activate only as a sorcery.")
		`,
		},
	}
}

var duchessWaywardTavernkeepToken = newDuchessWaywardTavernkeepToken()

func newDuchessWaywardTavernkeepToken() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name:     "Junk",
			Types:    []types.Card{types.Artifact},
			Subtypes: []types.Sub{types.Junk},
			ActivatedAbilities: []game.ActivatedAbility{
				game.ActivatedAbility{
					Text: "{T}, Sacrifice this token: Exile the top card of your library. You may play that card this turn. Activate only as a sorcery.",
					AdditionalCosts: []cost.Additional{
						{
							Kind: cost.AdditionalTap,
						},
						{
							Kind:   cost.AdditionalSacrificeSource,
							Text:   "Sacrifice this token",
							Amount: 1,
						},
					},
					ZoneOfFunction: zone.Battlefield,
					Timing:         game.SorceryOnly,
					Content: game.Mode{
						Sequence: []game.Instruction{
							{
								Primitive: game.ImpulseExile{
									Player:   game.ControllerReference(),
									Amount:   game.Fixed(1),
									Duration: game.DurationThisTurn,
								},
							},
						},
					}.Ability(),
				},
			},
		},
	}
}
