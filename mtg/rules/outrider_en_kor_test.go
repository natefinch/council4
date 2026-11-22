package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/cards"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/types"
)

// Source review: pinned Oracle corpus SHA-256
// 6da9ced0e19bd384771ab8f14051cc45a8dc90302082b875fa6e039d2399a1b4.
// Outrider en-Kor, Oracle 644b7aba-a7b8-4861-9e83-73329cf85be2, printing
// de9b2506-ead4-49ff-a733-66f3f13dbe17: {2}{W}, Creature — Kor Rebel Knight,
// 2/2; "Flanking (Whenever a creature without flanking blocks this creature,
// the blocking creature gets -1/-1 until end of turn.)" and "{0}: The next 1
// damage that would be dealt to this creature this turn is dealt to target
// creature you control instead."
// The exact Scryfall printing and rulings response were fetched 2026-11-21
// from https://api.scryfall.com/cards/de9b2506-ead4-49ff-a733-66f3f13dbe17
// and https://api.scryfall.com/cards/de9b2506-ead4-49ff-a733-66f3f13dbe17/rulings
// (HTTP 200; three rulings). The rulings specify that when multiple sources
// deal damage simultaneously, the controller chooses which source's damage is
// redirected; the original source deals the redirected damage; and no
// redirection occurs if the target has left the battlefield or is no longer a
// creature or planeswalker. Raw responses are retained under
// .cardwork/sources/en-kor/printing-de9b2506-ead4-49ff-a733-66f3f13dbe17.json
// and rulings-de9b2506-ead4-49ff-a733-66f3f13dbe17.json.
// Official Comprehensive Rules effective 2026-11-06,
// https://media.wizards.com/2026/downloads/MagicCompRules%2020261106.txt,
// verified local copy .cardwork/sources/en-kor/MagicCompRules-20261106.txt,
// SHA-256 524f5195a1e5f54bb13c735edf399ac717a4458e2713f398d1a3e6877186b829.
// Read §§601.2c, 602.1a, 602.2a-b, 509.3d, 510.1a-d, 510.2, 614.1a, 614.5,
// 614.9, and 702.25a-b: {0} is the activation cost and its creature target is
// chosen at activation; flanking triggers when a creature without flanking
// blocks Outrider and gives that blocker -1/-1; the shield redirects only the
// next 1 damage to the chosen creature, preserving the original damage event.
const outriderEnKorTestOracleID = "644b7aba-a7b8-4861-9e83-73329cf85be2"

func TestOutriderEnKorFlankingAndRedirect(t *testing.T) {
	registry := cards.NewDefaultRegistry()
	def, err := registry.LookupOracleID(outriderEnKorTestOracleID)
	if err != nil {
		t.Fatalf("exact Outrider en-Kor lookup failed: %v", err)
	}
	if def.Name != "Outrider en-Kor" {
		t.Fatalf("exact Outrider en-Kor lookup returned %q", def.Name)
	}

	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	outrider := addCombatPermanent(g, game.Player1, def)
	recipient := addCombatPermanent(g, game.Player1, redirectTestCard(
		"Outrider en-Kor's chosen creature", []types.Card{types.Creature}, 1, 5,
	))
	blocker := addCombatPermanent(g, game.Player2, redirectTestCard(
		"Nonflanking blocker", []types.Card{types.Creature}, 2, 4,
	))
	engine := NewEngine(nil)
	agents := [game.NumPlayers]PlayerAgent{}

	activateRedirect(t, engine, g, game.Player1, outrider, 0,
		[]game.Target{game.PermanentTarget(recipient.ObjectID)}, agents)
	resolveStackWithTriggers(engine, g, agents)

	outrider.SummoningSick = false
	g.Turn.ActivePlayer = game.Player1
	g.Turn.PriorityPlayer = game.Player1
	g.Turn.Phase = game.PhaseCombat
	g.Turn.Step = game.StepDeclareAttackers
	g.Combat = &game.CombatState{}
	declaration := action.DeclareAttackersAction{Attackers: []game.AttackDeclaration{{
		Attacker: outrider.ObjectID,
		Target:   game.AttackTarget{Player: game.Player2},
	}}}
	defender := &bandBlockingAgent{blocks: []game.BlockDeclaration{{
		Blocker:  blocker.ObjectID,
		Blocking: outrider.ObjectID,
	}}}
	policies := simPassPolicies()
	policies[game.Player2] = defender
	after, ok := NewEngine(nil).Simulator().ResolveCombatWithAttackers(g, game.Player1, declaration, policies)
	if !ok {
		t.Fatal("Outrider en-Kor combat was not accepted")
	}

	afterOutrider, ok := permanentByObjectID(after, outrider.ObjectID)
	if !ok {
		t.Fatal("Outrider en-Kor left the battlefield")
	}
	afterRecipient, ok := permanentByObjectID(after, recipient.ObjectID)
	if !ok {
		t.Fatal("chosen creature left the battlefield")
	}
	afterBlocker, ok := permanentByObjectID(after, blocker.ObjectID)
	if !ok {
		t.Fatal("flanked blocker left the battlefield")
	}
	power := effectivePower(after, afterBlocker)
	toughness, _ := effectiveToughness(after, afterBlocker)
	if power != 1 || toughness != 3 || afterBlocker.MarkedDamage != 2 {
		t.Fatalf("flanked blocker is %d/%d with %d damage, want 1/3 with 2 damage",
			power, toughness, afterBlocker.MarkedDamage)
	}
	if afterOutrider.MarkedDamage != 0 || afterRecipient.MarkedDamage != 1 ||
		len(after.PreventionShields) != 0 {
		t.Fatalf("after combat: Outrider damage %d, chosen creature damage %d, shields %+v; want 0, 1, none",
			afterOutrider.MarkedDamage, afterRecipient.MarkedDamage, after.PreventionShields)
	}
}
