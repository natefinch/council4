package rules

import (
	"slices"
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/compare"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

func TestPlayerTargetedSpellCreatesOneLegalActionPerAlivePlayer(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	spellID := addCardToHand(g, game.Player1, playerDamageSpell())
	if !engine.eliminatePlayer(g, game.Player4) {
		t.Fatal("eliminatePlayer() = false, want true")
	}
	g.Turn.Phase = game.PhasePrecombatMain
	g.Turn.Step = game.StepNone

	legal := engine.legalActions(g, game.Player1)

	var castTargets []game.Target
	for _, act := range legal {
		cast, ok := act.CastSpellPayload()
		if ok && cast.CardID == spellID {
			if len(cast.Targets) != 1 {
				t.Fatalf("cast targets = %d, want 1", len(cast.Targets))
			}
			castTargets = append(castTargets, cast.Targets[0])
		}
	}
	wantTargets := []game.Target{
		game.PlayerTarget(game.Player1),
		game.PlayerTarget(game.Player2),
		game.PlayerTarget(game.Player3),
	}
	if len(castTargets) != len(wantTargets) {
		t.Fatalf("cast actions = %d, want %d", len(castTargets), len(wantTargets))
	}
	for i, want := range wantTargets {
		if castTargets[i] != want {
			t.Fatalf("target %d = %+v, want %+v", i, castTargets[i], want)
		}
	}
}

func TestTargetedEffectUsesSelectedTarget(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	spellID := addCardToHand(g, game.Player1, playerDamageSpell())
	g.Turn.Phase = game.PhasePrecombatMain
	g.Turn.Step = game.StepNone

	if !engine.applyAction(g, game.Player1, action.CastSpell(spellID, []game.Target{game.PlayerTarget(game.Player2)}, 0, nil)) {
		t.Fatal("applyAction() = false, want true")
	}
	engine.resolveTopOfStack(g, &TurnLog{})

	if g.Players[game.Player1].Life != 40 {
		t.Fatalf("caster life = %d, want 40", g.Players[game.Player1].Life)
	}
	if g.Players[game.Player2].Life != 37 {
		t.Fatalf("target life = %d, want 37", g.Players[game.Player2].Life)
	}
}

func TestDeadPlayerTargetDoesNotApplyEffect(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	sourceID := addEffectSpellToStack(g, game.Player1, &game.Effect{
		Type:        game.EffectDamage,
		Amount:      3,
		TargetIndex: 0,
	}, []game.Target{game.PlayerTarget(game.Player2)})
	if !engine.eliminatePlayer(g, game.Player2) {
		t.Fatal("eliminatePlayer() = false, want true")
	}

	engine.resolveTopOfStack(g, &TurnLog{})

	if g.Players[game.Player2].Life != 40 {
		t.Fatalf("dead target life = %d, want 40", g.Players[game.Player2].Life)
	}
	if !g.Players[game.Player1].Graveyard.Contains(sourceID) {
		t.Fatal("spell did not move to graveyard")
	}
}

func TestTargetThatDiesBeforeResolutionDoesNotApplyEffect(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	spellID := addCardToHand(g, game.Player1, playerDamageSpell())
	g.Turn.Phase = game.PhasePrecombatMain
	g.Turn.Step = game.StepNone
	if !engine.applyAction(g, game.Player1, action.CastSpell(spellID, []game.Target{game.PlayerTarget(game.Player2)}, 0, nil)) {
		t.Fatal("applyAction() = false, want true")
	}
	if !engine.eliminatePlayer(g, game.Player2) {
		t.Fatal("eliminatePlayer() = false, want true")
	}

	engine.resolveTopOfStack(g, &TurnLog{})

	if g.Players[game.Player2].Life != 40 {
		t.Fatalf("dead target life = %d, want 40", g.Players[game.Player2].Life)
	}
	if !g.Players[game.Player1].Graveyard.Contains(spellID) {
		t.Fatal("spell did not move to graveyard")
	}
}

func TestPermanentTargetedSpellCreatesActionsForMatchingPermanents(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	creature := addCreaturePermanent(g, game.Player1)
	addBasicLandPermanent(g, game.Player1, types.Forest)
	addBasicLandPermanent(g, game.Player2, types.Forest)
	spellID := addCardToHand(g, game.Player1, permanentTargetSpell("creature"))
	g.Turn.Phase = game.PhasePrecombatMain
	g.Turn.Step = game.StepNone

	legal := engine.legalActions(g, game.Player1)

	want := action.CastSpell(spellID, []game.Target{game.PermanentTarget(creature.ObjectID)}, 0, nil)
	if !containsAction(legal, want) {
		t.Fatalf("legal actions did not include creature target action %+v", want)
	}
	for _, act := range legal {
		cast, ok := act.CastSpellPayload()
		if !ok || cast.CardID != spellID {
			continue
		}
		if len(cast.Targets) != 1 || cast.Targets[0] != game.PermanentTarget(creature.ObjectID) {
			t.Fatalf("unexpected cast target %+v", cast.Targets)
		}
	}
}

func TestOptionalPermanentTargetAllowsTargetOrNoTarget(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	creature := addCreaturePermanent(g, game.Player1)
	addBasicLandPermanent(g, game.Player1, types.Forest)
	spellID := addCardToHand(g, game.Player1, optionalPermanentTargetSpell("creature"))
	g.Turn.Phase = game.PhasePrecombatMain
	g.Turn.Step = game.StepNone

	legal := engine.legalActions(g, game.Player1)

	if !containsAction(legal, action.CastSpell(spellID, []game.Target{game.PermanentTarget(creature.ObjectID)}, 0, nil)) {
		t.Fatal("legal actions did not include optional target spell with a target")
	}
	if !containsAction(legal, action.CastSpell(spellID, nil, 0, nil)) {
		t.Fatal("legal actions did not include optional target spell with no target")
	}
	if !engine.applyAction(g, game.Player1, action.CastSpell(spellID, []game.Target{game.PermanentTarget(creature.ObjectID)}, 0, nil)) {
		t.Fatal("applyAction() with optional target = false, want true")
	}
}

func TestOptionalTargetWithNoCandidatesHasSingleNoTargetChoice(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})

	result := targetChoicesForSpecs(g, game.Player1, nil, 0, []game.TargetSpec{
		{MinTargets: 0, MaxTargets: 2, Constraint: "creature"},
	})

	if result.kind != targetLegalChoicesFound {
		t.Fatalf("kind = %v, want targetLegalChoicesFound", result.kind)
	}
	if len(result.choices) != 1 {
		t.Fatalf("choices = %d, want one no-target choice", len(result.choices))
	}
	if result.choices[0] != nil {
		t.Fatalf("choice = %+v, want nil no-target choice", result.choices[0])
	}
}

