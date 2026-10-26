package rules

// Source review (#3348, shared Engine operation; no card admitted):
// Oracle (pinned corpus): Tibalt's Trickery 6623cb85-fb1e-4561-8b21-0977c25096c2
// "... Then they put the exiled cards on the bottom of their library in a random
// order."; Trial of a Time Lord a5425d4f-cd6c-4240-89a4-dd13e0e1a95f "... the
// owner of each card exiled with this Saga puts that card on the bottom of their
// library."; Bloodbraid Elf 3f0c9466-5ab9-4205-a84f-b4b27b5a678e cascade.
// CR 702.85a (cascade): "Then put all cards exiled this way that weren't cast on
// the bottom of your library in a random order." CR 603.2c: an ability triggers
// only once each time its trigger event occurs. CR 400.6: replacement effects
// that affect a move are determined, with the owner's/controller's choices,
// before the event moves the object. CR 903.9b: a commander that would be put
// into its owner's library may be put into the command zone instead. CR 101.4
// and 616.1: simultaneous choices are made in APNAP order.
// "Put all ... on the bottom" is therefore one simultaneous zone change out of
// exile: every replacement/commander choice is made against the pre-move state,
// then every card moves with one SimultaneousID.

import (
	"fmt"
	"math/rand/v2"
	"slices"
	"testing"

	"github.com/natefinch/council4/mtg/cards"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/id"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
)

const (
	disposalTestLink    = game.LinkedKey("disposal-test-exiled")
	commanderZonePrompt = "Put your commander into the command zone?"
)

// disposalBottomWatcher adds a Player3 permanent whose one-or-more trigger fires
// once per simultaneous batch of cards put into a library from exile.
func disposalBottomWatcher(g *game.Game) *game.Permanent {
	return addTriggeredPermanent(g, game.Player3, &game.TriggerPattern{
		Event:         game.EventZoneChanged,
		MatchFromZone: true,
		FromZone:      zone.Exile,
		MatchToZone:   true,
		ToZone:        zone.Library,
		OneOrMore:     true,
	}, []game.Instruction{{Primitive: game.GainLife{Amount: game.Fixed(1), Player: game.ControllerReference()}}}, nil)
}

// disposalWatcherTriggers puts pending triggers on the stack and counts the
// watcher's triggers there.
func disposalWatcherTriggers(engine *Engine, g *game.Game, watcher *game.Permanent) int {
	engine.putTriggeredAbilitiesOnStack(g)
	count := 0
	for _, obj := range g.Stack.Objects() {
		if obj.Kind == game.StackTriggeredAbility && obj.SourceCardID == watcher.CardInstanceID {
			count++
		}
	}
	return count
}

// assertOneExileBatch checks that every card put from exile into a library or,
// by the commander replacement, the command zone did so in one nonzero
// simultaneous batch of wantMoves cards.
func assertOneExileBatch(t *testing.T, g *game.Game, wantMoves int) {
	t.Helper()
	batches := map[id.ID]int{}
	for _, event := range g.Events {
		if event.Kind == game.EventZoneChanged && event.FromZone == zone.Exile &&
			(event.ToZone == zone.Library || event.ToZone == zone.Command) {
			batches[event.SimultaneousID]++
		}
	}
	if len(batches) != 1 || batches[0] != 0 {
		t.Fatalf("exile departures by batch = %v, want one nonzero batch of %d", batches, wantMoves)
	}
	for _, count := range batches {
		if count != wantMoves {
			t.Fatalf("exile departures by batch = %v, want one nonzero batch of %d", batches, wantMoves)
		}
	}
}

func assertLibraryBottomSet(t *testing.T, g *game.Game, playerID game.PlayerID, want ...id.ID) {
	t.Helper()
	if bottom := libraryBottomIDs(g, playerID, len(want)); !sameCardIDs(bottom, want) {
		t.Fatalf("player %d library bottom %v, want exactly %v", playerID, bottom, want)
	}
}

func disposalVanilla(name string, cardTypes ...types.Card) *game.CardDef {
	return &game.CardDef{CardFace: game.CardFace{Name: name, Types: cardTypes}}
}

