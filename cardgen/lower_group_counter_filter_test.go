package cardgen

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
)

// groupCounterAddCounter lowers a single-face card and returns the AddCounter
// primitive of its first group counter placement, scanning every ability kind so
// callers can exercise both bare-spell and triggered placements.
func groupCounterAddCounter(t *testing.T, card *ScryfallCard) game.AddCounter {
	t.Helper()
	face := lowerSingleFace(t, card)
	var sequences [][]game.Instruction
	if face.SpellAbility.Exists {
		for _, mode := range face.SpellAbility.Val.Modes {
			sequences = append(sequences, mode.Sequence)
		}
	}
	for _, ability := range face.TriggeredAbilities {
		for _, mode := range ability.Content.Modes {
			sequences = append(sequences, mode.Sequence)
		}
	}
	for _, ability := range face.ActivatedAbilities {
		for _, mode := range ability.Content.Modes {
			sequences = append(sequences, mode.Sequence)
		}
	}
	for _, sequence := range sequences {
		for _, instruction := range sequence {
			if add, ok := instruction.Primitive.(game.AddCounter); ok {
				return add
			}
		}
	}
	t.Fatalf("no AddCounter primitive lowered for %q", card.OracleText)
	return game.AddCounter{}
}

// groupCounterSelection lowers a single-face instant whose only effect is a
// group counter placement and returns the AddCounter group's runtime selection.
func groupCounterSelection(t *testing.T, oracleText string) game.Selection {
	t.Helper()
	add := groupCounterAddCounter(t, &ScryfallCard{
		Name:       "Test Group Counter",
		Layout:     "normal",
		TypeLine:   "Instant",
		OracleText: oracleText,
	})
	if add.CounterKind != counter.PlusOnePlusOne || add.Amount.Value() != 1 {
		t.Fatalf("primitive = %+v, want one +1/+1 counter", add)
	}
	if !add.Group.Valid() {
		t.Fatalf("counter placement missing battlefield group: %+v", add)
	}
	return add.Group.Selection()
}

func TestLowerGroupCounterKeywordFilter(t *testing.T) {
	t.Parallel()
	selection := groupCounterSelection(t, "Put a +1/+1 counter on each creature you control with vigilance.")
	if selection.Controller != game.ControllerYou {
		t.Fatalf("controller = %v, want ControllerYou", selection.Controller)
	}
	if selection.Keyword != game.Vigilance {
		t.Fatalf("keyword = %v, want Vigilance", selection.Keyword)
	}
}

func TestLowerGroupCounterCounterFilter(t *testing.T) {
	t.Parallel()
	// The "with a +1/+1 counter on it" qualifier carries a trailing pronoun the
	// bare-spell amount path counts as a reference, so the real cards (Patron of
	// the Valiant, Twisted Spider-Clone) place this filtered group from a
	// triggered ability; exercise that shape.
	add := groupCounterAddCounter(t, &ScryfallCard{
		Name:       "Test Counter Filter",
		Layout:     "normal",
		TypeLine:   "Creature — Soldier",
		OracleText: "When this creature enters, put a +1/+1 counter on each creature you control with a +1/+1 counter on it.",
	})
	selection := add.Group.Selection()
	if selection.Controller != game.ControllerYou {
		t.Fatalf("controller = %v, want ControllerYou", selection.Controller)
	}
	if !selection.MatchCounter || selection.RequiredCounter != counter.PlusOnePlusOne {
		t.Fatalf("counter filter = (%v, %v), want a +1/+1 counter requirement", selection.MatchCounter, selection.RequiredCounter)
	}
}

func TestLowerGroupCounterMultiSubtypeFilter(t *testing.T) {
	t.Parallel()
	selection := groupCounterSelection(t, "Put a +1/+1 counter on each Pest, Bat, Insect, Snake, and Spider you control.")
	if selection.Controller != game.ControllerYou {
		t.Fatalf("controller = %v, want ControllerYou", selection.Controller)
	}
	want := []types.Sub{types.Pest, types.Bat, types.Insect, types.Snake, types.Spider}
	if len(selection.SubtypesAny) != len(want) {
		t.Fatalf("subtypes = %+v, want %+v", selection.SubtypesAny, want)
	}
	for i, sub := range want {
		if selection.SubtypesAny[i] != sub {
			t.Fatalf("subtype %d = %v, want %v", i, selection.SubtypesAny[i], sub)
		}
	}
}

