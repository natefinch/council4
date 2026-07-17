package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

func greatTrainHeistTreasure() *game.CardDef {
	return &game.CardDef{CardFace: game.CardFace{
		Name:     "Treasure",
		Types:    []types.Card{types.Artifact},
		Subtypes: []types.Sub{types.Treasure},
	}}
}

func greatTrainHeistDelayedTrigger(treasure *game.CardDef) game.DelayedTriggerDef {
	return game.DelayedTriggerDef{
		EventPattern: opt.Val(game.TriggerPattern{
			Event:                 game.EventDamageDealt,
			Controller:            game.TriggerControllerYou,
			Subject:               game.TriggerSubjectDamageSource,
			DamageRecipient:       game.DamageRecipientPlayer,
			DamageSourceSelection: game.Selection{RequiredTypes: []types.Card{types.Creature}},
			RequireCombatDamage:   true,
		}),
		EventPlayer: opt.Val(game.TargetPlayerReference(0)),
		Window:      game.DelayedWindowThisTurn,
		Content: game.Mode{Sequence: []game.Instruction{{Primitive: game.CreateToken{
			Amount:      game.Fixed(1),
			Source:      game.TokenDef(treasure),
			EntryTapped: true,
		}}}}.Ability(),
	}
}

func scheduleGreatTrainHeistTrigger(t *testing.T, g *game.Game, controller, target game.PlayerID, treasure *game.CardDef) {
	t.Helper()
	def := greatTrainHeistDelayedTrigger(treasure)
	if !scheduleDelayedTrigger(g, &game.StackObject{
		Controller: controller,
		Targets:    []game.Target{game.PlayerTarget(target)},
	}, &def) {
		t.Fatal("scheduleDelayedTrigger failed")
	}
}

func tappedTreasuresControlledBy(g *game.Game, controller game.PlayerID) (count int, allTapped bool) {
	allTapped = true
	for _, permanent := range g.Battlefield {
		if !permanent.Token || permanent.Controller != controller || permanent.TokenDef == nil ||
			permanent.TokenDef.Name != "Treasure" {
			continue
		}
		count++
		allTapped = allTapped && permanent.Tapped
	}
	return count, allTapped
}

func TestGreatTrainHeistGroupUntapAndFirstStrikeBuff(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	attacker := addCombatCreaturePermanentWithPower(g, game.Player1, 2)
	other := addCombatCreaturePermanentWithPower(g, game.Player1, 1)
	blocker := addCombatCreaturePermanentWithPower(g, game.Player2, 2)
	attacker.Tapped, other.Tapped, blocker.Tapped = true, true, true
	group := game.BattlefieldGroup(game.Selection{
		RequiredTypes: []types.Card{types.Creature},
		Controller:    game.ControllerYou,
	})

	addInstructionSpellToStackForController(g, game.Player1, []game.Instruction{
		{Primitive: game.Untap{Group: group}},
		{Primitive: game.ApplyContinuous{
			ContinuousEffects: []game.ContinuousEffect{
				{Group: group, Layer: game.LayerPowerToughnessModify, PowerDelta: 1},
				{Group: group, Layer: game.LayerAbility, AddKeywords: []game.Keyword{game.FirstStrike}},
			},
			Duration: game.DurationUntilEndOfTurn,
		}},
	}, nil)
	engine.resolveTopOfStack(g, &TurnLog{})

	if attacker.Tapped || other.Tapped || !blocker.Tapped {
		t.Fatalf("tapped state after untap: attacker=%v other=%v blocker=%v", attacker.Tapped, other.Tapped, blocker.Tapped)
	}
	if got := effectivePower(g, attacker); got != 3 {
		t.Fatalf("attacker power = %d, want 3", got)
	}
	if !hasKeyword(g, attacker, game.FirstStrike) || hasKeyword(g, blocker, game.FirstStrike) {
		t.Fatalf("first strike: attacker=%v blocker=%v", hasKeyword(g, attacker, game.FirstStrike), hasKeyword(g, blocker, game.FirstStrike))
	}

	blocker.Tapped = false
	g.Combat = &game.CombatState{
		Attackers: []game.AttackDeclaration{{Attacker: attacker.ObjectID, Target: game.AttackTarget{Player: game.Player2}}},
		Blockers:  []game.BlockDeclaration{{Blocker: blocker.ObjectID, Blocking: attacker.ObjectID}},
	}
	engine.resolveCombatDamagePass(g, firstStrikeCombatDamage, &TurnLog{})
	engine.applyStateBasedActionsWithDeaths(g)
	if _, ok := permanentByObjectID(g, blocker.ObjectID); ok {
		t.Fatal("blocker survived the granted first-strike damage")
	}
	if attacker.MarkedDamage != 0 {
		t.Fatalf("attacker marked damage = %d, want 0", attacker.MarkedDamage)
	}

	expireCleanupDurations(g)
	if got := effectivePower(g, attacker); got != 2 || hasKeyword(g, attacker, game.FirstStrike) {
		t.Fatalf("buff survived cleanup: power=%d firstStrike=%v", got, hasKeyword(g, attacker, game.FirstStrike))
	}
}

