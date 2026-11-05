package rules

// Source review (#3658, existing default-registry card defect; shared Engine
// operation): Trial of a Time Lord, pinned corpus
// 6da9ced0e19bd384771ab8f14051cc45a8dc90302082b875fa6e039d2399a1b4, oracle_id
// a5425d4f-cd6c-4240-89a4-dd13e0e1a95f, WHO 29 printing
// 52b68ed3-0d26-4334-8d0c-2d28ced8ee4c: "I, II, III — Exile target nontoken
// creature an opponent controls until this Saga leaves the battlefield. IV —
// Starting with you, each player votes for innocent or guilty. If guilty gets
// more votes, the owner of each card exiled with this Saga puts that card on the
// bottom of their library." Ruling 2023-10-13 (Scryfall rulings retrieved
// 2026-11-05): "The cards that didn't get put on the bottom of their owners'
// libraries will return to the battlefield as Trial of a Time Lord leaves the
// battlefield." CR effective 2026-09-25 (local copy SHA256
// 8d860e451f20f38865b725b42d82feb714c725373dd8f3b32b8652b3eeb070ca): 714.2c
// "I, II, III" is the same chapter ability for each number; 607.2a "exiled with
// [this object]" refers to cards in exile put there by the linked exiling
// ability, so every chapter resolution accumulates; 400.7 a card that left
// exile is a new object, and a new Saga occurrence is a different object;
// 701.38a votes start with the specified player in turn order; 610.3/610.3c
// cards still exiled return under their owners' control when the Saga leaves;
// 714.4 the Saga is sacrificed after chapter IV leaves the stack; 401.4 the
// owner of two or more cards put into the same library position at the same
// time arranges them in any order without revealing it; 702.140e/721 a merged
// permanent's components all move to exile, so each is "exiled with" the Saga.

import (
	"slices"
	"testing"

	"github.com/natefinch/council4/mtg/cards"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/id"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
)

const (
	trialInnocent = 0
	trialGuilty   = 1
)

func addTrialOfATimeLord(t *testing.T, g *game.Game) *game.Permanent {
	t.Helper()
	def := cards.NewDefaultRegistry().Lookup("Trial of a Time Lord")
	if def == nil {
		t.Fatal("Trial of a Time Lord is not registered")
	}
	return addCombatPermanent(g, game.Player1, def)
}

// trialChapterExiles adds one lore counter to saga, so its single opponent
// creature is the only legal target, and resolves the chapter.
func trialChapterExiles(t *testing.T, engine *Engine, g *game.Game, saga, victim *game.Permanent) {
	t.Helper()
	addCountersToPermanent(g, saga, counter.Lore, 1)
	if !engine.putTriggeredAbilitiesOnStack(g) {
		t.Fatalf("chapter %d did not trigger", saga.Counters.Get(counter.Lore))
	}
	engine.resolveTopOfStack(g, &TurnLog{})
	if !g.Players[victim.Owner].Exile.Contains(victim.CardInstanceID) {
		t.Fatalf("chapter %d did not exile its target", saga.Counters.Get(counter.Lore))
	}
}

// trialVerdict adds the fourth lore counter, resolves chapter IV with the
// scripted votes, and then applies state-based actions so the Saga is
// sacrificed (CR 714.4) and its remaining exiles return (CR 610.3).
func trialVerdict(t *testing.T, engine *Engine, g *game.Game, saga *game.Permanent, votes [game.NumPlayers]int) {
	t.Helper()
	addCountersToPermanent(g, saga, counter.Lore, 1)
	if !engine.putTriggeredAbilitiesOnStack(g) {
		t.Fatal("chapter IV did not trigger")
	}
	engine.resolveTopOfStackWithChoices(g, voteAgents(votes), &TurnLog{})
	if g.Stack.Size() != 0 {
		t.Fatalf("stack size %d after chapter IV, want empty", g.Stack.Size())
	}
	engine.applyStateBasedActions(g)
	if !g.Players[game.Player1].Graveyard.Contains(saga.CardInstanceID) {
		t.Fatal("Saga was not sacrificed after chapter IV")
	}
}

func assertTrialCardZone(t *testing.T, g *game.Game, cardID id.ID, want zone.Type) {
	t.Helper()
	if got, ok := cardZone(g, cardID); !ok || got != want {
		t.Fatalf("card %v zone = %v (ok=%v), want %v", cardID, got, ok, want)
	}
}

