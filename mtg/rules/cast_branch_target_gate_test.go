package rules

import (
	"reflect"
	"slices"
	"strings"
	"testing"

	"github.com/natefinch/council4/mtg/cardprogram"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/id"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
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
				{Primitive: game.MovePermanent{Object: game.TargetPermanentReference(0), Destination: zone.Hand}},
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
				{Primitive: game.MovePermanent{Object: game.TargetPermanentReference(0), Destination: zone.Hand}},
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
			Sequence: []game.Instruction{{Primitive: game.MovePermanent{Object: game.TargetPermanentReference(0), Destination: zone.Hand}}},
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

// Source review: Coastal Discovery, BFZ 73, oracle f7c84690-8c7c-41a6-b430-a9771030f903,
// printing c1df2411-0b0d-4b93-bf30-a22487b3c6ef. Pinned corpus SHA-256:
// 6da9ced0e19bd384771ab8f14051cc45a8dc90302082b875fa6e039d2399a1b4.
// Verified Scryfall /cards/c1df2411-0b0d-4b93-bf30-a22487b3c6ef and its /rulings
// on 2026-11-19: {3}{U} draws two; Awaken 4—{5}{U} adds a target land you
// control and animates it. The six 2015-08-25 rulings include no draw when
// the sole Awaken target becomes illegal. Official CR effective 2026-11-06,
// https://media.wizards.com/2026/downloads/MagicCompRules%2020261106.txt:
// 118.9a, 601.2b–f, 608.2b, 702.40a, 702.113a–b, 707.10, 707.12.
// Only the chosen alternative adds the target; a stack copy inherits the
// choice without casting. These fixtures isolate the Engine gate; they do
// not implement the Awaken animation or register Coastal Discovery.
func alternativeTargetScene(t *testing.T) (*Engine, *game.Game, *game.CardDef) {
	t.Helper()
	const gatedCost cost.AlternativeIndex = 1
	ref := game.ProgramReference{ID: "test/alternative-cost-target", Version: "1"}
	e := revealHandEngine(t, cardprogram.Program{Reference: ref, Resolve: func(ctx cardprogram.Context) error {
		if _, err := ctx.Execute(game.Draw{Player: game.ControllerReference(), Amount: game.Fixed(2)}); err != nil {
			return err
		}
		if ctx.View().Source().AlternativeCost != gatedCost {
			return nil
		}
		_, err := ctx.Execute(game.AddCounter{Object: game.TargetPermanentReference(0), Amount: game.Fixed(1), CounterKind: counter.PlusOnePlusOne})
		return err
	}})
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	setMainPhasePriority(g, game.Player1)
	for range 4 {
		addCardToLibraryNamed(g, game.Player1, "Draw fixture")
	}
	def := &game.CardDef{CardFace: game.CardFace{
		Name: "Alternative target fixture", Types: []types.Card{types.Sorcery},
		ManaCost:         opt.Val(cost.Mana{cost.O(3), cost.U}),
		AlternativeCosts: []cost.Alternative{{Label: "Alternate", ManaCost: opt.Val(cost.Mana{cost.O(5), cost.U})}},
		SpellAbility: opt.Val(game.Mode{Program: ref, Targets: []game.TargetSpec{{
			MinTargets: 1, MaxTargets: 1, Allow: game.TargetAllowPermanent, AlternativeCost: gatedCost,
			Selection: opt.Val(game.Selection{RequiredTypes: []types.Card{types.Land}, Controller: game.ControllerYou}),
		}}}.Ability()),
	}}
	return e, g, def
}

func TestAlternativeCostTargetAnnouncementAndResolution(t *testing.T) {
	for _, tt := range []struct {
		name        string
		generic     int
		ownLand     bool
		targeted    bool
		destination string
	}{
		{"unaffordable alternative", 3, true, false, "resolve"},
		{"no legal alternative target", 5, false, false, "resolve"},
		{"decline payable alternative", 5, true, false, "resolve"},
		{"pay alternative", 5, true, true, "resolve"},
		{"all targets illegal", 5, true, true, "fizzle"},
		{"counter before resolution", 5, true, true, "counter"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			e, g, def := alternativeTargetScene(t)
			def.AlternativeCosts[0].Condition = cost.AlternativeConditionYourTeamCastSpellThisTurn
			castSurgeSetupSpell(t, e, g, game.Player1)
			e.resolveTopOfStack(g, &TurnLog{})
			card := addCardToHand(g, game.Player1, def)
			opposing := addBasicLandPermanent(g, game.Player2, types.Island)
			opposingTarget := []game.Target{game.PermanentTarget(opposing.ObjectID)}
			var own *game.Permanent
			var ownTarget []game.Target
			if tt.ownLand {
				own = addBasicLandPermanent(g, game.Player1, types.Island)
				own.Tapped = true
				ownTarget = []game.Target{game.PermanentTarget(own.ObjectID)}
			}
			pool := &g.Players[game.Player1].ManaPool
			pool.Add(mana.C, tt.generic)
			pool.Add(mana.U, 1)
			base, alt := false, false
			for _, cast := range castActionsForCard(t, e, g, game.Player1, card) {
				switch {
				case len(cast.Targets) == 0:
					base = true
				case tt.ownLand && slices.Equal(cast.Targets, ownTarget):
					alt = true
				default:
					t.Fatalf("unexpected targets: %v", cast.Targets)
				}
			}
			if !base || alt != (tt.generic == 5 && tt.ownLand) {
				t.Fatalf("legal base/alternative actions = %v/%v", base, alt)
			}
			if e.applyAction(g, game.Player1, action.CastSpell(card, opposingTarget, 0, nil)) ||
				(tt.ownLand && tt.generic == 3 && e.applyAction(g, game.Player1, action.CastSpell(card, ownTarget, 0, nil))) {
				t.Fatal("invalid announcement was accepted")
			}
			var targets []game.Target
			var branch cost.AlternativeIndex
			wantSpent := 4
			if tt.targeted {
				targets, branch, wantSpent = ownTarget, 1, 6
			}
			if !e.applyAction(g, game.Player1, action.CastSpell(card, targets, 0, nil)) {
				t.Fatal("legal announcement was rejected")
			}
			spell, _ := g.Stack.Peek()
			if spell.AlternativeCost != branch || spell.ManaSpentToCast != wantSpent ||
				!slices.Equal(spell.Targets, targets) || !slices.Equal(spell.TargetCounts, []int{len(targets)}) {
				t.Fatalf("wrong paid or target branch: %+v", spell)
			}
			log := &TurnLog{}
			switch tt.destination {
			case "counter":
				if !counterStackObject(g, spell.ID) {
					t.Fatal("could not counter the spell")
				}
			case "fizzle":
				if !movePermanentToZone(g, own, zone.Graveyard) {
					t.Fatal("could not remove the only target")
				}
				e.resolveTopOfStack(g, log)
			default:
				e.resolveTopOfStack(g, log)
			}
			wantHand, wantCounters := 0, 0
			if tt.destination == "resolve" {
				wantHand = 2
				if tt.targeted {
					wantCounters = 1
				}
			}
			if g.Players[game.Player1].Hand.Size() != wantHand || !g.Players[game.Player1].Graveyard.Contains(card) ||
				(own != nil && own.Counters.Get(counter.PlusOnePlusOne) != wantCounters) {
				t.Fatal("wrong resolution, target effect, or destination")
			}
		})
	}
}

