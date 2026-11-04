package rules

import (
	"slices"
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/mana"
)

// Source review: issue #3600. Official CR effective 2026-09-25 (verified local
// copy), read 302.6, 500.8, 501.1, 502.1-502.4 and 505.1a. 302.6: a creature
// can't attack or activate a {T}/{Q} ability unless its controller has
// controlled it continuously since their most recent turn began. 500.8: effects
// add phases to a turn after a specified phase; 501.1/502.3: a beginning phase's
// untap step untaps the active player's permanents. An added beginning phase
// starts no turn, so it cannot end summoning sickness. Cards (pinned corpus):
// Sphinx of the Second Sun (oracle be2d74cd-76fb-4bca-99b5-07749aef4d06,
// {6}{U}{U} 6/6) "Flying / At the beginning of each of your postcombat main
// phases, there is an additional beginning phase after this phase." Rulings
// 2020-11-10: in the additional untap step your tapped permanents untap; it all
// happens during the current turn, so "until your next turn" effects don't
// expire. Shadow of the Second Sun (oracle
// 0da72736-574d-4d99-98ba-3a91c374cd10) has the same enchanted-player trigger
// and rulings (2024-06-07). Llanowar Elves (oracle
// 68954295-54e3-4303-a6bc-fc4547a4e3a3) "{T}: Add {G}." Control Magic (oracle
// cd0d7141-46d2-4aa3-bc77-6b3b4513803e) "Enchant creature / You control
// enchanted creature."

// TestSphinxExtraBeginningPhaseKeepsSummoningSickness runs Sphinx of the Second
// Sun's real postcombat trigger and the additional beginning phase it adds.
// Player1 casts Llanowar Elves and steals Player2's Elves with Control Magic
// during the turn. The added untap step untaps all three of Player1's tapped
// creatures, but only the Sphinx, controlled since the turn began, may attack;
// the entered and stolen Elves remain summoning sick until Player1's next turn.
func TestSphinxExtraBeginningPhaseKeepsSummoningSickness(t *testing.T) {
	g := sicknessMainPhaseGame(game.Player2)
	engine := choicesProgramEngine(t)
	sphinx := addCombatPermanent(g, game.Player1, registryCardDef(t, "Sphinx of the Second Sun"))
	stolen := addCombatPermanent(g, game.Player2, registryCardDef(t, "Llanowar Elves"))
	beginSicknessTurn(g, engine, game.Player1)

	castAndResolveRegistered(t, g, engine, game.Player1, "Control Magic", mana.U, 4, stolen)
	if effectiveController(g, stolen) != game.Player1 {
		t.Fatal("Control Magic did not give Player1 control of Player2's Elves")
	}
	entered := castRegisteredPermanent(t, g, engine, game.Player1, "Llanowar Elves", mana.G, 1)
	for _, creature := range []*game.Permanent{sphinx, stolen, entered} {
		creature.Tapped = true
	}

	addCardToLibrary(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Draw"}})
	log := TurnLog{}
	g.Turn.MainPhasesThisTurn = 1
	engine.runMainPhase(g, allPassAgents(), game.PhasePostcombatMain, &log)
	if !slices.Equal(g.Turn.ExtraPhases, []game.Phase{game.PhaseBeginning}) {
		t.Fatalf("Sphinx's postcombat trigger queued %v, want one beginning phase", g.Turn.ExtraPhases)
	}
	engine.runExtraPhases(g, allPassAgents(), &log)
	if got := phaseOrder(log); !slices.Equal(got, []game.Phase{game.PhasePostcombatMain, game.PhaseBeginning}) {
		t.Fatalf("phases = %v, want postcombat main then the additional beginning phase", got)
	}
	for _, creature := range []*game.Permanent{sphinx, stolen, entered} {
		if creature.Tapped {
			t.Fatalf("%s did not untap in the additional untap step", permanentName(g, creature))
		}
	}

	setSicknessMainPhase(g, game.Player1)
	g.Turn.MainPhasesThisTurn = 2
	if permanentSummoningSick(g, sphinx) || !creatureCanAttack(t, g, engine, game.Player1, sphinx) {
		t.Fatal("the Sphinx controlled since the turn began cannot attack")
	}
	for _, creature := range []*game.Permanent{stolen, entered} {
		if !permanentSummoningSick(g, creature) {
			t.Fatalf("the additional beginning phase cured %s's summoning sickness", permanentName(g, creature))
		}
		if creatureCanAttack(t, g, engine, game.Player1, creature) {
			t.Fatalf("%s attacked after the additional beginning phase without haste", permanentName(g, creature))
		}
	}
	if elvesPayForElves(t, g, engine, game.Player1) {
		t.Fatal("an Elves gained this turn tapped for mana after the additional beginning phase")
	}
	view := NewObservation(g, game.Player1).Battlefield()
	for _, creature := range []*game.Permanent{stolen, entered} {
		if i := slices.IndexFunc(view, func(p PermanentView) bool { return p.ObjectID == creature.ObjectID }); i < 0 || !view[i].SummoningSick {
			t.Fatalf("the observation did not report %s as summoning sick", permanentName(g, creature))
		}
	}

	beginSicknessTurn(g, engine, game.Player2)
	beginSicknessTurn(g, engine, game.Player1)
	for _, creature := range []*game.Permanent{stolen, entered} {
		if !creatureCanAttack(t, g, engine, game.Player1, creature) {
			t.Fatalf("%s controlled since Player1's next turn began cannot attack", permanentName(g, creature))
		}
	}
	if !elvesPayForElves(t, g, engine, game.Player1) {
		t.Fatal("an Elves controlled since Player1's next turn began cannot tap for mana")
	}
}

// castRegisteredPermanent casts and resolves an untargeted registered permanent
// spell and returns the permanent it became.
func castRegisteredPermanent(t *testing.T, g *game.Game, engine *Engine, caster game.PlayerID, name string, color mana.Color, amount int) *game.Permanent {
	t.Helper()
	spell := addCardToHand(g, caster, registryCardDef(t, name))
	g.Players[caster].ManaPool.Add(color, amount)
	g.Turn.PriorityPlayer = caster
	cast := action.CastSpell(spell, nil, 0, nil)
	if !engine.applyAction(g, caster, cast) {
		t.Fatalf("casting %s failed", name)
	}
	engine.resolveTopOfStack(g, &TurnLog{})
	i := slices.IndexFunc(g.Battlefield, func(p *game.Permanent) bool { return p.CardInstanceID == spell })
	if i < 0 {
		t.Fatalf("%s did not enter the battlefield", name)
	}
	return g.Battlefield[i]
}
