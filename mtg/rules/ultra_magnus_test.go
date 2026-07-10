package rules

import (
	"testing"

	cards "github.com/natefinch/council4/mtg/cards/u"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/compare"
	"github.com/natefinch/council4/mtg/game/id"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// newUltraMagnusFront puts the real Ultra Magnus, Tactician card onto the
// controller's battlefield as its front face so its "Whenever Ultra Magnus
// attacks, you may put an artifact creature card from your hand onto the
// battlefield tapped and attacking. If you do, convert Ultra Magnus at end of
// combat." trigger runs through the real resolution path.
func newUltraMagnusFront(g *game.Game, controller game.PlayerID) *game.Permanent {
	permanent := addCombatPermanent(g, controller, cards.UltraMagnusTactician())
	permanent.Face = game.FaceFront
	return permanent
}

// newUltraMagnusBack puts the real card onto the battlefield already converted
// to its back face, Ultra Magnus, Armored Carrier, so its "Formidable —
// Whenever Ultra Magnus attacks, attacking creatures you control gain
// indestructible until end of turn. If those creatures have total power 8 or
// greater, convert Ultra Magnus." trigger runs through the real path.
func newUltraMagnusBack(g *game.Game, controller game.PlayerID) *game.Permanent {
	permanent := addCombatPermanent(g, controller, cards.UltraMagnusTactician())
	permanent.Face = game.FaceBack
	permanent.Transformed = true
	return permanent
}

func ultraMagnusFrontObject(g *game.Game, permanent *game.Permanent) *game.StackObject {
	return &game.StackObject{
		ID:           g.IDGen.Next(),
		Kind:         game.StackTriggeredAbility,
		SourceID:     permanent.ObjectID,
		SourceCardID: permanent.CardInstanceID,
		Face:         game.FaceFront,
		Controller:   permanent.Controller,
	}
}

func artifactCreatureCard(g *game.Game, controller game.PlayerID, name string, power int) id.ID {
	return addCardToHand(g, controller, &game.CardDef{CardFace: game.CardFace{
		Name:      name,
		Types:     []types.Card{types.Artifact, types.Creature},
		Power:     opt.Val(game.PT{Value: power}),
		Toughness: opt.Val(game.PT{Value: power}),
	}})
}

func attackingCreature(g *game.Game, controller game.PlayerID, name string, power int) *game.Permanent {
	permanent := addCombatPermanent(g, controller, &game.CardDef{CardFace: game.CardFace{
		Name:      name,
		Types:     []types.Card{types.Creature},
		Power:     opt.Val(game.PT{Value: power}),
		Toughness: opt.Val(game.PT{Value: power}),
	}})
	return permanent
}

func markAttacking(g *game.Game, defender game.PlayerID, attackers ...*game.Permanent) {
	if g.Combat == nil {
		g.Combat = &game.CombatState{}
	}
	for _, attacker := range attackers {
		g.Combat.Attackers = append(g.Combat.Attackers, game.AttackDeclaration{
			Attacker: attacker.ObjectID,
			Target:   game.AttackTarget{Player: defender},
		})
	}
	g.Combat.AttackersDeclared = true
}

func permanentIsAttacking(g *game.Game, permanent *game.Permanent) bool {
	if g.Combat == nil {
		return false
	}
	for _, declaration := range g.Combat.Attackers {
		if declaration.Attacker == permanent.ObjectID {
			return true
		}
	}
	return false
}

// TestUltraMagnusPutsArtifactCreatureAndSchedulesConvert proves mechanic #1 (the
// artifact-creature-filtered put-from-hand tapped and attacking) and the "if you
// do" arm of mechanic #2: accepting the optional put moves the artifact creature
// onto the battlefield tapped and attacking while every ineligible card stays in
// hand, and schedules the end-of-combat delayed convert.
func TestUltraMagnusPutsArtifactCreatureAndSchedulesConvert(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	g.Turn.ActivePlayer = game.Player1
	engine := NewEngine(nil)
	um := newUltraMagnusFront(g, game.Player1)
	markAttacking(g, game.Player2, um)

	artifactCreature := artifactCreatureCard(g, game.Player1, "Servo", 1)
	plainCreature := addCardToHand(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:      "Grizzly Bears",
		Types:     []types.Card{types.Creature},
		Power:     opt.Val(game.PT{Value: 2}),
		Toughness: opt.Val(game.PT{Value: 2}),
	}})
	plainArtifact := addCardToHand(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:  "Mox",
		Types: []types.Card{types.Artifact},
	}})

	content := cards.UltraMagnusTactician().TriggeredAbilities[0].Content
	agents := [game.NumPlayers]PlayerAgent{game.Player1: defaultChoiceAgent{}}
	engine.resolveAbilityContentWithChoices(g, ultraMagnusFrontObject(g, um), content, agents, &TurnLog{})

	put, ok := reanimatedPermanent(g, artifactCreature)
	if !ok {
		t.Fatal("artifact creature was not put onto the battlefield")
	}
	if !put.Tapped {
		t.Fatal("put artifact creature did not enter tapped")
	}
	if !permanentIsAttacking(g, put) {
		t.Fatal("put artifact creature did not enter attacking")
	}
	if !g.Players[game.Player1].Hand.Contains(plainCreature) {
		t.Fatal("non-artifact creature was put onto the battlefield (filter too loose)")
	}
	if !g.Players[game.Player1].Hand.Contains(plainArtifact) {
		t.Fatal("non-creature artifact was put onto the battlefield (filter too loose)")
	}
	if len(g.DelayedTriggers) != 1 {
		t.Fatalf("delayed triggers = %d, want 1 (convert scheduled because a creature was put)", len(g.DelayedTriggers))
	}
	if um.Face != game.FaceFront || um.Transformed {
		t.Fatal("Ultra Magnus converted immediately instead of at end of combat")
	}
}