func TestTargetChoiceResultKinds(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	creature := addCreaturePermanent(g, game.Player1)

	tests := []struct {
		name     string
		specs    []game.TargetSpec
		wantKind targetChoiceKind
		wantErr  bool
	}{
		{
			name:     "no specs means no targets required",
			specs:    nil,
			wantKind: targetNoTargetsRequired,
		},
		{
			name:     "empty specs means no targets required",
			specs:    []game.TargetSpec{},
			wantKind: targetNoTargetsRequired,
		},
		{
			name: "required target with legal candidate",
			specs: []game.TargetSpec{
				{MinTargets: 1, MaxTargets: 1, Constraint: "creature"},
			},
			wantKind: targetLegalChoicesFound,
		},
		{
			name: "required target with no legal candidates",
			specs: []game.TargetSpec{
				{MinTargets: 1, MaxTargets: 1, Constraint: "planeswalker"},
			},
			wantKind: targetNoLegalChoices,
		},
		{
			name: "optional target with candidates produces legal choices",
			specs: []game.TargetSpec{
				{MinTargets: 0, MaxTargets: 1, Constraint: "creature"},
			},
			wantKind: targetLegalChoicesFound,
		},
		{
			name: "invalid spec max less than min returns error",
			specs: []game.TargetSpec{
				{MinTargets: 2, MaxTargets: 1, Constraint: "creature"},
			},
			wantKind: targetInvalidSpec,
			wantErr:  true,
		},
		{
			name: "invalid spec negative min returns error",
			specs: []game.TargetSpec{
				{MinTargets: -1, MaxTargets: 1, Constraint: "creature"},
			},
			wantKind: targetInvalidSpec,
			wantErr:  true,
		},
	}

	_ = creature // used as a board-state fixture for legal-candidate cases

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := targetChoicesForSpecs(g, game.Player1, nil, 0, tt.specs)
			if result.kind != tt.wantKind {
				t.Errorf("kind = %v, want %v", result.kind, tt.wantKind)
			}
			if tt.wantErr && result.err == nil {
				t.Error("err = nil, want non-nil error for invalid spec")
			}
			if !tt.wantErr && result.err != nil {
				t.Errorf("err = %v, want nil", result.err)
			}
			if tt.wantKind == targetNoTargetsRequired || tt.wantKind == targetLegalChoicesFound {
				if len(result.choices) == 0 {
					t.Error("choices is empty, want at least one choice")
				}
			}
			if tt.wantKind == targetNoLegalChoices || tt.wantKind == targetInvalidSpec {
				if len(result.choices) != 0 {
					t.Errorf("choices = %d, want 0 for %v", len(result.choices), tt.wantKind)
				}
			}
		})
	}
}

