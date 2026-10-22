package rules

import (
	"fmt"
	"reflect"
	"slices"
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/id"
	"github.com/natefinch/council4/mtg/game/zone"
)

func TestSimultaneousCommanderReplacementChoicesAPNAP(t *testing.T) {
	for _, seating := range []struct {
		name   string
		active game.PlayerID
		order  [game.NumPlayers]game.PlayerID
		want   []game.PlayerID
	}{
		{"default", game.Player1, [4]game.PlayerID{0, 1, 2, 3}, []game.PlayerID{1, 3}},
		{"active-fourth", game.Player4, [4]game.PlayerID{0, 1, 2, 3}, []game.PlayerID{3, 1}},
		{"reverse-seating", game.Player1, [4]game.PlayerID{0, 3, 2, 1}, []game.PlayerID{3, 1}},
		{"active-second-custom-seating", game.Player2, [4]game.PlayerID{0, 3, 1, 2}, []game.PlayerID{1, 3}},
	} {
		for _, destination := range []zone.Type{zone.Hand, zone.Library} {
			for mask := range 4 {
				for _, reverse := range []bool{false, true} {
					for _, changedControl := range []bool{false, true} {
						t.Run(fmt.Sprintf("%s/destination-%d/accept-%d/reverse-%t/control-%t", seating.name, destination, mask, reverse, changedControl), func(t *testing.T) {
							g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
							g.Turn.ActivePlayer = seating.active
							g.TurnOrder.Order = seating.order
							second := addCommanderPermanent(g, game.Player2)
							fourth := addCommanderPermanent(g, game.Player4)
							if changedControl {
								second.Controller, fourth.Controller = game.Player4, game.Player2
							}
							second.Counters.Add(counter.PlusOnePlusOne, 2)
							permanents := []*game.Permanent{second, fourth}
							if reverse {
								slices.Reverse(permanents)
							}
							before := make(map[id.ID]game.ObjectSnapshot)
							for _, permanent := range permanents {
								before[permanent.ObjectID] = snapshotPermanent(g, permanent, zone.Battlefield)
							}
							eventCount := len(g.Events)
							var choices []game.PlayerID
							agent := &controlScript{choice: func(obs PlayerObservation, request game.ChoiceRequest) []int {
								if request.Kind != game.ChoiceMay || obs.Player != request.Player || request.Private {
									t.Fatalf("unexpected choice: %+v", request)
								}
								choices = append(choices, request.Player)
								if len(g.Battlefield) != 2 || len(g.Events) != eventCount || len(g.LastKnownInformation) != 0 {
									t.Fatal("commander moved before all choices completed")
								}
								for _, permanent := range permanents {
									if !reflect.DeepEqual(snapshotPermanent(g, permanent, zone.Battlefield), before[permanent.ObjectID]) {
										t.Fatal("a choice saw a different pre-move battlefield")
									}
								}
								bit := 0
								if request.Player == game.Player4 {
									bit = 1
								}
								return []int{(mask >> bit) & 1}
							}}
							engine := NewEngine(nil)
							engine.setReplacementChoiceContext(g, [game.NumPlayers]PlayerAgent{agent, agent, agent, agent}, &TurnLog{})
							batch := g.IDGen.Next()
							results := movePermanentsToZoneSimultaneouslyWithResultsInBatch(g, permanents, destination, batch)
							if !slices.Equal(choices, seating.want) {
								t.Fatalf("commander choices = %v, want %v", choices, seating.want)
							}
							if len(results) != len(permanents) || len(g.Battlefield) != 0 {
								t.Fatal("batch did not move both permanents")
							}
							var events []game.Event
							for _, event := range g.Events[eventCount:] {
								if event.Kind == game.EventZoneChanged {
									events = append(events, event)
								}
							}
							if len(events) != len(results) {
								t.Fatalf("zone events = %d, want %d", len(events), len(results))
							}
							for i, result := range results {
								permanent := permanents[i]
								want := destination
								bit := 0
								if permanent.Owner == game.Player4 {
									bit = 1
								}
								if mask&(1<<bit) != 0 {
									want = zone.Command
								}
								cards, _ := destinationZone(g, permanent.Owner, want)
								card, _ := g.GetCardInstance(permanent.CardInstanceID)
								if result.permanent != permanent || !result.moved || result.destination != want ||
									!cards.Contains(permanent.CardInstanceID) ||
									!slices.Equal(result.arrivals, []game.ZoneCardSnapshot{{CardID: card.ID, ZoneVersion: card.ZoneVersion}}) {
									t.Fatalf("uncorrelated result %d: %+v", i, result)
								}
								event := events[i]
								if event.PermanentID != permanent.ObjectID || event.CardID != card.ID || event.Player != permanent.Owner ||
									event.Controller != permanent.Controller || event.ToZone != want || event.SimultaneousID != batch {
									t.Fatalf("uncorrelated simultaneous event %d: %+v", i, event)
								}
								lki := g.LastKnownInformation[permanent.ObjectID]
								if lki.Controller != permanent.Controller || lki.Owner != permanent.Owner ||
									!reflect.DeepEqual(lki.Counters, before[permanent.ObjectID].Counters) {
									t.Fatal("lost pre-move controller, owner or counters")
								}
							}
						})
					}
				}
			}
		}
	}
}

