package rules

import (
	"slices"
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/id"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// Source review: default-registry Rest in Peace (oracle
// 087f9ad7-e74f-40e2-8102-1ed2925d0418) "If a card or token would be put into a
// graveyard from anywhere, exile it instead." and Progenitus (oracle
// 2d7e00b6-12f0-4b03-82a6-50e3d1b5395e) "If Progenitus would be put into a
// graveyard from anywhere, reveal Progenitus and shuffle it into its owner's
// library instead.", both matching the pinned corpus. Comprehensive Rules
// effective September 25, 2026: CR 101.4 (choices first, then the actions happen
// simultaneously), CR 514.1 (cleanup discards enough cards), CR 614.1
// (replacement effects apply as the event happens), CR 616.1 (the affected
// object's owner chooses which replacement applies), CR 701.9a/b (discard;
// the affected player chooses by default) and CR 701.50d (connive N counts the
// nonland cards discarded this way).

// progenitusDiscardAgents answers every replacement choice through recorder and
// every card selection by picking the options naming Progenitus, so each route
// discards exactly the two Progenitus.
func progenitusDiscardAgents(recorder replacementChoiceRecorder) [game.NumPlayers]PlayerAgent {
	var agents [game.NumPlayers]PlayerAgent
	for _, player := range []game.PlayerID{game.Player1, game.Player2, game.Player3, game.Player4} {
		replacement := recorder.agent(player).(authoredChoicePolicy)
		agents[player] = authoredChoicePolicy{choose: func(obs PlayerObservation, req game.ChoiceRequest) []int {
			if req.Kind == game.ChoiceReplacement {
				return replacement.choose(obs, req)
			}
			var picks []int
			for _, option := range req.Options {
				if option.Card.Exists && option.Card.Val.Name == "Progenitus" {
					picks = append(picks, option.Index)
				}
			}
			if len(picks) == 0 {
				return req.DefaultSelection
			}
			return picks
		}}
	}
	return agents
}

// TestSinglePlayerMultiCardDiscardDecidesReplacementsBeforeMoving proves every
// route that has one player discard several chosen cards decides each card's
// replacement while both cards are still in hand, then discards them as one
// simultaneous cohort. Player 2 controls Rest in Peace; player 1 owns and
// discards two Progenitus, so player 1 makes both replacement choices.
func TestSinglePlayerMultiCardDiscardDecidesReplacementsBeforeMoving(t *testing.T) {
	type route struct {
		name string
		// setup puts the two Progenitus in player 1's hand (or library, to be
		// drawn) and returns them plus the cards that must stay in hand.
		setup func(*game.Game, hiddenReplacementCards) (progenitus, kept []id.ID)
		run   func(*testing.T, *Engine, *game.Game, [game.NumPlayers]PlayerAgent)
		check func(*testing.T, *game.Game)
	}
	handWithKeeper := func(g *game.Game, set hiddenReplacementCards) ([]id.ID, []id.ID) {
		progenitus := []id.ID{addCardToHand(g, game.Player1, set.progenitus), addCardToHand(g, game.Player1, set.progenitus)}
		return progenitus, []id.ID{addCardToHand(g, game.Player1, discardableCard("Kept card"))}
	}
	var conniver *game.Permanent
	routes := []route{
		{
			name:  "discard two chosen cards",
			setup: handWithKeeper,
			run: func(_ *testing.T, engine *Engine, g *game.Game, agents [game.NumPlayers]PlayerAgent) {
				addEffectSpellToStack(g, game.Player1, game.Discard{Amount: game.Fixed(2), Player: game.ControllerReference()}, nil)
				engine.resolveTopOfStackWithChoices(g, agents, &TurnLog{})
			},
		},
		{
			name:  "discard any number then draw that many",
			setup: handWithKeeper,
			run: func(_ *testing.T, engine *Engine, g *game.Game, agents [game.NumPlayers]PlayerAgent) {
				addCardToLibrary(g, game.Player1, discardableCard("Library card"))
				addCardToLibrary(g, game.Player1, discardableCard("Library card"))
				addEffectSpellToStack(g, game.Player1, game.DiscardThenDraw{Player: game.ControllerReference()}, nil)
				engine.resolveTopOfStackWithChoices(g, agents, &TurnLog{})
			},
			check: func(t *testing.T, g *game.Game) {
				if got := g.Players[game.Player1].Hand.Size(); got != 3 {
					t.Fatalf("hand size after discarding two and drawing = %d, want 3", got)
				}
			},
		},
		{
			name: "connive 2",
			setup: func(g *game.Game, set hiddenReplacementCards) ([]id.ID, []id.ID) {
				conniver = addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Conniver",
					Types:     []types.Card{types.Creature},
					Power:     opt.Val(game.PT{Value: 1}),
					Toughness: opt.Val(game.PT{Value: 1}),
				}})
				kept := []id.ID{addCardToHand(g, game.Player1, discardableCard("Kept card"))}
				progenitus := []id.ID{addCardToLibrary(g, game.Player1, set.progenitus), addCardToLibrary(g, game.Player1, set.progenitus)}
				return progenitus, kept
			},
			run: func(t *testing.T, engine *Engine, g *game.Game, agents [game.NumPlayers]PlayerAgent) {
				obj := &game.StackObject{
					Kind:         game.StackTriggeredAbility,
					SourceID:     conniver.ObjectID,
					SourceCardID: conniver.CardInstanceID,
					Controller:   game.Player1,
				}
				// As during a turn (runTurn), replacement choices reach the agents.
				engine.setReplacementChoiceContext(g, agents, &TurnLog{})
				defer g.ClearChoiceContext()
				engine.resolveInstructionWithChoices(g, obj, &game.Instruction{Primitive: game.Connive{
					Object: game.SourcePermanentReference(),
					Player: game.ControllerReference(),
					Amount: game.Fixed(2),
				}}, agents, &TurnLog{})
			},
			check: func(t *testing.T, _ *game.Game) {
				if got := conniver.Counters.Get(counter.PlusOnePlusOne); got != 2 {
					t.Fatalf("+1/+1 counters = %d, want 2 for two nonland cards discarded", got)
				}
			},
		},
		{
			name: "cleanup discard to maximum hand size",
			setup: func(g *game.Game, set hiddenReplacementCards) ([]id.ID, []id.ID) {
				progenitus := []id.ID{addCardToHand(g, game.Player1, set.progenitus), addCardToHand(g, game.Player1, set.progenitus)}
				var kept []id.ID
				for range maximumHandSize {
					kept = append(kept, addCardToHand(g, game.Player1, discardableCard("Kept card")))
				}
				return progenitus, kept
			},
			run: func(_ *testing.T, engine *Engine, g *game.Game, agents [game.NumPlayers]PlayerAgent) {
				engine.setReplacementChoiceContext(g, agents, &TurnLog{})
				defer g.ClearChoiceContext()
				engine.discardToMaximumHandSizeWithChoices(g, game.Player1, agents, &TurnLog{})
			},
			check: func(t *testing.T, g *game.Game) {
				if got := g.Players[game.Player1].Hand.Size(); got != maximumHandSize {
					t.Fatalf("hand size after cleanup = %d, want %d", got, maximumHandSize)
				}
			},
		},
	}
	for _, rt := range routes {
		for _, pick := range []struct {
			label string
			want  zone.Type
		}{
			{label: "Progenitus", want: zone.Library},
			{label: "exile it instead", want: zone.Exile},
		} {
			t.Run(rt.name+"/"+pick.label, func(t *testing.T) {
				g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
				g.Turn.ActivePlayer = game.Player1
				engine := NewEngine(nil)
				set := hiddenReplacementRegistry(t)
				resolveHiddenReplacementPermanent(t, g, engine, game.Player2, set.restInPeace)
				resolveWholeStack(t, engine, g)
				progenitus, kept := rt.setup(g, set)
				pending := map[id.ID]game.PlayerID{progenitus[0]: game.Player1, progenitus[1]: game.Player1}
				var asked []game.PlayerID
				recorder := replacementChoiceRecorder{t: t, g: g, pending: pending, pick: pick.label, asked: &asked}
				start := len(g.Events)

				rt.run(t, engine, g, progenitusDiscardAgents(recorder))

				if want := []game.PlayerID{game.Player1, game.Player1}; !slices.Equal(asked, want) {
					t.Fatalf("replacement choosers = %v, want %v (the discarding owner, once per Progenitus)", asked, want)
				}
				assertOneDiscardCohort(t, g, start, map[game.PlayerID]int{game.Player1: 2})
				// Read each destination from its zone change: a later draw may
				// return a shuffled-in Progenitus to the hand.
				for _, cardID := range progenitus {
					var got []zone.Type
					for _, event := range g.Events[start:] {
						if event.Kind == game.EventZoneChanged && event.FromZone == zone.Hand && event.CardID == cardID {
							got = append(got, event.ToZone)
						}
					}
					if !slices.Equal(got, []zone.Type{pick.want}) {
						t.Fatalf("Progenitus %d moved from hand to %v, want [%v]", cardID, got, pick.want)
					}
				}
				for _, cardID := range kept {
					if !g.Players[game.Player1].Hand.Contains(cardID) {
						t.Fatalf("unchosen card %d left the hand", cardID)
					}
				}
				if rt.check != nil {
					rt.check(t, g)
				}
			})
		}
	}
}