func TestMultiTargetSpellCreatesCombinations(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	first := addCreaturePermanent(g, game.Player1)
	second := addCreaturePermanent(g, game.Player1)
	third := addCreaturePermanent(g, game.Player2)
	addBasicLandPermanent(g, game.Player1, types.Forest)
	spellID := addCardToHand(g, game.Player1, permanentTargetSpellWithRange("creature", 2, 2))
	g.Turn.Phase = game.PhasePrecombatMain
	g.Turn.Step = game.StepNone

	legal := engine.legalActions(g, game.Player1)

	wantTargets := [][]game.Target{
		{game.PermanentTarget(first.ObjectID), game.PermanentTarget(second.ObjectID)},
		{game.PermanentTarget(first.ObjectID), game.PermanentTarget(third.ObjectID)},
		{game.PermanentTarget(second.ObjectID), game.PermanentTarget(third.ObjectID)},
	}
	for _, targets := range wantTargets {
		if !containsAction(legal, action.CastSpell(spellID, targets, 0, nil)) {
			t.Fatalf("legal actions did not include targets %+v", targets)
		}
	}
	if containsAction(legal, action.CastSpell(spellID, []game.Target{game.PermanentTarget(first.ObjectID)}, 0, nil)) {
		t.Fatal("legal actions included too few targets for two-target spell")
	}
}

func TestUpToTwoTargetsIncludesZeroOneAndTwoTargets(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	first := addCreaturePermanent(g, game.Player1)
	second := addCreaturePermanent(g, game.Player2)
	addBasicLandPermanent(g, game.Player1, types.Forest)
	spellID := addCardToHand(g, game.Player1, permanentTargetSpellWithRange("creature", 0, 2))
	g.Turn.Phase = game.PhasePrecombatMain
	g.Turn.Step = game.StepNone

	legal := engine.legalActions(g, game.Player1)

	for _, targets := range [][]game.Target{
		nil,
		{game.PermanentTarget(first.ObjectID)},
		{game.PermanentTarget(second.ObjectID)},
		{game.PermanentTarget(first.ObjectID), game.PermanentTarget(second.ObjectID)},
	} {
		if !containsAction(legal, action.CastSpell(spellID, targets, 0, nil)) {
			t.Fatalf("legal actions did not include targets %+v", targets)
		}
	}
}

func TestMixedTargetSlotsCreateCartesianProduct(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	creature := addCreaturePermanent(g, game.Player1)
	addBasicLandPermanent(g, game.Player1, types.Forest)
	spellID := addCardToHand(g, game.Player1, permanentTargetSpellWithSpecs([]game.TargetSpec{
		{MinTargets: 1, MaxTargets: 1, Constraint: "creature"},
		{MinTargets: 1, MaxTargets: 1, Constraint: "opponent"},
	}))
	g.Turn.Phase = game.PhasePrecombatMain
	g.Turn.Step = game.StepNone

	legal := engine.legalActions(g, game.Player1)

	for _, playerID := range []game.PlayerID{game.Player2, game.Player3, game.Player4} {
		targets := []game.Target{game.PermanentTarget(creature.ObjectID), game.PlayerTarget(playerID)}
		if !containsAction(legal, action.CastSpell(spellID, targets, 0, nil)) {
			t.Fatalf("legal actions did not include mixed targets %+v", targets)
		}
	}
}

