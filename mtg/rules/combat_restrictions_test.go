package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

func TestLegalDeclareAttackersIncludesSingleAttackerChoices(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	first := addCombatCreaturePermanentWithPower(g, game.Player1, 2)
	second := addCombatCreaturePermanentWithPower(g, game.Player1, 2)
	g.Combat = &game.CombatState{}
	g.Turn.Phase = game.PhaseCombat
	g.Turn.Step = game.StepDeclareAttackers
	g.Turn.ActivePlayer = game.Player1

	actions := legalDeclareAttackersActions(g, game.Player1)

	if !declareAttackersActionsContainTarget(actions, first.ObjectID, game.AttackTarget{Player: game.Player2}) {
		t.Fatal("legal attacks did not include first creature attacking alone")
	}
	if !declareAttackersActionsContainTarget(actions, second.ObjectID, game.AttackTarget{Player: game.Player2}) {
		t.Fatal("legal attacks did not include second creature attacking alone")
	}
}

func TestPhasedOutCreatureCannotAttackBlockOrBeAttacked(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	attacker := addCombatCreaturePermanentWithPower(g, game.Player1, 2)
	blocker := addCombatCreaturePermanentWithPower(g, game.Player2, 2)
	planeswalker := addCombatPermanent(g, game.Player2, &game.CardDef{CardFace: game.CardFace{Name: "Phased Walker",
		Types:   []types.Card{types.Planeswalker},
		Loyalty: opt.Val(3)},
	})
	attacker.PhasedOut = true
	blocker.PhasedOut = true
	planeswalker.PhasedOut = true
	g.Combat = &game.CombatState{}
	g.Turn.Phase = game.PhaseCombat
	g.Turn.Step = game.StepDeclareAttackers
	g.Turn.ActivePlayer = game.Player1

	if canAttackWith(g, attacker, game.Player1) {
		t.Fatal("phased-out creature can attack")
	}
	if canBlockWith(g, blocker, game.Player2) {
		t.Fatal("phased-out creature can block")
	}
	for _, target := range legalAttackTargets(g, game.Player1) {
		if target.PlaneswalkerID == planeswalker.ObjectID {
			t.Fatal("phased-out planeswalker is an attack target")
		}
	}
}

func TestStaticRuleEffectsCanProhibitAttackingAndBlocking(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	attacker := addCombatCreaturePermanentWithPower(g, game.Player2, 2)
	blocker := addCombatCreaturePermanentWithPower(g, game.Player2, 2)
	addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Pacifying Law",
		Types: []types.Card{types.Enchantment},
		StaticAbilities: []game.StaticAbility{{
			RuleEffects: []game.RuleEffect{
				{
					Kind:               game.RuleEffectCantAttack,
					AffectedController: game.ControllerOpponent,
					PermanentTypes:     []types.Card{types.Creature},
				},
				{
					Kind:               game.RuleEffectCantBlock,
					AffectedController: game.ControllerOpponent,
					PermanentTypes:     []types.Card{types.Creature},
				},
			},
		}}},
	})
	g.Combat = &game.CombatState{}
	g.Turn.Phase = game.PhaseCombat
	g.Turn.Step = game.StepDeclareAttackers
	g.Turn.ActivePlayer = game.Player2

	if canAttackWith(g, attacker, game.Player2) {
		t.Fatal("opponent creature could attack through cant-attack rule effect")
	}
	if canBlockWith(g, blocker, game.Player2) {
		t.Fatal("opponent creature could block through cant-block rule effect")
	}
}

