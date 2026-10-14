package rules

import (
	"slices"
	"testing"

	"github.com/natefinch/council4/mtg/cards/authored"
	cardi "github.com/natefinch/council4/mtg/cards/i"
	cardl "github.com/natefinch/council4/mtg/cards/l"
	cardr "github.com/natefinch/council4/mtg/cards/r"
	cardt "github.com/natefinch/council4/mtg/cards/t"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// Pinned Oracle SHA-256: 6da9ced0e19bd384771ab8f14051cc45a8dc90302082b875fa6e039d2399a1b4.
// Official CR 2026-09-25: 202.3b, 302.6, 613.6/7a/11, 702.10, 707.8a, 712.8e.
// Published Humility/Opalescence rulings: 2006-02-01 and 2009-10-01.
func TestHumilityRemovesThaliaTaxAndRhonasRestrictions(t *testing.T) {
	t.Parallel()
	g, engine := setupBestowMain(t)
	thalia := addCombatPermanent(g, game.Player2, cardt.ThaliaGuardianOfThraben())
	rhonas := addCombatPermanent(g, game.Player1, cardr.RhonasTheIndomitable())
	bolt := addCardToHand(g, game.Player1, cardl.LightningBolt())
	lifeBefore := g.Players[game.Player2].Life
	g.Players[game.Player1].ManaPool.Add(mana.R, 1)
	cast := action.CastSpell(bolt, []game.Target{game.PlayerTarget(game.Player2)}, 0, nil)
	if engine.applyAction(g, game.Player1, cast) || canAttackWith(g, rhonas, game.Player1) || canBlockWith(g, rhonas, game.Player1) {
		t.Fatal("printed Thalia tax or Rhonas restrictions did not function")
	}
	g.Players[game.Player1].ManaPool.Add(mana.C, 1)
	humility := challengeCastEnchantment(t, g, engine, game.Player1, authored.Humility.New())
	if !canAttackWith(g, rhonas, game.Player1) || !canBlockWith(g, rhonas, game.Player1) {
		t.Fatal("Rhonas retained removed attack/block restrictions")
	}
	g.Players[game.Player1].ManaPool = mana.Pool{}
	g.Players[game.Player1].ManaPool.Add(mana.R, 1)
	if !engine.applyAction(g, game.Player1, cast) {
		t.Fatal("Thalia's removed ability still taxed Lightning Bolt")
	}
	engine.resolveTopOfStack(g, &TurnLog{})
	if g.Players[game.Player2].Life != lifeBefore-3 {
		t.Fatal("untaxed Lightning Bolt did not resolve")
	}
	movePermanentToZone(g, humility, zone.Graveyard)
	if canAttackWith(g, rhonas, game.Player1) || canBlockWith(g, rhonas, game.Player1) {
		t.Fatal("Humility departure did not restore Rhonas restrictions")
	}
	bolt = addCardToHand(g, game.Player1, cardl.LightningBolt())
	g.Players[game.Player1].ManaPool.Add(mana.R, 1)
	if engine.applyAction(g, game.Player1, action.CastSpell(bolt, []game.Target{game.PlayerTarget(game.Player2)}, 0, nil)) {
		t.Fatal("Humility departure did not restore Thalia's tax")
	}
	if len(permanentEffectiveAbilitiesView(g, thalia)) == 0 {
		t.Fatal("Thalia's printed abilities failed to return")
	}
}

func TestHumilityAllowsLaterExternallyGrantedRuleAbility(t *testing.T) {
	t.Parallel()
	g, engine := setupBestowMain(t)
	target := addCombatPermanent(g, game.Player1, cardt.ThaliaGuardianOfThraben())
	g.Players[game.Player1].ManaPool.Add(mana.C, 1)
	challengeCastEnchantment(t, g, engine, game.Player1, authored.Humility.New())
	g.Players[game.Player1].ManaPool = mana.Pool{}
	grant := cardt.ThaliaGuardianOfThraben().StaticAbilities[1]
	aura := addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name: "Later tax-granting Aura", Types: []types.Card{types.Enchantment}, Subtypes: []types.Sub{types.Aura},
		StaticAbilities: []game.StaticAbility{{ContinuousEffects: []game.ContinuousEffect{{
			Layer: game.LayerAbility, Group: game.AttachedObjectGroup(game.SourcePermanentReference()),
			AddAbilities: []game.Ability{&grant},
		}}}},
	}})
	aura.AttachedTo = opt.Val(target.ObjectID)
	bolt := addCardToHand(g, game.Player1, cardl.LightningBolt())
	g.Players[game.Player1].ManaPool.Add(mana.R, 1)
	cast := action.CastSpell(bolt, []game.Target{game.PlayerTarget(game.Player2)}, 0, nil)
	if engine.applyAction(g, game.Player1, cast) {
		t.Fatal("a later granted Thalia ability did not produce its rule effect")
	}
	clone := g.Clone()
	clonedAura, _ := permanentByObjectID(clone, aura.ObjectID)
	movePermanentToZone(clone, clonedAura, zone.Graveyard)
	if !engine.applyAction(clone, game.Player1, cast) || engine.applyAction(g, game.Player1, cast) {
		t.Fatal("grant departure or clone isolation failed")
	}
}