func TestStructuredTargetPredicates(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	blackCreature := addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Black Creature",
		ManaCost:  opt.Val(cost.Mana{cost.O(2)}),
		Colors:    []color.Color{color.Black},
		Types:     []types.Card{types.Creature},
		Power:     opt.Val(game.PT{Value: 3}),
		Toughness: opt.Val(game.PT{Value: 3})},
	})
	whiteCreature := addCombatPermanent(g, game.Player2, &game.CardDef{CardFace: game.CardFace{Name: "White Creature",
		ManaCost:  opt.Val(cost.Mana{cost.O(4)}),
		Colors:    []color.Color{color.White},
		Types:     []types.Card{types.Creature},
		Power:     opt.Val(game.PT{Value: 2}),
		Toughness: opt.Val(game.PT{Value: 2}),
		Abilities: []game.AbilityDef{{Kind: game.StaticAbility, Keywords: []game.Keyword{game.Flying}}}},
	})
	whiteCreature.Tapped = true
	addBasicLandPermanent(g, game.Player1, types.Forest)
	spellID := addCardToHand(g, game.Player1, permanentTargetSpellWithSpecs([]game.TargetSpec{
		{
			MinTargets: 1,
			MaxTargets: 1,
			Constraint: "nonblack tapped creature with flying mana value 4 or less an opponent controls",
			Allow:      game.TargetAllowPermanent,
			Predicate: game.TargetPredicate{
				PermanentTypes: []types.Card{types.Creature},
				ExcludedColors: []color.Color{color.Black},
				Controller:     game.ControllerOpponent,
				Tapped:         game.TriTrue,
				Keyword:        game.Flying,
				ManaValue:      opt.Val(compare.Int{Op: compare.LessOrEqual, Value: 4}),
			},
		},
	}))
	g.Turn.Phase = game.PhasePrecombatMain
	g.Turn.Step = game.StepNone

	legal := engine.legalActions(g, game.Player1)

	if !containsAction(legal, action.CastSpell(spellID, []game.Target{game.PermanentTarget(whiteCreature.ObjectID)}, 0, nil)) {
		t.Fatal("structured predicate did not allow matching creature")
	}
	if containsAction(legal, action.CastSpell(spellID, []game.Target{game.PermanentTarget(blackCreature.ObjectID)}, 0, nil)) {
		t.Fatal("structured predicate allowed excluded black creature")
	}
}

func TestStructuredAllowPermanentWithoutTypePredicateAllowsAnyPermanent(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	artifact := addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Relic",
		Types: []types.Card{types.Artifact}},
	})
	addBasicLandPermanent(g, game.Player1, types.Forest)
	spellID := addCardToHand(g, game.Player1, permanentTargetSpellWithSpecs([]game.TargetSpec{
		{MinTargets: 1, MaxTargets: 1, Allow: game.TargetAllowPermanent},
	}))
	g.Turn.Phase = game.PhasePrecombatMain
	g.Turn.Step = game.StepNone

	legal := engine.legalActions(g, game.Player1)

	if !containsAction(legal, action.CastSpell(spellID, []game.Target{game.PermanentTarget(artifact.ObjectID)}, 0, nil)) {
		t.Fatal("structured permanent target without type predicate did not allow artifact")
	}
}

func TestStructuredAnyTargetAllowsOnlyDamageablePermanents(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	artifact := addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Relic",
		Types: []types.Card{types.Artifact}},
	})
	creature := addCreaturePermanent(g, game.Player1)
	addBasicLandPermanent(g, game.Player1, types.Forest)
	spellID := addCardToHand(g, game.Player1, permanentTargetSpellWithSpecs([]game.TargetSpec{
		{MinTargets: 1, MaxTargets: 1, Allow: game.TargetAllowPermanent | game.TargetAllowPlayer},
	}))
	g.Turn.Phase = game.PhasePrecombatMain
	g.Turn.Step = game.StepNone

	legal := engine.legalActions(g, game.Player1)

	if containsAction(legal, action.CastSpell(spellID, []game.Target{game.PermanentTarget(artifact.ObjectID)}, 0, nil)) {
		t.Fatal("structured any target allowed noncreature artifact")
	}
	if !containsAction(legal, action.CastSpell(spellID, []game.Target{game.PermanentTarget(creature.ObjectID)}, 0, nil)) {
		t.Fatal("structured any target did not allow creature")
	}
	if !containsAction(legal, action.CastSpell(spellID, []game.Target{game.PlayerTarget(game.Player2)}, 0, nil)) {
		t.Fatal("structured any target did not allow player")
	}
}

