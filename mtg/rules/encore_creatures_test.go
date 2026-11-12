package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/cards"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/id"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

func addEncoreMana(g *game.Game, generic int, colored ...mana.Color) {
	for range generic {
		g.Players[game.Player1].ManaPool.Add(mana.C, 1)
	}
	for _, color := range colored {
		g.Players[game.Player1].ManaPool.Add(color, 1)
	}
}

func playEncoreFromGraveyard(t *testing.T, oracleID, name string, generic int, colored []mana.Color, setup func(*game.Game)) (*game.Game, *Engine) {
	t.Helper()
	def, err := cards.NewDefaultRegistry().LookupOracleID(oracleID)
	if err != nil {
		t.Fatalf("exact %s registry lookup failed: %v", name, err)
	}
	if def.Name != name {
		t.Fatalf("exact Oracle ID returned %q, want %q", def.Name, name)
	}

	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	cardID := addCardToGraveyard(g, game.Player1, def)
	if setup != nil {
		setup(g)
	}
	addEncoreMana(g, generic, colored...)
	setMainPhasePriority(g, game.Player1)
	engine := NewEngine(nil)
	encore := action.ActivateAbility(cardID, 0, nil, 0)
	if !containsAction(engine.legalActions(g, game.Player1), encore) {
		t.Fatalf("%s Encore was not available at sorcery timing with its full cost", name)
	}
	if !engine.applyActionWithChoices(g, game.Player1, encore, simPassPolicies(), &TurnLog{}) {
		t.Fatalf("%s Encore activation failed", name)
	}
	if !g.Players[game.Player1].Exile.Contains(cardID) || g.Players[game.Player1].Graveyard.Contains(cardID) {
		t.Fatalf("%s was not exiled from its graveyard as the Encore cost", name)
	}
	if got := g.Players[game.Player1].ManaPool.Total(); got != 0 {
		t.Fatalf("%s Encore left %d mana after paying its printed cost", name, got)
	}
	resolveStackWithTriggers(engine, g, simPassPolicies())
	return g, engine
}

func declareEncoreCopiesAttacking(t *testing.T, g *game.Game, engine *Engine, name string) map[game.PlayerID]id.ID {
	t.Helper()
	tokenIDs := playerTokensNamed(g, game.Player1, name)
	if len(tokenIDs) == 0 {
		t.Fatalf("Encore created no %s token copy", name)
	}
	var declarations []game.AttackDeclaration
	byOpponent := make(map[game.PlayerID]id.ID, len(tokenIDs))
	for _, tokenID := range tokenIDs {
		token, ok := permanentByObjectID(g, tokenID)
		if !ok {
			t.Fatalf("Encore token %d is not on the battlefield", tokenID)
		}
		required, directed := directedAttackRequiredPlayers(g, token)
		if !directed || len(required) != 1 {
			t.Fatalf("Encore token %d has no single required opponent", tokenID)
		}
		byOpponent[required[0]] = tokenID
		declarations = append(declarations, game.AttackDeclaration{
			Attacker: tokenID, Target: game.AttackTarget{Player: required[0]},
		})
	}
	g.Turn.Phase = game.PhaseCombat
	g.Turn.Step = game.StepDeclareAttackers
	g.Combat = &game.CombatState{}
	payload := mustDeclareAttackersPayload(t, action.DeclareAttackers(declarations))
	if !engine.applyDeclareAttackers(g, game.Player1, payload) {
		t.Fatalf("could not declare the Encore copies of %s as attackers", name)
	}
	return byOpponent
}

