package cardgen

import (
	"testing"

	"github.com/natefinch/council4/cardgen/oracle/compiler"
	"github.com/natefinch/council4/cardgen/oracle/parser"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/zone"
)

func TestSinglePermanentFramePreservesVerbPolicies(t *testing.T) {
	t.Parallel()
	base := func() contentCtx {
		return contentCtx{content: compiler.AbilityContent{
			Effects: []compiler.CompiledEffect{{
				Kind: compiler.EffectDestroy, Exact: true, Context: parser.EffectContextController,
			}},
			Targets: []compiler.CompiledTarget{{
				Exact: true, Cardinality: compiler.TargetCardinality{Min: 1, Max: 1},
				Selector: compiler.CompiledSelector{Kind: compiler.SelectorCreature},
			}},
		}}
	}
	thatPlayer := func(ctx *contentCtx, event game.EventKind, binding compiler.ReferenceBinding) {
		ctx.triggerEvent = event
		ctx.content.Targets[0].Selector.Controller = compiler.ControllerThatPlayer
		ctx.content.References = []compiler.CompiledReference{{Kind: compiler.ReferenceThatPlayer, Binding: binding}}
	}
	for _, tc := range []struct {
		name      string
		mutate    func(*contentCtx)
		destroy   bool
		fixed     bool
		defending bool
	}{
		{name: "plain", destroy: true, fixed: true},
		{name: "inexact action", mutate: func(ctx *contentCtx) { ctx.content.Effects[0].Exact = false }},
		{name: "negated action", mutate: func(ctx *contentCtx) { ctx.content.Effects[0].Negated = true }},
		{name: "opponent actor", mutate: func(ctx *contentCtx) { ctx.content.Effects[0].Context = parser.EffectContextEachOpponent }},
		{name: "inexact target", mutate: func(ctx *contentCtx) { ctx.content.Targets[0].Exact = false }},
		{name: "no target", mutate: func(ctx *contentCtx) { ctx.content.Targets = nil }},
		{name: "unconsumed keyword", mutate: func(ctx *contentCtx) { ctx.content.Keywords = []compiler.CompiledKeyword{{}} }},
		{name: "unconsumed mode", mutate: func(ctx *contentCtx) { ctx.content.Modes = []compiler.CompiledMode{{}} }},
		{name: "unconsumed condition", mutate: func(ctx *contentCtx) { ctx.content.Conditions = []compiler.CompiledCondition{{}} }},
		{name: "raw effect optional remains destroy-owned", mutate: func(ctx *contentCtx) { ctx.content.Effects[0].Optional = true }, destroy: true},
		{name: "raw body optional remains destroy-owned", mutate: func(ctx *contentCtx) { ctx.optional = true }, destroy: true},
		{name: "raw extra action remains caller-owned", mutate: func(ctx *contentCtx) {
			ctx.content.Effects = append(ctx.content.Effects, ctx.content.Effects[0])
		}, destroy: true},
		{name: "redundant target reference only generic", mutate: func(ctx *contentCtx) {
			ctx.content.References = []compiler.CompiledReference{{Kind: compiler.ReferenceThatObject, Binding: compiler.ReferenceBindingTarget}}
		}, fixed: true},
		{name: "wrong target occurrence", mutate: func(ctx *contentCtx) {
			ctx.content.References = []compiler.CompiledReference{{Kind: compiler.ReferenceThatObject, Binding: compiler.ReferenceBindingTarget, Occurrence: 1}}
		}},
		{name: "event player", mutate: func(ctx *contentCtx) {
			thatPlayer(ctx, game.EventLifeGained, compiler.ReferenceBindingEventPlayer)
		}, destroy: true, fixed: true},
		{name: "unknown event player", mutate: func(ctx *contentCtx) {
			thatPlayer(ctx, game.EventUnknown, compiler.ReferenceBindingEventPlayer)
		}, fixed: true},
		{name: "event player wrong binding", mutate: func(ctx *contentCtx) {
			thatPlayer(ctx, game.EventLifeGained, compiler.ReferenceBindingTarget)
		}, fixed: true},
		{name: "attack defender", mutate: func(ctx *contentCtx) {
			thatPlayer(ctx, game.EventAttackerDeclared, compiler.ReferenceBindingAmbiguous)
		}, destroy: true, fixed: true, defending: true},
		{name: "extra defender reference", mutate: func(ctx *contentCtx) {
			thatPlayer(ctx, game.EventAttackerDeclared, compiler.ReferenceBindingEventPlayer)
			ctx.content.References = append(ctx.content.References, ctx.content.References[0])
		}, fixed: true},
		{name: "event player without controller qualifier", mutate: func(ctx *contentCtx) {
			ctx.triggerEvent = game.EventLifeGained
			ctx.content.References = []compiler.CompiledReference{{Kind: compiler.ReferenceThatPlayer, Binding: compiler.ReferenceBindingEventPlayer}}
		}},
		{name: "bare that-player controller", mutate: func(ctx *contentCtx) {
			ctx.content.Targets[0].Selector.Controller = compiler.ControllerThatPlayer
		}, destroy: true, fixed: true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			ctx := base()
			if tc.mutate != nil {
				tc.mutate(&ctx)
			}
			destroy, diagnostic := lowerFixedDestroySpell(ctx)
			if (diagnostic == nil) != tc.destroy {
				t.Fatalf("destroy accepted=%v, want %v: %+v", diagnostic == nil, tc.destroy, diagnostic)
			}
			fixed, diagnostic := lowerFixedPermanentTargetSpell(ctx, "Exile", func(object game.ObjectReference) game.Primitive {
				return game.MovePermanent{Object: object, Destination: zone.Exile}
			}, true)
			if (diagnostic == nil) != tc.fixed {
				t.Fatalf("generic accepted=%v, want %v: %+v", diagnostic == nil, tc.fixed, diagnostic)
			}
			for _, accepted := range []struct {
				content game.AbilityContent
				ok      bool
			}{{destroy, tc.destroy}, {fixed, tc.fixed}} {
				if !accepted.ok {
					continue
				}
				if len(accepted.content.Modes) != 1 ||
					len(accepted.content.Modes[0].Targets) != 1 ||
					len(accepted.content.Modes[0].Sequence) != 1 {
					t.Fatalf("not a singleton permanent frame: %+v", accepted.content)
				}
				// Raw adapters never own the outer "may" envelope.
				if accepted.content.Modes[0].Sequence[0].Optional {
					t.Fatal("raw adapter applied an optional envelope")
				}
			}
			if tc.destroy {
				selection := destroy.Modes[0].Targets[0].Selection.Val
				if selection.ControlledByDefendingPlayer != tc.defending {
					t.Fatalf("defender predicate=%v, want %v", selection.ControlledByDefendingPlayer, tc.defending)
				}
				wantEventPlayer := ctx.content.Targets[0].Selector.Controller == compiler.ControllerThatPlayer && !tc.defending
				if selection.ControlledByEventPlayer != wantEventPlayer {
					t.Fatalf("event-player predicate=%v, want %v", selection.ControlledByEventPlayer, wantEventPlayer)
				}
			}
		})
	}
}