// TestTrialOfATimeLordGuiltyBottomsEveryChapterExile: the default-registry
// Trial exiles a different opponent creature with each of chapters I, II and
// III, across two owners. A guilty majority puts every card still exiled with
// the Saga on the bottom of its owner's library; the chapter I card that left
// exile and was exiled again as a new object, and an unrelated exiled card,
// stay in exile, and nothing returns when the Saga is sacrificed.
func TestTrialOfATimeLordGuiltyBottomsEveryChapterExile(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	saga := addTrialOfATimeLord(t, g)
	addCardToLibraryNamed(g, game.Player2, "P2 Library")
	addCardToLibraryNamed(g, game.Player3, "P3 Library")
	unrelated := addExiledInstant(g, game.Player4, disposalVanilla("Unrelated Exile", types.Creature))

	first := addCombatCreaturePermanent(g, game.Player2)
	trialChapterExiles(t, engine, g, saga, first)
	second := addCombatCreaturePermanent(g, game.Player3)
	trialChapterExiles(t, engine, g, saga, second)
	third := addCombatCreaturePermanent(g, game.Player2)
	trialChapterExiles(t, engine, g, saga, third)

	// The first card leaves exile and is exiled again by something else: it is
	// a new object, no longer exiled with this Saga (CR 400.7, 607.2a).
	if !moveCardBetweenZones(g, game.Player2, first.CardInstanceID, zone.Exile, zone.Graveyard) ||
		!moveCardBetweenZones(g, game.Player2, first.CardInstanceID, zone.Graveyard, zone.Exile) {
		t.Fatal("moving the first exiled card out of and back into exile failed")
	}

	trialVerdict(t, engine, g, saga, [game.NumPlayers]int{trialGuilty, trialInnocent, trialGuilty, trialGuilty})

	assertLibraryBottomSet(t, g, game.Player2, third.CardInstanceID)
	assertLibraryBottomSet(t, g, game.Player3, second.CardInstanceID)
	assertTrialCardZone(t, g, first.CardInstanceID, zone.Exile)
	assertTrialCardZone(t, g, unrelated, zone.Exile)
	for _, cardID := range []id.ID{first.CardInstanceID, second.CardInstanceID, third.CardInstanceID} {
		if permanentByCardID(g, cardID) != nil {
			t.Fatalf("card %v returned to the battlefield after the Saga left", cardID)
		}
	}
	if g.Players[game.Player2].Library.Size() != 2 || g.Players[game.Player3].Library.Size() != 2 {
		t.Fatal("bottomed cards did not stay in their owners' libraries")
	}
}

// TestTrialOfATimeLordWithoutGuiltyMajorityReturnsEveryExile: an innocent
// majority or a tie disposes of nothing, so every creature the three chapters
// exiled returns under its owner's control when the Saga is sacrificed (ruling
// 2023-10-13, CR 610.3c).
func TestTrialOfATimeLordWithoutGuiltyMajorityReturnsEveryExile(t *testing.T) {
	for name, votes := range map[string][game.NumPlayers]int{
		"innocent": {trialGuilty, trialInnocent, trialInnocent, trialInnocent},
		"tie":      {trialGuilty, trialInnocent, trialGuilty, trialInnocent},
	} {
		t.Run(name, func(t *testing.T) {
			g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
			engine := NewEngine(nil)
			saga := addTrialOfATimeLord(t, g)
			victims := []*game.Permanent{addCombatCreaturePermanent(g, game.Player2)}
			trialChapterExiles(t, engine, g, saga, victims[0])
			victims = append(victims, addCombatCreaturePermanent(g, game.Player3))
			trialChapterExiles(t, engine, g, saga, victims[1])
			victims = append(victims, addCombatCreaturePermanent(g, game.Player4))
			trialChapterExiles(t, engine, g, saga, victims[2])

			trialVerdict(t, engine, g, saga, votes)

			for _, victim := range victims {
				returned := permanentByCardID(g, victim.CardInstanceID)
				if returned == nil {
					t.Fatalf("card %v did not return when the Saga left", victim.CardInstanceID)
				}
				if returned.Controller != victim.Owner {
					t.Fatalf("card %v returned under %v, want owner %v", victim.CardInstanceID, returned.Controller, victim.Owner)
				}
			}
		})
	}
}

