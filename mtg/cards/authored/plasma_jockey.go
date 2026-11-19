package authored

import (
	"github.com/natefinch/council4/mtg/cards/cardset"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

const (
	plasmaJockeyOracleID   = "dedf0d3a-090a-4cc3-b5ee-65b6cb990b72"
	plasmaJockeyAttackText = "Whenever this creature attacks, target creature an opponent controls can't block this turn."
	plasmaJockeyOracleText = `Whenever this creature attacks, target creature an opponent controls can't block this turn.
Blitz {2}{R} (If you cast this spell for its blitz cost, it gains haste and "When this creature dies, draw a card." Sacrifice it at the beginning of the next end step.)`
)

var PlasmaJockey = cardset.Entry{
	OracleID:     plasmaJockeyOracleID,
	Name:         "Plasma Jockey",
	New:          newPlasmaJockey,
	Verification: "mtg/rules/plasma_jockey_test.go: TestPlasmaJockeyBlitzAttackPreventsOpponentCreatureBlocking",
}

func newPlasmaJockey() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Red),
		CardFace: game.CardFace{
			Name:             "Plasma Jockey",
			ManaCost:         opt.Val(cost.Mana{cost.O(3), cost.R}),
			AlternativeCosts: []cost.Alternative{cost.BlitzAlternative(cost.Mana{cost.O(2), cost.R})},
			Types:            []types.Card{types.Creature},
			Subtypes:         []types.Sub{types.Lizard, types.Warrior},
			Colors:           []color.Color{color.Red},
			Power:            opt.Val(game.PT{Value: 3}),
			Toughness:        opt.Val(game.PT{Value: 1}),
			OracleText:       plasmaJockeyOracleText,
			TriggeredAbilities: []game.TriggeredAbility{{
				Text: plasmaJockeyAttackText,
				Trigger: game.TriggerCondition{
					Type: game.TriggerWhenever,
					Pattern: game.TriggerPattern{
						Event:  game.EventAttackerDeclared,
						Source: game.TriggerSourceSelf,
					},
				},
				Content: game.Mode{
					Targets: []game.TargetSpec{{
						MinTargets: 1,
						MaxTargets: 1,
						Constraint: "target creature an opponent controls",
						Allow:      game.TargetAllowPermanent,
						Selection: opt.Val(game.Selection{
							RequiredTypes: []types.Card{types.Creature},
							Controller:    game.ControllerOpponent,
						}),
					}},
					Sequence: []game.Instruction{{
						Primitive: game.ApplyRule{
							Object:      opt.Val(game.TargetPermanentReference(0)),
							RuleEffects: []game.RuleEffect{{Kind: game.RuleEffectCantBlock}},
							Duration:    game.DurationUntilEndOfTurn,
						},
					}},
				}.Ability(),
			}},
			StaticAbilities: []game.StaticAbility{game.BlitzStaticAbility()},
		},
	}
}
