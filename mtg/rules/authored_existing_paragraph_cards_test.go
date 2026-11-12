package rules

import (
	"slices"
	"testing"

	"github.com/natefinch/council4/mtg/cards"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// Source review: Touch of the Void (Oracle ID 6b530534-5c02-4874-9256-501102ef8a5f,
// printing 006ead4a-dc57-4856-8e13-235ba55483e6) is Devoid and deals 3 damage
// to any target, exiling a creature dealt that damage if it would die that turn.
// Scryfall rulings https://api.scryfall.com/cards/006ead4a-dc57-4856-8e13-235ba55483e6/rulings
// returned six WotC rulings dated 2015-08-25: Devoid is colorless in every zone,
// and the damaged creature is exiled if it would die for any reason that turn.
// Official CR effective 2026-09-25,
// https://media.wizards.com/2026/downloads/MagicCompRules%2020260925.txt
// (SHA-256 8d860e451f20f38865b725b42d82feb714c725373dd8f3b32b8652b3eeb070ca):
// read 601.2, 115.1a, 120.3-8, 608.2b-c, 614.1/4/6, 702.16a-b, 702.114a,
// and 704.5g; these require a colorless spell that can target protection from
// red, and only a creature it actually damaged has its later death redirected
// to exile for the turn.
//
// Source review: Rebel Salvo (Oracle ID a35b7fd3-fdef-4414-b720-978f7ce70bff,
// printing cd37d1b1-70ce-466e-890d-36be82433035) has Affinity for Equipment,
// deals 5 damage to a creature or planeswalker, and removes indestructible from
// that permanent until end of turn. Its Scryfall rulings endpoint
// https://api.scryfall.com/cards/cd37d1b1-70ce-466e-890d-36be82433035/rulings
// returned no rulings. The same official CR edition above: read 115.1a, 301.5,
// 601.2f, 608.2b-c, 613.1f, 702.12a-b, and 702.41a; affinity counts the
// caster's Equipment during casting, then damage and the ability removal resolve.
//
// Source review: Konstrari Improviser // Soul Tether (Oracle ID
// bad2e3b2-ad47-4e17-b35e-0fcc13e82937, printing
// 42e28bd2-486b-45d4-8840-6e33c19c2d57) enters prepared; its copy spell creates
// a Heartwood token. Scryfall rulings
// https://api.scryfall.com/cards/42e28bd2-486b-45d4-8840-6e33c19c2d57/rulings
// returned 16 WotC rulings dated 2026-08-21: the prepared copy is in exile,
// casting it unprepares the creature, and Heartwood is a red-green artifact
// with the Heartwood subtype and {T}: add {R} or {G}. The same official CR
// edition above: read 111.10x, 601.2, 608.2b, and 722.1-3c; these require the
// predefined token, and the prepared copy's exile lifetime and cast permission.
//
// Source review: Militant Inquisitor (Oracle ID d1e326e8-6e9e-4dbd-8c2f-0bef90db2bb7,
// printing 68d99f1f-80d6-4b5e-aeae-83fb8ae62cdc) gets +1/+0 for each Equipment
// its controller controls. Its Scryfall rulings URI
// https://api.scryfall.com/cards/68d99f1f-80d6-4b5e-aeae-83fb8ae62cdc/rulings
// returned two WotC rulings dated 2016-04-08: all Equipment count whether
// attached or not, and this bonus stacks with bonuses from attached Equipment.
// The same official CR edition above: read 301.5 and 613.1f/4c; the count is
// of controlled Equipment, and the continuous power increase is applied to
// this creature.
//
// Source review: Manifest Dread (Oracle ID df5a66d1-61f6-44d9-b464-3ee739e78dce,
// printing a649265b-6c32-49e7-b6cb-6086c40d26e8) manifests one of the top two
// cards and puts the other into its controller's graveyard. Scryfall rulings
// https://api.scryfall.com/cards/a649265b-6c32-49e7-b6cb-6086c40d26e8/rulings
// returned 16 WotC rulings dated 2024-09-20: the choice produces a 2/2 face-down
// creature, and a manifested creature card may be turned face up for its mana
// cost as a special action. The same official CR edition above: read 401.2,
// 701.40a-b, 701.62a-b, 708.2-9, and 116.2b; these require private access to
// the looked-at library cards, the two-card choice, face-down manifestation,
// and the subsequent face-up action.
//
// Source review: Devil's Play (Oracle ID ee9f0b29-8a54-4cb8-8e2c-7bd67c2184ba,
// printing d46473a7-69f5-492b-845d-c8862a1fec85) deals X damage to any target
// and has Flashback {X}{R}{R}{R}. Scryfall rulings
// https://api.scryfall.com/cards/d46473a7-69f5-492b-845d-c8862a1fec85/rulings
// returned six WotC rulings dated 2021-03-19: flashback permits casting from
// the graveyard for its alternative cost, preserves normal sorcery timing, and
// exiles the card when it leaves the stack. The same official CR edition above:
// read 107.3a, 601.2b, 601.2f-h, 608.2b-c, and 702.34a; X determines both the
// paid cost and damage, and a flashback spell is exiled after resolution.
//
// Source review: Oxidda Finisher (Oracle ID f308a1de-94a6-4f9f-8f46-6f2018959b40,
// printing 7177431e-5cc4-4ecd-8848-f51903289072) has Affinity for Equipment
// and trample. Its Scryfall rulings endpoint
// https://api.scryfall.com/cards/7177431e-5cc4-4ecd-8848-f51903289072/rulings
// returned no rulings. The same official CR edition above: read 301.5, 510.1,
// 601.2, 702.19a-b, and 702.41a; the spell's cost is reduced for controlled
// Equipment and excess combat damage from this attacking creature can be
// assigned to the defending player.

const (
	touchOfTheVoidTestOracleID     = "6b530534-5c02-4874-9256-501102ef8a5f"
	rebelSalvoTestOracleID         = "a35b7fd3-fdef-4414-b720-978f7ce70bff"
	konstrariTestOracleID          = "bad2e3b2-ad47-4e17-b35e-0fcc13e82937"
	militantInquisitorTestOracleID = "d1e326e8-6e9e-4dbd-8c2f-0bef90db2bb7"
	manifestDreadTestOracleID      = "df5a66d1-61f6-44d9-b464-3ee739e78dce"
	devilsPlayTestOracleID         = "ee9f0b29-8a54-4cb8-8e2c-7bd67c2184ba"
	oxiddaFinisherTestOracleID     = "f308a1de-94a6-4f9f-8f46-6f2018959b40"
)

func authoredParagraphCard(t *testing.T, oracleID, name string) *game.CardDef {
	t.Helper()
	def, err := cards.NewDefaultRegistry().LookupOracleID(oracleID)
	if err != nil {
		t.Fatalf("exact %s registry lookup failed: %v", name, err)
	}
	if def.Name != name {
		t.Fatalf("exact %s registry lookup returned %q", name, def.Name)
	}
	return def
}

func paragraphTestEquipment(name string) *game.CardDef {
	return &game.CardDef{CardFace: game.CardFace{
		Name:     name,
		Types:    []types.Card{types.Artifact},
		Subtypes: []types.Sub{types.Equipment},
	}}
}

func TestTouchOfTheVoidScenario(t *testing.T) {
	def := authoredParagraphCard(t, touchOfTheVoidTestOracleID, "Touch of the Void")
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	protected := addProtectionFromColorPermanent(g, game.Player2, color.Red)
	spell := addCardToHand(g, game.Player1, def)
	g.Players[game.Player1].ManaPool.Add(mana.C, 2)
	g.Players[game.Player1].ManaPool.Add(mana.R, 1)
	setMainPhasePriority(g, game.Player1)

	if !engine.applyAction(g, game.Player1, action.CastSpell(
		spell, []game.Target{game.PermanentTarget(protected.ObjectID)}, 0, nil,
	)) {
		t.Fatal("Touch of the Void did not cast targeting a creature with protection from red")
	}
	if g.Players[game.Player1].ManaPool.Total() != 0 {
		t.Fatal("Touch of the Void did not pay {2}{R}")
	}
	engine.resolveTopOfStack(g, &TurnLog{})
	engine.applyStateBasedActions(g)

	if !g.Players[game.Player2].Exile.Contains(protected.CardInstanceID) {
		_, onBattlefield := permanentByObjectID(g, protected.ObjectID)
		t.Fatalf("Touch of the Void target: exile=%v graveyard=%v battlefield=%v markedDamage=%d replacements=%d",
			g.Players[game.Player2].Exile.Contains(protected.CardInstanceID),
			g.Players[game.Player2].Graveyard.Contains(protected.CardInstanceID),
			onBattlefield, protected.MarkedDamage, len(g.ReplacementEffects))
	}
	if g.Players[game.Player2].Graveyard.Contains(protected.CardInstanceID) {
		t.Fatal("Touch of the Void put its lethal-damage target into the graveyard instead of exile")
	}
}

func TestTouchOfTheVoidOnlyExilesCreaturesItDamaged(t *testing.T) {
	def := authoredParagraphCard(t, touchOfTheVoidTestOracleID, "Touch of the Void")
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	prevented := addReplacementPermanent(t, g, game.Player2, &game.CardDef{CardFace: game.CardFace{
		Name:      "Damage-preventing creature",
		Types:     []types.Card{types.Creature},
		Power:     opt.Val(game.PT{Value: 2}),
		Toughness: opt.Val(game.PT{Value: 2}),
		ReplacementAbilities: []game.ReplacementAbility{
			game.DamagePreventionToPlusOneCountersReplacement(
				"Prevent all damage to this creature and put that many +1/+1 counters on it.",
				false, opt.V[game.Condition]{},
			),
		},
	}})
	spell := addCardToHand(g, game.Player1, def)
	g.Players[game.Player1].ManaPool.Add(mana.C, 2)
	g.Players[game.Player1].ManaPool.Add(mana.R, 1)
	setMainPhasePriority(g, game.Player1)

	if !engine.applyAction(g, game.Player1, action.CastSpell(
		spell, []game.Target{game.PermanentTarget(prevented.ObjectID)}, 0, nil,
	)) {
		t.Fatal("Touch of the Void did not cast targeting a damage-preventing creature")
	}
	engine.resolveTopOfStack(g, &TurnLog{})

	if prevented.MarkedDamage != 0 {
		t.Fatalf("Touch of the Void dealt %d damage through complete prevention", prevented.MarkedDamage)
	}
	for _, replacement := range g.ReplacementEffects {
		if replacement.SourceCardID == spell && replacement.AffectedObjectID == prevented.ObjectID &&
			replacement.ReplaceToZone == zone.Exile {
			t.Fatal("Touch of the Void created its exile replacement without dealing damage")
		}
	}
}

func TestRebelSalvoScenario(t *testing.T) {
	def := authoredParagraphCard(t, rebelSalvoTestOracleID, "Rebel Salvo")
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	addCombatPermanent(g, game.Player1, paragraphTestEquipment("Rebel Salvo Equipment"))
	addCombatPermanent(g, game.Player2, paragraphTestEquipment("Opponent Equipment"))
	target := addCombatPermanent(g, game.Player2, &game.CardDef{CardFace: game.CardFace{
		Name:      "Indestructible Salvo target",
		Types:     []types.Card{types.Creature},
		Power:     opt.Val(game.PT{Value: 5}),
		Toughness: opt.Val(game.PT{Value: 5}),
		StaticAbilities: []game.StaticAbility{
			game.IndestructibleStaticBody,
		},
	}})
	spell := addCardToHand(g, game.Player1, def)
	g.Players[game.Player1].ManaPool.Add(mana.C, 1)
	g.Players[game.Player1].ManaPool.Add(mana.R, 1)
	setMainPhasePriority(g, game.Player1)

	if !engine.applyAction(g, game.Player1, action.CastSpell(
		spell, []game.Target{game.PermanentTarget(target.ObjectID)}, 0, nil,
	)) {
		t.Fatal("Rebel Salvo did not cast for {1}{R} with one Equipment controlled")
	}
	if g.Players[game.Player1].ManaPool.Total() != 0 {
		t.Fatal("Rebel Salvo did not reduce its cost by exactly one for its controller's Equipment")
	}
	engine.resolveTopOfStack(g, &TurnLog{})
	engine.applyStateBasedActions(g)

	if !g.Players[game.Player2].Graveyard.Contains(target.CardInstanceID) {
		_, onBattlefield := permanentByObjectID(g, target.ObjectID)
		t.Fatalf("Rebel Salvo target: graveyard=%v exile=%v battlefield=%v markedDamage=%d indestructible=%v",
			g.Players[game.Player2].Graveyard.Contains(target.CardInstanceID),
			g.Players[game.Player2].Exile.Contains(target.CardInstanceID),
			onBattlefield, target.MarkedDamage, hasKeyword(g, target, game.Indestructible))
	}
}

func TestKonstrariImproviserSoulTetherScenario(t *testing.T) {
	def := authoredParagraphCard(t, konstrariTestOracleID, "Konstrari Improviser")
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	cardID := addCardToHand(g, game.Player1, def)
	g.Players[game.Player1].ManaPool.Add(mana.C, 1)
	g.Players[game.Player1].ManaPool.Add(mana.G, 1)
	setMainPhasePriority(g, game.Player1)

	if !engine.applyAction(g, game.Player1, action.CastSpell(cardID, nil, 0, nil)) {
		t.Fatal("Konstrari Improviser did not cast for {1}{G}")
	}
	if g.Players[game.Player1].ManaPool.Total() != 0 {
		t.Fatal("Konstrari Improviser did not pay {1}{G}")
	}
	engine.resolveTopOfStack(g, &TurnLog{})
	permanent := permanentForCard(g, cardID)
	if permanent == nil || !permanent.Prepared {
		t.Fatalf("Konstrari Improviser permanent = %+v, want prepared", permanent)
	}

	g.Players[game.Player1].ManaPool.Add(mana.C, 2)
	g.Players[game.Player1].ManaPool.Add(mana.R, 1)
	preparedCast := action.CastSpellFaceFromZone(
		cardID, zone.Battlefield, game.FaceAlternate, nil, 0, nil,
	)
	if !engine.applyAction(g, game.Player1, preparedCast) {
		t.Fatal("Soul Tether prepared copy did not cast for {2}{R}")
	}
	if g.Players[game.Player1].ManaPool.Total() != 0 {
		t.Fatal("Soul Tether prepared copy did not pay {2}{R}")
	}
	stacked, ok := g.Stack.Peek()
	if !ok || !stacked.Copy || stacked.SourceZone != zone.Exile || stacked.SourceCardID != cardID {
		t.Fatalf("Soul Tether stack object = %+v, want its prepared copy cast from exile", stacked)
	}
	if permanent.Prepared {
		t.Fatal("casting Soul Tether's prepared copy did not unprepare Konstrari Improviser")
	}
	engine.resolveTopOfStack(g, &TurnLog{})

	var heartwood *game.Permanent
	for _, candidate := range g.Battlefield {
		if candidate.TokenDef != nil && candidate.TokenDef.CardFace.Name == "Heartwood" {
			heartwood = candidate
			break
		}
	}
	if heartwood == nil || heartwood.Controller != game.Player1 {
		t.Fatalf("Soul Tether Heartwood token = %+v, want one controlled by its caster", heartwood)
	}
	if !slices.Contains(heartwood.TokenDef.CardFace.Types, types.Artifact) ||
		!slices.Contains(heartwood.TokenDef.CardFace.Subtypes, types.Sub("Heartwood")) ||
		!slices.Contains(heartwood.TokenDef.CardFace.Colors, color.Red) ||
		!slices.Contains(heartwood.TokenDef.CardFace.Colors, color.Green) ||
		len(heartwood.TokenDef.CardFace.ManaAbilities) != 1 {
		t.Fatalf("Soul Tether token definition = %+v, want a Heartwood artifact", heartwood.TokenDef.CardFace)
	}
}

func TestMilitantInquisitorScenario(t *testing.T) {
	def := authoredParagraphCard(t, militantInquisitorTestOracleID, "Militant Inquisitor")
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	addCombatPermanent(g, game.Player1, paragraphTestEquipment("First unattached Equipment"))
	addCombatPermanent(g, game.Player1, paragraphTestEquipment("Second unattached Equipment"))
	addCombatPermanent(g, game.Player2, paragraphTestEquipment("Opponent Equipment"))
	cardID := addCardToHand(g, game.Player1, def)
	g.Players[game.Player1].ManaPool.Add(mana.C, 2)
	g.Players[game.Player1].ManaPool.Add(mana.W, 1)
	setMainPhasePriority(g, game.Player1)

	if !engine.applyAction(g, game.Player1, action.CastSpell(cardID, nil, 0, nil)) {
		t.Fatal("Militant Inquisitor did not cast for {2}{W}")
	}
	if g.Players[game.Player1].ManaPool.Total() != 0 {
		t.Fatal("Militant Inquisitor did not pay {2}{W}")
	}
	engine.resolveTopOfStack(g, &TurnLog{})
	inquisitor := permanentForCard(g, cardID)
	if inquisitor == nil {
		t.Fatal("Militant Inquisitor did not enter the battlefield")
	}
	if power := effectivePower(g, inquisitor); power != 4 {
		t.Fatalf("Militant Inquisitor power = %d, want 4 from two controlled Equipment", power)
	}
	if toughness, ok := effectiveToughness(g, inquisitor); !ok || toughness != 3 {
		t.Fatalf("Militant Inquisitor toughness = %d, %v; want 3", toughness, ok)
	}
}

func TestManifestDreadScenario(t *testing.T) {
	def := authoredParagraphCard(t, manifestDreadTestOracleID, "Manifest Dread")
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	discarded := addCardToLibrary(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:  "Manifest Dread card to graveyard",
		Types: []types.Card{types.Instant},
	}})
	manifestedCard := addCardToLibrary(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:      "Manifest Dread creature",
		ManaCost:  opt.Val(cost.Mana{cost.O(1), cost.G}),
		Types:     []types.Card{types.Creature},
		Power:     opt.Val(game.PT{Value: 3}),
		Toughness: opt.Val(game.PT{Value: 3}),
	}})
	spell := addCardToHand(g, game.Player1, def)
	g.Players[game.Player1].ManaPool.Add(mana.C, 2)
	g.Players[game.Player1].ManaPool.Add(mana.G, 3)
	setMainPhasePriority(g, game.Player1)

	if !engine.applyAction(g, game.Player1, action.CastSpell(spell, nil, 0, nil)) {
		t.Fatal("Manifest Dread did not cast for {1}{G}")
	}
	if g.Players[game.Player1].ManaPool.Total() != 3 {
		t.Fatal("Manifest Dread did not pay {1}{G}")
	}
	agents := [game.NumPlayers]PlayerAgent{
		game.Player1: &choiceOnlyAgent{choices: [][]int{{0}}},
	}
	engine.resolveTopOfStackWithChoices(g, agents, &TurnLog{})

	if !g.Players[game.Player1].Graveyard.Contains(discarded) {
		t.Fatal("Manifest Dread did not put the unchosen card into its controller's graveyard")
	}
	permanent := permanentForCard(g, manifestedCard)
	if permanent == nil || !permanent.FaceDown || permanent.FaceDownKind != game.FaceDownManifest {
		t.Fatalf("manifested permanent = %+v, want the selected creature face down", permanent)
	}
	if power := effectivePower(g, permanent); power != 2 {
		t.Fatalf("face-down manifested creature power = %d, want 2", power)
	}
	if toughness, ok := effectiveToughness(g, permanent); !ok || toughness != 2 {
		t.Fatalf("face-down manifested creature toughness = %d, %v; want 2", toughness, ok)
	}

	setMainPhasePriority(g, game.Player1)
	if !engine.applyAction(g, game.Player1, action.TurnFaceUp(permanent.ObjectID)) {
		t.Fatal("Manifest Dread creature did not turn face up for its {1}{G} mana cost")
	}
	if permanent.FaceDown || g.Players[game.Player1].ManaPool.Total() != 1 {
		t.Fatalf("manifested creature face-down=%v, mana=%d; want face up after paying {1}{G}",
			permanent.FaceDown, g.Players[game.Player1].ManaPool.Total())
	}
	if got := effectivePower(g, permanent); got != 3 {
		t.Fatalf("face-up manifested creature power = %d, want 3", got)
	}
}

