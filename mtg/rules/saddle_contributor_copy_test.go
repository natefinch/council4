package rules

import (
	"slices"
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

const saddleCopyLink = game.LinkedKey("test-saddle-copy")
const saddleCopyResult = game.ResultKey("test-saddle-copy-created")

type saddleCopyTestFixture struct {
	game   *game.Game
	engine *Engine
	source *game.Permanent
	obj    *game.StackObject
}

func saddleCopyCreatureDef(name string, legendary, indestructible bool) *game.CardDef {
	def := &game.CardDef{CardFace: game.CardFace{
		Name:      name,
		Types:     []types.Card{types.Creature},
		Power:     opt.Val(game.PT{Value: 2}),
		Toughness: opt.Val(game.PT{Value: 2}),
	}}
	if legendary {
		def.Supertypes = []types.Super{types.Legendary}
	}
	if indestructible {
		def.StaticAbilities = []game.StaticAbility{game.IndestructibleStaticBody}
	}
	return def
}

func saddleCopyProcess() game.RepeatProcess {
	group := game.SaddleContributorsGroup(
		game.SourcePermanentReference(),
		game.Selection{
			RequiredTypes:     []types.Card{types.Creature},
			ExcludedSupertype: types.Legendary,
		},
	)
	return game.RepeatProcess{
		Times: game.Fixed(2),
		Body: game.Mode{Sequence: []game.Instruction{
			{
				Primitive: game.CreateToken{
					Amount: game.Fixed(1),
					Source: game.TokenCopyOf(game.TokenCopySpec{
						Source: game.TokenCopySourceChosenFromGroup,
						Group:  game.GroupRef(group),
					}),
					EntryTapped:        true,
					AttackSameAsSource: true,
					PublishLinked:      saddleCopyLink,
				},
				PublishResult: saddleCopyResult,
			},
			{
				Primitive: game.CreateDelayedTrigger{Trigger: game.DelayedTriggerDef{
					Timing:              game.DelayedAtBeginningOfNextEndStep,
					CapturedObjectGroup: opt.Val(game.LinkedObjectReference(string(saddleCopyLink))),
					Content: game.Mode{Sequence: []game.Instruction{{Primitive: game.Sacrifice{
						Group: game.CapturedObjectsGroup(),
					}}}}.Ability(),
				}},
				ResultGate: opt.Val(game.InstructionResultGate{
					Key:       saddleCopyResult,
					Succeeded: game.TriTrue,
				}),
			},
		}}.Ability(),
	}
}

func saddleCopyFixture(t *testing.T, target game.AttackTarget) saddleCopyTestFixture {
	t.Helper()
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	source := addCombatPermanent(g, game.Player1, saddleCopyCreatureDef("Test Mount", true, false))
	source.Saddled = true
	g.Turn.ActivePlayer = game.Player1
	g.Turn.Phase = game.PhaseCombat
	g.Combat = &game.CombatState{
		AttackersDeclared: true,
		Attackers: []game.AttackDeclaration{{
			Attacker: source.ObjectID,
			Target:   target,
		}},
	}
	obj := &game.StackObject{
		ID:              g.IDGen.Next(),
		Kind:            game.StackTriggeredAbility,
		Controller:      game.Player1,
		SourceID:        source.ObjectID,
		SourceCardID:    source.CardInstanceID,
		HasTriggerEvent: true,
		TriggerEvent: game.Event{
			Kind:         game.EventAttackerDeclared,
			Player:       target.Player,
			AttackTarget: target,
		},
	}
	return saddleCopyTestFixture{game: g, engine: engine, source: source, obj: obj}
}

func resolveSaddleCopyProcess(
	g *game.Game,
	engine *Engine,
	obj *game.StackObject,
	choices ...[]int,
) {
	agents := [game.NumPlayers]PlayerAgent{
		game.Player1: &choiceOnlyAgent{choices: choices},
	}
	instruction := game.Instruction{Primitive: saddleCopyProcess()}
	engine.resolveInstructionWithChoices(g, obj, &instruction, agents, &TurnLog{})
}

func copiedTokens(g *game.Game, names ...string) []*game.Permanent {
	var tokens []*game.Permanent
	for _, permanent := range g.Battlefield {
		if permanent == nil || !permanent.Token {
			continue
		}
		name := permanentName(g, permanent)
		if len(names) == 0 || slices.Contains(names, name) {
			tokens = append(tokens, permanent)
		}
	}
	return tokens
}

func TestSaddleActivationPreservesExactTappedContributorIdentity(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	source := addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:               "Test Mount",
		Types:              []types.Card{types.Creature},
		Power:              opt.Val(game.PT{Value: 2}),
		Toughness:          opt.Val(game.PT{Value: 2}),
		ActivatedAbilities: []game.ActivatedAbility{game.SaddleActivatedAbility(1)},
	}})
	contributor := addCombatPermanent(g, game.Player1, saddleCopyCreatureDef("Contributor", false, false))
	g.Turn.ActivePlayer = game.Player1
	g.Turn.Phase = game.PhasePrecombatMain
	g.Turn.Step = game.StepNone
	g.Turn.PriorityPlayer = game.Player1

	if !engine.applyAction(g, game.Player1, action.ActivateAbility(source.ObjectID, 0, nil, 0)) {
		t.Fatal("Saddle activation failed")
	}
	obj, ok := g.Stack.Peek()
	if !ok || !slices.Equal(obj.TappedAsCostIDs, []game.ObjectID{contributor.ObjectID}) {
		t.Fatalf("tapped cost IDs = %v, want exact contributor %d", obj.TappedAsCostIDs, contributor.ObjectID)
	}
	engine.resolveTopOfStack(g, &TurnLog{})
	if !source.Saddled || !slices.Equal(source.SaddleContributorIDs, []game.ObjectID{contributor.ObjectID}) {
		t.Fatalf("Saddle state = %v contributors = %v", source.Saddled, source.SaddleContributorIDs)
	}
}

