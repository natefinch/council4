package rules

import (
	"math/rand/v2"
	"strings"
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
)

func TestAttackTaxOpaqueReversalCanRedeclare(t *testing.T) {
	g, engine, _, _ := costedOpaqueFixture(t, rand.New(&costedOpaqueSource{word: 37}), "battlefield")
	ce := combatEngine{engine}
	addBasicLandPermanent(g, game.Player1, types.Island)
	addCombatPermanent(g, game.Player2, registryCardDef(t, "Propaganda"))
	attacker := addCombatPermanent(g, game.Player1, registryCardDef(t, "Grizzly Bears"))
	attacker.SummoningSick = false
	g.Combat = &game.CombatState{}
	g.Turn.Phase, g.Turn.Step, g.Turn.ActivePlayer = game.PhaseCombat, game.StepDeclareAttackers, game.Player1
	declared, opaque := 0, 0
	agent := &controlScript{
		action: func(_ PlayerObservation, legal []action.Action) action.Action {
			declared++
			for _, candidate := range legal {
				if declare, ok := candidate.DeclareAttackersPayload(); ok && len(declare.Attackers) == 1 &&
					declare.Attackers[0].Attacker == attacker.ObjectID && declare.Attackers[0].Target.Player == game.Player2 {
					return candidate
				}
			}
			t.Fatal("validated declaration was withheld after opaque reversal")
			return action.Pass()
		},
		choice: func(_ PlayerObservation, req game.ChoiceRequest) []int {
			if declared == 1 {
				for _, option := range req.Options {
					if strings.Contains(option.Label, "(opaque)") {
						opaque++
						return []int{option.Index}
					}
				}
			}
			return req.DefaultSelection
		},
	}
	ce.declareAttackers(g, reboundKickerAgents(agent), &TurnLog{})
	if declared != 2 || opaque != 1 || !g.Combat.AttackDeclarationComplete || len(g.Combat.Attackers) != 1 ||
		!attacker.Tapped || g.Players[game.Player1].Life != 39 || g.Players[game.Player1].ManaPool.Total() != 0 ||
		countTokenPermanentsNamed(g, "Opaque token") != 0 {
		t.Fatal("attack declaration did not reuse the shared opaque rollback and then complete normally")
	}
}

func TestRegisteredAttackTaxPlansTreasurePayment(t *testing.T) {
	for _, scenario := range []string{"pay", "decline attack", "unpayable"} {
		t.Run(scenario, func(t *testing.T) {
			g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
			engine := NewEngine(nil)
			ce := combatEngine{engine}
			attacker := addCombatPermanent(g, game.Player1, registryCardDef(t, "Grizzly Bears"))
			attacker.SummoningSick = false
			addCombatPermanent(g, game.Player2, registryCardDef(t, "Propaganda"))
			count := 2
			if scenario == "unpayable" {
				count = 1
			}
			for range count {
				if _, ok := createTokenPermanent(g, game.Player1, recklessLackeyTreasure(t)); !ok {
					t.Fatal("Treasure creation failed")
				}
			}
			g.Combat = &game.CombatState{}
			g.Turn.Phase, g.Turn.Step, g.Turn.ActivePlayer = game.PhaseCombat, game.StepDeclareAttackers, game.Player1
			declarations := []game.AttackDeclaration{{Attacker: attacker.ObjectID, Target: game.AttackTarget{Player: game.Player2}}}
			if scenario == "decline attack" {
				declarations = nil
			}
			declare := action.DeclareAttackersAction{Attackers: declarations}
			payable := scenario != "unpayable"
			if ce.attackersLegal(g, game.Player1, declare) != payable || countTokenPermanentsNamed(g, "Treasure") != count || attacker.Tapped {
				t.Fatal("attack availability lost its witness or mutated resources")
			}
			windows := 0
			agent := &costedObservationAgent{t: t, choose: func(_ PlayerObservation, req game.ChoiceRequest) []int {
				if req.Kind == game.ChoiceMay {
					t.Fatal("mandatory attack tax offered optional payment")
				}
				if req.Prompt == "Activate a mana ability while paying?" {
					windows++
					for _, option := range req.Options {
						if !strings.Contains(option.Label, "Treasure") {
							t.Fatal("attack tax offered an unpaid finish")
						}
					}
				}
				return req.DefaultSelection
			}}
			if ce.applyAttackersWithChoices(g, game.Player1, declare, reboundKickerAgents(agent), &TurnLog{}) != payable {
				t.Fatal("attack-tax application disagreed with the pure witness")
			}
			if scenario == "pay" {
				if windows != 2 || !attacker.Tapped || len(g.Combat.Attackers) != 1 || countTokenPermanentsNamed(g, "Treasure") != 0 ||
					g.Players[game.Player1].ManaPool.Total() != 0 {
					t.Fatal("attack tax was not paid with both planned Treasures")
				}
			} else if windows != 0 || attacker.Tapped || len(g.Combat.Attackers) != 0 || countTokenPermanentsNamed(g, "Treasure") != count {
				t.Fatal("unpaid/declined attack activated a source or declared an attacker")
			}
		})
	}
}

