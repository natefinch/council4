package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

const tapChosenGroupTestKey = game.ResultKey("optional-tap-group-count")

// tapChosenGroupAgent answers a ChoiceResolution tap-group request by selecting
// a scripted number of the offered options: every option when takeAll is set,
// otherwise the first take options in order. It selects nothing when take is
// zero and takeAll is false, modeling the "may tap X" decline.
type tapChosenGroupAgent struct {
	takeAll bool
	take    int
}

func (tapChosenGroupAgent) ChooseAction(PlayerObservation, []action.Action) action.Action {
	return action.Pass()
}

func (a tapChosenGroupAgent) ChooseChoice(_ PlayerObservation, request game.ChoiceRequest) []int {
	if request.Kind != game.ChoiceResolution {
		return nil
	}
	limit := a.take
	if a.takeAll {
		limit = len(request.Options)
	}
	if limit > len(request.Options) {
		limit = len(request.Options)
	}
	selected := make([]int, 0, limit)
	for i := 0; i < limit; i++ {
		selected = append(selected, request.Options[i].Index)
	}
	return selected
}

// addMyrCreature adds a Myr creature permanent with the given base power to
// controller's battlefield, tapping it when tapped is set.
func addMyrCreature(g *game.Game, controller game.PlayerID, power int, tapped bool) *game.Permanent {
	cardID := g.IDGen.Next()
	g.CardInstances[cardID] = &game.CardInstance{
		ID: cardID,
		Def: &game.CardDef{CardFace: game.CardFace{
			Name:      "Myr Token",
			Types:     []types.Card{types.Creature, types.Artifact},
			Subtypes:  []types.Sub{types.Sub("Myr")},
			Power:     opt.Val(game.PT{Value: power}),
			Toughness: opt.Val(game.PT{Value: power}),
		}},
		Owner: controller,
	}
	permanent := &game.Permanent{
		ObjectID:       g.IDGen.Next(),
		CardInstanceID: cardID,
		Owner:          controller,
		Controller:     controller,
		Tapped:         tapped,
	}
	g.Battlefield = append(g.Battlefield, permanent)
	return permanent
}

// battlesphereTapSelection is the "untapped Myr you control" tap group selection
// the recognizer projects from Myr Battlesphere's tap clause.
func battlesphereTapSelection() game.Selection {
	return game.Selection{
		SubtypesAny: []types.Sub{types.Sub("Myr")},
		Controller:  game.ControllerYou,
		Tapped:      game.TriFalse,
	}
}

// TestTapChosenGroupCandidatesExcludeTappedOwnedAndSubtype proves the tap group
// offers only the untapped Myr the resolving controller controls: a tapped Myr,
// a non-Myr, and an opponent's Myr are all excluded, and choosing every offered
// option taps exactly the controller's untapped Myr, publishing that count.
func TestTapChosenGroupCandidatesExcludeTappedOwnedAndSubtype(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	myrA := addMyrCreature(g, game.Player1, 1, false)
	myrB := addMyrCreature(g, game.Player1, 1, false)
	tappedMyr := addMyrCreature(g, game.Player1, 1, true)
	nonMyr := addCreaturePermanent(g, game.Player1)
	opponentMyr := addMyrCreature(g, game.Player2, 1, false)

	obj := &game.StackObject{ID: g.IDGen.Next(), Kind: game.StackTriggeredAbility, Controller: game.Player1}
	agents := [game.NumPlayers]PlayerAgent{game.Player1: tapChosenGroupAgent{takeAll: true}}
	r := &effectResolver{engine: NewEngine(nil), game: g, obj: obj, agents: agents, log: &TurnLog{}}

	resolved := handleTapChosenGroup(r, game.TapChosenGroup{
		ChooseFrom:   game.PlayerControlledGroup(game.ControllerReference(), battlesphereTapSelection()),
		PublishCount: tapChosenGroupTestKey,
	})

	if !resolved.succeeded {
		t.Fatal("tapping at least one Myr must report success")
	}
	if resolved.amount != 2 {
		t.Fatalf("tapped count = %d, want 2", resolved.amount)
	}
	if !myrA.Tapped || !myrB.Tapped {
		t.Fatal("both untapped controlled Myr must be tapped")
	}
	if !tappedMyr.Tapped {
		t.Fatal("the already-tapped Myr must remain tapped")
	}
	if nonMyr.Tapped {
		t.Fatal("a non-Myr must not be tapped")
	}
	if opponentMyr.Tapped {
		t.Fatal("an opponent's Myr must not be tapped")
	}
	if choice, ok := linkedResolutionChoice(obj, string(tapChosenGroupTestKey)); !ok ||
		choice.Kind != game.ResolutionChoiceNumber || choice.Number != 2 {
		t.Fatalf("published choice = %+v (ok=%v), want number 2", choice, ok)
	}
}

