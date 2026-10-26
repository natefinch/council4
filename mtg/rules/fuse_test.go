package rules

import (
	"slices"
	"testing"

	"github.com/natefinch/council4/mtg/cards"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/id"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// Source review: fuse (CR 702.102a-d) and split cards (CR 709.3, 709.4b-d)
// from the verified CR effective 2026-09-25, with 601.2c, 608.2b and 115.3.
// Scryfall rulings (all 2013-04-15) for Wear // Tear
// 9842734c-1eac-4509-a731-4c22017ae586, Far // Away
// e4b1ef6a-6f05-472e-aaed-1c24dd6605c7 and Protect // Serve
// 2e0cd5a5-3aee-4b32-873a-e5a40bfba3e9: both halves only from hand; pay both
// mana costs and the fused mana value is the total; each half may target the
// same object; the spell doesn't resolve only if all targets are illegal;
// left half first, then right; a fused spell of differently colored halves is
// multicolored; Away targets only a player, who chooses the sacrifice on
// resolution. Copies and free casts of fused spells are part 2 (#3429).

var fuseTestRegistry = cards.NewDefaultRegistry()

func fuseRegisteredCard(t *testing.T, leftName, rightName string) *game.CardDef {
	t.Helper()
	def := fuseTestRegistry.Lookup(leftName)
	if def == nil || def.Layout != game.LayoutSplit || !def.Alternate.Exists || def.Alternate.Val.Name != rightName {
		t.Fatalf("registry lookup %q = %+v, want split card %s // %s", leftName, def, leftName, rightName)
	}
	if !def.CanFuse() {
		t.Fatalf("%s // %s CanFuse = false, want true", leftName, rightName)
	}
	return def
}

func fuseRegisteredByName(t *testing.T, name string) *game.CardDef {
	t.Helper()
	def := fuseTestRegistry.Lookup(name)
	if def == nil {
		t.Fatalf("production %s is missing", name)
	}
	return def
}

func fuseMainPhaseGame() (*game.Game, *Engine) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	g.Turn.Phase = game.PhasePrecombatMain
	g.Turn.Step = game.StepNone
	return g, NewEngine(nil)
}

func fuseAddMana(g *game.Game, playerID game.PlayerID, colors ...mana.Color) {
	for _, c := range colors {
		g.Players[playerID].ManaPool.Add(c, 1)
	}
}

func fusedCastActions(actions []action.Action, cardID id.ID) []action.CastSpellAction {
	var casts []action.CastSpellAction
	for _, act := range actions {
		cast, ok := act.CastSpellPayload()
		if ok && cast.CardID == cardID && cast.Face == game.FaceFused {
			casts = append(casts, cast)
		}
	}
	return casts
}

func fuseCastFaceOffered(actions []action.Action, cardID id.ID, sourceZone zone.Type, face game.FaceIndex) bool {
	for _, act := range actions {
		cast, ok := act.CastSpellPayload()
		if ok && cast.CardID == cardID && cast.Face == face && normalizedCastSourceZone(cast) == sourceZone {
			return true
		}
	}
	return false
}

func fuseOnBattlefield(g *game.Game, objectID id.ID) bool {
	for _, permanent := range g.Battlefield {
		if permanent.ObjectID == objectID {
			return true
		}
	}
	return false
}

type wearTearScene struct {
	g           *game.Game
	engine      *Engine
	cardID      id.ID
	artifact    *game.Permanent
	enchantment *game.Permanent
}

func newWearTearScene(t *testing.T) wearTearScene {
	t.Helper()
	g, engine := fuseMainPhaseGame()
	return wearTearScene{
		g:           g,
		engine:      engine,
		cardID:      addCardToHand(g, game.Player1, fuseRegisteredCard(t, "Wear", "Tear")),
		artifact:    addControlledPermanent(g, game.Player2, types.Artifact),
		enchantment: addControlledPermanent(g, game.Player2, types.Enchantment),
	}
}

func (s wearTearScene) fusedTargets() []game.Target {
	return []game.Target{game.PermanentTarget(s.artifact.ObjectID), game.PermanentTarget(s.enchantment.ObjectID)}
}

