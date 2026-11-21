package rules

import (
	"reflect"
	"slices"
	"strings"
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/id"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/mtg/rules/payment"
	"github.com/natefinch/council4/opt"
)

// Source review: Channeled Dragonfire (oracle 91179937-6f55-4d1b-8f06-2efe1c8fea7a,
// printing 24204881-690c-4043-8771-20cb93385072), pinned corpus SHA-256
// 6da9ced0e19bd384771ab8f14051cc45a8dc90302082b875fa6e039d2399a1b4.
// https://api.scryfall.com/cards/24204881-690c-4043-8771-20cb93385072/rulings
// returned seven rulings: timing still applies, discounts are generic only,
// ordinary increases/reductions apply, and every stack departure exiles it.
// Official CR effective 2026-11-06:
// https://media.wizards.com/2026/downloads/MagicCompRules%2020261106.txt
// 702.180a-b and 601.2b-h: announce the exact optional creature before targets,
// evaluate power when locking the total, and tap the creature during payment.
// 107.1b, 118.7a, 118.9a-d: no negative reduction or colored discount, and
// only one alternative per cast. 707.10/12 and 702.40a: stack and storm copies
// keep choices but are not casts; casting a card copy is a new cast. 608.2b
// covers departure without resolution when every target becomes illegal.
// This is an unregistered Engine fixture for the Channeled Dragonfire play.
func harmonizeFixture() *game.CardDef {
	return &game.CardDef{CardFace: game.CardFace{
		Name: "Channeled Dragonfire", Types: []types.Card{types.Sorcery}, ManaCost: opt.Val(cost.Mana{cost.R}),
		AlternativeCosts: []cost.Alternative{cost.HarmonizeAlternative(cost.Mana{cost.O(5), cost.R, cost.R})},
		SpellAbility: opt.Val(game.Mode{
			Targets: []game.TargetSpec{{MinTargets: 1, MaxTargets: 1, Constraint: "any target"}},
			Sequence: []game.Instruction{{Primitive: game.Damage{
				Amount: game.Fixed(2), Recipient: game.AnyTargetDamageRecipient(0),
			}}},
		}.Ability()),
	}}
}

func harmonizeCreature(g *game.Game, name string, power int) *game.Permanent {
	return addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name: name, Types: []types.Card{types.Creature}, Power: opt.Val(game.PT{Value: power}), Toughness: opt.Val(game.PT{Value: 5}),
	}})
}

func harmonizeCast(card id.ID, from zone.Type) action.Action {
	return action.CastSpellFromZone(card, from, []game.Target{game.PlayerTarget(game.Player2)}, 0, nil)
}

func harmonizeAgent(t *testing.T, subject *game.Permanent, buyback bool) [game.NumPlayers]PlayerAgent {
	t.Helper()
	return [game.NumPlayers]PlayerAgent{game.Player1: &authoredCopyChoiceAgent{
		choose: func(request game.ChoiceRequest) []int {
			if request.Prompt != "Choose spell cost" {
				return request.DefaultSelection
			}
			for _, option := range request.Options {
				if strings.Contains(option.Label, "Buyback") != buyback {
					continue
				}
				if subject == nil && strings.Contains(option.Label, "tap no creature") ||
					subject != nil && option.Card.Exists && option.Card.Val.CardID == subject.CardInstanceID {
					return []int{option.Index}
				}
			}
			t.Fatalf("wanted Harmonize announcement is missing: %+v", request.Options)
			return nil
		},
	}}
}