// TestTapChosenGroupZeroDeclines proves declining every option publishes zero
// and reports the instruction as not succeeded, so an "If you do" payoff gated
// on the tap resolves to nothing.
func TestTapChosenGroupZeroDeclines(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	myr := addMyrCreature(g, game.Player1, 1, false)

	obj := &game.StackObject{ID: g.IDGen.Next(), Kind: game.StackTriggeredAbility, Controller: game.Player1}
	agents := [game.NumPlayers]PlayerAgent{game.Player1: tapChosenGroupAgent{take: 0}}
	r := &effectResolver{engine: NewEngine(nil), game: g, obj: obj, agents: agents, log: &TurnLog{}}

	resolved := handleTapChosenGroup(r, game.TapChosenGroup{
		ChooseFrom:   game.PlayerControlledGroup(game.ControllerReference(), battlesphereTapSelection()),
		PublishCount: tapChosenGroupTestKey,
	})

	if resolved.succeeded {
		t.Fatal("tapping nothing must not report success")
	}
	if resolved.amount != 0 {
		t.Fatalf("tapped count = %d, want 0", resolved.amount)
	}
	if myr.Tapped {
		t.Fatal("declining must leave the Myr untapped")
	}
	if choice, ok := linkedResolutionChoice(obj, string(tapChosenGroupTestKey)); !ok ||
		choice.Number != 0 {
		t.Fatalf("published choice = %+v (ok=%v), want number 0", choice, ok)
	}
}

// battlesphereAttackSequence is Myr Battlesphere's attack-trigger instruction
// sequence: tap any number of untapped Myr you control publishing X, then a
// source +X/+0 pump and X damage to the attacked defender, both gated on the tap
// having tapped at least one Myr, both reading the same published X.
func battlesphereAttackSequence() []game.Instruction {
	paidCount := game.Dynamic(game.DynamicAmount{
		Kind:      game.DynamicAmountChosenNumber,
		ResultKey: tapChosenGroupTestKey,
	})
	gate := opt.Val(game.InstructionResultGate{Key: tapChosenGroupTestKey, Succeeded: game.TriTrue})
	return []game.Instruction{
		{
			Primitive: game.TapChosenGroup{
				ChooseFrom:   game.PlayerControlledGroup(game.ControllerReference(), battlesphereTapSelection()),
				PublishCount: tapChosenGroupTestKey,
			},
			PublishResult: tapChosenGroupTestKey,
		},
		{
			Primitive: game.ModifyPT{
				Object:         game.SourcePermanentReference(),
				PowerDelta:     paidCount,
				ToughnessDelta: game.Fixed(0),
				Duration:       game.DurationUntilEndOfTurn,
			},
			ResultGate: gate,
		},
		{
			Primitive: game.Damage{
				Amount:       paidCount,
				Recipient:    game.AttackedDefenderDamageRecipient(),
				DamageSource: opt.Val(game.SourcePermanentReference()),
			},
			ResultGate: gate,
		},
	}
}

// pushBattlesphereAttackTrigger seeds a Battlesphere source permanent (a tapped
// attacking Myr with the given base power) and pushes its attack trigger whose
// captured event attacks target. The source is tapped so it is not among the
// untapped tap fodder.
func pushBattlesphereAttackTrigger(g *game.Game, controller game.PlayerID, target game.AttackTarget, power int) *game.Permanent {
	source := addMyrCreature(g, controller, power, true)
	trigger := game.TriggeredAbility{
		Trigger: game.TriggerCondition{
			Type:    game.TriggerWhenever,
			Pattern: game.TriggerPattern{Event: game.EventAttackerDeclared},
		},
		Content: game.Mode{Sequence: battlesphereAttackSequence()}.Ability(),
	}
	g.Stack.Push(&game.StackObject{
		ID:              g.IDGen.Next(),
		Kind:            game.StackTriggeredAbility,
		SourceID:        source.ObjectID,
		SourceCardID:    source.CardInstanceID,
		Controller:      controller,
		InlineTrigger:   &trigger,
		HasTriggerEvent: true,
		TriggerEvent: game.Event{
			Kind:         game.EventAttackerDeclared,
			PermanentID:  source.ObjectID,
			Controller:   controller,
			Player:       target.Player,
			AttackTarget: target,
		},
	})
	return source
}

