package rules

import (
	"slices"
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

func TestEligibleAttackersFiltersIllegalCreatures(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	eligible := addCombatCreaturePermanent(g, game.Player1)
	tapped := addCombatCreaturePermanent(g, game.Player1)
	tapped.Tapped = true
	sick := addCombatCreaturePermanent(g, game.Player1)
	sick.SummoningSick = true
	hasty := addCombatCreaturePermanent(g, game.Player1, game.Haste)
	hasty.SummoningSick = true
	defender := addCombatCreaturePermanent(g, game.Player1, game.Defender)
	opponent := addCombatCreaturePermanent(g, game.Player2)
	nonCreature := addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Relic",
		Types: []types.Card{types.Artifact}},
	})

	got := eligibleAttackers(g, game.Player1)

	if !slices.Equal(got, []*game.Permanent{eligible, hasty}) {
		t.Fatalf("eligible attackers = %v, want [%v %v]", permanentIDs(got), eligible.ObjectID, hasty.ObjectID)
	}
	for _, permanent := range []*game.Permanent{tapped, sick, defender, opponent, nonCreature} {
		if slices.Contains(got, permanent) {
			t.Fatalf("ineligible permanent %v was eligible", permanent.ObjectID)
		}
	}
}

func TestLegalDeclareAttackersActionsProductiveFirstThenNoAttacks(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	attacker1 := addCombatCreaturePermanent(g, game.Player1)
	attacker2 := addCombatCreaturePermanent(g, game.Player1)
	g.Turn.Phase = game.PhaseCombat
	g.Turn.Step = game.StepDeclareAttackers
	g.Combat = &game.CombatState{}
	g.Players[game.Player3].Eliminated = true
	g.TurnOrder.Eliminate(game.Player3)

	legal := legalDeclareAttackersActions(g, game.Player1)

	if len(legal) != 7 {
		t.Fatalf("legal declare attackers actions = %d, want 7", len(legal))
	}
	wantTargets := []game.PlayerID{game.Player2, game.Player4}
	for targetIndex, target := range wantTargets {
		allAttackersAction := targetIndex*3 + 2
		for i := targetIndex * 3; i <= allAttackersAction; i++ {
			if legal[i].Kind != action.ActionDeclareAttackers {
				t.Fatalf("action %d kind = %v, want declare attackers", i, legal[i].Kind)
			}
		}
		want := []game.AttackDeclaration{
			{Attacker: attacker1.ObjectID, Target: game.AttackTarget{Player: target}},
			{Attacker: attacker2.ObjectID, Target: game.AttackTarget{Player: target}},
		}
		attackers := mustDeclareAttackersPayload(t, legal[allAttackersAction])
		if !slices.Equal(attackers.Attackers, want) {
			t.Fatalf("action %d attackers = %+v, want %+v", allAttackersAction, attackers.Attackers, want)
		}
	}
	attackers := mustDeclareAttackersPayload(t, legal[6])
	if len(attackers.Attackers) != 0 {
		t.Fatalf("last declare attackers action = %+v, want no attacks", attackers.Attackers)
	}
}

func TestGoadedCreatureMustAttackIfAble(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	attacker := addCombatCreaturePermanent(g, game.Player1)
	attacker.Goaded = map[game.PlayerID]game.GoadStatus{game.Player2: {CreatedTurn: 1, ExpiresFor: game.Player2}}
	g.Turn.Phase = game.PhaseCombat
	g.Turn.Step = game.StepDeclareAttackers
	g.Combat = &game.CombatState{}
	engine := NewEngine(nil)

	legal := legalDeclareAttackersActions(g, game.Player1)

	for _, act := range legal {
		attackers := mustDeclareAttackersPayload(t, act)
		if len(attackers.Attackers) == 0 {
			t.Fatalf("legal actions included no attacks despite goaded eligible attacker: %+v", legal)
		}
	}
	if engine.applyDeclareAttackers(g, game.Player1, mustDeclareAttackersPayload(t, action.DeclareAttackers(nil))) {
		t.Fatal("applyDeclareAttackers() accepted no attacks with goaded eligible attacker")
	}
	legalAttack := mustDeclareAttackersPayload(t, action.DeclareAttackers([]game.AttackDeclaration{
		{Attacker: attacker.ObjectID, Target: game.AttackTarget{Player: game.Player3}},
	}))
	if !engine.applyDeclareAttackers(g, game.Player1, legalAttack) {
		t.Fatal("applyDeclareAttackers() rejected legal goaded attack")
	}
}

