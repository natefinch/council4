package rules

import (
	"strings"
	"testing"

	"github.com/natefinch/council4/mtg/cards"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/id"
	"github.com/natefinch/council4/mtg/game/zone"
)

// Replacement choices for a card leaving a hidden zone stay private (#3369).
//
// Source review:
//   - Comprehensive Rules effective 2026-09-25,
//     https://media.wizards.com/2026/downloads/MagicCompRules%2020260925.txt
//     (sha256 8d860e451f20f38865b725b42d82feb714c725373dd8f3b32b8652b3eeb070ca):
//     400.2 ("Library and hand are hidden zones"), 406.3 ("Cards 'exiled face
//     down' can't be examined by any player except when instructions allow
//     it"), 616.1 (the affected object's controller, or its owner if it has no
//     controller, or the affected player chooses one replacement to apply),
//     708.2 and 708.5 (a face-down permanent's identity is visible only to its
//     controller).
//   - Progenitus, oracle_id 2d7e00b6-12f0-4b03-82a6-50e3d1b5395e: "If
//     Progenitus would be put into a graveyard from anywhere, reveal
//     Progenitus and shuffle it into its owner's library instead."
//   - Rest in Peace, oracle_id 087f9ad7-e74f-40e2-8102-1ed2925d0418: "If a
//     card or token would be put into a graveyard from anywhere, exile it
//     instead."
//   - Leyline of the Void, oracle_id f4e32fc1-1b8d-441e-8e76-71f19f98e925: "If
//     a card would be put into an opponent's graveyard from anywhere, exile it
//     instead."
//   - Mind Rot, oracle_id ad44cf74-b717-48fb-9fa2-77512024d76a: "Target player
//     discards two cards."
//   - Glimpse the Unthinkable, oracle_id
//     552f0163-a19d-4671-888f-044fc0354875: "Target player mills ten cards."
//
// Progenitus's own replacement competes with Rest in Peace, and its option
// label names Progenitus. While the card is hidden, the owner still sees the
// labels, but the turn log and recorded decision must not name the card. A
// face-up card in a public zone keeps a public, labelled choice.

const hiddenReplacementPrivateLabel = "Private option"

type hiddenReplacementCards struct {
	registry                                           *cards.Registry
	progenitus, restInPeace, leyline, mindRot, glimpse *game.CardDef
}

func hiddenReplacementRegistry(t *testing.T) hiddenReplacementCards {
	t.Helper()
	registry := cards.NewDefaultRegistry()
	set := hiddenReplacementCards{
		registry:    registry,
		progenitus:  registry.Lookup("Progenitus"),
		restInPeace: registry.Lookup("Rest in Peace"),
		leyline:     registry.Lookup("Leyline of the Void"),
		mindRot:     registry.Lookup("Mind Rot"),
		glimpse:     registry.Lookup("Glimpse the Unthinkable"),
	}
	if set.progenitus == nil || set.restInPeace == nil || set.leyline == nil || set.mindRot == nil || set.glimpse == nil {
		t.Fatal("registered Progenitus, Rest in Peace, Leyline of the Void, Mind Rot or Glimpse the Unthinkable missing")
	}
	return set
}

// resolveHiddenReplacementPermanent resolves a registered permanent spell for
// controller, so its continuous replacement effect is registered.
func resolveHiddenReplacementPermanent(t *testing.T, g *game.Game, engine *Engine, controller game.PlayerID, def *game.CardDef) {
	t.Helper()
	before := len(g.ReplacementEffects)
	addImplementationSpellToStack(g, controller, def, nil)
	engine.resolveTopOfStack(g, &TurnLog{})
	if len(g.ReplacementEffects) != before+1 {
		t.Fatalf("%s replacement effects = %d, want %d", def.Name, len(g.ReplacementEffects), before+1)
	}
}

// hiddenReplacementAgents has chooser pick Rest in Peace or Progenitus when
// asked to order replacements, recording every replacement request.
func hiddenReplacementAgents(t *testing.T, chooser game.PlayerID, pickProgenitus bool, requests *[]game.ChoiceRequest) [game.NumPlayers]PlayerAgent {
	t.Helper()
	var agents [game.NumPlayers]PlayerAgent
	for _, player := range []game.PlayerID{game.Player1, game.Player2, game.Player3, game.Player4} {
		agents[player] = &authoredCopyChoiceAgent{choose: func(request game.ChoiceRequest) []int {
			if request.Kind != game.ChoiceReplacement {
				return request.DefaultSelection
			}
			if player != chooser || request.Player != chooser {
				t.Fatalf("%v was asked to order replacements for %v, want %v", player, request.Player, chooser)
			}
			*requests = append(*requests, request)
			for _, option := range request.Options {
				if strings.Contains(option.Label, "Progenitus") == pickProgenitus {
					return []int{option.Index}
				}
			}
			return request.DefaultSelection
		}}
	}
	return agents
}

