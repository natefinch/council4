package rules

import (
	"reflect"
	"slices"
	"strings"
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/id"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/mtg/rules/payment"
	"github.com/natefinch/council4/opt"
)

// Official CR effective 2026-11-06: 605.1a/b, 605.3a/b, 605.4,
// 106.12, 118.3c, 601.2g/h. City of Brass's ordinary tap trigger is not a
// mana ability; Wild Growth's tapped-for-mana trigger resolves immediately.
func TestPlannedManaRegisteredSources(t *testing.T) {
	tests := []struct {
		name  string
		mana  cost.Mana
		setup func(*game.Game, *game.Permanent)
		check func(*testing.T, *game.Game, *game.Permanent)
	}{
		{"Waterlogged Grove", cost.Mana{cost.U}, nil, func(t *testing.T, g *game.Game, p *game.Permanent) {
			if g.Players[0].Life != 39 || !p.Tapped {
				t.Fatal("life payment and tap were not committed")
			}
		}},
		{"Vivid Creek", cost.Mana{cost.R}, func(g *game.Game, p *game.Permanent) { p.Counters.Add(counter.Charge, 1) }, func(t *testing.T, g *game.Game, p *game.Permanent) {
			if p.Counters.Get(counter.Charge) != 0 || !p.Tapped {
				t.Fatal("charge counter and tap were not paid")
			}
		}},
		{"Prismatic Lens", cost.Mana{cost.U}, func(g *game.Game, p *game.Permanent) { g.Players[0].ManaPool.Add(mana.C, 1) }, nil},
		{"Lotus Petal", cost.Mana{cost.U}, nil, func(t *testing.T, g *game.Game, p *game.Permanent) {
			if _, ok := permanentByObjectID(g, p.ObjectID); ok || !g.Players[0].Graveyard.Contains(p.CardInstanceID) {
				t.Fatal("Petal was not sacrificed")
			}
		}},
		{"Shivan Reef", cost.Mana{cost.U}, nil, func(t *testing.T, g *game.Game, p *game.Permanent) {
			if g.Players[0].Life != 39 {
				t.Fatal("painland damage was not dealt")
			}
		}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
			e := NewEngine(nil)
			source := addCombatPermanent(g, game.Player1, registryCardDef(t, test.name))
			if test.setup != nil {
				test.setup(g, source)
			}
			before := g.Clone()
			planner := e.PaymentPlanner(g)
			req := payment.ForGeneric(payment.GenericRequest{PlayerID: game.Player1, Cost: &test.mana})
			plan, ok, reason := planner.Complete(req, payment.Partial{})
			if !ok {
				t.Fatal(reason)
			}
			if !reflect.DeepEqual(g.Clone(), before) {
				t.Fatal("Complete mutated game state")
			}
			if len(plan.Mana().Activations) != 1 {
				t.Fatalf("activations = %+v", plan.Mana().Activations)
			}
			if _, ok, reason := planner.Apply(plan); !ok {
				t.Fatal(reason)
			}
			if g.Players[0].ManaPool.Total() != 0 || g.Stack.Size() != 0 {
				t.Fatal("wrong remaining mana or mana ability on stack")
			}
			if test.check != nil {
				test.check(t, g, source)
			}
		})
	}
}

func TestPlannedManaWildGrowthAndReflection(t *testing.T) {
	for _, name := range []string{"Wild Growth", "Mana Reflection"} {
		t.Run(name, func(t *testing.T) {
			g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
			e := NewEngine(nil)
			forest := addCombatPermanent(g, game.Player1, registryCardDef(t, "Forest"))
			effect := addCombatPermanent(g, game.Player1, registryCardDef(t, name))
			if name == "Wild Growth" {
				effect.AttachedTo = opt.Val(forest.ObjectID)
			}
			m := cost.Mana{cost.G, cost.G}
			req := payment.ForGeneric(payment.GenericRequest{PlayerID: game.Player1, Cost: &m})
			before := g.Clone()
			plan, ok, reason := e.PaymentPlanner(g).Complete(req, payment.Partial{})
			if !ok {
				t.Fatal(reason)
			}
			if !reflect.DeepEqual(g.Clone(), before) {
				t.Fatal("forecast mutated live rules state")
			}
			if _, ok, reason := e.PaymentPlanner(g).Apply(plan); !ok {
				t.Fatal(reason)
			}
			if e.putTriggeredAbilitiesOnStack(g) || g.Stack.Size() != 0 || g.Players[0].ManaPool.Total() != 0 {
				t.Fatal("mana trigger did not resolve exactly once during activation")
			}
		})
	}
}