// TestUltraMagnusDeclinedPutSchedulesNoConvert proves the gate on mechanic #2:
// declining the optional put publishes a not-done result, so the "If you do"
// gate fails and no end-of-combat convert is scheduled.
func TestUltraMagnusDeclinedPutSchedulesNoConvert(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	g.Turn.ActivePlayer = game.Player1
	engine := NewEngine(nil)
	um := newUltraMagnusFront(g, game.Player1)
	markAttacking(g, game.Player2, um)
	artifactCreature := artifactCreatureCard(g, game.Player1, "Servo", 1)

	content := cards.UltraMagnusTactician().TriggeredAbilities[0].Content
	// The scripted agent answers the "you may" offer with "No".
	agents := [game.NumPlayers]PlayerAgent{game.Player1: &sequencedChoiceAgent{choices: [][]int{{0}}}}
	engine.resolveAbilityContentWithChoices(g, ultraMagnusFrontObject(g, um), content, agents, &TurnLog{})

	if !g.Players[game.Player1].Hand.Contains(artifactCreature) {
		t.Fatal("declined put still moved the artifact creature out of hand")
	}
	if len(g.DelayedTriggers) != 0 {
		t.Fatalf("delayed triggers = %d, want 0 (no put, so no convert scheduled)", len(g.DelayedTriggers))
	}
}

// TestUltraMagnusScheduledConvertFiresAtEndOfCombat proves the timing arm of
// mechanic #2: once scheduled, the delayed self-convert stays pending through the
// rest of combat and flips Ultra Magnus to its back face at the end-of-combat
// step, reusing the shared delayed-at-end-of-combat infrastructure.
func TestUltraMagnusScheduledConvertFiresAtEndOfCombat(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	um := newUltraMagnusFront(g, game.Player1)

	content := cards.UltraMagnusTactician().TriggeredAbilities[0].Content
	delayed, ok := content.Modes[0].Sequence[1].Primitive.(game.CreateDelayedTrigger)
	if !ok {
		t.Fatalf("front sequence[1] = %T, want game.CreateDelayedTrigger", content.Modes[0].Sequence[1].Primitive)
	}
	def := delayed.Trigger
	if !scheduleDelayedTrigger(g, ultraMagnusFrontObject(g, um), &def) {
		t.Fatal("scheduleDelayedTrigger failed")
	}
	if um.Face != game.FaceFront || um.Transformed {
		t.Fatal("Ultra Magnus converted before end of combat")
	}

	engine.runCombatPhase(g, allFirstLegalAgents(), &TurnLog{})

	if len(g.DelayedTriggers) != 0 {
		t.Fatalf("delayed triggers after end of combat = %d, want 0", len(g.DelayedTriggers))
	}
	if um.Face != game.FaceBack || !um.Transformed {
		t.Fatalf("Ultra Magnus face/transformed = %v/%v, want back/true (converts at end of combat)", um.Face, um.Transformed)
	}
}

