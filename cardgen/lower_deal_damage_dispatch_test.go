package cardgen

import (
	"testing"

	"github.com/natefinch/council4/cardgen/oracle/compiler"
	"github.com/natefinch/council4/cardgen/oracle/parser"
)

// The deal-damage dispatch assertions encode invariants lowerDealDamageSpell
// guarantees. They must panic when violated (an internal dispatch bug) rather than
// silently letting a would-be bug masquerade as an unsupported card.

func damageDispatchCtx(divided bool, reference parser.DamageRecipientReferenceKind) contentCtx {
	effect := compiler.CompiledEffect{Kind: compiler.EffectDealDamage, Divided: divided}
	effect.DamageRecipient.Reference = reference
	return contentCtx{content: compiler.AbilityContent{Effects: []compiler.CompiledEffect{effect}}}
}

func assertPanics(t *testing.T, name string, fn func()) {
	t.Helper()
	defer func() {
		if recover() == nil {
			t.Fatalf("%s: expected panic, got none", name)
		}
	}()
	fn()
}

func TestAssertDealDamageDispatch(t *testing.T) {
	// The dispatched-for shapes do not panic.
	assertDealDamageDispatch(damageDispatchCtx(false, parser.DamageRecipientReferenceYou), false)
	assertDealDamageDispatch(damageDispatchCtx(true, parser.DamageRecipientReferenceNone), true)

	assertPanics(t, "wrong effect count", func() {
		assertDealDamageDispatch(contentCtx{content: compiler.AbilityContent{}}, false)
	})
	assertPanics(t, "wrong kind", func() {
		ctx := damageDispatchCtx(false, parser.DamageRecipientReferenceYou)
		ctx.content.Effects[0].Kind = compiler.EffectDraw
		assertDealDamageDispatch(ctx, false)
	})
	assertPanics(t, "divided mismatch (want undivided)", func() {
		assertDealDamageDispatch(damageDispatchCtx(true, parser.DamageRecipientReferenceNone), false)
	})
	assertPanics(t, "divided mismatch (want divided)", func() {
		assertDealDamageDispatch(damageDispatchCtx(false, parser.DamageRecipientReferenceNone), true)
	})
}

func TestAssertUndividedRecipientDamageDispatch(t *testing.T) {
	assertUndividedRecipientDamageDispatch(
		damageDispatchCtx(false, parser.DamageRecipientReferenceYou),
		parser.DamageRecipientReferenceYou,
	)

	assertPanics(t, "divided", func() {
		assertUndividedRecipientDamageDispatch(
			damageDispatchCtx(true, parser.DamageRecipientReferenceYou),
			parser.DamageRecipientReferenceYou,
		)
	})
	assertPanics(t, "wrong recipient", func() {
		assertUndividedRecipientDamageDispatch(
			damageDispatchCtx(false, parser.DamageRecipientReferenceThatPlayer),
			parser.DamageRecipientReferenceYou,
		)
	})
}
