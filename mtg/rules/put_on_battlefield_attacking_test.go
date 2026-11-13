package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// Source review: Engine gap #3835 (PutOnBattlefield.EntryAttacking) is proved
// with the test-local Thunderkin Awakener shape (oracle
// c0171343-9c0f-4001-8f4e-a5af9256c4d2, pinned corpus Oracle "Haste / Whenever
// this creature attacks, choose target Elemental creature card in your graveyard
// with toughness less than this creature's toughness. Return that card to the
// battlefield tapped and attacking. Sacrifice it at the beginning of the next end
// step."). 2019-07-12 rulings: the controller chooses what the returned creature
// attacks, which need not match what Thunderkin attacks; it was never declared
// as an attacker, so "whenever ... attacks" abilities don't trigger for it; a
// returned creature that already left is not sacrificed. CR 2026-11-06: 506.3a
// (a noncreature permanent never becomes attacking), 506.3b, 508.4 (controller
// chooses the defender on entry; "attacking" but never "attacked"), 508.4a,
// 508.4c, 508.4d (entering after blockers are declared, it is unblocked). The
// defender choice is the shared entry-attack choice of created attacking tokens;
// planeswalker and battle destinations on that route are #3683.

// selfAttackLifeCreature is an Elemental creature card whose own "whenever this
// creature attacks" trigger would observe it being declared as an attacker.
func selfAttackLifeCreature(name string, power, toughness int) *game.CardDef {
	return &game.CardDef{CardFace: game.CardFace{
		Name:      name,
		Types:     []types.Card{types.Creature},
		Subtypes:  []types.Sub{types.Elemental},
		Power:     opt.Val(game.PT{Value: power}),
		Toughness: opt.Val(game.PT{Value: toughness}),
		TriggeredAbilities: []game.TriggeredAbility{{
			Trigger: game.TriggerCondition{
				Type: game.TriggerWhenever,
				Pattern: game.TriggerPattern{
					Event:  game.EventAttackerDeclared,
					Source: game.TriggerSourceSelf,
				},
			},
			Content: game.Mode{Sequence: []game.Instruction{{
				Primitive: game.GainLife{Amount: game.Fixed(5), Player: game.ControllerReference()},
			}}}.Ability(),
		}},
	}}
}

// The attack trigger returns the targeted graveyard card tapped and attacking
// the defender its controller chooses (here a different player than the source
// attacks), no attack trigger fires for it, and it is sacrificed at the
// beginning of the next end step.
func TestPutOnBattlefieldEntryAttackingReturnsTargetedCardAttackingChosenDefender(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	source := addCombatPermanent(g, game.Player1, thunderkinShapeDef())
	partner := addCombatCreaturePermanent(g, game.Player1)
	returnedCard := addCardToGraveyard(g, game.Player1, selfAttackLifeCreature("Returned Elemental", 3, 1))

	g.Turn.ActivePlayer = game.Player1
	g.Turn.PriorityPlayer = game.Player1
	g.Turn.Phase = game.PhaseCombat
	g.Turn.Step = game.StepDeclareAttackers
	g.Combat = &game.CombatState{}
	if !engine.applyDeclareAttackers(g, game.Player1, action.DeclareAttackersAction{Attackers: []game.AttackDeclaration{
		{Attacker: source.ObjectID, Target: game.AttackTarget{Player: game.Player2}},
		{Attacker: partner.ObjectID, Target: game.AttackTarget{Player: game.Player3}},
	}}) {
		t.Fatal("attackers could not be declared")
	}
	if !engine.putTriggeredAbilitiesOnStack(g) {
		t.Fatal("attack created no trigger")
	}
	trigger, ok := g.Stack.Peek()
	if !ok || trigger.SourceID != source.ObjectID {
		t.Fatalf("top of stack = %+v, want the Thunderkin shape trigger", trigger)
	}
	requireCardOnlyTarget(t, trigger, returnedCard)

	// Defenders are offered in turn order (Player2, Player3); choose Player3.
	var agents [game.NumPlayers]PlayerAgent
	agents[game.Player1] = scriptedChoiceAgent{answer: []int{1}}
	engine.resolveTopOfStackWithChoices(g, agents, &TurnLog{})

	returned, ok := reanimatedPermanent(g, returnedCard)
	if !ok || returned.Controller != game.Player1 || !returned.Tapped {
		t.Fatalf("returned permanent = %+v (present %v), want tapped under Player1", returned, ok)
	}
	attack, attacking := attackerDeclarationFor(g, returned.ObjectID)
	if !attacking || attack.Target != (game.AttackTarget{Player: game.Player3}) {
		t.Fatalf("returned attack = %+v (attacking %v), want attacking Player3", attack, attacking)
	}
	if len(g.Combat.Bands) != 0 {
		t.Fatalf("bands = %+v, want none for a creature entering attacking", g.Combat.Bands)
	}
	life := g.Players[game.Player1].Life
	if engine.putTriggeredAbilitiesOnStack(g) || !g.Stack.IsEmpty() {
		t.Fatalf("entering attacking triggered an attack ability; stack = %+v", g.Stack)
	}
	if g.Players[game.Player1].Life != life {
		t.Fatal("returned creature's attack trigger resolved although it never attacked")
	}

	engine.runEndingPhase(g, [game.NumPlayers]PlayerAgent{})
	if _, ok := permanentByObjectID(g, returned.ObjectID); ok {
		t.Fatal("returned creature was not sacrificed at the beginning of the next end step")
	}
	if z, ok := cardZone(g, returnedCard); !ok || z != zone.Graveyard {
		t.Fatalf("sacrificed card zone = %v (%v), want graveyard", z, ok)
	}
}

