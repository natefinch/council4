package rules

import (
	"slices"
	"testing"

	"github.com/natefinch/council4/mtg/game/zone"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/compare"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/id"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

func TestSingleModeContentDoesNotRequireModeChoice(t *testing.T) {
	content := game.Mode{
		Targets: []game.TargetSpec{{MinTargets: 1, MaxTargets: 1, Constraint: "player"}},
	}.Ability()

	choices := modeChoicesForContent(content)
	if len(choices) != 1 || len(choices[0]) != 0 {
		t.Fatalf("mode choices = %+v, want one choice with no selected modes", choices)
	}
	if !modesValidForContent(content, nil) {
		t.Fatal("single-mode content rejected an empty mode choice")
	}
	if modesValidForContent(content, []int{0}) {
		t.Fatal("single-mode content required an explicit mode selection")
	}
}

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
	sourceID := addEffectSpellToStack(g, game.Player1, game.Damage{
		Amount:    game.Fixed(3),
		Recipient: game.AnyTargetDamageRecipient(0),
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
		ManaCost:        opt.Val(cost.Mana{cost.O(4)}),
		Colors:          []color.Color{color.White},
		Types:           []types.Card{types.Creature},
		Power:           opt.Val(game.PT{Value: 2}),
		Toughness:       opt.Val(game.PT{Value: 2}),
		StaticAbilities: []game.StaticAbility{game.FlyingStaticBody}},
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
		ActivatedAbilities: []game.ActivatedAbility{{
			Content: game.Mode{
				Targets: []game.TargetSpec{{
					MinTargets: 1,
					MaxTargets: 1,
					Constraint: "another target creature",
					Allow:      game.TargetAllowPermanent,
					Predicate: game.TargetPredicate{
						PermanentTypes: []types.Card{types.Creature},
						Another:        true,
					},
				}},
				Sequence: []game.Instruction{{Primitive: game.Tap{Object: game.TargetPermanentReference(0)}}},
			}.Ability(),
		}}},
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
	sourceID := addEffectSpellToStack(g, game.Player1, game.Damage{
		Amount:    game.Fixed(3),
		Recipient: game.AnyTargetDamageRecipient(0),
	}, []game.Target{game.PermanentTarget(target.ObjectID)})
	card, ok := g.GetCardInstance(sourceID)
	if !ok {
		t.Fatal("source card instance not found")
	}
	// Set target spec on the spell's content to require a creature target
	card.Def.SpellAbility.Val.Modes[0].Targets = []game.TargetSpec{{MinTargets: 1, MaxTargets: 1, Constraint: "creature"}}
	if !movePermanentToZone(g, target, zone.Graveyard) {
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

func TestCardTargetedSpellCreatesActionsForMatchingGraveyardCards(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	instantID := addCardToGraveyard(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:  "Graveyard Instant",
		Types: []types.Card{types.Instant},
	}})
	addCardToGraveyard(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:  "Graveyard Land",
		Types: []types.Card{types.Land},
	}})
	battlefieldCreature := addCreaturePermanent(g, game.Player1)
	spellID := addCardToHand(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:  "Regrow Spell",
		Types: []types.Card{types.Sorcery},
		SpellAbility: opt.Val(game.Mode{
			Targets: []game.TargetSpec{{
				MinTargets: 1,
				MaxTargets: 1,
				Allow:      game.TargetAllowCard,
				TargetZone: zone.Graveyard,
				Selection:  opt.Val(game.Selection{RequiredTypesAny: []types.Card{types.Instant, types.Sorcery}, Controller: game.ControllerYou}),
			}},
			Sequence: []game.Instruction{{Primitive: game.MoveCard{
				Card:        game.CardReference{Kind: game.CardReferenceTarget},
				FromZone:    zone.Graveyard,
				Destination: zone.Hand,
			}}},
		}.Ability()),
	}})
	g.Turn.Phase = game.PhasePrecombatMain
	g.Turn.Step = game.StepNone

	legal := engine.legalActions(g, game.Player1)

	want := action.CastSpell(spellID, []game.Target{currentCardTarget(t, g, instantID)}, 0, nil)
	if !containsAction(legal, want) {
		t.Fatalf("legal actions did not include graveyard instant target action %+v", want)
	}
	for _, act := range legal {
		cast, ok := act.CastSpellPayload()
		if !ok || cast.CardID != spellID {
			continue
		}
		if len(cast.Targets) != 1 || cast.Targets[0] != currentCardTarget(t, g, instantID) {
			t.Fatalf("unexpected card target %+v; battlefield creature was %+v", cast.Targets, battlefieldCreature)
		}
	}
}

