package cardgen

import (
	"reflect"
	"strings"
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/types"
)

// TestLowerMassRegenerateGroup confirms the "Regenerate all/each <group>." mass
// forms lower to a single battlefield-group game.Regenerate, the same machinery
// the destroy/exile/tap/untap mass forms share.
func TestLowerMassRegenerateGroup(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name       string
		oracleText string
		selection  game.Selection
	}{
		{
			name:       "all creatures you control",
			oracleText: "Regenerate all creatures you control.",
			selection: game.Selection{
				RequiredTypes: []types.Card{types.Creature},
				Controller:    game.ControllerYou,
			},
		},
		{
			name:       "each creature you control",
			oracleText: "Regenerate each creature you control.",
			selection: game.Selection{
				RequiredTypes: []types.Card{types.Creature},
				Controller:    game.ControllerYou,
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			face := lowerSingleFace(t, &ScryfallCard{
				Name:       "Test Mass Regenerate",
				Layout:     "normal",
				TypeLine:   "Instant",
				OracleText: test.oracleText,
			})
			primitive := face.SpellAbility.Val.Modes[0].Sequence[0].Primitive
			regenerate, ok := primitive.(game.Regenerate)
			if !ok {
				t.Fatalf("primitive = %T, want game.Regenerate", primitive)
			}
			if regenerate.Group.Domain() != game.GroupDomainBattlefield {
				t.Fatalf("group domain = %v, want battlefield", regenerate.Group.Domain())
			}
			if selection := regenerate.Group.Selection(); !reflect.DeepEqual(selection, test.selection) {
				t.Fatalf("selection = %#v, want %#v", selection, test.selection)
			}
		})
	}
}

func TestGenerateRegenerateRecipients(t *testing.T) {
	cases := []struct {
		name     string
		typeLine string
		text     string
		wantRef  string
	}{
		{"Self This Creature", "Creature — Beast", "{G}: Regenerate this creature.", "game.SourcePermanentReference()"},
		{"Vorthos Troll", "Legendary Creature — Troll", "{G}: Regenerate Vorthos Troll.", "game.SourcePermanentReference()"},
		{"Mantle Cloak", "Enchantment — Aura", "Enchant creature\n{1}: Regenerate enchanted creature.", "game.SourceAttachedPermanentReference()"},
		{"Plate Guard", "Artifact — Equipment", "{2}: Regenerate equipped creature.\nEquip {3}", "game.SourceAttachedPermanentReference()"},
		{"Ward Spell", "Instant", "Regenerate target creature.", "game.TargetPermanentReference(0)"},
		{"Counter Troll", "Creature — Troll", "{1}, Remove a +1/+1 counter from this creature: Regenerate this creature.", "game.SourcePermanentReference()"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			card := &ScryfallCard{Name: c.name, Layout: "normal", TypeLine: c.typeLine, OracleText: c.text}
			source, diags, err := GenerateExecutableCardSource(card, "z")
			if err != nil {
				t.Fatal(err)
			}
			if len(diags) != 0 {
				t.Fatalf("diags = %#v", diags)
			}
			if !strings.Contains(source, "game.Regenerate{") {
				t.Fatalf("missing Regenerate primitive:\n%s", source)
			}
			if !strings.Contains(source, c.wantRef) {
				t.Fatalf("missing object ref %q:\n%s", c.wantRef, source)
			}
		})
	}
}
