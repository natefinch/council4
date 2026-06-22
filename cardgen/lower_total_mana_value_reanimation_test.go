package cardgen

import (
	goparser "go/parser"
	"go/token"
	"slices"
	"strings"
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
)

func TestLowerTotalManaValueGraveyardReanimation(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Dirge",
		Layout:     "normal",
		TypeLine:   "Sorcery",
		OracleText: "Return up to two creature cards with total mana value 4 or less from your graveyard to the battlefield.",
	})
	mode := face.SpellAbility.Val.Modes[0]
	if len(mode.Targets) != 0 {
		t.Fatalf("targets = %#v; want none (choose at resolution)", mode.Targets)
	}
	if len(mode.Sequence) != 1 {
		t.Fatalf("sequence = %#v; want one instruction", mode.Sequence)
	}
	prim, ok := mode.Sequence[0].Primitive.(game.ReturnFromGraveyard)
	if !ok {
		t.Fatalf("primitive = %T; want game.ReturnFromGraveyard", mode.Sequence[0].Primitive)
	}
	if prim.Player != game.ControllerReference() {
		t.Fatalf("player = %#v; want controller", prim.Player)
	}
	if prim.Destination != zone.Battlefield {
		t.Fatalf("destination = %v; want battlefield", prim.Destination)
	}
	if prim.Amount.Value() != 2 {
		t.Fatalf("amount = %#v; want fixed 2", prim.Amount)
	}
	if !prim.MaxTotalManaValue.Exists || prim.MaxTotalManaValue.Val != 4 {
		t.Fatalf("max total mana value = %#v; want 4", prim.MaxTotalManaValue)
	}
	if prim.EntryTapped {
		t.Fatal("entry tapped = true; want false")
	}
	if prim.Selection.ManaValue.Exists {
		t.Fatalf("selection carries a per-card mana value bound %#v; total cap must not lower to a per-card filter", prim.Selection.ManaValue)
	}
	if !slices.Equal(prim.Selection.RequiredTypes, []types.Card{types.Creature}) ||
		prim.Selection.Controller != game.ControllerYou {
		t.Fatalf("selection = %#v; want your creature cards", prim.Selection)
	}
	if err := game.ValidateInstructionSequence(mode.Sequence, mode.Targets); err != nil {
		t.Fatalf("invalid instruction sequence: %v", err)
	}
}

func TestLowerTotalManaValueGraveyardReanimationTapped(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Tapped Dirge",
		Layout:     "normal",
		TypeLine:   "Sorcery",
		OracleText: "Return up to three creature cards with total mana value 6 or less from your graveyard to the battlefield tapped.",
	})
	prim, ok := face.SpellAbility.Val.Modes[0].Sequence[0].Primitive.(game.ReturnFromGraveyard)
	if !ok {
		t.Fatalf("primitive = %T; want game.ReturnFromGraveyard", face.SpellAbility.Val.Modes[0].Sequence[0].Primitive)
	}
	if !prim.EntryTapped {
		t.Fatal("entry tapped = false; want true")
	}
	if prim.Amount.Value() != 3 || !prim.MaxTotalManaValue.Exists || prim.MaxTotalManaValue.Val != 6 {
		t.Fatalf("amount/cap = %#v / %#v; want 3 and 6", prim.Amount, prim.MaxTotalManaValue)
	}
}

func TestTotalManaValueReanimationRejectsTargeted(t *testing.T) {
	t.Parallel()
	// A targeted form ("up to two target creature cards ... with total mana
	// value ...") is not modeled by this non-target reanimation path and must
	// not generate a card that silently drops the set-sum constraint.
	_, diagnostics, err := GenerateExecutableCardSource(&ScryfallCard{
		Name:       "Test Targeted Total",
		Layout:     "normal",
		TypeLine:   "Sorcery",
		OracleText: "Return up to two target creature cards with total mana value 4 or less from your graveyard to the battlefield.",
	}, "t")
	if err != nil {
		t.Fatal(err)
	}
	if len(diagnostics) == 0 {
		t.Fatal("diagnostics = none; want the targeted total-mana-value return to fail closed")
	}
}

func TestGenerateTotalManaValueReanimationSource(t *testing.T) {
	t.Parallel()
	source, diagnostics, err := GenerateExecutableCardSource(&ScryfallCard{
		Name:     "Lively Dirge",
		Layout:   "normal",
		TypeLine: "Sorcery",
		ManaCost: "{1}{B}",
		OracleText: "Spree (Choose one or more additional costs.)\n" +
			"+ {1} — Search your library for a card, put it into your graveyard, then shuffle.\n" +
			"+ {2} — Return up to two creature cards with total mana value 4 or less from your graveyard to the battlefield.",
	}, "l")
	if err != nil {
		t.Fatal(err)
	}
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v; want Lively Dirge to fully generate", diagnostics)
	}
	if _, err := goparser.ParseFile(token.NewFileSet(), "lively_dirge.go", source, goparser.AllErrors); err != nil {
		t.Fatalf("generated source does not parse: %v\n%s", err, source)
	}
	for _, want := range []string{
		"game.ReturnFromGraveyard{",
		"game.Selection{RequiredTypes: []types.Card{types.Creature}, Controller: game.ControllerYou}",
		"Amount:            game.Fixed(2),",
		"Destination:       zone.Battlefield,",
		"MaxTotalManaValue: opt.Val(4),",
	} {
		if !strings.Contains(source, want) {
			t.Fatalf("generated source missing %q:\n%s", want, source)
		}
	}
}