func TestCardTargetedSpellMatchesCardsWithCyclingInGraveyard(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	cyclingID := addCardToGraveyard(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name: "Cycling Card",
		ActivatedAbilities: []game.ActivatedAbility{
			game.CyclingActivatedAbility(cost.Mana{cost.W}),
		},
	}})
	addCardToGraveyard(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name: "Plain Card",
	}})
	spellID := addCardToHand(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:  "Excavation",
		Types: []types.Card{types.Sorcery},
		SpellAbility: opt.Val(game.Mode{
			Targets: []game.TargetSpec{{
				MinTargets: 1,
				MaxTargets: 1,
				Allow:      game.TargetAllowCard,
				TargetZone: zone.Graveyard,
				Selection: opt.Val(game.Selection{
					Keyword:    game.Cycling,
					Controller: game.ControllerYou,
				}),
			}},
			Sequence: []game.Instruction{{Primitive: game.MoveCard{
				Card:        game.CardReference{Kind: game.CardReferenceTarget},
				FromZone:    zone.Graveyard,
				Destination: zone.Hand,
			}}},
		}.Ability()),
	}})
	g.Turn.Phase = game.PhasePrecombatMain
	g.Turn.Step = game.StepNone

	legal := engine.legalActions(g, game.Player1)

	want := action.CastSpell(spellID, []game.Target{currentCardTarget(t, g, cyclingID)}, 0, nil)
	if !containsAction(legal, want) {
		t.Fatalf("legal actions did not include cycling-card target action %+v", want)
	}
	for _, act := range legal {
		cast, ok := act.CastSpellPayload()
		if !ok || cast.CardID != spellID {
			continue
		}
		if len(cast.Targets) != 1 || cast.Targets[0] != currentCardTarget(t, g, cyclingID) {
			t.Fatalf("unexpected cycling card target %+v", cast.Targets)
		}
	}
}

func TestIndexedCardTargetReferencesMoveMultipleTargetCards(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	firstID := addCardToGraveyard(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "First Cycling"}})
	secondID := addCardToGraveyard(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Second Cycling"}})
	sourceID := addInstructionSpellToStackForController(g, game.Player1, []game.Instruction{
		{Primitive: game.MoveCard{
			Card:        game.CardReference{Kind: game.CardReferenceTarget},
			FromZone:    zone.Graveyard,
			Destination: zone.Hand,
		}},
		{Primitive: game.MoveCard{
			Card:        game.CardReference{Kind: game.CardReferenceTarget, TargetIndex: 1},
			FromZone:    zone.Graveyard,
			Destination: zone.Hand,
		}},
	}, []game.Target{currentCardTarget(t, g, firstID), currentCardTarget(t, g, secondID)})
	card, ok := g.GetCardInstance(sourceID)
	if !ok {
		t.Fatal("source card instance not found")
	}
	card.Def.SpellAbility.Val.Modes[0].Targets = []game.TargetSpec{{
		MinTargets: 0,
		MaxTargets: 2,
		Allow:      game.TargetAllowCard,
		TargetZone: zone.Graveyard,
	}}

	engine.resolveTopOfStack(g, &TurnLog{})

	if !g.Players[game.Player1].Hand.Contains(firstID) || !g.Players[game.Player1].Hand.Contains(secondID) {
		t.Fatalf("hand = %+v, want both target cards moved", g.Players[game.Player1].Hand.All())
	}
}

func currentCardTarget(t *testing.T, g *game.Game, cardID id.ID) game.Target {
	t.Helper()
	card, ok := g.GetCardInstance(cardID)
	if !ok {
		t.Fatalf("card %v not found", cardID)
	}
	return game.CardTargetWithZoneVersion(cardID, card.ZoneVersion)
}