func TestCantAttackStaticBodyProhibitsSourceFromAttacking(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	attacker := addCombatPermanent(g, game.Player2, &game.CardDef{CardFace: game.CardFace{
		Name:            "Pacifist Bear",
		Types:           []types.Card{types.Creature},
		Power:           opt.Val(game.PT{Value: 3}),
		Toughness:       opt.Val(game.PT{Value: 3}),
		StaticAbilities: []game.StaticAbility{game.CantAttackStaticBody},
	}})
	otherAttacker := addCombatCreaturePermanentWithPower(g, game.Player2, 2)
	g.Combat = &game.CombatState{}
	g.Turn.Phase = game.PhaseCombat
	g.Turn.Step = game.StepDeclareAttackers
	g.Turn.ActivePlayer = game.Player2

	if canAttackWith(g, attacker, game.Player2) {
		t.Fatal("source with cannot-attack static ability could attack")
	}
	if !canAttackWith(g, otherAttacker, game.Player2) {
		t.Fatal("cannot-attack static ability affected another creature")
	}
}

func TestCantBlockStaticBodyProhibitsSourceFromBlocking(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	blocker := addCombatPermanent(g, game.Player2, &game.CardDef{CardFace: game.CardFace{
		Name:            "Reluctant Bear",
		Types:           []types.Card{types.Creature},
		Power:           opt.Val(game.PT{Value: 3}),
		Toughness:       opt.Val(game.PT{Value: 3}),
		StaticAbilities: []game.StaticAbility{game.CantBlockStaticBody},
	}})
	otherBlocker := addCombatCreaturePermanentWithPower(g, game.Player2, 2)

	if canBlockWith(g, blocker, game.Player2) {
		t.Fatal("source with cannot-block static ability could block")
	}
	if !canBlockWith(g, otherBlocker, game.Player2) {
		t.Fatal("cannot-block static ability affected another creature")
	}
}

func TestCantBeBlockedStaticBodyProhibitsBlockingSource(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	attacker := addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:            "Elusive Bear",
		Types:           []types.Card{types.Creature},
		Power:           opt.Val(game.PT{Value: 2}),
		Toughness:       opt.Val(game.PT{Value: 2}),
		StaticAbilities: []game.StaticAbility{game.CantBeBlockedStaticBody},
	}})
	otherAttacker := addCombatCreaturePermanentWithPower(g, game.Player1, 2)
	blocker := addCombatCreaturePermanentWithPower(g, game.Player2, 2)

	if canBlockAttacker(g, blocker, attacker) {
		t.Fatal("source with cannot-be-blocked static ability could be blocked")
	}
	if !canBlockAttacker(g, blocker, otherAttacker) {
		t.Fatal("cannot-be-blocked static ability affected another creature")
	}
}

func TestCantAttackRuleCanApplyOnlyToSpecificDefender(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	attacker := addCombatCreaturePermanentWithPower(g, game.Player2, 2)
	addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "No Attacks Here",
		Types: []types.Card{types.Enchantment},
		StaticAbilities: []game.StaticAbility{{
			RuleEffects: []game.RuleEffect{{
				Kind:               game.RuleEffectCantAttack,
				AffectedController: game.ControllerOpponent,
				PermanentTypes:     []types.Card{types.Creature},
				DefendingPlayer:    game.PlayerYou,
			}},
		}}},
	})

	if !canAttackWith(g, attacker, game.Player2) {
		t.Fatal("target-specific cant-attack effect should not remove attack eligibility")
	}
	if canAttackTarget(g, attacker, game.AttackTarget{Player: game.Player1}) {
		t.Fatal("creature could attack protected player")
	}
	if !canAttackTarget(g, attacker, game.AttackTarget{Player: game.Player3}) {
		t.Fatal("creature could not attack unprotected player")
	}
}