func TestAnotherTargetPredicateExcludesSourcePermanent(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	source := addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Source Creature",
		Types: []types.Card{types.Creature},
		Abilities: []game.AbilityDef{
			{
				Kind: game.ActivatedAbility,
				Targets: []game.TargetSpec{
					{
						MinTargets: 1,
						MaxTargets: 1,
						Constraint: "another target creature",
						Allow:      game.TargetAllowPermanent,
						Predicate: game.TargetPredicate{
							PermanentTypes: []types.Card{types.Creature},
							Another:        true,
						},
					},
				},
				Effects: []game.Effect{{Type: game.EffectTap, TargetIndex: 0}},
			},
		}},
	})
	other := addCreaturePermanent(g, game.Player1)
	g.Turn.Phase = game.PhasePrecombatMain
	g.Turn.Step = game.StepNone

	legal := engine.legalActions(g, game.Player1)

	if containsAction(legal, action.ActivateAbility(source.ObjectID, 0, []game.Target{game.PermanentTarget(source.ObjectID)}, 0)) {
		t.Fatal("another target predicate allowed source permanent")
	}
	if !containsAction(legal, action.ActivateAbility(source.ObjectID, 0, []game.Target{game.PermanentTarget(other.ObjectID)}, 0)) {
		t.Fatal("another target predicate did not allow other creature")
	}
}

func TestPermanentTargetConstraintsCanRequireOpponentControlledNonland(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	ownCreature := addCreaturePermanent(g, game.Player1)
	opponentCreature := addCreaturePermanent(g, game.Player2)
	opponentLand := addBasicLandPermanent(g, game.Player2, types.Forest)
	addBasicLandPermanent(g, game.Player1, types.Forest)
	spellID := addCardToHand(g, game.Player1, permanentTargetSpell("nonland permanent an opponent controls"))
	g.Turn.Phase = game.PhasePrecombatMain
	g.Turn.Step = game.StepNone

	legal := engine.legalActions(g, game.Player1)

	want := action.CastSpell(spellID, []game.Target{game.PermanentTarget(opponentCreature.ObjectID)}, 0, nil)
	if !containsAction(legal, want) {
		t.Fatalf("legal actions did not include opponent nonland permanent target %+v", want)
	}
	for _, forbidden := range []game.Target{game.PermanentTarget(ownCreature.ObjectID), game.PermanentTarget(opponentLand.ObjectID)} {
		if containsAction(legal, action.CastSpell(spellID, []game.Target{forbidden}, 0, nil)) {
			t.Fatalf("legal actions included forbidden target %+v", forbidden)
		}
	}
}

func TestIllegalPermanentTargetIsRejectedDuringCast(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	land := addBasicLandPermanent(g, game.Player2, types.Forest)
	addBasicLandPermanent(g, game.Player1, types.Forest)
	spellID := addCardToHand(g, game.Player1, permanentTargetSpell("creature"))
	g.Turn.Phase = game.PhasePrecombatMain
	g.Turn.Step = game.StepNone

	if engine.applyAction(g, game.Player1, action.CastSpell(spellID, []game.Target{game.PermanentTarget(land.ObjectID)}, 0, nil)) {
		t.Fatal("cast with illegal permanent target succeeded")
	}
}