func TestCardTargetThatLeavesZoneBeforeResolutionCountersSpellByRules(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	targetID := addCardToGraveyard(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:  "Graveyard Instant",
		Types: []types.Card{types.Instant},
	}})
	sourceID := addEffectSpellToStack(g, game.Player1, game.MoveCard{
		Card:        game.CardReference{Kind: game.CardReferenceTarget},
		FromZone:    zone.Graveyard,
		Destination: zone.Hand,
	}, []game.Target{game.CardTarget(targetID)})
	card, ok := g.GetCardInstance(sourceID)
	if !ok {
		t.Fatal("source card instance not found")
	}
	card.Def.SpellAbility.Val.Modes[0].Targets = []game.TargetSpec{{
		MinTargets: 1,
		MaxTargets: 1,
		Allow:      game.TargetAllowCard,
		TargetZone: zone.Graveyard,
		Selection:  opt.Val(game.Selection{RequiredTypes: []types.Card{types.Instant}, Controller: game.ControllerYou}),
	}}
	g.Players[game.Player1].Graveyard.Remove(targetID)
	g.Players[game.Player1].Exile.Add(targetID)
	log := TurnLog{}

	engine.resolveTopOfStack(g, &log)

	if len(log.Resolves) != 1 || log.Resolves[0].Result != "countered by rules" {
		t.Fatalf("resolve log = %+v, want countered by rules", log.Resolves)
	}
	if !g.Players[game.Player1].Exile.Contains(targetID) {
		t.Fatal("target card left exile")
	}
}

func TestCardTargetThatLeavesAndReturnsBeforeResolutionCountersSpellByRules(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	targetID := addCardToGraveyard(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:  "Graveyard Instant",
		Types: []types.Card{types.Instant},
	}})
	sourceID := addEffectSpellToStack(g, game.Player1, game.MoveCard{
		Card:        game.CardReference{Kind: game.CardReferenceTarget},
		FromZone:    zone.Graveyard,
		Destination: zone.Hand,
	}, []game.Target{currentCardTarget(t, g, targetID)})
	card, ok := g.GetCardInstance(sourceID)
	if !ok {
		t.Fatal("source card instance not found")
	}
	card.Def.SpellAbility.Val.Modes[0].Targets = []game.TargetSpec{{
		MinTargets: 1,
		MaxTargets: 1,
		Allow:      game.TargetAllowCard,
		TargetZone: zone.Graveyard,
		Selection:  opt.Val(game.Selection{RequiredTypes: []types.Card{types.Instant}, Controller: game.ControllerYou}),
	}}
	if !moveCardBetweenZones(g, game.Player1, targetID, zone.Graveyard, zone.Exile) {
		t.Fatal("failed to move target to exile")
	}
	if !moveCardBetweenZones(g, game.Player1, targetID, zone.Exile, zone.Graveyard) {
		t.Fatal("failed to return target to graveyard")
	}
	log := TurnLog{}

	engine.resolveTopOfStack(g, &log)

	if len(log.Resolves) != 1 || log.Resolves[0].Result != "countered by rules" {
		t.Fatalf("resolve log = %+v, want countered by rules", log.Resolves)
	}
	if !g.Players[game.Player1].Graveyard.Contains(targetID) {
		t.Fatal("target card left graveyard")
	}
}

func TestMoveCardCanPutTargetOnBottomOfLibrary(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	targetID := addCardToGraveyard(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:  "Graveyard Card",
		Types: []types.Card{types.Instant},
	}})
	topID := addCardToLibrary(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Top Card"}})
	sourceID := addEffectSpellToStack(g, game.Player1, game.MoveCard{
		Card:              game.CardReference{Kind: game.CardReferenceTarget},
		FromZone:          zone.Graveyard,
		Destination:       zone.Library,
		DestinationBottom: true,
	}, []game.Target{currentCardTarget(t, g, targetID)})
	card, ok := g.GetCardInstance(sourceID)
	if !ok {
		t.Fatal("source card instance not found")
	}
	card.Def.SpellAbility.Val.Modes[0].Targets = []game.TargetSpec{{
		MinTargets: 1,
		MaxTargets: 1,
		Allow:      game.TargetAllowCard,
		TargetZone: zone.Graveyard,
	}}

	engine.resolveTopOfStack(g, &TurnLog{})

	if top, ok := g.Players[game.Player1].Library.Top(); !ok || top != topID {
		t.Fatalf("library top = %v, %v; want existing top %v", top, ok, topID)
	}
	if bottom, ok := g.Players[game.Player1].Library.Bottom(); !ok || bottom != targetID {
		t.Fatalf("library bottom = %v, %v; want target %v", bottom, ok, targetID)
	}
}

