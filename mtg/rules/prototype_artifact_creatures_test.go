package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

func castPrototypeArtifact(t *testing.T, g *game.Game, engine *Engine, name string, face game.FaceIndex) *game.Permanent {
	t.Helper()
	cardID := addCardToHand(g, game.Player1, registryCard(t, name))
	if !engine.applyAction(g, game.Player1, action.CastSpellFace(cardID, face, nil, 0, nil)) {
		t.Fatalf("%s cast using face %d was rejected", name, face)
	}
	spell, ok := g.Stack.Peek()
	if !ok || spell.SourceID != cardID || spell.Face != face {
		t.Fatalf("stack object = %+v, want %s face %d", spell, name, face)
	}
	resolveStackWithTriggers(engine, g, simPassPolicies())
	permanent := permanentForCard(g, cardID)
	if permanent == nil || permanent.Face != face {
		t.Fatalf("%s did not enter as face %d: %+v", name, face, permanent)
	}
	return permanent
}

func addPrototypeTestMana(g *game.Game, generic int, colored ...mana.Color) {
	pool := make([]mana.Color, generic, generic+len(colored))
	for i := range pool {
		pool[i] = mana.C
	}
	pool = append(pool, colored...)
	fuseAddMana(g, game.Player1, pool...)
}

func assertArtifactPrototypeStats(t *testing.T, g *game.Game, permanent *game.Permanent, face game.FaceIndex, power, toughness int) {
	t.Helper()
	if permanent.Face != face {
		t.Fatalf("permanent face = %d, want %d", permanent.Face, face)
	}
	if got := effectivePower(g, permanent); got != power {
		t.Fatalf("power = %d, want %d", got, power)
	}
	if got, _ := effectiveToughness(g, permanent); got != toughness {
		t.Fatalf("toughness = %d, want %d", got, toughness)
	}
	def, ok := permanentFaceDef(g, permanent)
	if !ok || !def.HasType(types.Artifact) || !def.HasType(types.Creature) {
		t.Fatalf("prototype did not retain artifact creature types: %+v", def)
	}
}

// Source review: pinned Oracle ID 02457c7f-eced-4bf3-9b61-309b4ceab2d5,
// printing 9d927dfd-ae66-4235-a61b-39c85d0c1222; Scryfall rulings
// https://api.scryfall.com/cards/9d927dfd-ae66-4235-a61b-39c85d0c1222/rulings
// (7, 2022-10-14) confirm prototype characteristics are used only for a
// prototyped spell/permanent and its printed abilities remain. Read the
// official Comprehensive Rules effective 2026-09-25 at
// .cardwork/prototype-source/MagicCompRules-20260925.txt (SHA-256
// 8d860e451f20f38865b725b42d82feb714c725373dd8f3b32b8652b3eeb070ca):
// 105.2, 601.2b/f, 702.2, 702.160a, 719.1-5, and 510.1. These specify the
// prototype's color/P/T/cost, retained deathtouch, and combat damage behavior.
func TestGoringWarplowScenario(t *testing.T) {
	g, engine := prototypeScenario(t)
	addPrototypeTestMana(g, 7, mana.B)
	prototyped := castPrototypeArtifact(t, g, engine, "Goring Warplow", game.FacePrototype)
	normal := castPrototypeArtifact(t, g, engine, "Goring Warplow", game.FaceFront)
	assertArtifactPrototypeStats(t, g, prototyped, game.FacePrototype, 1, 1)
	assertArtifactPrototypeStats(t, g, normal, game.FaceFront, 5, 4)
	if !hasKeyword(g, prototyped, game.Deathtouch) || !hasKeyword(g, normal, game.Deathtouch) {
		t.Fatal("Goring Warplow lost deathtouch in a cast mode")
	}

	prototyped.SummoningSick, normal.SummoningSick = false, false
	blocker := addCombatPermanent(g, game.Player2, &game.CardDef{CardFace: game.CardFace{
		Name: "2/2 blocker", Types: []types.Card{types.Creature},
		Power: opt.Val(game.PT{Value: 2}), Toughness: opt.Val(game.PT{Value: 2}),
	}})
	g.Turn.ActivePlayer, g.Turn.Phase, g.Turn.Step = game.Player1, game.PhaseCombat, game.StepDeclareAttackers
	g.Combat = &game.CombatState{}
	policies := simPassPolicies()
	policies[game.Player2] = blockingPolicy{blocker: blocker.ObjectID, attacker: prototyped.ObjectID}
	resolved, ok := engine.Simulator().ResolveCombatWithAttackers(g, game.Player1, action.DeclareAttackersAction{
		Attackers: []game.AttackDeclaration{
			{Attacker: prototyped.ObjectID, Target: game.AttackTarget{Player: game.Player2}},
			{Attacker: normal.ObjectID, Target: game.AttackTarget{Player: game.Player2}},
		},
	}, policies)
	if !ok {
		t.Fatal("Goring Warplows could not attack")
	}
	if permanentStillPresent(resolved, blocker.ObjectID) || permanentStillPresent(resolved, prototyped.ObjectID) {
		t.Fatal("the prototyped 1/1 did not trade with a 2/2 through deathtouch")
	}
	if !permanentStillPresent(resolved, normal.ObjectID) {
		t.Fatal("the unblocked normal Goring Warplow did not survive")
	}
	if got, want := resolved.Players[game.Player2].Life, g.Players[game.Player2].Life-5; got != want {
		t.Fatalf("defender life = %d, want %d from the normal 5/4", got, want)
	}
}

