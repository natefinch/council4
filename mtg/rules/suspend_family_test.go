package rules

import (
	"fmt"
	"testing"

	"github.com/natefinch/council4/mtg/cards"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/id"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// The saved batch rules pin is 2026-09-25; the official WotC rules page listed
// the 2026-11-06 edition as current, so this later edition was read.
// Source review: official WotC Comprehensive Rules TXT, effective 2026-11-06,
// https://media.wizards.com/2026/downloads/MagicCompRules%2020261106.txt,
// SHA-256 524f5195a1e5f54bb13c735edf399ac717a4458e2713f398d1a3e6877186b829.
// Oracle source review: pinned corpus /Users/natefinch/Library/Caches/council4/pinned/6da9ced0e19bd384771ab8f14051cc45a8dc90302082b875fa6e039d2399a1b4.json; current Scryfall card and rulings responses were compared on 2026-11-10.
// Full Scryfall printing and rulings responses (HTTP 200) are retained under .cardwork/suspend-family-sources; Oracle text was compared for all ten pinned identities.
func suspendFamilyCast(t *testing.T, name, oracleID string, counters int, payment []mana.Color, agent PlayerAgent) (*game.Game, *Engine, id.ID, *game.StackObject) {
	t.Helper()
	registry := cards.NewDefaultRegistry()
	def, err := registry.LookupOracleID(oracleID)
	if err != nil {
		t.Fatalf("exact %s default-registry lookup failed: %v", name, err)
	}
	if def.Name != name {
		t.Fatalf("exact Oracle ID %s returned %q, want %q", oracleID, def.Name, name)
	}

	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	cardID := addCardToHand(g, game.Player1, def)
	for _, color := range payment {
		g.Players[game.Player1].ManaPool.Add(color, 1)
	}
	setMainPhasePriority(g, game.Player1)
	engine := NewEngine(nil)
	if !engine.applyAction(g, game.Player1, action.SuspendCard(cardID)) {
		t.Fatalf("%s could not be suspended from hand", name)
	}
	if got := g.Players[game.Player1].ManaPool.Total(); got != 0 {
		t.Fatalf("%s left %d mana after paying its suspend cost, want 0", name, got)
	}
	if got := g.ExileCounterCount(cardID, counter.Time); got != counters || !cardIsSuspended(g, cardID) {
		t.Fatalf("%s time counters = %d, want suspended with %d", name, got, counters)
	}
	requireExileTimeCounterEvent(t, g, game.EventCountersAdded, cardID, 0, counters)
	if !g.Players[game.Player1].Exile.Contains(cardID) {
		t.Fatalf("%s was not exiled by its suspend action", name)
	}

	agents := suspendAgents(agent)
	for remaining := counters; remaining > 0; remaining-- {
		card, _ := g.GetCardInstance(cardID)
		zoneVersion := card.ZoneVersion
		engine.processSuspendUpkeep(g, game.Player1, agents, &TurnLog{})
		removed := requireExileTimeCounterEvent(t, g, game.EventCountersRemoved, cardID, remaining, 1)
		if removed.CardZoneVersion != zoneVersion {
			t.Fatalf("%s removal event zone version = %d, want the exiled card's %d", name, removed.CardZoneVersion, zoneVersion)
		}
		if remaining > 1 {
			if got := g.ExileCounterCount(cardID, counter.Time); got != remaining-1 {
				t.Fatalf("%s time counters after upkeep = %d, want %d", name, got, remaining-1)
			}
		}
	}
	obj, event := requireSuspendCast(t, g, cardID)
	if event.FromZone != zone.Exile {
		t.Fatalf("%s suspend cast came from %v, want exile", name, event.FromZone)
	}
	if obj.ManaSpentToCast != 0 {
		t.Fatalf("%s suspend cast spent %d mana, want the free cast", name, obj.ManaSpentToCast)
	}
	return g, engine, cardID, obj
}

func suspendFamilyPermanent(t *testing.T, g *game.Game, cardID id.ID) *game.Permanent {
	t.Helper()
	for _, permanent := range g.Battlefield {
		if permanent.CardInstanceID == cardID {
			return permanent
		}
	}
	t.Fatalf("suspended card %v did not become a battlefield permanent", cardID)
	return nil
}

func suspendFamilyCreatureDef(name string, power, toughness int, cardTypes []types.Card, colors []color.Color, keywords ...game.Keyword) *game.CardDef {
	def := &game.CardDef{CardFace: game.CardFace{
		Name:      name,
		Types:     cardTypes,
		Colors:    colors,
		Power:     opt.Val(game.PT{Value: power}),
		Toughness: opt.Val(game.PT{Value: toughness}),
	}}
	if len(keywords) != 0 {
		def.StaticAbilities = []game.StaticAbility{{KeywordAbilities: game.SimpleKeywords(keywords...)}}
	}
	return def
}

func suspendFamilyDeclareAttacks(t *testing.T, g *game.Game, engine *Engine, attackingPlayer game.PlayerID, target game.PlayerID, attackers ...*game.Permanent) {
	t.Helper()
	g.Turn.ActivePlayer = attackingPlayer
	g.Turn.PriorityPlayer = attackingPlayer
	g.Turn.Phase = game.PhaseCombat
	g.Turn.Step = game.StepDeclareAttackers
	g.Combat = &game.CombatState{}
	declarations := make([]game.AttackDeclaration, 0, len(attackers))
	for _, attacker := range attackers {
		declarations = append(declarations, game.AttackDeclaration{
			Attacker: attacker.ObjectID,
			Target:   game.AttackTarget{Player: target},
		})
	}
	payload, ok := action.DeclareAttackers(declarations).DeclareAttackersPayload()
	if !ok || !engine.applyDeclareAttackers(g, attackingPlayer, payload) {
		t.Fatalf("could not declare %d attacker(s)", len(attackers))
	}
}

func suspendFamilyDeclareBlocks(g *game.Game, engine *Engine, defendingPlayer game.PlayerID, declarations ...game.BlockDeclaration) bool {
	g.Turn.Step = game.StepDeclareBlockers
	g.Turn.PriorityPlayer = defendingPlayer
	payload, ok := action.DeclareBlockers(declarations).DeclareBlockersPayload()
	return ok && engine.applyDeclareBlockers(g, defendingPlayer, payload)
}

func suspendFamilyBlockControl(t *testing.T, attackerDef, blockerDef *game.CardDef) {
	t.Helper()
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	attacker := addCombatPermanent(g, game.Player1, attackerDef)
	blocker := addCombatPermanent(g, game.Player2, blockerDef)
	suspendFamilyDeclareAttacks(t, g, engine, game.Player1, game.Player2, attacker)
	if !suspendFamilyDeclareBlocks(g, engine, game.Player2, game.BlockDeclaration{
		Blocker: blocker.ObjectID, Blocking: attacker.ObjectID,
	}) {
		t.Fatal("control attacker was not blockable by the same creature")
	}
}

// Source review: Oracle 9728dec9-d482-4c7a-8cdc-44d010dc878d, printing 9ec075ba-db56-4dcf-b1dc-fe6270b7ab36; pinned text and https://api.scryfall.com/cards/9ec075ba-db56-4dcf-b1dc-fe6270b7ab36/rulings checked 2026-11-10 (16 rulings, including 2013).
// CR 2026-11-06 (official TXT URL and SHA-256 above), read 702.62a-d, 601.2, 116.2f, and 121.2; four suspend counters lead to a free targeted cast and three cards drawn by the chosen player.
func TestAncestralVisionSuspendTargetsPlayerAndDrawsThree(t *testing.T) {
	target := game.PlayerTarget(game.Player2)
	agent := &suspendTargetAgent{target: target}
	g, engine, cardID, spell := suspendFamilyCast(t, "Ancestral Vision", "9728dec9-d482-4c7a-8cdc-44d010dc878d", 4, []mana.Color{mana.U}, agent)
	if len(spell.Targets) != 1 || spell.Targets[0] != target {
		t.Fatalf("Ancestral Vision targets = %v, want Player2", spell.Targets)
	}

	drawn := make([]id.ID, 3)
	for i := range drawn {
		drawn[i] = addCardToLibrary(g, game.Player2, &game.CardDef{CardFace: game.CardFace{Name: fmt.Sprintf("Vision draw %d", i)}})
	}
	engine.resolveTopOfStack(g, &TurnLog{})
	for _, drawnID := range drawn {
		if !g.Players[game.Player2].Hand.Contains(drawnID) {
			t.Fatalf("target player did not draw card %v", drawnID)
		}
		if g.Players[game.Player1].Hand.Contains(drawnID) {
			t.Fatalf("non-target player drew card %v", drawnID)
		}
	}
	if !g.Players[game.Player1].Graveyard.Contains(cardID) {
		t.Fatal("resolved Ancestral Vision did not go to its owner's graveyard")
	}
}

// Source review: Oracle 9cdc9f99-c6fa-40cd-90b0-e47d43a8cc3c, printing 4a0c73a9-38a2-4a3d-938e-b81ba166f0ff; pinned text and https://api.scryfall.com/cards/4a0c73a9-38a2-4a3d-938e-b81ba166f0ff/rulings checked 2026-11-10 (13 rulings).
// CR 2026-11-06 (official TXT URL and SHA-256 above), read 702.62a-d, 701.23a-e, and 701.24a; this qualified hidden-zone search may fail to find, moves a chosen basic land to the battlefield, does not reveal it, then shuffles.
func TestSearchForTomorrowSuspendFindsBasicLand(t *testing.T) {
	g, engine, _, _ := suspendFamilyCast(t, "Search for Tomorrow", "9cdc9f99-c6fa-40cd-90b0-e47d43a8cc3c", 2, []mana.Color{mana.G}, nil)
	basicForest := &game.CardDef{CardFace: game.CardFace{
		Name:       "Basic Forest",
		Types:      []types.Card{types.Land},
		Supertypes: []types.Super{types.Basic},
		Subtypes:   []types.Sub{types.Forest},
	}}
	nonbasicForest := &game.CardDef{CardFace: game.CardFace{
		Name:     "Nonbasic Forest",
		Types:    []types.Card{types.Land},
		Subtypes: []types.Sub{types.Forest},
	}}
	basicID := addCardToLibrary(g, game.Player1, basicForest)
	nonbasicID := addCardToLibrary(g, game.Player1, nonbasicForest)
	searchAgent := &suspendFamilySearchAgent{}
	engine.resolveTopOfStackWithChoices(g, suspendAgents(searchAgent), &TurnLog{})
	if len(searchAgent.requests) != 1 || searchAgent.requests[0].Kind != game.ChoiceSearch || len(searchAgent.requests[0].Options) != 1 {
		t.Fatalf("Search for Tomorrow search choices = %+v, want the single eligible basic land", searchAgent.requests)
	}

	basic := suspendFamilyPermanent(t, g, basicID)
	if basic.Controller != game.Player1 {
		t.Fatalf("found basic land controller = %v, want Player1", basic.Controller)
	}
	if g.Players[game.Player1].Library.Contains(basicID) || !g.Players[game.Player1].Library.Contains(nonbasicID) {
		t.Fatal("Search for Tomorrow did not move only the eligible basic land from the library")
	}
	for _, event := range g.Events {
		if event.Kind == game.EventCardRevealed && event.CardID == basicID {
			t.Fatal("Search for Tomorrow revealed the found land")
		}
	}
}

// Source review: Oracle 9dea848b-d758-46d0-829d-b096061fd460, printing cdc9f54e-7fba-4f03-83ca-6d293dffc07a; pinned text and https://api.scryfall.com/cards/cdc9f54e-7fba-4f03-83ca-6d293dffc07a/rulings checked 2026-11-10 (14 rulings).
// CR 2026-11-06 (official TXT URL and SHA-256 above), read 602.2, 611.2a, 702.10, and 702.62a-d; after paying {U} for its suspend route, its {U} activation gives itself +1/+0 until end of turn.
func TestVisceridDeepwalkerSuspendPump(t *testing.T) {
	g, engine, cardID, _ := suspendFamilyCast(t, "Viscerid Deepwalker", "9dea848b-d758-46d0-829d-b096061fd460", 4, []mana.Color{mana.U}, nil)
	engine.resolveTopOfStack(g, &TurnLog{})
	creature := suspendFamilyPermanent(t, g, cardID)
	g.Players[game.Player1].ManaPool.Add(mana.U, 1)
	if !engine.applyAction(g, game.Player1, action.ActivateAbility(creature.ObjectID, 0, nil, 0)) {
		t.Fatal("Viscerid Deepwalker could not pay {U} to activate its pump ability")
	}
	engine.resolveTopOfStack(g, &TurnLog{})
	if got := effectivePower(g, creature); got != 3 {
		t.Fatalf("Viscerid Deepwalker power = %d after pumping, want 3", got)
	}
}

// Source review: Oracle a1472a2d-e6f7-4d65-8bb2-4ecfae2d2764, printing 55e0caef-215a-4433-8a85-322e62e3590d; pinned text and https://api.scryfall.com/cards/55e0caef-215a-4433-8a85-322e62e3590d/rulings checked 2026-11-10 (14 rulings).
// CR 2026-11-06 (official TXT URL and SHA-256 above), read 510.4, 702.7, and 702.62a-d; First Strike assigns damage in the first combat-damage step, so the 4/1 survives killing the 2/4 blocker before it can deal damage.
func TestKeldonHalberdierSuspendFirstStrike(t *testing.T) {
	g, engine, cardID, _ := suspendFamilyCast(t, "Keldon Halberdier", "a1472a2d-e6f7-4d65-8bb2-4ecfae2d2764", 4, []mana.Color{mana.R}, nil)
	engine.resolveTopOfStack(g, &TurnLog{})
	attacker := suspendFamilyPermanent(t, g, cardID)
	blocker := addCombatPermanent(g, game.Player2, suspendFamilyCreatureDef("Larger Blocker", 2, 4, []types.Card{types.Creature}, []color.Color{color.Green}))
	suspendFamilyDeclareAttacks(t, g, engine, game.Player1, game.Player2, attacker)
	if !suspendFamilyDeclareBlocks(g, engine, game.Player2, game.BlockDeclaration{Blocker: blocker.ObjectID, Blocking: attacker.ObjectID}) {
		t.Fatal("ordinary creature could not block Keldon Halberdier")
	}
	if !hasKeyword(g, attacker, game.FirstStrike) {
		t.Fatal("Keldon Halberdier did not have first strike")
	}
	g.Turn.Step = game.StepFirstStrikeDamage
	engine.resolveCombatDamagePass(g, firstStrikeCombatDamage, &TurnLog{})
	engine.applyStateBasedActions(g)
	g.Turn.Step = game.StepCombatDamage
	engine.resolveCombatDamagePass(g, normalCombatDamage, &TurnLog{})
	engine.applyStateBasedActions(g)
	if _, alive := permanentByObjectID(g, attacker.ObjectID); !alive {
		t.Fatal("Keldon Halberdier died before its first-strike damage could prevent the blocker's damage")
	}
	if _, alive := permanentByObjectID(g, blocker.ObjectID); alive {
		t.Fatal("Keldon Halberdier's first-strike damage did not destroy the blocker")
	}
}

// Source review: Oracle b3a84b75-a6a2-4521-a9f1-92596796c703, printing c8680e7c-f180-45b9-9af6-a8e93037b851; pinned text and https://api.scryfall.com/cards/c8680e7c-f180-45b9-9af6-a8e93037b851/rulings checked 2026-11-10 (14 rulings).
// CR 2026-11-06 (official TXT URL and SHA-256 above), read 509.1, 702.36, and 702.62a-d; a nonblack, nonartifact creature may not block Fear, while artifact and black creatures may.
func TestCorpulentCorpseSuspendFear(t *testing.T) {
	groundDef := suspendFamilyCreatureDef("Ground Blocker", 2, 2, []types.Card{types.Creature}, []color.Color{color.Green})
	suspendFamilyBlockControl(t,
		suspendFamilyCreatureDef("Vanilla Attacker", 3, 3, []types.Card{types.Creature}, []color.Color{color.Green}),
		groundDef,
	)

	g, engine, cardID, _ := suspendFamilyCast(t, "Corpulent Corpse", "b3a84b75-a6a2-4521-a9f1-92596796c703", 5, []mana.Color{mana.B}, nil)
	engine.resolveTopOfStack(g, &TurnLog{})
	attacker := suspendFamilyPermanent(t, g, cardID)
	groundBlocker := addCombatPermanent(g, game.Player2, groundDef)
	suspendFamilyDeclareAttacks(t, g, engine, game.Player1, game.Player2, attacker)
	if suspendFamilyDeclareBlocks(g, engine, game.Player2, game.BlockDeclaration{Blocker: groundBlocker.ObjectID, Blocking: attacker.ObjectID}) {
		t.Fatal("Fear allowed a nonblack, nonartifact creature to block Corpulent Corpse")
	}

	artifactBlocker := addCombatPermanent(g, game.Player2, suspendFamilyCreatureDef(
		"Artifact Blocker", 2, 2, []types.Card{types.Artifact, types.Creature}, nil,
	))
	blackBlocker := addCombatPermanent(g, game.Player2, suspendFamilyCreatureDef(
		"Black Blocker", 2, 2, []types.Card{types.Creature}, []color.Color{color.Black},
	))
	if !suspendFamilyDeclareBlocks(g, engine, game.Player2,
		game.BlockDeclaration{Blocker: artifactBlocker.ObjectID, Blocking: attacker.ObjectID},
		game.BlockDeclaration{Blocker: blackBlocker.ObjectID, Blocking: attacker.ObjectID},
	) {
		t.Fatal("Fear rejected an artifact creature and a black creature blocking Corpulent Corpse")
	}
}

// Source review: Oracle c77b89ec-4514-4e62-a508-2012f38a3e15, printing 3779a76a-0171-4714-9713-a65dc173a33f; pinned text and https://api.scryfall.com/cards/3779a76a-0171-4714-9713-a65dc173a33f/rulings checked 2026-11-10 (14 rulings).
// CR 2026-11-06 (official TXT URL and SHA-256 above), read 509.1, 702.9, and 702.62a-d; a nonflying creature cannot block the suspended 4/4 flyer, while a flying blocker can.
func TestErrantEphemeronSuspendFlying(t *testing.T) {
	groundDef := suspendFamilyCreatureDef("Ground Blocker", 2, 2, []types.Card{types.Creature}, []color.Color{color.Green})
	suspendFamilyBlockControl(t,
		suspendFamilyCreatureDef("Vanilla Attacker", 4, 4, []types.Card{types.Creature}, []color.Color{color.Blue}),
		groundDef,
	)

	g, engine, cardID, _ := suspendFamilyCast(t, "Errant Ephemeron", "c77b89ec-4514-4e62-a508-2012f38a3e15", 4, []mana.Color{mana.C, mana.U}, nil)
	engine.resolveTopOfStack(g, &TurnLog{})
	attacker := suspendFamilyPermanent(t, g, cardID)
	groundBlocker := addCombatPermanent(g, game.Player2, groundDef)
	suspendFamilyDeclareAttacks(t, g, engine, game.Player1, game.Player2, attacker)
	if suspendFamilyDeclareBlocks(g, engine, game.Player2, game.BlockDeclaration{Blocker: groundBlocker.ObjectID, Blocking: attacker.ObjectID}) {
		t.Fatal("a ground creature blocked Errant Ephemeron")
	}
	flyingBlocker := addCombatCreaturePermanentWithPower(g, game.Player2, 2, game.Flying)
	if !suspendFamilyDeclareBlocks(g, engine, game.Player2, game.BlockDeclaration{Blocker: flyingBlocker.ObjectID, Blocking: attacker.ObjectID}) {
		t.Fatal("a flying creature could not block Errant Ephemeron")
	}
}

type suspendFamilyColorAgent struct {
	optionIndex int
	requests    []game.ChoiceRequest
}

type suspendFamilySearchAgent struct {
	requests []game.ChoiceRequest
}

func (*suspendFamilySearchAgent) ChooseAction(PlayerObservation, []action.Action) action.Action {
	return action.Pass()
}

func (a *suspendFamilySearchAgent) ChooseChoice(_ PlayerObservation, request game.ChoiceRequest) []int {
	if request.Kind == game.ChoiceSearch {
		a.requests = append(a.requests, request)
		if len(request.Options) > 0 {
			return []int{request.Options[0].Index}
		}
	}
	return request.DefaultSelection
}

func (*suspendFamilyColorAgent) ChooseAction(PlayerObservation, []action.Action) action.Action {
	return action.Pass()
}

func (a *suspendFamilyColorAgent) ChooseChoice(_ PlayerObservation, request game.ChoiceRequest) []int {
	a.requests = append(a.requests, request)
	if len(request.Options) == 5 && a.optionIndex >= 0 && a.optionIndex < len(request.Options) {
		return []int{request.Options[a.optionIndex].Index}
	}
	return request.DefaultSelection
}

// Source review: Oracle cc319819-68f9-4c1e-8d60-25a5e3e50669, printing dcd05b01-dc73-4dd2-970a-32ec6e153c86; pinned text and full HTTP 200 response at https://api.scryfall.com/cards/dcd05b01-dc73-4dd2-970a-32ec6e153c86/rulings checked 2026-11-10 (15 rulings).
// CR 2026-11-06 (official TXT URL and SHA-256 above), read 605.1a, 605.3, and 702.62a-d; the artifact suspends for {0}, then its tap mana ability chooses one of five colors.
func TestMoxTantaliteZeroCostSuspendAndAnyColorMana(t *testing.T) {
	g, engine, cardID, _ := suspendFamilyCast(t, "Mox Tantalite", "cc319819-68f9-4c1e-8d60-25a5e3e50669", 3, nil, nil)
	engine.resolveTopOfStack(g, &TurnLog{})
	mox := suspendFamilyPermanent(t, g, cardID)
	agent := &suspendFamilyColorAgent{optionIndex: 4}
	if !engine.applyActionWithChoices(g, game.Player1, action.ActivateAbility(mox.ObjectID, 0, nil, 0), suspendAgents(agent), &TurnLog{}) {
		t.Fatal("Mox Tantalite could not activate its tap mana ability")
	}
	if len(agent.requests) != 1 || len(agent.requests[0].Options) != 5 {
		t.Fatalf("Mox Tantalite mana choices = %+v, want five colors", agent.requests)
	}
	if got := g.Players[game.Player1].ManaPool.Amount(mana.G); got != 1 {
		t.Fatalf("Mox Tantalite produced %d green mana, want 1", got)
	}
	if !mox.Tapped {
		t.Fatal("Mox Tantalite did not pay its tap cost")
	}
}

// Source review: Oracle dd491ac7-3c6c-4a47-94af-071d1edc5eb2, printing c3ab190b-c96f-4b00-8e97-33a11bb2ed08; pinned text and https://api.scryfall.com/cards/c3ab190b-c96f-4b00-8e97-33a11bb2ed08/rulings checked 2026-11-10 (17 rulings, including 2021-03-19).
// CR 2026-11-06 (official TXT URL and SHA-256 above), read 509.1, 702.28, and 702.62a-d; only creatures with shadow may block this suspended creature.
func TestInfiltratorIlKorSuspendShadow(t *testing.T) {
	groundDef := suspendFamilyCreatureDef("Ground Blocker", 2, 2, []types.Card{types.Creature}, []color.Color{color.Green})
	suspendFamilyBlockControl(t,
		suspendFamilyCreatureDef("Vanilla Attacker", 3, 3, []types.Card{types.Creature}, []color.Color{color.Blue}),
		groundDef,
	)

	g, engine, cardID, _ := suspendFamilyCast(t, "Infiltrator il-Kor", "dd491ac7-3c6c-4a47-94af-071d1edc5eb2", 2, []mana.Color{mana.C, mana.U}, nil)
	engine.resolveTopOfStack(g, &TurnLog{})
	attacker := suspendFamilyPermanent(t, g, cardID)
	groundBlocker := addCombatPermanent(g, game.Player2, groundDef)
	suspendFamilyDeclareAttacks(t, g, engine, game.Player1, game.Player2, attacker)
	if suspendFamilyDeclareBlocks(g, engine, game.Player2, game.BlockDeclaration{Blocker: groundBlocker.ObjectID, Blocking: attacker.ObjectID}) {
		t.Fatal("a creature without shadow blocked Infiltrator il-Kor")
	}
	shadowBlocker := addCombatCreaturePermanentWithPower(g, game.Player2, 2, game.Shadow)
	if !suspendFamilyDeclareBlocks(g, engine, game.Player2, game.BlockDeclaration{Blocker: shadowBlocker.ObjectID, Blocking: attacker.ObjectID}) {
		t.Fatal("a creature with shadow could not block Infiltrator il-Kor")
	}
}

// Source review: Oracle eb933723-b879-44c8-bbc8-d8c84ce5ab12, printing 54c32735-e8e7-46a1-84ea-92806f459e84; pinned text and https://api.scryfall.com/cards/54c32735-e8e7-46a1-84ea-92806f459e84/rulings checked 2026-11-10 (18 rulings, including 2013 and 2021-03-19).
// CR 2026-11-06 (official TXT URL and SHA-256 above), read 602.2, 615.1, 702.20, and 702.62a-d; vigilance leaves Ith untapped attacking, and its tap ability untaps its target while preventing combat damage both to and by it for the turn.
func TestIthHighArcanistVigilanceUntapsAndPreventsCombatDamage(t *testing.T) {
	g, engine, cardID, _ := suspendFamilyCast(t, "Ith, High Arcanist", "eb933723-b879-44c8-bbc8-d8c84ce5ab12", 4, []mana.Color{mana.W, mana.U}, nil)
	engine.resolveTopOfStack(g, &TurnLog{})
	ith := suspendFamilyPermanent(t, g, cardID)
	attacker := addCombatCreaturePermanentWithPower(g, game.Player1, 2)
	suspendFamilyDeclareAttacks(t, g, engine, game.Player1, game.Player2, ith, attacker)
	if ith.Tapped {
		t.Fatal("Ith was tapped attacking despite vigilance")
	}
	if !attacker.Tapped {
		t.Fatal("ordinary attacker was not tapped when declared")
	}

	g.Turn.PriorityPlayer = game.Player1
	if !engine.applyAction(g, game.Player1, action.ActivateAbility(
		ith.ObjectID, 0, []game.Target{game.PermanentTarget(attacker.ObjectID)}, 0,
	)) {
		t.Fatal("Ith could not activate its tap ability targeting the attacking creature")
	}
	engine.resolveTopOfStack(g, &TurnLog{})
	if !ith.Tapped || attacker.Tapped {
		t.Fatalf("Ith tapped=%t, target tapped=%t after resolution; want tapped source and untapped target", ith.Tapped, attacker.Tapped)
	}
	blocker := addCombatCreaturePermanentWithPower(g, game.Player2, 4)
	if !suspendFamilyDeclareBlocks(g, engine, game.Player2, game.BlockDeclaration{Blocker: blocker.ObjectID, Blocking: attacker.ObjectID}) {
		t.Fatal("untapped attacker could not be blocked")
	}
	g.Turn.Step = game.StepCombatDamage
	engine.resolveCombatDamage(g, &TurnLog{})
	engine.applyStateBasedActions(g)
	if _, alive := permanentByObjectID(g, attacker.ObjectID); !alive {
		t.Fatal("target attacker died despite Ith preventing damage dealt to it")
	}
	if blocker.MarkedDamage != 0 {
		t.Fatalf("blocker marked damage = %d, want Ith to prevent damage dealt by the target", blocker.MarkedDamage)
	}
	if got := g.Players[game.Player2].Life; got != 37 {
		t.Fatalf("defending player life = %d, want 37 from Ith's unblocked 3 damage", got)
	}
}

// Source review: Oracle f9deeea1-bf80-4832-aac3-373fc1f4b41d, printing 97dbd212-f1e8-429a-bf00-b2ea966d880e; pinned text and https://api.scryfall.com/cards/97dbd212-f1e8-429a-bf00-b2ea966d880e/rulings checked 2026-11-10 (14 rulings).
// CR 2026-11-06 (official TXT URL and SHA-256 above), read 605.1a, 605.3, and 702.62a-d; the green suspend cost yields a creature with haste, letting it tap immediately to choose any of the five mana colors.
func TestRiftSowerSuspendAnyColorMana(t *testing.T) {
	g, engine, cardID, _ := suspendFamilyCast(t, "Rift Sower", "f9deeea1-bf80-4832-aac3-373fc1f4b41d", 2, []mana.Color{mana.G}, nil)
	engine.resolveTopOfStack(g, &TurnLog{})
	riftSower := suspendFamilyPermanent(t, g, cardID)
	agent := &suspendFamilyColorAgent{optionIndex: 1}
	if !engine.applyActionWithChoices(g, game.Player1, action.ActivateAbility(riftSower.ObjectID, 0, nil, 0), suspendAgents(agent), &TurnLog{}) {
		t.Fatal("Rift Sower could not activate its tap mana ability immediately after its suspend cast")
	}
	if len(agent.requests) != 1 || len(agent.requests[0].Options) != 5 {
		t.Fatalf("Rift Sower mana choices = %+v, want five colors", agent.requests)
	}
	if got := g.Players[game.Player1].ManaPool.Amount(mana.U); got != 1 {
		t.Fatalf("Rift Sower produced %d blue mana, want 1", got)
	}
	if !riftSower.Tapped {
		t.Fatal("Rift Sower did not pay its tap cost")
	}
}

// requireExileTimeCounterEvent asserts the latest counter event of kind
// identifies the exiled card cardID, with no permanent, as changing amount time
// counters from previous, and returns it.
func requireExileTimeCounterEvent(t *testing.T, g *game.Game, kind game.EventKind, cardID id.ID, previous, amount int) game.Event {
	t.Helper()
	for i := len(g.Events) - 1; i >= 0; i-- {
		event := g.Events[i]
		if event.Kind != kind || event.CardID != cardID {
			continue
		}
		card, _ := g.GetCardInstance(cardID)
		if event.PermanentID != 0 || event.CounterKind != counter.Time ||
			event.PreviousCounterAmount != previous || event.Amount != amount || event.Player != card.Owner {
			t.Fatalf("%v event = %+v, want %d time counters changed from %d on exiled card %v", kind, event, amount, previous, cardID)
		}
		return event
	}
	t.Fatalf("no %v event for exiled card %v", kind, cardID)
	return game.Event{}
}
