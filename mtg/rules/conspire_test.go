package rules

import (
	"reflect"
	"strings"
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/id"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/mtg/rules/payment"
	"github.com/natefinch/council4/opt"
)

// Source review: Gleeful Sabotage, Oracle 25fe48be-95c4-4011-8c60-4628fb5ecfcb,
// pinned revision 7802c7f8-ca2a-4bac-a579-77a8204aa5d4. It destroys an artifact
// or enchantment and has Conspire. Live Scryfall /cards/7802c7f8-ca2a-4bac-a579-77a8204aa5d4/rulings
// returned an empty list. Official CR (effective 2026-11-06),
// https://media.wizards.com/2026/downloads/MagicCompRules%2020261106.txt:
// 105.2–3, 302.6, 601.2a–i, 603.2–4, 608.2h, 702.78a–b, 707.10/.10c/.12.
// The optional cost taps two of your untapped creatures each sharing a color
// with the spell, without a tap symbol; payment triggers an independent copy
// with optional new targets. Stack copies inherit the paid decision but are
// not cast. A cast card copy makes and pays its own decisions. Fixtures below
// are unregistered and isolate Engine behavior; they do not author the card.

func conspireTestSpell() *game.CardDef {
	def := simpleGainLifeInstant("Conspire test spell")
	def.ManaCost = opt.Val(cost.Mana{cost.O(1), cost.G})
	def.Colors = []color.Color{color.Green}
	def.StaticAbilities = []game.StaticAbility{game.ConspireStaticBody}
	return def
}

func conspireTestAgent(choose func(game.ChoiceRequest) []int) [game.NumPlayers]PlayerAgent {
	return [game.NumPlayers]PlayerAgent{game.Player1: &authoredCopyChoiceAgent{choose: choose}}
}

func chooseConspireCost(req game.ChoiceRequest) []int {
	if req.Prompt == "Choose spell cost" {
		for _, option := range req.Options {
			if strings.Contains(option.Label, "Conspire") {
				return []int{option.Index}
			}
		}
	}
	return req.DefaultSelection
}

func isConspireTrigger(obj *game.StackObject) bool {
	return obj != nil && obj.Kind == game.StackTriggeredAbility && game.BodyHasKeyword(obj.InlineTrigger, game.Conspire)
}