func TestOptionalManaTriggerDeclinesAndForecastIsOpaque(t *testing.T) {
	for _, accept := range []bool{false, true} {
		t.Run(map[bool]string{false: "decline", true: "accept"}[accept], func(t *testing.T) {
			g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
			e := NewEngine(nil)
			land := basicColorLand(g, game.Player1, "Island", types.Island, mana.U)
			addOptionalTriggeredPermanent(g, game.Player1, &game.TriggerPattern{
				Event: game.EventManaProduced, Controller: game.TriggerControllerYou, RequireManaProducedByLand: true,
			}, []game.Instruction{{Primitive: game.AddMana{ManaColor: mana.U, Amount: game.Fixed(1)}}}, nil)
			m := cost.Mana{cost.U, cost.U}
			before := g.Clone()
			if _, ok, _ := e.PaymentPlanner(g).Complete(payment.ForGeneric(payment.GenericRequest{PlayerID: game.Player1, Cost: &m}), payment.Partial{}); ok {
				t.Error("unplanned optional mana bonus was forecast as deterministic")
			}
			if !reflect.DeepEqual(before, g.Clone()) {
				t.Fatal("optional mana forecast mutated game")
			}
			decision := 0
			if accept {
				decision = 1
			}
			agent := &choiceOnlyAgent{choices: [][]int{{decision}}}
			agents := [game.NumPlayers]PlayerAgent{agent}
			if !e.applyActivateAbilityWithChoices(g, game.Player1, action.ActivateAbilityAction{SourceID: land.ObjectID}, agents, &TurnLog{}) {
				t.Fatal("mana activation failed")
			}
			if got := g.Players[0].ManaPool.Amount(mana.U); got != 1+decision {
				t.Errorf("blue mana = %d, want %d", got, 1+decision)
			}
			if agent.next != 1 || e.putTriggeredAbilitiesOnStack(g) || g.Stack.Size() != 0 {
				t.Fatal("optional mana trigger did not offer its decision exactly once during activation")
			}
		})
	}
}

func TestImmediateManaTriggerRechecksInterveningCondition(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	e := NewEngine(nil)
	land := basicColorLand(g, game.Player1, "Island", types.Island, mana.U)
	pattern := &game.TriggerPattern{Event: game.EventManaProduced, Controller: game.TriggerControllerYou, RequireManaProducedByLand: true}
	addTriggeredPermanent(g, game.Player1, pattern, []game.Instruction{
		{Primitive: game.LoseLife{Player: game.ControllerReference(), Amount: game.Fixed(1)}},
		{Primitive: game.AddMana{ManaColor: mana.U, Amount: game.Fixed(1)}},
	}, nil)
	addTriggeredPermanentWithCondition(g, game.Player1, pattern, 40, []game.Instruction{{Primitive: game.AddMana{ManaColor: mana.U, Amount: game.Fixed(1)}}})
	if !e.applyAction(g, game.Player1, action.ActivateAbility(land.ObjectID, 0, nil, 0)) {
		t.Fatal("mana activation failed")
	}
	if g.Players[0].Life != 39 || g.Players[0].ManaPool.Amount(mana.U) != 2 {
		t.Fatal("captured mana trigger ignored its now-false intervening condition")
	}
	if e.putTriggeredAbilitiesOnStack(g) || g.Stack.Size() != 0 {
		t.Fatal("immediate mana trigger was left on the stack")
	}
}

func TestPlannedManaHandSourceAndChoices(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	e := NewEngine(nil)
	card := addCardToHand(g, game.Player1, registryCardDef(t, "Simian Spirit Guide"))
	m := cost.Mana{cost.R}
	req := payment.ForGeneric(payment.GenericRequest{PlayerID: game.Player1, Cost: &m})
	plan, ok, reason := e.PaymentPlanner(g).Complete(req, payment.Partial{})
	if !ok {
		t.Fatal(reason)
	}
	activation := plan.Mana().Activations[0]
	if activation.Source != card || activation.SourceZone != zone.Hand {
		t.Fatal("hand source identity lost")
	}
	options := paymentManaAbilityOptions(g, game.Player1, []action.Action{action.ActivateAbility(card, activation.AbilityIndex, nil, 0)})
	if len(options) != 2 || !strings.Contains(options[1].Label, "Simian Spirit Guide") || !strings.Contains(options[1].Label, "Exile") {
		t.Fatalf("hand source label = %+v", options)
	}
	log := &TurnLog{}
	if _, ok, reason := e.PaymentPlanner(g, log).Apply(plan); !ok {
		t.Fatal(reason)
	}
	if len(log.Entries) != 1 || log.Entries[0].Action.PermanentSources[card] != card || !strings.Contains(log.Entries[0].Action.AbilityText, "Exile") {
		t.Fatal("planned hand activation did not retain its source-labelled normal mana log")
	}
	if !g.Players[0].Exile.Contains(card) || g.Players[0].ManaPool.Total() != 0 {
		t.Fatal("hand source was not paid exactly")
	}
}

