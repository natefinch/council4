package j

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// JinGitaxiasProgressTyrant is the card definition for Jin-Gitaxias, Progress Tyrant.
//
// Type: Legendary Creature — Phyrexian Praetor
// Cost: {5}{U}{U}
//
// Oracle text:
//
//	Whenever you cast an artifact, instant, or sorcery spell, copy that spell. You may choose new targets for the copy. This ability triggers only once each turn. (A copy of a permanent spell becomes a token.)
//	Whenever an opponent casts an artifact, instant, or sorcery spell, counter that spell. This ability triggers only once each turn.
var JinGitaxiasProgressTyrant = newJinGitaxiasProgressTyrant

func newJinGitaxiasProgressTyrant() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue),
		CardFace: game.CardFace{
			Name: "Jin-Gitaxias, Progress Tyrant",
			ManaCost: opt.Val(cost.Mana{
				cost.O(5),
				cost.U,
				cost.U,
			}),
			Colors:     []color.Color{color.Blue},
			Supertypes: []types.Super{types.Legendary},
			Types:      []types.Card{types.Creature},
			Subtypes:   []types.Sub{types.Phyrexian, types.Praetor},
			Power:      opt.Val(game.PT{Value: 5}),
			Toughness:  opt.Val(game.PT{Value: 5}),
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Trigger: game.TriggerCondition{
						Type: game.TriggerWhenever,
						Pattern: game.TriggerPattern{
							Event:         game.EventSpellCast,
							Controller:    game.TriggerControllerYou,
							CardSelection: game.Selection{RequiredTypesAny: []types.Card{types.Artifact, types.Instant, types.Sorcery}},
						},
					},
					MaxTriggersPerTurn: 1,
					Content: game.Mode{
						Sequence: []game.Instruction{
							{
								Primitive: game.CopyStackObject{
									Object:              game.EventStackObjectReference(),
									MayChooseNewTargets: true,
								},
							},
						},
					}.Ability(),
				},
				game.TriggeredAbility{
					Trigger: game.TriggerCondition{
						Type: game.TriggerWhenever,
						Pattern: game.TriggerPattern{
							Event:         game.EventSpellCast,
							Controller:    game.TriggerControllerOpponent,
							CardSelection: game.Selection{RequiredTypesAny: []types.Card{types.Artifact, types.Instant, types.Sorcery}},
						},
					},
					MaxTriggersPerTurn: 1,
					Content: game.Mode{
						Sequence: []game.Instruction{
							{
								Primitive: game.CounterObject{
									Object: game.EventStackObjectReference(),
								},
							},
						},
					}.Ability(),
				},
			},
			OracleText: `
			Whenever you cast an artifact, instant, or sorcery spell, copy that spell. You may choose new targets for the copy. This ability triggers only once each turn. (A copy of a permanent spell becomes a token.)
			Whenever an opponent casts an artifact, instant, or sorcery spell, counter that spell. This ability triggers only once each turn.
		`,
		},
	}
}
