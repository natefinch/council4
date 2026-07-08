package t

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// TheDawningArchaic is the card definition for The Dawning Archaic.
//
// Type: Legendary Creature — Avatar
// Cost: {10}
//
// Oracle text:
//
//	This spell costs {1} less to cast for each instant and sorcery card in your graveyard.
//	Reach
//	Whenever The Dawning Archaic attacks, you may cast target instant or sorcery card from your graveyard without paying its mana cost. If that spell would be put into your graveyard, exile it instead.
var TheDawningArchaic = newTheDawningArchaic

func newTheDawningArchaic() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name: "The Dawning Archaic",
			ManaCost: opt.Val(cost.Mana{
				cost.O(10),
			}),
			Supertypes: []types.Super{types.Legendary},
			Types:      []types.Card{types.Creature},
			Subtypes:   []types.Sub{types.Avatar},
			Power:      opt.Val(game.PT{Value: 7}),
			Toughness:  opt.Val(game.PT{Value: 7}),
			StaticAbilities: []game.StaticAbility{
				game.StaticAbility{
					RuleEffects: []game.RuleEffect{
						game.RuleEffect{
							Kind:           game.RuleEffectCostModifier,
							AffectedSource: true,
							CostModifier: game.CostModifier{
								Kind:               game.CostModifierSpell,
								PerObjectReduction: 1,
								CountSelection:     &game.Selection{RequiredTypesAny: []types.Card{types.Instant, types.Sorcery}},
								CountZone:          opt.Val(zone.Graveyard),
							},
						},
					},
				},
				game.ReachStaticBody,
			},
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Trigger: game.TriggerCondition{
						Type: game.TriggerWhenever,
						Pattern: game.TriggerPattern{
							Event:  game.EventAttackerDeclared,
							Source: game.TriggerSourceSelf,
						},
					},
					Content: game.Mode{
						Targets: []game.TargetSpec{
							game.TargetSpec{
								MinTargets: 1,
								MaxTargets: 1,
								Constraint: "target instant or sorcery card from your graveyard without paying its mana cost",
								Allow:      game.TargetAllowCard,
								TargetZone: zone.Graveyard,
								Selection:  opt.Val(game.Selection{RequiredTypesAny: []types.Card{types.Instant, types.Sorcery}, Controller: game.ControllerYou}),
							},
						},
						Sequence: []game.Instruction{
							{
								Primitive: game.CastForFree{
									Player:            game.ControllerReference(),
									Zone:              zone.Graveyard,
									Card:              game.CardReference{Kind: game.CardReferenceTarget},
									ExileOnResolution: true,
								},
								Optional: true,
							},
						},
					}.Ability(),
				},
			},
			OracleText: `
			This spell costs {1} less to cast for each instant and sorcery card in your graveyard.
			Reach
			Whenever The Dawning Archaic attacks, you may cast target instant or sorcery card from your graveyard without paying its mana cost. If that spell would be put into your graveyard, exile it instead.
		`,
		},
	}
}
