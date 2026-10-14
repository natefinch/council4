package rules

import (
	"slices"
	"testing"

	"github.com/natefinch/council4/mtg/cards/authored"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/zone"
)

func TestAuthoredSearchPartitionGraveyardOrderAfterReplacements(t *testing.T) {
	t.Parallel()
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := searchPartitionEngine(t, authored.Intuition)
	for _, name := range []string{"One", "Two", "Three"} {
		addCardToLibrary(g, game.Player1, vanillaCreature(name, 1, 1))
	}
	found := g.Players[game.Player1].Library.All()
	// The opponent's hand choice also arrives in the graveyard.
	g.ReplacementEffects = append(g.ReplacementEffects, game.ReplacementEffect{
		ID: g.IDGen.Next(), MatchEvent: game.EventZoneChanged, MatchFromZone: true, FromZone: zone.Library,
		MatchToZone: true, ToZone: zone.Hand, ReplaceToZone: zone.Graveyard,
	})
	ordered := false
	policies := [game.NumPlayers]PlayerAgent{
		game.Player1: &controlScript{choice: func(obs PlayerObservation, req game.ChoiceRequest) []int {
			if req.Kind == game.ChoiceSearch {
				return []int{0, 1, 2}
			}
			if req.Kind != game.ChoiceOrder || req.Player != game.Player1 ||
				len(req.Options) != 3 || req.MinChoices != 3 || req.MaxChoices != 3 ||
				len(obs.LibraryLookable(game.Player1)) != 0 {
				t.Fatal("owner did not order the actual post-replacement public arrivals")
			}
			ordered = true
			return []int{1, 2, 0}
		}},
		game.Player3: &controlScript{choice: func(PlayerObservation, game.ChoiceRequest) []int {
			return []int{0}
		}},
	}
	spell := addImplementationSpellToStack(g, game.Player1, authored.Intuition.New(), searchPartitionTargets(authored.Intuition))
	engine.resolveTopOfStackWithChoices(g, policies, &TurnLog{})
	if !ordered || !slices.Equal(g.Players[game.Player1].Graveyard.All(), []game.ObjectID{spell, found[1], found[2], found[0]}) {
		t.Fatalf("owner's graveyard ordering was not committed: %v", g.Players[game.Player1].Graveyard.All())
	}
	assertSearchPartitionBatch(t, g, found)
}
