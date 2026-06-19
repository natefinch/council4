package rules

import (
	"slices"
	"strings"
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/id"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// splitSearchAgent answers the three choices a split-destination search makes:
// which matching cards to find, which found card enters the primary slot when
// two are found, and which slot the lone found card fills when only one is
// found. It dispatches on the choice prompt.
type splitSearchAgent struct {
	find        []string // matching card names to find
	primaryCard string   // card to assign to the primary slot (two-card case)
	slot        string   // slot label to fill (one-card case)
}

func (*splitSearchAgent) ChooseAction(PlayerObservation, []action.Action) action.Action {
	return action.Pass()
}

func (a *splitSearchAgent) ChooseChoice(_ PlayerObservation, request game.ChoiceRequest) []int {
	switch {
	case strings.Contains(request.Prompt, "choose matching cards to find"):
		var out []int
		for _, option := range request.Options {
			if slices.Contains(a.find, option.Label) {
				out = append(out, option.Index)
			}
		}
		return out
	case strings.Contains(request.Prompt, "which card goes to"):
		for _, option := range request.Options {
			if option.Label == a.primaryCard {
				return []int{option.Index}
			}
		}
		return nil
	case strings.Contains(request.Prompt, "where to put the found card"):
		for _, option := range request.Options {
			if option.Label == a.slot {
				return []int{option.Index}
			}
		}
		return nil
	default:
		return nil
	}
}

func splitBasicLand(name string, sub types.Sub) *game.CardDef {
	return &game.CardDef{CardFace: game.CardFace{
		Name:       name,
		Supertypes: []types.Super{types.Basic},
		Types:      []types.Card{types.Land},
		Subtypes:   []types.Sub{sub},
	}}
}

func cultivateSpec() game.SearchSpec {
	return game.SearchSpec{
		SourceZone:       zone.Library,
		Destination:      zone.Battlefield,
		CardType:         opt.Val(types.Land),
		Supertype:        opt.Val(types.Basic),
		Reveal:           true,
		EntersTapped:     true,
		SplitDestination: opt.Val(game.SearchDestination{Zone: zone.Hand}),
	}
}

// handFirstSplitSpec is a split-destination search whose primary slot is hand
// and whose secondary slot is the tapped battlefield. Parser/lowering accept
// hand-first wordings, so the runtime must keep prompt and routing aligned for
// this orientation too.
func handFirstSplitSpec() game.SearchSpec {
	return game.SearchSpec{
		SourceZone:       zone.Library,
		Destination:      zone.Hand,
		CardType:         opt.Val(types.Land),
		Supertype:        opt.Val(types.Basic),
		Reveal:           true,
		EntersTapped:     false,
		SplitDestination: opt.Val(game.SearchDestination{Zone: zone.Battlefield, EntersTapped: true}),
	}
}

// TestSplitSearchTwoCardsHandFirstPrimary verifies a hand-first split spec: the
// primary slot is hand, so the player's chosen card goes to hand and the other
// enters the battlefield tapped. The agent selects Forest, which is not the
// fallback default (the most-recently-added Island sits at found[0]), so a pass
// requires the prompt-driven selection and proves prompt and routing both name
// the hand destination.
func TestSplitSearchTwoCardsHandFirstPrimary(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	forest := addCardToLibrary(g, game.Player1, splitBasicLand("Forest", types.Forest))
	island := addCardToLibrary(g, game.Player1, splitBasicLand("Island", types.Island))
	addEffectSpellToStack(g, game.Player1, game.Search{
		Amount: game.Fixed(2),
		Player: game.ControllerReference(),
		Spec:   handFirstSplitSpec(),
	}, nil)
	agents := [game.NumPlayers]PlayerAgent{game.Player1: &splitSearchAgent{find: []string{"Forest", "Island"}, primaryCard: "Forest"}}

	engine.resolveTopOfStackWithChoices(g, agents, &TurnLog{})

	if !g.Players[game.Player1].Hand.Contains(forest) {
		t.Fatal("the chosen primary-slot land did not go to hand")
	}
	if permanentForCard(g, forest) != nil {
		t.Fatal("the hand-slot land incorrectly entered the battlefield")
	}
	permanent := permanentForCard(g, island)
	if permanent == nil {
		t.Fatal("the other found land did not enter the battlefield")
	}
	if !permanent.Tapped {
		t.Fatal("the battlefield-slot land entered untapped, want tapped")
	}
	if g.Players[game.Player1].Hand.Contains(island) {
		t.Fatal("the battlefield-slot land incorrectly went to hand")
	}
}

// TestSplitSearchTwoCardsAssignsBattlefieldAndHand verifies the two-card flow of
// a Cultivate-style split tutor: the player chooses which found basic land
// enters the battlefield tapped, the other goes to hand, both are revealed, and
// the library is shuffled afterward.
func TestSplitSearchTwoCardsAssignsBattlefieldAndHand(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	forest := addCardToLibrary(g, game.Player1, splitBasicLand("Forest", types.Forest))
	island := addCardToLibrary(g, game.Player1, splitBasicLand("Island", types.Island))
	// Filler cards remain in the library so a reorder evidences the shuffle.
	var filler []id.ID
	for _, name := range []string{"Filler1", "Filler2", "Filler3", "Filler4", "Filler5", "Filler6"} {
		filler = append(filler, addCardToLibrary(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: name, Types: []types.Card{types.Creature}}}))
	}
	before := g.Players[game.Player1].Library.All()
	addEffectSpellToStack(g, game.Player1, game.Search{
		Amount: game.Fixed(2),
		Player: game.ControllerReference(),
		Spec:   cultivateSpec(),
	}, nil)
	agents := [game.NumPlayers]PlayerAgent{game.Player1: &splitSearchAgent{find: []string{"Forest", "Island"}, primaryCard: "Forest"}}

	engine.resolveTopOfStackWithChoices(g, agents, &TurnLog{})

	permanent := permanentForCard(g, forest)
	if permanent == nil {
		t.Fatal("the chosen basic land did not enter the battlefield")
	}
	if !permanent.Tapped {
		t.Fatal("the battlefield-slot land entered untapped, want tapped")
	}
	if !g.Players[game.Player1].Hand.Contains(island) {
		t.Fatal("the other found land did not go to hand")
	}
	if permanentForCard(g, island) != nil {
		t.Fatal("the hand-slot land incorrectly entered the battlefield")
	}
	if g.Players[game.Player1].Library.Contains(forest) || g.Players[game.Player1].Library.Contains(island) {
		t.Fatal("found lands were not removed from the library")
	}
	assertEvent(t, g.Events, game.EventCardRevealed, func(event game.Event) bool {
		return event.CardID == forest
	})
	assertEvent(t, g.Events, game.EventCardRevealed, func(event game.Event) bool {
		return event.CardID == island
	})
	after := g.Players[game.Player1].Library.All()
	if !slices.Contains(before, filler[0]) {
		t.Fatal("test setup: filler not in starting library")
	}
	if slices.Equal(before[2:], after) {
		t.Fatal("library order unchanged; split search did not shuffle")
	}
}