func (s wearTearScene) fusedAction() action.Action {
	return action.CastSpellFace(s.cardID, game.FaceFused, s.fusedTargets(), 0, game.FusedModes())
}

func TestFuseWearTearFusedCastDestroysBoth(t *testing.T) {
	s := newWearTearScene(t)
	fuseAddMana(s.g, game.Player1, mana.C, mana.R, mana.W)

	legal := s.engine.legalActions(s.g, game.Player1)
	if !actionsContain(legal, s.fusedAction()) {
		t.Fatalf("legal fused casts = %+v, want Wear // Tear fused at artifact and enchantment", fusedCastActions(legal, s.cardID))
	}
	if got := fusedCastActions(legal, s.cardID); len(got) != 1 {
		t.Fatalf("fused casts = %+v, want exactly the one legal target assignment", got)
	}
	if !s.engine.applyAction(s.g, game.Player1, s.fusedAction()) {
		t.Fatal("fused Wear // Tear cast = false, want true")
	}
	if total := s.g.Players[game.Player1].ManaPool.Total(); total != 0 {
		t.Fatalf("mana pool after fused cast = %d, want the combined {1}{R}{W} paid", total)
	}
	obj, ok := s.g.Stack.Peek()
	if !ok || obj.Face != game.FaceFused || !slices.Equal(obj.ChosenModes, game.FusedModes()) {
		t.Fatalf("stack object = %+v, want one fused spell with locked modes [0 1]", obj)
	}
	if s.g.Stack.Size() != 1 {
		t.Fatalf("stack size = %d, want one fused split spell", s.g.Stack.Size())
	}

	log := &TurnLog{}
	s.engine.resolveTopOfStack(s.g, log)
	if fuseOnBattlefield(s.g, s.artifact.ObjectID) || fuseOnBattlefield(s.g, s.enchantment.ObjectID) {
		t.Fatal("fused Wear // Tear did not destroy both the artifact and the enchantment")
	}
	if !s.g.Players[game.Player1].Graveyard.Contains(s.cardID) || s.g.Players[game.Player1].Graveyard.Size() != 1 {
		t.Fatalf("graveyard = %+v, want Wear // Tear exactly once", s.g.Players[game.Player1].Graveyard)
	}
	if len(log.Resolves) != 1 || log.Resolves[0].Result != "graveyard" {
		t.Fatalf("resolve log = %+v, want one resolved fused spell", log.Resolves)
	}
}