func TestConspirePaymentAndRetargetReplay(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	def := conspireTestSpell()
	def.Types = []types.Card{types.Sorcery}
	def.SpellAbility = opt.Val(game.Mode{
		Targets: []game.TargetSpec{{
			MinTargets: 1, MaxTargets: 1, Allow: game.TargetAllowPermanent,
			Selection: opt.Val(game.Selection{RequiredTypesAny: []types.Card{types.Artifact, types.Enchantment}}),
		}},
		Sequence: []game.Instruction{{Primitive: game.Destroy{
			Object: game.TargetPermanentReference(0),
		}}},
	}.Ability())
	if issues := game.ValidateCardDef(def); len(issues) != 0 {
		t.Fatalf("Conspire fixture rejected: %+v", issues)
	}
	cardID := addCardToHand(g, game.Player1, def)
	artifact := addNamedTypedPermanent(g, game.Player2, "Artifact target", types.Artifact)
	enchantment := addNamedTypedPermanent(g, game.Player2, "Enchantment target", types.Enchantment)
	first := addColoredCreaturePermanent(g, game.Player1, 2, color.Green)
	spare := addColoredCreaturePermanent(g, game.Player1, 2, color.Green)
	second := addColoredCreaturePermanent(g, game.Player1, 2, color.Green, color.Blue)
	first.SummoningSick, second.SummoningSick = true, true
	wrongColor := addColoredCreaturePermanent(g, game.Player1, 2, color.Blue)
	opposing := addColoredCreaturePermanent(g, game.Player2, 2, color.Green)
	preTapped := addColoredCreaturePermanent(g, game.Player1, 2, color.Green)
	preTapped.Tapped = true
	noncreature := addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name: "Green artifact", Types: []types.Card{types.Artifact}, Colors: []color.Color{color.Green},
	}})
	g.Players[game.Player1].ManaPool.Add(mana.G, 1)
	g.Players[game.Player1].ManaPool.Add(mana.C, 1)
	setMainPhasePriority(g, game.Player1)
	cast := action.CastSpell(cardID, []game.Target{game.PermanentTarget(artifact.ObjectID)}, 0, nil)
	initial := g.Clone()
	play := func(g *game.Game, agents [game.NumPlayers]PlayerAgent) *TurnLog {
		t.Helper()
		log := &TurnLog{}
		if !engine.applyActionWithChoices(g, game.Player1, cast, agents, log) {
			t.Fatal("could not cast the targeted spell")
		}
		original, _ := g.Stack.Peek()
		if original.ConspirePaid != 1 || original.ManaSpentToCast != 2 || g.Stack.Size() != 1 {
			t.Fatalf("expected a paid cast with no immediate copy: %+v", original)
		}
		for _, creature := range []*game.Permanent{first, second, spare, wrongColor, opposing, preTapped, noncreature} {
			p, _ := permanentByObjectID(g, creature.ObjectID)
			want := creature == first || creature == second || creature == preTapped
			if p.Tapped != want {
				t.Fatalf("object %d tapped=%v, want %v", p.ObjectID, p.Tapped, want)
			}
		}
		engine.putTriggeredAbilitiesOnStackWithChoices(g, agents, log)
		trigger, _ := g.Stack.Peek()
		if !isConspireTrigger(trigger) || trigger.TriggerEvent.StackObjectID != original.ID {
			t.Fatal("paid Conspire did not make its own self-cast trigger")
		}
		engine.resolveTopOfStackWithChoices(g, agents, log)
		copied, _ := g.Stack.Peek()
		if !copied.Copy || copied.CastCopy || copied.ConspirePaid != 1 ||
			copied.Targets[0].PermanentID != enchantment.ObjectID ||
			original.Targets[0].PermanentID != artifact.ObjectID {
			t.Fatal("copy lost the paid decision or failed to retarget independently")
		}
		if engine.putTriggeredAbilitiesOnStack(g) || spellCastEventCount(g) != 1 {
			t.Fatal("a copy put directly onto the stack counted as a cast")
		}
		engine.resolveTopOfStack(g, log)
		if !g.Players[game.Player2].Graveyard.Contains(enchantment.CardInstanceID) ||
			g.Players[game.Player2].Graveyard.Contains(artifact.CardInstanceID) ||
			g.Players[game.Player1].Graveyard.Contains(cardID) {
			t.Fatal("copy did not destroy only its new target")
		}
		engine.resolveTopOfStack(g, log)
		if !g.Players[game.Player2].Graveyard.Contains(artifact.CardInstanceID) ||
			!g.Players[game.Player1].Graveyard.Contains(cardID) {
			t.Fatal("original did not destroy its own target and finish in the graveyard")
		}
		return log
	}
	tapChoices := 0
	log := play(g, conspireTestAgent(func(req game.ChoiceRequest) []int {
		if req.Prompt == "Tap 2 permanents" {
			// Select the multicolored creature first and leave the other
			// eligible creature untapped, rather than taking the default pair.
			want := []id.ID{second.ObjectID, first.ObjectID}[tapChoices]
			tapChoices++
			selected := -1
			for _, option := range req.Options {
				switch option.Targets[0].PermanentID {
				case want:
					selected = option.Index
				case first.ObjectID, spare.ObjectID, second.ObjectID:
				default:
					t.Fatalf("ineligible creature offered for Conspire: %+v", option)
				}
			}
			if selected == -1 {
				t.Fatal("wanted payment creature was not offered")
			}
			return []int{selected}
		}
		if req.Prompt == "Choose new targets" {
			for _, option := range req.Options {
				if len(option.Targets) == 1 && option.Targets[0].PermanentID == enchantment.ObjectID {
					return []int{option.Index}
				}
			}
			t.Fatal("the enchantment must be offered as a fresh target")
		}
		return chooseConspireCost(req)
	}))
	if tapChoices != 2 {
		t.Fatal("caster did not get both Conspire payment choices")
	}
	replayed := 0
	logReplay := play(initial, conspireTestAgent(func(req game.ChoiceRequest) []int {
		if replayed >= len(log.Choices) {
			t.Fatal("replay requested an extra decision")
		}
		record := log.Choices[replayed]
		replayed++
		if req.Prompt != record.Request.Prompt || !reflect.DeepEqual(req.Options, record.Request.Options) {
			t.Fatalf("replay prompt/options differ: got %+v, recorded %+v", req, record.Request)
		}
		return record.Selected
	}))
	if replayed != len(log.Choices) || !reflect.DeepEqual(log.Choices, logReplay.Choices) {
		t.Fatal("replaying the cast from a clone diverged")
	}
}

