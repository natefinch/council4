package rules

import (
	"testing"

	cardo "github.com/natefinch/council4/mtg/cards/o"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/id"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// okoyeTargetAgent drives Okoye's beginning-of-combat trigger through the real
// target enumeration, choosing the creature identified by targetPick for the
// "put a +1/+1 counter on target creature" clause. It defers every other choice
// to its default selection so the delayed trigger, which needs no choices,
// resolves untouched.
type okoyeTargetAgent struct {
	targetPick id.ID
}

func (*okoyeTargetAgent) ChooseAction(PlayerObservation, []action.Action) action.Action {
	return action.Pass()
}

func (a *okoyeTargetAgent) ChooseChoice(_ PlayerObservation, request game.ChoiceRequest) []int {
	if request.Kind == game.ChoiceTarget {
		for i, option := range request.Options {
			if len(option.Targets) == 1 && option.Targets[0].PermanentID == a.targetPick {
				return []int{i}
			}
		}
		return []int{0}
	}
	return request.DefaultSelection
}

// newOkoyeCombat puts the real Okoye, Mighty and Adored onto Player1's
// battlefield next to a Player1 creature that will receive the counter, and
// stages Player1's beginning-of-combat step so a test can fire the real
// beginning-of-combat trigger.
func newOkoyeCombat(t *testing.T) (g *game.Game, engine *Engine, creature *game.Permanent) {
	t.Helper()
	g = game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine = NewEngine(nil)
	addCombatPermanent(g, game.Player1, cardo.OkoyeMightyAndAdored())
	creature = addCombatCreaturePermanentWithPower(g, game.Player1, 2)
	g.Turn.Phase = game.PhaseCombat
	g.Turn.Step = game.StepBeginningOfCombat
	g.Turn.ActivePlayer = game.Player1
	g.Combat = &game.CombatState{}
	return g, engine, creature
}

// fireOkoyeBeginningOfCombat drives Okoye's real beginning-of-combat trigger:
// it emits the turn-based event, puts the trigger on the stack (choosing target
// through the real enumeration), and resolves it, so the +1/+1 counter is placed
// and the delayed "attacks the monarch this turn" trigger is scheduled.
func fireOkoyeBeginningOfCombat(t *testing.T, g *game.Game, engine *Engine, target *game.Permanent) {
	t.Helper()
	emitBeginningOfStepEvent(g, game.StepBeginningOfCombat)
	agents := [game.NumPlayers]PlayerAgent{game.Player1: &okoyeTargetAgent{targetPick: target.ObjectID}}
	log := TurnLog{}
	if !engine.putTriggeredAbilitiesOnStackWithChoices(g, agents, &log) {
		t.Fatal("Okoye beginning-of-combat trigger was not put on the stack")
	}
	engine.resolveTopOfStackWithChoices(g, agents, &log)
}

// declareAttack declares attacker as attacking defender through the real
// declare-attackers engine path, emitting the real EventAttackerDeclared.
func declareAttack(t *testing.T, g *game.Game, engine *Engine, attacker *game.Permanent, defender game.PlayerID) {
	t.Helper()
	g.Turn.Step = game.StepDeclareAttackers
	attack := mustDeclareAttackersPayload(t, action.DeclareAttackers([]game.AttackDeclaration{
		{Attacker: attacker.ObjectID, Target: game.AttackTarget{Player: defender}},
	}))
	if !engine.applyDeclareAttackers(g, game.Player1, attack) {
		t.Fatalf("applyDeclareAttackers() rejected %v attacking %v", attacker.ObjectID, defender)
	}
}

// TestOkoyeBeginningOfCombatPlacesCounterAndSchedulesTrigger proves the
// beginning-of-combat trigger both puts a +1/+1 counter on the target creature
// and sets up the delayed attacks-the-monarch trigger, driven through the real
// trigger enumeration and resolution.
func TestOkoyeBeginningOfCombatPlacesCounterAndSchedulesTrigger(t *testing.T) {
	g, engine, creature := newOkoyeCombat(t)
	fireOkoyeBeginningOfCombat(t, g, engine, creature)

	if got := creature.Counters.Get(counter.PlusOnePlusOne); got != 1 {
		t.Fatalf("target creature +1/+1 counters = %d, want 1", got)
	}
	if len(g.DelayedTriggers) != 1 {
		t.Fatalf("scheduled delayed triggers = %d, want 1", len(g.DelayedTriggers))
	}
	delayed := g.DelayedTriggers[0]
	if !delayed.EventPattern.Exists || delayed.EventPattern.Val.Event != game.EventAttackerDeclared {
		t.Fatalf("delayed trigger pattern = %+v, want an EventAttackerDeclared pattern", delayed.EventPattern)
	}
	if delayed.BoundAttackerObjectID != creature.ObjectID {
		t.Fatalf("delayed trigger bound attacker = %v, want the countered creature %v", delayed.BoundAttackerObjectID, creature.ObjectID)
	}
}

// TestOkoyeCounteredCreatureAttackingMonarchGainsKeywords drives the full engine
// path: after the beginning-of-combat trigger places the counter and schedules
// the delayed trigger, the countered creature is declared attacking the living
// monarch, the delayed trigger fires through the real enumeration, and resolving
// it grants that creature double strike and trample until end of turn.
func TestOkoyeCounteredCreatureAttackingMonarchGainsKeywords(t *testing.T) {
	g, engine, creature := newOkoyeCombat(t)
	fireOkoyeBeginningOfCombat(t, g, engine, creature)

	if !setMonarch(g, game.Player2) {
		t.Fatal("setMonarch(Player2) = false")
	}
	declareAttack(t, g, engine, creature, game.Player2)

	agents := [game.NumPlayers]PlayerAgent{game.Player1: &okoyeTargetAgent{}}
	log := TurnLog{}
	if !engine.putTriggeredAbilitiesOnStackWithChoices(g, agents, &log) {
		t.Fatal("delayed attacks-the-monarch trigger did not fire on the real attack")
	}
	if g.Stack.Size() != 1 {
		t.Fatalf("stack size = %d, want 1 (the delayed grant)", g.Stack.Size())
	}
	engine.resolveTopOfStackWithChoices(g, agents, &log)

	if !hasKeyword(g, creature, game.DoubleStrike) {
		t.Fatal("countered creature attacking the monarch did not gain double strike")
	}
	if !hasKeyword(g, creature, game.Trample) {
		t.Fatal("countered creature attacking the monarch did not gain trample")
	}
}

// TestOkoyeCounteredCreatureAttackingNonMonarchGainsNothing proves the delayed
// trigger is gated on the attacked player being the monarch: when the countered
// creature attacks a non-monarch, no ability fires and it gains no keywords.
func TestOkoyeCounteredCreatureAttackingNonMonarchGainsNothing(t *testing.T) {
	g, engine, creature := newOkoyeCombat(t)
	fireOkoyeBeginningOfCombat(t, g, engine, creature)

	// Player3 holds the crown; the countered creature attacks Player2.
	if !setMonarch(g, game.Player3) {
		t.Fatal("setMonarch(Player3) = false")
	}
	declareAttack(t, g, engine, creature, game.Player2)

	agents := [game.NumPlayers]PlayerAgent{game.Player1: &okoyeTargetAgent{}}
	if engine.putTriggeredAbilitiesOnStackWithChoices(g, agents, &TurnLog{}) {
		t.Fatal("delayed trigger fired when the countered creature attacked a non-monarch")
	}
	if g.Stack.Size() != 0 {
		t.Fatalf("stack size = %d, want 0 (no trigger)", g.Stack.Size())
	}
	if hasKeyword(g, creature, game.DoubleStrike) || hasKeyword(g, creature, game.Trample) {
		t.Fatal("countered creature gained keywords despite not attacking the monarch")
	}
}

// TestOkoyeDifferentCreatureAttackingMonarchGainsNothing proves the delayed
// trigger is bound to the specific countered creature: when a different creature
// attacks the monarch, the trigger does not fire and neither creature gains
// keywords.
func TestOkoyeDifferentCreatureAttackingMonarchGainsNothing(t *testing.T) {
	g, engine, creature := newOkoyeCombat(t)
	other := addCombatCreaturePermanentWithPower(g, game.Player1, 2)
	fireOkoyeBeginningOfCombat(t, g, engine, creature)

	if !setMonarch(g, game.Player2) {
		t.Fatal("setMonarch(Player2) = false")
	}
	// The creature that never received the counter attacks the monarch.
	declareAttack(t, g, engine, other, game.Player2)

	agents := [game.NumPlayers]PlayerAgent{game.Player1: &okoyeTargetAgent{}}
	if engine.putTriggeredAbilitiesOnStackWithChoices(g, agents, &TurnLog{}) {
		t.Fatal("delayed trigger fired when a different creature attacked the monarch")
	}
	if g.Stack.Size() != 0 {
		t.Fatalf("stack size = %d, want 0 (no trigger)", g.Stack.Size())
	}
	if hasKeyword(g, other, game.DoubleStrike) || hasKeyword(g, other, game.Trample) {
		t.Fatal("the uncountered attacker gained keywords it should not have")
	}
	if hasKeyword(g, creature, game.DoubleStrike) || hasKeyword(g, creature, game.Trample) {
		t.Fatal("the countered creature gained keywords despite not attacking")
	}
}

// addCombatTokenCreaturePermanent puts a token creature onto controller's
// battlefield exactly the way real tokens exist (Token, TokenDef, and a zero
// CardInstanceID with no CardInstance entry), so a test can exercise the
// counter target being a token.
func addCombatTokenCreaturePermanent(g *game.Game, controller game.PlayerID, power int) *game.Permanent {
	pt := game.PT{Value: power}
	def := &game.CardDef{CardFace: game.CardFace{
		Name:      "Combat Token Creature",
		Types:     []types.Card{types.Creature},
		Power:     opt.Val(pt),
		Toughness: opt.Val(pt),
	}}
	permanent := &game.Permanent{
		ObjectID:   g.IDGen.Next(),
		Owner:      controller,
		Controller: controller,
		Token:      true,
		TokenDef:   def,
	}
	g.Battlefield = append(g.Battlefield, permanent)
	return permanent
}

// TestOkoyeCounteredTokenCreatureAttackingMonarchGainsKeywords proves the rider
// works when the +1/+1 counter is placed on a TOKEN creature (CardInstanceID
// == 0). The attacker binding must survive for a token: after the
// beginning-of-combat trigger counters the token, the token attacking the
// living monarch fires the delayed trigger through the real path and gains
// double strike and trample.
func TestOkoyeCounteredTokenCreatureAttackingMonarchGainsKeywords(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	addCombatPermanent(g, game.Player1, cardo.OkoyeMightyAndAdored())
	token := addCombatTokenCreaturePermanent(g, game.Player1, 2)
	g.Turn.Phase = game.PhaseCombat
	g.Turn.Step = game.StepBeginningOfCombat
	g.Turn.ActivePlayer = game.Player1
	g.Combat = &game.CombatState{}

	fireOkoyeBeginningOfCombat(t, g, engine, token)

	if got := token.Counters.Get(counter.PlusOnePlusOne); got != 1 {
		t.Fatalf("token creature +1/+1 counters = %d, want 1", got)
	}
	if len(g.DelayedTriggers) != 1 {
		t.Fatalf("scheduled delayed triggers = %d, want 1", len(g.DelayedTriggers))
	}
	if g.DelayedTriggers[0].BoundAttackerObjectID != token.ObjectID {
		t.Fatalf("delayed trigger bound attacker = %v, want the token %v", g.DelayedTriggers[0].BoundAttackerObjectID, token.ObjectID)
	}

	if !setMonarch(g, game.Player2) {
		t.Fatal("setMonarch(Player2) = false")
	}
	declareAttack(t, g, engine, token, game.Player2)

	agents := [game.NumPlayers]PlayerAgent{game.Player1: &okoyeTargetAgent{}}
	log := TurnLog{}
	if !engine.putTriggeredAbilitiesOnStackWithChoices(g, agents, &log) {
		t.Fatal("delayed attacks-the-monarch trigger did not fire when a token attacked the monarch")
	}
	engine.resolveTopOfStackWithChoices(g, agents, &log)

	if !hasKeyword(g, token, game.DoubleStrike) {
		t.Fatal("countered token attacking the monarch did not gain double strike")
	}
	if !hasKeyword(g, token, game.Trample) {
		t.Fatal("countered token attacking the monarch did not gain trample")
	}
}

// is limited to the turn it was created: once the this-turn window ends, the
// countered creature attacking the monarch fires nothing.
func TestOkoyeDelayedTriggerExpiresAfterItsTurn(t *testing.T) {
	g, engine, creature := newOkoyeCombat(t)
	fireOkoyeBeginningOfCombat(t, g, engine, creature)
	if len(g.DelayedTriggers) != 1 {
		t.Fatalf("scheduled delayed triggers = %d, want 1", len(g.DelayedTriggers))
	}

	// End-of-turn cleanup expires this-turn delayed triggers (CR 603.7b).
	expireEventDelayedTriggers(g)
	g.Turn.TurnNumber++
	if len(g.DelayedTriggers) != 0 {
		t.Fatalf("delayed triggers after expiry = %d, want 0", len(g.DelayedTriggers))
	}

	if !setMonarch(g, game.Player2) {
		t.Fatal("setMonarch(Player2) = false")
	}
	declareAttack(t, g, engine, creature, game.Player2)

	agents := [game.NumPlayers]PlayerAgent{game.Player1: &okoyeTargetAgent{}}
	if engine.putTriggeredAbilitiesOnStackWithChoices(g, agents, &TurnLog{}) {
		t.Fatal("expired delayed trigger fired on a later turn")
	}
	if hasKeyword(g, creature, game.DoubleStrike) || hasKeyword(g, creature, game.Trample) {
		t.Fatal("countered creature gained keywords from an expired delayed trigger")
	}
}
