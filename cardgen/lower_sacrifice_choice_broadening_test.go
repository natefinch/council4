package cardgen

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/types"
)

// TestLowerSacrificeChoiceHighAmountSequence verifies that a spelled sacrifice
// amount above the historical cap of two ("Each player sacrifices four lands of
// their choice.", "sacrifice three creatures.") reconstructs exactly and lowers
// to a SacrificePermanents with the fixed count, so the surrounding ordered
// sequence (Wildfire's mass damage, Greed's Gambit's leaves-the-battlefield
// edict) lowers end to end. The runtime primitive already carried an unbounded
// fixed amount; only the parser round-trip capped the value.
func TestLowerSacrificeChoiceHighAmountSequence(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Wildfire",
		Layout:     "normal",
		TypeLine:   "Sorcery",
		OracleText: "Each player sacrifices four lands of their choice. Test Wildfire deals 4 damage to each creature.",
	})
	if !face.SpellAbility.Exists {
		t.Fatal("spell ability missing")
	}
	mode := face.SpellAbility.Val.Modes[0]
	if len(mode.Sequence) != 2 {
		t.Fatalf("sequence = %d instructions, want 2 (sacrifice, damage)", len(mode.Sequence))
	}
	sacrifice, ok := mode.Sequence[0].Primitive.(game.SacrificePermanents)
	if !ok {
		t.Fatalf("first primitive = %T, want game.SacrificePermanents", mode.Sequence[0].Primitive)
	}
	if got := sacrifice.Amount; got != game.Fixed(4) {
		t.Fatalf("sacrifice amount = %+v, want Fixed(4)", got)
	}
	if sacrifice.PlayerGroup != game.AllPlayersReference() {
		t.Fatalf("sacrifice player group = %+v, want all players", sacrifice.PlayerGroup)
	}
}

// TestLowerSacrificeChoiceExcludedSubtype verifies the "non-<subtype>" sacrifice
// noun ("non-Zombie creature", "non-Vampire creature") reconstructs exactly and
// lowers to a SacrificePermanents whose Selection drops that creature subtype
// (Archdemon of Unx, Anowon the Ruin Sage, Dreadfeast Demon, Ruthless Winnower).
func TestLowerSacrificeChoiceExcludedSubtype(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Archdemon",
		Layout:     "normal",
		TypeLine:   "Creature",
		OracleText: "At the beginning of your upkeep, sacrifice a non-Zombie creature, then create a 2/2 black Zombie creature token.",
	})
	if len(face.TriggeredAbilities) != 1 {
		t.Fatalf("triggered abilities = %d, want 1", len(face.TriggeredAbilities))
	}
	mode := face.TriggeredAbilities[0].Content.Modes[0]
	if len(mode.Sequence) != 2 {
		t.Fatalf("sequence = %d instructions, want 2 (sacrifice, create token)", len(mode.Sequence))
	}
	sacrifice, ok := mode.Sequence[0].Primitive.(game.SacrificePermanents)
	if !ok {
		t.Fatalf("first primitive = %T, want game.SacrificePermanents", mode.Sequence[0].Primitive)
	}
	if sacrifice.Amount != game.Fixed(1) {
		t.Fatalf("sacrifice amount = %+v, want Fixed(1)", sacrifice.Amount)
	}
	if sacrifice.Selection.ExcludedSubtype != types.Sub("Zombie") {
		t.Fatalf("sacrifice excluded subtype = %q, want Zombie", sacrifice.Selection.ExcludedSubtype)
	}
	if len(sacrifice.Selection.RequiredTypes) != 1 || sacrifice.Selection.RequiredTypes[0] != types.Creature {
		t.Fatalf("sacrifice required types = %v, want [Creature]", sacrifice.Selection.RequiredTypes)
	}
}

// TestSacrificeChoiceExcludedSubtypeStaysUnsupported guards the fail-closed
// boundary: a sacrifice naming more than one excluded subtype has no canonical
// Oracle wording the round-trip reproduces, so it must stay unsupported rather
// than silently mislower.
func TestSacrificeChoiceExcludedSubtypeStaysUnsupported(t *testing.T) {
	t.Parallel()
	lowerSingleFaceExpectingUnsupported(t, &ScryfallCard{
		Name:       "Test Twin Exclusion",
		Layout:     "normal",
		TypeLine:   "Sorcery",
		OracleText: "Each player sacrifices a non-Zombie non-Vampire creature of their choice.",
	})
}

// TestLowerSacrificeChoicePlaneswalker verifies the single-type planeswalker
// sacrifice noun ("a planeswalker of their choice"; Angrath's Rampage,
// Sheoldred's Edict) reconstructs exactly and lowers to a SacrificePermanents
// whose Selection requires the Planeswalker card type.
func TestLowerSacrificeChoicePlaneswalker(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Rampage",
		Layout:     "normal",
		TypeLine:   "Sorcery",
		OracleText: "Target player sacrifices a planeswalker of their choice.",
	})
	if !face.SpellAbility.Exists {
		t.Fatal("spell ability missing")
	}
	sacrifice := firstSacrifice(t, face.SpellAbility.Val.Modes[0].Sequence)
	if len(sacrifice.Selection.RequiredTypes) != 1 || sacrifice.Selection.RequiredTypes[0] != types.Planeswalker {
		t.Fatalf("sacrifice required types = %v, want [Planeswalker]", sacrifice.Selection.RequiredTypes)
	}
	if sacrifice.Selection.TokenOnly {
		t.Fatalf("sacrifice selection unexpectedly token-only: %+v", sacrifice.Selection)
	}
}

