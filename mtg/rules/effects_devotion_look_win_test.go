package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/compare"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// addBlueDevotionPermanent puts a permanent controlled by controller onto the
// battlefield whose mana cost is pips blue mana symbols, contributing pips to
// the controller's devotion to blue.
func addBlueDevotionPermanent(g *game.Game, controller game.PlayerID, pips int) *game.Permanent {
	manaCost := make(cost.Mana, 0, pips)
	for range pips {
		manaCost = append(manaCost, cost.U)
	}
	return addCombatPermanent(g, controller, &game.CardDef{CardFace: game.CardFace{
		Name:     "Blue Devotion Permanent",
		Types:    []types.Card{types.Enchantment},
		ManaCost: opt.Val(manaCost),
	}})
}

// devotionLookWinSequence builds the two-instruction Thassa's Oracle body: a Dig
// looking at devotion-to-blue cards that keeps up to one on top and bottoms the
// rest, followed by a controller win gated on the controller's library being no
// larger than that same live devotion.
func devotionLookWinSequence() []game.Instruction {
	devotion := game.DynamicAmount{
		Kind:   game.DynamicAmountDevotion,
		Colors: []color.Color{color.Blue},
	}
	return []game.Instruction{
		{
			Primitive: game.Dig{
				Player:      game.ControllerReference(),
				Look:        game.Dynamic(devotion),
				Take:        game.Fixed(1),
				TakeUpTo:    true,
				Destination: zone.Library,
				Remainder:   game.DigRemainderLibraryBottom,
			},
		},
		{
			Primitive: game.PlayerWinsGame{Player: game.ControllerReference()},
			Condition: opt.Val(game.EffectCondition{
				Condition: opt.Val(game.Condition{
					Aggregates: []game.AggregateComparison{{
						Aggregate:   game.AggregateControllerLibrarySize,
						Op:          compare.LessOrEqual,
						ValueAmount: opt.Val(devotion),
					}},
				}),
			}),
		},
	}
}

// TestDevotionLookWinKeepsChosenOnTopAndBottomsRest proves the look step of
// Thassa's Oracle: with devotion 2 the controller looks at the top two cards,
// keeps the chosen one on top of the library, and bottoms the other. The library
// (four cards) is larger than the devotion (two), so the controller does not win.
func TestDevotionLookWinKeepsChosenOnTopAndBottomsRest(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	addBlueDevotionPermanent(g, game.Player1, 2)
	// Add bottom-to-top: bottom deepest, top last. peekLibrary returns top-first,
	// so the seen order is top, second, ...
	bottom := addCardToLibrary(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Bottom"}})
	third := addCardToLibrary(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Third"}})
	second := addCardToLibrary(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Second"}})
	top := addCardToLibrary(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Top"}})

	addInstructionSpellToStackForController(g, game.Player1, devotionLookWinSequence(), nil)
	log := TurnLog{}
	// Seen order is [top, second]; choosing index 1 keeps the second card on top.
	agents := [game.NumPlayers]PlayerAgent{game.Player1: &choiceOnlyAgent{choices: [][]int{{1}}}}

	engine.resolveTopOfStackWithChoices(g, agents, &log)

	player := g.Players[game.Player1]
	libraryTop, ok := player.Library.Top()
	if !ok || libraryTop != second {
		t.Fatalf("library top = %v, want the chosen card %v", libraryTop, second)
	}
	libraryBottom, ok := player.Library.Bottom()
	if !ok || libraryBottom != top {
		t.Fatalf("library bottom = %v, want the unchosen looked-at card %v", libraryBottom, top)
	}
	if player.Library.Size() != 4 {
		t.Fatalf("library size = %d, want 4 (the look never removes cards)", player.Library.Size())
	}
	if !player.Library.Contains(third) || !player.Library.Contains(bottom) {
		t.Fatal("look disturbed cards below the looked-at window")
	}
	if g.MarkedToLoseGame[game.Player2] {
		t.Fatal("opponents were marked to lose although devotion is below the library size")
	}
}

// TestDevotionLookWinWinsWhenDevotionAtLeastLibrarySize proves the win step: when
// devotion (three) is at least the library size (two), the controller wins and
// every opponent is marked to lose.
func TestDevotionLookWinWinsWhenDevotionAtLeastLibrarySize(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	addBlueDevotionPermanent(g, game.Player1, 3)
	addCardToLibrary(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "A"}})
	addCardToLibrary(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "B"}})

	addInstructionSpellToStackForController(g, game.Player1, devotionLookWinSequence(), nil)
	log := TurnLog{}
	agents := [game.NumPlayers]PlayerAgent{game.Player1: &choiceOnlyAgent{choices: [][]int{{0}}}}

	engine.resolveTopOfStackWithChoices(g, agents, &log)

	if g.MarkedToLoseGame[game.Player1] {
		t.Fatal("the winning controller was marked to lose")
	}
	for _, opponent := range []game.PlayerID{game.Player2, game.Player3, game.Player4} {
		if !g.MarkedToLoseGame[opponent] {
			t.Errorf("opponent %v was not marked to lose the game", opponent)
		}
	}
	engine.checkStateBasedActions(g)
	winner, ok := g.Winner()
	if !ok || winner.ID != game.Player1 {
		t.Fatalf("Winner() = %v, %v, want Player1 as sole survivor", winner, ok)
	}
}