func TestAlternativeCostTargetCombinedPaymentAndReplay(t *testing.T) {
	for _, generic := range []int{5, 6} {
		e, g, def := alternativeTargetScene(t)
		// A second cost with the same display label and price doesn't turn on
		// the gate. Buyback also shifts the payment option index away from
		// the declared alternative index.
		def.AlternativeCosts = append(def.AlternativeCosts, def.AlternativeCosts[0])
		def.AlternativeCosts[0].Condition = cost.AlternativeConditionYourTeamCastSpellThisTurn
		castSurgeSetupSpell(t, e, g, game.Player1)
		e.resolveTopOfStack(g, &TurnLog{})
		def.StaticAbilities = []game.StaticAbility{{KeywordAbilities: []game.KeywordAbility{game.BuybackKeyword{
			Cost: cost.Mana{cost.O(1)}, AdditionalCosts: []cost.Additional{{Kind: cost.AdditionalPayLife, Amount: 2}},
		}}}}
		card := addCardToHand(g, game.Player1, def)
		land := addBasicLandPermanent(g, game.Player1, types.Island)
		land.Tapped = true
		g.Players[game.Player1].ManaPool.Add(mana.C, generic)
		g.Players[game.Player1].ManaPool.Add(mana.U, 1)
		initial, life := g.Clone(), g.Players[game.Player1].Life
		cast := action.CastSpell(card, []game.Target{game.PermanentTarget(land.ObjectID)}, 0, nil)
		play := func(g *game.Game, agents [game.NumPlayers]PlayerAgent) *TurnLog {
			t.Helper()
			log := &TurnLog{}
			if !e.applyActionWithChoices(g, game.Player1, cast, agents, log) {
				t.Fatal("targeted alternative cast failed")
			}
			obj, _ := g.Stack.Peek()
			paid := generic == 6
			if obj.AlternativeCost != 1 || obj.BuybackPaid != paid || obj.ManaSpentToCast != generic+1 {
				t.Fatalf("wrong expanded payment option: %+v", obj)
			}
			e.resolveTopOfStack(g, log)
			wantLife := life
			if paid {
				wantLife -= 2
			}
			if g.Players[game.Player1].Life != wantLife || g.Players[game.Player1].Hand.Contains(card) != paid {
				t.Fatal("nonmana payment or buyback destination was lost")
			}
			return log
		}
		paymentChoices := 0
		log := play(g, independentBuybackAgent(t, func(req game.ChoiceRequest) string {
			if req.Prompt == "Choose spell cost" {
				paymentChoices++
				if len(req.Options) != 2 || !strings.Contains(req.Options[1].Label, "Buyback") {
					t.Fatalf("expected only the two targeted alternative payments, got %v", req.Options)
				}
				return req.Options[1].Label
			}
			return ""
		}))
		if (paymentChoices == 1) != (generic == 6) {
			t.Fatal("buyback was not gated by affordability")
		}
		replayed := 0
		replay := play(initial, conspireTestAgent(func(req game.ChoiceRequest) []int {
			if replayed >= len(log.Choices) {
				t.Fatal("unexpected replay choice")
			}
			record := log.Choices[replayed]
			replayed++
			return record.Selected
		}))
		if !reflect.DeepEqual(replay.Choices, log.Choices) {
			t.Fatal("replaying the expanded cost choice diverged")
		}
	}
}

