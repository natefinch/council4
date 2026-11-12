package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/cards"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
)

func runForMirrodinEquipmentScenario(
	t *testing.T,
	oracleID string,
	manaColor mana.Color,
	castCost int,
	equipCost int,
	wantPower int,
	wantToughness int,
	wantKeywords ...game.Keyword,
) {
	t.Helper()
	g, _ := setupBestowMain(t)
	registry := cards.NewDefaultRegistry()
	engine := NewEngineWithPrograms(nil, registry.Programs())
	def, err := registry.LookupOracleID(oracleID)
	if err != nil {
		t.Fatal(err)
	}
	if issues := game.ValidateCardDef(def); len(issues) != 0 {
		t.Fatalf("production declaration: %+v", issues)
	}

	target := addCombatCreaturePermanentWithPower(g, game.Player1, 2)
	cardID := addCardToHand(g, game.Player1, def)
	g.Players[game.Player1].ManaPool.Add(mana.C, 20)
	g.Players[game.Player1].ManaPool.Add(manaColor, 8)
	if !engine.applyAction(g, game.Player1, action.CastSpell(cardID, nil, 0, nil)) {
		t.Fatalf("%s could not be cast", def.Name)
	}
	if got := g.Players[game.Player1].ManaPool.Spent(); got != castCost {
		t.Fatalf("casting spent %d mana, want %d", got, castCost)
	}
	engine.resolveTopOfStack(g, &TurnLog{})
	equipment := permanentByCardID(g, cardID)
	if equipment == nil {
		t.Fatal("Equipment spell did not enter the battlefield")
	}
	if equipment.AttachedTo.Exists {
		t.Fatal("Equipment entered the battlefield already attached")
	}
	if !engine.putTriggeredAbilitiesOnStack(g) {
		t.Fatal("For Mirrodin! did not trigger on entry")
	}
	engine.resolveTopOfStack(g, &TurnLog{})

	var rebel *game.Permanent
	for _, permanent := range g.Battlefield {
		if permanent.Token && permanent.TokenDef != nil && permanent.TokenDef.Name == "Rebel Token" {
			rebel = permanent
			break
		}
	}
	if rebel == nil {
		t.Fatal("For Mirrodin! did not create its Rebel token")
	}
	if !permanentHasType(g, rebel, types.Creature) || !permanentHasSubtype(g, rebel, types.Rebel) {
		t.Fatal("created token is not a Rebel creature")
	}
	if got := permanentEffectiveColors(g, rebel); len(got) != 1 || got[0] != color.Red {
		t.Fatalf("Rebel token colors = %v, want red", got)
	}
	if !equipment.AttachedTo.Exists || equipment.AttachedTo.Val != rebel.ObjectID {
		t.Fatal("For Mirrodin! did not attach the Equipment to its Rebel")
	}
	assertForMirrodinStatsAndKeywords(t, g, rebel, wantPower, wantToughness, wantKeywords)

	if !engine.applyAction(g, game.Player1, action.ActivateAbility(
		equipment.ObjectID, 0, []game.Target{game.PermanentTarget(target.ObjectID)}, 0,
	)) {
		t.Fatalf("%s Equip ability could not target the other creature", def.Name)
	}
	if got := g.Players[game.Player1].ManaPool.Spent(); got != castCost+equipCost {
		t.Fatalf("casting and equipping spent %d mana, want %d", got, castCost+equipCost)
	}
	engine.resolveTopOfStack(g, &TurnLog{})
	if !equipment.AttachedTo.Exists || equipment.AttachedTo.Val != target.ObjectID {
		t.Fatal("Equip did not move the Equipment to its targeted creature")
	}
	assertForMirrodinStatsAndKeywords(t, g, target, wantPower, wantToughness, wantKeywords)
	if got, ok := effectiveToughness(g, rebel); !ok || effectivePower(g, rebel) != 2 || got != 2 {
		t.Fatalf("former equipped Rebel is %d/%d, want its unmodified 2/2", effectivePower(g, rebel), got)
	}
	for _, keyword := range wantKeywords {
		if hasKeyword(g, rebel, keyword) {
			t.Fatalf("former equipped Rebel still has %v", keyword)
		}
	}
}

func assertForMirrodinStatsAndKeywords(t *testing.T, g *game.Game, permanent *game.Permanent, power, toughness int, keywords []game.Keyword) {
	t.Helper()
	if got := effectivePower(g, permanent); got != power {
		t.Fatalf("%v power = %d, want %d", permanent.ObjectID, got, power)
	}
	if got, ok := effectiveToughness(g, permanent); !ok || got != toughness {
		t.Fatalf("%v toughness = %d (available %t), want %d", permanent.ObjectID, got, ok, toughness)
	}
	for _, keyword := range keywords {
		if !hasKeyword(g, permanent, keyword) {
			t.Fatalf("%v does not have %v", permanent.ObjectID, keyword)
		}
	}
}