// Source review: Scryfall printing 444b75d9-f7c7-4a13-9318-02d7bbd80c9c for Oracle ID 393ef069-ea6b-4969-b878-d942281cc918 was fetched 2026-11-12; its Oracle text matches the pinned printing. The rulings URI returned 7 WotC rulings dated 2020-11-10 and 2023-07-28. Official Wizards CR effective 2026-09-25 was read from the TXT linked by the archived official rules-page snapshot https://web.archive.org/web/20260928215933/https://magic.wizards.com/en/rules. Sections 702.141a, 307.5/602.5d, 508.1d, 707.2, 603.7a-c/513.1, 702.10/302.6, 111.1/111.4, 700.4, and 603.6c require graveyard activation and sorcery timing, copies of printed text, attack-if-able/haste, delayed next-end-step handling, token creation, and a dies trigger on battlefield-to-graveyard movement. Full responses: .cardwork/source-review/encore-20261112-1028/444b75d9-f7c7-4a13-9318-02d7bbd80c9c-{card,rulings}.json; CR SHA-256 8d860e451f20f38865b725b42d82feb714c725373dd8f3b32b8652b3eeb070ca.
func TestExquisiteHuntmasterEncoreCopyCreatesElfAfterDying(t *testing.T) {
	g, engine := playEncoreFromGraveyard(t, "393ef069-ea6b-4969-b878-d942281cc918", "Exquisite Huntmaster", 4, []mana.Color{mana.B}, nil)
	copies := playerTokensNamed(g, game.Player1, "Exquisite Huntmaster")
	if len(copies) == 0 {
		t.Fatal("Encore created no Exquisite Huntmaster copy")
	}
	if _, destroyed := destroyPermanent(g, copies[0]); !destroyed {
		t.Fatal("could not destroy an Encore copy")
	}
	resolveStackWithTriggers(engine, g, simPassPolicies())
	if elves := playerTokensNamed(g, game.Player1, "Elf Warrior"); len(elves) != 1 {
		t.Fatalf("the dying Encore copy created %d Elf Warrior tokens, want 1", len(elves))
	}
}

// Source review: Scryfall printing 08241f94-8b5e-4f9b-8120-8175e3256e35 for Oracle ID 4d211447-9f18-4d23-8f19-cdf6ad3c68ba was fetched 2026-11-12 and matches the pinned text; its rulings URI returned 8 WotC rulings dated 2020-11-10 and 2023-07-28, including that its characteristic-defining power/toughness ability functions in every zone. Official Wizards CR effective 2026-09-25 (TXT linked by https://web.archive.org/web/20260928215933/https://magic.wizards.com/en/rules) was read at 702.141a, 307.5/602.5d, 707.2, 604.3/604.3a, and 208.2a: Encore copies printed characteristics, while the life-total CDA defines both starred values everywhere. Full responses: .cardwork/source-review/encore-20261112-1028/08241f94-8b5e-4f9b-8120-8175e3256e35-{card,rulings}.json; CR SHA-256 8d860e451f20f38865b725b42d82feb714c725373dd8f3b32b8652b3eeb070ca.
func TestSoulOfEternityEncoreCopyUsesControllerLifeTotal(t *testing.T) {
	g, _ := playEncoreFromGraveyard(t, "4d211447-9f18-4d23-8f19-cdf6ad3c68ba", "Soul of Eternity", 7, []mana.Color{mana.W, mana.W}, nil)
	copies := playerTokensNamed(g, game.Player1, "Soul of Eternity")
	if len(copies) == 0 {
		t.Fatal("Encore created no Soul of Eternity copy")
	}
	copy, ok := permanentByObjectID(g, copies[0])
	if !ok {
		t.Fatal("Soul of Eternity Encore copy is not on the battlefield")
	}
	want := g.Players[game.Player1].Life
	power := effectivePower(g, copy)
	toughness, _ := effectiveToughness(g, copy)
	if power != want || toughness != want {
		t.Fatalf("Encore copy is %d/%d at life %d, want %d/%d", power, toughness, want, want, want)
	}
}

// Source review: Scryfall printing 41163f43-7357-4bbc-a1e7-9c8070135c93 for Oracle ID 4f57a320-b6b5-49de-af0a-f2e2649758a7 was fetched 2026-11-12; its rulings URI returned 10 WotC rulings, all dated 2024-04-12, including the permanent-card definition, X=0 for mana value outside the stack, and Aura entry choices. Official Wizards CR effective 2026-09-25 (TXT linked by https://web.archive.org/web/20260928215933/https://magic.wizards.com/en/rules) was read at 702.141a, 307.5/602.5d, 603.6a, 603.3d, 115.1d/115.2, 202.3, and 608.2b: each entering copy triggers, chooses its graveyard-card target on the stack, and affects it only if still legal on resolution. Full responses: .cardwork/source-review/encore-20261112-1028/41163f43-7357-4bbc-a1e7-9c8070135c93-{card,rulings}.json; CR SHA-256 8d860e451f20f38865b725b42d82feb714c725373dd8f3b32b8652b3eeb070ca.
func TestAngelOfIndemnityEncoreCopyReturnsTargetPermanentCard(t *testing.T) {
	var targetID id.ID
	g, _ := playEncoreFromGraveyard(t, "4f57a320-b6b5-49de-af0a-f2e2649758a7", "Angel of Indemnity", 6, []mana.Color{mana.W, mana.W}, func(g *game.Game) {
		targetID = addCardToGraveyard(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
			Name: "Four-Mana Creature", ManaCost: opt.Val(cost.Mana{cost.O(4)}),
			Types: []types.Card{types.Creature}, Power: opt.Val(game.PT{Value: 3}), Toughness: opt.Val(game.PT{Value: 3}),
		}})
	})
	if targetID == 0 || g.Players[game.Player1].Graveyard.Contains(targetID) {
		t.Fatal("the targeted permanent card did not return from the graveyard to the battlefield")
	}
	if permanentForCard(g, targetID) == nil {
		t.Fatal("the targeted permanent card is not on the battlefield")
	}
}