func TestOpalescenceBrokenSkyManaValueAndDoubleFacedTokenCopies(t *testing.T) {
	t.Parallel()
	g, _ := setupBestowMain(t)
	engine := choicesProgramEngine(t)
	challengeCastEnchantment(t, g, engine, game.Player1, authored.Humility.New())
	sky := addCombatPermanent(g, game.Player1, cardi.InvasionOfTolvada())
	if !transformPermanent(g, sky) {
		t.Fatal("Invasion of Tolvada could not transform")
	}
	challengeCastEnchantment(t, g, engine, game.Player1, authored.Opalescence.New())
	assertChallengeValues(t, g, sky, true, 5, false)
	addImplementationSpellToStack(g, game.Player1, authored.CacklingCounterpart.New(), []game.Target{game.PermanentTarget(sky.ObjectID)})
	engine.resolveTopOfStack(g, &TurnLog{})
	token := g.Battlefield[len(g.Battlefield)-1]
	if !token.Token || !token.TokenDef.IsTransformingDoubleFaced() || token.Face != game.FaceBack {
		t.Fatal("Cackling Counterpart lost the full double-faced token")
	}
	assertChallengeValues(t, g, token, true, 5, false)
	for _, source := range []*game.Permanent{sky, token} {
		snapshot := g.Clone()
		frozen, _ := permanentByObjectID(snapshot, source.ObjectID)
		t.Run(permanentEffectiveName(g, source), func(t *testing.T) {
			t.Parallel()
			mv, known := effectivePermanentManaValue(snapshot, frozen)
			if !known || mv != 5 {
				t.Fatalf("full transforming object mana value = %d/%v, want 5/true", mv, known)
			}
			clone := snapshot.Clone()
			copied, _ := permanentByObjectID(clone, frozen.ObjectID)
			if !transformPermanent(clone, copied) {
				t.Fatal("cloned transforming object could not turn front up")
			}
			mv, known = effectivePermanentManaValue(clone, copied)
			if !known || mv != 5 || frozen.Face != game.FaceBack {
				t.Fatal("transformation lost mana value or clone isolation")
			}
		})
	}
	// Becoming a copy of only a back face has MV 0, even on a physical DFC.
	for _, def := range []*game.CardDef{challengeEnchantment("Single-faced copy", 7), cardi.InvasionOfTolvada()} {
		holder := addCombatPermanent(g, game.Player2, def)
		becomeStaticCopy(t, g, holder, sky, false)
		mv, known := effectivePermanentManaValue(g, holder)
		values := effectivePermanentValues(g, holder)
		if !known || mv != 0 || values.power != 0 || values.toughness != 0 {
			t.Fatal("an individual back-face copy borrowed the original front mana cost")
		}
	}
	token.FaceDown = true
	if mv, known := effectivePermanentManaValue(g, token); !known || mv != 0 {
		t.Fatal("a face-down double-faced token retained its front mana value")
	}
}

