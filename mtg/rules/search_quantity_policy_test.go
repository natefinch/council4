package rules

import (
	"math/rand/v2"
	"reflect"
	"testing"

	"github.com/natefinch/council4/mtg/game"
)

func TestSearchQuantityLimitPolicy(t *testing.T) {
	t.Parallel()
	for _, tt := range []struct {
		name      string
		quantity  game.Quantity
		limit     int
		anyNumber bool
	}{
		{"dynamic zero", game.Dynamic(game.DynamicAmount{Kind: game.DynamicAmountX}), 0, false},
		{"dynamic negative", game.Dynamic(game.DynamicAmount{Kind: game.DynamicAmountX, Addend: -1}), 0, false},
		{"dynamic one", game.Dynamic(game.DynamicAmount{Kind: game.DynamicAmountX, Addend: 1}), 1, false},
		{"dynamic two", game.Dynamic(game.DynamicAmount{Kind: game.DynamicAmountX, Addend: 2}), 2, false},
		{"legacy fixed zero", game.Fixed(0), 1, false},
		{"legacy fixed negative", game.Fixed(-1), 1, false},
		{"any number", game.Fixed(0), 8, true},
	} {
		for _, group := range []bool{false, true} {
			t.Run(tt.name+map[bool]string{false: "/single", true: "/group"}[group], func(t *testing.T) {
				t.Parallel()
				g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
				engine := NewEngine(nil)
				engine.rng = rand.New(rand.NewPCG(1, 2))
				for player := range game.NumPlayers {
					for range 8 {
						addCardToLibrary(g, game.PlayerID(player), basicNamedLand("Forest"))
					}
				}
				expected := g.Clone()
				expected.Players[game.Player1].Library.Shuffle(rand.New(rand.NewPCG(1, 2)))
				mode := compileOwnedSearchMode(t, compiledOwnedDynamicSearch)
				search, ok := mode.Sequence[0].Primitive.(game.Search)
				if !ok {
					t.Fatalf("primitive=%T, want Search", mode.Sequence[0].Primitive)
				}
				search.Amount, search.Spec.AnyNumber = tt.quantity, tt.anyNumber
				if group {
					search.Player = game.PlayerReference{}
					search.PlayerGroup = game.AllPlayersReference()
				}
				obj := &game.StackObject{ID: g.IDGen.Next(), Controller: game.Player1}
				log := TurnLog{}
				r := &effectResolver{engine: engine, game: g, obj: obj, log: &log}
				resolved := r.quantity(tt.quantity)
				result := handleSearch(r, search)
				if result.amount != resolved || result.succeeded != (tt.limit > 0) {
					t.Fatal("search limit normalization altered the original scalar/action receipt")
				}
				events := 0
				for _, event := range g.Events {
					if event.Kind == game.EventLibrarySearched {
						events++
					}
				}
				wantEvents := 1
				if group {
					wantEvents = game.NumPlayers
				}
				if events != wantEvents {
					t.Fatalf("search events=%d, want %d", events, wantEvents)
				}
				for player := range game.NumPlayers {
					limit := 0
					if group || game.PlayerID(player) == game.Player1 {
						limit = tt.limit
					}
					if g.Players[player].Library.Size() != 8-limit {
						t.Fatalf("player%d found%d, want limit%d", player, 8-g.Players[player].Library.Size(), limit)
					}
				}
				if tt.limit == 0 {
					if len(log.Choices) != 0 {
						t.Fatal("zero search limit made a choice")
					}
					if !group && !reflect.DeepEqual(g.Players[game.Player1].Library.All(), expected.Players[game.Player1].Library.All()) {
						t.Fatal("zero search limit did not perform the normal shuffle")
					}
				}
			})
		}
	}
}

func TestSearchUnavailableQuantityDoesNotBecomeKnownZero(t *testing.T) {
	t.Parallel()
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	for range 3 {
		addCardToLibrary(g, game.Player1, basicNamedLand("Forest"))
	}
	mode := compileOwnedSearchMode(t, compiledOwnedDynamicSearch)
	search, ok := mode.Sequence[0].Primitive.(game.Search)
	if !ok {
		t.Fatalf("primitive=%T, want Search", mode.Sequence[0].Primitive)
	}
	search.Amount = game.Dynamic(game.DynamicAmount{
		Kind: game.DynamicAmountPreviousEffectResult, ResultKey: "unavailable", Selection: &game.Selection{},
	})
	obj := &game.StackObject{ID: g.IDGen.Next(), Controller: game.Player1}
	NewEngine(nil).resolveInstructionSequence(g, obj, []game.Instruction{{Primitive: search}}, [game.NumPlayers]PlayerAgent{}, &TurnLog{})
	for _, event := range g.Events {
		if event.Kind == game.EventLibrarySearched {
			t.Fatal("an unavailable quantity became an executed zero Search")
		}
	}
	if g.Players[game.Player1].Library.Size() != 3 {
		t.Fatal("an unavailable actual-card quantity searched or moved cards")
	}
}

func TestDynamicZeroSearchClearsOnlyItsOwnArrival(t *testing.T) {
	t.Parallel()
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	forest := addCardToLibrary(g, game.Player1, basicNamedLand("Forest"))
	second := addCardToLibrary(g, game.Player1, basicNamedLand("Forest"))
	obj := &game.StackObject{ID: g.IDGen.Next(), Controller: game.Player1, SourceID: g.IDGen.Next()}
	mode := compileOwnedSearchMode(t, compiledOwnedDynamicSearch)
	search, ok := mode.Sequence[0].Primitive.(game.Search)
	if !ok {
		t.Fatalf("primitive=%T, want Search", mode.Sequence[0].Primitive)
	}
	search.Amount, search.PublishLinked = game.Fixed(1), "search-arrival"
	engine := NewEngine(nil)
	log := TurnLog{}
	r := &effectResolver{engine: engine, game: g, obj: obj, log: &log}
	handleSearch(r, search)
	found := permanentForCard(g, forest)
	if found == nil {
		found = permanentForCard(g, second)
	}
	if found == nil {
		t.Fatal("missing initial publication")
	}
	found.Tapped = true
	rememberLinkedObject(g, linkedObjectSourceKey(g, obj, "independent"), permanentLinkedObjectRef(found))
	search.Amount = game.Dynamic(game.DynamicAmount{Kind: game.DynamicAmountX})
	engine.resolveInstructionSequence(g, obj, []game.Instruction{
		{Primitive: search},
		{Primitive: game.Untap{Object: game.LinkedObjectReference("search-arrival")}},
	}, [game.NumPlayers]PlayerAgent{}, &log)
	if !found.Tapped || g.Players[game.Player1].Library.Size() != 1 {
		t.Fatal("zero-limit Search reused an old arrived permanent or selected another card")
	}
	engine.resolveInstructionSequence(g, obj, []game.Instruction{
		{Primitive: game.Untap{Object: game.LinkedObjectReference("independent")}},
	}, [game.NumPlayers]PlayerAgent{}, &log)
	if found.Tapped {
		t.Fatal("zero-limit Search cleared an independent publication")
	}
}
