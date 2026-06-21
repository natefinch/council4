package cardgen

import (
	"github.com/natefinch/council4/cardgen/oracle/compiler"
	"github.com/natefinch/council4/cardgen/oracle/parser"
	"github.com/natefinch/council4/cardgen/oracle/shared"
	"github.com/natefinch/council4/mtg/game"
)

// lowerAttachSpell lowers "attach it/that <Equipment> to target <permanent> you
// control" — the enters-the-battlefield auto-attach trigger of Equipment such as
// Mithril Coat ("attach it ...") and Hammer of Nazahn ("Whenever ~ or another
// Equipment you control enters, you may attach that Equipment ..."). The
// entering Equipment is attached to the single chosen target without paying an
// Equip cost. The attachment object is whatever the back-reference denotes: the
// source permanent for a self-only trigger, or the triggering event permanent
// when "another Equipment you control" can enter. It fails closed for any other
// attach shape.
func lowerAttachSpell(ctx contentCtx) (game.AbilityContent, *shared.Diagnostic) {
	unsupported := func() (game.AbilityContent, *shared.Diagnostic) {
		return game.AbilityContent{}, contentDiagnostic(
			ctx,
			"unsupported attach effect",
			"the executable source backend supports only \"attach it/that <Equipment> to target <permanent> you control\" attaching the entering or source permanent",
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
		!attachReferencesEnteringObject(effect.References) {
		return unsupported()
	}
	attachment, ok := lowerObjectReference(effect.References[0],
		referenceLoweringContext{AllowSource: true, AllowEvent: true})
	if !ok {
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
				Attachment: attachment,
				Target:     game.TargetPermanentReference(0),
			},
		}},
	}.Ability(), nil
}

// attachReferencesEnteringObject reports whether the effect's only reference
// names the entering Equipment, either as the pronoun "it" or as the
// demonstrative "that <Equipment>". The compiler binds either wording to the
// source permanent (a self-only enters trigger) or to the triggering event
// permanent ("another Equipment you control enters"); both denote the Equipment
// being attached.
func attachReferencesEnteringObject(references []compiler.CompiledReference) bool {
	if len(references) != 1 {
		return false
	}
	reference := references[0]
	if reference.Binding != compiler.ReferenceBindingSource &&
		reference.Binding != compiler.ReferenceBindingEventPermanent {
		return false
	}
	itPronoun := reference.Kind == compiler.ReferencePronoun &&
		reference.Pronoun == compiler.ReferencePronounIt
	thatObject := reference.Kind == compiler.ReferenceThatObject
	return itPronoun || thatObject
}
