package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/cards"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// Source review: pinned printing cf76bee3-2b69-4c7b-9b7c-1a9f4bbcfde0,
// Oracle ID 1466687f-cfe1-4e6e-ad8e-dca5e8e4c06d, text: “Whenever you cast a
// blue spell or an Island you control enters, you gain 1 life.” Its
// https://api.scryfall.com/cards/cf76bee3-2b69-4c7b-9b7c-1a9f4bbcfde0/rulings
// response contains three WotC rulings dated 2014-07-18: any land with subtype
// Island counts (not only a land named Island); most nonbasic lands do not have
// basic land types; a cast card copy triggers, but a copied stack spell does not.
// The official Comprehensive Rules TXT, effective 2026-11-06,
// https://media.wizards.com/2026/downloads/MagicCompRules%2020261106.txt,
// §§105.1–2, 119.3, 205.3a–i, 305.1–6, 601.2/601.2i, 603.1–3, and 603.6/603.6a,
// requires matching the cast spell's color at casting, checking its Island
// subtype when it enters, and gaining one life for each triggered event.
func TestStaffOfTheMindMagusTriggersForBlueSpellsAndIslands(t *testing.T) {
	staffMagusTriggerScenario(t, "1466687f-cfe1-4e6e-ad8e-dca5e8e4c06d",
		"Staff of the Mind Magus", color.Blue, cost.Mana{cost.U}, mana.U, types.Island)
}

// Source review: pinned printing e64f31e8-0b49-4510-a4df-2e0bea793d3e,
// Oracle ID 6bdff38d-0a52-42e5-86ea-edd7029b44f6, text: “Whenever you cast a
// green spell or a Forest you control enters, you gain 1 life.” Its
// https://api.scryfall.com/cards/e64f31e8-0b49-4510-a4df-2e0bea793d3e/rulings
// response contains three WotC rulings dated 2014-07-18: any land with subtype
// Forest counts (not only a land named Forest); most nonbasic lands do not have
// basic land types; a cast card copy triggers, but a copied stack spell does not.
// The official Comprehensive Rules TXT, effective 2026-11-06,
// https://media.wizards.com/2026/downloads/MagicCompRules%2020261106.txt,
// §§105.1–2, 119.3, 205.3a–i, 305.1–6, 601.2/601.2i, 603.1–3, and 603.6/603.6a,
// requires matching the cast spell's color at casting, checking its Forest
// subtype when it enters, and gaining one life for each triggered event.
func TestStaffOfTheWildMagusTriggersForGreenSpellsAndForests(t *testing.T) {
	staffMagusTriggerScenario(t, "6bdff38d-0a52-42e5-86ea-edd7029b44f6",
		"Staff of the Wild Magus", color.Green, cost.Mana{cost.G}, mana.G, types.Forest)
}

// Source review: pinned printing e1f51ff4-9b5c-4e26-b811-4000a4c8f965,
// Oracle ID bd8bc927-e50a-4fab-9803-0602a841c7c2, text: “Whenever you cast a
// white spell or a Plains you control enters, you gain 1 life.” Its
// https://api.scryfall.com/cards/e1f51ff4-9b5c-4e26-b811-4000a4c8f965/rulings
// response contains three WotC rulings dated 2014-07-18: any land with subtype
// Plains counts (not only a land named Plains); most nonbasic lands do not have
// basic land types; a cast card copy triggers, but a copied stack spell does not.
// The official Comprehensive Rules TXT, effective 2026-11-06,
// https://media.wizards.com/2026/downloads/MagicCompRules%2020261106.txt,
// §§105.1–2, 119.3, 205.3a–i, 305.1–6, 601.2/601.2i, 603.1–3, and 603.6/603.6a,
// requires matching the cast spell's color at casting, checking its Plains
// subtype when it enters, and gaining one life for each triggered event.
func TestStaffOfTheSunMagusTriggersForWhiteSpellsAndPlains(t *testing.T) {
	staffMagusTriggerScenario(t, "bd8bc927-e50a-4fab-9803-0602a841c7c2",
		"Staff of the Sun Magus", color.White, cost.Mana{cost.W}, mana.W, types.Plains)
}

// Source review: pinned printing 84a341a4-8177-41cd-a3bc-eff5dee48c94,
// Oracle ID c3fda16e-fd56-434e-91ac-561ae0482f43, text: “Whenever you cast a
// black spell or a Swamp you control enters, you gain 1 life.” Its
// https://api.scryfall.com/cards/84a341a4-8177-41cd-a3bc-eff5dee48c94/rulings
// response contains three WotC rulings dated 2014-07-18: any land with subtype
// Swamp counts (not only a land named Swamp); most nonbasic lands do not have
// basic land types; a cast card copy triggers, but a copied stack spell does not.
// The official Comprehensive Rules TXT, effective 2026-11-06,
// https://media.wizards.com/2026/downloads/MagicCompRules%2020261106.txt,
// §§105.1–2, 119.3, 205.3a–i, 305.1–6, 601.2/601.2i, 603.1–3, and 603.6/603.6a,
// requires matching the cast spell's color at casting, checking its Swamp
// subtype when it enters, and gaining one life for each triggered event.
func TestStaffOfTheDeathMagusTriggersForBlackSpellsAndSwamps(t *testing.T) {
	staffMagusTriggerScenario(t, "c3fda16e-fd56-434e-91ac-561ae0482f43",
		"Staff of the Death Magus", color.Black, cost.Mana{cost.B}, mana.B, types.Swamp)
}

