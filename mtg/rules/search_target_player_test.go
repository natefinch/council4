package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/id"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// fetchedLandPermanent returns the battlefield permanent created from cardID, or
// nil when the card never entered the battlefield.
func fetchedLandPermanent(g *game.Game, cardID id.ID) *game.Permanent {
	for _, permanent := range g.Battlefield {
		if permanent.CardInstanceID == cardID {
			return permanent
		}
	}
	return nil
}

// TestTargetPlayerSearchEntersUnderSearcherControl proves the Fertilid form: a
// spell controlled by Player1 makes a target player (Player2) search their own
// library for a basic land and put it onto the battlefield tapped. The searcher
// is the chosen target player, so the land leaves Player2's library and enters
// under Player2's control — never Player1's.
func TestTargetPlayerSearchEntersUnderSearcherControl(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	forest := addCardToLibrary(g, game.Player2, basicForestDef())
	addInstructionSpellToStackForController(g, game.Player1, []game.Instruction{{
		Primitive: game.Search{
			Amount: game.Fixed(1),
			Player: game.TargetPlayerReference(0),
			Spec: game.SearchSpec{
				SourceZone:   zone.Library,
				Destination:  zone.Battlefield,
				CardType:     opt.Val(types.Land),
				Supertype:    opt.Val(types.Basic),
				EntersTapped: true,
			},
		},
	}}, []game.Target{game.PlayerTarget(game.Player2)})
	agents := [game.NumPlayers]PlayerAgent{
		// The searcher (Player2, the target) chooses which matching card to take.
		game.Player2: optionalSearchAgent{accept: true, wanted: "Forest"},
	}

	engine.resolveTopOfStackWithChoices(g, agents, &TurnLog{})

	if g.Players[game.Player2].Library.Contains(forest) {
		t.Fatal("target-player search left the basic land in the library")
	}
	landPermanent := fetchedLandPermanent(g, forest)
	if landPermanent == nil {
		t.Fatal("target-player search did not put the basic land onto the battlefield")
	}
	if landPermanent.Controller != game.Player2 {
		t.Fatalf("fetched land controller = %v, want Player2 (the searching target player)", landPermanent.Controller)
	}
	if !landPermanent.Tapped {
		t.Fatal("fetched land entered untapped, want tapped")
	}
}

// TestControllerSearchEntersUnderTargetControl proves the Yavimaya Dryad form:
// Player1 searches their own library for a Forest, but the found permanent enters
// under a chosen target player's (Player2's) control. The card leaves Player1's
// library — the searcher — yet enters tapped under Player2's control.
func TestControllerSearchEntersUnderTargetControl(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	forest := addCardToLibrary(g, game.Player1, basicForestDef())
	addInstructionSpellToStackForController(g, game.Player1, []game.Instruction{{
		Primitive: game.Search{
			Amount:     game.Fixed(1),
			Player:     game.ControllerReference(),
			Controller: opt.Val(game.TargetPlayerReference(0)),
			Spec: game.SearchSpec{
				SourceZone:   zone.Library,
				Destination:  zone.Battlefield,
				SubtypesAny:  []types.Sub{types.Forest},
				EntersTapped: true,
			},
		},
	}}, []game.Target{game.PlayerTarget(game.Player2)})
	agents := [game.NumPlayers]PlayerAgent{
		// The searcher is Player1 (the controller), so Player1 chooses the card.
		game.Player1: optionalSearchAgent{accept: true, wanted: "Forest"},
	}

	engine.resolveTopOfStackWithChoices(g, agents, &TurnLog{})

	if g.Players[game.Player1].Library.Contains(forest) {
		t.Fatal("controller search left the Forest in the searcher's library")
	}
	landPermanent := fetchedLandPermanent(g, forest)
	if landPermanent == nil {
		t.Fatal("controller search did not put the Forest onto the battlefield")
	}
	if landPermanent.Controller != game.Player2 {
		t.Fatalf("fetched land controller = %v, want Player2 (the named target player)", landPermanent.Controller)
	}
	if !landPermanent.Tapped {
		t.Fatal("fetched land entered untapped, want tapped")
	}
}
