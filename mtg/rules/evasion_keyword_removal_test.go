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

// Source review: pinned Oracle corpus 6da9ced0…9a1b4:
// Hammerheim (Oracle ID c7476beb-7923-4994-8476-bc69187ecb72; printing
// 773acde9-504c-4065-95e0-f96428f05a7d), "{T}: Add {R}. {T}: Target creature
// loses all landwalk abilities until end of turn." Its Scryfall rulings URI
// https://api.scryfall.com/cards/773acde9-504c-4065-95e0-f96428f05a7d/rulings
// returned two rulings: 2009-10-01 says landwalk removal is useful through
// declare attackers but too late after declare blockers begins; 2004-10-04 says
// a creature without landwalk may be targeted but is unaffected.
// Crash Landing (Oracle ID 5d36e5ee-1e23-48fe-ad13-081cc153342b; printing
// f8683546-85b7-4658-91be-91ef65fb3fc3), mana cost {2}{G}, "Target creature with flying
// loses flying until end of turn. Crash Landing deals damage to that creature
// equal to the number of Forests you control." Its Scryfall rulings URI
// https://api.scryfall.com/cards/f8683546-85b7-4658-91be-91ef65fb3fc3/rulings
// returned the 2006-02-01 ruling that it cannot target a creature without flying
// just to deal damage to it.
// Kinscaer Harpoonist (Oracle ID 8ee98d6a-6a7e-411d-818c-816d84cfb7cd; printing
// 13cfde34-f76b-47ab-b83e-1c188d61a613), "Flying. Whenever this creature
// attacks, you may have target creature lose flying until end of turn." Its
// Scryfall rulings URI
// https://api.scryfall.com/cards/13cfde34-f76b-47ab-b83e-1c188d61a613/rulings
// returned the 2008-05-01 ruling that a creature without flying may be targeted.
// Official CR, effective 2026-11-06:
// https://media.wizards.com/2026/downloads/MagicCompRules%2020261106.txt
// Read 113.10/113.10b (removal removes all instances), 613.1f (ability removal
// applies in layer 6), 611.2a (resolving effects last for their stated duration),
// 702.14a-e and 702.9a-c (landwalk/flying are evasion abilities), 601.2a-c and
// 602.2a-b (spell/ability announcement, target and cost), 508.1a-m and
// 603.2-603.3d (attack declaration triggers and putting triggers on the stack),
// 608.2b-c (target revalidation and instruction order), and 120.1/120.3e
// (creature damage is marked). These require removing every landwalk instance,
// validating Crash Landing's flying target, counting its controller's Forests
// as resolution performs damage after removal, and modeling Kinscaer's "may"
// as an optional resolving instruction with an any-creature target.

const (
	hammerheimTestOracleID         = "c7476beb-7923-4994-8476-bc69187ecb72"
	crashLandingTestOracleID       = "5d36e5ee-1e23-48fe-ad13-081cc153342b"
	kinscaerHarpoonistTestOracleID = "8ee98d6a-6a7e-411d-818c-816d84cfb7cd"
)

func evasionRemovalTestCard(t *testing.T, oracleID, name string) *game.CardDef {
	t.Helper()
	def, err := cards.NewDefaultRegistry().LookupOracleID(oracleID)
	if err != nil {
		t.Fatalf("default registry lookup for %s: %v", name, err)
	}
	if def.Name != name {
		t.Fatalf("default registry Oracle ID %s returned %q, want %q", oracleID, def.Name, name)
	}
	return def
}

func evasionTestCreature(name string, abilities ...game.StaticAbility) *game.CardDef {
	return &game.CardDef{CardFace: game.CardFace{
		Name:            name,
		Types:           []types.Card{types.Creature},
		Power:           opt.Val(game.PT{Value: 3}),
		Toughness:       opt.Val(game.PT{Value: 4}),
		StaticAbilities: abilities,
	}}
}

