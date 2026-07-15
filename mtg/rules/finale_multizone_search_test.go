package rules

import (
	"slices"
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/id"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// finaleSearchShuffleInstructions builds the first two instructions the generated
// Finale of Devastation produces: the folded multi-zone battlefield search that
// records whether the library was searched, and the ShuffleLibrary gated on that
// recorded result. It mirrors the generated card field-for-field so these runtime
// tests exercise the exact instruction shape the cardgen backend emits.
func finaleSearchShuffleInstructions() []game.Instruction {
	return []game.Instruction{
		{
			Primitive: game.Search{
				Player: game.ControllerReference(),
				Spec: game.SearchSpec{
					SourceZone:         zone.Library,
					Destination:        zone.Battlefield,
					Filter:             game.Selection{RequiredTypes: []types.Card{types.Creature}},
					MaxManaValueFromX:  true,
					AlsoGraveyard:      true,
					ConditionalShuffle: true,
				},
				Amount: game.Fixed(1),
			},
			PublishResult: game.ResultKey("multizone-search-library"),
		},
		{
			Primitive: game.ShuffleLibrary{Player: game.ControllerReference()},
			ResultGate: opt.Val(game.InstructionResultGate{
				Key:             game.ResultKey("multizone-search-library"),
				SearchedLibrary: game.TriTrue,
			}),
		},
	}
}

// finaleZoneSearchAgent answers the multi-zone search's two prompts: the zone
// selection (which of the library and graveyard to search) and the match
// selection (which found card to take, by name). An empty wanted declines to
// find, which the hidden-zone rules permit only when the library was searched.
type finaleZoneSearchAgent struct {
	zones  []int // 0 = library, 1 = graveyard
	wanted string
}

func (finaleZoneSearchAgent) ChooseAction(PlayerObservation, []action.Action) action.Action {
	return action.Pass()
}

func (a finaleZoneSearchAgent) ChooseChoice(_ PlayerObservation, request game.ChoiceRequest) []int {
	switch request.Kind {
	case game.ChoiceZoneSelection:
		return a.zones
	case game.ChoiceSearch:
		if a.wanted == "" {
			return []int{}
		}
		for _, option := range request.Options {
			if option.Label == a.wanted {
				return []int{option.Index}
			}
		}
		return nil
	default:
		return nil
	}
}

// finaleCreatureDef builds a creature card with the given name and mana value, so
// the search's "mana value X or less" bound and creature filter can be exercised.
func finaleCreatureDef(name string, manaValue int) *game.CardDef {
	manaCost := cost.Mana{}
	if manaValue > 0 {
		manaCost = cost.Mana{cost.O(manaValue)}
	}
	return &game.CardDef{CardFace: game.CardFace{
		Name:      name,
		ManaCost:  opt.Val(manaCost),
		Types:     []types.Card{types.Creature},
		Power:     opt.Val(game.PT{Value: 1}),
		Toughness: opt.Val(game.PT{Value: 1}),
	}}
}

// finaleLandDef builds a non-creature land, used as library filler that the
// creature search never matches so the library stays untouched by card removal
// while still being searched (isolating the conditional shuffle).
func finaleLandDef(name string) *game.CardDef {
	return &game.CardDef{CardFace: game.CardFace{Name: name, Types: []types.Card{types.Land}}}
}

func libraryOrder(g *game.Game, playerID game.PlayerID) []id.ID {
	return slices.Clone(g.Players[playerID].Library.All())
}

// castFinaleSearch pushes the Finale search+shuffle spell for Player1, sets its
// chosen X, and resolves it with the given zone/match agent.
func castFinaleSearch(t *testing.T, g *game.Game, x int, agent finaleZoneSearchAgent) {
	t.Helper()
	addInstructionSpellToStack(g, finaleSearchShuffleInstructions())
	obj, ok := g.Stack.Peek()
	if !ok {
		t.Fatal("expected the Finale spell on the stack")
	}
	obj.XValue = x
	var agents [game.NumPlayers]PlayerAgent
	agents[game.Player1] = agent
	NewEngine(nil).resolveTopOfStackWithChoices(g, agents, &TurnLog{})
}

// TestFinaleSearchLibraryOnlyFindsCreatureOntoBattlefield proves the library-only
// zone choice finds a creature within the X bound and puts it onto the
// battlefield, leaving the library.
func TestFinaleSearchLibraryOnlyFindsCreatureOntoBattlefield(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	target := addCardToLibrary(g, game.Player1, finaleCreatureDef("Library Bear", 2))
	addCardToLibrary(g, game.Player1, finaleLandDef("Filler Land"))

	castFinaleSearch(t, g, 5, finaleZoneSearchAgent{zones: []int{0}, wanted: "Library Bear"})

	if permanentForCard(g, target) == nil {
		t.Fatal("library creature did not enter the battlefield")
	}
	if g.Players[game.Player1].Library.Contains(target) {
		t.Fatal("found creature was left in the library")
	}
}

// TestFinaleSearchGraveyardOnlyFindsAndDoesNotShuffle proves the graveyard-only
// zone choice finds from the graveyard, puts the creature onto the battlefield,
// and does NOT shuffle the library (the library was never searched).
func TestFinaleSearchGraveyardOnlyFindsAndDoesNotShuffle(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	for i := range 6 {
		addCardToLibrary(g, game.Player1, finaleLandDef("Filler "+string(rune('A'+i))))
	}
	target := addCardToGraveyard(g, game.Player1, finaleCreatureDef("Grave Bear", 2))
	before := libraryOrder(g, game.Player1)

	castFinaleSearch(t, g, 5, finaleZoneSearchAgent{zones: []int{1}, wanted: "Grave Bear"})

	if permanentForCard(g, target) == nil {
		t.Fatal("graveyard creature did not enter the battlefield")
	}
	if g.Players[game.Player1].Graveyard.Contains(target) {
		t.Fatal("found creature was left in the graveyard")
	}
	if got := libraryOrder(g, game.Player1); !slices.Equal(got, before) {
		t.Fatalf("library order = %v, want unshuffled %v (graveyard-only must not shuffle)", got, before)
	}
}

// TestFinaleSearchBothChooseLibrarySource proves that when both zones are
// searched and the player takes the library creature, it enters the battlefield
// while the graveyard creature is left untouched.
func TestFinaleSearchBothChooseLibrarySource(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	libTarget := addCardToLibrary(g, game.Player1, finaleCreatureDef("Library Bear", 2))
	gyTarget := addCardToGraveyard(g, game.Player1, finaleCreatureDef("Grave Bear", 2))

	castFinaleSearch(t, g, 5, finaleZoneSearchAgent{zones: []int{0, 1}, wanted: "Library Bear"})

	if permanentForCard(g, libTarget) == nil {
		t.Fatal("chosen library creature did not enter the battlefield")
	}
	if g.Players[game.Player1].Library.Contains(libTarget) {
		t.Fatal("chosen library creature was left in the library")
	}
	if !g.Players[game.Player1].Graveyard.Contains(gyTarget) {
		t.Fatal("unchosen graveyard creature was incorrectly removed")
	}
}

// TestFinaleSearchBothChooseGraveyardSourceStillShuffles proves the "If you
// search your library this way, shuffle." step fires whenever the library was
// among the searched zones, even when the found card is taken from the graveyard.
func TestFinaleSearchBothChooseGraveyardSourceStillShuffles(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	for i := range 6 {
		addCardToLibrary(g, game.Player1, finaleLandDef("Filler "+string(rune('A'+i))))
	}
	gyTarget := addCardToGraveyard(g, game.Player1, finaleCreatureDef("Grave Bear", 2))
	before := libraryOrder(g, game.Player1)

	castFinaleSearch(t, g, 5, finaleZoneSearchAgent{zones: []int{0, 1}, wanted: "Grave Bear"})

	if permanentForCard(g, gyTarget) == nil {
		t.Fatal("chosen graveyard creature did not enter the battlefield")
	}
	if got := libraryOrder(g, game.Player1); slices.Equal(got, before) {
		t.Fatalf("library order = %v, want shuffled (library was searched)", got)
	}
}

// TestFinaleSearchBothNoFindStillShuffles proves the shuffle fires when the
// library was searched even if the player finds no card at all.
func TestFinaleSearchBothNoFindStillShuffles(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	for i := range 6 {
		addCardToLibrary(g, game.Player1, finaleLandDef("Filler "+string(rune('A'+i))))
	}
	gyTarget := addCardToGraveyard(g, game.Player1, finaleCreatureDef("Grave Bear", 2))
	before := libraryOrder(g, game.Player1)

	castFinaleSearch(t, g, 5, finaleZoneSearchAgent{zones: []int{0, 1}, wanted: ""})

	if !g.Players[game.Player1].Graveyard.Contains(gyTarget) {
		t.Fatal("declining to find still removed the graveyard creature")
	}
	if got := libraryOrder(g, game.Player1); slices.Equal(got, before) {
		t.Fatalf("library order = %v, want shuffled (library was searched despite no find)", got)
	}
}

// TestFinaleSearchHiddenZoneMayFailToFind proves searching the hidden library
// permits declining to find even when a legal creature is present (CR 701.19e).
func TestFinaleSearchHiddenZoneMayFailToFind(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	target := addCardToLibrary(g, game.Player1, finaleCreatureDef("Library Bear", 2))

	castFinaleSearch(t, g, 5, finaleZoneSearchAgent{zones: []int{0}, wanted: ""})

	if permanentForCard(g, target) != nil {
		t.Fatal("declining a hidden-zone search still put a creature onto the battlefield")
	}
	if !g.Players[game.Player1].Library.Contains(target) {
		t.Fatal("declining a hidden-zone search still removed the creature from the library")
	}
}

// TestFinaleSearchGraveyardOnlyMustFind proves searching only the public
// graveyard must find a legal creature when one exists: an attempt to decline
// falls back to finding, honoring public-zone legality.
func TestFinaleSearchGraveyardOnlyMustFind(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	target := addCardToGraveyard(g, game.Player1, finaleCreatureDef("Grave Bear", 2))

	castFinaleSearch(t, g, 5, finaleZoneSearchAgent{zones: []int{1}, wanted: ""})

	if permanentForCard(g, target) == nil {
		t.Fatal("a public graveyard-only search failed to find a legal creature")
	}
	if g.Players[game.Player1].Graveyard.Contains(target) {
		t.Fatal("the found creature was left in the graveyard")
	}
}

// TestFinaleSearchManaValueBoundFromX proves the "mana value X or less" bound
// resolves from the resolving spell's chosen X: with X=3 a mana-value-2 creature
// is findable but a mana-value-8 creature is not.
func TestFinaleSearchManaValueBoundFromX(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	cheap := addCardToLibrary(g, game.Player1, finaleCreatureDef("Small Bear", 2))
	pricey := addCardToLibrary(g, game.Player1, finaleCreatureDef("Giant Bear", 8))

	castFinaleSearch(t, g, 3, finaleZoneSearchAgent{zones: []int{0}, wanted: "Giant Bear"})

	if permanentForCard(g, pricey) != nil {
		t.Fatal("a creature above the X mana-value bound was incorrectly findable")
	}
	if !g.Players[game.Player1].Library.Contains(pricey) {
		t.Fatal("the out-of-bound creature left the library")
	}
	// The in-bound creature remains available; a second cast can take it.
	castFinaleSearch(t, g, 3, finaleZoneSearchAgent{zones: []int{0}, wanted: "Small Bear"})
	if permanentForCard(g, cheap) == nil {
		t.Fatal("a creature within the X mana-value bound was not findable")
	}
}

// TestFinaleSearchManaValueBoundXZero proves the dynamic bound holds at X=0: only
// a mana-value-0 creature is findable.
func TestFinaleSearchManaValueBoundXZero(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	free := addCardToLibrary(g, game.Player1, finaleCreatureDef("Free Bear", 0))
	oneDrop := addCardToLibrary(g, game.Player1, finaleCreatureDef("One Bear", 1))

	castFinaleSearch(t, g, 0, finaleZoneSearchAgent{zones: []int{0}, wanted: "One Bear"})
	if permanentForCard(g, oneDrop) != nil {
		t.Fatal("a mana-value-1 creature was findable at X=0")
	}

	castFinaleSearch(t, g, 0, finaleZoneSearchAgent{zones: []int{0}, wanted: "Free Bear"})
	if permanentForCard(g, free) == nil {
		t.Fatal("a mana-value-0 creature was not findable at X=0")
	}
}

// TestFinaleSearchExactlyOneCard proves the search finds exactly one creature even
// when multiple match, leaving the rest in the library (Amount is Fixed(1)).
func TestFinaleSearchExactlyOneCard(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	first := addCardToLibrary(g, game.Player1, finaleCreatureDef("First Bear", 2))
	second := addCardToLibrary(g, game.Player1, finaleCreatureDef("Second Bear", 2))

	castFinaleSearch(t, g, 5, finaleZoneSearchAgent{zones: []int{0}, wanted: "First Bear"})

	if permanentForCard(g, first) == nil {
		t.Fatal("the chosen creature did not enter the battlefield")
	}
	if !g.Players[game.Player1].Library.Contains(second) {
		t.Fatal("a second matching creature was incorrectly found by a one-card search")
	}
}

// TestFinaleSearchResolutionsAreIsolated proves the searched-library result is
// scoped to each resolution: a first graveyard-only cast does not shuffle, and a
// later library cast shuffles, with neither leaking the other's result.
func TestFinaleSearchResolutionsAreIsolated(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	for i := range 6 {
		addCardToLibrary(g, game.Player1, finaleLandDef("Filler "+string(rune('A'+i))))
	}
	graveFirst := addCardToGraveyard(g, game.Player1, finaleCreatureDef("Grave Bear", 2))
	beforeFirst := libraryOrder(g, game.Player1)

	// First resolution: graveyard-only, must not shuffle.
	castFinaleSearch(t, g, 5, finaleZoneSearchAgent{zones: []int{1}, wanted: "Grave Bear"})
	if permanentForCard(g, graveFirst) == nil {
		t.Fatal("first cast did not find the graveyard creature")
	}
	if got := libraryOrder(g, game.Player1); !slices.Equal(got, beforeFirst) {
		t.Fatalf("first (graveyard-only) cast shuffled the library: %v", got)
	}

	// Second resolution: library searched, must shuffle, independent of the first.
	beforeSecond := libraryOrder(g, game.Player1)
	castFinaleSearch(t, g, 5, finaleZoneSearchAgent{zones: []int{0}, wanted: ""})
	if got := libraryOrder(g, game.Player1); slices.Equal(got, beforeSecond) {
		t.Fatalf("second (library) cast did not shuffle the library: %v", got)
	}
}
