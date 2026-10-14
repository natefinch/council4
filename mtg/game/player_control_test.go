package game

import (
	"encoding/json"
	"reflect"
	"testing"
)

func TestPlayerControlDataCloneSnapshotAndScopedOverride(t *testing.T) {
	g := NewGame([NumPlayers]PlayerConfig{})
	g.Turn.TurnNumber = 4
	g.Turn.ActivePlayer = Player2
	g.PlayerControlEffects = []PlayerControlEffect{
		{ID: 20, SourceCardID: 1, SourceObjectID: 3, Controller: Player1, Player: Player2, CreatedTurn: 3},
		{ID: 21, SourceCardID: 2, SourceObjectID: 2, Controller: Player3, Player: Player2, CreatedTurn: 3},
		{ID: 22, SourceCardID: 4, SourceObjectID: 4, Controller: Player1, Player: Player3, CreatedTurn: 4},
	}
	g.BeginPlayerControlTurn()
	g.BeginPlayerControlTurn()
	if len(g.PlayerControlEffects) != 2 || g.DecisionMaker(Player2, DecisionRules) != Player3 ||
		g.DecisionMaker(Player2, DecisionPersonal) != Player2 {
		t.Fatal("next-turn competition, idempotence, or personal ownership failed")
	}
	g.SearchControlScopes = []SearchControlScope{{Player: Player2, Controller: Player4, Override: true}}
	if g.DecisionMaker(Player2, DecisionRules) != Player4 || !g.CanSeePlayerInformation(Player4, Player2) ||
		g.CanSeePlayerInformation(Player2, Player4) || g.CanSeePlayerInformation(Player3, Player2) {
		t.Fatal("scoped control did not override turn control directionally")
	}
	g.SetChoiceContext(&struct{ value int }{value: 1})
	clone := g.Clone()
	if _, exists := clone.ChoiceContext(); exists {
		t.Fatal("clone retained transient live context")
	}
	encoded, err := json.Marshal(g.PlayerControlEffects)
	if err != nil {
		t.Fatal(err)
	}
	var restored []PlayerControlEffect
	if err := json.Unmarshal(encoded, &restored); err != nil || !reflect.DeepEqual(restored, g.PlayerControlEffects) {
		t.Fatal("data-only effect did not round-trip")
	}
	clone.SearchControlScopes[0].Controller = Player1
	clone.PlayerControlEffects[1].Controller = Player4
	clone.SkippedTurns = map[PlayerID]int{Player2: 2}
	second := clone.Clone()
	second.SkippedTurns[Player2] = 1
	if g.SearchControlScopes[0].Controller != Player4 || clone.SkippedTurns[Player2] != 2 {
		t.Fatal("clone aliased decision scopes or skipped turns")
	}
	g.SearchControlScopes = nil
	if g.DecisionMaker(Player2, DecisionRules) != Player3 {
		t.Fatal("turn control did not resume after scoped control")
	}
	g.Turn.TurnNumber++
	g.Turn.ActivePlayer = Player3
	g.BeginPlayerControlTurn()
	if len(g.PlayerControlEffects) != 1 || g.DecisionMaker(Player2, DecisionRules) != Player2 ||
		g.DecisionMaker(Player3, DecisionRules) != Player1 {
		t.Fatal("active duration expiry or pending next-turn identity failed")
	}
}

func TestPlayerControlNestedDecisionAndVisibility(t *testing.T) {
	g := NewGame([NumPlayers]PlayerConfig{})
	g.PlayerControlEffects = []PlayerControlEffect{
		{ID: 1, Player: Player2, Controller: Player3, Turn: 1},
	}
	g.SearchControlScopes = []SearchControlScope{{Player: Player3, Controller: Player1, Override: true}}
	if g.DecisionMaker(Player2, DecisionRules) != Player1 ||
		!g.CanSeePlayerInformation(Player1, Player2) || !g.CanSeePlayerInformation(Player3, Player2) ||
		g.CanSeePlayerInformation(Player2, Player3) || g.CanSeePlayerInformation(Player4, Player2) {
		t.Fatal("nested control conflated object actors with ultimate decision maker or visibility")
	}
	g.EndPlayerControls(Player3)
	g.Players[Player3].Eliminated = true
	if g.DecisionMaker(Player2, DecisionRules) != Player2 || g.CanSeePlayerInformation(Player1, Player2) {
		t.Fatal("departing controller left decision/information control behind")
	}
}
