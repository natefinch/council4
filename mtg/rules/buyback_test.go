package rules

import (
	"slices"
	"strings"
	"testing"

	"github.com/natefinch/council4/mtg/cardprogram"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/mtg/rules/payment"
	"github.com/natefinch/council4/opt"
)

// Minimal Engine proving fixture, not the production registry definition.
// Named production Reiterate scenarios are in buyback_interaction_test.go.
// Oracle ID 38394ee3-6726-4f91-bc25-36bce0c6aab9; printing
// 99302c41-434f-41ff-a61a-2c3681a0c135; text SHA-256
// d3770caf39a9493ba336c30a906cd13abd83176c4cb886158501096f59d1cc41.
// Corpus SHA-256 6da9ced0e19bd384771ab8f14051cc45a8dc90302082b875fa6e039d2399a1b4.
// Read the complete seven-ruling response on 2026-10-23:
// https://api.scryfall.com/cards/99302c41-434f-41ff-a61a-2c3681a0c135/rulings
// All seven are WotC 2021-03-19: copies are not casts and retain modes, X,
// targets, divisions and additional-cost decisions, without paying again.
// Official CR effective 2026-09-25, verified supplied high-review-rules.txt
// SHA-256 8d860e451f20f38865b725b42d82feb714c725373dd8f3b32b8652b3eeb070ca:
// read 601.2a-i (announce and pay combined costs), 608.2a-p (all-illegal
// spells do not resolve; final graveyard move), 702.27a (resolution-only
// buyback), 707.10/a-c (copy choices/subjects, no physical card), 400.7
// (new zone incarnations), 614.15 and 616.1a-g (buyback is not a resolving
// self-replacement; the affected spell's controller orders competing redirects),
// and 702.34a (Flashback exiles any departure from the stack).
// Repair review 2026-10-23 additionally read 707.12/a (casting a card copy
// follows ordinary casting/payment) and 118.9d (additional costs and modifiers
// still apply when casting without paying the mana cost).
// Further repair review read 702.85a: a genuinely cast card copy triggers
// cascade; a non-cast stack/storm copy does not (707.10/12).
// Read 603.2/3 for shared cast-event capture: cast-only and self-cast abilities
// see genuine card-copy casts, while non-cast copies only produce copy events.
const reiterateFixtureText = "Buyback {3} (You may pay an additional {3} as you cast this spell. If you do, put this card into your hand as it resolves.)\nCopy target instant or sorcery spell. You may choose new targets for the copy."

var reiterateFixtureProgram = game.ProgramReference{ID: "test/reiterate/spell/0", Version: "1"}

func reiterateFixture() *game.CardDef {
	return &game.CardDef{ColorIdentity: color.NewIdentity(color.Red), CardFace: game.CardFace{
		Name: "Reiterate", ManaCost: opt.Val(cost.Mana{cost.O(1), cost.R, cost.R}),
		Types: []types.Card{types.Instant}, Colors: []color.Color{color.Red},
		OracleText: reiterateFixtureText,
		StaticAbilities: []game.StaticAbility{{
			Text: "Buyback {3}", ZoneOfFunction: zone.Stack,
			KeywordAbilities: []game.KeywordAbility{game.BuybackKeyword{Cost: cost.Mana{cost.O(3)}}},
		}},
		SpellAbility: opt.Val(game.Mode{
			Text: "Copy target instant or sorcery spell. You may choose new targets for the copy.",
			Targets: []game.TargetSpec{{
				MinTargets: 1, MaxTargets: 1, Allow: game.TargetAllowStackObject,
				Predicate: game.TargetPredicate{
					StackObjectKinds:  []game.StackObjectKind{game.StackSpell},
					SpellCardTypesAny: []types.Card{types.Instant, types.Sorcery},
				},
			}},
			Program: reiterateFixtureProgram,
		}.Ability()),
	}}
}

func buybackEngine(t *testing.T) *Engine {
	t.Helper()
	catalog, err := cardprogram.NewCatalog([]cardprogram.Program{{
		Reference: reiterateFixtureProgram,
		Resolve: func(ctx cardprogram.Context) error {
			_, err := ctx.Copy(game.CopyStackObject{
				Object: game.TargetStackObjectReference(0), MayChooseNewTargets: true,
			}, ctx.View().Source().Controller)
			return err
		},
	}})
	if err != nil {
		t.Fatal(err)
	}
	return NewEngineWithPrograms(nil, catalog)
}

