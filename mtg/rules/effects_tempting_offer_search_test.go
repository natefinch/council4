package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// temptingOfferSearchInstruction models the Tempt with Discovery idiom: the
// acting player searches their own library for a land card and puts it onto the
// battlefield. The acting player is addressed through GroupOfferMemberReference()
// so the controller searches for the base and each reward while each accepting
// opponent searches their own library, and every found land enters under the
// searcher's control.
func temptingOfferSearchInstruction() game.Instruction {
	return game.Instruction{
		Primitive: game.Search{
			Player: game.GroupOfferMemberReference(),
			Spec: game.SearchSpec{
				SourceZone:  zone.Library,
				Destination: zone.Battlefield,
				Filter:      game.Selection{RequiredTypes: []types.Card{types.Land}},
			},
			Amount: game.Fixed(1),
		},
		Optional:           true,
		OptionalActorGroup: opt.Val(game.OpponentsReference()),
		TemptingOffer:      true,
	}
}

func permanentsByController(g *game.Game) map[game.PlayerID]int {
	counts := make(map[game.PlayerID]int)
	for _, permanent := range g.Battlefield {
		counts[permanent.Controller]++
	}
	return counts
}

// librarySearchCounts counts the EventLibrarySearched events per searching
// player. searchLibrary emits exactly one such event and shuffles the searched
// library each time it runs, so this count is also the number of times each
// player shuffled: a player who never searched never shuffled.
func librarySearchCounts(g *game.Game) map[game.PlayerID]int {
	counts := make(map[game.PlayerID]int)
	for _, event := range g.Events {
		if event.Kind == game.EventLibrarySearched {
			counts[event.Player]++
		}
	}
	return counts
}

func addForests(g *game.Game, playerID game.PlayerID, n int) {
	for range n {
		addCardToLibrary(g, playerID, basicForestDef())
	}
}

// TestTemptingOfferSearchOwnershipRepeatAndShuffle proves the search idiom (Tempt
// with Discovery) with a mixed field: every searcher searches their OWN library
// and puts the found land under their OWN control, the controller repeats the
// search exactly once per accepting opponent, and only the players who searched
// (the controller and the accepters) shuffle.
func TestTemptingOfferSearchOwnershipRepeatAndShuffle(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	// The controller searches once for the base plus once per accepting opponent
	// (two here), so it needs three fetchable lands in its own library.
	addForests(g, game.Player1, 3)
	addCardToLibrary(g, game.Player2, basicIslandDef())
	addCardToLibrary(g, game.Player3, basicIslandDef())
	addCardToLibrary(g, game.Player4, basicIslandDef())
	addInstructionSpellToStack(g, []game.Instruction{temptingOfferSearchInstruction()})

	// Player2 and Player3 accept and fetch their Island; Player4 declines.
	agents := [game.NumPlayers]PlayerAgent{
		game.Player1: optionalSearchAgent{wanted: "Forest"},
		game.Player2: optionalSearchAgent{accept: true, wanted: "Island"},
		game.Player3: optionalSearchAgent{accept: true, wanted: "Island"},
		game.Player4: optionalSearchAgent{accept: false, wanted: "Island"},
	}

	engine.resolveTopOfStackWithChoices(g, agents, &TurnLog{})

	permanents := permanentsByController(g)
	// Controller: one base land plus one reward land per accepting opponent (2).
	if permanents[game.Player1] != 3 {
		t.Fatalf("controller controls %d lands, want 3 (1 base + 2 rewards)", permanents[game.Player1])
	}
	if permanents[game.Player2] != 1 {
		t.Fatalf("Player2 controls %d lands, want 1 (accepted, own library)", permanents[game.Player2])
	}
	if permanents[game.Player3] != 1 {
		t.Fatalf("Player3 controls %d lands, want 1 (accepted, own library)", permanents[game.Player3])
	}
	if permanents[game.Player4] != 0 {
		t.Fatalf("Player4 controls %d lands, want 0 (declined)", permanents[game.Player4])
	}

	// Every land on the battlefield is a Forest under the controller or an Island
	// under its own opponent: no searcher reached another player's library.
	for _, permanent := range g.Battlefield {
		name := g.CardInstances[permanent.CardInstanceID].Def.Name
		switch permanent.Controller {
		case game.Player1:
			if name != "Forest" {
				t.Fatalf("controller land is %q, want Forest from its own library", name)
			}
		default:
			if name != "Island" {
				t.Fatalf("opponent %v land is %q, want Island from its own library", permanent.Controller, name)
			}
		}
	}

	searches := librarySearchCounts(g)
	// The controller searched (and shuffled) for the base plus each reward (3);
	// each accepter searched once; the decliner never searched, so never shuffled.
	if searches[game.Player1] != 3 {
		t.Fatalf("controller searched %d times, want 3 (1 base + 2 rewards)", searches[game.Player1])
	}
	if searches[game.Player2] != 1 || searches[game.Player3] != 1 {
		t.Fatalf("accepter searches = %d/%d, want 1/1", searches[game.Player2], searches[game.Player3])
	}
	if searches[game.Player4] != 0 {
		t.Fatalf("decliner searched %d times, want 0 (no shuffle)", searches[game.Player4])
	}
}

