package rules

import (
	"strings"
	"testing"

	"github.com/natefinch/council4/mtg/cards"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/id"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
)

// Rebound's resolution exile competes with other graveyard replacements
// (#3535).
//
// Source review:
//   - Comprehensive Rules effective 2026-09-25,
//     https://media.wizards.com/2026/downloads/MagicCompRules%2020260925.txt
//     (sha256 8d860e451f20f38865b725b42d82feb714c725373dd8f3b32b8652b3eeb070ca):
//     702.88a ("Rebound" means "If this spell was cast from your hand,
//     instead of putting it into your graveyard as it resolves, exile it and,
//     at the beginning of your next upkeep, you may cast this card from exile
//     without paying its mana cost."), 614.15 and 616.1a (only
//     self-replacement effects are applied first), 616.1/616.1e-f (the
//     affected object's controller chooses one applicable replacement, then
//     repeats with the modified event) and 400.7 (a card that changes zones
//     becomes a new object).
//   - Trumpeting Herd, oracle_id 69501650-ed48-4ebf-9287-b0338c5bc5d5,
//     Scryfall rulings 2019-06-14 for b0f3b68e-f616-4687-bc2d-075165162cd1:
//     "Create a 3/3 green Elephant creature token. Rebound"; a countered
//     rebound spell doesn't resolve, so none of its effects happen, including
//     rebound, and it is put into its owner's graveyard; the recast is
//     optional and a card left in exile gets no later chance.
//   - Staggershock, oracle_id 056c3b7d-b603-40b8-8404-18c2eb7e7129 (print
//     56976d52-053d-49e9-94c0-f21dc13029ac): "Staggershock deals 2 damage to
//     any target. Rebound", used for the all-targets-illegal route (CR 608.2b:
//     the spell doesn't resolve and is put into its owner's graveyard).
//   - Dauthi Voidwalker, oracle_id f1c2dbe2-fbe0-4058-bdf1-91d1b1832786: "If a
//     card would be put into an opponent's graveyard from anywhere, instead
//     exile it with a void counter on it."
//
// Rebound is a static ability of the spell, not a self-replacement, so with an
// opponent's Dauthi Voidwalker the resolving spell's controller chooses which
// applies. The delayed recast exists only when rebound itself exiled the card.

func reboundReplacementRegistry(t *testing.T) (*cards.Registry, *game.CardDef, *game.CardDef) {
	t.Helper()
	registry := cards.NewDefaultRegistry()
	herd := registry.Lookup("Trumpeting Herd")
	dauthi := registry.Lookup("Dauthi Voidwalker")
	if herd == nil || dauthi == nil || !herd.HasKeyword(game.Rebound) {
		t.Fatal("registered Trumpeting Herd with rebound or Dauthi Voidwalker missing")
	}
	return registry, herd, dauthi
}

// castTrumpetingHerd has Player1 cast registered Trumpeting Herd from hand.
func castTrumpetingHerd(t *testing.T, g *game.Game, engine *Engine, herd *game.CardDef) (id.ID, *game.StackObject) {
	t.Helper()
	cardID := addCardToHand(g, game.Player1, herd)
	for range 4 {
		addBasicLandPermanent(g, game.Player1, types.Forest)
	}
	giveMainPhasePriority(g, game.Player1)
	if !engine.applyAction(g, game.Player1, action.CastSpell(cardID, nil, 0, nil)) {
		t.Fatal("Player1 could not cast Trumpeting Herd from hand")
	}
	obj, ok := g.Stack.Peek()
	if !ok || obj.SourceID != cardID || obj.SourceZone != zone.Hand {
		t.Fatalf("stack object = %+v, want Trumpeting Herd cast from hand", obj)
	}
	return cardID, obj
}