func castBuybackSubject(t *testing.T, g *game.Game, engine *Engine) *game.StackObject {
	t.Helper()
	def := &game.CardDef{CardFace: game.CardFace{
		Name: "Untargeted buyback subject", Types: []types.Card{types.Instant},
		ManaCost: opt.Val(cost.Mana{cost.O(0)}),
		SpellAbility: opt.Val(game.Mode{Sequence: []game.Instruction{{
			Primitive: game.GainLife{Amount: game.Fixed(4), Player: game.ControllerReference()},
		}}}.Ability()),
	}}
	card := addCardToHand(g, game.Player2, def)
	setMainPhasePriority(g, game.Player2)
	if !engine.applyAction(g, game.Player2, action.CastSpell(card, nil, 0, nil)) {
		t.Fatal("subject's real cast failed")
	}
	obj, _ := g.Stack.Peek()
	g.Turn.PriorityPlayer = game.Player1
	return obj
}

func buybackAgents(t *testing.T, paid bool) [game.NumPlayers]PlayerAgent {
	t.Helper()
	return [game.NumPlayers]PlayerAgent{game.Player1: &authoredCopyChoiceAgent{
		choose: func(request game.ChoiceRequest) []int {
			if request.Kind == game.ChoicePayment {
				for _, option := range request.Options {
					if strings.Contains(option.Label, "Buyback") == paid {
						return []int{option.Index}
					}
				}
				t.Fatalf("requested buyback payment branch missing: %+v", request)
			}
			return request.DefaultSelection
		},
	}}
}

func fundBuyback(g *game.Game, generic int) {
	g.Players[game.Player1].ManaPool.Add(mana.R, 2)
	g.Players[game.Player1].ManaPool.Add(mana.C, generic)
}

func TestReiterateBuybackOptionalPaymentAndResolution(t *testing.T) {
	t.Parallel()
	for _, tc := range []struct {
		name    string
		generic int
		paid    bool
		choices int
		spent   int
	}{
		{name: "paid", generic: 4, paid: true, choices: 1, spent: 6},
		{name: "declined", generic: 4, choices: 1, spent: 3},
		{name: "unaffordable optional cost", generic: 1, spent: 3},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
			engine := buybackEngine(t)
			subject := castBuybackSubject(t, g, engine)
			cardID := addCardToHand(g, game.Player1, reiterateFixture())
			fundBuyback(g, tc.generic)
			log := &TurnLog{}
			agents := buybackAgents(t, tc.paid)
			if !engine.applyActionWithChoices(g, game.Player1, action.CastSpell(cardID,
				[]game.Target{game.StackObjectTarget(subject.ID)}, 0, nil), agents, log) {
				t.Fatal("Reiterate real cast failed")
			}
			obj, _ := g.Stack.Peek()
			if obj.BuybackPaid != tc.paid || obj.ManaSpentToCast != tc.spent ||
				g.Players[game.Player1].ManaPool.Total() != tc.generic+2-tc.spent ||
				len(log.Choices) != tc.choices {
				t.Fatalf("payment choice or combined cost wrong: object=%+v choices=%+v", obj, log.Choices)
			}
			for _, choice := range log.Choices {
				if choice.UsedFallback {
					t.Fatal("buyback choice bypassed its agent")
				}
			}
			engine.resolveTopOfStackWithChoices(g, agents, log)
			if g.Players[game.Player1].Hand.Contains(cardID) != tc.paid ||
				g.Players[game.Player1].Graveyard.Contains(cardID) == tc.paid ||
				g.Stack.Size() != 2 {
				t.Fatal("paid/unpaid destination or copy creation wrong")
			}
			copyObj, _ := g.Stack.Peek()
			if !copyObj.Copy || copyObj.Controller != game.Player1 {
				t.Fatal("Reiterate did not create the target spell's copy")
			}
			engine.resolveTopOfStack(g, log)
			engine.resolveTopOfStack(g, log)
			if g.Players[game.Player1].Life != 44 || g.Players[game.Player2].Life != 44 ||
				countEventsOfKind(g.Events, game.EventSpellCast) != 2 ||
				countEventsOfKind(g.Events, game.EventSpellCopied) != 1 ||
				countEventsOfKind(g.Events, game.EventSpellResolved) != 3 ||
				g.CardInstances[cardID].ZoneVersion != 2 {
				t.Fatal("copy effects, cast/resolution events or source incarnation wrong")
			}
		})
	}
}