// Source review: OracleID 1fd04d9e-cf89-49ba-a932-5d278f65bf77, printing revision 3006ea5a-5391-41eb-b0c8-092741dca2eb; rulings https://api.scryfall.com/cards/3006ea5a-5391-41eb-b0c8-092741dca2eb/rulings retrieved 2026-11-12 (3 rulings, all 2023-02-04: Rebel enters 2/2 before attachment; Equipment remains if Rebel dies; multiple tokens attach it to one). Read Wizards CR effective 2026-09-25 at https://media.wizards.com/2026/downloads/MagicCompRules%2020260925.txt (SHA-256 8d860e451f20f38865b725b42d82feb714c725373dd8f3b32b8652b3eeb070ca), sections 111.4, 117.1, 301.5b-c, 307.1, 601.2f-h, 602.2, 603.2c/603.3, 611.3a/c, 613.4c/613.5, 702.6a, 702.163a: cast and pay printed costs, trigger on entry to create the named red 2/2 Rebel then attach, and use Equip's sorcery-only target procedure; the attached stat and vigilance effects update with attachment.
func TestMirranBardicheScenario(t *testing.T) {
	runForMirrodinEquipmentScenario(t, "1fd04d9e-cf89-49ba-a932-5d278f65bf77", mana.W, 5, 4, 4, 3, game.Vigilance)
}

// Source review: OracleID 2f6c21a1-d618-497c-bfde-c0f2e1014d99, printing revision 71c1a1d0-5616-42f8-a59c-42c1ccd48d26; rulings https://api.scryfall.com/cards/71c1a1d0-5616-42f8-a59c-42c1ccd48d26/rulings retrieved 2026-11-12 (3 rulings, all 2023-02-04: Rebel enters 2/2 before attachment; Equipment remains if Rebel dies; multiple tokens attach it to one). Read Wizards CR effective 2026-09-25 at https://media.wizards.com/2026/downloads/MagicCompRules%2020260925.txt (SHA-256 8d860e451f20f38865b725b42d82feb714c725373dd8f3b32b8652b3eeb070ca), sections 111.4, 117.1, 301.5b-c, 307.1, 601.2f-h, 602.2, 603.2c/603.3, 611.3a/c, 613.4c/613.5, 702.6a, 702.163a: cast and pay printed costs, trigger on entry to create the named red 2/2 Rebel then attach, and use Equip's sorcery-only target procedure; the attached stat and trample effects update with attachment.
func TestSylvokBattleChairScenario(t *testing.T) {
	runForMirrodinEquipmentScenario(t, "2f6c21a1-d618-497c-bfde-c0f2e1014d99", mana.G, 6, 7, 6, 6, game.Trample)
}

// Source review: OracleID 4c35f41a-55e4-49ca-b211-f57d7b1b90f6, printing revision f4057210-ef07-4496-94c8-95c3b807c23c; rulings https://api.scryfall.com/cards/f4057210-ef07-4496-94c8-95c3b807c23c/rulings retrieved 2026-11-12 (3 rulings, all 2023-02-04: Rebel enters 2/2 before attachment; Equipment remains if Rebel dies; multiple tokens attach it to one). Read Wizards CR effective 2026-09-25 at https://media.wizards.com/2026/downloads/MagicCompRules%2020260925.txt (SHA-256 8d860e451f20f38865b725b42d82feb714c725373dd8f3b32b8652b3eeb070ca), sections 111.4, 117.1, 301.5b-c, 307.1, 601.2f-h, 602.2, 603.2c/603.3, 611.3a/c, 613.4c/613.5, 702.6a, 702.163a: cast and pay printed costs, trigger on entry to create the named red 2/2 Rebel then attach, and use Equip's sorcery-only target procedure; the attached toughness bonus updates with attachment.
func TestGoldwardensHelmScenario(t *testing.T) {
	runForMirrodinEquipmentScenario(t, "4c35f41a-55e4-49ca-b211-f57d7b1b90f6", mana.W, 3, 2, 2, 3)
}