func TestDevilsPlayScenario(t *testing.T) {
	def := authoredParagraphCard(t, devilsPlayTestOracleID, "Devil's Play")
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	cardID := addCardToHand(g, game.Player1, def)
	g.Players[game.Player1].ManaPool.Add(mana.R, 3)
	setMainPhasePriority(g, game.Player1)

	if !engine.applyAction(g, game.Player1, action.CastSpell(
		cardID, []game.Target{game.PlayerTarget(game.Player2)}, 2, nil,
	)) {
		t.Fatal("Devil's Play did not cast for X=2 and {X}{R}")
	}
	if g.Players[game.Player1].ManaPool.Total() != 0 {
		t.Fatal("Devil's Play did not pay {2}{R}")
	}
	engine.resolveTopOfStack(g, &TurnLog{})
	if g.Players[game.Player2].Life != 38 ||
		!g.Players[game.Player1].Graveyard.Contains(cardID) {
		t.Fatalf("first Devil's Play result: opponent life=%d; card in graveyard=%v, want 38 and true",
			g.Players[game.Player2].Life, g.Players[game.Player1].Graveyard.Contains(cardID))
	}

	g.Players[game.Player1].ManaPool.Add(mana.R, 5)
	setMainPhasePriority(g, game.Player1)
	if !engine.applyAction(g, game.Player1, action.CastSpellFaceFromZone(
		cardID, zone.Graveyard, game.FaceFront,
		[]game.Target{game.PlayerTarget(game.Player2)}, 2, nil,
	)) {
		t.Fatal("Devil's Play did not flash back for X=2 and {X}{R}{R}{R}")
	}
	if g.Players[game.Player1].ManaPool.Total() != 0 {
		t.Fatal("Devil's Play flashback did not pay {2}{R}{R}{R}")
	}
	engine.resolveTopOfStack(g, &TurnLog{})
	if g.Players[game.Player2].Life != 36 {
		t.Fatalf("opponent life after flashback = %d, want 36 after two 2-damage resolutions", g.Players[game.Player2].Life)
	}
	if !g.Players[game.Player1].Exile.Contains(cardID) ||
		g.Players[game.Player1].Graveyard.Contains(cardID) {
		t.Fatal("Devil's Play cast with flashback was not exiled after resolution")
	}
}