func TestReiterateBuybackCounterAndAllIllegal(t *testing.T) {
	t.Parallel()
	for _, countered := range []bool{false, true} {
		name := "all illegal"
		if countered {
			name = "countered"
		}
		t.Run(name, func(t *testing.T) {
			g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
			engine := buybackEngine(t)
			subject := castBuybackSubject(t, g, engine)
			cardID := addCardToHand(g, game.Player1, reiterateFixture())
			fundBuyback(g, 4)
			if !engine.applyActionWithChoices(g, game.Player1, action.CastSpell(cardID,
				[]game.Target{game.StackObjectTarget(subject.ID)}, 0, nil), buybackAgents(t, true), &TurnLog{}) {
				t.Fatal("paid Reiterate cast failed")
			}
			obj, _ := g.Stack.Peek()
			target := subject.ID
			if countered {
				target = obj.ID
			}
			addEffectSpellToStack(g, game.Player3, game.CounterObject{
				Object: game.TargetStackObjectReference(0),
			}, []game.Target{game.StackObjectTarget(target)})
			engine.resolveTopOfStack(g, &TurnLog{})
			if !countered {
				engine.resolveTopOfStack(g, &TurnLog{})
			}
			if !obj.BuybackPaid || !g.Players[game.Player1].Graveyard.Contains(cardID) ||
				g.Players[game.Player1].Hand.Contains(cardID) ||
				countEventsOfKind(g.Events, game.EventSpellCopied) != 0 {
				t.Fatal("a spell that did not resolve returned via buyback or copied its target")
			}
		})
	}
}

func TestReiterateBuybackCopiesClonesAndRecast(t *testing.T) {
	t.Parallel()
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := buybackEngine(t)
	subject := castBuybackSubject(t, g, engine)
	def := reiterateFixture()
	// A separate mandatory payment subject proves copying preserves paid facts,
	// rather than re-paying or looking up the now-departed incarnation.
	def.AdditionalCosts = []cost.Additional{{
		Kind: cost.AdditionalDiscard, Amount: 1, SubjectKey: "fixture-discard",
	}}
	discarded := addCardToHand(g, game.Player1, vanillaCreature("Discarded", 2, 2))
	cardID := addCardToHand(g, game.Player1, def)
	fundBuyback(g, 4)
	agents := buybackAgents(t, true)
	if !engine.applyActionWithChoices(g, game.Player1, action.CastSpell(cardID,
		[]game.Target{game.StackObjectTarget(subject.ID)}, 0, nil), agents, &TurnLog{}) {
		t.Fatal("paid Reiterate with mandatory discard failed")
	}
	original, _ := g.Stack.Peek()
	if len(original.PaidCostSubjects) != 1 || original.PaidCostSubjects[0].Snapshot.CardID != discarded {
		t.Fatal("cast did not retain the actual discard payment")
	}
	addEffectSpellToStack(g, game.Player3, game.CopyStackObject{
		Object: game.TargetStackObjectReference(0),
	}, []game.Target{game.StackObjectTarget(original.ID)})
	engine.resolveTopOfStack(g, &TurnLog{})
	copyObj, _ := g.Stack.Peek()
	if !copyObj.Copy || !copyObj.BuybackPaid || len(copyObj.PaidCostSubjects) != 1 ||
		copyObj.PaidCostSubjects[0].Snapshot.CardID != discarded {
		t.Fatal("stack copy lost buyback announcement or original payment subjects")
	}
	clone := g.Clone()
	cloneCopy, _ := clone.Stack.Peek()
	cloneCopy.PaidCostSubjects[0].Key = "clone-only"
	if copyObj.PaidCostSubjects[0].Key != "fixture-discard" {
		t.Fatal("clone aliased the original payment subjects")
	}
	for _, branch := range []*game.Game{g, clone} {
		engine.resolveTopOfStackWithChoices(branch, agents, &TurnLog{})
		if branch.Players[game.Player1].Hand.Contains(cardID) ||
			branch.Players[game.Player3].Hand.Contains(cardID) ||
			branch.Stack.Size() != 3 || branch.Players[game.Player1].ManaPool.Total() != 0 {
			t.Fatal("copied Reiterate re-paid costs or returned the original's physical card")
		}
		engine.resolveTopOfStack(branch, &TurnLog{})                    // copied subject
		engine.resolveTopOfStackWithChoices(branch, agents, &TurnLog{}) // original Reiterate
		if !branch.Players[game.Player1].Hand.Contains(cardID) ||
			branch.CardInstances[cardID].ZoneVersion != 2 ||
			branch.Players[game.Player3].Life != 44 {
			t.Fatal("copy-relative effects or original buyback return wrong")
		}
		// A new cast of the same physical card starts without the old buyback
		// announcement, even though an old copy inherited it.
		addCardToHand(branch, game.Player1, vanillaCreature("Second discard", 1, 1))
		fundBuyback(branch, 1)
		branch.Turn.PriorityPlayer = game.Player1
		if !engine.applyActionWithChoices(branch, game.Player1, action.CastSpell(cardID,
			[]game.Target{game.StackObjectTarget(subject.ID)}, 0, nil), buybackAgents(t, false), &TurnLog{}) {
			t.Fatal("unpaid recast failed")
		}
		recast, _ := branch.Stack.Peek()
		if recast.BuybackPaid || recast.ID == original.ID ||
			branch.CardInstances[cardID].ZoneVersion != 3 {
			t.Fatal("new source incarnation inherited old buyback state")
		}
		engine.resolveTopOfStack(branch, &TurnLog{})
		if !branch.Players[game.Player1].Graveyard.Contains(cardID) ||
			branch.CardInstances[cardID].ZoneVersion != 4 {
			t.Fatal("unpaid new incarnation did not reach graveyard")
		}
	}
}