func TestGoadedCreatureAttacksNonGoadingPlayerIfAble(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	attacker := addCombatCreaturePermanent(g, game.Player1)
	attacker.Goaded = map[game.PlayerID]game.GoadStatus{game.Player2: {CreatedTurn: 1, ExpiresFor: game.Player2}}
	g.Turn.Phase = game.PhaseCombat
	g.Turn.Step = game.StepDeclareAttackers
	g.Combat = &game.CombatState{}
	engine := NewEngine(nil)

	legal := legalDeclareAttackersActions(g, game.Player1)

	for _, act := range legal {
		attackers := mustDeclareAttackersPayload(t, act)
		for _, attack := range attackers.Attackers {
			if attack.Target.Player == game.Player2 {
				t.Fatalf("legal actions included attack at goading player while alternatives exist: %+v", legal)
			}
		}
	}
	goadingAttack := mustDeclareAttackersPayload(t, action.DeclareAttackers([]game.AttackDeclaration{
		{Attacker: attacker.ObjectID, Target: game.AttackTarget{Player: game.Player2}},
	}))
	if engine.applyDeclareAttackers(g, game.Player1, goadingAttack) {
		t.Fatal("applyDeclareAttackers() accepted goaded attack at goading player while alternatives exist")
	}
	nonGoadingAttack := mustDeclareAttackersPayload(t, action.DeclareAttackers([]game.AttackDeclaration{
		{Attacker: attacker.ObjectID, Target: game.AttackTarget{Player: game.Player4}},
	}))
	if !engine.applyDeclareAttackers(g, game.Player1, nonGoadingAttack) {
		t.Fatal("applyDeclareAttackers() rejected attack at non-goading player")
	}
}

func TestGoadedByTwoPlayersMustAttackRemainingNonGoadingOpponentIfAble(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	attacker := addCombatCreaturePermanent(g, game.Player1)
	attacker.Goaded = map[game.PlayerID]game.GoadStatus{
		game.Player2: {CreatedTurn: 1, ExpiresFor: game.Player2},
		game.Player3: {CreatedTurn: 1, ExpiresFor: game.Player3},
	}
	g.Turn.Phase = game.PhaseCombat
	g.Turn.Step = game.StepDeclareAttackers
	g.Combat = &game.CombatState{}
	engine := NewEngine(nil)

	legal := legalDeclareAttackersActions(g, game.Player1)

	if len(legal) != 1 {
		t.Fatalf("legal actions = %d, want only attack at remaining non-goading opponent", len(legal))
	}
	want := action.DeclareAttackers([]game.AttackDeclaration{
		{Attacker: attacker.ObjectID, Target: game.AttackTarget{Player: game.Player4}},
	})
	if !actionsEqual(legal[0], want) {
		t.Fatalf("legal action = %+v, want %+v", legal[0], want)
	}
	goadingAttack := mustDeclareAttackersPayload(t, action.DeclareAttackers([]game.AttackDeclaration{
		{Attacker: attacker.ObjectID, Target: game.AttackTarget{Player: game.Player2}},
	}))
	if engine.applyDeclareAttackers(g, game.Player1, goadingAttack) {
		t.Fatal("applyDeclareAttackers() accepted attack at goading player while remaining opponent exists")
	}
	if !engine.applyDeclareAttackers(g, game.Player1, mustDeclareAttackersPayload(t, want)) {
		t.Fatal("applyDeclareAttackers() rejected attack at remaining non-goading opponent")
	}
}

func TestGoadDoesNotRequireAttackTaxForNonGoadingOpponent(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	attacker := addCombatCreaturePermanent(g, game.Player1)
	attacker.Goaded = map[game.PlayerID]game.GoadStatus{game.Player2: {CreatedTurn: 1, ExpiresFor: game.Player2}}
	for _, defender := range []game.PlayerID{game.Player3, game.Player4} {
		addAttackTaxPermanent(g, defender, 1)
	}
	g.Turn.Phase = game.PhaseCombat
	g.Turn.Step = game.StepDeclareAttackers
	g.Combat = &game.CombatState{}
	engine := NewEngine(nil)

	legal := legalDeclareAttackersActions(g, game.Player1)
	want := action.DeclareAttackers([]game.AttackDeclaration{{
		Attacker: attacker.ObjectID,
		Target:   game.AttackTarget{Player: game.Player2},
	}})
	if len(legal) != 1 || !actionsEqual(legal[0], want) {
		t.Fatalf("legal actions = %+v, want attack at goading player", legal)
	}
	if !engine.applyDeclareAttackers(g, game.Player1, mustDeclareAttackersPayload(t, want)) {
		t.Fatal("applyDeclareAttackers rejected goading player when other opponents required attack taxes")
	}
}