func TestPermanentTargetThatLeavesBeforeResolutionCountersSpellByRules(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	target := addCreaturePermanent(g, game.Player2)
	sourceID := addEffectSpellToStack(g, game.Player1, &game.Effect{
		Type:        game.EffectDamage,
		Amount:      3,
		TargetIndex: 0,
	}, []game.Target{game.PermanentTarget(target.ObjectID)})
	card, ok := g.GetCardInstance(sourceID)
	if !ok {
		t.Fatal("source card instance not found")
	}
	card.Def.Abilities[0].Targets = []game.TargetSpec{{MinTargets: 1, MaxTargets: 1, Constraint: "creature"}}
	if !movePermanentToZone(g, target, game.ZoneGraveyard) {
		t.Fatal("movePermanentToZone() = false, want true")
	}
	log := TurnLog{}

	engine.resolveTopOfStack(g, &log)

	if len(log.Resolves) != 1 || log.Resolves[0].Result != "countered by rules" {
		t.Fatalf("resolve log = %+v, want countered by rules", log.Resolves)
	}
	if !g.Players[game.Player1].Graveyard.Contains(sourceID) {
		t.Fatal("countered spell did not move to graveyard")
	}
}

func TestPermanentTargetedDamageMarksDamageOnResolution(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	target := addCreaturePermanent(g, game.Player2)
	sourceID := addEffectSpellToStack(g, game.Player1, &game.Effect{
		Type:        game.EffectDamage,
		Amount:      3,
		TargetIndex: 0,
	}, []game.Target{game.PermanentTarget(target.ObjectID)})
	card, ok := g.GetCardInstance(sourceID)
	if !ok {
		t.Fatal("source card instance not found")
	}
	card.Def.Abilities[0].Targets = []game.TargetSpec{{MinTargets: 1, MaxTargets: 1, Constraint: "creature"}}

	engine.resolveTopOfStack(g, &TurnLog{})

	if target.MarkedDamage != 3 {
		t.Fatalf("marked damage = %d, want 3", target.MarkedDamage)
	}
	if !g.Players[game.Player1].Graveyard.Contains(sourceID) {
		t.Fatal("resolved spell did not move to graveyard")
	}
}

func TestTargetChoiceKindsAtActionEnumerationLevel(t *testing.T) {
	tests := []struct {
		name            string
		setupSpell      func() *game.CardDef
		setupBoard      func(g *game.Game)
		wantCastActions int // number of cast-spell actions for the spell
	}{
		{
			name: "no targets required produces one cast action",
			setupSpell: func() *game.CardDef {
				return &game.CardDef{CardFace: game.CardFace{Name: "Shock No Target",
					Types: []types.Card{types.Sorcery},
					Abilities: []game.AbilityDef{
						{Kind: game.SpellAbility},
					}},
				}
			},
			setupBoard:      func(g *game.Game) {},
			wantCastActions: 1,
		},
		{
			name: "required target with one legal candidate produces one cast action",
			setupSpell: func() *game.CardDef {
				return permanentTargetSpell("creature")
			},
			setupBoard: func(g *game.Game) {
				addCreaturePermanent(g, game.Player2)
				addBasicLandPermanent(g, game.Player1, types.Forest)
			},
			wantCastActions: 1,
		},
		{
			name: "required target with no legal candidates produces no cast actions",
			setupSpell: func() *game.CardDef {
				return permanentTargetSpell("planeswalker")
			},
			setupBoard:      func(g *game.Game) {},
			wantCastActions: 0,
		},
		{
			name: "invalid target spec (min > max) produces no cast actions",
			setupSpell: func() *game.CardDef {
				return permanentTargetSpellWithRange("creature", 3, 1)
			},
			setupBoard: func(g *game.Game) {
				addCreaturePermanent(g, game.Player2)
			},
			wantCastActions: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
			engine := NewEngine(nil)
			spellID := addCardToHand(g, game.Player1, tt.setupSpell())
			tt.setupBoard(g)
			g.Turn.Phase = game.PhasePrecombatMain
			g.Turn.Step = game.StepNone

			legal := engine.legalActions(g, game.Player1)

			var castCount int
			for _, act := range legal {
				if cast, ok := act.CastSpellPayload(); ok && cast.CardID == spellID {
					castCount++
				}
			}
			if castCount != tt.wantCastActions {
				t.Errorf("cast actions = %d, want %d", castCount, tt.wantCastActions)
			}
		})
	}
}

