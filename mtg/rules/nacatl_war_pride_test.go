package rules

import (
	"slices"
	"testing"

	ncards "github.com/natefinch/council4/mtg/cards/n"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/id"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
)

func TestExactlyOneBlockRequirementRejectsZeroAndMultiple(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	attacker := addCombatPermanent(g, game.Player1, ncards.NacatlWarPride())
	first := addCombatCreaturePermanentWithPower(g, game.Player2, 2)
	second := addCombatCreaturePermanentWithPower(g, game.Player2, 2)
	g.Turn.Phase = game.PhaseCombat
	g.Turn.Step = game.StepDeclareBlockers
	g.Combat = &game.CombatState{Attackers: []game.AttackDeclaration{{
		Attacker: attacker.ObjectID,
		Target:   game.AttackTarget{Player: game.Player2},
	}}}
	engine := NewEngine(nil)
	one := action.DeclareBlockers([]game.BlockDeclaration{{Blocker: first.ObjectID, Blocking: attacker.ObjectID}})
	two := action.DeclareBlockers([]game.BlockDeclaration{
		{Blocker: first.ObjectID, Blocking: attacker.ObjectID},
		{Blocker: second.ObjectID, Blocking: attacker.ObjectID},
	})
	if engine.applyDeclareBlockers(g, game.Player2, mustDeclareBlockersPayload(t, action.DeclareBlockers(nil))) {
		t.Fatal("accepted no blockers while an exactly-one requirement was satisfiable")
	}
	if engine.applyDeclareBlockers(g, game.Player2, mustDeclareBlockersPayload(t, two)) {
		t.Fatal("accepted two blockers against an exactly-one attacker")
	}
	if !engine.applyDeclareBlockers(g, game.Player2, mustDeclareBlockersPayload(t, one)) {
		t.Fatal("rejected exactly one legal blocker")
	}
}

func TestMultipleExactlyOneRequirementsMaximizeSatisfiedAttackers(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	firstAttacker := addCombatPermanent(g, game.Player1, ncards.NacatlWarPride())
	secondAttacker := addCombatPermanent(g, game.Player1, ncards.NacatlWarPride())
	firstBlocker := addCombatCreaturePermanentWithPower(g, game.Player2, 2)
	secondBlocker := addCombatCreaturePermanentWithPower(g, game.Player2, 2)
	g.Turn.Phase = game.PhaseCombat
	g.Turn.Step = game.StepDeclareBlockers
	g.Combat = &game.CombatState{Attackers: []game.AttackDeclaration{
		{Attacker: firstAttacker.ObjectID, Target: game.AttackTarget{Player: game.Player2}},
		{Attacker: secondAttacker.ObjectID, Target: game.AttackTarget{Player: game.Player2}},
	}}
	engine := NewEngine(nil)
	partial := action.DeclareBlockers([]game.BlockDeclaration{{
		Blocker: firstBlocker.ObjectID, Blocking: firstAttacker.ObjectID,
	}})
	if engine.applyDeclareBlockers(g, game.Player2, mustDeclareBlockersPayload(t, partial)) {
		t.Fatal("accepted a declaration satisfying only one of two satisfiable requirements")
	}
	complete := action.DeclareBlockers([]game.BlockDeclaration{
		{Blocker: firstBlocker.ObjectID, Blocking: firstAttacker.ObjectID},
		{Blocker: secondBlocker.ObjectID, Blocking: secondAttacker.ObjectID},
	})
	if !slices.ContainsFunc(legalDeclareBlockersActions(g, game.Player2), func(candidate action.Action) bool {
		return actionsEqual(candidate, complete)
	}) {
		t.Fatal("legal actions omitted the maximum-satisfaction declaration")
	}
	swapped := action.DeclareBlockers([]game.BlockDeclaration{
		{Blocker: secondBlocker.ObjectID, Blocking: firstAttacker.ObjectID},
		{Blocker: firstBlocker.ObjectID, Blocking: secondAttacker.ObjectID},
	})
	if !slices.ContainsFunc(legalDeclareBlockersActions(g, game.Player2), func(candidate action.Action) bool {
		return actionsEqual(candidate, swapped)
	}) {
		t.Fatal("legal actions omitted an alternate maximum-satisfaction assignment")
	}
	if !engine.applyDeclareBlockers(g, game.Player2, mustDeclareBlockersPayload(t, complete)) {
		t.Fatal("rejected a declaration satisfying both exactly-one requirements")
	}
}