// Source review: pinned Oracle ID 0c460bdd-460b-4312-9ed2-32f3809a7d36,
// printing 74dfabfd-e13b-4512-a733-abe514be6404; Scryfall rulings
// https://api.scryfall.com/cards/74dfabfd-e13b-4512-a733-abe514be6404/rulings
// (7, 2022-10-14) cover the normal/prototype characteristics and retained
// abilities. Read CR 105.2, 603.2, 601.2b/f, 702.4, 702.160a, 719.1-5,
// 121.2, and 510.1 in the verified 2026-09-25 document. The ETB trigger draws
// one card per entry; double strike assigns combat damage in both damage steps.
func TestCombatThresherScenario(t *testing.T) {
	g, engine := prototypeScenario(t)
	drawn1 := addCardToLibrary(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "First draw"}})
	drawn2 := addCardToLibrary(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Second draw"}})
	addPrototypeTestMana(g, 9, mana.W)
	prototyped := castPrototypeArtifact(t, g, engine, "Combat Thresher", game.FacePrototype)
	normal := castPrototypeArtifact(t, g, engine, "Combat Thresher", game.FaceFront)
	assertArtifactPrototypeStats(t, g, prototyped, game.FacePrototype, 1, 1)
	assertArtifactPrototypeStats(t, g, normal, game.FaceFront, 3, 3)
	if !g.Players[game.Player1].Hand.Contains(drawn1) || !g.Players[game.Player1].Hand.Contains(drawn2) {
		t.Fatal("each Combat Thresher entry should draw one card")
	}
	if !hasKeyword(g, prototyped, game.DoubleStrike) || !hasKeyword(g, normal, game.DoubleStrike) {
		t.Fatal("Combat Thresher lost double strike in a cast mode")
	}

	prototyped.SummoningSick, normal.SummoningSick = false, false
	g.Turn.ActivePlayer, g.Turn.Phase, g.Turn.Step = game.Player1, game.PhaseCombat, game.StepDeclareAttackers
	g.Combat = &game.CombatState{}
	life := g.Players[game.Player2].Life
	resolved, ok := engine.Simulator().ResolveCombatWithAttackers(g, game.Player1, action.DeclareAttackersAction{
		Attackers: []game.AttackDeclaration{
			{Attacker: prototyped.ObjectID, Target: game.AttackTarget{Player: game.Player2}},
			{Attacker: normal.ObjectID, Target: game.AttackTarget{Player: game.Player2}},
		},
	}, simPassPolicies())
	if !ok || resolved.Players[game.Player2].Life != life-8 {
		t.Fatalf("double-strike combat resolved=%v, defender life=%d; want %d", ok, resolved.Players[game.Player2].Life, life-8)
	}
}