// reboundOrderAgents makes chooser pick rebound or the other offered
// replacement, recording each replacement request.
func reboundOrderAgents(t *testing.T, chooser game.PlayerID, chooseRebound bool, requests *[]game.ChoiceRequest) [game.NumPlayers]PlayerAgent {
	t.Helper()
	var agents [game.NumPlayers]PlayerAgent
	for _, player := range []game.PlayerID{game.Player1, game.Player2, game.Player3, game.Player4} {
		agents[player] = &authoredCopyChoiceAgent{choose: func(request game.ChoiceRequest) []int {
			if request.Kind != game.ChoiceReplacement {
				return request.DefaultSelection
			}
			if player != chooser {
				t.Fatalf("%v was asked to order replacements, want %v", player, chooser)
			}
			*requests = append(*requests, request)
			for _, option := range request.Options {
				if (option.Label == reboundReplacementLabel) == chooseRebound {
					return []int{option.Index}
				}
			}
			t.Fatalf("replacement options %+v lack the requested choice", request.Options)
			return nil
		}}
	}
	return agents
}

func requireReboundTwoWayChoice(t *testing.T, requests []game.ChoiceRequest, chooser game.PlayerID, other string) {
	t.Helper()
	if len(requests) != 1 || requests[0].Player != chooser || len(requests[0].Options) != 2 {
		t.Fatalf("replacement requests = %+v, want one two-way choice for %v", requests, chooser)
	}
	if requests[0].Options[0].Label != reboundReplacementLabel || requests[0].Options[1].Label != other {
		t.Fatalf("replacement labels = %q, %q, want rebound then %q",
			requests[0].Options[0].Label, requests[0].Options[1].Label, other)
	}
}

func controlledElephants(g *game.Game, player game.PlayerID) int {
	count := 0
	for _, permanent := range g.Battlefield {
		if permanent.Controller == player && permanent.TokenDef != nil && permanent.TokenDef.Name == "Elephant" {
			count++
		}
	}
	return count
}

// assertReboundUpkeep runs Player1's next-upkeep rebound and reports whether
// the card was recast from exile.
func assertReboundUpkeep(t *testing.T, g *game.Game, engine *Engine, cardID id.ID, wantRecast bool) {
	t.Helper()
	stackBefore := g.Stack.Size()
	runReboundUpkeep(engine, g, game.Player1, [game.NumPlayers]PlayerAgent{})
	if reboundTriggerPending(g, cardID) {
		t.Fatal("rebound tracking survived the upkeep")
	}
	obj, ok := g.Stack.Peek()
	recast := g.Stack.Size() == stackBefore+1 && ok && obj.SourceID == cardID && obj.SourceZone == zone.Exile
	if recast != wantRecast {
		t.Fatalf("rebound recast = %v (stack %+v), want %v", recast, obj, wantRecast)
	}
}

func TestReboundCompetesWithDauthiVoidwalker(t *testing.T) {
	t.Parallel()
	for _, chooseRebound := range []bool{true, false} {
		name := "Dauthi first"
		if chooseRebound {
			name = "rebound first"
		}
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			registry, herd, dauthi := reboundReplacementRegistry(t)
			g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
			engine := NewEngineWithPrograms(nil, registry.Programs())
			resolveDauthiVoidwalker(t, g, engine, game.Player2, dauthi)
			cardID, _ := castTrumpetingHerd(t, g, engine, herd)
			var requests []game.ChoiceRequest
			engine.resolveTopOfStackWithChoices(g, reboundOrderAgents(t, game.Player1, chooseRebound, &requests), &TurnLog{})
			requireReboundTwoWayChoice(t, requests, game.Player1, g.ReplacementEffects[0].Description)
			if controlledElephants(g, game.Player1) != 1 {
				t.Fatal("Trumpeting Herd did not create its Elephant")
			}
			// Whichever applies first, the other no longer applies: the card is
			// no longer going to a graveyard (CR 616.1f).
			voidCounters := 1
			if chooseRebound {
				voidCounters = 0
			}
			if !g.Players[game.Player1].Exile.Contains(cardID) || g.Players[game.Player1].Graveyard.Contains(cardID) {
				t.Fatal("Trumpeting Herd did not end in its owner's exile")
			}
			if got := g.ExileCounterCount(cardID, counter.Void); got != voidCounters {
				t.Fatalf("void counters = %d, want %d", got, voidCounters)
			}
			if tracked := reboundTriggerPending(g, cardID); tracked != chooseRebound {
				t.Fatalf("rebound tracked = %v, want %v", tracked, chooseRebound)
			}
			assertReboundUpkeep(t, g, engine, cardID, chooseRebound)
			if chooseRebound {
				// The recast is not from hand, so Dauthi alone exiles it.
				engine.resolveTopOfStackWithChoices(g, reboundOrderAgents(t, game.Player1, false, &requests), &TurnLog{})
				if controlledElephants(g, game.Player1) != 2 || g.ExileCounterCount(cardID, counter.Void) != 1 ||
					len(requests) != 1 || reboundTriggerPending(g, cardID) {
					t.Fatal("rebound recast did not resolve once and fall to Dauthi Voidwalker")
				}
			}
		})
	}
}