// TestCantAttackYouAttachedAuraProtectsOnlyItsController models the Vow cycle:
// an Aura that reads "Enchanted creature can't attack you or planeswalkers you
// control" lowers to a can't-attack rule effect on the attached object restricted
// to the Aura controller. The enchanted creature stays able to attack other
// players.
func TestCantAttackYouAttachedAuraProtectsOnlyItsController(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	enchanted := addCombatCreaturePermanentWithPower(g, game.Player2, 3)
	aura := addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:     "Vow of Duty",
		Types:    []types.Card{types.Enchantment},
		Subtypes: []types.Sub{types.Aura},
		StaticAbilities: []game.StaticAbility{
			{
				KeywordAbilities: []game.KeywordAbility{game.EnchantKeyword{Target: game.TargetSpec{
					Allow:     game.TargetAllowPermanent,
					Predicate: game.TargetPredicate{PermanentTypes: []types.Card{types.Creature}},
				}}},
			},
			{
				RuleEffects: []game.RuleEffect{{
					Kind:             game.RuleEffectCantAttack,
					AffectedAttached: true,
					DefendingPlayer:  game.PlayerYou,
				}},
			},
		},
	}})
	if !attachPermanent(g, aura, enchanted) {
		t.Fatal("attachPermanent(aura, enchanted) = false")
	}
	g.Turn.Phase = game.PhaseCombat
	g.Turn.Step = game.StepDeclareAttackers
	g.Turn.ActivePlayer = game.Player2
	g.Combat = &game.CombatState{}

	if !canAttackWith(g, enchanted, game.Player2) {
		t.Fatal("defender-restricted cant-attack effect should not remove attack eligibility")
	}
	if canAttackTarget(g, enchanted, game.AttackTarget{Player: game.Player1}) {
		t.Fatal("enchanted creature could attack the Aura's controller")
	}
	if !canAttackTarget(g, enchanted, game.AttackTarget{Player: game.Player3}) {
		t.Fatal("enchanted creature could not attack an unprotected player")
	}
}

func TestEliminatedPlayerCleanupRemovesCombatAndStackObjects(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	attacker := addCombatCreaturePermanentWithPower(g, game.Player1, 2)
	blocker := addCombatCreaturePermanentWithPower(g, game.Player2, 2)
	owned := addCombatCreaturePermanentWithPower(g, game.Player2, 2)
	controlled := addCombatCreaturePermanentWithPower(g, game.Player1, 2)
	controlled.Controller = game.Player2
	g.Combat = &game.CombatState{
		Attackers: []game.AttackDeclaration{{Attacker: attacker.ObjectID, Target: game.AttackTarget{Player: game.Player2}}},
		Blockers:  []game.BlockDeclaration{{Blocker: blocker.ObjectID, Blocking: attacker.ObjectID}},
	}
	g.Stack.Push(&game.StackObject{ID: g.IDGen.Next(), Controller: game.Player2})

	engine.eliminatePlayer(g, game.Player2)

	if len(g.Combat.Attackers) != 0 || len(g.Combat.Blockers) != 0 {
		t.Fatalf("combat after elimination attackers=%+v blockers=%+v, want cleared", g.Combat.Attackers, g.Combat.Blockers)
	}
	if g.Stack.Size() != 0 {
		t.Fatalf("stack size after elimination = %d, want 0", g.Stack.Size())
	}
	if _, ok := permanentByObjectID(g, owned.ObjectID); ok || !g.Players[game.Player2].Exile.Contains(owned.CardInstanceID) {
		t.Fatal("eliminated player's owned permanent did not leave battlefield")
	}
	if _, ok := permanentByObjectID(g, controlled.ObjectID); !ok || controlled.Controller != game.Player1 {
		t.Fatalf("controlled permanent after elimination = %+v, want returned to owner control", controlled)
	}
}

