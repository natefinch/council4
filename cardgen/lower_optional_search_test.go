package cardgen

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
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

// TestLowerOptionalSearchSubtypeToHand verifies a now-supported subtype tutor
// ("a Goblin card") lowers optionally to a subtype-filtered hand search.
func TestLowerOptionalSearchSubtypeToHand(t *testing.T) {
	t.Parallel()
	faces, diagnostics := lowerExecutableFaces(&ScryfallCard{
		Name:       "Goblin Caller",
		Layout:     "normal",
		TypeLine:   "Creature — Goblin",
		OracleText: "When this creature enters, you may search your library for a Goblin card, reveal it, put it into your hand, then shuffle.",
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
		SubtypesAny: []types.Sub{types.Goblin},
		Reveal:      true,
	}
	if !searchSpecEqual(search.Spec, want) {
		t.Errorf("spec = %+v, want %+v", search.Spec, want)
	}
}

// TestLowerOptionalSearchUnsupportedFilterFailsClosed confirms a tutor whose
// filter the runtime cannot model (a multi-type union the single-type SearchSpec
// cannot express) stays unsupported even when wrapped in "you may", rather than
// lowering to a silently-wrong search.
func TestLowerOptionalSearchUnsupportedFilterFailsClosed(t *testing.T) {
	t.Parallel()
	_, diagnostics := lowerExecutableFaces(&ScryfallCard{
		Name:       "Goblin Caller",
		Layout:     "normal",
		TypeLine:   "Creature — Goblin",
		OracleText: "When this creature enters, you may search your library for an artifact creature card, reveal it, put it into your hand, then shuffle.",
		Power:      new("1"),
		Toughness:  new("1"),
	})
	if len(diagnostics) == 0 {
		t.Fatal("expected an unsupported diagnostic for a multi-type-union tutor, got none")
	}
}

// TestLowerOptionalSearchPermanentToBattlefield verifies an optional permanent
// tutor with a subtype and a "with mana value N or less" rider lowers to an
// Optional Search whose spec carries Permanent, the subtype, and the mana-value
// bound, exercising the permanent/MaxManaValue envelope through the "you may"
// path.
func TestLowerOptionalSearchPermanentToBattlefield(t *testing.T) {
	t.Parallel()
	faces, diagnostics := lowerExecutableFaces(&ScryfallCard{
		Name:       "Rebel Caller",
		Layout:     "normal",
		TypeLine:   "Creature — Rebel",
		OracleText: "When this creature enters, you may search your library for a Rebel permanent card with mana value 3 or less, put it onto the battlefield, then shuffle.",
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
		SourceZone:   zone.Library,
		Destination:  zone.Battlefield,
		Permanent:    true,
		SubtypesAny:  []types.Sub{types.Rebel},
		MaxManaValue: opt.Val(3),
	}
	if !searchSpecEqual(search.Spec, want) {
		t.Errorf("spec = %+v, want %+v", search.Spec, want)
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

// TestLowerControllerPaidCostThenSearch verifies the generic "you may <cost>. If
// you do, <search>" gate: an optional non-mana cost (sacrifice a land) folds
// into a resolution Pay whose success gates the whole multi-effect search
// consequence (search/put/shuffle merged into one Search instruction). This is
// the Springbloom Druid shape, where the ordered optional path cannot merge the
// multi-effect search but the paid path lowers it as standalone content.
func TestLowerControllerPaidCostThenSearch(t *testing.T) {
	t.Parallel()
	faces, diagnostics := lowerExecutableFaces(&ScryfallCard{
		Name:       "Springbloom Druid",
		Layout:     "normal",
		TypeLine:   "Creature — Human Druid",
		OracleText: "When this creature enters, you may sacrifice a land. If you do, search your library for up to two basic land cards, put them onto the battlefield tapped, then shuffle.",
		Power:      new("2"),
		Toughness:  new("2"),
	})
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	seq := faces[0].TriggeredAbilities[0].Content.Modes[0].Sequence
	if len(seq) != 2 {
		t.Fatalf("sequence = %#v, want two instructions", seq)
	}
	pay, ok := seq[0].Primitive.(game.Pay)
	if !ok {
		t.Fatalf("instruction[0] = %T, want game.Pay", seq[0].Primitive)
	}
	if seq[0].PublishResult != controllerPaidResultKey {
		t.Fatalf("instruction[0].PublishResult = %q, want %q", seq[0].PublishResult, controllerPaidResultKey)
	}
	if len(pay.Payment.AdditionalCosts) != 1 ||
		pay.Payment.AdditionalCosts[0].Kind != cost.AdditionalSacrifice {
		t.Fatalf("payment = %+v, want one sacrifice additional cost", pay.Payment)
	}
	if pay.Payment.ManaCost.Exists {
		t.Fatalf("payment carries mana cost %+v, want none", pay.Payment.ManaCost)
	}
	search, ok := seq[1].Primitive.(game.Search)
	if !ok {
		t.Fatalf("instruction[1] = %T, want game.Search", seq[1].Primitive)
	}
	if !seq[1].ResultGate.Exists ||
		seq[1].ResultGate.Val.Key != controllerPaidResultKey ||
		seq[1].ResultGate.Val.Succeeded != game.TriTrue {
		t.Fatalf("instruction[1].ResultGate = %#v, want succeeded gate on %q", seq[1].ResultGate, controllerPaidResultKey)
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

// TestLowerControllerPaidManaThenSearch verifies the mana-cost variant of the
// paid-cost-then-search gate ("you may pay {2}. If you do, <search>") folds the
// mana payment and gates the merged search instruction on it.
func TestLowerControllerPaidManaThenSearch(t *testing.T) {
	t.Parallel()
	faces, diagnostics := lowerExecutableFaces(&ScryfallCard{
		Name:       "Paid Tutor",
		Layout:     "normal",
		TypeLine:   "Creature — Human Druid",
		OracleText: "When this creature enters, you may pay {2}. If you do, search your library for up to two basic land cards, put them onto the battlefield tapped, then shuffle.",
		Power:      new("2"),
		Toughness:  new("2"),
	})
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	seq := faces[0].TriggeredAbilities[0].Content.Modes[0].Sequence
	if len(seq) != 2 {
		t.Fatalf("sequence = %#v, want two instructions", seq)
	}
	pay, ok := seq[0].Primitive.(game.Pay)
	if !ok {
		t.Fatalf("instruction[0] = %T, want game.Pay", seq[0].Primitive)
	}
	if !pay.Payment.ManaCost.Exists {
		t.Fatalf("payment = %+v, want a mana cost", pay.Payment)
	}
	if _, ok := seq[1].Primitive.(game.Search); !ok {
		t.Fatalf("instruction[1] = %T, want game.Search", seq[1].Primitive)
	}
	if !seq[1].ResultGate.Exists || seq[1].ResultGate.Val.Key != controllerPaidResultKey {
		t.Fatalf("instruction[1].ResultGate = %#v, want gate on %q", seq[1].ResultGate, controllerPaidResultKey)
	}
}