func TestSimultaneousCommanderReplacementReplayAndClone(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	fourth := addCommanderPermanent(g, game.Player4)
	second := addCommanderPermanent(g, game.Player2)
	targets := []id.ID{fourth.ObjectID, second.ObjectID}
	replay := g.Clone()
	var duringChoice *game.Game
	run := func(state *game.Game, accept bool, recorded *TurnLog) *TurnLog {
		t.Helper()
		var permanents []*game.Permanent
		for _, objectID := range targets {
			permanent, ok := permanentByObjectID(state, objectID)
			if !ok {
				t.Fatal("clone lost a pre-move permanent")
			}
			permanents = append(permanents, permanent)
		}
		agent := &controlScript{choice: func(_ PlayerObservation, request game.ChoiceRequest) []int {
			if duringChoice == nil {
				duringChoice = state.Clone()
			}
			if accept {
				return []int{1}
			}
			return []int{0}
		}}
		engine := NewEngine(nil)
		log := &TurnLog{}
		agents := [game.NumPlayers]PlayerAgent{agent, agent, agent, agent}
		if recorded != nil {
			var decisions [game.NumPlayers][][]int
			for _, decision := range recorded.Choices {
				decisions[decision.DecisionMaker] = append(decisions[decision.DecisionMaker], slices.Clone(decision.Selected))
			}
			for player := range game.NumPlayers {
				agents[player] = &choiceOnlyAgent{choices: decisions[player]}
			}
		}
		engine.setReplacementChoiceContext(state, agents, log)
		if !movePermanentsToZoneSimultaneously(state, permanents, zone.Library) {
			t.Fatal("batch move failed")
		}
		return log
	}
	firstLog := run(g, true, nil)
	replayLog := run(replay, true, firstLog)
	if !reflect.DeepEqual(firstLog, replayLog) || !reflect.DeepEqual(g.Events, replay.Events) ||
		!reflect.DeepEqual(g.LastKnownInformation, replay.LastKnownInformation) {
		t.Fatal("identical decisions did not replay deterministically")
	}
	if duringChoice == nil || len(duringChoice.Battlefield) != 2 {
		t.Fatal("clone taken during choices did not retain the pre-move battlefield")
	}
	branchLog := run(duringChoice, false, nil)
	if len(branchLog.Choices) != 2 || len(firstLog.Choices) != 2 {
		t.Fatal("one game's staged decisions escaped into another")
	}
	for _, permanent := range []*game.Permanent{second, fourth} {
		if !g.Players[permanent.Owner].CommandZone.Contains(permanent.CardInstanceID) ||
			!replay.Players[permanent.Owner].CommandZone.Contains(permanent.CardInstanceID) ||
			!duringChoice.Players[permanent.Owner].Library.Contains(permanent.CardInstanceID) ||
			duringChoice.Players[permanent.Owner].CommandZone.Contains(permanent.CardInstanceID) {
			t.Fatal("a clone reused or mutated another game's destination decisions")
		}
	}
}

