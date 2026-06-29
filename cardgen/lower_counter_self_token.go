package cardgen

import (
	"github.com/natefinch/council4/cardgen/oracle/compiler"
	"github.com/natefinch/council4/cardgen/oracle/parser"
	"github.com/natefinch/council4/mtg/game"
)

// lowerCounterThenSelfTokenSequence lowers the ordered counter-then-create pair
// "Counter target <filter> spell. [You] create <token>." into a CounterObject
// followed by a controller-recipient CreateToken. It mirrors the
// target-controller sibling lowerCounterThenTargetControllerTokenSequence, but
// the created token is owned by the spell's caster (Geist Snatch, Summoner's
// Bane, Launch Mishap, Hornswoggle), so no recipient reference is emitted.
func lowerCounterThenSelfTokenSequence(ctx contentCtx) (game.AbilityContent, bool) {
	if !isCounterThenCreateSequence(ctx.content) ||
		!hasExactSelfCounterTokenEnvelope(ctx) {
		return game.AbilityContent{}, false
	}
	counterEffect := &ctx.content.Effects[0]
	tokenEffect := &ctx.content.Effects[1]
	target := ctx.content.Targets[0]
	if !isExactMandatoryCounterEffect(counterEffect, target) {
		return game.AbilityContent{}, false
	}
	targetSpec, ok := stackSpellTargetSpec(target)
	if !ok {
		return game.AbilityContent{}, false
	}
	tokenInstruction, ok := selfTokenInstruction(ctx, tokenEffect)
	if !ok {
		return game.AbilityContent{}, false
	}
	return game.Mode{
		Targets: []game.TargetSpec{targetSpec},
		Sequence: []game.Instruction{
			{Primitive: game.CounterObject{Object: game.TargetStackObjectReference(0)}},
			tokenInstruction,
		},
	}.Ability(), true
}

// hasExactSelfCounterTokenEnvelope accepts the counter-then-self-token body's
// envelope: a single stack-spell target, the lone counter target reference, and
// no other riders. The token recipient is the controller, so unlike the
// target-controller sibling the create effect contributes no extra reference.
func hasExactSelfCounterTokenEnvelope(ctx contentCtx) bool {
	return !ctx.optional &&
		len(ctx.content.Targets) == 1 &&
		len(ctx.content.Conditions) == 0 &&
		len(ctx.content.Keywords) == 0 &&
		len(ctx.content.Modes) == 0 &&
		len(ctx.content.References) == 0
}

// selfTokenInstruction builds the controller-recipient CreateToken instruction.
// It accepts the same unmodified creature and predefined-artifact token shapes
// as standalone token creation with a fixed count, and rejects tapped,
// attacking, copy, and choice forms.
func selfTokenInstruction(ctx contentCtx, tokenEffect *compiler.CompiledEffect) (game.Instruction, bool) {
	if tokenEffect.Context != parser.EffectContextController ||
		!isExactMandatoryEffect(tokenEffect) ||
		len(tokenEffect.Targets) != 0 ||
		len(tokenEffect.References) != 0 ||
		tokenEffect.TokenCopyOfTarget ||
		tokenEffect.TokenChoice ||
		tokenEffect.Selector.Tapped ||
		tokenEffect.Selector.Attacking {
		return game.Instruction{}, false
	}
	def, ok := synthesizeCreatureTokenDef(tokenEffect, nil)
	if !ok {
		def, ok = synthesizeNamedArtifactTokenDef(tokenEffect)
	}
	if !ok {
		return game.Instruction{}, false
	}
	if !tokenEffect.Amount.Known || tokenEffect.Amount.Value < 1 {
		return game.Instruction{}, false
	}
	return game.Instruction{Primitive: game.CreateToken{
		Amount: game.Fixed(tokenEffect.Amount.Value),
		Source: game.TokenDef(def),
	}}, true
}
