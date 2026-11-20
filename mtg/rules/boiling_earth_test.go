package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/cardprogram"
	"github.com/natefinch/council4/mtg/cards"
	"github.com/natefinch/council4/mtg/cards/authored"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
)

// Source review: exact BFZ #142 printing cdaab44c-4ce1-43fb-915c-c687fe8559ce,
// Oracle ID fb34b671-b61a-47e2-90fa-dbe5cf6d3743; pinned corpus SHA-256
// 6da9ced0e19bd384771ab8f14051cc45a8dc90302082b875fa6e039d2399a1b4. Current
// Scryfall /cards/cdaab44c-4ce1-43fb-915c-c687fe8559ce verified 2026-11-19:
// "{1}{R} Sorcery. Boiling Earth deals 1 damage to each creature your opponents
// control. Awaken 4—{6}{R} (If you cast this spell for {6}{R}, also put four
// +1/+1 counters on target land you control and it becomes a 0/0 Elemental
// creature with haste. It's still a land.)" Its rulings URI
// https://api.scryfall.com/cards/cdaab44c-4ce1-43fb-915c-c687fe8559ce/rulings
// returned six WotC rulings dated 2015-08-25: the Awaken cost is optional; only
// that branch requires the land target and gets both effects; illegal targets
// are unaffected and an illegal sole Awaken target prevents resolution; the
// land retains its other types, subtypes and mana abilities, and gains no color.
// Retrieved card/rulings responses are retained under .cardwork/source-review.
// Official Wizards Comprehensive Rules effective 2026-11-06,
// https://media.wizards.com/2026/downloads/MagicCompRules%2020261106.txt,
// read 601.2b-c/f-h, 608.2b/f, 122.1a, 205.1b, 613.1d/g, 613.4b-c, 702.10b,
// 702.113a-b, 120.1, 120.3e and 120.4b. These rules require choosing the Awaken
// alternative at cast and adding its target only on that branch; if its sole
// target is illegal on resolution, no part resolves. Damage marks each creature
// simultaneously, and Awaken adds four counters, makes the land a 0/0 Elemental
// creature with haste, and leaves it a land.
func boilingEarthEngine(t *testing.T) (*Engine, *game.CardDef) {
	t.Helper()
	catalog, err := cardprogram.NewCatalog(authored.BoilingEarth.Programs)
	if err != nil {
		t.Fatal(err)
	}
	def, err := cards.NewDefaultRegistry().LookupOracleID(authored.BoilingEarth.OracleID)
	if err != nil {
		t.Fatal(err)
	}
	if def.Name != "Boiling Earth" {
		t.Fatalf("registered card name = %q, want Boiling Earth", def.Name)
	}
	return NewEngineWithPrograms(nil, catalog), def
}

