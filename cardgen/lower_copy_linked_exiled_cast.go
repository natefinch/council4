package cardgen

import (
	"github.com/natefinch/council4/cardgen/oracle/compiler"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/opt"
)

// copyLinkedExiledCardResultKey gates the imprint cast-the-copy instruction on
// the preceding copy-consent instruction: the controller may cast the copy only
// if a linked exiled card existed and was copied ("If you do, ...").
const copyLinkedExiledCardResultKey = game.ResultKey("imprint-copy-made")

// lowerCopyLinkedExiledCardCast lowers the imprint copy/cast idiom "You may copy
// the exiled card. If you do, you may cast the copy without paying its mana
// cost." (CR 707.12; Isochron Scepter, Spellbinder) into a two-instruction
// sequence composing the generic CopyCard and PlayLinkedExiledCard primitives.
//
// The parser marks the two sentences with typed flags (CopyLinkedExiledCard on
// the EffectCopyStackObject copy-consent effect, CastLinkedExiledCopy on the
// EffectCast cast-the-copy effect) and consumes the "its" possessive, so the
// body carries no bound reference. The compiler emits the "If you do" reflexive
// gate as a PriorInstructionAccepted condition, which this lowering subsumes by
// wiring the cast instruction's result gate to the copy instruction.
//
// The first instruction offers the controller the copy (Optional consent) and
// succeeds only when a card linked to this source under the imprint link still
// rests in exile. The second instruction, gated on that success, casts a free
// copy of the imprinted exiled card, choosing first-legal targets with X as 0
// and leaving the original card in exile. Any other shape fails closed.
func lowerCopyLinkedExiledCardCast(ctx contentCtx) (game.AbilityContent, bool) {
	if len(ctx.content.Effects) != 2 ||
		len(ctx.content.Targets) != 0 ||
		len(ctx.content.Modes) != 0 ||
		len(ctx.content.Keywords) != 0 {
		return game.AbilityContent{}, false
	}
	copyEffect := ctx.content.Effects[0]
	castEffect := ctx.content.Effects[1]
	if !copyLinkedExiledCardEffect(copyEffect) || !castLinkedExiledCopyEffect(castEffect) {
		return game.AbilityContent{}, false
	}
	if !copyLinkedExiledCardConditions(ctx.content.Conditions) {
		return game.AbilityContent{}, false
	}
	sequence := []game.Instruction{
		{
			Optional: copyEffect.Optional,
			Primitive: game.CopyCard{
				Player: game.ControllerReference(),
				LinkID: imprintLinkKey,
			},
			PublishResult: copyLinkedExiledCardResultKey,
		},
		{
			Optional: castEffect.Optional,
			ResultGate: opt.Val(game.InstructionResultGate{
				Key:       copyLinkedExiledCardResultKey,
				Succeeded: game.TriTrue,
			}),
			Primitive: game.PlayLinkedExiledCard{
				Player:                game.ControllerReference(),
				LinkID:                imprintLinkKey,
				Copy:                  true,
				WithoutPayingManaCost: true,
			},
		},
	}
	return game.Mode{Sequence: sequence}.Ability(), true
}

// recognizeCopyLinkedExiledCardCast reports whether an ability body is exactly
// the imprint copy/cast idiom: a two-effect body pairing the typed
// copy-linked-exiled-card consent with the cast-linked-exiled-copy consequence
// under only its implied prior-instruction gate. It backs both the activation
// reference gate (the idiom owns its consumed "its" possessive) and the lowerer.
func recognizeCopyLinkedExiledCardCast(content compiler.AbilityContent) bool {
	return len(content.Effects) == 2 &&
		len(content.Targets) == 0 &&
		len(content.Modes) == 0 &&
		len(content.Keywords) == 0 &&
		copyLinkedExiledCardEffect(content.Effects[0]) &&
		castLinkedExiledCopyEffect(content.Effects[1]) &&
		copyLinkedExiledCardConditions(content.Conditions)
}

// copyLinkedExiledCardEffect reports whether an effect is the typed imprint
// copy-consent effect ("You may copy the exiled card.").
func copyLinkedExiledCardEffect(effect compiler.CompiledEffect) bool {
	return effect.Kind == compiler.EffectCopyStackObject &&
		effect.CopyLinkedExiledCard &&
		!effect.Negated &&
		len(effect.Targets) == 0 &&
		len(effect.References) == 0
}

// castLinkedExiledCopyEffect reports whether an effect is the typed imprint
// cast-the-copy effect ("... you may cast the copy without paying its mana
// cost.").
func castLinkedExiledCopyEffect(effect compiler.CompiledEffect) bool {
	return effect.Kind == compiler.EffectCast &&
		effect.CastLinkedExiledCopy &&
		effect.CastWithoutPayingManaCost &&
		!effect.Negated &&
		len(effect.Targets) == 0 &&
		len(effect.References) == 0
}

// copyLinkedExiledCardConditions reports whether the body carries only its
// implied "If you do" reflexive gate (a single PriorInstructionAccepted
// condition), which the sequence subsumes through the cast instruction's result
// gate. Any other condition is rejected.
func copyLinkedExiledCardConditions(conditions []compiler.CompiledCondition) bool {
	switch len(conditions) {
	case 0:
		return true
	case 1:
		return conditions[0].Predicate == compiler.ConditionPredicatePriorInstructionAccepted
	default:
		return false
	}
}