func TestAttackTaxReservesAttackersJointlyWithManaSacrifices(t *testing.T) {
	for _, scenario := range []string{"attacker alone", "spare creature", "two attackers"} {
		t.Run(scenario, func(t *testing.T) {
			g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
			engine := NewEngine(nil)
			ce := combatEngine{engine}
			addCombatPermanent(g, game.Player1, registryCardDef(t, "Ashnod's Altar"))
			addCombatPermanent(g, game.Player2, registryCardDef(t, "Propaganda"))
			attacker := addCombatPermanent(g, game.Player1, registryCardDef(t, "Grizzly Bears"))
			attacker.SummoningSick = false
			declarations := []game.AttackDeclaration{{Attacker: attacker.ObjectID, Target: game.AttackTarget{Player: game.Player2}}}
			var spare *game.Permanent
			if scenario != "attacker alone" {
				spare = addCombatPermanent(g, game.Player1, registryCardDef(t, "Grizzly Bears"))
			}
			if scenario == "two attackers" {
				second := addCombatPermanent(g, game.Player1, registryCardDef(t, "Grizzly Bears"))
				second.SummoningSick = false
				declarations = append(declarations, game.AttackDeclaration{Attacker: second.ObjectID, Target: game.AttackTarget{Player: game.Player2}})
			}
			g.Combat = &game.CombatState{}
			g.Turn.Phase, g.Turn.Step, g.Turn.ActivePlayer = game.PhaseCombat, game.StepDeclareAttackers, game.Player1
			declare := action.DeclareAttackersAction{Attackers: declarations}
			if legal := ce.attackersLegal(g, game.Player1, declare); legal != (scenario == "spare creature") {
				t.Fatalf("attack legality = %v; attacker or spare was double-used", legal)
			}
			prompts := 0
			agent := &costedObservationAgent{t: t, choose: func(_ PlayerObservation, req game.ChoiceRequest) []int {
				prompts++
				for _, option := range req.Options {
					for _, target := range option.Targets {
						for _, declaration := range declarations {
							if target.PermanentID == declaration.Attacker {
								t.Fatal("declared attacker was offered to fund its own tax")
							}
						}
					}
				}
				return req.DefaultSelection
			}}
			paid := ce.applyAttackersWithChoices(g, game.Player1, declare, reboundKickerAgents(agent), &TurnLog{})
			if scenario == "spare creature" {
				if !paid || prompts == 0 || !attacker.Tapped || !g.Players[game.Player1].Graveyard.Contains(spare.CardInstanceID) ||
					g.Players[game.Player1].Graveyard.Contains(attacker.CardInstanceID) || g.Players[game.Player1].ManaPool.Total() != 0 {
					t.Fatal("Altar did not sacrifice only the spare creature to pay the whole tax")
				}
			} else if paid || prompts != 0 || attacker.Tapped || g.Players[game.Player1].Graveyard.Size() != 0 {
				t.Fatal("impossible joint payment prompted or consumed a creature")
			}
		})
	}
}

