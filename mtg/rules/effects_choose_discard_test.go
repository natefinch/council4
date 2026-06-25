package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// handDiscardCard builds a CardDef with a fixed mana value and types for the
// targeted-discard runtime tests.
func handDiscardCard(name string, manaValue int, cardTypes ...types.Card) *game.CardDef {
	return &game.CardDef{CardFace: game.CardFace{
		Name:     name,
		ManaCost: opt.Val(cost.Mana{cost.O(manaValue)}),
		Types:    cardTypes,
	}}
}

// TestChooseDiscardFromHandDuressExcludesCreatures proves the Duress shape lets
// the resolving controller pick a noncreature card from the target's hand to
// discard, leaving creature cards untouched.
func TestChooseDiscardFromHandDuressExcludesCreatures(t *testing.T) {
	t.Parallel()
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	addEffectSpellToStack(g, game.Player1, game.ChooseDiscardFromHand{
		Player:          game.TargetPlayerReference(0),
		ExcludeCreature: true,
	}, []game.Target{game.PlayerTarget(game.Player2)})

	creature := addCardToHand(g, game.Player2, handDiscardCard("Bear", 2, types.Creature))
	spell := addCardToHand(g, game.Player2, handDiscardCard("Bolt", 1, types.Instant))

	engine.resolveTopOfStack(g, &TurnLog{})

	if !g.Players[game.Player2].Hand.Contains(creature) {
		t.Fatal("creature card was discarded; Duress must not discard creatures")
	}
	if g.Players[game.Player2].Hand.Contains(spell) {
		t.Fatal("noncreature card still in hand; want it discarded")
	}
	if !g.Players[game.Player2].Graveyard.Contains(spell) {
		t.Fatal("noncreature card not in graveyard after discard")
	}
}

// TestChooseDiscardFromHandManaValueBound proves the Inquisition of Kozilek
// shape only offers cards whose mana value is within the bound, leaving costlier
// cards in hand.
func TestChooseDiscardFromHandManaValueBound(t *testing.T) {
	t.Parallel()
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	addEffectSpellToStack(g, game.Player1, game.ChooseDiscardFromHand{
		Player:       game.TargetPlayerReference(0),
		ExcludeLand:  true,
		MaxManaValue: opt.Val(3),
	}, []game.Target{game.PlayerTarget(game.Player2)})

	cheap := addCardToHand(g, game.Player2, handDiscardCard("Cheap Spell", 2, types.Sorcery))
	expensive := addCardToHand(g, game.Player2, handDiscardCard("Expensive Spell", 6, types.Sorcery))

	log := TurnLog{}
	engine.resolveTopOfStack(g, &log)

	if g.Players[game.Player2].Hand.Contains(cheap) {
		t.Fatal("eligible cheap card not discarded")
	}
	if !g.Players[game.Player2].Hand.Contains(expensive) {
		t.Fatal("over-bound card discarded; mana value filter ignored")
	}
}

// TestChooseDiscardFromHandThoughtseizeLifeLoss proves the Thoughtseize shape
// discards a nonland card and the controller loses 2 life from the trailing
// rider.
func TestChooseDiscardFromHandThoughtseizeLifeLoss(t *testing.T) {
	t.Parallel()
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	startLife := g.Players[game.Player1].Life
	addInstructionSpellToStackForController(g, game.Player1, []game.Instruction{
		{Primitive: game.ChooseDiscardFromHand{
			Player:      game.TargetPlayerReference(0),
			ExcludeLand: true,
		}},
		{Primitive: game.LoseLife{Amount: game.Fixed(2), Player: game.ControllerReference()}},
	}, []game.Target{game.PlayerTarget(game.Player2)})

	land := addCardToHand(g, game.Player2, handDiscardCard("Forest", 0, types.Land))
	spell := addCardToHand(g, game.Player2, handDiscardCard("Bolt", 1, types.Instant))

	engine.resolveTopOfStack(g, &TurnLog{})

	if !g.Players[game.Player2].Hand.Contains(land) {
		t.Fatal("land card discarded; Thoughtseize must not discard lands")
	}
	if !g.Players[game.Player2].Graveyard.Contains(spell) {
		t.Fatal("nonland card not discarded")
	}
	if got := g.Players[game.Player1].Life; got != startLife-2 {
		t.Fatalf("controller life = %d, want %d", got, startLife-2)
	}
}
