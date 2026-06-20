package cardgen

import (
	"github.com/natefinch/council4/cardgen/oracle/compiler"
	"github.com/natefinch/council4/cardgen/oracle/parser"
	"github.com/natefinch/council4/cardgen/oracle/shared"
	"github.com/natefinch/council4/mtg/game"
)

// lowerAttachSpell lowers "attach it to target <permanent> you control" — the
// enters-the-battlefield auto-attach trigger of Equipment such as Mithril Coat.
// The source permanent ("it") is attached to the single chosen target without
// paying an Equip cost. It fails closed for any other attach shape.
func lowerAttachSpell(ctx contentCtx) (game.AbilityContent, *shared.Diagnostic) {
	unsupported := func() (game.AbilityContent, *shared.Diagnostic) {
		return game.AbilityContent{}, contentDiagnostic(
			ctx,
			"unsupported attach effect",
			"the executable source backend supports only \"attach it to target <permanent> you control\" attaching the source permanent",
		)
	}
	effect := ctx.content.Effects[0]
	if ctx.optional ||
		len(ctx.content.Effects) != 1 ||
		len(ctx.content.Targets) != 1 ||
		len(ctx.content.Modes) != 0 ||
		len(ctx.content.Conditions) != 0 ||
		len(ctx.content.Keywords) != 0 ||
		effect.Negated ||
		effect.Optional ||
		effect.Context != parser.EffectContextController ||
		!attachReferencesSource(effect.References) {
		return unsupported()
	}
	targetSpec, ok := permanentTargetSpec(ctx.content.Targets[0])
	if !ok {
		return unsupported()
	}
	return game.Mode{
		Targets: []game.TargetSpec{targetSpec},
		Sequence: []game.Instruction{{
			Primitive: game.Attach{
				Attachment: game.SourcePermanentReference(),
				Target:     game.TargetPermanentReference(0),
			},
		}},
	}.Ability(), nil
}

// attachReferencesSource reports whether the effect's only reference is the
// pronoun "it" naming the source permanent (the entering Equipment). The
// compiler binds that pronoun to the source or, for a self enters trigger, the
// triggering event permanent — both resolve to the Equipment being attached.
func attachReferencesSource(references []compiler.CompiledReference) bool {
	if len(references) != 1 {
		return false
	}
	reference := references[0]
	return reference.Kind == compiler.ReferencePronoun &&
		reference.Pronoun == compiler.ReferencePronounIt &&
		(reference.Binding == compiler.ReferenceBindingSource ||
			reference.Binding == compiler.ReferenceBindingEventPermanent)
}