func TestCopiedOpalescenceUsesNewAbilityTimestampAndPhysicalExclusion(t *testing.T) {
	t.Parallel()
	g, engine := setupBestowMain(t)
	holder := addCombatPermanent(g, game.Player1, challengeEnchantment("Old physical enchantment", 6))
	opal := challengeCastEnchantment(t, g, engine, game.Player1, authored.Opalescence.New())
	humility := challengeCastEnchantment(t, g, engine, game.Player1, authored.Humility.New())
	physicalTimestamp := holder.Timestamp()
	becomeStaticCopy(t, g, holder, opal, true)
	assertChallengeValues(t, g, opal, true, 4, false)
	assertChallengeValues(t, g, humility, true, 4, false)
	assertChallengeValues(t, g, holder, true, 1, false)
	if holder.Timestamp() != physicalTimestamp {
		t.Fatal("copying rewrote the physical source timestamp")
	}
	clone := g.Clone()
	for _, original := range []*game.Permanent{opal, humility, holder} {
		copied, _ := permanentByObjectID(clone, original.ObjectID)
		if effectivePower(g, original) != effectivePower(clone, copied) {
			t.Fatal("copied static ability ordering was not clone-safe")
		}
	}
	expireCleanupDurations(g)
	assertChallengeValues(t, g, opal, false, 0, true)
	assertChallengeValues(t, g, humility, true, 1, false)
	clonedOpal, _ := permanentByObjectID(clone, opal.ObjectID)
	assertChallengeValues(t, clone, clonedOpal, true, 4, false)
}

func challengeGrantedActivation(t *testing.T, g *game.Game, permanent *game.Permanent, untap bool) action.Action {
	t.Helper()
	for index, ability := range permanentEffectiveAbilitiesView(g, permanent) {
		if !untap {
			if _, ok := ability.(*game.ManaAbility); ok {
				return action.ActivateAbility(permanent.ObjectID, index, nil, 0)
			}
		} else if body, ok := ability.(*game.ActivatedAbility); ok && slices.ContainsFunc(body.AdditionalCosts, func(c cost.Additional) bool {
			return c.Kind == cost.AdditionalUntap
		}) {
			return action.ActivateAbility(permanent.ObjectID, index, nil, 0)
		}
	}
	t.Fatal("later granted activation missing")
	return action.Pass()
}