func TestSimultaneousCommanderReplacementComposesWithZoneReplacement(t *testing.T) {
	for _, accept := range []bool{false, true} {
		t.Run(fmt.Sprintf("accept-%t", accept), func(t *testing.T) {
			g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
			fourth := addCommanderPermanent(g, game.Player4)
			second := addCommanderPermanent(g, game.Player2)
			fourth.Controller, second.Controller = game.Player1, game.Player3
			// The first departing permanent supplies both ordinary replacements.
			for _, destination := range []zone.Type{zone.Library, zone.Hand} {
				g.ReplacementEffects = append(g.ReplacementEffects, game.ReplacementEffect{
					ID: g.IDGen.Next(), SourceObjectID: fourth.ObjectID,
					MatchEvent: game.EventZoneChanged, MatchFromZone: true, FromZone: zone.Battlefield,
					MatchToZone: true, ToZone: zone.Graveyard, ReplaceToZone: destination,
					Description: fmt.Sprintf("redirect to %d", destination),
				})
			}
			var owners, controllers []game.PlayerID
			agent := &controlScript{choice: func(_ PlayerObservation, request game.ChoiceRequest) []int {
				if len(g.Battlefield) != 2 {
					t.Fatal("replacement source left before all decisions")
				}
				if request.Kind == game.ChoiceReplacement {
					controllers = append(controllers, request.Player)
					return []int{0}
				}
				if request.Kind != game.ChoiceMay {
					t.Fatalf("unexpected choice: %+v", request)
				}
				owners = append(owners, request.Player)
				if accept {
					return []int{1}
				}
				return []int{0}
			}}
			engine := NewEngine(nil)
			engine.setReplacementChoiceContext(g, [game.NumPlayers]PlayerAgent{agent, agent, agent, agent}, &TurnLog{})
			results := movePermanentsToZoneSimultaneouslyWithResults(g, []*game.Permanent{fourth, second}, zone.Graveyard)
			if !slices.Equal(controllers, []game.PlayerID{game.Player1, game.Player3}) ||
				!slices.Equal(owners, []game.PlayerID{game.Player2, game.Player4}) {
				t.Fatalf("ordinary chooser = %v, commander owners = %v", controllers, owners)
			}
			if len(results) != 2 || len(g.ReplacementDecisions) != 2 {
				t.Fatal("replacement composition lost a move or an ordinary selection")
			}
			want := zone.Library
			if accept {
				want = zone.Command
			}
			for _, result := range results {
				cards, _ := destinationZone(g, result.permanent.Owner, want)
				if !result.moved || result.destination != want || !cards.Contains(result.permanent.CardInstanceID) {
					t.Fatal("commander choice did not compose with the ordinary library redirect")
				}
			}
			for _, decision := range g.ReplacementDecisions {
				if decision.UsedFallback || !slices.Equal(decision.Selected, []int{0}) {
					t.Fatal("ordinary controller choice was not honored")
				}
			}
		})
	}
}