func TestHammerheimRemovesAllLandwalkAbilities(t *testing.T) {
	def := evasionRemovalTestCard(t, hammerheimTestOracleID, "Hammerheim")
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	hammerheim := addCombatPermanent(g, game.Player1, def)
	target := addCombatPermanent(g, game.Player2, evasionTestCreature("Landwalker",
		game.ForestwalkStaticBody, game.IslandwalkStaticBody, game.FlyingStaticBody,
	))

	if !hasKeyword(g, target, game.Landwalk) || !hasKeyword(g, target, game.Flying) {
		t.Fatal("test target must start with landwalk and flying")
	}
	setMainPhasePriority(g, game.Player1)
	if !engine.applyAction(g, game.Player1, action.ActivateAbility(
		hammerheim.ObjectID, 0, []game.Target{game.PermanentTarget(target.ObjectID)}, 0,
	)) {
		t.Fatal("Hammerheim's targeted tap ability was not accepted")
	}
	if !hammerheim.Tapped {
		t.Fatal("Hammerheim did not pay its tap cost")
	}
	engine.resolveTopOfStack(g, &TurnLog{})

	if hasKeyword(g, target, game.Landwalk) {
		t.Fatal("target retained landwalk after Hammerheim resolved")
	}
	if !hasKeyword(g, target, game.Flying) {
		t.Fatal("Hammerheim removed an unrelated flying ability")
	}
}

func TestHammerheimCanTargetCreatureWithoutLandwalk(t *testing.T) {
	def := evasionRemovalTestCard(t, hammerheimTestOracleID, "Hammerheim")
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	hammerheim := addCombatPermanent(g, game.Player1, def)
	target := addCombatPermanent(g, game.Player2, evasionTestCreature("Ordinary creature"))

	setMainPhasePriority(g, game.Player1)
	if !engine.applyAction(g, game.Player1, action.ActivateAbility(
		hammerheim.ObjectID, 0, []game.Target{game.PermanentTarget(target.ObjectID)}, 0,
	)) {
		t.Fatal("Hammerheim could not target a creature without landwalk")
	}
	engine.resolveTopOfStack(g, &TurnLog{})
	if hasKeyword(g, target, game.Landwalk) || hasKeyword(g, target, game.Flying) {
		t.Fatal("the target unexpectedly gained an evasion ability")
	}
}

func TestHammerheimAddsRedMana(t *testing.T) {
	def := evasionRemovalTestCard(t, hammerheimTestOracleID, "Hammerheim")
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	hammerheim := addCombatPermanent(g, game.Player1, def)
	setMainPhasePriority(g, game.Player1)

	if !engine.applyAction(g, game.Player1, action.ActivateAbility(
		hammerheim.ObjectID, def.ManaAbilityIndex(0), nil, 0,
	)) {
		t.Fatal("Hammerheim's red-mana ability was not accepted")
	}
	if !hammerheim.Tapped {
		t.Fatal("Hammerheim did not pay its tap cost for mana")
	}
	if got := g.Players[game.Player1].ManaPool.Amount(mana.R); got != 1 {
		t.Fatalf("red mana added = %d, want 1", got)
	}
}

func TestCrashLandingRemovesFlyingThenDealsDamageFromForests(t *testing.T) {
	def := evasionRemovalTestCard(t, crashLandingTestOracleID, "Crash Landing")
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	for range 2 {
		addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
			Name: "Forest", Types: []types.Card{types.Land}, Subtypes: []types.Sub{types.Forest},
		}})
	}
	addCombatPermanent(g, game.Player2, &game.CardDef{CardFace: game.CardFace{
		Name: "Opponent's Forest", Types: []types.Card{types.Land}, Subtypes: []types.Sub{types.Forest},
	}})
	target := addCombatPermanent(g, game.Player2, evasionTestCreature("Flying target", game.FlyingStaticBody))
	spellID := addCardToHand(g, game.Player1, def)
	g.Players[game.Player1].ManaPool.Add(mana.G, 1)
	g.Players[game.Player1].ManaPool.Add(mana.C, 2)
	setMainPhasePriority(g, game.Player1)

	if !engine.applyAction(g, game.Player1, action.CastSpell(
		spellID, []game.Target{game.PermanentTarget(target.ObjectID)}, 0, nil,
	)) {
		t.Fatal("Crash Landing could not be cast targeting a creature with flying")
	}
	engine.resolveTopOfStack(g, &TurnLog{})

	if hasKeyword(g, target, game.Flying) {
		t.Fatal("Crash Landing did not remove flying")
	}
	if target.MarkedDamage != 2 {
		t.Fatalf("damage marked = %d, want 2 from the caster's two Forests", target.MarkedDamage)
	}
}

type kinscaerMayPolicy struct {
	choiceOnlyAgent
	accept            bool
	target            game.Target
	targetChoiceCount int
	targetChosen      bool
	mayChoiceCount    int
	mayChoice         int
}