func TestGoadDoesNotForceIllegalAttacks(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	defender := addCombatCreaturePermanent(g, game.Player1, game.Defender)
	defender.Goaded = map[game.PlayerID]game.GoadStatus{game.Player2: {CreatedTurn: 1, ExpiresFor: game.Player2}}
	g.Turn.Phase = game.PhaseCombat
	g.Turn.Step = game.StepDeclareAttackers
	g.Combat = &game.CombatState{}
	engine := NewEngine(nil)

	legal := legalDeclareAttackersActions(g, game.Player1)

	attackers := mustDeclareAttackersPayload(t, legal[0])
	if len(legal) != 1 || len(attackers.Attackers) != 0 {
		t.Fatalf("legal actions = %+v, want only no attacks", legal)
	}
	if !engine.applyDeclareAttackers(g, game.Player1, mustDeclareAttackersPayload(t, action.DeclareAttackers(nil))) {
		t.Fatal("applyDeclareAttackers() rejected no attacks when goaded creature could not legally attack")
	}
}

func TestMustAttackStaticBodyRequiresSourceToAttackIfAble(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	attacker := addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:            "Reckless Bear",
		Types:           []types.Card{types.Creature},
		Power:           opt.Val(game.PT{Value: 3}),
		Toughness:       opt.Val(game.PT{Value: 2}),
		StaticAbilities: []game.StaticAbility{game.MustAttackStaticBody},
	}})
	otherAttacker := addCombatCreaturePermanentWithPower(g, game.Player1, 2)
	g.Turn.Phase = game.PhaseCombat
	g.Turn.Step = game.StepDeclareAttackers
	g.Combat = &game.CombatState{}
	engine := NewEngine(nil)

	legal := legalDeclareAttackersActions(g, game.Player1)
	for _, act := range legal {
		declarations := mustDeclareAttackersPayload(t, act)
		if !slices.ContainsFunc(declarations.Attackers, func(declaration game.AttackDeclaration) bool {
			return declaration.Attacker == attacker.ObjectID
		}) {
			t.Fatalf("legal action omitted required attacker: %+v", declarations.Attackers)
		}
	}
	otherOnly := mustDeclareAttackersPayload(t, action.DeclareAttackers([]game.AttackDeclaration{{
		Attacker: otherAttacker.ObjectID,
		Target:   game.AttackTarget{Player: game.Player2},
	}}))
	if engine.applyDeclareAttackers(g, game.Player1, otherOnly) {
		t.Fatal("applyDeclareAttackers accepted attack without required source")
	}
	requiredAttack := mustDeclareAttackersPayload(t, action.DeclareAttackers([]game.AttackDeclaration{{
		Attacker: attacker.ObjectID,
		Target:   game.AttackTarget{Player: game.Player2},
	}}))
	if !engine.applyDeclareAttackers(g, game.Player1, requiredAttack) {
		t.Fatal("applyDeclareAttackers rejected required source attack")
	}
}