func TestBuybackCompetesWithGraveyardReplacement(t *testing.T) {
	t.Parallel()
	for _, choose := range []string{"Buyback", "Graveyard exile"} {
		t.Run(choose, func(t *testing.T) {
			g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
			engine := buybackEngine(t)
			subject := castBuybackSubject(t, g, engine)
			cardID := addCardToHand(g, game.Player1, reiterateFixture())
			g.CardInstances[cardID].Owner = game.Player4
			fundBuyback(g, 4)
			if !engine.applyActionWithChoices(g, game.Player1, action.CastSpell(cardID,
				[]game.Target{game.StackObjectTarget(subject.ID)}, 0, nil), buybackAgents(t, true), &TurnLog{}) {
				t.Fatal("paid cast failed")
			}
			g.ReplacementEffects = append(g.ReplacementEffects, game.ReplacementEffect{
				ID: g.IDGen.Next(), Description: "Graveyard exile", MatchEvent: game.EventZoneChanged,
				MatchToZone: true, ToZone: zone.Graveyard, ReplaceToZone: zone.Exile,
			})
			log := &TurnLog{}
			choices := 0
			agents := [game.NumPlayers]PlayerAgent{game.Player1: &authoredCopyChoiceAgent{
				choose: func(request game.ChoiceRequest) []int {
					if request.Kind == game.ChoiceReplacement {
						choices++
						if request.Player != game.Player1 || len(request.Options) != 2 {
							t.Fatal("spell controller did not choose both competing replacements")
						}
						for _, option := range request.Options {
							if option.Label == choose {
								return []int{option.Index}
							}
						}
						t.Fatal("desired replacement absent")
					}
					return request.DefaultSelection
				},
			}}
			engine.resolveTopOfStackWithChoices(g, agents, log)
			if choices != 1 || g.Players[game.Player4].Hand.Contains(cardID) != (choose == "Buyback") ||
				g.Players[game.Player4].Exile.Contains(cardID) != (choose != "Buyback") ||
				g.Players[game.Player1].Hand.Contains(cardID) || len(g.ReplacementEffects) != 1 {
				t.Fatal("replacement ordering, physical owner or local effect lifetime wrong")
			}
		})
	}
}

