package cardgen

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

func TestDepartureCharacteristicGeneratedRoundTrip(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping generated-package semantic round-trip in short mode")
	}
	cards := []*ScryfallCard{{
		Name: "RT Departure Capture", Layout: "normal", TypeLine: "Creature", Power: new("6"), Toughness: new("6"),
		OracleText: "When this creature enters, return target creature card from your graveyard to your hand. You gain life equal to that card's power.",
	}, {
		Name: "RT Departure Toughness", Layout: "normal", TypeLine: "Creature", Power: new("6"), Toughness: new("6"),
		OracleText: "When this creature enters, return target creature card from your graveyard to your hand. You gain life equal to that card's toughness.",
	}}
	dir, pkg := writeCardRoundTripPackage(t, cards)
	source := fmt.Sprintf(`package %s

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
)

func TestDepartureCaptureWiring(t *testing.T) {
	for _, tc := range []struct {
		def *game.CardDef
		power bool
	}{
		{RTDepartureCapture(), true},
		{RTDepartureToughness(), false},
	} {
		sequence := tc.def.TriggeredAbilities[0].Content.Modes[0].Sequence
		if len(sequence) != 2 {
			t.Fatalf("sequence = %%+v, want move and consumer", sequence)
		}
		move := sequence[0].Primitive.(game.MoveCard)
		key, other := move.PublishDepartureCharacteristics.Power, move.PublishDepartureCharacteristics.Toughness
		if !tc.power {
			key, other = other, key
		}
		if key == "" || other != "" || !sequence[0].LocalProducts.HasResult(key) {
			t.Fatalf("generated capture/local declaration lost: %%+v", sequence[0])
		}
		gain := sequence[1].Primitive.(game.GainLife)
		amount := gain.Amount.DynamicAmount()
		if !amount.Exists || amount.Val.Kind != game.DynamicAmountPreviousEffectResult ||
			amount.Val.ResultKey != key || !sequence[1].ResultGate.Exists ||
			sequence[1].ResultGate.Val.Key != key || !sequence[1].ResultGate.Val.AmountAvailable {
			t.Fatalf("generated consumer lost exact scalar or availability: %%+v", sequence[1])
		}
	}
}
`, pkg)
	if err := os.WriteFile(filepath.Join(dir, "semantic_test.go"), []byte(source), 0o600); err != nil {
		t.Fatal(err)
	}
	cmd := exec.CommandContext(context.Background(), "go", "test", "-count=1", "./"+filepath.Base(dir))
	if output, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("departure-characteristic generated round-trip: %v\n%s", err, output)
	}
}
