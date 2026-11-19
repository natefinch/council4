package rules

import (
	"fmt"
	"testing"

	"github.com/natefinch/council4/mtg/cards"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
)

func castPhasingCreature(t *testing.T, oracleID, name string, genericCost int) (*game.Game, *Engine, *game.Permanent) {
	t.Helper()

	def, err := cards.NewDefaultRegistry().LookupOracleID(oracleID)
	if err != nil {
		t.Fatalf("exact %s registry lookup failed: %v", name, err)
	}
	if def.Name != name {
		t.Fatalf("exact lookup returned %q, want %q", def.Name, name)
	}

	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	cardID := addCardToHand(g, game.Player1, def)
	g.Players[game.Player1].ManaPool.Add(mana.C, genericCost)
	g.Players[game.Player1].ManaPool.Add(mana.U, 1)
	setMainPhasePriority(g, game.Player1)

	if !engine.applyAction(g, game.Player1, action.CastSpell(cardID, nil, 0, nil)) {
		t.Fatalf("could not cast %s for {%d}{U}", name, genericCost)
	}
	if !g.Players[game.Player1].ManaPool.IsEmpty() {
		t.Fatalf("casting %s did not spend its full {%d}{U} cost", name, genericCost)
	}
	engine.resolveTopOfStack(g, &TurnLog{})

	permanent := permanentByCardID(g, cardID)
	if permanent == nil {
		t.Fatalf("%s did not resolve onto the battlefield", name)
	}
	return g, engine, permanent
}

// Source review: pinned Oracle printing revision d0f34166-8946-4bcc-84af-3540c42ac7f7,
// Oracle ID 19edc98c-edc2-47bf-b2da-00f536d7c158: Merfolk Raiders costs {1}{U},
// is a 2/3 Merfolk Soldier, and has islandwalk and phasing. Its Scryfall rulings
// URI, https://api.scryfall.com/cards/d0f34166-8946-4bcc-84af-3540c42ac7f7/rulings,
// returned no rulings on 2026-11-19. Official Comprehensive Rules effective
// 2026-09-25 (the corpus pin), read from
// https://media.wizards.com/2026/downloads/MagicCompRules%2020260925.docx
// (SHA-256 201037368d45407f271733a0575b756b3a868ce5e8024cf24aa1f30d48ecb960):
// 302.1-2 and 601.2 govern casting/resolution; 113.3d, 702.14a-c, and 509.1a-b
// make its landwalk a live evasion effect; 502.1, 502.3, and 702.26a-d require
// its phasing before untap without changing zones. The scenario pays {1}{U},
// resolves the exact registered card, and checks that a blocker able to block
// before the defender gains an Island can't block it afterward.
func TestMerfolkRaidersIslandwalkPreventsBlockingAfterCast(t *testing.T) {
	g, _, merfolk := castPhasingCreature(t, "19edc98c-edc2-47bf-b2da-00f536d7c158", "Merfolk Raiders", 1)
	if !hasKeyword(g, merfolk, game.Landwalk) || !hasKeyword(g, merfolk, game.Phasing) {
		t.Fatal("Merfolk Raiders did not have its printed islandwalk and phasing abilities")
	}

	blocker := addCombatCreaturePermanentWithPower(g, game.Player2, 2)
	if !canBlockAttacker(g, blocker, merfolk) {
		t.Fatal("control blocker could not block Merfolk Raiders before player 2 controlled an Island")
	}
	addLandPermanent(g, game.Player2, "Defending Island", types.Island)
	if canBlockAttacker(g, blocker, merfolk) {
		t.Fatal("Merfolk Raiders could be blocked while the defending player controlled an Island")
	}
}

// Source review: pinned Oracle printing revision c24823df-5651-4578-a0c8-f9f52f66abe4,
// Oracle ID 3605c5ac-3b35-41a5-9266-286fb1d11716: Teferi's Drake costs {2}{U},
// is a 3/2 Drake, and has flying and phasing. Its Scryfall rulings URI,
// https://api.scryfall.com/cards/c24823df-5651-4578-a0c8-f9f52f66abe4/rulings,
// returned no rulings on 2026-11-19. Official Comprehensive Rules effective
// 2026-09-25 (the corpus pin), read from
// https://media.wizards.com/2026/downloads/MagicCompRules%2020260925.docx
// (SHA-256 201037368d45407f271733a0575b756b3a868ce5e8024cf24aa1f30d48ecb960):
// 302.1-2 and 601.2 govern casting/resolution; 113.3d, 702.9a-b, and 509.1a-b
// make flying an evasion ability; 502.1, 502.3, and 702.26a-d govern its
// before-untap phasing. This scenario pays {2}{U}, resolves the exact registered
// card, and observes both of its printed abilities on the resulting permanent.
func TestTeferisDrakeCastsWithFlyingAndPhasing(t *testing.T) {
	g, _, drake := castPhasingCreature(t, "3605c5ac-3b35-41a5-9266-286fb1d11716", "Teferi's Drake", 2)
	if !hasKeyword(g, drake, game.Flying) || !hasKeyword(g, drake, game.Phasing) {
		t.Fatal("Teferi's Drake did not have its printed flying and phasing abilities")
	}
}

