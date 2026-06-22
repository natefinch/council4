package rules

import (
	"math/rand/v2"
	"slices"
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
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
					Reveal:              test.reveal,
					Filter: game.Selection{
						RequiredTypesAny: []types.Card{types.Artifact, types.Enchantment},
					},
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

// TestSearchLibraryColorFilterOffersOnlyMatchingColor verifies the ColorsAny
// search filter (Green Sun's Zenith family): a "green creature card" tutor finds
// a green creature but never an off-color creature or an off-type green card.
func TestSearchLibraryColorFilterOffersOnlyMatchingColor(t *testing.T) {
	greenCreatureSpec := game.SearchSpec{
		SourceZone:          zone.Library,
		Destination:         zone.Library,
		DestinationPosition: game.SearchPositionTop,
		Filter: game.Selection{
			RequiredTypes: []types.Card{types.Creature},
			ColorsAny:     []color.Color{color.Green},
		},
	}

	t.Run("finds the green creature, leaves off-filter cards", func(t *testing.T) {
		g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
		engine := NewEngine(nil)
		addCardToLibrary(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
			Name: "GreenBear", Types: []types.Card{types.Creature}, Colors: []color.Color{color.Green},
		}})
		whiteBear := addCardToLibrary(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
			Name: "WhiteBear", Types: []types.Card{types.Creature}, Colors: []color.Color{color.White},
		}})
		greenRelic := addCardToLibrary(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
			Name: "GreenRelic", Types: []types.Card{types.Artifact}, Colors: []color.Color{color.Green},
		}})
		addEffectSpellToStack(g, game.Player1, game.Search{
			Amount: game.Fixed(1),
			Player: game.ControllerReference(),
			Spec:   greenCreatureSpec,
		}, nil)
		agents := [game.NumPlayers]PlayerAgent{game.Player1: &searchByNameAgent{wanted: "GreenBear"}}

		engine.resolveTopOfStackWithChoices(g, agents, &TurnLog{})

		topID, ok := g.Players[game.Player1].Library.Top()
		if !ok {
			t.Fatal("library unexpectedly empty")
		}
		top, _ := g.GetCardInstance(topID)
		if top.Def.Name != "GreenBear" {
			t.Fatalf("top card = %q, want GreenBear", top.Def.Name)
		}
		if !g.Players[game.Player1].Library.Contains(whiteBear) ||
			!g.Players[game.Player1].Library.Contains(greenRelic) {
			t.Fatal("the off-color creature and off-type green card must remain in the library")
		}
	})

	t.Run("no green creature is a legal fail to find", func(t *testing.T) {
		g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
		engine := NewEngine(nil)
		whiteBear := addCardToLibrary(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
			Name: "WhiteBear", Types: []types.Card{types.Creature}, Colors: []color.Color{color.White},
		}})
		greenRelic := addCardToLibrary(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
			Name: "GreenRelic", Types: []types.Card{types.Artifact}, Colors: []color.Color{color.Green},
		}})
		addEffectSpellToStack(g, game.Player1, game.Search{
			Amount: game.Fixed(1),
			Player: game.ControllerReference(),
			Spec:   greenCreatureSpec,
		}, nil)
		agents := [game.NumPlayers]PlayerAgent{game.Player1: &searchByNameAgent{}}

		engine.resolveTopOfStackWithChoices(g, agents, &TurnLog{})

		if !g.Players[game.Player1].Library.Contains(whiteBear) ||
			!g.Players[game.Player1].Library.Contains(greenRelic) {
			t.Fatal("no green creature exists; the search must not remove a non-matching card")
		}
	})
}

func TestUnrestrictedExactSearchCannotDeclineNonemptyLibrary(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	addCardToLibrary(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Only Card"}})
	addEffectSpellToStack(g, game.Player1, game.Search{
		Amount: game.Fixed(1),
		Player: game.ControllerReference(),
		Spec: game.SearchSpec{
			SourceZone:          zone.Library,
			Destination:         zone.Library,
			DestinationPosition: game.SearchPositionTop,
		},
	}, nil)
	agents := [game.NumPlayers]PlayerAgent{game.Player1: &searchByNameAgent{}}
	log := &TurnLog{}

	engine.resolveTopOfStackWithChoices(g, agents, log)

	if len(log.Choices) != 1 {
		t.Fatalf("choices = %#v, want one required search choice", log.Choices)
	}
	choice := log.Choices[0]
	if choice.Request.MinChoices != 1 || len(choice.Selected) != 1 || !choice.UsedFallback {
		t.Fatalf("choice = %#v, want declined choice rejected and one card selected", choice)
	}
}

func TestUnrestrictedExactSearchAllowsEmptyLibrary(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	addEffectSpellToStack(g, game.Player1, game.Search{
		Amount: game.Fixed(1),
		Player: game.ControllerReference(),
		Spec: game.SearchSpec{
			SourceZone:          zone.Library,
			Destination:         zone.Library,
			DestinationPosition: game.SearchPositionTop,
		},
	}, nil)
	log := &TurnLog{}

	engine.resolveTopOfStackWithChoices(g, [game.NumPlayers]PlayerAgent{}, log)

	if len(log.Choices) != 0 {
		t.Fatalf("choices = %#v, want no choice for an empty library", log.Choices)
	}
	if _, ok := g.Players[game.Player1].Library.Top(); ok {
		t.Fatal("empty library search unexpectedly added a card")
	}
}

func TestUpToSearchStillAllowsFailToFind(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	first := addCardToLibrary(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name: "First", Types: []types.Card{types.Creature},
	}})
	second := addCardToLibrary(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name: "Second", Types: []types.Card{types.Creature},
	}})
	addEffectSpellToStack(g, game.Player1, game.Search{
		Amount: game.Fixed(2),
		Player: game.ControllerReference(),
		Spec: game.SearchSpec{
			SourceZone:  zone.Library,
			Destination: zone.Hand,
			Filter: game.Selection{
				RequiredTypes: []types.Card{types.Creature},
			},
		},
	}, nil)
	agents := [game.NumPlayers]PlayerAgent{game.Player1: &searchByNameAgent{}}
	log := &TurnLog{}

	engine.resolveTopOfStackWithChoices(g, agents, log)

	if !g.Players[game.Player1].Library.Contains(first) ||
		!g.Players[game.Player1].Library.Contains(second) {
		t.Fatal("up-to search did not allow legal fail-to-find")
	}
	if len(log.Choices) != 1 || log.Choices[0].Request.MinChoices != 0 {
		t.Fatalf("choices = %#v, want optional up-to search choice", log.Choices)
	}
}