func TestPutOnBattlefieldCanUseTargetedGraveyardCard(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	targetID := addCardToGraveyard(g, game.Player2, &game.CardDef{CardFace: game.CardFace{
		Name:      "Opponent Graveyard Creature",
		Types:     []types.Card{types.Creature},
		Power:     opt.Val(game.PT{Value: 2}),
		Toughness: opt.Val(game.PT{Value: 2}),
	}})
	sourceID := addEffectSpellToStack(g, game.Player1, game.PutOnBattlefield{
		Source:    game.CardBattlefieldSource(game.CardReference{Kind: game.CardReferenceTarget}),
		Recipient: opt.Val(game.ControllerReference()),
	}, []game.Target{currentCardTarget(t, g, targetID)})
	card, ok := g.GetCardInstance(sourceID)
	if !ok {
		t.Fatal("source card instance not found")
	}
	card.Def.SpellAbility.Val.Modes[0].Targets = []game.TargetSpec{{
		MinTargets: 1,
		MaxTargets: 1,
		Allow:      game.TargetAllowCard,
		TargetZone: zone.Graveyard,
		Selection:  opt.Val(game.Selection{RequiredTypes: []types.Card{types.Creature}}),
	}}

	engine.resolveTopOfStack(g, &TurnLog{})

	if g.Players[game.Player2].Graveyard.Contains(targetID) {
		t.Fatal("target card remained in graveyard")
	}
	permanent := permanentByCardID(g, targetID)
	if permanent == nil {
		t.Fatal("target card was not put onto the battlefield")
	}
	if permanent.Controller != game.Player1 {
		t.Fatalf("permanent controller = %v, want Player1", permanent.Controller)
	}
}

func TestPutOnBattlefieldEntryOptionsAreAtomic(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	sourceID := addCardToGraveyard(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:      "Recursive Construct",
		Types:     []types.Card{types.Artifact, types.Creature},
		Power:     opt.Val(game.PT{Value: 2}),
		Toughness: opt.Val(game.PT{Value: 2}),
	}})
	g.Stack.Push(&game.StackObject{
		ID:           g.IDGen.Next(),
		Kind:         game.StackActivatedAbility,
		SourceID:     sourceID,
		SourceCardID: sourceID,
		Controller:   game.Player1,
		InlineActivated: &game.ActivatedAbility{Content: game.Mode{Sequence: []game.Instruction{{Primitive: game.PutOnBattlefield{
			Source:        game.CardBattlefieldSource(game.CardReference{Kind: game.CardReferenceSource}),
			EntryTapped:   true,
			EntryCounters: []game.CounterPlacement{{Kind: counter.PlusOnePlusOne, Amount: 2}},
		}}}}.Ability()},
	})

	engine.resolveTopOfStack(g, &TurnLog{})

	permanent := permanentByCardID(g, sourceID)
	if permanent == nil {
		t.Fatal("returned card not on battlefield")
	}
	if !permanent.Tapped {
		t.Fatal("returned permanent did not enter tapped")
	}
	if got := permanent.Counters.Get(counter.PlusOnePlusOne); got != 2 {
		t.Fatalf("+1/+1 counters = %d, want 2", got)
	}
	for _, event := range g.Events {
		if event.Kind == game.EventPermanentTapped || event.Kind == game.EventCountersAdded {
			t.Fatalf("entry option emitted follow-up event: %+v", event)
		}
	}
}

