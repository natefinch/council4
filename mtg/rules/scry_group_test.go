package rules

import (
	"fmt"
	"reflect"
	"slices"
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/id"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

func TestGroupScryStagesChoicesBeforeLibraryChanges(t *testing.T) {
	for _, optional := range []bool{false, true} {
		name := "mandatory"
		if optional {
			name = "optional"
		}
		t.Run(name, func(t *testing.T) {
			g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
			g.Turn.ActivePlayer = game.Player3
			var libraries [game.NumPlayers][]id.ID
			for player := range game.NumPlayers {
				bottom := addCardToLibrary(g, game.PlayerID(player), vanillaCreature("Scry secret bottom", 2, 2))
				top := addCardToLibrary(g, game.PlayerID(player), vanillaCreature("Scry secret top", 2, 2))
				libraries[player] = []id.ID{top, bottom}
			}
			instruction := game.Instruction{
				Primitive: game.Scry{Player: game.GroupOfferMemberReference(), Amount: game.Fixed(1)},
			}
			if optional {
				instruction.Optional = true
				instruction.OptionalActorGroup = opt.Val(game.OpponentsReference())
			} else {
				instruction.ForEachPlayerGroup = opt.Val(game.OpponentsReference())
			}
			def := &game.CardDef{CardFace: game.CardFace{
				Name: "Group scry", Types: []types.Card{types.Sorcery},
				SpellAbility: opt.Val(game.Mode{Sequence: []game.Instruction{instruction}}.Ability()),
			}}
			addImplementationSpellToStack(g, game.Player2, def, nil)
			var mayPlayers, scryPlayers []game.PlayerID
			order := []game.PlayerID{game.Player3, game.Player4, game.Player1}
			agent := &controlScript{choice: func(_ PlayerObservation, request game.ChoiceRequest) []int {
				for player := range game.NumPlayers {
					if !slices.Equal(g.Players[player].Library.All(), libraries[player]) {
						t.Fatalf("player %d's library changed before every scry placement choice", player)
					}
				}
				if countEventsOfKind(g.Events, game.EventScry) != 0 {
					t.Fatal("a scry was published before the simultaneous placement")
				}
				if request.Kind == game.ChoiceMay {
					if len(scryPlayers) != 0 {
						t.Fatal("opponent looked before all participation decisions")
					}
					if len(request.ScryDecisions) != len(mayPlayers) {
						t.Fatal("later participation choice lost prior public decisions")
					}
					for index, decision := range request.ScryDecisions {
						if decision.Player != mayPlayers[index] || !decision.Accepted || decision.PlacementChosen {
							t.Fatal("participation history is not public, exact, or in APNAP order")
						}
					}
					mayPlayers = append(mayPlayers, request.Player)
					return []int{1}
				}
				if request.Kind != game.ChoiceScry ||
					!request.Subject.Exists || request.Subject.Val.CardID != libraries[request.Player][0] {
					t.Fatal("participant did not receive its own actual scry subject")
				}
				if optional && !slices.Equal(mayPlayers, order) {
					t.Fatalf("participation decisions = %v, want APNAP %v before looking", mayPlayers, order)
				}
				if len(request.ScryDecisions) != len(order) {
					t.Fatal("scry chooser lost public participant information")
				}
				for index, decision := range request.ScryDecisions {
					chosen := index < len(scryPlayers)
					if decision.Player != order[index] || !decision.Accepted ||
						decision.PlacementChosen != chosen || decision.TopCount != 0 ||
						(chosen && decision.BottomCount != 1) || (!chosen && decision.BottomCount != 0) {
						t.Fatal("scry chooser lost earlier public placement counts")
					}
				}
				request.ScryDecisions[0].Accepted = false
				scryPlayers = append(scryPlayers, request.Player)
				return []int{1}
			}}
			agents := [game.NumPlayers]PlayerAgent{agent, agent, agent, agent}
			engine := NewEngine(nil)
			clone := g.Clone()
			log := &TurnLog{}
			engine.resolveTopOfStackWithChoices(g, agents, log)
			if !slices.Equal(scryPlayers, order) {
				t.Fatalf("placement choices = %v, want APNAP %v", scryPlayers, order)
			}
			for player := range game.NumPlayers {
				want := slices.Clone(libraries[player])
				if game.PlayerID(player) != game.Player2 {
					slices.Reverse(want)
				}
				if !slices.Equal(g.Players[player].Library.All(), want) ||
					!slices.Equal(clone.Players[player].Library.All(), libraries[player]) {
					t.Fatal("simultaneous scry changed wrong library or shared mutable clone state")
				}
			}
			var scripts [game.NumPlayers]PlayerAgent
			for player := range game.NumPlayers {
				var answers [][]int
				for _, decision := range log.Choices {
					if decision.Request.Player == game.PlayerID(player) {
						answers = append(answers, slices.Clone(decision.Selected))
					}
				}
				scripts[player] = &choiceOnlyAgent{choices: answers}
			}
			replayLog := &TurnLog{}
			engine.resolveTopOfStackWithChoices(clone, scripts, replayLog)
			if !reflect.DeepEqual(log, replayLog) {
				t.Fatal("group-scry replay changed the recorded decisions")
			}
			for player := range game.NumPlayers {
				if !slices.Equal(g.Players[player].Library.All(), clone.Players[player].Library.All()) {
					t.Fatal("group-scry replay did not reproduce library placements")
				}
			}
			if countEventsOfKind(g.Events, game.EventScry) != 3 {
				t.Fatal("simultaneous scry event or in-library semantics are incorrect")
			}
			var batch id.ID
			for _, event := range g.Events {
				if event.Kind == game.EventScry {
					if batch == 0 {
						batch = event.SimultaneousID
					}
					if batch == 0 || event.SimultaneousID != batch {
						t.Fatal("simultaneous scries lost their common event cohort")
					}
				}
				if event.Kind == game.EventZoneChanged && event.FromZone == zone.Library {
					t.Fatal("scry created a zone change for a reordered library card")
				}
			}

			assertPrivateScryLog(t, log)
		})
	}
}

func TestGroupScryDeclineEmptyShortAndControlled(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	g.Turn.TurnNumber = 1
	g.Turn.ActivePlayer = game.Player3
	g.PlayerControlEffects = []game.PlayerControlEffect{{
		ID: g.IDGen.Next(), Player: game.Player3, Controller: game.Player4, Turn: 1,
	}}
	short := addCardToLibrary(g, game.Player3, vanillaCreature("Scry secret short", 2, 2))
	declined := addCardToLibrary(g, game.Player4, vanillaCreature("Scry secret declined", 2, 2))
	addInstructionSpellToStack(g, []game.Instruction{{
		Primitive: game.Scry{Player: game.GroupOfferMemberReference(), Amount: game.Fixed(3)},
		Optional:  true, OptionalActorGroup: opt.Val(game.OpponentsReference()), PublishResult: "scried",
	}})
	// Opponents are 3, 4, 2 in APNAP order; player 2 accepts with an empty library.
	var actors []game.PlayerID
	scryRequests := 0
	agent := &controlScript{choice: func(obs PlayerObservation, request game.ChoiceRequest) []int {
		if request.Player == game.Player3 && (obs.Player != game.Player3 || obs.DecisionMaker != game.Player4) {
			t.Fatal("controlled scry lost actor versus decision-maker identity")
		}
		if request.Kind == game.ChoiceMay {
			actors = append(actors, request.Player)
			if request.Player == game.Player4 {
				return []int{0}
			}
			return []int{1}
		}
		scryRequests++
		if request.Kind != game.ChoiceScry || request.Player != game.Player3 ||
			!request.Subject.Exists || request.Subject.Val.CardID != short ||
			!slices.Equal(actors, []game.PlayerID{game.Player3, game.Player4, game.Player2}) {
			t.Fatal("short/empty/declined group scry used incorrect choices")
		}
		if len(request.ScryDecisions) != 3 || request.ScryDecisions[1].Accepted {
			t.Fatal("decline was not visible before private placement")
		}
		return []int{1}
	}}
	log := &TurnLog{}
	NewEngine(nil).resolveTopOfStackWithChoices(g, [game.NumPlayers]PlayerAgent{agent, agent, nil, agent}, log)
	if scryRequests != 1 || !g.Players[game.Player3].Library.Contains(short) ||
		!g.Players[game.Player4].Library.Contains(declined) || countEventsOfKind(g.Events, game.EventScry) != 2 {
		t.Fatal("declined, empty, or short-library scry semantics changed")
	}
	for _, event := range g.Events {
		if event.Kind == game.EventScry && (event.Amount != 3 || event.Player == game.Player4) {
			t.Fatal("group scry event used actual library length or a declined participant")
		}
	}
	assertPrivateScryLog(t, log)
}

func TestGroupScryPartitionsOrdersAndPublicCounts(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	var libraries [game.NumPlayers][]id.ID
	for player := range game.NumPlayers {
		libraries[player] = make([]id.ID, 5)
		for index := 4; index >= 0; index-- {
			libraries[player][index] = addCardToLibrary(g, game.PlayerID(player), vanillaCreature("Scry secret partition", 2, 2))
		}
	}
	addInstructionSpellToStack(g, []game.Instruction{{
		Primitive:          game.Scry{Player: game.GroupOfferMemberReference(), Amount: game.Fixed(4)},
		ForEachPlayerGroup: opt.Val(game.OpponentsReference()), PublishResult: "scried",
	}})
	obj, _ := g.Stack.Peek()
	choices := 0
	agent := &controlScript{choice: func(_ PlayerObservation, request game.ChoiceRequest) []int {
		for player := range game.NumPlayers {
			if !slices.Equal(g.Players[player].Library.All(), libraries[player]) {
				t.Fatal("a group partition/order committed before later players finished")
			}
		}
		for _, decision := range request.ScryDecisions {
			if decision.Player < request.Player {
				if !decision.PlacementChosen || decision.TopCount != 2 || decision.BottomCount != 2 {
					t.Fatal("later scry chooser cannot see earlier public partition counts")
				}
			} else if decision.PlacementChosen {
				t.Fatal("scry exposed a future placement decision")
			}
		}
		choices++
		if request.Kind == game.ChoiceScry {
			if len(request.Options) != 4 || !request.Options[0].Card.Exists {
				t.Fatal("partition chooser lost its actual private cards")
			}
			return []int{0, 2}
		}
		if request.Kind != game.ChoiceOrder || len(request.Options) != 2 || !request.Options[0].Card.Exists {
			t.Fatal("top/bottom order chooser lost its private cards")
		}
		return []int{1, 0}
	}}
	log := &TurnLog{}
	NewEngine(nil).resolveTopOfStackWithChoices(g, [game.NumPlayers]PlayerAgent{agent, agent, agent, agent}, log)
	for _, player := range []game.PlayerID{game.Player2, game.Player3, game.Player4} {
		cards := libraries[player]
		if !slices.Equal(g.Players[player].Library.All(), []id.ID{cards[3], cards[1], cards[4], cards[2], cards[0]}) {
			t.Fatal("group scry lost the chosen partition or one of its two orders")
		}
	}
	receipt := obj.ResolutionResults["scried"]
	if choices != 9 || !receipt.Accepted || !receipt.Succeeded || receipt.Amount != 3 ||
		receipt.AcceptedActors.Count() != 3 {
		t.Fatal("group scry lost normal group-result publication")
	}
	assertPrivateScryLog(t, log)
}

func TestGroupScryFallbackAndDeclinedPublication(t *testing.T) {
	for _, decline := range []bool{false, true} {
		t.Run(fmt.Sprint(decline), func(t *testing.T) {
			g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
			for _, player := range []game.PlayerID{game.Player2, game.Player3, game.Player4} {
				for range 2 {
					addCardToLibrary(g, player, vanillaCreature("Scry secret fallback", 2, 2))
				}
			}
			addInstructionSpellToStack(g, []game.Instruction{{
				Primitive: game.Scry{Player: game.GroupOfferMemberReference(), Amount: game.Fixed(2)},
				Optional:  true, OptionalActorGroup: opt.Val(game.OpponentsReference()), PublishResult: "scried",
			}})
			obj, _ := g.Stack.Peek()
			agents := [game.NumPlayers]PlayerAgent{}
			if decline {
				agent := &choiceOnlyAgent{choices: [][]int{{0}}}
				agents[game.Player2] = agent
				agents[game.Player3] = &choiceOnlyAgent{choices: [][]int{{0}}}
				agents[game.Player4] = &choiceOnlyAgent{choices: [][]int{{0}}}
			}
			log := &TurnLog{}
			NewEngine(nil).resolveTopOfStackWithChoices(g, agents, log)
			receipt := obj.ResolutionResults["scried"]
			if decline {
				if len(log.Choices) != 3 || receipt.Accepted || receipt.Succeeded ||
					receipt.Amount != 0 || countEventsOfKind(g.Events, game.EventScry) != 0 {
					t.Fatal("declined group scry published success or prompted private placement")
				}
			} else if len(log.Choices) != 9 || g.Execution.ChoiceFallbacks != 9 ||
				!receipt.Accepted || !receipt.Succeeded || receipt.Amount != 3 ||
				countEventsOfKind(g.Events, game.EventScry) != 3 {
				t.Fatal("ordinary group-scry fallbacks lost participation, ordering, or result publication")
			}
			assertPrivateScryLog(t, log)
		})
	}
}
