package cardgen

import (
	"reflect"
	"strings"
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/types"
)

func TestLowerMultiColorMultiSubtypeToken(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Multi",
		Layout:     "normal",
		TypeLine:   "Sorcery",
		OracleText: "Create a 1/1 blue and white Bird creature token.",
		Colors:     []string{"W", "U"},
	})
	create, ok := face.SpellAbility.Val.Modes[0].Sequence[0].Primitive.(game.CreateToken)
	if !ok {
		t.Fatalf("primitive = %T, want game.CreateToken", face.SpellAbility.Val.Modes[0].Sequence[0].Primitive)
	}
	def, _ := create.Source.TokenDefRef()
	if len(def.Colors) != 2 || def.Colors[0] != color.Blue || def.Colors[1] != color.White {
		t.Fatalf("token colors = %v, want [Blue White]", def.Colors)
	}
	if def.Name != "Bird" {
		t.Fatalf("token name = %q, want Bird", def.Name)
	}
}

func TestLowerMultipleCreatureTokens(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Tokens",
		Layout:     "normal",
		TypeLine:   "Sorcery",
		OracleText: "Create two 1/1 white Soldier creature tokens.",
		Colors:     []string{"W"},
	})
	if !face.SpellAbility.Exists {
		t.Fatal("spell ability not lowered")
	}
	create, ok := face.SpellAbility.Val.Modes[0].Sequence[0].Primitive.(game.CreateToken)
	if !ok {
		t.Fatalf("primitive = %T, want game.CreateToken", face.SpellAbility.Val.Modes[0].Sequence[0].Primitive)
	}
	if create.Amount.Value() != 2 {
		t.Fatalf("amount = %d, want 2", create.Amount.Value())
	}
}

func TestLowerSingleCreatureToken(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Token",
		Layout:     "normal",
		TypeLine:   "Sorcery",
		OracleText: "Create a 1/1 white Soldier creature token.",
		Colors:     []string{"W"},
	})
	if !face.SpellAbility.Exists {
		t.Fatal("spell ability not lowered")
	}
	mode := face.SpellAbility.Val.Modes[0]
	if len(mode.Sequence) != 1 {
		t.Fatalf("sequence = %+v, want one instruction", mode.Sequence)
	}
	create, ok := mode.Sequence[0].Primitive.(game.CreateToken)
	if !ok {
		t.Fatalf("primitive = %T, want game.CreateToken", mode.Sequence[0].Primitive)
	}
	if create.Amount.Value() != 1 {
		t.Fatalf("amount = %d, want 1", create.Amount.Value())
	}
	def, ok := create.Source.TokenDefRef()
	if !ok {
		t.Fatal("token source is not a token definition")
	}
	if def.Name != "Soldier" ||
		!def.Power.Exists || def.Power.Val.Value != 1 ||
		!def.Toughness.Exists || def.Toughness.Val.Value != 1 {
		t.Fatalf("token def = %+v, want 1/1 Soldier", def.CardFace)
	}
	if len(def.Types) != 1 || def.Types[0] != types.Creature {
		t.Fatalf("token types = %v, want [Creature]", def.Types)
	}
	if len(def.Subtypes) != 1 || def.Subtypes[0] != types.Soldier {
		t.Fatalf("token subtypes = %v, want [Soldier]", def.Subtypes)
	}
	if len(def.Colors) != 1 || def.Colors[0] != color.White {
		t.Fatalf("token colors = %v, want [White]", def.Colors)
	}
}

func TestGenerateExecutableCardSourceTokenReferencedControllerRecipient(t *testing.T) {
	t.Parallel()
	source, diagnostics, err := GenerateExecutableCardSource(&ScryfallCard{
		Name:       "Test Within",
		Layout:     "normal",
		ManaCost:   "{2}{G}",
		TypeLine:   "Instant",
		OracleText: "Destroy target permanent. Its controller creates a 3/3 green Beast creature token.",
		Colors:     []string{"G"},
	}, "t")
	if err != nil {
		t.Fatal(err)
	}
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	for _, wanted := range []string{
		"Primitive: game.Destroy{",
		"Primitive: game.CreateToken{",
		"Recipient: opt.Val(game.ObjectControllerReference(game.TargetPermanentReference(0))),",
		`Name:      "Beast",`,
	} {
		if !strings.Contains(source, wanted) {
			t.Fatalf("source missing %q:\n%s", wanted, source)
		}
	}
}

