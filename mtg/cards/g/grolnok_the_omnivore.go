package g

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// GrolnokTheOmnivore is the card definition for Grolnok, the Omnivore.
//
// Type: Legendary Creature — Frog
// Cost: {2}{G}{U}
//
// Oracle text:
//
//	Whenever a Frog you control attacks, mill three cards.
//	Whenever a permanent card is put into your graveyard from your library, exile it with a croak counter on it.
//	You may play lands and cast spells from among cards you own in exile with croak counters on them.
var GrolnokTheOmnivore = newGrolnokTheOmnivore

func newGrolnokTheOmnivore() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue, color.Green),
		CardFace: game.CardFace{
			Name: "Grolnok, the Omnivore",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.G,
				cost.U,
			}),
			Colors:     []color.Color{color.Green, color.Blue},
			Supertypes: []types.Super{types.Legendary},
			Types:      []types.Card{types.Creature},
			Subtypes:   []types.Sub{types.Frog},
			Power:      opt.Val(game.PT{Value: 3}),
			Toughness:  opt.Val(game.PT{Value: 3}),
			StaticAbilities: []game.StaticAbility{
				game.StaticAbility{
					RuleEffects: []game.RuleEffect{
						game.RuleEffect{
							Kind:               game.RuleEffectPlayLandsFromZone,
							AffectedPlayer:     game.PlayerYou,
							CastFromZone:       zone.Exile,
							PermanentTypes:     []types.Card{types.Land},
							ExileCounterFilter: opt.Val(counter.Croak),
						},
						game.RuleEffect{
							Kind:               game.RuleEffectCastSpellsFromZone,
							AffectedPlayer:     game.PlayerYou,
							CastFromZone:       zone.Exile,
							ExileCounterFilter: opt.Val(counter.Croak),
						},
					},
				},
			},
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Trigger: game.TriggerCondition{
						Type: game.TriggerWhenever,
						Pattern: game.TriggerPattern{
							Event:            game.EventAttackerDeclared,
							Controller:       game.TriggerControllerYou,
							SubjectSelection: game.Selection{SubtypesAny: []types.Sub{types.Sub("Frog")}},
						},
					},
					Content: game.Mode{
						Sequence: []game.Instruction{
							{
								Primitive: game.Mill{
									Amount: game.Fixed(3),
									Player: game.ControllerReference(),
								},
							},
						},
					}.Ability(),
				},
				game.TriggeredAbility{
					Trigger: game.TriggerCondition{
						Type: game.TriggerWhenever,
						Pattern: game.TriggerPattern{
							Event:            game.EventZoneChanged,
							Player:           game.TriggerPlayerYou,
							MatchFromZone:    true,
							FromZone:         zone.Library,
							MatchToZone:      true,
							ToZone:           zone.Graveyard,
							SubjectSelection: game.Selection{RequiredTypesAny: []types.Card{types.Artifact, types.Battle, types.Creature, types.Enchantment, types.Land, types.Planeswalker}},
						},
					},
					Content: game.Mode{
						Sequence: []game.Instruction{
							{
								Primitive: game.MoveCard{
									Card:        game.CardReference{Kind: game.CardReferenceEvent},
									FromZone:    zone.Graveyard,
									Destination: zone.Exile,
									Counter:     opt.Val(counter.Croak),
								},
							},
						},
					}.Ability(),
				},
			},
			OracleText: `
			Whenever a Frog you control attacks, mill three cards.
			Whenever a permanent card is put into your graveyard from your library, exile it with a croak counter on it.
			You may play lands and cast spells from among cards you own in exile with croak counters on them.
		`,
		},
	}
}