func TestSinglePermanentFrameCompiledCapabilities(t *testing.T) {
	t.Parallel()
	for _, tc := range []struct {
		body string
		want []string
	}{
		{"Destroy target tapped nonland permanent an opponent controls.", []string{
			"Modes[0].Targets[0].Selection.Val.Tapped = game.TriTrue",
			"Modes[0].Targets[0].Selection.Val.ExcludedTypes[0] = types.Land",
			"Modes[0].Targets[0].Selection.Val.Controller = game.ControllerOpponent",
			"Modes[0].Sequence[0].Primitive.(game.Destroy).Object.kind = game.ObjectReferenceTargetPermanent",
		}},
		{"Destroy target creature if it's blue.", []string{
			"Modes[0].Sequence[0].Primitive.(game.Destroy)",
			"Modes[0].Sequence[0].Condition.Val.Condition.Val.ObjectMatches.Val.ColorsAny[0] = color.Blue",
			"Modes[0].Sequence[0].Condition.Val.Object.kind = game.ObjectReferenceTargetPermanent",
		}},
		{"You may destroy target creature.", []string{
			"Modes[0].Sequence[0].Primitive.(game.Destroy)",
			"Modes[0].Sequence[0].Optional = true",
		}},
		{"Destroy target creature. It can't be regenerated.", []string{
			"Modes[0].Sequence[0].Primitive.(game.Destroy).PreventRegeneration = true",
		}},
		{"Destroy target nonland permanent and all other permanents with the same name as that permanent.", []string{
			"Modes[0].Sequence[0].Primitive.(game.Destroy).Group.domain = game.GroupDomainSameName",
		}},
		{"Destroy X target artifacts.", []string{
			"Modes[0].Targets[0].CountEqualsX = true",
			"Modes[0].Sequence[0].Primitive.(game.Destroy).Object.kind = game.ObjectReferenceAllTargetPermanents",
		}},
		{"Exile target Elf.", []string{
			"Modes[0].Targets[0].Selection.Val.SubtypesAny[0] = types.Elf",
			"Modes[0].Sequence[0].Primitive.(game.MovePermanent).Object.kind = game.ObjectReferenceTargetPermanent",
		}},
		{"Tap target nonland permanent.", []string{
			"Modes[0].Targets[0].Selection.Val.ExcludedTypes[0] = types.Land",
			"Modes[0].Sequence[0].Primitive.(game.Tap).Object.kind = game.ObjectReferenceTargetPermanent",
		}},
		{"Regenerate target creature with a +1/+1 counter on it.", []string{
			"Modes[0].Targets[0].Selection.Val.MatchCounter = true",
			"Modes[0].Sequence[0].Primitive.(game.Regenerate).Object.kind = game.ObjectReferenceTargetPermanent",
		}},
	} {
		t.Run(tc.body, func(t *testing.T) {
			t.Parallel()
			card := targetCountTestCard(tc.body)
			assertCardPaths(t, card, tc.want...)
			assertCardPathsAbsent(t, card, "Modes[1]", "Modes[0].Targets[1]", "Modes[0].Sequence[1]")
		})
	}

	for _, body := range []string{
		"Destroy target creature or instant.",
		"Destroy target creature that entered this turn.",
		"Destroy target creature if it's monocolored.",
	} {
		t.Run(body, func(t *testing.T) {
			t.Parallel()
			assertCardUnsupported(t, targetCountTestCard(body), "unsupported")
		})
	}
}