func TestSourceCardReferenceRequiresSameGraveyardIncarnation(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	sourceID := addCardToGraveyard(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:      "Recursive Construct",
		Types:     []types.Card{types.Artifact, types.Creature},
		Power:     opt.Val(game.PT{Value: 2}),
		Toughness: opt.Val(game.PT{Value: 2}),
	}})
	source, ok := g.GetCardInstance(sourceID)
	if !ok {
		t.Fatal("source card instance not found")
	}
	g.Stack.Push(&game.StackObject{
		ID:                g.IDGen.Next(),
		Kind:              game.StackActivatedAbility,
		SourceID:          sourceID,
		SourceCardID:      sourceID,
		SourceZone:        zone.Graveyard,
		SourceZoneVersion: source.ZoneVersion,
		Controller:        game.Player1,
		InlineActivated: &game.ActivatedAbility{Content: game.Mode{Sequence: []game.Instruction{{Primitive: game.PutOnBattlefield{
			Source: game.CardBattlefieldSource(game.CardReference{Kind: game.CardReferenceSource}),
		}}}}.Ability()},
	})
	if !moveCardBetweenZones(g, game.Player1, sourceID, zone.Graveyard, zone.Exile) {
		t.Fatal("failed to move source to exile")
	}
	if !moveCardBetweenZones(g, game.Player1, sourceID, zone.Exile, zone.Graveyard) {
		t.Fatal("failed to return source to graveyard")
	}
	log := TurnLog{}

	engine.resolveTopOfStack(g, &log)

	if permanentByCardID(g, sourceID) != nil {
		t.Fatal("stale source incarnation returned to battlefield")
	}
	if !g.Players[game.Player1].Graveyard.Contains(sourceID) {
		t.Fatal("source card left graveyard")
	}
}

func TestSpellResolvesForRemainingLegalTargetWithoutShiftingSlots(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	shrouded := addCombatCreaturePermanentWithPower(g, game.Player2, 2)
	legal := addCombatCreaturePermanentWithPower(g, game.Player3, 2)
	sourceID := addInstructionSpellToStackForController(g, game.Player1, []game.Instruction{
		{Primitive: game.Damage{Amount: game.Fixed(2), Recipient: game.AnyTargetDamageRecipient(0)}},
		{Primitive: game.Damage{Amount: game.Fixed(3), Recipient: game.AnyTargetDamageRecipient(1)}},
	}, []game.Target{
		game.PermanentTarget(shrouded.ObjectID),
		game.PermanentTarget(legal.ObjectID),
	})
	card, ok := g.GetCardInstance(sourceID)
	if !ok {
		t.Fatal("source card instance not found")
	}
	card.Def.SpellAbility.Val.Modes[0].Targets = []game.TargetSpec{
		{MinTargets: 1, MaxTargets: 1, Constraint: "creature"},
		{MinTargets: 1, MaxTargets: 1, Constraint: "creature"},
	}
	obj, ok := g.Stack.Peek()
	if !ok {
		t.Fatal("spell was not put on stack")
	}
	obj.TargetCounts = []int{1, 1}
	addShroudGranter(g, game.Player2)
	log := TurnLog{}

	engine.resolveTopOfStack(g, &log)

	if len(log.Resolves) != 1 || log.Resolves[0].Result != "graveyard" {
		t.Fatalf("resolve log = %+v, want spell resolved to graveyard", log.Resolves)
	}
	if shrouded.MarkedDamage != 0 {
		t.Fatalf("shrouded target marked damage = %d, want 0", shrouded.MarkedDamage)
	}
	if legal.MarkedDamage != 3 {
		t.Fatalf("legal target marked damage = %d, want 3", legal.MarkedDamage)
	}
}

