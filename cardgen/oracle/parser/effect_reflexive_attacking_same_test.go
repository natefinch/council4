package parser

import (
	"testing"

	"github.com/natefinch/council4/cardgen/oracle/shared"
)

// TestCreditEachOpponentAttackingSameRiderAccepts verifies the trailing "Each
// opponent attacking that player does the same." sentence folds onto the lone
// controller create-token effect of an enchanted-player combat trigger (Curse of
// Opulence): the create records the rider span and stays exact, and the rider
// sentence is marked and emptied so nothing downstream treats it as a separate
// effect.
func TestCreditEachOpponentAttackingSameRiderAccepts(t *testing.T) {
	t.Parallel()
	document, diagnostics := Parse(
		"Whenever enchanted player is attacked, create a Gold token. Each opponent attacking that player does the same.",
		Context{CardName: "Curse of Opulence"})
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	ability := document.Abilities[0]
	if ability.Trigger == nil || ability.Trigger.TriggerEvent == nil || !ability.Trigger.TriggerEvent.EnchantedPlayerIsAttacked {
		t.Fatalf("trigger = %#v, want enchanted-player-is-attacked", ability.Trigger)
	}
	if len(ability.Sentences) != 2 {
		t.Fatalf("sentences = %d, want 2 (create + rider)", len(ability.Sentences))
	}
	effects := ability.Sentences[0].Effects
	if len(effects) != 1 || effects[0].Kind != EffectCreate {
		t.Fatalf("first sentence effects = %#v, want a single create", effects)
	}
	effect := effects[0]
	if effect.EachOpponentAttackingSameRiderSpan == (shared.Span{}) {
		t.Error("EachOpponentAttackingSameRiderSpan is unset, want the rider sentence span")
	}
	if effect.HasUnrecognizedSibling {
		t.Error("HasUnrecognizedSibling = true, want false after crediting the rider")
	}
	if !effect.Exact {
		t.Error("Exact = false, want true for the credited create")
	}
	if !ability.Sentences[1].EachOpponentAttackingSameRider {
		t.Error("rider sentence EachOpponentAttackingSameRider = false, want true")
	}
	if len(ability.Sentences[1].Effects) != 0 {
		t.Errorf("rider sentence effects = %#v, want cleared", ability.Sentences[1].Effects)
	}
}

// TestCreditEachOpponentAttackingSameRiderRequiresTrigger verifies the rider is
// not credited when the reflexive sentence is not attached to an
// enchanted-player-is-attacked trigger. "That player" then has no attack-target
// antecedent, so the create must stay a plain controller create with the rider
// left as an unrecognized sibling that fails closed downstream.
func TestCreditEachOpponentAttackingSameRiderRequiresTrigger(t *testing.T) {
	t.Parallel()
	document, _ := Parse(
		"Create a Gold token. Each opponent attacking that player does the same.",
		Context{InstantOrSorcery: true, CardName: "Not A Curse"})
	for _, ability := range document.Abilities {
		for _, sentence := range ability.Sentences {
			if sentence.EachOpponentAttackingSameRider {
				t.Fatal("rider credited without an enchanted-player trigger, want fail closed")
			}
			for _, effect := range sentence.Effects {
				if effect.EachOpponentAttackingSameRiderSpan != (shared.Span{}) {
					t.Fatal("create recorded a rider span without an enchanted-player trigger, want fail closed")
				}
			}
		}
	}
}

// TestCreditEachOpponentAttackingSameRiderRequiresCreate verifies the rider is
// not credited when the preceding effect is not a controller create-token. The
// "does the same" anaphor only has a token creation to widen; a non-create
// effect leaves the rider uncredited so the ability fails closed rather than
// silently attaching the group creation to an unrelated effect.
func TestCreditEachOpponentAttackingSameRiderRequiresCreate(t *testing.T) {
	t.Parallel()
	document, _ := Parse(
		"Whenever enchanted player is attacked, draw a card. Each opponent attacking that player does the same.",
		Context{CardName: "Curse of Verbosity"})
	for _, ability := range document.Abilities {
		for _, sentence := range ability.Sentences {
			if sentence.EachOpponentAttackingSameRider {
				t.Fatal("rider credited onto a non-create effect, want fail closed")
			}
		}
	}
}
