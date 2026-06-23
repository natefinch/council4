package cardgen

import (
	"strings"
	"testing"
)

// TestGenerateExecutableCardSourceYenna exercises the three Yenna, Redtooth
// Regent features together: a same-name target restriction ("target enchantment
// you control that doesn't have the same name as another permanent you
// control"), a separate-sentence copy of that target ("Create a token that's a
// copy of it"), and a resolving conditional gating later effects on the created
// token's subtype ("If the token is an Aura, untap Yenna, then scry 2").
func TestGenerateExecutableCardSourceYenna(t *testing.T) {
	t.Parallel()
	source, diagnostics, err := GenerateExecutableCardSource(&ScryfallCard{
		Name:       "Yenna, Redtooth Regent",
		Layout:     "normal",
		ManaCost:   "{2}{G}{W}",
		TypeLine:   "Legendary Creature — Elf Noble",
		Power:      new("4"),
		Toughness:  new("4"),
		OracleText: "{2}, {T}: Choose target enchantment you control that doesn't have the same name as another permanent you control. Create a token that's a copy of it, except it isn't legendary. If the token is an Aura, untap Yenna, then scry 2. Activate only as a sorcery.",
		Colors:     []string{"G", "W"},
	}, "y")
	if err != nil {
		t.Fatal(err)
	}
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	for _, wanted := range []string{
		"Timing:          game.SorceryOnly,",
		"PermanentTypes:            []types.Card{types.Enchantment},",
		"Controller:                game.ControllerYou,",
		"NameUniqueAmongControlled: true,",
		"Primitive: game.CreateToken{",
		"Source: game.TokenCopyOf(game.TokenCopySpec{",
		"Object:          game.TargetPermanentReference(0),",
		"SetNotLegendary: true,",
		"PublishLinked: game.LinkedKey(\"created-token\"),",
		"Primitive: game.Untap{",
		"Object: game.SourcePermanentReference(),",
		"Object:        opt.Val(game.LinkedObjectReference(\"created-token\")),",
		"ObjectMatches: opt.Val(game.Selection{SubtypesAny: []types.Sub{types.Sub(\"Aura\")}}),",
		"Primitive: game.Scry{",
	} {
		if !strings.Contains(source, wanted) {
			t.Fatalf("source missing %q:\n%s", wanted, source)
		}
	}
}
