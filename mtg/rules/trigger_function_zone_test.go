package rules

import (
	"testing"

	"github.com/natefinch/council4/cardgen"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/zone"
)

func TestCompiledGraveyardFunctionZoneCapturesExactSource(t *testing.T) {
	def := compiledReferenceRoleOriginal(t, "Squee, Goblin Nabob")
	if def.TriggeredAbilities[0].ZoneOfFunction != zone.Graveyard {
		t.Fatalf("zone metadata missing: %+v", def.TriggeredAbilities[0])
	}
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	card := addCardToGraveyard(g, game.Player2, def)
	event := game.Event{Kind: game.EventBeginningOfStep, Step: game.StepUpkeep, Player: game.Player2}
	if pending := graveyardStepSourceTriggers(g, event); len(pending) != 1 {
		t.Fatalf("captured %d triggers for card %v", len(pending), card)
	}
}

func TestCompiledRecurringGraveyardTriggers(t *testing.T) {
	for _, tc := range []struct {
		name   string
		step   game.Step
		toHand bool
		pay    bool
		tapped bool
	}{
		{"Squee, Goblin Nabob", game.StepUpkeep, true, false, false},
		{"Flamewake Phoenix", game.StepBeginningOfCombat, false, true, false},
		{"Lightning Phoenix", game.StepEnd, false, true, false},
		{"Silversmote Ghoul", game.StepEnd, false, false, true},
	} {
		def := compiledReferenceRoleOriginal(t, tc.name)
		for _, scenario := range []string{"matching", "wrong player", "wrong step", "wrong zone", "condition below", "declined", "unaffordable", "condition changed"} {
			t.Run(tc.name+"/"+scenario, func(t *testing.T) {
				g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
				card := addCardToGraveyard(g, game.Player2, def)
				decoy := addCardToHand(g, game.Player1, def)
				version := g.CardInstances[card].ZoneVersion
				power := 4
				amount := 3
				if scenario == "condition below" {
					power, amount = 3, 2
				}
				ferocious := addCombatPermanent(g, game.Player2, vanillaCreature("Ferocious", power, power))
				g.AppendEvent(game.Event{Kind: game.EventLifeGained, Player: game.Player2, Amount: amount})
				g.AppendEvent(game.Event{Kind: game.EventDamageDealt, DamageRecipient: game.DamageRecipientPlayer, Player: game.Player3, Amount: amount})
				g.Players[game.Player2].ManaPool.Add(mana.R, 1)
				if scenario == "unaffordable" {
					g.Players[game.Player2].ManaPool.Empty()
				}
				event := game.Event{Kind: game.EventBeginningOfStep, Step: tc.step, Player: game.Player2, Controller: game.Player1}
				if scenario == "wrong player" {
					event.Player = game.Player1
				}
				if scenario == "wrong step" {
					event.Step = game.StepDraw
				}
				if scenario == "wrong zone" && !moveCardBetweenZonesWithPlacement(g, game.Player2, card, zone.Graveyard, zone.Hand, false) {
					t.Fatal("failed to move source out of its function zone")
				}
				emitEvent(g, event)
				engine := NewEngine(nil)
				agents := [game.NumPlayers]PlayerAgent{}
				agents[game.Player2] = &repeatLandChoiceAgent{mayAnswers: []bool{scenario != "declined"}}
				queued := engine.putTriggeredAbilitiesOnStackWithChoices(g, agents, &TurnLog{})
				wantQueued := scenario != "wrong player" && scenario != "wrong step" && scenario != "wrong zone" &&
					(scenario != "condition below" || tc.toHand)
				if queued != wantQueued {
					t.Fatalf("queued=%v want=%v", queued, wantQueued)
				}
				if !queued {
					return
				}
				obj, ok := g.Stack.Peek()
				if !ok || obj.Controller != game.Player2 || obj.SourceCardID != card ||
					obj.SourceZone != zone.Graveyard || obj.SourceZoneVersion != version {
					t.Fatal("lost original owner/card/zone/version")
				}
				if scenario == "condition changed" && tc.name == "Flamewake Phoenix" {
					if !movePermanentToZone(g, ferocious, zone.Exile) {
						t.Fatal("failed to remove ferocious")
					}
				}
				engine.resolveTopOfStackWithChoices(g, agents, &TurnLog{})
				declined := scenario == "declined" && (tc.toHand || tc.pay)
				unaffordable := scenario == "unaffordable" && tc.pay
				conditionChanged := scenario == "condition changed" && tc.name == "Flamewake Phoenix"
				wantReturned := !declined && !unaffordable && !conditionChanged
				returned := g.Players[game.Player2].Hand.Contains(card)
				if !tc.toHand {
					permanent, exists := findPermanentByCardID(g, card)
					returned = exists
					if exists && (permanent.Controller != game.Player2 || permanent.Tapped != tc.tapped) {
						t.Fatal("return lost controller or tapped state")
					}
				}
				if returned != wantReturned || !g.Players[game.Player1].Hand.Contains(decoy) {
					t.Fatalf("returned=%v want=%v; same-name decoy must remain untouched", returned, wantReturned)
				}
			})
		}
	}
}

func TestGraveyardTriggerSnapshotSurvivesDepartureWithoutRebinding(t *testing.T) {
	def := compileUnlessCard(t, cardgen.ScryfallCard{
		Name: "Snapshot Source", Layout: "normal", TypeLine: "Creature", Power: new("1"), Toughness: new("1"),
		OracleText: "At the beginning of your upkeep, return this card from your graveyard to your hand. You gain 1 life.",
	})
	for _, timing := range []string{"before queue", "before resolution", "copied stack"} {
		t.Run(timing, func(t *testing.T) {
			g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
			card := addCardToGraveyard(g, game.Player2, def)
			decoy := addCardToHand(g, game.Player2, def)
			version := g.CardInstances[card].ZoneVersion
			emitEvent(g, game.Event{Kind: game.EventBeginningOfStep, Step: game.StepUpkeep, Player: game.Player2})
			leaveAndReturn := func() {
				t.Helper()
				if !moveCardBetweenZonesWithPlacement(g, game.Player2, card, zone.Graveyard, zone.Exile, false) ||
					!moveCardBetweenZonesWithPlacement(g, game.Player2, card, zone.Exile, zone.Graveyard, false) {
					t.Fatal("failed to change source incarnation")
				}
			}
			if timing == "before queue" {
				leaveAndReturn()
				g = g.Clone()
			}
			engine := NewEngine(nil)
			if !engine.putTriggeredAbilitiesOnStack(g) {
				t.Fatal("captured ability vanished with its source")
			}
			obj, ok := g.Stack.Peek()
			if !ok || obj.SourceZoneVersion != version || obj.SourceZone != zone.Graveyard {
				t.Fatal("captured original source incarnation was rebound")
			}
			if timing != "before queue" {
				leaveAndReturn()
			}
			if timing == "copied stack" {
				copied := game.NewStackObjectCopy(obj, g.IDGen.Next())
				_, _ = g.Stack.Pop()
				g.Stack.Push(copied)
			}
			engine.resolveTopOfStack(g, &TurnLog{})
			if !g.Players[game.Player2].Graveyard.Contains(card) || !g.Players[game.Player2].Hand.Contains(decoy) ||
				g.Players[game.Player2].Life != 41 {
				t.Fatal("departed trigger must resolve independently, without moving a new incarnation or decoy")
			}
		})
	}
}
