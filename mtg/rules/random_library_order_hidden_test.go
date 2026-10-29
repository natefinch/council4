package rules

// Source review (#3437, shared Engine operation; no card admitted):
// Oracle (pinned corpus 6da9ced0e19bd384771ab8f14051cc45a8dc90302082b875fa6e039d2399a1b4):
//   - Bloodbraid Elf 3f0c9466-5ab9-4205-a84f-b4b27b5a678e {2}{R}{G}: "Cascade
//     (... Put the exiled cards on the bottom in a random order.)"
//   - Tana, the Bloodsower a7dff624-1f6b-4e00-b13b-0b0d69a48ec2 {2}{R}{G} and
//     Kraum, Ludevic's Opus 5f694f2d-9349-4988-b2b4-788136d4293e {3}{U}{R}: both
//     "Partner (You can have two commanders if both have partner.)"
//   - Curator of Sun's Creation 0e636c98-67c7-4f02-b909-5d43e276fa41 (discover).
//   - Tibalt's Trickery 6623cb85-fb1e-4561-8b21-0977c25096c2: "Then they put the
//     exiled cards on the bottom of their library in a random order."
// Comprehensive Rules (MagicCompRules 20260925): 401.2 "Each library must be
// kept in a single face-down pile. Players can't look at or change the order of
// cards in a library."; 702.85a cascade and 701.57a discover put the remaining
// exiled cards "on the bottom of your library in a random order"; 400.6 the
// replacement choices are made before the event moves the object; 903.9b a
// commander that would be put into its owner's library may be put into the
// command zone instead.
// So the commander choices for a random-order bottoming are made in an order
// that does not depend on the hidden random order (exile/linked order), and
// neither those choices nor the zone-change events reveal the library order.

import (
	"math/rand/v2"
	"slices"
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/id"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
)

const hiddenOrderSeeds = 64

// declineCommanderAgent declines every commander command-zone replacement and
// records the order of the commander subjects it was asked about.
func declineCommanderAgent(subjects *[]id.ID) *controlScript {
	return &controlScript{choice: func(_ PlayerObservation, req game.ChoiceRequest) []int {
		if req.Prompt != commanderZonePrompt {
			return fallbackChoice(req)
		}
		if req.Subject.Exists {
			*subjects = append(*subjects, req.Subject.Val.CardID)
		}
		return []int{0}
	}}
}

// relativeOrder returns the members of cards in the order they appear in all.
func relativeOrder(all, cards []id.ID) []id.ID {
	var order []id.ID
	for _, cardID := range all {
		if slices.Contains(cards, cardID) {
			order = append(order, cardID)
		}
	}
	return order
}

// exileToLibraryOrder returns the card order of the exile-to-library zone-change
// events.
func exileToLibraryOrder(g *game.Game) []id.ID {
	var order []id.ID
	for _, event := range g.Events {
		if event.Kind == game.EventZoneChanged && event.FromZone == zone.Exile && event.ToZone == zone.Library {
			order = append(order, event.CardID)
		}
	}
	return order
}

// loggedCommanderSubjects returns the commander-choice subjects recorded in log.
func loggedCommanderSubjects(log *TurnLog) []id.ID {
	var subjects []id.ID
	for _, choice := range log.Choices {
		if choice.Request.Prompt == commanderZonePrompt && choice.Request.Subject.Exists {
			subjects = append(subjects, choice.Request.Subject.Val.CardID)
		}
	}
	return subjects
}

// hiddenOrderRun is one seeded random bottoming of two declined commanders:
// want is the public (exile or linked) order they were offered in.
type hiddenOrderRun struct {
	want, asked, logged, events, library []id.ID
}

