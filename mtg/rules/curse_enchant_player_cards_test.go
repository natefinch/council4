package rules

import (
	"testing"

	cardsc "github.com/natefinch/council4/mtg/cards/c"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/types"
)

// tokenNamesByController groups the names of every token permanent on the
// battlefield by the player controlling it, so a test can assert both how many
// tokens each player received and what token they are.
func tokenNamesByController(g *game.Game) map[game.PlayerID][]string {
	names := make(map[game.PlayerID][]string)
	for _, permanent := range g.Battlefield {
		if !permanent.Token || permanent.TokenDef == nil {
			continue
		}
		names[permanent.Controller] = append(names[permanent.Controller], permanent.TokenDef.Name)
	}
	return names
}

// resolveCurseAttackTrigger drives the resolution of an enchanted-player combat
// curse's triggered ability against a combat where Player3 attacks the enchanted
// Player2 with two creatures, Player4 attacks Player2 both directly and through a
// planeswalker, and the controller Player1 attacks Player3. It resolves the real
// generated card's ability content — both the controller effect and the folded
// "Each opponent attacking that player does the same." group effect — so the
// returned battlefield reflects exactly what the generated card produces.
func resolveCurseAttackTrigger(t *testing.T, curse *game.CardDef) *game.Game {
	t.Helper()
	return resolveCurseAttackTriggerWithSetup(t, curse, nil)
}

// resolveCurseAttackTriggerWithSetup extends resolveCurseAttackTrigger with a
// setup hook that runs after the combat is declared but before the curse's
// ability resolves, so a test can seed the recipients' libraries or add tapped
// permanents. It resolves the full instruction — not merely its primitive — so an
// instruction-level ForEachPlayerGroup (Curse of Bounty's per-attacker untap)
// iterates its members rather than resolving once against the controller.
func resolveCurseAttackTriggerWithSetup(t *testing.T, curse *game.CardDef, setup func(g *game.Game)) *game.Game {
	t.Helper()
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})

	p3a := addCombatCreaturePermanent(g, game.Player3)
	p3b := addCombatCreaturePermanent(g, game.Player3)
	p4a := addCombatCreaturePermanent(g, game.Player4)
	p4pw := addCombatCreaturePermanent(g, game.Player4)
	p1a := addCombatCreaturePermanent(g, game.Player1)
	g.Combat = &game.CombatState{Attackers: []game.AttackDeclaration{
		{Attacker: p3a.ObjectID, Target: game.AttackTarget{Player: game.Player2}},
		{Attacker: p3b.ObjectID, Target: game.AttackTarget{Player: game.Player2}},
		{Attacker: p4a.ObjectID, Target: game.AttackTarget{Player: game.Player2}},
		{Attacker: p4pw.ObjectID, Target: game.AttackTarget{Player: game.Player2, PlaneswalkerID: g.IDGen.Next()}},
		{Attacker: p1a.ObjectID, Target: game.AttackTarget{Player: game.Player3}},
	}}

	if setup != nil {
		setup(g)
	}
	resolveCurseAttackedAbility(t, g, curse)
	return g
}

// resolveCurseAttackedAbility resolves curse's single enchanted-player-attacked
// triggered ability against the combat already configured in g, with Player1 as
// the curse's controller and Player2 as the enchanted, attacked player. It asserts
// the trigger's once-per-combat enchanted-player pattern and resolves each of the
// ability's instructions in order, passing the full instruction so an
// instruction-level ForEachPlayerGroup iterates its members. Edge-case tests build
// an arbitrary combat and battlefield in g, then call this to exercise the real
// generated card content against it.
func resolveCurseAttackedAbility(t *testing.T, g *game.Game, curse *game.CardDef) {
	t.Helper()
	engine := NewEngine(nil)
	obj := &game.StackObject{
		Controller:      game.Player1,
		HasTriggerEvent: true,
		TriggerEvent: game.Event{
			Kind:         game.EventAttackerDeclared,
			AttackTarget: game.AttackTarget{Player: game.Player2},
		},
	}
	if len(curse.TriggeredAbilities) != 1 {
		t.Fatalf("%s triggered abilities = %d, want 1", curse.Name, len(curse.TriggeredAbilities))
	}
	pattern := curse.TriggeredAbilities[0].Trigger.Pattern
	if pattern.Event != game.EventAttackerDeclared || !pattern.AttackedPlayerIsSourceEnchantedPlayer || !pattern.OneOrMore {
		t.Fatalf("%s trigger pattern = %+v, want attacker-declared enchanted-player once-per-combat", curse.Name, pattern)
	}
	content := curse.TriggeredAbilities[0].Content
	if len(content.Modes) != 1 {
		t.Fatalf("%s content modes = %d, want 1", curse.Name, len(content.Modes))
	}
	log := &TurnLog{}
	for i := range content.Modes[0].Sequence {
		engine.resolveInstructionWithChoices(g, obj, &content.Modes[0].Sequence[i], [game.NumPlayers]PlayerAgent{}, log)
	}
}