// TestDevotionLookWinNoWinWhenLibraryLargerThanDevotion proves the controller
// does not win when the library (three) exceeds their devotion (one).
func TestDevotionLookWinNoWinWhenLibraryLargerThanDevotion(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	addBlueDevotionPermanent(g, game.Player1, 1)
	addCardToLibrary(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "A"}})
	addCardToLibrary(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "B"}})
	addCardToLibrary(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "C"}})

	addInstructionSpellToStackForController(g, game.Player1, devotionLookWinSequence(), nil)
	log := TurnLog{}
	agents := [game.NumPlayers]PlayerAgent{game.Player1: &choiceOnlyAgent{choices: [][]int{{0}}}}

	engine.resolveTopOfStackWithChoices(g, agents, &log)

	for _, opponent := range []game.PlayerID{game.Player2, game.Player3, game.Player4} {
		if g.MarkedToLoseGame[opponent] {
			t.Errorf("opponent %v was marked to lose although the library exceeds devotion", opponent)
		}
	}
}

// TestDevotionLookWinEmptyLibraryStillWins proves the classic empty-library
// Thassa's Oracle win: with an empty library the look is a harmless no-op and the
// controller wins because devotion (zero) is at least the library size (zero).
func TestDevotionLookWinEmptyLibraryStillWins(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	// No blue permanents: devotion is zero. No library cards: size is zero.

	addInstructionSpellToStackForController(g, game.Player1, devotionLookWinSequence(), nil)
	log := TurnLog{}

	engine.resolveTopOfStackWithChoices(g, [game.NumPlayers]PlayerAgent{}, &log)

	if g.MarkedToLoseGame[game.Player1] {
		t.Fatal("the winning controller was marked to lose")
	}
	for _, opponent := range []game.PlayerID{game.Player2, game.Player3, game.Player4} {
		if !g.MarkedToLoseGame[opponent] {
			t.Errorf("opponent %v was not marked to lose with an empty library", opponent)
		}
	}
}

// TestDevotionLookWinZeroDevotionNonEmptyLibraryNoWin proves that zero devotion
// with a non-empty library neither looks at any card nor wins, because the
// library size (one) exceeds the devotion (zero).
func TestDevotionLookWinZeroDevotionNonEmptyLibraryNoWin(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	only := addCardToLibrary(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Only"}})

	addInstructionSpellToStackForController(g, game.Player1, devotionLookWinSequence(), nil)
	log := TurnLog{}

	engine.resolveTopOfStackWithChoices(g, [game.NumPlayers]PlayerAgent{}, &log)

	player := g.Players[game.Player1]
	if player.Library.Size() != 1 || !player.Library.Contains(only) {
		t.Fatalf("library = %d cards, want the single card untouched", player.Library.Size())
	}
	for _, opponent := range []game.PlayerID{game.Player2, game.Player3, game.Player4} {
		if g.MarkedToLoseGame[opponent] {
			t.Errorf("opponent %v was marked to lose although devotion is below the library size", opponent)
		}
	}
}