// assertRandomOrderHidden checks, across seeds, that the commander choices, the
// logged choices and the zone-change events always follow the public order,
// while the final library order of the two commanders really is random: both
// relative orders occur.
func assertRandomOrderHidden(t *testing.T, run func(t *testing.T, seed uint64) hiddenOrderRun) {
	t.Helper()
	seen := map[int]bool{}
	for seed := uint64(1); seed <= hiddenOrderSeeds; seed++ {
		got := run(t, seed)
		if len(got.want) != 2 {
			t.Fatalf("seed %d: public order %v, want two commanders", seed, got.want)
		}
		if !slices.Equal(got.asked, got.want) {
			t.Fatalf("seed %d: commander choices asked for %v, want public order %v", seed, got.asked, got.want)
		}
		if !slices.Equal(got.logged, got.want) {
			t.Fatalf("seed %d: logged commander choices %v, want public order %v", seed, got.logged, got.want)
		}
		if !slices.Equal(got.events, got.want) {
			t.Fatalf("seed %d: exile-to-library events %v, want public order %v", seed, got.events, got.want)
		}
		order := relativeOrder(got.library, got.want)
		if !sameCardIDs(order, got.want) {
			t.Fatalf("seed %d: library holds %v, want both declined commanders %v", seed, order, got.want)
		}
		seen[slices.Index(got.want, order[0])] = true
	}
	if len(seen) != 2 {
		t.Fatalf("library order never varied across %d seeds; want a random bottom order", hiddenOrderSeeds)
	}
}

// TestBloodbraidElfCascadeCommanderChoicesHideRandomBottomOrder reproduces
// #3437 with the default-registry cards: Tana and Kraum are the only cards in
// the library, Bloodbraid Elf's cascade exiles both without a hit, and both
// command-zone replacements are declined. The choices and events always follow
// the exile order, independent of the random library order.
func TestBloodbraidElfCascadeCommanderChoicesHideRandomBottomOrder(t *testing.T) {
	assertRandomOrderHidden(t, func(t *testing.T, seed uint64) hiddenOrderRun {
		f := newCascadeTriggerFixture(t)
		f.engine = NewEngineWithPrograms(rand.New(rand.NewPCG(seed, seed)), f.registry.Programs())
		g := f.game
		kraum := addCardToLibrary(g, game.Player1, f.def(t, "Kraum, Ludevic's Opus"))
		tana := addCardToLibrary(g, game.Player1, f.def(t, "Tana, the Bloodsower"))
		g.CommanderIDs = map[id.ID]bool{tana: true, kraum: true}
		f.castBloodbraid(t)

		var asked []id.ID
		agent := declineCommanderAgent(&asked)
		agents := [game.NumPlayers]PlayerAgent{agent, agent, agent, agent}
		log := &TurnLog{}
		f.engine.putTriggeredAbilitiesOnStackWithChoices(g, agents, log)
		if top, ok := g.Stack.Peek(); !ok || !isCascadeTrigger(top) {
			t.Fatalf("stack top = %+v, want Bloodbraid Elf's cascade trigger", top)
		}
		f.engine.resolveTopOfStackWithChoices(g, agents, log)
		if g.Players[game.Player1].Exile.Size() != 0 || g.Stack.Size() != 1 {
			t.Fatal("cascade did not bottom both commanders and leave only Bloodbraid Elf on the stack")
		}
		return hiddenOrderRun{want: []id.ID{tana, kraum}, asked: asked, logged: loggedCommanderSubjects(log),
			events: exileToLibraryOrder(g), library: g.Players[game.Player1].Library.All()}
	})
}

// TestDiscoverCommanderChoicesHideRandomBottomOrder: discover's remaining exiled
// cards (CR 701.57a) use the same random bottoming. Discover 3 finds no hit
// above an unrelated library card, so both exiled commanders are offered in
// exile order and then bottomed beneath it in a hidden random order.
func TestDiscoverCommanderChoicesHideRandomBottomOrder(t *testing.T) {
	assertRandomOrderHidden(t, func(t *testing.T, seed uint64) hiddenOrderRun {
		g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
		engine := NewEngine(rand.New(rand.NewPCG(seed, seed)))
		addEffectSpellToStack(g, game.Player1, game.DiscoverCards{Amount: game.Fixed(3)}, nil)
		second := addCardToLibrary(g, game.Player1, simpleGainLifeInstantWithManaValue("Second Partner", 5))
		first := addCardToLibrary(g, game.Player1, simpleGainLifeInstantWithManaValue("First Partner", 4))
		g.CommanderIDs = map[id.ID]bool{first: true, second: true}

		var asked []id.ID
		agent := declineCommanderAgent(&asked)
		log := &TurnLog{}
		engine.resolveTopOfStackWithChoices(g, [game.NumPlayers]PlayerAgent{agent, agent, agent, agent}, log)
		if g.Players[game.Player1].Exile.Size() != 0 || g.Stack.Size() != 0 {
			t.Fatal("discover did not bottom both exiled commanders")
		}
		return hiddenOrderRun{want: []id.ID{first, second}, asked: asked, logged: loggedCommanderSubjects(log),
			events: exileToLibraryOrder(g), library: g.Players[game.Player1].Library.All()}
	})
}