func TestSinglePermanentFrameDestroyColorPolicy(t *testing.T) {
	t.Parallel()
	for _, tc := range []struct {
		name   string
		mutate func(*contentCtx)
		ok     bool
	}{
		{name: "single color gate", ok: true},
		{name: "multiple color gates", mutate: func(ctx *contentCtx) {
			ctx.content.Conditions = append(ctx.content.Conditions, ctx.content.Conditions[0])
		}},
		{name: "unavailable color selection", mutate: func(ctx *contentCtx) {
			ctx.content.Conditions[0].Selection = compiler.ConditionSelection{}
		}},
		{name: "wrong predicate", mutate: func(ctx *contentCtx) {
			ctx.content.Conditions[0].Predicate = compiler.ConditionPredicateUnsupported
		}},
		{name: "unconsumed target reference", mutate: func(ctx *contentCtx) {
			ctx.content.References = []compiler.CompiledReference{{Kind: compiler.ReferenceThatObject, Binding: compiler.ReferenceBindingTarget}}
		}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			compilation, diagnostics := compileTestOracle(
				"Destroy target creature if it's blue.",
				parser.Context{InstantOrSorcery: true}, compiler.Context{},
			)
			if len(diagnostics) != 0 {
				t.Fatalf("compile diagnostics: %+v", diagnostics)
			}
			ctx := contentCtx{content: compilation.Abilities[0].Content}
			if tc.mutate != nil {
				tc.mutate(&ctx)
			}
			content, diagnostic := lowerFixedDestroySpell(ctx)
			if (diagnostic == nil) != tc.ok {
				t.Fatalf("destroy accepted=%v, want %v: %+v", diagnostic == nil, tc.ok, diagnostic)
			}
			if tc.ok && content.Modes[0].Sequence[0].Condition.Val.Object != game.TargetPermanentReference(0) {
				t.Fatal("color gate did not name the singleton target")
			}
		})
	}
}
