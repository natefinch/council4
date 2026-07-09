package cardgen

import (
	"github.com/natefinch/council4/cardgen/oracle/compiler"
	"github.com/natefinch/council4/cardgen/oracle/parser"
	"github.com/natefinch/council4/cardgen/oracle/shared"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// lowerKeywordShareTrigger lowers the recognized team keyword-sharing construct
// (Odric, Lunarch Marshal) into a phase/step triggered ability whose resolving
// sequence grants each shared keyword to all the controller's creatures until
// end of turn, gated on the controller controlling a creature that already has
// that keyword. Each shared keyword lowers to one Instruction whose Condition
// tests "a creature you control has <KW>" (ControlsMatching a creature with the
// keyword, controller-scoped) wrapping an ApplyContinuous that adds the keyword
// to the "creatures you control" battlefield group. Because each keyword's gate
// tests only that same keyword, the grants never feed one another, so evaluating
// them in sequence is equivalent to the printed simultaneous check.
//
// It fails closed (returning a diagnostic, so the card does not generate) for
// any keyword the lowering cannot both map to a runtime keyword and grant as a
// simple static keyword, never silently dropping a keyword or its gate.
func lowerKeywordShareTrigger(
	ability compiler.CompiledAbility,
	pattern game.TriggerPattern,
) (game.TriggeredAbility, *shared.Diagnostic) {
	const summary = "unsupported phase/step trigger phrase"
	if ability.KeywordShare == nil || len(ability.KeywordShare.Keywords) == 0 {
		return game.TriggeredAbility{}, executableDiagnostic(
			ability,
			summary,
			"the executable source backend requires at least one shared keyword",
		)
	}
	sequence := make([]game.Instruction, 0, len(ability.KeywordShare.Keywords))
	for _, kind := range ability.KeywordShare.Keywords {
		keyword, ok := keywordShareRuntimeKeyword(kind)
		if !ok {
			return game.TriggeredAbility{}, executableDiagnostic(
				ability,
				summary,
				"the executable source backend does not support granting a shared keyword",
			)
		}
		sequence = append(sequence, game.Instruction{
			Condition: opt.Val(game.EffectCondition{
				Condition: opt.Val(game.Condition{
					ControlsMatching: opt.Val(game.SelectionCount{
						Selection: game.Selection{
							RequiredTypes: []types.Card{types.Creature},
							Keyword:       keyword,
						},
					}),
				}),
			}),
			Primitive: game.ApplyContinuous{
				ContinuousEffects: []game.ContinuousEffect{
					{
						Layer: game.LayerAbility,
						Group: game.BattlefieldGroup(game.Selection{
							RequiredTypes: []types.Card{types.Creature},
							Controller:    game.ControllerYou,
						}),
						AddKeywords: []game.Keyword{keyword},
					},
				},
				Duration: game.DurationUntilEndOfTurn,
			},
		})
	}
	return game.TriggeredAbility{
		Text: ability.Text,
		Trigger: game.TriggerCondition{
			Type:    game.TriggerAt,
			Pattern: pattern,
		},
		Content: game.Mode{Sequence: sequence}.Ability(),
	}, nil
}

// keywordShareRuntimeKeyword maps a shared keyword kind to its runtime keyword,
// succeeding only for a keyword the engine both recognizes and supports as a
// grantable simple static keyword. This is the lowering's fail-closed authority
// for the keyword-share construct, mirroring the parser's recognition whitelist.
func keywordShareRuntimeKeyword(kind parser.KeywordKind) (game.Keyword, bool) {
	keyword, ok := runtimeKeyword(kind)
	if !ok || !mixedStaticKeywordImplemented(keyword) {
		return game.KeywordNone, false
	}
	return keyword, true
}