func TestReboundReplacementDefaultWithoutAgentAndClone(t *testing.T) {
	t.Parallel()
	registry, herd, dauthi := reboundReplacementRegistry(t)
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngineWithPrograms(nil, registry.Programs())
	resolveDauthiVoidwalker(t, g, engine, game.Player2, dauthi)
	cardID, _ := castTrumpetingHerd(t, g, engine, herd)
	replay := g.Clone()
	engine.resolveTopOfStack(g, &TurnLog{})
	decisions := g.ReplacementDecisions
	if len(decisions) == 0 || !decisions[len(decisions)-1].UsedFallback ||
		decisions[len(decisions)-1].Options[0] != reboundReplacementLabel {
		t.Fatalf("replacement decisions = %+v, want a rebound-first fallback", decisions)
	}
	if !reboundTriggerPending(g, cardID) || g.ExileCounterCount(cardID, counter.Void) != 0 {
		t.Fatal("fallback did not apply rebound")
	}
	// A clone taken before resolution makes its own choice independently.
	var requests []game.ChoiceRequest
	engine.resolveTopOfStackWithChoices(replay, reboundOrderAgents(t, game.Player1, false, &requests), &TurnLog{})
	requireReboundTwoWayChoice(t, requests, game.Player1, replay.ReplacementEffects[0].Description)
	if reboundTriggerPending(replay, cardID) || replay.ExileCounterCount(cardID, counter.Void) != 1 {
		t.Fatal("cloned game did not resolve its own Dauthi-first choice")
	}
	if !reboundTriggerPending(g, cardID) {
		t.Fatal("clone's resolution changed the original's rebound tracking")
	}
	// A clone taken after rebound keeps the delayed recast.
	after := g.Clone()
	assertReboundUpkeep(t, after, engine, cardID, true)
	assertReboundUpkeep(t, g, engine, cardID, true)
}