// Source review: pinned Oracle printing revision e04bba8a-48ee-4981-adc2-4f82c0f2c1bd,
// Oracle ID d999f2a3-5cb4-4f12-a79a-c5db0a300620: Tolarian Drake costs {2}{U},
// is a 2/4 Drake, and has flying and phasing. Its Scryfall rulings URI,
// https://api.scryfall.com/cards/e04bba8a-48ee-4981-adc2-4f82c0f2c1bd/rulings,
// returned no rulings on 2026-11-19. Official Comprehensive Rules effective
// 2026-09-25 (the corpus pin), read from
// https://media.wizards.com/2026/downloads/MagicCompRules%2020260925.docx
// (SHA-256 201037368d45407f271733a0575b756b3a868ce5e8024cf24aa1f30d48ecb960):
// 302.1-2 and 601.2 govern casting/resolution; 113.3d, 702.9a-b, and 509.1a-b
// make flying an evasion ability; 502.1, 502.3, and 702.26a-d govern its
// before-untap phasing. The scenario pays {2}{U}, resolves Tolarian Drake, and
// compares an otherwise legal ground block before checking flying's restriction.
func TestTolarianDrakeFlyingPreventsGroundBlockAfterCast(t *testing.T) {
	g, _, drake := castPhasingCreature(t, "d999f2a3-5cb4-4f12-a79a-c5db0a300620", "Tolarian Drake", 2)
	if !hasKeyword(g, drake, game.Flying) || !hasKeyword(g, drake, game.Phasing) {
		t.Fatal("Tolarian Drake did not have its printed flying and phasing abilities")
	}

	blocker := addCombatCreaturePermanentWithPower(g, game.Player2, 2)
	groundCreature := addCombatCreaturePermanentWithPower(g, game.Player1, 2)
	if !canBlockAttacker(g, blocker, groundCreature) {
		t.Fatal("setup ground blocker could not block a nonflying creature")
	}
	if canBlockAttacker(g, blocker, drake) {
		t.Fatal("ground creature could block Tolarian Drake with flying")
	}
}

// Source review: pinned Oracle printing revision 450b79d9-5ab8-4699-8052-a278c316a5c3,
// Oracle ID e11926bf-ca14-4e1e-aca5-9ffedb91868c: Sandbar Crocodile costs {4}{U},
// is a 6/5 Crocodile, and has phasing. Its Scryfall rulings URI,
// https://api.scryfall.com/cards/450b79d9-5ab8-4699-8052-a278c316a5c3/rulings,
// returned no rulings on 2026-11-19. Official Comprehensive Rules effective
// 2026-09-25 (the corpus pin), read from
// https://media.wizards.com/2026/downloads/MagicCompRules%2020260925.docx
// (SHA-256 201037368d45407f271733a0575b756b3a868ce5e8024cf24aa1f30d48ecb960):
// 302.1-2 and 601.2 govern casting/resolution; 113.3d, 502.1, 502.3, and
// 702.26a-d require the permanent to phase out before its controller untaps,
// become nonexistent without changing zones, and retain its object identity.
// The scenario pays {4}{U}, resolves Sandbar Crocodile, skips the opponent's
// phasing event, then observes it phased out but still tapped on its next untap.
func TestSandbarCrocodilePhasesOutBeforeItsControllerUntaps(t *testing.T) {
	g, engine, crocodile := castPhasingCreature(t, "e11926bf-ca14-4e1e-aca5-9ffedb91868c", "Sandbar Crocodile", 4)
	if !hasKeyword(g, crocodile, game.Phasing) {
		t.Fatal("Sandbar Crocodile did not have phasing after resolving")
	}
	objectID := crocodile.ObjectID
	setPermanentTapped(g, crocodile, true)

	for player := game.Player1; player < game.NumPlayers; player++ {
		addCardToLibrary(g, player, &game.CardDef{CardFace: game.CardFace{
			Name: fmt.Sprintf("Sandbar library card for player %d", player),
		}})
		addCardToLibrary(g, player, &game.CardDef{CardFace: game.CardFace{
			Name: fmt.Sprintf("Sandbar next library card for player %d", player),
		}})
	}

	g.Turn.ActivePlayer = game.Player2
	g.Turn.PriorityPlayer = game.Player2
	engine.runBeginningPhase(g, [game.NumPlayers]PlayerAgent{}, &TurnLog{})
	if crocodile.PhasedOut || !crocodile.Tapped {
		t.Fatal("Sandbar Crocodile phased out or untapped during its opponent's untap")
	}

	g.Turn.ActivePlayer = game.Player1
	g.Turn.PriorityPlayer = game.Player1
	engine.runBeginningPhase(g, [game.NumPlayers]PlayerAgent{}, &TurnLog{})
	returned := permanentByCardID(g, crocodile.CardInstanceID)
	if returned == nil || returned.ObjectID != objectID || !returned.PhasedOut || !returned.Tapped {
		t.Fatal("Sandbar Crocodile did not phase out before its controller untapped as the same tapped permanent")
	}
}
