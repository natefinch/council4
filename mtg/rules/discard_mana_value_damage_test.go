package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// kujataChapterThreeSequence builds Summon: Kujata's chapter III instruction
// sequence by hand: discard a card (publishing it under linkedKey), draw two,
// then deal damage to each opponent equal to the discarded card's mana value.
func kujataChapterThreeSequence(linkedKey game.LinkedKey) []game.Instruction {
	return []game.Instruction{
		{Primitive: game.Discard{
			Amount:        game.Fixed(1),
			Player:        game.ControllerReference(),
			PublishLinked: linkedKey,
		}},
		{Primitive: game.Draw{
			Amount: game.Fixed(2),
			Player: game.ControllerReference(),
		}},
		{Primitive: game.Damage{
			Amount: game.Dynamic(game.DynamicAmount{
				Kind:   game.DynamicAmountObjectManaValue,
				Object: game.LinkedObjectReference(string(linkedKey)),
			}),
			Recipient:    game.PlayerGroupDamageRecipient(game.OpponentsReference()),
			DamageSource: opt.Val(game.SourcePermanentReference()),
		}},
	}
}

// TestDiscardThenManaValueDamageToOpponents proves the reflexive
// discard-then-mana-value-damage sequence (Summon: Kujata chapter III): the
// discarded card is published under a linked key and the follow-up damage reads
// that card's mana value, dealing it to each opponent even though the card has
// left the hand for the graveyard.
func TestDiscardThenManaValueDamageToOpponents(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	const linkedKey = game.LinkedKey("discarded-card-mana-value")

	kujata := addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:  "Summon: Kujata",
		Types: []types.Card{types.Creature},
	}})
	obj := &game.StackObject{
		Kind:         game.StackTriggeredAbility,
		SourceID:     kujata.ObjectID,
		SourceCardID: kujata.CardInstanceID,
		Controller:   game.Player1,
	}

	addCardToHand(g, game.Player1, evidenceCard("Discarded Spell", 3))
	addCardToLibrary(g, game.Player1, evidenceCard("Library One", 1))
	addCardToLibrary(g, game.Player1, evidenceCard("Library Two", 1))

	startLife := g.Players[game.Player2].Life
	startHand := g.Players[game.Player1].Hand.Size()

	log := TurnLog{}
	for _, instr := range kujataChapterThreeSequence(linkedKey) {
		engine.resolveInstructionWithChoices(g, obj, &instr, [game.NumPlayers]PlayerAgent{}, &log)
	}

	if got := g.Players[game.Player2].Life; got != startLife-3 {
		t.Fatalf("opponent life = %d, want %d (discarded card mana value 3)", got, startLife-3)
	}
	// One card discarded, two drawn: net +1 in hand.
	if got := g.Players[game.Player1].Hand.Size(); got != startHand+1 {
		t.Fatalf("controller hand size = %d, want %d", got, startHand+1)
	}
	if got := g.Players[game.Player1].Graveyard.Size(); got != 1 {
		t.Fatalf("graveyard size = %d, want 1 discarded card", got)
	}
}

// TestDiscardThenManaValueDamageEmptyHandDealsNoDamage proves the sequence fails
// closed when the controller has no card to discard: nothing is published, so
// the mana-value damage resolves to zero and opponents are untouched.
func TestDiscardThenManaValueDamageEmptyHandDealsNoDamage(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	const linkedKey = game.LinkedKey("discarded-card-mana-value")

	kujata := addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:  "Summon: Kujata",
		Types: []types.Card{types.Creature},
	}})
	obj := &game.StackObject{
		Kind:         game.StackTriggeredAbility,
		SourceID:     kujata.ObjectID,
		SourceCardID: kujata.CardInstanceID,
		Controller:   game.Player1,
	}
	addCardToLibrary(g, game.Player1, evidenceCard("Library One", 1))
	addCardToLibrary(g, game.Player1, evidenceCard("Library Two", 1))

	startLife := g.Players[game.Player2].Life
	log := TurnLog{}
	for _, instr := range kujataChapterThreeSequence(linkedKey) {
		engine.resolveInstructionWithChoices(g, obj, &instr, [game.NumPlayers]PlayerAgent{}, &log)
	}

	if got := g.Players[game.Player2].Life; got != startLife {
		t.Fatalf("opponent life = %d, want %d (no card discarded)", got, startLife)
	}
}
