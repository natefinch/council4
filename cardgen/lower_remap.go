package cardgen

import (
	"slices"

	"github.com/natefinch/council4/cardgen/oracle/compiler"
	"github.com/natefinch/council4/cardgen/oracle/parser"
	"github.com/natefinch/council4/cardgen/oracle/shared"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/opt"
)

// remapping to all target references in sequence. Unlike rebaseTargetedSequence
// which adds a uniform offset, this function looks up each local target index
// in localToGame and replaces it with the corresponding accumulated game index.
// This is needed for mixed inherited+owned target clauses where inherited
// targets live at their original accumulated indices while newly-owned targets
// start at a later position.
func remapTargetedSequence(sequence []game.Instruction, localToGame []int) bool {
	for i := range sequence {
		primitive, ok := remapTargetedPrimitive(sequence[i].Primitive, localToGame)
		if !ok {
			return false
		}
		sequence[i].Primitive = primitive
	}
	return true
}

func remapTargetedPrimitive(primitive game.Primitive, localToGame []int) (game.Primitive, bool) {
	// Explicit allowlist — same set as rebaseTargetedPrimitive.
	if value, ok := primitive.(game.Damage); ok {
		recipient, ok := remapDamageRecipient(value.Recipient, localToGame)
		if !ok {
			return nil, false
		}
		value.Recipient = recipient
		if value.DamageSource.Exists {
			source, ok := remapObjectReference(value.DamageSource.Val, localToGame)
			if !ok {
				return nil, false
			}
			value.DamageSource = opt.Val(source)
		}
		return value, true
	}
	if value, ok := primitive.(game.Destroy); ok {
		if value.Group.Valid() {
			return nil, false
		}
		value.Object, ok = remapObjectReference(value.Object, localToGame)
		return value, ok
	}
	if value, ok := primitive.(game.AddCounter); ok {
		value.Object, ok = remapObjectReference(value.Object, localToGame)
		return value, ok
	}
	if value, ok := primitive.(game.AddPlayerCounter); ok {
		value.Player, ok = remapPlayerReference(value.Player, localToGame)
		return value, ok
	}
	if value, ok := primitive.(game.ModifyPT); ok {
		value.Object, ok = remapObjectReference(value.Object, localToGame)
		return value, ok
	}
	if value, ok := primitive.(game.Fight); ok {
		var ok bool
		value.Object, ok = remapObjectReference(value.Object, localToGame)
		if !ok {
			return nil, false
		}
		value.RelatedObject, ok = remapObjectReference(value.RelatedObject, localToGame)
		return value, ok
	}
	if value, ok := primitive.(game.Tap); ok {
		value.Object, ok = remapObjectReference(value.Object, localToGame)
		return value, ok
	}
	if value, ok := primitive.(game.Untap); ok {
		if value.Group.Valid() {
			return nil, false
		}
		value.Object, ok = remapObjectReference(value.Object, localToGame)
		return value, ok
	}
	if value, ok := primitive.(game.Exile); ok {
		if value.Group.Valid() {
			return nil, false
		}
		value.Object, ok = remapObjectReference(value.Object, localToGame)
		return value, ok
	}
	if value, ok := primitive.(game.Bounce); ok {
		if value.Group.Valid() {
			return nil, false
		}
		value.Object, ok = remapObjectReference(value.Object, localToGame)
		return value, ok
	}
	if value, ok := primitive.(game.CounterObject); ok {
		value.Object, ok = remapObjectReference(value.Object, localToGame)
		return value, ok
	}
	if value, ok := primitive.(game.Regenerate); ok {
		value.Object, ok = remapObjectReference(value.Object, localToGame)
		return value, ok
	}
	if value, ok := primitive.(game.Draw); ok {
		value.Player, ok = remapPlayerReference(value.Player, localToGame)
		return value, ok
	}
	if value, ok := primitive.(game.Discard); ok {
		value.Player, ok = remapPlayerReference(value.Player, localToGame)
		return value, ok
	}
	if value, ok := primitive.(game.Mill); ok {
		value.Player, ok = remapPlayerReference(value.Player, localToGame)
		return value, ok
	}
	if value, ok := primitive.(game.GainLife); ok {
		value.Player, ok = remapPlayerReference(value.Player, localToGame)
		return value, ok
	}
	if value, ok := primitive.(game.LoseLife); ok {
		value.Player, ok = remapPlayerReference(value.Player, localToGame)
		return value, ok
	}
	if value, ok := primitive.(game.CreateDelayedTrigger); ok {
		return value, true
	}
	return nil, false
}