// TestBattlesphereAttackPumpsAndBurnsForTapCount proves the composed attack
// trigger reuses the published tap count X for both the source pump and the
// defender damage: tapping three untapped Myr pumps the source +3/+0 and deals 3
// damage to the defending player.
func TestBattlesphereAttackPumpsAndBurnsForTapCount(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	source := pushBattlesphereAttackTrigger(g, game.Player1, game.AttackTarget{Player: game.Player2}, 4)
	addMyrCreature(g, game.Player1, 1, false)
	addMyrCreature(g, game.Player1, 1, false)
	addMyrCreature(g, game.Player1, 1, false)

	startLife := g.Players[game.Player2].Life
	agents := [game.NumPlayers]PlayerAgent{game.Player1: tapChosenGroupAgent{takeAll: true}}
	engine.resolveTopOfStackWithChoices(g, agents, &TurnLog{})

	if got := effectivePower(g, source); got != 7 {
		t.Fatalf("source power after tapping three Myr = %d, want 7 (4+3)", got)
	}
	if got := startLife - g.Players[game.Player2].Life; got != 3 {
		t.Fatalf("defender life lost = %d, want 3", got)
	}
}

// TestBattlesphereAttackDeclineDoesNothing proves declining the tap leaves the
// source unpumped and deals no damage, because the pump and damage are gated on
// the tap having tapped at least one Myr.
func TestBattlesphereAttackDeclineDoesNothing(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	source := pushBattlesphereAttackTrigger(g, game.Player1, game.AttackTarget{Player: game.Player2}, 4)
	addMyrCreature(g, game.Player1, 1, false)

	startLife := g.Players[game.Player2].Life
	agents := [game.NumPlayers]PlayerAgent{game.Player1: tapChosenGroupAgent{take: 0}}
	engine.resolveTopOfStackWithChoices(g, agents, &TurnLog{})

	if got := effectivePower(g, source); got != 4 {
		t.Fatalf("source power after declining = %d, want 4", got)
	}
	if got := g.Players[game.Player2].Life; got != startLife {
		t.Fatalf("defender life = %d, want unchanged %d", got, startLife)
	}
}

// TestBattlesphereAttackDamageSurvivesSourceLeaving proves the damage still
// resolves for the tapped count when the source has left the battlefield before
// resolution (the pump fizzles as a harmless no-op), matching the ruling that a
// gone Battlesphere still deals its X damage.
func TestBattlesphereAttackDamageSurvivesSourceLeaving(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	source := pushBattlesphereAttackTrigger(g, game.Player1, game.AttackTarget{Player: game.Player2}, 4)
	addMyrCreature(g, game.Player1, 1, false)
	addMyrCreature(g, game.Player1, 1, false)

	// The source leaves the battlefield before the trigger resolves.
	if !movePermanentToZone(g, source, zone.Graveyard) {
		t.Fatal("movePermanentToZone failed")
	}

	startLife := g.Players[game.Player2].Life
	agents := [game.NumPlayers]PlayerAgent{game.Player1: tapChosenGroupAgent{takeAll: true}}
	engine.resolveTopOfStackWithChoices(g, agents, &TurnLog{})

	if got := startLife - g.Players[game.Player2].Life; got != 2 {
		t.Fatalf("defender life lost with source gone = %d, want 2", got)
	}
}

// addLoyaltyPlaneswalker adds a planeswalker permanent with the given starting
// loyalty to controller's battlefield.
func addLoyaltyPlaneswalker(g *game.Game, controller game.PlayerID, loyalty int) *game.Permanent {
	permanent := addCombatPermanent(g, controller, &game.CardDef{CardFace: game.CardFace{
		Name:    "Test Planeswalker",
		Types:   []types.Card{types.Planeswalker},
		Loyalty: opt.Val(loyalty),
	}})
	permanent.Counters.Add(counter.Loyalty, loyalty)
	return permanent
}

// TestBattlesphereAttackDamagesAttackedPlaneswalker proves the attacked-defender
// recipient routes the X damage to the planeswalker the source is attacking, not
// the defending player: tapping two Myr removes two loyalty from the planeswalker
// and leaves the defending player's life unchanged.
func TestBattlesphereAttackDamagesAttackedPlaneswalker(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	planeswalker := addLoyaltyPlaneswalker(g, game.Player2, 5)
	source := pushBattlesphereAttackTrigger(g, game.Player1, game.AttackTarget{Player: game.Player2, PlaneswalkerID: planeswalker.ObjectID}, 4)
	addMyrCreature(g, game.Player1, 1, false)
	addMyrCreature(g, game.Player1, 1, false)

	startLife := g.Players[game.Player2].Life
	agents := [game.NumPlayers]PlayerAgent{game.Player1: tapChosenGroupAgent{takeAll: true}}
	engine.resolveTopOfStackWithChoices(g, agents, &TurnLog{})

	if got := effectivePower(g, source); got != 6 {
		t.Fatalf("source power after tapping two Myr = %d, want 6 (4+2)", got)
	}
	if got := planeswalker.Counters.Get(counter.Loyalty); got != 3 {
		t.Fatalf("planeswalker loyalty after 2 damage = %d, want 3 (5-2)", got)
	}
	if got := g.Players[game.Player2].Life; got != startLife {
		t.Fatalf("defender life = %d, want unchanged %d", got, startLife)
	}
}

