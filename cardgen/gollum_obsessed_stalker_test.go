package cardgen

import (
	"testing"

	"github.com/natefinch/council4/cardgen/oracle/compiler"
	"github.com/natefinch/council4/cardgen/oracle/parser"
	"github.com/natefinch/council4/mtg/game"
)

// TestGollumStalkerSemanticShape verifies the parser recognizes the "each
// opponent dealt combat damage this game by a creature named <Name>" recipient,
// preserves the creature name, and marks the effect exact so it is covered.
func TestGollumStalkerSemanticShape(t *testing.T) {
	t.Parallel()
	compilation, diagnostics := compileTestOracle(
		"At the beginning of your end step, each opponent dealt combat damage this game by a creature named Gollum, Obsessed Stalker loses life equal to the amount of life you gained this turn.",
		parser.Context{CardName: "Gollum, Obsessed Stalker"},
		compiler.Context{},
	)
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	if len(compilation.Abilities) != 1 {
		t.Fatalf("abilities = %#v", compilation.Abilities)
	}
	effects := compilation.Abilities[0].Content.Effects
	if len(effects) != 1 {
		t.Fatalf("effects = %#v", effects)
	}
	effect := effects[0]
	if effect.Context != parser.EffectContextEachOpponentDealtCombatDamageByNamed {
		t.Fatalf("context = %q, want EffectContextEachOpponentDealtCombatDamageByNamed", effect.Context)
	}
	if effect.CombatDamageSourceName != "Gollum, Obsessed Stalker" {
		t.Fatalf("combat-damage source name = %q, want %q", effect.CombatDamageSourceName, "Gollum, Obsessed Stalker")
	}
	if !effect.Exact {
		t.Fatal("effect must be exact so the source is covered")
	}
}

// TestGollumStalkerDynamicLifeLoss verifies Gollum's real end-step trigger lowers
// to a group life loss routed to the opponents dealt combat damage this game by a
// creature named Gollum, Obsessed Stalker, for the amount of life gained this turn.
func TestGollumStalkerDynamicLifeLoss(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Gollum, Obsessed Stalker",
		Layout:     "normal",
		TypeLine:   "Legendary Creature — Halfling Horror",
		Power:      new("1"),
		Toughness:  new("1"),
		OracleText: "Skulk (This creature can't be blocked by creatures with greater power.)\nAt the beginning of your end step, each opponent dealt combat damage this game by a creature named Gollum, Obsessed Stalker loses life equal to the amount of life you gained this turn.",
	})
	if len(face.TriggeredAbilities) != 1 {
		t.Fatalf("triggered abilities = %d, want 1", len(face.TriggeredAbilities))
	}
	ability := face.TriggeredAbilities[0]
	if ability.Trigger.Pattern.Event != game.EventBeginningOfStep ||
		ability.Trigger.Pattern.Step != game.StepEnd ||
		ability.Trigger.Pattern.Controller != game.TriggerControllerYou {
		t.Fatalf("trigger pattern = %#v, want your end step", ability.Trigger.Pattern)
	}
	mode := ability.Content.Modes[0]
	if len(mode.Sequence) != 1 {
		t.Fatalf("sequence len = %d, want 1", len(mode.Sequence))
	}
	loseLife, ok := mode.Sequence[0].Primitive.(game.LoseLife)
	if !ok {
		t.Fatalf("sequence[0] = %T, want game.LoseLife", mode.Sequence[0].Primitive)
	}
	if loseLife.PlayerGroup.Kind != game.PlayerGroupReferenceOpponentsDealtCombatDamageThisGameByNamed {
		t.Fatalf("player group kind = %v, want OpponentsDealtCombatDamageThisGameByNamed", loseLife.PlayerGroup.Kind)
	}
	if loseLife.PlayerGroup.Name != "Gollum, Obsessed Stalker" {
		t.Fatalf("player group name = %q, want %q", loseLife.PlayerGroup.Name, "Gollum, Obsessed Stalker")
	}
	dynamic := loseLife.Amount.DynamicAmount()
	if !dynamic.Exists || dynamic.Val.Kind != game.DynamicAmountLifeGainedThisTurn {
		t.Fatalf("lose-life amount = %#v, want DynamicAmountLifeGainedThisTurn", loseLife.Amount)
	}
}

// TestGollumStalkerFixedLifeLoss verifies the same group recipient carries a fixed
// life amount, proving the recipient is a broad, typed capability rather than a
// Gollum-only lowerer.
func TestGollumStalkerFixedLifeLoss(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Gollum, Obsessed Stalker",
		Layout:     "normal",
		TypeLine:   "Legendary Creature — Halfling Horror",
		Power:      new("1"),
		Toughness:  new("1"),
		OracleText: "At the beginning of your end step, each opponent dealt combat damage this game by a creature named Gollum, Obsessed Stalker loses 2 life.",
	})
	if len(face.TriggeredAbilities) != 1 {
		t.Fatalf("triggered abilities = %d, want 1", len(face.TriggeredAbilities))
	}
	prim := face.TriggeredAbilities[0].Content.Modes[0].Sequence[0].Primitive
	loseLife, ok := prim.(game.LoseLife)
	if !ok {
		t.Fatalf("sequence[0] = %T, want game.LoseLife", prim)
	}
	if loseLife.PlayerGroup.Kind != game.PlayerGroupReferenceOpponentsDealtCombatDamageThisGameByNamed ||
		loseLife.PlayerGroup.Name != "Gollum, Obsessed Stalker" {
		t.Fatalf("player group = %#v, want named combat-damage group for Gollum", loseLife.PlayerGroup)
	}
	if loseLife.Amount.IsDynamic() || loseLife.Amount.Value() != 2 {
		t.Fatalf("lose-life amount = %#v, want fixed 2", loseLife.Amount)
	}
}