// Source review: pinned Oracle ID 19f344ee-031d-4346-9553-c5d32b02df43,
// printing e4e5529a-848b-486c-97fb-1def35df7837; Scryfall rulings
// https://api.scryfall.com/cards/e4e5529a-848b-486c-97fb-1def35df7837/rulings
// (7, 2022-10-14) describe prototype's alternate stack/battlefield
// characteristics and the card's normal identity elsewhere. Read CR 105.2,
// 302.6, 508.1a, 702.10, 702.160a, and 719.1-5 in the verified 2026-09-25
// rules. Haste removes the summoning-sickness attack restriction in 508.1a.
func TestBlitzAutomatonScenario(t *testing.T) {
	g, engine := prototypeScenario(t)
	addPrototypeTestMana(g, 9, mana.R)
	prototyped := castPrototypeArtifact(t, g, engine, "Blitz Automaton", game.FacePrototype)
	normal := castPrototypeArtifact(t, g, engine, "Blitz Automaton", game.FaceFront)
	assertArtifactPrototypeStats(t, g, prototyped, game.FacePrototype, 3, 2)
	assertArtifactPrototypeStats(t, g, normal, game.FaceFront, 6, 4)
	if !hasKeyword(g, prototyped, game.Haste) || !hasKeyword(g, normal, game.Haste) {
		t.Fatal("Blitz Automaton lost haste in a cast mode")
	}

	g.Turn.ActivePlayer, g.Turn.Phase, g.Turn.Step = game.Player1, game.PhaseCombat, game.StepDeclareAttackers
	g.Combat = &game.CombatState{}
	life := g.Players[game.Player2].Life
	resolved, ok := engine.Simulator().ResolveCombatWithAttackers(g, game.Player1, action.DeclareAttackersAction{
		Attackers: []game.AttackDeclaration{
			{Attacker: prototyped.ObjectID, Target: game.AttackTarget{Player: game.Player2}},
			{Attacker: normal.ObjectID, Target: game.AttackTarget{Player: game.Player2}},
		},
	}, simPassPolicies())
	if !ok || resolved.Players[game.Player2].Life != life-9 {
		t.Fatalf("same-turn hasty attack resolved=%v, defender life=%d; want %d", ok, resolved.Players[game.Player2].Life, life-9)
	}
}

// Source review: pinned Oracle ID 225f53ef-da42-47a0-b773-d2aad3338c72,
// printing 50cc9b90-792e-4cf7-ab8c-cb616d94092a; Scryfall rulings
// https://api.scryfall.com/cards/50cc9b90-792e-4cf7-ab8c-cb616d94092a/rulings
// (7, 2022-10-14) confirm a prototyped permanent keeps its printed abilities.
// Read CR 105.2, 302.6, 605.1a/605.3b, 613.4c, 702.160a, and 719.1-5 in the
// verified 2026-09-25 rules. The tap ability is a mana ability, and its amount
// uses the source creature's current power in either cast mode.
func TestCradleClearcutterScenario(t *testing.T) {
	g, engine := prototypeScenario(t)
	addPrototypeTestMana(g, 8, mana.G)
	prototyped := castPrototypeArtifact(t, g, engine, "Cradle Clearcutter", game.FacePrototype)
	normal := castPrototypeArtifact(t, g, engine, "Cradle Clearcutter", game.FaceFront)
	assertArtifactPrototypeStats(t, g, prototyped, game.FacePrototype, 1, 3)
	assertArtifactPrototypeStats(t, g, normal, game.FaceFront, 3, 6)
	prototyped.SummoningSick, normal.SummoningSick = false, false
	setMainPhasePriority(g, game.Player1)

	if !engine.applyAction(g, game.Player1, action.ActivateAbility(prototyped.ObjectID, 0, nil, 0)) {
		t.Fatal("could not activate prototype's power-based mana ability")
	}
	if !engine.applyAction(g, game.Player1, action.ActivateAbility(normal.ObjectID, 0, nil, 0)) {
		t.Fatal("could not activate normal creature's power-based mana ability")
	}
	if got := g.Players[game.Player1].ManaPool.Amount(mana.G); got != 4 {
		t.Fatalf("green mana produced = %d, want 1 + 3 from the two powers", got)
	}
	if !prototyped.Tapped || !normal.Tapped {
		t.Fatal("the tap-symbol mana abilities did not tap their sources")
	}
}