func TestExactlyOneAndMenaceUnsatisfiableFailsOpen(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	def := ncards.NacatlWarPride()
	def.StaticAbilities = append(def.StaticAbilities, game.MenaceStaticBody)
	attacker := addCombatPermanent(g, game.Player1, def)
	addCombatCreaturePermanentWithPower(g, game.Player2, 2)
	addCombatCreaturePermanentWithPower(g, game.Player2, 2)
	g.Turn.Phase = game.PhaseCombat
	g.Turn.Step = game.StepDeclareBlockers
	g.Combat = &game.CombatState{Attackers: []game.AttackDeclaration{{
		Attacker: attacker.ObjectID,
		Target:   game.AttackTarget{Player: game.Player2},
	}}}
	if !NewEngine(nil).applyDeclareBlockers(g, game.Player2, mustDeclareBlockersPayload(t, action.DeclareBlockers(nil))) {
		t.Fatal("rejected no blocks when menace and exactly-one made the requirement impossible")
	}
}

func TestNacatlWarPrideCreatesCopiedTappedAttackersAndCapturesDoubledBatch(t *testing.T) {
	for _, target := range []struct {
		name string
		make func(*game.Game) game.AttackTarget
	}{
		{name: "planeswalker", make: func(g *game.Game) game.AttackTarget {
			permanent := addCombatPermanent(g, game.Player2, &game.CardDef{CardFace: game.CardFace{
				Name: "Defended Walker", Types: []types.Card{types.Planeswalker},
			}})
			return game.AttackTarget{Player: game.Player2, PlaneswalkerID: permanent.ObjectID}
		}},
		{name: "battle", make: func(g *game.Game) game.AttackTarget {
			permanent := addCombatPermanent(g, game.Player2, &game.CardDef{CardFace: game.CardFace{
				Name: "Defended Battle", Types: []types.Card{types.Battle},
			}})
			return game.AttackTarget{Player: game.Player2, BattleID: permanent.ObjectID}
		}},
	} {
		t.Run(target.name, func(t *testing.T) {
			g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
			engine := NewEngine(nil)
			attacker := addCombatPermanent(g, game.Player1, ncards.NacatlWarPride())
			addCombatCreaturePermanentWithPower(g, game.Player2, 2)
			addCombatCreaturePermanentWithPower(g, game.Player2, 2)
			addReplacementPermanent(t, g, game.Player1, tokenDoublingReplacementCardDef())
			attackTarget := target.make(g)
			g.Combat = &game.CombatState{Attackers: []game.AttackDeclaration{{
				Attacker: attacker.ObjectID, Target: attackTarget,
			}}, AttackersDeclared: true}
			g.Turn.ActivePlayer = game.Player1
			g.Turn.Phase = game.PhaseCombat
			attacker.Controller = game.Player3
			obj := nacatlTriggerObject(g, attacker, attackTarget)
			log := TurnLog{}
			engine.resolveAbilityContentWithChoices(g, obj, ncards.NacatlWarPride().TriggeredAbilities[0].Content, [game.NumPlayers]PlayerAgent{}, &log)

			tokens := nacatlTokens(g)
			if len(tokens) != 4 {
				t.Fatalf("created %d tokens, want 4 after doubling two defenders", len(tokens))
			}
			if len(log.Choices) != 0 {
				t.Fatalf("entry attacking prompted %d times, want none", len(log.Choices))
			}
			for _, token := range tokens {
				if token.Controller != game.Player1 || !token.Tapped {
					t.Fatalf("token = %+v, want tapped under trigger controller", token)
				}
				face, ok := permanentFaceDef(g, token)
				if !ok || face.Name != "Nacatl War-Pride" ||
					len(face.StaticAbilities) != 1 ||
					len(face.TriggeredAbilities) != 1 {
					t.Fatalf("token copy characteristics = %+v", face)
				}
				declaration, ok := attackerDeclarationFor(g, token.ObjectID)
				if !ok || !declaration.Target.IsPlayerAttack() || declaration.Target.Player != game.Player2 {
					t.Fatalf("token attack = %+v, want defending player directly", declaration)
				}
			}
			if len(g.DelayedTriggers) != 1 || len(g.DelayedTriggers[0].CapturedObjectIDs) != 4 {
				t.Fatalf("delayed triggers = %+v", g.DelayedTriggers)
			}
		})
	}
}