func TestSimultaneousMixedZoneReplacementChoicesAPNAP(t *testing.T) {
	for _, seating := range []struct {
		name        string
		active      game.PlayerID
		order       [game.NumPlayers]game.PlayerID
		normalWant  []game.PlayerID
		changedWant []game.PlayerID
	}{
		{"default", game.Player1, [4]game.PlayerID{0, 1, 2, 3}, []game.PlayerID{1, 3}, []game.PlayerID{1, 2}},
		{"active-fourth", game.Player4, [4]game.PlayerID{0, 1, 2, 3}, []game.PlayerID{3, 1}, []game.PlayerID{1, 2}},
		{"reverse-seating", game.Player1, [4]game.PlayerID{0, 3, 2, 1}, []game.PlayerID{3, 1}, []game.PlayerID{2, 1}},
		{"active-second-custom-seating", game.Player2, [4]game.PlayerID{0, 3, 1, 2}, []game.PlayerID{1, 3}, []game.PlayerID{1, 2}},
	} {
		for _, destination := range []zone.Type{zone.Hand, zone.Library} {
			for _, reverse := range []bool{false, true} {
				for _, changedControl := range []bool{false, true} {
					for _, accept := range []bool{false, true} {
						t.Run(fmt.Sprintf("%s/destination-%d/reverse-%t/control-%t/accept-%t", seating.name, destination, reverse, changedControl, accept), func(t *testing.T) {
							g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
							g.Turn.ActivePlayer, g.TurnOrder.Order = seating.active, seating.order
							commander := addCommanderPermanent(g, game.Player2)
							ordinary := addCombatPermanent(g, game.Player4, vanillaCreature("Ordinary return", 2, 2))
							want := seating.normalWant
							if changedControl {
								commander.Controller, ordinary.Controller = game.Player4, game.Player3
								want = seating.changedWant
							}
							redirect := zone.Library
							if destination == zone.Library {
								redirect = zone.Hand
							}
							for _, to := range []zone.Type{redirect, zone.Exile} {
								g.ReplacementEffects = append(g.ReplacementEffects, game.ReplacementEffect{
									ID: g.IDGen.Next(), SourceObjectID: commander.ObjectID, AffectedObjectID: ordinary.ObjectID,
									MatchEvent: game.EventZoneChanged, MatchFromZone: true, FromZone: zone.Battlefield,
									MatchToZone: true, ToZone: destination, ReplaceToZone: to,
									Description: fmt.Sprintf("redirect ordinary creature to %d", to),
								})
							}
							permanents := []*game.Permanent{commander, ordinary}
							if reverse {
								slices.Reverse(permanents)
							}
							before := []game.ObjectSnapshot{
								snapshotPermanent(g, permanents[0], zone.Battlefield),
								snapshotPermanent(g, permanents[1], zone.Battlefield),
							}
							eventCount := len(g.Events)
							var players []game.PlayerID
							var kinds []game.ChoiceKind
							agent := &controlScript{choice: func(obs PlayerObservation, request game.ChoiceRequest) []int {
								if obs.Player != request.Player || request.Private ||
									len(g.Battlefield) != 2 || len(g.Events) != eventCount || len(g.LastKnownInformation) != 0 {
									t.Fatal("choices did not precede all movement")
								}
								for i, permanent := range permanents {
									if !reflect.DeepEqual(snapshotPermanent(g, permanent, zone.Battlefield), before[i]) {
										t.Fatal("replacement choice saw a different pre-move battlefield")
									}
								}
								players = append(players, request.Player)
								kinds = append(kinds, request.Kind)
								if request.Kind == game.ChoiceMay && accept {
									return []int{1}
								}
								return []int{0}
							}}
							engine := NewEngine(nil)
							log := &TurnLog{}
							engine.setReplacementChoiceContext(g, [game.NumPlayers]PlayerAgent{agent, agent, agent, agent}, log)
							batch := g.IDGen.Next()
							results := movePermanentsToZoneSimultaneouslyWithResultsInBatch(g, permanents, destination, batch)
							wantKinds := make([]game.ChoiceKind, len(want))
							for i, player := range want {
								wantKinds[i] = game.ChoiceReplacement
								if player == commander.Owner {
									wantKinds[i] = game.ChoiceMay
								}
							}
							if !slices.Equal(players, want) || !slices.Equal(kinds, wantKinds) {
								t.Fatalf("mixed choices players=%v kinds=%v, want players=%v kinds=%v", players, kinds, want, wantKinds)
							}
							if len(results) != 2 || len(g.Battlefield) != 0 || len(log.Choices) != 2 ||
								len(g.ReplacementDecisions) != 1 || g.ReplacementDecisions[0].UsedFallback ||
								!slices.Equal(g.ReplacementDecisions[0].Selected, []int{0}) {
								t.Fatal("lost a move or ignored the ordinary replacement choice")
							}
							var events []game.Event
							for _, event := range g.Events[eventCount:] {
								if event.Kind == game.EventZoneChanged {
									events = append(events, event)
								}
							}
							if len(events) != 2 {
								t.Fatalf("departures = %d, want 2", len(events))
							}
							for i, permanent := range permanents {
								to := redirect
								if permanent == commander {
									to = destination
									if accept {
										to = zone.Command
									}
								}
								cards, _ := destinationZone(g, permanent.Owner, to)
								result, event := results[i], events[i]
								card, _ := g.GetCardInstance(permanent.CardInstanceID)
								if !result.moved || result.permanent != permanent || result.destination != to ||
									!cards.Contains(permanent.CardInstanceID) ||
									!slices.Equal(result.arrivals, []game.ZoneCardSnapshot{{CardID: card.ID, ZoneVersion: card.ZoneVersion}}) ||
									event.PermanentID != permanent.ObjectID || event.Player != permanent.Owner ||
									event.Controller != permanent.Controller || event.ToZone != to || event.SimultaneousID != batch {
									t.Fatalf("mixed result/event %d lost its input correspondence or owner-relative destination", i)
								}
							}
						})
					}
				}
			}
		}
	}
}