// Source review: pinned Oracle ID 61df991d-32e7-4b6b-b8e6-700ba915c3f2,
// printing 8348f73d-4226-4296-9d51-a2c497ee8270; Scryfall rulings
// https://api.scryfall.com/cards/8348f73d-4226-4296-9d51-a2c497ee8270/rulings
// (7, 2022-10-14) cover prototype's alternate characteristics and retained
// abilities. Read CR 105.2, 509.1, 510.1, 702.9, 702.17, 702.19,
// 702.160a, and 719.1-5 in the verified 2026-09-25 rules. Reach permits a
// Rust Goliath block against a flyer; trample assigns excess combat damage.
func TestRustGoliathScenario(t *testing.T) {
	g, engine := prototypeScenario(t)
	addPrototypeTestMana(g, 13, mana.G, mana.G)
	prototyped := castPrototypeArtifact(t, g, engine, "Rust Goliath", game.FacePrototype)
	normal := castPrototypeArtifact(t, g, engine, "Rust Goliath", game.FaceFront)
	assertArtifactPrototypeStats(t, g, prototyped, game.FacePrototype, 3, 5)
	assertArtifactPrototypeStats(t, g, normal, game.FaceFront, 10, 10)
	if !hasKeyword(g, prototyped, game.Reach) || !hasKeyword(g, normal, game.Reach) ||
		!hasKeyword(g, prototyped, game.Trample) || !hasKeyword(g, normal, game.Trample) {
		t.Fatal("Rust Goliath lost reach or trample in a cast mode")
	}

	prototyped.SummoningSick = false
	blocker := addCombatPermanent(g, game.Player2, &game.CardDef{CardFace: game.CardFace{
		Name: "1/1 ground blocker", Types: []types.Card{types.Creature},
		Power: opt.Val(game.PT{Value: 1}), Toughness: opt.Val(game.PT{Value: 1}),
	}})
	g.Turn.ActivePlayer, g.Turn.Phase, g.Turn.Step = game.Player1, game.PhaseCombat, game.StepDeclareAttackers
	g.Combat = &game.CombatState{}
	policies := simPassPolicies()
	policies[game.Player2] = blockingPolicy{blocker: blocker.ObjectID, attacker: prototyped.ObjectID}
	life := g.Players[game.Player2].Life
	trampleResult, ok := engine.Simulator().ResolveCombatWithAttackers(g, game.Player1, action.DeclareAttackersAction{
		Attackers: []game.AttackDeclaration{{Attacker: prototyped.ObjectID, Target: game.AttackTarget{Player: game.Player2}}},
	}, policies)
	if !ok || trampleResult.Players[game.Player2].Life != life-2 || permanentStillPresent(trampleResult, blocker.ObjectID) {
		t.Fatalf("prototype trample result resolved=%v, defender life=%d, blocker remains=%v; want 2 damage and a dead blocker", ok, trampleResult.Players[game.Player2].Life, permanentStillPresent(trampleResult, blocker.ObjectID))
	}

	flyer := addCombatPermanent(g, game.Player2, &game.CardDef{CardFace: game.CardFace{
		Name: "1/1 flyer", Types: []types.Card{types.Creature},
		Power: opt.Val(game.PT{Value: 1}), Toughness: opt.Val(game.PT{Value: 1}),
		StaticAbilities: []game.StaticAbility{game.FlyingStaticBody},
	}})
	g.Turn.ActivePlayer, g.Turn.Phase, g.Turn.Step = game.Player2, game.PhaseCombat, game.StepDeclareAttackers
	g.Combat = &game.CombatState{}
	policies = simPassPolicies()
	policies[game.Player1] = blockingPolicy{blocker: normal.ObjectID, attacker: flyer.ObjectID}
	reachResult, ok := engine.Simulator().ResolveCombatWithAttackers(g, game.Player2, action.DeclareAttackersAction{
		Attackers: []game.AttackDeclaration{{Attacker: flyer.ObjectID, Target: game.AttackTarget{Player: game.Player1}}},
	}, policies)
	if !ok || permanentStillPresent(reachResult, flyer.ObjectID) || !permanentStillPresent(reachResult, normal.ObjectID) {
		t.Fatalf("normal Rust Goliath did not legally block and defeat the flyer: resolved=%v", ok)
	}
}

