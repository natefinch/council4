package cardgen

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// TestLowerOptionalSearchSpell verifies that a resolving optional library-search
// tutor ("You may search your library for ...") lowers to a single game.Search
// instruction marked Optional, so the engine asks the controller whether to
// perform the whole tutor.
func TestLowerOptionalSearchSpell(t *testing.T) {
	t.Parallel()
	faces, diagnostics := lowerExecutableFaces(&ScryfallCard{
		Name:       "Optional Tutor",
		Layout:     "normal",
		TypeLine:   "Creature — Elf",
		OracleText: "When this creature enters, you may search your library for a basic land card, put it onto the battlefield tapped, then shuffle.",
		Power:      new("1"),
		Toughness:  new("1"),
	})
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	if len(faces) != 1 || len(faces[0].TriggeredAbilities) != 1 {
		t.Fatalf("faces = %#v", faces)
	}
	seq := faces[0].TriggeredAbilities[0].Content.Modes[0].Sequence
	if len(seq) != 1 {
		t.Fatalf("sequence = %#v, want one instruction", seq)
	}
	if !seq[0].Optional {
		t.Error("Search instruction Optional = false, want true")
	}
	search, ok := seq[0].Primitive.(game.Search)
	if !ok {
		t.Fatalf("primitive = %#v, want game.Search", seq[0].Primitive)
	}
	want := game.SearchSpec{
		SourceZone:   zone.Library,
		Destination:  zone.Battlefield,
		CardType:     opt.Val(types.Land),
		Supertype:    opt.Val(types.Basic),
		EntersTapped: true,
	}
	if !searchSpecEqual(search.Spec, want) {
		t.Errorf("spec = %+v, want %+v", search.Spec, want)
	}
}

// TestLowerOptionalSearchToHand verifies the hand-destination, reveal-first tutor
// shape also lowers optionally.
func TestLowerOptionalSearchToHand(t *testing.T) {
	t.Parallel()
	faces, diagnostics := lowerExecutableFaces(&ScryfallCard{
		Name:       "Optional Hand Tutor",
		Layout:     "normal",
		TypeLine:   "Creature — Human",
		OracleText: "When this creature enters, you may search your library for a creature card, reveal it, put it into your hand, then shuffle.",
		Power:      new("1"),
		Toughness:  new("1"),
	})
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	seq := faces[0].TriggeredAbilities[0].Content.Modes[0].Sequence
	if len(seq) != 1 || !seq[0].Optional {
		t.Fatalf("sequence = %#v, want one Optional instruction", seq)
	}
	search, ok := seq[0].Primitive.(game.Search)
	if !ok {
		t.Fatalf("primitive = %#v, want game.Search", seq[0].Primitive)
	}
	want := game.SearchSpec{
		SourceZone:  zone.Library,
		Destination: zone.Hand,
		CardType:    opt.Val(types.Creature),
		Reveal:      true,
	}
	if !searchSpecEqual(search.Spec, want) {
		t.Errorf("spec = %+v, want %+v", search.Spec, want)
	}
}

// TestLowerOptionalSearchUnsupportedFilterFailsClosed confirms a tutor whose
// filter the runtime cannot model (a creature subtype) stays unsupported even
// when wrapped in "you may", rather than lowering to a silently-wrong search.
func TestLowerOptionalSearchUnsupportedFilterFailsClosed(t *testing.T) {
	t.Parallel()
	_, diagnostics := lowerExecutableFaces(&ScryfallCard{
		Name:       "Goblin Caller",
		Layout:     "normal",
		TypeLine:   "Creature — Goblin",
		OracleText: "When this creature enters, you may search your library for a Goblin card, reveal it, put it into your hand, then shuffle.",
		Power:      new("1"),
		Toughness:  new("1"),
	})
	if len(diagnostics) == 0 {
		t.Fatal("expected an unsupported diagnostic for a subtype tutor, got none")
	}
}

// TestLowerTrailingBareOptionalEffect verifies that a multi-effect sequence whose
// final effect carries resolving optionality with no "if you do" gate marks only
// that effect's instruction Optional while the preceding mandatory effect lowers
// unchanged.
func TestLowerTrailingBareOptionalEffect(t *testing.T) {
	t.Parallel()
	faces, diagnostics := lowerExecutableFaces(&ScryfallCard{
		Name:       "Surveil Drawer",
		Layout:     "normal",
		TypeLine:   "Sorcery",
		OracleText: "Surveil 2. You may draw a card.",
	})
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	seq := faces[0].SpellAbility.Val.Modes[0].Sequence
	if len(seq) != 2 {
		t.Fatalf("sequence = %#v, want two instructions", seq)
	}
	if seq[0].Optional {
		t.Error("first (Surveil) instruction Optional = true, want false")
	}
	if !seq[1].Optional {
		t.Error("trailing (Draw) instruction Optional = false, want true")
	}
	if _, ok := seq[1].Primitive.(game.Draw); !ok {
		t.Errorf("trailing primitive = %#v, want game.Draw", seq[1].Primitive)
	}
}
