package rules

import (
	"slices"
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/id"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// Source review: issue #3934, the exert-as-it-attacks optional cost to attack.
// Pinned-corpus Oracle text (fixtures below are unregistered; a later card unit
// admits them): Hooded Brawler AKH 173 printing
// 1b4dab03-abf1-4c8f-983e-77dffcfb1578 (oracle 4e40a68a-5e0c-4f4f-9757-02f8139797e2),
// Glorybringer TDC 215 printing 06f90d62-6d21-47b1-a427-eb25a42f4dcb (oracle
// b75c3902-633e-4d24-acde-d7a9cc8f466e), Ahn-Crop Champion AKH 194 printing
// 2ddda94a-4b2d-47b0-8d2c-3ad70a0d5584 (oracle c03d073e-ea59-4873-952a-bcf2730f1115)
// and Watchful Naga AKH 193 printing b0ae361e-20e2-4388-bd82-ce041f427495
// (oracle 21932be9-cdfb-48d9-87c7-3d62eca98150). Scryfall rulings
// (https://api.scryfall.com/cards/<printing>/rulings) returned 4, 5, 7 and 4
// WotC rulings dated 2017-04-18: exert is chosen only as attackers are
// declared, so a creature put onto the battlefield attacking can't be exerted;
// it can be exerted with no legal target for its trigger; an exerted creature
// already untapped at that untap step simply expires; tap-and-freeze effects
// aren't exert; a creature stolen until end of turn and exerted untaps during
// its owner's untap step; untapping an attacker doesn't remove it from combat.
// Registered Threaten (oracle 3dc8d15a-3ca7-4d1f-a819-c1b7f947a59e) and Pride
// Sovereign (oracle a131c32d-2b4f-4ee4-aab2-ab05ab010978) drive the control
// and cost-route scenarios. CR 2026-11-06 from the verified official copy:
// 701.43a-d, 508.1f/g/j/k/m, 502.3, 603.2, 603.3d, 603.12, 607.2h, 302.6.

const (
	hoodedBrawlerText   = "You may exert this creature as it attacks. When you do, it gets +2/+2 until end of turn. (An exerted creature won't untap during your next untap step.)"
	glorybringerText    = "You may exert this creature as it attacks. When you do, it deals 4 damage to target non-Dragon creature an opponent controls. (An exerted creature won't untap during your next untap step.)"
	ahnCropChampionText = "You may exert this creature as it attacks. When you do, untap all other creatures you control. (An exerted creature won't untap during your next untap step.)"
	watchfulNagaText    = "You may exert this creature as it attacks. When you do, draw a card. (An exerted creature won't untap during your next untap step.)"
)

func exertCreatureFixture(name string, manaCost cost.Mana, colors []color.Color, subtypes []types.Sub, power, toughness int, text string, whenYouDo game.AbilityContent, keywords ...game.Keyword) *game.CardDef {
	statics := []game.StaticAbility{}
	if len(keywords) > 0 {
		statics = append(statics, game.StaticAbility{KeywordAbilities: game.SimpleKeywords(keywords...)})
	}
	statics = append(statics, game.StaticAbility{
		Text:             text,
		ExertAsItAttacks: opt.Val(game.ExertAttackCost{WhenYouDo: opt.Val(whenYouDo)}),
	})
	return &game.CardDef{CardFace: game.CardFace{
		Name:            name,
		ManaCost:        opt.Val(manaCost),
		Colors:          colors,
		Types:           []types.Card{types.Creature},
		Subtypes:        subtypes,
		Power:           opt.Val(game.PT{Value: power}),
		Toughness:       opt.Val(game.PT{Value: toughness}),
		StaticAbilities: statics,
		OracleText:      text,
	}}
}

func hoodedBrawlerFixture() *game.CardDef {
	return exertCreatureFixture("Hooded Brawler", cost.Mana{cost.O(2), cost.G}, []color.Color{color.Green},
		[]types.Sub{"Snake", "Warrior"}, 3, 2, hoodedBrawlerText,
		game.Mode{Sequence: []game.Instruction{{Primitive: game.ModifyPT{
			Object:         game.SourcePermanentReference(),
			PowerDelta:     game.Fixed(2),
			ToughnessDelta: game.Fixed(2),
			Duration:       game.DurationUntilEndOfTurn,
		}}}}.Ability())
}

func glorybringerFixture() *game.CardDef {
	return exertCreatureFixture("Glorybringer", cost.Mana{cost.O(3), cost.R, cost.R}, []color.Color{color.Red},
		[]types.Sub{"Dragon"}, 4, 4, glorybringerText,
		game.Mode{
			Targets: []game.TargetSpec{{
				MinTargets: 1,
				MaxTargets: 1,
				Constraint: "target non-Dragon creature an opponent controls",
				Allow:      game.TargetAllowPermanent,
				Selection: opt.Val(game.Selection{
					RequiredTypesAny: []types.Card{types.Creature},
					Controller:       game.ControllerOpponent,
					ExcludedSubtype:  types.Sub("Dragon"),
				}),
			}},
			Sequence: []game.Instruction{{Primitive: game.Damage{
				Amount:       game.Fixed(4),
				Recipient:    game.ObjectDamageRecipient(game.TargetPermanentReference(0)),
				DamageSource: opt.Val(game.SourcePermanentReference()),
			}}},
		}.Ability(), game.Flying, game.Haste)
}

func ahnCropChampionFixture() *game.CardDef {
	return exertCreatureFixture("Ahn-Crop Champion", cost.Mana{cost.O(2), cost.G, cost.W}, []color.Color{color.Green, color.White},
		[]types.Sub{"Human", "Warrior"}, 4, 4, ahnCropChampionText,
		game.Mode{Sequence: []game.Instruction{{Primitive: game.Untap{
			Group: game.BattlefieldGroup(game.Selection{
				RequiredTypes: []types.Card{types.Creature},
				Controller:    game.ControllerYou,
				ExcludeSource: true,
			}),
		}}}}.Ability())
}

func watchfulNagaFixture() *game.CardDef {
	return exertCreatureFixture("Watchful Naga", cost.Mana{cost.O(2), cost.G}, []color.Color{color.Green},
		[]types.Sub{"Snake", "Wizard"}, 2, 2, watchfulNagaText,
		game.Mode{Sequence: []game.Instruction{{Primitive: game.Draw{
			Amount: game.Fixed(1),
			Player: game.ControllerReference(),
		}}}}.Ability())
}

func pickExertOptions(attackers ...*game.Permanent) func(game.ChoiceRequest) []int {
	return func(request game.ChoiceRequest) []int {
		var picked []int
		for _, option := range request.Options {
			for _, attacker := range attackers {
				if option.Card.Exists && option.Card.Val.CardID == attacker.CardInstanceID {
					picked = append(picked, option.Index)
				}
			}
		}
		return picked
	}
}

func declareExertAttack(t *testing.T, g *game.Game, engine *Engine, playerID game.PlayerID, agents [game.NumPlayers]PlayerAgent, defender game.PlayerID, attackers ...*game.Permanent) {
	t.Helper()
	declarations := make([]game.AttackDeclaration, len(attackers))
	for i, attacker := range attackers {
		declarations[i] = attackAt(attacker, game.AttackTarget{Player: defender})
	}
	if !engine.applyActionWithChoices(g, playerID, action.DeclareAttackersWithBands(declarations, nil), agents, &TurnLog{}) {
		t.Fatal("attack declaration was rejected")
	}
}

func runUntapStepFor(g *game.Game, engine *Engine, playerID game.PlayerID) {
	g.Turn.ActivePlayer, g.Turn.PriorityPlayer = playerID, playerID
	addCardToLibraryNamed(g, playerID, "Draw Step Card")
	engine.runBeginningPhase(g, [game.NumPlayers]PlayerAgent{}, &TurnLog{})
}

func TestHoodedBrawlerExertGetsBonusBeforeBlockersAndSkipsNextUntap(t *testing.T) {
	for _, exert := range []bool{true, false} {
		name := "declined"
		if exert {
			name = "exerted"
		}
		t.Run(name, func(t *testing.T) {
			g := newEnlistCombatGame()
			brawler := addCombatPermanent(g, game.Player1, hoodedBrawlerFixture())
			policy := &chosenTapPolicy{pick: func(game.ChoiceRequest) []int { return nil }}
			if exert {
				policy.pick = pickExertOptions(brawler)
			}
			var agents [game.NumPlayers]PlayerAgent
			agents[game.Player1] = policy
			engine := NewEngine(nil)

			declareExertAttack(t, g, engine, game.Player1, agents, game.Player2, brawler)

			requests := policy.tapRequests()
			if len(requests) != 1 || requests[0].MinChoices != 0 || requests[0].MaxChoices != 1 ||
				len(requests[0].DefaultSelection) != 0 || !slices.Equal(optionCardIDs(requests[0]), []id.ID{brawler.CardInstanceID}) {
				t.Fatalf("exert requests = %+v, want one optional choice offering Hooded Brawler", requests)
			}
			if brawler.ExertedBy[game.Player1] != exert {
				t.Fatalf("exerted = %v, want %v", brawler.ExertedBy[game.Player1], exert)
			}
			engine.putTriggeredAbilitiesOnStackWithChoices(g, agents, &TurnLog{})
			wantStack := 0
			if exert {
				wantStack = 1
			}
			if g.Stack.Size() != wantStack || g.Turn.Step != game.StepDeclareAttackers {
				t.Fatalf("stack = %d in step %v, want %d in declare attackers", g.Stack.Size(), g.Turn.Step, wantStack)
			}
			if exert {
				engine.resolveTopOfStackWithChoices(g, agents, &TurnLog{})
			}
			power := effectivePower(g, brawler)
			toughness, _ := effectiveToughness(g, brawler)
			if wantPower, wantToughness := 3+2*wantStack, 2+2*wantStack; power != wantPower || toughness != wantToughness {
				t.Fatalf("Hooded Brawler is %d/%d before blockers, want %d/%d", power, toughness, wantPower, wantToughness)
			}

			g.Combat = nil
			runUntapStepFor(g, engine, game.Player1)
			if brawler.Tapped != exert || brawler.ExertedBy[game.Player1] {
				t.Fatalf("after next untap step tapped=%v exerted=%v, want tapped=%v and exert expired", brawler.Tapped, brawler.ExertedBy[game.Player1], exert)
			}
			runUntapStepFor(g, engine, game.Player1)
			if brawler.Tapped {
				t.Fatal("Hooded Brawler did not untap during the following untap step")
			}
		})
	}
}

func TestGlorybringerExertDealsDamageOrExertsWithoutLegalTarget(t *testing.T) {
	for _, hasTarget := range []bool{true, false} {
		name := "no legal target"
		if hasTarget {
			name = "target"
		}
		t.Run(name, func(t *testing.T) {
			g := newEnlistCombatGame()
			glorybringer := addCombatPermanent(g, game.Player1, glorybringerFixture())
			glorybringer.SummoningSick = true
			ownCreature := addCombatCreaturePermanentWithPower(g, game.Player1, 5)
			opposingDragon := addCombatPermanent(g, game.Player2, &game.CardDef{CardFace: game.CardFace{
				Name:      "Opposing Dragon",
				Types:     []types.Card{types.Creature},
				Subtypes:  []types.Sub{"Dragon"},
				Power:     opt.Val(game.PT{Value: 5}),
				Toughness: opt.Val(game.PT{Value: 5}),
			}})
			var victim *game.Permanent
			if hasTarget {
				victim = addCombatCreaturePermanentWithPower(g, game.Player2, 5)
			}
			policy := &chosenTapPolicy{pick: pickExertOptions(glorybringer)}
			var agents [game.NumPlayers]PlayerAgent
			agents[game.Player1] = policy
			engine := NewEngine(nil)

			declareExertAttack(t, g, engine, game.Player1, agents, game.Player2, glorybringer)
			if !glorybringer.ExertedBy[game.Player1] {
				t.Fatal("hasty Glorybringer was not exerted")
			}
			engine.putTriggeredAbilitiesOnStackWithChoices(g, agents, &TurnLog{})
			if hasTarget {
				if g.Stack.Size() != 1 {
					t.Fatalf("stack = %d, want Glorybringer's trigger", g.Stack.Size())
				}
				top, _ := g.Stack.Peek()
				if len(top.Targets) != 1 || top.Targets[0].PermanentID != victim.ObjectID {
					t.Fatalf("trigger targets = %+v, want the opponent's non-Dragon creature", top.Targets)
				}
				engine.resolveTopOfStackWithChoices(g, agents, &TurnLog{})
				if victim.MarkedDamage != 4 {
					t.Fatalf("target has %d damage, want 4 from Glorybringer", victim.MarkedDamage)
				}
			} else if g.Stack.Size() != 0 || len(g.PendingReflexiveTriggers) != 0 {
				t.Fatalf("trigger without a legal target was not removed (stack=%d)", g.Stack.Size())
			}
			if ownCreature.MarkedDamage != 0 || opposingDragon.MarkedDamage != 0 || glorybringer.MarkedDamage != 0 {
				t.Fatal("Glorybringer damaged an illegal target")
			}

			g.Combat = nil
			runUntapStepFor(g, engine, game.Player1)
			if !glorybringer.Tapped {
				t.Fatal("exerted Glorybringer untapped during its controller's next untap step")
			}
		})
	}
}

func TestAhnCropChampionAndWatchfulNagaExertTogether(t *testing.T) {
	g := newEnlistCombatGame()
	champion := addCombatPermanent(g, game.Player1, ahnCropChampionFixture())
	naga := addCombatPermanent(g, game.Player1, watchfulNagaFixture())
	tappedOwn := addCombatCreaturePermanentWithPower(g, game.Player1, 1)
	tappedOwn.Tapped = true
	tappedOpponent := addCombatCreaturePermanentWithPower(g, game.Player2, 1)
	tappedOpponent.Tapped = true
	addCardToLibraryNamed(g, game.Player1, "Naga Draw")
	policy := &chosenTapPolicy{pick: pickExertOptions(champion, naga)}
	var agents [game.NumPlayers]PlayerAgent
	agents[game.Player1] = policy
	engine := NewEngine(nil)

	declareExertAttack(t, g, engine, game.Player1, agents, game.Player2, champion, naga)

	requests := policy.tapRequests()
	if len(requests) != 1 || requests[0].MaxChoices != 2 ||
		!slices.Equal(optionCardIDs(requests[0]), []id.ID{champion.CardInstanceID, naga.CardInstanceID}) {
		t.Fatalf("exert requests = %+v, want one choice offering both attackers", requests)
	}
	if !champion.ExertedBy[game.Player1] || !naga.ExertedBy[game.Player1] {
		t.Fatal("both attackers were not exerted")
	}
	clone := g.Clone()
	for _, original := range []*game.Permanent{champion, naga} {
		copied, ok := permanentByObjectID(clone, original.ObjectID)
		if !ok || !copied.ExertedBy[game.Player1] {
			t.Fatal("game clone lost an exert record")
		}
	}
	engine.putTriggeredAbilitiesOnStackWithChoices(g, agents, &TurnLog{})
	if g.Stack.Size() != 2 || g.Turn.Step != game.StepDeclareAttackers {
		t.Fatalf("stack = %d in step %v, want both exert triggers in declare attackers", g.Stack.Size(), g.Turn.Step)
	}
	engine.resolveTopOfStackWithChoices(g, agents, &TurnLog{})
	engine.resolveTopOfStackWithChoices(g, agents, &TurnLog{})

	if !champion.Tapped || naga.Tapped || tappedOwn.Tapped || !tappedOpponent.Tapped {
		t.Fatalf("tapped champion=%v naga=%v own=%v opponent=%v, want only other creatures you control untapped",
			champion.Tapped, naga.Tapped, tappedOwn.Tapped, tappedOpponent.Tapped)
	}
	if !slices.ContainsFunc(g.Combat.Attackers, func(a game.AttackDeclaration) bool { return a.Attacker == naga.ObjectID }) {
		t.Fatal("untapping Watchful Naga removed it from combat")
	}
	if got := g.Players[game.Player1].Hand.Size(); got != 1 {
		t.Fatalf("hand = %d, want Watchful Naga's draw", got)
	}
	g.Combat = nil
	naga.Tapped = true
	runUntapStepFor(g, engine, game.Player1)
	if !champion.Tapped || !naga.Tapped {
		t.Fatal("an exerted attacker untapped during the next untap step")
	}
}

func TestExertByThiefUntapsDuringOwnersUntapStep(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	g.Turn.TurnNumber = 2
	g.Turn.ActivePlayer, g.Turn.PriorityPlayer = game.Player2, game.Player2
	g.Turn.Phase, g.Turn.Step = game.PhasePrecombatMain, game.StepNone
	engine := NewEngine(nil)
	brawler := addCombatPermanent(g, game.Player1, hoodedBrawlerFixture())
	threaten := addCardToHand(g, game.Player2, registryCardDef(t, "Threaten"))
	g.Players[game.Player2].ManaPool.Add(mana.R, 3)
	if !engine.applyAction(g, game.Player2, action.CastSpell(threaten, []game.Target{game.PermanentTarget(brawler.ObjectID)}, 0, nil)) {
		t.Fatal("casting Threaten on Hooded Brawler failed")
	}
	engine.resolveTopOfStack(g, &TurnLog{})
	if effectiveController(g, brawler) != game.Player2 {
		t.Fatal("Threaten did not steal Hooded Brawler")
	}

	g.Turn.Phase, g.Turn.Step = game.PhaseCombat, game.StepDeclareAttackers
	g.Combat = &game.CombatState{}
	policy := &chosenTapPolicy{pick: pickExertOptions(brawler)}
	var agents [game.NumPlayers]PlayerAgent
	agents[game.Player2] = policy
	declareExertAttack(t, g, engine, game.Player2, agents, game.Player1, brawler)
	if !brawler.ExertedBy[game.Player2] || brawler.ExertedBy[game.Player1] {
		t.Fatalf("exert record = %v, want exerted by the thief", brawler.ExertedBy)
	}
	engine.putTriggeredAbilitiesOnStackWithChoices(g, agents, &TurnLog{})
	engine.resolveTopOfStackWithChoices(g, agents, &TurnLog{})
	if power := effectivePower(g, brawler); power != 5 {
		t.Fatalf("stolen Hooded Brawler power = %d, want 5", power)
	}

	engine.runCleanupStep(g, agents, &TurnLog{})
	if effectiveController(g, brawler) != game.Player1 || !brawler.Tapped {
		t.Fatal("control did not return to the owner at end of turn")
	}
	runUntapStepFor(g, engine, game.Player1)
	if brawler.Tapped {
		t.Fatal("Hooded Brawler did not untap during its owner's untap step")
	}
	if !brawler.ExertedBy[game.Player2] {
		t.Fatal("the thief's exert expired before the thief's next untap step")
	}
	runUntapStepFor(g, engine, game.Player2)
	if brawler.ExertedBy[game.Player2] {
		t.Fatal("the thief's exert did not expire during the thief's untap step")
	}
}

func TestPrideSovereignExertCostUsesSharedExert(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	g.Turn.ActivePlayer, g.Turn.PriorityPlayer = game.Player1, game.Player1
	g.Turn.Phase, g.Turn.Step = game.PhasePrecombatMain, game.StepNone
	engine := NewEngine(nil)
	sovereign := addCombatPermanent(g, game.Player1, registryCardDef(t, "Pride Sovereign"))
	g.Players[game.Player1].ManaPool.Add(mana.W, 1)
	if !engine.applyAction(g, game.Player1, action.ActivateAbility(sovereign.ObjectID, 0, nil, 0)) {
		t.Fatal("activating Pride Sovereign failed")
	}
	if !sovereign.Tapped || !sovereign.ExertedBy[game.Player1] || sovereign.SkipsNextUntap {
		t.Fatalf("tapped=%v exertedBy=%v freeze=%v, want tapped and exerted by its controller",
			sovereign.Tapped, sovereign.ExertedBy, sovereign.SkipsNextUntap)
	}
	engine.resolveTopOfStack(g, &TurnLog{})
	cats := 0
	for _, permanent := range g.Battlefield {
		if permanent.TokenDef != nil && permanent.TokenDef.Name == "Cat" {
			cats++
		}
	}
	if cats != 2 {
		t.Fatalf("cat tokens = %d, want 2", cats)
	}
	runUntapStepFor(g, engine, game.Player1)
	if !sovereign.Tapped {
		t.Fatal("exerted Pride Sovereign untapped during its controller's next untap step")
	}
	runUntapStepFor(g, engine, game.Player1)
	if sovereign.Tapped {
		t.Fatal("Pride Sovereign did not untap during the following untap step")
	}
}

func TestExertAttackCostDeclinedWithoutAgent(t *testing.T) {
	g := newEnlistCombatGame()
	brawler := addCombatPermanent(g, game.Player1, hoodedBrawlerFixture())
	startingLife := g.Players[game.Player2].Life
	attack := action.DeclareAttackersWithBands([]game.AttackDeclaration{attackAt(brawler, game.AttackTarget{Player: game.Player2})}, nil)
	payload, _ := attack.DeclareAttackersPayload()

	after, ok := NewEngine(nil).Simulator().ResolveCombatWithAttackers(g, game.Player1, payload, [game.NumPlayers]PlayerAgent{})
	if !ok {
		t.Fatal("simulated attack was rejected")
	}
	simulated, _ := permanentByObjectID(after, brawler.ObjectID)
	if simulated.ExertedBy[game.Player1] || startingLife-after.Players[game.Player2].Life != 3 {
		t.Fatalf("agentless simulation exerted=%v dealt %d, want declined exert and 3 damage",
			simulated.ExertedBy[game.Player1], startingLife-after.Players[game.Player2].Life)
	}

	if !(combatEngine{NewEngine(nil)}).applyAttackers(g, game.Player1, payload) {
		t.Fatal("agentless attack declaration was rejected")
	}
	if brawler.ExertedBy[game.Player1] || len(g.PendingReflexiveTriggers) != 0 {
		t.Fatal("agentless attack declaration exerted Hooded Brawler")
	}
}

func TestExertedCreatureUntapsDuringAnotherPlayersUntapStepWithSeedborn(t *testing.T) {
	g := newEnlistCombatGame()
	addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:            "Seedborn Muse",
		Types:           []types.Card{types.Creature},
		Power:           opt.Val(game.PT{Value: 2}),
		Toughness:       opt.Val(game.PT{Value: 4}),
		StaticAbilities: []game.StaticAbility{seedbornStaticBody()},
	}})
	brawler := addCombatPermanent(g, game.Player1, hoodedBrawlerFixture())
	var agents [game.NumPlayers]PlayerAgent
	agents[game.Player1] = &chosenTapPolicy{pick: pickExertOptions(brawler)}
	engine := NewEngine(nil)
	declareExertAttack(t, g, engine, game.Player1, agents, game.Player2, brawler)
	g.Combat = nil

	runUntapStepFor(g, engine, game.Player2)
	if brawler.Tapped || !brawler.ExertedBy[game.Player1] {
		t.Fatalf("tapped=%v exerted=%v, want Seedborn untap with exert still pending for its controller", brawler.Tapped, brawler.ExertedBy)
	}
	brawler.Tapped = true
	runUntapStepFor(g, engine, game.Player1)
	if !brawler.Tapped || brawler.ExertedBy[game.Player1] {
		t.Fatal("exert did not stop the untap during its controller's own next untap step")
	}
}