// Source review: pinned Oracle ID ccc7de36-aafc-4d87-80f7-fc2effc976eb,
// printing 981fdd21-a650-4360-9ea5-550399d1da91; Scryfall rulings
// https://api.scryfall.com/cards/981fdd21-a650-4360-9ea5-550399d1da91/rulings
// (8, latest 2024-06-07) include that affinity still reduces a prototyped
// spell's cost. Read CR 105.2, 601.2b/f, 702.41a, 702.160a, and 719.1-5 in
// the verified 2026-09-25 rules. The scenario pays only {R} for the prototype
// with three artifacts, then only {3} for the normal spell after the first
// Frogmyr adds a fourth artifact.
func TestFrogmyrEnforcerScenario(t *testing.T) {
	g, engine := prototypeScenario(t)
	for i := 0; i < 3; i++ {
		addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
			Name: "Artifact", Types: []types.Card{types.Artifact},
		}})
	}
	fuseAddMana(g, game.Player1, mana.R)
	prototyped := castPrototypeArtifact(t, g, engine, "Frogmyr Enforcer", game.FacePrototype)
	assertArtifactPrototypeStats(t, g, prototyped, game.FacePrototype, 2, 2)

	fuseAddMana(g, game.Player1, mana.C, mana.C, mana.C)
	normal := castPrototypeArtifact(t, g, engine, "Frogmyr Enforcer", game.FaceFront)
	assertArtifactPrototypeStats(t, g, normal, game.FaceFront, 4, 4)
}

// Source review: pinned Oracle ID dc8cca38-df2a-4cef-973c-a482505be66f,
// printing f7761ed0-a784-4dfb-ab6c-0f4b9a411cf3; Scryfall rulings
// https://api.scryfall.com/cards/f7761ed0-a784-4dfb-ab6c-0f4b9a411cf3/rulings
// (7, 2022-10-14) confirm the prototype face retains flying and its activated
// ability. Read CR 105.2, 302.6, 508.1a, 602.2, 611.2a, 613.4c, 702.9,
// 702.160a, and 719.1-5 in the verified 2026-09-25 rules. Both modes are
// pumped through the Engine; a ground blocker cannot block either flyer.
func TestFallajiDragonEngineScenario(t *testing.T) {
	g, engine := prototypeScenario(t)
	addPrototypeTestMana(g, 14, mana.R)
	prototyped := castPrototypeArtifact(t, g, engine, "Fallaji Dragon Engine", game.FacePrototype)
	normal := castPrototypeArtifact(t, g, engine, "Fallaji Dragon Engine", game.FaceFront)
	assertArtifactPrototypeStats(t, g, prototyped, game.FacePrototype, 1, 3)
	assertArtifactPrototypeStats(t, g, normal, game.FaceFront, 5, 5)
	if !hasKeyword(g, prototyped, game.Flying) || !hasKeyword(g, normal, game.Flying) {
		t.Fatal("Fallaji Dragon Engine lost flying in a cast mode")
	}
	fuseAddMana(g, game.Player1, mana.C, mana.C, mana.C, mana.C)
	for _, permanent := range []*game.Permanent{prototyped, normal} {
		if !engine.applyAction(g, game.Player1, action.ActivateAbility(permanent.ObjectID, 0, nil, 0)) {
			t.Fatal("could not activate Fallaji Dragon Engine's pump ability")
		}
		resolveStackWithTriggers(engine, g, simPassPolicies())
	}
	assertArtifactPrototypeStats(t, g, prototyped, game.FacePrototype, 2, 3)
	assertArtifactPrototypeStats(t, g, normal, game.FaceFront, 6, 5)

	prototyped.SummoningSick, normal.SummoningSick = false, false
	groundBlocker := addCombatPermanent(g, game.Player2, &game.CardDef{CardFace: game.CardFace{
		Name: "2/2 ground blocker", Types: []types.Card{types.Creature},
		Power: opt.Val(game.PT{Value: 2}), Toughness: opt.Val(game.PT{Value: 2}),
	}})
	g.Turn.ActivePlayer, g.Turn.Phase, g.Turn.Step = game.Player1, game.PhaseCombat, game.StepDeclareAttackers
	g.Combat = &game.CombatState{}
	policies := simPassPolicies()
	policies[game.Player2] = blockingPolicy{blocker: groundBlocker.ObjectID, attacker: prototyped.ObjectID}
	life := g.Players[game.Player2].Life
	resolved, ok := engine.Simulator().ResolveCombatWithAttackers(g, game.Player1, action.DeclareAttackersAction{
		Attackers: []game.AttackDeclaration{
			{Attacker: prototyped.ObjectID, Target: game.AttackTarget{Player: game.Player2}},
			{Attacker: normal.ObjectID, Target: game.AttackTarget{Player: game.Player2}},
		},
	}, policies)
	if !ok || resolved.Players[game.Player2].Life != life-8 {
		t.Fatalf("flying combat resolved=%v, defender life=%d; want %d through a ground block attempt", ok, resolved.Players[game.Player2].Life, life-8)
	}
}

