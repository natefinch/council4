package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/cards"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
)

// Source review: exact BFZ #84 printing f9ae6f63-d1b2-4bf9-a423-ed6fbb540d4b,
// Oracle ID e9d62416-c4b1-4982-9593-174d46e975c7; pinned corpus SHA-256
// 6da9ced0e19bd384771ab8f14051cc45a8dc90302082b875fa6e039d2399a1b4. Current
// Scryfall record verified 2026-11-19 at
// https://api.scryfall.com/cards/f9ae6f63-d1b2-4bf9-a423-ed6fbb540d4b:
// {U} Sorcery. Tap target creature. It doesn't untap during its controller's
// next untap step. Awaken 3—{4}{U} (If cast for {4}{U}, also put three +1/+1
// counters on target land you control and it becomes a 0/0 Elemental creature
// with haste. It's still a land.) Rulings URI
// https://api.scryfall.com/cards/f9ae6f63-d1b2-4bf9-a423-ed6fbb540d4b/rulings
// returned eight rulings, all dated 2015-08-25: an already-tapped creature may
// be targeted and is still frozen; the effect tracks the creature, not its
// controller; Awaken is optional and its land target is required only when the
// alternative is paid; the land retains its other types, subtypes, mana
// abilities and color. The raw Scryfall responses are retained under
// .cardwork/rush-of-ice-{scryfall,rulings}.json.
// Official Wizards Comprehensive Rules effective 2026-11-06, verified in
// .cardwork/MagicCompRules-20261106.txt (SHA-256
// 524f5195a1e5f54bb13c735edf399ac717a4458e2713f398d1a3e6877186b829) and at
// https://media.wizards.com/2026/downloads/MagicCompRules%2020261106.txt. Read
// 115.1a; 601.2b-c/e/f/h; 608.2b-c; 502.3; 701.26a; 122.1a; 205.1b;
// 611.2a; 613.1d/f/g and 613.4b-c; 702.10b-c; 702.113a-b. These require the
// first creature target on either cast branch, the land target only for Awaken,
// text-order resolution with illegal targets ignored, one skipped untap for the
// tracked creature, and the land's permanent 0/0 Elemental/haste effect with
// three counters while retaining its land identity.
func rushOfIceEngine(t *testing.T) (*Engine, *game.CardDef) {
	t.Helper()
	registry := cards.NewDefaultRegistry()
	def, err := registry.LookupOracleID("e9d62416-c4b1-4982-9593-174d46e975c7")
	if err != nil {
		t.Fatalf("lookup Rush of Ice by Oracle ID: %v", err)
	}
	if def.Name != "Rush of Ice" {
		t.Fatalf("Oracle ID resolved to %q, want Rush of Ice", def.Name)
	}
	return NewEngineWithPrograms(nil, registry.Programs()), def
}

func TestRushOfIceScenario(t *testing.T) {
	t.Run("Awaken taps and freezes a creature and animates a land", func(t *testing.T) {
		engine, def := rushOfIceEngine(t)
		g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
		spellID := addCardToHand(g, game.Player1, def)
		target := addCombatPermanent(g, game.Player2, vanillaCreature("Rush of Ice target", 2, 2))
		land := addBasicLandPermanent(g, game.Player1, types.Forest)
		g.Players[game.Player1].ManaPool.Add(mana.C, 4)
		g.Players[game.Player1].ManaPool.Add(mana.U, 1)
		setMainPhasePriority(g, game.Player1)

		cast := action.CastSpell(spellID, []game.Target{
			game.PermanentTarget(target.ObjectID),
			game.PermanentTarget(land.ObjectID),
		}, 0, nil)
		if !containsAction(engine.legalActions(g, game.Player1), cast) {
			t.Fatal("Awaken cast with {4}{U}, a creature, and a controlled Forest is not legal")
		}
		if !engine.applyAction(g, game.Player1, cast) {
			t.Fatal("Awaken cast failed")
		}
		spell, ok := g.Stack.Peek()
		if !ok || spell.AlternativeCost != 1 || spell.ManaSpentToCast != 5 ||
			len(spell.Targets) != 2 || spell.Targets[0].PermanentID != target.ObjectID ||
			spell.Targets[1].PermanentID != land.ObjectID {
			t.Fatalf("cast spell = %+v, want Awaken alternative, five mana, creature then land targets", spell)
		}
		engine.resolveTopOfStack(g, &TurnLog{})

		if !target.Tapped || !target.SkipsNextUntap {
			t.Fatal("Rush of Ice did not tap its target and prevent its next untap")
		}
		if land.Counters.Get(counter.PlusOnePlusOne) != 3 ||
			!permanentHasType(g, land, types.Land) || !permanentHasType(g, land, types.Creature) ||
			!permanentHasSubtype(g, land, types.Forest) || !permanentHasSubtype(g, land, types.Elemental) {
			t.Fatal("Awakened Forest did not gain three counters and become a land Elemental creature")
		}
		if got := effectivePower(g, land); got != 3 {
			t.Fatalf("Awakened Forest power = %d, want 3", got)
		}
		if got := effectiveToughnessMust(t, g, land); got != 3 {
			t.Fatalf("Awakened Forest toughness = %d, want 3", got)
		}
		if !hasKeyword(g, land, game.Haste) {
			t.Fatal("Awakened Forest does not have haste")
		}

		g.Turn.ActivePlayer = game.Player2
		g.Turn.PriorityPlayer = game.Player2
		addCardToLibrary(g, game.Player2, &game.CardDef{CardFace: game.CardFace{Name: "Rush of Ice untap-step draw"}})
		engine.runBeginningPhase(g, [game.NumPlayers]PlayerAgent{}, &TurnLog{})
		if !target.Tapped || target.SkipsNextUntap {
			t.Fatal("target did not stay tapped through its controller's next untap step")
		}
	})

	t.Run("mana cost needs only the creature target", func(t *testing.T) {
		engine, def := rushOfIceEngine(t)
		g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
		spellID := addCardToHand(g, game.Player1, def)
		target := addCombatPermanent(g, game.Player2, vanillaCreature("Rush of Ice target", 2, 2))
		land := addBasicLandPermanent(g, game.Player1, types.Forest)
		g.Players[game.Player1].ManaPool.Add(mana.U, 1)
		setMainPhasePriority(g, game.Player1)

		cast := action.CastSpell(spellID, []game.Target{game.PermanentTarget(target.ObjectID)}, 0, nil)
		if !containsAction(engine.legalActions(g, game.Player1), cast) {
			t.Fatal("mana-cost cast should be legal with only its creature target")
		}
		if !engine.applyAction(g, game.Player1, cast) {
			t.Fatal("mana-cost cast failed")
		}
		spell, ok := g.Stack.Peek()
		if !ok || spell.AlternativeCost != 0 || spell.ManaSpentToCast != 1 ||
			len(spell.Targets) != 1 || spell.Targets[0].PermanentID != target.ObjectID {
			t.Fatalf("cast spell = %+v, want mana cost, one mana, one creature target", spell)
		}
		engine.resolveTopOfStack(g, &TurnLog{})

		if !target.Tapped || !target.SkipsNextUntap {
			t.Fatal("mana-cost Rush of Ice did not tap and freeze its creature target")
		}
		if land.Counters.Get(counter.PlusOnePlusOne) != 0 || permanentHasType(g, land, types.Creature) {
			t.Fatal("mana-cost Rush of Ice applied its Awaken effect")
		}
	})
}