func TestNacatlWarPrideDelayedExileKeepsBatchesIndependent(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	attacker := addCombatPermanent(g, game.Player1, ncards.NacatlWarPride())
	firstDefender := addCombatCreaturePermanentWithPower(g, game.Player2, 2)
	target := game.AttackTarget{Player: game.Player2}
	g.Combat = &game.CombatState{Attackers: []game.AttackDeclaration{{
		Attacker: attacker.ObjectID, Target: target,
	}}, AttackersDeclared: true}
	g.Turn.ActivePlayer = game.Player1
	g.Turn.Phase = game.PhaseCombat
	content := ncards.NacatlWarPride().TriggeredAbilities[0].Content
	engine.resolveAbilityContentWithChoices(g, nacatlTriggerObject(g, attacker, target), content, [game.NumPlayers]PlayerAgent{}, &TurnLog{})
	firstBatch := append([]id.ID(nil), g.DelayedTriggers[0].CapturedObjectIDs...)

	secondDefender := addCombatCreaturePermanentWithPower(g, game.Player2, 2)
	engine.resolveAbilityContentWithChoices(g, nacatlTriggerObject(g, attacker, target), content, [game.NumPlayers]PlayerAgent{}, &TurnLog{})
	if len(g.DelayedTriggers) != 2 ||
		len(firstBatch) != 1 ||
		len(g.DelayedTriggers[1].CapturedObjectIDs) != 2 {
		t.Fatalf("captured batches = first %v, delayed %+v", firstBatch, g.DelayedTriggers)
	}
	secondBatch := append([]id.ID(nil), g.DelayedTriggers[1].CapturedObjectIDs...)
	if !movePermanentToZone(g, firstDefender, zone.Graveyard) ||
		!movePermanentToZone(g, secondDefender, zone.Graveyard) {
		t.Fatal("could not remove defending creatures")
	}
	engine.resolveAbilityContentWithChoices(g, nacatlTriggerObject(g, attacker, target), content, [game.NumPlayers]PlayerAgent{}, &TurnLog{})
	if len(g.DelayedTriggers) != 3 || len(g.DelayedTriggers[2].CapturedObjectIDs) != 0 {
		t.Fatalf("zero-token resolution captured stale batch: %+v", g.DelayedTriggers)
	}
	resolveNacatlDelayedExile(engine, g, g.DelayedTriggers[0])
	for _, objectID := range firstBatch {
		if _, ok := permanentByObjectID(g, objectID); ok {
			t.Fatalf("first-batch token %d survived its delayed exile", objectID)
		}
	}
	for _, objectID := range secondBatch {
		if _, ok := permanentByObjectID(g, objectID); !ok {
			t.Fatalf("second-batch token %d was cross-contaminated", objectID)
		}
	}
	resolveNacatlDelayedExile(engine, g, g.DelayedTriggers[1])
	for _, objectID := range secondBatch {
		if _, ok := permanentByObjectID(g, objectID); ok {
			t.Fatalf("second-batch token %d survived its delayed exile", objectID)
		}
	}
}

