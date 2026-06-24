package cardgen

import (
	"github.com/natefinch/council4/cardgen/oracle/compiler"
	"github.com/natefinch/council4/cardgen/oracle/parser"
	"github.com/natefinch/council4/mtg/game"
)

// manifestedCreatureLinkKey links a freshly manifested creature to a following
// "put <n> +1/+1 counters on that creature" clause so the counters land on
// exactly that creature. The runtime scopes the key per source object, so a
// fixed string is unambiguous across cards.
const manifestedCreatureLinkKey = "manifested-creature"

// lowerManifestDreadThenCountersSequence lowers the ordered pair "Manifest
// dread, then put <n> +1/+1 counters on that creature." (Weight Room). The first
// clause manifests the top card after a dread look; the second places a fixed
// number of +1/+1 counters on that just-manifested creature. The counter
// clause's "that creature" back-reference binds to the manifest instruction's
// result, realized by publishing the manifested creature under a link key and
// resolving the counter clause's object reference to that linked creature. It
// supports only the singular "that creature" back-reference onto a single
// known-kind, placement-supported counter and fails closed for any other shape,
// plural back-references, unknown amounts, or unsupported counter kinds.
func lowerManifestDreadThenCountersSequence(ctx contentCtx) (game.AbilityContent, bool) {
	if len(ctx.content.Effects) != 2 ||
		len(ctx.content.Targets) != 0 ||
		len(ctx.content.Conditions) != 0 ||
		len(ctx.content.Keywords) != 0 ||
		len(ctx.content.Modes) != 0 ||
		ctx.optional {
		return game.AbilityContent{}, false
	}
	manifestEffect := ctx.content.Effects[0]
	counterEffect := ctx.content.Effects[1]
	if manifestEffect.Kind != compiler.EffectManifestDread ||
		manifestEffect.Negated ||
		manifestEffect.Optional ||
		manifestEffect.Context != parser.EffectContextController {
		return game.AbilityContent{}, false
	}
	if counterEffect.Kind != compiler.EffectPut ||
		counterEffect.Negated ||
		counterEffect.Optional ||
		counterEffect.Context != parser.EffectContextController ||
		counterEffect.Duration != compiler.DurationNone ||
		!counterEffect.CounterKindKnown ||
		!compiler.CounterKindPlacementSupported(counterEffect.CounterKind) ||
		counterEffect.CounterKind.PlayerOnly() ||
		!counterEffect.Amount.Known ||
		counterEffect.Amount.Value < 1 {
		return game.AbilityContent{}, false
	}
	// "that creature" must bind to the manifest instruction's result (the
	// just-manifested creature). The reference lives at the ability level, not
	// in the counter effect's own subject references, so validate it there.
	if len(ctx.content.References) != 1 ||
		!referencesBindTo(ctx.content.References, compiler.ReferenceBindingPriorInstructionResult, 0) {
		return game.AbilityContent{}, false
	}
	return game.Mode{
		Sequence: []game.Instruction{
			{Primitive: game.Manifest{Dread: true, PublishLinked: manifestedCreatureLinkKey}},
			{Primitive: game.AddCounter{
				Amount:      game.Fixed(counterEffect.Amount.Value),
				Object:      game.LinkedObjectReference(string(manifestedCreatureLinkKey)),
				CounterKind: counterEffect.CounterKind,
			}},
		},
	}.Ability(), true
}