// TestTemptingOfferSearchAllAccept proves that when every opponent accepts, the
// controller repeats the search once per opponent (three rewards) and each
// opponent fetches from their own library.
func TestTemptingOfferSearchAllAccept(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	// One base search plus one reward per opponent (three) needs four lands.
	addForests(g, game.Player1, 4)
	addCardToLibrary(g, game.Player2, basicIslandDef())
	addCardToLibrary(g, game.Player3, basicIslandDef())
	addCardToLibrary(g, game.Player4, basicIslandDef())
	addInstructionSpellToStack(g, []game.Instruction{temptingOfferSearchInstruction()})

	agents := [game.NumPlayers]PlayerAgent{
		game.Player1: optionalSearchAgent{wanted: "Forest"},
		game.Player2: optionalSearchAgent{accept: true, wanted: "Island"},
		game.Player3: optionalSearchAgent{accept: true, wanted: "Island"},
		game.Player4: optionalSearchAgent{accept: true, wanted: "Island"},
	}

	engine.resolveTopOfStackWithChoices(g, agents, &TurnLog{})

	permanents := permanentsByController(g)
	if permanents[game.Player1] != 4 {
		t.Fatalf("controller controls %d lands, want 4 (1 base + 3 rewards)", permanents[game.Player1])
	}
	for _, pid := range []game.PlayerID{game.Player2, game.Player3, game.Player4} {
		if permanents[pid] != 1 {
			t.Fatalf("opponent %v controls %d lands, want 1", pid, permanents[pid])
		}
	}
}

// TestTemptingOfferSearchAllDecline proves that when no opponent accepts, only
// the controller searches once for the base land and no reward search runs, so
// only the controller shuffles.
func TestTemptingOfferSearchAllDecline(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	addForests(g, game.Player1, 4)
	addCardToLibrary(g, game.Player2, basicIslandDef())
	addCardToLibrary(g, game.Player3, basicIslandDef())
	addCardToLibrary(g, game.Player4, basicIslandDef())
	addInstructionSpellToStack(g, []game.Instruction{temptingOfferSearchInstruction()})

	agents := [game.NumPlayers]PlayerAgent{
		game.Player1: optionalSearchAgent{wanted: "Forest"},
		game.Player2: optionalSearchAgent{accept: false, wanted: "Island"},
		game.Player3: optionalSearchAgent{accept: false, wanted: "Island"},
		game.Player4: optionalSearchAgent{accept: false, wanted: "Island"},
	}

	engine.resolveTopOfStackWithChoices(g, agents, &TurnLog{})

	permanents := permanentsByController(g)
	if permanents[game.Player1] != 1 {
		t.Fatalf("controller controls %d lands, want 1 (base only)", permanents[game.Player1])
	}
	for _, pid := range []game.PlayerID{game.Player2, game.Player3, game.Player4} {
		if permanents[pid] != 0 {
			t.Fatalf("declining opponent %v controls %d lands, want 0", pid, permanents[pid])
		}
	}
	searches := librarySearchCounts(g)
	if searches[game.Player1] != 1 {
		t.Fatalf("controller searched %d times, want 1 (base only)", searches[game.Player1])
	}
	for _, pid := range []game.PlayerID{game.Player2, game.Player3, game.Player4} {
		if searches[pid] != 0 {
			t.Fatalf("declining opponent %v searched %d times, want 0 (no shuffle)", pid, searches[pid])
		}
	}
}

// TestTemptingOfferSearchAcceptedButEmptyStillCounts proves the reward repeat
// follows acceptance, not a successful find (Tempt with Discovery's official
// ruling): an opponent who accepts and searches but finds no land still searched
// a library this way, so the controller repeats for that opponent and that
// opponent still shuffled.
func TestTemptingOfferSearchAcceptedButEmptyStillCounts(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	// The controller searches for the base plus one reward for the sole accepter.
	addForests(g, game.Player1, 2)
	// Player2 accepts but its library holds no land, so its search finds nothing.
	addCardToLibrary(g, game.Player2, &game.CardDef{CardFace: game.CardFace{
		Name:  "Bear",
		Types: []types.Card{types.Creature},
	}})
	addInstructionSpellToStack(g, []game.Instruction{temptingOfferSearchInstruction()})

	agents := [game.NumPlayers]PlayerAgent{
		game.Player1: optionalSearchAgent{wanted: "Forest"},
		game.Player2: optionalSearchAgent{accept: true, wanted: "Forest"},
		game.Player3: optionalSearchAgent{accept: false, wanted: "Island"},
		game.Player4: optionalSearchAgent{accept: false, wanted: "Island"},
	}

	engine.resolveTopOfStackWithChoices(g, agents, &TurnLog{})

	permanents := permanentsByController(g)
	// The controller repeated its search once for the accepting opponent even
	// though that opponent found nothing: one base plus one reward.
	if permanents[game.Player1] != 2 {
		t.Fatalf("controller controls %d lands, want 2 (1 base + 1 reward for the empty-search accepter)", permanents[game.Player1])
	}
	if permanents[game.Player2] != 0 {
		t.Fatalf("Player2 controls %d lands, want 0 (accepted but found none)", permanents[game.Player2])
	}
	searches := librarySearchCounts(g)
	// Player2 searched (and shuffled) despite finding nothing.
	if searches[game.Player2] != 1 {
		t.Fatalf("empty-search accepter searched %d times, want 1 (still shuffled)", searches[game.Player2])
	}
	if searches[game.Player1] != 2 {
		t.Fatalf("controller searched %d times, want 2 (1 base + 1 reward)", searches[game.Player1])
	}
}