func TestSpellIsCounteredWhenAllTargetsGainShroud(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	first := addCombatCreaturePermanentWithPower(g, game.Player2, 2)
	second := addCombatCreaturePermanentWithPower(g, game.Player2, 2)
	sourceID := addInstructionSpellToStackForController(g, game.Player1, []game.Instruction{
		{Primitive: game.Damage{Amount: game.Fixed(2), Recipient: game.AnyTargetDamageRecipient(0)}},
		{Primitive: game.Damage{Amount: game.Fixed(3), Recipient: game.AnyTargetDamageRecipient(1)}},
	}, []game.Target{
		game.PermanentTarget(first.ObjectID),
		game.PermanentTarget(second.ObjectID),
	})
	card, ok := g.GetCardInstance(sourceID)
	if !ok {
		t.Fatal("source card instance not found")
	}
	card.Def.SpellAbility.Val.Modes[0].Targets = []game.TargetSpec{
		{MinTargets: 1, MaxTargets: 1, Constraint: "creature"},
		{MinTargets: 1, MaxTargets: 1, Constraint: "creature"},
	}
	obj, ok := g.Stack.Peek()
	if !ok {
		t.Fatal("spell was not put on stack")
	}
	obj.TargetCounts = []int{1, 1}
	addShroudGranter(g, game.Player2)
	log := TurnLog{}

	engine.resolveTopOfStack(g, &log)

	if len(log.Resolves) != 1 || log.Resolves[0].Result != "countered by rules" {
		t.Fatalf("resolve log = %+v, want countered by rules", log.Resolves)
	}
	if first.MarkedDamage != 0 || second.MarkedDamage != 0 {
		t.Fatalf("marked damage = %d, %d; want no damage", first.MarkedDamage, second.MarkedDamage)
	}
}

func TestPermanentTargetedDamageMarksDamageOnResolution(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	target := addCreaturePermanent(g, game.Player2)
	sourceID := addEffectSpellToStack(g, game.Player1, game.Damage{
		Amount:    game.Fixed(3),
		Recipient: game.AnyTargetDamageRecipient(0),
	}, []game.Target{game.PermanentTarget(target.ObjectID)})
	card, ok := g.GetCardInstance(sourceID)
	if !ok {
		t.Fatal("source card instance not found")
	}
	// Set target spec on the spell's content to require a creature target
	card.Def.SpellAbility.Val.Modes[0].Targets = []game.TargetSpec{{MinTargets: 1, MaxTargets: 1, Constraint: "creature"}}

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
					Types:        []types.Card{types.Sorcery},
					SpellAbility: opt.Val(game.AbilityContent{})},
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
		ActivatedAbilities: []game.ActivatedAbility{{
			Content: game.Mode{
				Targets:  []game.TargetSpec{{MinTargets: 3, MaxTargets: 1, Constraint: "creature"}},
				Sequence: []game.Instruction{{Primitive: game.Tap{Object: game.TargetPermanentReference(0)}}},
			}.Ability(),
		}}},
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
	source := opponentChosenTargetAbilitySource()
	content := source.ActivatedAbilities[0].Content
	if content.IsModal() || len(content.Modes) != 1 || len(content.Modes[0].Targets) < 2 {
		t.Fatal("source card missing expected ability targets")
	}
	spec := content.Modes[0].Targets[1]
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
		SpellAbility: opt.Val(game.Mode{
			Targets:  []game.TargetSpec{{MinTargets: 1, MaxTargets: 1, Constraint: "player"}},
			Sequence: []game.Instruction{{Primitive: game.Damage{Amount: game.Fixed(3), Recipient: game.AnyTargetDamageRecipient(0)}}},
		}.Ability())},
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
		SpellAbility: opt.Val(game.Mode{
			Targets: specs,
		}.Ability())},
	}
}

func opponentChosenTargetAbilitySource() *game.CardDef {
	return &game.CardDef{CardFace: game.CardFace{Name: "Arena-like Land",
		Types: []types.Card{types.Land},
		ActivatedAbilities: []game.ActivatedAbility{{
			Content: game.Mode{
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
				Sequence: []game.Instruction{
					{Primitive: game.Tap{Object: game.TargetPermanentReference(0)}},
					{Primitive: game.Tap{Object: game.TargetPermanentReference(1)}},
					{Primitive: game.Fight{}},
				},
			}.Ability(),
		}}},
	}
}

func addShroudGranter(g *game.Game, controller game.PlayerID) *game.Permanent {
	return addCombatPermanent(g, controller, &game.CardDef{CardFace: game.CardFace{
		Name:  "Shroud Granter",
		Types: []types.Card{types.Enchantment},
		StaticAbilities: []game.StaticAbility{{
			ContinuousEffects: []game.ContinuousEffect{{
				Layer: game.LayerAbility,
				Group: game.ObjectControlledGroup(
					game.SourcePermanentReference(),
					game.Selection{RequiredTypes: []types.Card{types.Creature}},
				),
				AddKeywords: []game.Keyword{game.Shroud},
			}},
		}},
	}})
}