// TestLowerSacrificeChoiceCreatureToken verifies the card-type token sacrifice
// noun ("a creature token of their choice"; Gaius van Baelsar, Sheoldred's
// Edict) reconstructs exactly and lowers to a SacrificePermanents whose
// Selection requires the Creature card type and restricts to tokens.
func TestLowerSacrificeChoiceCreatureToken(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Gaius",
		Layout:     "normal",
		TypeLine:   "Sorcery",
		OracleText: "Each opponent sacrifices a creature token of their choice.",
	})
	if !face.SpellAbility.Exists {
		t.Fatal("spell ability missing")
	}
	sacrifice := firstSacrifice(t, face.SpellAbility.Val.Modes[0].Sequence)
	if len(sacrifice.Selection.RequiredTypes) != 1 || sacrifice.Selection.RequiredTypes[0] != types.Creature {
		t.Fatalf("sacrifice required types = %v, want [Creature]", sacrifice.Selection.RequiredTypes)
	}
	if !sacrifice.Selection.TokenOnly {
		t.Fatalf("sacrifice selection = %+v, want token-only", sacrifice.Selection)
	}
	if sacrifice.Selection.NonToken {
		t.Fatalf("sacrifice selection = %+v, want token-only not nontoken", sacrifice.Selection)
	}
}

func firstSacrifice(t *testing.T, sequence []game.Instruction) game.SacrificePermanents {
	t.Helper()
	if len(sequence) == 0 {
		t.Fatal("empty instruction sequence")
	}
	sacrifice, ok := sequence[0].Primitive.(game.SacrificePermanents)
	if !ok {
		t.Fatalf("first primitive = %T, want game.SacrificePermanents", sequence[0].Primitive)
	}
	return sacrifice
}

// TestLowerEachPlayerSacrificesThirteen verifies Blasphemous Edict's edict line
// ("Each player sacrifices thirteen creatures of their choice.") reconstructs
// exactly and lowers to an all-players SacrificePermanents with a fixed count of
// thirteen — the exact-sacrifice-choice amount cap was raised to the full
// spelled-cardinal vocabulary (one … twenty), and the runtime primitive already
// carried an unbounded fixed amount.
func TestLowerEachPlayerSacrificesThirteen(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Edict",
		Layout:     "normal",
		TypeLine:   "Sorcery",
		OracleText: "Each player sacrifices thirteen creatures of their choice.",
	})
	if !face.SpellAbility.Exists {
		t.Fatal("spell ability missing")
	}
	sacrifice := firstSacrifice(t, face.SpellAbility.Val.Modes[0].Sequence)
	if sacrifice.Amount != game.Fixed(13) {
		t.Fatalf("sacrifice amount = %+v, want Fixed(13)", sacrifice.Amount)
	}
	if sacrifice.PlayerGroup != game.AllPlayersReference() {
		t.Fatalf("sacrifice player group = %+v, want all players", sacrifice.PlayerGroup)
	}
	if len(sacrifice.Selection.RequiredTypes) != 1 || sacrifice.Selection.RequiredTypes[0] != types.Creature {
		t.Fatalf("sacrifice required types = %v, want [Creature]", sacrifice.Selection.RequiredTypes)
	}
}

// TestLowerSacrificeChoiceUpperVocabularyBound verifies the top of the raised
// spelled-cardinal range ("twenty") still reconstructs and lowers, so the cap
// spans the full CardinalWordValue vocabulary rather than stopping short.
func TestLowerSacrificeChoiceUpperVocabularyBound(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Twenty",
		Layout:     "normal",
		TypeLine:   "Sorcery",
		OracleText: "Each player sacrifices twenty creatures of their choice.",
	})
	if !face.SpellAbility.Exists {
		t.Fatal("spell ability missing")
	}
	sacrifice := firstSacrifice(t, face.SpellAbility.Val.Modes[0].Sequence)
	if sacrifice.Amount != game.Fixed(20) {
		t.Fatalf("sacrifice amount = %+v, want Fixed(20)", sacrifice.Amount)
	}
}

// TestLowerSacrificeChoiceBeyondVocabularyStaysUnsupported guards the fail-closed
// boundary of the raised cap: a spelled amount outside the CardinalWordValue
// vocabulary ("thirty") is not a known amount, so the edict must stay
// unsupported rather than lower with a bogus or unbounded count.
func TestLowerSacrificeChoiceBeyondVocabularyStaysUnsupported(t *testing.T) {
	t.Parallel()
	lowerSingleFaceExpectingUnsupported(t, &ScryfallCard{
		Name:       "Test Thirty",
		Layout:     "normal",
		TypeLine:   "Sorcery",
		OracleText: "Each player sacrifices thirty creatures of their choice.",
	})
}