func remapDamageRecipient(recipient game.DamageRecipient, localToGame []int) (game.DamageRecipient, bool) {
	if object, ok := recipient.AnyTargetObjectReference(); ok {
		idx := object.TargetIndex()
		if idx < 0 || idx >= len(localToGame) {
			return game.DamageRecipient{}, false
		}
		return game.AnyTargetDamageRecipient(localToGame[idx]), true
	}
	if object, ok := recipient.ObjectReference(); ok {
		remapped, valid := remapObjectReference(object, localToGame)
		return game.ObjectDamageRecipient(remapped), valid
	}
	if player, ok := recipient.PlayerReference(); ok {
		remapped, valid := remapPlayerReference(player, localToGame)
		return game.PlayerDamageRecipient(remapped), valid
	}
	return game.DamageRecipient{}, false
}

func remapObjectReference(reference game.ObjectReference, localToGame []int) (game.ObjectReference, bool) {
	switch reference.Kind() {
	case game.ObjectReferenceTargetPermanent:
		idx := reference.TargetIndex()
		if idx < 0 || idx >= len(localToGame) {
			return game.ObjectReference{}, false
		}
		return game.TargetPermanentReference(localToGame[idx]), true
	case game.ObjectReferenceTargetStackObject:
		idx := reference.TargetIndex()
		if idx < 0 || idx >= len(localToGame) {
			return game.ObjectReference{}, false
		}
		return game.TargetStackObjectReference(localToGame[idx]), true
	case game.ObjectReferenceTargetAttachedPermanent:
		idx := reference.TargetIndex()
		if idx < 0 || idx >= len(localToGame) {
			return game.ObjectReference{}, false
		}
		return game.TargetAttachedPermanentReference(localToGame[idx]), true
	default:
		return reference, len(reference.Validate()) == 0
	}
}

func remapPlayerReference(reference game.PlayerReference, localToGame []int) (game.PlayerReference, bool) {
	switch reference.Kind() {
	case game.PlayerReferenceTargetPlayer:
		idx := reference.TargetIndex()
		if idx < 0 || idx >= len(localToGame) {
			return game.PlayerReference{}, false
		}
		return game.TargetPlayerReference(localToGame[idx]), true
	case game.PlayerReferenceObjectController, game.PlayerReferenceObjectOwner:
		object, ok := reference.Object()
		if !ok {
			return game.PlayerReference{}, false
		}
		object, ok = remapObjectReference(object, localToGame)
		if !ok {
			return game.PlayerReference{}, false
		}
		if reference.Kind() == game.PlayerReferenceObjectController {
			return game.ObjectControllerReference(object), true
		}
		return game.ObjectOwnerReference(object), true
	default:
		return reference, len(reference.Validate()) == 0
	}
}

func rebaseTargetedSequence(sequence []game.Instruction, offset int) bool {
	for i := range sequence {
		primitive, ok := rebaseTargetedPrimitive(sequence[i].Primitive, offset)
		if !ok {
			return false
		}
		sequence[i].Primitive = primitive
	}
	return true
}

