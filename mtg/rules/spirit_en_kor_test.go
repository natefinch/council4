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
// Spirit en-Kor, Oracle f066174a-a959-4365-920f-c04506d94a5a, printing
// 8c57ca87-75b0-4058-acd8-5394545d0846 (Tempest Remastered #34): {3}{W},
// Creature — Kor Spirit, 2/2; "Flying / {0}: The next 1 damage that would be
// dealt to this creature this turn is dealt to target creature you control
// instead." The pinned Oracle record and the Scryfall printing fetched
// 2026-11-21 agree; its exact response is retained at
// .cardwork/sources/en-kor/printing-8c57ca87-75b0-4058-acd8-5394545d0846.json.
// The current Scryfall rulings response (HTTP 200, five WotC rulings) was
// rechecked 2026-11-21 at
// https://api.scryfall.com/cards/8c57ca87-75b0-4058-acd8-5394545d0846/rulings
// and is retained at
// .cardwork/sources/en-kor/rulings-8c57ca87-75b0-4058-acd8-5394545d0846.json.
// Those rulings confirm that redirection preserves combat damage, does not
// change trample assignment to the defending player, can target Spirit itself
// or send damage above the recipient's toughness, and can be activated
// repeatedly before damage.
// Official Comprehensive Rules effective 2026-11-06,
// https://media.wizards.com/2026/downloads/MagicCompRules%2020261106.txt,
// verified local copy .cardwork/sources/en-kor/MagicCompRules-20261106.txt,
// SHA-256 524f5195a1e5f54bb13c735edf399ac717a4458e2713f398d1a3e6877186b829.
// Read §§115.1c and 601.2c/602.2a-b: {0} is the activation cost and its
// creature target is chosen during activation. §§120.1 and 120.3e identify the
// damage source and mark damage on the creature that receives it. §§614.1a,
// 614.5, 614.9 and 616.1 define the "instead" redirection replacement, its
// eligibility and ordering. §702.9b requires flying attackers to be blocked
// only by creatures with flying or reach.
func TestSpiritEnKorFlyingAndRedirectsOneCombatDamage(t *testing.T) {
	registry := cards.NewDefaultRegistry()
	def, err := registry.LookupOracleID("f066174a-a959-4365-920f-c04506d94a5a")
	if err != nil {
		t.Fatalf("exact Spirit en-Kor lookup failed: %v", err)
	}
	if def.Name != "Spirit en-Kor" {
		t.Fatalf("exact Spirit en-Kor lookup returned %q", def.Name)
	}

	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	spirit := addCombatPermanent(g, game.Player1, def)
	recipient := addCombatPermanent(g, game.Player1, redirectTestCard(
		"Spirit en-Kor's chosen creature", []types.Card{types.Creature}, 1, 5,
	))
	ground := addCombatCreaturePermanentWithPower(g, game.Player2, 2)
	flyingBlocker := addCombatCreaturePermanentWithPower(g, game.Player2, 1, game.Flying)
	engine := NewEngine(nil)
	target := game.PermanentTarget(recipient.ObjectID)
	agents := [game.NumPlayers]PlayerAgent{}

	activateRedirect(t, engine, g, game.Player1, spirit, 0, []game.Target{target}, agents)
	resolveStackWithTriggers(engine, g, agents)

	g.Turn.Phase = game.PhaseCombat
	g.Turn.Step = game.StepDeclareBlockers
	g.Combat = &game.CombatState{
		Attackers: []game.AttackDeclaration{{
			Attacker: spirit.ObjectID,
			Target:   game.AttackTarget{Player: game.Player2},
		}},
	}
	groundBlock := mustDeclareBlockersPayload(t, action.DeclareBlockers([]game.BlockDeclaration{
		{Blocker: ground.ObjectID, Blocking: spirit.ObjectID},
	}))
	if engine.applyDeclareBlockers(g, game.Player2, groundBlock) {
		t.Fatal("ground creature blocked Spirit en-Kor")
	}
	flyingBlock := mustDeclareBlockersPayload(t, action.DeclareBlockers([]game.BlockDeclaration{
		{Blocker: flyingBlocker.ObjectID, Blocking: spirit.ObjectID},
	}))
	if !engine.applyDeclareBlockers(g, game.Player2, flyingBlock) {
		t.Fatal("flying creature could not block Spirit en-Kor")
	}

	g.Turn.Step = game.StepCombatDamage
	engine.resolveCombatDamage(g, &TurnLog{})

	if spirit.MarkedDamage != 0 || recipient.MarkedDamage != 1 || len(g.PreventionShields) != 0 {
		t.Fatalf("after combat: Spirit damage %d, chosen creature damage %d, shields %+v; want 0, 1, none",
			spirit.MarkedDamage, recipient.MarkedDamage, g.PreventionShields)
	}
}
