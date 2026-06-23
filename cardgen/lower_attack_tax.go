package cardgen

import (
	"github.com/natefinch/council4/cardgen/oracle/compiler"
	"github.com/natefinch/council4/cardgen/oracle/parser"
	"github.com/natefinch/council4/cardgen/oracle/shared"
	"github.com/natefinch/council4/mtg/game"
)

// lowerAttackTaxSpell lowers the resolving, duration-bounded attack-tax effect
// "Until your next turn, creatures can't attack you unless their controller pays
// {N} for each of those creatures." (Summon: Yojimbo chapters II/III) into an
// ApplyRule that installs a controller-scoped RuleEffectAttackTax for the
// recognized duration. Unlike the continuous Propaganda-style static attack tax,
// this is a one-shot resolving installation; the runtime collects the applied
// rule effect alongside static ones when computing the per-attacker tax, and the
// "until your next turn" rule effect expires at the start of the controller's
// next turn. Targets, references, conditions, keywords, modes, a negation, a
// non-positive amount, or an unsupported duration fail closed.
func lowerAttackTaxSpell(ctx contentCtx) (game.AbilityContent, *shared.Diagnostic) {
	effect := ctx.content.Effects[0]
	if !effect.Exact ||
		effect.Negated ||
		effect.AttackTaxGeneric <= 0 ||
		effect.Context != parser.EffectContextController ||
		effect.Duration != compiler.DurationUntilYourNextTurn ||
		len(ctx.content.Targets) != 0 ||
		len(ctx.content.Keywords) != 0 ||
		len(ctx.content.Modes) != 0 {
		return game.AbilityContent{}, contentDiagnostic(
			ctx,
			"unsupported attack tax effect",
			"the executable source backend supports only the exact \"Until your next turn, creatures can't attack you unless their controller pays {N} for each of those creatures.\" resolving attack tax",
		)
	}
	return game.Mode{Sequence: []game.Instruction{{
		Primitive: game.ApplyRule{
			RuleEffects: []game.RuleEffect{{
				Kind:             game.RuleEffectAttackTax,
				AffectedPlayer:   game.PlayerYou,
				AttackTaxGeneric: effect.AttackTaxGeneric,
			}},
			Duration: game.DurationUntilYourNextTurn,
		},
	}}}.Ability(), nil
}