func TestReboundCounteredOrFizzledSpellNeverRebounds(t *testing.T) {
	t.Parallel()
	for _, withDauthi := range []bool{false, true} {
		name := "no replacement"
		if withDauthi {
			name = "Dauthi Voidwalker"
		}
		t.Run("countered/"+name, func(t *testing.T) {
			t.Parallel()
			registry, herd, dauthi := reboundReplacementRegistry(t)
			g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
			engine := NewEngineWithPrograms(nil, registry.Programs())
			if withDauthi {
				resolveDauthiVoidwalker(t, g, engine, game.Player2, dauthi)
			}
			cardID, obj := castTrumpetingHerd(t, g, engine, herd)
			addEffectSpellToStack(g, game.Player2, game.CounterObject{Object: game.TargetStackObjectReference(0)},
				[]game.Target{game.StackObjectTarget(obj.ID)})
			engine.resolveTopOfStack(g, &TurnLog{})
			if g.Stack.Size() != 0 || controlledElephants(g, game.Player1) != 0 {
				t.Fatal("countered Trumpeting Herd resolved")
			}
			assertNoRebound(t, g, cardID, withDauthi)
		})
		t.Run("fizzled/"+name, func(t *testing.T) {
			t.Parallel()
			registry, _, dauthi := reboundReplacementRegistry(t)
			staggershock := registry.Lookup("Staggershock")
			if staggershock == nil || !staggershock.HasKeyword(game.Rebound) {
				t.Fatal("registered Staggershock with rebound missing")
			}
			g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
			engine := NewEngineWithPrograms(nil, registry.Programs())
			if withDauthi {
				resolveDauthiVoidwalker(t, g, engine, game.Player2, dauthi)
			}
			creature := addCreaturePermanent(g, game.Player3)
			cardID := addCardToHand(g, game.Player1, staggershock)
			for range 3 {
				addBasicLandPermanent(g, game.Player1, types.Mountain)
			}
			giveMainPhasePriority(g, game.Player1)
			if !engine.applyAction(g, game.Player1, action.CastSpell(cardID,
				[]game.Target{game.PermanentTarget(creature.ObjectID)}, 0, nil)) {
				t.Fatal("Player1 could not cast Staggershock")
			}
			if !movePermanentToZone(g, creature, zone.Exile) {
				t.Fatal("could not remove Staggershock's target")
			}
			engine.resolveTopOfStack(g, &TurnLog{})
			assertNoRebound(t, g, cardID, withDauthi)
		})
	}
}

func assertNoRebound(t *testing.T, g *game.Game, cardID id.ID, withDauthi bool) {
	t.Helper()
	owner := g.Players[game.Player1]
	if withDauthi {
		if !owner.Exile.Contains(cardID) || g.ExileCounterCount(cardID, counter.Void) != 1 {
			t.Fatal("Dauthi Voidwalker did not exile the unresolved spell with a void counter")
		}
	} else if !owner.Graveyard.Contains(cardID) {
		t.Fatal("unresolved rebound spell did not go to its owner's graveyard")
	}
	if reboundTriggerPending(g, cardID) {
		t.Fatalf("delayed triggers = %+v for an unresolved spell", g.DelayedTriggers)
	}
}

func TestReboundStolenSpellAndCopyDoNotRebound(t *testing.T) {
	t.Parallel()
	t.Run("stolen", func(t *testing.T) {
		t.Parallel()
		registry, herd, _ := reboundReplacementRegistry(t)
		g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
		engine := NewEngineWithPrograms(nil, registry.Programs())
		cardID, obj := castTrumpetingHerd(t, g, engine, herd)
		stealStackSpell(t, g, engine, obj.ID, game.Player2)
		engine.resolveTopOfStack(g, &TurnLog{})
		// Player2 didn't cast it from their hand, so "your hand" fails.
		if controlledElephants(g, game.Player2) != 1 || !g.Players[game.Player1].Graveyard.Contains(cardID) ||
			reboundTriggerPending(g, cardID) {
			t.Fatal("stolen rebound spell rebounded or left its owner's graveyard")
		}
	})
	t.Run("copy", func(t *testing.T) {
		t.Parallel()
		registry, herd, _ := reboundReplacementRegistry(t)
		g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
		engine := NewEngineWithPrograms(nil, registry.Programs())
		cardID, obj := castTrumpetingHerd(t, g, engine, herd)
		copied := copyStackObject(engine, g, obj)
		if copied == nil || !copied.Copy {
			t.Fatalf("top of stack = %+v, want a copy of Trumpeting Herd", copied)
		}
		engine.resolveTopOfStack(g, &TurnLog{})
		if controlledElephants(g, game.Player1) != 1 || reboundTriggerPending(g, cardID) {
			t.Fatal("copy of a rebound spell did not resolve without rebound")
		}
		engine.resolveTopOfStack(g, &TurnLog{})
		if controlledElephants(g, game.Player1) != 2 || !g.Players[game.Player1].Exile.Contains(cardID) {
			t.Fatal("original rebound spell did not rebound after its copy")
		}
		if !reboundTriggerPending(g, cardID) {
			t.Fatal("original rebound spell was not scheduled")
		}
	})
}