func TestAttackTaxTotalStaysLockedWhileManaSourcesLeave(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	ce := combatEngine{engine}
	attacker := addCombatCreaturePermanentWithPower(g, game.Player1, 2)
	// A resource-dependent tax makes any accidental recalculation observable:
	// sacrificing the payer's Treasures must not reduce the locked {2} total.
	addCombatPermanent(g, game.Player2, &game.CardDef{CardFace: game.CardFace{
		Name: "Artifact-count attack tax", Types: []types.Card{types.Enchantment},
		StaticAbilities: []game.StaticAbility{{RuleEffects: []game.RuleEffect{{
			Kind: game.RuleEffectAttackTaxPerCreature, AffectedPlayer: game.PlayerYou,
			CardSelection: game.Selection{RequiredTypes: []types.Card{types.Artifact}, Controller: game.ControllerOpponent},
		}}}},
	}})
	for range 2 {
		if _, ok := createTokenPermanent(g, game.Player1, recklessLackeyTreasure(t)); !ok {
			t.Fatal("Treasure creation failed")
		}
	}
	g.Combat = &game.CombatState{}
	g.Turn.Phase, g.Turn.Step, g.Turn.ActivePlayer = game.PhaseCombat, game.StepDeclareAttackers, game.Player1
	declare := action.DeclareAttackersAction{Attackers: []game.AttackDeclaration{{Attacker: attacker.ObjectID, Target: game.AttackTarget{Player: game.Player2}}}}
	if !ce.applyAttackersWithChoices(g, game.Player1, declare, [game.NumPlayers]PlayerAgent{}, &TurnLog{}) ||
		countTokenPermanentsNamed(g, "Treasure") != 0 || g.Players[game.Player1].ManaPool.Total() != 0 || !attacker.Tapped {
		t.Fatal("changing the resource-dependent tax changed the locked payment")
	}
}

func TestAttackTaxTotalStaysLockedWhileOpaqueSourceLeaves(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	ce := combatEngine{engine}
	attacker := addCombatCreaturePermanentWithPower(g, game.Player1, 2)
	grove := addCombatPermanent(g, game.Player1, registryCardDef(t, "Waterlogged Grove"))
	island := addBasicLandPermanent(g, game.Player1, types.Island)
	opaque := addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name: "Opaque token-maker", Types: []types.Card{types.Artifact},
		ManaAbilities: []game.ManaAbility{{
			AdditionalCosts: []cost.Additional{{Kind: cost.AdditionalSacrificeSource, Amount: 1}},
			Content: game.Mode{Sequence: []game.Instruction{
				{Primitive: game.CreateToken{Amount: game.Fixed(1), Source: game.TokenDef(vanillaCreature("Mana companion", 1, 1))}},
				{Primitive: game.AddMana{Amount: game.Fixed(2), ManaColor: mana.C}},
			}}.Ability(),
		}},
	}})
	addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Spare artifact", Types: []types.Card{types.Artifact}}})
	addCombatPermanent(g, game.Player2, &game.CardDef{CardFace: game.CardFace{
		Name: "Artifact-count attack tax", Types: []types.Card{types.Enchantment},
		StaticAbilities: []game.StaticAbility{{RuleEffects: []game.RuleEffect{{
			Kind: game.RuleEffectAttackTaxPerCreature, AffectedPlayer: game.PlayerYou,
			CardSelection: game.Selection{RequiredTypes: []types.Card{types.Artifact}, Controller: game.ControllerOpponent},
		}}}},
	}})
	g.Combat = &game.CombatState{}
	g.Turn.Phase, g.Turn.Step, g.Turn.ActivePlayer = game.PhaseCombat, game.StepDeclareAttackers, game.Player1
	declare := action.DeclareAttackersAction{Attackers: []game.AttackDeclaration{{Attacker: attacker.ObjectID, Target: game.AttackTarget{Player: game.Player2}}}}
	selected := 0
	agent := &costedObservationAgent{t: t, choose: func(_ PlayerObservation, req game.ChoiceRequest) []int {
		for _, option := range req.Options {
			if strings.Contains(option.Label, "Opaque token-maker") && strings.Contains(option.Label, "(opaque)") {
				selected++
				return []int{option.Index}
			}
		}
		return req.DefaultSelection
	}}
	if !ce.applyAttackersWithChoices(g, game.Player1, declare, reboundKickerAgents(agent), &TurnLog{}) || selected != 1 ||
		!attacker.Tapped || grove.Tapped || island.Tapped || !g.Players[game.Player1].Graveyard.Contains(opaque.CardInstanceID) ||
		g.Players[game.Player1].ManaPool.Total() != 0 || countTokenPermanentsNamed(g, "Mana companion") != 1 {
		t.Fatal("opaque activation changed the locked attack total or spent unchosen sources")
	}
	if current, ok := ce.attackTaxCost(g, declare.Attackers); !ok || current.String() != "{1}" {
		t.Fatal("fixture did not reduce the live tax after locking it at {2}")
	}
}