func TestBoilingEarthScenario(t *testing.T) {
	t.Run("mana cost damages opposing creatures without a target", func(t *testing.T) {
		engine, def := boilingEarthEngine(t)
		g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
		spellID := addCardToHand(g, game.Player1, def)
		land := addBasicLandPermanent(g, game.Player1, types.Forest)
		ownCreature := addCombatPermanent(g, game.Player1, vanillaCreature("Boiling Earth own creature", 2, 2))
		opponentCreature := addCombatPermanent(g, game.Player2, vanillaCreature("Boiling Earth opponent creature", 2, 2))
		otherOpponentCreature := addCombatPermanent(g, game.Player3, vanillaCreature("Boiling Earth other opponent creature", 2, 2))
		g.Players[game.Player1].ManaPool.Add(mana.C, 1)
		g.Players[game.Player1].ManaPool.Add(mana.R, 1)
		setMainPhasePriority(g, game.Player1)

		cast := action.CastSpell(spellID, nil, 0, nil)
		if !containsAction(engine.legalActions(g, game.Player1), cast) {
			t.Fatal("mana-cost cast should be legal without the Awaken target")
		}
		if !engine.applyAction(g, game.Player1, cast) {
			t.Fatal("mana-cost cast failed")
		}
		spell, ok := g.Stack.Peek()
		if !ok || spell.AlternativeCost != 0 || spell.ManaSpentToCast != 2 || len(spell.Targets) != 0 {
			t.Fatalf("cast spell = %+v, want mana cost, two mana spent, no targets", spell)
		}
		engine.resolveTopOfStack(g, &TurnLog{})

		if opponentCreature.MarkedDamage != 1 || otherOpponentCreature.MarkedDamage != 1 {
			t.Fatalf("opposing creature damage = (%d, %d), want (1, 1)", opponentCreature.MarkedDamage, otherOpponentCreature.MarkedDamage)
		}
		if ownCreature.MarkedDamage != 0 {
			t.Fatalf("own creature marked damage = %d, want 0", ownCreature.MarkedDamage)
		}
		if permanentHasType(g, land, types.Creature) || land.Counters.Get(counter.PlusOnePlusOne) != 0 {
			t.Fatal("mana-cost cast applied the Awaken effect")
		}
		if !g.Players[game.Player1].Graveyard.Contains(spellID) {
			t.Fatal("resolved Boiling Earth did not go to its owner's graveyard")
		}
	})

	t.Run("Awaken alternative damages and animates a land", func(t *testing.T) {
		engine, def := boilingEarthEngine(t)
		g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
		spellID := addCardToHand(g, game.Player1, def)
		land := addBasicLandPermanent(g, game.Player1, types.Forest)
		ownCreature := addCombatPermanent(g, game.Player1, vanillaCreature("Boiling Earth own creature", 2, 2))
		opponentCreature := addCombatPermanent(g, game.Player2, vanillaCreature("Boiling Earth opponent creature", 2, 2))
		otherOpponentCreature := addCombatPermanent(g, game.Player3, vanillaCreature("Boiling Earth other opponent creature", 2, 2))
		g.Players[game.Player1].ManaPool.Add(mana.C, 6)
		g.Players[game.Player1].ManaPool.Add(mana.R, 1)
		setMainPhasePriority(g, game.Player1)

		cast := action.CastSpell(spellID, []game.Target{game.PermanentTarget(land.ObjectID)}, 0, nil)
		if !containsAction(engine.legalActions(g, game.Player1), cast) {
			t.Fatal("Awaken cast with {6}{R} and a controlled Forest is not legal")
		}
		if !engine.applyAction(g, game.Player1, cast) {
			t.Fatal("Awaken cast failed")
		}
		spell, ok := g.Stack.Peek()
		if !ok || spell.AlternativeCost != 1 || spell.ManaSpentToCast != 7 || len(spell.Targets) != 1 {
			t.Fatalf("cast spell = %+v, want Awaken alternative, seven mana spent, one target", spell)
		}
		engine.resolveTopOfStack(g, &TurnLog{})

		if opponentCreature.MarkedDamage != 1 || otherOpponentCreature.MarkedDamage != 1 {
			t.Fatalf("opposing creature damage = (%d, %d), want (1, 1)", opponentCreature.MarkedDamage, otherOpponentCreature.MarkedDamage)
		}
		if ownCreature.MarkedDamage != 0 {
			t.Fatalf("own creature marked damage = %d, want 0", ownCreature.MarkedDamage)
		}
		if land.Counters.Get(counter.PlusOnePlusOne) != 4 ||
			!permanentHasType(g, land, types.Land) || !permanentHasType(g, land, types.Creature) ||
			!permanentHasSubtype(g, land, types.Forest) || !permanentHasSubtype(g, land, types.Elemental) {
			t.Fatal("Awakened Forest did not gain four counters and become a land Elemental creature")
		}
		if got := effectivePower(g, land); got != 4 {
			t.Fatalf("Awakened Forest power = %d, want 4", got)
		}
		if got := effectiveToughnessMust(t, g, land); got != 4 {
			t.Fatalf("Awakened Forest toughness = %d, want 4", got)
		}
		if !hasKeyword(g, land, game.Haste) {
			t.Fatal("Awakened Forest does not have haste")
		}
		if !g.Players[game.Player1].Graveyard.Contains(spellID) {
			t.Fatal("resolved Boiling Earth did not go to its owner's graveyard")
		}
	})
}
