package parser

import (
	"slices"
	"testing"
)

func keepOnePerTypeEffect(t *testing.T, text string) *EffectSyntax {
	t.Helper()
	document, diagnostics := Parse(text, Context{})
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v, want none", diagnostics)
	}
	if len(document.Abilities) != 1 || len(document.Abilities[0].Sentences) != 1 {
		t.Fatalf("abilities = %#v, want one ability with one sentence", document.Abilities)
	}
	effects := document.Abilities[0].Sentences[0].Effects
	if len(effects) != 1 {
		t.Fatalf("effects = %#v, want one effect", effects)
	}
	return &effects[0]
}

// TestParseControllerChoosesKeepSequence proves the parser folds the two-sentence
// controller-chooses keep-one-per-type form (Tragic Arrogance) into the generic
// KeepOnePerType payload: the effect's controller chooses for every player, the
// listed permanent types are recorded in order, and only nonland permanents are
// sacrificed. The zero-effect "you choose" prelude is credited so the card is
// fully covered rather than left with an unrecognized sibling.
func TestParseControllerChoosesKeepSequence(t *testing.T) {
	text := "For each player, you choose from among the permanents that player controls an artifact, a creature, an enchantment, and a planeswalker. Then each player sacrifices all other nonland permanents they control."
	document, diagnostics := Parse(text, Context{})
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v, want none", diagnostics)
	}
	if len(document.Abilities) != 1 || len(document.Abilities[0].Sentences) != 2 {
		t.Fatalf("abilities = %#v, want one ability with two sentences", document.Abilities)
	}
	sentences := document.Abilities[0].Sentences
	if !sentences[0].ControllerChoosesKeepPrelude {
		t.Error("prelude sentence ControllerChoosesKeepPrelude = false, want true (credited)")
	}
	if len(sentences[1].Effects) != 1 {
		t.Fatalf("sacrifice sentence effects = %#v, want one", sentences[1].Effects)
	}
	effect := sentences[1].Effects[0]
	if !effect.Exact {
		t.Fatal("sacrifice effect.Exact = false, want true")
	}
	keep := effect.KeepOnePerType
	if keep == nil {
		t.Fatal("effect.KeepOnePerType = nil, want payload")
	}
	if keep.Scope != KeepScopeAllPlayers {
		t.Errorf("scope = %v, want KeepScopeAllPlayers", keep.Scope)
	}
	if !keep.ControllerChoosesForAll {
		t.Error("controllerChoosesForAll = false, want true")
	}
	if !keep.NonlandOnly {
		t.Error("nonlandOnly = false, want true")
	}
	want := []CardType{CardTypeArtifact, CardTypeCreature, CardTypeEnchantment, CardTypePlaneswalker}
	if !slices.Equal(keep.Types, want) {
		t.Errorf("types = %v, want %v", keep.Types, want)
	}
}

// TestParseControllerChoosesKeepFailsClosed proves the two-sentence controller-
// chooses form is verbatim: any deviation — an each-player choose verb, a
// non-permanent type, an altered sacrifice pool, or a missing follow-up sentence
// — leaves the sacrifice effect unrecognized so no card with different rules is
// lowered as the controller-chooses family.
func TestParseControllerChoosesKeepFailsClosed(t *testing.T) {
	cases := map[string]string{
		"each player chooses, not controller": "For each player, each player chooses from among the permanents that player controls an artifact, a creature, an enchantment, and a planeswalker. Then each player sacrifices all other nonland permanents they control.",
		"non-permanent type in prelude":       "For each player, you choose from among the permanents that player controls an artifact, an instant, an enchantment, and a planeswalker. Then each player sacrifices all other nonland permanents they control.",
		"sacrifice pool not nonland":          "For each player, you choose from among the permanents that player controls an artifact, a creature, an enchantment, and a planeswalker. Then each player sacrifices the rest.",
		"prelude without sacrifice follow-up": "For each player, you choose from among the permanents that player controls an artifact, a creature, an enchantment, and a planeswalker.",
	}
	for name, text := range cases {
		t.Run(name, func(t *testing.T) {
			document, _ := Parse(text, Context{})
			for _, ability := range document.Abilities {
				for _, sentence := range ability.Sentences {
					for i := range sentence.Effects {
						if sentence.Effects[i].KeepOnePerType != nil && sentence.Effects[i].KeepOnePerType.ControllerChoosesForAll {
							t.Fatal("altered wording was recognized as controller-chooses keep-one-per-type, want fail-closed")
						}
					}
				}
			}
		})
	}
}

