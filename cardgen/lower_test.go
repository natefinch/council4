package cardgen

import (
	"testing"

	"github.com/natefinch/council4/cardgen/oracle/compiler"
	"github.com/natefinch/council4/cardgen/oracle/parser"
	"github.com/natefinch/council4/cardgen/oracle/shared"
	"github.com/natefinch/council4/mtg/game"
)

func lowerSingleFace(t *testing.T, card *ScryfallCard) loweredFaceAbilities {
	t.Helper()
	faces, diagnostics := lowerExecutableFaces(card)
	if len(diagnostics) != 0 {
		t.Fatalf("unexpected diagnostics: %#v", diagnostics)
	}

	if len(faces) == 0 {
		t.Fatal("no faces lowered")
	}
	return faces[0]
}

func TestLowerEventPlayerCoordinatedSubjectInTrigger(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Watcher",
		Layout:     "normal",
		TypeLine:   "Enchantment",
		OracleText: "Whenever a player draws a card, they discard a card, then draw a card.",
	})
	mode := face.TriggeredAbilities[0].Content.Modes[0]
	if len(mode.Sequence) != 2 {
		t.Fatalf("sequence = %#v", mode.Sequence)
	}
	draw, ok := mode.Sequence[1].Primitive.(game.Draw)
	if !ok || draw.Player != game.EventPlayerReference() {
		t.Fatalf("coordinated draw = %#v", mode.Sequence[1].Primitive)
	}
}

func compileTestOracle(source string, parserContext parser.Context, compilerContext compiler.Context) (compiler.Compilation, []shared.Diagnostic) {
	document, diagnostics := parser.Parse(source, parserContext)
	compilation, compilerDiagnostics := compiler.Compile(document, compilerContext)
	return compilation, append(diagnostics, compilerDiagnostics...)
}

func lowerKeywordForTest(t *testing.T, oracleText string, kind game.Keyword) game.KeywordAbility {
	t.Helper()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Parameterized Creature",
		Layout:     "normal",
		TypeLine:   "Creature — Test",
		OracleText: oracleText,
		Power:      new("2"),
		Toughness:  new("2"),
	})
	if len(face.StaticAbilities) != 1 {
		t.Fatalf("static abilities = %d, want 1", len(face.StaticAbilities))
	}
	keyword, ok := game.BodyKeywordAbility(face.StaticAbilities[0].Body, kind)
	if !ok {
		t.Fatalf("%v keyword not found in %#v", kind, face.StaticAbilities[0].Body)
	}
	return keyword
}

func counterPlacementAmount(primitive game.Primitive) (game.Quantity, bool) {
	switch primitive.Kind() {
	case game.PrimitiveAddCounter:
		add, ok := primitive.(game.AddCounter)
		return add.Amount, ok
	case game.PrimitiveAddPlayerCounter:
		add, ok := primitive.(game.AddPlayerCounter)
		return add.Amount, ok
	default:
		return game.Quantity{}, false
	}
}

// checkGainControlSequence validates the standard gain-control sequence:
//
//	Instruction 0: ApplyContinuous (LayerControl, NewController = Player1)
//	Instruction 1 (optional): Untap
//	Instruction 2 (optional): ApplyContinuous (LayerAbility, AddKeywords = [Haste])
func checkGainControlPrimitive(t *testing.T, mode game.Mode, seqIdx int, duration game.EffectDuration) {
	t.Helper()
	prim, ok := mode.Sequence[seqIdx].Primitive.(game.ApplyContinuous)
	if !ok {
		t.Fatalf("sequence[%d] = %T, want game.ApplyContinuous", seqIdx, mode.Sequence[seqIdx].Primitive)
	}
	if !prim.Object.Exists || prim.Object.Val != game.TargetPermanentReference(0) {
		t.Fatalf("ApplyContinuous.Object = %v, want TargetPermanentReference(0)", prim.Object)
	}
	if len(prim.ContinuousEffects) != 1 {
		t.Fatalf("ContinuousEffects len = %d, want 1", len(prim.ContinuousEffects))
	}
	eff := prim.ContinuousEffects[0]
	if eff.Layer != game.LayerControl {
		t.Fatalf("Layer = %v, want LayerControl", eff.Layer)
	}
	if !eff.NewController.Exists || eff.NewController.Val != game.Player1 {
		t.Fatalf("NewController = %v, want Player1", eff.NewController)
	}
	if prim.Duration != duration {
		t.Fatalf("Duration = %v, want %v", prim.Duration, duration)
	}
}

func checkUntapPrimitive(t *testing.T, mode game.Mode, seqIdx int) {
	t.Helper()
	untap, ok := mode.Sequence[seqIdx].Primitive.(game.Untap)
	if !ok {
		t.Fatalf("sequence[%d] = %T, want game.Untap", seqIdx, mode.Sequence[seqIdx].Primitive)
	}
	if untap.Object != game.TargetPermanentReference(0) {
		t.Fatalf("Untap.Object = %v, want TargetPermanentReference(0)", untap.Object)
	}
}

func checkKeywordGrantPrimitive(t *testing.T, mode game.Mode, seqIdx int, keyword game.Keyword) {
	t.Helper()
	prim, ok := mode.Sequence[seqIdx].Primitive.(game.ApplyContinuous)
	if !ok {
		t.Fatalf("sequence[%d] = %T, want game.ApplyContinuous (keyword grant)", seqIdx, mode.Sequence[seqIdx].Primitive)
	}
	if !prim.Object.Exists || prim.Object.Val != game.TargetPermanentReference(0) {
		t.Fatalf("keyword grant Object = %v, want TargetPermanentReference(0)", prim.Object)
	}
	if len(prim.ContinuousEffects) != 1 {
		t.Fatalf("keyword grant ContinuousEffects len = %d, want 1", len(prim.ContinuousEffects))
	}
	eff := prim.ContinuousEffects[0]
	if eff.Layer != game.LayerAbility {
		t.Fatalf("keyword grant Layer = %v, want LayerAbility", eff.Layer)
	}
	if len(eff.AddKeywords) != 1 || eff.AddKeywords[0] != keyword {
		t.Fatalf("AddKeywords = %v, want [%v]", eff.AddKeywords, keyword)
	}
	if prim.Duration != game.DurationUntilEndOfTurn {
		t.Fatalf("keyword grant Duration = %v, want DurationUntilEndOfTurn", prim.Duration)
	}
}
