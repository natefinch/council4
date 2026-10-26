package rules

import (
	"slices"
	"strings"
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/id"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// Source review: fused split spells on copy and free-cast routes (#3429, part
// 2 of #1577), from the verified CR effective 2026-09-25: 702.102a (fuse only
// from the caster's own hand; the choice is made before the card goes on the
// stack), 702.102b-d, 709.3a (only the chosen half is evaluated when deciding
// whether it can be cast), 709.4d (a fused spell has the combined
// characteristics), 707.10 (a copy copies characteristics and all decisions,
// including modes and targets, and isn't cast), 707.10c (any number of the
// copy's targets may be changed), 601.3e and 118.9c (a free cast doesn't change
// the mana cost an effect's limit checks). Scryfall rulings: Wear // Tear
// 9842734c-1eac-4509-a731-4c22017ae586 (2013-04-15) — both halves only from
// hand, fused mana value is the total, and a fuse card cast from hand without
// paying its mana cost may still be fused; Twincast (2021-03-19), Reverberate
// and Swarm Intelligence (2017-07-14) — the copy isn't cast, keeps its modes,
// and may change any number of targets; Electrodominance (2024-04-12) — the
// free spell is cast from hand during resolution; Kari Zev's Expertise
// (2017-04-18, 2017-02-09) — either half or, with fuse, both halves, subject
// to the spell's mana value limit, so a 3-mana-value fused Wear // Tear exceeds
// "mana value 2 or less" while each half alone does not. Storm cannot reach a
// fused spell: CanFuse admits only halves whose statics are fuse or split
// second. No Omniscience-style priority permission to cast from hand without
// paying exists in the Engine; it is not covered here.

type fuseCopyScene struct {
	g            *game.Game
	engine       *Engine
	wearTear     id.ID
	artifacts    [2]*game.Permanent
	enchantments [2]*game.Permanent
}

func newFuseCopyScene(t *testing.T) fuseCopyScene {
	t.Helper()
	g, _ := fuseMainPhaseGame()
	g.Turn.PriorityPlayer = game.Player1
	s := fuseCopyScene{
		g:        g,
		engine:   NewEngineWithPrograms(nil, fuseTestRegistry.Programs()),
		wearTear: addCardToHand(g, game.Player1, fuseRegisteredCard(t, "Wear", "Tear")),
	}
	for i := range 2 {
		s.artifacts[i] = addControlledPermanent(g, game.Player2, types.Artifact)
		s.enchantments[i] = addControlledPermanent(g, game.Player2, types.Enchantment)
	}
	return s
}

func (s fuseCopyScene) castFused(t *testing.T) *game.StackObject {
	t.Helper()
	fuseAddMana(s.g, game.Player1, mana.C, mana.R, mana.W)
	targets := []game.Target{game.PermanentTarget(s.artifacts[0].ObjectID), game.PermanentTarget(s.enchantments[0].ObjectID)}
	if !s.engine.applyAction(s.g, game.Player1, action.CastSpellFace(s.wearTear, game.FaceFused, targets, 0, game.FusedModes())) {
		t.Fatal("fused Wear // Tear cast = false, want true")
	}
	obj, _ := s.g.Stack.Peek()
	return obj
}

func (s fuseCopyScene) destroyed(objects ...*game.Permanent) bool {
	for _, permanent := range objects {
		if fuseOnBattlefield(s.g, permanent.ObjectID) {
			return false
		}
	}
	return true
}

func fuseTargetPair(targets []game.Target) [2]id.ID {
	if len(targets) != 2 {
		return [2]id.ID{}
	}
	return [2]id.ID{targets[0].PermanentID, targets[1].PermanentID}
}

func assertFusedCopy(t *testing.T, g *game.Game, copyObj *game.StackObject, want [2]id.ID) {
	t.Helper()
	if copyObj == nil || !copyObj.Copy || copyObj.Face != game.FaceFused || !slices.Equal(copyObj.ChosenModes, game.FusedModes()) ||
		copyObj.Controller != game.Player1 || fuseTargetPair(copyObj.Targets) != want || !slices.Equal(copyObj.TargetCounts, []int{1, 1}) {
		t.Fatalf("copy = %+v, want a fused copy controlled by Player1 with modes [0 1] and one target per half %v", copyObj, want)
	}
	if manaValue, ok := stackObjectManaValue(g, copyObj); !ok || manaValue != 3 {
		t.Fatalf("copy mana value = %d, %v; want the combined 3", manaValue, ok)
	}
	if colors, _ := stackObjectColors(g, copyObj); !slices.Equal(colors, []color.Color{color.White, color.Red}) {
		t.Fatalf("copy colors = %v, want white and red", colors)
	}
}

func TestFuseCopyKeepsBothHalvesAndRetargetsEachHalf(t *testing.T) {
	for _, copier := range []struct {
		name  string
		color mana.Color
	}{{"Twincast", mana.U}, {"Reverberate", mana.R}} {
		t.Run(copier.name, func(t *testing.T) {
			s := newFuseCopyScene(t)
			original := s.castFused(t)
			copierID := addCardToHand(s.g, game.Player1, fuseRegisteredByName(t, copier.name))
			fuseAddMana(s.g, game.Player1, copier.color, copier.color)
			if !s.engine.applyAction(s.g, game.Player1, action.CastSpell(copierID, []game.Target{game.StackObjectTarget(original.ID)}, 0, nil)) {
				t.Fatalf("%s targeting the fused spell = false, want true", copier.name)
			}

			a, e := s.artifacts, s.enchantments
			want := [2]id.ID{a[1].ObjectID, e[1].ObjectID}
			var offered [][2]id.ID
			agent := &authoredCopyChoiceAgent{choose: func(request game.ChoiceRequest) []int {
				if request.Kind != game.ChoiceTarget {
					t.Fatalf("unexpected choice %+v", request)
				}
				chosen := -1
				for _, option := range request.Options {
					pair := fuseTargetPair(option.Targets)
					offered = append(offered, pair)
					if pair == want {
						chosen = option.Index
					}
				}
				return []int{chosen}
			}}
			s.engine.resolveTopOfStackWithChoices(s.g, [game.NumPlayers]PlayerAgent{game.Player1: agent}, &TurnLog{})
			for _, partial := range [][2]id.ID{{a[1].ObjectID, e[0].ObjectID}, {a[0].ObjectID, e[1].ObjectID}, want} {
				if !slices.Contains(offered, partial) {
					t.Fatalf("retarget options = %v, want each half retargetable independently (missing %v)", offered, partial)
				}
			}
			if s.g.Stack.Size() != 2 {
				t.Fatalf("stack size = %d, want the copy above the original", s.g.Stack.Size())
			}
			copyObj, _ := s.g.Stack.Peek()
			assertFusedCopy(t, s.g, copyObj, want)
			if fuseTargetPair(original.Targets) != [2]id.ID{a[0].ObjectID, e[0].ObjectID} {
				t.Fatalf("original targets = %v, want unchanged", original.Targets)
			}
			if got := countEventsOfKind(s.g.Events, game.EventSpellCast); got != 2 {
				t.Fatalf("spell cast events = %d, want 2 (the copy is not cast)", got)
			}
			if got := countEventsOfKind(s.g.Events, game.EventSpellCopied); got != 1 {
				t.Fatalf("spell copied events = %d, want 1", got)
			}

			s.engine.resolveTopOfStack(s.g, &TurnLog{})
			if !s.destroyed(a[1], e[1]) || s.destroyed(a[0]) || s.destroyed(e[0]) {
				t.Fatal("fused copy did not destroy exactly its new artifact and enchantment targets")
			}
			s.engine.resolveTopOfStack(s.g, &TurnLog{})
			if !s.destroyed(a[0], e[0]) || !s.g.Stack.IsEmpty() {
				t.Fatal("original fused spell did not destroy its own targets")
			}
			graveyard := s.g.Players[game.Player1].Graveyard
			if !graveyard.Contains(s.wearTear) || !graveyard.Contains(copierID) || graveyard.Size() != 2 {
				t.Fatalf("graveyard = %+v, want Wear // Tear once and the copier; the copy ceases to exist", graveyard)
			}
		})
	}
}

func TestFuseCopyKeepTargetsAndCloneReplay(t *testing.T) {
	s := newFuseCopyScene(t)
	original := s.castFused(t)
	copierID := addCardToHand(s.g, game.Player1, fuseRegisteredByName(t, "Twincast"))
	fuseAddMana(s.g, game.Player1, mana.U, mana.U)
	if !s.engine.applyAction(s.g, game.Player1, action.CastSpell(copierID, []game.Target{game.StackObjectTarget(original.ID)}, 0, nil)) {
		t.Fatal("Twincast cast = false, want true")
	}
	a, e := s.artifacts, s.enchantments
	partial := [2]id.ID{a[0].ObjectID, e[1].ObjectID}
	agent := &authoredCopyChoiceAgent{choose: func(request game.ChoiceRequest) []int {
		for _, option := range request.Options {
			if fuseTargetPair(option.Targets) == partial {
				return []int{option.Index}
			}
		}
		t.Fatalf("partial retarget %v missing from %+v", partial, request)
		return nil
	}}
	s.engine.resolveTopOfStackWithChoices(s.g, [game.NumPlayers]PlayerAgent{game.Player1: agent}, &TurnLog{})
	copyObj, _ := s.g.Stack.Peek()
	assertFusedCopy(t, s.g, copyObj, partial)

	clone := s.g.Clone()
	cloned, _ := clone.Stack.Peek()
	assertFusedCopy(t, clone, cloned, partial)
	cloneScene := s
	cloneScene.g = clone
	for _, scene := range []fuseCopyScene{s, cloneScene} {
		log := &TurnLog{}
		scene.engine.resolveTopOfStack(scene.g, log)
		if !scene.destroyed(a[0], e[1]) || scene.destroyed(a[1]) || scene.destroyed(e[0]) {
			t.Fatal("fused copy with one retargeted half resolved differently from its choices")
		}
		scene.engine.resolveTopOfStack(scene.g, log)
		if scene.destroyed(a[1]) || !scene.destroyed(e[0]) || !scene.g.Stack.IsEmpty() {
			t.Fatal("original fused spell must still destroy its legal enchantment after the copy took its artifact")
		}
		if len(log.Resolves) != 2 || log.Resolves[1].Result != "graveyard" || scene.g.Players[game.Player1].Graveyard.Size() != 2 {
			t.Fatalf("resolve log = %+v, want copy then original in every game copy", log.Resolves)
		}
	}
}

func TestFuseSwarmIntelligenceCopyIsNotCast(t *testing.T) {
	s := newFuseCopyScene(t)
	addCombatPermanent(s.g, game.Player1, fuseRegisteredByName(t, "Swarm Intelligence"))
	viashino := addCombatPermanent(s.g, game.Player1, fuseRegisteredByName(t, "Gloryscale Viashino"))
	basePower := effectivePower(s.g, viashino)
	s.castFused(t)

	a, e := s.artifacts, s.enchantments
	want := [2]id.ID{a[1].ObjectID, e[1].ObjectID}
	agent := &authoredCopyChoiceAgent{choose: func(request game.ChoiceRequest) []int {
		switch request.Kind {
		case game.ChoiceMay:
			return []int{1}
		case game.ChoiceOrder:
			order := make([]int, len(request.Options))
			for i := range order {
				order[i] = i
			}
			return order
		case game.ChoiceTarget:
			for _, option := range request.Options {
				if fuseTargetPair(option.Targets) == want {
					return []int{option.Index}
				}
			}
		}
		t.Fatalf("unexpected choice %+v", request)
		return nil
	}}
	agents := [game.NumPlayers]PlayerAgent{game.Player1: agent}
	if !s.engine.putTriggeredAbilitiesOnStackWithChoices(s.g, agents, &TurnLog{}) || s.g.Stack.Size() != 3 {
		t.Fatalf("stack size = %d, want the fused spell plus Swarm Intelligence and Gloryscale Viashino triggers", s.g.Stack.Size())
	}
	var copies []*game.StackObject
	for steps := 0; !s.g.Stack.IsEmpty() && steps < 10; steps++ {
		s.engine.resolveTopOfStackWithChoices(s.g, agents, &TurnLog{})
		if s.engine.putTriggeredAbilitiesOnStackWithChoices(s.g, agents, &TurnLog{}) {
			t.Fatal("a copy of the fused spell triggered a cast ability")
		}
		for _, obj := range s.g.Stack.Objects() {
			if obj.Copy && !slices.Contains(copies, obj) {
				copies = append(copies, obj)
				assertFusedCopy(t, s.g, obj, want)
			}
		}
	}
	if len(copies) != 1 {
		t.Fatalf("copies seen = %d, want one Swarm Intelligence copy", len(copies))
	}
	if !s.destroyed(a[0], a[1], e[0], e[1]) {
		t.Fatal("fused spell and its copy did not destroy all four targets")
	}
	if got := effectivePower(s.g, viashino); got != basePower+3 {
		t.Fatalf("Gloryscale Viashino power = %d, want %d (one trigger for the cast only)", got, basePower+3)
	}
	if countEventsOfKind(s.g.Events, game.EventSpellCast) != 1 || countEventsOfKind(s.g.Events, game.EventSpellCopied) != 1 {
		t.Fatal("want exactly one cast and one copy event")
	}
}

type fuseFreeCastAgent struct {
	t         *testing.T
	pick      string
	castLabel []string
}

func (*fuseFreeCastAgent) ChooseAction(PlayerObservation, []action.Action) action.Action {
	return action.Pass()
}

func (a *fuseFreeCastAgent) ChooseChoice(_ PlayerObservation, request game.ChoiceRequest) []int {
	switch {
	case request.Kind == game.ChoiceMay:
		return []int{1}
	case strings.HasPrefix(request.Prompt, "Choose a spell to cast"):
		for _, option := range request.Options {
			if strings.HasPrefix(option.Label, "Wear") {
				return []int{option.Index}
			}
		}
	case strings.HasPrefix(request.Prompt, "Choose how to cast"):
		for _, option := range request.Options {
			a.castLabel = append(a.castLabel, option.Label)
		}
		for _, option := range request.Options {
			if strings.HasPrefix(option.Label, a.pick) {
				return []int{option.Index}
			}
		}
	}
	a.t.Fatalf("unexpected choice %+v", request)
	return nil
}

func fuseLabelsOffer(labels []string, prefix string) bool {
	return slices.ContainsFunc(labels, func(label string) bool { return strings.HasPrefix(label, prefix) })
}

func TestFuseElectrodominanceFreeCastFromHand(t *testing.T) {
	for _, tc := range []struct {
		x     int
		fused bool
	}{{x: 3, fused: true}, {x: 2, fused: false}} {
		t.Run(map[bool]string{true: "X=3 fused", false: "X=2 halves only"}[tc.fused], func(t *testing.T) {
			s := newFuseCopyScene(t)
			electro := addCardToHand(s.g, game.Player1, fuseRegisteredByName(t, "Electrodominance"))
			for range tc.x {
				fuseAddMana(s.g, game.Player1, mana.C)
			}
			fuseAddMana(s.g, game.Player1, mana.R, mana.R)
			if !s.engine.applyAction(s.g, game.Player1, action.CastSpell(electro, []game.Target{game.PlayerTarget(game.Player2)}, tc.x, nil)) {
				t.Fatalf("Electrodominance X=%d cast = false, want true", tc.x)
			}
			pick := "Wear // Tear (fused)"
			if !tc.fused {
				pick = "Tear"
			}
			agent := &fuseFreeCastAgent{t: t, pick: pick}
			s.engine.resolveTopOfStackWithChoices(s.g, [game.NumPlayers]PlayerAgent{game.Player1: agent}, &TurnLog{})
			if fuseLabelsOffer(agent.castLabel, "Wear // Tear (fused)") != tc.fused ||
				!fuseLabelsOffer(agent.castLabel, "Wear ") || !fuseLabelsOffer(agent.castLabel, "Tear ") {
				t.Fatalf("X=%d free-cast options = %q, want Wear and Tear alone, fused only when X >= 3", tc.x, agent.castLabel)
			}
			obj, ok := s.g.Stack.Peek()
			wantFace := game.FaceFused
			if !tc.fused {
				wantFace = game.FaceAlternate
			}
			if !ok || obj.Copy || obj.Face != wantFace || s.g.Stack.Size() != 1 {
				t.Fatalf("stack = %+v, want the free Wear // Tear cast with face %v", obj, wantFace)
			}
			if s.g.Players[game.Player1].ManaPool.Total() != 0 || countEventsOfKind(s.g.Events, game.EventSpellCast) != 2 {
				t.Fatal("free Wear // Tear was not cast without paying mana")
			}
			if tc.fused {
				if manaValue, _ := stackObjectManaValue(s.g, obj); manaValue != 3 {
					t.Fatalf("free fused mana value = %d, want 3 (a free cast keeps the mana cost)", manaValue)
				}
			}
			wantTargets := 1
			if tc.fused {
				wantTargets = 2
			}
			chosen := map[id.ID]bool{}
			for _, target := range obj.Targets {
				chosen[target.PermanentID] = true
			}
			if len(chosen) != wantTargets {
				t.Fatalf("free spell targets = %v, want %d", obj.Targets, wantTargets)
			}
			s.engine.resolveTopOfStack(s.g, &TurnLog{})
			for _, permanent := range []*game.Permanent{s.artifacts[0], s.artifacts[1], s.enchantments[0], s.enchantments[1]} {
				if s.destroyed(permanent) != chosen[permanent.ObjectID] {
					t.Fatalf("free Wear // Tear destroyed %v = %v, want only its chosen targets %v", permanent.ObjectID, s.destroyed(permanent), obj.Targets)
				}
			}
			if s.g.Players[game.Player2].Life != 40-tc.x {
				t.Fatalf("Player2 life = %d, want %d", s.g.Players[game.Player2].Life, 40-tc.x)
			}
		})
	}
}

func TestFuseKariZevsExpertiseOffersHalvesOnly(t *testing.T) {
	s := newFuseCopyScene(t)
	creature := addCombatCreaturePermanent(s.g, game.Player2)
	kari := addCardToHand(s.g, game.Player1, fuseRegisteredByName(t, "Kari Zev's Expertise"))
	fuseAddMana(s.g, game.Player1, mana.C, mana.C, mana.R, mana.R)
	if !s.engine.applyAction(s.g, game.Player1, action.CastSpell(kari, []game.Target{game.PermanentTarget(creature.ObjectID)}, 0, nil)) {
		t.Fatal("Kari Zev's Expertise cast = false, want true")
	}
	agent := &fuseFreeCastAgent{t: t, pick: "Wear "}
	s.engine.resolveTopOfStackWithChoices(s.g, [game.NumPlayers]PlayerAgent{game.Player1: agent}, &TurnLog{})
	if fuseLabelsOffer(agent.castLabel, "Wear // Tear") || !fuseLabelsOffer(agent.castLabel, "Wear ") || !fuseLabelsOffer(agent.castLabel, "Tear ") {
		t.Fatalf("Kari Zev's Expertise options = %q, want Wear or Tear alone, never fused (mana value 3 > 2)", agent.castLabel)
	}
	obj, ok := s.g.Stack.Peek()
	if !ok || obj.Face != game.FaceFront || s.g.Stack.Size() != 1 {
		t.Fatalf("stack = %+v, want Wear cast alone", obj)
	}
}

func TestFuseFreeCastFusesOnlyFromCastersOwnHand(t *testing.T) {
	g, engine := fuseMainPhaseGame()
	def := fuseRegisteredCard(t, "Wear", "Tear")
	addControlledPermanent(g, game.Player2, types.Artifact)
	addControlledPermanent(g, game.Player2, types.Enchantment)
	own := addCardToHand(g, game.Player1, def)
	other := addCardToHand(g, game.Player2, def)
	faces := func(cardID id.ID, player game.PlayerID, from zone.Type) []game.FaceIndex {
		card, _ := g.GetCardInstance(cardID)
		var got []game.FaceIndex
		for _, option := range engine.freeCastOptionsForCard(g, player, card, from, game.Selection{}, opt.V[int]{}) {
			if !slices.Contains(got, option.cast.Face) {
				got = append(got, option.cast.Face)
			}
		}
		return got
	}
	if got := faces(own, game.Player1, zone.Hand); !slices.Contains(got, game.FaceFused) {
		t.Fatalf("own-hand free cast faces = %v, want fused offered", got)
	}
	for _, tc := range []struct {
		name   string
		cardID id.ID
		player game.PlayerID
		from   zone.Type
	}{
		{"graveyard", own, game.Player1, zone.Graveyard},
		{"exile", own, game.Player1, zone.Exile},
		{"library", own, game.Player1, zone.Library},
		{"another player's hand", other, game.Player1, zone.Hand},
	} {
		got := faces(tc.cardID, tc.player, tc.from)
		if slices.Contains(got, game.FaceFused) || !slices.Contains(got, game.FaceFront) || !slices.Contains(got, game.FaceAlternate) {
			t.Fatalf("%s free cast faces = %v, want Wear and Tear alone, never fused", tc.name, got)
		}
	}
}