// TestUltraMagnusArmoredCarrierConvertsAtHighPower proves mechanic #3 (the
// attack trigger grants indestructible to the attacking creatures you control
// until end of turn) and the satisfied arm of mechanic #4: when those attackers
// have total power 8 or greater, the conditional convert flips Ultra Magnus back
// to its front face.
func TestUltraMagnusArmoredCarrierConvertsAtHighPower(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	g.Turn.ActivePlayer = game.Player1
	engine := NewEngine(nil)
	um := newUltraMagnusBack(g, game.Player1)
	first := attackingCreature(g, game.Player1, "Warpath Bruiser", 4)
	second := attackingCreature(g, game.Player1, "Convoy Guardian", 4)
	markAttacking(g, game.Player2, first, second)

	content := cards.UltraMagnusTactician().Back.Val.TriggeredAbilities[0].Content
	engine.resolveAbilityContentWithChoices(g, ultraMagnusFrontObject(g, um), content, [game.NumPlayers]PlayerAgent{}, &TurnLog{})

	if !hasKeyword(g, first, game.Indestructible) || !hasKeyword(g, second, game.Indestructible) {
		t.Fatal("attacking creatures did not gain indestructible")
	}
	if um.Face != game.FaceFront || um.Transformed {
		t.Fatalf("Ultra Magnus face/transformed = %v/%v, want front/false (total power 8 converts)", um.Face, um.Transformed)
	}
}

// TestUltraMagnusArmoredCarrierKeepsFaceBelowPower proves mechanic #3 still
// grants indestructible while the unsatisfied arm of mechanic #4 leaves Ultra
// Magnus on its back face: when the attacking creatures' total power is below 8
// the conditional convert does not fire.
func TestUltraMagnusArmoredCarrierKeepsFaceBelowPower(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	g.Turn.ActivePlayer = game.Player1
	engine := NewEngine(nil)
	um := newUltraMagnusBack(g, game.Player1)
	first := attackingCreature(g, game.Player1, "Scout Rotor", 3)
	second := attackingCreature(g, game.Player1, "Picket Drone", 3)
	markAttacking(g, game.Player2, first, second)

	content := cards.UltraMagnusTactician().Back.Val.TriggeredAbilities[0].Content
	engine.resolveAbilityContentWithChoices(g, ultraMagnusFrontObject(g, um), content, [game.NumPlayers]PlayerAgent{}, &TurnLog{})

	if !hasKeyword(g, first, game.Indestructible) || !hasKeyword(g, second, game.Indestructible) {
		t.Fatal("attacking creatures did not gain indestructible")
	}
	if um.Face != game.FaceBack || !um.Transformed {
		t.Fatalf("Ultra Magnus face/transformed = %v/%v, want back/true (total power 6 does not convert)", um.Face, um.Transformed)
	}
}

