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

func baelothTestDef() *game.CardDef {
	treasure := &game.CardDef{CardFace: game.CardFace{
		Name:  string(types.Treasure),
		Types: []types.Card{types.Artifact},
	}}
	return &game.CardDef{CardFace: game.CardFace{
		Name:      "Baeloth Barrityl, Entertainer",
		Types:     []types.Card{types.Creature},
		Power:     opt.Val(game.PT{Value: 2}),
		Toughness: opt.Val(game.PT{Value: 5}),
		StaticAbilities: []game.StaticAbility{{
			RuleEffects: []game.RuleEffect{{
				Kind:               game.RuleEffectGoaded,
				AffectedController: game.ControllerAny,
				PermanentTypes:     []types.Card{types.Creature},
				AffectedSelection: game.Selection{
					Controller:          game.ControllerOpponent,
					PowerLessThanSource: true,
				},
			}},
		}},
		TriggeredAbilities: []game.TriggeredAbility{{
			Trigger: game.TriggerCondition{
				Type: game.TriggerWhenever,
				Pattern: game.TriggerPattern{
					Event: game.EventPermanentDied,
					SubjectSelection: game.Selection{
						RequiredTypes: []types.Card{types.Creature},
						MatchGoaded:   true,
						CombatState:   game.CombatStateAttackingOrBlocking,
					},
				},
			},
			Content: game.Mode{Sequence: []game.Instruction{{
				Primitive: game.CreateToken{
					Amount: game.Fixed(1),
					Source: game.TokenDef(treasure),
				},
			}}}.Ability(),
		}},
	}}
}

func treasureCount(g *game.Game, controller game.PlayerID) int {
	count := 0
	for _, permanent := range g.Battlefield {
		if effectiveController(g, permanent) == controller &&
			permanent.TokenDef != nil &&
			permanent.TokenDef.Name == string(types.Treasure) {
			count++
		}
	}
	return count
}

func TestBaelothContinuousGoadTracksPowerControllerTypeAndSource(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	source := addCombatPermanent(g, game.Player1, baelothTestDef())
	lowP2Def := vanillaCreature("Low P2", 1, 1)
	lowP2 := addCombatPermanent(g, game.Player2, lowP2Def)
	tieP2 := addCombatPermanent(g, game.Player2, vanillaCreature("Tie P2", 2, 2))
	lowP3 := addCombatPermanent(g, game.Player3, vanillaCreature("Low P3", 1, 1))
	ownLow := addCombatPermanent(g, game.Player1, vanillaCreature("Own Low", 1, 1))

	if !wasGoadedByNow(g, lowP2, game.Player1) || !wasGoadedByNow(g, lowP3, game.Player1) {
		t.Fatal("opponents' lower-power creatures are not goaded by the source controller")
	}
	if isGoadedNow(g, tieP2) || isGoadedNow(g, ownLow) {
		t.Fatal("power tie or source controller's creature was goaded")
	}

	source.Counters.Add(counter.PlusOnePlusOne, 1)
	if !wasGoadedByNow(g, tieP2, game.Player1) {
		t.Fatal("raising the source's effective power did not update the goaded group")
	}

	lowP2Def.Types = []types.Card{types.Artifact}
	if isGoadedNow(g, lowP2) {
		t.Fatal("object remained goaded after it stopped being a creature")
	}
	lowP2Def.Types = []types.Card{types.Creature}

	source.Controller = game.Player3
	if isGoadedNow(g, lowP3) {
		t.Fatal("new controller's creature remained in the opponent-controlled group")
	}
	if !wasGoadedByNow(g, ownLow, game.Player3) || !wasGoadedByNow(g, lowP2, game.Player3) {
		t.Fatal("control change did not immediately rebind opponent scope and goad attribution")
	}

	if _, ok := removePermanentFromBattlefield(g, source.ObjectID); !ok {
		t.Fatal("failed to remove source")
	}
	if isGoadedNow(g, lowP2) || isGoadedNow(g, ownLow) {
		t.Fatal("continuous goad persisted after its source left the battlefield")
	}
}

func TestBaelothContinuousAndPermanentGoadAttributionCompose(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	addCombatPermanent(g, game.Player1, baelothTestDef())
	second := addCombatPermanent(g, game.Player3, baelothTestDef())
	second.Counters.Add(counter.PlusOnePlusOne, 1)
	creature := addCombatPermanent(g, game.Player2, vanillaCreature("Entertainer's Target", 1, 1))
	creature.Goaded = map[game.PlayerID]game.GoadStatus{
		game.Player4: {RestOfGame: true},
	}

	for _, player := range []game.PlayerID{game.Player1, game.Player3, game.Player4} {
		if !wasGoadedByNow(g, creature, player) {
			t.Fatalf("creature is not currently goaded by %v", player)
		}
	}

	creature.Counters.Add(counter.PlusOnePlusOne, 3)
	if wasGoadedByNow(g, creature, game.Player1) || wasGoadedByNow(g, creature, game.Player3) {
		t.Fatal("continuous goad attribution remained after the creature crossed both thresholds")
	}
	if !wasGoadedByNow(g, creature, game.Player4) || !isGoadedNow(g, creature) {
		t.Fatal("independent permanent goad was lost when continuous goads stopped applying")
	}
}

