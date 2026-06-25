package cardgen

import (
	"strings"
	"testing"

	"github.com/natefinch/council4/cardgen/oracle/compiler"
	"github.com/natefinch/council4/cardgen/oracle/parser"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/counter"
)

// TestGenerateExecutableCardSourceRoomDoorUnlockSplit verifies that a Duskmourn
// Room split enchantment whose halves carry "When you unlock this door, ..."
// triggers generates. The door-unlock trigger lowers onto the self
// enters-the-battlefield event (the cast face unlocks its door as it enters),
// the Moldering Gym half reuses the basic-land tutor, and the Weight Room half
// uses the manifest-dread-then-counters sequence.
func TestGenerateExecutableCardSourceRoomDoorUnlockSplit(t *testing.T) {
	t.Parallel()
	card := &ScryfallCard{
		Name:          "Moldering Gym // Weight Room",
		Layout:        "split",
		ColorIdentity: []string{"G"},
		CardFaces: []ScryfallCardFace{
			{
				Name:       "Moldering Gym",
				ManaCost:   "{2}{G}",
				TypeLine:   "Enchantment — Room",
				OracleText: "When you unlock this door, search your library for a basic land card, put it onto the battlefield tapped, then shuffle.\n(You may cast either half. That door unlocks on the battlefield. As a sorcery, you may pay the mana cost of a locked door to unlock it.)",
			},
			{
				Name:       "Weight Room",
				ManaCost:   "{5}{G}",
				TypeLine:   "Enchantment — Room",
				OracleText: "When you unlock this door, manifest dread, then put three +1/+1 counters on that creature.\n(You may cast either half. That door unlocks on the battlefield. As a sorcery, you may pay the mana cost of a locked door to unlock it.)",
			},
		},
	}

	source, diagnostics, err := GenerateExecutableCardSource(card, "m")
	if err != nil {
		t.Fatal(err)
	}
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	for _, wanted := range []string{
		`Name: "Moldering Gym",`,
		`Name: "Weight Room",`,
		"Subtypes: []types.Sub{types.Room},",
		"Layout: game.LayoutSplit,",
		"Event:  game.EventPermanentEnteredBattlefield,",
		"Source: game.TriggerSourceSelf,",
		"Primitive: game.Search{",
		"Primitive: game.Manifest{",
		`PublishLinked: game.LinkedKey("manifested-creature"),`,
		"Primitive: game.AddCounter{",
		"Amount:      game.Fixed(3),",
		`Object:      game.LinkedObjectReference("manifested-creature"),`,
		"CounterKind: counter.PlusOnePlusOne,",
	} {
		if !strings.Contains(source, wanted) {
			t.Fatalf("source missing %q:\n%s", wanted, source)
		}
	}
}

func compiledWeightRoomBodyContentCtx(t *testing.T) contentCtx {
	t.Helper()
	compilation, diagnostics := compileTestOracle(
		"Manifest dread, then put three +1/+1 counters on that creature.",
		parser.Context{},
		compiler.Context{},
	)
	if len(diagnostics) != 0 {
		t.Fatalf("compile diagnostics = %#v", diagnostics)
	}
	content := compilation.Abilities[0].Content
	return contentCtx{span: content.Span, content: content}
}

func TestLowerManifestDreadThenCountersSequence(t *testing.T) {
	t.Parallel()
	ctx := compiledWeightRoomBodyContentCtx(t)
	content, ok := lowerManifestDreadThenCountersSequence(ctx)
	if !ok {
		t.Fatal("manifest-dread-then-counters sequence did not lower")
	}
	if len(content.Modes) != 1 || len(content.Modes[0].Sequence) != 2 {
		t.Fatalf("content = %#v, want one mode of two instructions", content)
	}
	manifest, ok := content.Modes[0].Sequence[0].Primitive.(game.Manifest)
	if !ok {
		t.Fatalf("first instruction = %#v, want game.Manifest", content.Modes[0].Sequence[0].Primitive)
	}
	if !manifest.Dread || manifest.PublishLinked != game.LinkedKey(manifestedCreatureLinkKey) {
		t.Fatalf("manifest = %#v, want dread with published link", manifest)
	}
	counters, ok := content.Modes[0].Sequence[1].Primitive.(game.AddCounter)
	if !ok {
		t.Fatalf("second instruction = %#v, want game.AddCounter", content.Modes[0].Sequence[1].Primitive)
	}
	if counters.CounterKind != counter.PlusOnePlusOne {
		t.Fatalf("counter kind = %v, want +1/+1", counters.CounterKind)
	}
}

func TestLowerManifestDreadThenCountersSequenceFailsClosed(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name   string
		mutate func(ctx *contentCtx)
	}{
		{"optional", func(ctx *contentCtx) { ctx.optional = true }},
		{"extra effect", func(ctx *contentCtx) {
			ctx.content.Effects = append(ctx.content.Effects, ctx.content.Effects[1])
		}},
		{"negated counter", func(ctx *contentCtx) { ctx.content.Effects[1].Negated = true }},
		{"unknown amount", func(ctx *contentCtx) { ctx.content.Effects[1].Amount.Known = false }},
		{"unknown counter kind", func(ctx *contentCtx) { ctx.content.Effects[1].CounterKindKnown = false }},
		{"no reference", func(ctx *contentCtx) { ctx.content.References = nil }},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			ctx := compiledWeightRoomBodyContentCtx(t)
			test.mutate(&ctx)
			if _, ok := lowerManifestDreadThenCountersSequence(ctx); ok {
				t.Fatal("lowered unsupported boundary mutation")
			}
		})
	}
}
