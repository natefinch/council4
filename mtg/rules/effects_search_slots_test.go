package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/id"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
)

// slotSearchAgent answers the staged per-slot choices of a heterogeneous
// multi-slot search. Each slot offers only library cards matching that slot's
// filter; the agent picks the next wanted card present, or declines. It is
// stateful so it never re-picks a card across slots.
type slotSearchAgent struct {
	want   []string
	picked map[string]bool
}

func newSlotSearchAgent(want ...string) *slotSearchAgent {
	return &slotSearchAgent{want: want, picked: map[string]bool{}}
}

func (*slotSearchAgent) ChooseAction(PlayerObservation, []action.Action) action.Action {
	return action.Pass()
}

func (a *slotSearchAgent) ChooseChoice(_ PlayerObservation, request game.ChoiceRequest) []int {
	for _, option := range request.Options {
		if a.picked[option.Label] {
			continue
		}
		for _, name := range a.want {
			if option.Label == name {
				a.picked[name] = true
				return []int{option.Index}
			}
		}
	}
	return []int{}
}

// krosanVergeSearchSpec is the typed search Krosan Verge's activated ability
// lowers to: one Forest card and one Plains card, both onto the battlefield
// tapped, then shuffle.
func krosanVergeSearchSpec() game.SearchSpec {
	return game.SearchSpec{
		SourceZone:   zone.Library,
		Destination:  zone.Battlefield,
		EntersTapped: true,
		SlotFilters: []game.Selection{
			{SubtypesAny: []types.Sub{types.Forest}},
			{SubtypesAny: []types.Sub{types.Plains}},
		},
	}
}

// TestSlotSearchFindsOneCardPerSlotTapped verifies a heterogeneous two-slot
// search finds one card matching each slot filter, both enter the battlefield
// tapped, and the matching but unchosen cards stay in the library.
func TestSlotSearchFindsOneCardPerSlotTapped(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	forestA := addCardToLibrary(g, game.Player1, basicLandWithSubtypes("Forest A", types.Forest))
	forestB := addCardToLibrary(g, game.Player1, basicLandWithSubtypes("Forest B", types.Forest))
	plainsA := addCardToLibrary(g, game.Player1, basicLandWithSubtypes("Plains A", types.Plains))
	plainsB := addCardToLibrary(g, game.Player1, basicLandWithSubtypes("Plains B", types.Plains))

	addEffectSpellToStack(g, game.Player1, game.Search{
		Amount: game.Fixed(2),
		Player: game.ControllerReference(),
		Spec:   krosanVergeSearchSpec(),
	}, nil)
	agents := [game.NumPlayers]PlayerAgent{game.Player1: newSlotSearchAgent("Forest A", "Plains A")}

	engine.resolveTopOfStackWithChoices(g, agents, &TurnLog{})

	for _, want := range []id.ID{forestA, plainsA} {
		permanent := permanentForCard(g, want)
		if permanent == nil {
			t.Fatalf("slot land %v did not enter the battlefield", want)
		}
		if !permanent.Tapped {
			t.Fatalf("slot land %v entered untapped, want tapped", want)
		}
		if g.Players[game.Player1].Library.Contains(want) {
			t.Fatalf("found land %v was not removed from the library", want)
		}
	}
	for _, stay := range []id.ID{forestB, plainsB} {
		if !g.Players[game.Player1].Library.Contains(stay) {
			t.Fatalf("unchosen land %v left the library", stay)
		}
	}
}

// TestSlotSearchMayFailToFindSlot verifies the searching player may decline a
// slot: with no Plains in the library, the Forest slot still resolves and the
// Plains slot simply finds nothing.
func TestSlotSearchMayFailToFindSlot(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	forest := addCardToLibrary(g, game.Player1, basicLandWithSubtypes("Forest", types.Forest))
	creature := addCardToLibrary(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Bear", Types: []types.Card{types.Creature}}})

	addEffectSpellToStack(g, game.Player1, game.Search{
		Amount: game.Fixed(2),
		Player: game.ControllerReference(),
		Spec:   krosanVergeSearchSpec(),
	}, nil)
	agents := [game.NumPlayers]PlayerAgent{game.Player1: newSlotSearchAgent("Forest")}

	engine.resolveTopOfStackWithChoices(g, agents, &TurnLog{})

	if permanentForCard(g, forest) == nil {
		t.Fatal("Forest slot did not enter the battlefield")
	}
	if !g.Players[game.Player1].Library.Contains(creature) {
		t.Fatal("non-matching card left the library")
	}
}

// TestSlotSearchEachSlotTakesDistinctCard verifies the same card cannot fill two
// slots: with one Forest in a library and the agent wanting it for both slots,
// only one copy enters and the agent cannot double-claim it.
func TestSlotSearchEachSlotTakesDistinctCard(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	// A land that is both Forest and Plains matches both slot filters, but once
	// taken for the first slot it must not be offered for the second.
	dual := addCardToLibrary(g, game.Player1, basicLandWithSubtypes("Forest Plains", types.Forest, types.Plains))

	addEffectSpellToStack(g, game.Player1, game.Search{
		Amount: game.Fixed(2),
		Player: game.ControllerReference(),
		Spec:   krosanVergeSearchSpec(),
	}, nil)
	agents := [game.NumPlayers]PlayerAgent{game.Player1: newSlotSearchAgent("Forest Plains")}

	engine.resolveTopOfStackWithChoices(g, agents, &TurnLog{})

	if permanentForCard(g, dual) == nil {
		t.Fatal("dual land did not enter the battlefield for its single slot")
	}
	if g.Players[game.Player1].Library.Contains(dual) {
		t.Fatal("dual land was not removed from the library")
	}
}
