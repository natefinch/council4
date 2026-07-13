package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/compare"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// boromirTrigger builds Boromir, Warden of the Tower's cast trigger: an opponent
// cast trigger gated by "if no mana was spent to cast it" whose body counters the
// triggering stack spell through the event stack reference. The source belongs to
// Player1; the countered spell to Player2 (an opponent). manaSpent seeds the
// triggering event's recorded actual mana spend (0 for free/zero-mana, alternative
// {0}, and Convoke/Improvise casts paid entirely without mana). When uncounterable
// is set the spell carries a can't-be-countered rule effect. It returns the source
// permanent, the triggering spell's stack object, and its owning card id.
func boromirTrigger(g *game.Game, manaSpent int, uncounterable bool) (*game.Permanent, *game.StackObject, game.PlayerID) {
	source := addCreaturePermanent(g, game.Player1)

	spellSourceID := g.IDGen.Next()
	g.CardInstances[spellSourceID] = &game.CardInstance{
		ID: spellSourceID,
		Def: &game.CardDef{CardFace: game.CardFace{
			Name:  "Opponent Spell",
			Types: []types.Card{types.Instant},
		}},
		Owner: game.Player2,
	}
	castSpell := &game.StackObject{
		ID:         g.IDGen.Next(),
		Kind:       game.StackSpell,
		SourceID:   spellSourceID,
		Controller: game.Player2,
	}
	if uncounterable {
		castSpell.RuleEffects = []game.RuleEffect{{Kind: game.RuleEffectCantBeCountered}}
	}
	g.Stack.Push(castSpell)

	trigger := game.TriggeredAbility{
		Trigger: game.TriggerCondition{
			Type: game.TriggerWhenever,
			Pattern: game.TriggerPattern{
				Event:      game.EventSpellCast,
				Controller: game.TriggerControllerOpponent,
			},
			InterveningCondition: opt.Val(game.Condition{
				Aggregates: []game.AggregateComparison{{
					Aggregate: game.AggregateEventSpellManaSpentToCast,
					Op:        compare.LessOrEqual,
					Value:     0,
				}},
			}),
		},
		Content: game.Mode{Sequence: []game.Instruction{{
			Primitive: game.CounterObject{Object: game.EventStackObjectReference()},
		}}}.Ability(),
	}
	ability := &game.StackObject{
		ID:              g.IDGen.Next(),
		Kind:            game.StackTriggeredAbility,
		SourceID:        source.ObjectID,
		SourceCardID:    source.CardInstanceID,
		Controller:      game.Player1,
		InlineTrigger:   &trigger,
		HasTriggerEvent: true,
		TriggerEvent: game.Event{
			Kind:            game.EventSpellCast,
			Controller:      game.Player2,
			StackObjectID:   castSpell.ID,
			ManaSpentToCast: opt.Val(manaSpent),
		},
	}
	g.Stack.Push(ability)
	return source, castSpell, game.Player2
}

// TestBoromirCountersOpponentZeroManaSpell proves the trigger counters an
// opponent's spell cast with no mana spent, putting it into its owner's
// graveyard. This is the free/zero-mana, alternative {0}, and Convoke/Improvise
// zero-mana path, all of which record ManaSpentToCast == 0.
func TestBoromirCountersOpponentZeroManaSpell(t *testing.T) {
	t.Parallel()
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	_, castSpell, owner := boromirTrigger(g, 0, false)
	spellCardID := castSpell.SourceID

	engine.resolveTopOfStack(g, &TurnLog{})

	if _, ok := stackObjectByID(g, castSpell.ID); ok {
		t.Fatal("zero-mana opponent spell was not countered")
	}
	if !g.Players[owner].Graveyard.Contains(spellCardID) {
		t.Fatal("countered spell was not put into its owner's graveyard")
	}
}

// TestBoromirDoesNotCounterManaSpentSpell proves the intervening condition gates
// the counter: an opponent spell cast with mana spent is not countered when the
// trigger resolves (CR 603.4 re-check), so it stays on the stack.
func TestBoromirDoesNotCounterManaSpentSpell(t *testing.T) {
	t.Parallel()
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	_, castSpell, _ := boromirTrigger(g, 3, false)

	engine.resolveTopOfStack(g, &TurnLog{})

	if _, ok := stackObjectByID(g, castSpell.ID); !ok {
		t.Fatal("mana-spent spell was countered; intervening condition did not gate the counter")
	}
}