// requireReplacementChoice checks the one replacement decision: its chooser
// saw the real labels, and the log and the recorded decision are private (no
// label naming the card) exactly when private is set.
func requireReplacementChoice(t *testing.T, g *game.Game, log *TurnLog, requests []game.ChoiceRequest, chooser game.PlayerID, private bool, wantLabel string) {
	t.Helper()
	if len(requests) != 1 || requests[0].Private != private || len(requests[0].Options) != 2 {
		t.Fatalf("replacement requests = %+v, want one two-way choice with Private=%v", requests, private)
	}
	seen := false
	for _, option := range requests[0].Options {
		seen = seen || strings.Contains(option.Label, wantLabel)
	}
	if !seen {
		t.Fatalf("chooser options %+v lack %q", requests[0].Options, wantLabel)
	}
	var logged []game.ChoiceDecision
	for _, decision := range log.Choices {
		if decision.Request.Kind == game.ChoiceReplacement {
			logged = append(logged, decision)
		}
	}
	if len(logged) != 1 || logged[0].Request.Player != chooser || logged[0].Request.Private != private {
		t.Fatalf("logged replacement choices = %+v, want one for %v with Private=%v", logged, chooser, private)
	}
	if len(g.ReplacementDecisions) != 1 || g.ReplacementDecisions[0].Player != chooser ||
		len(g.ReplacementDecisions[0].Options) != 2 || g.ReplacementDecisions[0].UsedFallback {
		t.Fatalf("replacement decisions = %+v, want one agent decision by %v", g.ReplacementDecisions, chooser)
	}
	if g.ReplacementDecisions[0].Selected[0] != logged[0].Selected[0] {
		t.Fatalf("recorded selection %v differs from logged selection %v", g.ReplacementDecisions[0].Selected, logged[0].Selected)
	}
	for i, option := range logged[0].Request.Options {
		recorded := g.ReplacementDecisions[0].Options[i]
		if private {
			if option.Label != hiddenReplacementPrivateLabel || recorded != hiddenReplacementPrivateLabel ||
				logged[0].Request.Prompt != "Private decision" {
				t.Fatalf("private choice leaked: logged %+v, recorded %q", logged[0].Request, recorded)
			}
			continue
		}
		if option.Label != requests[0].Options[i].Label || recorded != requests[0].Options[i].Label {
			t.Fatalf("public choice lost its labels: logged %+v, recorded %q", logged[0].Request, recorded)
		}
	}
}

func hiddenReplacementOutcome(t *testing.T, g *game.Game, owner game.PlayerID, cardID id.ID, pickProgenitus bool) {
	t.Helper()
	player := g.Players[owner]
	want, other := player.Exile, player.Library
	if pickProgenitus {
		want, other = player.Library, player.Exile
	}
	if !want.Contains(cardID) || other.Contains(cardID) || player.Graveyard.Contains(cardID) || player.Hand.Contains(cardID) {
		t.Fatalf("Progenitus did not follow the chosen replacement (Progenitus first=%v)", pickProgenitus)
	}
}

func hiddenReplacementChoiceName(pickProgenitus bool) string {
	if pickProgenitus {
		return "Progenitus first"
	}
	return "Rest in Peace first"
}

// Mind Rot discards a hand together (the discard batch path); Progenitus's
// choice is private.
func TestHiddenReplacementChoiceMindRotDiscardIsPrivate(t *testing.T) {
	t.Parallel()
	for _, pickProgenitus := range []bool{false, true} {
		t.Run(hiddenReplacementChoiceName(pickProgenitus), func(t *testing.T) {
			t.Parallel()
			set := hiddenReplacementRegistry(t)
			g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
			engine := NewEngineWithPrograms(nil, set.registry.Programs())
			resolveHiddenReplacementPermanent(t, g, engine, game.Player1, set.restInPeace)
			progenitus := addCardToHand(g, game.Player2, set.progenitus)
			other := addCardToHand(g, game.Player2, vanillaCreature("Hidden Bear", 2, 2))
			addImplementationSpellToStack(g, game.Player1, set.mindRot,
				[]game.Target{{Kind: game.TargetPlayer, PlayerID: game.Player2}})
			var requests []game.ChoiceRequest
			log := &TurnLog{}
			engine.resolveTopOfStackWithChoices(g, hiddenReplacementAgents(t, game.Player2, pickProgenitus, &requests), log)
			requireReplacementChoice(t, g, log, requests, game.Player2, true, "Progenitus")
			hiddenReplacementOutcome(t, g, game.Player2, progenitus, pickProgenitus)
			if !g.Players[game.Player2].Exile.Contains(other) || g.Players[game.Player2].Hand.Size() != 0 {
				t.Fatal("Mind Rot did not discard both cards into Rest in Peace's exile")
			}
		})
	}
}