// TestGeneratedCurseOfOpulenceCreatesGoldForControllerAndAttackers proves the
// generated Curse of Opulence resolves its "Whenever enchanted player is
// attacked, create a Gold token. Each opponent attacking that player does the
// same." ability to one Gold token for the controller and one for each opponent
// attacking the enchanted player, and none for the enchanted player or for an
// opponent attacking someone else. This is the end-to-end proof that the folded
// reflexive rider lowers to the reusable opponents-attacking group.
func TestGeneratedCurseOfOpulenceCreatesGoldForControllerAndAttackers(t *testing.T) {
	g := resolveCurseAttackTrigger(t, cardsc.CurseOfOpulence())

	names := tokenNamesByController(g)
	if got := names[game.Player1]; len(got) != 1 || got[0] != "Gold" {
		t.Fatalf("controller Player1 tokens = %v, want one Gold", got)
	}
	if got := names[game.Player3]; len(got) != 1 || got[0] != "Gold" {
		t.Fatalf("attacking opponent Player3 tokens = %v, want one Gold", got)
	}
	if got := names[game.Player4]; len(got) != 1 || got[0] != "Gold" {
		t.Fatalf("attacking opponent Player4 tokens = %v, want one Gold", got)
	}
	if got := names[game.Player2]; len(got) != 0 {
		t.Fatalf("enchanted Player2 tokens = %v, want none", got)
	}

	// The created token is the Gold artifact, not a placeholder.
	for _, permanent := range g.Battlefield {
		if !permanent.Token {
			continue
		}
		def := permanent.TokenDef
		if !def.HasType(types.Artifact) || !def.HasSubtype(types.Gold) {
			t.Fatalf("Gold token def = %v/%v, want Artifact Gold", def.Types, def.Subtypes)
		}
	}
}

// TestGeneratedCurseOfDisturbanceCreatesZombiesForControllerAndAttackers proves
// the generated Curse of Disturbance resolves the same reflexive shape with its
// 2/2 black Zombie token: the controller and each opponent attacking the
// enchanted player each get one Zombie, confirming the "does the same" rider is
// generic across the token the curse creates rather than specific to Gold.
func TestGeneratedCurseOfDisturbanceCreatesZombiesForControllerAndAttackers(t *testing.T) {
	g := resolveCurseAttackTrigger(t, cardsc.CurseOfDisturbance())

	names := tokenNamesByController(g)
	for _, player := range []game.PlayerID{game.Player1, game.Player3, game.Player4} {
		if got := names[player]; len(got) != 1 || got[0] != "Zombie" {
			t.Fatalf("player %v tokens = %v, want one Zombie", player, got)
		}
	}
	if got := names[game.Player2]; len(got) != 0 {
		t.Fatalf("enchanted Player2 tokens = %v, want none", got)
	}

	for _, permanent := range g.Battlefield {
		if !permanent.Token {
			continue
		}
		def := permanent.TokenDef
		if !def.HasType(types.Creature) || !def.HasSubtype(types.Zombie) {
			t.Fatalf("Zombie token def = %v/%v, want Creature Zombie", def.Types, def.Subtypes)
		}
		if !def.Power.Exists || def.Power.Val.Value != 2 || !def.Toughness.Exists || def.Toughness.Val.Value != 2 {
			t.Fatalf("Zombie token P/T = %v/%v, want 2/2", def.Power, def.Toughness)
		}
	}
}