func (p *kinscaerMayPolicy) ChooseChoice(_ PlayerObservation, request game.ChoiceRequest) []int {
	switch request.Kind {
	case game.ChoiceMay:
		p.mayChoiceCount++
		if p.accept {
			p.mayChoice = 1
			return []int{1}
		}
		p.mayChoice = 0
		return []int{0}
	case game.ChoiceTarget:
		p.targetChoiceCount++
		for i, option := range request.Options {
			for _, target := range option.Targets {
				if target == p.target {
					p.targetChosen = true
					return []int{i}
				}
			}
		}
	}
	return request.DefaultSelection
}

func resolveKinscaerAttack(t *testing.T, targetFlying, accept bool) (*game.Game, *game.Permanent) {
	t.Helper()
	def := evasionRemovalTestCard(t, kinscaerHarpoonistTestOracleID, "Kinscaer Harpoonist")
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	attacker := addCombatPermanent(g, game.Player1, def)
	var targetAbilities []game.StaticAbility
	if targetFlying {
		targetAbilities = append(targetAbilities, game.FlyingStaticBody)
	}
	target := addCombatPermanent(g, game.Player2, evasionTestCreature("Target creature", targetAbilities...))
	if !hasKeyword(g, attacker, game.Flying) {
		t.Fatal("Kinscaer Harpoonist did not have flying")
	}

	g.Turn.ActivePlayer = game.Player1
	g.Turn.Phase = game.PhaseCombat
	g.Turn.Step = game.StepDeclareAttackers
	g.Combat = &game.CombatState{}
	declareAttack(t, g, engine, attacker, game.Player2)
	policy := &kinscaerMayPolicy{
		accept: accept,
		target: game.PermanentTarget(target.ObjectID),
	}
	agents := [game.NumPlayers]PlayerAgent{game.Player1: policy}
	if !engine.putTriggeredAbilitiesOnStackWithChoices(g, agents, &TurnLog{}) {
		t.Fatal("Kinscaer Harpoonist's attack trigger was not put on the stack")
	}
	if g.Stack.Size() != 1 {
		t.Fatalf("stack size after attack = %d, want 1", g.Stack.Size())
	}
	engine.resolveTopOfStackWithChoices(g, agents, &TurnLog{})
	if policy.targetChoiceCount != 1 || !policy.targetChosen {
		t.Fatalf("target choice count=%d chosen=%v, want the requested creature",
			policy.targetChoiceCount, policy.targetChosen)
	}
	wantMayChoice := 0
	if accept {
		wantMayChoice = 1
	}
	if policy.mayChoiceCount != 1 || policy.mayChoice != wantMayChoice {
		t.Fatalf("may choice count=%d choice=%d, want one choice of %d",
			policy.mayChoiceCount, policy.mayChoice, wantMayChoice)
	}
	return g, target
}

func TestKinscaerHarpoonistAttackMayRemoveFlying(t *testing.T) {
	g, target := resolveKinscaerAttack(t, true, true)
	if hasKeyword(g, target, game.Flying) {
		t.Fatal("accepted Kinscaer Harpoonist trigger did not remove flying")
	}
}

func TestKinscaerHarpoonistMayTargetCreatureWithoutFlying(t *testing.T) {
	g, target := resolveKinscaerAttack(t, false, true)
	if hasKeyword(g, target, game.Flying) {
		t.Fatal("a creature without flying unexpectedly gained flying")
	}
}

func TestKinscaerHarpoonistMayDeclineFlyingRemoval(t *testing.T) {
	g, target := resolveKinscaerAttack(t, true, false)
	if !hasKeyword(g, target, game.Flying) {
		t.Fatal("declining the optional effect removed flying")
	}
}

