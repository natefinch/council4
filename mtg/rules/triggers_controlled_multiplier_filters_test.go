package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/compare"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// selfEntersCreatureWithPower is a vanilla creature of the given base power whose
// self-enter "gain 1 life" trigger a power-bounded doubler may multiply.
func selfEntersCreatureWithPower(name string, power int) *game.CardDef {
	return &game.CardDef{CardFace: game.CardFace{
		Name:               name,
		Types:              []types.Card{types.Creature},
		Power:              opt.Val(game.PT{Value: power}),
		Toughness:          opt.Val(game.PT{Value: power}),
		TriggeredAbilities: []game.TriggeredAbility{selfEnterTriggerAbility()},
	}}
}

// powerFilter is Delney, Streetwise Lookout's source filter: "a creature you
// control with power 2 or less". It carries no ExcludeSource, because the Oracle
// text says "a creature" (not "another"), so the doubler doubles its own
// controlled triggers too.
func powerFilter() game.Selection {
	return game.Selection{
		RequiredTypes: []types.Card{types.Creature},
		Power:         opt.Val(compare.Int{Op: compare.LessOrEqual, Value: 2}),
	}
}

// TestControlledPermanentRuleEffectPowerFilter verifies the power-bounded source
// filter (Delney, Streetwise Lookout: "a creature you control with power 2 or
// less"). A creature with effective power 2 or less has its trigger doubled; a
// power-3 creature does not.
func TestControlledPermanentRuleEffectPowerFilter(t *testing.T) {
	for name, tc := range map[string]struct {
		power     int
		wantStack int
	}{
		"power 2 is doubled":     {power: 2, wantStack: 2},
		"power 1 is doubled":     {power: 1, wantStack: 2},
		"power 3 is not doubled": {power: 3, wantStack: 1},
	} {
		t.Run(name, func(t *testing.T) {
			g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
			engine := NewEngine(nil)
			addCombatPermanent(g, game.Player1, controlledTriggerDoublerDef(powerFilter()))
			source := addCombatPermanent(g, game.Player1, selfEntersCreatureWithPower("Power Source", tc.power))

			emitSelfEnter(g, source)
			if !engine.putTriggeredAbilitiesOnStack(g) {
				t.Fatal("controlled trigger was not put on the stack")
			}
			if got := g.Stack.Size(); got != tc.wantStack {
				t.Fatalf("stack size = %d, want %d", got, tc.wantStack)
			}
		})
	}
}

// TestControlledPermanentRuleEffectPowerFilterDoublesOwnTrigger verifies Delney's
// self-inclusion: Delney is itself a power-2 creature and its filter reads "a
// creature" (no "another"), so it doubles its own triggered abilities.
func TestControlledPermanentRuleEffectPowerFilterDoublesOwnTrigger(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	def := selfEntersCreatureWithPower("Delney", 2)
	def.StaticAbilities = []game.StaticAbility{{
		RuleEffects: []game.RuleEffect{{
			Kind:              game.RuleEffectAdditionalTriggerForControlledPermanent,
			AffectedSelection: powerFilter(),
		}},
	}}
	source := addCombatPermanent(g, game.Player1, def)

	emitSelfEnter(g, source)
	if !engine.putTriggeredAbilitiesOnStack(g) {
		t.Fatal("controlled trigger was not put on the stack")
	}
	if got := g.Stack.Size(); got != 2 {
		t.Fatalf("stack size = %d, want 2 (a power-2 doubler doubles its own trigger)", got)
	}
}

// castOrCopyTriggerSourceDef is a creature with a "whenever <controller> casts or
// copies a spell, gain 1 life" trigger carrying no card-type filter, letting a
// test drive the magecraft doubler's causal gate purely from the emitted event's
// controller and card types.
func castOrCopyTriggerSourceDef(name string, controller game.TriggerControllerFilter) *game.CardDef {
	return &game.CardDef{CardFace: game.CardFace{
		Name:  name,
		Types: []types.Card{types.Creature},
		TriggeredAbilities: []game.TriggeredAbility{{
			Trigger: game.TriggerCondition{Type: game.TriggerWhenever, Pattern: game.TriggerPattern{
				Event:          game.EventSpellCast,
				Controller:     controller,
				MatchSpellCopy: true,
			}},
			Content: game.Mode{Sequence: []game.Instruction{
				{Primitive: game.GainLife{Amount: game.Fixed(1), Player: game.ControllerReference()}},
			}}.Ability(),
		}},
	}}
}

// magecraftDoublerDef is Veyran, Voice of Duality's doubler: "If you casting or
// copying an instant or sorcery spell causes a triggered ability of a permanent
// you control to trigger, that ability triggers an additional time." Its empty
// selection matches any permanent you control (Veyran included).
func magecraftDoublerDef() *game.CardDef {
	return &game.CardDef{CardFace: game.CardFace{
		Name:  "Magecraft Doubler",
		Types: []types.Card{types.Enchantment},
		StaticAbilities: []game.StaticAbility{{
			RuleEffects: []game.RuleEffect{{
				Kind:                                 game.RuleEffectAdditionalTriggerForControlledPermanent,
				TriggerCauseCastOrCopyInstantSorcery: true,
			}},
		}},
	}}
}