func TestAlternativeCostTargetInvalidDeclaration(t *testing.T) {
	_, _, def := alternativeTargetScene(t)
	if issues := game.ValidateCardDef(def); len(issues) != 0 {
		t.Fatalf("valid alternative target rejected: %v", issues)
	}
	def.SpellAbility.Val.Modes[0].Targets[0].AlternativeCost = 2
	if !slices.ContainsFunc(game.ValidateCardDef(def), func(issue game.CardDefIssue) bool {
		return issue.Code == game.CardDefIssueInvalidTargetSpec && strings.HasSuffix(issue.Path, ".AlternativeCost")
	}) {
		t.Fatal("target referencing a nonexistent alternative cost was admitted")
	}
}

func TestAlternativeCostTargetStackCopies(t *testing.T) {
	for _, route := range []string{"direct", "storm"} {
		t.Run(route, func(t *testing.T) {
			e, g, def := alternativeTargetScene(t)
			def.AlternativeCosts[0].Condition = cost.AlternativeConditionYourTeamCastSpellThisTurn
			if route == "storm" {
				def.StaticAbilities = []game.StaticAbility{{KeywordAbilities: game.SimpleKeywords(game.Storm)}}
			}
			castSurgeSetupSpell(t, e, g, game.Player1)
			e.resolveTopOfStack(g, &TurnLog{})
			card := addCardToHand(g, game.Player1, def)
			first := addBasicLandPermanent(g, game.Player1, types.Island)
			second := addBasicLandPermanent(g, game.Player1, types.Island)
			g.Players[game.Player1].ManaPool.Add(mana.C, 5)
			g.Players[game.Player1].ManaPool.Add(mana.U, 1)
			if !e.applyAction(g, game.Player1, action.CastSpell(card, []game.Target{game.PermanentTarget(first.ObjectID)}, 0, nil)) {
				t.Fatal("targeted original cast failed")
			}
			original, _ := g.Stack.Peek()
			if route == "storm" {
				e.putTriggeredAbilitiesOnStackWithChoices(g, [game.NumPlayers]PlayerAgent{}, &TurnLog{})
			} else {
				copier := &game.CardDef{CardFace: game.CardFace{
					Name: "Copy fixture", Types: []types.Card{types.Instant},
					ManaCost: opt.Val(cost.Mana{}),
					SpellAbility: opt.Val(game.Mode{
						Targets: []game.TargetSpec{{
							Allow: game.TargetAllowStackObject, MinTargets: 1, MaxTargets: 1,
							Predicate: game.TargetPredicate{StackObjectKinds: []game.StackObjectKind{game.StackSpell}},
						}},
						Sequence: []game.Instruction{{Primitive: game.CopyStackObject{
							Object: game.TargetStackObjectReference(0), MayChooseNewTargets: true,
						}}},
					}.Ability()),
				}}
				copierID := addCardToHand(g, game.Player1, copier)
				if !e.applyAction(g, game.Player1, action.CastSpell(copierID, []game.Target{game.StackObjectTarget(original.ID)}, 0, nil)) {
					t.Fatal("failed to cast the copying spell")
				}
			}
			log := &TurnLog{}
			retargeted := false
			e.resolveTopOfStackWithChoices(g, conspireTestAgent(func(req game.ChoiceRequest) []int {
				if req.Prompt == "Choose new targets" {
					for _, candidate := range req.Options {
						if len(candidate.Targets) == 1 && candidate.Targets[0].PermanentID == second.ObjectID {
							retargeted = true
							return []int{candidate.Index}
						}
					}
				}
				return req.DefaultSelection
			}), log)
			copy, _ := g.Stack.Peek()
			if !retargeted || g.Stack.Size() != 2 || copy.ID == original.ID || !copy.Copy ||
				copy.AlternativeCost != 1 || spellCastEvents(g, copy.ID) != 0 || spellCastEvents(g, original.ID) != 1 {
				t.Fatal("copy did not inherit the selected cost branch and offer a new target without being cast")
			}
			for _, branch := range []*game.Game{g, g.Clone()} {
				e.resolveTopOfStack(branch, log)
				a, _ := permanentByObjectID(branch, first.ObjectID)
				b, _ := permanentByObjectID(branch, second.ObjectID)
				if a.Counters.Get(counter.PlusOnePlusOne) != 0 || b.Counters.Get(counter.PlusOnePlusOne) != 1 {
					t.Fatal("copy did not affect just its new target")
				}
				e.resolveTopOfStack(branch, log)
				if a.Counters.Get(counter.PlusOnePlusOne) != 1 || branch.Players[game.Player1].Hand.Size() != 4 {
					t.Fatal("original lost its own target or copied branch did not resolve")
				}
			}
		})
	}
}