func TestReiterateBuybackCopyOutlivesOriginalIncarnation(t *testing.T) {
	t.Parallel()
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := buybackEngine(t)
	subject := castBuybackSubject(t, g, engine)
	cardID := addCardToHand(g, game.Player1, reiterateFixture())
	fundBuyback(g, 4)
	if !engine.applyActionWithChoices(g, game.Player1, action.CastSpell(cardID,
		[]game.Target{game.StackObjectTarget(subject.ID)}, 0, nil), buybackAgents(t, true), &TurnLog{}) {
		t.Fatal("paid cast failed")
	}
	original, _ := g.Stack.Peek()
	addEffectSpellToStack(g, game.Player3, game.CopyStackObject{
		Object: game.TargetStackObjectReference(0),
	}, []game.Target{game.StackObjectTarget(original.ID)})
	engine.resolveTopOfStack(g, &TurnLog{})
	copyObj, _ := g.Stack.Peek()
	if !copyObj.BuybackPaid || !bounceStackSpellToHand(g, original) {
		t.Fatal("could not copy and return the original spell")
	}
	fundBuyback(g, 1)
	g.Turn.PriorityPlayer = game.Player1
	if !engine.applyAction(g, game.Player1, action.CastSpell(cardID,
		[]game.Target{game.StackObjectTarget(subject.ID)}, 0, nil)) {
		t.Fatal("new unpaid cast failed")
	}
	engine.resolveTopOfStack(g, &TurnLog{}) // new Reiterate incarnation
	engine.resolveTopOfStack(g, &TurnLog{}) // new subject copy
	if !g.Players[game.Player1].Graveyard.Contains(cardID) || g.CardInstances[cardID].ZoneVersion != 4 {
		t.Fatal("new incarnation did not go to the graveyard")
	}
	engine.resolveTopOfStack(g, &TurnLog{}) // old paid Reiterate copy
	if !g.Players[game.Player1].Graveyard.Contains(cardID) ||
		g.CardInstances[cardID].ZoneVersion != 4 || g.Players[game.Player1].ManaPool.Total() != 0 {
		t.Fatal("old copy moved or re-paid the newer physical incarnation")
	}
	created, _ := g.Stack.Peek()
	if !created.Copy || created.Controller != game.Player3 || g.Stack.Size() != 2 {
		t.Fatal("copy lost its target body after the physical source left the stack")
	}
}

func TestBuybackCombinedKickerFlashbackAndFreeCast(t *testing.T) {
	t.Parallel()
	for _, path := range []string{"kicker", "flashback", "free", "paid during resolution"} {
		t.Run(path, func(t *testing.T) {
			g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
			engine := buybackEngine(t)
			subject := castBuybackSubject(t, g, engine)
			def := reiterateFixture()
			def.StaticAbilities = append(def.StaticAbilities, game.StaticAbility{
				KeywordAbilities: []game.KeywordAbility{game.KickerKeyword{Cost: cost.Mana{cost.O(1)}}},
			})
			source := zone.Hand
			if path == "flashback" {
				source = zone.Graveyard
				def.StaticAbilities = append(def.StaticAbilities, game.StaticAbility{
					KeywordAbilities: game.SimpleKeywords(game.Flashback),
				})
				def.AlternativeCosts = []cost.Alternative{{
					Label: "Flashback", Mechanic: cost.AlternativeMechanicFlashback,
					ManaCost: opt.Val(cost.Mana{cost.R, cost.R}),
				}}
			}
			cardID := addCardToHand(g, game.Player1, def)
			if source == zone.Graveyard {
				g.Players[game.Player1].Hand.Remove(cardID)
				g.Players[game.Player1].Graveyard.Add(cardID)
			}
			fundBuyback(g, 5)
			agents := buybackAgents(t, true)
			targets := []game.Target{game.StackObjectTarget(subject.ID)}
			cast := action.CastSpellAction{CardID: cardID, SourceZone: source, Targets: targets}
			wantSpent := 6
			ok := false
			if path == "kicker" {
				cast.KickerPaid = true
				wantSpent = 7
			}
			if path == "flashback" {
				wantSpent = 5
			}
			if path == "free" {
				wantSpent = 3
				ok = engine.castFreeOption(g, g.Players[game.Player1], game.Player1,
					g.CardInstances[cardID], cast, false, agents, &TurnLog{})
			} else if path == "paid during resolution" {
				ok = engine.castPaidSpellFromSource(g, g.Players[game.Player1], game.Player1,
					cardID, source, agents, &TurnLog{})
			} else {
				ok = engine.applyCastSpellWithChoices(g, game.Player1, cast, agents, &TurnLog{})
			}
			if !ok {
				t.Fatal("combined buyback cast failed")
			}
			obj, _ := g.Stack.Peek()
			if !obj.BuybackPaid || obj.ManaSpentToCast != wantSpent ||
				obj.KickerPaid != (path == "kicker") || obj.Flashback != (path == "flashback") {
				t.Fatalf("combined cost or announced state wrong: %+v", obj)
			}
			engine.resolveTopOfStackWithChoices(g, agents, &TurnLog{})
			if g.Players[game.Player1].Exile.Contains(cardID) != (path == "flashback") ||
				g.Players[game.Player1].Hand.Contains(cardID) != (path != "flashback") {
				t.Fatal("Flashback or buyback destination changed")
			}
		})
	}
}

