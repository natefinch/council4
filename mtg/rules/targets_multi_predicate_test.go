package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game/zone"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/compare"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

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
			result := targetChoicesForSpecs(g, game.Player1, nil, 0, game.Event{}, tt.specs)
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
			Selection: opt.Val(game.Selection{
				RequiredTypesAny: []types.Card{types.Creature},
				ExcludedColors:   []color.Color{color.Black},
				Controller:       game.ControllerOpponent,
				Tapped:           game.TriTrue,
				Keyword:          game.Flying,
				ManaValue:        opt.Val(compare.Int{Op: compare.LessOrEqual, Value: 4}),
			}),
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
					Selection: opt.Val(game.Selection{
						RequiredTypesAny: []types.Card{types.Creature},
						ExcludeSource:    true,
					}),
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
