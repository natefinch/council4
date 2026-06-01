package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/mana"
)

func TestMadnessDiscardGoesToExile(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	cardID := addCardToHand(g, game.Player1, madnessSorcery(mana.Cost{mana.GenericMana(1)}))

	if !discardCardFromHand(g, game.Player1, cardID) {
		t.Fatal("discardCardFromHand() = false, want true")
	}

	if g.Players[game.Player1].Graveyard.Contains(cardID) {
		t.Fatal("madness card went to graveyard instead of exile")
	}
	if !g.Players[game.Player1].Exile.Contains(cardID) {
		t.Fatal("madness card did not go to exile")
	}
	assertEvent(t, g.Events, game.EventCardDiscarded, func(event game.GameEvent) bool {
		return event.CardID == cardID && event.FromZone == game.ZoneHand && event.ToZone == game.ZoneExile
	})
}

func TestMadnessTriggerCastsCardFromExile(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	cardID := addCardToHand(g, game.Player1, madnessSorcery(mana.Cost{mana.ColoredMana(mana.Green)}))
	forest := addBasicLandPermanent(g, game.Player1, game.LandSubtypeForest)

	discardCardFromHand(g, game.Player1, cardID)
	if !engine.putTriggeredAbilitiesOnStack(g) {
		t.Fatal("madness trigger was not put on the stack")
	}
	engine.resolveTopOfStack(g, &TurnLog{})

	if !forest.Tapped {
		t.Fatal("madness cost did not tap mana source")
	}
	if g.Players[game.Player1].Exile.Contains(cardID) {
		t.Fatal("cast madness card remained in exile")
	}
	obj, ok := g.Stack.Peek()
	if !ok || obj.Kind != game.StackSpell || obj.SourceID != cardID {
		t.Fatalf("stack top = %+v, want madness spell", obj)
	}
	assertEvent(t, g.Events, game.EventSpellCast, func(event game.GameEvent) bool {
		return event.CardID == cardID && event.FromZone == game.ZoneExile && event.ToZone == game.ZoneStack
	})
}

func TestUnpayableMadnessTriggerMovesCardToGraveyard(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	cardID := addCardToHand(g, game.Player1, madnessSorcery(mana.Cost{mana.GenericMana(1)}))

	discardCardFromHand(g, game.Player1, cardID)
	engine.putTriggeredAbilitiesOnStack(g)
	engine.resolveTopOfStack(g, &TurnLog{})

	if g.Players[game.Player1].Exile.Contains(cardID) {
		t.Fatal("unpayable madness card remained in exile")
	}
	if !g.Players[game.Player1].Graveyard.Contains(cardID) {
		t.Fatal("unpayable madness card did not move to graveyard")
	}
	if g.Stack.Size() != 0 {
		t.Fatalf("stack size = %d, want no madness spell", g.Stack.Size())
	}
}

func TestDeclinedMadnessTriggerMovesCardToGraveyard(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	cardID := addCardToHand(g, game.Player1, madnessSorcery(mana.Cost{mana.ColoredMana(mana.Green)}))
	addBasicLandPermanent(g, game.Player1, game.LandSubtypeForest)
	agents := [game.NumPlayers]PlayerAgent{game.Player1: &choiceOnlyAgent{choices: [][]int{{0}}}}

	discardCardFromHand(g, game.Player1, cardID)
	engine.putTriggeredAbilitiesOnStack(g)
	engine.resolveTopOfStackWithChoices(g, agents, &TurnLog{})

	if g.Players[game.Player1].Exile.Contains(cardID) {
		t.Fatal("declined madness card remained in exile")
	}
	if !g.Players[game.Player1].Graveyard.Contains(cardID) {
		t.Fatal("declined madness card did not move to graveyard")
	}
	if g.Stack.Size() != 0 {
		t.Fatalf("stack size = %d, want no madness spell", g.Stack.Size())
	}
}

func madnessSorcery(cost mana.Cost) *game.CardDef {
	return &game.CardDef{
		Name:  "Madness Sorcery",
		Types: []game.CardType{game.TypeSorcery},
		Abilities: []game.AbilityDef{{
			Kind:        game.StaticAbility,
			Keywords:    []game.Keyword{game.Madness},
			MadnessCost: optCost(cost),
		}},
	}
}