// Source review: pinned Oracle corpus SHA-256
// 6da9ced0e19bd384771ab8f14051cc45a8dc90302082b875fa6e039d2399a1b4,
// Burning Palm Efreet Oracle ID 55cbc0e8-eeec-433a-b70c-3501238b658b,
// printing revision a194488a-1d6f-4bc3-8f30-82e2c3c91389:
// "{1}{R}{R}: This creature deals 2 damage to target creature with flying and
// that creature loses flying until end of turn." Retrieved its Scryfall rulings
// URI https://api.scryfall.com/cards/a194488a-1d6f-4bc3-8f30-82e2c3c91389/rulings
// on 2026-11-10; it returned an empty list (raw response retained at
// .cardwork/source-review/burning-palm-rulings.json). Read the official Wizards
// Comprehensive Rules, effective 2026-11-06, at
// https://media.wizards.com/2026/downloads/MagicCompRules%2020261106.txt
// (verified local copy: .cardwork/source-review/MagicCompRules-20261106.txt):
// 113.10b, 120.3e, 514.2, 601.2c/f-h, 602.2a-b, 608.2b-c, 611.2a,
// 613.1f, 613.6, and 702.9a-c. These require paying the activated cost and
// choosing a flying creature, checking target legality on resolution, dealing
// damage before removing all instances of flying in layer 6 through cleanup.
func TestBurningPalmEfreetDealsDamageAndRemovesFlying(t *testing.T) {
	def := evasionRemovalTestCard(t, "55cbc0e8-eeec-433a-b70c-3501238b658b", "Burning Palm Efreet")
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	source := addCombatPermanent(g, game.Player1, def)
	target := addCombatPermanent(g, game.Player2, evasionTestCreature("Flying target", game.FlyingStaticBody))
	g.Players[game.Player1].ManaPool.Add(mana.C, 1)
	g.Players[game.Player1].ManaPool.Add(mana.R, 2)
	setMainPhasePriority(g, game.Player1)

	if !engine.applyAction(g, game.Player1, action.ActivateAbility(
		source.ObjectID, 0, []game.Target{game.PermanentTarget(target.ObjectID)}, 0,
	)) {
		t.Fatal("Burning Palm Efreet's {1}{R}{R} ability was not accepted")
	}
	if g.Players[game.Player1].ManaPool.Amount(mana.C) != 0 ||
		g.Players[game.Player1].ManaPool.Amount(mana.R) != 0 {
		t.Fatal("Burning Palm Efreet did not pay {1}{R}{R}")
	}
	engine.resolveTopOfStack(g, &TurnLog{})

	if target.MarkedDamage != 2 {
		t.Fatalf("damage marked = %d, want 2", target.MarkedDamage)
	}
	if hasKeyword(g, target, game.Flying) {
		t.Fatal("Burning Palm Efreet did not remove flying")
	}
}

// Source review: pinned Oracle corpus SHA-256
// 6da9ced0e19bd384771ab8f14051cc45a8dc90302082b875fa6e039d2399a1b4,
// Wind Shear Oracle ID fa683059-b9de-4063-ae71-597ed4e2df30, printing revision
// b8324f44-c7f5-41ee-bc8d-16822bd8942f: "Attacking creatures with flying get
// -2/-2 and lose flying until end of turn." Retrieved its Scryfall rulings URI
// https://api.scryfall.com/cards/b8324f44-c7f5-41ee-bc8d-16822bd8942f/rulings
// on 2026-11-10; its sole 2004-10-04 ruling confirms both effects last until
// end of turn (raw response retained at
// .cardwork/source-review/wind-shear-rulings.json). Read the official Wizards
// Comprehensive Rules, effective 2026-11-06, at
// https://media.wizards.com/2026/downloads/MagicCompRules%2020261106.txt
// (verified local copy: .cardwork/source-review/MagicCompRules-20261106.txt):
// 113.10b, 120.3e, 506.4a-b, 508.1a-k/m, 514.2, 601.2a-c/f-h,
// 608.2b-c, 611.2a/c, 613.1f-g, 613.4c, 613.6, and 702.9a-c. These require
// selecting the attacking-with-flying group at resolution, fixing the set
// affected across ability-removal and power/toughness layers, and ending both
// effects at cleanup; losing flying does not remove a declared attacker.
func TestWindShearAffectsOnlyAttackingCreaturesWithFlying(t *testing.T) {
	def := evasionRemovalTestCard(t, "fa683059-b9de-4063-ae71-597ed4e2df30", "Wind Shear")
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	flyingAttacker := addCombatPermanent(g, game.Player2, evasionTestCreature("Flying attacker", game.FlyingStaticBody))
	groundAttacker := addCombatPermanent(g, game.Player2, evasionTestCreature("Ground attacker"))
	flyingNonattacker := addCombatPermanent(g, game.Player2, evasionTestCreature("Flying nonattacker", game.FlyingStaticBody))

	g.Turn.ActivePlayer = game.Player2
	g.Turn.Phase = game.PhaseCombat
	g.Turn.Step = game.StepDeclareAttackers
	g.Combat = &game.CombatState{}
	attackers := mustDeclareAttackersPayload(t, action.DeclareAttackers([]game.AttackDeclaration{
		{Attacker: flyingAttacker.ObjectID, Target: game.AttackTarget{Player: game.Player1}},
		{Attacker: groundAttacker.ObjectID, Target: game.AttackTarget{Player: game.Player1}},
	}))
	if !engine.applyDeclareAttackers(g, game.Player2, attackers) {
		t.Fatal("test attackers were not declared")
	}

	spellID := addCardToHand(g, game.Player1, def)
	g.Players[game.Player1].ManaPool.Add(mana.C, 2)
	g.Players[game.Player1].ManaPool.Add(mana.G, 1)
	g.Turn.Step = game.StepDeclareBlockers
	g.Turn.PriorityPlayer = game.Player1
	if !engine.applyAction(g, game.Player1, action.CastSpell(spellID, nil, 0, nil)) {
		t.Fatal("Wind Shear could not be cast for {2}{G}")
	}
	if g.Players[game.Player1].ManaPool.Amount(mana.C) != 0 ||
		g.Players[game.Player1].ManaPool.Amount(mana.G) != 0 {
		t.Fatal("Wind Shear did not pay {2}{G}")
	}
	engine.resolveTopOfStack(g, &TurnLog{})

	flyingToughness, flyingToughnessOK := effectiveToughness(g, flyingAttacker)
	if hasKeyword(g, flyingAttacker, game.Flying) ||
		effectivePower(g, flyingAttacker) != 1 || !flyingToughnessOK || flyingToughness != 2 {
		t.Fatalf("attacking flyer after Wind Shear: flying=%v power=%d toughness=%d, want no flying and 1/2",
			hasKeyword(g, flyingAttacker, game.Flying), effectivePower(g, flyingAttacker), flyingToughness)
	}
	groundToughness, groundToughnessOK := effectiveToughness(g, groundAttacker)
	if hasKeyword(g, groundAttacker, game.Flying) ||
		effectivePower(g, groundAttacker) != 3 || !groundToughnessOK || groundToughness != 4 {
		t.Fatal("Wind Shear affected an attacking creature without flying")
	}
	nonattackerToughness, nonattackerToughnessOK := effectiveToughness(g, flyingNonattacker)
	if !hasKeyword(g, flyingNonattacker, game.Flying) ||
		effectivePower(g, flyingNonattacker) != 3 || !nonattackerToughnessOK || nonattackerToughness != 4 {
		t.Fatal("Wind Shear affected a flying creature that was not attacking")
	}
}