// TestOpponentControlledStaticMustAttackForcesOpponentCreatures proves the
// continuous opponent-scoped forced-attack static the lowering emits for
// "Creatures your opponents control attack each combat if able." (Angler Turtle)
// — a RuleEffectMustAttack whose affected-permanent Selection scopes the
// controller to the opponent relation — forces an opponent's creature to attack
// while leaving the source controller's own creatures free not to attack.
func TestOpponentControlledStaticMustAttackForcesOpponentCreatures(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	addCombatPermanent(g, game.Player2, &game.CardDef{CardFace: game.CardFace{
		Name:  "Angler Turtle",
		Types: []types.Card{types.Creature},
		StaticAbilities: []game.StaticAbility{{
			RuleEffects: []game.RuleEffect{{
				Kind:              game.RuleEffectMustAttack,
				PermanentTypes:    []types.Card{types.Creature},
				AffectedSelection: game.Selection{Controller: game.ControllerOpponent},
			}},
		}},
	}})
	forced := addCombatCreaturePermanent(g, game.Player1)
	ownTurtleSide := addCombatCreaturePermanent(g, game.Player2)
	g.Turn.Phase = game.PhaseCombat
	g.Turn.Step = game.StepDeclareAttackers
	g.Combat = &game.CombatState{}

	// The turtle's controller's opponent (Player1) is forced to attack; the
	// turtle's controller's own creature is not, because the affected-permanent
	// Selection scopes the rule to the opponent of the static's source.
	if !attackerMustAttack(g, forced) {
		t.Fatalf("opponent creature %d was not forced to attack", forced.ObjectID)
	}
	if attackerMustAttack(g, ownTurtleSide) {
		t.Fatalf("source controller's creature %d was wrongly forced to attack", ownTurtleSide.ObjectID)
	}

	g.Turn.ActivePlayer = game.Player1
	legal := legalDeclareAttackersActions(g, game.Player1)
	if len(legal) == 0 {
		t.Fatal("no legal declare-attackers actions")
	}
	for _, act := range legal {
		declarations := mustDeclareAttackersPayload(t, act)
		if !slices.ContainsFunc(declarations.Attackers, func(declaration game.AttackDeclaration) bool {
			return declaration.Attacker == forced.ObjectID
		}) {
			t.Fatalf("legal action omitted forced opponent attacker: %+v", declarations.Attackers)
		}
	}
}

func TestConditionalMixedStaticValuesAffectCharacteristicsAndAttackLegality(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	attacker := addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:      "Delirious Attacker",
		Types:     []types.Card{types.Creature},
		Power:     opt.Val(game.PT{Value: 1}),
		Toughness: opt.Val(game.PT{Value: 1}),
		StaticAbilities: []game.StaticAbility{{
			Condition: opt.Val(game.Condition{ControllerGraveyardCardTypeCountAtLeast: 4}),
			ContinuousEffects: []game.ContinuousEffect{
				{
					Layer:          game.LayerPowerToughnessModify,
					AffectedSource: true,
					PowerDelta:     2,
					ToughnessDelta: 2,
				},
				{
					Layer:          game.LayerAbility,
					AffectedSource: true,
					AddKeywords:    []game.Keyword{game.Flying},
				},
			},
			RuleEffects: []game.RuleEffect{{
				Kind:           game.RuleEffectMustAttack,
				AffectedSource: true,
			}},
		}},
	}})
	g.Turn.Phase = game.PhaseCombat
	g.Turn.Step = game.StepDeclareAttackers
	g.Combat = &game.CombatState{}

	if got := effectivePower(g, attacker); got != 1 {
		t.Fatalf("power before condition = %d, want 1", got)
	}
	if hasKeyword(g, attacker, game.Flying) {
		t.Fatal("attacker had flying before condition")
	}
	if !slices.ContainsFunc(legalDeclareAttackersActions(g, game.Player1), func(act action.Action) bool {
		return len(mustDeclareAttackersPayload(t, act).Attackers) == 0
	}) {
		t.Fatal("no-attack action was not legal before condition")
	}

	for name, cardType := range map[string]types.Card{
		"Relic":  types.Artifact,
		"Bear":   types.Creature,
		"Lesson": types.Sorcery,
		"Trick":  types.Instant,
	} {
		addCardToGraveyard(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
			Name:  name,
			Types: []types.Card{cardType},
		}})
	}

	if got := effectivePower(g, attacker); got != 3 {
		t.Fatalf("power while condition holds = %d, want 3", got)
	}
	if !hasKeyword(g, attacker, game.Flying) {
		t.Fatal("attacker did not have flying while condition held")
	}
	for _, act := range legalDeclareAttackersActions(g, game.Player1) {
		declarations := mustDeclareAttackersPayload(t, act)
		if !slices.ContainsFunc(declarations.Attackers, func(declaration game.AttackDeclaration) bool {
			return declaration.Attacker == attacker.ObjectID
		}) {
			t.Fatalf("legal action omitted conditionally required attacker: %+v", declarations.Attackers)
		}
	}
}

