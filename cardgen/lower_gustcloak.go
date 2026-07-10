package cardgen

import (
	"github.com/natefinch/council4/cardgen/oracle/compiler"
	"github.com/natefinch/council4/cardgen/oracle/parser"
	"github.com/natefinch/council4/mtg/game"
)

// optionalUntapRemoveFromCombatBody reports whether content is the Gustcloak
// whole-body optional sequence "you may untap it and remove it from combat."
// The trigger owns the single choice; both effects share one sentence and act on
// the source permanent.
func optionalUntapRemoveFromCombatBody(content compiler.AbilityContent) bool {
	if len(content.Effects) != 2 ||
		len(content.Targets) != 0 ||
		len(content.Conditions) != 0 ||
		len(content.Keywords) != 0 ||
		len(content.Modes) != 0 ||
		!gustcloakReferences(content.References) {
		return false
	}
	untap := content.Effects[0]
	remove := content.Effects[1]
	return untap.Kind == compiler.EffectUntap &&
		untap.Optional &&
		untap.Context == parser.EffectContextController &&
		remove.Kind == compiler.EffectRemoveFromCombat &&
		!remove.Optional &&
		remove.Exact &&
		remove.Context == parser.EffectContextController &&
		untap.Span == remove.Span
}

func gustcloakReferences(references []compiler.CompiledReference) bool {
	if len(references) < 2 || len(references) > 3 {
		return false
	}
	source, event := 0, 0
	for i := range references {
		switch references[i].Binding {
		case compiler.ReferenceBindingSource:
			source++
		case compiler.ReferenceBindingEventPermanent:
			event++
		default:
			return false
		}
	}
	return event == 2 && (source == 0 || source == 1)
}

// lowerOptionalUntapRemoveFromCombat lowers the two mandatory instructions that
// execute after the enclosing optional trigger is accepted.
func lowerOptionalUntapRemoveFromCombat(ctx contentCtx) (game.AbilityContent, bool) {
	if ctx.optional || !optionalUntapRemoveFromCombatBody(ctx.content) {
		return game.AbilityContent{}, false
	}
	source := game.EventPermanentReference()
	return game.Mode{Sequence: []game.Instruction{
		{Primitive: game.Untap{Object: source}},
		{Primitive: game.RemoveFromCombat{Object: source}},
	}}.Ability(), true
}
