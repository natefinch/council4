package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// pathToExileSequence builds the Path to Exile rider body: exile the target,
// then the affected permanent's controller may fetch a basic land onto the
// battlefield tapped. The optional choice and the search both name the removal
// target's controller via ObjectControllerReference(TargetPermanentReference(0)),
// so the affected player — not the spell's controller — decides and searches.
func pathToExileSequence() []game.Instruction {
	searcher := game.ObjectControllerReference(game.TargetPermanentReference(0))
	return []game.Instruction{
		{Primitive: game.Exile{Object: game.TargetPermanentReference(0)}},
		{
			Optional:      true,
			OptionalActor: opt.Val(searcher),
			Primitive: game.Search{
				Amount: game.Fixed(1),
				Player: searcher,
				Spec: game.SearchSpec{
					SourceZone:   zone.Library,
					Destination:  zone.Battlefield,
					EntersTapped: true,
					Filter: game.Selection{
						RequiredTypes: []types.Card{types.Land},
						Supertypes:    []types.Super{types.Basic},
					},
				},
			},
		},
	}
}

func basicForestDef() *game.CardDef {
	return &game.CardDef{CardFace: game.CardFace{
		Name:       "Forest",
		Supertypes: []types.Super{types.Basic},
		Types:      []types.Card{types.Land},
		Subtypes:   []types.Sub{types.Forest},
	}}
}

// TestRemovalThenControllerSearchActorDeclines proves the decline branch of the
// "Its controller may search" rider: the spell's controller (Player1) exiles a
// creature controlled by Player2, and Player2 — the affected permanent's
// controller — declines the optional fetch, so the basic land stays in Player2's
// library and nothing enters the battlefield.
func TestRemovalThenControllerSearchActorDeclines(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	victim := addCreaturePermanent(g, game.Player2)
	forest := addCardToLibrary(g, game.Player2, basicForestDef())
	addInstructionSpellToStackForController(g, game.Player1, pathToExileSequence(),
		[]game.Target{game.PermanentTarget(victim.ObjectID)})
	agents := [game.NumPlayers]PlayerAgent{
		// Player1 (the spell's controller) would accept; the test proves the
		// choice is routed to Player2, the affected permanent's controller, whose
		// decline must win.
		game.Player1: optionalSearchAgent{accept: true, wanted: "Forest"},
		game.Player2: optionalSearchAgent{accept: false, wanted: "Forest"},
	}

	engine.resolveTopOfStackWithChoices(g, agents, &TurnLog{})

	if _, ok := permanentByObjectID(g, victim.ObjectID); ok {
		t.Fatal("targeted creature was not exiled")
	}
	if !g.Players[game.Player2].Library.Contains(forest) {
		t.Fatal("declining the controller search still moved the basic land out of the library")
	}
	for _, permanent := range g.Battlefield {
		if permanent.CardInstanceID == forest {
			t.Fatal("declining the controller search still put a land onto the battlefield")
		}
	}
}

// TestRemovalThenControllerSearchActorAccepts proves the accept branch: the
// affected permanent's controller (Player2) chooses to fetch, and the basic land
// leaves Player2's library and enters the battlefield tapped under Player2's
// control — even though the searcher reference is read from last-known
// information after the targeted creature has already been exiled.
func TestRemovalThenControllerSearchActorAccepts(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	victim := addCreaturePermanent(g, game.Player2)
	forest := addCardToLibrary(g, game.Player2, basicForestDef())
	addInstructionSpellToStackForController(g, game.Player1, pathToExileSequence(),
		[]game.Target{game.PermanentTarget(victim.ObjectID)})
	agents := [game.NumPlayers]PlayerAgent{
		// Player1 (the spell's controller) would decline; the test proves the
		// choice is routed to Player2, whose acceptance must drive the fetch.
		game.Player1: optionalSearchAgent{accept: false, wanted: "Forest"},
		game.Player2: optionalSearchAgent{accept: true, wanted: "Forest"},
	}

	engine.resolveTopOfStackWithChoices(g, agents, &TurnLog{})

	if g.Players[game.Player2].Library.Contains(forest) {
		t.Fatal("accepting the controller search left the basic land in the library")
	}
	var landPermanent *game.Permanent
	for _, permanent := range g.Battlefield {
		if permanent.CardInstanceID == forest {
			landPermanent = permanent
			break
		}
	}
	if landPermanent == nil {
		t.Fatal("accepting the controller search did not put the basic land onto the battlefield")
	}
	if landPermanent.Controller != game.Player2 {
		t.Fatalf("fetched land controller = %v, want Player2 (the affected player)", landPermanent.Controller)
	}
	if !landPermanent.Tapped {
		t.Fatal("fetched land entered untapped, want tapped")
	}
}