// TestTrialOfATimeLordVerdictIgnoresAnotherSagasExiles: "exiled with this Saga"
// is scoped to one Saga object (CR 607.2a). A second Trial's guilty verdict
// bottoms only the card it exiled; the first Trial's card stays exiled.
func TestTrialOfATimeLordVerdictIgnoresAnotherSagasExiles(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	other := addTrialOfATimeLord(t, g)
	saga := addTrialOfATimeLord(t, g)

	otherVictim := addCombatCreaturePermanent(g, game.Player2)
	trialChapterExiles(t, engine, g, other, otherVictim)
	victim := addCombatCreaturePermanent(g, game.Player3)
	trialChapterExiles(t, engine, g, saga, victim)
	// With no opponent creature left, chapters II and III have no legal target
	// and are removed from the stack as they would be put on it.
	for range 2 {
		addCountersToPermanent(g, saga, counter.Lore, 1)
		engine.putTriggeredAbilitiesOnStack(g)
		for g.Stack.Size() > 0 {
			engine.resolveTopOfStack(g, &TurnLog{})
		}
	}

	trialVerdict(t, engine, g, saga, [game.NumPlayers]int{trialGuilty, trialGuilty, trialGuilty, trialGuilty})

	assertLibraryBottomSet(t, g, game.Player3, victim.CardInstanceID)
	assertTrialCardZone(t, g, otherVictim.CardInstanceID, zone.Exile)
	if _, ok := permanentByObjectID(g, other.ObjectID); !ok {
		t.Fatal("the other Trial left the battlefield")
	}
}

// trialOwnerAgent votes, answers the commander command-zone option, and orders
// its cards arriving together on the bottom of its library (top card first),
// recording each ordering request it receives in asked.
type trialOwnerAgent struct {
	player  game.PlayerID
	vote    int
	order   []id.ID
	command bool
	asked   *[]game.PlayerID
	offered [][]id.ID
}

func (*trialOwnerAgent) ChooseAction(PlayerObservation, []action.Action) action.Action {
	return actionBuild.pass()
}

func (a *trialOwnerAgent) ChooseChoice(_ PlayerObservation, req game.ChoiceRequest) []int {
	switch req.Kind {
	case game.ChoiceVote:
		return []int{a.vote}
	case game.ChoiceMay:
		if a.command {
			return []int{1}
		}
		return []int{0}
	case game.ChoiceOrder:
		*a.asked = append(*a.asked, a.player)
		var offered []id.ID
		for _, option := range req.Options {
			offered = append(offered, option.Card.Val.CardID)
		}
		a.offered = append(a.offered, offered)
		var result []int
		for _, cardID := range a.order {
			for _, option := range req.Options {
				if option.Card.Exists && option.Card.Val.CardID == cardID {
					result = append(result, option.Index)
				}
			}
		}
		return result
	}
	return req.DefaultSelection
}

func trialOwnerAgents(asked *[]game.PlayerID) ([game.NumPlayers]PlayerAgent, [game.NumPlayers]*trialOwnerAgent) {
	var agents [game.NumPlayers]PlayerAgent
	var owners [game.NumPlayers]*trialOwnerAgent
	for player := range game.NumPlayers {
		owners[player] = &trialOwnerAgent{player: game.PlayerID(player), vote: trialGuilty, asked: asked}
		agents[player] = owners[player]
	}
	return agents, owners
}

func trialResolveVerdict(t *testing.T, engine *Engine, g *game.Game, saga *game.Permanent, agents [game.NumPlayers]PlayerAgent) *TurnLog {
	t.Helper()
	addCountersToPermanent(g, saga, counter.Lore, 1)
	if !engine.putTriggeredAbilitiesOnStack(g) {
		t.Fatal("chapter IV did not trigger")
	}
	log := &TurnLog{}
	engine.resolveTopOfStackWithChoices(g, agents, log)
	engine.applyStateBasedActions(g)
	return log
}

