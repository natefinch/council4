package a

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// ArtisanOfKozilek is the card definition for Artisan of Kozilek.
//
// Type: Creature — Eldrazi
// Cost: {9}
//
// Oracle text:
//
//	When you cast this spell, you may return target creature card from your graveyard to the battlefield.
//	Annihilator 2 (Whenever this creature attacks, defending player sacrifices two permanents of their choice.)
var ArtisanOfKozilek = newArtisanOfKozilek()

func newArtisanOfKozilek() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name: "Artisan of Kozilek",
			ManaCost: opt.Val(cost.Mana{
				cost.O(9),
			}),
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Eldrazi},
			Power:     opt.Val(game.PT{Value: 10}),
			Toughness: opt.Val(game.PT{Value: 9}),
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Trigger: game.TriggerCondition{
						Type: game.TriggerWhen,
						Pattern: game.TriggerPattern{
							Event:       game.EventSpellCast,
							Source:      game.TriggerSourceSelf,
							Controller:  game.TriggerControllerYou,
							SelfWasCast: true,
						},
					},
					Optional: true,
					Content: game.Mode{
						Targets: []game.TargetSpec{
							game.TargetSpec{
								MinTargets: 1,
								MaxTargets: 1,
								Constraint: "target creature card from your graveyard",
								Allow:      game.TargetAllowCard,
								TargetZone: zone.Graveyard,
								Selection:  opt.Val(game.Selection{RequiredTypes: []types.Card{types.Creature}, Controller: game.ControllerYou}),
							},
						},
						Sequence: []game.Instruction{
							{
								Primitive: game.PutOnBattlefield{
									Source: game.CardBattlefieldSource(game.CardReference{Kind: game.CardReferenceTarget}),
								},
							},
						},
					}.Ability(),
				},
				game.TriggeredAbility{
					Trigger: game.TriggerCondition{
						Type: game.TriggerWhenever,
						Pattern: game.TriggerPattern{
							Event:  game.EventAttackerDeclared,
							Source: game.TriggerSourceSelf,
						},
					},
					Content: game.Mode{
						Sequence: []game.Instruction{
							{
								Primitive: game.SacrificePermanents{
									Amount: game.Fixed(2),
									Player: game.DefendingPlayerReference(),
								},
							},
						},
					}.Ability(),
				},
			},
			OracleText: `
			When you cast this spell, you may return target creature card from your graveyard to the battlefield.
			Annihilator 2 (Whenever this creature attacks, defending player sacrifices two permanents of their choice.)
		`,
		},
	}
}