func TestBuybackFailedPaymentIsAtomic(t *testing.T) {
	t.Parallel()
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := buybackEngine(t)
	subject := castBuybackSubject(t, g, engine)
	cardID := addCardToHand(g, game.Player1, reiterateFixture())
	fundBuyback(g, 4)
	request := payment.SpellRequest{PlayerID: game.Player1, CardID: cardID,
		Card: g.CardInstances[cardID].Def, SourceZone: zone.Hand}
	options := paymentOrch.planner(g).PayableSpellOptions(request)
	if len(options) != 2 {
		t.Fatal("otherwise legal paid/unpaid choices absent")
	}
	request.Prefs = &payment.Preferences{AlternativeIndex: options[1].Index}
	g.Players[game.Player1].ManaPool = mana.NewPool()
	fundBuyback(g, 1)
	before := g.Clone()
	if _, ok := paymentOrch.paySpellCosts(g, request); ok {
		t.Fatal("stale buyback payment silently fell back to unpaid")
	}
	if !slices.Equal(g.Players[game.Player1].Hand.All(), before.Players[game.Player1].Hand.All()) ||
		g.Players[game.Player1].ManaPool.Total() != 3 || g.Stack.Size() != 1 ||
		len(g.Events) != len(before.Events) {
		t.Fatal("unpayable buyback changed resources or zones")
	}
	g.Players[game.Player1].ManaPool = mana.NewPool()
	g.Players[game.Player1].ManaPool.Add(mana.R, 2)
	if engine.applyActionWithChoices(g, game.Player1, action.CastSpell(cardID,
		[]game.Target{game.StackObjectTarget(subject.ID)}, 0, nil), buybackAgents(t, true), &TurnLog{}) ||
		!g.Players[game.Player1].Hand.Contains(cardID) || g.Stack.Size() != 1 ||
		g.Players[game.Player1].ManaPool.Total() != 2 || g.CardInstances[cardID].ZoneVersion != 0 {
		t.Fatal("unpayable base cost mutated the game")
	}
}

func TestBuybackNonManaComponentUsesActualPaymentSubjects(t *testing.T) {
	t.Parallel()
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	def := &game.CardDef{CardFace: game.CardFace{
		Name: "Non-mana buyback probe", Types: []types.Card{types.Instant},
		ManaCost: opt.Val(cost.Mana{cost.O(0)}),
		StaticAbilities: []game.StaticAbility{{
			KeywordAbilities: []game.KeywordAbility{game.BuybackKeyword{
				AdditionalCosts: []cost.Additional{{
					Kind: cost.AdditionalDiscard, Amount: 1, SubjectKey: "buyback-discard",
				}},
			}},
		}},
		SpellAbility: opt.Val(game.Mode{Sequence: []game.Instruction{{
			Primitive: game.GainLife{Amount: game.Fixed(1), Player: game.ControllerReference()},
		}}}.Ability()),
	}}
	cardID := addCardToHand(g, game.Player1, def)
	setMainPhasePriority(g, game.Player1)
	request := payment.SpellRequest{PlayerID: game.Player1, CardID: cardID,
		Card: def, SourceZone: zone.Hand}
	if options := paymentOrch.planner(g).PayableSpellOptions(request); len(options) != 1 {
		t.Fatal("buyback payment selected the card being cast as its own discard")
	}
	discarded := addCardToHand(g, game.Player1, vanillaCreature("Discard subject", 1, 1))
	if !engine.applyActionWithChoices(g, game.Player1, action.CastSpell(cardID, nil, 0, nil),
		buybackAgents(t, true), &TurnLog{}) {
		t.Fatal("non-mana buyback cast failed")
	}
	obj, _ := g.Stack.Peek()
	if !obj.BuybackPaid || len(obj.PaidCostSubjects) != 1 ||
		obj.PaidCostSubjects[0].Key != "buyback-discard" ||
		obj.PaidCostSubjects[0].Snapshot.CardID != discarded ||
		!g.Players[game.Player1].Graveyard.Contains(discarded) {
		t.Fatal("non-mana buyback cost lost its actual payment subject")
	}
	engine.resolveTopOfStack(g, &TurnLog{})
	if g.Players[game.Player1].Life != 41 || !g.Players[game.Player1].Hand.Contains(cardID) {
		t.Fatal("non-mana buyback did not resolve and return")
	}
}