// Source review: OracleID 4cd4c6da-a84e-4cba-893b-799a7cb5c411, printing revision 5456b036-231e-4a64-b060-0709f5254664; rulings https://api.scryfall.com/cards/5456b036-231e-4a64-b060-0709f5254664/rulings retrieved 2026-11-12 (3 rulings, all 2023-02-04: Rebel enters 2/2 before attachment; Equipment remains if Rebel dies; multiple tokens attach it to one). Read Wizards CR effective 2026-09-25 at https://media.wizards.com/2026/downloads/MagicCompRules%2020260925.txt (SHA-256 8d860e451f20f38865b725b42d82feb714c725373dd8f3b32b8652b3eeb070ca), sections 111.4, 117.1, 301.5b-c, 307.1, 601.2f-h, 602.2, 603.2c/603.3, 611.3a/c, 613.4c/613.5, 702.6a, 702.163a: cast and pay printed costs, trigger on entry to create the named red 2/2 Rebel then attach, and use Equip's sorcery-only target procedure; attached stats, flying, and haste update with attachment.
func TestDragonwingGliderScenario(t *testing.T) {
	runForMirrodinEquipmentScenario(t, "4cd4c6da-a84e-4cba-893b-799a7cb5c411", mana.R, 5, 5, 4, 4, game.Flying, game.Haste)
}

// Source review: OracleID 6253f9d4-83ff-4a0b-b465-e9f9a565f664, printing revision f05078dd-b928-4b3f-9636-f299ebac180b; rulings https://api.scryfall.com/cards/f05078dd-b928-4b3f-9636-f299ebac180b/rulings retrieved 2026-11-12 (3 rulings, all 2023-02-04: Rebel enters 2/2 before attachment; Equipment remains if Rebel dies; multiple tokens attach it to one). Read Wizards CR effective 2026-09-25 at https://media.wizards.com/2026/downloads/MagicCompRules%2020260925.txt (SHA-256 8d860e451f20f38865b725b42d82feb714c725373dd8f3b32b8652b3eeb070ca), sections 111.4, 117.1, 301.5b-c, 307.1, 601.2f-h, 602.2, 603.2c/603.3, 611.3a/c, 613.4c/613.5, 702.6a, 702.163a: cast and pay printed costs, trigger on entry to create the named red 2/2 Rebel then attach, and use Equip's sorcery-only target procedure; the attached power bonus updates with attachment.
func TestVulshokSplitterScenario(t *testing.T) {
	runForMirrodinEquipmentScenario(t, "6253f9d4-83ff-4a0b-b465-e9f9a565f664", mana.R, 4, 3, 4, 2)
}

// Source review: OracleID 93c2b91d-7d4e-483c-a11f-4a525f881503, printing revision de1d02d1-91dc-47d6-bdbe-87602428abfb; rulings https://api.scryfall.com/cards/de1d02d1-91dc-47d6-bdbe-87602428abfb/rulings retrieved 2026-11-12 (3 rulings, all 2023-02-04: Rebel enters 2/2 before attachment; Equipment remains if Rebel dies; multiple tokens attach it to one). Read Wizards CR effective 2026-09-25 at https://media.wizards.com/2026/downloads/MagicCompRules%2020260925.txt (SHA-256 8d860e451f20f38865b725b42d82feb714c725373dd8f3b32b8652b3eeb070ca), sections 111.4, 117.1, 301.5b-c, 307.1, 601.2f-h, 602.2, 603.2c/603.3, 611.3a/c, 613.4c/613.5, 702.6a, 702.163a: cast and pay printed costs, trigger on entry to create the named red 2/2 Rebel then attach, and use Equip's sorcery-only target procedure; the attached stat bonus updates with attachment.
func TestBarbedBatterfistScenario(t *testing.T) {
	runForMirrodinEquipmentScenario(t, "93c2b91d-7d4e-483c-a11f-4a525f881503", mana.R, 2, 1, 3, 1)
}

// Source review: OracleID bd31c248-8b3f-4fa5-a639-19b0e37b15c3, printing revision 5e50c148-9f30-4131-8ec1-85e9ba8cafbb; rulings https://api.scryfall.com/cards/5e50c148-9f30-4131-8ec1-85e9ba8cafbb/rulings retrieved 2026-11-12 (3 rulings, all 2023-02-04: Rebel enters 2/2 before attachment; Equipment remains if Rebel dies; multiple tokens attach it to one). Read Wizards CR effective 2026-09-25 at https://media.wizards.com/2026/downloads/MagicCompRules%2020260925.txt (SHA-256 8d860e451f20f38865b725b42d82feb714c725373dd8f3b32b8652b3eeb070ca), sections 111.4, 117.1, 301.5b-c, 307.1, 601.2f-h, 602.2, 603.2c/603.3, 611.3a/c, 613.4c/613.5, 702.6a, 702.163a: cast and pay printed costs, trigger on entry to create the named red 2/2 Rebel then attach, and use Equip's sorcery-only target procedure; the +1/+1-per-controlled-creature modifier recalculates and follows the attachment.
func TestKembasBannerScenario(t *testing.T) {
	runForMirrodinEquipmentScenario(t, "bd31c248-8b3f-4fa5-a639-19b0e37b15c3", mana.W, 4, 3, 4, 4)
}