func TestConspireCombinedAndUnavailableCosts(t *testing.T) {
	for _, tc := range []struct {
		name       string
		eligible   int
		colors     []color.Color
		choosePaid bool
		wantOffer  bool
	}{
		{"combined", 2, []color.Color{color.Green, color.Blue}, true, true},
		{"decline", 2, []color.Color{color.Green, color.Blue}, false, true},
		{"one-eligible", 1, []color.Color{color.Green, color.Blue}, false, false},
		{"colorless-spell", 2, nil, false, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
			engine := NewEngine(nil)
			def := conspireTestSpell()
			def.Colors = tc.colors
			def.AdditionalCosts = []cost.Additional{{Kind: cost.AdditionalDiscard, Amount: 1}}
			def.AdditionalCostChoices = []cost.AdditionalChoice{{Options: []cost.AdditionalChoiceOption{
				{Label: "Two life", Costs: []cost.Additional{{Kind: cost.AdditionalPayLife, Amount: 2}}},
				{Label: "One mana", Mana: cost.Mana{cost.O(1)}},
			}}}
			cardID := addCardToHand(g, game.Player1, def)
			fodder := addCardToHand(g, game.Player1, vanillaCreature("Discard this", 1, 1))
			first := addColoredCreaturePermanent(g, game.Player1, 2, color.Green)
			secondColor := color.Blue
			if tc.eligible < 2 {
				secondColor = color.Red
			}
			second := addColoredCreaturePermanent(g, game.Player1, 2, secondColor)
			g.Players[game.Player1].ManaPool.Add(mana.G, 1)
			g.Players[game.Player1].ManaPool.Add(mana.C, 2)
			offer := false
			agents := conspireTestAgent(func(req game.ChoiceRequest) []int {
				if req.Prompt == "Choose spell cost" {
					want := "One mana"
					if tc.choosePaid {
						want += " + Conspire"
					}
					selection := -1
					for _, option := range req.Options {
						offer = offer || strings.Contains(option.Label, "Conspire")
						if option.Label == want {
							selection = option.Index
						}
					}
					if selection < 0 {
						t.Fatalf("missing choice %q in %+v", want, req.Options)
					}
					return []int{selection}
				}
				return req.DefaultSelection
			})
			if !engine.applyActionWithChoices(g, game.Player1, action.CastSpell(cardID, nil, 0, nil), agents, &TurnLog{}) {
				t.Fatal("otherwise legal cast failed")
			}
			obj, _ := g.Stack.Peek()
			if offer != tc.wantOffer || (obj.ConspirePaid == 1) != tc.choosePaid ||
				first.Tapped != tc.choosePaid || second.Tapped != tc.choosePaid ||
				obj.ManaSpentToCast != 3 || g.Players[game.Player1].Life != 40 ||
				!g.Players[game.Player1].Graveyard.Contains(fodder) {
				t.Fatal("available/selected Conspire branch displaced or mispaid another cost")
			}
			if engine.putTriggeredAbilitiesOnStack(g) != tc.choosePaid {
				t.Fatal("wrong Conspire trigger outcome")
			}
		})
	}
}

func TestConspirePaymentRejectsInvalidOrStaleSubjects(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	def := conspireTestSpell()
	cardID := addCardToHand(g, game.Player1, def)
	first := addColoredCreaturePermanent(g, game.Player1, 2, color.Green)
	second := addColoredCreaturePermanent(g, game.Player1, 2, color.Green)
	spare := addColoredCreaturePermanent(g, game.Player1, 2, color.Green)
	wrong := addColoredCreaturePermanent(g, game.Player1, 2, color.Red)
	g.Players[game.Player1].ManaPool.Add(mana.G, 1)
	g.Players[game.Player1].ManaPool.Add(mana.C, 1)
	planner := NewEngine(nil).PaymentPlanner(g)
	req := payment.ForSpell(payment.SpellRequest{PlayerID: game.Player1, CardID: cardID, Card: def, SourceZone: zone.Hand})
	partial := payment.Partial{Option: opt.Val(1)}
	step := payment.Step{Kind: payment.StepComponent, Index: 0}
	with := func(ids ...id.ID) payment.Partial {
		return partial.With(step, payment.Choice{Component: payment.ComponentChoice{Objects: ids, Finished: true}})
	}
	for _, pick := range [][]id.ID{{first.ObjectID, first.ObjectID}, {first.ObjectID, wrong.ObjectID}} {
		if _, ok, _ := planner.Complete(req, with(pick...)); ok {
			t.Fatalf("invalid pair accepted: %v", pick)
		}
	}
	plan, ok, reason := planner.Complete(req, with(first.ObjectID, second.ObjectID))
	if !ok {
		t.Fatalf("valid pair rejected: %v", reason)
	}
	second.Tapped = true
	if _, ok, _ := planner.Apply(plan); ok || first.Tapped || spare.Tapped || g.Players[game.Player1].ManaPool.Total() != 2 {
		t.Fatal("stale tap must invalidate the plan without charging or substituting another creature")
	}
}