// addScaledEnchantmentAttackTaxPermanent adds the Sphere of Safety static: a
// creature can't attack the controller or a planeswalker they control unless its
// controller pays generic mana equal to the number of enchantments the
// controller controls. The permanent is itself an enchantment, so it counts
// toward its own scaling.
func addScaledEnchantmentAttackTaxPermanent(g *game.Game, controller game.PlayerID) *game.Permanent {
	return addCombatPermanent(g, controller, &game.CardDef{CardFace: game.CardFace{
		Name:  "Sphere of Safety",
		Types: []types.Card{types.Enchantment},
		StaticAbilities: []game.StaticAbility{{
			RuleEffects: []game.RuleEffect{{
				Kind:                           game.RuleEffectAttackTaxPerCreature,
				AffectedPlayer:                 game.PlayerYou,
				AttackTaxIncludesPlaneswalkers: true,
				CardSelection: game.Selection{
					RequiredTypes: []types.Card{types.Enchantment},
					Controller:    game.ControllerYou,
				},
			}},
		}},
	}})
}

// addFixedPlaneswalkerAttackTaxPermanent adds the Baird, Steward of Argive
// static: a creature can't attack the controller or a planeswalker they control
// unless its controller pays a fixed {1} per attacker.
func addFixedPlaneswalkerAttackTaxPermanent(g *game.Game, controller game.PlayerID) *game.Permanent {
	return addCombatPermanent(g, controller, &game.CardDef{CardFace: game.CardFace{
		Name:  "Baird, Steward of Argive",
		Types: []types.Card{types.Creature},
		StaticAbilities: []game.StaticAbility{{
			RuleEffects: []game.RuleEffect{{
				Kind:                           game.RuleEffectAttackTaxPerCreature,
				AffectedPlayer:                 game.PlayerYou,
				AttackTaxIncludesPlaneswalkers: true,
				AttackTaxGeneric:               1,
			}},
		}},
	}})
}

// addDomainAttackTaxPermanent adds the Collective Restraint static: a creature
// can't attack the controller unless its controller pays generic mana equal to
// the number of basic land types among lands the controller controls.
// Planeswalkers are not protected.
func addDomainAttackTaxPermanent(g *game.Game, controller game.PlayerID) *game.Permanent {
	return addCombatPermanent(g, controller, &game.CardDef{CardFace: game.CardFace{
		Name:  "Collective Restraint",
		Types: []types.Card{types.Enchantment},
		StaticAbilities: []game.StaticAbility{{
			RuleEffects: []game.RuleEffect{{
				Kind:                  game.RuleEffectAttackTaxPerCreature,
				AffectedPlayer:        game.PlayerYou,
				AttackTaxScaledAmount: game.AggregateControllerBasicLandTypeCount,
			}},
		}},
	}})
}

