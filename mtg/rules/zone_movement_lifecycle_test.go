package rules

import (
	"strings"
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/zone"
)

func TestSharedZoneArrivalsDeferCommanderSBAAcrossClones(t *testing.T) {
	for _, from := range []zone.Type{zone.Hand, zone.Library, zone.Graveyard, zone.Exile} {
		for _, destination := range []zone.Type{zone.Graveyard, zone.Exile} {
			if from == destination && from != zone.Exile {
				continue
			}
			t.Run(from.String()+" to "+destination.String(), func(t *testing.T) {
				g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
				engine := NewEngine(nil)
				cardID := addCommanderCardToLibrary(g, game.Player1)
				player := g.Players[game.Player1]
				if from != zone.Library {
					player.Library.Remove(cardID)
					source, _ := destinationZone(g, game.Player1, from)
					source.Add(cardID)
				}
				owner := &zoneMovementChoices{command: true}
				engine.setReplacementChoiceContext(g, [game.NumPlayers]PlayerAgent{game.Player1: owner}, &TurnLog{})
				if !moveCardBetweenZonesInBatch(g, game.Player1, cardID, from, destination, false, 0) {
					t.Fatal("shared move failed")
				}
				arrival, _ := destinationZone(g, game.Player1, destination)
				if owner.asked != 0 || !arrival.Contains(cardID) || player.CommandZone.Contains(cardID) {
					t.Fatal("graveyard/exile arrival ran a premature commander SBA")
				}
				g.ClearChoiceContext()
				branch := g.Clone()
				branchOwner := &zoneMovementChoices{command: true}
				engine.setReplacementChoiceContext(branch, [game.NumPlayers]PlayerAgent{game.Player1: branchOwner}, &TurnLog{})
				engine.applyStateBasedActions(branch)
				engine.applyStateBasedActions(branch)
				if branchOwner.asked != 1 || !branch.Players[game.Player1].CommandZone.Contains(cardID) {
					t.Fatal("clone did not offer the arrival's SBA exactly once")
				}
				if !arrival.Contains(cardID) || owner.asked != 0 || g.CommanderEventCursor != 0 {
					t.Fatal("clone consumed the original's pending arrival")
				}
				engine.setReplacementChoiceContext(g, [game.NumPlayers]PlayerAgent{game.Player1: owner}, &TurnLog{})
				engine.applyStateBasedActions(g)
				engine.applyStateBasedActions(g)
				if owner.asked != 1 || !player.CommandZone.Contains(cardID) {
					t.Fatal("original did not independently offer its SBA exactly once")
				}
			})
		}
	}
}

func TestExileToExileInvalidatesOldPermissionsAndHandles(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	cardID := addCardToLibrary(g, game.Player1, vanillaCreature("New occurrence", 2, 2))
	if !moveCardBetweenZones(g, game.Player1, cardID, zone.Library, zone.Exile) {
		t.Fatal("setup move failed")
	}
	card, _ := g.GetCardInstance(cardID)
	original := game.ZoneCardSnapshot{CardID: cardID, ZoneVersion: card.ZoneVersion}
	g.PlottedCards = map[game.ObjectID]int{cardID: 1}
	g.ForetoldCards = map[game.ObjectID]int{cardID: 1}
	g.SuspendedCards = map[game.ObjectID]game.SuspendedCard{cardID: {Owner: game.Player1, Controller: game.Player1, TimeCounters: 1}}
	g.RuleEffects = []game.RuleEffect{{
		ID: g.IDGen.Next(), Kind: game.RuleEffectCastFromZone, Controller: game.Player1,
		CastFromZone: zone.Exile, AffectedCardID: cardID, Duration: game.DurationPermanent,
	}}
	g.AddExileCounterFromController(cardID, counter.Void, 1, game.Player2)
	if !moveCardBetweenZones(g, game.Player1, cardID, zone.Exile, zone.Exile) {
		t.Fatal("exile-to-exile move failed")
	}
	_, plotted := g.PlottedCards[cardID]
	_, foretold := g.ForetoldCards[cardID]
	_, suspended := g.SuspendedCards[cardID]
	_, counterSource := g.ExileCounterExiledBy[cardID]
	if plotted || foretold || suspended || counterSource || len(g.RuleEffects) != 0 {
		t.Fatal("old exile occurrence's permissions or markers survived")
	}
	if card.ZoneVersion != original.ZoneVersion+1 {
		t.Fatal("exile-to-exile move did not create a new incarnation")
	}
	if bottomExiledCards(g, g.Players[game.Player1], game.Player1, []game.ZoneCardSnapshot{original}, nil) {
		t.Fatal("an old exile handle followed the physical card into its new occurrence")
	}
	if !g.Players[game.Player1].Exile.Contains(cardID) {
		t.Fatal("stale handle removed the new exile occurrence")
	}
}