// TestLinkedRandomDisposalCommanderChoicesHideRandomBottomOrder: a linked
// random-order disposal (Tibalt's Trickery's PutLinkedExiledCardsInLibrary
// with RandomOrder) offers each commander in linked order and then randomly
// orders the cards that actually arrive beneath the existing library.
func TestLinkedRandomDisposalCommanderChoicesHideRandomBottomOrder(t *testing.T) {
	assertRandomOrderHidden(t, func(t *testing.T, seed uint64) hiddenOrderRun {
		g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
		engine := NewEngine(rand.New(rand.NewPCG(seed, seed)))
		existing := addCardToLibraryNamed(g, game.Player1, "Library Card")
		first := addExiledInstant(g, game.Player1, disposalVanilla("First Partner", types.Creature))
		second := addExiledInstant(g, game.Player1, disposalVanilla("Second Partner", types.Creature))
		g.CommanderIDs = map[id.ID]bool{first: true, second: true}
		pushLinkedDisposal(g, true, first, second)

		var asked []id.ID
		agent := declineCommanderAgent(&asked)
		log := &TurnLog{}
		engine.resolveTopOfStackWithChoices(g, [game.NumPlayers]PlayerAgent{agent, agent, agent, agent}, log)
		library := g.Players[game.Player1].Library.All()
		if len(library) != 3 || library[0] != existing {
			t.Fatalf("library = %v, want %v above both bottomed commanders", library, existing)
		}
		return hiddenOrderRun{want: []id.ID{first, second}, asked: asked, logged: loggedCommanderSubjects(log),
			events: exileToLibraryOrder(g), library: library}
	})
}

// TestRandomBottomingRandomizesOnlyArrivedCards: a card the owner sends to the
// command zone is not part of the random library order, and cards already in
// the library keep their places while the arrivals are permuted among
// themselves beneath them.
func TestRandomBottomingRandomizesOnlyArrivedCards(t *testing.T) {
	seen := map[id.ID]bool{}
	for seed := uint64(1); seed <= hiddenOrderSeeds; seed++ {
		g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
		engine := NewEngine(rand.New(rand.NewPCG(seed, seed)))
		bottom := addCardToLibraryNamed(g, game.Player1, "Bottom Library Card")
		top := addCardToLibraryNamed(g, game.Player1, "Top Library Card")
		commander := addExiledInstant(g, game.Player1, disposalVanilla("Commander", types.Creature))
		plain := []id.ID{
			addExiledInstant(g, game.Player1, disposalVanilla("Exiled A", types.Instant)),
			addExiledInstant(g, game.Player1, disposalVanilla("Exiled B", types.Instant)),
		}
		g.CommanderIDs = map[id.ID]bool{commander: true}
		pushLinkedDisposal(g, true, plain[0], commander, plain[1])
		engine.resolveTopOfStackWithChoices(g, [game.NumPlayers]PlayerAgent{}, &TurnLog{})

		if !g.Players[game.Player1].CommandZone.Contains(commander) {
			t.Fatalf("seed %d: commander was not put into the command zone", seed)
		}
		library := g.Players[game.Player1].Library.All()
		if len(library) != 4 || library[0] != top || library[1] != bottom || !sameCardIDs(library[2:], plain) {
			t.Fatalf("seed %d: library = %v, want %v, %v then %v in any order", seed, library, top, bottom, plain)
		}
		seen[library[2]] = true
	}
	if len(seen) != 2 {
		t.Fatal("the arrived cards were never randomly reordered")
	}
}