func TestSimultaneousMixedZoneReplacementRestartsAPNAP(t *testing.T) {
	for _, reverse := range []bool{false, true} {
		t.Run(fmt.Sprintf("reverse-%t", reverse), func(t *testing.T) {
			g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
			commander := addCommanderPermanent(g, game.Player2)
			commander.Controller = game.Player3
			ordinary := addCombatPermanent(g, game.Player4, vanillaCreature("Ordinary return", 2, 2))
			for _, to := range []zone.Type{zone.Library, zone.Exile} {
				g.ReplacementEffects = append(g.ReplacementEffects, game.ReplacementEffect{
					ID: g.IDGen.Next(), MatchEvent: game.EventZoneChanged,
					MatchFromZone: true, FromZone: zone.Battlefield, MatchToZone: true, ToZone: zone.Hand,
					ReplaceToZone: to, Description: fmt.Sprintf("redirect to %d", to),
				})
			}
			var players []game.PlayerID
			var kinds []game.ChoiceKind
			agent := &controlScript{choice: func(_ PlayerObservation, request game.ChoiceRequest) []int {
				if len(g.Battlefield) != 2 {
					t.Fatal("movement preceded the pending replacement decisions")
				}
				players = append(players, request.Player)
				kinds = append(kinds, request.Kind)
				if request.Kind == game.ChoiceMay {
					return []int{1}
				}
				return []int{0}
			}}
			engine := NewEngine(nil)
			engine.setReplacementChoiceContext(g, [game.NumPlayers]PlayerAgent{agent, agent, agent, agent}, &TurnLog{})
			permanents := []*game.Permanent{commander, ordinary}
			if reverse {
				slices.Reverse(permanents)
			}
			results := movePermanentsToZoneSimultaneouslyWithResults(g, permanents, zone.Hand)
			if !slices.Equal(players, []game.PlayerID{game.Player3, game.Player2, game.Player4}) ||
				!slices.Equal(kinds, []game.ChoiceKind{game.ChoiceReplacement, game.ChoiceMay, game.ChoiceReplacement}) {
				t.Fatalf("newly exposed owner choice did not restart APNAP (101.4d): players=%v kinds=%v", players, kinds)
			}
			if len(results) != 2 || !g.Players[game.Player2].CommandZone.Contains(commander.CardInstanceID) ||
				!g.Players[game.Player4].Library.Contains(ordinary.CardInstanceID) {
				t.Fatal("APNAP restart lost an ordinary or commander destination")
			}
		})
	}
}