func TestConspireOtherCastRoutes(t *testing.T) {
	for _, route := range []string{"paid-from-exile", "free-from-exile", "cast-card-copy", "prepared-cast-copy"} {
		t.Run(route, func(t *testing.T) {
			g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
			engine := NewEngine(nil)
			def := conspireTestSpell()
			paymentColor := color.Green
			if route == "prepared-cast-copy" {
				paymentColor = color.White
			}
			first := addColoredCreaturePermanent(g, game.Player1, 2, paymentColor)
			second := addColoredCreaturePermanent(g, game.Player1, 2, paymentColor)
			agents := conspireTestAgent(chooseConspireCost)
			log := &TurnLog{}
			castCopy := route == "cast-card-copy" || route == "prepared-cast-copy"
			wantMana := 0
			var source id.ID
			var ok bool
			if route == "prepared-cast-copy" {
				prepared := prepareCard()
				prepared.Alternate.Val.StaticAbilities = def.StaticAbilities
				permanent := addPreparedPermanent(g, game.Player1, prepared)
				source = permanent.CardInstanceID
				setMainPhasePriority(g, game.Player1)
				cast := action.CastSpellFaceFromZone(source, zone.Battlefield, game.FaceAlternate, nil, 0, nil)
				ok = engine.applyActionWithChoices(g, game.Player1, cast, agents, log)
				wantMana = 1
			} else {
				source = addCardToHand(g, game.Player1, def)
				if !moveCardBetweenZones(g, game.Player1, source, zone.Hand, zone.Exile) {
					t.Fatal("could not set up exiled spell")
				}
				if route == "paid-from-exile" {
					g.Players[game.Player1].ManaPool.Add(mana.G, 1)
					g.Players[game.Player1].ManaPool.Add(mana.C, 1)
					ok = engine.castPaidSpellFromSource(g, g.Players[game.Player1], game.Player1, source, zone.Exile, agents, log)
					wantMana = 2
				} else {
					ok = engine.castFreeOptionWithCopy(g, g.Players[game.Player1], game.Player1, g.CardInstances[source],
						action.CastSpellAction{CardID: source, SourceZone: zone.Exile}, false, castCopy, agents, log)
				}
			}
			if !ok {
				t.Fatal("otherwise legal cast failed")
			}
			obj, _ := g.Stack.Peek()
			if obj.ConspirePaid != 1 || obj.CastCopy != castCopy || obj.ManaSpentToCast != wantMana || !first.Tapped || !second.Tapped {
				t.Fatalf("cast route missed the optional payment or copy identity: %+v", obj)
			}
			if !engine.putTriggeredAbilitiesOnStackWithChoices(g, agents, log) || g.Stack.Size() != 2 {
				t.Fatal("genuinely cast spell did not trigger Conspire")
			}
			engine.resolveTopOfStackWithChoices(g, agents, log)
			copied, _ := g.Stack.Peek()
			if !copied.Copy || copied.CastCopy || copied.ConspirePaid != 1 || spellCastEventCount(g) != 1 {
				t.Fatal("Conspire copy inherited the wrong state or counted as a cast")
			}
			engine.resolveTopOfStack(g, log)
			engine.resolveTopOfStack(g, log)
			if g.Players[game.Player1].Graveyard.Contains(source) == castCopy {
				t.Fatal("cast route and copies handled the physical source incorrectly")
			}
		})
	}
}

func TestConspireWithMutateAlternativeCost(t *testing.T) {
	fixture := newMutateFixture(t)
	g := fixture.game
	def := g.CardInstances[fixture.mutatorID].Def
	def.StaticAbilities = append(def.StaticAbilities, game.ConspireStaticBody)
	first := addColoredCreaturePermanent(g, game.Player1, 2, color.Green)
	second := addColoredCreaturePermanent(g, game.Player1, 2, color.Green)
	agents := conspireTestAgent(chooseConspireCost)
	log := &TurnLog{}
	if !fixture.engine.applyActionWithChoices(g, game.Player1,
		action.CastMutateSpell(fixture.mutatorID, fixture.target.ObjectID), agents, log) {
		t.Fatal("mutate cast with Conspire failed")
	}
	original, _ := g.Stack.Peek()
	if !original.Mutate || original.ConspirePaid != 1 || !fixture.forest.Tapped ||
		!first.Tapped || !second.Tapped {
		t.Fatal("mutate alternative cost and Conspire were not jointly paid")
	}
	fixture.engine.putTriggeredAbilitiesOnStackWithChoices(g, agents, log)
	trigger, _ := g.Stack.Peek()
	if !isConspireTrigger(trigger) {
		t.Fatal("paid mutate cast did not trigger Conspire")
	}
	fixture.engine.resolveTopOfStackWithChoices(g, agents, log)
	copied, _ := g.Stack.Peek()
	if !copied.Copy || !copied.Mutate || copied.ConspirePaid != 1 || copied.MutateTargetID != original.MutateTargetID {
		t.Fatal("copy lost either the alternative or additional cost decision")
	}
}