func TestReboundEndsWhenCardLeavesExile(t *testing.T) {
	t.Parallel()
	registry, herd, _ := reboundReplacementRegistry(t)
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngineWithPrograms(nil, registry.Programs())
	cardID, _ := castTrumpetingHerd(t, g, engine, herd)
	engine.resolveTopOfStack(g, &TurnLog{})
	if !reboundTriggerPending(g, cardID) {
		t.Fatal("Trumpeting Herd did not rebound")
	}
	// The card returned to exile is a new object (CR 400.7) that rebound's
	// delayed trigger can't find.
	if !moveCardBetweenZones(g, game.Player1, cardID, zone.Exile, zone.Hand) ||
		!moveCardBetweenZones(g, game.Player1, cardID, zone.Hand, zone.Exile) {
		t.Fatal("could not move Trumpeting Herd out of exile and back")
	}
	assertReboundUpkeep(t, g, engine, cardID, false)
}

func reboundBuybackFixture() *game.CardDef {
	def := reboundSorcery()
	def.Name = "Rebound Buyback Sorcery"
	def.StaticAbilities = append(def.StaticAbilities, game.StaticAbility{
		KeywordAbilities: []game.KeywordAbility{game.BuybackKeyword{Cost: cost.Mana{cost.O(1)}}},
	})
	return def
}

func TestReboundCompetesWithBuyback(t *testing.T) {
	t.Parallel()
	for _, chooseRebound := range []bool{true, false} {
		name := "buyback first"
		if chooseRebound {
			name = "rebound first"
		}
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
			engine := NewEngine(nil)
			cardID := addCardToHand(g, game.Player1, reboundBuybackFixture())
			g.Players[game.Player1].ManaPool.Add(mana.G, 1)
			g.Players[game.Player1].ManaPool.Add(mana.C, 1)
			giveMainPhasePriority(g, game.Player1)
			payBuyback := [game.NumPlayers]PlayerAgent{game.Player1: &authoredCopyChoiceAgent{
				choose: func(request game.ChoiceRequest) []int {
					if request.Kind == game.ChoicePayment {
						for _, option := range request.Options {
							if strings.Contains(option.Label, "Buyback") {
								return []int{option.Index}
							}
						}
						t.Fatalf("buyback payment branch missing: %+v", request)
					}
					return request.DefaultSelection
				},
			}}
			if !engine.applyActionWithChoices(g, game.Player1, action.CastSpell(cardID, nil, 0, nil), payBuyback, &TurnLog{}) {
				t.Fatal("cast with buyback failed")
			}
			if obj, ok := g.Stack.Peek(); !ok || !obj.BuybackPaid {
				t.Fatalf("stack object = %+v, want buyback paid", obj)
			}
			var requests []game.ChoiceRequest
			engine.resolveTopOfStackWithChoices(g, reboundOrderAgents(t, game.Player1, chooseRebound, &requests), &TurnLog{})
			requireReboundTwoWayChoice(t, requests, game.Player1, "Buyback")
			owner := g.Players[game.Player1]
			tracked := reboundTriggerPending(g, cardID)
			if owner.Exile.Contains(cardID) != chooseRebound || owner.Hand.Contains(cardID) == chooseRebound ||
				tracked != chooseRebound {
				t.Fatalf("exile=%v hand=%v rebound=%v, want rebound %v",
					owner.Exile.Contains(cardID), owner.Hand.Contains(cardID), tracked, chooseRebound)
			}
		})
	}
}