func rebaseTargetedPrimitive(primitive game.Primitive, offset int) (game.Primitive, bool) {
	// Keep this as an explicit allowlist so a new target-bearing primitive cannot
	// silently retain a clause-local target index.
	if value, ok := primitive.(game.Damage); ok {
		recipient, ok := rebaseDamageRecipient(value.Recipient, offset)
		if !ok {
			return nil, false
		}
		value.Recipient = recipient
		if value.DamageSource.Exists {
			source, ok := rebaseObjectReference(value.DamageSource.Val, offset)
			if !ok {
				return nil, false
			}
			value.DamageSource = opt.Val(source)
		}
		return value, true
	}
	if value, ok := primitive.(game.Destroy); ok {
		if value.Group.Valid() {
			return nil, false
		}
		value.Object, ok = rebaseObjectReference(value.Object, offset)
		return value, ok
	}
	if value, ok := primitive.(game.AddCounter); ok {
		value.Object, ok = rebaseObjectReference(value.Object, offset)
		return value, ok
	}
	if value, ok := primitive.(game.AddPlayerCounter); ok {
		value.Player, ok = rebasePlayerReference(value.Player, offset)
		return value, ok
	}
	if value, ok := primitive.(game.ModifyPT); ok {
		value.Object, ok = rebaseObjectReference(value.Object, offset)
		return value, ok
	}
	if value, ok := primitive.(game.Fight); ok {
		var ok bool
		value.Object, ok = rebaseObjectReference(value.Object, offset)
		if !ok {
			return nil, false
		}
		value.RelatedObject, ok = rebaseObjectReference(value.RelatedObject, offset)
		return value, ok
	}
	if value, ok := primitive.(game.Tap); ok {
		value.Object, ok = rebaseObjectReference(value.Object, offset)
		return value, ok
	}
	if value, ok := primitive.(game.Untap); ok {
		if value.Group.Valid() {
			return nil, false
		}
		value.Object, ok = rebaseObjectReference(value.Object, offset)
		return value, ok
	}
	if value, ok := primitive.(game.Exile); ok {
		if value.Group.Valid() {
			return nil, false
		}
		value.Object, ok = rebaseObjectReference(value.Object, offset)
		return value, ok
	}
	if value, ok := primitive.(game.Bounce); ok {
		if value.Group.Valid() {
			return nil, false
		}
		value.Object, ok = rebaseObjectReference(value.Object, offset)
		return value, ok
	}
	if value, ok := primitive.(game.CounterObject); ok {
		value.Object, ok = rebaseObjectReference(value.Object, offset)
		return value, ok
	}
	if value, ok := primitive.(game.Regenerate); ok {
		value.Object, ok = rebaseObjectReference(value.Object, offset)
		return value, ok
	}
	if value, ok := primitive.(game.Draw); ok {
		value.Player, ok = rebasePlayerReference(value.Player, offset)
		return value, ok
	}
	if value, ok := primitive.(game.Discard); ok {
		value.Player, ok = rebasePlayerReference(value.Player, offset)
		return value, ok
	}
	if value, ok := primitive.(game.Mill); ok {
		value.Player, ok = rebasePlayerReference(value.Player, offset)
		return value, ok
	}
	if value, ok := primitive.(game.GainLife); ok {
		value.Player, ok = rebasePlayerReference(value.Player, offset)
		return value, ok
	}
	if value, ok := primitive.(game.LoseLife); ok {
		value.Player, ok = rebasePlayerReference(value.Player, offset)
		return value, ok
	}
	if value, ok := primitive.(game.CreateDelayedTrigger); ok {
		return value, true
	}
	if value, ok := primitive.(game.ApplyContinuous); ok {
		if value.Object.Exists {
			rebased, ok := rebaseObjectReference(value.Object.Val, offset)
			if !ok {
				return nil, false
			}
			value.Object = opt.Val(rebased)
		}
		return value, true
	}
	return nil, false
}

func rebaseDamageRecipient(recipient game.DamageRecipient, offset int) (game.DamageRecipient, bool) {
	if object, ok := recipient.AnyTargetObjectReference(); ok {
		return game.AnyTargetDamageRecipient(object.TargetIndex() + offset), true
	}
	if object, ok := recipient.ObjectReference(); ok {
		rebased, valid := rebaseObjectReference(object, offset)
		return game.ObjectDamageRecipient(rebased), valid
	}
	if player, ok := recipient.PlayerReference(); ok {
		rebased, valid := rebasePlayerReference(player, offset)
		return game.PlayerDamageRecipient(rebased), valid
	}
	return game.DamageRecipient{}, false
}

func rebaseObjectReference(reference game.ObjectReference, offset int) (game.ObjectReference, bool) {
	switch reference.Kind() {
	case game.ObjectReferenceTargetPermanent:
		return game.TargetPermanentReference(reference.TargetIndex() + offset), true
	case game.ObjectReferenceTargetStackObject:
		return game.TargetStackObjectReference(reference.TargetIndex() + offset), true
	case game.ObjectReferenceTargetAttachedPermanent:
		return game.TargetAttachedPermanentReference(reference.TargetIndex() + offset), true
	default:
		return reference, len(reference.Validate()) == 0
	}
}

func rebasePlayerReference(reference game.PlayerReference, offset int) (game.PlayerReference, bool) {
	switch reference.Kind() {
	case game.PlayerReferenceTargetPlayer:
		return game.TargetPlayerReference(reference.TargetIndex() + offset), true
	case game.PlayerReferenceObjectController, game.PlayerReferenceObjectOwner:
		object, ok := reference.Object()
		if !ok {
			return game.PlayerReference{}, false
		}
		object, ok = rebaseObjectReference(object, offset)
		if !ok {
			return game.PlayerReference{}, false
		}
		if reference.Kind() == game.PlayerReferenceObjectController {
			return game.ObjectControllerReference(object), true
		}
		return game.ObjectOwnerReference(object), true
	default:
		return reference, len(reference.Validate()) == 0
	}
}