func TestAttackTaxFiltersAndChargesDeclareAttackers(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	attacker := addCombatCreaturePermanentWithPower(g, game.Player1, 2)
	addAttackTaxPermanent(g, game.Player2, 1)
	g.Combat = &game.CombatState{}
	g.Turn.Phase = game.PhaseCombat
	g.Turn.Step = game.StepDeclareAttackers
	g.Turn.ActivePlayer = game.Player1

	actions := legalDeclareAttackersActions(g, game.Player1)
	if declareAttackersActionsContainTarget(actions, attacker.ObjectID, game.AttackTarget{Player: game.Player2}) {
		t.Fatal("taxed attack was legal without mana")
	}
	forest := addBasicLandPermanent(g, game.Player1, types.Forest)
	actions = legalDeclareAttackersActions(g, game.Player1)
	if !declareAttackersActionsContainTarget(actions, attacker.ObjectID, game.AttackTarget{Player: game.Player2}) {
		t.Fatal("taxed attack was not legal with mana available")
	}

	declare := mustDeclareAttackersPayload(t, action.DeclareAttackers([]game.AttackDeclaration{{Attacker: attacker.ObjectID, Target: game.AttackTarget{Player: game.Player2}}}))
	if !engine.applyDeclareAttackers(g, game.Player1, declare) {
		t.Fatal("applyDeclareAttackers() = false, want tax payment to succeed")
	}
	if !forest.Tapped {
		t.Fatal("attack tax did not tap mana source")
	}
}

func addAttackTaxPermanent(g *game.Game, controller game.PlayerID, amount int) *game.Permanent {
	return addCombatPermanent(g, controller, &game.CardDef{CardFace: game.CardFace{
		Name:  "Attack Tax",
		Types: []types.Card{types.Enchantment},
		StaticAbilities: []game.StaticAbility{{
			RuleEffects: []game.RuleEffect{{
				Kind:             game.RuleEffectAttackTax,
				AffectedPlayer:   game.PlayerYou,
				AttackTaxGeneric: amount,
			}},
		}},
	}})
}

func TestAttackTaxCannotBePaidByDeclaredAttackerManaAbility(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	manaDork := addManaAbilityPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Mana Dork",
		Types:           []types.Card{types.Creature},
		Power:           opt.Val(game.PT{Value: 1}),
		Toughness:       opt.Val(game.PT{Value: 1}),
		StaticAbilities: []game.StaticAbility{game.HasteStaticBody}},
	}, mana.G, 1)
	manaDork.SummoningSick = false
	addAttackTaxPermanent(g, game.Player2, 1)
	g.Combat = &game.CombatState{}
	g.Turn.Phase = game.PhaseCombat
	g.Turn.Step = game.StepDeclareAttackers
	g.Turn.ActivePlayer = game.Player1

	actions := legalDeclareAttackersActions(g, game.Player1)

	if declareAttackersActionsContainTarget(actions, manaDork.ObjectID, game.AttackTarget{Player: game.Player2}) {
		t.Fatal("taxed attack was legal by using the declared attacker as its own mana source")
	}
}

func TestAttackTaxesStackPerDeclaredAttacker(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	first := addCombatCreaturePermanentWithPower(g, game.Player1, 2)
	second := addCombatCreaturePermanentWithPower(g, game.Player1, 2)
	addAttackTaxPermanent(g, game.Player2, 1)
	addAttackTaxPermanent(g, game.Player2, 2)
	var forests []*game.Permanent
	for range 5 {
		forests = append(forests, addBasicLandPermanent(g, game.Player1, types.Forest))
	}
	g.Combat = &game.CombatState{}
	g.Turn.Phase = game.PhaseCombat
	g.Turn.Step = game.StepDeclareAttackers
	g.Turn.ActivePlayer = game.Player1
	declarations := []game.AttackDeclaration{
		{Attacker: first.ObjectID, Target: game.AttackTarget{Player: game.Player2}},
		{Attacker: second.ObjectID, Target: game.AttackTarget{Player: game.Player2}},
	}

	declare := mustDeclareAttackersPayload(t, action.DeclareAttackers(declarations))
	if engine.applyDeclareAttackers(g, game.Player1, declare) {
		t.Fatal("applyDeclareAttackers() = true with only five mana, want stacked per-attacker tax to make declaration illegal")
	}
	for _, forest := range forests {
		if forest.Tapped {
			t.Fatal("failed attack declaration spent mana")
		}
	}

	if len(g.Combat.Attackers) != 0 {
		t.Fatalf("combat attackers = %#v after failed declaration", g.Combat.Attackers)
	}

	forests = append(forests, addBasicLandPermanent(g, game.Player1, types.Forest))
	if !engine.applyDeclareAttackers(g, game.Player1, declare) {
		t.Fatal("applyDeclareAttackers() = false with six mana, want payment to succeed")
	}
	for _, forest := range forests {
		if !forest.Tapped {
			t.Fatal("successful stacked attack-tax payment left a mana source untapped")
		}
	}
}

