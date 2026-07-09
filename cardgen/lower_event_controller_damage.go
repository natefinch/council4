package cardgen

import (
	"fmt"

	"github.com/natefinch/council4/cardgen/oracle/compiler"
	"github.com/natefinch/council4/cardgen/oracle/parser"
	"github.com/natefinch/council4/mtg/game"
)

// lowerEventPermanentControllerDamageSpell lowers a triggered "deals N damage to
// that <object>'s controller/owner" effect whose recipient is the player who
// controls (or owns) the permanent that fired the trigger, as in "Whenever a
// land enters, this artifact deals 2 damage to that land's controller." (Ankh of
// Mishra), "Whenever a creature dies, this artifact deals 2 damage to that
// creature's controller." (Dingus Staff), "Whenever a creature blocks, this
// enchantment deals 1 damage to that creature's controller." (Heat of Battle),
// or "Whenever a goaded creature attacks, it deals 1 damage to its controller."
// (Vengeful Ancestor).
//
// The recipient reference is the possessive "that <noun>'s"/"its" object whose
// binding resolves to the triggering event's permanent
// (ReferenceBindingEventPermanent); the damage recipient is the controller or
// owner of that permanent. The damage subject (the source) is the ability's own
// permanent ("this artifact" or the card's own name), a bare "it" bound to the
// source, or a bare "it" bound to that same triggering event permanent (Vengeful
// Ancestor, where the goaded attacker both fires the trigger and deals the
// damage). The amount is a fixed value (>= 1) or X.
//
// It returns ok=false for every shape outside that template — a recipient that is
// not the event-bound controller/owner, a recipient that is the resolving source
// or a chosen target, a source-power or other dynamic amount, any rider, target,
// recipient selector, condition, keyword, or mode — leaving the standard damage
// paths to handle their own effects and emit their own diagnostics.
func lowerEventPermanentControllerDamageSpell(ctx contentCtx) (game.AbilityContent, bool) {
	assertDealDamageDispatch(ctx, false)
	effect := ctx.content.Effects[0]
	player, ok := eventPermanentControllerOwnerRecipient(effect.DamageRecipient.Reference, ctx.content.References)
	if !ok {
		return game.AbilityContent{}, false
	}
	if !effect.Exact ||
		effect.Negated ||
		len(effect.DamageRiders) != 0 ||
		len(effect.DamageRecipient.GroupSelectors) != 0 ||
		len(ctx.content.Targets) != 0 ||
		len(ctx.content.Conditions) != 0 ||
		len(abilityKeywordsExcludingSelectorPredicates(ctx.content)) != 0 ||
		len(ctx.content.Modes) != 0 {
		return game.AbilityContent{}, false
	}
	amount, ok := eventPermanentControllerDamageAmount(effect)
	if !ok {
		return game.AbilityContent{}, false
	}
	sourceReferences := damageSourceReferencesExcludingRecipient(ctx.content.References)
	if len(sourceReferences) != 1 || !exactDamageSourceSyntax(sourceReferences) {
		return game.AbilityContent{}, false
	}
	damage := game.Damage{
		Amount:       amount,
		Recipient:    game.PlayerDamageRecipient(player),
		DamageSource: primaryDamageSource(sourceReferences),
	}
	return game.Mode{
		Sequence: []game.Instruction{{Primitive: damage}},
	}.Ability(), true
}

