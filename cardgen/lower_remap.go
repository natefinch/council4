package cardgen

import (
	"slices"

	"github.com/natefinch/council4/cardgen/oracle/compiler"
	"github.com/natefinch/council4/cardgen/oracle/parser"
	"github.com/natefinch/council4/cardgen/oracle/shared"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/opt"
)

// targetIndexKind identifies the numbering domain of a clause-local target
// index so a transform can map it correctly. Object, player, stack, and
// attached-permanent target references share the stack object's global target
// list (targetIndexObject); card-reference target indices are numbered among
// card targets only (targetIndexCard).
type targetIndexKind uint8

const (
	targetIndexObject targetIndexKind = iota
	targetIndexCard
)

// targetIndexTransform maps a clause-local target index in the given numbering
// domain to its accumulated game position, returning false if it cannot be
// expressed (the carrying primitive then fails closed).
type targetIndexTransform func(kind targetIndexKind, old int) (int, bool)

// remapTargetedSequence rewrites every target reference in a clause's primitives
// by looking each clause-local index up in localToGame. Unlike
// rebaseTargetedSequence, which adds a uniform offset, this replaces each local
// target index with the corresponding accumulated game index. This is needed for
// mixed inherited+owned target clauses where inherited targets live at their
// original accumulated indices while newly-owned targets start at a later
// position. Card-reference target indices are numbered among card targets only,
// which the global localToGame table cannot express, so a card-moving primitive
// in a mixed clause fails closed rather than risk a wrong slot.
func remapTargetedSequence(sequence []game.Instruction, localToGame []int) bool {
	transform := func(kind targetIndexKind, old int) (int, bool) {
		if kind != targetIndexObject {
			return 0, false
		}
		if old < 0 || old >= len(localToGame) {
			return 0, false
		}
		return localToGame[old], true
	}
	return transformTargetedSequence(sequence, transform)
}

// rebaseTargetedSequence shifts every target reference in a clause's primitives
// to its accumulated game position. offset is the number of preceding
// accumulated target specs (the base for object/player target indices, which are
// global positions in the stack object's target list). cardOffset is the number
// of preceding accumulated card target specs (the base for card-reference target
// indices, which the runtime counts among card targets only). The two bases
// coincide unless a non-card target spec precedes a card reference.
func rebaseTargetedSequence(sequence []game.Instruction, offset, cardOffset int) bool {
	return transformTargetedSequence(sequence, rebaseTransform(offset, cardOffset))
}

// rebaseTransform builds the uniform-offset transform shared by the rebase entry
// points: object-domain indices shift by offset, card-domain indices by
// cardOffset, and both always succeed.
func rebaseTransform(offset, cardOffset int) targetIndexTransform {
	return func(kind targetIndexKind, old int) (int, bool) {
		if kind == targetIndexCard {
			return old + cardOffset, true
		}
		return old + offset, true
	}
}

func transformTargetedSequence(sequence []game.Instruction, transform targetIndexTransform) bool {
	for i := range sequence {
		primitive, ok := transformPrimitiveTargetIndices(sequence[i].Primitive, transform)
		if !ok {
			return false
		}
		sequence[i].Primitive = primitive
	}
	return true
}

// rebaseTargetedPrimitive rebases a single primitive's target references by a
// uniform offset. It is a thin adapter over the shared
// transformPrimitiveTargetIndices walker, retained for targeted tests.
func rebaseTargetedPrimitive(primitive game.Primitive, offset, cardOffset int) (game.Primitive, bool) {
	return transformPrimitiveTargetIndices(primitive, rebaseTransform(offset, cardOffset))
}

