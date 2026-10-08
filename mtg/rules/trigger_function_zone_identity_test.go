package rules

import (
	"testing"

	"github.com/natefinch/council4/cardgen"
	"github.com/natefinch/council4/mtg/game"
)

func TestGraveyardTriggerRefusesWrongSourceIdentity(t *testing.T) {
	def := compiledReferenceRoleOriginal(t, "Squee, Goblin Nabob")
	for _, corruption := range []string{"missing", "wrong ID", "wrong owner", "wrong event"} {
		t.Run(corruption, func(t *testing.T) {
			g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
			card := addCardToGraveyard(g, game.Player2, def)
			event := game.Event{Kind: game.EventBeginningOfStep, Step: game.StepUpkeep, Player: game.Player2}
			switch corruption {
			case "missing":
				delete(g.CardInstances, card)
			case "wrong ID":
				g.CardInstances[card].ID = g.IDGen.Next()
			case "wrong owner":
				g.CardInstances[card].Owner = game.Player3
			case "wrong event":
				event.Kind = game.EventLifeGained
			default:
				t.Fatal("unknown corruption")
			}
			if got := graveyardStepSourceTriggers(g, event); len(got) != 0 {
				t.Fatal("invalid source identity or event acquired a graveyard trigger")
			}
		})
	}
}

func TestBattlefieldAndGraveyardStepSourcesStayDistinct(t *testing.T) {
	ordinary := compileUnlessCard(t, cardgen.ScryfallCard{
		Name: "Ordinary Source", Layout: "normal", TypeLine: "Creature", Power: new("1"), Toughness: new("1"),
		OracleText: "At the beginning of your upkeep, you gain 1 life.",
	})
	recurring := compiledReferenceRoleOriginal(t, "Squee, Goblin Nabob")
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	live := addCombatPermanent(g, game.Player1, ordinary)
	live.Controller = game.Player2
	addCombatPermanent(g, game.Player2, recurring)
	addCardToGraveyard(g, game.Player2, ordinary)
	source := addCardToGraveyard(g, game.Player2, recurring)
	event := game.Event{Kind: game.EventBeginningOfStep, Step: game.StepUpkeep, Player: game.Player2, Controller: game.Player1}
	emitEvent(g, event)
	captured := g.Events[len(g.Events)-1].TriggeredAbilities
	if len(captured) != 2 {
		t.Fatalf("captured %d triggers, want exactly controlled battlefield and owned graveyard sources", len(captured))
	}
	seenLive, seenGraveyard := false, false
	for _, trigger := range captured {
		if trigger.Controller != game.Player2 {
			t.Fatal("step event controller replaced the source controller/owner")
		}
		seenLive = seenLive || trigger.SourceID == live.ObjectID
		seenGraveyard = seenGraveyard || trigger.SourceCardID == source
	}
	if !seenLive || !seenGraveyard {
		t.Fatal("default and explicit source zones were mixed")
	}
}