// pushLinkedDisposal puts a triggered ability from a vanilla Player1 enchantment
// on the stack that resolves PutLinkedExiledCardsInLibrary, with linked
// published to its link in order.
func pushLinkedDisposal(g *game.Game, randomOrder bool, linked ...id.ID) {
	source := addCombatPermanent(g, game.Player1, disposalVanilla("Linked Source", types.Enchantment))
	trigger := game.TriggeredAbility{Content: game.Mode{Sequence: []game.Instruction{{
		Primitive: game.PutLinkedExiledCardsInLibrary{LinkedKey: disposalTestLink, Bottom: true, RandomOrder: randomOrder},
	}}}.Ability()}
	obj := &game.StackObject{
		ID: g.IDGen.Next(), Kind: game.StackTriggeredAbility, SourceID: source.ObjectID,
		SourceCardID: source.CardInstanceID, Controller: game.Player1, InlineTrigger: &trigger,
	}
	key := linkedObjectSourceKey(g, obj, string(disposalTestLink))
	for _, cardID := range linked {
		publishIterativeCard(g, key, cardID)
	}
	g.Stack.Push(obj)
}

// TestTibaltsTrickeryRemainderMovesAsOneBatch: the default-registry Tibalt's
// Trickery puts its remainder (the cards exiled this way, plus the found card
// when its cast is declined) on the bottom as one simultaneous event, so a
// one-or-more exile-to-library watcher triggers once.
func TestTibaltsTrickeryRemainderMovesAsOneBatch(t *testing.T) {
	def := cards.NewDefaultRegistry().Lookup("Tibalt's Trickery")
	if def == nil {
		t.Fatal("Tibalt's Trickery is not registered")
	}
	for _, castFound := range []bool{false, true} {
		t.Run(fmt.Sprintf("castFound=%v", castFound), func(t *testing.T) {
			g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
			engine := NewEngine(tibaltSeedForMillCount(t, 1))
			watcher := disposalBottomWatcher(g)
			victim, _ := pushTibaltVictim(g, game.Player2, "Tibalt Victim")
			addCardToLibraryNamed(g, game.Player2, "Rest")
			found := addCardToLibraryNamed(g, game.Player2, "Different Name")
			sameName := addCardToLibraryNamed(g, game.Player2, "Tibalt Victim")
			land := addLandToLibraryNamed(g, game.Player2, "Forest")
			addCardToLibraryNamed(g, game.Player2, "Milled Top")
			sourceID := g.IDGen.Next()
			g.CardInstances[sourceID] = &game.CardInstance{ID: sourceID, Def: def, Owner: game.Player1}
			g.Stack.Push(&game.StackObject{
				ID: g.IDGen.Next(), Kind: game.StackSpell, SourceID: sourceID, Controller: game.Player1,
				Targets: []game.Target{game.StackObjectTarget(victim.ID)}, TargetCounts: []int{1},
			})

			agents := [game.NumPlayers]PlayerAgent{game.Player2: &iterativeLibraryAgent{mayAnswers: []bool{castFound}}}
			engine.resolveTopOfStackWithChoices(g, agents, &TurnLog{})

			want := []id.ID{land, sameName}
			if castFound {
				if _, ok := tibaltCastSpell(g, found); !ok {
					t.Fatal("found card was not cast")
				}
			} else {
				want = append(want, found)
			}
			assertLibraryBottomSet(t, g, game.Player2, want...)
			assertOneExileBatch(t, g, len(want))
			if got := disposalWatcherTriggers(engine, g, watcher); got != 1 {
				t.Fatalf("one-or-more exile-to-library trigger fired %d times, want once", got)
			}
		})
	}
}

// TestPutLinkedExiledCardsInLibraryBatchesAcrossOwners: a linked disposal moves
// every owner's cards into their own library as one batch, in a random order
// per owner and in the ordered form (Trial of a Time Lord), where the cards are
// bottomed in linked order.
func TestPutLinkedExiledCardsInLibraryBatchesAcrossOwners(t *testing.T) {
	for _, randomOrder := range []bool{true, false} {
		t.Run(fmt.Sprintf("random=%v", randomOrder), func(t *testing.T) {
			g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
			engine := NewEngine(rand.New(rand.NewPCG(19, 19)))
			watcher := disposalBottomWatcher(g)
			addCardToLibraryNamed(g, game.Player1, "P1 Library")
			addCardToLibraryNamed(g, game.Player2, "P2 Library")
			mine := []id.ID{
				addExiledInstant(g, game.Player1, disposalVanilla("Mine 1", types.Creature)),
				addExiledInstant(g, game.Player1, disposalVanilla("Mine 2", types.Creature)),
			}
			theirs := []id.ID{
				addExiledInstant(g, game.Player2, disposalVanilla("Theirs 1", types.Creature)),
				addExiledInstant(g, game.Player2, disposalVanilla("Theirs 2", types.Creature)),
			}
			pushLinkedDisposal(g, randomOrder, mine[0], theirs[0], mine[1], theirs[1])

			engine.resolveTopOfStackWithChoices(g, [game.NumPlayers]PlayerAgent{}, &TurnLog{})

			assertLibraryBottomSet(t, g, game.Player1, mine...)
			assertLibraryBottomSet(t, g, game.Player2, theirs...)
			if !randomOrder {
				if got := libraryBottomIDs(g, game.Player1, 2); !slices.Equal(got, mine) {
					t.Fatalf("player 1 bottom %v, want linked order %v", got, mine)
				}
				if got := libraryBottomIDs(g, game.Player2, 2); !slices.Equal(got, theirs) {
					t.Fatalf("player 2 bottom %v, want linked order %v", got, theirs)
				}
			}
			if g.Players[game.Player1].Exile.Size() != 0 || g.Players[game.Player2].Exile.Size() != 0 {
				t.Fatal("linked cards remained in exile")
			}
			assertOneExileBatch(t, g, 4)
			if got := disposalWatcherTriggers(engine, g, watcher); got != 1 {
				t.Fatalf("one-or-more trigger fired %d times across owners, want once", got)
			}
		})
	}
}