// transformPrimitiveTargetIndices is the single target-index traversal shared by
// the remap (lookup-table) and rebase (uniform-offset) paths. It walks every
// target-bearing primitive variant — including nested damage recipients and
// amounts, prevent-damage shields, continuous-effect and token recipients, and
// zone-movement card/player references — applying transform to each clause-local
// target index in its proper numbering domain. This is the ONE place to extend
// when a new target-bearing primitive kind is added; the completeness test in
// lower_remap_test.go guards every variant so a new primitive cannot silently
// omit transform support. Keep it as an explicit allowlist so an unhandled
// target-bearing primitive fails closed rather than retain a clause-local index.
func transformPrimitiveTargetIndices(primitive game.Primitive, transform targetIndexTransform) (game.Primitive, bool) {
	if value, ok := primitive.(game.Damage); ok {
		recipient, ok := transformDamageRecipient(value.Recipient, transform)
		if !ok {
			return nil, false
		}
		value.Recipient = recipient
		if value.DamageSource.Exists {
			source, ok := transformObjectReference(value.DamageSource.Val, transform)
			if !ok {
				return nil, false
			}
			value.DamageSource = opt.Val(source)
		}
		amount, ok := transformQuantity(value.Amount, transform)
		if !ok {
			return nil, false
		}
		value.Amount = amount
		return value, true
	}
	if value, ok := primitive.(game.Destroy); ok {
		if value.Group.Valid() {
			return nil, false
		}
		value.Object, ok = transformObjectReference(value.Object, transform)
		return value, ok
	}
	if value, ok := primitive.(game.AddCounter); ok {
		value.Object, ok = transformObjectReference(value.Object, transform)
		if !ok {
			return nil, false
		}
		value.Amount, ok = transformQuantity(value.Amount, transform)
		return value, ok
	}
	if value, ok := primitive.(game.AddPlayerCounter); ok {
		value.Player, ok = transformPlayerReference(value.Player, transform)
		return value, ok
	}
	if value, ok := primitive.(game.ModifyPT); ok {
		value.Object, ok = transformObjectReference(value.Object, transform)
		return value, ok
	}
	if value, ok := primitive.(game.Fight); ok {
		var ok bool
		value.Object, ok = transformObjectReference(value.Object, transform)
		if !ok {
			return nil, false
		}
		value.RelatedObject, ok = transformObjectReference(value.RelatedObject, transform)
		return value, ok
	}
	if value, ok := primitive.(game.Tap); ok {
		value.Object, ok = transformObjectReference(value.Object, transform)
		return value, ok
	}
	if value, ok := primitive.(game.TapOrUntap); ok {
		value.Object, ok = transformObjectReference(value.Object, transform)
		return value, ok
	}
	if value, ok := primitive.(game.SkipNextUntap); ok {
		value.Object, ok = transformObjectReference(value.Object, transform)
		return value, ok
	}
	if value, ok := primitive.(game.PreventDamage); ok {
		return transformPreventDamage(value, transform)
	}
	if value, ok := primitive.(game.Untap); ok {
		if value.Group.Valid() {
			return nil, false
		}
		value.Object, ok = transformObjectReference(value.Object, transform)
		return value, ok
	}
	if value, ok := primitive.(game.RemoveFromCombat); ok {
		value.Object, ok = transformObjectReference(value.Object, transform)
		return value, ok
	}
	if value, ok := primitive.(game.LookAtHand); ok {
		value.Player, ok = transformPlayerReference(value.Player, transform)
		return value, ok
	}
	if value, ok := primitive.(game.BecomeMonarch); ok {
		value.Player, ok = transformPlayerReference(value.Player, transform)
		return value, ok
	}
	if value, ok := primitive.(game.Exile); ok {
		if value.Group.Valid() {
			return nil, false
		}
		value.Object, ok = transformObjectReference(value.Object, transform)
		return value, ok
	}
	if value, ok := primitive.(game.Bounce); ok {
		if value.Group.Valid() {
			return nil, false
		}
		value.Object, ok = transformObjectReference(value.Object, transform)
		return value, ok
	}
	if value, ok := primitive.(game.CounterObject); ok {
		value.Object, ok = transformObjectReference(value.Object, transform)
		return value, ok
	}
	if value, ok := primitive.(game.Regenerate); ok {
		value.Object, ok = transformObjectReference(value.Object, transform)
		return value, ok
	}
	if value, ok := primitive.(game.Attach); ok {
		// Attachment names the entering or source Equipment (a source/event
		// reference) and passes through unchanged; Target carries the chosen
		// permanent's clause-local target index, which is rewritten here so an
		// auto-attach clause can appear in an ordered sequence ("attach it to
		// target creature you control. That creature gains <keyword> ...").
		attachment, ok := transformObjectReference(value.Attachment, transform)
		if !ok {
			return nil, false
		}
		value.Attachment = attachment
		value.Target, ok = transformObjectReference(value.Target, transform)
		return value, ok
	}
	if value, ok := primitive.(game.Draw); ok {
		if value.PlayerGroup.Kind != game.PlayerGroupReferenceNone {
			return value, true
		}
		value.Player, ok = transformPlayerReference(value.Player, transform)
		return value, ok
	}
	if value, ok := primitive.(game.Discard); ok {
		if value.PlayerGroup.Kind != game.PlayerGroupReferenceNone {
			return value, true
		}
		value.Player, ok = transformPlayerReference(value.Player, transform)
		return value, ok
	}
	if value, ok := primitive.(game.Mill); ok {
		if value.PlayerGroup.Kind != game.PlayerGroupReferenceNone {
			return value, true
		}
		value.Player, ok = transformPlayerReference(value.Player, transform)
		return value, ok
	}
	if value, ok := primitive.(game.ExileTopOfLibrary); ok {
		if value.PlayerGroup.Kind != game.PlayerGroupReferenceNone {
			return value, true
		}
		value.Player, ok = transformPlayerReference(value.Player, transform)
		return value, ok
	}
	if value, ok := primitive.(game.RevealUntil); ok {
		if value.PlayerGroup.Kind != game.PlayerGroupReferenceNone {
			return value, true
		}
		value.Player, ok = transformPlayerReference(value.Player, transform)
		return value, ok
	}
	if value, ok := primitive.(game.RevealTopPartition); ok {
		value.Player, ok = transformPlayerReference(value.Player, transform)
		return value, ok
	}
	if value, ok := primitive.(game.GainLife); ok {
		value.Player, ok = transformPlayerReference(value.Player, transform)
		return value, ok
	}
	if value, ok := primitive.(game.LoseLife); ok {
		value.Player, ok = transformPlayerReference(value.Player, transform)
		return value, ok
	}
	if value, ok := primitive.(game.SacrificePermanents); ok {
		// The player-group form ("Each opponent sacrifices ...") carries no
		// clause-local target index; only the single-player form ("Target player
		// sacrifices ...") does, so transform that and leave the group form
		// unchanged. Selection is a permanent filter and carries no target index.
		if value.Player.Kind() == game.PlayerReferenceNone {
			return value, true
		}
		value.Player, ok = transformPlayerReference(value.Player, transform)
		return value, ok
	}
	if value, ok := primitive.(game.CreateDelayedTrigger); ok {
		return value, true
	}
	if value, ok := primitive.(game.ApplyContinuous); ok {
		if value.Object.Exists {
			transformed, ok := transformObjectReference(value.Object.Val, transform)
			if !ok {
				return nil, false
			}
			value.Object = opt.Val(transformed)
		}
		return value, true
	}
	if value, ok := primitive.(game.ApplyRule); ok {
		if value.Object.Exists {
			transformed, ok := transformObjectReference(value.Object.Val, transform)
			if !ok {
				return nil, false
			}
			value.Object = opt.Val(transformed)
		}
		return value, true
	}
	if value, ok := primitive.(game.CreateToken); ok {
		if value.Recipient.Exists {
			transformed, ok := transformPlayerReference(value.Recipient.Val, transform)
			if !ok {
				return nil, false
			}
			value.Recipient = opt.Val(transformed)
		}
		return value, true
	}
	return transformZonePrimitiveTargetIndices(primitive, transform)
}

