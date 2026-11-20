package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/cards"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// Source review: Oracle ID 16df20e2-5228-4543-84b7-74a3170d9d31, Oath of the
// Gatewatch #102, printing revision 4e269989-bb22-4da4-a374-434a572e8e8f;
// pinned corpus SHA-256 6da9ced0e19bd384771ab8f14051cc45a8dc90302082b875fa6e039d2399a1b4.
// Current Scryfall record fetched 2026-11-20 from
// https://api.scryfall.com/cards/4e269989-bb22-4da4-a374-434a572e8e8f:
// "{4}{R} Sorcery. Surge {1}{R} (You may cast this spell for its surge cost if
// you or a teammate has cast another spell this turn.) Boulder Salvo deals 4
// damage to target creature."
// Rulings fetched the same day from
// https://api.scryfall.com/cards/4e269989-bb22-4da4-a374-434a572e8e8f/rulings:
// four WotC rulings dated 2016-01-22. Surge is only an alternative cost for
// this card; the qualifying prior spell may have resolved, been countered, or
// (for instants) remain on the stack; copies of a Surge instant or sorcery
// inherit the paid Surge cost; paying Surge doesn't change mana cost or mana
// value.
// Official Wizards Comprehensive Rules effective 2026-11-06,
// https://media.wizards.com/2026/downloads/MagicCompRules%2020261106.txt.
// Read 102.3-4, 107.3a, 115.1, 118.9-9d, 601.2b,c,f-h, 608.2b, 120.3a,
// 120.4a-d, 120.5 and 702.117a: Surge replaces the printed cost only after a
// qualifying cast; the spell chooses one creature target on announcement,
// rechecks it on resolution, then deals four damage.
const (
	boulderSalvoTestOracleID    = "16df20e2-5228-4543-84b7-74a3170d9d31"
	fallOfTheTitansTestOracleID = "55e564b4-7c08-4109-bd67-da5f31fe002e"
)

func registeredSurgeDamageCard(t *testing.T, oracleID, name string) *game.CardDef {
	t.Helper()
	def, err := cards.NewDefaultRegistry().LookupOracleID(oracleID)
	if err != nil {
		t.Fatalf("exact %s default-registry lookup failed: %v", name, err)
	}
	if def.Name != name {
		t.Fatalf("exact lookup returned %q, want %q", def.Name, name)
	}
	return def
}

func TestBoulderSalvoSurgeTargetCreatureDamage(t *testing.T) {
	def := registeredSurgeDamageCard(t, boulderSalvoTestOracleID, "Boulder Salvo")
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	setMainPhasePriority(g, game.Player1)
	castSurgeSetupSpell(t, engine, g, game.Player1)
	engine.resolveTopOfStack(g, &TurnLog{})

	target := addCombatPermanent(g, game.Player2, &game.CardDef{CardFace: game.CardFace{
		Name:      "Boulder Salvo target",
		Types:     []types.Card{types.Creature},
		Toughness: opt.Val(game.PT{Value: 5}),
	}})
	cardID := addCardToHand(g, game.Player1, def)
	g.Players[game.Player1].ManaPool.Add(mana.C, 1)
	g.Players[game.Player1].ManaPool.Add(mana.R, 1)
	cast := action.CastSpell(cardID, []game.Target{game.PermanentTarget(target.ObjectID)}, 0, nil)
	if !engine.applyAction(g, game.Player1, cast) {
		t.Fatal("Boulder Salvo was not castable for its Surge cost after the earlier spell")
	}
	spell, ok := g.Stack.Peek()
	if !ok || spell.AlternativeCost != 1 || spell.ManaSpentToCast != 2 ||
		g.Players[game.Player1].ManaPool.Total() != 0 {
		t.Fatalf("Boulder Salvo cast did not pay {1}{R} Surge: spell=%+v, mana=%d",
			spell, g.Players[game.Player1].ManaPool.Total())
	}

	engine.resolveTopOfStack(g, &TurnLog{})
	damaged, present := permanentByObjectID(g, target.ObjectID)
	if !present || damaged.MarkedDamage != 4 {
		t.Fatalf("target creature marked damage = %+v (present %t), want 4",
			damaged, present)
	}
}

