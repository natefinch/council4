package rules

import (
	"testing"

	"github.com/natefinch/council4/cardgen"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/zone"
)

func compiledObservationBronco(t *testing.T) *game.CardDef {
	t.Helper()
	return compileUnlessCard(t, cardgen.ScryfallCard{
		Name: "Caustic Bronco", Layout: "normal", TypeLine: "Creature — Horse", ManaCost: "{1}{B}",
		Power: new("2"), Toughness: new("2"),
		OracleText: "Whenever this creature attacks, reveal the top card of your library and put it into your hand. You lose life equal to that card's mana value if this creature isn't saddled. Otherwise, each opponent loses that much life.\nSaddle 3 (Tap any number of other creatures you control with total power 3 or more: This Mount becomes saddled until end of turn. Saddle only as a sorcery.)",
	})
}

func TestCompiledRevealPostMoveManaValueIsolation(t *testing.T) {
	bronco := compiledObservationBronco(t)
	divination := compileUnlessCard(t, cardgen.ScryfallCard{
		Name: "Divination", Layout: "normal", TypeLine: "Sorcery", ManaCost: "{2}{U}", OracleText: "Draw two cards.",
	})
	forest := compileUnlessCard(t, cardgen.ScryfallCard{
		Name: "Forest", Layout: "normal", TypeLine: "Basic Land — Forest", OracleText: "{T}: Add {G}.",
	})
	for _, outcome := range []string{"repeat", "cloned pending trigger", "replaced hand move", "empty", "known zero"} {
		t.Run(outcome, func(t *testing.T) {
			g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
			source := addCombatPermanent(g, game.Player1, bronco)
			observed := divination
			if outcome == "known zero" {
				observed = forest
			}
			if outcome == "replaced hand move" {
				g.ReplacementEffects = append(g.ReplacementEffects, game.ReplacementEffect{
					ID: g.IDGen.Next(), MatchEvent: game.EventZoneChanged,
					MatchFromZone: true, FromZone: zone.Library,
					MatchToZone: true, ToZone: zone.Hand, ReplaceToZone: zone.Graveyard,
				})
			}
			iterations := 1
			if outcome == "repeat" {
				iterations = 2
			}
			wantLife := 40
			for range iterations {
				card := addCardToLibrary(g, game.Player1, observed)
				if outcome == "empty" {
					g.Players[game.Player1].Library.Remove(card)
				}
				engine := NewEngine(nil)
				if !sixAttacks(g, engine, source) {
					t.Fatal("compiled attack trigger did not fire")
				}
				if outcome == "cloned pending trigger" {
					g = g.Clone()
				}
				obj, ok := g.Stack.Peek()
				if !ok {
					t.Fatal("pending actual trigger missing")
				}
				const scalar = "sequence-effect-0-mana-value"
				obj.ResolvedAmounts = map[string]int{scalar: 91, "other": 92}
				obj.ResolutionResults = map[string]game.InstructionResolutionResult{scalar: {Succeeded: true, Amount: 91}}
				obj.ResolvedExcessDamage = map[string]int{scalar: 93}
				engine.resolveTopOfStack(g, &TurnLog{})
				if outcome != "empty" {
					wantLife -= observed.ManaValue()
				}
				if g.Players[game.Player1].Life != wantLife {
					t.Errorf("life=%d, want %d; current observation confused with parent, prior iteration, or hand success",
						g.Players[game.Player1].Life, wantLife)
				}
				inHand := outcome != "empty" && outcome != "replaced hand move"
				if g.Players[game.Player1].Hand.Contains(card) != inHand ||
					g.Players[game.Player1].Graveyard.Contains(card) != (outcome == "replaced hand move") {
					t.Error("actual observed-card destination did not match the real move/replacement")
				}
				if obj.ResolvedAmounts[scalar] != 91 || obj.ResolvedAmounts["other"] != 92 ||
					obj.ResolutionResults[scalar].Amount != 91 || obj.ResolvedExcessDamage[scalar] != 93 ||
					len(obj.LocalLinkedProducts) != 0 || len(g.LinkedObjects) != 0 {
					t.Error("sequence scalar/identity scope leaked or failed to restore enclosing products")
				}
				for player := game.Player2; player <= game.Player4; player++ {
					if g.Players[player].Life != 40 {
						t.Error("unsaddled trigger changed an opponent's life")
					}
				}
			}
		})
	}
}

func TestCompiledRevealPostMoveManaValueAttack(t *testing.T) {
	bronco := compiledObservationBronco(t)
	divination := compileUnlessCard(t, cardgen.ScryfallCard{
		Name: "Divination", Layout: "normal", TypeLine: "Sorcery", ManaCost: "{2}{U}", OracleText: "Draw two cards.",
	})
	if len(bronco.TriggeredAbilities) != 1 || len(bronco.ActivatedAbilities) != 1 ||
		bronco.TriggeredAbilities[0].Trigger.Pattern.Event != game.EventAttackerDeclared {
		t.Fatal("complete compiled Bronco lost its attack trigger or Saddle")
	}
	for _, saddled := range []bool{false, true} {
		name := "unsaddled"
		if saddled {
			name = "saddled"
		}
		t.Run(name, func(t *testing.T) {
			g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
			for _, player := range g.Players {
				player.Life = 40
			}
			card := addCardToLibrary(g, game.Player1, divination)
			decoy := addCardToLibrary(g, game.Player3, divination)
			source := addCombatPermanent(g, game.Player1, bronco)
			source.Saddled = saddled
			engine := NewEngine(nil)
			if !sixAttacks(g, engine, source) || g.Stack.Size() != 1 {
				t.Fatal("actual attack event did not put the compiled trigger on the stack")
			}
			obj, ok := g.Stack.Peek()
			if !ok || obj.Controller != game.Player1 || obj.SourceID != source.ObjectID {
				t.Fatal("compiled attack trigger has the wrong source/controller")
			}
			engine.resolveTopOfStack(g, &TurnLog{})
			if !g.Players[game.Player1].Hand.Contains(card) || g.Players[game.Player1].Library.Contains(card) ||
				g.CardInstances[card].Owner != game.Player1 || !g.Players[game.Player3].Library.Contains(decoy) {
				t.Error("the actual revealed card did not reach its owner's hand, or another library was changed")
			}
			lost := 0
			for playerID, player := range g.Players {
				want := 40
				if (game.PlayerID(playerID) == game.Player1) != saddled {
					want -= divination.ManaValue()
					lost++
				}
				if player.Life != want {
					t.Errorf("saddled=%v player=%v life=%d, want %d from actual compiled mana value %d",
						saddled, playerID, player.Life, want, divination.ManaValue())
				}
			}
			reveals, lifeEvents := 0, 0
			for _, event := range g.Events {
				if event.Kind == game.EventCardRevealed {
					reveals++
					if event.CardID != card || event.Player != game.Player1 {
						t.Errorf("wrong public revealed-card event: %#v", event)
					}
				}
				if event.Kind == game.EventLifeLost {
					lifeEvents++
					if event.Amount != divination.ManaValue() || (event.Player == game.Player1) == saddled {
						t.Errorf("wrong postmove mana-value life event: %#v", event)
					}
				}
			}
			if reveals != 1 || lifeEvents != lost {
				t.Errorf("reveal/life events=%d/%d, want 1/%d", reveals, lifeEvents, lost)
			}
		})
	}
}
