package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
)

// Source review: issue #3601. Official CR effective 2026-09-25
// (https://media.wizards.com/2026/downloads/MagicCompRules%2020260925.txt),
// read 302.6, 305.6, 305.7, 605.3a, 611.2c and 702.10b-c. 305.6: a land with a
// basic land type has the intrinsic ability "{T}: Add [mana symbol]". 302.6: a
// creature's activated ability with {T} in its cost can't be activated unless
// the creature has been under its controller's control continuously since
// their most recent turn began. 702.10c: haste lifts that restriction. 605.3a:
// mana abilities may be activated while paying a cost, so automatic payment
// must offer only sources whose mana ability is activatable. Cards (pinned
// corpus): Forest (oracle b34bb2dc-c1af-4d77-b0b3-a0fb342a5fc6) "({T}: Add
// {G}.)" (rulings: none). Vivify (oracle 7b58e277-a969-4c7f-8143-ee4f6bd1e3ed,
// {2}{G} instant) "Target land becomes a 3/3 creature until end of turn. It's
// still a land. Draw a card." (ruling 2008-08-01: a noncreature permanent that
// turns into a creature can use its {T} abilities only if its controller has
// controlled it continuously since their most recent turn began; how long it
// has been a creature doesn't matter). Expedite (oracle
// 3501a839-eef5-44e4-8637-b5754780454e, {R} instant) "Target creature gains
// haste until end of turn. Draw a card." (rulings: none). Llanowar Elves,
// Control Magic and Word of Seizing are reviewed in
// summoning_sickness_control_test.go.

// playRegisteredForest plays a default-registry Forest from player's hand
// through the land-play action and returns the new permanent.
func playRegisteredForest(t *testing.T, g *game.Game, engine *Engine, player game.PlayerID) *game.Permanent {
	t.Helper()
	cardID := addCardToHand(g, player, registryCardDef(t, "Forest"))
	if !engine.applyAction(g, player, action.PlayLand(cardID)) {
		t.Fatal("playing the Forest failed")
	}
	forest, ok := findPermanentByCardID(g, cardID)
	if !ok {
		t.Fatal("the played Forest is not on the battlefield")
	}
	return forest
}

// animateWithVivify resolves the registered Vivify on land for caster and
// checks it became a summoning-sickness-relevant creature land.
func animateWithVivify(t *testing.T, g *game.Game, engine *Engine, caster game.PlayerID, land *game.Permanent) {
	t.Helper()
	addCardToLibrary(g, caster, &game.CardDef{CardFace: game.CardFace{Name: "Draw"}})
	castAndResolveRegistered(t, g, engine, caster, "Vivify", mana.G, 3, land)
	if !permanentHasType(g, land, types.Creature) || !permanentHasType(g, land, types.Land) {
		t.Fatal("Vivify did not make the Forest a land creature")
	}
	if !g.Players[caster].ManaPool.IsEmpty() {
		t.Fatal("fixture: mana is still floating after Vivify")
	}
}

// forestPaysForElves reports whether player may cast a registered Llanowar
// Elves using forest as the only mana source. The legal-action enumeration and
// the cast driver must agree; a successful cast must actually tap the Forest.
func forestPaysForElves(t *testing.T, g *game.Game, engine *Engine, player game.PlayerID, forest *game.Permanent) bool {
	t.Helper()
	g.Turn.PriorityPlayer = player
	spell := addCardToHand(g, player, registryCardDef(t, "Llanowar Elves"))
	cast := action.CastSpell(spell, nil, 0, nil)
	offered := containsAction(engine.legalActions(g, player), cast)
	accepted := engine.applyAction(g, player, cast)
	if offered != accepted {
		t.Fatalf("cast enumeration offered %v but the cast driver accepted %v", offered, accepted)
	}
	if !accepted {
		if forest.Tapped {
			t.Fatal("a rejected cast tapped the Forest")
		}
		g.Players[player].Hand.Remove(spell)
		return false
	}
	if !forest.Tapped {
		t.Fatal("the Elves were cast without tapping the Forest")
	}
	top, ok := g.Stack.Peek()
	if !ok || top.SourceID != spell {
		t.Fatal("Llanowar Elves is not on the stack after casting")
	}
	return true
}