func contextForEffect(
	ctx contentCtx,
	effect *compiler.CompiledEffect,
) contentCtx {
	ctx.text = effect.Text
	ctx.span = effect.Span
	resolvedEffect := *effect
	resolvedEffect.RequiresOrderedLowering = false
	ctx.content.Effects = []compiler.CompiledEffect{resolvedEffect}
	ctx.content.Targets = effect.Targets
	ctx.content.Keywords = keywordsWithinSpan(ctx.content.Keywords, effect.ClauseSpan)
	ctx.content.References = effect.References
	return ctx
}

func targetsWithinSpan(targets []compiler.CompiledTarget, span shared.Span) []compiler.CompiledTarget {
	var within []compiler.CompiledTarget
	for _, target := range targets {
		if spanCovered(target.Span, []shared.Span{span}) {
			within = append(within, target)
		}
	}
	return within
}

func keywordsWithinSpan(keywords []compiler.CompiledKeyword, span shared.Span) []compiler.CompiledKeyword {
	var within []compiler.CompiledKeyword
	for _, keyword := range keywords {
		if spanCovered(keyword.Span, []shared.Span{span}) {
			within = append(within, keyword)
		}
	}
	return within
}

func referencesWithinSpan(references []compiler.CompiledReference, span shared.Span) []compiler.CompiledReference {
	var within []compiler.CompiledReference
	for _, reference := range references {
		if spanCovered(reference.Span, []shared.Span{span}) {
			within = append(within, reference)
		}
	}
	return within
}

func syntaxWithinSpan(syntax *parser.Ability, span shared.Span) parser.Ability {
	result := *syntax
	result.Span = span
	result.Text = ""
	result.Tokens = slices.DeleteFunc(
		append([]shared.Token(nil), syntax.Tokens...),
		func(token shared.Token) bool {
			return !spanCovered(token.Span, []shared.Span{span})
		},
	)
	return result
}

// splitEffectSyntaxes clips retained diagnostic syntax to parser-owned clause
// spans. Appending sentence punctuation does not derive semantic ownership.
func splitEffectSyntaxes(syntax *parser.Ability, effects []compiler.CompiledEffect) []parser.Ability {
	clauses := make([]parser.Ability, len(effects))
	for i := range effects {
		clauses[i] = syntaxWithinSpan(syntax, effects[i].ClauseSpan)
		if len(clauses[i].Tokens) == 0 || clauses[i].Tokens[len(clauses[i].Tokens)-1].Kind == shared.Period {
			continue
		}
		sentence := syntaxWithinSpan(syntax, effects[i].Span)
		if len(sentence.Tokens) > 0 && sentence.Tokens[len(sentence.Tokens)-1].Kind == shared.Period {
			clauses[i].Tokens = append(clauses[i].Tokens, sentence.Tokens[len(sentence.Tokens)-1])
		}
	}
	return clauses
}

func priorSubjectTargets(effects []compiler.CompiledEffect, index int) []compiler.CompiledTarget {
	for i := index - 1; i >= 0; i-- {
		if len(effects[i].SubjectTargets) > 0 {
			return effects[i].SubjectTargets
		}
		if effects[i].Context != parser.EffectContextPriorSubject {
			break
		}
	}
	return nil
}

func priorSubjectContext(effects []compiler.CompiledEffect, index int) parser.EffectContextKind {
	for i := index - 1; i >= 0; i-- {
		if effects[i].Context != parser.EffectContextPriorSubject {
			return effects[i].Context
		}
	}
	return parser.EffectContextUnknown
}

func priorSubjectReferences(effects []compiler.CompiledEffect, index int) []compiler.CompiledReference {
	for i := index - 1; i >= 0; i-- {
		if len(effects[i].SubjectReferences) > 0 {
			return effects[i].SubjectReferences
		}
		if effects[i].Context != parser.EffectContextPriorSubject {
			break
		}
	}
	return nil
}

func unsupportedEffectSequenceDiagnostic(ctx contentCtx) *shared.Diagnostic {
	return contentDiagnostic(
		ctx,
		"unsupported ordered effect sequence",
		"the executable source backend supports only exact ordered sequences of independently supported effects",
	)
}
