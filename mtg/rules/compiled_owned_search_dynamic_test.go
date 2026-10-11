package rules

import (
	"fmt"
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/id"
	"github.com/natefinch/council4/mtg/game/types"
)

const compiledOwnedDynamicSearch = "Search your library for up to X basic land cards, where X is the number of lands you control, put them onto the battlefield tapped, then shuffle."

func TestCompiledOwnedDynamicSearchActualCounts(t *testing.T) {
	t.Parallel()
	for _, prefix := range []string{"", "You gain 1 life. "} {
		for amount := range 3 {
			t.Run(fmt.Sprintf("prefix=%q/count=%d", prefix, amount), func(t *testing.T) {
				t.Parallel()
				mode := compileOwnedSearchMode(t, prefix+compiledOwnedDynamicSearch)
				g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
				for range amount {
					addBasicLandPermanent(g, game.Player1, types.Island)
				}
				for range 3 {
					addCardToLibrary(g, game.Player1, basicNamedLand("Forest"))
				}
				addInstructionSpellToStack(g, mode.Sequence)
				log := TurnLog{}
				NewEngine(nil).resolveTopOfStackWithChoices(g, [game.NumPlayers]PlayerAgent{}, &log)
				if g.Players[game.Player1].Library.Size() != 3-amount {
					t.Fatalf("searched=%d, want current rules-derived count %d", 3-g.Players[game.Player1].Library.Size(), amount)
				}
				for _, choice := range log.Choices {
					if choice.Request.Kind == game.ChoiceSearch && choice.Request.MaxChoices != amount {
						t.Fatalf("choice maximum=%d, want %d", choice.Request.MaxChoices, amount)
					}
				}
				searches := 0
				for _, event := range g.Events {
					if event.Kind == game.EventLibrarySearched {
						searches++
					}
				}
				if searches != 1 || amount == 0 && len(log.Choices) != 0 {
					t.Fatal("zero-limit Search lost its performed event or made a choice")
				}
			})
		}
	}
}

func TestCompiledOwnedDynamicSearchAfterStateChange(t *testing.T) {
	t.Parallel()
	for amount := range 3 {
		t.Run(fmt.Sprintf("after-departure=%d", amount), func(t *testing.T) {
			t.Parallel()
			mode := compileOwnedSearchMode(t, "Exile target land. "+compiledOwnedDynamicSearch+" You gain 1 life.")
			g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
			target := addBasicLandPermanent(g, game.Player1, types.Island)
			for range amount {
				addBasicLandPermanent(g, game.Player1, types.Island)
			}
			for range 3 {
				addBasicLandPermanent(g, game.Player2, types.Swamp)
			}
			var candidates []id.ID
			for range 3 {
				cardID := addCardToLibrary(g, game.Player1, basicNamedLand("Forest"))
				g.CardInstances[cardID].Owner = game.Player4
				candidates = append(candidates, cardID)
			}
			source := addInstructionSpellToStackForController(g, game.Player1, mode.Sequence,
				[]game.Target{game.PermanentTarget(target.ObjectID)})
			g.CardInstances[source].Owner = game.Player3
			clone := g.Clone()
			for _, state := range []*game.Game{g, clone} {
				NewEngine(nil).resolveTopOfStackWithChoices(state, [game.NumPlayers]PlayerAgent{}, &TurnLog{})
				found := 0
				for _, cardID := range candidates {
					if permanent := permanentForCard(state, cardID); permanent != nil {
						found++
						if permanent.Owner != game.Player4 || permanent.Controller != game.Player1 || !permanent.Tapped {
							t.Fatal("dynamic count confused library/physical ownership or entry controller")
						}
					}
				}
				if found != amount || state.Players[game.Player1].Life != 41 {
					t.Fatalf("found=%d, want current post-departure count%d and independent rider", found, amount)
				}
			}
		})
	}
}