// Source review: Scryfall printing 9d3fb5f1-baf5-4934-aac7-ffbaa85b2428 for Oracle ID 6680a5f7-3ac1-4550-a29a-9553ad7ae465 was fetched 2026-11-12 and matches the pinned Oracle text; its rulings URI returned 7 WotC rulings dated 2020-11-10 and 2023-07-28, including that changing a blocker after it blocks does not undo that block. Official Wizards CR effective 2026-09-25 (TXT linked by https://web.archive.org/web/20260928215933/https://magic.wizards.com/en/rules) was read at 702.141a, 307.5/602.5d, 508.1d/508.2a, 603.3d, 611.2a, and 613.4c: Encore tokens can attack when able, attacking triggers occur at declaration, targets are chosen for the trigger, and -1/-1 lasts through the end of turn. Full responses: .cardwork/source-review/encore-20261112-1028/9d3fb5f1-baf5-4934-aac7-ffbaa85b2428-{card,rulings}.json; CR SHA-256 8d860e451f20f38865b725b42d82feb714c725373dd8f3b32b8652b3eeb070ca.
func TestBriarbladeAdeptEncoreCopyWeakensCreatureWhenAttacking(t *testing.T) {
	var target *game.Permanent
	g, engine := playEncoreFromGraveyard(t, "6680a5f7-3ac1-4550-a29a-9553ad7ae465", "Briarblade Adept", 3, []mana.Color{mana.B}, func(g *game.Game) {
		target = addCombatCreaturePermanentWithPower(g, game.Player2, 5)
	})
	if target == nil {
		t.Fatal("opponent target was not created")
	}
	declareEncoreCopiesAttacking(t, g, engine, "Briarblade Adept")
	agents := simPassPolicies()
	log := &TurnLog{}
	if !engine.putTriggeredAbilitiesOnStackWithChoices(g, agents, log) {
		t.Fatal("Briarblade Adept's attack trigger was not put on the stack")
	}
	engine.resolveTopOfStackWithChoices(g, agents, log)
	power := effectivePower(g, target)
	toughness, _ := effectiveToughness(g, target)
	if power != 4 || toughness != 4 {
		t.Fatalf("target creature became %d/%d, want 4/4 after one resolved trigger", power, toughness)
	}
}

// Source review: Scryfall printing 0f020359-5ad1-43c4-8588-172be635cccc for Oracle ID 7d9fc9e7-d80b-49c3-871c-ed25b3059ae8 was fetched 2026-11-12; its rulings URI returned 7 WotC rulings dated 2020-11-10 and 2023-07-28. Official Wizards CR effective 2026-09-25 (TXT linked by https://web.archive.org/web/20260928215933/https://magic.wizards.com/en/rules) was read at 702.141a, 307.5/602.5d, 707.2, 603.7a-c/513.1, 700.4/603.6c, and 111.10a: the copy carries the printed dies trigger; a dying token creates the predefined Treasure artifact with its mana ability. Full responses: .cardwork/source-review/encore-20261112-1028/0f020359-5ad1-43c4-8588-172be635cccc-{card,rulings}.json; CR SHA-256 8d860e451f20f38865b725b42d82feb714c725373dd8f3b32b8652b3eeb070ca.
func TestImpulsivePilfererEncoreCopyCreatesTreasureAfterDying(t *testing.T) {
	g, engine := playEncoreFromGraveyard(t, "7d9fc9e7-d80b-49c3-871c-ed25b3059ae8", "Impulsive Pilferer", 3, []mana.Color{mana.R}, nil)
	copies := playerTokensNamed(g, game.Player1, "Impulsive Pilferer")
	if len(copies) == 0 {
		t.Fatal("Encore created no Impulsive Pilferer copy")
	}
	if _, destroyed := destroyPermanent(g, copies[0]); !destroyed {
		t.Fatal("could not destroy an Encore copy")
	}
	resolveStackWithTriggers(engine, g, simPassPolicies())
	if treasures := playerTokensNamed(g, game.Player1, "Treasure"); len(treasures) != 1 {
		t.Fatalf("the dying Encore copy created %d Treasures, want 1", len(treasures))
	}
}

