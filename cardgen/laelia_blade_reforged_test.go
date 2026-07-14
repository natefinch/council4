package cardgen

import (
	"strings"
	"testing"
)

// TestGenerateLaeliaMultiOriginExileTrigger covers the multi-origin zone-change
// trigger "Whenever one or more cards are put into exile from your library
// and/or your graveyard, ..." (Laelia, the Blade Reforged). The origin union
// lowers to a FromZones set composed with the single exile destination and the
// one-or-more batch coalescing, while the attack and counter bodies lower
// through the existing ImpulseExile and AddCounter components.
func TestGenerateLaeliaMultiOriginExileTrigger(t *testing.T) {
	t.Parallel()
	source, diagnostics, err := GenerateExecutableCardSource(&ScryfallCard{
		Name:      "Laelia, the Blade Reforged",
		Layout:    "normal",
		TypeLine:  "Legendary Creature — Spirit Soldier",
		ManaCost:  "{2}{R}",
		Colors:    []string{"R"},
		Power:     new("2"),
		Toughness: new("2"),
		OracleText: "Haste\n" +
			"Whenever Laelia, the Blade Reforged attacks, exile the top card of your library. You may play that card this turn.\n" +
			"Whenever one or more cards are put into exile from your library and/or your graveyard, put a +1/+1 counter on Laelia, the Blade Reforged.",
	}, "l")
	if err != nil {
		t.Fatal(err)
	}
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	for _, want := range []string{
		"game.HasteStaticBody,",
		"Event: game.EventAttackerDeclared,",
		"Primitive: game.ImpulseExile{",
		"Event: game.EventZoneChanged,",
		"Player: game.TriggerPlayerYou,",
		"FromZones: []zone.Type{zone.Library, zone.Graveyard},",
		"MatchToZone: true,",
		"ToZone: zone.Exile,",
		"OneOrMore: true,",
		"SubjectSelection: game.Selection{NonToken: true},",
		"Primitive: game.AddCounter{",
		"CounterKind: counter.PlusOnePlusOne,",
	} {
		if !strings.Contains(spaceCollapsed(source), spaceCollapsed(want)) {
			t.Fatalf("source missing %q:\n%s", want, source)
		}
	}
	// The multi-origin union must not also emit a single from-zone filter.
	if strings.Contains(spaceCollapsed(source), spaceCollapsed("MatchFromZone: true,")) {
		t.Fatalf("multi-origin trigger must not set MatchFromZone:\n%s", source)
	}
}