// TestBattlesphereAttackPlaneswalkerDamageSurvivesSourceLeaving proves the damage
// still reaches the attacked planeswalker (not the defending player) when the
// source has left the battlefield before resolution, using the attack target
// captured on the trigger event.
func TestBattlesphereAttackPlaneswalkerDamageSurvivesSourceLeaving(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	planeswalker := addLoyaltyPlaneswalker(g, game.Player2, 5)
	source := pushBattlesphereAttackTrigger(g, game.Player1, game.AttackTarget{Player: game.Player2, PlaneswalkerID: planeswalker.ObjectID}, 4)
	addMyrCreature(g, game.Player1, 1, false)
	addMyrCreature(g, game.Player1, 1, false)

	if !movePermanentToZone(g, source, zone.Graveyard) {
		t.Fatal("movePermanentToZone failed")
	}

	startLife := g.Players[game.Player2].Life
	agents := [game.NumPlayers]PlayerAgent{game.Player1: tapChosenGroupAgent{takeAll: true}}
	engine.resolveTopOfStackWithChoices(g, agents, &TurnLog{})

	if got := planeswalker.Counters.Get(counter.Loyalty); got != 3 {
		t.Fatalf("planeswalker loyalty with source gone = %d, want 3 (5-2)", got)
	}
	if got := g.Players[game.Player2].Life; got != startLife {
		t.Fatalf("defender life = %d, want unchanged %d (damage must not redirect to the player)", got, startLife)
	}
}

// TestBattlesphereAttackPlaneswalkerGoneDealsNoDamage proves the effect deals no
// damage — and never redirects to the defending player — when the attacked
// planeswalker has left the battlefield before resolution.
func TestBattlesphereAttackPlaneswalkerGoneDealsNoDamage(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	planeswalker := addLoyaltyPlaneswalker(g, game.Player2, 5)
	pushBattlesphereAttackTrigger(g, game.Player1, game.AttackTarget{Player: game.Player2, PlaneswalkerID: planeswalker.ObjectID}, 4)
	addMyrCreature(g, game.Player1, 1, false)
	addMyrCreature(g, game.Player1, 1, false)

	if !movePermanentToZone(g, planeswalker, zone.Graveyard) {
		t.Fatal("movePermanentToZone failed")
	}

	startLife := g.Players[game.Player2].Life
	agents := [game.NumPlayers]PlayerAgent{game.Player1: tapChosenGroupAgent{takeAll: true}}
	engine.resolveTopOfStackWithChoices(g, agents, &TurnLog{})

	if got := g.Players[game.Player2].Life; got != startLife {
		t.Fatalf("defender life = %d, want unchanged %d (no damage when the planeswalker is gone)", got, startLife)
	}
}

// TestBattlesphereAttackDamagesAttackedBattle proves the attacked-defender
// recipient routes the X damage to the battle the source is attacking, removing
// defense counters and leaving the battle's protector's life unchanged.
func TestBattlesphereAttackDamagesAttackedBattle(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	battle := addCombatPermanent(g, game.Player2, &game.CardDef{CardFace: game.CardFace{
		Name:    "Test Battle",
		Types:   []types.Card{types.Battle},
		Defense: opt.Val(5),
	}})
	battle.Counters.Add(counter.Defense, 5)
	pushBattlesphereAttackTrigger(g, game.Player1, game.AttackTarget{Player: game.Player2, BattleID: battle.ObjectID}, 4)
	addMyrCreature(g, game.Player1, 1, false)
	addMyrCreature(g, game.Player1, 1, false)

	startLife := g.Players[game.Player2].Life
	agents := [game.NumPlayers]PlayerAgent{game.Player1: tapChosenGroupAgent{takeAll: true}}
	engine.resolveTopOfStackWithChoices(g, agents, &TurnLog{})

	if got := battle.Counters.Get(counter.Defense); got != 3 {
		t.Fatalf("battle defense after 2 damage = %d, want 3 (5-2)", got)
	}
	if got := g.Players[game.Player2].Life; got != startLife {
		t.Fatalf("defender life = %d, want unchanged %d", got, startLife)
	}
}