func TestGreatTrainHeistDelayedTriggerCapturesExactPlayerAndRepeats(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	treasure := greatTrainHeistTreasure()
	first := addCombatCreaturePermanentWithPower(g, game.Player1, 2)
	second := addCombatCreaturePermanentWithPower(g, game.Player1, 3)
	scheduleGreatTrainHeistTrigger(t, g, game.Player1, game.Player2, treasure)

	dealPlayerDamage(g, first.CardInstanceID, first.ObjectID, game.Player1, game.Player3, 2, true)
	dealPlayerDamage(g, second.CardInstanceID, second.ObjectID, game.Player1, game.Player2, 3, false)
	if engine.putTriggeredAbilitiesOnStack(g) {
		t.Fatal("trigger fired for the wrong player or noncombat damage")
	}

	dealPlayerDamage(g, first.CardInstanceID, first.ObjectID, game.Player1, game.Player2, 2, true)
	if !engine.putTriggeredAbilitiesOnStack(g) || g.Stack.Size() != 1 {
		t.Fatalf("stack size = %d, want one trigger for first combat-damage event", g.Stack.Size())
	}
	engine.resolveTopOfStack(g, &TurnLog{})

	dealPlayerDamage(g, second.CardInstanceID, second.ObjectID, game.Player1, game.Player2, 3, true)
	if !engine.putTriggeredAbilitiesOnStack(g) || g.Stack.Size() != 1 {
		t.Fatalf("stack size = %d, want one trigger for second combat-damage event", g.Stack.Size())
	}
	engine.resolveTopOfStack(g, &TurnLog{})

	if count, allTapped := tappedTreasuresControlledBy(g, game.Player1); count != 2 || !allTapped {
		t.Fatalf("Player1 tapped Treasures = %d allTapped=%v, want 2 true", count, allTapped)
	}
	if len(g.DelayedTriggers) != 1 {
		t.Fatalf("delayed triggers = %d, want repeating trigger to remain", len(g.DelayedTriggers))
	}
	expireEventDelayedTriggers(g)
	if len(g.DelayedTriggers) != 0 {
		t.Fatalf("delayed triggers after cleanup = %d, want 0", len(g.DelayedTriggers))
	}
}

func TestGreatTrainHeistDelayedTriggerFiresInBothDamageSteps(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	addCombatCreaturePermanentWithPower(g, game.Player1, 2, game.DoubleStrike)
	scheduleGreatTrainHeistTrigger(t, g, game.Player1, game.Player2, greatTrainHeistTreasure())

	engine.runCombatPhase(g, allFirstLegalAgents(), &TurnLog{})

	if count, allTapped := tappedTreasuresControlledBy(g, game.Player1); count != 2 || !allTapped {
		t.Fatalf("double-strike tapped Treasures = %d allTapped=%v, want 2 true", count, allTapped)
	}
}