func emitControllerSpellCast(g *game.Game, kind game.EventKind, controller game.PlayerID, cardType types.Card) {
	emitEvent(g, game.Event{
		Kind:       kind,
		Controller: controller,
		Player:     controller,
		CardTypes:  []types.Card{cardType},
	})
}

// TestControlledPermanentRuleEffectMagecraftCause verifies Veyran's causal gate:
// only a triggered ability caused by you casting or copying an instant or sorcery
// is doubled. Non-cast causes, casts of other spell types, and opponents' casts
// are ignored, and the doubler doubles its own magecraft trigger (self-inclusion).
func TestControlledPermanentRuleEffectMagecraftCause(t *testing.T) {
	t.Run("doubles a trigger caused by casting an instant", func(t *testing.T) {
		g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
		engine := NewEngine(nil)
		addCombatPermanent(g, game.Player1, magecraftDoublerDef())
		addCombatPermanent(g, game.Player1, castOrCopyTriggerSourceDef("Caster", game.TriggerControllerYou))

		emitControllerSpellCast(g, game.EventSpellCast, game.Player1, types.Instant)
		if !engine.putTriggeredAbilitiesOnStack(g) {
			t.Fatal("cast trigger was not put on the stack")
		}
		if got := g.Stack.Size(); got != 2 {
			t.Fatalf("stack size = %d, want 2 (instant-cast cause is doubled)", got)
		}
	})

	t.Run("doubles a trigger caused by copying a sorcery", func(t *testing.T) {
		g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
		engine := NewEngine(nil)
		addCombatPermanent(g, game.Player1, magecraftDoublerDef())
		addCombatPermanent(g, game.Player1, castOrCopyTriggerSourceDef("Caster", game.TriggerControllerYou))

		emitControllerSpellCast(g, game.EventSpellCopied, game.Player1, types.Sorcery)
		if !engine.putTriggeredAbilitiesOnStack(g) {
			t.Fatal("copy trigger was not put on the stack")
		}
		if got := g.Stack.Size(); got != 2 {
			t.Fatalf("stack size = %d, want 2 (sorcery-copy cause is doubled)", got)
		}
	})

	t.Run("ignores a cast of a noninstant nonsorcery spell", func(t *testing.T) {
		g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
		engine := NewEngine(nil)
		addCombatPermanent(g, game.Player1, magecraftDoublerDef())
		addCombatPermanent(g, game.Player1, castOrCopyTriggerSourceDef("Caster", game.TriggerControllerYou))

		emitControllerSpellCast(g, game.EventSpellCast, game.Player1, types.Artifact)
		if !engine.putTriggeredAbilitiesOnStack(g) {
			t.Fatal("cast trigger was not put on the stack")
		}
		if got := g.Stack.Size(); got != 1 {
			t.Fatalf("stack size = %d, want 1 (an artifact cast is not an instant or sorcery)", got)
		}
	})

	t.Run("ignores a non-cast cause", func(t *testing.T) {
		g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
		engine := NewEngine(nil)
		addCombatPermanent(g, game.Player1, magecraftDoublerDef())
		source := addCombatPermanent(g, game.Player1, selfEntersTypedTriggerSourceDef(
			"ETB Source", nil, []types.Card{types.Creature}, nil))

		emitSelfEnter(g, source)
		if !engine.putTriggeredAbilitiesOnStack(g) {
			t.Fatal("controlled trigger was not put on the stack")
		}
		if got := g.Stack.Size(); got != 1 {
			t.Fatalf("stack size = %d, want 1 (an enters cause is not a spell cast)", got)
		}
	})

	t.Run("ignores an instant cast by an opponent", func(t *testing.T) {
		g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
		engine := NewEngine(nil)
		addCombatPermanent(g, game.Player1, magecraftDoublerDef())
		addCombatPermanent(g, game.Player1, castOrCopyTriggerSourceDef("Opponent Watcher", game.TriggerControllerOpponent))

		emitControllerSpellCast(g, game.EventSpellCast, game.Player2, types.Instant)
		if !engine.putTriggeredAbilitiesOnStack(g) {
			t.Fatal("opponent-cast trigger was not put on the stack")
		}
		if got := g.Stack.Size(); got != 1 {
			t.Fatalf("stack size = %d, want 1 (only your own casts double, not an opponent's)", got)
		}
	})

	t.Run("doubles the doubler's own magecraft trigger", func(t *testing.T) {
		g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
		engine := NewEngine(nil)
		def := castOrCopyTriggerSourceDef("Veyran", game.TriggerControllerYou)
		def.StaticAbilities = []game.StaticAbility{{
			RuleEffects: []game.RuleEffect{{
				Kind:                                 game.RuleEffectAdditionalTriggerForControlledPermanent,
				TriggerCauseCastOrCopyInstantSorcery: true,
			}},
		}}
		addCombatPermanent(g, game.Player1, def)

		emitControllerSpellCast(g, game.EventSpellCast, game.Player1, types.Instant)
		if !engine.putTriggeredAbilitiesOnStack(g) {
			t.Fatal("own magecraft trigger was not put on the stack")
		}
		if got := g.Stack.Size(); got != 2 {
			t.Fatalf("stack size = %d, want 2 (the doubler doubles its own magecraft trigger)", got)
		}
	})
}
