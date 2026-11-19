package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/cards"
	"github.com/natefinch/council4/mtg/cards/authored"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/mana"
)

// Source review: Oracle ID 4d2abfe1-f5f8-4ed5-915d-bfe98c6321ee, Night Clubber,
// pinned printing 59fd3f8a-a13b-4a0a-bb27-b9246949ea7b; current Scryfall Oracle
// text and rulings retrieved 2026-11-19 from
// https://api.scryfall.com/cards/59fd3f8a-a13b-4a0a-bb27-b9246949ea7b and
// /rulings. All five WotC rulings dated 2022-04-29 confirm ordinary timing,
// sacrifice only if the original permanent remains, no required attack, copies
// don't inherit Blitz, and draw on any death.
// Official CR effective 2026-09-25:
// https://media.wizards.com/2026/downloads/MagicCompRules%2020260925.txt
// (verified local SHA-256 8d860e451f20f38865b725b42d82feb714c725373dd8f3b32b8652b3eeb070ca).
// Read 603.2/3 (the entry ability triggers and resolves independently),
// 611.2a/c (duration through end of turn; affected set determined as the effect
// begins), 613.4c and 704.5f (layer-7c P/T modification and zero-toughness SBA),
// 118.9, 601.2b/f-h and 702.152a (Blitz alternative payment, haste, dies-draw,
// and next-end-step sacrifice), 702.10b/302.6, 513.1, 514.2, and 121.1.
func TestNightClubberBlitzEntryGivesOpponentsMinusOneUntilEndOfTurnAndDraws(t *testing.T) {
	registry := cards.NewDefaultRegistry()
	def, err := registry.LookupOracleID(authored.NightClubber.OracleID)
	if err != nil || def.Name != "Night Clubber" {
		t.Fatalf("checked Night Clubber identity lookup = %v, %v", def, err)
	}

	engine := NewEngineWithPrograms(nil, registry.Programs())
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	setMainPhasePriority(g, game.Player1)
	ownCreature := addCombatPermanent(g, game.Player1, vanillaCreature("Own Creature", 2, 2))
	opponentTwo := addCombatPermanent(g, game.Player2, vanillaCreature("Opponent Two", 2, 2))
	opponentThree := addCombatPermanent(g, game.Player3, vanillaCreature("Opponent Three", 2, 2))
	opponentFour := addCombatPermanent(g, game.Player4, vanillaCreature("Opponent Four", 1, 1))
	card := addCardToHand(g, game.Player1, def)
	draw := addCardToLibraryNamed(g, game.Player1, "Night Clubber Blitz draw")
	g.Players[game.Player1].ManaPool.Add(mana.B, 1)
	g.Players[game.Player1].ManaPool.Add(mana.C, 2)

	if !engine.applyActionWithChoices(g, game.Player1, action.CastSpell(card, nil, 0, nil), blitzAgents("Blitz"), &TurnLog{}) {
		t.Fatal("could not cast Night Clubber for its Blitz cost")
	}
	spell, ok := g.Stack.Peek()
	if !ok || !spell.Blitzed || g.Players[game.Player1].ManaPool.Total() != 0 {
		t.Fatalf("Night Clubber did not pay its {2}{B} Blitz cost: spell=%+v mana=%d", spell, g.Players[game.Player1].ManaPool.Total())
	}
	engine.resolveTopOfStack(g, &TurnLog{})
	permanent := permanentForCard(g, card)
	if permanent == nil || !hasKeyword(g, permanent, game.Haste) {
		t.Fatal("Blitzed Night Clubber did not enter with haste")
	}
	if !engine.putTriggeredAbilitiesOnStack(g) {
		t.Fatal("Night Clubber's entry ability did not trigger")
	}
	trigger, ok := g.Stack.Peek()
	if !ok || trigger.Kind != game.StackTriggeredAbility || len(trigger.Targets) != 0 {
		t.Fatalf("Night Clubber entry ability should be nontargeted: %+v", trigger)
	}
	engine.resolveTopOfStack(g, &TurnLog{})

	for _, creature := range []*game.Permanent{opponentTwo, opponentThree} {
		if got := effectivePower(g, creature); got != 1 {
			t.Fatalf("opposing creature power = %d, want 1", got)
		}
		if got, valid := effectiveToughness(g, creature); !valid || got != 1 {
			t.Fatalf("opposing creature toughness = %d (valid=%t), want 1", got, valid)
		}
	}
	if effectivePower(g, ownCreature) != 2 {
		t.Fatal("Night Clubber affected a creature its controller controls")
	}
	engine.applyStateBasedActions(g)
	if _, ok := permanentByObjectID(g, opponentFour.ObjectID); ok {
		t.Fatal("the opposing 1/1 survived Night Clubber's -1/-1 effect")
	}

	blitzEndStep(engine, g)
	resolveStackWithTriggers(engine, g, [game.NumPlayers]PlayerAgent{})
	if !g.Players[game.Player1].Graveyard.Contains(card) || !g.Players[game.Player1].Hand.Contains(draw) {
		t.Fatal("Night Clubber's Blitz sacrifice did not cause its death-draw trigger")
	}
	for _, creature := range []*game.Permanent{opponentTwo, opponentThree} {
		if got := effectivePower(g, creature); got != 1 {
			t.Fatalf("Night Clubber's effect ended before cleanup: opposing creature power = %d, want 1", got)
		}
	}
	engine.runCleanupStep(g, [game.NumPlayers]PlayerAgent{}, &TurnLog{})
	for _, creature := range []*game.Permanent{opponentTwo, opponentThree} {
		if got := effectivePower(g, creature); got != 2 {
			t.Fatalf("Night Clubber's effect remained after cleanup: opposing creature power = %d, want 2", got)
		}
	}
}
