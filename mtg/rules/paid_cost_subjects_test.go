package rules

import (
	"slices"
	"testing"

	"github.com/natefinch/council4/cardgen"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/compare"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/id"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/mtg/rules/payment"
	"github.com/natefinch/council4/opt"
)

func compiledPaidSubjectCard(t *testing.T, text, typeLine string) *game.CardDef {
	t.Helper()
	defs, diagnostics, err := cardgen.CompileCardDefs(&cardgen.ScryfallCard{
		Name: "Paid Subject", Layout: "normal", TypeLine: typeLine, OracleText: text,
	})
	if err != nil || len(diagnostics) != 0 || len(defs) != 1 {
		t.Fatalf("compile: %v; diagnostics %#v", err, diagnostics)
	}
	return defs[0]
}

func TestPaidCostActivationFreezesActualSacrifice(t *testing.T) {
	for _, qualifying := range []bool{false, true} {
		t.Run(map[bool]string{false: "predicate false still pays", true: "modified stolen creature"}[qualifying], func(t *testing.T) {
			g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
			engine := NewEngine(nil)
			source := addCombatPermanent(g, game.Player1, compiledPaidSubjectCard(t,
				"Sacrifice a creature: You gain 1 life. If the sacrificed creature was red, draw a card.", "Artifact"))
			paid := addCombatPermanent(g, game.Player2, &game.CardDef{CardFace: game.CardFace{
				Name: "Paid Bear", Types: []types.Card{types.Creature}, Colors: []color.Color{color.Blue},
				Power: opt.Val(game.PT{Value: 0}), Toughness: opt.Val(game.PT{Value: 2}),
			}})
			paid.Controller = game.Player1
			if qualifying {
				g.ContinuousEffects = append(g.ContinuousEffects, game.ContinuousEffect{
					Layer: game.LayerColor, AffectedObjectID: paid.ObjectID,
					SetColors: []color.Color{color.Red},
				})
			}
			drawID := addCardToLibrary(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Drawn", Types: []types.Card{types.Land}}})
			setSorcerySpeedTurn(g, game.Player1)
			life := g.Players[game.Player1].Life
			if !engine.applyAction(g, game.Player1, action.ActivateAbility(source.ObjectID, 0, nil, 0)) {
				t.Fatal("activation refused based on resolving predicate")
			}
			obj, ok := g.Stack.Peek()
			if !ok || len(obj.PaidCostSubjects) != 1 {
				t.Fatal("actual payment was not captured")
			}
			fact := obj.PaidCostSubjects[0]
			if fact.Snapshot.ObjectID != paid.ObjectID || fact.Snapshot.Owner != game.Player2 ||
				fact.Snapshot.Controller != game.Player1 || !fact.Snapshot.Power.Exists || fact.Snapshot.Power.Val != 0 {
				t.Fatalf("incorrect paid snapshot: %#v", fact)
			}
			if !g.Players[game.Player2].Graveyard.Contains(paid.CardInstanceID) {
				t.Fatal("stolen cost did not go to owner's graveyard")
			}
			// Reincarnate the physical card and overwrite global LKI. Neither is
			// the frozen payment subject attached to this stack object.
			card, _ := g.GetCardInstance(paid.CardInstanceID)
			card.Def = &game.CardDef{CardFace: game.CardFace{Name: "Later", Types: []types.Card{types.Artifact}, Colors: []color.Color{color.Black}}}
			delete(g.LastKnownInformation, paid.ObjectID)
			if !moveCardBetweenZones(g, game.Player2, paid.CardInstanceID, zone.Graveyard, zone.Exile) {
				t.Fatal("moving discarded incarnation failed")
			}
			if !slices.Equal(fact.Snapshot.Colors, map[bool][]color.Color{false: {color.Blue}, true: {color.Red}}[qualifying]) {
				t.Fatalf("cost snapshot colors = %v", fact.Snapshot.Colors)
			}
			engine.resolveTopOfStack(g, &TurnLog{})
			if g.Players[game.Player1].Life != life+1 || g.Players[game.Player1].Hand.Contains(drawID) != qualifying {
				t.Fatal("body did not use frozen paid traits and frozen ability controller")
			}
		})
	}
}

func TestPaidCostDiscardCastAndActivation(t *testing.T) {
	for _, spell := range []bool{false, true} {
		for _, land := range []bool{false, true} {
			g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
			engine := NewEngine(nil)
			text := "Discard a card: You gain 1 life. If the discarded card wasn't a land card, draw a card."
			typeLine := "Artifact"
			if spell {
				text = "As an additional cost to cast this spell, discard a card.\nYou gain 1 life. If the discarded card wasn't a land card, draw a card."
				typeLine = "Sorcery"
			}
			def := compiledPaidSubjectCard(t, text, typeLine)
			cardType := types.Instant
			if land {
				cardType = types.Land
			}
			paidID := addCardToHand(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Cost Card", Types: []types.Card{cardType}}})
			card, _ := g.GetCardInstance(paidID)
			card.ZoneVersion = 0
			drawID := addCardToLibrary(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Reward", Types: []types.Card{types.Land}}})
			setSorcerySpeedTurn(g, game.Player1)
			if spell {
				spellID := addCardToHand(g, game.Player1, def)
				if !engine.applyAction(g, game.Player1, action.CastSpell(spellID, nil, 0, nil)) {
					t.Fatal("cast failed")
				}
			} else {
				source := addCombatPermanent(g, game.Player1, def)
				if !engine.applyAction(g, game.Player1, action.ActivateAbility(source.ObjectID, 0, nil, 0)) {
					t.Fatal("activation failed")
				}
			}
			obj, ok := g.Stack.Peek()
			if !ok || len(obj.PaidCostSubjects) != 1 {
				t.Fatal("missing actual discarded cost")
			}
			fact := obj.PaidCostSubjects[0]
			if fact.Snapshot.CardID != paidID || !fact.CardZoneVersion.Exists || fact.CardZoneVersion.Val != 0 {
				t.Fatalf("zero-version paid identity unavailable: %#v", fact)
			}
			card.Def = &game.CardDef{CardFace: game.CardFace{Name: "Later Opposite", Types: []types.Card{types.Land}}}
			card.ZoneVersion += 2
			engine.resolveTopOfStack(g, &TurnLog{})
			if g.Players[game.Player1].Hand.Contains(drawID) == land {
				t.Fatalf("spell=%v land=%v: predicate reread later card", spell, land)
			}
		}
	}
}

func TestPaidCostReferenceRejectsAmbiguityAndUnknown(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	ref := game.PaidCostReference("component", game.PaidCostSacrifice, types.Creature)
	condition := opt.Val(game.Condition{
		Object: opt.Val(ref), ObjectMatches: opt.Val(game.Selection{SubtypesAny: []types.Sub{types.Human}}),
		Negate: true,
	})
	fact := game.PaidCostSubject{
		Key: "component", Kind: game.PaidCostSacrifice, CharacteristicsKnown: true,
		Snapshot: game.ObjectSnapshot{ObjectID: 21, Types: []types.Card{types.Creature}},
	}
	for _, test := range []struct {
		name  string
		facts []game.PaidCostSubject
		want  bool
	}{
		{"known nonhuman", []game.PaidCostSubject{fact}, true},
		{"missing", nil, false},
		{"multiple actual members", []game.PaidCostSubject{fact, fact}, false},
		{"wrong kind", []game.PaidCostSubject{{Key: "component", Kind: game.PaidCostDiscard, CharacteristicsKnown: true, Snapshot: fact.Snapshot}}, false},
		{"unknown traits", []game.PaidCostSubject{{Key: "component", Kind: game.PaidCostSacrifice, Snapshot: fact.Snapshot}}, false},
		{"noun mismatch", []game.PaidCostSubject{{Key: "component", Kind: game.PaidCostSacrifice, CharacteristicsKnown: true, Snapshot: game.ObjectSnapshot{ObjectID: 22, Types: []types.Card{types.Artifact}}}}, false},
		{"different component", []game.PaidCostSubject{{Key: "other", Kind: game.PaidCostSacrifice, CharacteristicsKnown: true, Snapshot: fact.Snapshot}}, false},
	} {
		t.Run(test.name, func(t *testing.T) {
			obj := &game.StackObject{Controller: game.Player1, PaidCostSubjects: test.facts}
			if got := conditionSatisfied(g, conditionContext{controller: game.Player1, obj: obj}, condition); got != test.want {
				t.Fatalf("predicate = %v, want %v", got, test.want)
			}
		})
	}
	fact.Snapshot.Power = opt.Val(0)
	obj := &game.StackObject{PaidCostSubjects: []game.PaidCostSubject{fact}}
	numeric := opt.Val(game.Condition{
		Object: opt.Val(ref), ObjectMatches: opt.Val(game.Selection{Power: opt.Val(compare.Int{Op: compare.Equal, Value: 0})}),
	})
	if !conditionSatisfied(g, conditionContext{obj: obj}, numeric) {
		t.Fatal("known zero power did not match zero")
	}
	obj.PaidCostSubjects[0].Snapshot.Power = opt.V[int]{}
	numeric.Val.Negate = true
	if conditionSatisfied(g, conditionContext{obj: obj}, numeric) {
		t.Fatal("unavailable power matched a negated comparison")
	}
}

func TestPaidCostPaymentFailureAndIndependentComponents(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	source := addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Source", Types: []types.Card{types.Artifact}}})
	paid := addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Bear", Types: []types.Card{types.Creature}}})
	first := cost.Additional{Kind: cost.AdditionalSacrifice, MatchPermanentType: true, PermanentType: types.Creature, SubjectKey: "first"}
	// Failed mana payment must neither consume nor publish a selected cost.
	result, ok := paymentOrch.payAbilityCosts(g, payment.AbilityRequest{
		PlayerID: game.Player1, Source: source, ManaCost: opt.Val(cost.Mana{cost.O(9)}), AdditionalCosts: []cost.Additional{first},
	})
	if ok || len(result.subjects) != 0 || !activeBattlefieldPermanent(paid) {
		t.Fatal("failed payment consumed or published subject")
	}
	cardID := addCardToHand(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Discarded", Types: []types.Card{types.Land}}})
	result, ok = paymentOrch.payAbilityCosts(g, payment.AbilityRequest{
		PlayerID: game.Player1, Source: source, AdditionalCosts: []cost.Additional{first, {Kind: cost.AdditionalDiscard, Source: zone.Hand, SubjectKey: "second"}},
		Prefs: &payment.Preferences{SacrificeChoices: []id.ID{paid.ObjectID}, DiscardChoices: []id.ID{cardID}, StrictReplay: true},
	})
	if !ok || len(result.subjects) != 2 || result.subjects[0].Key != "first" || result.subjects[1].Key != "second" {
		t.Fatalf("components aliased or payment failed: %#v", result)
	}
}

func TestPaidCostSubjectIsNotTheConsumerTarget(t *testing.T) {
	for _, qualifies := range []bool{false, true} {
		g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
		engine := NewEngine(nil)
		source := addCombatPermanent(g, game.Player1, compiledPaidSubjectCard(t,
			"Sacrifice a creature: Tap target artifact. Put a +1/+1 counter on target creature if the sacrificed creature was a Human.", "Artifact"))
		paidSubtype, targetSubtype := types.Bear, types.Human
		if qualifies {
			paidSubtype, targetSubtype = targetSubtype, paidSubtype
		}
		paid := addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
			Name: "Paid", Types: []types.Card{types.Creature}, Subtypes: []types.Sub{paidSubtype},
		}})
		artifact := addCombatPermanent(g, game.Player2, &game.CardDef{CardFace: game.CardFace{Name: "First Target", Types: []types.Card{types.Artifact}}})
		target := addCombatPermanent(g, game.Player2, &game.CardDef{CardFace: game.CardFace{
			Name: "Second Target", Types: []types.Card{types.Creature}, Subtypes: []types.Sub{targetSubtype},
		}})
		setSorcerySpeedTurn(g, game.Player1)
		if !engine.applyAction(g, game.Player1, action.ActivateAbility(source.ObjectID, 0,
			[]game.Target{game.PermanentTarget(artifact.ObjectID), game.PermanentTarget(target.ObjectID)}, 0)) {
			t.Fatal("multi-target activation failed")
		}
		if !g.Players[game.Player1].Graveyard.Contains(paid.CardInstanceID) {
			t.Fatal("wrong actual cost member was paid")
		}
		engine.resolveTopOfStack(g, &TurnLog{})
		expected := 0
		if qualifies {
			expected = 1
		}
		if !artifact.Tapped || target.Counters.Get(counter.PlusOnePlusOne) != expected || source.Counters.Get(counter.PlusOnePlusOne) != 0 {
			t.Fatal("cost subject, unconditional rider and nonzero consumer target were conflated")
		}
	}
}

func TestPaidCostIndependentConsumersShareOnlyCapturedSubject(t *testing.T) {
	for _, colors := range [][]color.Color{{color.Red}, {color.Black}, {color.Red, color.Black}, nil} {
		g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
		engine := NewEngine(nil)
		source := addCombatPermanent(g, game.Player1, compiledPaidSubjectCard(t,
			"Sacrifice a creature: You gain 2 life if the sacrificed creature was red. You gain 3 life if the sacrificed creature was black.", "Artifact"))
		addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Paid", Types: []types.Card{types.Creature}, Colors: colors}})
		setSorcerySpeedTurn(g, game.Player1)
		life := g.Players[game.Player1].Life
		if !engine.applyAction(g, game.Player1, action.ActivateAbility(source.ObjectID, 0, nil, 0)) {
			t.Fatal("activation failed")
		}
		source.Controller = game.Player2
		engine.resolveTopOfStack(g, &TurnLog{})
		expected := life
		if slices.Contains(colors, color.Red) {
			expected += 2
		}
		if slices.Contains(colors, color.Black) {
			expected += 3
		}
		if g.Players[game.Player1].Life != expected {
			t.Fatal("independent paid-subject predicates or ability controller were conflated")
		}
	}
}