func TestGenerateExecutableCardSourceTokenReferencedControllerRebased(t *testing.T) {
	t.Parallel()
	// The recipient must point at the destroyed permanent (game target 1), not the
	// tapped creature (game target 0): the antecedent index is rebased.
	source, diagnostics, err := GenerateExecutableCardSource(&ScryfallCard{
		Name:       "Test Rebase",
		Layout:     "normal",
		ManaCost:   "{2}{G}",
		TypeLine:   "Instant",
		OracleText: "Tap target creature. Destroy target permanent. Its controller creates a 3/3 green Beast creature token.",
		Colors:     []string{"G"},
	}, "t")
	if err != nil {
		t.Fatal(err)
	}
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	if !strings.Contains(source, "Recipient: opt.Val(game.ObjectControllerReference(game.TargetPermanentReference(1))),") {
		t.Fatalf("recipient not rebased to target 1:\n%s", source)
	}
}

func TestGenerateExecutableCardSourceCreatureTokenCompiles(t *testing.T) {
	t.Parallel()
	source, diagnostics, err := GenerateExecutableCardSource(&ScryfallCard{
		Name:       "Test Token",
		Layout:     "normal",
		ManaCost:   "{1}{W}",
		TypeLine:   "Sorcery",
		OracleText: "Create a 2/2 green Bear creature token.",
		Colors:     []string{"G"},
	}, "t")
	if err != nil {
		t.Fatal(err)
	}
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	for _, wanted := range []string{
		"Primitive: game.CreateToken{",
		"Source: game.TokenDef(testTokenToken)",
		"var testTokenToken = &game.CardDef{",
		`Name:      "Bear",`,
		"Subtypes:  []types.Sub{types.Bear},",
		"Power:     opt.Val(game.PT{Value: 2}),",
	} {
		if !strings.Contains(source, wanted) {
			t.Fatalf("source missing %q:\n%s", wanted, source)
		}
	}
}

func TestLowerCreatureTokenWithKeyword(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Keyword Token",
		Layout:     "normal",
		TypeLine:   "Sorcery",
		OracleText: "Create a 4/4 red Dragon creature token with flying.",
		Colors:     []string{"R"},
	})
	if !face.SpellAbility.Exists {
		t.Fatal("spell ability not lowered")
	}
	create, ok := face.SpellAbility.Val.Modes[0].Sequence[0].Primitive.(game.CreateToken)
	if !ok {
		t.Fatalf("primitive = %T, want game.CreateToken", face.SpellAbility.Val.Modes[0].Sequence[0].Primitive)
	}
	def, ok := create.Source.TokenDefRef()
	if !ok {
		t.Fatal("token source is not a token definition")
	}
	if len(def.StaticAbilities) != 1 {
		t.Fatalf("token static abilities = %v, want one (flying)", def.StaticAbilities)
	}
	if !reflect.DeepEqual(def.StaticAbilities[0], game.FlyingStaticBody) {
		t.Fatalf("token static ability = %+v, want game.FlyingStaticBody", def.StaticAbilities[0])
	}
}

func TestGenerateExecutableCardSourceKeywordTokenCompiles(t *testing.T) {
	t.Parallel()
	source, diagnostics, err := GenerateExecutableCardSource(&ScryfallCard{
		Name:       "Test Keyword Token",
		Layout:     "normal",
		ManaCost:   "{1}{R}",
		TypeLine:   "Sorcery",
		OracleText: "Create a 4/4 red Dragon creature token with flying.",
		Colors:     []string{"R"},
	}, "t")
	if err != nil {
		t.Fatal(err)
	}
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	for _, wanted := range []string{
		"StaticAbilities: []game.StaticAbility{",
		"game.FlyingStaticBody,",
	} {
		if !strings.Contains(source, wanted) {
			t.Fatalf("source missing %q:\n%s", wanted, source)
		}
	}
}

func TestCreateTokenFailsClosedForUnsupportedShapes(t *testing.T) {
	t.Parallel()
	for _, oracle := range []string{
		"Create a Treasure token.", // named, no P/T
		"Create a 1/1 white Soldier creature token with flying and vigilance.", // multiple keywords
	} {
		_, diagnostics, err := GenerateExecutableCardSource(&ScryfallCard{
			Name:       "Test Token",
			Layout:     "normal",
			TypeLine:   "Sorcery",
			OracleText: oracle,
		}, "t")
		if err != nil {
			t.Fatalf("%q: %v", oracle, err)
		}
		if len(diagnostics) == 0 {
			t.Fatalf("%q: expected fail-closed, got supported", oracle)
		}
	}
}
