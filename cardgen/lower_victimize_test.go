package cardgen

import (
	goparser "go/parser"
	"go/token"
	"slices"
	"strings"
	"testing"

	"github.com/natefinch/council4/cardgen/oracle/compiler"
	"github.com/natefinch/council4/cardgen/oracle/parser"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
)

const victimizeText = "Choose two target creature cards in your graveyard. Sacrifice a creature. If you do, return the chosen cards to the battlefield tapped."

func TestLowerSacrificeConditionedReanimation(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Twin Revival",
		Layout:     "normal",
		TypeLine:   "Sorcery",
		OracleText: victimizeText,
	})
	mode := face.SpellAbility.Val.Modes[0]
	if len(mode.Targets) != 1 {
		t.Fatalf("targets = %#v; want one target group", mode.Targets)
	}
	target := mode.Targets[0]
	if target.MinTargets != 2 ||
		target.MaxTargets != 2 ||
		target.Allow != game.TargetAllowCard ||
		target.TargetZone != zone.Graveyard ||
		!target.Selection.Exists ||
		!slices.Equal(target.Selection.Val.RequiredTypes, []types.Card{types.Creature}) {
		t.Fatalf("target = %#v; want exactly two creature cards in your graveyard", target)
	}
	if len(mode.Sequence) != 2 {
		t.Fatalf("sequence = %#v; want sacrifice then simultaneous reanimation", mode.Sequence)
	}
	if err := game.ValidateInstructionSequence(mode.Sequence, mode.Targets); err != nil {
		t.Fatalf("invalid instruction sequence: %v", err)
	}
	sacrifice, ok := mode.Sequence[0].Primitive.(game.SacrificePermanents)
	if !ok ||
		sacrifice.Player != game.ControllerReference() ||
		sacrifice.Amount.Value() != 1 ||
		!slices.Equal(sacrifice.Selection.RequiredTypes, []types.Card{types.Creature}) {
		t.Fatalf("sacrifice = %#v; want controller sacrifice one creature", mode.Sequence[0])
	}
	if mode.Sequence[0].Optional ||
		mode.Sequence[0].PublishResult != sacrificeSucceededResultKey {
		t.Fatalf("sacrifice envelope = %#v", mode.Sequence[0])
	}
	instruction := mode.Sequence[1]
	put, ok := instruction.Primitive.(game.PutOnBattlefield)
	if !ok || len(put.Sources) != 2 || !put.EntryTapped {
		t.Fatalf("reanimation = %#v", instruction)
	}
	for i, source := range put.Sources {
		if source != game.CardBattlefieldSource(game.CardReference{
			Kind:        game.CardReferenceTarget,
			TargetIndex: i,
		}) {
			t.Fatalf("reanimation source %d = %#v", i, source)
		}
	}
	if !instruction.ResultGate.Exists ||
		instruction.ResultGate.Val.Key != sacrificeSucceededResultKey ||
		instruction.ResultGate.Val.Succeeded != game.TriTrue {
		t.Fatalf("reanimation gate = %#v", instruction.ResultGate)
	}
}

func TestGenerateSacrificeConditionedReanimationSource(t *testing.T) {
	t.Parallel()
	source, diagnostics, err := GenerateExecutableCardSource(&ScryfallCard{
		Name:       "Twin Revival",
		Layout:     "normal",
		TypeLine:   "Sorcery",
		OracleText: victimizeText,
	}, "c")
	if err != nil {
		t.Fatal(err)
	}
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	if _, err := goparser.ParseFile(token.NewFileSet(), "twin_revival.go", source, goparser.AllErrors); err != nil {
		t.Fatalf("generated source does not parse: %v\n%s", err, source)
	}
	for _, want := range []string{
		"MinTargets: 2",
		"MaxTargets: 2",
		"game.SacrificePermanents",
		`PublishResult: game.ResultKey("sacrifice-succeeded")`,
		"[]game.BattlefieldSource{game.CardBattlefieldSource",
		"game.CardReference{Kind: game.CardReferenceTarget, TargetIndex: 1}",
		"EntryTapped: true",
		"ResultGate: opt.Val(game.InstructionResultGate",
		"Key:",
		"Succeeded: game.TriTrue",
	} {
		if !strings.Contains(source, want) {
			t.Fatalf("generated source missing %q:\n%s", want, source)
		}
	}
}

func TestLowerSacrificeConditionedReanimationIsTextBlind(t *testing.T) {
	t.Parallel()
	document, diagnostics := parser.Parse(
		victimizeText,
		parser.Context{InstantOrSorcery: true},
	)
	if len(diagnostics) != 0 {
		t.Fatalf("parse diagnostics = %#v", diagnostics)
	}
	compilation, diagnostics := compiler.Compile(document, compiler.Context{})
	if len(diagnostics) != 0 {
		t.Fatalf("compile diagnostics = %#v", diagnostics)
	}
	content := compilation.Abilities[0].Content
	content.Targets[0].Text = "lowering must not inspect target text"
	content.References[0].Text = "or reference text"
	for i := range content.Effects {
		content.Effects[i].Text = "or effect text"
	}
	content.Conditions[0].Text = "or condition text"

	lowered, ok := lowerSacrificeConditionedReanimationSequence(contentCtx{
		content: content,
	})
	if !ok || len(lowered.Modes) != 1 || len(lowered.Modes[0].Sequence) != 2 {
		t.Fatalf("lowered = %#v, ok = %v; want typed sacrifice and simultaneous-return sequence", lowered, ok)
	}
}

func TestLowerSacrificeConditionedReanimationFailsClosed(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		text string
	}{
		{
			name: "wrong target cardinality",
			text: "Choose one target creature card in your graveyard. Sacrifice a creature. If you do, return the chosen cards to the battlefield tapped.",
		},
		{
			name: "unsupported sacrifice selection",
			text: "Choose two target creature cards in your graveyard. Sacrifice an artifact. If you do, return the chosen cards to the battlefield tapped.",
		},
		{
			name: "negative condition",
			text: "Choose two target creature cards in your graveyard. Sacrifice a creature. If you don't, return the chosen cards to the battlefield tapped.",
		},
		{
			name: "missing condition",
			text: "Choose two target creature cards in your graveyard. Sacrifice a creature. Return the chosen cards to the battlefield tapped.",
		},
		{
			name: "untapped return",
			text: "Choose two target creature cards in your graveyard. Sacrifice a creature. If you do, return the chosen cards to the battlefield.",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			_, diagnostics, err := GenerateExecutableCardSource(&ScryfallCard{
				Name:       "Boundary Test",
				Layout:     "normal",
				TypeLine:   "Sorcery",
				OracleText: test.text,
			}, "c")
			if err != nil {
				t.Fatal(err)
			}
			if len(diagnostics) == 0 {
				t.Fatal("unsupported sacrifice-conditioned sequence lowered without diagnostic")
			}
		})
	}
}