func TestNacatlWarPrideCopiesSourceLastKnownInformation(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	attacker := addCombatPermanent(g, game.Player1, ncards.NacatlWarPride())
	addCombatCreaturePermanentWithPower(g, game.Player2, 2)
	target := game.AttackTarget{Player: game.Player2}
	g.Combat = &game.CombatState{Attackers: []game.AttackDeclaration{{
		Attacker: attacker.ObjectID, Target: target,
	}}, AttackersDeclared: true}
	g.Turn.Phase = game.PhaseCombat
	obj := nacatlTriggerObject(g, attacker, target)
	if !movePermanentToZone(g, attacker, zone.Graveyard) {
		t.Fatal("could not move source to graveyard")
	}

	engine.resolveAbilityContentWithChoices(g, obj, ncards.NacatlWarPride().TriggeredAbilities[0].Content, [game.NumPlayers]PlayerAgent{}, &TurnLog{})

	tokens := nacatlTokens(g)
	if len(tokens) != 1 {
		t.Fatalf("created %d tokens, want one copied from source LKI", len(tokens))
	}
	face, ok := permanentFaceDef(g, tokens[0])
	if !ok || face.Name != "Nacatl War-Pride" ||
		len(face.StaticAbilities) != 1 ||
		len(face.TriggeredAbilities) != 1 {
		t.Fatalf("token copy characteristics = %+v", face)
	}
	declaration, ok := attackerDeclarationFor(g, tokens[0].ObjectID)
	if !ok || declaration.Target != target {
		t.Fatalf("token attack = %+v, want defending player", declaration)
	}
}

func TestNacatlWarPrideZeroDefendersCreatesNothing(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	attacker := addCombatPermanent(g, game.Player1, ncards.NacatlWarPride())
	target := game.AttackTarget{Player: game.Player2}
	g.Combat = &game.CombatState{Attackers: []game.AttackDeclaration{{
		Attacker: attacker.ObjectID, Target: target,
	}}, AttackersDeclared: true}
	g.Turn.ActivePlayer = game.Player1
	g.Turn.Phase = game.PhaseCombat
	engine.resolveAbilityContentWithChoices(g, nacatlTriggerObject(g, attacker, target), ncards.NacatlWarPride().TriggeredAbilities[0].Content, [game.NumPlayers]PlayerAgent{}, &TurnLog{})
	if tokens := nacatlTokens(g); len(tokens) != 0 {
		t.Fatalf("created %d tokens with zero defending creatures", len(tokens))
	}
	if len(g.DelayedTriggers) != 1 || len(g.DelayedTriggers[0].CapturedObjectIDs) != 0 {
		t.Fatalf("empty delayed capture = %+v", g.DelayedTriggers)
	}
}

func nacatlTriggerObject(g *game.Game, source *game.Permanent, target game.AttackTarget) *game.StackObject {
	return &game.StackObject{
		ID:              g.IDGen.Next(),
		Kind:            game.StackTriggeredAbility,
		Controller:      game.Player1,
		SourceID:        source.ObjectID,
		SourceCardID:    source.CardInstanceID,
		HasTriggerEvent: true,
		TriggerEvent: game.Event{
			Kind:           game.EventAttackerDeclared,
			SourceObjectID: source.ObjectID,
			Player:         target.Player,
			AttackTarget:   target,
		},
	}
}

func nacatlTokens(g *game.Game) []*game.Permanent {
	var tokens []*game.Permanent
	for _, permanent := range g.Battlefield {
		if permanent.Token {
			tokens = append(tokens, permanent)
		}
	}
	return tokens
}

func resolveNacatlDelayedExile(engine *Engine, g *game.Game, delayed game.DelayedTrigger) {
	engine.resolveAbilityContentWithChoices(g, &game.StackObject{
		Controller:        delayed.Controller,
		SourceID:          delayed.SourceObjectID,
		SourceCardID:      delayed.SourceID,
		CapturedObjectIDs: append([]id.ID(nil), delayed.CapturedObjectIDs...),
	}, delayed.Ability.Content, [game.NumPlayers]PlayerAgent{}, &TurnLog{})
}