func TestInvalidTargetSpecAbilityProducesNoActivateActions(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	source := addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Broken Ability Source",
		Types: []types.Card{types.Creature},
		Abilities: []game.AbilityDef{
			{
				Kind: game.ActivatedAbility,
				Targets: []game.TargetSpec{
					{MinTargets: 3, MaxTargets: 1, Constraint: "creature"},
				},
				Effects: []game.Effect{{Type: game.EffectTap, TargetIndex: 0}},
			},
		}},
	})
	addCreaturePermanent(g, game.Player2)
	g.Turn.Phase = game.PhasePrecombatMain
	g.Turn.Step = game.StepNone

	legal := engine.legalActions(g, game.Player1)

	for _, act := range legal {
		if activate, ok := act.ActivateAbilityPayload(); ok && activate.SourceID == source.ObjectID {
			t.Fatalf("invalid ability target spec produced activate action: %+v", act)
		}
	}
}

func TestOpponentChosenTargetSlotUsesDeferredLegalAction(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	source := addCombatPermanent(g, game.Player1, opponentChosenTargetAbilitySource())
	own := addCreaturePermanent(g, game.Player1)
	addCreaturePermanent(g, game.Player2)
	addCreaturePermanent(g, game.Player2)
	g.Turn.Phase = game.PhasePrecombatMain
	g.Turn.Step = game.StepNone

	legal := engine.legalActions(g, game.Player1)

	var matching []action.ActivateAbilityAction
	for _, act := range legal {
		activate, ok := act.ActivateAbilityPayload()
		if ok && activate.SourceID == source.ObjectID {
			matching = append(matching, activate)
		}
	}
	if len(matching) != 1 {
		t.Fatalf("activate actions = %d, want 1 canonical action", len(matching))
	}
	if got := matching[0].Targets; len(got) != 2 || got[0] != game.PermanentTarget(own.ObjectID) || got[1].Kind != game.TargetDeferred {
		t.Fatalf("targets = %+v, want own creature plus deferred opponent-chosen slot", got)
	}
}

func TestOpponentChosenTargetSlotIsChosenDuringAnnouncement(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	source := addCombatPermanent(g, game.Player1, opponentChosenTargetAbilitySource())
	own := addCreaturePermanent(g, game.Player1)
	placeholder := addCreaturePermanent(g, game.Player2)
	_ = addCreaturePermanent(g, game.Player3)
	chosen := addCreaturePermanent(g, game.Player3)
	g.Turn.Phase = game.PhasePrecombatMain
	g.Turn.Step = game.StepNone
	agents := [game.NumPlayers]PlayerAgent{
		game.Player1: &choiceOnlyAgent{choices: [][]int{{1}}},
		game.Player3: &choiceOnlyAgent{choices: [][]int{{1}}},
	}
	log := TurnLog{}

	ok := engine.applyActionWithChoices(g, game.Player1, action.ActivateAbility(source.ObjectID, 0, []game.Target{
		game.PermanentTarget(own.ObjectID),
		game.PermanentTarget(placeholder.ObjectID),
	}, 0), agents, &log)

	if !ok {
		t.Fatal("applyActionWithChoices() = false, want true")
	}
	obj, ok := g.Stack.Peek()
	if !ok {
		t.Fatal("activated ability was not put on the stack")
	}
	if got := obj.Targets; len(got) != 2 || got[0] != game.PermanentTarget(own.ObjectID) || got[1] != game.PermanentTarget(chosen.ObjectID) {
		t.Fatalf("stack targets = %+v, want opponent's chosen target %d", got, chosen.ObjectID)
	}
	if len(log.Choices) != 2 || log.Choices[0].Request.Player != game.Player1 || log.Choices[1].Request.Player != game.Player3 {
		t.Fatalf("choice log = %+v, want controller opponent choice then opponent target choice", log.Choices)
	}
}