// Source review: Scryfall printing 56b2b413-e00f-440d-9d27-92eb069b8c71 for Oracle ID 92c0346d-c375-4851-9806-5d3b9b83425f was fetched 2026-11-12; its rulings URI returned 7 WotC rulings dated 2020-11-10 and 2023-07-28. Official Wizards CR effective 2026-09-25 (TXT linked by https://web.archive.org/web/20260928215933/https://magic.wizards.com/en/rules) was read at 702.141a, 307.5/602.5d, 603.6a, 603.3d, and 122.1a: each entering copy triggers and puts a +1/+1 counter on the target creature chosen as the trigger is stacked. Full responses: .cardwork/source-review/encore-20261112-1028/56b2b413-e00f-440d-9d27-92eb069b8c71-{card,rulings}.json; CR SHA-256 8d860e451f20f38865b725b42d82feb714c725373dd8f3b32b8652b3eeb070ca.
func TestKinsbaileCourierEncoreCopyAddsCounterOnEntry(t *testing.T) {
	var target *game.Permanent
	g, _ := playEncoreFromGraveyard(t, "92c0346d-c375-4851-9806-5d3b9b83425f", "Kinsbaile Courier", 2, []mana.Color{mana.W}, func(g *game.Game) {
		target = addCombatCreaturePermanentWithPower(g, game.Player1, 2)
	})
	if target == nil {
		t.Fatal("target creature was not created")
	}
	current, ok := permanentByObjectID(g, target.ObjectID)
	if !ok {
		t.Fatal("target creature left the battlefield")
	}
	if got := current.Counters.Get(counter.PlusOnePlusOne); got == 0 {
		t.Fatal("Kinsbaile Courier's Encore entry trigger did not put a +1/+1 counter on the target")
	}
}

// Source review: Scryfall printing b46d498b-dac3-4ab0-bcd6-514025a046aa for Oracle ID 9b984236-39cf-4552-827b-e81c26cfb388 was fetched 2026-11-12; its rulings URI returned 8 WotC rulings, all dated 2024-06-07, including the full Encore definition, graveyard-exile cost, opponent count, and copy-of-printed-values guidance. Official Wizards CR effective 2026-09-25 (TXT linked by https://web.archive.org/web/20260928215933/https://magic.wizards.com/en/rules) was read at 702.141a, 307.5/602.5d, 603.6a, 707.2, 111.1/111.4, and 702.9: each entrant triggers its Dragon creation, and the specified red 5/5 Dragon token has flying. Full responses: .cardwork/source-review/encore-20261112-1028/b46d498b-dac3-4ab0-bcd6-514025a046aa-{card,rulings}.json; CR SHA-256 8d860e451f20f38865b725b42d82feb714c725373dd8f3b32b8652b3eeb070ca.
func TestBroodmateTyrantEncoreCopyCreatesFlyingDragon(t *testing.T) {
	g, _ := playEncoreFromGraveyard(t, "9b984236-39cf-4552-827b-e81c26cfb388", "Broodmate Tyrant", 5, []mana.Color{mana.B, mana.R, mana.G}, nil)
	dragons := playerTokensNamed(g, game.Player1, "Dragon")
	if len(dragons) == 0 {
		t.Fatal("Broodmate Tyrant's Encore entry trigger created no Dragon")
	}
	dragon, ok := permanentByObjectID(g, dragons[0])
	if !ok {
		t.Fatal("created Dragon is not on the battlefield")
	}
	power := effectivePower(g, dragon)
	toughness, _ := effectiveToughness(g, dragon)
	if power != 5 || toughness != 5 || !hasKeyword(g, dragon, game.Flying) {
		t.Fatalf("created token is %d/%d flying=%t, want a 5/5 flying Dragon", power, toughness, hasKeyword(g, dragon, game.Flying))
	}
}

