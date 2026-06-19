package parser

import "testing"

// TestParsePlayerGraveyardExileExact recognizes the whole-graveyard exile
// "Exile target player's graveyard." (and its opponent variant) as exact and
// records the typed owner relation.
func TestParsePlayerGraveyardExileExact(t *testing.T) {
	t.Parallel()
	tests := []struct {
		source string
		owner  GraveyardZoneExileKind
	}{
		{"Exile target player's graveyard.", GraveyardZoneExileTargetPlayer},
		{"Exile target opponent's graveyard.", GraveyardZoneExileTargetOpponent},
	}
	for _, test := range tests {
		t.Run(test.source, func(t *testing.T) {
			t.Parallel()
			document, _ := Parse(test.source, Context{InstantOrSorcery: true})
			effects := document.Abilities[0].Sentences[0].Effects
			if len(effects) != 1 {
				t.Fatalf("effects = %#v, want one", effects)
			}
			if !effects[0].Exact {
				t.Fatalf("effect not exact: %#v", effects[0])
			}
			if effects[0].GraveyardZoneExile != test.owner {
				t.Fatalf("owner = %q, want %q", effects[0].GraveyardZoneExile, test.owner)
			}
		})
	}
}

// TestParsePlayerGraveyardExileFailsClosed documents that the referenced-player,
// each-player, all-graveyards, and single-card forms are not recognized as a
// whole-graveyard player-targeted exile.
func TestParsePlayerGraveyardExileFailsClosed(t *testing.T) {
	t.Parallel()
	for _, source := range []string{
		"Exile that player's graveyard.",
		"Exile each player's graveyard.",
		"Exile all graveyards.",
		"Exile target card from a graveyard.",
		"Exile up to two target cards from a single graveyard.",
	} {
		t.Run(source, func(t *testing.T) {
			t.Parallel()
			document, _ := Parse(source, Context{InstantOrSorcery: true})
			for _, sentence := range document.Abilities[0].Sentences {
				for i := range sentence.Effects {
					if sentence.Effects[i].GraveyardZoneExile != GraveyardZoneExileNone {
						t.Fatalf("effect %q recognized as player-graveyard exile: %#v", source, sentence.Effects[i])
					}
				}
			}
		})
	}
}
