package cardgen

import "github.com/natefinch/council4/cardgen/oracle/compiler"

func halfPlayerAmountRecipientReferences(ctx contentCtx, effect compiler.CompiledEffect) []compiler.CompiledReference {
	var target *compiler.CompiledTarget
	if len(ctx.content.Targets) == 1 {
		target = &ctx.content.Targets[0]
	}
	var remaining []compiler.CompiledReference
	for _, reference := range ctx.content.References {
		if ctx.content.OwnsSubject(reference) && reference.HalfPlayerAmountRecipient(effect, target) {
			continue
		}
		remaining = append(remaining, reference)
	}
	return remaining
}