func TestPlannedManaTreasureOptionalAndDisjoint(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	e := NewEngine(nil)
	treasure := addTreasureTokenPermanent(g, game.Player1)
	treasure.TokenDef = testTreasureToken()
	land := addCombatPermanent(g, game.Player1, registryCardDef(t, "Island"))
	m := cost.Mana{cost.U}
	req := payment.ForGeneric(payment.GenericRequest{PlayerID: game.Player1, Cost: &m})
	planner := e.PaymentPlanner(g)
	plan, ok, reason := planner.Complete(req, payment.Partial{})
	if !ok {
		t.Fatal(reason)
	}
	if len(plan.Mana().Activations) != 1 || plan.Mana().Activations[0].Source != land.ObjectID {
		t.Fatal("unnecessary Treasure consumed")
	}
	index := treasure.TokenDef.ManaAbilityIndex(0)
	step := payment.Step{Kind: payment.StepManaAbility, Source: treasure.ObjectID, Index: index}
	options := planner.Options(req, payment.Partial{}, step)
	if len(options) != 2 {
		t.Fatalf("strategic options = %+v", options)
	}
	chosen := options[slices.IndexFunc(options, func(c payment.Choice) bool { return c.UseManaAbility })]
	if len(chosen.Plan.Mana().Activations) != 1 || chosen.Plan.Mana().Activations[0].Source != treasure.ObjectID {
		t.Fatal("selected Treasure not used alone")
	}
	requireApplyPayment(t, g, chosen.Plan)
	if _, ok := permanentByObjectID(g, treasure.ObjectID); ok || land.Tapped {
		t.Fatal("wrong strategic source applied")
	}
	land.Tapped = true
	m = cost.Mana{cost.U, cost.U}
	if _, ok, _ := planner.Complete(req, payment.Partial{}); ok {
		t.Fatal("removed resource was used twice")
	}
}

func TestPlannedManaOpaqueAndLibraryBoundary(t *testing.T) {
	body := game.TapManaAbility(mana.U)
	body.Content.Modes[0].Sequence = append(body.Content.Modes[0].Sequence, game.Instruction{Primitive: game.Untap{Object: game.SourcePermanentReference()}})
	if _, class := payment.ClassifyManaAbility(&body); class != payment.OpaqueManaAbility {
		t.Fatal("unrecognized shape not opaque")
	}
	body.Content.Modes[0].Sequence[1].Primitive = game.Draw{Amount: game.Fixed(1), Player: game.ControllerReference()}
	if _, class := payment.ClassifyManaAbility(&body); class != payment.NotManaAbility || manaBodyHasAddManaEffect(&body) {
		t.Fatal("library-moving activated ability classified as mana")
	}
}

func TestPlannedManaCounterAdditionAndLifeReservation(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	e := NewEngine(nil)
	body := painlandColoredManaAbility(mana.U, 1)
	body.Content.Modes[0].Sequence = append(body.Content.Modes[0].Sequence, game.Instruction{Primitive: game.AddCounter{Object: game.SourcePermanentReference(), CounterKind: counter.Charge, Amount: game.Fixed(1)}})
	source := addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Counter pain", Types: []types.Card{types.Artifact}, ManaAbilities: []game.ManaAbility{body}}})
	m := cost.Mana{cost.U}
	req := payment.ForGeneric(payment.GenericRequest{PlayerID: game.Player1, Cost: &m, AdditionalCosts: []cost.Additional{{Kind: cost.AdditionalPayLife, Amount: 1}}})
	g.Players[0].Life = 1
	if _, ok, _ := e.PaymentPlanner(g).Complete(req, payment.Partial{}); ok {
		t.Fatal("damage reused life reserved for the cost")
	}
	g.Players[0].Life = 2
	plan, ok, reason := e.PaymentPlanner(g).Complete(req, payment.Partial{})
	if !ok {
		t.Fatal(reason)
	}
	if _, ok, reason := e.PaymentPlanner(g).Apply(plan); !ok {
		t.Fatal(reason)
	}
	if source.Counters.Get(counter.Charge) != 1 || g.Players[0].Life != 0 {
		t.Fatal("counter effect or joint life payment not exact")
	}
}

