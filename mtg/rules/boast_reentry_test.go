package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/cards"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/mana"
)

// Source review: Usher of the Fallen KHM 35 printing
// 24e58478-c6a8-4f86-854a-a489c99bd777 (Oracle a6eb06dc-a62d-4fdb-b336-e304d8d68c92);
// https://api.scryfall.com/cards/24e58478-c6a8-4f86-854a-a489c99bd777/rulings
// returned four WotC rulings dated 2021-02-05 (boast usable any time after the
// creature was declared as an attacker; a creature put onto the battlefield
// attacking was never declared and can't boast). Verified CR effective
// 2026-09-25: 702.142a (boast = activate only if this creature attacked this
// turn), 400.7 (a moved object is a new object with no memory of its previous
// existence; no exception covers attack history), 201.5 and 508.4.

func usherBoastScenario(t *testing.T) (*game.Game, *Engine, *game.Permanent) {
	t.Helper()
	def := cards.NewDefaultRegistry().Lookup("Usher of the Fallen")
	if def == nil {
		t.Fatal("production Usher of the Fallen is missing from the default registry")
	}
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	usher := addCombatPermanent(g, game.Player1, def)
	usher.SummoningSick = false
	return g, NewEngine(nil), usher
}

func usherDeclareAttack(t *testing.T, g *game.Game, engine *Engine, attacker *game.Permanent) {
	t.Helper()
	g.Turn.ActivePlayer = game.Player1
	g.Turn.PriorityPlayer = game.Player1
	g.Turn.Phase = game.PhaseCombat
	g.Turn.Step = game.StepDeclareAttackers
	g.Combat = &game.CombatState{}
	attack, ok := action.DeclareAttackers([]game.AttackDeclaration{{
		Attacker: attacker.ObjectID, Target: game.AttackTarget{Player: game.Player2},
	}}).DeclareAttackersPayload()
	if !ok || !engine.applyDeclareAttackers(g, game.Player1, attack) {
		t.Fatal("Usher of the Fallen could not be declared as an attacker")
	}
	g.Combat = nil
	g.Turn.Phase = game.PhasePostcombatMain
	g.Turn.Step = game.StepNone
}

func usherBoastAction(g *game.Game, usher *game.Permanent) action.Action {
	g.Players[game.Player1].ManaPool.Add(mana.W, 2)
	return action.ActivateAbility(usher.ObjectID, 0, nil, 0)
}

