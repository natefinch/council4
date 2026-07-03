package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// TestCantBlockThisTurnRuleEffectProhibitsBlocking models the runtime behavior of
// the temporary "<targets> can't block this turn." resolving effect: an
// unconditional RuleEffectCantBlock scoped to a single affected creature (as the
// ApplyRule lowering produces, one per target) stops that creature from blocking
// any attacker, while an unaffected creature blocks normally.
func TestCantBlockThisTurnRuleEffectProhibitsBlocking(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	restricted := addCombatCreaturePermanentWithPower(g, game.Player2, 2)
	free := addCombatCreaturePermanentWithPower(g, game.Player2, 2)

	g.RuleEffects = append(g.RuleEffects, game.RuleEffect{
		ID:               g.IDGen.Next(),
		Kind:             game.RuleEffectCantBlock,
		Controller:       game.Player1,
		AffectedObjectID: restricted.ObjectID,
		Duration:         game.DurationThisTurn,
		CreatedTurn:      g.Turn.TurnNumber,
	})

	if canBlockWith(g, restricted, game.Player2) {
		t.Fatal("can't-block-this-turn rule effect let the affected creature block")
	}
	if !canBlockWith(g, free, game.Player2) {
		t.Fatal("can't-block-this-turn rule effect stopped an unaffected creature from blocking")
	}
}

// TestCantBlockThisTurnMultiTargetUnfilledSlotDoesNotRestrictOthers covers the
// "Up to N target creatures can't block this turn." multi-target lowering, which
// emits one ApplyRule per target slot. When the controller chooses fewer targets
// than the maximum, the declined slots must no-op: an unresolved object-scoped
// can't-block rule effect must never apply to every creature on the battlefield.
func TestCantBlockThisTurnMultiTargetUnfilledSlotDoesNotRestrictOthers(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	chosen := addCombatCreaturePermanentWithPower(g, game.Player2, 2)
	bystander := addCombatCreaturePermanentWithPower(g, game.Player2, 2)
	obj := &game.StackObject{
		Controller: game.Player1,
		Targets:    []game.Target{game.PermanentTarget(chosen.ObjectID)},
	}

	// The lowering for "Up to three target creatures can't block this turn."
	// emits ApplyRule for slots 0, 1, and 2; here only slot 0 was chosen.
	for i := range 3 {
		resolveInstruction(engine, g, obj, game.ApplyRule{
			Object: opt.Val(game.TargetPermanentReference(i)),
			RuleEffects: []game.RuleEffect{
				{Kind: game.RuleEffectCantBlock},
			},
			Duration: game.DurationThisTurn,
		}, nil)
	}

	if canBlockWith(g, chosen, game.Player2) {
		t.Fatal("chosen target was not restricted from blocking")
	}
	if !canBlockWith(g, bystander, game.Player2) {
		t.Fatal("an unfilled target slot wrongly restricted a non-targeted creature from blocking")
	}
}

// TestGroupCantBlockThisTurnControllerScope models the runtime behavior of the
// group-scoped "Creatures your opponents control can't block this turn."
// (Cosmotronic Wave, Hazardous Blast) resolving effect: an object-less
// RuleEffectCantBlock scoped by ControllerOpponent stops every creature an
// opponent of the caster controls from blocking, while the caster's own
// creatures block normally.
func TestGroupCantBlockThisTurnControllerScope(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	opponentCreature := addCombatCreaturePermanentWithPower(g, game.Player2, 2)
	casterCreature := addCombatCreaturePermanentWithPower(g, game.Player1, 2)
	obj := &game.StackObject{Controller: game.Player1}

	resolveInstruction(engine, g, obj, game.ApplyRule{
		RuleEffects: []game.RuleEffect{{
			Kind:               game.RuleEffectCantBlock,
			AffectedController: game.ControllerOpponent,
			PermanentTypes:     []types.Card{types.Creature},
		}},
		Duration: game.DurationThisTurn,
	}, nil)

	if canBlockWith(g, opponentCreature, game.Player2) {
		t.Fatal("group can't-block let an opponent-controlled creature block")
	}
	if !canBlockWith(g, casterCreature, game.Player1) {
		t.Fatal("group can't-block wrongly restricted the caster's own creature")
	}
}

// TestGroupCantBlockThisTurnKeywordFilter models the runtime behavior of the
// keyword-filtered group spell "Creatures without flying can't block this turn."
// (Falter, Magmatic Chasm, Seismic Stomp): an object-less RuleEffectCantBlock
// carrying an ExcludedKeyword: Flying affected Selection stops every non-flying
// creature from blocking while flying creatures block normally.
func TestGroupCantBlockThisTurnKeywordFilter(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	grounded := addCombatCreaturePermanentWithPower(g, game.Player2, 2)
	flyer := addCombatCreaturePermanentWithPower(g, game.Player2, 2, game.Flying)
	obj := &game.StackObject{Controller: game.Player1}

	resolveInstruction(engine, g, obj, game.ApplyRule{
		RuleEffects: []game.RuleEffect{{
			Kind:              game.RuleEffectCantBlock,
			PermanentTypes:    []types.Card{types.Creature},
			AffectedSelection: game.Selection{ExcludedKeyword: game.Flying},
		}},
		Duration: game.DurationThisTurn,
	}, nil)

	if canBlockWith(g, grounded, game.Player2) {
		t.Fatal("group can't-block let a non-flying creature block")
	}
	if !canBlockWith(g, flyer, game.Player2) {
		t.Fatal("group can't-block wrongly restricted a flying creature")
	}
}