func TestHarmonizeReductionMustUseChosenCreaturePower(t *testing.T) {
	for _, tc := range []struct {
		name                   string
		generic, red, selected int
		spent                  int
		soleLegalChoice        bool
		noCreatures            bool
	}{
		{name: "no creature", generic: 5, red: 2, spent: 7, noCreatures: true},
		{name: "decline despite creatures", generic: 5, red: 2, spent: 7},
		{name: "choose power one", generic: 4, red: 2, selected: 1, spent: 6},
		{name: "choose power five leaving a larger creature", red: 2, selected: 2, spent: 2},
		{name: "one-power creature cannot borrow power five", red: 2, selected: 2, spent: 2, soleLegalChoice: true},
		{name: "colored shortfall", generic: 6, red: 1},
		{name: "generic shortfall", generic: 3, red: 2, selected: 1},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g, engine := game.NewGame([game.NumPlayers]game.PlayerConfig{}), NewEngine(nil)
			setMainPhasePriority(g, game.Player1)
			player := g.Players[game.Player1]
			def := harmonizeFixture()
			card := addCardToGraveyard(g, game.Player1, def)
			var creatures []*game.Permanent
			if !tc.noCreatures {
				creatures = append(creatures, harmonizeCreature(g, "Small", 1))
				creatures[0].SummoningSick = true
				if tc.name != "generic shortfall" {
					creatures = append(creatures, harmonizeCreature(g, "Large", 5))
				}
				if !tc.soleLegalChoice && tc.selected == 2 {
					creatures = append(creatures, harmonizeCreature(g, "Largest", 8))
				}
			}
			player.ManaPool.Add(mana.R, tc.red)
			player.ManaPool.Add(mana.C, tc.generic)
			var selected *game.Permanent
			var expectedID id.ID
			if tc.selected > 0 {
				selected = creatures[tc.selected-1]
				expectedID = selected.ObjectID
			}
			if tc.soleLegalChoice {
				planner := engine.PaymentPlanner(g)
				req := payment.SpellRequest{PlayerID: game.Player1, CardID: card, Card: def, SourceZone: zone.Graveyard}
				payable := planner.PayableSpellOptions(req)
				if len(payable) != 1 || payable[0].HarmonizeCreature != opt.Val(expectedID) {
					t.Fatalf("{R}{R} should buy exactly the 5-power option, not 1-power or decline: %+v", payable)
				}
				for _, option := range planner.SpellCostOptions(req) {
					_, ok, _ := planner.Complete(payment.ForSpell(req), payment.Partial{Option: opt.Val(option.Index)})
					if ok != (option.HarmonizeCreature == opt.Val(expectedID)) {
						t.Fatalf("proposed %v: payable = %v", option.HarmonizeCreature, ok)
					}
				}
			}
			cast := harmonizeCast(card, zone.Graveyard)
			if listed := containsAction(engine.legalActions(g, game.Player1), cast); listed != (tc.spent != 0) {
				t.Fatalf("Harmonize action listed = %v, want payable = %v", listed, tc.spent != 0)
			}
			if ok := engine.applyActionWithChoices(g, game.Player1, cast, harmonizeAgent(t, selected, false), &TurnLog{}); ok != (tc.spent != 0) {
				t.Fatalf("Harmonize cast = %v, want payable = %v", ok, tc.spent != 0)
			}
			for _, creature := range creatures {
				if creature.Tapped != (tc.spent != 0 && creature == selected) {
					t.Fatalf("wrong tap payment for creature %d", creature.ObjectID)
				}
			}
			if tc.spent == 0 {
				if !player.Graveyard.Contains(card) || spellCastEventCount(g) != 0 {
					t.Fatal("a failed proposal changed the spell's zone or counted as a cast")
				}
				return
			}
			spell, _ := g.Stack.Peek()
			if spell.AlternativeCost != 1 || spell.HarmonizeCreature != opt.Val(expectedID) ||
				spell.ManaSpentToCast != tc.spent || stackObjectKnownManaValue(g, spell) != opt.Val(1) {
				t.Fatalf("wrong Harmonize payment or mana value: %+v", spell)
			}
			engine.resolveTopOfStack(g, &TurnLog{})
			if g.Players[game.Player2].Life != 38 || !player.Exile.Contains(card) {
				t.Fatal("Harmonize spell did not deal 2 damage and go to exile")
			}
		})
	}
}

