package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/cards"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/id"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

func adventureBatchCard(t *testing.T, oracleID, name string) *game.CardDef {
	t.Helper()
	def, err := cards.NewDefaultRegistry().LookupOracleID(oracleID)
	if err != nil {
		t.Fatalf("default-registry lookup for %s (%s): %v", name, oracleID, err)
	}
	if def.Name != name {
		t.Fatalf("oracle %s resolved to %q, want %q", oracleID, def.Name, name)
	}
	return def
}

func adventureBatchPermanent(g *game.Game, cardID id.ID) *game.Permanent {
	for _, permanent := range g.Battlefield {
		if permanent.CardInstanceID == cardID {
			return permanent
		}
	}
	return nil
}

func adventureBatchColoredMana(g *game.Game, player game.PlayerID) int {
	pool := &g.Players[player].ManaPool
	return pool.Amount(mana.W) + pool.Amount(mana.U) + pool.Amount(mana.B) +
		pool.Amount(mana.R) + pool.Amount(mana.G)
}

// Source review:
// Rosethorn Acolyte // Seasonal Ritual, Oracle ID 639a8804-8513-4ba4-8683-ea1e2e5a9530,
// pinned corpus SHA-256 6da9ced0e19bd384771ab8f14051cc45a8dc90302082b875fa6e039d2399a1b4,
// printing 2a0d430f-da84-4752-940c-8457c525aac9 (2019-10-04):
// Rosethorn Acolyte ({2}{G}, 2/3 Creature — Elf Druid), "{T}: Add one mana of
// any color."; Seasonal Ritual ({G}, Sorcery — Adventure), "Add one mana of any
// color. (Then exile this card. You may cast the creature later from exile.)"
// Scryfall rulings retrieved 2026-11-13 from
// https://api.scryfall.com/cards/2a0d430f-da84-4752-940c-8457c525aac9/rulings
// (10 entries; response retained at
// .cardwork/source-review/adventure-20261113/2a0d430f-da84-4752-940c-8457c525aac9-rulings.json).
// Its Adventure rulings require using the alternative characteristics on the
// stack, exiling the card when the Adventure resolves, and granting that spell's
// controller permission to cast its creature face from exile.
// Wizards Comprehensive Rules effective 2026-11-06, official TXT at
// https://media.wizards.com/2026/downloads/MagicCompRules%2020261106.txt,
// verified local copy .cardwork/MagicCompRules-20261106.txt (SHA-256
// 524f5195a1e5f54bb13c735edf399ac717a4458e2713f398d1a3e6877186b829), linked
// from https://magic.wizards.com/en/rules. Read 106.3-4, 601.2f, 605.1a,
// 716.3a-d, and 716.4: spell and mana-ability effects put their produced mana
// in the pool; the tap ability is a mana ability; the Adventure face supplies
// its own casting characteristics, exiles on resolution, and grants that
// resolving spell's controller permission to play the normal creature face.
func TestRosethornAcolyteSeasonalRitualScenario(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	def := adventureBatchCard(t, "639a8804-8513-4ba4-8683-ea1e2e5a9530", "Rosethorn Acolyte")
	acolyteID := addCardToHand(g, game.Player1, def)
	ritualID := addCardToHand(g, game.Player1, def)
	g.Players[game.Player1].ManaPool.Add(mana.C, 2)
	g.Players[game.Player1].ManaPool.Add(mana.G, 1)
	giveMainPhasePriority(g, game.Player1)

	if !engine.applyAction(g, game.Player1, action.CastSpellFace(acolyteID, game.FaceFront, nil, 0, nil)) {
		t.Fatal("Rosethorn Acolyte did not cast for {2}{G}")
	}
	if got := g.Players[game.Player1].ManaPool.Total(); got != 0 {
		t.Fatalf("mana after casting Rosethorn Acolyte = %d, want 0", got)
	}
	engine.resolveTopOfStack(g, &TurnLog{})
	acolyte := adventureBatchPermanent(g, acolyteID)
	if acolyte == nil {
		t.Fatal("Rosethorn Acolyte did not enter the battlefield")
	}
	acolyte.SummoningSick = false
	manaAbilityChoices := [game.NumPlayers]PlayerAgent{
		game.Player1: &choiceOnlyAgent{choices: [][]int{{3}}},
	}
	if !engine.applyActionWithChoices(
		g, game.Player1, action.ActivateAbility(acolyte.ObjectID, def.ManaAbilityIndex(0), nil, 0),
		manaAbilityChoices, &TurnLog{},
	) {
		t.Fatal("Rosethorn Acolyte could not activate its any-color mana ability")
	}
	if !acolyte.Tapped {
		t.Fatal("Rosethorn Acolyte did not tap for its mana ability")
	}
	if got := g.Players[game.Player1].ManaPool.Total(); got != 1 ||
		adventureBatchColoredMana(g, game.Player1) != 1 || g.Players[game.Player1].ManaPool.Amount(mana.R) != 1 {
		t.Fatalf("Rosethorn Acolyte mana = %d total, %d colored, want one colored mana", got, adventureBatchColoredMana(g, game.Player1))
	}

	g.Players[game.Player1].ManaPool.Empty()
	g.Players[game.Player1].ManaPool.Add(mana.G, 1)
	giveMainPhasePriority(g, game.Player1)
	if !engine.applyAction(g, game.Player1, action.CastSpellFace(ritualID, game.FaceAlternate, nil, 0, nil)) {
		t.Fatal("Seasonal Ritual did not cast for {G}")
	}
	if got := g.Players[game.Player1].ManaPool.Total(); got != 0 {
		t.Fatalf("mana after casting Seasonal Ritual = %d, want 0", got)
	}
	ritualChoices := [game.NumPlayers]PlayerAgent{
		game.Player1: &choiceOnlyAgent{choices: [][]int{{1}}},
	}
	engine.resolveTopOfStackWithChoices(g, ritualChoices, &TurnLog{})
	if !g.Players[game.Player1].Exile.Contains(ritualID) {
		t.Fatal("resolved Seasonal Ritual was not exiled")
	}
	if got := g.Players[game.Player1].ManaPool.Total(); got != 1 ||
		adventureBatchColoredMana(g, game.Player1) != 1 || g.Players[game.Player1].ManaPool.Amount(mana.U) != 1 {
		t.Fatalf("Seasonal Ritual mana = %d total, %d colored, want one colored mana", got, adventureBatchColoredMana(g, game.Player1))
	}

	g.Players[game.Player1].ManaPool.Empty()
	g.Players[game.Player1].ManaPool.Add(mana.C, 2)
	g.Players[game.Player1].ManaPool.Add(mana.G, 1)
	giveMainPhasePriority(g, game.Player1)
	if !engine.applyAction(g, game.Player1, action.CastSpellFaceFromZone(ritualID, zone.Exile, game.FaceFront, nil, 0, nil)) {
		t.Fatal("Rosethorn Acolyte could not be cast from Seasonal Ritual's exile")
	}
	engine.resolveTopOfStack(g, &TurnLog{})
	if adventureBatchPermanent(g, ritualID) == nil {
		t.Fatal("Seasonal Ritual's creature face did not enter the battlefield")
	}
}

