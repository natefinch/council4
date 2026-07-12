package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/id"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// kickerContingentTargetSpell is an instant whose base effect bounces a creature
// (spec 0, always required) and whose kicked effect deals 2 damage to another
// creature (spec 1, gated on the spell being kicked and distinct from the first
// target). It models Jilt's shape: the second, kicker-only target must
// participate in announcement only when the kicker cost is paid (CR 601.2c).
func kickerContingentTargetSpell() *game.CardDef {
	return &game.CardDef{CardFace: game.CardFace{
		Name:     "Kicker Contingent Target",
		ManaCost: greenCost(),
		Types:    []types.Card{types.Instant},
		StaticAbilities: []game.StaticAbility{{
			KeywordAbilities: []game.KeywordAbility{game.KickerKeyword{Cost: greenCost().Val}},
		}},
		SpellAbility: opt.Val(game.Mode{
			Targets: []game.TargetSpec{
				{MinTargets: 1, MaxTargets: 1, Constraint: "creature"},
				{MinTargets: 1, MaxTargets: 1, Constraint: "creature", DistinctFromPriorTargets: true, Gate: game.TargetGateSpellKicked},
			},
			Sequence: []game.Instruction{
				{Primitive: game.Bounce{Object: game.TargetPermanentReference(0)}},
				{
					Primitive: game.Damage{Recipient: game.AnyTargetDamageRecipient(1), Amount: game.Fixed(2)},
					Condition: opt.Val(game.EffectCondition{Condition: opt.Val(game.Condition{SpellWasKicked: true})}),
				},
			},
		}.Ability()),
	}}
}

// giftContingentTargetSpell is an instant whose base effect bounces a creature
// (spec 0, always required) and whose promised-only effect destroys an artifact
// (spec 1, gated on the gift being promised). It models the "additional target
// if the gift was promised" shape: the promised-only target must participate in
// announcement only when the gift is promised (CR 601.2c, CR 702.171).
func giftContingentTargetSpell() *game.CardDef {
	delivery := game.Mode{Sequence: []game.Instruction{{
		Primitive: game.Draw{Amount: game.Fixed(1), Player: game.GiftRecipientReference()},
	}}}.Ability()
	return &game.CardDef{CardFace: game.CardFace{
		Name:     "Gift Contingent Target",
		ManaCost: greenCost(),
		Types:    []types.Card{types.Instant},
		StaticAbilities: []game.StaticAbility{{
			KeywordAbilities: []game.KeywordAbility{game.GiftKeyword{Delivery: delivery}},
		}},
		SpellAbility: opt.Val(game.Mode{
			Targets: []game.TargetSpec{
				{MinTargets: 1, MaxTargets: 1, Constraint: "creature"},
				{MinTargets: 1, MaxTargets: 1, Constraint: "artifact", Gate: game.TargetGateGiftPromised},
			},
			Sequence: []game.Instruction{
				{Primitive: game.Bounce{Object: game.TargetPermanentReference(0)}},
				{
					Primitive: game.Destroy{Object: game.TargetPermanentReference(1)},
					Condition: opt.Val(game.EffectCondition{Condition: opt.Val(game.Condition{GiftPromised: true})}),
				},
			},
		}.Ability()),
	}}
}

// castActionsForCard returns every cast-spell action the engine enumerates for a
// specific card in hand.
func castActionsForCard(t *testing.T, engine *Engine, g *game.Game, player game.PlayerID, cardID id.ID) []action.CastSpellAction {
	t.Helper()
	var casts []action.CastSpellAction
	for _, act := range engine.legalActions(g, player) {
		if cast, ok := act.CastSpellPayload(); ok && cast.CardID == cardID {
			casts = append(casts, cast)
		}
	}
	return casts
}