func TestDrawLogsDoNotReuseAnEarlierFailedDraw(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	log := TurnLog{}
	agents := [game.NumPlayers]PlayerAgent{}
	engine.drawCardWithReplacements(g, game.Player1, agents, &log, false)
	drawn := addCardToLibrary(g, game.Player1, vanillaCreature("Later draw", 2, 2))
	engine.drawCardWithReplacements(g, game.Player1, agents, &log, false)
	redirected := addCardToLibrary(g, game.Player1, vanillaCreature("Replaced draw", 2, 2))
	g.ReplacementEffects = []game.ReplacementEffect{{
		ID: g.IDGen.Next(), MatchEvent: game.EventZoneChanged,
		MatchFromZone: true, FromZone: zone.Library,
		MatchToZone: true, ToZone: zone.Hand, ReplaceToZone: zone.Exile,
	}}
	engine.drawCardWithReplacements(g, game.Player1, agents, &log, false)
	if len(log.Draws) != 3 || !log.Draws[0].Failed || log.Draws[1].Failed || log.Draws[2].Failed {
		t.Fatalf("per-draw failures = %+v, want only the empty-library attempt marked failed", log.Draws)
	}
	if !g.Players[game.Player1].Hand.Contains(drawn) || !g.Players[game.Player1].Exile.Contains(redirected) ||
		!g.FailedDraws[game.Player1] {
		t.Fatal("draw logging changed the actual moves or erased the pending SBA loss")
	}
}

func TestLibraryReorderKeepsCommanderIncarnation(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	untouched := addCardToLibrary(g, game.Player1, vanillaCreature("Untouched", 2, 2))
	commander := addCommanderCardToLibrary(g, game.Player1)
	card, _ := g.GetCardInstance(commander)
	before := card.ZoneVersion
	addEffectSpellToStack(g, game.Player1, game.Dig{
		Player: game.ControllerReference(), Look: game.Fixed(1), Take: game.Fixed(0),
		Slots: []game.DigSlot{{Count: game.Fixed(1), Destination: zone.Library, Bottom: true}},
	}, nil)
	owner := &zoneMovementChoices{command: true}
	engine.resolveTopOfStackWithChoices(g, [game.NumPlayers]PlayerAgent{game.Player1: owner}, &TurnLog{})
	player := g.Players[game.Player1]
	top, _ := player.Library.Top()
	bottom, _ := player.Library.Bottom()
	if top != untouched || bottom != commander || owner.asked != 0 || card.ZoneVersion != before {
		t.Fatal("reordering the commander was treated as a new zone arrival")
	}
	for _, event := range g.Events {
		if event.Kind == game.EventZoneChanged && event.CardID == commander {
			t.Fatal("library reorder emitted a zone-change event")
		}
	}
}

func TestHideawayDoesNotMoveOrExposeUnchosenCards(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	commander := addCommanderCardToLibrary(g, game.Player1)
	chosen := addCardToLibrary(g, game.Player1, vanillaCreature("Private hideaway choice", 2, 2))
	source := addHideawayLandPermanent(g, game.Player1)
	owner := &zoneMovementChoices{command: true}
	log := TurnLog{}
	agents := [game.NumPlayers]PlayerAgent{game.Player1: owner}
	engine.setReplacementChoiceContext(g, agents, &log)
	engine.resolveInstructionWithChoices(g, hideawaySourceObject(source),
		&game.Instruction{Primitive: game.HideawayExile{Amount: game.Fixed(2)}}, agents, &log)
	engine.applyStateBasedActions(g)
	player := g.Players[game.Player1]
	if !player.Library.Contains(commander) || owner.asked != 0 || !player.Exile.IsFaceDown(chosen) {
		t.Fatal("hideaway moved the unchosen commander or failed to hide the chosen card")
	}
	for _, event := range g.Events {
		if event.CardID == commander && event.Kind == game.EventZoneChanged {
			t.Fatal("hideaway manufactured a zone change for an unchosen card")
		}
	}
	if reachableNames(NewObservation(g, game.Player2))["Private hideaway choice"] {
		t.Fatal("hideaway exposed its private choice to an opponent")
	}
	for _, choice := range log.Choices {
		for _, option := range choice.Request.Options {
			if option.Card.Exists || option.Label == "Private hideaway choice" {
				t.Fatal("hideaway's private choice leaked into the log")
			}
		}
	}
}