// eventPermanentControllerOwnerRecipient resolves the player recipient of a
// "deals N damage to that <object>'s controller/owner" effect when the referenced
// object is the triggering event's permanent. It returns the controller or owner
// of EventPermanentReference() per the recipient role, requiring exactly one
// event-permanent recipient reference. It fails closed (ok=false) for a role that
// is not controller or owner and for any reference set that does not carry a sole
// event-permanent recipient reference.
func eventPermanentControllerOwnerRecipient(
	role parser.DamageRecipientReferenceKind,
	references []compiler.CompiledReference,
) (game.PlayerReference, bool) {
	if role != parser.DamageRecipientReferenceController &&
		role != parser.DamageRecipientReferenceOwner {
		return game.PlayerReference{}, false
	}
	recipient, ok := soleEventPermanentRecipientReference(references)
	if !ok {
		return game.PlayerReference{}, false
	}
	object, ok := lowerObjectReference(recipient, referenceLoweringContext{AllowEvent: true})
	if !ok {
		return game.PlayerReference{}, false
	}
	switch role {
	case parser.DamageRecipientReferenceController:
		return game.ObjectControllerReference(object), true
	case parser.DamageRecipientReferenceOwner:
		return game.ObjectOwnerReference(object), true
	default:
		// Unreachable: the guard above already fail-closed any role other than
		// controller or owner, so reaching here is an internal bug.
		panic(fmt.Sprintf("eventPermanentControllerOwnerRecipient: role %v passed the controller/owner guard but is neither", role))
	}
}

// soleEventPermanentRecipientReference returns the lone "that <noun>'s"/"its"
// reference bound to the triggering event's permanent, the recipient antecedent
// of "deals N damage to that creature's controller." It fails closed unless
// exactly one such reference exists, so a clause carrying a second event-bound
// referent (a source-power amount referent) is not mistaken for the recipient.
func soleEventPermanentRecipientReference(
	references []compiler.CompiledReference,
) (compiler.CompiledReference, bool) {
	var recipient compiler.CompiledReference
	found := false
	for _, reference := range references {
		if !eventPermanentRecipientReference(reference) {
			continue
		}
		if found {
			return compiler.CompiledReference{}, false
		}
		recipient = reference
		found = true
	}
	return recipient, found
}

// eventPermanentRecipientReference reports whether reference is the possessive
// "that <noun>'s"/"its" antecedent bound to the triggering event's permanent that
// names a damage recipient's controller or owner. A controller/owner recipient
// antecedent is always possessive, so a bare "it" (ReferencePronounIt) is not
// matched: that pronoun names the damage source ("it deals N damage to its
// controller", Vengeful Ancestor) and is retained by
// damageSourceReferencesExcludingRecipient.
func eventPermanentRecipientReference(reference compiler.CompiledReference) bool {
	if reference.Binding != compiler.ReferenceBindingEventPermanent {
		return false
	}
	switch reference.Kind {
	case compiler.ReferenceThatObject:
		return true
	case compiler.ReferencePronoun:
		return reference.Pronoun == compiler.ReferencePronounIts
	default:
		return false
	}
}

// damageSourceReferencesExcludingRecipient returns the references that describe
// the damage subject, dropping the event-permanent recipient antecedent so the
// damage-source exactness check sees only the subject reference.
func damageSourceReferencesExcludingRecipient(
	references []compiler.CompiledReference,
) []compiler.CompiledReference {
	sources := make([]compiler.CompiledReference, 0, len(references))
	for _, reference := range references {
		if eventPermanentRecipientReference(reference) {
			continue
		}
		sources = append(sources, reference)
	}
	return sources
}

// eventPermanentControllerDamageAmount resolves the quantity a controller/owner
// damage effect deals. It accepts only a fixed value (>= 1) or bare X, failing
// closed for a non-positive fixed value or any dynamic amount so the recipient
// antecedent is never confused with a source-power amount referent.
func eventPermanentControllerDamageAmount(effect compiler.CompiledEffect) (game.Quantity, bool) {
	if effect.Amount.DynamicKind != compiler.DynamicAmountNone {
		return game.Quantity{}, false
	}
	if effect.Amount.Known {
		if effect.Amount.Value < 1 {
			return game.Quantity{}, false
		}
		return game.Fixed(effect.Amount.Value), true
	}
	if !effect.Amount.VariableX {
		return game.Quantity{}, false
	}
	return game.Dynamic(game.DynamicAmount{Kind: game.DynamicAmountX}), true
}
