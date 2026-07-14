package parser

import (
	"testing"

	"github.com/natefinch/council4/cardgen/oracle/shared"
)

// creditedRiderAbility parses text as a single triggered ability and returns it,
// failing the test on any diagnostic or an ability count other than one. It backs
// the reflexive attacking-opponent rider tests, which each assert one enchanted
// player combat trigger.
func creditedRiderAbility(t *testing.T, text, cardName string) Ability {
	t.Helper()
	document, diagnostics := Parse(text, Context{CardName: cardName})
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	if len(document.Abilities) != 1 {
		t.Fatalf("abilities = %d, want 1", len(document.Abilities))
	}
	ability := document.Abilities[0]
	if ability.Trigger == nil || ability.Trigger.TriggerEvent == nil ||
		!ability.Trigger.TriggerEvent.EnchantedPlayerIsAttacked {
		t.Fatalf("trigger = %#v, want enchanted-player-is-attacked", ability.Trigger)
	}
	return ability
}

// assertRiderFolded asserts the ability's first sentence holds a single credited
// controller effect of the wanted kind — recording the rider span, cleared of the
// unrecognized-sibling flag, and marked exact — and that its second sentence is
// the marked, emptied rider sentence. It is the shared positive assertion for the
// "does the same." family (create/gain/draw) and the explicit Bounty untap rider.
func assertRiderFolded(t *testing.T, ability Ability, wantKind EffectKind) {
	t.Helper()
	if len(ability.Sentences) != 2 {
		t.Fatalf("sentences = %d, want 2 (controller effect + rider)", len(ability.Sentences))
	}
	effects := ability.Sentences[0].Effects
	if len(effects) != 1 || effects[0].Kind != wantKind {
		t.Fatalf("first sentence effects = %#v, want a single %s", effects, wantKind)
	}
	effect := effects[0]
	if effect.EachOpponentAttackingSameRiderSpan == (shared.Span{}) {
		t.Error("EachOpponentAttackingSameRiderSpan is unset, want the rider sentence span")
	}
	if effect.HasUnrecognizedSibling {
		t.Error("HasUnrecognizedSibling = true, want false after crediting the rider")
	}
	if !effect.Exact {
		t.Error("Exact = false, want true for the credited effect")
	}
	if effect.RequiresOrderedLowering {
		t.Error("RequiresOrderedLowering = true, want false after folding the rider to one effect")
	}
	if !ability.Sentences[1].EachOpponentAttackingSameRider {
		t.Error("rider sentence EachOpponentAttackingSameRider = false, want true")
	}
	if len(ability.Sentences[1].Effects) != 0 {
		t.Errorf("rider sentence effects = %#v, want cleared", ability.Sentences[1].Effects)
	}
}

// TestCreditEachOpponentAttackingSameRiderAcceptsCreate verifies the trailing
// "Each opponent attacking that player does the same." sentence folds onto the
// lone controller create-token effect of an enchanted-player combat trigger
// (Curse of Opulence).
func TestCreditEachOpponentAttackingSameRiderAcceptsCreate(t *testing.T) {
	t.Parallel()
	ability := creditedRiderAbility(t,
		"Whenever enchanted player is attacked, create a Gold token. Each opponent attacking that player does the same.",
		"Curse of Opulence")
	assertRiderFolded(t, ability, EffectCreate)
}

// TestCreditEachOpponentAttackingSameRiderAcceptsGainLife verifies the anaphoric
// rider folds onto a lone controller gain-life effect (Curse of Vitality), the
// generalization of the create-token family to a life gain.
func TestCreditEachOpponentAttackingSameRiderAcceptsGainLife(t *testing.T) {
	t.Parallel()
	ability := creditedRiderAbility(t,
		"Whenever enchanted player is attacked, you gain 2 life. Each opponent attacking that player does the same.",
		"Curse of Vitality")
	assertRiderFolded(t, ability, EffectGain)
	if !ability.Sentences[0].Effects[0].LifeObject {
		t.Error("LifeObject = false, want true for a gain-life effect")
	}
}

// TestCreditEachOpponentAttackingSameRiderAcceptsDraw verifies the anaphoric
// rider folds onto a lone controller draw effect (Curse of Verbosity), the
// generalization of the create-token family to a card draw.
func TestCreditEachOpponentAttackingSameRiderAcceptsDraw(t *testing.T) {
	t.Parallel()
	ability := creditedRiderAbility(t,
		"Whenever enchanted player is attacked, you draw a card. Each opponent attacking that player does the same.",
		"Curse of Verbosity")
	assertRiderFolded(t, ability, EffectDraw)
}

// TestCreditEachOpponentAttackingUntapRiderAccepts verifies the explicit "Each
// opponent attacking that player untaps all nonland permanents they control."
// rider folds onto the lone controller untap of the same nonland group (Curse of
// Bounty). Unlike the anaphoric family the rider spells its action out, so the
// recognizer matches its parsed each-opponent untap against the controller untap.
func TestCreditEachOpponentAttackingUntapRiderAccepts(t *testing.T) {
	t.Parallel()
	ability := creditedRiderAbility(t,
		"Whenever enchanted player is attacked, untap all nonland permanents you control. Each opponent attacking that player untaps all nonland permanents they control.",
		"Curse of Bounty")
	assertRiderFolded(t, ability, EffectUntap)
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
	assertRiderUncredited(t, document)
}

// TestCreditEachOpponentAttackingSameRiderRequiresSupportedEffect verifies the
// anaphoric rider is not credited when the lone controller effect is not one the
// "does the same." anaphor can widen — a create-token, gain-life, or draw. A
// life-loss effect is a controller effect of an unsupported kind, so the rider
// stays uncredited and the ability fails closed rather than silently attaching a
// group effect to it.
func TestCreditEachOpponentAttackingSameRiderRequiresSupportedEffect(t *testing.T) {
	t.Parallel()
	document, _ := Parse(
		"Whenever enchanted player is attacked, you lose 2 life. Each opponent attacking that player does the same.",
		Context{CardName: "Not A Curse"})
	assertRiderUncredited(t, document)
}

// TestCreditEachOpponentAttackingUntapRiderRequiresMatchingGroup verifies the
// explicit untap rider is not credited when the rider untaps a different group
// than the controller untap. "All creatures you control" and "all nonland
// permanents they control" name different permanent sets, so the recognizer
// leaves the rider uncredited and the ability fails closed rather than mirroring
// an untap onto a mismatched group.
func TestCreditEachOpponentAttackingUntapRiderRequiresMatchingGroup(t *testing.T) {
	t.Parallel()
	document, _ := Parse(
		"Whenever enchanted player is attacked, untap all creatures you control. Each opponent attacking that player untaps all nonland permanents they control.",
		Context{CardName: "Not A Curse"})
	assertRiderUncredited(t, document)
}

// assertRiderUncredited asserts no sentence of any ability in the document was
// marked as a folded reflexive rider and no effect recorded a rider span, the
// shared fail-closed assertion for the near-miss rider tests.
func assertRiderUncredited(t *testing.T, document Document) {
	t.Helper()
	for _, ability := range document.Abilities {
		for _, sentence := range ability.Sentences {
			if sentence.EachOpponentAttackingSameRider {
				t.Fatal("rider credited on an unsupported shape, want fail closed")
			}
			for _, effect := range sentence.Effects {
				if effect.EachOpponentAttackingSameRiderSpan != (shared.Span{}) {
					t.Fatal("effect recorded a rider span on an unsupported shape, want fail closed")
				}
			}
		}
	}
}