func TestReiterateBuybackTotalIncludesTaxesAndReductions(t *testing.T) {
	t.Parallel()
	for _, tc := range []struct {
		name                                string
		increase, reduction, generic, spent int
	}{
		{name: "tax", increase: 2, generic: 6, spent: 8},
		{name: "reduction", reduction: 2, generic: 2, spent: 4},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
			engine := buybackEngine(t)
			subject := castBuybackSubject(t, g, engine)
			cardID := addCardToHand(g, game.Player1, reiterateFixture())
			g.CostModifiers = []game.CostModifier{{
				Kind: game.CostModifierSpell, GenericIncrease: tc.increase, GenericReduction: tc.reduction,
			}}
			fundBuyback(g, tc.generic)
			if !engine.applyActionWithChoices(g, game.Player1, action.CastSpell(cardID,
				[]game.Target{game.StackObjectTarget(subject.ID)}, 0, nil), buybackAgents(t, true), &TurnLog{}) {
				t.Fatal("modified paid Reiterate cast failed")
			}
			obj, _ := g.Stack.Peek()
			if !obj.BuybackPaid || obj.ManaSpentToCast != tc.spent ||
				g.Players[game.Player1].ManaPool.Total() != 0 {
				t.Fatal("tax/reduction applied before or separately from buyback")
			}
			engine.resolveTopOfStack(g, &TurnLog{})
			if !g.Players[game.Player1].Hand.Contains(cardID) {
				t.Fatal("modified paid spell lost buyback")
			}
		})
	}
}

func TestBuybackCastCopyUsesSharedCastHooks(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := buybackEngine(t)
	castBuybackSubject(t, g, engine)
	def := reiterateFixture()
	def.StaticAbilities = append(def.StaticAbilities, game.StormStaticBody)
	def.TriggeredAbilities = []game.TriggeredAbility{{
		Trigger: game.TriggerCondition{Pattern: game.TriggerPattern{
			Event: game.EventSpellCast, Source: game.TriggerSourceSelf, SelfWasCast: true,
		}},
		Content: game.Mode{Sequence: []game.Instruction{{
			Primitive: game.GainLife{Amount: game.Fixed(1), Player: game.ControllerReference()},
		}}}.Ability(),
	}}
	cardID := addExiledInstant(g, game.Player4, def)
	castOnly := addTriggeredPermanent(g, game.Player1, instantSorceryCastOnlyTrigger(), []game.Instruction{{
		Primitive: game.GainLife{Amount: game.Fixed(2), Player: game.ControllerReference()},
	}}, nil)
	magecraft := addTriggeredPermanent(g, game.Player1, instantSorceryCopyTrigger(), []game.Instruction{{
		Primitive: game.GainLife{Amount: game.Fixed(4), Player: game.ControllerReference()},
	}}, nil)
	g.RuleEffects = append(g.RuleEffects, nextSpellUncounterableEffect(g))
	g.TriggerEventCursor = len(g.Events)
	g.Players[game.Player1].ManaPool.Add(mana.C, 3)
	if !engine.castFreeCopyOfCard(g, game.Player1, cardID, buybackAgents(t, true), &TurnLog{}) {
		t.Fatal("genuine card-copy cast failed")
	}
	objects := g.Stack.Objects()
	castCopy, stormCopy := objects[1], objects[2]
	if stackSpellCanBeCountered(g, castCopy) || !stackSpellCanBeCountered(g, stormCopy) ||
		len(g.RuleEffects) != 0 {
		t.Fatal("genuine cast did not consume the shared next-spell hook independently of its storm copy")
	}
	for _, branch := range []*game.Game{g, g.Clone()} {
		if !engine.putTriggeredAbilitiesOnStack(branch) {
			t.Fatal("shared cast-event hooks captured no triggers")
		}
		self, cast, copies := 0, 0, 0
		for _, obj := range branch.Stack.Objects() {
			if obj.Kind != game.StackTriggeredAbility {
				continue
			}
			switch obj.SourceID {
			case castOnly.ObjectID:
				cast++
			case magecraft.ObjectID:
				copies++
			case castCopy.ID:
				if obj.SourceCardID != 0 || obj.SourceTokenDef == nil {
					t.Fatal("cast-copy self trigger borrowed the imprinted physical source")
				}
				self++
			default:
				t.Fatal("non-cast storm copy produced a self-cast trigger")
			}
		}
		if self != 1 || cast != 1 || copies != 2 {
			t.Fatalf("shared boundary conflated cast/copy triggers: self=%d cast=%d magecraft=%d", self, cast, copies)
		}
		for range 4 {
			engine.resolveTopOfStack(branch, &TurnLog{})
		}
		if branch.Players[game.Player1].Life != 51 || branch.Stack.Size() != 3 ||
			!branch.Players[game.Player4].Exile.Contains(cardID) ||
			branch.CardInstances[cardID].ZoneVersion != 0 {
			t.Fatal("shared captured triggers lost their controller/source across cloning")
		}
	}
}