func TestFuseWearTearFusedOptionWithheld(t *testing.T) {
	t.Run("only one half affordable", func(t *testing.T) {
		s := newWearTearScene(t)
		fuseAddMana(s.g, game.Player1, mana.C, mana.R)
		legal := s.engine.legalActions(s.g, game.Player1)
		if !fuseCastFaceOffered(legal, s.cardID, zone.Hand, game.FaceFront) {
			t.Fatalf("legal actions = %+v, want Wear alone", legal)
		}
		if got := fusedCastActions(legal, s.cardID); len(got) != 0 {
			t.Fatalf("fused casts = %+v, want none with only {1}{R}", got)
		}
		if s.engine.applyAction(s.g, game.Player1, s.fusedAction()) {
			t.Fatal("fused cast with only {1}{R} = true, want false")
		}
	})
	t.Run("right half has no legal target", func(t *testing.T) {
		g, engine := fuseMainPhaseGame()
		cardID := addCardToHand(g, game.Player1, fuseRegisteredCard(t, "Wear", "Tear"))
		addControlledPermanent(g, game.Player2, types.Artifact)
		fuseAddMana(g, game.Player1, mana.C, mana.R, mana.W)
		legal := engine.legalActions(g, game.Player1)
		if !fuseCastFaceOffered(legal, cardID, zone.Hand, game.FaceFront) {
			t.Fatalf("legal actions = %+v, want Wear alone", legal)
		}
		if got := fusedCastActions(legal, cardID); len(got) != 0 {
			t.Fatalf("fused casts = %+v, want none without an enchantment", got)
		}
	})
	for _, sourceZone := range []zone.Type{zone.Graveyard, zone.Exile} {
		t.Run("from "+sourceZone.String(), func(t *testing.T) {
			s := newWearTearScene(t)
			player := s.g.Players[game.Player1]
			player.Hand.Remove(s.cardID)
			if sourceZone == zone.Graveyard {
				player.Graveyard.Add(s.cardID)
			} else {
				player.Exile.Add(s.cardID)
			}
			// Even a permission naming every face, including the fused
			// view, cannot fuse outside the hand.
			for _, face := range []game.FaceIndex{game.FaceFront, game.FaceAlternate, game.FaceFused} {
				s.g.RuleEffects = append(s.g.RuleEffects, game.RuleEffect{
					ID:             s.g.IDGen.Next(),
					Kind:           game.RuleEffectCastFromZone,
					Controller:     game.Player1,
					AffectedPlayer: game.PlayerYou,
					CastFromZone:   sourceZone,
					CastFace:       opt.Val(face),
				})
			}
			fuseAddMana(s.g, game.Player1, mana.C, mana.R, mana.W)
			legal := s.engine.legalActions(s.g, game.Player1)
			if !fuseCastFaceOffered(legal, s.cardID, sourceZone, game.FaceFront) ||
				!fuseCastFaceOffered(legal, s.cardID, sourceZone, game.FaceAlternate) {
				t.Fatalf("legal actions = %+v, want Wear and Tear castable alone from %s", legal, sourceZone)
			}
			if got := fusedCastActions(legal, s.cardID); len(got) != 0 {
				t.Fatalf("fused casts from %s = %+v, want none (CR 702.102a)", sourceZone, got)
			}
			fused := action.CastSpellFaceFromZone(s.cardID, sourceZone, game.FaceFused, s.fusedTargets(), 0, game.FusedModes())
			if s.engine.applyAction(s.g, game.Player1, fused) {
				t.Fatalf("fused cast from %s = true, want false", sourceZone)
			}
		})
	}
	t.Run("modes must be the locked left-then-right pair", func(t *testing.T) {
		s := newWearTearScene(t)
		fuseAddMana(s.g, game.Player1, mana.C, mana.R, mana.W)
		for _, modes := range [][]int{nil, {0}, {1}, {1, 0}} {
			if s.engine.canCastSpellFaceFromZoneWithKicker(s.g, game.Player1, s.cardID, zone.Hand, game.FaceFused, s.fusedTargets(), 0, modes, false) {
				t.Fatalf("fused cast with modes %v = legal, want only [0 1]", modes)
			}
		}
	})
}

func TestFuseFarAwayResolvesLeftHalfFirst(t *testing.T) {
	g, engine := fuseMainPhaseGame()
	cardID := addCardToHand(g, game.Player1, fuseRegisteredCard(t, "Far", "Away"))
	bounced := addCombatCreaturePermanent(g, game.Player2)
	remaining := addCombatCreaturePermanent(g, game.Player2)
	fuseAddMana(g, game.Player1, mana.C, mana.U, mana.C, mana.C, mana.B)

	fused := action.CastSpellFace(cardID, game.FaceFused,
		[]game.Target{game.PermanentTarget(bounced.ObjectID), game.PlayerTarget(game.Player2)}, 0, game.FusedModes())
	if !actionsContain(engine.legalActions(g, game.Player1), fused) {
		t.Fatal("fused Far // Away with {1}{U}{2}{B} is not offered")
	}
	if !engine.applyAction(g, game.Player1, fused) {
		t.Fatal("fused Far // Away cast = false, want true")
	}
	if total := g.Players[game.Player1].ManaPool.Total(); total != 0 {
		t.Fatalf("mana pool after fused cast = %d, want both halves' costs paid", total)
	}
	engine.resolveTopOfStack(g, &TurnLog{})

	bouncedCard := bounced.CardInstanceID
	if !g.Players[game.Player2].Hand.Contains(bouncedCard) {
		t.Fatal("Far did not return its target to its owner's hand before Away resolved")
	}
	if fuseOnBattlefield(g, remaining.ObjectID) || !g.Players[game.Player2].Graveyard.Contains(remaining.CardInstanceID) {
		t.Fatal("Away did not make the targeted player sacrifice the remaining creature")
	}
	if !g.Players[game.Player1].Graveyard.Contains(cardID) {
		t.Fatal("Far // Away did not go to its owner's graveyard")
	}
}