// TestVivifiedNewlyPlayedForestCannotPayWhileSick is the #3601 scenario: a
// Forest played this turn and animated by Vivify is a creature that hasn't been
// controlled since the turn began, so its {T}: Add {G} can't pay for Llanowar
// Elves, either automatically or by direct activation.
func TestVivifiedNewlyPlayedForestCannotPayWhileSick(t *testing.T) {
	g := sicknessMainPhaseGame(game.Player1)
	engine := choicesProgramEngine(t)
	forest := playRegisteredForest(t, g, engine, game.Player1)
	animateWithVivify(t, g, engine, game.Player1, forest)
	if !permanentSummoningSick(g, forest) {
		t.Fatal("fixture: the newly played Forest is not summoning sick")
	}

	if forestPaysForElves(t, g, engine, game.Player1, forest) {
		t.Fatal("the summoning-sick animated Forest paid for Llanowar Elves")
	}
	if engine.applyAction(g, game.Player1, action.ActivateAbility(forest.ObjectID, 0, nil, 0)) {
		t.Fatal("the summoning-sick animated Forest's mana ability was activated")
	}
}

// TestNewlyPlayedForestPaysWhileNotACreature is the control: summoning
// sickness only restricts creatures, so the same un-animated Forest pays.
func TestNewlyPlayedForestPaysWhileNotACreature(t *testing.T) {
	g := sicknessMainPhaseGame(game.Player1)
	engine := choicesProgramEngine(t)
	forest := playRegisteredForest(t, g, engine, game.Player1)

	if !forestPaysForElves(t, g, engine, game.Player1, forest) {
		t.Fatal("the newly played noncreature Forest could not pay")
	}
}

// TestExpeditedVivifiedForestPaysWithHaste gives the newly played animated
// Forest haste with the registered Expedite, so it may tap for mana (CR
// 702.10c).
func TestExpeditedVivifiedForestPaysWithHaste(t *testing.T) {
	g := sicknessMainPhaseGame(game.Player1)
	engine := choicesProgramEngine(t)
	forest := playRegisteredForest(t, g, engine, game.Player1)
	animateWithVivify(t, g, engine, game.Player1, forest)
	addCardToLibrary(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Draw"}})
	castAndResolveRegistered(t, g, engine, game.Player1, "Expedite", mana.R, 1, forest)
	if !hasKeyword(g, forest, game.Haste) {
		t.Fatal("Expedite did not give the animated Forest haste")
	}

	if !forestPaysForElves(t, g, engine, game.Player1, forest) {
		t.Fatal("the hasty animated Forest could not pay")
	}
}

// TestVivifiedEstablishedForestPays animates a Forest its controller has
// controlled since the turn began; per Vivify's ruling it doesn't matter how
// long it has been a creature, so it pays.
func TestVivifiedEstablishedForestPays(t *testing.T) {
	g := sicknessMainPhaseGame(game.Player1)
	engine := choicesProgramEngine(t)
	forest := addCombatPermanent(g, game.Player1, registryCardDef(t, "Forest"))
	controlledSinceTurnBegan(g, forest)
	animateWithVivify(t, g, engine, game.Player1, forest)

	if !forestPaysForElves(t, g, engine, game.Player1, forest) {
		t.Fatal("the established animated Forest could not pay")
	}
}

// TestStolenVivifiedForestSickness has Player2 animate Player1's established
// Forest with Vivify and then take it. Under Control Magic it hasn't been
// controlled by Player2 since their turn began and can't pay; under Word of
// Seizing it has haste and can.
func TestStolenVivifiedForestSickness(t *testing.T) {
	for _, tc := range []struct {
		steal string
		pays  bool
	}{
		{steal: "Control Magic", pays: false},
		{steal: "Word of Seizing", pays: true},
	} {
		t.Run(tc.steal, func(t *testing.T) {
			g := sicknessMainPhaseGame(game.Player2)
			engine := choicesProgramEngine(t)
			forest := addCombatPermanent(g, game.Player1, registryCardDef(t, "Forest"))
			controlledSinceTurnBegan(g, forest)
			animateWithVivify(t, g, engine, game.Player2, forest)
			if tc.steal == "Control Magic" {
				castAndResolveRegistered(t, g, engine, game.Player2, "Control Magic", mana.U, 4, forest)
			} else {
				castAndResolveWordOfSeizing(t, g, engine, game.Player2, forest)
			}
			if effectiveController(g, forest) != game.Player2 {
				t.Fatalf("%s did not give Player2 the animated Forest", tc.steal)
			}
			if !permanentSummoningSick(g, forest) {
				t.Fatal("the stolen animated Forest is not summoning sick")
			}

			if got := forestPaysForElves(t, g, engine, game.Player2, forest); got != tc.pays {
				t.Fatalf("stolen animated Forest paid = %v, want %v", got, tc.pays)
			}
		})
	}
}