// transformZonePrimitiveTargetIndices handles the zone-movement primitives whose
// target references span players and cards, split out of
// transformPrimitiveTargetIndices to keep that allowlist's maintainability index
// within bounds.
func transformZonePrimitiveTargetIndices(primitive game.Primitive, transform targetIndexTransform) (game.Primitive, bool) {
	if value, ok := primitive.(game.MoveCard); ok {
		// The player-zone group form ("Exile target player's graveyard.") carries
		// a target-bearing Player reference; transform it against the accumulated
		// target list and fail closed if it cannot be expressed. The single-card
		// form leaves Player unset and transforms its Card slot instead.
		if value.Player.Kind() != game.PlayerReferenceNone {
			transformed, ok := transformPlayerReference(value.Player, transform)
			if !ok {
				return nil, false
			}
			value.Player = transformed
			return value, true
		}
		card, ok := transformCardReference(value.Card, transform)
		if !ok {
			return nil, false
		}
		value.Card = card
		return value, true
	}
	if value, ok := primitive.(game.PutOnBattlefield); ok {
		// Entry counters and continuous effects may embed their own target
		// references; transforming those is not modeled, so fail closed rather
		// than leave a clause-local index pointing at the wrong accumulated
		// target.
		if len(value.EntryCounters) != 0 || len(value.ContinuousEffects) != 0 {
			return nil, false
		}
		source, ok := transformBattlefieldSource(value.Source, transform)
		if !ok {
			return nil, false
		}
		value.Source = source
		if value.Recipient.Exists {
			recipient, ok := transformPlayerReference(value.Recipient.Val, transform)
			if !ok {
				return nil, false
			}
			value.Recipient = opt.Val(recipient)
		}
		return value, true
	}
	return nil, false
}

