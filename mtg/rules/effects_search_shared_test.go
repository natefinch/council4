package rules

import (
	"slices"
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/id"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// correlatedSearchAgent answers the staged choices of a "that share a land type"
// search. Each stage offers only cards that can still share a land subtype with
// those already chosen, so the agent simply picks the next wanted card present in
// the offered pool, or declines (choosing none) to stop the search. It is
// stateful so it does not re-pick a card across stages.
type correlatedSearchAgent struct {
	want   []string
	picked map[string]bool
}

func newCorrelatedSearchAgent(want ...string) *correlatedSearchAgent {
	return &correlatedSearchAgent{want: want, picked: map[string]bool{}}
}

func (*correlatedSearchAgent) ChooseAction(PlayerObservation, []action.Action) action.Action {
	return action.Pass()
}

func (a *correlatedSearchAgent) ChooseChoice(_ PlayerObservation, request game.ChoiceRequest) []int {
	for _, option := range request.Options {
		if a.picked[option.Label] {
			continue
		}
		if slices.Contains(a.want, option.Label) {
			a.picked[option.Label] = true
			return []int{option.Index}
		}
	}
	return []int{}
}

// myriadLandscapeSearchSpec is the typed search Myriad Landscape's activated
// ability lowers to: up to two basic land cards that share a land subtype, both
// onto the battlefield tapped, then shuffle.
func myriadLandscapeSearchSpec() game.SearchSpec {
	return game.SearchSpec{
		SourceZone:    zone.Library,
		Destination:   zone.Battlefield,
		CardType:      opt.Val(types.Land),
		Supertype:     opt.Val(types.Basic),
		EntersTapped:  true,
		SharedSubtype: true,
	}
}

func basicLandWithSubtypes(name string, subs ...types.Sub) *game.CardDef {
	return &game.CardDef{CardFace: game.CardFace{
		Name:       name,
		Supertypes: []types.Super{types.Basic},
		Types:      []types.Card{types.Land},
		Subtypes:   subs,
	}}
}

// TestSharedSubtypeSearchTwoSharingCardsEnterTapped verifies a valid pair: two
// basic lands that share a land subtype both enter the battlefield tapped, leave
// the library, and the library is shuffled afterward.
func TestSharedSubtypeSearchTwoSharingCardsEnterTapped(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	forestA := addCardToLibrary(g, game.Player1, basicLandWithSubtypes("Forest A", types.Forest))
	forestB := addCardToLibrary(g, game.Player1, basicLandWithSubtypes("Forest B", types.Forest))
	var filler []id.ID
	for _, name := range []string{"F1", "F2", "F3", "F4", "F5", "F6"} {
		filler = append(filler, addCardToLibrary(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: name, Types: []types.Card{types.Creature}}}))
	}
	before := g.Players[game.Player1].Library.All()
	addEffectSpellToStack(g, game.Player1, game.Search{
		Amount: game.Fixed(2),
		Player: game.ControllerReference(),
		Spec:   myriadLandscapeSearchSpec(),
	}, nil)
	agents := [game.NumPlayers]PlayerAgent{game.Player1: newCorrelatedSearchAgent("Forest A", "Forest B")}

	engine.resolveTopOfStackWithChoices(g, agents, &TurnLog{})

	for _, want := range []id.ID{forestA, forestB} {
		permanent := permanentForCard(g, want)
		if permanent == nil {
			t.Fatalf("shared-subtype land %v did not enter the battlefield", want)
		}
		if !permanent.Tapped {
			t.Fatalf("shared-subtype land %v entered untapped, want tapped", want)
		}
		if g.Players[game.Player1].Library.Contains(want) {
			t.Fatalf("found land %v was not removed from the library", want)
		}
	}
	if len(filler) == 0 || !slices.Contains(before, filler[0]) {
		t.Fatal("test setup: filler not in starting library")
	}
	after := g.Players[game.Player1].Library.All()
	if slices.Equal(before[2:], after) {
		t.Fatal("library order unchanged; correlated search did not shuffle")
	}
}

// TestSharedSubtypeSearchDualBasicIntersection verifies the correlation honors
// dual basics carrying more than one land subtype: two lands whose only common
// subtype is Island still form a legal pair and both enter the battlefield.
func TestSharedSubtypeSearchDualBasicIntersection(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	forestIsland := addCardToLibrary(g, game.Player1, basicLandWithSubtypes("Forest Island", types.Forest, types.Island))
	plainsIsland := addCardToLibrary(g, game.Player1, basicLandWithSubtypes("Plains Island", types.Plains, types.Island))
	addEffectSpellToStack(g, game.Player1, game.Search{
		Amount: game.Fixed(2),
		Player: game.ControllerReference(),
		Spec:   myriadLandscapeSearchSpec(),
	}, nil)
	agents := [game.NumPlayers]PlayerAgent{game.Player1: newCorrelatedSearchAgent("Forest Island", "Plains Island")}

	engine.resolveTopOfStackWithChoices(g, agents, &TurnLog{})

	for _, want := range []id.ID{forestIsland, plainsIsland} {
		permanent := permanentForCard(g, want)
		if permanent == nil {
			t.Fatalf("dual-basic land %v sharing only Island did not enter the battlefield", want)
		}
		if !permanent.Tapped {
			t.Fatalf("dual-basic land %v entered untapped, want tapped", want)
		}
	}
}

