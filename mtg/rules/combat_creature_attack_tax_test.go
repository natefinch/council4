package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/types"
)

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