func TestSpecialActionExilesUseSharedMovement(t *testing.T) {
	operations := []struct {
		name     string
		def      *game.CardDef
		action   func(game.ObjectID) action.Action
		faceDown bool
	}{
		{"plot", plotSorcery(cost.Mana{cost.O(1)}), action.PlotCard, false},
		{"foretell", foretellInstant(cost.Mana{cost.U}), action.ForetellCard, true},
		{"suspend", suspendSorcery(2, cost.Mana{cost.O(1)}), action.SuspendCard, false},
	}
	for _, operation := range operations {
		t.Run(operation.name, func(t *testing.T) {
			for _, redirect := range []bool{false, true} {
				name := "exiled"
				if redirect {
					name = "redirected"
				}
				t.Run(name, func(t *testing.T) {
					g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
					engine := NewEngine(nil)
					cardID := addCardToHand(g, game.Player1, operation.def)
					setSorceryTiming(g, game.Player1)
					g.Players[game.Player1].ManaPool.Add(mana.C, 2)
					if redirect {
						g.ReplacementEffects = []game.ReplacementEffect{{
							ID: g.IDGen.Next(), MatchEvent: game.EventZoneChanged,
							MatchFromZone: true, FromZone: zone.Hand,
							MatchToZone: true, ToZone: zone.Exile, ReplaceToZone: zone.Hand,
						}}
					}
					if !engine.applyAction(g, game.Player1, operation.action(cardID)) {
						t.Fatal("special action failed")
					}
					player := g.Players[game.Player1]
					if redirect {
						_, plotted := g.PlottedCards[cardID]
						_, foretold := g.ForetoldCards[cardID]
						_, suspended := g.SuspendedCards[cardID]
						if !player.Hand.Contains(cardID) || player.Exile.Contains(cardID) || plotted || foretold || suspended {
							t.Fatal("special action followed the requested rather than actual destination")
						}
					} else if !player.Exile.Contains(cardID) || player.Exile.IsFaceDown(cardID) != operation.faceDown {
						t.Fatal("special action lost its actual exile placement or visibility")
					}
				})
			}
		})
	}
}

func TestIterativeZoneReplacementWithoutProgressIsExplicitlyUnsupported(t *testing.T) {
	for _, operation := range []struct {
		name      string
		primitive game.Primitive
		target    zone.Type
	}{
		{"reveal", game.RevealUntil{
			Player: game.ControllerReference(), Until: game.Selection{Name: "Absent"}, Destination: zone.Hand,
		}, zone.Hand},
		{"discover", game.DiscoverCards{Amount: game.Fixed(100)}, zone.Exile},
	} {
		t.Run(operation.name, func(t *testing.T) {
			g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
			addCardToLibrary(g, game.Player1, vanillaCreature("Stays put", 2, 2))
			g.ReplacementEffects = []game.ReplacementEffect{{
				ID: g.IDGen.Next(), MatchEvent: game.EventZoneChanged,
				MatchFromZone: true, FromZone: zone.Library,
				MatchToZone: true, ToZone: operation.target, ReplaceToZone: zone.Library,
			}}
			addEffectSpellToStack(g, game.Player1, operation.primitive, nil)
			defer func() {
				failure, ok := recover().(UnsupportedError)
				if !ok || !strings.Contains(failure.Reason, "advancing") {
					t.Fatalf("failure = %+v, want an explicit unsupported continuation, not a hang or success", failure)
				}
			}()
			NewEngine(nil).resolveTopOfStack(g, &TurnLog{})
		})
	}
}
