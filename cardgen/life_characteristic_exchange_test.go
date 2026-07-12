package cardgen

import (
	"strings"
	"testing"

	"github.com/natefinch/council4/mtg/game"
)

func TestLowerLifeCharacteristicExchangeFamily(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name           string
		manaCost       string
		oracleText     string
		characteristic game.SourcePowerToughness
		targetOpponent bool
	}{
		{
			name:           "Tree of Redemption",
			manaCost:       "{3}{G}",
			oracleText:     "Defender\n{T}: Exchange your life total with this creature's toughness.",
			characteristic: game.SourceToughness,
		},
		{
			name:           "Tree of Perdition",
			manaCost:       "{3}{B}",
			oracleText:     "Defender\n{T}: Exchange target opponent's life total with this creature's toughness.",
			characteristic: game.SourceToughness,
			targetOpponent: true,
		},
		{
			name:           "Evra, Halcyon Witness",
			manaCost:       "{4}{W}{W}",
			oracleText:     "Lifelink\n{4}: Exchange your life total with Evra's power.",
			characteristic: game.SourcePower,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			face := lowerSingleFace(t, &ScryfallCard{
				Name:       test.name,
				Layout:     "normal",
				TypeLine:   "Creature — Test",
				ManaCost:   test.manaCost,
				Power:      new("0"),
				Toughness:  new("13"),
				OracleText: test.oracleText,
			})
			mode := face.ActivatedAbilities[0].Content.Modes[0]
			exchange, ok := mode.Sequence[0].Primitive.(game.ExchangeLifeTotalWithSourceCharacteristic)
			if !ok || exchange.Characteristic != test.characteristic {
				t.Fatalf("exchange = %#v", mode.Sequence[0].Primitive)
			}
			if test.targetOpponent {
				if len(mode.Targets) != 1 ||
					exchange.Player != game.TargetPlayerReference(0) ||
					!mode.Targets[0].Selection.Exists ||
					mode.Targets[0].Selection.Val.Player != game.PlayerOpponent {
					t.Fatalf("targeted exchange = %#v, targets = %#v", exchange, mode.Targets)
				}
			} else if len(mode.Targets) != 0 || exchange.Player != game.ControllerReference() {
				t.Fatalf("controller exchange = %#v, targets = %#v", exchange, mode.Targets)
			}
		})
	}
}

func TestGenerateLifeCharacteristicExchange(t *testing.T) {
	t.Parallel()
	source, diagnostics, err := GenerateExecutableCardSource(&ScryfallCard{
		Name:       "Tree of Perdition",
		Layout:     "normal",
		TypeLine:   "Creature — Plant",
		ManaCost:   "{3}{B}",
		Power:      new("0"),
		Toughness:  new("13"),
		OracleText: "Defender\n{T}: Exchange target opponent's life total with this creature's toughness.",
	}, "t")
	if err != nil {
		t.Fatal(err)
	}
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	for _, wanted := range []string{
		"Primitive: game.ExchangeLifeTotalWithSourceCharacteristic{",
		"game.TargetPlayerReference(0)",
		"Characteristic: game.SourceToughness",
	} {
		if !strings.Contains(source, wanted) {
			t.Fatalf("source missing %q:\n%s", wanted, source)
		}
	}
}
