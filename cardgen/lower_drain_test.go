package cardgen

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/counter"
)

func drainInstructions(t *testing.T, content game.AbilityContent) (game.LoseLife, game.GainLife) {
	t.Helper()
	mode := content.Modes[0]
	if len(mode.Sequence) != 2 {
		t.Fatalf("sequence = %#v, want two instructions", mode.Sequence)
	}
	lose, ok := mode.Sequence[0].Primitive.(game.LoseLife)
	if !ok {
		t.Fatalf("first primitive = %T, want game.LoseLife", mode.Sequence[0].Primitive)
	}
	gain, ok := mode.Sequence[1].Primitive.(game.GainLife)
	if !ok {
		t.Fatalf("second primitive = %T, want game.GainLife", mode.Sequence[1].Primitive)
	}
	return lose, gain
}

// TestLowerFellBeastDrain verifies the anchor card: "Whenever this creature
// enters or attacks, target opponent loses X life and you gain X life, where X
// is the number of +1/+1 counters on it." The drain targets one opponent, loses
// the source's +1/+1 counter count from them, and gains that same amount for the
// controller.
func TestLowerFellBeastDrain(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:     "Test Fell Beast",
		Layout:   "normal",
		TypeLine: "Creature",
		OracleText: "Flying\nDevour 1 (As this creature enters, you may sacrifice any number of creatures. " +
			"It enters with that many +1/+1 counters on it.)\n" +
			"Whenever this creature enters or attacks, target opponent loses X life and you gain X life, " +
			"where X is the number of +1/+1 counters on it.",
	})
	if len(face.TriggeredAbilities) != 1 {
		t.Fatalf("triggered abilities = %d, want 1 (the enters-or-attacks drain)", len(face.TriggeredAbilities))
	}
	content := face.TriggeredAbilities[0].Content
	if len(content.Modes[0].Targets) != 1 {
		t.Fatalf("targets = %#v, want one target opponent", content.Modes[0].Targets)
	}
	lose, gain := drainInstructions(t, content)
	if lose.Player != game.TargetPlayerReference(0) {
		t.Fatalf("lose player = %+v, want target player 0", lose.Player)
	}
	if gain.Player != game.ControllerReference() {
		t.Fatalf("gain player = %+v, want controller", gain.Player)
	}
	for label, amount := range map[string]game.Quantity{"lose": lose.Amount, "gain": gain.Amount} {
		dynamic := amount.DynamicAmount()
		if !dynamic.Exists {
			t.Fatalf("%s amount = %+v, want dynamic source counter count", label, amount)
		}
		got := dynamic.Val
		if got.Kind != game.DynamicAmountObjectCounters ||
			got.Multiplier != 1 ||
			got.Object != game.SourcePermanentReference() ||
			got.CounterKind != counter.PlusOnePlusOne {
			t.Fatalf("%s dynamic = %+v, want source +1/+1 counter count", label, got)
		}
	}
}

// TestLowerEachOpponentDrainVariants verifies group drains "each opponent loses
// X life and you gain X life" across supported dynamic-X definitions and the
// fixed-amount form, each emitting an opponents-group LoseLife paired with a
// controller GainLife that share one amount.
func TestLowerEachOpponentDrainVariants(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name       string
		oracleText string
		wantKind   game.DynamicAmountKind
		wantFixed  int
	}{
		{
			name:       "basic land types",
			oracleText: "When this creature enters, each opponent loses X life and you gain X life, where X is the number of basic land types among lands you control.",
			wantKind:   game.DynamicAmountControllerBasicLandTypeCount,
		},
		{
			name:       "source power",
			oracleText: "When this creature dies, each opponent loses X life and you gain X life, where X is its power.",
			wantKind:   game.DynamicAmountObjectPower,
		},
		{
			name:       "fixed amount",
			oracleText: "When this creature enters, each opponent loses 2 life and you gain 2 life.",
			wantFixed:  2,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			face := lowerSingleFace(t, &ScryfallCard{
				Name:       "Test Drain " + test.name,
				Layout:     "normal",
				TypeLine:   "Creature",
				OracleText: test.oracleText,
			})
			content := face.TriggeredAbilities[0].Content
			if len(content.Modes[0].Targets) != 0 {
				t.Fatalf("targets = %#v, want none", content.Modes[0].Targets)
			}
			lose, gain := drainInstructions(t, content)
			if lose.PlayerGroup != game.OpponentsReference() {
				t.Fatalf("lose group = %+v, want opponents", lose.PlayerGroup)
			}
			if gain.Player != game.ControllerReference() {
				t.Fatalf("gain player = %+v, want controller", gain.Player)
			}
			if lose.Amount != gain.Amount {
				t.Fatalf("lose amount %+v != gain amount %+v", lose.Amount, gain.Amount)
			}
			if test.wantFixed != 0 {
				if lose.Amount != game.Fixed(test.wantFixed) {
					t.Fatalf("amount = %+v, want fixed %d", lose.Amount, test.wantFixed)
				}
				return
			}
			dynamic := lose.Amount.DynamicAmount()
			if !dynamic.Exists || dynamic.Val.Kind != test.wantKind {
				t.Fatalf("amount = %+v, want dynamic kind %v", lose.Amount, test.wantKind)
			}
		})
	}
}