// TestTrialOfATimeLordOwnersOrderTheirBottomedCards: chapters I, II and III
// exile two of Player2's creatures and Player3's merged creature, whose two
// components are both exiled with the Saga (CR 607.2a). A guilty verdict puts
// all four cards on the bottom simultaneously; each owner, in APNAP order, is
// privately offered exactly their own arriving cards and arranges them
// (CR 401.4). Players with no arriving cards are not asked, and the ordering is
// not in the public log, so it does not reveal the hidden library order.
func TestTrialOfATimeLordOwnersOrderTheirBottomedCards(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	saga := addTrialOfATimeLord(t, g)
	p2Top := addCardToLibraryNamed(g, game.Player2, "P2 Library")
	p3Top := addCardToLibraryNamed(g, game.Player3, "P3 Library")

	first := addCombatCreaturePermanent(g, game.Player2)
	trialChapterExiles(t, engine, g, saga, first)
	second := addCombatCreaturePermanent(g, game.Player2)
	trialChapterExiles(t, engine, g, saga, second)
	merged := addCombatCreaturePermanent(g, game.Player3)
	lower := addCardToHand(g, game.Player3, shepherdVictimDef("Merged Lower"))
	g.Players[game.Player3].Hand.Remove(lower)
	merged.MergedCards = []game.MergedCard{{CardInstanceID: lower, Owner: game.Player3}}
	trialChapterExiles(t, engine, g, saga, merged)
	if !g.Players[game.Player3].Exile.Contains(lower) {
		t.Fatal("merged component was not exiled with the Saga")
	}

	var asked []game.PlayerID
	agents, owners := trialOwnerAgents(&asked)
	owners[game.Player2].order = []id.ID{second.CardInstanceID, first.CardInstanceID}
	owners[game.Player3].order = []id.ID{lower, merged.CardInstanceID}
	log := trialResolveVerdict(t, engine, g, saga, agents)

	if !slices.Equal(asked, []game.PlayerID{game.Player2, game.Player3}) {
		t.Fatalf("ordering asked %v, want Player2 then Player3", asked)
	}
	for player, want := range map[game.PlayerID][]id.ID{
		game.Player2: {first.CardInstanceID, second.CardInstanceID},
		game.Player3: {merged.CardInstanceID, lower},
	} {
		offered := owners[player].offered
		if len(offered) != 1 || len(offered[0]) != 2 || !slices.Contains(offered[0], want[0]) || !slices.Contains(offered[0], want[1]) {
			t.Fatalf("player %v offered %v, want exactly its arriving cards %v", player, offered, want)
		}
	}
	if got := g.Players[game.Player2].Library.All(); !slices.Equal(got, []id.ID{p2Top, second.CardInstanceID, first.CardInstanceID}) {
		t.Fatalf("Player2 library = %v, want chosen bottom order", got)
	}
	if got := g.Players[game.Player3].Library.All(); !slices.Equal(got, []id.ID{p3Top, lower, merged.CardInstanceID}) {
		t.Fatalf("Player3 library = %v, want chosen bottom order", got)
	}
	for _, choice := range log.Choices {
		if choice.Request.Kind == game.ChoiceOrder {
			t.Fatalf("library ordering was publicly logged: %+v", choice)
		}
	}
}

// TestTrialOfATimeLordOrderingFollowsReplacements: Player2's two cards are put
// on the bottom together, but one is Player2's commander and its owner puts it
// into the command zone instead (CR 903.9b). Only one card arrives in each of
// Player2's and Player3's libraries, so neither owner has an order to choose.
func TestTrialOfATimeLordOrderingFollowsReplacements(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	saga := addTrialOfATimeLord(t, g)
	commander := addCombatCreaturePermanent(g, game.Player2)
	trackCommanderID(g, game.Player2, commander.CardInstanceID)
	trialChapterExiles(t, engine, g, saga, commander)
	other := addCombatCreaturePermanent(g, game.Player2)
	trialChapterExiles(t, engine, g, saga, other)
	single := addCombatCreaturePermanent(g, game.Player3)
	trialChapterExiles(t, engine, g, saga, single)

	var asked []game.PlayerID
	agents, owners := trialOwnerAgents(&asked)
	owners[game.Player2].command = true
	trialResolveVerdict(t, engine, g, saga, agents)

	if len(asked) != 0 {
		t.Fatalf("ordering asked %v for a single arriving card", asked)
	}
	assertTrialCardZone(t, g, commander.CardInstanceID, zone.Command)
	if got := g.Players[game.Player2].Library.All(); !slices.Equal(got, []id.ID{other.CardInstanceID}) {
		t.Fatalf("Player2 library = %v, want only the non-commander card", got)
	}
	assertLibraryBottomSet(t, g, game.Player3, single.CardInstanceID)
}