func TestPlannedManaEffectsUnlockNonmanaCosts(t *testing.T) {
	for _, life := range []bool{false, true} {
		t.Run(map[bool]string{false: "counter", true: "life"}[life], func(t *testing.T) {
			g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
			e := NewEngine(nil)
			body := game.TapManaAbility(mana.U)
			body.Content.Modes[0].Sequence = append(body.Content.Modes[0].Sequence, game.Instruction{Primitive: game.AddCounter{Object: game.SourcePermanentReference(), CounterKind: counter.Charge, Amount: game.Fixed(1)}})
			source := addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Counter producer", Types: []types.Card{types.Artifact}, ManaAbilities: []game.ManaAbility{body}}})
			req := payment.ForAbility(payment.AbilityRequest{PlayerID: game.Player1, Source: source, ManaCost: opt.Val(cost.Mana{cost.U}), AdditionalCosts: []cost.Additional{{Kind: cost.AdditionalRemoveCounter, Amount: 1, AnyCounterKind: true}}})
			if life {
				g.Battlefield = nil
				source = addCombatPermanent(g, game.Player1, registryCardDef(t, "The Great Henge"))
				req = payment.ForAbility(payment.AbilityRequest{PlayerID: game.Player1, Source: source, ManaCost: opt.Val(cost.Mana{cost.G, cost.G}), AdditionalCosts: []cost.Additional{{Kind: cost.AdditionalPayLife, Amount: 42}}})
			}
			before := g.Clone()
			plan, ok, reason := e.PaymentPlanner(g).Complete(req, payment.Partial{})
			if !ok {
				t.Fatal(reason)
			}
			if !life {
				options := e.PaymentPlanner(g).Options(req, payment.Partial{}, payment.Step{Kind: payment.StepComponent, Index: 0})
				if len(options) != 1 || !slices.Equal(options[0].Component.Counters, []payment.CounterRemovalChoice{{Permanent: source.ObjectID, Kind: counter.Charge}}) {
					t.Fatalf("projected counter options = %+v, want the produced charge counter", options)
				}
				plan = options[0].Plan
			}
			if !reflect.DeepEqual(before, g.Clone()) {
				t.Fatal("resource-unlocking forecast mutated game")
			}
			if _, ok, reason := e.PaymentPlanner(g).Apply(plan); !ok {
				t.Fatal(reason)
			}
			if !source.Counters.IsEmpty() || g.Players[0].ManaPool.Total() != 0 || life && g.Players[0].Life != 0 {
				t.Fatal("unlocked costs were not paid exactly")
			}
		})
	}
}

func TestPlannedManaComponentOptionsRetainAlternateSourcesAndKinds(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	e := NewEngine(nil)
	var expected []payment.CounterRemovalChoice
	for _, kinds := range [][]counter.Kind{{counter.Charge, counter.Stun}, {counter.Charge}} {
		var bodies []game.ManaAbility
		for _, kind := range kinds {
			body := game.TapManaAbility(mana.U)
			body.Content.Modes[0].Sequence = append(body.Content.Modes[0].Sequence, game.Instruction{Primitive: game.AddCounter{Object: game.SourcePermanentReference(), CounterKind: kind, Amount: game.Fixed(1)}})
			bodies = append(bodies, body)
		}
		source := addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Counter producer", Types: []types.Card{types.Artifact}, ManaAbilities: bodies}})
		for _, kind := range kinds {
			expected = append(expected, payment.CounterRemovalChoice{Permanent: source.ObjectID, Kind: kind})
		}
	}
	m := cost.Mana{cost.U}
	req := payment.ForGeneric(payment.GenericRequest{PlayerID: game.Player1, Cost: &m, AdditionalCosts: []cost.Additional{{Kind: cost.AdditionalRemoveCounterAmong, Amount: 1, AnyCounterKind: true, MatchPermanentType: true, PermanentType: types.Artifact}}})
	before := g.Clone()
	options := e.PaymentPlanner(g).Options(req, payment.Partial{}, payment.Step{Kind: payment.StepComponent, Index: 0})
	if len(options) != len(expected) {
		t.Fatalf("projected options = %+v, want %d alternate source/kind choices", options, len(expected))
	}
	for _, removal := range expected {
		index := slices.IndexFunc(options, func(choice payment.Choice) bool {
			return choice.Component.Finished && slices.Equal(choice.Component.Counters, []payment.CounterRemovalChoice{removal})
		})
		if index < 0 {
			t.Fatalf("missing projected option %+v", removal)
		}
		clone := g.Clone()
		if _, ok, reason := e.PaymentPlanner(clone).Apply(options[index].Plan); !ok {
			t.Fatal(reason)
		}
		for _, source := range clone.Battlefield {
			if !source.Counters.IsEmpty() {
				t.Fatal("offered witness did not remove its produced counter")
			}
		}
	}
	if !reflect.DeepEqual(before, g.Clone()) {
		t.Fatal("projected choice enumeration mutated game")
	}
}