func TestHarmonizeCombinedCostsAndReplay(t *testing.T) {
	g, engine := game.NewGame([game.NumPlayers]game.PlayerConfig{}), NewEngine(nil)
	setMainPhasePriority(g, game.Player1)
	def := harmonizeFixture()
	def.StaticAbilities = []game.StaticAbility{{KeywordAbilities: []game.KeywordAbility{
		game.BuybackKeyword{Cost: cost.Mana{cost.O(1)}, AdditionalCosts: []cost.Additional{
			{Kind: cost.AdditionalPayLife, Amount: 2, ChoiceGroup: 1},
			{Kind: cost.AdditionalPayLife, Amount: 5, ChoiceGroup: 1},
		}},
	}}}
	def.AdditionalCosts = []cost.Additional{{Kind: cost.AdditionalTapPermanents, Amount: 1,
		MatchPermanentType: true, PermanentType: types.Creature}}
	card := addCardToGraveyard(g, game.Player1, def)
	small := harmonizeCreature(g, "Small", 1)
	large := harmonizeCreature(g, "Large", 6)
	g.CostModifiers = append(g.CostModifiers,
		game.CostModifier{Kind: game.CostModifierSpell, GenericIncrease: 3},
		game.CostModifier{Kind: game.CostModifierSpell, GenericReduction: 2, MinimumGeneric: 1})
	g.Players[game.Player1].ManaPool.Add(mana.R, 2)
	g.Players[game.Player1].ManaPool.Add(mana.C, 1)
	play := func(state *game.Game, agents [game.NumPlayers]PlayerAgent) *TurnLog {
		t.Helper()
		log := &TurnLog{}
		if !engine.applyActionWithChoices(state, game.Player1, harmonizeCast(card, zone.Graveyard), agents, log) {
			t.Fatal("combined Harmonize and buyback cast failed")
		}
		spell, _ := state.Stack.Peek()
		for _, p := range []*game.Permanent{small, large} {
			permanent, ok := permanentByObjectID(state, p.ObjectID)
			if !ok || !permanent.Tapped {
				t.Fatalf("one of the two distinct tap payments is missing: %d", p.ObjectID)
			}
		}
		if spell.AlternativeCost != 1 || spell.HarmonizeCreature != opt.Val(large.ObjectID) ||
			!spell.BuybackPaid || spell.ManaSpentToCast != 3 || state.Players[game.Player1].Life != 38 {
			t.Fatalf("combined costs or original alternative identity were lost: %+v", spell)
		}
		return log
	}
	replay := g.Clone()
	log := play(g, harmonizeAgent(t, large, true))
	cursor := 0
	replayed := play(replay, [game.NumPlayers]PlayerAgent{game.Player1: &authoredCopyChoiceAgent{
		choose: func(req game.ChoiceRequest) []int {
			if cursor >= len(log.Choices) {
				t.Fatalf("unexpected replay prompt: %s", req.Prompt)
			}
			selected := log.Choices[cursor].Selected
			cursor++
			return selected
		},
	}})
	if cursor == 0 || cursor != len(log.Choices) || !reflect.DeepEqual(log.Choices, replayed.Choices) {
		t.Fatal("Harmonize announcement was not faithfully replayed on the clone")
	}
}

func TestHarmonizeStackCopiesAndCounteredOriginal(t *testing.T) {
	g, engine := game.NewGame([game.NumPlayers]game.PlayerConfig{}), NewEngine(nil)
	setMainPhasePriority(g, game.Player1)
	castInstantFor(t, engine, g, game.Player1, "Earlier spell")
	engine.resolveTopOfStack(g, &TurnLog{})
	def := harmonizeFixture()
	def.StaticAbilities = []game.StaticAbility{game.StormStaticBody}
	card := addCardToGraveyard(g, game.Player1, def)
	creature := harmonizeCreature(g, "Large", 5)
	g.Players[game.Player1].ManaPool.Add(mana.R, 2)
	if !engine.applyAction(g, game.Player1, harmonizeCast(card, zone.Graveyard)) {
		t.Fatal("Harmonize plus storm cast failed")
	}
	original, _ := g.Stack.Peek()
	engine.putTriggeredAbilitiesOnStack(g)
	addEffectSpellToStack(g, game.Player3, game.CopyStackObject{
		Object: game.TargetStackObjectReference(0),
	}, []game.Target{game.StackObjectTarget(original.ID)})
	engine.resolveTopOfStack(g, &TurnLog{})
	direct, _ := g.Stack.Peek()
	if !direct.Copy || direct.HarmonizeCreature != opt.Val(creature.ObjectID) ||
		direct.AlternativeCost != 1 || spellCastEvents(g, direct.ID) != 0 {
		t.Fatalf("direct copy did not preserve the original casting choices: %+v", direct)
	}
	engine.resolveTopOfStack(g, &TurnLog{})
	for _, branch := range []*game.Game{g, g.Clone()} {
		if !counterStackObjectWithDestination(branch, original.ID, false, game.CounteredSpellHand) ||
			!branch.Players[game.Player1].Exile.Contains(card) {
			t.Fatal("a countered Harmonize spell did not get exiled instead of returning to hand")
		}
		engine.resolveTopOfStack(branch, &TurnLog{})
		copies := stormCopiesOf(branch, card)
		if len(copies) != 1 || copies[0].HarmonizeCreature != opt.Val(creature.ObjectID) ||
			copies[0].AlternativeCost != 1 || copies[0].CastCopy ||
			!slices.Equal(copies[0].Targets, original.Targets) || spellCastEventCount(branch) != 2 {
			t.Fatalf("storm copy lost casting decisions, paid again, or counted as cast: %+v", copies)
		}
		engine.resolveTopOfStack(branch, &TurnLog{})
		if branch.Players[game.Player2].Life != 36 || !branch.Players[game.Player1].Exile.Contains(card) {
			t.Fatal("storm copy did not resolve independently of the exiled card")
		}
	}
}