func addPlainEnchantmentPermanent(g *game.Game, controller game.PlayerID) *game.Permanent {
	return addCombatPermanent(g, controller, &game.CardDef{CardFace: game.CardFace{
		Name:  "Idle Enchantment",
		Types: []types.Card{types.Enchantment},
	}})
}

func TestScaledEnchantmentAttackTaxChargesPerEnchantment(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	attacker := addCombatCreaturePermanentWithPower(g, game.Player1, 2)
	addScaledEnchantmentAttackTaxPermanent(g, game.Player2)
	g.Combat = &game.CombatState{}
	g.Turn.Phase = game.PhaseCombat
	g.Turn.Step = game.StepDeclareAttackers
	g.Turn.ActivePlayer = game.Player1

	// One enchantment (the Sphere itself) means a {1} tax: illegal without mana.
	actions := legalDeclareAttackersActions(g, game.Player1)
	if declareAttackersActionsContainTarget(actions, attacker.ObjectID, game.AttackTarget{Player: game.Player2}) {
		t.Fatal("scaled-taxed attack was legal without mana")
	}
	forest := addBasicLandPermanent(g, game.Player1, types.Forest)
	actions = legalDeclareAttackersActions(g, game.Player1)
	if !declareAttackersActionsContainTarget(actions, attacker.ObjectID, game.AttackTarget{Player: game.Player2}) {
		t.Fatal("scaled-taxed attack was not legal with one mana available")
	}
	_ = forest

	// A second enchantment raises the tax to {2}: one Forest is no longer enough.
	addPlainEnchantmentPermanent(g, game.Player2)
	actions = legalDeclareAttackersActions(g, game.Player1)
	if declareAttackersActionsContainTarget(actions, attacker.ObjectID, game.AttackTarget{Player: game.Player2}) {
		t.Fatal("two-enchantment scaled tax was legal with only one mana")
	}
	addBasicLandPermanent(g, game.Player1, types.Forest)
	actions = legalDeclareAttackersActions(g, game.Player1)
	if !declareAttackersActionsContainTarget(actions, attacker.ObjectID, game.AttackTarget{Player: game.Player2}) {
		t.Fatal("two-enchantment scaled tax was not legal with two mana")
	}
}

func TestPerCreatureAttackTaxAppliesPredicate(t *testing.T) {
	for name, tc := range map[string]struct {
		includePlaneswalkers bool
		target               game.AttackTarget
		want                 bool
	}{
		"planeswalker-inclusive: direct attack on controller":   {includePlaneswalkers: true, target: game.AttackTarget{Player: game.Player2}, want: true},
		"planeswalker-inclusive: planeswalker the owner keeps":  {includePlaneswalkers: true, target: game.AttackTarget{Player: game.Player2, PlaneswalkerID: 7}, want: true},
		"planeswalker-inclusive: battle the controller guards":  {includePlaneswalkers: true, target: game.AttackTarget{Player: game.Player2, BattleID: 9}, want: false},
		"planeswalker-inclusive: attack on a different player":  {includePlaneswalkers: true, target: game.AttackTarget{Player: game.Player1}, want: false},
		"planeswalker-inclusive: target left combat":            {includePlaneswalkers: true, target: game.AttackTarget{Player: game.Player2, NoTarget: true}, want: false},
		"player-only: direct attack on controller":              {target: game.AttackTarget{Player: game.Player2}, want: true},
		"player-only: planeswalker the controller owns is free": {target: game.AttackTarget{Player: game.Player2, PlaneswalkerID: 7}, want: false},
		"player-only: attack on a different player":             {target: game.AttackTarget{Player: game.Player1}, want: false},
	} {
		t.Run(name, func(t *testing.T) {
			effect := &game.RuleEffect{
				Kind:                           game.RuleEffectAttackTaxPerCreature,
				Controller:                     game.Player2,
				AffectedPlayer:                 game.PlayerYou,
				AttackTaxIncludesPlaneswalkers: tc.includePlaneswalkers,
			}
			declaration := game.AttackDeclaration{Target: tc.target}
			if got := ruleEffectPerCreatureAttackTaxApplies(effect, declaration); got != tc.want {
				t.Fatalf("applies = %v, want %v", got, tc.want)
			}
		})
	}
}

