package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/compare"
	"github.com/natefinch/council4/opt"
)

// attackBatchDeclared appends an EventAttackerDeclared batch sharing a fresh
// SimultaneousID for every target and returns a StackObject whose TriggerEvent
// is the first declared attacker, mirroring how the runtime resolves an
// attack-batch trigger controlled by controller. It backs the Firemane Commando
// gate "if none of those creatures attacked you".
func attackBatchDeclared(g *game.Game, controller game.PlayerID, targets ...game.AttackTarget) *game.StackObject {
	batch := g.IDGen.Next()
	var first game.Event
	for i, target := range targets {
		event := game.Event{
			Kind:           game.EventAttackerDeclared,
			PermanentID:    g.IDGen.Next(),
			AttackTarget:   target,
			SimultaneousID: batch,
		}
		if len(targets) == 1 {
			event.SimultaneousID = 0
		}
		g.AppendEvent(event)
		if i == 0 {
			first = event
		}
	}
	return &game.StackObject{
		Controller:      controller,
		HasTriggerEvent: true,
		TriggerEvent:    first,
	}
}

// TestAttackersInBatchAttackedControllerDirectly covers the reusable attack-batch
// relation behind Firemane Commando's "if none of those creatures attacked you":
// only a direct attack on the controlling player counts, so attacks on other
// players, on any planeswalker (even the controller's own), or on a battle are
// excluded (CR 508.1).
func TestAttackersInBatchAttackedControllerDirectly(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name    string
		targets []game.AttackTarget
		want    int
	}{
		{
			name: "all away from controller",
			targets: []game.AttackTarget{
				{Player: game.Player2},
				{Player: game.Player3},
			},
			want: 0,
		},
		{
			name: "one attacks controller directly",
			targets: []game.AttackTarget{
				{Player: game.Player1},
				{Player: game.Player3},
			},
			want: 1,
		},
		{
			name: "planeswalker of controller does not count",
			targets: []game.AttackTarget{
				{Player: game.Player1, PlaneswalkerID: 999},
				{Player: game.Player3},
			},
			want: 0,
		},
		{
			name: "battle toward controller does not count",
			targets: []game.AttackTarget{
				{Player: game.Player1, BattleID: 777},
				{Player: game.Player2},
			},
			want: 0,
		},
		{
			name: "both attack controller directly",
			targets: []game.AttackTarget{
				{Player: game.Player1},
				{Player: game.Player1},
			},
			want: 2,
		},
		{
			name:    "single attacker toward controller",
			targets: []game.AttackTarget{{Player: game.Player1}},
			want:    1,
		},
		{
			name:    "single attacker away",
			targets: []game.AttackTarget{{Player: game.Player2}},
			want:    0,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
			obj := attackBatchDeclared(g, game.Player1, tc.targets...)
			if got := attackersInBatchAttackedControllerDirectly(g, &obj.TriggerEvent, game.Player1); got != tc.want {
				t.Fatalf("count = %d, want %d", got, tc.want)
			}
		})
	}
}

// firemaneGate builds the effect-gate condition Firemane Commando lowers to:
// "if none of those creatures attacked you" holds when at most zero attackers in
// the triggering batch attacked the controller directly.
func firemaneGate() opt.V[game.EffectCondition] {
	return opt.Val(game.EffectCondition{
		Text: "if none of those creatures attacked you",
		Condition: opt.Val(game.Condition{
			Aggregates: []game.AggregateComparison{{
				Aggregate: game.AggregateAttackersInBatchAttackedController,
				Op:        compare.LessOrEqual,
				Value:     0,
			}},
		}),
	})
}

// TestFiremaneGateReadsTriggerEventBatch proves the effect-gate is evaluated
// against the resolving stack object's trigger event even though
// effectConditionSatisfied does not set conditionContext.event, so the attacking
// player's draw is allowed only when no attacker in that batch attacked the
// ability controller. Attacks on other players and on planeswalkers do not
// suppress the draw; a single direct attack on the controller does.
func TestFiremaneGateReadsTriggerEventBatch(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name    string
		targets []game.AttackTarget
		want    bool
	}{
		{
			name: "all attackers away, draw allowed",
			targets: []game.AttackTarget{
				{Player: game.Player2},
				{Player: game.Player3},
			},
			want: true,
		},
		{
			name: "planeswalker of controller attacked, draw allowed",
			targets: []game.AttackTarget{
				{Player: game.Player1, PlaneswalkerID: 999},
				{Player: game.Player3},
			},
			want: true,
		},
		{
			name: "one attacker hits controller, draw suppressed",
			targets: []game.AttackTarget{
				{Player: game.Player1},
				{Player: game.Player3},
			},
			want: false,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
			obj := attackBatchDeclared(g, game.Player1, tc.targets...)
			if got := effectConditionSatisfied(g, obj, firemaneGate()); got != tc.want {
				t.Fatalf("gate satisfied = %v, want %v", got, tc.want)
			}
		})
	}
}

// TestFiremaneGateControllerIdentity proves the gate is controller-relative: the
// same attack batch suppresses the draw for the player it attacked while
// allowing it for a bystander controller. This is what lets each Firemane
// Commando evaluate its own "attacked you" independently in multiplayer.
func TestFiremaneGateControllerIdentity(t *testing.T) {
	t.Parallel()
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	targets := []game.AttackTarget{
		{Player: game.Player1},
		{Player: game.Player3},
	}

	attacked := attackBatchDeclared(g, game.Player1, targets...)
	if effectConditionSatisfied(g, attacked, firemaneGate()) {
		t.Fatal("Player1 was attacked directly; gate must fail")
	}

	bystander := attackBatchDeclared(g, game.Player2, targets...)
	if !effectConditionSatisfied(g, bystander, firemaneGate()) {
		t.Fatal("Player2 was not attacked; gate must hold")
	}
}