// transformPreventDamage transforms a prevent-damage primitive's target
// reference. The shield can reference either a target player or a target object;
// only one is set, so transform whichever the primitive carries. A global shield
// carries no target index and passes through.
func transformPreventDamage(value game.PreventDamage, transform targetIndexTransform) (game.Primitive, bool) {
	if value.Global {
		return value, true
	}
	var ok bool
	if _, anyTarget := value.AnyTarget.AnyTargetObjectReference(); anyTarget {
		value.AnyTarget, ok = transformDamageRecipient(value.AnyTarget, transform)
		return value, ok
	}
	if value.Player.Kind() != game.PlayerReferenceNone {
		value.Player, ok = transformPlayerReference(value.Player, transform)
		return value, ok
	}
	value.Object, ok = transformObjectReference(value.Object, transform)
	return value, ok
}

func transformDamageRecipient(recipient game.DamageRecipient, transform targetIndexTransform) (game.DamageRecipient, bool) {
	if object, ok := recipient.AnyTargetObjectReference(); ok {
		idx, ok := transform(targetIndexObject, object.TargetIndex())
		if !ok {
			return game.DamageRecipient{}, false
		}
		return game.AnyTargetDamageRecipient(idx), true
	}
	if object, ok := recipient.ObjectReference(); ok {
		transformed, valid := transformObjectReference(object, transform)
		return game.ObjectDamageRecipient(transformed), valid
	}
	if player, ok := recipient.PlayerReference(); ok {
		transformed, valid := transformPlayerReference(player, transform)
		return game.PlayerDamageRecipient(transformed), valid
	}
	if group, ok := recipient.GroupReference(); ok {
		transformed, valid := transformGroupReference(group, transform)
		return game.GroupDamageRecipient(transformed), valid
	}
	if _, ok := recipient.PlayerGroupReference(); ok {
		// An opponents/all-players group carries no target index to transform.
		return recipient, true
	}
	return game.DamageRecipient{}, false
}

// objectReferenceCarriesTargetIndex reports whether an object reference embeds a
// clause-local target index that must be transformed when a primitive moves into
// an accumulated sequence target list.
func objectReferenceCarriesTargetIndex(reference game.ObjectReference) bool {
	switch reference.Kind() {
	case game.ObjectReferenceTargetPermanent,
		game.ObjectReferenceTargetStackObject,
		game.ObjectReferenceTargetAttachedPermanent,
		game.ObjectReferenceTargetObject:
		return true
	default:
		return false
	}
}

// transformQuantity transforms a Quantity whose dynamic formula reads a target's
// value (e.g. DynamicAmountObjectPower for "equal to its power"). Fixed amounts
// and dynamic formulas that do not reference a target are returned unchanged.
func transformQuantity(amount game.Quantity, transform targetIndexTransform) (game.Quantity, bool) {
	dynamic := amount.DynamicAmount()
	if !dynamic.Exists || !objectReferenceCarriesTargetIndex(dynamic.Val.Object) {
		return amount, true
	}
	object, ok := transformObjectReference(dynamic.Val.Object, transform)
	if !ok {
		return game.Quantity{}, false
	}
	value := dynamic.Val
	value.Object = object
	return game.Dynamic(value), true
}