// Source review: pinned Oracle ID f376f4c6-4526-4ae3-a84a-e8ea50fe78b6,
// printing 34e09173-a92d-4cc3-9d02-bf18ab0850ec; the task claim's printing
// UUID 34e09173-a92d-4cc7-80f7-fc2effc976eb returns 404, while the pinned
// corpus/name lookup yields the valid printing and rulings URI below. Scryfall
// rulings https://api.scryfall.com/cards/34e09173-a92d-4cc3-9d02-bf18ab0850ec/rulings
// (7, 2022-10-14) confirm prototype's normal characteristics outside the
// stack/battlefield and retained abilities. Read CR 105.2, 302.6, 502.3,
// 602.2, 702.160a, and 719.1-5 in the verified 2026-09-25 rules. The untap
// step lock applies to both modes; {3} explicitly untaps either creature.
func TestDepthChargeColossusScenario(t *testing.T) {
	g, engine := prototypeScenario(t)
	addCardToLibrary(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Draw for next turn"}})
	addPrototypeTestMana(g, 19, mana.U, mana.U)
	prototyped := castPrototypeArtifact(t, g, engine, "Depth Charge Colossus", game.FacePrototype)
	normal := castPrototypeArtifact(t, g, engine, "Depth Charge Colossus", game.FaceFront)
	assertArtifactPrototypeStats(t, g, prototyped, game.FacePrototype, 6, 6)
	assertArtifactPrototypeStats(t, g, normal, game.FaceFront, 9, 9)
	prototyped.Tapped, normal.Tapped = true, true
	g.Turn.ActivePlayer = game.Player1
	engine.runBeginningPhase(g, [game.NumPlayers]PlayerAgent{}, &TurnLog{})
	if !prototyped.Tapped || !normal.Tapped {
		t.Fatal("Depth Charge Colossus untapped during its controller's untap step")
	}

	setMainPhasePriority(g, game.Player1)
	addPrototypeTestMana(g, 6)
	for _, permanent := range []*game.Permanent{prototyped, normal} {
		if !engine.applyAction(g, game.Player1, action.ActivateAbility(permanent.ObjectID, 0, nil, 0)) {
			t.Fatal("could not activate Depth Charge Colossus's {3} untap ability")
		}
		resolveStackWithTriggers(engine, g, simPassPolicies())
	}
	if prototyped.Tapped || normal.Tapped {
		t.Fatal("the {3} ability did not untap both cast modes")
	}
}
