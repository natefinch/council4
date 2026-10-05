package cardgen

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/mana"
)

// TestGenerateExecutableCardSourceMidnightClock exercises the full reusable
// pipeline Midnight Clock relies on: a mana ability, an activated ability and an
// each-upkeep trigger that both place an hour counter on the source, and a
// self-sourced Nth-counter threshold trigger whose content shuffles the
// controller's hand and graveyard into their library, draws a fixed number of
// cards, then exiles the source permanent.
func TestGenerateExecutableCardSourceMidnightClock(t *testing.T) {
	t.Parallel()
	card := &ScryfallCard{
		Name:     "Midnight Clock",
		Layout:   "normal",
		ManaCost: "{2}{U}",
		TypeLine: "Artifact",
		OracleText: "{T}: Add {U}.\n" +
			"{2}{U}: Put an hour counter on this artifact.\n" +
			"At the beginning of each upkeep, put an hour counter on this artifact.\n" +
			"When the twelfth hour counter is put on this artifact, shuffle your hand and graveyard into your library, then draw seven cards. Exile this artifact.",
	}
	assertCardPaths(t, card,
		"ManaAbilities[0].AdditionalCosts[0].Kind = cost.AdditionalTap",
		"ActivatedAbilities[0].Content.Modes[0].Sequence[0].Primitive.(game.AddCounter).CounterKind = counter.Hour",
		"ActivatedAbilities[0].Content.Modes[0].Sequence[0].Primitive.(game.AddCounter).Object.kind = game.ObjectReferenceSourcePermanent",
		"TriggeredAbilities[0].Trigger.Pattern.Event = game.EventBeginningOfStep",
		"TriggeredAbilities[0].Trigger.Pattern.Step = game.StepUpkeep",
		"TriggeredAbilities[0].Content.Modes[0].Sequence[0].Primitive.(game.AddCounter).CounterKind = counter.Hour",
		"TriggeredAbilities[1].Trigger.Pattern.Event = game.EventCountersAdded",
		"TriggeredAbilities[1].Trigger.Pattern.Source = game.TriggerSourceSelf",
		"TriggeredAbilities[1].Trigger.Pattern.MatchCounterKind = true",
		"TriggeredAbilities[1].Trigger.Pattern.CounterKind = counter.Hour",
		"TriggeredAbilities[1].Trigger.Pattern.CounterThreshold = 12",
		"TriggeredAbilities[1].Content.Modes[0].Sequence[0].Primitive.(game.ShuffleGraveyardIntoLibrary).IncludeHand = true",
		"TriggeredAbilities[1].Content.Modes[0].Sequence[2].Primitive.(game.MovePermanent).Object.kind = game.ObjectReferenceSourcePermanent",
	)
	assertCardPathsAbsent(t, card, "TriggeredAbilities[0].Trigger.Pattern.Player")
	face := lowerSingleFace(t, card)
	addMana := captureTestPrimitive[game.AddMana](t, face.ManaAbilities[0].Content.Modes[0].Sequence[0].Primitive)
	draw := captureTestPrimitive[game.Draw](t, face.TriggeredAbilities[1].Content.Modes[0].Sequence[1].Primitive)
	if addMana.ManaColor != mana.U || addMana.Amount != game.Fixed(1) || draw.Amount != game.Fixed(7) {
		t.Fatal("Midnight Clock lost its fixed mana output or seven-card draw")
	}
}