// TestSharedSubtypeSearchInvalidPairPrevented verifies the staged choice prevents
// an illegal pair: when the only other candidate shares no land subtype with the
// first chosen card, it is never offered, so the search finds one card rather
// than choosing two and silently dropping one.
func TestSharedSubtypeSearchInvalidPairPrevented(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	forest := addCardToLibrary(g, game.Player1, basicLandWithSubtypes("Forest", types.Forest))
	island := addCardToLibrary(g, game.Player1, basicLandWithSubtypes("Island", types.Island))
	addEffectSpellToStack(g, game.Player1, game.Search{
		Amount: game.Fixed(2),
		Player: game.ControllerReference(),
		Spec:   myriadLandscapeSearchSpec(),
	}, nil)
	// The agent wants both lands, but they share no land subtype, so only the
	// first-offered card can be found.
	agents := [game.NumPlayers]PlayerAgent{game.Player1: newCorrelatedSearchAgent("Forest", "Island")}

	engine.resolveTopOfStackWithChoices(g, agents, &TurnLog{})

	forestEntered := permanentForCard(g, forest) != nil
	islandEntered := permanentForCard(g, island) != nil
	if forestEntered && islandEntered {
		t.Fatal("an illegal non-sharing pair was put onto the battlefield")
	}
	if !forestEntered && !islandEntered {
		t.Fatal("the first chosen land did not enter the battlefield")
	}
	// Exactly one card is found; the non-sharing card is never offered for the
	// second pick, so it stays in the library rather than being chosen and
	// silently dropped.
	if forestEntered && !g.Players[game.Player1].Library.Contains(island) {
		t.Fatal("the non-sharing second land was found instead of being prevented")
	}
	if islandEntered && !g.Players[game.Player1].Library.Contains(forest) {
		t.Fatal("the non-sharing second land was found instead of being prevented")
	}
}

// TestSharedSubtypeSearchOneCardIsValid verifies a single found card always
// satisfies the correlation: it enters the battlefield tapped with no second
// card required.
func TestSharedSubtypeSearchOneCardIsValid(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	forest := addCardToLibrary(g, game.Player1, basicLandWithSubtypes("Forest", types.Forest))
	addEffectSpellToStack(g, game.Player1, game.Search{
		Amount: game.Fixed(2),
		Player: game.ControllerReference(),
		Spec:   myriadLandscapeSearchSpec(),
	}, nil)
	agents := [game.NumPlayers]PlayerAgent{game.Player1: newCorrelatedSearchAgent("Forest")}

	engine.resolveTopOfStackWithChoices(g, agents, &TurnLog{})

	permanent := permanentForCard(g, forest)
	if permanent == nil {
		t.Fatal("the lone shared-subtype land did not enter the battlefield")
	}
	if !permanent.Tapped {
		t.Fatal("the lone shared-subtype land entered untapped, want tapped")
	}
}

// TestSharedSubtypeSearchZeroFailToFind verifies a legal fail-to-find: declining
// at the first stage leaves the matching lands in the library and the battlefield
// empty.
func TestSharedSubtypeSearchZeroFailToFind(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	forest := addCardToLibrary(g, game.Player1, basicLandWithSubtypes("Forest", types.Forest))
	addEffectSpellToStack(g, game.Player1, game.Search{
		Amount: game.Fixed(2),
		Player: game.ControllerReference(),
		Spec:   myriadLandscapeSearchSpec(),
	}, nil)
	agents := [game.NumPlayers]PlayerAgent{game.Player1: newCorrelatedSearchAgent()}

	engine.resolveTopOfStackWithChoices(g, agents, &TurnLog{})

	if !g.Players[game.Player1].Library.Contains(forest) {
		t.Fatal("fail-to-find removed the matching land from the library")
	}
	if permanentForCard(g, forest) != nil {
		t.Fatal("fail-to-find put the matching land onto the battlefield")
	}
}

// TestSharedSubtypeSearchIgnoresUnrelatedCards verifies the basic-land filter
// still gates candidates: a nonbasic land and a creature never match, so they
// are neither offered nor found.
func TestSharedSubtypeSearchIgnoresUnrelatedCards(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	forest := addCardToLibrary(g, game.Player1, basicLandWithSubtypes("Forest", types.Forest))
	nonbasic := addCardToLibrary(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:  "Wastes Nonbasic",
		Types: []types.Card{types.Land},
	}})
	creature := addCardToLibrary(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:  "Grizzly Bears",
		Types: []types.Card{types.Creature},
	}})
	addEffectSpellToStack(g, game.Player1, game.Search{
		Amount: game.Fixed(2),
		Player: game.ControllerReference(),
		Spec:   myriadLandscapeSearchSpec(),
	}, nil)
	agents := [game.NumPlayers]PlayerAgent{game.Player1: newCorrelatedSearchAgent("Forest", "Wastes Nonbasic", "Grizzly Bears")}

	engine.resolveTopOfStackWithChoices(g, agents, &TurnLog{})

	if permanentForCard(g, forest) == nil {
		t.Fatal("the basic land matching the filter did not enter the battlefield")
	}
	if !g.Players[game.Player1].Library.Contains(nonbasic) {
		t.Fatal("a nonbasic land matched a basic-only correlated search filter")
	}
	if !g.Players[game.Player1].Library.Contains(creature) {
		t.Fatal("a creature matched a basic-land correlated search filter")
	}
}
