package parser

import (
	"fmt"
	"strings"
)

// exactDualReferencedCounterPlacementEffectSyntax recognizes the counter
// placement that names two distinct referenced permanents — the triggering
// event permanent and the source itself — placing the same counter on each
// ("put a +1/+1 counter on that creature and a +1/+1 counter on this creature."
// — Juniper Order Ranger; "put a +1/+1 counter on that creature and a +1/+1
// counter on X-23." — X-23, Deadly Weapon). The parser models both recipients as
// one EffectPut carrying a single counter kind and amount with two references, so
// the same kind and count apply to each.
//
// It requires a single recognized counter kind, a fixed positive amount, no
// targets, and exactly one event demonstrative ("that creature"/"it") plus one
// self reference ("this <object>" or the card's own name), then reconstructs the
// canonical "Put <amount> <kind> counter(s) on <obj1> and <amount> <kind>
// counter(s) on <obj2>." clause and compares it byte-for-byte to the source.
// Every other shape — a kind choice, an attached recipient, a dynamic amount, a
// negated or optional effect, two self or two event references — fails the
// round-trip and stays unsupported.
func exactDualReferencedCounterPlacementEffectSyntax(effect *EffectSyntax) bool {
	if !effect.CounterKnown ||
		effect.Negated ||
		effect.Optional ||
		effect.CounterRecipientAttached ||
		len(effect.CounterKindChoices) != 0 ||
		len(effect.Targets) != 0 ||
		len(effect.References) != 2 {
		return false
	}
	if !effect.Amount.Known ||
		effect.Amount.Value < 1 ||
		effect.Amount.DynamicForm != EffectDynamicAmountFormNone ||
		effect.Amount.DynamicKind != EffectDynamicAmountNone {
		return false
	}
	objects, ok := dualCounterPlacementObjects(effect.References)
	if !ok {
		return false
	}
	noun := "counters"
	if effect.Amount.Value == 1 {
		noun = "counter"
	}
	amount := effectAmountSourceText(effect)
	kind := effect.CounterKind.String()
	half := func(object string) string {
		return fmt.Sprintf("%s %s %s on %s", amount, kind, noun, object)
	}
	expected := fmt.Sprintf("Put %s and %s.", half(objects[0]), half(objects[1]))
	return strings.EqualFold(exactEffectClauseText(effect), expected)
}

// dualCounterPlacementObjects reconstructs the two recipient phrases of a dual
// referenced counter placement, in source order. It requires exactly one event
// demonstrative reference ("that creature"/"it") and one self reference ("this
// <object>" or the card's own name); any other pairing, including two event or
// two self references, fails closed.
func dualCounterPlacementObjects(references []Reference) ([2]string, bool) {
	if len(references) != 2 {
		return [2]string{}, false
	}
	ordered := []Reference{references[0], references[1]}
	if ordered[1].Span.Start.Offset < ordered[0].Span.Start.Offset {
		ordered[0], ordered[1] = ordered[1], ordered[0]
	}
	var events, selves int
	var objects [2]string
	for index, reference := range ordered {
		switch reference.Kind {
		case ReferenceThatObject:
			events++
		case ReferencePronoun:
			if reference.Pronoun != PronounIt {
				return [2]string{}, false
			}
			events++
		case ReferenceThisObject, ReferenceSelfName:
			selves++
		default:
			return [2]string{}, false
		}
		objects[index] = joinedEffectText(reference.Tokens)
	}
	if events != 1 || selves != 1 {
		return [2]string{}, false
	}
	return objects, true
}
