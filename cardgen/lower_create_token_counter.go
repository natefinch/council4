package cardgen

import (
	"github.com/natefinch/council4/cardgen/oracle/compiler"
	"github.com/natefinch/council4/cardgen/oracle/parser"
	"github.com/natefinch/council4/mtg/game"
)

// lowerCreateTokenThenCountersSequence lowers the ordered pair "Create a <token>.
// Put <n> +1/+1 counters on it." (Fractal Summoning, Match the Odds, Leyline
// Invocation) into a token creation that publishes its result under a link key,
// followed by a counter placement whose recipient resolves to that just-created
// token. The counter clause's singular back-reference ("it" / "that token")
// names the one token the prior clause created; the lowering realizes it by
// publishing the token under createdTokenLinkKey and pointing the AddCounter at
// that linked object, mirroring lowerManifestDreadThenCountersSequence. It is
// restricted to a single-token creation (so the singular back-reference is
// unambiguous) and a controller-context placement of a fixed, variable-X, or
// recognized dynamic count of a placement-supported permanent counter kind,
// failing closed for plural tokens, player counters, durations, targets,
// conditions, modes, or any other shape.
func lowerCreateTokenThenCountersSequence(ctx contentCtx) (game.AbilityContent, bool) {
	if len(ctx.content.Effects) != 2 ||
		len(ctx.content.Targets) != 0 ||
		len(ctx.content.Conditions) != 0 ||
		len(ctx.content.Modes) != 0 ||
		ctx.optional {
		return game.AbilityContent{}, false
	}
	createEffect := ctx.content.Effects[0]
	counterEffect := ctx.content.Effects[1]
	if createEffect.Kind != compiler.EffectCreate ||
		counterEffect.Kind != compiler.EffectPut {
		return game.AbilityContent{}, false
	}
	// The creation must make exactly one token so the singular back-reference
	// "it"/"that token" denotes that one token without ambiguity.
	if !createEffect.Amount.Known || createEffect.Amount.Value != 1 {
		return game.AbilityContent{}, false
	}
	if counterEffect.Negated ||
		counterEffect.Optional ||
		!counterEffect.Exact ||
		counterEffect.Context != parser.EffectContextController ||
		counterEffect.Duration != compiler.DurationNone ||
		!counterEffect.CounterKindKnown ||
		!compiler.CounterKindPlacementSupported(counterEffect.CounterKind) ||
		counterEffect.CounterKind.PlayerOnly() {
		return game.AbilityContent{}, false
	}
	if !createdTokenSingularBackReference(counterEffect.References) {
		return game.AbilityContent{}, false
	}
	amount, ok := createdTokenCounterAmount(counterEffect.Amount)
	if !ok {
		return game.AbilityContent{}, false
	}
	createContent, diagnostic := lowerCreateTokenSpellLinked(
		contextForEffect(ctx, &createEffect), createdTokenLinkKey)
	if diagnostic != nil ||
		len(createContent.Modes) != 1 ||
		len(createContent.Modes[0].Sequence) != 1 ||
		len(createContent.Modes[0].Targets) != 0 {
		return game.AbilityContent{}, false
	}
	return game.Mode{
		Sequence: []game.Instruction{
			createContent.Modes[0].Sequence[0],
			{Primitive: game.AddCounter{
				Amount:      amount,
				Object:      game.LinkedObjectReference(string(createdTokenLinkKey)),
				CounterKind: counterEffect.CounterKind,
			}},
		},
	}.Ability(), true
}

// createdTokenSingularBackReference reports whether the counter clause's only
// reference is the singular object back-reference that names the just-created
// token: the pronoun "it" or the demonstrative "that token". The compiler does
// not bind these to a prior-instruction result here (no target precedes them and
// the creating clause is the nearest antecedent), so the lowering reconstructs
// the link structurally. Possessive ("its"), plural ("them"/"those"), and
// subject ("they") forms are rejected because their referent or count is not a
// single created token.
func createdTokenSingularBackReference(references []compiler.CompiledReference) bool {
	if len(references) != 1 {
		return false
	}
	reference := references[0]
	if reference.Kind == compiler.ReferenceThatObject {
		return true
	}
	return reference.Kind == compiler.ReferencePronoun &&
		reference.Pronoun == compiler.ReferencePronounIt
}

// createdTokenCounterAmount resolves the counter count placed on the created
// token. It accepts a fixed positive amount, the spell's variable X, or a
// recognized controller- or opponent-relative dynamic count, mirroring the
// single-target counter-placement amount handling. It fails closed for
// non-positive fixed amounts and for dynamic counts that read a permanent's own
// characteristics, whose object referent the created-token link does not model.
func createdTokenCounterAmount(amount compiler.CompiledAmount) (game.Quantity, bool) {
	switch {
	case amount.Known:
		if amount.Value < 1 {
			return game.Quantity{}, false
		}
		return game.Fixed(amount.Value), true
	case amount.VariableX:
		if amount.Multiplier > 1 || amount.Addend != 0 {
			return game.Quantity{}, false
		}
		return game.Dynamic(game.DynamicAmount{Kind: game.DynamicAmountX}), true
	case amount.DynamicKind != compiler.DynamicAmountNone:
		if createdTokenCounterAmountReadsObject(amount.DynamicKind) {
			return game.Quantity{}, false
		}
		dynamic, ok := lowerDynamicAmount(amount, game.SourcePermanentReference())
		if !ok {
			return game.Quantity{}, false
		}
		return game.Dynamic(dynamic), true
	default:
		return game.Quantity{}, false
	}
}

// createdTokenCounterAmountReadsObject reports whether a dynamic counter count
// reads a permanent's own power, toughness, or counter total. Those forms need
// an object referent the created-token link does not supply, so the placement
// fails closed rather than silently reading the spell source's characteristics.
func createdTokenCounterAmountReadsObject(kind compiler.DynamicAmountKind) bool {
	switch kind {
	case compiler.DynamicAmountSourcePower,
		compiler.DynamicAmountSourceToughness,
		compiler.DynamicAmountSourceManaValue,
		compiler.DynamicAmountSourceCounterCount:
		return true
	default:
		return false
	}
}