// TestLinkedRandomDisposalIsReproducibleAndCloneSafe: the random bottom order is
// drawn from the engine RNG per owner, so equal seeds give equal orders on the
// original game and on a clone taken before resolution, and the order really
// varies with the seed.
func TestLinkedRandomDisposalIsReproducibleAndCloneSafe(t *testing.T) {
	setup := func() (*game.Game, []id.ID, []id.ID) {
		g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
		var mine, theirs []id.ID
		for i := range 5 {
			mine = append(mine, addExiledInstant(g, game.Player1, disposalVanilla(fmt.Sprintf("Mine %d", i), types.Creature)))
			theirs = append(theirs, addExiledInstant(g, game.Player2, disposalVanilla(fmt.Sprintf("Theirs %d", i), types.Creature)))
		}
		var linked []id.ID
		for i := range mine {
			linked = append(linked, mine[i], theirs[i])
		}
		pushLinkedDisposal(g, true, linked...)
		return g, mine, theirs
	}
	resolve := func(g *game.Game, seed uint64) [2][]id.ID {
		NewEngine(rand.New(rand.NewPCG(seed, seed))).resolveTopOfStackWithChoices(g, [game.NumPlayers]PlayerAgent{}, &TurnLog{})
		return [2][]id.ID{g.Players[game.Player1].Library.All(), g.Players[game.Player2].Library.All()}
	}
	varied := false
	for seed := uint64(1); seed <= 8; seed++ {
		g, mine, theirs := setup()
		clone := g.Clone()
		got := resolve(g, seed)
		if !sameCardIDs(got[0], mine) || !sameCardIDs(got[1], theirs) {
			t.Fatalf("seed %d: libraries %v, want each owner's own cards %v / %v", seed, got, mine, theirs)
		}
		if cloned := resolve(clone, seed); !slices.Equal(cloned[0], got[0]) || !slices.Equal(cloned[1], got[1]) {
			t.Fatalf("seed %d: clone order %v differs from original %v", seed, cloned, got)
		}
		if again, _, _ := setup(); !slices.Equal(resolve(again, seed)[0], got[0]) {
			t.Fatalf("seed %d: order is not reproducible", seed)
		}
		varied = varied || !slices.Equal(got[0], mine) || !slices.Equal(got[1], theirs)
	}
	if !varied {
		t.Fatal("random disposal never changed the linked order")
	}
}

