package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/cards"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/mana"
)

// Source review: Odyssey #304, printing revision 1fd66433-b0e9-453d-9a27-ea0efb158dac,
// Oracle ID 364b3231-c0e7-45a8-90b9-a1cdb584dd7c; pinned Oracle text:
// "{2}, {T}, Sacrifice this artifact: Add {R}{G}. Draw a card. (Activate only as an instant.)"
// Rulings at https://api.scryfall.com/cards/1fd66433-b0e9-453d-9a27-ea0efb158dac/rulings
// retrieved 2026-11-21: one WotC ruling dated 2008-08-01 says this is a mana ability
// and says the drawn card is not seen until the spell/ability being paid for is finished.
// Pinned card record and raw ruling response are retained in
// .cardwork/source-review/odyssey-eggs/mossfire-egg/.
// Official CR: Wizards https://magic.wizards.com/en/rules; run pin effective 2026-09-25
// (https://media.wizards.com/2026/downloads/MagicCompRules%2020260925.pdf) and current
// edition effective 2026-11-06 (https://media.wizards.com/2026/downloads/MagicCompRules%2020261106.pdf).
// Both local PDFs are retained under .cardwork/source-review; read 113.3b, 118.1-3,
// 121.1, 601.2f-h, 602.1a-b, 602.2a-b, 602.5e, 605.1a, 605.3b, and 608.2c in both
// editions; the cited passages are identical. CR 121.1 defines drawing as moving the
// library's top card to hand, which excludes this ability from mana abilities under
// 605.1a despite the older ruling. It therefore uses the ordinary activated-ability
// stack and instant-speed timing, resolving Add {R}, Add {G}, then Draw in text order.
func TestMossfireEggPaysCostsThenAddsManaAndDraws(t *testing.T) {
	runOdysseyEggScenario(t, "Mossfire Egg", mossfireEggTestOracleID, mana.R, mana.G)
}

// Source review: Odyssey #311, printing revision 70e888f9-c583-478a-9ef0-2e89db8a8dbb,
// Oracle ID 80a49a1b-a202-4c14-b093-dc76eb0f42c7; pinned Oracle text:
// "{2}, {T}, Sacrifice this artifact: Add {G}{W}. Draw a card. (Activate only as an instant.)"
// Rulings at https://api.scryfall.com/cards/70e888f9-c583-478a-9ef0-2e89db8a8dbb/rulings
// retrieved 2026-11-21: one WotC ruling dated 2008-08-01 says this is a mana ability
// and says the drawn card is not seen until the spell/ability being paid for is finished.
// Pinned card record and raw ruling response are retained in
// .cardwork/source-review/odyssey-eggs/sungrass-egg/.
// Official CR: Wizards https://magic.wizards.com/en/rules; run pin effective 2026-09-25
// (https://media.wizards.com/2026/downloads/MagicCompRules%2020260925.pdf) and current
// edition effective 2026-11-06 (https://media.wizards.com/2026/downloads/MagicCompRules%2020261106.pdf).
// Both local PDFs are retained under .cardwork/source-review; read 113.3b, 118.1-3,
// 121.1, 601.2f-h, 602.1a-b, 602.2a-b, 602.5e, 605.1a, 605.3b, and 608.2c in both
// editions; the cited passages are identical. CR 121.1 defines drawing as moving the
// library's top card to hand, which excludes this ability from mana abilities under
// 605.1a despite the older ruling. It therefore uses the ordinary activated-ability
// stack and instant-speed timing, resolving Add {G}, Add {W}, then Draw in text order.
func TestSungrassEggPaysCostsThenAddsManaAndDraws(t *testing.T) {
	runOdysseyEggScenario(t, "Sungrass Egg", sungrassEggTestOracleID, mana.G, mana.W)
}

// Source review: Odyssey #308, printing revision 044bbf15-e71d-4f47-bac3-bba8021118bd,
// Oracle ID 9b58e7fa-4259-4d6b-8b5d-33fb37fe489f; pinned Oracle text:
// "{2}, {T}, Sacrifice this artifact: Add {B}{R}. Draw a card. (Activate only as an instant.)"
// Rulings at https://api.scryfall.com/cards/044bbf15-e71d-4f47-bac3-bba8021118bd/rulings
// retrieved 2026-11-21: one WotC ruling dated 2008-08-01 says this is a mana ability
// and says the drawn card is not seen until the spell/ability being paid for is finished.
// Pinned card record and raw ruling response are retained in
// .cardwork/source-review/odyssey-eggs/shadowblood-egg/.
// Official CR: Wizards https://magic.wizards.com/en/rules; run pin effective 2026-09-25
// (https://media.wizards.com/2026/downloads/MagicCompRules%2020260925.pdf) and current
// edition effective 2026-11-06 (https://media.wizards.com/2026/downloads/MagicCompRules%2020261106.pdf).
// Both local PDFs are retained under .cardwork/source-review; read 113.3b, 118.1-3,
// 121.1, 601.2f-h, 602.1a-b, 602.2a-b, 602.5e, 605.1a, 605.3b, and 608.2c in both
// editions; the cited passages are identical. CR 121.1 defines drawing as moving the
// library's top card to hand, which excludes this ability from mana abilities under
// 605.1a despite the older ruling. It therefore uses the ordinary activated-ability
// stack and instant-speed timing, resolving Add {B}, Add {R}, then Draw in text order.
func TestShadowbloodEggPaysCostsThenAddsManaAndDraws(t *testing.T) {
	runOdysseyEggScenario(t, "Shadowblood Egg", shadowbloodEggTestOracleID, mana.B, mana.R)
}

