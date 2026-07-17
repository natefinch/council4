package cardgen

import (
	"strings"
	"testing"
)

const furyStormOracle = "When you cast this spell, copy it for each time you've cast your commander from the command zone this game. You may choose new targets for the copies.\n" +
	"Copy target instant or sorcery spell. You may choose new targets for the copy."

func TestGenerateFuryStormSource(t *testing.T) {
	t.Parallel()
	source, diagnostics, err := GenerateExecutableCardSource(&ScryfallCard{
		Name:       "Fury Storm",
		Layout:     "normal",
		ManaCost:   "{2}{R}{R}",
		TypeLine:   "Instant",
		OracleText: furyStormOracle,
	}, "f")
	if err != nil {
		t.Fatal(err)
	}
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	for _, want := range []string{
		"SelfWasCast: true,",
		"Object: game.EventStackObjectReference(),",
		"DynamicCount: game.Dynamic(game.DynamicAmount{",
		"Kind: game.DynamicAmountCommanderCastCount,",
		"MayChooseNewTargets: true,",
		"Object: game.TargetStackObjectReference(0),",
		"SpellCardTypesAny: []types.Card{types.Instant, types.Sorcery},",
	} {
		if !strings.Contains(spaceCollapsed(source), spaceCollapsed(want)) {
			t.Fatalf("source missing %q:\n%s", want, source)
		}
	}
}