// Source review:
// Crystal Dragon // Rob the Hoard, Oracle ID 56b0cb64-b886-413e-bb20-392bf12d3926,
// pinned corpus SHA-256 6da9ced0e19bd384771ab8f14051cc45a8dc90302082b875fa6e039d2399a1b4,
// printing 384ec8b7-472e-4026-9632-349ab9df2b27 (2022-06-10):
// Crystal Dragon ({4}{W}{W}, 4/4 Creature — Dragon), "Flying, vigilance";
// Rob the Hoard ({1}{W}, Sorcery — Adventure), "Return target artifact,
// enchantment, or legendary card from your graveyard to your hand. (Then exile
// this card. You may cast the creature later from exile.)"
// Scryfall rulings retrieved 2026-11-13 from
// https://api.scryfall.com/cards/384ec8b7-472e-4026-9632-349ab9df2b27/rulings
// (11 entries; response retained at
// .cardwork/source-review/adventure-20261113/384ec8b7-472e-4026-9632-349ab9df2b27-rulings.json).
// The Adventure rulings require that the graveyard card have its normal
// characteristics outside the stack, and that the resolving Adventure be
// exiled for its controller to cast the permanent face later.
// Wizards Comprehensive Rules effective 2026-11-06, official TXT at
// https://media.wizards.com/2026/downloads/MagicCompRules%2020261106.txt,
// verified local copy .cardwork/MagicCompRules-20261106.txt (SHA-256
// 524f5195a1e5f54bb13c735edf399ac717a4458e2713f398d1a3e6877186b829), linked
// from https://magic.wizards.com/en/rules. Read 601.2c/f, 608.2b-c, 702.9b,
// 702.20b, and 716.3a-d/716.4: choose the appropriate target while casting,
// recheck its legality at resolution, return the card in the written order,
// use the alternate characteristics only for the Adventure spell, exile it on
// resolution, and retain flying/vigilance on the creature face.
func TestCrystalDragonRobTheHoardScenario(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	def := adventureBatchCard(t, "56b0cb64-b886-413e-bb20-392bf12d3926", "Crystal Dragon")
	dragonID := addCardToHand(g, game.Player1, def)
	relicID := addCardToGraveyard(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:  "Graveyard Relic",
		Types: []types.Card{types.Artifact},
	}})
	g.Players[game.Player1].ManaPool.Add(mana.C, 1)
	g.Players[game.Player1].ManaPool.Add(mana.W, 1)
	giveMainPhasePriority(g, game.Player1)

	if !engine.applyAction(g, game.Player1, action.CastSpellFace(
		dragonID, game.FaceAlternate, []game.Target{game.CardTarget(relicID)}, 0, nil,
	)) {
		t.Fatal("Rob the Hoard did not cast for {1}{W} targeting the artifact in its controller's graveyard")
	}
	if got := g.Players[game.Player1].ManaPool.Total(); got != 0 {
		t.Fatalf("mana after casting Rob the Hoard = %d, want 0", got)
	}
	engine.resolveTopOfStack(g, &TurnLog{})
	if !g.Players[game.Player1].Hand.Contains(relicID) {
		t.Fatal("Rob the Hoard did not return its artifact target to hand")
	}
	if !g.Players[game.Player1].Exile.Contains(dragonID) {
		t.Fatal("resolved Rob the Hoard was not exiled")
	}

	g.Players[game.Player1].ManaPool.Add(mana.C, 4)
	g.Players[game.Player1].ManaPool.Add(mana.W, 2)
	giveMainPhasePriority(g, game.Player1)
	if !engine.applyAction(g, game.Player1, action.CastSpellFaceFromZone(dragonID, zone.Exile, game.FaceFront, nil, 0, nil)) {
		t.Fatal("Crystal Dragon could not be cast from Rob the Hoard's exile")
	}
	if got := g.Players[game.Player1].ManaPool.Total(); got != 0 {
		t.Fatalf("mana after casting Crystal Dragon = %d, want 0", got)
	}
	engine.resolveTopOfStack(g, &TurnLog{})
	dragon := adventureBatchPermanent(g, dragonID)
	if dragon == nil {
		t.Fatal("Crystal Dragon did not enter the battlefield")
	}
	if !hasKeyword(g, dragon, game.Flying) || !hasKeyword(g, dragon, game.Vigilance) {
		t.Fatal("Crystal Dragon does not have its printed flying and vigilance abilities")
	}
}

