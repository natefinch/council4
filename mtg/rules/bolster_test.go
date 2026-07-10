package rules

import (
	"testing"

	cards "github.com/natefinch/council4/mtg/cards/o"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/counter"
)

// newOptimusFront puts the real Optimus Prime, Hero card onto the controller's
// battlefield as its front face so its "At the beginning of each end step,
// bolster 1." trigger runs through the real resolution path.
func newOptimusFront(g *game.Game, controller game.PlayerID) *game.Permanent {
	permanent := addCombatPermanent(g, controller, cards.OptimusPrimeHero())
	permanent.Face = game.FaceFront
	return permanent
}

// newOptimusBack puts the real card onto the battlefield already converted to
// its back face, Optimus Prime, Autobot Leader, so its "Whenever you attack,
// bolster 2. The chosen creature gains trample until end of turn. When that
// creature deals combat damage to a player this turn, convert Optimus Prime."
// trigger runs through the real path.
func newOptimusBack(g *game.Game, controller game.PlayerID) *game.Permanent {
	permanent := addCombatPermanent(g, controller, cards.OptimusPrimeHero())
	permanent.Face = game.FaceBack
	permanent.Transformed = true
	return permanent
}

func optimusObject(g *game.Game, permanent *game.Permanent) *game.StackObject {
	return &game.StackObject{
		ID:           g.IDGen.Next(),
		Kind:         game.StackTriggeredAbility,
		SourceID:     permanent.ObjectID,
		SourceCardID: permanent.CardInstanceID,
		Face:         permanent.Face,
		Controller:   permanent.Controller,
	}
}

// TestBolsterCountersLeastToughnessCreatureYouControl proves the reusable
// bolster keyword action (mechanic #1): bolster 1 puts a +1/+1 counter on the
// single creature with the least toughness among the creatures the resolving
// player controls, ignoring higher-toughness creatures and creatures an opponent
// controls even when the opponent's creature has less toughness overall.
func TestBolsterCountersLeastToughnessCreatureYouControl(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	g.Turn.ActivePlayer = game.Player1
	engine := NewEngine(nil)
	optimus := newOptimusFront(g, game.Player1)
	tall := addCreatureWithPowerToughness(g, game.Player1, 2, 5)
	short := addCreatureWithPowerToughness(g, game.Player1, 2, 2)
	mid := addCreatureWithPowerToughness(g, game.Player1, 2, 4)
	enemy := addCreatureWithPowerToughness(g, game.Player2, 1, 1)

	content := cards.OptimusPrimeHero().TriggeredAbilities[0].Content
	engine.resolveAbilityContentWithChoices(g, optimusObject(g, optimus), content, [game.NumPlayers]PlayerAgent{}, &TurnLog{})

	if got := short.Counters.Get(counter.PlusOnePlusOne); got != 1 {
		t.Fatalf("least-toughness creature +1/+1 counters = %d, want 1", got)
	}
	if got := tall.Counters.Get(counter.PlusOnePlusOne); got != 0 {
		t.Fatalf("higher-toughness creature +1/+1 counters = %d, want 0", got)
	}
	if got := mid.Counters.Get(counter.PlusOnePlusOne); got != 0 {
		t.Fatalf("middle-toughness creature +1/+1 counters = %d, want 0", got)
	}
	if got := optimus.Counters.Get(counter.PlusOnePlusOne); got != 0 {
		t.Fatalf("Optimus (toughness 8) +1/+1 counters = %d, want 0", got)
	}
	if got := enemy.Counters.Get(counter.PlusOnePlusOne); got != 0 {
		t.Fatalf("opponent creature +1/+1 counters = %d, want 0 (bolster must scope to creatures you control)", got)
	}
}

// TestBolsterAddsAmountCounters proves bolster N places N counters at once
// (bolster 2 on the back-face attack trigger), not one counter N times, and that
// the counter amount reads from the primitive's quantity.
func TestBolsterAddsAmountCounters(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	g.Turn.ActivePlayer = game.Player1
	engine := NewEngine(nil)
	optimus := newOptimusBack(g, game.Player1)
	short := addCreatureWithPowerToughness(g, game.Player1, 2, 1)

	content := cards.OptimusPrimeHero().Back.Val.TriggeredAbilities[0].Content
	engine.resolveAbilityContentWithChoices(g, optimusObject(g, optimus), content, [game.NumPlayers]PlayerAgent{}, &TurnLog{})

	if got := short.Counters.Get(counter.PlusOnePlusOne); got != 2 {
		t.Fatalf("bolster 2 +1/+1 counters = %d, want 2", got)
	}
}

