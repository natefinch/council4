package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
)

// selfEnterTriggerAbility is the self-enter "gain 1 life" triggered ability shared
// by the exclusion test doublers and sources.
func selfEnterTriggerAbility() game.TriggeredAbility {
	return game.TriggeredAbility{
		Trigger: game.TriggerCondition{Type: game.TriggerWhenever, Pattern: game.TriggerPattern{
			Event:       game.EventZoneChanged,
			Source:      game.TriggerSourceSelf,
			MatchToZone: true,
			ToZone:      zone.Battlefield,
		}},
		Content: game.Mode{Sequence: []game.Instruction{
			{Primitive: game.GainLife{Amount: game.Fixed(1), Player: game.ControllerReference()}},
		}}.Ability(),
	}
}

// excludingDoublerDef builds a controlled-permanent trigger doubler carrying the
// given filter that also has its own self-enter triggered ability, so a test can
// observe whether the doubler doubles its own trigger.
func excludingDoublerDef(name string, subtypes []types.Sub, filter game.Selection) *game.CardDef {
	return &game.CardDef{CardFace: game.CardFace{
		Name:     name,
		Types:    []types.Card{types.Creature},
		Subtypes: subtypes,
		StaticAbilities: []game.StaticAbility{{
			RuleEffects: []game.RuleEffect{{
				Kind:              game.RuleEffectAdditionalTriggerForControlledPermanent,
				AffectedSelection: filter,
			}},
		}},
		TriggeredAbilities: []game.TriggeredAbility{selfEnterTriggerAbility()},
	}}
}

// TestControlledPermanentRuleEffectExcludesSelfForAnother verifies "another ...
// you control" (ExcludeSource) does not double the doubler's own trigger but does
// double another matching permanent's trigger (Twinflame Travelers).
func TestControlledPermanentRuleEffectExcludesSelfForAnother(t *testing.T) {
	elementalAnother := game.Selection{
		SubtypesAny:   []types.Sub{types.Sub("Elemental")},
		ExcludeSource: true,
	}

	t.Run("own trigger is not doubled", func(t *testing.T) {
		g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
		engine := NewEngine(nil)
		doubler := addCombatPermanent(g, game.Player1, excludingDoublerDef(
			"Twinflame", []types.Sub{types.Sub("Elemental")}, elementalAnother))

		emitSelfEnter(g, doubler)
		if !engine.putTriggeredAbilitiesOnStack(g) {
			t.Fatal("controlled trigger was not put on the stack")
		}
		if got := g.Stack.Size(); got != 1 {
			t.Fatalf("stack size = %d, want 1 (the doubler does not double its own trigger)", got)
		}
	})

	t.Run("another matching trigger is doubled", func(t *testing.T) {
		g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
		engine := NewEngine(nil)
		addCombatPermanent(g, game.Player1, excludingDoublerDef(
			"Twinflame", []types.Sub{types.Sub("Elemental")}, elementalAnother))
		source := addCombatPermanent(g, game.Player1, selfEntersTypedTriggerSourceDef(
			"Other Elemental", nil, []types.Card{types.Creature}, []types.Sub{types.Sub("Elemental")}))

		emitSelfEnter(g, source)
		if !engine.putTriggeredAbilitiesOnStack(g) {
			t.Fatal("controlled trigger was not put on the stack")
		}
		if got := g.Stack.Size(); got != 2 {
			t.Fatalf("stack size = %d, want 2 (another Elemental's trigger is doubled)", got)
		}
	})
}

// TestControlledPermanentRuleEffectDisjunctionMixedExclusion verifies the
// "a Shaman or another Wizard you control" AnyOf filter (Harmonic Prodigy):
// the Shaman branch includes any Shaman, while the Wizard branch excludes the
// doubler itself.
func TestControlledPermanentRuleEffectDisjunctionMixedExclusion(t *testing.T) {
	harmonic := game.Selection{AnyOf: []game.Selection{
		{SubtypesAny: []types.Sub{types.Sub("Shaman")}},
		{SubtypesAny: []types.Sub{types.Sub("Wizard")}, ExcludeSource: true},
	}}

	t.Run("own wizard trigger excluded by the wizard branch", func(t *testing.T) {
		g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
		engine := NewEngine(nil)
		doubler := addCombatPermanent(g, game.Player1, excludingDoublerDef(
			"Harmonic Prodigy", []types.Sub{types.Sub("Wizard")}, harmonic))

		emitSelfEnter(g, doubler)
		if !engine.putTriggeredAbilitiesOnStack(g) {
			t.Fatal("controlled trigger was not put on the stack")
		}
		if got := g.Stack.Size(); got != 1 {
			t.Fatalf("stack size = %d, want 1 (a non-Shaman Wizard doubler does not double its own trigger)", got)
		}
	})

	t.Run("own shaman trigger included by the shaman branch", func(t *testing.T) {
		g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
		engine := NewEngine(nil)
		doubler := addCombatPermanent(g, game.Player1, excludingDoublerDef(
			"Shaman Doubler", []types.Sub{types.Sub("Shaman")}, harmonic))

		emitSelfEnter(g, doubler)
		if !engine.putTriggeredAbilitiesOnStack(g) {
			t.Fatal("controlled trigger was not put on the stack")
		}
		if got := g.Stack.Size(); got != 2 {
			t.Fatalf("stack size = %d, want 2 (the Shaman branch includes the doubler itself)", got)
		}
	})

	t.Run("another wizard trigger doubled by the wizard branch", func(t *testing.T) {
		g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
		engine := NewEngine(nil)
		addCombatPermanent(g, game.Player1, excludingDoublerDef(
			"Harmonic Prodigy", []types.Sub{types.Sub("Wizard")}, harmonic))
		source := addCombatPermanent(g, game.Player1, selfEntersTypedTriggerSourceDef(
			"Other Wizard", nil, []types.Card{types.Creature}, []types.Sub{types.Sub("Wizard")}))

		emitSelfEnter(g, source)
		if !engine.putTriggeredAbilitiesOnStack(g) {
			t.Fatal("controlled trigger was not put on the stack")
		}
		if got := g.Stack.Size(); got != 2 {
			t.Fatalf("stack size = %d, want 2 (another Wizard's trigger is doubled)", got)
		}
	})
}
