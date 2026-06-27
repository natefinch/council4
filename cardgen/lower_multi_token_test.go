package cardgen

import (
	"reflect"
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/types"
)

// multiTokenCreateInstructions returns the CreateToken primitives lowered from a
// single multi-token "Create a X, a Y, and a Z." clause, asserting the spell
// ability lowered to exactly one mode whose sequence is entirely CreateToken
// instructions.
func multiTokenCreateInstructions(t *testing.T, face loweredFaceAbilities) []game.CreateToken {
	t.Helper()
	if !face.SpellAbility.Exists {
		t.Fatal("spell ability not lowered")
	}
	mode := face.SpellAbility.Val.Modes[0]
	creates := make([]game.CreateToken, 0, len(mode.Sequence))
	for i, inst := range mode.Sequence {
		create, ok := inst.Primitive.(game.CreateToken)
		if !ok {
			t.Fatalf("sequence[%d] = %T, want game.CreateToken", i, inst.Primitive)
		}
		creates = append(creates, create)
	}
	return creates
}

func multiTokenDef(t *testing.T, create game.CreateToken) *game.CardDef {
	t.Helper()
	if create.Amount.Value() != 1 {
		t.Fatalf("amount = %d, want 1 (each article creates exactly one token)", create.Amount.Value())
	}
	def, ok := create.Source.TokenDefRef()
	if !ok {
		t.Fatal("token source is not a token definition")
	}
	return def
}

// TestLowerMultiTokenDistinctCreatures proves a single clause that creates three
// distinct creature tokens ("Create a 1/1 green Snake creature token, a 2/2
// green Wolf creature token, and a 3/3 green Elephant creature token.", as on
// Bestial Menace) lowers to a sequence of three separate CreateToken
// instructions, each carrying its own token definition.
func TestLowerMultiTokenDistinctCreatures(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Bestial",
		Layout:     "normal",
		TypeLine:   "Sorcery",
		OracleText: "Create a 1/1 green Snake creature token, a 2/2 green Wolf creature token, and a 3/3 green Elephant creature token.",
		Colors:     []string{"G"},
	})
	creates := multiTokenCreateInstructions(t, face)
	if len(creates) != 3 {
		t.Fatalf("sequence length = %d, want 3", len(creates))
	}
	want := []struct {
		name string
		sub  types.Sub
		pt   int
	}{
		{"Snake", types.Snake, 1},
		{"Wolf", types.Wolf, 2},
		{"Elephant", types.Elephant, 3},
	}
	for i, w := range want {
		def := multiTokenDef(t, creates[i])
		if def.Name != w.name {
			t.Errorf("token[%d] name = %q, want %q", i, def.Name, w.name)
		}
		if !def.Power.Exists || def.Power.Val.Value != w.pt ||
			!def.Toughness.Exists || def.Toughness.Val.Value != w.pt {
			t.Errorf("token[%d] PT = %+v/%+v, want %d/%d", i, def.Power, def.Toughness, w.pt, w.pt)
		}
		if len(def.Subtypes) != 1 || def.Subtypes[0] != w.sub {
			t.Errorf("token[%d] subtypes = %v, want [%v]", i, def.Subtypes, w.sub)
		}
		if len(def.Colors) != 1 || def.Colors[0] != color.Green {
			t.Errorf("token[%d] colors = %v, want [Green]", i, def.Colors)
		}
	}
}

// TestLowerMultiTokenPerTokenKeyword proves a per-token "with <keyword>" rider
// attaches only to the token it modifies. Forbidden Friendship's "Create a 1/1
// red Dinosaur creature token with haste and a 1/1 white Human Soldier creature
// token." grants haste to the Dinosaur and nothing to the Human Soldier.
func TestLowerMultiTokenPerTokenKeyword(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Friendship",
		Layout:     "normal",
		TypeLine:   "Sorcery",
		OracleText: "Create a 1/1 red Dinosaur creature token with haste and a 1/1 white Human Soldier creature token.",
		Colors:     []string{"R"},
	})
	creates := multiTokenCreateInstructions(t, face)
	if len(creates) != 2 {
		t.Fatalf("sequence length = %d, want 2", len(creates))
	}
	dino := multiTokenDef(t, creates[0])
	if dino.Name != "Dinosaur" {
		t.Fatalf("token[0] name = %q, want Dinosaur", dino.Name)
	}
	if len(dino.StaticAbilities) != 1 || !reflect.DeepEqual(dino.StaticAbilities[0], game.HasteStaticBody) {
		t.Fatalf("token[0] statics = %v, want [Haste]", dino.StaticAbilities)
	}
	human := multiTokenDef(t, creates[1])
	if human.Name != "Human Soldier" {
		t.Fatalf("token[1] name = %q, want Human Soldier", human.Name)
	}
	if len(human.StaticAbilities) != 0 {
		t.Fatalf("token[1] statics = %v, want none", human.StaticAbilities)
	}
}

// TestLowerMultiTokenRendererCollisionFailsClosed proves the fail-closed guard
// against the renderer's tokenDefKey limitation. Wurmcoil Engine creates two
// "Phyrexian Wurm" tokens that share name, colorless color, and 3/3 P/T but
// differ in keywords (deathtouch vs. lifelink). Because tokenDefKey omits static
// abilities, those two definitions would collapse to a single var and render
// identically, so lowering must fail closed rather than emit a wrong card.
func TestLowerMultiTokenRendererCollisionFailsClosed(t *testing.T) {
	t.Parallel()
	face := lowerSingleFaceExpectingUnsupported(t, &ScryfallCard{
		Name:       "Test Wurmcoil",
		Layout:     "normal",
		TypeLine:   "Sorcery",
		OracleText: "Create a 3/3 colorless Phyrexian Wurm artifact creature token with deathtouch and a 3/3 colorless Phyrexian Wurm artifact creature token with lifelink.",
	})
	if face.SpellAbility.Exists {
		t.Fatal("spell ability lowered, want fail-closed for colliding token defs")
	}
}