// Source review: Odyssey #299, printing revision fb988572-1e1d-4b1d-8dc7-78bda966554e,
// Oracle ID c4893a24-3cbc-4011-bef4-50e0c4dce16e; pinned Oracle text:
// "{2}, {T}, Sacrifice this artifact: Add {U}{B}. Draw a card. (Activate only as an instant.)"
// Rulings at https://api.scryfall.com/cards/fb988572-1e1d-4b1d-8dc7-78bda966554e/rulings
// retrieved 2026-11-21: one WotC ruling dated 2008-08-01 says this is a mana ability
// and says the drawn card is not seen until the spell/ability being paid for is finished.
// Pinned card record and raw ruling response are retained in
// .cardwork/source-review/odyssey-eggs/darkwater-egg/.
// Official CR: Wizards https://magic.wizards.com/en/rules; run pin effective 2026-09-25
// (https://media.wizards.com/2026/downloads/MagicCompRules%2020260925.pdf) and current
// edition effective 2026-11-06 (https://media.wizards.com/2026/downloads/MagicCompRules%2020261106.pdf).
// Both local PDFs are retained under .cardwork/source-review; read 113.3b, 118.1-3,
// 121.1, 601.2f-h, 602.1a-b, 602.2a-b, 602.5e, 605.1a, 605.3b, and 608.2c in both
// editions; the cited passages are identical. CR 121.1 defines drawing as moving the
// library's top card to hand, which excludes this ability from mana abilities under
// 605.1a despite the older ruling. It therefore uses the ordinary activated-ability
// stack and instant-speed timing, resolving Add {U}, Add {B}, then Draw in text order.
func TestDarkwaterEggPaysCostsThenAddsManaAndDraws(t *testing.T) {
	runOdysseyEggScenario(t, "Darkwater Egg", darkwaterEggTestOracleID, mana.U, mana.B)
}

const (
	mossfireEggTestOracleID    = "364b3231-c0e7-45a8-90b9-a1cdb584dd7c"
	sungrassEggTestOracleID    = "80a49a1b-a202-4c14-b093-dc76eb0f42c7"
	shadowbloodEggTestOracleID = "9b58e7fa-4259-4d6b-8b5d-33fb37fe489f"
	darkwaterEggTestOracleID   = "c4893a24-3cbc-4011-bef4-50e0c4dce16e"
)

func runOdysseyEggScenario(t *testing.T, name, oracleID string, first, second mana.Color) {
	t.Helper()
	def, err := cards.NewDefaultRegistry().LookupOracleID(oracleID)
	if err != nil {
		t.Fatalf("exact %s registry lookup failed: %v", name, err)
	}
	if def.Name != name {
		t.Fatalf("exact %s lookup returned %q", name, def.Name)
	}

	g := mainPhaseGame(game.Player1)
	setMainPhasePriority(g, game.Player1)
	engine := NewEngine(nil)
	source := addCombatPermanent(g, game.Player1, def)
	drawnCard := addCardToLibrary(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: name + " draw"}})
	g.Players[game.Player1].ManaPool.Add(mana.C, 2)

	activation := action.ActivateAbility(source.ObjectID, def.ActivatedAbilityIndex(0), nil, 0)
	if !actionsContain(engine.legalActions(g, game.Player1), activation) ||
		!engine.applyAction(g, game.Player1, activation) {
		t.Fatalf("%s's {2}, tap, and sacrifice activation was not legally completed", name)
	}
	if !source.Tapped || !g.Players[game.Player1].Graveyard.Contains(source.CardInstanceID) ||
		g.Players[game.Player1].ManaPool.Total() != 0 || g.Players[game.Player1].Hand.Size() != 0 ||
		!g.Players[game.Player1].Library.Contains(drawnCard) || g.Stack.Size() != 1 {
		t.Fatalf("%s did not pay {2}, tap and sacrifice itself, and put the ability on the stack before its effects", name)
	}

	engine.resolveTopOfStack(g, &TurnLog{})
	if g.Players[game.Player1].ManaPool.Amount(first) != 1 ||
		g.Players[game.Player1].ManaPool.Amount(second) != 1 ||
		!g.Players[game.Player1].Hand.Contains(drawnCard) ||
		g.Players[game.Player1].Library.Size() != 0 || g.Stack.Size() != 0 {
		t.Fatalf("resolving %s did not add its two colors, draw the top card, and remove the ability from the stack", name)
	}
}