// TestBoromirFailsClosedOnUncounterableSpell proves an uncounterable triggering
// spell is left untouched rather than incorrectly removed.
func TestBoromirFailsClosedOnUncounterableSpell(t *testing.T) {
	t.Parallel()
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	_, castSpell, _ := boromirTrigger(g, 0, true)

	engine.resolveTopOfStack(g, &TurnLog{})

	if _, ok := stackObjectByID(g, castSpell.ID); !ok {
		t.Fatal("uncounterable spell was removed from the stack; counter must fail closed")
	}
}

// TestBoromirFailsClosedWhenSpellLeftStack proves the counter has no incorrect
// effect when the triggering spell already left the stack before the trigger
// resolves (it resolved or was otherwise removed). The trigger resolves harmlessly.
func TestBoromirFailsClosedWhenSpellLeftStack(t *testing.T) {
	t.Parallel()
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	_, castSpell, _ := boromirTrigger(g, 0, false)
	// Remove the triggering spell from the stack before the trigger resolves.
	if _, ok := g.Stack.RemoveByID(castSpell.ID); !ok {
		t.Fatal("failed to remove triggering spell for setup")
	}
	sizeBefore := g.Stack.Size()

	engine.resolveTopOfStack(g, &TurnLog{})

	if got := g.Stack.Size(); got != sizeBefore-1 {
		t.Fatalf("stack size = %d, want %d (only the resolved trigger leaves)", got, sizeBefore-1)
	}
}

// TestBoromirTriggerControllerFilter proves the opponent controller filter fires
// on an opponent's cast but never on the source controller's own cast, so a
// controller's own zero-mana spell is not countered.
func TestBoromirTriggerControllerFilter(t *testing.T) {
	t.Parallel()
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	source := addCreaturePermanent(g, game.Player1)
	pattern := &game.TriggerPattern{Event: game.EventSpellCast, Controller: game.TriggerControllerOpponent}

	opponentCast := game.Event{Kind: game.EventSpellCast, Controller: game.Player2}
	if !triggerMatchesEventForController(g, source, game.Player1, pattern, opponentCast) {
		t.Fatal("opponent cast did not match the opponent controller filter")
	}
	ownCast := game.Event{Kind: game.EventSpellCast, Controller: game.Player1}
	if triggerMatchesEventForController(g, source, game.Player1, pattern, ownCast) {
		t.Fatal("controller's own cast matched the opponent controller filter")
	}
}

// TestBoromirInterveningConditionOnManaSpent proves the no-mana intervening
// condition is satisfied only when the triggering event recorded zero actual
// mana spent, and fails closed when the spend amount is absent. Convoke/Improvise
// casts paid entirely by tapping creatures/artifacts and alternative {0} casts
// all record ManaSpentToCast == 0 and so satisfy it.
func TestBoromirInterveningConditionOnManaSpent(t *testing.T) {
	t.Parallel()
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	trigger := &game.TriggerCondition{
		InterveningCondition: opt.Val(game.Condition{
			Aggregates: []game.AggregateComparison{{
				Aggregate: game.AggregateEventSpellManaSpentToCast,
				Op:        compare.LessOrEqual,
				Value:     0,
			}},
		}),
	}
	cases := []struct {
		name  string
		event *game.Event
		want  bool
	}{
		{"zero mana (free/convoke/alternative {0})", &game.Event{Kind: game.EventSpellCast, ManaSpentToCast: opt.Val(0)}, true},
		{"mana spent", &game.Event{Kind: game.EventSpellCast, ManaSpentToCast: opt.Val(2)}, false},
		{"absent spend fails closed", &game.Event{Kind: game.EventSpellCast}, false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			if got := triggerInterveningIf(g, nil, game.Player1, trigger, tc.event); got != tc.want {
				t.Fatalf("triggerInterveningIf = %v, want %v", got, tc.want)
			}
		})
	}
}

// TestBoromirIgnoresSpellCopies proves the trigger does not fire on a spell copy.
// A copy of a spell is put onto the stack without being cast (CR 707.10c), so it
// generates no spell-cast event and the cast-trigger pattern never matches it.
func TestBoromirIgnoresSpellCopies(t *testing.T) {
	t.Parallel()
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	source := addCreaturePermanent(g, game.Player1)
	pattern := &game.TriggerPattern{Event: game.EventSpellCast, Controller: game.TriggerControllerOpponent}

	// A spell copy produces no EventSpellCast; a non-cast event never matches the
	// cast-trigger pattern.
	nonCast := game.Event{Kind: game.EventZoneChanged, Controller: game.Player2}
	if triggerMatchesEventForController(g, source, game.Player1, pattern, nonCast) {
		t.Fatal("cast trigger matched a non-cast event; spell copies must not trigger it")
	}
}