func TestFuseWearTearTargetLoss(t *testing.T) {
	t.Run("one half loses its target", func(t *testing.T) {
		s := newWearTearScene(t)
		responseID := addCardToHand(s.g, game.Player1, fuseRegisteredCard(t, "Wear", "Tear"))
		fuseAddMana(s.g, game.Player1, mana.C, mana.R, mana.W)
		if !s.engine.applyAction(s.g, game.Player1, s.fusedAction()) {
			t.Fatal("fused cast = false, want true")
		}
		fusedObj, _ := s.g.Stack.Peek()
		fuseAddMana(s.g, game.Player1, mana.C, mana.R)
		wear := action.CastSpellFace(responseID, game.FaceFront, []game.Target{game.PermanentTarget(s.artifact.ObjectID)}, 0, nil)
		if !s.engine.applyAction(s.g, game.Player1, wear) {
			t.Fatal("responding Wear cast = false, want true")
		}
		log := &TurnLog{}
		s.engine.resolveTopOfStack(s.g, log)
		if fuseOnBattlefield(s.g, s.artifact.ObjectID) {
			t.Fatal("responding Wear did not destroy the artifact")
		}
		s.engine.resolveTopOfStack(s.g, log)
		if fuseOnBattlefield(s.g, s.enchantment.ObjectID) {
			t.Fatal("fused Tear half did not destroy the enchantment after Wear's target became illegal")
		}
		if len(log.Resolves) != 2 || log.Resolves[1].StackObjectID != fusedObj.ID || log.Resolves[1].Result != "graveyard" {
			t.Fatalf("resolve log = %+v, want the fused spell to resolve", log.Resolves)
		}
		if !s.g.Players[game.Player1].Graveyard.Contains(s.cardID) {
			t.Fatal("fused card did not go to the graveyard")
		}
	})
	t.Run("both halves lose their targets", func(t *testing.T) {
		s := newWearTearScene(t)
		responseID := addCardToHand(s.g, game.Player1, fuseRegisteredCard(t, "Wear", "Tear"))
		bystander := addCombatCreaturePermanent(s.g, game.Player2)
		fuseAddMana(s.g, game.Player1, mana.C, mana.R, mana.W)
		if !s.engine.applyAction(s.g, game.Player1, s.fusedAction()) {
			t.Fatal("fused cast = false, want true")
		}
		fusedObj, _ := s.g.Stack.Peek()
		fuseAddMana(s.g, game.Player1, mana.C, mana.R, mana.W)
		response := action.CastSpellFace(responseID, game.FaceFused, s.fusedTargets(), 0, game.FusedModes())
		if !s.engine.applyAction(s.g, game.Player1, response) {
			t.Fatal("responding fused cast = false, want true")
		}
		log := &TurnLog{}
		s.engine.resolveTopOfStack(s.g, log)
		eventsBefore := len(s.g.Events)
		s.engine.resolveTopOfStack(s.g, log)
		if len(log.Resolves) != 2 || log.Resolves[1].StackObjectID != fusedObj.ID || log.Resolves[1].Result != "countered by rules" {
			t.Fatalf("resolve log = %+v, want the first fused spell not to resolve (CR 608.2b)", log.Resolves)
		}
		for _, event := range s.g.Events[eventsBefore:] {
			if event.Kind == game.EventSpellResolved {
				t.Fatalf("event %+v, want no resolution for a spell with only illegal targets", event)
			}
		}
		if !fuseOnBattlefield(s.g, bystander.ObjectID) {
			t.Fatal("unrelated permanent was affected")
		}
		graveyard := s.g.Players[game.Player1].Graveyard
		if !graveyard.Contains(s.cardID) || !graveyard.Contains(responseID) {
			t.Fatal("both Wear // Tear cards should be in their owner's graveyard")
		}
	})
}

