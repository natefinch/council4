package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
)

// exileForPlayDiscardPermanent gives playerID a battlefield permanent whose
// discard trigger exiles the just-discarded card from the graveyard and grants
// permission to play (or, when cast is set, cast) it this turn. It models
// Containment Construct and Conspiracy Theorist's reflexive bodies.
func exileForPlayDiscardPermanent(g *game.Game, playerID game.PlayerID, cast bool) *game.Permanent {
	return addOptionalTriggeredPermanent(g, playerID,
		&game.TriggerPattern{Event: game.EventCardDiscarded, Player: game.TriggerPlayerYou},
		[]game.Instruction{{
			Optional: true,
			Primitive: game.ExileForPlay{
				Card:     game.CardReference{Kind: game.CardReferenceEvent},
				FromZone: zone.Graveyard,
				Duration: game.DurationThisTurn,
				Cast:     cast,
			},
		}}, nil)
}

func TestExileForPlayExilesDiscardedCardAndGrantsPlay(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	exileForPlayDiscardPermanent(g, game.Player1, false)
	forestID := addCardToHand(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name: "Forest", Types: []types.Card{types.Land}}})

	if !discardCardFromHand(g, game.Player1, forestID) {
		t.Fatal("discardCardFromHand() = false, want true")
	}
	agents := [game.NumPlayers]PlayerAgent{game.Player1: &choiceOnlyAgent{choices: [][]int{{1}}}}
	if !engine.putTriggeredAbilitiesOnStackWithChoices(g, agents, &TurnLog{}) {
		t.Fatal("discard trigger was not put on the stack")
	}
	engine.resolveTopOfStackWithChoices(g, agents, &TurnLog{})

	if g.Players[game.Player1].Graveyard.Contains(forestID) {
		t.Fatal("discarded card remained in the graveyard after the exile")
	}
	if !g.Players[game.Player1].Exile.Contains(forestID) {
		t.Fatal("discarded card was not exiled")
	}
	if !canPlayLandFromZoneByRuleEffect(g, game.Player1, forestID, zone.Exile) {
		t.Fatal("exiled discarded land is not playable despite the play permission")
	}
	if !hasCastFromZoneRuleEffect(g, game.Player1, forestID, zone.Exile, game.FaceFront) {
		t.Fatal("play permission does not allow casting the exiled card")
	}
}

func TestExileForPlayDeclinedLeavesCardInGraveyard(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	exileForPlayDiscardPermanent(g, game.Player1, false)
	forestID := addCardToHand(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name: "Forest", Types: []types.Card{types.Land}}})

	discardCardFromHand(g, game.Player1, forestID)
	agents := [game.NumPlayers]PlayerAgent{game.Player1: &choiceOnlyAgent{choices: [][]int{{0}}}}
	engine.putTriggeredAbilitiesOnStackWithChoices(g, agents, &TurnLog{})
	engine.resolveTopOfStackWithChoices(g, agents, &TurnLog{})

	if !g.Players[game.Player1].Graveyard.Contains(forestID) {
		t.Fatal("declined exile moved the card out of the graveyard")
	}
	if g.Players[game.Player1].Exile.Contains(forestID) {
		t.Fatal("declined exile still exiled the card")
	}
	if canPlayLandFromZoneByRuleEffect(g, game.Player1, forestID, zone.Exile) {
		t.Fatal("declined exile still granted play permission")
	}
}

func TestExileForPlayCastVariantGrantsCastOnly(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	exileForPlayDiscardPermanent(g, game.Player1, true)
	spellID := addCardToHand(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name: "Test Sorcery", Types: []types.Card{types.Sorcery}}})

	discardCardFromHand(g, game.Player1, spellID)
	agents := [game.NumPlayers]PlayerAgent{game.Player1: &choiceOnlyAgent{choices: [][]int{{1}}}}
	engine.putTriggeredAbilitiesOnStackWithChoices(g, agents, &TurnLog{})
	engine.resolveTopOfStackWithChoices(g, agents, &TurnLog{})

	if !g.Players[game.Player1].Exile.Contains(spellID) {
		t.Fatal("discarded card was not exiled")
	}
	if !hasCastFromZoneRuleEffect(g, game.Player1, spellID, zone.Exile, game.FaceFront) {
		t.Fatal("cast variant did not grant cast permission for the exiled card")
	}
	if canPlayLandFromZoneByRuleEffect(g, game.Player1, spellID, zone.Exile) {
		t.Fatal("cast variant granted land-play permission, but should grant cast only")
	}
}

