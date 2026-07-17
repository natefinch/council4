package cardgen

import (
	"strings"
	"testing"

	"github.com/natefinch/council4/mtg/game"
)

const electrodominanceOracle = "Electrodominance deals X damage to any target. You may cast a spell with mana value X or less from your hand without paying its mana cost."

func TestLowerElectrodominanceComposesDamageAndFreeCast(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Electrodominance",
		Layout:     "normal",
		ManaCost:   "{X}{R}{R}",
		TypeLine:   "Instant",
		OracleText: electrodominanceOracle,
		Colors:     []string{"R"},
	})
	mode := face.SpellAbility.Val.Modes[0]
	if len(mode.Targets) != 1 || len(mode.Sequence) != 2 {
		t.Fatalf("mode = %#v", mode)
	}
	damage, ok := mode.Sequence[0].Primitive.(game.Damage)
	if !ok || damage.Amount.DynamicAmount().Val.Kind != game.DynamicAmountX {
		t.Fatalf("damage primitive = %#v", mode.Sequence[0].Primitive)
	}
	cast, ok := mode.Sequence[1].Primitive.(game.CastForFree)
	if !ok || !cast.MaxManaValueFromX || !mode.Sequence[1].Optional {
		t.Fatalf("free-cast instruction = %#v", mode.Sequence[1])
	}
}

func TestGenerateElectrodominanceSourceIsTextBlindAndComposable(t *testing.T) {
	t.Parallel()
	source, diagnostics, err := GenerateExecutableCardSource(&ScryfallCard{
		Name:       "Electrodominance",
		Layout:     "normal",
		ManaCost:   "{X}{R}{R}",
		TypeLine:   "Instant",
		OracleText: electrodominanceOracle,
		Colors:     []string{"R"},
	}, "e")
	if err != nil {
		t.Fatal(err)
	}
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	for _, want := range []string{
		"game.AnyTargetDamageRecipient(0)",
		"Kind: game.DynamicAmountX",
		"Primitive: game.CastForFree",
		"MaxManaValueFromX: true",
		"Optional: true",
	} {
		if !strings.Contains(source, want) {
			t.Fatalf("source missing %q:\n%s", want, source)
		}
	}
}