func TestFuseProtectServeSameTarget(t *testing.T) {
	g, engine := fuseMainPhaseGame()
	cardID := addCardToHand(g, game.Player1, fuseRegisteredCard(t, "Protect", "Serve"))
	creature := addCombatCreaturePermanentWithPower(g, game.Player2, 5)
	basePower := effectivePower(g, creature)
	baseToughness, _ := effectiveToughness(g, creature)
	fuseAddMana(g, game.Player1, mana.C, mana.C, mana.W, mana.C, mana.U)

	target := game.PermanentTarget(creature.ObjectID)
	fused := action.CastSpellFace(cardID, game.FaceFused, []game.Target{target, target}, 0, game.FusedModes())
	if !actionsContain(engine.legalActions(g, game.Player1), fused) {
		t.Fatal("fused Protect // Serve targeting one creature with both halves is not offered (CR 115.3)")
	}
	if !engine.applyAction(g, game.Player1, fused) {
		t.Fatal("fused Protect // Serve cast = false, want true")
	}
	engine.resolveTopOfStack(g, &TurnLog{})
	if got := effectivePower(g, creature); got != basePower+2-6 {
		t.Fatalf("power = %d, want %d after +2 and -6", got, basePower+2-6)
	}
	if got, _ := effectiveToughness(g, creature); got != baseToughness+4 {
		t.Fatalf("toughness = %d, want %d after +4", got, baseToughness+4)
	}
}

func TestFuseCombinedCharacteristicsOnStack(t *testing.T) {
	t.Run("mana value and colors", func(t *testing.T) {
		for _, fusedCast := range []bool{true, false} {
			s := newWearTearScene(t)
			snareID := addCardToHand(s.g, game.Player2, fuseRegisteredByName(t, "Spell Snare"))
			fuseAddMana(s.g, game.Player1, mana.C, mana.R, mana.W)
			fuseAddMana(s.g, game.Player2, mana.U)
			cast := s.fusedAction()
			if !fusedCast {
				cast = action.CastSpellFace(s.cardID, game.FaceFront, []game.Target{game.PermanentTarget(s.artifact.ObjectID)}, 0, nil)
			}
			if !s.engine.applyAction(s.g, game.Player1, cast) {
				t.Fatalf("cast (fused=%v) = false, want true", fusedCast)
			}
			obj, _ := s.g.Stack.Peek()
			manaValue, ok := stackObjectManaValue(s.g, obj)
			wantManaValue := 2
			wantColors := []color.Color{color.Red}
			if fusedCast {
				wantManaValue = 3
				wantColors = []color.Color{color.White, color.Red}
			}
			if !ok || manaValue != wantManaValue {
				t.Fatalf("fused=%v mana value = %d, %v; want %d", fusedCast, manaValue, ok, wantManaValue)
			}
			if got, _ := stackObjectColors(s.g, obj); !slices.Equal(got, wantColors) {
				t.Fatalf("fused=%v colors = %v, want %v", fusedCast, got, wantColors)
			}
			s.g.Turn.PriorityPlayer = game.Player2
			snared := false
			for _, act := range s.engine.legalActions(s.g, game.Player2) {
				snare, ok := act.CastSpellPayload()
				if ok && snare.CardID == snareID && slices.Contains(snare.Targets, game.StackObjectTarget(obj.ID)) {
					snared = true
				}
			}
			if snared == fusedCast {
				t.Fatalf("fused=%v Spell Snare can target = %v, want %v", fusedCast, snared, !fusedCast)
			}
		}
	})
	t.Run("multicolored cast trigger", func(t *testing.T) {
		for _, fusedCast := range []bool{true, false} {
			s := newWearTearScene(t)
			viashino := addCombatPermanent(s.g, game.Player1, fuseRegisteredByName(t, "Gloryscale Viashino"))
			basePower := effectivePower(s.g, viashino)
			fuseAddMana(s.g, game.Player1, mana.C, mana.R, mana.W)
			cast := s.fusedAction()
			if !fusedCast {
				cast = action.CastSpellFace(s.cardID, game.FaceFront, []game.Target{game.PermanentTarget(s.artifact.ObjectID)}, 0, nil)
			}
			if !s.engine.applyAction(s.g, game.Player1, cast) {
				t.Fatalf("cast (fused=%v) = false, want true", fusedCast)
			}
			triggered := s.engine.putTriggeredAbilitiesOnStack(s.g)
			if triggered != fusedCast {
				t.Fatalf("fused=%v Gloryscale Viashino triggered = %v, want %v", fusedCast, triggered, fusedCast)
			}
			wantStack := 1
			if fusedCast {
				wantStack = 2
			}
			if s.g.Stack.Size() != wantStack {
				t.Fatalf("fused=%v stack size = %d, want %d (one trigger per fused cast)", fusedCast, s.g.Stack.Size(), wantStack)
			}
			for s.g.Stack.Size() > 0 {
				s.engine.resolveTopOfStack(s.g, &TurnLog{})
			}
			wantPower := basePower
			if fusedCast {
				wantPower += 3
			}
			if got := effectivePower(s.g, viashino); got != wantPower {
				t.Fatalf("fused=%v Viashino power = %d, want %d", fusedCast, got, wantPower)
			}
		}
	})
}