func TestSimultaneousMixedZoneReplacementReplayAndClone(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	commander := addCommanderPermanent(g, game.Player2)
	ordinary := addCombatPermanent(g, game.Player4, vanillaCreature("Ordinary return", 2, 2))
	commander.Controller, ordinary.Controller = game.Player4, game.Player3
	for _, to := range []zone.Type{zone.Library, zone.Exile} {
		g.ReplacementEffects = append(g.ReplacementEffects, game.ReplacementEffect{
			ID: g.IDGen.Next(), AffectedObjectID: ordinary.ObjectID, MatchEvent: game.EventZoneChanged,
			MatchFromZone: true, FromZone: zone.Battlefield, MatchToZone: true, ToZone: zone.Hand,
			ReplaceToZone: to, Description: fmt.Sprintf("redirect to %d", to),
		})
	}
	replay := g.Clone()
	var branch *game.Game
	run := func(state *game.Game, commanderSelection, ordinarySelection int) *TurnLog {
		t.Helper()
		var players []game.PlayerID
		agent := &controlScript{choice: func(_ PlayerObservation, request game.ChoiceRequest) []int {
			players = append(players, request.Player)
			if request.Kind == game.ChoiceMay {
				return []int{commanderSelection}
			}
			if branch == nil {
				branch = state.Clone()
			}
			return []int{ordinarySelection}
		}}
		engine := NewEngine(nil)
		log := &TurnLog{}
		engine.setReplacementChoiceContext(state, [game.NumPlayers]PlayerAgent{agent, agent, agent, agent}, log)
		var permanents []*game.Permanent
		for _, objectID := range []id.ID{ordinary.ObjectID, commander.ObjectID} {
			permanent, ok := permanentByObjectID(state, objectID)
			if !ok {
				t.Fatal("clone lost a pending permanent")
			}
			permanents = append(permanents, permanent)
		}
		if !movePermanentsToZoneSimultaneously(state, permanents, zone.Hand) ||
			!slices.Equal(players, []game.PlayerID{game.Player2, game.Player3}) || len(log.Choices) != 2 {
			t.Fatal("mixed decisions did not run independently in APNAP order")
		}
		return log
	}
	log, replayLog := run(g, 1, 0), run(replay, 1, 0)
	if !reflect.DeepEqual(log, replayLog) || !reflect.DeepEqual(g.Events, replay.Events) ||
		!reflect.DeepEqual(g.ReplacementDecisions, replay.ReplacementDecisions) ||
		!reflect.DeepEqual(g.LastKnownInformation, replay.LastKnownInformation) {
		t.Fatal("mixed replacement decisions did not replay deterministically")
	}
	if branch == nil || len(branch.Battlefield) != 2 {
		t.Fatal("clone taken during an ordinary decision lost the pre-move battlefield")
	}
	run(branch, 0, 1)
	for _, state := range []*game.Game{g, replay} {
		if !state.Players[game.Player2].CommandZone.Contains(commander.CardInstanceID) ||
			!state.Players[game.Player4].Library.Contains(ordinary.CardInstanceID) {
			t.Fatal("opposite-decision clone changed another game's destinations")
		}
	}
	if !branch.Players[game.Player2].Hand.Contains(commander.CardInstanceID) ||
		!branch.Players[game.Player4].Exile.Contains(ordinary.CardInstanceID) {
		t.Fatal("clone reused pending commander or ordinary replacement state")
	}
}

func TestSimultaneousCommanderReplacementMergedComponent(t *testing.T) {
	for _, destination := range []zone.Type{zone.Hand, zone.Library} {
		for _, accept := range []bool{false, true} {
			t.Run(fmt.Sprintf("destination-%d/accept-%t", destination, accept), func(t *testing.T) {
				g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
				top := addCombatPermanent(g, game.Player4, vanillaCreature("Merged top", 3, 3))
				lower := addCommanderCardToHand(g, game.Player4)
				g.Players[game.Player4].Hand.Remove(lower)
				top.MergedCards = []game.MergedCard{{CardInstanceID: lower, Owner: game.Player4}}
				second := addCommanderPermanent(g, game.Player2)
				var choices []game.PlayerID
				agent := &controlScript{choice: func(_ PlayerObservation, request game.ChoiceRequest) []int {
					choices = append(choices, request.Player)
					if len(g.Battlefield) != 2 {
						t.Fatal("merged permanent moved before all owners chose")
					}
					if accept {
						return []int{1}
					}
					return []int{0}
				}}
				engine := NewEngine(nil)
				engine.setReplacementChoiceContext(g, [game.NumPlayers]PlayerAgent{agent, agent, agent, agent}, &TurnLog{})
				results := movePermanentsToZoneSimultaneouslyWithResults(g, []*game.Permanent{top, second}, destination)
				if !slices.Equal(choices, []game.PlayerID{game.Player2, game.Player4}) || len(results) != 2 ||
					results[0].permanent != top || !results[0].moved || results[0].destination != destination ||
					len(results[0].arrivals) != 2 {
					t.Fatal("merged commander choices or physical arrival correlation changed")
				}
				cards, _ := destinationZone(g, game.Player4, destination)
				commanderDestination := destination
				if accept {
					commanderDestination = zone.Command
				}
				commanderCards, _ := destinationZone(g, game.Player4, commanderDestination)
				if !cards.Contains(top.CardInstanceID) || !commanderCards.Contains(lower) {
					t.Fatal("merged commander redirected its noncommander component")
				}
				var batch id.ID
				count := 0
				for _, event := range g.Events {
					if event.Kind == game.EventZoneChanged && event.FromZone == zone.Battlefield {
						if batch == 0 {
							batch = event.SimultaneousID
						}
						if batch == 0 || event.SimultaneousID != batch {
							t.Fatal("merged components lost their shared simultaneous identity")
						}
						count++
					}
				}
				if count != 3 {
					t.Fatalf("physical departures = %d, want 3", count)
				}
			})
		}
	}
}