func TestPlannedManaCounterResourcesCannotBeReused(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	e := NewEngine(nil)
	body := game.ManaAbility{AdditionalCosts: []cost.Additional{{Kind: cost.AdditionalRemoveCounter, CounterKind: counter.Charge, Amount: 1}}, Content: game.Mode{Sequence: []game.Instruction{{Primitive: game.AddMana{ManaColor: mana.G, Amount: game.Fixed(1)}}}}.Ability()}
	source := addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Charged mana", Types: []types.Card{types.Artifact}, ManaAbilities: []game.ManaAbility{body}}})
	source.Counters.Add(counter.Charge, 2)
	req := payment.ForAbility(payment.AbilityRequest{PlayerID: game.Player1, Source: source, ManaCost: opt.Val(cost.Mana{cost.G, cost.G}), AdditionalCosts: []cost.Additional{{Kind: cost.AdditionalRemoveCounter, CounterKind: counter.Charge, Amount: 1}}})
	if _, ok, _ := e.PaymentPlanner(g).Complete(req, payment.Partial{}); ok {
		t.Fatal("same charge reserved for mana and the outer cost")
	}
	req = payment.ForAbility(payment.AbilityRequest{PlayerID: game.Player1, Source: source, ManaCost: opt.Val(cost.Mana{cost.G}), AdditionalCosts: []cost.Additional{{Kind: cost.AdditionalRemoveCounter, CounterKind: counter.Charge, Amount: 1}}})
	plan, ok, reason := e.PaymentPlanner(g).Complete(req, payment.Partial{})
	if !ok {
		t.Fatal(reason)
	}
	if _, ok, reason := e.PaymentPlanner(g).Apply(plan); !ok {
		t.Fatal(reason)
	}
	if !source.Counters.IsEmpty() {
		t.Fatal("joint charge reservation not committed")
	}
}

func TestPlannedManaRiderReceiptAndFilterConsumption(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{{Commander: registryCardDef(t, "The Council of Four")}})
	e := NewEngine(nil)
	def := registryCardDef(t, "Path of Ancestry")
	source := addCombatPermanent(g, game.Player1, def)
	spell := vanillaCreature("Rider spell", 1, 1)
	spell.ManaCost = opt.Val(cost.Mana{cost.U})
	card := addCardToHand(g, game.Player1, spell)
	req := payment.ForSpell(payment.SpellRequest{PlayerID: game.Player1, CardID: card, Card: spell, SourceZone: zone.Hand})
	plan, ok, reason := e.PaymentPlanner(g).Complete(req, payment.Partial{})
	if !ok {
		t.Fatal(reason)
	}
	applied, ok, reason := e.PaymentPlanner(g).Apply(plan)
	if !ok {
		t.Fatal(reason)
	}
	unit := mana.Unit{Color: mana.U}
	if applied.ManaBefore[unit] != 1 || len(g.Players[0].ManaRiders) != 1 || g.Players[0].ManaPool.Total() != 0 {
		t.Fatal("planned rider mana missing from spend receipt")
	}
	consumeManaSpendRidersForPayment(g, game.Player1, nil, applied.ManaBefore, applied.Spell.PoolSpend)
	if len(g.Players[0].ManaRiders) != 0 {
		t.Fatal("receipt did not consume produced rider")
	}
	source.Tapped = true
	addCombatPermanent(g, game.Player1, registryCardDef(t, "Prismatic Lens"))
	g.Players[0].ManaPool.Add(mana.C, 1)
	g.Players[0].ManaRiders = []game.ManaRiderInstance{{Unit: mana.Unit{Color: mana.C}, Rider: scryRider()}}
	m := cost.Mana{cost.U}
	req = payment.ForGeneric(payment.GenericRequest{PlayerID: game.Player1, Cost: &m})
	plan, ok, reason = e.PaymentPlanner(g).Complete(req, payment.Partial{})
	if !ok {
		t.Fatal(reason)
	}
	if _, ok, reason := e.PaymentPlanner(g).Apply(plan); !ok {
		t.Fatal(reason)
	}
	if len(g.Players[0].ManaRiders) != 0 || g.Players[0].ManaPool.Total() != 0 {
		t.Fatal("filter retained a rider after spending its input")
	}
}

