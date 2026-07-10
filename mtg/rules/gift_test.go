package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/opt"
)

// TestEffectConditionGiftPromisedGatesResolution verifies that an effect gated
// on the "if the gift was promised" condition (CR 702.171) resolves only when
// the resolving stack object recorded a promised gift. The promoted rider (a
// draw) runs when GiftPromised is true and is skipped otherwise. Unlike the
// kicker gate, a copy of a promised spell is itself promised to the same
// opponent (CR 707.10), so the promoted rider still runs on a promised copy and
// the "if the gift wasn't promised" penalty does not. The negated form models
// that penalty clause, which fires only when no gift was promised.
func TestEffectConditionGiftPromisedGatesResolution(t *testing.T) {
	gatedDraw := func(negate bool) *game.Instruction {
		return &game.Instruction{
			Primitive: game.Draw{Amount: game.Fixed(1), Player: game.ControllerReference()},
			Condition: opt.Val(game.EffectCondition{
				Condition: opt.Val(game.Condition{Negate: negate, GiftPromised: true}),
			}),
		}
	}

	resolve := func(negate, promised, isCopy bool) int {
		g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
		engine := NewEngine(nil)
		addCardToLibrary(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Drawn"}})
		obj := &game.StackObject{
			Kind:         game.StackSpell,
			Controller:   game.Player1,
			GiftPromised: promised,
			Copy:         isCopy,
		}
		engine.resolveInstructionWithChoices(g, obj, gatedDraw(negate), [game.NumPlayers]PlayerAgent{}, &TurnLog{})
		return g.Players[game.Player1].Hand.Size()
	}

	if got := resolve(false, true, false); got != 1 {
		t.Fatalf("promised bonus draw: hand size = %d, want 1", got)
	}
	if got := resolve(false, false, false); got != 0 {
		t.Fatalf("unpromised bonus draw: hand size = %d, want 0", got)
	}
	if got := resolve(false, true, true); got != 1 {
		t.Fatalf("copied promised bonus draw: hand size = %d, want 1 (a copy of a promised spell is promised)", got)
	}
	if got := resolve(true, false, false); got != 1 {
		t.Fatalf("unpromised penalty draw: hand size = %d, want 1", got)
	}
	if got := resolve(true, true, false); got != 0 {
		t.Fatalf("promised penalty draw: hand size = %d, want 0", got)
	}
	if got := resolve(true, true, true); got != 0 {
		t.Fatalf("copied promised penalty draw: hand size = %d, want 0 (penalty does not fire for a promised copy)", got)
	}
}

// giftSpellDef builds an instant with a Gift keyword action that draws the
// promised opponent a card, a base effect that gains its controller 1 life, and
// a promoted effect gated on "if the gift was promised" that gains 2 more.
func giftSpellDef() *game.CardDef {
	delivery := game.Mode{Sequence: []game.Instruction{{
		Primitive: game.Draw{Amount: game.Fixed(1), Player: game.GiftRecipientReference()},
	}}}.Ability()
	return &game.CardDef{
		CardFace: game.CardFace{
			Name: "Test Gift Spell",
			StaticAbilities: []game.StaticAbility{{
				KeywordAbilities: []game.KeywordAbility{game.GiftKeyword{Delivery: delivery}},
			}},
			SpellAbility: opt.Val(game.Mode{Sequence: []game.Instruction{
				{Primitive: game.GainLife{Amount: game.Fixed(1), Player: game.ControllerReference()}},
				{
					Primitive: game.GainLife{Amount: game.Fixed(2), Player: game.ControllerReference()},
					Condition: opt.Val(game.EffectCondition{
						Condition: opt.Val(game.Condition{GiftPromised: true}),
					}),
				},
			}}.Ability()),
		},
	}
}

// TestGiftDeliveryAndBonusOnPromise verifies the two halves of the Gift keyword
// action at spell resolution: promising the gift delivers it to the chosen
// opponent (the recipient draws a card) and enables the "if the gift was
// promised" bonus, while not promising delivers nothing and applies only the
// base effect.
func TestGiftDeliveryAndBonusOnPromise(t *testing.T) {
	resolve := func(promised bool) (recipientDrew int, controllerLife int) {
		g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
		engine := NewEngine(nil)
		g.Players[game.Player1].Life = 10
		cardID := addCardToLibrary(g, game.Player1, giftSpellDef())
		card, _ := g.GetCardInstance(cardID)
		// The recipient needs a card in library to receive the gifted draw.
		addCardToLibrary(g, game.Player2, &game.CardDef{CardFace: game.CardFace{Name: "Gifted"}})
		obj := &game.StackObject{
			Kind:          game.StackSpell,
			SourceID:      cardID,
			Controller:    game.Player1,
			GiftPromised:  promised,
			GiftRecipient: game.Player2,
		}
		engine.resolveSpellEffectsWithChoices(g, obj, card, [game.NumPlayers]PlayerAgent{}, &TurnLog{})
		return g.Players[game.Player2].Hand.Size(), g.Players[game.Player1].Life
	}

	drew, life := resolve(true)
	if drew != 1 {
		t.Fatalf("promised gift: recipient hand size = %d, want 1 (gift delivered)", drew)
	}
	if life != 13 {
		t.Fatalf("promised gift: controller life = %d, want 13 (1 base + 2 bonus)", life)
	}

	drew, life = resolve(false)
	if drew != 0 {
		t.Fatalf("unpromised gift: recipient hand size = %d, want 0 (no gift delivered)", drew)
	}
	if life != 11 {
		t.Fatalf("unpromised gift: controller life = %d, want 11 (base effect only)", life)
	}
}

// TestGiftRecipientReferenceResolvesToPromisedOpponent verifies the gift
// recipient player reference resolves to the resolving spell's captured
// recipient only when a gift was promised, mirroring the reference machinery the
// gift delivery relies on.
func TestGiftRecipientReferenceResolvesToPromisedOpponent(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})

	promised := &game.StackObject{Controller: game.Player1, GiftPromised: true, GiftRecipient: game.Player3}
	if got, ok := resolvePlayerReference(g, promised, game.GiftRecipientReference()); !ok || got != game.Player3 {
		t.Fatalf("promised recipient reference = (%v, %v), want (Player3, true)", got, ok)
	}

	unpromised := &game.StackObject{Controller: game.Player1, GiftPromised: false, GiftRecipient: game.Player3}
	if _, ok := resolvePlayerReference(g, unpromised, game.GiftRecipientReference()); ok {
		t.Fatal("unpromised recipient reference resolved, want unresolved")
	}
}
