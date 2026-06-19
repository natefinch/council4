package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/opt"
)

// TestBuffThenUntapSequenceResolvesInOrder proves that an ordered spell sequence
// whose first instruction buffs a target creature and whose second instruction
// untaps that same target ("Target creature gets +2/+2 until end of turn. Untap
// that creature.") applies both instructions to the shared target in order: the
// creature ends the resolution both pumped and untapped.
func TestBuffThenUntapSequenceResolvesInOrder(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	creature := addCombatCreaturePermanentWithPower(g, game.Player1, 2)
	creature.Tapped = true

	addInstructionSpellToStackForController(g, game.Player1, []game.Instruction{
		{Primitive: game.ModifyPT{
			Object:         game.TargetPermanentReference(0),
			PowerDelta:     game.Fixed(2),
			ToughnessDelta: game.Fixed(2),
			Duration:       game.DurationUntilEndOfTurn,
		}},
		{Primitive: game.Untap{Object: game.TargetPermanentReference(0)}},
	}, []game.Target{game.PermanentTarget(creature.ObjectID)})

	engine.resolveTopOfStack(g, &TurnLog{})

	if got := effectivePower(g, creature); got != 4 {
		t.Fatalf("effective power = %d, want 4 (2 base + 2 buff)", got)
	}
	if creature.Tapped {
		t.Fatal("second sequence instruction did not untap the buffed creature")
	}
}

// TestMultiTargetCombinedBuffSequenceBuffsEachChosenTarget proves that the
// multi-target combined buff "Up to two target creatures each get +1/+1 and gain
// lifelink until end of turn." applies an independent power/toughness and
// keyword grant to each chosen target slot.
func TestMultiTargetCombinedBuffSequenceBuffsEachChosenTarget(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	first := addCombatCreaturePermanentWithPower(g, game.Player1, 2)
	second := addCombatCreaturePermanentWithPower(g, game.Player1, 3)

	buffSlot := func(index int) game.Instruction {
		return game.Instruction{Primitive: game.ApplyContinuous{
			Object: opt.Val(game.TargetPermanentReference(index)),
			ContinuousEffects: []game.ContinuousEffect{
				{Layer: game.LayerPowerToughnessModify, PowerDelta: 1, ToughnessDelta: 1},
				{Layer: game.LayerAbility, AddKeywords: []game.Keyword{game.Lifelink}},
			},
			Duration: game.DurationUntilEndOfTurn,
		}}
	}

	addInstructionSpellToStackForController(g, game.Player1,
		[]game.Instruction{buffSlot(0), buffSlot(1)},
		[]game.Target{
			game.PermanentTarget(first.ObjectID),
			game.PermanentTarget(second.ObjectID),
		},
	)

	engine.resolveTopOfStack(g, &TurnLog{})

	if got := effectivePower(g, first); got != 3 {
		t.Fatalf("first creature effective power = %d, want 3", got)
	}
	if got := effectivePower(g, second); got != 4 {
		t.Fatalf("second creature effective power = %d, want 4", got)
	}
	if !hasKeyword(g, first, game.Lifelink) {
		t.Fatal("first creature did not gain lifelink")
	}
	if !hasKeyword(g, second, game.Lifelink) {
		t.Fatal("second creature did not gain lifelink")
	}
}

// TestMultiInstructionClauseThenLifeGainResolvesInOrder proves that a spell
// whose leading clause lowers to more than one instruction — "Up to two target
// creatures each get +1/+2 until end of turn." expands to one ModifyPT per
// target — still resolves the trailing independent life-gain instruction. Both
// chosen creatures end pumped and the controller gains the 2 life (Tandem
// Tactics shape).
func TestMultiInstructionClauseThenLifeGainResolvesInOrder(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	first := addCombatCreaturePermanentWithPower(g, game.Player1, 2)
	second := addCombatCreaturePermanentWithPower(g, game.Player1, 3)
	startLife := g.Players[game.Player1].Life

	addInstructionSpellToStackForController(g, game.Player1, []game.Instruction{
		{Primitive: game.ModifyPT{
			Object:         game.TargetPermanentReference(0),
			PowerDelta:     game.Fixed(1),
			ToughnessDelta: game.Fixed(2),
			Duration:       game.DurationUntilEndOfTurn,
		}},
		{Primitive: game.ModifyPT{
			Object:         game.TargetPermanentReference(1),
			PowerDelta:     game.Fixed(1),
			ToughnessDelta: game.Fixed(2),
			Duration:       game.DurationUntilEndOfTurn,
		}},
		{Primitive: game.GainLife{Amount: game.Fixed(2), Player: game.ControllerReference()}},
	}, []game.Target{
		game.PermanentTarget(first.ObjectID),
		game.PermanentTarget(second.ObjectID),
	})

	engine.resolveTopOfStack(g, &TurnLog{})

	if got := effectivePower(g, first); got != 3 {
		t.Fatalf("first creature effective power = %d, want 3 (2 base + 1 buff)", got)
	}
	if got := effectivePower(g, second); got != 4 {
		t.Fatalf("second creature effective power = %d, want 4 (3 base + 1 buff)", got)
	}
	if got := g.Players[game.Player1].Life; got != startLife+2 {
		t.Fatalf("controller life = %d, want %d (gained 2)", got, startLife+2)
	}
}