// Glimpse the Unthinkable mills ten cards as one batch; Progenitus, milled
// last, gets a private choice.
func TestHiddenReplacementChoiceGlimpseMillIsPrivate(t *testing.T) {
	t.Parallel()
	for _, pickProgenitus := range []bool{false, true} {
		t.Run(hiddenReplacementChoiceName(pickProgenitus), func(t *testing.T) {
			t.Parallel()
			set := hiddenReplacementRegistry(t)
			g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
			engine := NewEngineWithPrograms(nil, set.registry.Programs())
			resolveHiddenReplacementPermanent(t, g, engine, game.Player1, set.restInPeace)
			progenitus := addCardToLibrary(g, game.Player2, set.progenitus)
			var others []id.ID
			for range 9 {
				others = append(others, addCardToLibrary(g, game.Player2, vanillaCreature("Hidden Bear", 2, 2)))
			}
			addImplementationSpellToStack(g, game.Player1, set.glimpse,
				[]game.Target{{Kind: game.TargetPlayer, PlayerID: game.Player2}})
			var requests []game.ChoiceRequest
			log := &TurnLog{}
			engine.resolveTopOfStackWithChoices(g, hiddenReplacementAgents(t, game.Player2, pickProgenitus, &requests), log)
			requireReplacementChoice(t, g, log, requests, game.Player2, true, "Progenitus")
			hiddenReplacementOutcome(t, g, game.Player2, progenitus, pickProgenitus)
			for _, cardID := range others {
				if !g.Players[game.Player2].Exile.Contains(cardID) {
					t.Fatal("Glimpse the Unthinkable did not mill the other cards into Rest in Peace's exile")
				}
			}
		})
	}
}

// Cards moving together out of several zones (the simultaneous card batch
// path) each get a choice whose privacy follows that card's own visibility.
func TestHiddenReplacementChoiceSimultaneousBatchPerCard(t *testing.T) {
	t.Parallel()
	set := hiddenReplacementRegistry(t)
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngineWithPrograms(nil, set.registry.Programs())
	resolveHiddenReplacementPermanent(t, g, engine, game.Player1, set.restInPeace)
	fromLibrary := addCardToLibrary(g, game.Player2, set.progenitus)
	faceDown := addCardToExile(g, game.Player2, set.progenitus)
	g.Players[game.Player2].Exile.SetFaceDown(faceDown, true)
	faceUp := addCardToExile(g, game.Player2, set.progenitus)
	var moves []simultaneousCardMove
	for _, move := range []struct {
		cardID id.ID
		from   zone.Type
	}{{fromLibrary, zone.Library}, {faceDown, zone.Exile}, {faceUp, zone.Exile}} {
		card, _ := g.GetCardInstance(move.cardID)
		moves = append(moves, simultaneousCardMove{
			card: game.ZoneCardSnapshot{CardID: card.ID, ZoneVersion: card.ZoneVersion}, fromZone: move.from,
		})
	}
	var requests []game.ChoiceRequest
	log := &TurnLog{}
	engine.setReplacementChoiceContext(g, hiddenReplacementAgents(t, game.Player2, false, &requests), log)
	if !moveCardsBetweenZonesSimultaneously(g, moves, zone.Graveyard, false, g.IDGen.Next(), nil) {
		t.Fatal("the batch did not move")
	}
	g.ClearChoiceContext()
	wantPrivate := []bool{true, true, false}
	if len(requests) != 3 || len(log.Choices) != 3 || len(g.ReplacementDecisions) != 3 {
		t.Fatalf("requests %d, logged %d, recorded %d; want one choice per card", len(requests), len(log.Choices), len(g.ReplacementDecisions))
	}
	for i, private := range wantPrivate {
		logged := log.Choices[i].Request
		recorded := g.ReplacementDecisions[i].Options
		if requests[i].Private != private || logged.Private != private ||
			!strings.Contains(requests[i].Options[1].Label, "Progenitus") {
			t.Fatalf("choice %d = %+v, want Private=%v with the chooser seeing Progenitus", i, requests[i], private)
		}
		leaked := strings.Contains(logged.Options[1].Label, "Progenitus") || strings.Contains(recorded[1], "Progenitus")
		if leaked == private || private && (logged.Options[0].Label != hiddenReplacementPrivateLabel || recorded[0] != hiddenReplacementPrivateLabel) {
			t.Fatalf("choice %d logged %+v and recorded %q, want labels withheld exactly when private", i, logged.Options, recorded)
		}
	}
	for _, cardID := range []id.ID{fromLibrary, faceDown, faceUp} {
		if !g.Players[game.Player2].Exile.Contains(cardID) {
			t.Fatal("Rest in Peace did not exile every batch member")
		}
	}
}