// exileForPlayBatchDiscardPermanent models Conspiracy Theorist's "Whenever you
// discard one or more nonland cards, you may exile one of them from your
// graveyard. If you do, you may cast it this turn." discard trigger: a
// one-or-more batch trigger whose SelectFromBatch ExileForPlay lets the
// controller choose which discarded card to exile and gain cast permission for.
func exileForPlayBatchDiscardPermanent(g *game.Game, playerID game.PlayerID) *game.Permanent {
	return addOptionalTriggeredPermanent(g, playerID,
		&game.TriggerPattern{Event: game.EventCardDiscarded, Player: game.TriggerPlayerYou, OneOrMore: true},
		[]game.Instruction{{
			Optional: true,
			Primitive: game.ExileForPlay{
				Card:            game.CardReference{Kind: game.CardReferenceEvent},
				FromZone:        zone.Graveyard,
				Duration:        game.DurationThisTurn,
				Cast:            true,
				SelectFromBatch: true,
			},
		}}, nil)
}

func TestExileForPlayBatchSelectsChosenDiscardedCard(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	exileForPlayBatchDiscardPermanent(g, game.Player1)
	firstID := addCardToHand(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name: "First Sorcery", Types: []types.Card{types.Sorcery}}})
	secondID := addCardToHand(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name: "Second Sorcery", Types: []types.Card{types.Sorcery}}})

	simultaneousID := g.IDGen.Next()
	if !discardCardFromHandInBatch(g, game.Player1, firstID, simultaneousID) ||
		!discardCardFromHandInBatch(g, game.Player1, secondID, simultaneousID) {
		t.Fatal("batched discard failed")
	}
	// The first two choices accept the optional trigger and optional effect;
	// the third selects the second discarded card from the batch.
	agents := [game.NumPlayers]PlayerAgent{game.Player1: &choiceOnlyAgent{choices: [][]int{{1}, {1}, {1}}}}
	if !engine.putTriggeredAbilitiesOnStackWithChoices(g, agents, &TurnLog{}) {
		t.Fatal("discard trigger was not put on the stack")
	}
	engine.resolveTopOfStackWithChoices(g, agents, &TurnLog{})

	if !g.Players[game.Player1].Exile.Contains(secondID) {
		t.Fatal("chosen discarded card was not exiled")
	}
	if g.Players[game.Player1].Exile.Contains(firstID) {
		t.Fatal("unchosen discarded card was exiled")
	}
	if !g.Players[game.Player1].Graveyard.Contains(firstID) {
		t.Fatal("unchosen discarded card left the graveyard")
	}
	if !hasCastFromZoneRuleEffect(g, game.Player1, secondID, zone.Exile, game.FaceFront) {
		t.Fatal("cast permission was not granted for the chosen exiled card")
	}
}

func TestExileForPlayBatchDeclinedExilesNothing(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	exileForPlayBatchDiscardPermanent(g, game.Player1)
	firstID := addCardToHand(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name: "First Sorcery", Types: []types.Card{types.Sorcery}}})
	secondID := addCardToHand(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name: "Second Sorcery", Types: []types.Card{types.Sorcery}}})

	simultaneousID := g.IDGen.Next()
	discardCardFromHandInBatch(g, game.Player1, firstID, simultaneousID)
	discardCardFromHandInBatch(g, game.Player1, secondID, simultaneousID)
	agents := [game.NumPlayers]PlayerAgent{game.Player1: &choiceOnlyAgent{choices: [][]int{{0}}}}
	engine.putTriggeredAbilitiesOnStackWithChoices(g, agents, &TurnLog{})
	engine.resolveTopOfStackWithChoices(g, agents, &TurnLog{})

	if g.Players[game.Player1].Exile.Contains(firstID) || g.Players[game.Player1].Exile.Contains(secondID) {
		t.Fatal("declined batch exile still exiled a card")
	}
	if !g.Players[game.Player1].Graveyard.Contains(firstID) || !g.Players[game.Player1].Graveyard.Contains(secondID) {
		t.Fatal("declined batch exile moved cards out of the graveyard")
	}
}