func TestAttackTaxTracksAffectedPlayerAndDirectPlayerAttacks(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	attacker := addCombatCreaturePermanentWithPower(g, game.Player1, 2)
	taxSource := addAttackTaxPermanent(g, game.Player2, 2)
	planeswalker := addCombatPermanent(g, game.Player2, &game.CardDef{CardFace: game.CardFace{
		Name:  "Defended Planeswalker",
		Types: []types.Card{types.Planeswalker},
	}})
	g.Combat = &game.CombatState{}
	g.Turn.Phase = game.PhaseCombat
	g.Turn.Step = game.StepDeclareAttackers
	g.Turn.ActivePlayer = game.Player1

	actions := legalDeclareAttackersActions(g, game.Player1)
	if declareAttackersActionsContainTarget(actions, attacker.ObjectID, game.AttackTarget{Player: game.Player2}) {
		t.Fatal("direct attack on affected player was legal without mana")
	}
	if !declareAttackersActionsContainTarget(actions, attacker.ObjectID, game.AttackTarget{Player: game.Player3}) {
		t.Fatal("attack on another opponent was taxed")
	}
	if !declareAttackersActionsContainTarget(actions, attacker.ObjectID, game.AttackTarget{
		Player:         game.Player2,
		PlaneswalkerID: planeswalker.ObjectID,
	}) {
		t.Fatal("planeswalker attack was taxed by a player-only attack cost")
	}

	taxSource.Controller = game.Player3
	actions = legalDeclareAttackersActions(g, game.Player1)
	if !declareAttackersActionsContainTarget(actions, attacker.ObjectID, game.AttackTarget{Player: game.Player2}) {
		t.Fatal("old controller remained affected after control changed")
	}
	if declareAttackersActionsContainTarget(actions, attacker.ObjectID, game.AttackTarget{Player: game.Player3}) {
		t.Fatal("new controller was not affected after control changed")
	}
}

func TestAttackTaxPaymentCanBeDeclinedByDeclaringNoAttack(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	attacker := addCombatCreaturePermanentWithPower(g, game.Player1, 2)
	addAttackTaxPermanent(g, game.Player2, 2)
	firstForest := addBasicLandPermanent(g, game.Player1, types.Forest)
	secondForest := addBasicLandPermanent(g, game.Player1, types.Forest)
	g.Combat = &game.CombatState{}
	g.Turn.Phase = game.PhaseCombat
	g.Turn.Step = game.StepDeclareAttackers
	g.Turn.ActivePlayer = game.Player1

	actions := legalDeclareAttackersActions(g, game.Player1)
	if !declareAttackersActionsContainTarget(actions, attacker.ObjectID, game.AttackTarget{Player: game.Player2}) {
		t.Fatal("payable taxed attack was not offered")
	}
	noAttack := mustDeclareAttackersPayload(t, action.DeclareAttackers(nil))
	if !engine.applyDeclareAttackers(g, game.Player1, noAttack) {
		t.Fatal("declaring no attackers was illegal")
	}
	if firstForest.Tapped || secondForest.Tapped {
		t.Fatal("declining the taxed attack spent mana")
	}
	if len(g.Combat.Attackers) != 0 {
		t.Fatalf("combat attackers = %#v, want none", g.Combat.Attackers)
	}
}
