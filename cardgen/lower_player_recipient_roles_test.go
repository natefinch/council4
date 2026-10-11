package cardgen

import (
	"testing"

	"github.com/natefinch/council4/cardgen/oracle/compiler"
	"github.com/natefinch/council4/cardgen/oracle/parser"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/zone"
)

func TestSharedPlayerRecipientTargetRoleProof(t *testing.T) {
	t.Parallel()
	target := compiler.CompiledTarget{
		Exact: true, Cardinality: compiler.TargetCardinality{Min: 1, Max: 1},
		Selector: compiler.CompiledSelector{Kind: compiler.SelectorOpponent},
	}
	reference := compiler.CompiledReference{Kind: compiler.ReferencePronoun, Pronoun: compiler.ReferencePronounThey,
		Binding: compiler.ReferenceBindingTarget, Occurrence: 2}
	ctx := contentCtx{content: compiler.AbilityContent{
		Effects: []compiler.CompiledEffect{{Context: parser.EffectContextEventPlayer}},
		Targets: []compiler.CompiledTarget{target},
	}}
	recipient, ok := resolvePlayerRecipient(ctx, []compiler.CompiledReference{reference})
	if !ok || recipient.player != game.TargetPlayerReference(2) || len(recipient.targets) != 0 {
		t.Fatalf("target actor=%+v ok=%v, want existing player slot 2, not an event player/new target", recipient, ok)
	}
	for _, tt := range []struct {
		name string
		edit func(*contentCtx, *compiler.CompiledReference)
	}{
		{"missing target domain", func(c *contentCtx, _ *compiler.CompiledReference) { c.content.Targets = nil }},
		{"permanent subject", func(c *contentCtx, _ *compiler.CompiledReference) { c.content.Targets[0] = targetCreatureFixture() }},
		{"card subject", func(c *contentCtx, _ *compiler.CompiledReference) {
			c.content.Targets[0].Selector = compiler.CompiledSelector{Kind: compiler.SelectorCard, Zone: zone.Graveyard}
		}},
		{"plural players", func(c *contentCtx, _ *compiler.CompiledReference) { c.content.Targets[0].Cardinality.Max = 2 }},
		{"optional player", func(c *contentCtx, _ *compiler.CompiledReference) { c.content.Targets[0].Cardinality.Min = 0 }},
		{"nonexact target", func(c *contentCtx, _ *compiler.CompiledReference) { c.content.Targets[0].Exact = false }},
		{"negative occurrence", func(_ *contentCtx, r *compiler.CompiledReference) { r.Occurrence = -1 }},
		{"foreign result binding", func(_ *contentCtx, r *compiler.CompiledReference) {
			r.Binding = compiler.ReferenceBindingPriorInstructionResult
		}},
		{"foreign object pronoun", func(_ *contentCtx, r *compiler.CompiledReference) { r.Pronoun = compiler.ReferencePronounIt }},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			mutated := ctx
			mutated.content.Targets = append([]compiler.CompiledTarget(nil), ctx.content.Targets...)
			ref := reference
			tt.edit(&mutated, &ref)
			if recipient, ok := resolvePlayerRecipient(mutated, []compiler.CompiledReference{ref}); ok {
				t.Fatalf("unsupported PLAYER-domain proof accepted: %+v", recipient)
			}
		})
	}
}

func TestSharedPlayerRecipientLifeZoneEventBoundary(t *testing.T) {
	t.Parallel()
	for _, tt := range []struct {
		name, text string
		kind       game.EventKind
		binding    compiler.ReferenceBinding
		supported  bool
	}{
		{"life zone player", "You gain 2 life.", game.EventZoneChanged, compiler.ReferenceBindingEventPermanent, true},
		{"life wrong trigger", "You gain 2 life.", game.EventSpellCast, compiler.ReferenceBindingEventPermanent, false},
		{"life unknown trigger", "You gain 2 life.", game.EventUnknown, compiler.ReferenceBindingEventPermanent, false},
		{"life foreign source", "You gain 2 life.", game.EventZoneChanged, compiler.ReferenceBindingSource, false},
		{"draw is not a life projection", "Draw a card.", game.EventZoneChanged, compiler.ReferenceBindingEventPermanent, false},
		{"mill is not a life projection", "Mill two cards.", game.EventZoneChanged, compiler.ReferenceBindingEventPermanent, false},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			compilation, diagnostics := compileTestOracle(tt.text, parser.Context{InstantOrSorcery: true}, compiler.Context{})
			if len(diagnostics) != 0 {
				t.Fatal(diagnostics)
			}
			content := compilation.Abilities[0].Content
			content.Effects[0].Context = parser.EffectContextReferencedPlayer
			content.References = []compiler.CompiledReference{{Kind: compiler.ReferenceThatPlayer, Binding: tt.binding}}
			ctx := contentCtx{content: content, enclosingKind: compiler.AbilityTriggered, triggerEvent: tt.kind}
			lowered, diagnostic := lowerImmediateSingleEffectSpell("Opaque", ctx, &parser.Ability{})
			if (diagnostic == nil) != tt.supported {
				t.Fatalf("supported=%v diagnostic=%+v", tt.supported, diagnostic)
			}
			if tt.supported {
				primitive, ok := lowered.Modes[0].Sequence[0].Primitive.(game.GainLife)
				if !ok || primitive.Player != game.EventPlayerReference() {
					t.Fatalf("life zone actor=%+v", lowered)
				}
			}
		})
	}
}