func TestConspireInstancesPaySeparately(t *testing.T) {
	for _, creatures := range []int{3, 4} {
		t.Run(map[int]string{3: "cannot-reuse-creature", 4: "two-paid-instances"}[creatures], func(t *testing.T) {
			g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
			engine := NewEngine(nil)
			def := conspireTestSpell()
			def.StaticAbilities = append(def.StaticAbilities, game.ConspireStaticBody)
			cardID := addCardToHand(g, game.Player1, def)
			for range creatures {
				addColoredCreaturePermanent(g, game.Player1, 2, color.Green)
			}
			g.Players[game.Player1].ManaPool.Add(mana.G, 1)
			g.Players[game.Player1].ManaPool.Add(mana.C, 1)
			want := creatures / 2
			agents := conspireTestAgent(func(req game.ChoiceRequest) []int {
				if req.Prompt == "Choose spell cost" {
					if len(req.Options) != want+1 {
						t.Fatalf("want only the normal cost and %d payable branches: %+v", want, req.Options)
					}
					return []int{req.Options[len(req.Options)-1].Index}
				}
				return req.DefaultSelection
			})
			if !engine.applyActionWithChoices(g, game.Player1, action.CastSpell(cardID, nil, 0, nil), agents, &TurnLog{}) {
				t.Fatal("cast with separately paid Conspire instances failed")
			}
			original, _ := g.Stack.Peek()
			taps := 0
			for _, p := range g.Battlefield {
				if p.Tapped {
					taps++
				}
			}
			engine.putTriggeredAbilitiesOnStack(g)
			if original.ConspirePaid != want || taps != 2*want || g.Stack.Size() != 1+want {
				t.Fatalf("paid=%d tapped=%d triggers=%d, want %d independent payments/triggers", original.ConspirePaid, taps, g.Stack.Size()-1, want)
			}
		})
	}
}

func TestConspireStormCopiesAndCounteredSource(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	def := conspireTestSpell()
	def.StaticAbilities = append(def.StaticAbilities, game.StormStaticBody)
	first := addCardToHand(g, game.Player1, simpleGainLifeInstant("Earlier spell"))
	cardID := addCardToHand(g, game.Player1, def)
	addColoredCreaturePermanent(g, game.Player1, 2, color.Green)
	addColoredCreaturePermanent(g, game.Player1, 2, color.Green)
	g.Players[game.Player1].ManaPool.Add(mana.G, 1)
	g.Players[game.Player1].ManaPool.Add(mana.C, 1)
	agents := conspireTestAgent(chooseConspireCost)
	log := &TurnLog{}
	if !engine.applyAction(g, game.Player1, action.CastSpell(first, nil, 0, nil)) {
		t.Fatal("first spell cast failed")
	}
	engine.resolveTopOfStack(g, log)
	if !engine.applyActionWithChoices(g, game.Player1, action.CastSpell(cardID, nil, 0, nil), agents, log) {
		t.Fatal("paid storm/Conspire cast failed")
	}
	original, _ := g.Stack.Peek()
	engine.putTriggeredAbilitiesOnStackWithChoices(g, agents, log)
	clone := g.Clone()
	if !counterStackObject(g, original.ID) || !g.Players[game.Player1].Graveyard.Contains(cardID) {
		t.Fatal("failed to counter physical original under its cast triggers")
	}
	for _, branch := range []*game.Game{g, clone} {
		for range 2 {
			top, _ := branch.Stack.Peek()
			if !isConspireTrigger(top) && !isStormTrigger(top) {
				t.Fatal("expected independently stacked Conspire and Storm triggers")
			}
			engine.resolveTopOfStack(branch, log)
			copied, _ := branch.Stack.Peek()
			if !copied.Copy || copied.CastCopy || copied.ConspirePaid != 1 {
				t.Fatal("cast trigger failed to copy the paid decision")
			}
			engine.resolveTopOfStack(branch, log)
		}
		if spellCastEventCount(branch) != 2 || engine.putTriggeredAbilitiesOnStack(branch) {
			t.Fatal("copies counted as casts or triggered Conspire again")
		}
	}
}
