package cardgen

import (
	"testing"

	"github.com/natefinch/council4/cardgen/oracle/compiler"
	"github.com/natefinch/council4/mtg/game"
)

func TestPlainTargetEffectConsumptionBoundary(t *testing.T) {
	t.Parallel()
	base := func() contentCtx {
		return contentCtx{content: compiler.AbilityContent{
			Effects: []compiler.CompiledEffect{{Exact: true}},
			Targets: []compiler.CompiledTarget{{Exact: true, Cardinality: compiler.TargetCardinality{Min: 1, Max: 1}}},
		}}
	}
	for _, tc := range []struct {
		name   string
		mutate func(*contentCtx)
		ok     bool
	}{
		{"plain", func(*contentCtx) {}, true},
		{"inexact", func(ctx *contentCtx) { ctx.content.Effects[0].Exact = false }, false},
		{"negated", func(ctx *contentCtx) { ctx.content.Effects[0].Negated = true }, false},
		{"another effect", func(ctx *contentCtx) {
			ctx.content.Effects = append(ctx.content.Effects, compiler.CompiledEffect{Exact: true})
		}, false},
		{"no target", func(ctx *contentCtx) { ctx.content.Targets = nil }, false},
		{"extra target", func(ctx *contentCtx) { ctx.content.Targets = append(ctx.content.Targets, ctx.content.Targets[0]) }, false},
		{"reference", func(ctx *contentCtx) { ctx.content.References = []compiler.CompiledReference{{}} }, false},
		{"condition", func(ctx *contentCtx) { ctx.content.Conditions = []compiler.CompiledCondition{{}} }, false},
		{"mode", func(ctx *contentCtx) { ctx.content.Modes = []compiler.CompiledMode{{}} }, false},
		{"payload-owned keyword", func(ctx *contentCtx) { ctx.content.Keywords = []compiler.CompiledKeyword{{}} }, true},
		{"payload-owned optional", func(ctx *contentCtx) { ctx.optional = true }, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			ctx := base()
			tc.mutate(&ctx)
			if got := plainTargetEffect(ctx); got != tc.ok {
				t.Fatalf("plain target consumption=%v, want %v", got, tc.ok)
			}
		})
	}
}

func TestSharedObjectFanoutKeepsPayloadAndTargetRange(t *testing.T) {
	t.Parallel()
	for _, tc := range []struct {
		body string
		want []string
	}{
		{"Two target creatures each get +2/-1 until end of turn.", []string{
			"Modes[0].Targets[0].MinTargets = 2",
			"Modes[0].Sequence[0].Primitive.(game.ModifyPT).PowerDelta.fixed = 2",
			"Modes[0].Sequence[1].Primitive.(game.ModifyPT).ToughnessDelta.fixed = -1",
		}},
		{"Up to two target creatures each gain flying until end of turn.", []string{
			"Modes[0].Targets[0].MaxTargets = 2",
			"Modes[0].Sequence[1].Primitive.(game.ApplyContinuous).Object.Val.targetIndex = 1",
			"Modes[0].Sequence[1].Primitive.(game.ApplyContinuous).Duration = game.DurationUntilEndOfTurn",
		}},
		{"Put two +1/+1 counters on each of up to two target creatures.", []string{
			"Modes[0].Targets[0].MaxTargets = 2",
			"Modes[0].Sequence[1].Primitive.(game.AddCounter).Amount.fixed = 2",
			"Modes[0].Sequence[1].Primitive.(game.AddCounter).Object.targetIndex = 1",
		}},
		{"If you control an artifact, two target creatures each get +2/+2 until end of turn.", []string{
			"Modes[0].Sequence[0].Primitive.(game.ModifyPT)",
			"Modes[0].Sequence[0].Condition",
			"Modes[0].Sequence[1].Primitive.(game.ModifyPT)",
		}},
		{"You may put two +1/+1 counters on each of up to two target creatures.", []string{
			"Modes[0].Sequence[0].Primitive.(game.AddCounter)",
			"Modes[0].Sequence[0].Optional = true",
			"Modes[0].Sequence[1].Primitive.(game.AddCounter)",
		}},
	} {
		t.Run(tc.body, func(t *testing.T) {
			t.Parallel()
			card := targetCountTestCard(tc.body)
			assertCardPaths(t, card, tc.want...)
			assertCardPathsAbsent(t, card, "Sequence[2]", "Targets[1]")
		})
	}
}

func TestSharedObjectFanoutKeepsUnboundedAndExactnessRefusals(t *testing.T) {
	t.Parallel()
	for _, tc := range []struct {
		name  string
		count compiler.TargetCardinality
		exact bool
		ok    bool
		slots int
	}{
		{"fixed", compiler.TargetCardinality{Min: 2, Max: 2}, true, true, 2},
		{"optional", compiler.TargetCardinality{Min: 0, Max: 2}, true, true, 2},
		{"exact X", compiler.TargetCardinality{Min: 0, Max: 99, CountEqualsX: true}, true, true, 20},
		{"unbounded", compiler.TargetCardinality{Min: 0, Max: 99}, true, false, 0},
		{"upper-bound X", compiler.TargetCardinality{Min: 0, Max: 99, MaxX: true}, true, false, 0},
		{"inexact target", compiler.TargetCardinality{Min: 2, Max: 2}, false, false, 0},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			target := compiler.CompiledTarget{
				Exact: tc.exact, Cardinality: tc.count,
				Selector: compiler.CompiledSelector{Kind: compiler.SelectorCreature},
			}
			content, ok := selectedGroupTargetMode(target, func(object game.ObjectReference) game.Primitive {
				return game.ModifyPT{Object: object, PowerDelta: game.Fixed(2), ToughnessDelta: game.Fixed(-1), Duration: game.DurationUntilEndOfTurn}
			}, false)
			if ok != tc.ok {
				t.Fatalf("mode accepted=%v, want %v", ok, tc.ok)
			}
			if ok && len(content.Modes[0].Sequence) != tc.slots {
				t.Fatalf("slots=%d, want %d", len(content.Modes[0].Sequence), tc.slots)
			}
		})
	}
}