func TestBaelothContinuousGoadUsesNegativeEffectivePower(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	source := addCombatPermanent(g, game.Player1, baelothTestDef())
	source.Counters.Add(counter.MinusOneMinusOne, 4)
	lower := addCombatPermanent(g, game.Player2, vanillaCreature("Lower", -3, 1))
	tie := addCombatPermanent(g, game.Player2, vanillaCreature("Tie", -2, 1))
	higher := addCombatPermanent(g, game.Player2, vanillaCreature("Higher", -1, 1))

	if !isGoadedNow(g, lower) {
		t.Fatal("-3-power creature is not goaded by a -2-power source")
	}
	if isGoadedNow(g, tie) || isGoadedNow(g, higher) {
		t.Fatal("tie or greater negative-power creature was goaded")
	}
}

func TestBaelothDeathTriggerUsesEventTimeGoadCombatAndController(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	source := addCombatPermanent(g, game.Player3, baelothTestDef())
	attacker := addCombatPermanent(g, game.Player2, vanillaCreature("Attacker", 1, 1))
	blocker := addCombatPermanent(g, game.Player4, vanillaCreature("Blocker", 1, 1))
	highAttacker := addCombatPermanent(g, game.Player2, vanillaCreature("Not Goaded", 2, 2))
	g.Combat = &game.CombatState{
		Attackers: []game.AttackDeclaration{
			{Attacker: attacker.ObjectID, Target: game.AttackTarget{Player: game.Player1}},
			{Attacker: highAttacker.ObjectID, Target: game.AttackTarget{Player: game.Player1}},
		},
		Blockers: []game.BlockDeclaration{{Blocker: blocker.ObjectID, Blocking: attacker.ObjectID}},
	}

	if !movePermanentsToZoneSimultaneously(g, []*game.Permanent{attacker, blocker, highAttacker}, zone.Graveyard) {
		t.Fatal("simultaneous deaths failed")
	}
	if !engine.putTriggeredAbilitiesOnStack(g) {
		t.Fatal("goaded attacking and blocking deaths created no triggers")
	}
	if _, ok := g.Stack.Peek(); !ok {
		t.Fatal("first death trigger missing")
	}
	engine.resolveTopOfStack(g, &TurnLog{})
	if _, ok := g.Stack.Peek(); !ok {
		t.Fatal("second death trigger missing")
	}
	engine.resolveTopOfStack(g, &TurnLog{})
	if trigger, ok := g.Stack.Peek(); ok {
		t.Fatalf("unexpected third death trigger = %#v", trigger)
	}
	if got := treasureCount(g, game.Player3); got != 2 {
		t.Fatalf("source controller's Treasure count = %d, want 2", got)
	}
	for _, player := range []game.PlayerID{game.Player1, game.Player2, game.Player4} {
		if got := treasureCount(g, player); got != 0 {
			t.Fatalf("Player %v Treasure count = %d, want 0", player, got)
		}
	}
	if effectiveController(g, source) != game.Player3 {
		t.Fatal("source controller changed unexpectedly")
	}
}

func TestGoadedAttackEventCapturesControllerAndUsesLiveControllerAtResolution(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	treasure := &game.CardDef{CardFace: game.CardFace{Name: string(types.Treasure), Types: []types.Card{types.Artifact}}}
	sourceDef := baelothTestDef()
	sourceDef.TriggeredAbilities = []game.TriggeredAbility{{
		Trigger: game.TriggerCondition{
			Type: game.TriggerWhenever,
			Pattern: game.TriggerPattern{
				Event: game.EventAttackerDeclared,
				SubjectSelection: game.Selection{
					RequiredTypes: []types.Card{types.Creature},
					MatchGoaded:   true,
				},
			},
		},
		Content: game.Mode{Sequence: []game.Instruction{{
			Primitive: game.CreateToken{
				Amount:    game.Fixed(1),
				Source:    game.TokenDef(treasure),
				Recipient: opt.Val(game.ObjectControllerReference(game.EventPermanentReference())),
			},
		}}}.Ability(),
	}}
	addCombatPermanent(g, game.Player1, sourceDef)
	attacker := addCombatPermanent(g, game.Player2, vanillaCreature("Goaded Attacker", 1, 1))
	g.Turn.ActivePlayer = game.Player2
	g.Turn.Phase = game.PhaseCombat
	g.Turn.Step = game.StepDeclareAttackers
	g.Combat = &game.CombatState{}

	declaration := game.AttackDeclaration{
		Attacker: attacker.ObjectID,
		Target:   game.AttackTarget{Player: game.Player3},
	}
	if !engine.applyDeclareAttackers(g, game.Player2, action.DeclareAttackersAction{
		Attackers: []game.AttackDeclaration{declaration},
	}) {
		t.Fatal("goaded attack declaration was rejected")
	}
	if !engine.putTriggeredAbilitiesOnStack(g) {
		t.Fatal("goaded attack created no trigger")
	}
	trigger, ok := g.Stack.Peek()
	if !ok || !trigger.TriggerEvent.SubjectGoaded || trigger.TriggerEvent.Controller != game.Player2 {
		t.Fatalf("attack trigger event = %#v, want goaded Player2 attacker snapshot", trigger)
	}

	attacker.Controller = game.Player4
	engine.resolveTopOfStack(g, &TurnLog{})
	if got := treasureCount(g, game.Player2); got != 0 {
		t.Fatalf("former attacking controller Treasure count = %d, want 0", got)
	}
	if got := treasureCount(g, game.Player4); got != 1 {
		t.Fatalf("controller at resolution Treasure count = %d, want 1", got)
	}
}