func TestHumilityPrintedHastePreservesControlHistoryForAttackTapUntap(t *testing.T) {
	t.Parallel()
	g, engine := setupBestowMain(t)
	humility := challengeCastEnchantment(t, g, engine, game.Player1, authored.Humility.New())
	card := addCardToHand(g, game.Player1, authored.Skullbriar.New())
	g.Players[game.Player1].ManaPool.Add(mana.B, 1)
	g.Players[game.Player1].ManaPool.Add(mana.G, 1)
	if !engine.applyAction(g, game.Player1, action.CastSpell(card, nil, 0, nil)) {
		t.Fatal("Skullbriar could not be cast under Humility")
	}
	engine.resolveTopOfStack(g, &TurnLog{})
	skullbriar := permanentByCardID(g, card)
	if skullbriar == nil || !skullbriar.SummoningSick || canAttackWith(g, skullbriar, game.Player1) {
		t.Fatal("printed haste erased control history despite Humility")
	}
	tap := game.TapManaAbility(mana.R)
	untap := game.ActivatedAbility{
		AdditionalCosts: []cost.Additional{{Kind: cost.AdditionalUntap}},
		Content: game.Mode{Sequence: []game.Instruction{{Primitive: game.GainLife{
			Player: game.ControllerReference(), Amount: game.Fixed(1),
		}}}}.Ability(),
	}
	effectID := g.IDGen.Next()
	g.ContinuousEffects = append(g.ContinuousEffects, game.ContinuousEffect{
		ID: effectID, Timestamp: game.Timestamp(effectID), AffectedObjectID: skullbriar.ObjectID,
		Layer: game.LayerAbility, AddAbilities: []game.Ability{&tap, &untap},
	})
	for _, wantHaste := range []bool{false, true, false, true} {
		humility.PhasedOut = wantHaste
		skullbriar.Tapped = false
		if canAttackWith(g, skullbriar, game.Player1) != wantHaste || !skullbriar.SummoningSick {
			t.Fatal("losing/restoring haste changed control history or attack eligibility")
		}
		g.Players[game.Player1].ManaPool = mana.Pool{}
		bolt := addCardToHand(g, game.Player1, cardl.LightningBolt())
		if engine.applyAction(g, game.Player1, action.CastSpell(bolt, []game.Target{game.PlayerTarget(game.Player2)}, 0, nil)) != wantHaste {
			t.Fatal("automatic mana payment failed to consult current haste")
		}
		if wantHaste {
			engine.resolveTopOfStack(g, &TurnLog{})
		}
		for _, untap := range []bool{false, true} {
			skullbriar.Tapped = untap
			act := challengeGrantedActivation(t, g, skullbriar, untap)
			legal := containsAction(engine.legalActions(g, game.Player1), act)
			if !untap {
				payload, _ := act.ActivateAbilityPayload()
				legal = canActivateManaAbility(g, game.Player1, skullbriar, &tap, payload.AbilityIndex)
			}
			applied := engine.applyAction(g, game.Player1, act)
			if legal != wantHaste || applied != wantHaste {
				t.Fatalf("tap/untap payment: untap=%v haste=%v legal=%v applied=%v action=%+v", untap, wantHaste, legal, applied, act)
			}
			if wantHaste && untap {
				engine.resolveTopOfStack(g, &TurnLog{})
			}
		}
	}
	humility.PhasedOut = false
	clone := g.Clone()
	clonedSkullbriar, _ := permanentByObjectID(clone, skullbriar.ObjectID)
	// A real control exchange must restart history independently of haste.
	other := addCombatCreaturePermanent(g, game.Player2)
	skullbriar.SummoningSick = false
	ctx := &programContext{engine: engine, game: g, active: true, log: &TurnLog{},
		obj: &game.StackObject{Controller: game.Player1, SourceID: skullbriar.ObjectID,
			Targets: []game.Target{game.PermanentTarget(other.ObjectID)}},
		specs: []game.TargetSpec{{MinTargets: 1, MaxTargets: 1, Allow: game.TargetAllowPermanent}},
	}
	outcome, err := ctx.Exchange(game.SourcePermanentReference(), game.TargetPermanentReference(0))
	if err != nil || !outcome.Exchanged || !skullbriar.SummoningSick || canAttackWith(g, skullbriar, game.Player2) {
		t.Fatal("control exchange did not restart sickness under Humility")
	}
	for _, player := range []game.PlayerID{game.Player1, game.Player2} {
		addCardToLibrary(g, player, vanillaCreature("Draw", 2, 2))
		g.Turn.ActivePlayer, g.Turn.PriorityPlayer = player, player
		engine.runBeginningPhase(g, [game.NumPlayers]PlayerAgent{}, &TurnLog{})
		wantSick := player != game.Player2
		if skullbriar.SummoningSick != wantSick || !clonedSkullbriar.SummoningSick {
			t.Fatal("turn start cleared another controller's history or changed a clone")
		}
	}
	if !canAttackWith(g, skullbriar, game.Player2) || hasKeyword(g, skullbriar, game.Haste) {
		t.Fatal("continuously controlled creature could not attack without haste")
	}
}