func transformObjectReference(reference game.ObjectReference, transform targetIndexTransform) (game.ObjectReference, bool) {
	switch reference.Kind() {
	case game.ObjectReferenceTargetPermanent:
		idx, ok := transform(targetIndexObject, reference.TargetIndex())
		if !ok {
			return game.ObjectReference{}, false
		}
		return game.TargetPermanentReference(idx), true
	case game.ObjectReferenceTargetStackObject:
		idx, ok := transform(targetIndexObject, reference.TargetIndex())
		if !ok {
			return game.ObjectReference{}, false
		}
		return game.TargetStackObjectReference(idx), true
	case game.ObjectReferenceTargetObject:
		idx, ok := transform(targetIndexObject, reference.TargetIndex())
		if !ok {
			return game.ObjectReference{}, false
		}
		return game.TargetObjectReference(idx), true
	case game.ObjectReferenceTargetAttachedPermanent:
		idx, ok := transform(targetIndexObject, reference.TargetIndex())
		if !ok {
			return game.ObjectReference{}, false
		}
		return game.TargetAttachedPermanentReference(idx), true
	default:
		return reference, len(reference.Validate()) == 0
	}
}

func transformPlayerReference(reference game.PlayerReference, transform targetIndexTransform) (game.PlayerReference, bool) {
	switch reference.Kind() {
	case game.PlayerReferenceTargetPlayer:
		idx, ok := transform(targetIndexObject, reference.TargetIndex())
		if !ok {
			return game.PlayerReference{}, false
		}
		return game.TargetPlayerReference(idx), true
	case game.PlayerReferenceObjectController, game.PlayerReferenceObjectOwner:
		object, ok := reference.Object()
		if !ok {
			return game.PlayerReference{}, false
		}
		object, ok = transformObjectReference(object, transform)
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

// transformCardReference transforms a target-card reference's slot in the card
// numbering domain. Non-target card references (source, event, linked) carry no
// target index and pass through unchanged.
func transformCardReference(reference game.CardReference, transform targetIndexTransform) (game.CardReference, bool) {
	if reference.Kind != game.CardReferenceTarget {
		return reference, true
	}
	idx, ok := transform(targetIndexCard, reference.TargetIndex)
	if !ok {
		return game.CardReference{}, false
	}
	reference.TargetIndex = idx
	return reference, true
}

// transformBattlefieldSource transforms the card slot of a card-backed
// battlefield source in the card numbering domain. Linked sources carry no
// target index and pass through.
func transformBattlefieldSource(source game.BattlefieldSource, transform targetIndexTransform) (game.BattlefieldSource, bool) {
	if card, ok := source.CardRef(); ok {
		transformed, ok := transformCardReference(card, transform)
		if !ok {
			return game.BattlefieldSource{}, false
		}
		return game.CardBattlefieldSource(transformed), true
	}
	return source, source.Valid()
}

// transformGroupReference transforms the anchor, exclusion, and player-anchor
// references of a battlefield, object-controlled, or player-controlled group,
// leaving its characteristic Selection (which carries no target index)
// unchanged. It backs the inherited source-power group damage shape ("It deals
// damage equal to its power to each other creature."), whose recipient group
// excludes the dealing target, and the targeted-player counter placement ("Put a
// +1/+1 counter on each creature target player controls."), whose recipient
// group's player anchor moves with the remapped player target slot. It fails
// closed for any other group domain.
func transformGroupReference(group game.GroupReference, transform targetIndexTransform) (game.GroupReference, bool) {
	selection := group.Selection()
	var anchor opt.V[game.ObjectReference]
	if a, ok := group.Anchor(); ok {
		transformed, valid := transformObjectReference(a, transform)
		if !valid {
			return game.GroupReference{}, false
		}
		anchor = opt.Val(transformed)
	}
	var exclude opt.V[game.ObjectReference]
	if e, ok := group.Exclusion(); ok {
		transformed, valid := transformObjectReference(e, transform)
		if !valid {
			return game.GroupReference{}, false
		}
		exclude = opt.Val(transformed)
	}
	switch group.Domain() {
	case game.GroupDomainBattlefield:
		if exclude.Exists {
			return game.BattlefieldGroupExcluding(selection, exclude.Val), true
		}
		return game.BattlefieldGroup(selection), true
	case game.GroupDomainObjectControlled:
		if !anchor.Exists {
			return game.GroupReference{}, false
		}
		if exclude.Exists {
			return game.ObjectControlledGroupExcluding(anchor.Val, selection, exclude.Val), true
		}
		return game.ObjectControlledGroup(anchor.Val, selection), true
	case game.GroupDomainPlayerControlled:
		player, ok := group.PlayerAnchor()
		if !ok {
			return game.GroupReference{}, false
		}
		player, ok = transformPlayerReference(player, transform)
		if !ok {
			return game.GroupReference{}, false
		}
		if exclude.Exists {
			return game.PlayerControlledGroupExcluding(player, selection, exclude.Val), true
		}
		return game.PlayerControlledGroup(player, selection), true
	case game.GroupDomainSameName:
		if !anchor.Exists {
			return game.GroupReference{}, false
		}
		return game.SameNamePermanentGroup(anchor.Val, selection), true
	default:
		return game.GroupReference{}, false
	}
}

func contextForEffect(
	ctx contentCtx,
	effect *compiler.CompiledEffect,
) contentCtx {
	ctx.text = effect.Text
	ctx.span = effect.Span
	ctx.sequenceClause = true
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

// referencesOutsideSpan returns the references whose source span is not covered
// by span, the complement of referencesWithinSpan.
func referencesOutsideSpan(references []compiler.CompiledReference, span shared.Span) []compiler.CompiledReference {
	var outside []compiler.CompiledReference
	for _, reference := range references {
		if !spanCovered(reference.Span, []shared.Span{span}) {
			outside = append(outside, reference)
		}
	}
	return outside
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

// unsupportedEffectSequenceDiagnostic reports that an ordered effect sequence
// could not be lowered. The category distinguishes the specific blocker so the
// support report can break the otherwise-opaque reason into actionable
// sub-categories: "sub-effect — <inner reason>" when one clause needs
// single-effect support not yet available, or "structural — <reason>" when a
// sequence-machinery limitation rejects an otherwise-supported sequence.
func unsupportedEffectSequenceDiagnostic(ctx contentCtx, category string) *shared.Diagnostic {
	return contentDiagnostic(
		ctx,
		"unsupported ordered effect sequence",
		category,
	)
}

// sequenceClauseCategory names the blocker when an ordered-sequence clause could
// not be lowered: the inner sub-effect reason when the clause itself is
// unsupported, otherwise the structural shape limitation.
func sequenceClauseCategory(diagnostic *shared.Diagnostic) string {
	if diagnostic != nil {
		return "sub-effect — " + diagnostic.Summary
	}
	return "structural — clause produced modal/shared/multi-mode content"
}

// appendClauseReason records one failing ordered-sequence clause. It wraps the
// clause's blocker as an ordered-sequence category (preserving the exact primary
// reason a first-failure bail used to return) and carries forward any Additional
// reasons the clause itself collected (e.g. a modal clause blocked on several
// modes), so completeness composes across nested fan-out.
func appendClauseReason(reasons []shared.Diagnostic, ctx contentCtx, clause *shared.Diagnostic) []shared.Diagnostic {
	reasons = append(reasons, *unsupportedEffectSequenceDiagnostic(ctx, sequenceClauseCategory(clause)))
	if clause != nil {
		reasons = append(reasons, clause.Additional...)
	}
	return reasons
}

// combineReasons folds a non-empty list of blocker reasons into a single primary
// diagnostic carrying the rest as Additional. The first reason stays primary so
// the card's headline blocker is exactly the one a first-failure bail reported;
// the remaining distinct reasons ride along so the report lists every blocker.
func combineReasons(reasons []shared.Diagnostic) *shared.Diagnostic {
	deduped := dedupeReasons(reasons)
	primary := deduped[0]
	primary.Additional = deduped[1:]
	return &primary
}

// dedupeReasons drops repeated blocker reasons (same summary and detail), keeping
// first-seen order, so a construct blocked identically on several sub-parts reports
// that blocker once.
func dedupeReasons(reasons []shared.Diagnostic) []shared.Diagnostic {
	deduped := make([]shared.Diagnostic, 0, len(reasons))
	for _, reason := range reasons {
		seen := false
		for _, existing := range deduped {
			if existing.Summary == reason.Summary && existing.Detail == reason.Detail {
				seen = true
				break
			}
		}
		if !seen {
			deduped = append(deduped, reason)
		}
	}
	return deduped
}
