package rules

import (
	"math/rand/v2"
	"slices"
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/id"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
)

func TestSearchLibraryTopShufflesBeforeReplacingFoundCard(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	seed1, seed2 := uint64(17), uint64(29)
	engine := NewEngine(rand.New(rand.NewPCG(seed1, seed2)))
	var cards []id.ID
	for _, name := range []string{"One", "Two", "Three", "Four", "Wanted"} {
		cards = append(cards, addCardToLibrary(g, game.Player1, &game.CardDef{
			CardFace: game.CardFace{Name: name},
		}))
	}
	wanted := cards[len(cards)-1]
	before := g.Players[game.Player1].Library.All()
	addEffectSpellToStack(g, game.Player1, game.Search{
		Amount: game.Fixed(1),
		Player: game.ControllerReference(),
		Spec: game.SearchSpec{
			SourceZone:          zone.Library,
			Destination:         zone.Library,
			DestinationPosition: game.SearchPositionTop,
		},
	}, nil)
	agents := [game.NumPlayers]PlayerAgent{game.Player1: &searchByNameAgent{wanted: "Wanted"}}

	engine.resolveTopOfStackWithChoices(g, agents, &TurnLog{})

	remaining := slices.DeleteFunc(slices.Clone(before), func(cardID id.ID) bool { return cardID == wanted })
	rand.New(rand.NewPCG(seed1, seed2)).Shuffle(len(remaining), func(i, j int) {
		remaining[i], remaining[j] = remaining[j], remaining[i]
	})
	want := append([]id.ID{wanted}, remaining...)
	if got := g.Players[game.Player1].Library.All(); !slices.Equal(got, want) {
		t.Fatalf("library = %v, want shuffle then selected card on top %v", got, want)
	}
}

func TestSearchLibraryTopTypeUnionRevealAndFailToFind(t *testing.T) {
	tests := []struct {
		name       string
		wanted     string
		reveal     bool
		wantTop    string
		wantReveal bool
	}{
		{name: "artifact revealed", wanted: "Relic", reveal: true, wantTop: "Relic", wantReveal: true},
		{name: "enchantment hidden", wanted: "Oath", wantTop: "Oath"},
		{name: "legal fail to find", wanted: ""},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
			engine := NewEngine(nil)
			relic := addCardToLibrary(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
				Name: "Relic", Types: []types.Card{types.Artifact},
			}})
			oath := addCardToLibrary(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
				Name: "Oath", Types: []types.Card{types.Enchantment},
			}})
			addCardToLibrary(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
				Name: "Bear", Types: []types.Card{types.Creature},
			}})
			addEffectSpellToStack(g, game.Player1, game.Search{
				Amount: game.Fixed(1),
				Player: game.ControllerReference(),
				Spec: game.SearchSpec{
					SourceZone:          zone.Library,
					Destination:         zone.Library,
					DestinationPosition: game.SearchPositionTop,
					CardTypesAny:        []types.Card{types.Artifact, types.Enchantment},
					Reveal:              test.reveal,
				},
			}, nil)
			agents := [game.NumPlayers]PlayerAgent{game.Player1: &searchByNameAgent{wanted: test.wanted}}

			engine.resolveTopOfStackWithChoices(g, agents, &TurnLog{})

			if test.wantTop != "" {
				topID, ok := g.Players[game.Player1].Library.Top()
				if !ok {
					t.Fatal("library unexpectedly empty")
				}
				top, _ := g.GetCardInstance(topID)
				if top.Def.Name != test.wantTop {
					t.Fatalf("top card = %q, want %q", top.Def.Name, test.wantTop)
				}
			} else if !g.Players[game.Player1].Library.Contains(relic) ||
				!g.Players[game.Player1].Library.Contains(oath) {
				t.Fatal("legal fail to find removed a matching card")
			}
			revealed := false
			for _, event := range g.Events {
				if event.Kind == game.EventCardRevealed && (event.CardID == relic || event.CardID == oath) {
					revealed = true
				}
			}
			if revealed != test.wantReveal {
				t.Fatalf("revealed = %v, want %v", revealed, test.wantReveal)
			}
		})
	}
}