func TestHarmonizeLocksEffectivePowerBeforeManaAbilities(t *testing.T) {
	g, engine := game.NewGame([game.NumPlayers]game.PlayerConfig{}), NewEngine(nil)
	setMainPhasePriority(g, game.Player1)
	card := addCardToGraveyard(g, game.Player1, harmonizeFixture())
	creature := harmonizeCreature(g, "Chosen creature", 4)
	source := addComplexManaAbilityPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name: "Sacrificial anthem", Types: []types.Card{types.Artifact},
		StaticAbilities: []game.StaticAbility{{ContinuousEffects: []game.ContinuousEffect{{
			Duration: game.DurationPermanent, Layer: game.LayerPowerToughnessModify,
			Group:      game.ObjectControlledGroup(game.SourcePermanentReference(), game.Selection{RequiredTypes: []types.Card{types.Creature}}),
			PowerDelta: 1,
		}}}},
	}}, &game.ManaAbility{
		Text:            "Sacrifice this artifact: Add {R}.",
		AdditionalCosts: []cost.Additional{{Kind: cost.AdditionalSacrificeSource, Amount: 1}},
		Content: game.Mode{Sequence: []game.Instruction{{Primitive: game.AddMana{
			Player: opt.Val(game.ControllerReference()), ManaColor: mana.R, Amount: game.Fixed(1),
		}}}}.Ability(),
	})
	player := g.Players[game.Player1]
	player.ManaPool.Add(mana.R, 1)
	if !engine.applyAction(g, game.Player1, harmonizeCast(card, zone.Graveyard)) {
		t.Fatal("losing the anthem while making the second red mana changed the locked total")
	}
	spell, _ := g.Stack.Peek()
	if !creature.Tapped || !player.Graveyard.Contains(source.CardInstanceID) ||
		spell.HarmonizeCreature != opt.Val(creature.ObjectID) || spell.ManaSpentToCast != 2 {
		t.Fatal("did not lock power 5, sacrifice the anthem for mana, then tap the chosen creature")
	}
}

func TestHarmonizeAndIndependentFreeGraveyardRoute(t *testing.T) {
	for _, harmonize := range []bool{false, true} {
		g, engine := game.NewGame([game.NumPlayers]game.PlayerConfig{}), NewEngine(nil)
		setMainPhasePriority(g, game.Player1)
		player := g.Players[game.Player1]
		card := addCardToGraveyard(g, game.Player1, harmonizeFixture())
		creature := harmonizeCreature(g, "Large", 5)
		player.ManaPool.Add(mana.R, 2)
		g.RuleEffects = append(g.RuleEffects, game.RuleEffect{
			ID: g.IDGen.Next(), Kind: game.RuleEffectPlayFromZone, Controller: game.Player1,
			AffectedPlayer: game.PlayerYou, Duration: game.DurationPermanent, CastFromZone: zone.Graveyard,
			AffectedCardID: card, WithoutPayingManaCost: true,
		})
		agents := [game.NumPlayers]PlayerAgent{game.Player1: &authoredCopyChoiceAgent{
			choose: func(req game.ChoiceRequest) []int {
				if req.Prompt == "Choose spell cost" {
					if len(req.Options) != 2 {
						t.Fatalf("expected separate free and Harmonize routes, got %+v", req.Options)
					}
					for _, option := range req.Options {
						if strings.Contains(option.Label, "Harmonize") == harmonize {
							return []int{option.Index}
						}
					}
					t.Fatalf("expected route missing: %+v", req.Options)
				}
				return req.DefaultSelection
			},
		}}
		log := &TurnLog{}
		if !engine.applyActionWithChoices(g, game.Player1, harmonizeCast(card, zone.Graveyard), agents, log) {
			t.Fatal("graveyard route failed")
		}
		spell, _ := g.Stack.Peek()
		wantSpent := 0
		if harmonize {
			wantSpent = 2
		}
		if !slices.ContainsFunc(log.Choices, func(c game.ChoiceDecision) bool { return c.Request.Prompt == "Choose spell cost" }) ||
			spell.HarmonizeCreature.Exists != harmonize || creature.Tapped != harmonize || spell.ManaSpentToCast != wantSpent {
			t.Fatalf("routes mixed their payment or choice: %+v", spell)
		}
		engine.resolveTopOfStack(g, log)
		if player.Exile.Contains(card) != harmonize || player.Graveyard.Contains(card) == harmonize {
			t.Fatal("wrong stack departure for the selected route")
		}
	}
}

