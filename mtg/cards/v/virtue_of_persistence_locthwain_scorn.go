package v

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// VirtueOfPersistence is the card definition for Virtue of Persistence // Locthwain Scorn.
//
// Type: Enchantment // Sorcery — Adventure
// Cost: {5}{B}{B} // {1}{B}
// Face: Locthwain Scorn — Sorcery — Adventure ({1}{B})
//
// Oracle text:
//
//	At the beginning of your upkeep, put target creature card from a graveyard onto the battlefield under your control.
var VirtueOfPersistence = newVirtueOfPersistence

func newVirtueOfPersistence() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Black),
		CardFace: game.CardFace{
			Name: "Virtue of Persistence",
			ManaCost: opt.Val(cost.Mana{
				cost.O(5),
				cost.B,
				cost.B,
			}),
			Colors: []color.Color{color.Black},
			Types:  []types.Card{types.Enchantment},
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Text: "At the beginning of your upkeep, put target creature card from a graveyard onto the battlefield under your control.",
					Trigger: game.TriggerCondition{
						Type: game.TriggerAt,
						Pattern: game.TriggerPattern{
							Event:           game.EventBeginningOfStep,
							Controller:      game.TriggerControllerYou,
							StackObjectKind: game.StackSpell,
							CounterKind:     counter.PlusOnePlusOne,
							Step:            game.StepUpkeep,
						},
					},
					Content: game.AbilityContent{
						Modes: []game.Mode{
							game.Mode{
								Targets: []game.TargetSpec{
									game.TargetSpec{
										MinTargets: 1,
										MaxTargets: 1,
										Constraint: "target creature card from a graveyard",
										Allow:      game.TargetAllowCard,
										Selection:  opt.Val(game.Selection{RequiredTypes: []types.Card{types.Creature}}),
										TargetZone: zone.Graveyard,
									},
								},
								Sequence: []game.Instruction{
									game.Instruction{
										Primitive: game.PutOnBattlefield{
											Source:    game.CardBattlefieldSource(game.CardReference{Kind: game.CardReferenceTarget}),
											Recipient: opt.Val(game.ControllerReference()),
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
			At the beginning of your upkeep, put target creature card from a graveyard onto the battlefield under your control.
		`,
		},
		Layout: game.LayoutAdventure,
		Alternate: opt.Val(game.CardFace{
			Name: "Locthwain Scorn",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
				cost.B,
			}),
			Colors:   []color.Color{color.Black},
			Types:    []types.Card{types.Sorcery},
			Subtypes: []types.Sub{types.Adventure},
			SpellAbility: opt.Val(game.Mode{
				Targets: []game.TargetSpec{
					game.TargetSpec{
						MinTargets: 1,
						MaxTargets: 1,
						Constraint: "target creature",
						Allow:      game.TargetAllowPermanent,
						Selection:  opt.Val(game.Selection{RequiredTypesAny: []types.Card{types.Creature}}),
					},
				},
				Sequence: []game.Instruction{
					{
						Primitive: game.ModifyPT{
							Object:         game.TargetPermanentReference(0),
							PowerDelta:     game.Fixed(-3),
							ToughnessDelta: game.Fixed(-3),
							Duration:       game.DurationUntilEndOfTurn,
						},
					},
					{
						Primitive: game.GainLife{
							Amount: game.Fixed(2),
							Player: game.ControllerReference(),
						},
					},
				},
			}.Ability()),
			OracleText: `
			Target creature gets -3/-3 until end of turn. You gain 2 life.
		`,
		}),
	}
}