// Source review: Oracle ID 55e564b4-7c08-4109-bd67-da5f31fe002e, Oath of the
// Gatewatch #109, printing revision 8c21da73-0dad-4b2b-8cc6-61861bed9410;
// pinned corpus SHA-256 6da9ced0e19bd384771ab8f14051cc45a8dc90302082b875fa6e039d2399a1b4.
// Current Scryfall record fetched 2026-11-20 from
// https://api.scryfall.com/cards/8c21da73-0dad-4b2b-8cc6-61861bed9410:
// "{X}{X}{R} Instant. Surge {X}{R} (You may cast this spell for its surge cost
// if you or a teammate has cast another spell this turn.) Fall of the Titans
// deals X damage to each of up to two targets."
// Rulings fetched the same day from
// https://api.scryfall.com/cards/8c21da73-0dad-4b2b-8cc6-61861bed9410/rulings:
// five WotC rulings dated 2016-01-22. The mana value uses the printed {X}{X}{R}
// cost (so X=4 has mana value 9), regardless of Surge; Surge is only an
// alternative cost, the qualifying prior spell may have resolved, been
// countered, or remain on the stack, Surge copies inherit its paid status, and
// casting for Surge doesn't change mana cost or mana value.
// Official Wizards Comprehensive Rules effective 2026-11-06,
// https://media.wizards.com/2026/downloads/MagicCompRules%2020261106.txt.
// Read 102.3-4, 107.3a, 115.1, 115.4, 118.9-9d, 202.3, 601.2b,c,f-h,
// 608.2b, 120.3a-d, 120.4a-d, 120.5 and 702.117a: announce one to two
// distinct legal targets, retain the chosen X on the spell, and deal that full
// amount separately to each legal target. Its alternative cost doesn't change
// the printed mana value.
func TestFallOfTheTitansSurgeEachTargetXDamage(t *testing.T) {
	def := registeredSurgeDamageCard(t, fallOfTheTitansTestOracleID, "Fall of the Titans")
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	setMainPhasePriority(g, game.Player1)
	castSurgeSetupSpell(t, engine, g, game.Player1)
	engine.resolveTopOfStack(g, &TurnLog{})

	target := addCombatPermanent(g, game.Player2, &game.CardDef{CardFace: game.CardFace{
		Name:      "Fall of the Titans target",
		Types:     []types.Card{types.Creature},
		Toughness: opt.Val(game.PT{Value: 5}),
	}})
	cardID := addCardToHand(g, game.Player1, def)
	g.Players[game.Player1].ManaPool.Add(mana.C, 2)
	g.Players[game.Player1].ManaPool.Add(mana.R, 1)
	startingLife := g.Players[game.Player2].Life
	cast := action.CastSpell(cardID, []game.Target{
		game.PlayerTarget(game.Player2),
		game.PermanentTarget(target.ObjectID),
	}, 2, nil)
	if !engine.applyAction(g, game.Player1, cast) {
		t.Fatal("Fall of the Titans was not castable for Surge {X}{R} with X=2")
	}
	spell, ok := g.Stack.Peek()
	if !ok || spell.AlternativeCost != 1 || spell.XValue != 2 ||
		spell.ManaSpentToCast != 3 || g.Players[game.Player1].ManaPool.Total() != 0 {
		t.Fatalf("Fall of the Titans did not pay {X}{R} Surge with X=2: spell=%+v, mana=%d",
			spell, g.Players[game.Player1].ManaPool.Total())
	}
	if value, valid := stackObjectManaValue(g, spell); !valid || value != 5 {
		t.Fatalf("Fall of the Titans mana value = %d (%t), want 5 from printed {X}{X}{R}",
			value, valid)
	}

	engine.resolveTopOfStack(g, &TurnLog{})
	damaged, present := permanentByObjectID(g, target.ObjectID)
	if !present || damaged.MarkedDamage != 2 || g.Players[game.Player2].Life != startingLife-2 {
		t.Fatalf("X damage results: creature=%+v (present %t), opponent life=%d; want 2 marked and %d",
			damaged, present, g.Players[game.Player2].Life, startingLife-2)
	}
}