// TestParseKeepOnePerTypeSacrifice proves the parser recognizes each printed
// member of the "keep one of each type" sacrifice family and records the exact
// affected scope, ordered permanent types, and nonland-pool flag, so the
// compiler and lowerer can build the generic KeepOnePerType primitive.
func TestParseKeepOnePerTypeSacrifice(t *testing.T) {
	cases := map[string]struct {
		text        string
		scope       KeepScope
		types       []CardType
		nonlandOnly bool
	}{
		"each permanent type": {
			text:  "Each opponent chooses a permanent they control of each permanent type and sacrifices the rest.",
			scope: KeepScopeOpponents,
			types: []CardType{CardTypeArtifact, CardTypeBattle, CardTypeCreature, CardTypeEnchantment, CardTypeLand, CardTypePlaneswalker},
		},
		"listed types, all permanents": {
			text:  "Each player chooses from among the permanents they control an artifact, a creature, an enchantment, and a land, then sacrifices the rest.",
			scope: KeepScopeAllPlayers,
			types: []CardType{CardTypeArtifact, CardTypeCreature, CardTypeEnchantment, CardTypeLand},
		},
		"listed types, nonland permanents": {
			text:        "When this creature enters, each player chooses an artifact, a creature, an enchantment, and a planeswalker from among the nonland permanents they control, then sacrifices the rest.",
			scope:       KeepScopeAllPlayers,
			types:       []CardType{CardTypeArtifact, CardTypeCreature, CardTypeEnchantment, CardTypePlaneswalker},
			nonlandOnly: true,
		},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			effect := keepOnePerTypeEffect(t, tc.text)
			if !effect.Exact {
				t.Fatal("effect.Exact = false, want true")
			}
			keep := effect.KeepOnePerType
			if keep == nil {
				t.Fatal("effect.KeepOnePerType = nil, want payload")
			}
			if keep.Scope != tc.scope {
				t.Errorf("scope = %v, want %v", keep.Scope, tc.scope)
			}
			if keep.NonlandOnly != tc.nonlandOnly {
				t.Errorf("nonlandOnly = %v, want %v", keep.NonlandOnly, tc.nonlandOnly)
			}
			if !slices.Equal(keep.Types, tc.types) {
				t.Errorf("types = %v, want %v", keep.Types, tc.types)
			}
		})
	}
}

// TestParseKeepOnePerTypeFailsClosed proves the recognizer is verbatim: any
// deviation from the printed wording — a non-permanent type, a dropped Oxford
// comma, an altered sacrifice object, or a wrong article — leaves the sentence
// unrecognized so no card with different rules is silently lowered as the family.
func TestParseKeepOnePerTypeFailsClosed(t *testing.T) {
	cases := map[string]string{
		"non-permanent type in list":  "Each player chooses from among the permanents they control an artifact, an instant, an enchantment, and a land, then sacrifices the rest.",
		"dropped oxford comma":        "Each player chooses from among the permanents they control an artifact, a creature, an enchantment and a land, then sacrifices the rest.",
		"wrong article":               "Each player chooses from among the permanents they control a artifact, a creature, an enchantment, and a land, then sacrifices the rest.",
		"sacrifices a single instead": "Each opponent chooses a permanent they control of each permanent type and sacrifices a permanent.",
		"keeps two instead of one":    "Each opponent chooses two permanents they control of each permanent type and sacrifices the rest.",
	}
	for name, text := range cases {
		t.Run(name, func(t *testing.T) {
			document, _ := Parse(text, Context{})
			for _, ability := range document.Abilities {
				for _, sentence := range ability.Sentences {
					for i := range sentence.Effects {
						if sentence.Effects[i].KeepOnePerType != nil {
							t.Fatal("altered wording was recognized as keep-one-per-type, want fail-closed")
						}
					}
				}
			}
		})
	}
}