func TestLowerGroupCounterNamedFilter(t *testing.T) {
	t.Parallel()
	add := groupCounterAddCounter(t, &ScryfallCard{
		Name:       "Test Named Filter",
		Layout:     "normal",
		TypeLine:   "Creature — Cat",
		OracleText: "When this creature enters, put a +1/+1 counter on each other creature you control named Charmed Stray.",
	})
	selection := add.Group.Selection()
	if selection.Controller != game.ControllerYou {
		t.Fatalf("controller = %v, want ControllerYou", selection.Controller)
	}
	if selection.Name != "Charmed Stray" {
		t.Fatalf("name = %q, want %q", selection.Name, "Charmed Stray")
	}
	exclusion, ok := add.Group.Exclusion()
	if !ok || exclusion != game.SourcePermanentReference() {
		t.Fatalf("exclusion = (%v, %v), want the source permanent excluded from the \"other\" group", exclusion, ok)
	}
}

func TestLowerGroupCounterOpponentControlled(t *testing.T) {
	t.Parallel()
	selection := groupCounterSelection(t, "Put a +1/+1 counter on each creature your opponents control.")
	if selection.Controller != game.ControllerOpponent {
		t.Fatalf("controller = %v, want ControllerOpponent", selection.Controller)
	}
}

// TestLowerGroupCounterEachOpponentControlled exercises the distributive "each
// opponent controls" opponent wording (Aku Djinn). It denotes the same
// opponent-controlled group as the plural "your opponents control" form and
// lowers identically to ControllerOpponent.
func TestLowerGroupCounterEachOpponentControlled(t *testing.T) {
	t.Parallel()
	add := groupCounterAddCounter(t, &ScryfallCard{
		Name:       "Test Each Opponent Group",
		Layout:     "normal",
		TypeLine:   "Creature — Djinn",
		OracleText: "At the beginning of your upkeep, put a +1/+1 counter on each creature each opponent controls.",
	})
	selection := add.Group.Selection()
	if selection.Controller != game.ControllerOpponent {
		t.Fatalf("controller = %v, want ControllerOpponent", selection.Controller)
	}
}

// TestLowerGroupCounterDynamicSourcePower exercises a dynamic-X group counter
// placement whose count is the source permanent's power (Ouroboroid). The "where
// X is this creature's power" count phrase trails the recipient group, so the
// recipient must lower to the plain "each creature you control" group without
// folding the count subject into its filter.
func TestLowerGroupCounterDynamicSourcePower(t *testing.T) {
	t.Parallel()
	add := groupCounterAddCounter(t, &ScryfallCard{
		Name:       "Test Dynamic Source Power Group",
		Layout:     "normal",
		TypeLine:   "Creature — Ooze",
		OracleText: "At the beginning of combat on your turn, put X +1/+1 counters on each creature you control, where X is this creature's power.",
	})
	if !add.Amount.IsDynamic() {
		t.Fatalf("amount = %+v, want a dynamic count", add.Amount)
	}
	if selection := add.Group.Selection(); selection.Controller != game.ControllerYou {
		t.Fatalf("controller = %v, want ControllerYou", selection.Controller)
	}
}

// TestLowerGroupCounterDynamicCountSubtype exercises a dynamic-X group counter
// placement whose count names a subtype ("the number of Shrines you control",
// Southern Air Temple). The subtype in the count phrase must not bleed into the
// recipient group, which stays the plain "each creature you control" group.
func TestLowerGroupCounterDynamicCountSubtype(t *testing.T) {
	t.Parallel()
	add := groupCounterAddCounter(t, &ScryfallCard{
		Name:       "Southern Air Temple",
		Layout:     "normal",
		TypeLine:   "Enchantment — Shrine",
		OracleText: "When Southern Air Temple enters, put X +1/+1 counters on each creature you control, where X is the number of Shrines you control.",
	})
	if !add.Amount.IsDynamic() {
		t.Fatalf("amount = %+v, want a dynamic count", add.Amount)
	}
	selection := add.Group.Selection()
	if selection.Controller != game.ControllerYou {
		t.Fatalf("controller = %v, want ControllerYou", selection.Controller)
	}
	if len(selection.SubtypesAny) != 0 {
		t.Fatalf("subtypes = %+v, want none (the count subtype must not fold into the recipient)", selection.SubtypesAny)
	}
}

// TestLowerGroupCounterTargetPlayerControlledRejected confirms a group whose
// controller is a runtime-chosen target player ("each creature target player
// controls") fails closed: the recipient has no static controller relation and
// no target-player group binding exists, so it must stay unsupported rather than
// lower to a wrong group.
func TestLowerGroupCounterTargetPlayerControlledRejected(t *testing.T) {
	t.Parallel()
	lowerSingleFaceExpectingUnsupported(t, &ScryfallCard{
		Name:       "Test Target Player Group",
		Layout:     "normal",
		TypeLine:   "Instant",
		OracleText: "Put a +1/+1 counter on each creature target player controls.",
	})
}
