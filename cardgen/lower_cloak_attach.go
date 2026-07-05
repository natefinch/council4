package cardgen

import (
	"github.com/natefinch/council4/cardgen/oracle/compiler"
	"github.com/natefinch/council4/cardgen/oracle/parser"
	"github.com/natefinch/council4/mtg/game"
)

// lowerCloakThenAttachSequence lowers the ordered pair "Cloak the top card of
// your library, then attach this Equipment to it." (Cryptic Coat). The first
// clause cloaks the top card — putting it onto the battlefield face down as a
// 2/2 creature with ward {2} — and the second fastens the entering Equipment
// onto that just-cloaked permanent. The attach clause names two objects: the
// entering Equipment ("this" / "this Equipment") and the cloaked permanent
// ("it"), neither of which is a chosen target. The lowering reconstructs the
// link structurally — publishing the cloaked permanent under
// manifestedCreatureLinkKey and pointing the Attach at that linked object —
// mirroring lowerCreateTokenThenAttachSequence and lowerManifestDreadThenCountersSequence.
// It fails closed for targets, conditions, modes, durations, optional bodies, or
// any other shape.
func lowerCloakThenAttachSequence(ctx contentCtx) (game.AbilityContent, bool) {
	if len(ctx.content.Effects) != 2 ||
		len(ctx.content.Targets) != 0 ||
		len(ctx.content.Conditions) != 0 ||
		len(ctx.content.Modes) != 0 ||
		ctx.optional {
		return game.AbilityContent{}, false
	}
	cloakEffect := ctx.content.Effects[0]
	attachEffect := ctx.content.Effects[1]
	if cloakEffect.Kind != compiler.EffectCloak ||
		attachEffect.Kind != compiler.EffectAttach ||
		cloakEffect.Negated ||
		attachEffect.Negated ||
		cloakEffect.Optional ||
		attachEffect.Optional ||
		!cloakEffect.Exact ||
		cloakEffect.Context != parser.EffectContextController ||
		attachEffect.Context != parser.EffectContextController ||
		attachEffect.Duration != compiler.DurationNone {
		return game.AbilityContent{}, false
	}
	attachment, ok := createdTokenAttachAttachment(attachEffect.References)
	if !ok {
		return game.AbilityContent{}, false
	}
	return game.Mode{
		Sequence: []game.Instruction{
			{Primitive: game.Manifest{Cloak: true, PublishLinked: manifestedCreatureLinkKey}},
			{Primitive: game.Attach{
				Attachment: attachment,
				Target:     game.LinkedObjectReference(string(manifestedCreatureLinkKey)),
			}},
		},
	}.Ability(), true
}