// TestLinkedLibraryDisposalDecidesReplacementsInAPNAPBeforeMoving: each owner's
// commander choice (CR 903.9b) is made in APNAP order against the pre-move
// state, with every card still in exile, before any card moves (CR 101.4, 400.6,
// 616.1). The declined commander joins its owner's other card on the library
// bottom in the same batch.
func TestLinkedLibraryDisposalDecidesReplacementsInAPNAPBeforeMoving(t *testing.T) {
	for _, active := range []game.PlayerID{game.Player1, game.Player3} {
		for _, randomOrder := range []bool{true, false} {
			t.Run(fmt.Sprintf("active=%d/random=%v", active, randomOrder), func(t *testing.T) {
				g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
				g.Turn.ActivePlayer = active
				engine := NewEngine(rand.New(rand.NewPCG(22, 22)))
				watcher := disposalBottomWatcher(g)
				commanders := map[game.PlayerID]id.ID{}
				g.CommanderIDs = map[id.ID]bool{}
				var linked []id.ID
				for _, owner := range []game.PlayerID{game.Player4, game.Player2, game.Player1} {
					cardID := addExiledInstant(g, owner, disposalVanilla(fmt.Sprintf("Commander %d", owner), types.Creature))
					commanders[owner] = cardID
					g.CommanderIDs[cardID] = true
					linked = append(linked, cardID)
				}
				other := addExiledInstant(g, game.Player2, disposalVanilla("Other", types.Creature))
				linked = append(linked, other)
				pushLinkedDisposal(g, randomOrder, linked...)

				var order []game.PlayerID
				agent := &controlScript{choice: func(_ PlayerObservation, req game.ChoiceRequest) []int {
					if req.Prompt != commanderZonePrompt {
						return fallbackChoice(req)
					}
					order = append(order, req.Player)
					for owner, cardID := range commanders {
						if !g.Players[owner].Exile.Contains(cardID) {
							t.Errorf("player %d's commander left exile before player %d chose", owner, req.Player)
						}
					}
					if !g.Players[game.Player2].Exile.Contains(other) {
						t.Errorf("a batch member left exile before player %d chose", req.Player)
					}
					if req.Player == game.Player2 {
						return []int{0}
					}
					return []int{1}
				}}
				engine.resolveTopOfStackWithChoices(g, [game.NumPlayers]PlayerAgent{agent, agent, agent, agent}, &TurnLog{})

				want := playersInAPNAPOrder(g, []game.PlayerID{game.Player1, game.Player2, game.Player4})
				if !slices.Equal(order, want) {
					t.Fatalf("commander choices by %v, want APNAP %v", order, want)
				}
				for _, owner := range []game.PlayerID{game.Player1, game.Player4} {
					if !g.Players[owner].CommandZone.Contains(commanders[owner]) {
						t.Fatalf("player %d's commander was not put in the command zone", owner)
					}
				}
				assertLibraryBottomSet(t, g, game.Player2, commanders[game.Player2], other)
				assertOneExileBatch(t, g, len(linked))
				if got := disposalWatcherTriggers(engine, g, watcher); got != 1 {
					t.Fatalf("one-or-more exile-to-library trigger fired %d times, want once", got)
				}
			})
		}
	}
}

// TestBottomExiledCardsDecidesEveryReplacementBeforeMoving: the cascade and
// discover remainder is one simultaneous move. With two exiled partner
// commanders of one owner, both choices are made while every card is still in
// exile, then the cards move in one batch.
func TestBottomExiledCardsDecidesEveryReplacementBeforeMoving(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(rand.New(rand.NewPCG(23, 23)))
	keep := addExiledInstant(g, game.Player1, disposalVanilla("Kept Partner", types.Creature))
	command := addExiledInstant(g, game.Player1, disposalVanilla("Commanded Partner", types.Creature))
	rest := addExiledInstant(g, game.Player1, disposalVanilla("Revealed Instant", types.Instant))
	g.CommanderIDs = map[id.ID]bool{keep: true, command: true}
	exiled := []id.ID{keep, command, rest}
	choices := 0
	agent := &controlScript{choice: func(_ PlayerObservation, req game.ChoiceRequest) []int {
		if req.Prompt != commanderZonePrompt {
			return fallbackChoice(req)
		}
		choices++
		for _, cardID := range exiled {
			if !g.Players[game.Player1].Exile.Contains(cardID) {
				t.Errorf("card %d left exile before every commander choice was made", cardID)
			}
		}
		if req.Subject.Exists && req.Subject.Val.CardID == command {
			return []int{1}
		}
		return []int{0}
	}}
	engine.setReplacementChoiceContext(g, [game.NumPlayers]PlayerAgent{agent, agent, agent, agent}, &TurnLog{})
	var snapshots []game.ZoneCardSnapshot
	for _, cardID := range exiled {
		snapshots = append(snapshots, game.ZoneCardSnapshot{CardID: cardID})
	}
	if !bottomExiledCards(g, g.Players[game.Player1], snapshots, engine.rng) {
		t.Fatal("no exiled card moved")
	}
	if choices != 2 {
		t.Fatalf("commander choices = %d, want 2", choices)
	}
	if !g.Players[game.Player1].CommandZone.Contains(command) {
		t.Fatal("the chosen partner was not put in the command zone")
	}
	assertLibraryBottomSet(t, g, game.Player1, keep, rest)
	assertOneExileBatch(t, g, len(exiled))
}