func TestSaddleContributorHistoryUnionsActivationsAndResets(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	source := addCombatPermanent(g, game.Player1, saddleCopyCreatureDef("Test Mount", true, false))
	first := addCombatPermanent(g, game.Player1, saddleCopyCreatureDef("First", false, false))
	second := addCombatPermanent(g, game.Player1, saddleCopyCreatureDef("Second", false, false))
	resolve := func(ids ...game.ObjectID) {
		obj := &game.StackObject{
			Kind:            game.StackActivatedAbility,
			Controller:      game.Player1,
			SourceID:        source.ObjectID,
			SourceCardID:    source.CardInstanceID,
			TappedAsCostIDs: ids,
		}
		resolveInstruction(engine, g, obj, game.BecomeSaddled{Object: game.SourcePermanentReference()}, &TurnLog{})
	}
	resolve(first.ObjectID)
	resolve(first.ObjectID, second.ObjectID)
	if !slices.Equal(source.SaddleContributorIDs, []game.ObjectID{first.ObjectID, second.ObjectID}) {
		t.Fatalf("contributors = %v, want stable union", source.SaddleContributorIDs)
	}
	engine.runEndingPhase(g, [game.NumPlayers]PlayerAgent{})
	if source.Saddled || source.SaddleContributorIDs != nil {
		t.Fatalf("cleanup retained Saddle state: saddled=%v contributors=%v", source.Saddled, source.SaddleContributorIDs)
	}
}

func TestRepeatedSaddleCopyAllowsSameOrDifferentChoices(t *testing.T) {
	for _, test := range []struct {
		name       string
		choices    [][]int
		wantCounts map[string]int
	}{
		{name: "same contributor twice", choices: [][]int{{0}, {0}}, wantCounts: map[string]int{"Bear": 2}},
		{name: "new choice each iteration", choices: [][]int{{0}, {1}}, wantCounts: map[string]int{"Bear": 1, "Wolf": 1}},
	} {
		t.Run(test.name, func(t *testing.T) {
			fixture := saddleCopyFixture(t, game.AttackTarget{Player: game.Player2})
			g, engine, source, obj := fixture.game, fixture.engine, fixture.source, fixture.obj
			bear := addCombatPermanent(g, game.Player1, saddleCopyCreatureDef("Bear", false, false))
			wolf := addCombatPermanent(g, game.Player1, saddleCopyCreatureDef("Wolf", false, false))
			source.SaddleContributorIDs = []game.ObjectID{bear.ObjectID, wolf.ObjectID}

			resolveSaddleCopyProcess(g, engine, obj, test.choices...)

			for name, want := range test.wantCounts {
				if got := len(copiedTokens(g, name)); got != want {
					t.Errorf("%s tokens = %d, want %d", name, got, want)
				}
			}
			if len(g.DelayedTriggers) != 2 {
				t.Fatalf("delayed triggers = %d, want one independently captured batch per iteration", len(g.DelayedTriggers))
			}
		})
	}
}