func TestFixedPlaneswalkerAttackTaxChargesPlaneswalkerAttack(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	attacker := addCombatCreaturePermanentWithPower(g, game.Player1, 2)
	addFixedPlaneswalkerAttackTaxPermanent(g, game.Player2)
	planeswalker := addCombatPermanent(g, game.Player2, &game.CardDef{CardFace: game.CardFace{
		Name:  "Guarded Planeswalker",
		Types: []types.Card{types.Planeswalker},
	}})
	g.Combat = &game.CombatState{}
	g.Turn.Phase = game.PhaseCombat
	g.Turn.Step = game.StepDeclareAttackers
	g.Turn.ActivePlayer = game.Player1

	// Attacking the controller's planeswalker owes the {1} tax: illegal unpaid.
	target := game.AttackTarget{Player: game.Player2, PlaneswalkerID: planeswalker.ObjectID}
	actions := legalDeclareAttackersActions(g, game.Player1)
	if declareAttackersActionsContainTarget(actions, attacker.ObjectID, target) {
		t.Fatal("fixed planeswalker-inclusive tax allowed a planeswalker attack without mana")
	}
	addBasicLandPermanent(g, game.Player1, types.Plains)
	declare := mustDeclareAttackersPayload(t, action.DeclareAttackers([]game.AttackDeclaration{
		{Attacker: attacker.ObjectID, Target: target},
	}))
	if !engine.applyDeclareAttackers(g, game.Player1, declare) {
		t.Fatal("fixed planeswalker-inclusive tax rejected a planeswalker attack with one mana available")
	}
}

func TestDomainAttackTaxDoesNotChargePlaneswalkerAttack(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	attacker := addCombatCreaturePermanentWithPower(g, game.Player1, 2)
	addDomainAttackTaxPermanent(g, game.Player2)
	// Two basic land types would make a domain tax of {2} on a player attack.
	addBasicLandPermanent(g, game.Player2, types.Plains)
	addBasicLandPermanent(g, game.Player2, types.Island)
	planeswalker := addCombatPermanent(g, game.Player2, &game.CardDef{CardFace: game.CardFace{
		Name:  "Unguarded Planeswalker",
		Types: []types.Card{types.Planeswalker},
	}})
	g.Combat = &game.CombatState{}
	g.Turn.Phase = game.PhaseCombat
	g.Turn.Step = game.StepDeclareAttackers
	g.Turn.ActivePlayer = game.Player1

	// The player-only domain tax never protects planeswalkers, so the attack is
	// free even with no mana available.
	declare := mustDeclareAttackersPayload(t, action.DeclareAttackers([]game.AttackDeclaration{
		{Attacker: attacker.ObjectID, Target: game.AttackTarget{Player: game.Player2, PlaneswalkerID: planeswalker.ObjectID}},
	}))
	if !engine.applyDeclareAttackers(g, game.Player1, declare) {
		t.Fatal("player-only domain tax charged a planeswalker attack")
	}
}

func TestScaledEnchantmentAttackTaxDoesNotChargeAttacksOnOthers(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	attacker := addCombatCreaturePermanentWithPower(g, game.Player1, 2)
	// Player3 controls the Sphere; an attack on Player2 is not taxed by it.
	addScaledEnchantmentAttackTaxPermanent(g, game.Player3)
	g.Combat = &game.CombatState{}
	g.Turn.Phase = game.PhaseCombat
	g.Turn.Step = game.StepDeclareAttackers
	g.Turn.ActivePlayer = game.Player1

	declare := mustDeclareAttackersPayload(t, action.DeclareAttackers([]game.AttackDeclaration{
		{Attacker: attacker.ObjectID, Target: game.AttackTarget{Player: game.Player2}},
	}))
	if !engine.applyDeclareAttackers(g, game.Player1, declare) {
		t.Fatal("attack on an unprotected player required a tax payment")
	}
}
