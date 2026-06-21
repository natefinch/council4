package cardgen

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

// roundTripCards are representative cards exercising the mana, static, and spell
// ability categories through the full typed pipeline.
var roundTripCards = []*ScryfallCard{
	{
		Name:       "RT Bear",
		Layout:     "normal",
		TypeLine:   "Creature — Bear",
		ManaCost:   "{1}{G}",
		Colors:     []string{"G"},
		OracleText: "Flying\nVigilance",
		Power:      new("2"),
		Toughness:  new("2"),
	},
	{
		Name:       "RT Land",
		Layout:     "normal",
		TypeLine:   "Land",
		OracleText: "{T}: Add {G}.",
	},
	{
		Name:       "RT Bolt",
		Layout:     "normal",
		TypeLine:   "Instant",
		ManaCost:   "{R}",
		Colors:     []string{"R"},
		OracleText: "RT Bolt deals 3 damage to any target.",
	},
	{
		Name:       "RT Bog",
		Layout:     "normal",
		TypeLine:   "Land",
		OracleText: "When this land enters, exile target player's graveyard.",
	},
	{
		Name:       "RT Ozolith",
		Layout:     "normal",
		TypeLine:   "Legendary Artifact",
		OracleText: "Whenever a creature you control leaves the battlefield, if it had counters on it, put those counters on RT Ozolith.\nAt the beginning of combat on your turn, if RT Ozolith has counters on it, you may move all counters from RT Ozolith onto target creature.",
	},
	{
		Name:       "RT Nesting Grounds",
		Layout:     "normal",
		TypeLine:   "Land",
		OracleText: "{T}: Add {C}.\n{1}, {T}: Move a counter from target permanent you control onto a second target permanent. Activate only as a sorcery.",
	},
	{
		Name:       "RT Reaver Cleaver",
		Layout:     "normal",
		TypeLine:   "Legendary Artifact — Equipment",
		OracleText: "Equipped creature gets +1/+1 and has trample and \"Whenever this creature deals combat damage to a player or planeswalker, create that many Treasure tokens.\"\nEquip {3}",
	},
	{
		// Exercises MassReturnFromGraveyard, whose rendered Destination zone
		// literal requires the zone import (regression guard for #995).
		Name:       "RT Replenish",
		Layout:     "normal",
		TypeLine:   "Sorcery",
		ManaCost:   "{3}{W}",
		Colors:     []string{"W"},
		OracleText: "Return all enchantment cards from your graveyard to the battlefield.",
	},
	{
		Name:       "RT Brotherhood Regalia",
		Layout:     "normal",
		TypeLine:   "Artifact — Equipment",
		OracleText: "Equipped creature has ward {2}, is an Assassin in addition to its other types, and can't be blocked.\nEquip legendary creature {1}\nEquip {3}",
	},
}

// writeRoundTripPackage generates source for roundTripCards into a fresh package
// directory inside the module and returns the directory and package name.
func writeRoundTripPackage(t *testing.T) (dir, pkgName string) {
	t.Helper()
	suffix := filepath.Base(t.TempDir())
	dir = filepath.Join(".", "roundtrippkg"+suffix)
	if err := os.Mkdir(dir, 0o700); err != nil {
		t.Fatalf("Mkdir: %v", err)
	}
	t.Cleanup(func() {
		_ = os.RemoveAll(dir)
	})

	pkgName = filepath.Base(dir)
	for _, card := range roundTripCards {
		source, diagnostics, err := GenerateExecutableCardSource(card, pkgName)
		if err != nil {
			t.Fatalf("GenerateExecutableCardSource(%q): %v", card.Name, err)
		}
		if len(diagnostics) != 0 {
			t.Fatalf("GenerateExecutableCardSource(%q) diagnostics: %#v", card.Name, diagnostics)
		}
		file := filepath.Join(dir, CardNameToVarName(card.Name)+".go")
		if err := os.WriteFile(file, []byte(source), 0o600); err != nil {
			t.Fatalf("WriteFile: %v", err)
		}
	}
	return dir, pkgName
}

// TestRoundTripCompiles generates executable card source, writes it to a fresh
// package directory inside the module, and runs `go build` to verify the
// rendered output is valid, compilable Go.
func TestRoundTripCompiles(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping go build round-trip in short mode")
	}

	dir, _ := writeRoundTripPackage(t)
	cmd := exec.CommandContext(context.Background(), "go", "build", "./"+filepath.Base(dir))
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("go build failed: %v\n%s", err, out)
	}
}

// TestRoundTripSemantic generates executable card source, writes a semantic test
// alongside it in the same package, and runs `go test` so the generated vars are
// checked for the actual typed structure they must round-trip to — not merely
// that they compile.
func TestRoundTripSemantic(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping go test round-trip in short mode")
	}

	dir, pkgName := writeRoundTripPackage(t)
	testFile := filepath.Join(dir, "semantic_test.go")
	if err := os.WriteFile(testFile, []byte(semanticTestSource(pkgName)), 0o600); err != nil {
		t.Fatalf("WriteFile semantic test: %v", err)
	}

	cmd := exec.CommandContext(context.Background(), "go", "test", "-count=1", "./"+filepath.Base(dir))
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("go test failed: %v\n%s", err, out)
	}
}

// semanticTestSource returns the source of a test file, in package pkgName, that
// directly inspects the generated vars to confirm they round-trip to the
// expected typed structure.
func semanticTestSource(pkgName string) string {
	return fmt.Sprintf(`package %s

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/mana"
)

func TestRTLandSemantic(t *testing.T) {
	if RTLand.CardFace.Name != "RT Land" {
		t.Fatalf("name = %%q", RTLand.CardFace.Name)
	}
	if len(RTLand.CardFace.ManaAbilities) != 1 {
		t.Fatalf("mana abilities = %%d", len(RTLand.CardFace.ManaAbilities))
	}
	prim := RTLand.CardFace.ManaAbilities[0].Content.Modes[0].Sequence[0].Primitive
	add, ok := prim.(game.AddMana)
	if !ok {
		t.Fatalf("primitive type = %%T", prim)
	}
	if add.ManaColor != mana.G {
		t.Fatalf("mana color = %%q", add.ManaColor)
	}
}

func TestRTBearSemantic(t *testing.T) {
	if len(RTBear.CardFace.StaticAbilities) != 2 {
		t.Fatalf("static abilities = %%d", len(RTBear.CardFace.StaticAbilities))
	}
	keywords := RTBear.CardFace.StaticAbilities[0].KeywordAbilities
	if len(keywords) != 1 {
		t.Fatalf("keyword abilities = %%d", len(keywords))
	}
	keyword, ok := keywords[0].(game.SimpleKeyword)
	if !ok || keyword.Kind != game.Flying {
		t.Fatalf("keyword[0] = %%#v", keywords[0])
	}
}

func TestRTBoltSemantic(t *testing.T) {
	if !RTBolt.CardFace.SpellAbility.Exists {
		t.Fatal("spell ability missing")
	}
	mode := RTBolt.CardFace.SpellAbility.Val.Modes[0]
	if len(mode.Targets) != 1 {
		t.Fatalf("targets = %%d", len(mode.Targets))
	}
	damage, ok := mode.Sequence[0].Primitive.(game.Damage)
	if !ok {
		t.Fatalf("primitive type = %%T", mode.Sequence[0].Primitive)
	}
	if damage.Amount.Value() != 3 {
		t.Fatalf("damage amount = %%d", damage.Amount.Value())
	}
}
`, pkgName)
}