// TestUltraMagnusArmoredCarrierIgnoresOpponentAttackerPower pins the
// controller-scoping invariant of the shared total-power condition (mechanic #4,
// reused by #2847): the convert counts only the attacking creatures the
// controller controls, never an opponent's attackers. Here the controller's
// attackers total only 6 while an opposing attacker would push the combined
// board total to 11; because the gate is controller-scoped, 6 < 8 leaves Ultra
// Magnus on its back face. This test would fail if the controller-scoping
// (allowed-player filter) were ever removed. It likewise confirms mechanic #3
// grants indestructible only to the controller's attackers, not the opponent's.
func TestUltraMagnusArmoredCarrierIgnoresOpponentAttackerPower(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	g.Turn.ActivePlayer = game.Player1
	engine := NewEngine(nil)
	um := newUltraMagnusBack(g, game.Player1)
	first := attackingCreature(g, game.Player1, "Scout Rotor", 3)
	second := attackingCreature(g, game.Player1, "Picket Drone", 3)
	markAttacking(g, game.Player2, first, second)
	opponent := attackingCreature(g, game.Player2, "Insecticon Swarm", 5)
	markAttacking(g, game.Player1, opponent)

	content := cards.UltraMagnusTactician().Back.Val.TriggeredAbilities[0].Content
	engine.resolveAbilityContentWithChoices(g, ultraMagnusFrontObject(g, um), content, [game.NumPlayers]PlayerAgent{}, &TurnLog{})

	if !hasKeyword(g, first, game.Indestructible) || !hasKeyword(g, second, game.Indestructible) {
		t.Fatal("controller's attacking creatures did not gain indestructible")
	}
	if hasKeyword(g, opponent, game.Indestructible) {
		t.Fatal("opponent's attacker gained indestructible; the grant must be controller-scoped")
	}
	if um.Face != game.FaceBack || !um.Transformed {
		t.Fatalf("Ultra Magnus face/transformed = %v/%v, want back/true (controller total power 6 < 8 despite an opposing attacker)", um.Face, um.Transformed)
	}
}

// TestTotalPowerAttackingConditionBothBounds proves the reusable total-power
// condition recognizer (mechanic #4) evaluates both bounds correctly against the
// attacking creatures the controller controls: an "N or greater" lower bound and
// an "N or less" upper bound. Ultra Magnus only exercises the lower bound, so the
// upper bound is validated directly here for the shared mechanic (#2847 reuse).
func TestTotalPowerAttackingConditionBothBounds(t *testing.T) {
	newState := func(powers ...int) (*game.Game, conditionContext) {
		g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
		g.Turn.ActivePlayer = game.Player1
		attackers := make([]*game.Permanent, 0, len(powers))
		for _, power := range powers {
			attackers = append(attackers, attackingCreature(g, game.Player1, "Attacker", power))
		}
		markAttacking(g, game.Player2, attackers...)
		return g, conditionContext{controller: game.Player1}
	}
	atLeast8 := opt.Val(game.Condition{ControlsMatching: opt.Val(game.SelectionCount{
		Selection:  game.Selection{RequiredTypes: []types.Card{types.Creature}, CombatState: game.CombatStateAttacking},
		TotalPower: opt.Val(compare.Int{Op: compare.GreaterOrEqual, Value: 8}),
	})})
	atMost8 := opt.Val(game.Condition{ControlsMatching: opt.Val(game.SelectionCount{
		Selection:  game.Selection{RequiredTypes: []types.Card{types.Creature}, CombatState: game.CombatStateAttacking},
		TotalPower: opt.Val(compare.Int{Op: compare.LessOrEqual, Value: 8}),
	})})

	t.Run("lower bound met", func(t *testing.T) {
		g, ctx := newState(4, 4)
		if !conditionSatisfied(g, ctx, atLeast8) {
			t.Fatal("total power 8 did not satisfy >= 8")
		}
	})
	t.Run("lower bound unmet", func(t *testing.T) {
		g, ctx := newState(3, 3)
		if conditionSatisfied(g, ctx, atLeast8) {
			t.Fatal("total power 6 satisfied >= 8")
		}
	})
	t.Run("upper bound met", func(t *testing.T) {
		g, ctx := newState(3, 3)
		if !conditionSatisfied(g, ctx, atMost8) {
			t.Fatal("total power 6 did not satisfy <= 8")
		}
	})
	t.Run("upper bound at boundary", func(t *testing.T) {
		g, ctx := newState(4, 4)
		if !conditionSatisfied(g, ctx, atMost8) {
			t.Fatal("total power 8 did not satisfy <= 8")
		}
	})
	t.Run("upper bound exceeded", func(t *testing.T) {
		g, ctx := newState(6, 5)
		if conditionSatisfied(g, ctx, atMost8) {
			t.Fatal("total power 11 satisfied <= 8")
		}
	})
}