func TestMustAttackStaticBodyDoesNotForceIllegalAttack(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:            "Defensive Bear",
		Types:           []types.Card{types.Creature},
		Power:           opt.Val(game.PT{Value: 3}),
		Toughness:       opt.Val(game.PT{Value: 2}),
		StaticAbilities: []game.StaticAbility{game.DefenderStaticBody, game.MustAttackStaticBody},
	}})
	g.Turn.Phase = game.PhaseCombat
	g.Turn.Step = game.StepDeclareAttackers
	g.Combat = &game.CombatState{}
	engine := NewEngine(nil)

	legal := legalDeclareAttackersActions(g, game.Player1)
	declarations := mustDeclareAttackersPayload(t, legal[0])
	if len(legal) != 1 || len(declarations.Attackers) != 0 {
		t.Fatalf("legal actions = %+v, want only no attacks", legal)
	}
	if !engine.applyDeclareAttackers(g, game.Player1, declarations) {
		t.Fatal("applyDeclareAttackers rejected no attacks when required creature could not attack")
	}
}

func TestMustAttackStaticBodyDoesNotRequirePayingAttackTax(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:            "Reckless Bear",
		Types:           []types.Card{types.Creature},
		Power:           opt.Val(game.PT{Value: 3}),
		Toughness:       opt.Val(game.PT{Value: 2}),
		StaticAbilities: []game.StaticAbility{game.MustAttackStaticBody},
	}})
	for _, defender := range []game.PlayerID{game.Player2, game.Player3, game.Player4} {
		addAttackTaxPermanent(g, defender, 1)
	}
	addBasicLandPermanent(g, game.Player1, types.Forest)
	g.Turn.Phase = game.PhaseCombat
	g.Turn.Step = game.StepDeclareAttackers
	g.Combat = &game.CombatState{}
	engine := NewEngine(nil)

	legal := legalDeclareAttackersActions(g, game.Player1)
	var noAttack action.DeclareAttackersAction
	foundNoAttack := false
	for _, act := range legal {
		declarations := mustDeclareAttackersPayload(t, act)
		if len(declarations.Attackers) == 0 {
			noAttack = declarations
			foundNoAttack = true
			break
		}
	}
	if !foundNoAttack {
		t.Fatalf("legal actions = %+v, want no-attack option despite available mana", legal)
	}
	if !engine.applyDeclareAttackers(g, game.Player1, noAttack) {
		t.Fatal("applyDeclareAttackers rejected declining to pay an attack tax")
	}
}

func TestApplyDeclareAttackersTapsNormalButNotVigilanceAttackers(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	normal := addCombatCreaturePermanent(g, game.Player1)
	vigilance := addCombatCreaturePermanent(g, game.Player1, game.Vigilance)
	g.Turn.Phase = game.PhaseCombat
	g.Turn.Step = game.StepDeclareAttackers
	g.Combat = &game.CombatState{}
	engine := NewEngine(nil)
	declare := mustDeclareAttackersPayload(t, action.DeclareAttackers([]game.AttackDeclaration{
		{Attacker: normal.ObjectID, Target: game.AttackTarget{Player: game.Player2}},
		{Attacker: vigilance.ObjectID, Target: game.AttackTarget{Player: game.Player2}},
	}))

	if !engine.applyDeclareAttackers(g, game.Player1, declare) {
		t.Fatal("applyDeclareAttackers() = false, want true")
	}
	if !normal.Tapped {
		t.Fatal("normal attacker was not tapped")
	}
	if vigilance.Tapped {
		t.Fatal("vigilance attacker was tapped")
	}
	if !slices.Equal(g.Combat.Attackers, declare.Attackers) {
		t.Fatalf("combat attackers = %+v, want %+v", g.Combat.Attackers, declare.Attackers)
	}
}

