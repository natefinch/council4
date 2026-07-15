package cardgen

import (
	"slices"
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/compare"
	"github.com/natefinch/council4/mtg/game/types"
)

const ingaAndEsikaText = "Creatures you control have vigilance and \"{T}: Add one mana of any color. Spend this mana only to cast a creature spell.\"\n" +
	"Whenever you cast a creature spell, if three or more mana from creatures was spent to cast it, draw a card."

// TestLowerIngaAndEsika proves the full parser→compiler→lowering pipeline turns
// Inga and Esika's real Oracle text into its spell-cast draw trigger gated on
// creature mana: the trigger fires on casting a creature spell, its intervening
// "if" lowers to an AggregateEventSpellManaFromCreaturesSpentToCast >= 3
// comparison, and its body draws a card. The static that grants vigilance and
// the restricted "creature mana" ability lowers alongside it.
func TestLowerIngaAndEsika(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Inga and Esika",
		Layout:     "normal",
		TypeLine:   "Legendary Creature — Human God",
		ManaCost:   "{2}{G}{U}",
		OracleText: ingaAndEsikaText,
		Power:      new("4"),
		Toughness:  new("4"),
	})

	if len(face.TriggeredAbilities) != 1 {
		t.Fatalf("triggered abilities = %d, want 1", len(face.TriggeredAbilities))
	}
	ta := face.TriggeredAbilities[0]
	if ta.Trigger.Type != game.TriggerWhenever || ta.Trigger.Pattern.Event != game.EventSpellCast {
		t.Fatalf("trigger = {%v %v}, want {TriggerWhenever EventSpellCast}", ta.Trigger.Type, ta.Trigger.Pattern.Event)
	}
	if !castTriggerRequiresCreature(ta.Trigger.Pattern) {
		t.Fatalf("trigger pattern does not require a creature spell: %#v", ta.Trigger.Pattern)
	}

	if !ta.Trigger.InterveningCondition.Exists {
		t.Fatalf("intervening condition missing; trigger = %+v", ta.Trigger)
	}
	aggregates := ta.Trigger.InterveningCondition.Val.Aggregates
	if len(aggregates) != 1 {
		t.Fatalf("aggregates = %#v, want exactly one", aggregates)
	}
	got := aggregates[0]
	if got.Aggregate != game.AggregateEventSpellManaFromCreaturesSpentToCast {
		t.Errorf("aggregate = %v, want AggregateEventSpellManaFromCreaturesSpentToCast", got.Aggregate)
	}
	if got.Op != compare.GreaterOrEqual || got.Value != 3 {
		t.Errorf("comparison = {Op:%v Value:%d}, want {GreaterOrEqual 3}", got.Op, got.Value)
	}

	if len(ta.Content.Modes) != 1 || len(ta.Content.Modes[0].Sequence) == 0 {
		t.Fatalf("trigger content = %#v, want a single-mode draw body", ta.Content)
	}
	if _, ok := ta.Content.Modes[0].Sequence[0].Primitive.(game.Draw); !ok {
		t.Fatalf("trigger body primitive = %#v, want game.Draw", ta.Content.Modes[0].Sequence[0].Primitive)
	}

	if len(face.StaticAbilities) != 1 {
		t.Fatalf("static abilities = %d, want 1 (grants vigilance + creature mana ability)", len(face.StaticAbilities))
	}
}

// castTriggerRequiresCreature reports whether a spell-cast trigger pattern is
// restricted to creature spells through either the legacy card-type filter or
// the Selection-based form.
func castTriggerRequiresCreature(pattern game.TriggerPattern) bool {
	if slices.Contains(pattern.RequireCardTypes, types.Creature) {
		return true
	}
	return slices.Contains(pattern.CardSelection.RequiredTypes, types.Creature)
}