// Ruling 2019-07-12: a returned creature that already left the battlefield is
// not sacrificed; its card stays where it went.
func TestPutOnBattlefieldEntryAttackingDepartedCreatureIsNotSacrificed(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	source := addCombatPermanent(g, game.Player1, thunderkinShapeDef())
	returnedCard := addCardToGraveyard(g, game.Player1, toughnessTestCreature("Returned Elemental", types.Elemental, 1, 1))

	declareThunderkinAttack(t, g, engine, source)
	engine.resolveTopOfStack(g, &TurnLog{})
	returned, ok := reanimatedPermanent(g, returnedCard)
	if !ok {
		t.Fatal("target was not returned")
	}
	if _, attacking := attackerDeclarationFor(g, returned.ObjectID); !attacking {
		t.Fatal("returned creature is not attacking")
	}
	if !movePermanentToZone(g, returned, zone.Hand) {
		t.Fatal("could not return the creature to its owner's hand")
	}
	if _, attacking := attackerDeclarationFor(g, returned.ObjectID); attacking {
		t.Fatal("departed creature remained attacking")
	}

	engine.runEndingPhase(g, [game.NumPlayers]PlayerAgent{})
	if z, ok := cardZone(g, returnedCard); !ok || z != zone.Hand {
		t.Fatalf("departed card zone = %v (%v), want hand", z, ok)
	}
}

// A simultaneous batch entering after blockers are declared puts its creature
// in combat unblocked (CR 508.4d), so it deals combat damage to the chosen
// defender, while a noncreature card in the same batch enters without
// attacking (CR 506.3a).
func TestPutOnBattlefieldEntryAttackingBatchAfterBlockersIsUnblocked(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	attacker := addCombatCreaturePermanentWithPower(g, game.Player1, 2)
	blocker := addCombatCreaturePermanentWithPower(g, game.Player2, 2)
	creatureCard := addCardToGraveyard(g, game.Player1, toughnessTestCreature("Late Attacker", types.Elemental, 4, 4))
	artifactCard := addCardToGraveyard(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:  "Returned Relic",
		Types: []types.Card{types.Artifact},
	}})

	g.Turn.ActivePlayer = game.Player1
	g.Turn.Phase = game.PhaseCombat
	g.Turn.Step = game.StepDeclareBlockers
	g.Combat = &game.CombatState{
		AttackDeclarationComplete: true,
		AttackersDeclared:         true,
		Attackers:                 []game.AttackDeclaration{{Attacker: attacker.ObjectID, Target: game.AttackTarget{Player: game.Player2}}},
		PlayersAttacked:           map[game.PlayerID]bool{game.Player2: true},
		Blockers:                  []game.BlockDeclaration{{Blocker: blocker.ObjectID, Blocking: attacker.ObjectID}},
		BlockedAttackers:          map[game.ObjectID]bool{attacker.ObjectID: true},
		BlockerOrder:              map[game.ObjectID][]game.ObjectID{attacker.ObjectID: {blocker.ObjectID}},
	}
	g.Combat.BlockDeclarationsComplete[game.Player2] = true

	obj := &game.StackObject{Controller: game.Player1, Targets: []game.Target{
		currentCardTarget(t, g, creatureCard),
		currentCardTarget(t, g, artifactCard),
	}}
	resolveInstruction(engine, g, obj, game.PutOnBattlefield{
		Sources: []game.BattlefieldSource{
			game.CardBattlefieldSource(game.CardReference{Kind: game.CardReferenceTarget, TargetIndex: 0}),
			game.CardBattlefieldSource(game.CardReference{Kind: game.CardReferenceTarget, TargetIndex: 1}),
		},
		EntryTapped:    true,
		EntryAttacking: true,
	}, &TurnLog{})

	late, ok := reanimatedPermanent(g, creatureCard)
	if !ok || !late.Tapped {
		t.Fatalf("creature = %+v (present %v), want it returned tapped", late, ok)
	}
	relic, ok := reanimatedPermanent(g, artifactCard)
	if !ok {
		t.Fatal("artifact card did not enter the battlefield")
	}
	if _, attacking := attackerDeclarationFor(g, relic.ObjectID); attacking {
		t.Fatal("noncreature permanent entered attacking")
	}
	if attack, attacking := attackerDeclarationFor(g, late.ObjectID); !attacking || attack.Target.Player != game.Player2 {
		t.Fatalf("late attack = %+v (attacking %v), want attacking Player2", attack, attacking)
	}
	if g.Combat.BlockedAttackers[late.ObjectID] {
		t.Fatal("creature entering after blockers became blocked")
	}

	life := g.Players[game.Player2].Life
	engine.resolveCombatDamage(g, &TurnLog{})
	if got := life - g.Players[game.Player2].Life; got != 4 {
		t.Fatalf("defender lost %d life, want 4 from the unblocked late attacker", got)
	}
}
