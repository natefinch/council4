package cardgen

import (
	"testing"

	"github.com/natefinch/council4/cardgen/oracle/compiler"
	"github.com/natefinch/council4/cardgen/oracle/parser"
	"github.com/natefinch/council4/cardgen/oracle/shared"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/opt"
)

func TestOptionalEnteredObjectReferenceUsesTypedOccurrence(t *testing.T) {
	document, diagnostics := parser.Parse(
		"You may return target creature card from your graveyard to the battlefield and gain 1 life. It gains haste until end of turn.",
		parser.Context{InstantOrSorcery: true})
	if len(diagnostics) != 0 {
		t.Fatal(diagnostics)
	}
	compilation, diagnostics := compiler.Compile(document, compiler.Context{})
	if len(diagnostics) != 0 {
		t.Fatal(diagnostics)
	}
	content := compilation.Abilities[0].Content
	content.Targets[0].Span = shared.Span{}
	for i := range content.Effects {
		content.Effects[i].Text = "opaque"
		content.Effects[i].Span = shared.Span{}
		content.Effects[i].ClauseSpan = shared.Span{}
	}
	for _, mapped := range []bool{true, false} {
		t.Run(map[bool]string{true: "opaque provenance", false: "missing target identity"}[mapped], func(t *testing.T) {
			ctx := contextForEffect(contentCtx{content: content}, &content.Effects[2])
			ctx.content.Targets = content.Targets
			sequence := []game.Instruction{
				{Primitive: game.PutOnBattlefield{
					Source: game.CardBattlefieldSource(game.CardReference{Kind: game.CardReferenceTarget}),
				}, Optional: true, PublishOptionalDecision: "choice"},
				{Primitive: game.GainLife{Player: game.ControllerReference(), Amount: game.Fixed(1)}, OptionalDecisionGate: "choice"},
				{Primitive: game.GainLife{Player: game.ControllerReference(), Amount: game.Fixed(2)}},
			}
			targetIndices := map[shared.Span]int{}
			if mapped {
				targetIndices[content.Targets[0].Span] = 1
			}
			reason := bindOptionalEnteredObjectReference(&ctx, sequence, [][2]int{{0, 1}, {1, 3}}, targetIndices,
				[]game.TargetSpec{{Allow: game.TargetAllowPermanent}, {Allow: game.TargetAllowCard}})
			if !mapped {
				if reason == "" || game.PublishedLinkedKey(sequence[0].Primitive) != "" ||
					ctx.content.References[0].Binding != compiler.ReferenceBindingTarget {
					t.Fatal("missing target mapping was accepted or changed the original publication/reference")
				}
				return
			}
			key := sequenceProductKey(0)
			if reason != "" || ctx.priorInstruction != 0 || ctx.priorLinkedKey != key ||
				game.PublishedLinkedKey(sequence[0].Primitive) != key ||
				len(ctx.content.Targets) != 0 ||
				ctx.content.References[0].Binding != compiler.ReferenceBindingPriorInstructionResult ||
				ctx.content.References[0].PriorInstruction != 0 {
				t.Fatalf("typed earlier card occurrence did not bind the actual product: %s", reason)
			}
		})
	}
}

func TestOptionalEnteredObjectPublicationRefusals(t *testing.T) {
	document, diagnostics := parser.Parse(
		"You may return target creature card from your graveyard to the battlefield and gain 1 life. It gains haste until end of turn.",
		parser.Context{InstantOrSorcery: true})
	if len(diagnostics) != 0 {
		t.Fatal(diagnostics)
	}
	compilation, diagnostics := compiler.Compile(document, compiler.Context{})
	if len(diagnostics) != 0 {
		t.Fatal(diagnostics)
	}
	content := compilation.Abilities[0].Content
	for _, tt := range []struct {
		name string
		key  game.LinkedKey
		span [2]int
	}{
		{"competing persistent publication", "persistent", [2]int{0, 1}},
		{"expanded producer without aggregate", "", [2]int{0, 2}},
	} {
		t.Run(tt.name, func(t *testing.T) {
			ctx := contextForEffect(contentCtx{content: content}, &content.Effects[2])
			ctx.content.Targets = content.Targets
			producer := game.PutOnBattlefield{
				Source:        game.CardBattlefieldSource(game.CardReference{Kind: game.CardReferenceTarget}),
				PublishLinked: tt.key,
			}
			sequence := []game.Instruction{
				{Primitive: producer, Optional: true, PublishOptionalDecision: "choice"},
				{Primitive: game.GainLife{Player: game.ControllerReference(), Amount: game.Fixed(1)}, OptionalDecisionGate: "choice"},
			}
			reason := bindOptionalEnteredObjectReference(&ctx, sequence, [][2]int{tt.span, {1, 2}},
				map[shared.Span]int{content.Targets[0].Span: 0},
				[]game.TargetSpec{{Allow: game.TargetAllowCard, MinTargets: 1, MaxTargets: 1}})
			if reason == "" || game.PublishedLinkedKey(sequence[0].Primitive) != tt.key ||
				ctx.content.References[0].Binding != compiler.ReferenceBindingTarget {
				t.Fatal("unproven publication was accepted or an existing link was changed")
			}
		})
	}
}

func TestOptionalLinkedPublicationAvailability(t *testing.T) {
	producer := game.Instruction{
		Primitive: game.PutOnBattlefield{
			Source:        game.CardBattlefieldSource(game.CardReference{Kind: game.CardReferenceTarget}),
			PublishLinked: "entered",
		},
		Optional: true, PublishOptionalDecision: "choice",
	}
	rider := game.Instruction{Primitive: game.ApplyContinuous{
		Object: opt.Val(game.LinkedObjectReference("entered")),
	}}
	alias := rider
	apply := alias.Primitive.(game.ApplyContinuous)
	apply.PublishLinked = "shim"
	alias.Primitive = apply
	persistentProducer := producer
	persistentProducer.Optional = false
	persistentProducer.PublishOptionalDecision = ""
	gatedAlias := alias
	gatedAlias.OptionalDecisionGate = "choice"
	for _, tt := range []struct {
		name     string
		sequence []game.Instruction
		modeled  bool
	}{
		{"actual product consumer", []game.Instruction{producer, rider}, true},
		{"unproven chained publication", []game.Instruction{producer, alias}, false},
		{"unproven gated shim", []game.Instruction{producer, gatedAlias}, false},
		{"existing persistent publication untouched", []game.Instruction{persistentProducer, alias}, true},
	} {
		t.Run(tt.name, func(t *testing.T) {
			if got := optionalLinkedPublicationsModeled(tt.sequence); got != tt.modeled {
				t.Fatalf("modeled = %v, want %v", got, tt.modeled)
			}
		})
	}
}
