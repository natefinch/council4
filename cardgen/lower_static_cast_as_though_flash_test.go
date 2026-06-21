package cardgen

import (
	"reflect"
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/types"
)

// TestLowerStaticCastAsThoughFlash proves the static "You may cast [<filter>]
// spells as though they had flash." lowers to a single
// RuleEffectCastSpellsAsThoughFlash scoped to the controller, carrying the
// optional card-type and subtype filters.
func TestLowerStaticCastAsThoughFlash(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name       string
		typeLine   string
		oracle     string
		spellTypes []types.Card
		subtypes   []types.Sub
	}{
		{
			name:     "all spells",
			typeLine: "Artifact",
			oracle:   "You may cast spells as though they had flash.",
		},
		{
			name:       "sorcery spells",
			typeLine:   "Creature — Dragon",
			oracle:     "You may cast sorcery spells as though they had flash.",
			spellTypes: []types.Card{types.Sorcery},
		},
		{
			name:     "aura and equipment spells",
			typeLine: "Artifact — Equipment",
			oracle:   "You may cast Aura and Equipment spells as though they had flash.",
			subtypes: []types.Sub{types.Sub("Aura"), types.Sub("Equipment")},
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			face := lowerSingleFace(t, &ScryfallCard{
				Name:       "Test Flash Static",
				Layout:     "normal",
				TypeLine:   tc.typeLine,
				OracleText: tc.oracle,
			})
			if len(face.StaticAbilities) != 1 {
				t.Fatalf("static abilities = %d, want 1", len(face.StaticAbilities))
			}
			effects := face.StaticAbilities[0].Body.RuleEffects
			if len(effects) != 1 {
				t.Fatalf("rule effects = %#v, want one", effects)
			}
			effect := effects[0]
			if effect.Kind != game.RuleEffectCastSpellsAsThoughFlash {
				t.Fatalf("kind = %v, want cast as though flash", effect.Kind)
			}
			if effect.AffectedPlayer != game.PlayerYou {
				t.Fatalf("affected player = %v, want you", effect.AffectedPlayer)
			}
			if len(tc.spellTypes) == 0 {
				if len(effect.SpellTypes) != 0 {
					t.Fatalf("spell types = %#v, want none", effect.SpellTypes)
				}
			} else if !reflect.DeepEqual(effect.SpellTypes, tc.spellTypes) {
				t.Fatalf("spell types = %#v, want %#v", effect.SpellTypes, tc.spellTypes)
			}
			if len(tc.subtypes) == 0 {
				if len(effect.SpellSubtypes) != 0 {
					t.Fatalf("subtypes = %#v, want none", effect.SpellSubtypes)
				}
			} else if !reflect.DeepEqual(effect.SpellSubtypes, tc.subtypes) {
				t.Fatalf("subtypes = %#v, want %#v", effect.SpellSubtypes, tc.subtypes)
			}
		})
	}
}
