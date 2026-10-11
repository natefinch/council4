package cardgen

import (
	"testing"

	"github.com/natefinch/council4/cardgen/oracle/compiler"
	"github.com/natefinch/council4/cardgen/oracle/parser"
)

func TestHalfPlayerAmountOwnedCardDefs(t *testing.T) {
	for _, text := range []string{
		"Target opponent loses half their life, rounded up.",
		"Target opponent loses half their life, rounded down.",
		"Target opponent loses 2 life. Scry 1. That player loses half their life, rounded up.",
		"Target opponent mills half their library, rounded up.",
		"Target opponent loses 2 life. Scry 1. That player mills half their library, rounded down.",
	} {
		card := &ScryfallCard{Name: "Half Player Amount", Layout: "normal", TypeLine: "Instant", OracleText: text}
		paths := cardDefPaths(t, card)
		if len(paths) == 0 {
			t.Fatal("half-player capability emitted no CardDef paths")
		}
	}
	for _, text := range []string{
		"Each opponent loses half their life, rounded up.",
		"Each opponent mills half their library, rounded up.",
		"Target opponent loses half its life, rounded up.",
		"Target opponent mills half its library, rounded up.",
		"Target opponent loses half your life, rounded up.",
		"Target opponent mills half your library, rounded up.",
		"Target opponent loses half their life.",
	} {
		assertCardUnsupported(t, &ScryfallCard{Name: "Half Player Near Miss", Layout: "normal", TypeLine: "Instant", OracleText: text})
	}
}

func TestHalfPlayerAmountProofRejectsForeignOwnership(t *testing.T) {
	compiled, diagnostics := compileTestOracle("Target opponent loses half their life, rounded up.",
		parser.Context{InstantOrSorcery: true}, compiler.Context{})
	if len(diagnostics) != 0 {
		t.Fatal(diagnostics)
	}
	content := compiled.Abilities[0].Content
	effect, reference, target := content.Effects[0], content.References[0], content.Targets[0]
	if reference.NodeID != 0 || effect.Amount.ReferenceNodeID != 0 ||
		!reference.HalfPlayerAmountRecipient(effect, &target) {
		t.Fatal("canonical first-reference NodeID zero did not receive an owned same-player amount proof")
	}
	for _, tt := range []struct {
		name string
		edit func(*compiler.CompiledReference, *compiler.CompiledEffect, *compiler.CompiledTarget)
	}{
		{"default reference", func(r *compiler.CompiledReference, _ *compiler.CompiledEffect, _ *compiler.CompiledTarget) {
			*r = compiler.CompiledReference{}
		}},
		{"foreign amount node", func(_ *compiler.CompiledReference, e *compiler.CompiledEffect, _ *compiler.CompiledTarget) {
			e.Amount.ReferenceNodeID++
		}},
		{"foreign reference node", func(r *compiler.CompiledReference, _ *compiler.CompiledEffect, _ *compiler.CompiledTarget) {
			r.NodeID++
		}},
		{"foreign order", func(r *compiler.CompiledReference, _ *compiler.CompiledEffect, _ *compiler.CompiledTarget) {
			r.Order.Start++
		}},
		{"foreign clause", func(_ *compiler.CompiledReference, e *compiler.CompiledEffect, _ *compiler.CompiledTarget) {
			e.ClauseID++
		}},
		{"foreign event binding", func(r *compiler.CompiledReference, _ *compiler.CompiledEffect, _ *compiler.CompiledTarget) {
			r.Binding = compiler.ReferenceBindingEventPlayer
		}},
		{"foreign cost binding", func(r *compiler.CompiledReference, _ *compiler.CompiledEffect, _ *compiler.CompiledTarget) {
			r.Binding = compiler.ReferenceBindingPaidCost
		}},
		{"foreign card target", func(_ *compiler.CompiledReference, _ *compiler.CompiledEffect, t *compiler.CompiledTarget) {
			t.Selector.Kind = compiler.SelectorCard
		}},
		{"foreign occurrence", func(r *compiler.CompiledReference, _ *compiler.CompiledEffect, _ *compiler.CompiledTarget) {
			r.Occurrence++
		}},
		{"plural targets", func(_ *compiler.CompiledReference, _ *compiler.CompiledEffect, t *compiler.CompiledTarget) {
			t.Cardinality.Max = 2
		}},
		{"changed rounding", func(_ *compiler.CompiledReference, e *compiler.CompiledEffect, _ *compiler.CompiledTarget) {
			e.Amount.RoundUp = false
		}},
	} {
		t.Run(tt.name, func(t *testing.T) {
			r, e, target := reference, effect, target
			tt.edit(&r, &e, &target)
			if r.HalfPlayerAmountRecipient(e, &target) {
				t.Fatal("foreign/default amount ownership was accepted")
			}
		})
	}
	remaining := halfPlayerAmountRecipientReferences(contentCtx{content: content}, effect)
	if len(remaining) != 0 {
		t.Fatal("proved amount operand was not consumed")
	}
	content.References = append(content.References, compiler.CompiledReference{Binding: compiler.ReferenceBindingEventPermanent})
	remaining = halfPlayerAmountRecipientReferences(contentCtx{content: content}, effect)
	if len(remaining) != 1 {
		t.Fatal("unknown extra reference was consumed along with the owned amount")
	}
	if _, ok := resolvePlayerRecipient(contentCtx{content: content}, remaining); ok {
		t.Fatal("unconsumed foreign reference was permitted as a player actor")
	}
}

func TestHalfPlayerAmountTargetCardDefFields(t *testing.T) {
	card := &ScryfallCard{Name: "Half Life Fields", Layout: "normal", TypeLine: "Instant",
		OracleText: "Target opponent loses half their life, rounded up."}
	assertCardPaths(t, card,
		"Sequence[0].Primitive.(game.LoseLife).Player.kind = game.PlayerReferenceTargetPlayer",
		"Sequence[0].Primitive.(game.LoseLife).Amount.dynamic.Kind = game.DynamicAmountPlayerLife",
		"Sequence[0].Primitive.(game.LoseLife).Amount.dynamic.Divisor = 2",
		"Sequence[0].Primitive.(game.LoseLife).Amount.dynamic.RoundUp = true",
		"Sequence[0].Primitive.(game.LoseLife).Amount.dynamic.Player.kind = game.PlayerReferenceTargetPlayer")
	assertCardPathsAbsent(t, card, "Player.kind = game.PlayerReferenceEventPlayer")
}