func TestPlannedManaReadOnlyViewAndStalePlan(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	e := NewEngine(nil)
	source := addCombatPermanent(g, game.Player1, registryCardDef(t, "Vivid Creek"))
	source.Counters.Add(counter.Charge, 1)
	state := &rulesPaymentState{g: g, engine: e}
	view := state.ProjectResources(game.ResourceProjection{Permanents: map[id.ID]game.PermanentResources{source.ObjectID: {Counters: map[counter.Kind]int{counter.Charge: 3}}}}).(*rulesPaymentState)
	condition := opt.Val(game.Condition{SourceCounterKindKnown: true, SourceCounterKind: counter.Charge, SourceCountersAtLeast: 2})
	if !view.ActivationConditionSatisfied(game.Player1, source, condition) || state.ActivationConditionSatisfied(game.Player1, source, condition) {
		t.Fatal("view did not translate Permanent by object identity")
	}
	func() {
		defer func() {
			if recover() == nil {
				t.Error("view allowed a mutation")
			}
		}()
		view.SetTappedForMana(source)
	}()
	m := cost.Mana{cost.R}
	req := payment.ForGeneric(payment.GenericRequest{PlayerID: game.Player1, Cost: &m})
	plan, ok, reason := e.PaymentPlanner(g).Complete(req, payment.Partial{})
	if !ok {
		t.Fatal(reason)
	}
	if _, ok, _ := payment.New(view).Apply(plan); ok {
		t.Fatal("Apply accepted an evaluation-only state")
	}
	source.Counters.Remove(counter.Charge, 1)
	before := g.Clone()
	if _, ok, _ := e.PaymentPlanner(g).Apply(plan); ok || !reflect.DeepEqual(before, g.Clone()) {
		t.Fatal("stale activation mutated the game or selected another source")
	}
}

func TestPlannedManaRecipientAndSacrificedSourceReplacement(t *testing.T) {
	for _, sacrificed := range []bool{false, true} {
		t.Run(map[bool]string{false: "opponent Aura", true: "Petal replacement"}[sacrificed], func(t *testing.T) {
			g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
			e := NewEngine(nil)
			m := cost.Mana{cost.G, cost.G}
			if sacrificed {
				addCombatPermanent(g, game.Player1, registryCardDef(t, "Lotus Petal"))
				addCombatPermanent(g, game.Player1, registryCardDef(t, "Mana Reflection"))
			} else {
				forest := addCombatPermanent(g, game.Player1, registryCardDef(t, "Forest"))
				aura := addCombatPermanent(g, game.Player2, registryCardDef(t, "Wild Growth"))
				aura.AttachedTo = opt.Val(forest.ObjectID)
			}
			req := payment.ForGeneric(payment.GenericRequest{PlayerID: game.Player1, Cost: &m})
			plan, ok, reason := e.PaymentPlanner(g).Complete(req, payment.Partial{})
			if !ok {
				t.Fatal(reason)
			}
			if _, ok, reason := e.PaymentPlanner(g).Apply(plan); !ok {
				t.Fatal(reason)
			}
			if g.Players[0].ManaPool.Total() != 0 || g.Players[1].ManaPool.Total() != 0 || e.putTriggeredAbilitiesOnStack(g) || g.Stack.Size() != 0 {
				t.Fatal("forecast and immediate recipient/replacement behavior diverged")
			}
		})
	}
}

func TestPlannedManaCounterReplacementIsPureOrOpaque(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	e := NewEngine(nil)
	addReplacementPermanent(t, g, game.Player1, doublingSeasonCounterReplacementCardDef())
	body := game.TapManaAbility(mana.U)
	body.Content.Modes[0].Sequence = append(body.Content.Modes[0].Sequence, game.Instruction{Primitive: game.AddCounter{Object: game.SourcePermanentReference(), CounterKind: counter.Charge, Amount: game.Fixed(1)}})
	source := addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Charge production", Types: []types.Card{types.Artifact}, ManaAbilities: []game.ManaAbility{body}}})
	m := cost.Mana{cost.U}
	req := payment.ForGeneric(payment.GenericRequest{PlayerID: game.Player1, Cost: &m})
	before := g.Clone()
	plan, ok, reason := e.PaymentPlanner(g).Complete(req, payment.Partial{})
	if !ok {
		t.Fatal(reason)
	}
	if !reflect.DeepEqual(before, g.Clone()) {
		t.Fatal("counter replacement prediction made a choice or mutated the game")
	}
	if _, ok, reason := e.PaymentPlanner(g).Apply(plan); !ok {
		t.Fatal(reason)
	}
	if source.Counters.Get(counter.Charge) != 2 {
		t.Fatal("canonical counter replacement not applied")
	}
	source.Tapped = false
	addReplacementPermanent(t, g, game.Player1, doublingSeasonCounterReplacementCardDef())
	before = g.Clone()
	if _, ok, _ := e.PaymentPlanner(g).Complete(req, payment.Partial{}); ok || !reflect.DeepEqual(before, g.Clone()) {
		t.Fatal("selectable replacement ordering was not opaque and pure")
	}
}