func TestHarmonizeDuringPaidResolutionAndIllegalTarget(t *testing.T) {
	g, engine := game.NewGame([game.NumPlayers]game.PlayerConfig{}), NewEngine(nil)
	setMainPhasePriority(g, game.Player2)
	player := g.Players[game.Player1]
	card := addCardToGraveyard(g, game.Player1, harmonizeFixture())
	creature := harmonizeCreature(g, "Tap me, then target me", 5)
	player.ManaPool.Add(mana.R, 2)
	agents := [game.NumPlayers]PlayerAgent{game.Player1: &authoredCopyChoiceAgent{
		choose: func(req game.ChoiceRequest) []int {
			if req.Kind == game.ChoiceTarget {
				for _, option := range req.Options {
					if slices.Equal(option.Targets, []game.Target{game.PermanentTarget(creature.ObjectID)}) {
						return []int{option.Index}
					}
				}
				t.Fatal("an announced cost subject should remain targetable")
			}
			return req.DefaultSelection
		},
	}}
	if !engine.castPaidTargetedSpell(g, game.Player1, card, zone.Graveyard, agents, &TurnLog{}) {
		t.Fatal("paid cast during resolution did not offer Harmonize")
	}
	spell, _ := g.Stack.Peek()
	if spell.HarmonizeCreature != opt.Val(creature.ObjectID) || spell.AlternativeCost != 1 ||
		!creature.Tapped || !slices.Equal(spell.Targets, []game.Target{game.PermanentTarget(creature.ObjectID)}) {
		t.Fatalf("paid effect cast lost announcement or tap: %+v", spell)
	}
	if _, ok := destroyPermanent(g, creature.ObjectID); !ok {
		t.Fatal("could not remove the only target")
	}
	engine.resolveTopOfStack(g, &TurnLog{})
	if !player.Exile.Contains(card) || countEventsOfKind(g.Events, game.EventSpellResolved) != 0 {
		t.Fatal("a Harmonize spell with no legal targets did not go directly to exile")
	}
}

func TestHarmonizeDoesNotApplyToHandOrFreeCardCopy(t *testing.T) {
	for _, from := range []zone.Type{zone.Hand, zone.Exile} {
		t.Run(from.String(), func(t *testing.T) {
			g, engine := game.NewGame([game.NumPlayers]game.PlayerConfig{}), NewEngine(nil)
			setMainPhasePriority(g, game.Player1)
			creature := harmonizeCreature(g, "Large", 5)
			def := harmonizeFixture()
			var spell *game.StackObject
			if from == zone.Hand {
				card := addCardToHand(g, game.Player1, def)
				g.Players[game.Player1].ManaPool.Add(mana.R, 1)
				if !engine.applyAction(g, game.Player1, harmonizeCast(card, zone.Hand)) {
					t.Fatal("ordinary cast failed")
				}
				spell, _ = g.Stack.Peek()
				engine.resolveTopOfStack(g, &TurnLog{})
				if !g.Players[game.Player1].Graveyard.Contains(card) || spell.ManaSpentToCast != 1 {
					t.Fatal("ordinary cast used the wrong cost or destination")
				}
			} else {
				card := addCardToExile(g, game.Player1, def)
				spell = castCopyOfExiledCard(t, engine, g, card)
				if spell.ManaSpentToCast != 0 || !g.Players[game.Player1].Exile.Contains(card) {
					t.Fatal("free copy paid mana or moved the physical card")
				}
			}
			if creature.Tapped || spell.HarmonizeCreature.Exists || spell.AlternativeCost != 0 || spellCastEvents(g, spell.ID) != 1 {
				t.Fatal("other cast route used Harmonize or was not counted as a cast")
			}
		})
	}
}