func TestSaddleCopyContributorIdentityAndControlRules(t *testing.T) {
	fixture := saddleCopyFixture(t, game.AttackTarget{Player: game.Player2})
	g, engine, source, obj := fixture.game, fixture.engine, fixture.source, fixture.obj
	survivor := addCombatPermanent(g, game.Player1, saddleCopyCreatureDef("Survivor", false, false))
	survivor.Controller = game.Player3
	departed := addCombatPermanent(g, game.Player1, saddleCopyCreatureDef("Departed", false, false))
	legendary := addCombatPermanent(g, game.Player1, saddleCopyCreatureDef("Legend", true, false))
	source.SaddleContributorIDs = []game.ObjectID{survivor.ObjectID, departed.ObjectID, legendary.ObjectID}
	if !movePermanentToZone(g, departed, zone.Graveyard) {
		t.Fatal("moving contributor off battlefield failed")
	}
	reentered := addCombatPermanent(g, game.Player1, saddleCopyCreatureDef("Departed", false, false))
	source.Controller = game.Player2

	resolveSaddleCopyProcess(g, engine, obj, []int{0}, []int{0})

	if got := len(copiedTokens(g, "Survivor")); got != 2 {
		t.Fatalf("Survivor tokens = %d, want 2 despite contributor/source control changes", got)
	}
	if got := len(copiedTokens(g, "Departed")); got != 0 {
		t.Fatalf("Departed tokens = %d, leave/reenter object %d must not match recorded object %d", got, reentered.ObjectID, departed.ObjectID)
	}
	if got := len(copiedTokens(g, "Legend")); got != 0 {
		t.Fatalf("Legend tokens = %d, legendary contributors are ineligible", got)
	}
	for _, token := range copiedTokens(g, "Survivor") {
		if token.Controller != game.Player1 {
			t.Errorf("token controller = %v, want trigger controller Player1", token.Controller)
		}
	}
}

func TestSaddleCopySupportsTokenContributorAndSameDefender(t *testing.T) {
	targets := []game.AttackTarget{
		{Player: game.Player2},
		{Player: game.Player2, PlaneswalkerID: 777},
		{Player: game.Player3, BattleID: 888},
	}
	for _, target := range targets {
		fixture := saddleCopyFixture(t, target)
		g, engine, source, obj := fixture.game, fixture.engine, fixture.source, fixture.obj
		if target.PlaneswalkerID != 0 {
			permanent := addCombatPermanent(g, target.Player, &game.CardDef{CardFace: game.CardFace{
				Name:  "Defending Planeswalker",
				Types: []types.Card{types.Planeswalker},
			}})
			g.Combat.Attackers[0].Target.PlaneswalkerID = permanent.ObjectID
			obj.TriggerEvent.AttackTarget.PlaneswalkerID = permanent.ObjectID
			target.PlaneswalkerID = permanent.ObjectID
		}
		if target.BattleID != 0 {
			permanent := addCombatPermanent(g, target.Player, &game.CardDef{CardFace: game.CardFace{
				Name:  "Defending Battle",
				Types: []types.Card{types.Battle},
			}})
			g.Combat.Attackers[0].Target.BattleID = permanent.ObjectID
			obj.TriggerEvent.AttackTarget.BattleID = permanent.ObjectID
			target.BattleID = permanent.ObjectID
		}
		contributor := addCombatPermanent(g, game.Player1, saddleCopyCreatureDef("Creature Token", false, false))
		contributor.Token = true
		source.SaddleContributorIDs = []game.ObjectID{contributor.ObjectID}

		resolveSaddleCopyProcess(g, engine, obj, []int{0}, []int{0})

		for _, token := range copiedTokens(g, "Creature Token") {
			if token.ObjectID == contributor.ObjectID {
				continue
			}
			if !token.Tapped {
				t.Error("copy token entered untapped")
			}
			declaration, ok := attackerDeclarationFor(g, token.ObjectID)
			if !ok || declaration.Target != target {
				t.Errorf("token attack = %#v, want exact source target %#v", declaration.Target, target)
			}
		}
	}
}

func TestSaddleCopyNoValidContributorIsNoOp(t *testing.T) {
	fixture := saddleCopyFixture(t, game.AttackTarget{Player: game.Player2})
	g, engine, source, obj := fixture.game, fixture.engine, fixture.source, fixture.obj
	departed := addCombatPermanent(g, game.Player1, saddleCopyCreatureDef("Departed", false, false))
	legendary := addCombatPermanent(g, game.Player1, saddleCopyCreatureDef("Legend", true, false))
	source.SaddleContributorIDs = []game.ObjectID{departed.ObjectID, legendary.ObjectID}
	if !movePermanentToZone(g, departed, zone.Graveyard) {
		t.Fatal("moving contributor off battlefield failed")
	}

	resolveSaddleCopyProcess(g, engine, obj)

	if got := len(copiedTokens(g)); got != 0 {
		t.Fatalf("created %d tokens without a valid contributor", got)
	}
	if len(g.DelayedTriggers) != 0 {
		t.Fatalf("scheduled %d delayed triggers without a created token", len(g.DelayedTriggers))
	}
}