func TestGreatTrainHeistDelayedTriggerUsesEventTimeControlAndSurvivesSource(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	treasure := greatTrainHeistTreasure()
	source := addCombatCreaturePermanent(g, game.Player1)
	creature := addCombatCreaturePermanentWithPower(g, game.Player1, 2)
	def := greatTrainHeistDelayedTrigger(treasure)
	if !scheduleDelayedTrigger(g, &game.StackObject{
		Kind:         game.StackActivatedAbility,
		Controller:   game.Player1,
		SourceID:     source.ObjectID,
		SourceCardID: source.CardInstanceID,
		Targets:      []game.Target{game.PlayerTarget(game.Player2)},
	}, &def) {
		t.Fatal("scheduleDelayedTrigger failed")
	}
	if _, ok := removePermanentFromBattlefield(g, source.ObjectID); !ok {
		t.Fatal("failed to remove source")
	}

	dealPlayerDamage(g, creature.CardInstanceID, creature.ObjectID, game.Player2, game.Player2, 2, true)
	if engine.putTriggeredAbilitiesOnStack(g) {
		t.Fatal("trigger fired after the damage source changed away from the trigger controller")
	}
	dealPlayerDamage(g, creature.CardInstanceID, creature.ObjectID, game.Player1, game.Player2, 2, true)
	if !engine.putTriggeredAbilitiesOnStack(g) {
		t.Fatal("trigger did not use event-time control or stopped after its source left")
	}
	engine.resolveTopOfStack(g, &TurnLog{})
	if count, _ := tappedTreasuresControlledBy(g, game.Player1); count != 1 {
		t.Fatalf("Player1 Treasures = %d, want 1", count)
	}
}

func TestGreatTrainHeistTreasureUsesTokenReplacements(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	treasure := greatTrainHeistTreasure()
	addReplacementPermanent(t, g, game.Player1, tokenDoublingReplacementCardDef())
	creature := addCombatCreaturePermanentWithPower(g, game.Player1, 2)
	scheduleGreatTrainHeistTrigger(t, g, game.Player1, game.Player2, treasure)

	dealPlayerDamage(g, creature.CardInstanceID, creature.ObjectID, game.Player1, game.Player2, 2, true)
	if !engine.putTriggeredAbilitiesOnStack(g) {
		t.Fatal("combat-damage trigger was not put on the stack")
	}
	engine.resolveTopOfStack(g, &TurnLog{})

	if count, allTapped := tappedTreasuresControlledBy(g, game.Player1); count != 2 || !allTapped {
		t.Fatalf("doubled tapped Treasures = %d allTapped=%v, want 2 true", count, allTapped)
	}
}

func TestGreatTrainHeistDelayedTriggerRequiresResolvableTarget(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	def := greatTrainHeistDelayedTrigger(greatTrainHeistTreasure())
	if scheduleDelayedTrigger(g, &game.StackObject{Controller: game.Player1}, &def) {
		t.Fatal("scheduled a player-bound trigger without a resolvable target")
	}
}

func TestGreatTrainHeistFizzlesWhenItsOnlyTargetBecomesIllegal(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	creature := addCombatCreaturePermanent(g, game.Player1)
	creature.Tapped = true
	group := game.BattlefieldGroup(game.Selection{
		RequiredTypes: []types.Card{types.Creature},
		Controller:    game.ControllerYou,
	})
	card := &game.CardDef{CardFace: game.CardFace{
		Name:  "Modal Heist",
		Types: []types.Card{types.Instant},
		SpellAbility: opt.Val(game.AbilityContent{
			MinModes: 1,
			MaxModes: 2,
			Modes: []game.Mode{
				{Sequence: []game.Instruction{{Primitive: game.Untap{Group: group}}}},
				{
					Targets: []game.TargetSpec{{
						MinTargets: 1,
						MaxTargets: 1,
						Allow:      game.TargetAllowPlayer,
						Selection:  opt.Val(game.Selection{Controller: game.ControllerOpponent}),
					}},
					Sequence: []game.Instruction{{Primitive: game.CreateDelayedTrigger{
						Trigger: greatTrainHeistDelayedTrigger(greatTrainHeistTreasure()),
					}}},
				},
			},
		}),
	}}
	addImplementationSpellToStack(g, game.Player1, card, []game.Target{game.PlayerTarget(game.Player2)})
	obj, ok := g.Stack.Peek()
	if !ok {
		t.Fatal("spell was not put on the stack")
	}
	obj.ChosenModes = []int{0, 1}
	obj.TargetCounts = []int{1}
	g.Players[game.Player2].Eliminated = true

	engine.resolveTopOfStack(g, &TurnLog{})

	if !creature.Tapped {
		t.Fatal("untargeted mode resolved even though every spell target was illegal")
	}
	if len(g.DelayedTriggers) != 0 {
		t.Fatalf("delayed triggers = %d, want none after the spell fizzled", len(g.DelayedTriggers))
	}
}