// Source review: pinned printing 0fd137db-296a-4c17-ba46-8b189d96c1f9,
// Oracle ID fc71710a-8294-4370-89da-7ab1c2b38d19, text: “Whenever you cast a
// red spell or a Mountain you control enters, you gain 1 life.” Its
// https://api.scryfall.com/cards/0fd137db-296a-4c17-ba46-8b189d96c1f9/rulings
// response contains three WotC rulings dated 2014-07-18: any land with subtype
// Mountain counts (not only a land named Mountain); most nonbasic lands do not
// have basic land types; a cast card copy triggers, but a copied stack spell does not.
// The official Comprehensive Rules TXT, effective 2026-11-06,
// https://media.wizards.com/2026/downloads/MagicCompRules%2020261106.txt,
// §§105.1–2, 119.3, 205.3a–i, 305.1–6, 601.2/601.2i, 603.1–3, and 603.6/603.6a,
// requires matching the cast spell's color at casting, checking its Mountain
// subtype when it enters, and gaining one life for each triggered event.
func TestStaffOfTheFlameMagusTriggersForRedSpellsAndMountains(t *testing.T) {
	staffMagusTriggerScenario(t, "fc71710a-8294-4370-89da-7ab1c2b38d19",
		"Staff of the Flame Magus", color.Red, cost.Mana{cost.R}, mana.R, types.Mountain)
}

func staffMagusTriggerScenario(
	t *testing.T,
	oracleID string,
	name string,
	spellColor color.Color,
	spellCost cost.Mana,
	spellMana mana.Color,
	landType types.Sub,
) {
	t.Helper()

	def, err := cards.NewDefaultRegistry().LookupOracleID(oracleID)
	if err != nil {
		t.Fatalf("exact %s registry lookup failed: %v", name, err)
	}
	if def.Name != name {
		t.Fatalf("exact %s lookup returned %q", name, def.Name)
	}

	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	staffID := addCardToHand(g, game.Player1, def)
	g.Players[game.Player1].ManaPool.Add(mana.C, 3)
	setMainPhasePriority(g, game.Player1)

	if !engine.applyAction(g, game.Player1, action.CastSpell(staffID, nil, 0, nil)) {
		t.Fatalf("casting %s for {3} failed", name)
	}
	if got := g.Players[game.Player1].ManaPool.Amount(mana.C); got != 0 {
		t.Fatalf("%s left %d generic mana, want 0 after paying {3}", name, got)
	}
	engine.resolveTopOfStack(g, &TurnLog{})

	lifeBefore := g.Players[game.Player1].Life
	spellID := addCardToHand(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:     name + " color trigger probe",
		ManaCost: opt.Val(spellCost),
		Types:    []types.Card{types.Instant},
		Colors:   []color.Color{spellColor},
	}})
	g.Players[game.Player1].ManaPool.Add(spellMana, 1)
	setMainPhasePriority(g, game.Player1)
	if !engine.applyAction(g, game.Player1, action.CastSpell(spellID, nil, 0, nil)) {
		t.Fatalf("casting a %s spell did not succeed", spellColor)
	}
	if !engine.putTriggeredAbilitiesOnStack(g) {
		t.Fatalf("%s did not trigger for its matching spell", name)
	}
	engine.resolveTopOfStack(g, &TurnLog{})
	if got := g.Players[game.Player1].Life; got != lifeBefore+1 {
		t.Fatalf("life after casting %s spell = %d, want %d", spellColor, got, lifeBefore+1)
	}
	engine.resolveTopOfStack(g, &TurnLog{})

	landID := addCardToHand(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:       string(landType),
		Supertypes: []types.Super{types.Basic},
		Types:      []types.Card{types.Land},
		Subtypes:   []types.Sub{landType},
	}})
	setMainPhasePriority(g, game.Player1)
	if !engine.applyAction(g, game.Player1, action.PlayLand(landID)) {
		t.Fatalf("playing a %s failed", landType)
	}
	if !engine.putTriggeredAbilitiesOnStack(g) {
		t.Fatalf("%s did not trigger when its %s entered", name, landType)
	}
	engine.resolveTopOfStack(g, &TurnLog{})
	if got := g.Players[game.Player1].Life; got != lifeBefore+2 {
		t.Fatalf("life after %s entered = %d, want %d", landType, got, lifeBefore+2)
	}
}
