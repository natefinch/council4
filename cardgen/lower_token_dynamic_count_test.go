package cardgen

import (
	"strings"
	"testing"

	"github.com/natefinch/council4/mtg/game"
)

// TestLowerNamedTokenEqualToCount verifies that a predefined/named artifact
// token ("a number of Food tokens equal to ...") accepts the same dynamic
// "equal to" count form already accepted by the creature-token path. It backs
// Killer Service and Gluttonous Troll, whose ETB creates Food equal to the
// number of opponents you have.
func TestLowerNamedTokenEqualToCount(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Named Equal Count",
		Layout:     "normal",
		TypeLine:   "Sorcery",
		OracleText: "Create a number of Food tokens equal to the number of opponents you have.",
	})
	create := createTokenPrimitive(t, face)
	def, ok := create.Source.TokenDefRef()
	if !ok {
		t.Fatal("token source is not a token definition")
	}
	if def.Name != "Food" {
		t.Fatalf("token name = %q, want Food", def.Name)
	}
	if !create.Amount.IsDynamic() ||
		create.Amount.DynamicAmount().Val.Kind != game.DynamicAmountOpponentCount {
		t.Fatalf("amount = %+v, want a dynamic opponent count", create.Amount)
	}
}

// TestGenerateCreateThatManyFromCounterTrigger verifies "create that many"
// resolves to the triggering event's measured quantity (here, the number of
// -1/-1 counters placed) rather than only combat damage. It backs Nest of
// Scarabs.
func TestGenerateCreateThatManyFromCounterTrigger(t *testing.T) {
	t.Parallel()
	source, diagnostics, err := GenerateExecutableCardSource(&ScryfallCard{
		Name:       "Test That Many Counters",
		Layout:     "normal",
		ManaCost:   "{2}",
		TypeLine:   "Artifact",
		OracleText: "Whenever you put one or more -1/-1 counters on a creature, create that many 1/1 black Insect creature tokens.",
	}, "t")
	if err != nil {
		t.Fatal(err)
	}
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	if !strings.Contains(source, "game.CreateToken{") {
		t.Fatalf("source missing CreateToken primitive:\n%s", source)
	}
	if !strings.Contains(source, "game.DynamicAmount") {
		t.Fatalf("source missing dynamic create amount:\n%s", source)
	}
}

// TestGenerateCreateThatManyFromDiscardTrigger verifies "create that many"
// resolves to the number of cards discarded. It backs Cryptcaller Chariot.
func TestGenerateCreateThatManyFromDiscardTrigger(t *testing.T) {
	t.Parallel()
	source, diagnostics, err := GenerateExecutableCardSource(&ScryfallCard{
		Name:       "Test That Many Discard",
		Layout:     "normal",
		ManaCost:   "{2}{B}",
		TypeLine:   "Artifact",
		OracleText: "Whenever you discard one or more cards, create that many tapped 2/2 black Zombie creature tokens.",
	}, "t")
	if err != nil {
		t.Fatal(err)
	}
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	if !strings.Contains(source, "game.CreateToken{") {
		t.Fatalf("source missing CreateToken primitive:\n%s", source)
	}
}