// Source review: Scryfall printing 3eb7af70-e024-44bf-a52b-209074b005a3 for Oracle ID dad2965a-ceb4-42ac-bb62-de7eea500011 was fetched 2026-11-12; its rulings URI returned 8 WotC rulings dated 2020-11-10 and 2023-07-28, including the rule that a blocker that already blocked remains blocking if its power later changes. Official Wizards CR effective 2026-09-25 (TXT linked by https://web.archive.org/web/20260928215933/https://magic.wizards.com/en/rules) was read at 702.141a, 307.5/602.5d, 508.1d, 509.1b, 707.2, and 603.7a-c/513.1: each Encore copy is a printed-value token with an attack-if-able requirement; a creature with this static restriction rejects blockers whose current power is 2 or less. Full responses: .cardwork/source-review/encore-20261112-1028/3eb7af70-e024-44bf-a52b-209074b005a3-{card,rulings}.json; CR SHA-256 8d860e451f20f38865b725b42d82feb714c725373dd8f3b32b8652b3eeb070ca.
func TestFinCladeFugitivesEncoreCopyRejectsSmallBlocker(t *testing.T) {
	g, engine := playEncoreFromGraveyard(t, "dad2965a-ceb4-42ac-bb62-de7eea500011", "Fin-Clade Fugitives", 4, []mana.Color{mana.G}, nil)
	byOpponent := declareEncoreCopiesAttacking(t, g, engine, "Fin-Clade Fugitives")
	attackerID, ok := byOpponent[game.Player2]
	if !ok {
		t.Fatal("no Fin-Clade Fugitives copy is attacking Player2")
	}
	attacker, ok := permanentByObjectID(g, attackerID)
	if !ok {
		t.Fatal("attacking Fin-Clade Fugitives copy is missing")
	}
	weak := addCombatCreaturePermanentWithPower(g, game.Player2, 2)
	strong := addCombatCreaturePermanentWithPower(g, game.Player2, 3)
	g.Turn.Step = game.StepDeclareBlockers
	weakBlock := mustDeclareBlockersPayload(t, action.DeclareBlockers([]game.BlockDeclaration{{Blocker: weak.ObjectID, Blocking: attacker.ObjectID}}))
	if engine.applyDeclareBlockers(g, game.Player2, weakBlock) {
		t.Fatal("a power-2 creature illegally blocked Fin-Clade Fugitives")
	}
	strongBlock := mustDeclareBlockersPayload(t, action.DeclareBlockers([]game.BlockDeclaration{{Blocker: strong.ObjectID, Blocking: attacker.ObjectID}}))
	if !engine.applyDeclareBlockers(g, game.Player2, strongBlock) {
		t.Fatal("a power-3 creature could not block Fin-Clade Fugitives")
	}
}

// Source review: Scryfall printing c3b74135-b8db-4485-a4b1-7aaa3ef28072 for Oracle ID f27972da-ad2f-4df5-82d3-c909cce8cc91 was fetched 2026-11-12; its rulings URI returned 7 WotC rulings dated 2020-11-10 and 2023-07-28. Official Wizards CR effective 2026-09-25 (TXT linked by https://web.archive.org/web/20260928215933/https://magic.wizards.com/en/rules) was read at 702.141a, 307.5/602.5d, 707.2, 603.7a-c/513.1, 700.4/603.6c, and 121.1: the copied printed dies ability triggers when its token is put into a graveyard from the battlefield and draws the top card of its controller's library. Full responses: .cardwork/source-review/encore-20261112-1028/c3b74135-b8db-4485-a4b1-7aaa3ef28072-{card,rulings}.json; CR SHA-256 8d860e451f20f38865b725b42d82feb714c725373dd8f3b32b8652b3eeb070ca.
func TestTroveTrackerEncoreCopyDrawsWhenItDies(t *testing.T) {
	var drawn id.ID
	g, engine := playEncoreFromGraveyard(t, "f27972da-ad2f-4df5-82d3-c909cce8cc91", "Trove Tracker", 5, []mana.Color{mana.U, mana.U}, func(g *game.Game) {
		drawn = addCardToLibrary(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Trove Tracker draw"}})
	})
	copies := playerTokensNamed(g, game.Player1, "Trove Tracker")
	if len(copies) == 0 {
		t.Fatal("Encore created no Trove Tracker copy")
	}
	if _, destroyed := destroyPermanent(g, copies[0]); !destroyed {
		t.Fatal("could not destroy an Encore copy")
	}
	resolveStackWithTriggers(engine, g, simPassPolicies())
	if drawn == 0 || !g.Players[game.Player1].Hand.Contains(drawn) {
		t.Fatal("the dying Encore copy did not draw the top library card")
	}
}