func blinkUsher(t *testing.T, g *game.Game, engine *Engine, usher *game.Permanent) *game.Permanent {
	t.Helper()
	obj := &game.StackObject{
		ID:         g.IDGen.Next(),
		Kind:       game.StackSpell,
		SourceID:   addCardInstance(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Blink Spell"}}),
		Controller: game.Player1,
		Targets:    []game.Target{game.PermanentTarget(usher.ObjectID)},
	}
	instrs := immediateBlinkInstructions(game.PutOnBattlefield{})
	for i := range instrs {
		engine.resolveInstructionWithChoices(g, obj, &instrs[i], [game.NumPlayers]PlayerAgent{}, &TurnLog{})
	}
	if _, ok := permanentByObjectID(g, usher.ObjectID); ok {
		t.Fatal("original Usher object remained on the battlefield after exile")
	}
	for _, permanent := range g.Battlefield {
		if permanent.CardInstanceID == usher.CardInstanceID {
			if permanent.ObjectID == usher.ObjectID {
				t.Fatal("returned Usher reused the original object identity")
			}
			return permanent
		}
	}
	t.Fatal("Usher of the Fallen did not return to the battlefield")
	return nil
}

func countHumanWarriorTokens(g *game.Game) int {
	count := 0
	for _, permanent := range g.Battlefield {
		if permanent.Token && permanentTokenName(permanent) == "Human Warrior" {
			count++
		}
	}
	return count
}

// TestUsherOfTheFallenReenteredCannotBoastFromEarlierAttack proves that boast
// tracks the attacking battlefield object, not its card: after Usher attacks,
// leaves and returns, the new object did not attack this turn (CR 400.7,
// 702.142a), and only its own later attack makes boast legal again.
func TestUsherOfTheFallenReenteredCannotBoastFromEarlierAttack(t *testing.T) {
	g, engine, usher := usherBoastScenario(t)
	if containsAction(engine.legalActions(g, game.Player1), usherBoastAction(g, usher)) {
		t.Fatal("Usher boast was legal before Usher attacked")
	}

	usherDeclareAttack(t, g, engine, usher)
	if !containsAction(engine.legalActions(g, game.Player1), usherBoastAction(g, usher)) {
		t.Fatal("Usher boast was not legal after Usher attacked")
	}

	returned := blinkUsher(t, g, engine, usher)
	returned.SummoningSick = false
	if containsAction(engine.legalActions(g, game.Player1), usherBoastAction(g, returned)) {
		t.Fatal("reentered Usher inherited its earlier incarnation's attack for boast")
	}

	usherDeclareAttack(t, g, engine, returned)
	act := usherBoastAction(g, returned)
	if !containsAction(engine.legalActions(g, game.Player1), act) {
		t.Fatal("reentered Usher boast was not legal after the new object attacked")
	}
	if !engine.applyAction(g, game.Player1, act) {
		t.Fatal("reentered Usher boast activation failed after it attacked")
	}
	engine.resolveTopOfStack(g, &TurnLog{})
	if got := countHumanWarriorTokens(g); got != 1 {
		t.Fatalf("Human Warrior tokens = %d, want 1 after boast resolved", got)
	}
}

// TestSelfCombatTriggerMatchingUsesObjectIdentityWhenKnown pins the shared
// matcher: combat events carrying object identities match the source object
// only, while events without object identity keep the card-instance fallback.
func TestSelfCombatTriggerMatchingUsesObjectIdentityWhenKnown(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	source := addCombatCreaturePermanentWithPower(g, game.Player1, 2)
	earlier := source.ObjectID + 1000
	for _, kind := range []game.EventKind{
		game.EventAttackerDeclared,
		game.EventBlockerDeclared,
		game.EventAttackerBecameBlocked,
		game.EventAttackerBecameUnblocked,
	} {
		stale := game.Event{Kind: kind, SourceID: source.CardInstanceID, SourceObjectID: earlier, PermanentID: earlier}
		if triggerSourceMatches(g, source, game.TriggerSourceSelf, game.TriggerSubjectDefault, stale) {
			t.Fatalf("%v of an earlier object with the same card matched the current source", kind)
		}
		current := game.Event{Kind: kind, SourceID: source.CardInstanceID, SourceObjectID: source.ObjectID, PermanentID: source.ObjectID}
		if !triggerSourceMatches(g, source, game.TriggerSourceSelf, game.TriggerSubjectDefault, current) {
			t.Fatalf("%v of the current source object did not match", kind)
		}
		cardOnly := game.Event{Kind: kind, SourceID: source.CardInstanceID}
		if !triggerSourceMatches(g, source, game.TriggerSourceSelf, game.TriggerSubjectDefault, cardOnly) {
			t.Fatalf("%v without object identity lost the card-instance fallback", kind)
		}
	}
	blockedAttacker := game.Event{
		Kind:              game.EventBlockerDeclared,
		SourceObjectID:    source.ObjectID + 2000,
		PermanentID:       source.ObjectID + 2000,
		BlockedAttackerID: source.ObjectID,
	}
	if !triggerSourceMatches(g, source, game.TriggerSourceSelf, game.TriggerSubjectBlockedAttacker, blockedAttacker) {
		t.Fatal("blocker declaration against the current source attacker did not match its blocked-attacker subject")
	}
	blockedAttacker.BlockedAttackerID = earlier
	if triggerSourceMatches(g, source, game.TriggerSourceSelf, game.TriggerSubjectBlockedAttacker, blockedAttacker) {
		t.Fatal("blocker declaration against an earlier attacker object matched the current source")
	}
}
