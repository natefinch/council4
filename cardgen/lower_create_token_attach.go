package cardgen

import (
	"github.com/natefinch/council4/cardgen/oracle/compiler"
	"github.com/natefinch/council4/cardgen/oracle/parser"
	"github.com/natefinch/council4/mtg/game"
)

// lowerCreateTokenThenAttachSequence lowers the ordered pair "Create a <token>,
// then attach this Equipment to it." (Barbed Spike, Ancestral Blade, Headsplitter
// and the Living weapon / For Mirrodin! reminder shape) into a token creation that
// publishes its result under a link key, followed by an attach that fastens the
// entering Equipment to that just-created token. The attach clause names two
// objects: the entering Equipment ("this" / "this Equipment") and the created
// token ("it"), neither of which is a chosen target. The lowering reconstructs the
// link structurally — publishing the token under createdTokenLinkKey and pointing
// the Attach at that linked object — mirroring lowerCreateTokenThenCountersSequence
// and the engine's Living weapon template. It is restricted to a single-token
// creation (so the "it" back-reference is unambiguous) and a controller-context
// attach of the source/entering Equipment, failing closed for plural tokens,
// targets, conditions, modes, durations, or any other shape.
func lowerCreateTokenThenAttachSequence(ctx contentCtx) (game.AbilityContent, bool) {
	if len(ctx.content.Effects) != 2 ||
		len(ctx.content.Targets) != 0 ||
		len(ctx.content.Conditions) != 0 ||
		len(ctx.content.Modes) != 0 ||
		ctx.optional {
		return game.AbilityContent{}, false
	}
	createEffect := ctx.content.Effects[0]
	attachEffect := ctx.content.Effects[1]
	if createEffect.Kind != compiler.EffectCreate ||
		attachEffect.Kind != compiler.EffectAttach ||
		createEffect.Negated ||
		attachEffect.Negated ||
		createEffect.Optional ||
		attachEffect.Optional ||
		attachEffect.Context != parser.EffectContextController ||
		attachEffect.Duration != compiler.DurationNone {
		return game.AbilityContent{}, false
	}
	// The creation must make exactly one token so the singular "it"
	// back-reference denotes that one token without ambiguity.
	if !createEffect.Amount.Known || createEffect.Amount.Value != 1 {
		return game.AbilityContent{}, false
	}
	attachment, ok := createdTokenAttachAttachment(attachEffect.References)
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
			{Primitive: game.Attach{
				Attachment: attachment,
				Target:     game.LinkedObjectReference(string(createdTokenLinkKey)),
			}},
		},
	}.Ability(), true
}

// createdTokenAttachAttachment resolves the entering Equipment fastened by a
// "create a token, then attach this Equipment to it" clause. The clause names two
// objects: the singular created-token back-reference ("it") and the entering
// Equipment ("this" / "this Equipment"), bound to the source or triggering event
// permanent. It returns the Equipment's object reference and fails closed unless
// exactly those two references are present.
func createdTokenAttachAttachment(references []compiler.CompiledReference) (game.ObjectReference, bool) {
	if len(references) != 2 {
		return game.ObjectReference{}, false
	}
	var equipment compiler.CompiledReference
	tokenRefs := 0
	for _, reference := range references {
		if reference.Kind == compiler.ReferencePronoun &&
			reference.Pronoun == compiler.ReferencePronounIt {
			tokenRefs++
			continue
		}
		equipment = reference
	}
	if tokenRefs != 1 {
		return game.ObjectReference{}, false
	}
	if equipment.Kind != compiler.ReferenceThisObject &&
		equipment.Kind != compiler.ReferenceThatObject {
		return game.ObjectReference{}, false
	}
	return lowerObjectReference(equipment,
		referenceLoweringContext{AllowSource: true, AllowEvent: true})
}