func TestApplyDeclareAttackersInvalidDoesNotMutate(t *testing.T) {
	tests := []struct {
		name    string
		declare func(*game.Game, *game.Permanent) action.DeclareAttackersAction
	}{
		{
			name: "duplicate attacker",
			declare: func(g *game.Game, attacker *game.Permanent) action.DeclareAttackersAction {
				return mustDeclareAttackersPayload(t, action.DeclareAttackers([]game.AttackDeclaration{
					{Attacker: attacker.ObjectID, Target: game.AttackTarget{Player: game.Player2}},
					{Attacker: attacker.ObjectID, Target: game.AttackTarget{Player: game.Player3}},
				}))
			},
		},
		{
			name: "dead defending player",
			declare: func(g *game.Game, attacker *game.Permanent) action.DeclareAttackersAction {
				g.Players[game.Player2].Eliminated = true
				g.TurnOrder.Eliminate(game.Player2)
				return mustDeclareAttackersPayload(t, action.DeclareAttackers([]game.AttackDeclaration{
					{Attacker: attacker.ObjectID, Target: game.AttackTarget{Player: game.Player2}},
				}))
			},
		},
		{
			name: "planeswalker target",
			declare: func(g *game.Game, attacker *game.Permanent) action.DeclareAttackersAction {
				return mustDeclareAttackersPayload(t, action.DeclareAttackers([]game.AttackDeclaration{
					{Attacker: attacker.ObjectID, Target: game.AttackTarget{Player: game.Player2, PlaneswalkerID: 99}},
				}))
			},
		},
		{
			name: "summoning sick attacker",
			declare: func(g *game.Game, attacker *game.Permanent) action.DeclareAttackersAction {
				attacker.SummoningSick = true
				return mustDeclareAttackersPayload(t, action.DeclareAttackers([]game.AttackDeclaration{
					{Attacker: attacker.ObjectID, Target: game.AttackTarget{Player: game.Player2}},
				}))
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
			attacker := addCombatCreaturePermanent(g, game.Player1)
			g.Turn.Phase = game.PhaseCombat
			g.Turn.Step = game.StepDeclareAttackers
			g.Combat = &game.CombatState{}
			engine := NewEngine(nil)

			if engine.applyDeclareAttackers(g, game.Player1, tt.declare(g, attacker)) {
				t.Fatal("applyDeclareAttackers() = true, want false")
			}
			if len(g.Combat.Attackers) != 0 {
				t.Fatalf("combat attackers = %+v, want none", g.Combat.Attackers)
			}
			if attacker.Tapped {
				t.Fatal("attacker was tapped by invalid declaration")
			}
		})
	}
}

func TestDeclareAttackersCanTargetPlaneswalkersAndBattles(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	attacker := addCombatCreaturePermanent(g, game.Player1)
	planeswalker := addCombatPermanent(g, game.Player2, &game.CardDef{CardFace: game.CardFace{Name: "Test Planeswalker",
		Types:   []types.Card{types.Planeswalker},
		Loyalty: opt.Val(3)},
	})
	battle := addCombatPermanent(g, game.Player3, &game.CardDef{CardFace: game.CardFace{Name: "Test Battle",
		Types:   []types.Card{types.Battle},
		Defense: opt.Val(4)},
	})
	g.Turn.Phase = game.PhaseCombat
	g.Turn.Step = game.StepDeclareAttackers
	g.Combat = &game.CombatState{}
	engine := NewEngine(nil)

	legal := legalDeclareAttackersActions(g, game.Player1)

	wantPlaneswalker := game.AttackTarget{Player: game.Player2, PlaneswalkerID: planeswalker.ObjectID}
	wantBattle := game.AttackTarget{Player: game.Player3, BattleID: battle.ObjectID}
	if !declareAttackersActionsContainTarget(legal, attacker.ObjectID, wantPlaneswalker) {
		t.Fatalf("legal actions = %+v, want planeswalker target %v", legal, wantPlaneswalker)
	}
	if !declareAttackersActionsContainTarget(legal, attacker.ObjectID, wantBattle) {
		t.Fatalf("legal actions = %+v, want battle target %v", legal, wantBattle)
	}
	declare := mustDeclareAttackersPayload(t, action.DeclareAttackers([]game.AttackDeclaration{
		{Attacker: attacker.ObjectID, Target: wantPlaneswalker},
	}))
	if !engine.applyDeclareAttackers(g, game.Player1, declare) {
		t.Fatal("applyDeclareAttackers() rejected valid planeswalker target")
	}
}