// TestKickerContingentTargetAnnouncedOnlyWhenKicked covers the kicker scenarios
// (#2919): the kicker-only target is absent when the spell is not kicked and
// required — with an "another target" distinctness constraint — when it is.
func TestKickerContingentTargetAnnouncedOnlyWhenKicked(t *testing.T) {
	// Scenario 3: with a single creature on the battlefield the unkicked branch
	// is legal with one target and the kicked branch, needing a distinct second
	// creature, has no legal cast.
	t.Run("one creature: unkicked legal, kicked impossible", func(t *testing.T) {
		g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
		engine := NewEngine(nil)
		spellID := addCardToHand(g, game.Player1, kickerContingentTargetSpell())
		addCreaturePermanent(g, game.Player2)
		addBasicLandPermanent(g, game.Player1, types.Forest)
		addBasicLandPermanent(g, game.Player1, types.Forest)
		g.Turn.Phase = game.PhasePrecombatMain
		g.Turn.Step = game.StepNone
		g.Turn.PriorityPlayer = game.Player1

		casts := castActionsForCard(t, engine, g, game.Player1, spellID)
		var unkicked, kicked int
		for _, cast := range casts {
			if cast.KickerPaid {
				kicked++
				continue
			}
			unkicked++
			if len(cast.Targets) != 1 {
				t.Errorf("unkicked cast has %d targets, want 1 (kicker-only target absent)", len(cast.Targets))
			}
		}
		if unkicked == 0 {
			t.Error("no unkicked cast action; spell should be legal with a single creature")
		}
		if kicked != 0 {
			t.Errorf("kicked cast actions = %d, want 0 (no distinct second creature for 'another target')", kicked)
		}
	})

	// Scenario 4: with two creatures the kicked branch requires a second target
	// distinct from the first, and never duplicates the first ("another target").
	t.Run("two creatures: kicked requires a distinct second target", func(t *testing.T) {
		g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
		engine := NewEngine(nil)
		spellID := addCardToHand(g, game.Player1, kickerContingentTargetSpell())
		addCreaturePermanent(g, game.Player2)
		addCreaturePermanent(g, game.Player2)
		addBasicLandPermanent(g, game.Player1, types.Forest)
		addBasicLandPermanent(g, game.Player1, types.Forest)
		g.Turn.Phase = game.PhasePrecombatMain
		g.Turn.Step = game.StepNone
		g.Turn.PriorityPlayer = game.Player1

		casts := castActionsForCard(t, engine, g, game.Player1, spellID)
		var unkicked, kicked int
		for _, cast := range casts {
			if !cast.KickerPaid {
				unkicked++
				if len(cast.Targets) != 1 {
					t.Errorf("unkicked cast has %d targets, want 1", len(cast.Targets))
				}
				continue
			}
			kicked++
			if len(cast.Targets) != 2 {
				t.Fatalf("kicked cast has %d targets, want 2 (base + kicker-only)", len(cast.Targets))
			}
			if cast.Targets[0] == cast.Targets[1] {
				t.Errorf("kicked cast duplicated a target %v; 'another target' must be distinct", cast.Targets[0])
			}
		}
		if unkicked == 0 {
			t.Error("no unkicked cast action")
		}
		if kicked == 0 {
			t.Error("no kicked cast action with two distinct creatures")
		}
	})
}

// TestGiftContingentTargetAnnouncedOnlyWhenPromised covers the Gift scenarios
// (#2928): the promised-only target is absent (and the spell still legal) when
// the gift is not promised, and required and chosen when it is.
func TestGiftContingentTargetAnnouncedOnlyWhenPromised(t *testing.T) {
	// Scenario 1: only a creature is on the battlefield (no artifact). The
	// unpromised branch is legal with one target even though the promised-only
	// artifact target has no legal candidate; the promised branch is impossible.
	t.Run("no artifact: unpromised legal, promised impossible", func(t *testing.T) {
		g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
		engine := NewEngine(nil)
		spellID := addCardToHand(g, game.Player1, giftContingentTargetSpell())
		addCreaturePermanent(g, game.Player2)
		addBasicLandPermanent(g, game.Player1, types.Forest)
		g.Turn.Phase = game.PhasePrecombatMain
		g.Turn.Step = game.StepNone
		g.Turn.PriorityPlayer = game.Player1

		casts := castActionsForCard(t, engine, g, game.Player1, spellID)
		var unpromised, promised int
		for _, cast := range casts {
			if cast.GiftPromised {
				promised++
				continue
			}
			unpromised++
			if len(cast.Targets) != 1 {
				t.Errorf("unpromised cast has %d targets, want 1 (promised-only target absent)", len(cast.Targets))
			}
		}
		if unpromised == 0 {
			t.Error("no unpromised cast action; spell should be legal with no legal promised-only target")
		}
		if promised != 0 {
			t.Errorf("promised cast actions = %d, want 0 (no legal artifact target)", promised)
		}
	})

	// Scenario 2: with a creature and an artifact on the battlefield the promised
	// branch requires and chooses both targets; the unpromised branch chooses
	// only the base target.
	t.Run("artifact present: promised requires the extra target", func(t *testing.T) {
		g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
		engine := NewEngine(nil)
		spellID := addCardToHand(g, game.Player1, giftContingentTargetSpell())
		addCreaturePermanent(g, game.Player2)
		addArtifactPermanent(g, game.Player2)
		addBasicLandPermanent(g, game.Player1, types.Forest)
		g.Turn.Phase = game.PhasePrecombatMain
		g.Turn.Step = game.StepNone
		g.Turn.PriorityPlayer = game.Player1

		casts := castActionsForCard(t, engine, g, game.Player1, spellID)
		var unpromised, promised int
		for _, cast := range casts {
			if !cast.GiftPromised {
				unpromised++
				if len(cast.Targets) != 1 {
					t.Errorf("unpromised cast has %d targets, want 1", len(cast.Targets))
				}
				continue
			}
			promised++
			if len(cast.Targets) != 2 {
				t.Errorf("promised cast has %d targets, want 2 (base + promised-only)", len(cast.Targets))
			}
		}
		if unpromised == 0 {
			t.Error("no unpromised cast action")
		}
		if promised == 0 {
			t.Error("no promised cast action with a legal artifact target")
		}
	})
}