func TestAlternativeCostTargetOtherCastRoutes(t *testing.T) {
	for _, route := range []string{"paid exile", "prepared copy", "free physical", "free exile copy"} {
		t.Run(route, func(t *testing.T) {
			e, g, def := alternativeTargetScene(t)
			def.AlternativeCosts[0].Condition = cost.AlternativeConditionYourTeamCastSpellThisTurn
			land := addBasicLandPermanent(g, game.Player1, types.Island)
			targets := []game.Target{game.PermanentTarget(land.ObjectID)}
			g.Players[game.Player1].ManaPool.Add(mana.C, 5)
			g.Players[game.Player1].ManaPool.Add(mana.U, 1)
			wantBranch := cost.AlternativeIndex(1)
			wantCopy := route == "prepared copy" || route == "free exile copy"
			var ok bool
			castPaid := func(cast action.Action) bool {
				t.Helper()
				if containsAction(e.legalActions(g, game.Player1), cast) {
					t.Fatal("paid alternative was offered without its required cast history")
				}
				castSurgeSetupSpell(t, e, g, game.Player1)
				e.resolveTopOfStack(g, &TurnLog{})
				return containsAction(e.legalActions(g, game.Player1), cast) && e.applyAction(g, game.Player1, cast)
			}
			switch route {
			case "paid exile":
				card := addCardToExile(g, game.Player1, def)
				g.RuleEffects = append(g.RuleEffects, castSpellsFromExileEffect(g, game.Player1))
				cast := action.CastSpellFaceFromZone(card, zone.Exile, game.FaceFront, targets, 0, nil)
				ok = castPaid(cast)
			case "prepared copy":
				source := prepareCard()
				source.Alternate = opt.Val(def.CardFace)
				permanent := addCombatPermanent(g, game.Player1, source)
				permanent.Prepared = true
				cast := action.CastSpellFaceFromZone(permanent.CardInstanceID, zone.Battlefield, game.FaceAlternate, targets, 0, nil)
				ok = castPaid(cast) && !permanent.Prepared
			case "free physical", "free exile copy":
				// Even when Surge is available, a required free cast cannot pay it.
				castSurgeSetupSpell(t, e, g, game.Player1)
				e.resolveTopOfStack(g, &TurnLog{})
				wantBranch = 0
				card := addCardToExile(g, game.Player1, def)
				if wantCopy {
					ok = e.castFreeCopyOfCard(g, game.Player1, card, [game.NumPlayers]PlayerAgent{}, &TurnLog{})
				} else {
					ok = e.castFreeOption(g, g.Players[game.Player1], game.Player1, g.CardInstances[card],
						action.CastSpellAction{CardID: card, SourceZone: zone.Exile}, false, [game.NumPlayers]PlayerAgent{}, &TurnLog{})
				}
			}
			if !ok {
				t.Fatal("expected cast route was not applied")
			}
			obj, _ := g.Stack.Peek()
			wantSpent, wantTargets := 6, 1
			if wantBranch == 0 {
				wantSpent, wantTargets = 0, 0
			}
			if obj.AlternativeCost != wantBranch || obj.CastCopy != wantCopy || obj.ManaSpentToCast != wantSpent ||
				len(obj.Targets) != wantTargets || spellCastEvents(g, obj.ID) != 1 {
				t.Fatalf("wrong cast branch, payment, targets, or cast events: %+v", obj)
			}
			e.resolveTopOfStack(g, &TurnLog{})
			if land.Counters.Get(counter.PlusOnePlusOne) != wantTargets || g.Players[game.Player1].Hand.Size() != 2 {
				t.Fatal("wrong effect for selected cost branch")
			}
		})
	}
}