// Source review: pinned Oracle corpus SHA-256
// 6da9ced0e19bd384771ab8f14051cc45a8dc90302082b875fa6e039d2399a1b4,
// Leering Gargoyle Oracle ID fc0907eb-a3fc-40ef-950d-b05aee9d5da7, printing
// revision a05d1cac-7012-4d22-82d3-0f82b168fe68: "Flying\n{T}: This creature
// gets -2/+2 and loses flying until end of turn." Retrieved its Scryfall rulings
// URI https://api.scryfall.com/cards/a05d1cac-7012-4d22-82d3-0f82b168fe68/rulings
// on 2026-11-10; it returned an empty list (raw response retained at
// .cardwork/source-review/leering-gargoyle-rulings.json). Read the official
// Wizards Comprehensive Rules, effective 2026-11-06, at
// https://media.wizards.com/2026/downloads/MagicCompRules%2020261106.txt
// (verified local copy: .cardwork/source-review/MagicCompRules-20261106.txt):
// 113.10b, 514.2, 601.2c/f-h, 602.2a-b, 608.2c, 611.2a,
// 613.1f-g, 613.4c, 613.6, and 702.9a-c. These require paying the tap cost,
// removing all instances of flying in layer 6, applying -2/+2 in layer 7c,
// and ending both effects during cleanup.
func TestLeeringGargoyleTapGetsMinusTwoPlusTwoAndLosesFlying(t *testing.T) {
	def := evasionRemovalTestCard(t, "fc0907eb-a3fc-40ef-950d-b05aee9d5da7", "Leering Gargoyle")
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	source := addCombatPermanent(g, game.Player1, def)
	if !hasKeyword(g, source, game.Flying) {
		t.Fatal("Leering Gargoyle did not start with flying")
	}
	setMainPhasePriority(g, game.Player1)

	if !engine.applyAction(g, game.Player1, action.ActivateAbility(source.ObjectID, 0, nil, 0)) {
		t.Fatal("Leering Gargoyle's tap ability was not accepted")
	}
	if !source.Tapped {
		t.Fatal("Leering Gargoyle did not pay its tap cost")
	}
	engine.resolveTopOfStack(g, &TurnLog{})

	toughness, ok := effectiveToughness(g, source)
	if !ok || effectivePower(g, source) != 0 || toughness != 4 {
		t.Fatalf("Leering Gargoyle after activation is %d/%d, want 0/4",
			effectivePower(g, source), toughness)
	}
	if hasKeyword(g, source, game.Flying) {
		t.Fatal("Leering Gargoyle retained flying after activating")
	}
}
