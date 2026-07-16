package cardgen

import (
	"strings"
	"testing"
)

// TestGenerateExecutableCardSourceHellkiteTyrant covers Hellkite Tyrant, the
// canonical event-player mass gain-control card: "Whenever this creature deals
// combat damage to a player, gain control of all artifacts that player controls."
// The combat-damage trigger lowers to a LayerControl continuous effect over a
// PlayerControlledGroup anchored on the triggering event's player (the damaged
// player), with the resolving controller as the new controller (the Player1
// sentinel) and a permanent duration. The permanent-duration effect snapshots the
// damaged player's artifacts at resolution, so the flying/trample body and the
// separate upkeep win condition round-trip unchanged alongside it.
func TestGenerateExecutableCardSourceHellkiteTyrant(t *testing.T) {
	t.Parallel()
	card := &ScryfallCard{
		Name:       "Hellkite Tyrant",
		Layout:     "normal",
		TypeLine:   "Creature — Dragon",
		ManaCost:   "{4}{R}{R}",
		OracleText: "Flying, trample\nWhenever Hellkite Tyrant deals combat damage to a player, gain control of all artifacts that player controls.\nAt the beginning of your upkeep, if you control twenty or more artifacts, you win the game.",
	}
	source, diagnostics, err := GenerateExecutableCardSource(card, "h")
	if err != nil {
		t.Fatal(err)
	}
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	for _, wanted := range []string{
		"Event:               game.EventDamageDealt",
		"RequireCombatDamage: true",
		"DamageRecipient:     game.DamageRecipientPlayer",
		"Primitive: game.ApplyContinuous{",
		"Layer:         game.LayerControl,",
		"NewController: opt.Val(game.Player1),",
		"Group:         game.PlayerControlledGroup(game.EventPlayerReference(), game.Selection{RequiredTypes: []types.Card{types.Artifact}}),",
		"Duration: game.DurationPermanent,",
		"Primitive: game.PlayerWinsGame{",
	} {
		if !strings.Contains(source, wanted) {
			t.Fatalf("source missing %q:\n%s", wanted, source)
		}
	}
}

// TestGenerateExecutableCardSourceAuraThief covers the plain-selector mass
// gain-control form: Aura Thief's "When this creature dies, you gain control of
// all enchantments." The unqualified "all enchantments" group carries no
// controller relationship, so it lowers to a battlefield-wide LayerControl over a
// BattlefieldGroup of enchantments with no target, distinct from the
// player-anchored forms.
func TestGenerateExecutableCardSourceAuraThief(t *testing.T) {
	t.Parallel()
	card := &ScryfallCard{
		Name:       "Aura Thief",
		Layout:     "normal",
		TypeLine:   "Creature — Illusion",
		ManaCost:   "{2}{U}{U}",
		OracleText: "Flying\nWhen this creature dies, you gain control of all enchantments. (You don't get to move Auras.)",
	}
	source, diagnostics, err := GenerateExecutableCardSource(card, "a")
	if err != nil {
		t.Fatal(err)
	}
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	for _, wanted := range []string{
		"Layer:         game.LayerControl,",
		"NewController: opt.Val(game.Player1),",
		"Group:         game.BattlefieldGroup(game.Selection{RequiredTypes: []types.Card{types.Enchantment}}),",
		"Duration: game.DurationPermanent,",
	} {
		if !strings.Contains(source, wanted) {
			t.Fatalf("source missing %q:\n%s", wanted, source)
		}
	}
}

// TestGenerateExecutableCardSourceTargetOpponentMassControl covers the
// targeted-player mass gain-control form ("Gain control of all creatures target
// opponent controls", Ashiok, Sculptor of Fears' ultimate). The targeted opponent
// supplies the group's controller relationship through TargetPlayerReference(0),
// and the card gains a single player target constrained to an opponent.
func TestGenerateExecutableCardSourceTargetOpponentMassControl(t *testing.T) {
	t.Parallel()
	card := &ScryfallCard{
		Name:       "Target Opponent Mass Control Probe",
		Layout:     "normal",
		TypeLine:   "Sorcery",
		ManaCost:   "{3}{U}{B}",
		OracleText: "Gain control of all creatures target opponent controls.",
	}
	source, diagnostics, err := GenerateExecutableCardSource(card, "t")
	if err != nil {
		t.Fatal(err)
	}
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	for _, wanted := range []string{
		"Constraint: \"target opponent\",",
		"Allow:      game.TargetAllowPlayer,",
		"Selection:  opt.Val(game.Selection{Player: game.PlayerOpponent}),",
		"Layer:         game.LayerControl,",
		"NewController: opt.Val(game.Player1),",
		"Group:         game.PlayerControlledGroup(game.TargetPlayerReference(0), game.Selection{RequiredTypes: []types.Card{types.Creature}}),",
		"Duration: game.DurationPermanent,",
	} {
		if !strings.Contains(source, wanted) {
			t.Fatalf("source missing %q:\n%s", wanted, source)
		}
	}
}

// TestGenerateExecutableCardSourceMassControlEventPlayerSpellFailsClosed asserts
// the event-player mass gain-control path stays scoped to triggers whose event
// binds "that player". In a spell body there is no triggering event, so the
// compiler never binds "that player" to ReferenceBindingEventPlayer;
// eventPlayerControlledMassGroup rejects the surviving reference and, with no
// selector controller and no player target, the card fails closed rather than
// resolving the group against a nonexistent event player.
func TestGenerateExecutableCardSourceMassControlEventPlayerSpellFailsClosed(t *testing.T) {
	t.Parallel()
	card := &ScryfallCard{
		Name:       "Mass Control Event Player Probe",
		Layout:     "normal",
		TypeLine:   "Sorcery",
		OracleText: "Gain control of all artifacts that player controls.",
	}
	_, diagnostics, err := GenerateExecutableCardSource(card, "m")
	if err != nil {
		t.Fatal(err)
	}
	if len(diagnostics) == 0 {
		t.Fatal("expected an unsupported diagnostic for a mass gain-control spell with an unbound event player, got none")
	}
}

// TestGenerateExecutableCardSourceMassControlUntapRiderFailsClosed asserts the
// mass gain-control path does not silently drop a composed rider. Karrthus,
// Tyrant of Jund's "gain control of all Dragons, then untap all Dragons" is an
// ordered two-effect sequence whose second effect the single-control lowering
// cannot represent, so the whole ability fails closed rather than lowering a
// partial control-only effect.
func TestGenerateExecutableCardSourceMassControlUntapRiderFailsClosed(t *testing.T) {
	t.Parallel()
	card := &ScryfallCard{
		Name:       "Mass Control Untap Rider Probe",
		Layout:     "normal",
		TypeLine:   "Sorcery",
		OracleText: "Gain control of all Dragons, then untap all Dragons.",
	}
	_, diagnostics, err := GenerateExecutableCardSource(card, "u")
	if err != nil {
		t.Fatal(err)
	}
	if len(diagnostics) == 0 {
		t.Fatal("expected an unsupported diagnostic for a mass gain-control with an untap rider, got none")
	}
}