func TestHumilityRiotHasteGrantCanBeRemovedWithoutErasingHistory(t *testing.T) {
	t.Parallel()
	g, engine := setupBestowMain(t)
	permanent := enterRiotCreature(t, g, engine, riotCreatureDef(), [game.NumPlayers]PlayerAgent{
		game.Player1: &choiceOnlyAgent{choices: [][]int{{1}}},
	})
	if !permanent.SummoningSick || !canAttackWith(g, permanent, game.Player1) {
		t.Fatal("riot did not grant haste independently of history")
	}
	humility := challengeCastEnchantment(t, g, engine, game.Player1, authored.Humility.New())
	if canAttackWith(g, permanent, game.Player1) || !permanent.SummoningSick {
		t.Fatal("Humility could not remove riot's haste restriction bypass")
	}
	movePermanentToZone(g, humility, zone.Graveyard)
	if !hasKeyword(g, permanent, game.Haste) || !canAttackWith(g, permanent, game.Player1) {
		t.Fatal("riot's permanent haste grant was not restored")
	}
}

func TestCurrentHasteKeywordGovernsRiotAndRemovedHastePayments(t *testing.T) {
	t.Parallel()
	for _, granted := range []bool{true, false} {
		t.Run(map[bool]string{true: "riot effect grant", false: "printed haste removed"}[granted], func(t *testing.T) {
			t.Parallel()
			g, engine := setupBestowMain(t)
			def := riotCreatureDef()
			tap := game.TapManaAbility(mana.R)
			untap := game.ActivatedAbility{
				AdditionalCosts: []cost.Additional{{Kind: cost.AdditionalUntap}},
				Content: game.Mode{Sequence: []game.Instruction{{Primitive: game.GainLife{
					Player: game.ControllerReference(), Amount: game.Fixed(1),
				}}}}.Ability(),
			}
			def.ManaAbilities = []game.ManaAbility{tap}
			def.ActivatedAbilities = []game.ActivatedAbility{untap}
			if !granted {
				def.StaticAbilities = []game.StaticAbility{game.HasteStaticBody}
			}
			permanent := enterRiotCreature(t, g, engine, def, [game.NumPlayers]PlayerAgent{
				game.Player1: &choiceOnlyAgent{choices: [][]int{{1}}},
			})
			if !granted {
				effectID := g.IDGen.Next()
				g.ContinuousEffects = append(g.ContinuousEffects, game.ContinuousEffect{
					ID: effectID, Timestamp: game.Timestamp(effectID), AffectedObjectID: permanent.ObjectID,
					Layer: game.LayerAbility, RemoveKeywords: []game.Keyword{game.Haste},
				})
			}
			if hasKeyword(g, permanent, game.Haste) != granted || !permanent.SummoningSick {
				t.Fatal("fixture did not distinguish keyword state from ability bodies")
			}
			bolt := addCardToHand(g, game.Player1, cardl.LightningBolt())
			if engine.applyAction(g, game.Player1, action.CastSpell(bolt, []game.Target{game.PlayerTarget(game.Player2)}, 0, nil)) != granted {
				t.Fatal("automatic mana payment ignored current haste keyword")
			}
			if granted {
				engine.resolveTopOfStack(g, &TurnLog{})
			}
			for _, untapCost := range []bool{false, true} {
				permanent.Tapped = untapCost
				act := challengeGrantedActivation(t, g, permanent, untapCost)
				if applied := engine.applyAction(g, game.Player1, act); applied != granted {
					t.Fatalf("tap/untap payment ignored keyword: untap=%v applied=%v", untapCost, applied)
				}
				if granted && untapCost {
					engine.resolveTopOfStack(g, &TurnLog{})
				}
			}
		})
	}
}
