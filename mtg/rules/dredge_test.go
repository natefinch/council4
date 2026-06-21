package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
)

// dredgeChoiceAgent always selects the first offered Dredge option, modeling a
// player who chooses to dredge instead of drawing.
type dredgeChoiceAgent struct{}

func (dredgeChoiceAgent) ChooseAction(PlayerObservation, []action.Action) action.Action {
	return action.Action{}
}

func (dredgeChoiceAgent) ChooseChoice(_ PlayerObservation, request game.ChoiceRequest) []int {
	if request.Kind == game.ChoiceZoneSelection && len(request.Options) > 0 {
		return []int{request.Options[0].Index}
	}
	return nil
}

func dredgeCardDef(n int) *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name:            "Test Dredger",
			StaticAbilities: []game.StaticAbility{game.DredgeStaticAbility(n)},
		},
	}
}

func TestDredgeReplacesDrawWhenChosen(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	for range 5 {
		addCardToLibrary(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Card"}})
	}
	dredgerID := addCardToGraveyard(g, game.Player1, dredgeCardDef(3))

	agents := [game.NumPlayers]PlayerAgent{game.Player1: dredgeChoiceAgent{}}
	log := TurnLog{}
	engine.drawCards(g, game.Player1, 1, agents, &log)

	player := g.Players[game.Player1]
	if got := player.Library.Size(); got != 2 {
		t.Fatalf("library size = %d, want 2 (milled three)", got)
	}
	if got := player.Graveyard.Size(); got != 3 {
		t.Fatalf("graveyard size = %d, want 3 (three milled cards)", got)
	}
	if got := player.Hand.Size(); got != 1 {
		t.Fatalf("hand size = %d, want 1 (the dredged card)", got)
	}
	if !player.Hand.Contains(dredgerID) {
		t.Fatalf("hand %v does not contain dredged card %v", player.Hand.All(), dredgerID)
	}
}

func TestDredgeDeclinedDrawsNormally(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	for range 5 {
		addCardToLibrary(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Card"}})
	}
	dredgerID := addCardToGraveyard(g, game.Player1, dredgeCardDef(3))

	log := TurnLog{}
	engine.drawCards(g, game.Player1, 1, [game.NumPlayers]PlayerAgent{}, &log)

	player := g.Players[game.Player1]
	if got := player.Library.Size(); got != 4 {
		t.Fatalf("library size = %d, want 4 (one normal draw, no mill)", got)
	}
	if got := player.Hand.Size(); got != 1 {
		t.Fatalf("hand size = %d, want 1 (one drawn card)", got)
	}
	if got := player.Graveyard.Size(); got != 1 || !player.Graveyard.Contains(dredgerID) {
		t.Fatalf("graveyard = %v, want only the undredged card", player.Graveyard.All())
	}
}