func TestSaddleCopyDelayedSacrificeCapturesDoubledBatches(t *testing.T) {
	fixture := saddleCopyFixture(t, game.AttackTarget{Player: game.Player2})
	g, engine, source, obj := fixture.game, fixture.engine, fixture.source, fixture.obj
	addReplacementPermanent(t, g, game.Player1, tokenDoublingReplacementCardDef())
	contributor := addCombatPermanent(g, game.Player1, saddleCopyCreatureDef("Indestructible", false, true))
	source.SaddleContributorIDs = []game.ObjectID{contributor.ObjectID}

	resolveSaddleCopyProcess(g, engine, obj, []int{0}, []int{0})
	tokens := copiedTokens(g, "Indestructible")
	if len(tokens) != 4 {
		t.Fatalf("tokens = %d, want two independently doubled batches", len(tokens))
	}
	if len(g.DelayedTriggers) != 2 {
		t.Fatalf("delayed triggers = %d, want 2", len(g.DelayedTriggers))
	}
	if linked := linkedObjects(g, linkedObjectSourceKey(g, obj, string(saddleCopyLink))); len(linked) != 2 {
		t.Fatalf("current linked batch = %v, want second doubled batch; obj source=%d card=%d all=%v",
			linked, obj.SourceID, obj.SourceCardID, g.LinkedObjects)
	}
	for i, delayed := range g.DelayedTriggers {
		if len(delayed.CapturedObjectIDs) != 2 {
			t.Fatalf("delayed trigger %d captured %v, want one doubled batch", i, delayed.CapturedObjectIDs)
		}
	}

	engine.runEndingPhase(g, [game.NumPlayers]PlayerAgent{})

	for _, token := range tokens {
		if _, ok := permanentByObjectID(g, token.ObjectID); ok {
			t.Errorf("indestructible doubled token %d (controller %v) was not sacrificed; delayed=%d stack=%d",
				token.ObjectID, token.Controller, len(g.DelayedTriggers), g.Stack.Size())
		}
	}
}

func TestSaddleHistoryBelongsToExactMountObject(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	mount := addCombatPermanent(g, game.Player1, saddleCopyCreatureDef("Mount", true, false))
	contributor := addCombatPermanent(g, game.Player1, saddleCopyCreatureDef("Contributor", false, false))
	mount.Saddled = true
	mount.SaddleContributorIDs = []game.ObjectID{contributor.ObjectID}
	if !movePermanentToZone(g, mount, zone.Graveyard) {
		t.Fatal("moving Mount off battlefield failed")
	}

	returned := addCombatPermanent(g, game.Player1, saddleCopyCreatureDef("Mount", true, false))
	if returned.Saddled || len(returned.SaddleContributorIDs) != 0 {
		t.Fatalf("new Mount object inherited old history: saddled=%v contributors=%v", returned.Saddled, returned.SaddleContributorIDs)
	}
}

func TestSaddleCopyUsesMountLastKnownContributorHistory(t *testing.T) {
	fixture := saddleCopyFixture(t, game.AttackTarget{Player: game.Player2})
	contributor := addCombatPermanent(fixture.game, game.Player1, saddleCopyCreatureDef("Contributor", false, false))
	fixture.source.SaddleContributorIDs = []game.ObjectID{contributor.ObjectID}
	if !movePermanentToZone(fixture.game, fixture.source, zone.Graveyard) {
		t.Fatal("moving Mount off battlefield failed")
	}

	resolveSaddleCopyProcess(fixture.game, fixture.engine, fixture.obj, []int{0}, []int{0})

	tokens := copiedTokens(fixture.game, "Contributor")
	if len(tokens) != 2 {
		t.Fatalf("tokens = %d, want 2 from the departed Mount's source-object history", len(tokens))
	}
	for _, token := range tokens {
		declaration, ok := attackerDeclarationFor(fixture.game, token.ObjectID)
		if !ok || declaration.Target.Player != game.Player2 {
			t.Errorf("token %d did not use the trigger-event defender after the Mount left", token.ObjectID)
		}
	}
}