func TestBuybackLinkedCardCopyPaysAndPreservesExile(t *testing.T) {
	t.Parallel()
	for _, tc := range []struct {
		name      string
		mana      int
		paid      bool
		unpayable bool
		discard   bool
	}{
		{name: "paid", mana: 4, paid: true},
		{name: "declined", mana: 4},
		{name: "optional unavailable", mana: 1},
		{name: "tax unpayable", unpayable: true},
		{name: "mandatory discard unpayable", mana: 4, discard: true, unpayable: true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
			engine := buybackEngine(t)
			castBuybackSubject(t, g, engine)
			def := reiterateFixture()
			if tc.discard {
				def.AdditionalCosts = []cost.Additional{{Kind: cost.AdditionalDiscard, Amount: 1}}
			}
			cardID := addCardToHand(g, game.Player4, def)
			if !moveCardBetweenZones(g, game.Player4, cardID, zone.Hand, zone.Exile) {
				t.Fatal("could not exile the imprinted card")
			}
			version := g.CardInstances[cardID].ZoneVersion
			source := addCombatPermanent(g, game.Player1, isochronScepterDef())
			obj := activatedObjFor(source)
			imprintLink(g, obj, cardID)
			g.Players[game.Player1].ManaPool.Add(mana.C, tc.mana)
			g.CostModifiers = []game.CostModifier{{Kind: game.CostModifierSpell, GenericIncrease: 1}}
			before := len(g.Events)
			engine.resolveAbilityContentWithChoices(g, obj, imprintCopyCastAbility(),
				buybackAgents(t, tc.paid), &TurnLog{})
			if !g.Players[game.Player4].Exile.Contains(cardID) ||
				g.CardInstances[cardID].ZoneVersion != version ||
				countEventsOfKind(g.Events[before:], game.EventCardPlayedFromExile) != 0 {
				t.Fatal("casting the copy changed the imprinted physical card's incarnation or play history")
			}
			for _, event := range g.Events[before:] {
				if event.Kind == game.EventZoneChanged && event.CardID == cardID {
					t.Fatal("cast copy reported a physical card leaving exile")
				}
			}
			if tc.unpayable {
				if g.Stack.Size() != 1 || g.Players[game.Player1].ManaPool.Total() != tc.mana ||
					countEventsOfKind(g.Events[before:], game.EventSpellCast) != 0 {
					t.Fatal("unpayable card-copy cast changed resources, stack or cast history")
				}
				return
			}
			copied, _ := g.Stack.Peek()
			spent := 1
			if tc.paid {
				spent += 3
			}
			if g.Stack.Size() != 2 || !copied.Copy || copied.BuybackPaid != tc.paid ||
				copied.ManaSpentToCast != spent || !copied.CopyOwner.Exists ||
				copied.CopyOwner.Val != game.Player1 ||
				g.Players[game.Player1].ManaPool.Total() != tc.mana-spent ||
				countEventsOfKind(g.Events[before:], game.EventSpellCast) != 1 {
				t.Fatal("cast card copy bypassed the ordinary total-cost/payment or ownership rules")
			}
			engine.resolveTopOfStack(g, &TurnLog{})
			if !g.Players[game.Player4].Exile.Contains(cardID) ||
				g.CardInstances[cardID].ZoneVersion != version ||
				g.Players[game.Player1].Hand.Contains(cardID) ||
				g.Players[game.Player4].Hand.Contains(cardID) {
				t.Fatal("resolving a paid/unpaid copy moved the imprinted physical card")
			}
		})
	}
}
