package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/types"
)

// TestSacrificePermanentsAllColored proves the All Is Dust mass edict
// (SacrificePermanents with All set over the all-players group and a Colored
// selection): every colored permanent each player controls is sacrificed at
// once while colorless permanents and lands survive, with no per-player choice.
func TestSacrificePermanentsAllColored(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)

	coloredP1 := addColoredPermanent(g, game.Player1, "Green Creature", []color.Color{color.Green}, []types.Card{types.Creature}, nil)
	colorlessP1 := addColoredPermanent(g, game.Player1, "Colorless Creature", nil, []types.Card{types.Creature}, nil)
	landP1 := addColoredPermanent(g, game.Player1, "Test Land", nil, []types.Card{types.Land}, nil)
	coloredP2 := addColoredPermanent(g, game.Player2, "Multicolor Artifact", []color.Color{color.Blue, color.Red}, []types.Card{types.Artifact}, nil)
	colorlessP2 := addColoredPermanent(g, game.Player2, "Colorless Artifact", nil, []types.Card{types.Artifact}, nil)

	addEffectSpellToStack(g, game.Player1, game.SacrificePermanents{
		All:         true,
		PlayerGroup: game.AllPlayersReference(),
		Selection:   game.Selection{Colored: true},
	}, nil)

	log := TurnLog{}
	engine.resolveTopOfStackWithChoices(g, [game.NumPlayers]PlayerAgent{}, &log)

	for _, choice := range log.Choices {
		if choice.Request.Kind == game.ChoicePayment {
			t.Fatalf("mass sacrifice asked for a choice: %+v", choice.Request)
		}
	}

	for _, sacrificed := range []*game.Permanent{coloredP1, coloredP2} {
		if _, ok := permanentByObjectID(g, sacrificed.ObjectID); ok {
			t.Fatalf("colored permanent %v survived, want sacrificed", sacrificed.ObjectID)
		}
	}
	for _, survivor := range []*game.Permanent{colorlessP1, landP1, colorlessP2} {
		if _, ok := permanentByObjectID(g, survivor.ObjectID); !ok {
			t.Fatalf("colorless permanent %v was sacrificed, want survived", survivor.ObjectID)
		}
	}
}
