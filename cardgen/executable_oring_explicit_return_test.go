package cardgen

import (
	"strings"
	"testing"
)

func TestGenerateExecutableExplicitORingReturn(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name       string
		typeLine   string
		manaCost   string
		oracleText string
		power      *string
		toughness  *string
	}{
		{
			name:     "Journey to Nowhere",
			typeLine: "Enchantment",
			manaCost: "{1}{W}",
			oracleText: "When Journey to Nowhere enters, exile target creature.\n" +
				"When Journey to Nowhere leaves the battlefield, return the exiled card to the battlefield under its owner's control.",
		},
		{
			name:     "Oblivion Ring",
			typeLine: "Enchantment",
			manaCost: "{2}{W}",
			oracleText: "When Oblivion Ring enters, exile another target nonland permanent.\n" +
				"When Oblivion Ring leaves the battlefield, return the exiled card to the battlefield under its owner's control.",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			card := &ScryfallCard{
				Name:       test.name,
				Layout:     "normal",
				ManaCost:   test.manaCost,
				TypeLine:   test.typeLine,
				OracleText: test.oracleText,
				Colors:     []string{"W"},
			}
			source, diagnostics, err := GenerateExecutableCardSource(card, "o")
			if err != nil {
				t.Fatal(err)
			}
			if len(diagnostics) != 0 {
				t.Fatalf("diagnostics = %#v", diagnostics)
			}
			for _, wanted := range []string{
				"Primitive: game.Exile",
				"Object:         game.TargetPermanentReference(0)",
				`ExileLinkedKey: game.LinkedKey("exile-until-leaves")`,
				"Primitive: game.PutOnBattlefield",
				`game.LinkedBattlefieldSource(game.LinkedKey("exile-until-leaves"))`,
				"game.EventZoneChanged",
			} {
				if !strings.Contains(source, wanted) {
					t.Fatalf("source missing %q:\n%s", wanted, source)
				}
			}
		})
	}
}

// TestGenerateExecutableExileUntilLeavesTargetWordings exercises the broadened
// target wordings of the single-ability O-Ring exile-until-leaves form: the "up
// to one target" cardinality and the optional "you may exile ..." offer. Both
// must still lower to a linked exile under the exile-until-leaves key with the
// synthesized leaves-the-battlefield return trigger.
func TestGenerateExecutableExileUntilLeavesTargetWordings(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name       string
		typeLine   string
		manaCost   string
		oracleText string
		wantOpt    bool
		wantUpTo   bool
	}{
		{
			name:       "Touch the Spirit Realm",
			typeLine:   "Enchantment",
			manaCost:   "{1}{W}",
			oracleText: "When Touch the Spirit Realm enters, exile up to one target artifact or creature until Touch the Spirit Realm leaves the battlefield.",
			wantUpTo:   true,
		},
		{
			name:       "Angel of Sanctions",
			typeLine:   "Creature — Angel",
			manaCost:   "{3}{W}{W}",
			oracleText: "When Angel of Sanctions enters, you may exile target nonland permanent an opponent controls until Angel of Sanctions leaves the battlefield.",
			wantOpt:    true,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			power, toughness := "4", "4"
			card := &ScryfallCard{
				Name:       test.name,
				Layout:     "normal",
				ManaCost:   test.manaCost,
				TypeLine:   test.typeLine,
				OracleText: test.oracleText,
				Colors:     []string{"W"},
				Power:      &power,
				Toughness:  &toughness,
			}
			source, diagnostics, err := GenerateExecutableCardSource(card, "o")
			if err != nil {
				t.Fatal(err)
			}
			if len(diagnostics) != 0 {
				t.Fatalf("diagnostics = %#v", diagnostics)
			}
			wanted := []string{
				"Primitive: game.Exile",
				`ExileLinkedKey: game.LinkedKey("exile-until-leaves")`,
				"Primitive: game.PutOnBattlefield",
				`game.LinkedBattlefieldSource(game.LinkedKey("exile-until-leaves"))`,
				"game.EventZoneChanged",
			}
			if test.wantOpt {
				wanted = append(wanted, "Optional: true,")
			}
			if test.wantUpTo {
				wanted = append(wanted, "MinTargets: 0,", "MaxTargets: 1,")
			}
			for _, want := range wanted {
				if !strings.Contains(source, want) {
					t.Fatalf("source missing %q:\n%s", want, source)
				}
			}
			if !test.wantOpt && strings.Contains(source, "Optional: true,") {
				t.Fatalf("unexpected Optional on non-optional wording:\n%s", source)
			}
		})
	}
}
