package cardgen

import (
	"github.com/natefinch/council4/cardgen/oracle/compiler"
	"github.com/natefinch/council4/cardgen/oracle/parser"
	"github.com/natefinch/council4/cardgen/oracle/shared"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/opt"
)

// lowerDelayedTriggerContent lowers a single EffectDelayedTrigger effect
// ("Whenever you cast a spell this turn, ...", "When you next cast a creature
// spell this turn, ...") into a game.CreateDelayedTrigger instruction. The
// nested triggered ability the parser reparsed is compiled and lowered
// recursively; its trigger pattern and content become the event-based delayed
// trigger's pattern and body, scoped to the rest of the turn. It fails closed
// when the effect carries outer targets, references, conditions, keywords, or
// modes, or when the nested ability does not lower to a plain event trigger.
func lowerDelayedTriggerContent(ctx contentCtx) (game.AbilityContent, *shared.Diagnostic) {
	effect := ctx.content.Effects[0]
	unsupported := func(detail string) (game.AbilityContent, *shared.Diagnostic) {
		return game.AbilityContent{}, contentDiagnostic(ctx, "unsupported delayed trigger", detail)
	}
	if effect.DelayedTriggerAbility == nil ||
		effect.Negated ||
		effect.Optional ||
		len(ctx.content.Conditions) != 0 ||
		len(ctx.content.Keywords) != 0 ||
		len(ctx.content.Modes) != 0 ||
		len(ctx.content.Targets) != 0 ||
		len(ctx.content.References) != 0 {
		return unsupported("the executable source backend supports only an unconditional, untargeted event-based delayed trigger")
	}
	triggered, ok := lowerDelayedTriggerInner(effect.DelayedTriggerAbility)
	if !ok {
		return unsupported("the nested triggered ability did not lower to a plain event trigger")
	}
	return game.Mode{
		Sequence: []game.Instruction{{
			Primitive: game.CreateDelayedTrigger{Trigger: game.DelayedTriggerDef{
				EventPattern: opt.Val(triggered.Trigger.Pattern),
				OneShot:      effect.DelayedTriggerOneShot,
				Window:       game.DelayedWindowThisTurn,
				Content:      triggered.Content,
			}},
		}},
	}.Ability(), nil
}

// lowerDelayedTriggerInner compiles and lowers the nested triggered ability of a
// delayed trigger, returning the plain triggered ability whose pattern and
// content the delayed trigger reuses. It mirrors attachTokenGrantedAbility's
// recursive compile + lower of an already-parsed inner document, and fails
// closed when the inner document does not compile to exactly one plain triggered
// ability the delayed-trigger runtime can carry (no intervening-if, optional,
// keyword, or per-turn-limit machinery, which DelayedTriggerDef cannot model).
func lowerDelayedTriggerInner(granted *parser.StaticGrantedAbilitySyntax) (game.TriggeredAbility, bool) {
	innerDocument, innerDiags := granted.Inner()
	if len(innerDiags) != 0 {
		return game.TriggeredAbility{}, false
	}
	innerComp, compilerDiags := compiler.Compile(innerDocument, compiler.Context{})
	if len(compilerDiags) != 0 ||
		len(innerComp.Abilities) != 1 ||
		len(innerComp.Syntax.Abilities) != 1 {
		return game.TriggeredAbility{}, false
	}
	lowered, diagnostic := lowerExecutableAbility("", false, nil, innerComp.Abilities[0], &innerComp.Syntax.Abilities[0])
	if diagnostic != nil || !lowered.triggeredAbility.Exists {
		return game.TriggeredAbility{}, false
	}
	triggered := lowered.triggeredAbility.Val
	if triggered.Optional ||
		triggered.MaxTriggersPerTurn != 0 ||
		len(triggered.KeywordAbilities) != 0 ||
		triggered.Trigger.InterveningIf != "" ||
		triggered.Trigger.InterveningCondition.Exists {
		return game.TriggeredAbility{}, false
	}
	return triggered, true
}