func TestPlannedManaRepeatingDelayedHighTide(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	e := NewEngine(nil)
	highTideDelayedTrigger(g, game.Player1)
	addCombatPermanent(g, game.Player1, registryCardDef(t, "Island"))
	m := cost.Mana{cost.U, cost.U}
	req := payment.ForGeneric(payment.GenericRequest{PlayerID: game.Player1, Cost: &m})
	before := g.Clone()
	plan, ok, reason := e.PaymentPlanner(g).Complete(req, payment.Partial{})
	if !ok {
		t.Fatal(reason)
	}
	if !reflect.DeepEqual(before, g.Clone()) {
		t.Fatal("High Tide forecast drained a delayed trigger")
	}
	if _, ok, reason := e.PaymentPlanner(g).Apply(plan); !ok {
		t.Fatal(reason)
	}
	if g.Players[0].ManaPool.Total() != 0 || e.putTriggeredAbilitiesOnStack(g) || g.Stack.Size() != 0 || len(g.DelayedTriggers) != 1 {
		t.Fatal("repeating delayed mana did not resolve immediately exactly once")
	}
}

func TestPlannedManaThenConvokeUsesLivePermanent(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	e := NewEngine(nil)
	body := game.ManaAbility{AdditionalCosts: []cost.Additional{{Kind: cost.AdditionalRemoveCounter, CounterKind: counter.Charge, Amount: 1}}, Content: game.Mode{Sequence: []game.Instruction{{Primitive: game.AddMana{ManaColor: mana.C, Amount: game.Fixed(1)}}}}.Ability()}
	def := vanillaCreature("Counter mana creature", 1, 1)
	def.Colors = []color.Color{color.Green}
	def.ManaAbilities = []game.ManaAbility{body}
	source := addCombatPermanent(g, game.Player1, def)
	source.Counters.Add(counter.Charge, 1)
	spell := &game.CardDef{CardFace: game.CardFace{Name: "Convoke payment", ManaCost: opt.Val(cost.Mana{cost.G, cost.O(1)}), Types: []types.Card{types.Instant}, StaticAbilities: []game.StaticAbility{{ZoneOfFunction: zone.Stack, KeywordAbilities: game.SimpleKeywords(game.Convoke)}}}}
	card := addCardToHand(g, game.Player1, spell)
	req := payment.ForSpell(payment.SpellRequest{PlayerID: game.Player1, CardID: card, SourceZone: zone.Hand, Card: spell})
	plan, ok, reason := e.PaymentPlanner(g).Complete(req, payment.Partial{})
	if !ok {
		t.Fatal(reason)
	}
	if len(plan.Mana().Activations) != 1 || !slices.Equal(plan.Mana().Convoke, []id.ID{source.ObjectID}) {
		t.Fatal("expected counter mana followed by convoke")
	}
	if _, ok, reason := e.PaymentPlanner(g).Apply(plan); !ok {
		t.Fatal(reason)
	}
	if !source.Tapped || !source.Counters.IsEmpty() || g.Players[0].ManaPool.Total() != 0 {
		t.Fatal("Apply mutated a projected permanent rather than the live resource")
	}
}

func TestPlannedManaMemoDistinguishesSpendRestrictions(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	e := NewEngine(nil)
	plain := game.ManaAbility{AdditionalCosts: []cost.Additional{{Kind: cost.AdditionalRemoveCounter, CounterKind: counter.Charge, Amount: 1}}, Content: game.Mode{Sequence: []game.Instruction{{Primitive: game.AddMana{ManaColor: mana.C, Amount: game.Fixed(1)}}}}.Ability()}
	restricted := plain
	restricted.Content.Modes = slices.Clone(plain.Content.Modes)
	restricted.Content.Modes[0].Sequence = []game.Instruction{{Primitive: game.AddMana{ManaColor: mana.C, Amount: game.Fixed(1), SpendRider: opt.Val(game.ManaSpendRider{Condition: game.ManaSpendCastArtifactSpellOnly, Restriction: game.ManaSpendRestrictedToCondition})}}}
	def := &game.CardDef{CardFace: game.CardFace{Name: "Restricted or plain", Types: []types.Card{types.Artifact}, ManaAbilities: []game.ManaAbility{restricted, plain}}}
	source := addCombatPermanent(g, game.Player1, def)
	source.Counters.Add(counter.Charge, 1)
	m := cost.Mana{cost.O(1)}
	req := payment.ForGeneric(payment.GenericRequest{PlayerID: game.Player1, Cost: &m})
	plan, ok, reason := e.PaymentPlanner(g).Complete(req, payment.Partial{})
	if !ok {
		t.Fatal(reason)
	}
	if plan.Mana().Activations[0].AbilityIndex != 1 {
		t.Fatal("restricted mana was used for an ineligible payment")
	}
	if _, ok, reason := e.PaymentPlanner(g).Apply(plan); !ok {
		t.Fatal(reason)
	}
}