// TestLowerTargetPlayerDrainGreatestPower verifies the target-player drain with
// the "greatest power among creatures you control" dynamic: "target player loses
// X life and you gain X life".
func TestLowerTargetPlayerDrainGreatestPower(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Essence Drain",
		Layout:     "normal",
		TypeLine:   "Creature",
		OracleText: "When this creature enters, target player loses X life and you gain X life, where X is the greatest power among creatures you control.",
	})
	content := face.TriggeredAbilities[0].Content
	if len(content.Modes[0].Targets) != 1 {
		t.Fatalf("targets = %#v, want one target player", content.Modes[0].Targets)
	}
	lose, gain := drainInstructions(t, content)
	if lose.Player != game.TargetPlayerReference(0) {
		t.Fatalf("lose player = %+v, want target player 0", lose.Player)
	}
	if lose.Amount != gain.Amount {
		t.Fatalf("lose amount %+v != gain amount %+v", lose.Amount, gain.Amount)
	}
	dynamic := lose.Amount.DynamicAmount()
	if !dynamic.Exists || dynamic.Val.Kind != game.DynamicAmountGreatestPowerInGroup {
		t.Fatalf("amount = %+v, want greatest-power-in-group", lose.Amount)
	}
}

// TestLowerDrainFailsClosed verifies drains that should not lower keep failing
// the round-trip: mismatched fixed amounts and an unparsed "where X is ..."
// definition (which leaves the gain clause non-exact) both stay unsupported.
func TestLowerDrainFailsClosed(t *testing.T) {
	t.Parallel()
	tests := []string{
		"When this creature enters, each opponent loses X life and you gain X life, where X is the number of creatures in your party.",
		"When this creature enters, each opponent loses X life and you gain X life, where X is the number of times this creature has mutated.",
	}
	for _, oracleText := range tests {
		t.Run(oracleText, func(t *testing.T) {
			t.Parallel()
			lowerSingleFaceExpectingUnsupported(t, &ScryfallCard{
				Name:       "Test Unsupported Drain",
				Layout:     "normal",
				TypeLine:   "Creature",
				OracleText: oracleText,
			})
		})
	}
}

// TestLowerEventControllerDrain verifies the punisher drain whose drained
// subject is the controller of a triggering or related permanent rather than a
// target or group. Revenge of Ravens' "Whenever a creature attacks you ..., that
// creature's controller loses 1 life and you gain 1 life." drains the attacking
// creature's controller through an event-permanent reference and gains the same
// life for the source's controller, with no targets.
func TestLowerEventControllerDrain(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Raven Revenge",
		Layout:     "normal",
		TypeLine:   "Enchantment",
		OracleText: "Whenever a creature attacks you or a planeswalker you control, that creature's controller loses 1 life and you gain 1 life.",
	})
	content := face.TriggeredAbilities[0].Content
	if len(content.Modes[0].Targets) != 0 {
		t.Fatalf("targets = %#v, want no targets", content.Modes[0].Targets)
	}
	lose, gain := drainInstructions(t, content)
	if lose.Player != game.ObjectControllerReference(game.EventPermanentReference()) {
		t.Fatalf("lose player = %+v, want controller of event permanent", lose.Player)
	}
	if gain.Player != game.ControllerReference() {
		t.Fatalf("gain player = %+v, want controller", gain.Player)
	}
	if lose.Amount != gain.Amount {
		t.Fatalf("lose amount %+v != gain amount %+v", lose.Amount, gain.Amount)
	}
	if lose.Amount != game.Fixed(1) {
		t.Fatalf("amount = %+v, want fixed 1", lose.Amount)
	}
}