func TestFuseCostModifierAppliesOnceToTotal(t *testing.T) {
	for _, tc := range []struct {
		name  string
		pool  []mana.Color
		legal bool
	}{
		{name: "reduced total paid", pool: []mana.Color{mana.U, mana.B, mana.C, mana.C}, legal: true},
		{name: "reduction is not applied per half", pool: []mana.Color{mana.U, mana.B, mana.C}, legal: false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g, engine := fuseMainPhaseGame()
			addCombatPermanent(g, game.Player1, fuseRegisteredByName(t, "Goblin Electromancer"))
			cardID := addCardToHand(g, game.Player1, fuseRegisteredCard(t, "Far", "Away"))
			creature := addCombatCreaturePermanent(g, game.Player2)
			fuseAddMana(g, game.Player1, tc.pool...)
			fused := action.CastSpellFace(cardID, game.FaceFused,
				[]game.Target{game.PermanentTarget(creature.ObjectID), game.PlayerTarget(game.Player2)}, 0, game.FusedModes())
			if got := actionsContain(engine.legalActions(g, game.Player1), fused); got != tc.legal {
				t.Fatalf("fused Far // Away with Goblin Electromancer and %v offered = %v, want %v", tc.pool, got, tc.legal)
			}
			if !tc.legal {
				return
			}
			if !engine.applyAction(g, game.Player1, fused) {
				t.Fatal("fused cast = false, want true")
			}
			if total := g.Players[game.Player1].ManaPool.Total(); total != 0 {
				t.Fatalf("mana pool = %d, want the reduced {2}{U}{B} total paid", total)
			}
		})
	}
}

func TestFuseFusedSpellSurvivesClone(t *testing.T) {
	s := newWearTearScene(t)
	fuseAddMana(s.g, game.Player1, mana.C, mana.R, mana.W)
	if !s.engine.applyAction(s.g, game.Player1, s.fusedAction()) {
		t.Fatal("fused cast = false, want true")
	}
	clone := s.g.Clone()
	original, _ := s.g.Stack.Peek()
	cloned, ok := clone.Stack.Peek()
	if !ok || cloned.Face != game.FaceFused || !slices.Equal(cloned.ChosenModes, original.ChosenModes) ||
		!slices.Equal(cloned.Targets, original.Targets) || !slices.Equal(cloned.TargetCounts, original.TargetCounts) {
		t.Fatalf("cloned stack object = %+v, want fused choice, modes and targets of %+v", cloned, original)
	}

	for _, g := range []*game.Game{s.g, clone} {
		log := &TurnLog{}
		s.engine.resolveTopOfStack(g, log)
		if fuseOnBattlefield(g, s.artifact.ObjectID) || fuseOnBattlefield(g, s.enchantment.ObjectID) {
			t.Fatal("fused spell did not destroy both targets in every game copy")
		}
		if !g.Players[game.Player1].Graveyard.Contains(s.cardID) || len(log.Resolves) != 1 || log.Resolves[0].Result != "graveyard" {
			t.Fatalf("resolve log = %+v, want identical fused resolution", log.Resolves)
		}
	}
}