// TestMustAttackAttachedRequiresEnchantedCreatureToAttackIfAble verifies that an
// Aura mapping to an AffectedAttached must-attack rule effect forces the
// enchanted creature into every legal attack declaration while leaving other
// creatures free to stay back.
func TestMustAttackAttachedRequiresEnchantedCreatureToAttackIfAble(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	enchanted := addCombatCreaturePermanentWithPower(g, game.Player1, 3)
	other := addCombatCreaturePermanentWithPower(g, game.Player1, 2)
	aura := addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:     "Undying Rage",
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
				RuleEffects: []game.RuleEffect{
					{Kind: game.RuleEffectMustAttack, AffectedAttached: true},
				},
			},
		},
	}})
	if !attachPermanent(g, aura, enchanted) {
		t.Fatal("attachPermanent(aura, enchanted) = false")
	}
	g.Turn.Phase = game.PhaseCombat
	g.Turn.Step = game.StepDeclareAttackers
	g.Turn.ActivePlayer = game.Player1
	g.Combat = &game.CombatState{}

	legal := legalDeclareAttackersActions(g, game.Player1)
	for _, act := range legal {
		declarations := mustDeclareAttackersPayload(t, act)
		if !slices.ContainsFunc(declarations.Attackers, func(declaration game.AttackDeclaration) bool {
			return declaration.Attacker == enchanted.ObjectID
		}) {
			t.Fatalf("legal action omitted required enchanted attacker: %+v", declarations.Attackers)
		}
	}
	if !slices.ContainsFunc(legal, func(act action.Action) bool {
		declarations := mustDeclareAttackersPayload(t, act)
		return !slices.ContainsFunc(declarations.Attackers, func(declaration game.AttackDeclaration) bool {
			return declaration.Attacker == other.ObjectID
		})
	}) {
		t.Fatal("attached must-attack rule forced an unrelated creature to attack")
	}
}

// TestGroupMustAttackRuleForcesOpponentCreaturesToAttack verifies that a
// one-shot, turn-scoped RuleEffectMustAttack scoped to the controller's
// opponents (Bident of Thassa: "Creatures your opponents control attack this
// turn if able.") forces an opponent's creature to attack while leaving the
// controller's own creatures free not to attack.
func TestGroupMustAttackRuleForcesOpponentCreaturesToAttack(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	forced := addCombatCreaturePermanent(g, game.Player1)
	g.Turn.Phase = game.PhaseCombat
	g.Turn.Step = game.StepDeclareAttackers
	g.Combat = &game.CombatState{}
	g.RuleEffects = append(g.RuleEffects, game.RuleEffect{
		ID:                 g.IDGen.Next(),
		Kind:               game.RuleEffectMustAttack,
		Controller:         game.Player2,
		AffectedController: game.ControllerOpponent,
		PermanentTypes:     []types.Card{types.Creature},
	})

	legal := legalDeclareAttackersActions(g, game.Player1)
	if len(legal) == 0 {
		t.Fatal("no legal declare-attackers actions")
	}
	for _, act := range legal {
		declarations := mustDeclareAttackersPayload(t, act)
		if len(declarations.Attackers) == 0 {
			t.Fatalf("legal actions included no attacks despite forced opponent creature: %+v", legal)
		}
		if !slices.ContainsFunc(declarations.Attackers, func(declaration game.AttackDeclaration) bool {
			return declaration.Attacker == forced.ObjectID
		}) {
			t.Fatalf("legal action omitted forced attacker: %+v", declarations.Attackers)
		}
	}
}

// TestGroupMustAttackRuleDoesNotForceControllerCreatures verifies that the
// opponents-scoped forced-attack rule leaves the rule's own controller's
// creatures free not to attack.
func TestGroupMustAttackRuleDoesNotForceControllerCreatures(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	own := addCombatCreaturePermanent(g, game.Player2)
	g.Turn.ActivePlayer = game.Player2
	g.Turn.Phase = game.PhaseCombat
	g.Turn.Step = game.StepDeclareAttackers
	g.Combat = &game.CombatState{}
	g.RuleEffects = append(g.RuleEffects, game.RuleEffect{
		ID:                 g.IDGen.Next(),
		Kind:               game.RuleEffectMustAttack,
		Controller:         game.Player2,
		AffectedController: game.ControllerOpponent,
		PermanentTypes:     []types.Card{types.Creature},
	})

	legal := legalDeclareAttackersActions(g, game.Player2)
	if !slices.ContainsFunc(legal, func(act action.Action) bool {
		return len(mustDeclareAttackersPayload(t, act).Attackers) == 0
	}) {
		t.Fatalf("controller's own creature %d was wrongly forced to attack", own.ObjectID)
	}
}
