package cardgen

import (
	"reflect"
	"strings"
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/types"
)

func alandraSkyDreamerCard() *ScryfallCard {
	return &ScryfallCard{
		Name:      "Alandra, Sky Dreamer",
		Layout:    "normal",
		TypeLine:  "Legendary Creature — Merfolk Wizard",
		ManaCost:  "{2}{U}{U}",
		Power:     new("2"),
		Toughness: new("4"),
		OracleText: "Whenever you draw your second card each turn, create a 2/2 blue Drake creature token with flying.\n" +
			"Whenever you draw your fifth card each turn, Alandra and Drakes you control each get +X/+X until end of turn, where X is the number of cards in your hand.",
	}
}

// TestLowerAlandraCoordinatedSelfGroupPump proves the coordinated "<self> and
// <group> each get +X/+X" pump (Alandra, Sky Dreamer) lowers to a two-instruction
// sequence: a ModifyPT that pumps the source permanent once and an ApplyContinuous
// that pumps every OTHER Drake the controller controls, both by the dynamic
// hand-size amount. Splitting the source from a source-excluding group is what
// keeps the coordinated subject from double-pumping the source.
func TestLowerAlandraCoordinatedSelfGroupPump(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, alandraSkyDreamerCard())
	if len(face.TriggeredAbilities) != 2 {
		t.Fatalf("triggered abilities = %d, want 2", len(face.TriggeredAbilities))
	}
	pump := face.TriggeredAbilities[1]
	if pump.Trigger.Pattern.PlayerEventOrdinalThisTurn != 5 {
		t.Fatalf("pump trigger ordinal = %d, want 5", pump.Trigger.Pattern.PlayerEventOrdinalThisTurn)
	}
	mode := pump.Content.Modes[0]
	if len(mode.Sequence) != 2 {
		t.Fatalf("pump sequence = %#v, want 2 instructions", mode.Sequence)
	}

	modifyPT, ok := mode.Sequence[0].Primitive.(game.ModifyPT)
	if !ok {
		t.Fatalf("sequence[0] = %T, want game.ModifyPT", mode.Sequence[0].Primitive)
	}
	if modifyPT.Object != game.SourcePermanentReference() {
		t.Fatalf("source pump Object = %v, want SourcePermanentReference", modifyPT.Object)
	}
	if modifyPT.Duration != game.DurationUntilEndOfTurn {
		t.Fatalf("source pump Duration = %v, want DurationUntilEndOfTurn", modifyPT.Duration)
	}
	assertHandSizeDynamic(t, "source power", modifyPT.PowerDelta)
	assertHandSizeDynamic(t, "source toughness", modifyPT.ToughnessDelta)

	apply, ok := mode.Sequence[1].Primitive.(game.ApplyContinuous)
	if !ok {
		t.Fatalf("sequence[1] = %T, want game.ApplyContinuous", mode.Sequence[1].Primitive)
	}
	if apply.Duration != game.DurationUntilEndOfTurn {
		t.Fatalf("group pump Duration = %v, want DurationUntilEndOfTurn", apply.Duration)
	}
	if len(apply.ContinuousEffects) != 1 {
		t.Fatalf("group ContinuousEffects = %d, want 1", len(apply.ContinuousEffects))
	}
	effect := apply.ContinuousEffects[0]
	if effect.Layer != game.LayerPowerToughnessModify {
		t.Fatalf("group Layer = %v, want LayerPowerToughnessModify", effect.Layer)
	}
	wantGroup := game.BattlefieldGroupExcluding(
		game.Selection{SubtypesAny: []types.Sub{types.Drake}, Controller: game.ControllerYou},
		game.SourcePermanentReference(),
	)
	if !reflect.DeepEqual(effect.Group, wantGroup) {
		t.Fatalf("group = %#v, want other Drakes you control excluding source", effect.Group)
	}
	if !effect.PowerDeltaDynamic.Exists || effect.PowerDeltaDynamic.Val.Kind != game.DynamicAmountCountCardsInZone {
		t.Fatalf("group PowerDeltaDynamic = %#v, want hand-size count", effect.PowerDeltaDynamic)
	}
	if !effect.ToughnessDeltaDynamic.Exists || effect.ToughnessDeltaDynamic.Val.Kind != game.DynamicAmountCountCardsInZone {
		t.Fatalf("group ToughnessDeltaDynamic = %#v, want hand-size count", effect.ToughnessDeltaDynamic)
	}
}

// TestGenerateAlandraSourceShape guards the generated source string against the
// two-instruction coordinated-pump shape, the drake token, and the ordinal-draw
// triggers, so a regression that drops the source pump or the excluding group is
// caught at the source level.
func TestGenerateAlandraSourceShape(t *testing.T) {
	t.Parallel()
	source, diagnostics, err := GenerateExecutableCardSource(alandraSkyDreamerCard(), "a")
	if err != nil {
		t.Fatal(err)
	}
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	for _, wanted := range []string{
		"PlayerEventOrdinalThisTurn: 2",
		"PlayerEventOrdinalThisTurn: 5",
		"Primitive: game.ModifyPT{",
		"Object: game.SourcePermanentReference(),",
		"game.DynamicAmountCountCardsInZone",
		"game.BattlefieldGroupExcluding(game.Selection{SubtypesAny: []types.Sub{types.Sub(\"Drake\")}, Controller: game.ControllerYou}, game.SourcePermanentReference())",
	} {
		if !strings.Contains(source, wanted) {
			t.Fatalf("source missing %q:\n%s", wanted, source)
		}
	}
}

// TestAlandraCoordinatedPumpFailsClosedForUnmappedGroup proves a coordinated
// "<self> and <group> each get" pump whose group has no source-excluding variant
// (opponent-controlled creatures), and the "each"-less near-miss that would drop
// the self conjunct, both fail closed with a diagnostic rather than silently
// narrowing to the source alone or a wrong group.
func TestAlandraCoordinatedPumpFailsClosedForUnmappedGroup(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name       string
		oracleText string
	}{
		{
			name: "group has no source-excluding variant",
			oracleText: "Whenever you draw your fifth card each turn, Alandra and creatures your opponents control each get +X/+X until end of turn, " +
				"where X is the number of cards in your hand.",
		},
		{
			name:       "missing each drops the self conjunct",
			oracleText: "Whenever you draw your fifth card each turn, Alandra and Drakes you control get +1/+1 until end of turn.",
		},
	}
	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			lowerSingleFaceExpectingUnsupported(t, &ScryfallCard{
				Name:       "Alandra, Sky Dreamer",
				Layout:     "normal",
				TypeLine:   "Legendary Creature — Merfolk Wizard",
				ManaCost:   "{2}{U}{U}",
				Power:      new("2"),
				Toughness:  new("4"),
				OracleText: test.oracleText,
			})
		})
	}
}

func assertHandSizeDynamic(t *testing.T, label string, quantity game.Quantity) {
	t.Helper()
	if !quantity.IsDynamic() {
		t.Fatalf("%s delta = %#v, want dynamic hand-size amount", label, quantity)
	}
	dynamic := quantity.DynamicAmount()
	if !dynamic.Exists || dynamic.Val.Kind != game.DynamicAmountCountCardsInZone {
		t.Fatalf("%s dynamic = %#v, want DynamicAmountCountCardsInZone", label, dynamic)
	}
}