// Source review:
// Pollen-Shield Hare // Hare Raising, Oracle ID 3ef02f2b-b611-40c3-a1d5-8b34989643cf,
// pinned corpus SHA-256 6da9ced0e19bd384771ab8f14051cc45a8dc90302082b875fa6e039d2399a1b4,
// printing 139f31f5-28f0-4e90-abdf-3f6ed0992dea (2023-09-08):
// Pollen-Shield Hare ({1}{W}, 2/2 Creature — Rabbit), "Creature tokens you
// control get +1/+1."; Hare Raising ({G}, Sorcery — Adventure), "Target creature
// you control gains vigilance and gets +X/+X until end of turn, where X is the
// number of creatures you control."
// Scryfall rulings retrieved 2026-11-13 from
// https://api.scryfall.com/cards/139f31f5-28f0-4e90-abdf-3f6ed0992dea/rulings
// (13 entries; response retained at
// .cardwork/source-review/adventure-20261113/139f31f5-28f0-4e90-abdf-3f6ed0992dea-rulings.json).
// Its 2023-09-01 ruling requires the spell to determine X only once as Hare
// Raising resolves; its Adventure rulings require the resolved Adventure to be
// exiled with permission granted to its resolving controller.
// Wizards Comprehensive Rules effective 2026-11-06, official TXT at
// https://media.wizards.com/2026/downloads/MagicCompRules%2020261106.txt,
// verified local copy .cardwork/MagicCompRules-20261106.txt (SHA-256
// 524f5195a1e5f54bb13c735edf399ac717a4458e2713f398d1a3e6877186b829), linked
// from https://magic.wizards.com/en/rules. Read 601.2c, 604.1-2, 608.2b-c,
// 611.2c, 613.4c, 702.20a-b, 716.3a-d/716.4, and 514.2: the spell needs an
// appropriate creature target, X is computed on resolution, the static ability
// continuously modifies controlled creature tokens, the target gets vigilance
// and the power/toughness bonus until cleanup, and a resolved Adventure is
// exiled with a later creature-face permission.
func TestPollenShieldHareRaisingScenario(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	def := adventureBatchCard(t, "3ef02f2b-b611-40c3-a1d5-8b34989643cf", "Pollen-Shield Hare")
	hareID := addCardToHand(g, game.Player1, def)
	g.Players[game.Player1].ManaPool.Add(mana.C, 1)
	g.Players[game.Player1].ManaPool.Add(mana.W, 1)
	giveMainPhasePriority(g, game.Player1)

	if !engine.applyAction(g, game.Player1, action.CastSpellFace(hareID, game.FaceFront, nil, 0, nil)) {
		t.Fatal("Pollen-Shield Hare did not cast for {1}{W}")
	}
	engine.resolveTopOfStack(g, &TurnLog{})
	if adventureBatchPermanent(g, hareID) == nil {
		t.Fatal("Pollen-Shield Hare did not enter the battlefield")
	}

	tokenDef := &game.CardDef{CardFace: game.CardFace{
		Name:      "Hare token",
		Types:     []types.Card{types.Creature},
		Power:     opt.Val(game.PT{Value: 1}),
		Toughness: opt.Val(game.PT{Value: 1}),
	}}
	token, ok := createTokenPermanent(g, game.Player1, tokenDef)
	if !ok {
		t.Fatal("could not create a controlled creature token for Hare Raising")
	}
	otherCreature := addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:      "Other creature",
		Types:     []types.Card{types.Creature},
		Power:     opt.Val(game.PT{Value: 1}),
		Toughness: opt.Val(game.PT{Value: 1}),
	}})
	opposingToken, ok := createTokenPermanent(g, game.Player2, tokenDef)
	if !ok {
		t.Fatal("could not create an opposing creature token")
	}
	tokenToughness, tokenToughnessOK := effectiveToughness(g, token)
	if power := effectivePower(g, token); power != 2 || !tokenToughnessOK || tokenToughness != 2 {
		t.Fatalf("controlled token is %d/%d, want 2/2 from Pollen-Shield Hare", power, tokenToughness)
	}
	if effectivePower(g, otherCreature) != 1 || effectivePower(g, opposingToken) != 1 {
		t.Fatal("Pollen-Shield Hare modified a nontoken creature or an opponent's token")
	}

	raisingID := addCardToHand(g, game.Player1, def)
	g.Players[game.Player1].ManaPool.Add(mana.G, 1)
	giveMainPhasePriority(g, game.Player1)
	if !engine.applyAction(g, game.Player1, action.CastSpellFace(
		raisingID, game.FaceAlternate, []game.Target{game.PermanentTarget(token.ObjectID)}, 0, nil,
	)) {
		t.Fatal("Hare Raising did not cast for {G} targeting a creature its controller controls")
	}
	if got := g.Players[game.Player1].ManaPool.Total(); got != 0 {
		t.Fatalf("mana after casting Hare Raising = %d, want 0", got)
	}
	engine.resolveTopOfStack(g, &TurnLog{})
	if !g.Players[game.Player1].Exile.Contains(raisingID) {
		t.Fatal("resolved Hare Raising was not exiled")
	}
	toughness, toughnessOK := effectiveToughness(g, token)
	if effectivePower(g, token) != 5 || !toughnessOK || toughness != 5 {
		t.Fatalf("Hare Raising target is %d/%d, want 5/5 (+1 anthem and +3 for three controlled creatures)", effectivePower(g, token), toughness)
	}
	if !hasKeyword(g, token, game.Vigilance) {
		t.Fatal("Hare Raising target did not gain vigilance")
	}
}