// A single card moving from a library, hand or face-down exile to a graveyard
// (the single-move path) gets a private choice; the same card face up in exile
// gets a public one.
func TestHiddenReplacementChoiceSingleMoveFollowsZoneVisibility(t *testing.T) {
	t.Parallel()
	for _, tc := range []struct {
		name     string
		from     zone.Type
		faceDown bool
		private  bool
	}{
		{"library", zone.Library, false, true},
		{"hand", zone.Hand, false, true},
		{"face-down exile", zone.Exile, true, true},
		{"face-up exile", zone.Exile, false, false},
	} {
		for _, pickProgenitus := range []bool{false, true} {
			t.Run(tc.name+"/"+hiddenReplacementChoiceName(pickProgenitus), func(t *testing.T) {
				t.Parallel()
				set := hiddenReplacementRegistry(t)
				g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
				engine := NewEngineWithPrograms(nil, set.registry.Programs())
				resolveHiddenReplacementPermanent(t, g, engine, game.Player1, set.restInPeace)
				var progenitus id.ID
				switch tc.from {
				case zone.Library:
					progenitus = addCardToLibrary(g, game.Player2, set.progenitus)
				case zone.Hand:
					progenitus = addCardToHand(g, game.Player2, set.progenitus)
				default:
					progenitus = addCardToExile(g, game.Player2, set.progenitus)
					g.Players[game.Player2].Exile.SetFaceDown(progenitus, tc.faceDown)
				}
				var requests []game.ChoiceRequest
				log := &TurnLog{}
				engine.setReplacementChoiceContext(g, hiddenReplacementAgents(t, game.Player2, pickProgenitus, &requests), log)
				if !moveCardBetweenZonesInBatch(g, game.Player2, progenitus, tc.from, zone.Graveyard, false, 0) {
					t.Fatal("Progenitus did not move")
				}
				g.ClearChoiceContext()
				requireReplacementChoice(t, g, log, requests, game.Player2, tc.private, "Progenitus")
				hiddenReplacementOutcome(t, g, game.Player2, progenitus, pickProgenitus)
			})
		}
	}
}

// Permanents leaving the battlefield together (the permanent batch path): a
// face-up Progenitus keeps a public choice; a face-down one, whose identity
// only its controller may know, gets a private one between Rest in Peace and
// Leyline of the Void (a face-down permanent has no replacement of its own).
func TestHiddenReplacementChoicePermanentBatchFollowsFaceDown(t *testing.T) {
	t.Parallel()
	for _, faceDown := range []bool{false, true} {
		name := "face up"
		if faceDown {
			name = "face down"
		}
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			set := hiddenReplacementRegistry(t)
			g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
			engine := NewEngineWithPrograms(nil, set.registry.Programs())
			resolveHiddenReplacementPermanent(t, g, engine, game.Player1, set.restInPeace)
			var permanent *game.Permanent
			wantLabel := "Progenitus"
			if faceDown {
				resolveHiddenReplacementPermanent(t, g, engine, game.Player1, set.leyline)
				permanent = addFaceDownPermanent(g, game.Player2, set.progenitus, game.FaceDownMorph)
				wantLabel = "opponent's graveyard"
			} else {
				addImplementationSpellToStack(g, game.Player2, set.progenitus, nil)
				engine.resolveTopOfStack(g, &TurnLog{})
				for _, candidate := range g.Battlefield {
					if candidate.Controller == game.Player2 {
						permanent = candidate
					}
				}
			}
			if permanent == nil || permanent.FaceDown != faceDown {
				t.Fatal("Progenitus is not on the battlefield in the expected face")
			}
			cardID := permanent.CardInstanceID
			var requests []game.ChoiceRequest
			log := &TurnLog{}
			engine.setReplacementChoiceContext(g, hiddenReplacementAgents(t, game.Player2, !faceDown, &requests), log)
			results := movePermanentsToZoneSimultaneouslyWithResultsInBatch(g, []*game.Permanent{permanent}, zone.Graveyard, g.IDGen.Next())
			g.ClearChoiceContext()
			if len(results) != 1 || !results[0].moved {
				t.Fatal("Progenitus did not leave the battlefield")
			}
			requireReplacementChoice(t, g, log, requests, game.Player2, faceDown, wantLabel)
			if faceDown {
				if !g.Players[game.Player2].Exile.Contains(cardID) {
					t.Fatal("face-down Progenitus was not exiled")
				}
				return
			}
			hiddenReplacementOutcome(t, g, game.Player2, cardID, true)
		})
	}
}
