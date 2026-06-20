package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/types"
)

func TestResolveTopOfStackAppendsResolveLog(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	sourceID := addCardToHand(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Test Spell"}})
	stackObjectID := g.IDGen.Next()
	g.Stack.Push(&game.StackObject{
		ID:         stackObjectID,
		Kind:       game.StackSpell,
		SourceID:   sourceID,
		Controller: game.Player1,
	})
	engine := NewEngine(nil)
	log := TurnLog{}

	engine.resolveTopOfStack(g, &log)

	if !g.Stack.IsEmpty() {
		t.Fatal("stack is not empty after resolution")
	}
	if len(log.Resolves) != 1 {
		t.Fatalf("resolve logs = %d, want 1", len(log.Resolves))
	}
	got := log.Resolves[0]
	if got.StackObjectID != stackObjectID {
		t.Fatalf("stack object ID = %v, want %v", got.StackObjectID, stackObjectID)
	}
	if got.SourceID != sourceID {
		t.Fatalf("source ID = %v, want %v", got.SourceID, sourceID)
	}
	if got.Controller != game.Player1 {
		t.Fatalf("controller = %v, want %v", got.Controller, game.Player1)
	}
	if got.Kind != game.StackSpell {
		t.Fatalf("kind = %v, want %v", got.Kind, game.StackSpell)
	}
	if got.Result != "resolved" {
		t.Fatalf("result = %q, want %q", got.Result, "resolved")
	}
}

func TestResolveCreatureSpellMovesCardToBattlefield(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	spellID := addCardToHand(g, game.Player1, greenCreature())
	addBasicLandPermanent(g, game.Player1, types.Forest)
	g.Turn.Phase = game.PhasePrecombatMain
	g.Turn.Step = game.StepNone
	if !engine.applyAction(g, game.Player1, action.CastSpell(spellID, nil, 0, nil)) {
		t.Fatal("applyAction() = false, want true")
	}
	log := TurnLog{}

	engine.resolveTopOfStack(g, &log)

	if g.Stack.Size() != 0 {
		t.Fatalf("stack size = %d, want 0", g.Stack.Size())
	}
	if g.Players[game.Player1].Hand.Contains(spellID) {
		t.Fatal("creature card remained in hand")
	}
	if len(g.Battlefield) != 2 {
		t.Fatalf("battlefield permanents = %d, want 2", len(g.Battlefield))
	}
	permanent := g.Battlefield[1]
	if permanent.CardInstanceID != spellID {
		t.Fatalf("permanent card ID = %v, want %v", permanent.CardInstanceID, spellID)
	}
	if permanent.Owner != game.Player1 {
		t.Fatalf("permanent owner = %v, want %v", permanent.Owner, game.Player1)
	}
	if permanent.Controller != game.Player1 {
		t.Fatalf("permanent controller = %v, want %v", permanent.Controller, game.Player1)
	}
	if !permanent.SummoningSick {
		t.Fatal("creature permanent was not summoning sick")
	}
	if len(g.Events) < 2 {
		t.Fatal("missing permanent-enter event")
	}
	enter := g.Events[len(g.Events)-2]
	if enter.Kind != game.EventPermanentEnteredBattlefield ||
		!enter.EnterWasCast ||
		!enter.EnterHasCastController ||
		enter.EnterCastController != game.Player1 {
		t.Fatalf("enter event = %+v, want cast permanent enter", enter)
	}
	if len(log.Resolves) != 1 {
		t.Fatalf("resolve logs = %d, want 1", len(log.Resolves))
	}
	if log.Resolves[0].SourceID != spellID || log.Resolves[0].Controller != game.Player1 {
		t.Fatalf("resolve log = %+v, want source %v controller %v", log.Resolves[0], spellID, game.Player1)
	}
	if log.Resolves[0].Result != "battlefield" {
		t.Fatalf("resolve result = %q, want %q", log.Resolves[0].Result, "battlefield")
	}
}

func TestResolveSorcerySpellMovesCardToGraveyard(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	spellID := addCardToHand(g, game.Player1, greenSorcery())
	addBasicLandPermanent(g, game.Player1, types.Forest)
	g.Turn.Phase = game.PhasePrecombatMain
	g.Turn.Step = game.StepNone
	if !engine.applyAction(g, game.Player1, action.CastSpell(spellID, nil, 0, nil)) {
		t.Fatal("applyAction() = false, want true")
	}
	log := TurnLog{}

	engine.resolveTopOfStack(g, &log)

	if g.Stack.Size() != 0 {
		t.Fatalf("stack size = %d, want 0", g.Stack.Size())
	}
	if !g.Players[game.Player1].Graveyard.Contains(spellID) {
		t.Fatal("sorcery card was not moved to graveyard")
	}
	if len(log.Resolves) != 1 {
		t.Fatalf("resolve logs = %d, want 1", len(log.Resolves))
	}
	if log.Resolves[0].Result != "graveyard" {
		t.Fatalf("resolve result = %q, want %q", log.Resolves[0].Result, "graveyard")
	}
}