// TestCascadeRemainderMovesAsOneBatch drives a real cascade cast that finds no
// eligible card and whose remainder holds two partner commanders: both commander
// choices see every revealed card still in exile, the accepted one goes to the
// command zone and the rest are bottomed in one batch that a one-or-more watcher
// sees once.
func TestCascadeRemainderMovesAsOneBatch(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(rand.New(rand.NewPCG(24, 24)))
	watcher := disposalBottomWatcher(g)
	cascadeID := addCardToHand(g, game.Player1, cascadeSpell(3))
	keep := addCardToLibrary(g, game.Player1, simpleGainLifeInstantWithManaValue("Kept Partner", 6))
	land := addCardToLibrary(g, game.Player1, disposalVanilla("Island", types.Land))
	command := addCardToLibrary(g, game.Player1, simpleGainLifeInstantWithManaValue("Commanded Partner", 5))
	expensive := addCardToLibrary(g, game.Player1, simpleGainLifeInstantWithManaValue("Too Expensive", 4))
	g.CommanderIDs = map[id.ID]bool{keep: true, command: true}
	g.Players[game.Player1].ManaPool.Add(mana.C, 3)
	g.Turn.PriorityPlayer = game.Player1
	revealed := []id.ID{expensive, command, land, keep}
	choices := 0
	agent := &controlScript{choice: func(_ PlayerObservation, req game.ChoiceRequest) []int {
		if req.Prompt != commanderZonePrompt {
			return fallbackChoice(req)
		}
		choices++
		for _, cardID := range revealed {
			if !g.Players[game.Player1].Exile.Contains(cardID) {
				t.Errorf("card %d left exile before every commander choice was made", cardID)
			}
		}
		if req.Subject.Exists && req.Subject.Val.CardID == command {
			return []int{1}
		}
		return []int{0}
	}}

	if !engine.applyActionWithChoices(g, game.Player1, action.CastSpell(cascadeID, nil, 0, nil),
		[game.NumPlayers]PlayerAgent{agent, agent, agent, agent}, &TurnLog{}) {
		t.Fatal("cascade spell cast failed")
	}
	agents := [game.NumPlayers]PlayerAgent{agent, agent, agent, agent}
	resolveCascadeTrigger(t, engine, g, &agents)

	if choices != 2 {
		t.Fatalf("commander choices = %d, want 2", choices)
	}
	if !g.Players[game.Player1].CommandZone.Contains(command) {
		t.Fatal("the chosen partner was not put in the command zone")
	}
	if g.Stack.Size() != 1 {
		t.Fatalf("stack size = %d, want only the cascade spell", g.Stack.Size())
	}
	if got := g.Players[game.Player1].Library.All(); !sameCardIDs(got, []id.ID{expensive, land, keep}) {
		t.Fatalf("library = %v, want the uncommanded remainder", got)
	}
	assertOneExileBatch(t, g, len(revealed))
	if got := disposalWatcherTriggers(engine, g, watcher); got != 1 {
		t.Fatalf("one-or-more exile-to-library trigger fired %d times, want once", got)
	}
}

// TestLinkedDisposalSkipsCardThatLeftExile: a linked card that already left exile
// (here moved exile-to-exile, a new incarnation) is skipped and not reclaimed;
// the rest of the group still moves together.
func TestLinkedDisposalSkipsCardThatLeftExile(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(rand.New(rand.NewPCG(25, 25)))
	var linked []id.ID
	for _, name := range []string{"Re-exiled", "Disposed"} {
		cardID := addCardToLibrary(g, game.Player1, disposalVanilla(name, types.Creature))
		if !moveCardBetweenZones(g, game.Player1, cardID, zone.Library, zone.Exile) {
			t.Fatal("setup exile failed")
		}
		linked = append(linked, cardID)
	}
	gone, stays := linked[0], linked[1]
	pushLinkedDisposal(g, false, gone, stays)
	if !moveCardBetweenZones(g, game.Player1, gone, zone.Exile, zone.Exile) {
		t.Fatal("exile-to-exile move failed")
	}
	engine.resolveTopOfStackWithChoices(g, [game.NumPlayers]PlayerAgent{}, &TurnLog{})
	if !g.Players[game.Player1].Exile.Contains(gone) || g.Players[game.Player1].Library.Contains(gone) {
		t.Fatal("a card that left exile was reclaimed by the stale link")
	}
	if got := g.Players[game.Player1].Library.All(); !slices.Equal(got, []id.ID{stays}) {
		t.Fatalf("library = %v, want only the still-linked card", got)
	}
}