func TestOpponentChosenTargetSlotFallsBackDeterministically(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	source := addCombatPermanent(g, game.Player1, opponentChosenTargetAbilitySource())
	own := addCreaturePermanent(g, game.Player1)
	fallback := addCreaturePermanent(g, game.Player2)
	addCreaturePermanent(g, game.Player3)
	g.Turn.Phase = game.PhasePrecombatMain
	g.Turn.Step = game.StepNone

	ok := engine.applyActionWithChoices(g, game.Player1, action.ActivateAbility(source.ObjectID, 0, []game.Target{
		game.PermanentTarget(own.ObjectID),
		game.DeferredTarget(),
	}, 0), [game.NumPlayers]PlayerAgent{}, &TurnLog{})

	if !ok {
		t.Fatal("applyActionWithChoices() = false, want true")
	}
	obj, ok := g.Stack.Peek()
	if !ok {
		t.Fatal("activated ability was not put on the stack")
	}
	if got := obj.Targets[1]; got != game.PermanentTarget(fallback.ObjectID) {
		t.Fatalf("opponent-chosen target = %+v, want first Player2 creature %d", got, fallback.ObjectID)
	}
}

func TestOpponentChosenTargetSlotKeepsSourceControllerProtection(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	spec := opponentChosenTargetAbilitySource().Abilities[0].Targets[1]
	source := opponentChosenTargetAbilitySource()
	hexproof := addHexproofPermanent(g, game.Player2)
	normal := addCreaturePermanent(g, game.Player2)

	candidates := targetCandidatesForSpecChosenBy(g, game.Player1, game.Player2, source, 0, &spec)

	if slices.Contains(candidates, game.PermanentTarget(hexproof.ObjectID)) {
		t.Fatal("opponent chooser could choose hexproof creature against source controller")
	}
	if !slices.Contains(candidates, game.PermanentTarget(normal.ObjectID)) {
		t.Fatal("opponent chooser could not choose non-hexproof creature they control")
	}
}

func playerDamageSpell() *game.CardDef {
	return &game.CardDef{CardFace: game.CardFace{Types: []types.Card{types.Sorcery},
		Abilities: []game.AbilityDef{
			{
				Kind: game.SpellAbility,
				Targets: []game.TargetSpec{
					{MinTargets: 1, MaxTargets: 1, Constraint: "player"},
				},
				Effects: []game.Effect{
					{Type: game.EffectDamage, Amount: 3, TargetIndex: 0},
				},
			},
		}},
	}
}

func permanentTargetSpell(constraint string) *game.CardDef {
	return permanentTargetSpellWithRange(constraint, 1, 1)
}

func optionalPermanentTargetSpell(constraint string) *game.CardDef {
	return permanentTargetSpellWithRange(constraint, 0, 1)
}

func permanentTargetSpellWithRange(constraint string, minTargets, maxTargets int) *game.CardDef {
	return permanentTargetSpellWithSpecs([]game.TargetSpec{
		{MinTargets: minTargets, MaxTargets: maxTargets, Constraint: constraint},
	})
}

func permanentTargetSpellWithSpecs(specs []game.TargetSpec) *game.CardDef {
	return &game.CardDef{CardFace: game.CardFace{Name: "Permanent Target Spell",
		ManaCost: greenCost(),
		Types:    []types.Card{types.Sorcery},
		Abilities: []game.AbilityDef{
			{
				Kind:    game.SpellAbility,
				Targets: specs,
			},
		}},
	}
}

func opponentChosenTargetAbilitySource() *game.CardDef {
	return &game.CardDef{CardFace: game.CardFace{Name: "Arena-like Land",
		Types: []types.Card{types.Land},
		Abilities: []game.AbilityDef{
			{
				Kind: game.ActivatedAbility,
				Targets: []game.TargetSpec{
					{
						MinTargets: 1,
						MaxTargets: 1,
						Allow:      game.TargetAllowPermanent,
						Predicate: game.TargetPredicate{
							PermanentTypes: []types.Card{types.Creature},
							Controller:     game.ControllerYou,
						},
					},
					{
						MinTargets: 1,
						MaxTargets: 1,
						Allow:      game.TargetAllowPermanent,
						Predicate: game.TargetPredicate{
							PermanentTypes: []types.Card{types.Creature},
							Controller:     game.ControllerYou,
						},
						Chooser: game.TargetChooserOpponent,
					},
				},
				Effects: []game.Effect{
					{Type: game.EffectTap, TargetIndex: 0},
					{Type: game.EffectTap, TargetIndex: 1},
					{Type: game.EffectFight},
				},
			},
		}},
	}
}