// TestGatedTargetSpecInactiveOnStackObjectBranch proves that a spell copied or
// resolved on the branch that does not have its gated target never treats that
// target spec as active: the stack object's announced specs neutralize the
// inactive spec (min/max zero) and the copy, which inherits the cast branch,
// resolves the same specs. Flipping only the branch field re-activates the spec,
// showing the activity is derived from the cast branch, not carried accidentally.
func TestGatedTargetSpecInactiveOnStackObjectBranch(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	cardID := addCardToHand(g, game.Player1, kickerContingentTargetSpell())

	unkicked := &game.StackObject{
		Kind:       game.StackSpell,
		SourceID:   cardID,
		Controller: game.Player1,
		KickerPaid: false,
	}
	unkickedSpecs, ok := stackObjectTargetSpecs(g, unkicked)
	if !ok {
		t.Fatal("stackObjectTargetSpecs failed for the unkicked spell")
	}
	if len(unkickedSpecs.specs) != 2 {
		t.Fatalf("unkicked specs len = %d, want 2 (both ordinal slots preserved)", len(unkickedSpecs.specs))
	}
	if unkickedSpecs.specs[1].MaxTargets != 0 || unkickedSpecs.specs[1].MinTargets != 0 {
		t.Errorf("unkicked kicker spec = %+v, want neutralized (min/max 0)", unkickedSpecs.specs[1])
	}
	if unkickedSpecs.specs[0].MaxTargets != 1 {
		t.Errorf("unkicked base spec = %+v, want the always-required creature target", unkickedSpecs.specs[0])
	}

	// A copy inherits KickerPaid=false, so it resolves the same neutralized spec.
	copySpecs, ok := stackObjectTargetSpecs(g, &game.StackObject{
		Kind:       game.StackSpell,
		SourceID:   cardID,
		Controller: game.Player1,
		KickerPaid: false,
	})
	if !ok {
		t.Fatal("stackObjectTargetSpecs failed for the copy")
	}
	if copySpecs.specs[1].MaxTargets != 0 {
		t.Errorf("copy kicker spec = %+v, want neutralized like the original unkicked cast", copySpecs.specs[1])
	}

	// Flipping only the branch field re-activates the gated spec, confirming
	// activity comes from the branch and is not accidentally sticky.
	kickedSpecs, ok := stackObjectTargetSpecs(g, &game.StackObject{
		Kind:       game.StackSpell,
		SourceID:   cardID,
		Controller: game.Player1,
		KickerPaid: true,
	})
	if !ok {
		t.Fatal("stackObjectTargetSpecs failed for the kicked spell")
	}
	if kickedSpecs.specs[1].MaxTargets != 1 {
		t.Errorf("kicked kicker spec = %+v, want the active distinct-creature target", kickedSpecs.specs[1])
	}
}

// TestUngatedSpecsUnaffectedByBranch proves that ordinary, ungated target specs
// — every existing card — take the identity path: applyCastBranchToSpecs returns
// the same slice unchanged on every branch, and the target-slot remap is the
// identity, so existing behavior is byte-for-byte preserved.
func TestUngatedSpecsUnaffectedByBranch(t *testing.T) {
	specs := []game.TargetSpec{
		{MinTargets: 1, MaxTargets: 1, Constraint: "creature"},
		{MinTargets: 1, MaxTargets: 2, Constraint: "creature"},
	}
	if specsHaveGate(specs) {
		t.Fatal("ungated specs reported as gated")
	}
	for _, branch := range []game.CastBranch{{}, {Kicked: true}, {GiftPromised: true}, {Kicked: true, GiftPromised: true}} {
		got := applyCastBranchToSpecs(specs, branch)
		if &got[0] != &specs[0] {
			t.Errorf("branch %+v: ungated specs not returned as the identity slice", branch)
		}
		for i := range 3 {
			if slot := gatedTargetSlot(specs, branch, i); slot != i {
				t.Errorf("branch %+v: gatedTargetSlot(%d) = %d, want identity", branch, i, slot)
			}
		}
	}
}

// TestInvalidTargetGateFailsCardValidation covers the fail-closed requirement at
// the game-model boundary: a target spec carrying a gate outside the known enum
// is rejected by card validation, so a gate that cannot be represented keeps its
// card unsupported rather than silently mis-announcing targets.
func TestInvalidTargetGateFailsCardValidation(t *testing.T) {
	def := &game.CardDef{CardFace: game.CardFace{
		Name:  "Broken Gate Spell",
		Types: []types.Card{types.Instant},
		SpellAbility: opt.Val(game.Mode{
			Targets: []game.TargetSpec{
				{MinTargets: 1, MaxTargets: 1, Constraint: "creature", Gate: game.TargetGate(99)},
			},
			Sequence: []game.Instruction{{Primitive: game.Bounce{Object: game.TargetPermanentReference(0)}}},
		}.Ability()),
	}}

	issues := game.ValidateCardDef(def)
	found := false
	for _, issue := range issues {
		if issue.Code == game.CardDefIssueInvalidTargetSpec {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("expected an invalid-target-spec issue for an unknown gate, got %+v", issues)
	}
}