func TestPlannedManaTapTriggerSeesProductionTimeResources(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	e := NewEngine(nil)
	addCombatPermanent(g, game.Player1, registryCardDef(t, "Waterlogged Grove"))
	addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Production-time bonus", Types: []types.Card{types.Enchantment}, TriggeredAbilities: []game.TriggeredAbility{{
		Trigger: game.TriggerCondition{Type: game.TriggerWhenever, Pattern: game.TriggerPattern{Event: game.EventPermanentTapped, RequireTappedForMana: true, Controller: game.TriggerControllerYou}, InterveningCondition: opt.Val(game.Condition{AnyPlayerLifeAtMost: 39})},
		Content: game.Mode{Sequence: []game.Instruction{{Primitive: game.AddMana{ManaColor: mana.U, Amount: game.Fixed(1)}}}}.Ability(),
	}}}})
	m := cost.Mana{cost.U, cost.U}
	req := payment.ForGeneric(payment.GenericRequest{PlayerID: game.Player1, Cost: &m})
	plan, ok, reason := e.PaymentPlanner(g).Complete(req, payment.Partial{})
	if !ok {
		t.Fatal(reason)
	}
	if _, ok, reason := e.PaymentPlanner(g).Apply(plan); !ok {
		t.Fatal(reason)
	}
	if g.Players[0].Life != 39 || g.Players[0].ManaPool.Total() != 0 || e.putTriggeredAbilitiesOnStack(g) {
		t.Fatal("CR 106.12a mana trigger was evaluated at tap-cost time rather than production")
	}
}

func TestPlannedManaOneShotDelayedCannotBeUsedTwice(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	e := NewEngine(nil)
	highTideDelayedTrigger(g, game.Player1)
	g.DelayedTriggers[0].OneShot = true
	addCombatPermanent(g, game.Player1, registryCardDef(t, "Island"))
	addCombatPermanent(g, game.Player1, registryCardDef(t, "Island"))
	m := cost.Mana{cost.U, cost.U, cost.U, cost.U}
	req := payment.ForGeneric(payment.GenericRequest{PlayerID: game.Player1, Cost: &m})
	if _, ok, _ := e.PaymentPlanner(g).Complete(req, payment.Partial{}); ok {
		t.Fatal("one-shot delayed mana was counted for both Islands")
	}
	m = cost.Mana{cost.U, cost.U, cost.U}
	req = payment.ForGeneric(payment.GenericRequest{PlayerID: game.Player1, Cost: &m})
	plan, ok, reason := e.PaymentPlanner(g).Complete(req, payment.Partial{})
	if !ok {
		t.Fatal(reason)
	}
	if _, ok, reason := e.PaymentPlanner(g).Apply(plan); !ok {
		t.Fatal(reason)
	}
	if len(g.DelayedTriggers) != 0 || g.Players[0].ManaPool.Total() != 0 || e.putTriggeredAbilitiesOnStack(g) {
		t.Fatal("one-shot delayed mana did not apply exactly once")
	}
}

func TestPlannedManaStunReplacementDoesNotFreeConvoke(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	g.Turn.TurnNumber = 2
	e := NewEngine(nil)
	def := vanillaCreature("Untap mana creature", 1, 1)
	def.Colors = []color.Color{color.Green}
	def.ManaAbilities = []game.ManaAbility{{AdditionalCosts: []cost.Additional{{Kind: cost.AdditionalUntap, Amount: 1}}, Content: game.Mode{Sequence: []game.Instruction{{Primitive: game.AddMana{ManaColor: mana.C, Amount: game.Fixed(1)}}}}.Ability()}}
	source := addCombatPermanent(g, game.Player1, def)
	source.Tapped = true
	source.Counters.Add(counter.Stun, 1)
	spell := &game.CardDef{CardFace: game.CardFace{Name: "Stun convoke", ManaCost: opt.Val(cost.Mana{cost.G, cost.O(1)}), Types: []types.Card{types.Instant}, StaticAbilities: []game.StaticAbility{{ZoneOfFunction: zone.Stack, KeywordAbilities: game.SimpleKeywords(game.Convoke)}}}}
	card := addCardToHand(g, game.Player1, spell)
	req := payment.ForSpell(payment.SpellRequest{PlayerID: game.Player1, CardID: card, SourceZone: zone.Hand, Card: spell})
	if _, ok, _ := e.PaymentPlanner(g).Complete(req, payment.Partial{}); ok {
		t.Fatal("replaced untap made a tapped creature available for convoke")
	}
	m := cost.Mana{cost.C}
	req = payment.ForGeneric(payment.GenericRequest{PlayerID: game.Player1, Cost: &m})
	plan, ok, reason := e.PaymentPlanner(g).Complete(req, payment.Partial{})
	if !ok {
		t.Fatal(reason)
	}
	if _, ok, reason := e.PaymentPlanner(g).Apply(plan); !ok {
		t.Fatal(reason)
	}
	if !source.Tapped || source.Counters.Has(counter.Stun) || g.Players[0].ManaPool.Total() != 0 {
		t.Fatal("canonical stun replacement and mana production diverged")
	}
}