// TestSplitSearchOneCardPlayerChoosesBattlefield verifies the one-card flow:
// when only one basic land is found, the player may choose to put it onto the
// battlefield tapped rather than into hand (CR 701.19).
func TestSplitSearchOneCardPlayerChoosesBattlefield(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	forest := addCardToLibrary(g, game.Player1, splitBasicLand("Forest", types.Forest))
	addEffectSpellToStack(g, game.Player1, game.Search{
		Amount: game.Fixed(2),
		Player: game.ControllerReference(),
		Spec:   cultivateSpec(),
	}, nil)
	agents := [game.NumPlayers]PlayerAgent{game.Player1: &splitSearchAgent{find: []string{"Forest"}, slot: "battlefield tapped"}}

	engine.resolveTopOfStackWithChoices(g, agents, &TurnLog{})

	permanent := permanentForCard(g, forest)
	if permanent == nil {
		t.Fatal("the lone found land did not enter the battlefield")
	}
	if !permanent.Tapped {
		t.Fatal("the lone found land entered untapped, want tapped")
	}
	if g.Players[game.Player1].Hand.Contains(forest) {
		t.Fatal("the lone found land went to hand despite the battlefield choice")
	}
}

// TestSplitSearchOneCardPlayerChoosesHand verifies the complementary one-card
// choice: the lone found land may instead go to hand.
func TestSplitSearchOneCardPlayerChoosesHand(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	forest := addCardToLibrary(g, game.Player1, splitBasicLand("Forest", types.Forest))
	addEffectSpellToStack(g, game.Player1, game.Search{
		Amount: game.Fixed(2),
		Player: game.ControllerReference(),
		Spec:   cultivateSpec(),
	}, nil)
	agents := [game.NumPlayers]PlayerAgent{game.Player1: &splitSearchAgent{find: []string{"Forest"}, slot: "hand"}}

	engine.resolveTopOfStackWithChoices(g, agents, &TurnLog{})

	if !g.Players[game.Player1].Hand.Contains(forest) {
		t.Fatal("the lone found land did not go to hand on the hand choice")
	}
	if permanentForCard(g, forest) != nil {
		t.Fatal("the lone found land entered the battlefield despite the hand choice")
	}
}

// TestSplitSearchZeroCardsFailToFind verifies a legal fail-to-find: the player
// may decline to find any basic land, leaving the matching land in the library,
// nothing in hand, and no permanent.
func TestSplitSearchZeroCardsFailToFind(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	forest := addCardToLibrary(g, game.Player1, splitBasicLand("Forest", types.Forest))
	addEffectSpellToStack(g, game.Player1, game.Search{
		Amount: game.Fixed(2),
		Player: game.ControllerReference(),
		Spec:   cultivateSpec(),
	}, nil)
	agents := [game.NumPlayers]PlayerAgent{game.Player1: &splitSearchAgent{find: nil}}

	engine.resolveTopOfStackWithChoices(g, agents, &TurnLog{})

	if !g.Players[game.Player1].Library.Contains(forest) {
		t.Fatal("fail-to-find removed the matching land from the library")
	}
	if g.Players[game.Player1].Hand.Contains(forest) {
		t.Fatal("fail-to-find moved the matching land to hand")
	}
	if permanentForCard(g, forest) != nil {
		t.Fatal("fail-to-find put the matching land onto the battlefield")
	}
	assertNoEvent(t, g.Events, game.EventCardRevealed, func(event game.Event) bool {
		return event.CardID == forest
	})
}
