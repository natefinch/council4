package cardgen

import (
	"github.com/natefinch/council4/cardgen/oracle/compiler"
	"github.com/natefinch/council4/cardgen/oracle/parser"
	"github.com/natefinch/council4/cardgen/oracle/shared"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/types"
)

// lowerResolvingCostModifier lowers the one-shot, duration-bounded resolved cost
// modifier "<type> spells <caster> cast cost {N} more/less to cast until your
// next turn." (Elspeth Conquers Death chapter II: "Noncreature spells your
// opponents cast cost {2} more to cast until your next turn.") to an ApplyRule
// that creates a continuous RuleEffectCostModifier scoped to the affected
// casters' spells and bounded by the stated duration. The affected casters are
// the controller's opponents (PlayerOpponent), the controller (PlayerYou), or
// every player (PlayerAny); the optional single card-type filter narrows or
// exempts the affected spells via the rule effect's SpellTypes/ExcludedSpellTypes.
// Targets, references, conditions, modes, a negation, or an unsupported duration
// fail closed.
func lowerResolvingCostModifier(ctx contentCtx) (game.AbilityContent, *shared.Diagnostic) {
	effect := ctx.content.Effects[0]
	duration, ok := lowerResolvingCostModifierDuration(effect.Duration)
	if !effect.Exact ||
		effect.Negated ||
		!ok ||
		effect.Context != parser.EffectContextController ||
		effect.ResolvingCostModifierAmount <= 0 ||
		len(ctx.content.Targets) != 0 ||
		len(ctx.content.References) != 0 ||
		len(ctx.content.Conditions) != 0 ||
		len(ctx.content.Keywords) != 0 ||
		len(ctx.content.Modes) != 0 {
		return game.AbilityContent{}, contentDiagnostic(
			ctx,
			"unsupported resolving cost modifier",
			"the executable source backend supports only the exact duration-bounded spell cost-increase or cost-reduction effect",
		)
	}
	affected := game.PlayerAny
	switch effect.ResolvingCostModifierCaster {
	case parser.ResolvingCostModifierCasterOpponents:
		affected = game.PlayerOpponent
	case parser.ResolvingCostModifierCasterController:
		affected = game.PlayerYou
	case parser.ResolvingCostModifierCasterAllPlayers:
		affected = game.PlayerAny
	default:
		return game.AbilityContent{}, contentDiagnostic(
			ctx,
			"unsupported resolving cost modifier",
			"the executable source backend supports only opponents, controller, or all-players cast-cost modifiers",
		)
	}
	modifier := game.CostModifier{Kind: game.CostModifierSpell}
	if effect.ResolvingCostModifierIncrease {
		modifier.GenericIncrease = effect.ResolvingCostModifierAmount
	} else {
		modifier.GenericReduction = effect.ResolvingCostModifierAmount
	}
	return game.Mode{Sequence: []game.Instruction{{
		Primitive: game.ApplyRule{
			RuleEffects: []game.RuleEffect{{
				Kind:               game.RuleEffectCostModifier,
				AffectedPlayer:     affected,
				SpellTypes:         append([]types.Card(nil), effect.ResolvingCostModifierRequiredTypes...),
				ExcludedSpellTypes: append([]types.Card(nil), effect.ResolvingCostModifierExcludedTypes...),
				CostModifier:       modifier,
			}},
			Duration: duration,
		},
	}}}.Ability(), nil
}

// lowerResolvingCostModifierDuration maps the two durations a resolving cost
// modifier may carry to their runtime equivalents. Any other duration fails
// closed.
func lowerResolvingCostModifierDuration(duration compiler.DurationKind) (game.EffectDuration, bool) {
	switch duration {
	case compiler.DurationUntilYourNextTurn:
		return game.DurationUntilYourNextTurn, true
	case compiler.DurationThisTurn:
		return game.DurationThisTurn, true
	default:
		return game.DurationPermanent, false
	}
}