// TestBolsterChosenCreatureGainsTrample proves mechanic #2: the back-face rider
// "The chosen creature gains trample until end of turn." binds "the chosen
// creature" to the just-bolstered creature through the shared linked-object
// reference and grants it trample, leaving other creatures the controller
// controls untouched.
func TestBolsterChosenCreatureGainsTrample(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	g.Turn.ActivePlayer = game.Player1
	engine := NewEngine(nil)
	optimus := newOptimusBack(g, game.Player1)
	tall := addCreatureWithPowerToughness(g, game.Player1, 2, 5)
	short := addCreatureWithPowerToughness(g, game.Player1, 2, 1)

	content := cards.OptimusPrimeHero().Back.Val.TriggeredAbilities[0].Content
	engine.resolveAbilityContentWithChoices(g, optimusObject(g, optimus), content, [game.NumPlayers]PlayerAgent{}, &TurnLog{})

	if !hasKeyword(g, short, game.Trample) {
		t.Fatal("bolstered creature did not gain trample")
	}
	if hasKeyword(g, tall, game.Trample) {
		t.Fatal("a creature that was not bolstered gained trample")
	}
}

// TestBolsterChosenCreatureDelayedConvertFiresForBoundCreature proves mechanic
// #3: the back-face rider schedules a delayed "When that creature deals combat
// damage to a player this turn, convert Optimus Prime." trigger bound to the
// bolstered creature. The trigger ignores combat damage from another creature and
// noncombat damage from the bound creature, and converts Optimus only when the
// bound creature deals combat damage to a player.
func TestBolsterChosenCreatureDelayedConvertFiresForBoundCreature(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	g.Turn.ActivePlayer = game.Player1
	engine := NewEngine(nil)
	optimus := newOptimusBack(g, game.Player1)
	other := addCreatureWithPowerToughness(g, game.Player1, 2, 5)
	chosen := addCreatureWithPowerToughness(g, game.Player1, 2, 1)

	content := cards.OptimusPrimeHero().Back.Val.TriggeredAbilities[0].Content
	engine.resolveAbilityContentWithChoices(g, optimusObject(g, optimus), content, [game.NumPlayers]PlayerAgent{}, &TurnLog{})

	if len(g.DelayedTriggers) != 1 {
		t.Fatalf("delayed triggers = %d, want 1 (combat-damage convert scheduled)", len(g.DelayedTriggers))
	}

	emitEvent(g, game.Event{
		Kind:            game.EventDamageDealt,
		SourceObjectID:  other.ObjectID,
		CombatDamage:    true,
		DamageRecipient: game.DamageRecipientPlayer,
		Player:          game.Player2,
	})
	if engine.putTriggeredAbilitiesOnStack(g) {
		t.Fatal("combat damage from a different creature fired the convert trigger")
	}

	emitEvent(g, game.Event{
		Kind:            game.EventDamageDealt,
		SourceObjectID:  chosen.ObjectID,
		DamageRecipient: game.DamageRecipientPlayer,
		Player:          game.Player2,
	})
	if engine.putTriggeredAbilitiesOnStack(g) {
		t.Fatal("noncombat damage from the bound creature fired the convert trigger")
	}

	emitEvent(g, game.Event{
		Kind:            game.EventDamageDealt,
		SourceObjectID:  chosen.ObjectID,
		CombatDamage:    true,
		DamageRecipient: game.DamageRecipientPlayer,
		Player:          game.Player2,
	})
	if !engine.putTriggeredAbilitiesOnStack(g) {
		t.Fatal("bound creature's combat damage to a player did not fire the convert trigger")
	}
	if g.Stack.Size() != 1 {
		t.Fatalf("stack size = %d, want 1 (convert on the stack)", g.Stack.Size())
	}

	engine.resolveTopOfStack(g, &TurnLog{})
	if optimus.Face != game.FaceFront || optimus.Transformed {
		t.Fatalf("Optimus face/transformed = %v/%v, want front/false (converted from its back face)", optimus.Face, optimus.Transformed)
	}
}