func TestOxiddaFinisherScenario(t *testing.T) {
	def := authoredParagraphCard(t, oxiddaFinisherTestOracleID, "Oxidda Finisher")
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	addCombatPermanent(g, game.Player1, paragraphTestEquipment("First Oxidda Equipment"))
	addCombatPermanent(g, game.Player1, paragraphTestEquipment("Second Oxidda Equipment"))
	addCombatPermanent(g, game.Player2, paragraphTestEquipment("Opponent Equipment"))
	cardID := addCardToHand(g, game.Player1, def)
	g.Players[game.Player1].ManaPool.Add(mana.C, 3)
	g.Players[game.Player1].ManaPool.Add(mana.R, 2)
	setMainPhasePriority(g, game.Player1)

	if !engine.applyAction(g, game.Player1, action.CastSpell(cardID, nil, 0, nil)) {
		t.Fatal("Oxidda Finisher did not cast for {3}{R}{R} with two controlled Equipment")
	}
	if g.Players[game.Player1].ManaPool.Total() != 0 {
		t.Fatal("Oxidda Finisher did not reduce its cost by exactly two for its controller's Equipment")
	}
	engine.resolveTopOfStack(g, &TurnLog{})
	finisher := permanentForCard(g, cardID)
	if finisher == nil || !hasKeyword(g, finisher, game.Trample) {
		t.Fatalf("Oxidda Finisher = %+v, want a creature with trample", finisher)
	}
	if power := effectivePower(g, finisher); power != 7 {
		t.Fatalf("Oxidda Finisher power = %d, want 7", power)
	}

	blocker := addCombatCreaturePermanentWithPower(g, game.Player2, 3)
	g.Combat = sharedBlockerCombat(g, blocker, finisher)
	engine.resolveCombatDamage(g, &TurnLog{})
	engine.applyStateBasedActions(g)
	if g.Players[game.Player2].Life != 36 {
		t.Fatalf("defending player life = %d, want 36 after 3 blocker damage and 4 trample damage", g.Players[game.Player2].Life)
	}
	if permanentWithCardOnBattlefield(g, blocker.CardInstanceID) {
		t.Fatal("Oxidda Finisher did not assign lethal damage to its blocker")
	}
}
