package rules

import (
	"slices"
	"strings"
	"testing"

	"github.com/natefinch/council4/mtg/cards"
	"github.com/natefinch/council4/mtg/cards/authored"
	"github.com/natefinch/council4/mtg/cards/cardset"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/id"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// Pinned Oracle source: /Users/natefinch/.copilot/session-state/2d21805f-c9bb-4910-a872-eb7b8d5d9459/files/oracle-cards.json,
// SHA-256 6da9ced0e19bd384771ab8f14051cc45a8dc90302082b875fa6e039d2399a1b4.
// Reiterate: Oracle 38394ee3-6726-4f91-bc25-36bce0c6aab9, printing
// 99302c41-434f-41ff-a61a-2c3681a0c135, rulings
// https://api.scryfall.com/cards/99302c41-434f-41ff-a61a-2c3681a0c135/rulings,
// retrieved 2026-10-23; full seven-ruling response retained at
// /Users/natefinch/.copilot/session-state/25e78302-d27a-418d-9506-e2571ee328eb/files/sources/reiterate-rulings.json
// (SHA-256 74138c43c897cacd937f2b207741bad6c5be78d4831ed8b34e537e284e4f4b3a).
// WotC dated all seven 2021-03-19: copies aren't cast, preserve modes/X/cost
// decisions, and may keep or legally change targets.
// Allay: Oracle 664d982a-5cd8-4289-afb8-fa2e083042c9, printing
// f20a1c6d-ec6a-4bd6-b3b2-b997f71d41fc, rulings
// https://api.scryfall.com/cards/f20a1c6d-ec6a-4bd6-b3b2-b997f71d41fc/rulings,
// retrieved 2026-10-23; complete response is empty at
// /Users/natefinch/.copilot/session-state/25e78302-d27a-418d-9506-e2571ee328eb/files/sources/allay-rulings.json
// (SHA-256 6959514f6cea4012b6cceb5fe573de3433c3927da3183236a64ab122bc02ef10).
// Capsize: Oracle 77637eff-2963-4402-88f3-ca346f762fc8, printing
// f36f6545-e4e0-4dc8-acea-e527e43f3e14, rulings
// https://api.scryfall.com/cards/f36f6545-e4e0-4dc8-acea-e527e43f3e14/rulings,
// retrieved 2026-10-23; complete response is empty at
// /Users/natefinch/.copilot/session-state/25e78302-d27a-418d-9506-e2571ee328eb/files/sources/capsize-rulings.json
// (SHA-256 6959514f6cea4012b6cceb5fe573de3433c3927da3183236a64ab122bc02ef10).
// Official Comprehensive Rules:
// https://media.wizards.com/2026/downloads/MagicCompRules%2020260925.txt,
// verified local copy /Users/natefinch/.copilot/session-state/2d21805f-c9bb-4910-a872-eb7b8d5d9459/files/high-review-rules.txt,
// effective 2026-09-25, SHA-256
// 8d860e451f20f38865b725b42d82feb714c725373dd8f3b32b8652b3eeb070ca.
// Read 118.3/8 and 601.2a-i for complete optional additional-cost payment;
// 608.2b/c/f/n and 400.7 for target revalidation, resolution order, stack
// copies, zone changes and normal spell cleanup; 614.1/4/15, 615.1/4, and
// 616.1a-f for replacement timing and ordering; 701.8/19, 702.12, 702.16,
// 702.27a, 707.10/a-c,
// 122.1c, and 903.9b for destruction, protection, buyback, copy semantics,
// shield counters and Commander return choices.

type buybackInteractionCase struct {
	entry      cardset.Entry
	oracleID   string
	oracleText string
	manaCost   cost.Mana
	color      color.Color
	manaToPay  []mana.Color
	baseMana   int
}

func buybackInteractionCases() []buybackInteractionCase {
	return []buybackInteractionCase{
		{
			entry: authored.Reiterate, oracleID: "38394ee3-6726-4f91-bc25-36bce0c6aab9",
			oracleText: "Buyback {3} (You may pay an additional {3} as you cast this spell. If you do, put this card into your hand as it resolves.)\nCopy target instant or sorcery spell. You may choose new targets for the copy.",
			manaCost:   cost.Mana{cost.O(1), cost.R, cost.R}, color: color.Red,
			manaToPay: []mana.Color{mana.C, mana.R, mana.R}, baseMana: 3,
		},
		{
			entry: authored.Allay, oracleID: "664d982a-5cd8-4289-afb8-fa2e083042c9",
			oracleText: "Buyback {3} (You may pay an additional {3} as you cast this spell. If you do, put this card into your hand as it resolves.)\nDestroy target enchantment.",
			manaCost:   cost.Mana{cost.O(1), cost.W}, color: color.White,
			manaToPay: []mana.Color{mana.C, mana.W}, baseMana: 2,
		},
		{
			entry: authored.Capsize, oracleID: "77637eff-2963-4402-88f3-ca346f762fc8",
			oracleText: "Buyback {3} (You may pay an additional {3} as you cast this spell. If you do, put this card into your hand as it resolves.)\nReturn target permanent to its owner's hand.",
			manaCost:   cost.Mana{cost.O(1), cost.U, cost.U}, color: color.Blue,
			manaToPay: []mana.Color{mana.C, mana.U, mana.U}, baseMana: 3,
		},
	}
}

func buybackInteractionRuntime(t *testing.T) (*cards.Registry, *Engine) {
	t.Helper()
	registry := cards.NewDefaultRegistry()
	return registry, NewEngineWithPrograms(nil, registry.Programs())
}

func buybackInteractionDef(t *testing.T, registry *cards.Registry, tc buybackInteractionCase) *game.CardDef {
	t.Helper()
	def, err := registry.LookupOracleID(tc.oracleID)
	if err != nil || def == nil || def.Name != tc.entry.Name {
		t.Fatalf("exact production identity %s: %v", tc.oracleID, err)
	}
	return def
}

func buybackInteractionChoiceAgent(t *testing.T, payer game.PlayerID, paid bool, retarget ...game.Target) [game.NumPlayers]PlayerAgent {
	t.Helper()
	agent := &authoredCopyChoiceAgent{choose: func(request game.ChoiceRequest) []int {
		if request.Kind == game.ChoiceTarget && len(retarget) != 0 {
			return chooseAuthoredTargets(t, request, retarget...)
		}
		if request.Kind == game.ChoicePayment {
			for _, option := range request.Options {
				if strings.Contains(option.Label, "Buyback") {
					if paid {
						return []int{option.Index}
					}
					continue
				}
				if !paid {
					return []int{option.Index}
				}
			}
		}
		return request.DefaultSelection
	}}
	var agents [game.NumPlayers]PlayerAgent
	agents[payer] = agent
	return agents
}

func addBuybackInteractionMana(g *game.Game, tc buybackInteractionCase, buyback, includeGeneric bool) {
	for _, color := range tc.manaToPay {
		if color == mana.C && !includeGeneric {
			continue
		}
		g.Players[game.Player1].ManaPool.Add(color, 1)
	}
	if buyback {
		g.Players[game.Player1].ManaPool.Add(mana.C, 3)
	}
}

func addReiterateSubjectSpell(t *testing.T, g *game.Game, engine *Engine) *game.StackObject {
	t.Helper()
	def := &game.CardDef{CardFace: game.CardFace{
		Name:         "Real instant spell",
		Types:        []types.Card{types.Instant},
		ManaCost:     opt.Val(cost.Mana{cost.O(0)}),
		SpellAbility: opt.Val(game.AbilityContent{}),
	}}
	cardID := addCardToHand(g, game.Player2, def)
	setMainPhasePriority(g, game.Player2)
	if !engine.applyAction(g, game.Player2, action.CastSpell(cardID, nil, 0, nil)) {
		t.Fatal("real instant subject failed to cast")
	}
	subject, ok := g.Stack.Peek()
	if !ok || subject.Kind != game.StackSpell {
		t.Fatal("cast subject is not an instant spell on the stack")
	}
	g.Turn.Phase, g.Turn.Step, g.Turn.PriorityPlayer =
		game.PhaseCombat, game.StepDeclareBlockers, game.Player1
	return subject
}

func buybackInteractionTarget(t *testing.T, g *game.Game, engine *Engine, tc buybackInteractionCase) game.Target {
	t.Helper()
	if tc.entry.OracleID == authored.Reiterate.OracleID {
		return game.StackObjectTarget(addReiterateSubjectSpell(t, g, engine).ID)
	}
	target := addCombatPermanent(g, game.Player2, &game.CardDef{CardFace: game.CardFace{
		Name:  tc.entry.Name + " target",
		Types: []types.Card{types.Enchantment},
	}})
	return game.PermanentTarget(target.ObjectID)
}

func castProductionBuyback(t *testing.T, g *game.Game, engine *Engine, registry *cards.Registry, tc buybackInteractionCase, payer game.PlayerID, target game.Target, paid bool, manaColors []mana.Color) (id.ID, *game.StackObject) {
	t.Helper()
	def := buybackInteractionDef(t, registry, tc)
	cardID := addCardToHand(g, payer, def)
	for _, manaColor := range manaColors {
		g.Players[payer].ManaPool.Add(manaColor, 1)
	}
	setMainPhasePriority(g, g.Turn.ActivePlayer)
	g.Turn.Phase, g.Turn.Step, g.Turn.PriorityPlayer =
		game.PhaseCombat, game.StepDeclareBlockers, payer
	agents := buybackInteractionChoiceAgent(t, payer, paid)
	if !engine.applyActionWithChoices(g, payer,
		action.CastSpell(cardID, []game.Target{target}, 0, nil), agents, &TurnLog{}) {
		t.Fatalf("%s real cast failed", tc.entry.Name)
	}
	spell, ok := g.Stack.Peek()
	if !ok || spell.SourceID != cardID || spell.Kind != game.StackSpell {
		t.Fatalf("%s did not produce its physical spell object", tc.entry.Name)
	}
	return cardID, spell
}

func TestBuybackInteractionRegistryIdentityAndCosts(t *testing.T) {
	t.Parallel()
	registry, _ := buybackInteractionRuntime(t)
	for _, tc := range buybackInteractionCases() {
		t.Run(tc.entry.Name, func(t *testing.T) {
			t.Parallel()
			def := buybackInteractionDef(t, registry, tc)
			if !def.ManaCost.Exists || !slices.Equal(def.ManaCost.Val, tc.manaCost) ||
				!slices.Equal(def.Types, []types.Card{types.Instant}) ||
				!slices.Equal(def.Colors, []color.Color{tc.color}) ||
				def.OracleText != tc.oracleText {
				t.Fatalf("printed definition differs from exact Oracle text: %+v", def.CardFace)
			}
			keyword, ok := def.BuybackKeyword()
			if !ok || !slices.Equal(keyword.Cost, cost.Mana{cost.O(3)}) ||
				!def.HasKeyword(game.Buyback) {
				t.Fatal("buyback's printed additional cost is not inspectable")
			}
			if len(def.StaticAbilities) != 1 ||
				def.StaticAbilities[0].ZoneOfFunction != zone.Stack ||
				def.StaticAbilities[0].Text != "Buyback {3}" {
				t.Fatal("buyback is not declared as a stack static ability")
			}
			if err := registry.Programs().ValidateCard(def); err != nil {
				t.Fatalf("registered program validation: %v", err)
			}
		})
	}
}

func TestBuybackInteractionCastsPaymentsAndInstantTiming(t *testing.T) {
	t.Parallel()
	for _, tc := range buybackInteractionCases() {
		for _, payment := range []struct {
			name              string
			provideBuyback    bool
			paid              bool
			cast              bool
			leftAfterBaseCost int
		}{
			{name: "paid", provideBuyback: true, paid: true, cast: true},
			{name: "declined", provideBuyback: true, cast: true, leftAfterBaseCost: 3},
			{name: "unaffordable optional cost", cast: true},
			{name: "insufficient normal cost", cast: false},
		} {
			t.Run(tc.entry.Name+"/"+payment.name, func(t *testing.T) {
				registry, engine := buybackInteractionRuntime(t)
				g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
				target := buybackInteractionTarget(t, g, engine, tc)
				if payment.provideBuyback {
					addBuybackInteractionMana(g, tc, true, true)
				} else if payment.cast {
					addBuybackInteractionMana(g, tc, false, true)
				} else {
					for _, color := range tc.manaToPay {
						if color != mana.C {
							g.Players[game.Player1].ManaPool.Add(color, 1)
						}
					}
				}
				if tc.entry.OracleID != authored.Reiterate.OracleID {
					setMainPhasePriority(g, game.Player2)
					g.Turn.Phase, g.Turn.Step, g.Turn.PriorityPlayer =
						game.PhaseCombat, game.StepDeclareBlockers, game.Player1
				}
				def := buybackInteractionDef(t, registry, tc)
				cardID := addCardToHand(g, game.Player1, def)
				beforeMana := g.Players[game.Player1].ManaPool.Total()
				agents := buybackInteractionChoiceAgent(t, game.Player1, payment.paid)
				act := action.CastSpell(cardID, []game.Target{target}, 0, nil)
				before := g.Clone()
				cast := engine.applyActionWithChoices(g, game.Player1, act, agents, &TurnLog{})
				if cast != payment.cast {
					t.Fatalf("cast = %v, want %v", cast, payment.cast)
				}
				if !cast {
					if !g.Players[game.Player1].Hand.Contains(cardID) ||
						g.Players[game.Player1].ManaPool.Total() != beforeMana ||
						g.Stack.Size() != before.Stack.Size() {
						t.Fatal("failed payment was not atomic")
					}
					if branch, accepted := engine.Simulator().Apply(before, game.Player1, act, agents); accepted || branch != nil {
						t.Fatal("Simulator accepted the same underfunded cast")
					}
					return
				}
				spell, ok := g.Stack.Peek()
				wantSpent := tc.baseMana
				if payment.paid {
					wantSpent += 3
				}
				if !ok || spell.SourceID != cardID ||
					spell.BuybackPaid != payment.paid ||
					spell.ManaSpentToCast != wantSpent ||
					g.Players[game.Player1].ManaPool.Total() != beforeMana-wantSpent {
					t.Fatalf("wrong paid cast state: %+v pool=%d", spell, g.Players[game.Player1].ManaPool.Total())
				}
				branch, accepted := engine.Simulator().Apply(before, game.Player1, act, agents)
				if !accepted || branch == nil {
					t.Fatal("Simulator rejected the same payable production cast")
				}
				simulated, ok := branch.Stack.Peek()
				if !ok || simulated.SourceID != cardID ||
					simulated.BuybackPaid != spell.BuybackPaid ||
					simulated.ManaSpentToCast != spell.ManaSpentToCast ||
					branch.Players[game.Player1].ManaPool.Total() != g.Players[game.Player1].ManaPool.Total() {
					t.Fatal("live and Simulator cast/payment results diverged")
				}
				if payment.leftAfterBaseCost != 0 &&
					g.Players[game.Player1].ManaPool.Total() != payment.leftAfterBaseCost {
					t.Fatalf("declined buyback left %d mana, want %d",
						g.Players[game.Player1].ManaPool.Total(), payment.leftAfterBaseCost)
				}
				if !payment.paid {
					engine.resolveTopOfStack(g, &TurnLog{})
					if !g.Players[game.Player1].Graveyard.Contains(cardID) ||
						g.Players[game.Player1].Hand.Contains(cardID) {
						t.Fatal("declined or unaffordable buyback did not use the normal graveyard destination")
					}
				}
			})
		}
	}
}

func TestBuybackInteractionAllayProtectionAndDestructionReplacement(t *testing.T) {
	t.Parallel()
	for _, scenario := range []string{"indestructible", "shield", "regeneration", "graveyard redirects to exile"} {
		t.Run(scenario, func(t *testing.T) {
			registry, engine := buybackInteractionRuntime(t)
			g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
			def := &game.CardDef{CardFace: game.CardFace{
				Name: "Enchanting creature", Types: []types.Card{types.Enchantment, types.Creature},
			}}
			target := addCombatPermanent(g, game.Player2, def)
			switch scenario {
			case "indestructible":
				def.StaticAbilities = []game.StaticAbility{game.IndestructibleStaticBody}
			case "shield":
				target.Counters.Add(counter.Shield, 1)
			case "regeneration":
				target.RegenerationShields = 1
			case "graveyard redirects to exile":
				g.ReplacementEffects = append(g.ReplacementEffects, game.ReplacementEffect{
					ID: g.IDGen.Next(), MatchEvent: game.EventZoneChanged,
					MatchFromZone: true, FromZone: zone.Battlefield,
					MatchToZone: true, ToZone: zone.Graveyard,
					ReplaceToZone: zone.Exile, AffectedObjectID: target.ObjectID,
					Duration: game.DurationPermanent,
				})
			}
			cardID, spell := castProductionBuyback(t, g, engine, registry, buybackInteractionCases()[1],
				game.Player1, game.PermanentTarget(target.ObjectID), true,
				[]mana.Color{mana.C, mana.W, mana.C, mana.C, mana.C})
			engine.resolveTopOfStack(g, &TurnLog{})
			if !g.Players[game.Player1].Hand.Contains(cardID) ||
				g.Players[game.Player1].Graveyard.Contains(cardID) ||
				!spell.BuybackPaid {
				t.Fatal("successful Allay resolution did not apply its paid buyback replacement")
			}
			switch scenario {
			case "indestructible":
				if _, live := permanentByObjectID(g, target.ObjectID); !live {
					t.Fatal("destroy failed to respect indestructible")
				}
			case "shield":
				if _, live := permanentByObjectID(g, target.ObjectID); !live ||
					target.Counters.Get(counter.Shield) != 0 {
					t.Fatal("shield counter did not replace the destruction")
				}
			case "regeneration":
				if _, live := permanentByObjectID(g, target.ObjectID); !live ||
					target.RegenerationShields != 0 || !target.Tapped {
					t.Fatal("regeneration did not replace the destruction")
				}
			case "graveyard redirects to exile":
				if _, live := permanentByObjectID(g, target.ObjectID); live ||
					!g.Players[game.Player2].Exile.Contains(target.CardInstanceID) {
					t.Fatal("destruction bypassed the battlefield-to-graveyard replacement")
				}
			}
		})
	}

	t.Run("protection makes the target illegal before payment", func(t *testing.T) {
		registry, engine := buybackInteractionRuntime(t)
		g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
		targetDef := &game.CardDef{CardFace: game.CardFace{
			Name:  "Protected enchantment",
			Types: []types.Card{types.Enchantment},
			StaticAbilities: []game.StaticAbility{{
				KeywordAbilities: []game.KeywordAbility{
					game.ProtectionKeyword{FromColors: []color.Color{color.White}},
				},
			}},
		}}
		target := addCombatPermanent(g, game.Player2, targetDef)
		setMainPhasePriority(g, game.Player2)
		g.Turn.Phase, g.Turn.Step, g.Turn.PriorityPlayer =
			game.PhaseCombat, game.StepDeclareBlockers, game.Player1
		cardID := addCardToHand(g, game.Player1, buybackInteractionDef(t, registry, buybackInteractionCases()[1]))
		for _, color := range []mana.Color{mana.C, mana.W, mana.C, mana.C, mana.C} {
			g.Players[game.Player1].ManaPool.Add(color, 1)
		}
		manaBefore := g.Players[game.Player1].ManaPool.Total()
		if engine.applyAction(g, game.Player1,
			action.CastSpell(cardID, []game.Target{game.PermanentTarget(target.ObjectID)}, 0, nil)) {
			t.Fatal("Allay targeted an enchantment with protection from white")
		}
		if !g.Players[game.Player1].Hand.Contains(cardID) ||
			g.Players[game.Player1].ManaPool.Total() != manaBefore ||
			g.Stack.Size() != 0 {
			t.Fatal("illegal protected target consumed the spell or its payment")
		}
	})
}

func TestBuybackInteractionAllayCountered(t *testing.T) {
	t.Parallel()
	registry, engine := buybackInteractionRuntime(t)
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	target := addCombatPermanent(g, game.Player3, &game.CardDef{CardFace: game.CardFace{
		Name: "Counter target enchantment", Types: []types.Card{types.Enchantment},
	}})
	allayID, allay := castProductionBuyback(t, g, engine, registry, buybackInteractionCases()[1],
		game.Player1, game.PermanentTarget(target.ObjectID), true,
		[]mana.Color{mana.C, mana.W, mana.C, mana.C, mana.C})
	counterspell, err := registry.LookupChecked("Counterspell")
	if err != nil || counterspell == nil {
		t.Fatalf("production Counterspell: %v", err)
	}
	counterspellID := addCardToHand(g, game.Player2, counterspell)
	g.Players[game.Player2].ManaPool.Add(mana.U, 2)
	g.Turn.PriorityPlayer = game.Player2
	if !engine.applyAction(g, game.Player2,
		action.CastSpell(counterspellID, []game.Target{game.StackObjectTarget(allay.ID)}, 0, nil)) {
		t.Fatal("real Counterspell could not counter the paid Allay")
	}
	engine.resolveTopOfStack(g, &TurnLog{})
	_, targetLive := permanentByObjectID(g, target.ObjectID)
	if !g.Players[game.Player1].Graveyard.Contains(allayID) ||
		g.Players[game.Player1].Hand.Contains(allayID) ||
		!targetLive {
		t.Fatal("countered paid Allay returned to hand or destroyed its target")
	}
}

func TestBuybackInteractionReiterateCopiesPaidSpellChoicesAndClone(t *testing.T) {
	t.Parallel()
	registry, engine := buybackInteractionRuntime(t)
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	first := addCombatPermanent(g, game.Player3, vanillaCreature("Original target A", 6, 6))
	second := addCombatPermanent(g, game.Player3, vanillaCreature("Original target B", 6, 6))
	retarget := addCombatPermanent(g, game.Player3, vanillaCreature("Copy target C", 6, 6))
	x := game.Dynamic(game.DynamicAmount{Kind: game.DynamicAmountX})
	targetSpec := game.TargetSpec{
		MinTargets: 1, MaxTargets: 1, Allow: game.TargetAllowPermanent,
		Selection: opt.Val(game.Selection{RequiredTypes: []types.Card{types.Creature}}),
	}
	subjectDef := &game.CardDef{CardFace: game.CardFace{
		Name: "Modal X subject", Types: []types.Card{types.Instant},
		ManaCost: opt.Val(cost.Mana{cost.X}),
		SpellAbility: opt.Val(game.AbilityContent{
			MinModes: 1, MaxModes: 1,
			Modes: []game.Mode{
				{Sequence: []game.Instruction{{Primitive: game.GainLife{
					Player: game.ControllerReference(), Amount: game.Fixed(99),
				}}}},
				{Targets: []game.TargetSpec{targetSpec, targetSpec}, Sequence: []game.Instruction{
					{Primitive: game.Damage{Amount: x, Recipient: game.AnyTargetDamageRecipient(0)}},
					{Primitive: game.Damage{Amount: x, Recipient: game.AnyTargetDamageRecipient(1)}},
				}},
			},
		}),
	}}
	subjectDef.AdditionalCosts = []cost.Additional{{
		Kind: cost.AdditionalDiscard, Amount: 1, SubjectKey: "reiterate-copy-discard",
	}}
	discarded := addCardToHand(g, game.Player2, vanillaCreature("Copied cost subject", 1, 1))
	subjectCardID := addCardToHand(g, game.Player2, subjectDef)
	g.Players[game.Player2].ManaPool.Add(mana.C, 3)
	setMainPhasePriority(g, game.Player2)
	subjectAgents := [game.NumPlayers]PlayerAgent{
		game.Player2: &authoredCopyChoiceAgent{choose: func(request game.ChoiceRequest) []int {
			return request.DefaultSelection
		}},
	}
	if !engine.applyActionWithChoices(g, game.Player2,
		action.CastSpell(subjectCardID, []game.Target{
			game.PermanentTarget(first.ObjectID), game.PermanentTarget(second.ObjectID),
		}, 3, []int{1}), subjectAgents, &TurnLog{}) {
		t.Fatal("real modal X spell with a discard additional cost did not cast")
	}
	subject, _ := g.Stack.Peek()
	if subject.XValue != 3 || !slices.Equal(subject.ChosenModes, []int{1}) ||
		len(subject.PaidCostSubjects) != 1 ||
		subject.PaidCostSubjects[0].Snapshot.CardID != discarded ||
		!g.Players[game.Player2].Graveyard.Contains(discarded) {
		t.Fatal("subject spell did not preserve its actual X, mode, and paid discard")
	}
	g.Turn.Phase, g.Turn.Step, g.Turn.PriorityPlayer =
		game.PhaseCombat, game.StepDeclareBlockers, game.Player1
	reiterateID, reiterate := castProductionBuyback(t, g, engine, registry, buybackInteractionCases()[0],
		game.Player1, game.StackObjectTarget(subject.ID), true,
		[]mana.Color{mana.C, mana.R, mana.R, mana.C, mana.C, mana.C})
	if !reiterate.BuybackPaid || reiterate.ManaSpentToCast != 6 {
		t.Fatal("real Reiterate did not pay its printed mana and buyback costs")
	}
	agents := buybackInteractionChoiceAgent(t, game.Player1, true,
		game.PermanentTarget(first.ObjectID), game.PermanentTarget(retarget.ObjectID))
	engine.resolveTopOfStackWithChoices(g, agents, &TurnLog{})
	copied, ok := g.Stack.Peek()
	if !ok || !copied.Copy || copied.XValue != 3 ||
		!slices.Equal(copied.ChosenModes, []int{1}) ||
		copied.Targets[0].PermanentID != first.ObjectID ||
		copied.Targets[1].PermanentID != retarget.ObjectID ||
		len(copied.PaidCostSubjects) != 1 ||
		copied.PaidCostSubjects[0].Snapshot.CardID != discarded {
		t.Fatal("Reiterate's copy lost spell identity, X, mode, legal new targets, or paid-cost subject")
	}
	if !g.Players[game.Player1].Hand.Contains(reiterateID) ||
		g.Players[game.Player1].Graveyard.Contains(reiterateID) ||
		countEventsOfKind(g.Events, game.EventSpellCast) != 2 ||
		countEventsOfKind(g.Events, game.EventSpellCopied) != 1 ||
		first.MarkedDamage != 0 || second.MarkedDamage != 0 || retarget.MarkedDamage != 0 {
		t.Fatal("Reiterate was recast, failed buyback, or resolved the copy before its target")
	}
	clone := g.Clone()
	engine.resolveTopOfStack(clone, &TurnLog{})
	if first.MarkedDamage != 0 || retarget.MarkedDamage != 0 ||
		clone.Stack.Size() != 1 || !clone.Players[game.Player1].Hand.Contains(reiterateID) ||
		clone.CardInstances[reiterateID].ZoneVersion != g.CardInstances[reiterateID].ZoneVersion {
		t.Fatal("resolving the cloned state mutated the live state or lost Reiterate's new incarnation")
	}
	engine.resolveTopOfStack(g, &TurnLog{})
	if first.MarkedDamage != 3 || second.MarkedDamage != 0 || retarget.MarkedDamage != 3 ||
		countEventsOfKind(g.Events, game.EventSpellCast) != 2 ||
		countEventsOfKind(g.Events, game.EventSpellCopied) != 1 {
		t.Fatal("copy did not resolve with the selected targets without recasting")
	}
	engine.resolveTopOfStack(g, &TurnLog{})
	if first.MarkedDamage != 6 || second.MarkedDamage != 3 ||
		!g.Players[game.Player2].Graveyard.Contains(subjectCardID) {
		t.Fatal("copy changed the original spell's targets, mode, or X")
	}
}

func TestBuybackInteractionReiteratePaidBuybackCopyHasNoPhysicalCard(t *testing.T) {
	t.Parallel()
	registry, engine := buybackInteractionRuntime(t)
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	first := addCombatPermanent(g, game.Player3, &game.CardDef{CardFace: game.CardFace{
		Name: "First enchantment", Types: []types.Card{types.Enchantment},
	}})
	second := addCombatPermanent(g, game.Player3, &game.CardDef{CardFace: game.CardFace{
		Name: "Second enchantment", Types: []types.Card{types.Enchantment},
	}})
	allayDef := buybackInteractionDef(t, registry, buybackInteractionCases()[1])
	allayID := addCardToHand(g, game.Player2, allayDef)
	for _, color := range []mana.Color{mana.C, mana.W, mana.C, mana.C, mana.C} {
		g.Players[game.Player2].ManaPool.Add(color, 1)
	}
	setMainPhasePriority(g, game.Player2)
	if !engine.applyActionWithChoices(g, game.Player2,
		action.CastSpell(allayID, []game.Target{game.PermanentTarget(first.ObjectID)}, 0, nil),
		buybackInteractionChoiceAgent(t, game.Player2, true), &TurnLog{}) {
		t.Fatal("paid Allay subject failed to cast")
	}
	allay, _ := g.Stack.Peek()
	g.Turn.Phase, g.Turn.Step, g.Turn.PriorityPlayer =
		game.PhaseCombat, game.StepDeclareBlockers, game.Player1
	reiterateID, _ := castProductionBuyback(t, g, engine, registry, buybackInteractionCases()[0],
		game.Player1, game.StackObjectTarget(allay.ID), true,
		[]mana.Color{mana.C, mana.R, mana.R, mana.C, mana.C, mana.C})
	engine.resolveTopOfStackWithChoices(g,
		buybackInteractionChoiceAgent(t, game.Player1, true, game.PermanentTarget(second.ObjectID)),
		&TurnLog{})
	copied, ok := g.Stack.Peek()
	if !ok || !copied.Copy || !copied.BuybackPaid ||
		!g.Players[game.Player1].Hand.Contains(reiterateID) ||
		g.Players[game.Player2].Hand.Contains(allayID) {
		t.Fatal("spell copy lost buyback state, or returned a physical card before the original resolved")
	}
	engine.resolveTopOfStack(g, &TurnLog{})
	_, firstLive := permanentByObjectID(g, first.ObjectID)
	if _, live := permanentByObjectID(g, second.ObjectID); live ||
		!firstLive ||
		g.Players[game.Player2].Hand.Contains(allayID) {
		t.Fatal("copy resolution moved the original Allay card or applied the original target")
	}
	engine.resolveTopOfStack(g, &TurnLog{})
	if _, live := permanentByObjectID(g, first.ObjectID); live ||
		!g.Players[game.Player2].Hand.Contains(allayID) ||
		g.CardInstances[allayID].ZoneVersion != 2 ||
		countEventsOfKind(g.Events, game.EventSpellCast) != 2 ||
		countEventsOfKind(g.Events, game.EventSpellCopied) != 1 {
		t.Fatal("only the original resolving Allay should return its physical card")
	}
}

func TestBuybackInteractionReiterateRejectsAbilityTargets(t *testing.T) {
	t.Parallel()
	registry, engine := buybackInteractionRuntime(t)
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	source := addCombatPermanent(g, game.Player2, &game.CardDef{CardFace: game.CardFace{
		Name:  "Activated ability source",
		Types: []types.Card{types.Creature},
		ActivatedAbilities: []game.ActivatedAbility{{
			Content: game.AbilityContent{},
		}},
	}})
	setMainPhasePriority(g, game.Player2)
	if !engine.applyAction(g, game.Player2, action.ActivateAbility(source.ObjectID, 0, nil, 0)) {
		t.Fatal("real activated ability failed to reach the stack")
	}
	ability, _ := g.Stack.Peek()
	if ability.Kind != game.StackActivatedAbility {
		t.Fatal("fixture did not create an activated ability")
	}
	g.Turn.Phase, g.Turn.Step, g.Turn.PriorityPlayer =
		game.PhaseCombat, game.StepDeclareBlockers, game.Player1
	cardID := addCardToHand(g, game.Player1, buybackInteractionDef(t, registry, buybackInteractionCases()[0]))
	for _, color := range []mana.Color{mana.C, mana.R, mana.R} {
		g.Players[game.Player1].ManaPool.Add(color, 1)
	}
	manaBefore := g.Players[game.Player1].ManaPool.Total()
	if engine.applyAction(g, game.Player1,
		action.CastSpell(cardID, []game.Target{game.StackObjectTarget(ability.ID)}, 0, nil)) {
		t.Fatal("Reiterate targeted an activated ability instead of an instant or sorcery spell")
	}
	if !g.Players[game.Player1].Hand.Contains(cardID) ||
		g.Players[game.Player1].ManaPool.Total() != manaBefore ||
		g.Stack.Size() != 1 {
		t.Fatal("illegal ability target consumed Reiterate or its normal payment")
	}
}

func TestBuybackInteractionReiterateAllIllegalTargetDoesNotBuyBack(t *testing.T) {
	t.Parallel()
	registry, engine := buybackInteractionRuntime(t)
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	subject := addReiterateSubjectSpell(t, g, engine)
	reiterateID, _ := castProductionBuyback(t, g, engine, registry, buybackInteractionCases()[0],
		game.Player1, game.StackObjectTarget(subject.ID), true,
		[]mana.Color{mana.C, mana.R, mana.R, mana.C, mana.C, mana.C})
	counterspell, err := registry.LookupChecked("Counterspell")
	if err != nil || counterspell == nil {
		t.Fatalf("production Counterspell: %v", err)
	}
	counterspellID := addCardToHand(g, game.Player3, counterspell)
	g.Players[game.Player3].ManaPool.Add(mana.U, 2)
	g.Turn.PriorityPlayer = game.Player3
	if !engine.applyAction(g, game.Player3,
		action.CastSpell(counterspellID, []game.Target{game.StackObjectTarget(subject.ID)}, 0, nil)) {
		t.Fatal("real Counterspell could not make Reiterate's target illegal")
	}
	engine.resolveTopOfStack(g, &TurnLog{})
	engine.resolveTopOfStack(g, &TurnLog{})
	if !g.Players[game.Player2].Graveyard.Contains(subject.SourceID) ||
		!g.Players[game.Player1].Graveyard.Contains(reiterateID) ||
		g.Players[game.Player1].Hand.Contains(reiterateID) ||
		countEventsOfKind(g.Events, game.EventSpellCopied) != 0 {
		t.Fatal("Reiterate copied a departed target or used buyback without resolving")
	}
}

func TestBuybackInteractionTargetIncarnations(t *testing.T) {
	t.Parallel()
	for _, tc := range []buybackInteractionCase{buybackInteractionCases()[1], buybackInteractionCases()[2]} {
		t.Run(tc.entry.Name, func(t *testing.T) {
			registry, engine := buybackInteractionRuntime(t)
			g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
			kind := types.Enchantment
			if tc.entry.OracleID == authored.Capsize.OracleID {
				kind = types.Artifact
			}
			target := addCombatPermanent(g, game.Player2, &game.CardDef{CardFace: game.CardFace{
				Name: "Exact incarnation target", Types: []types.Card{kind},
			}})
			sourceID, spell := castProductionBuyback(t, g, engine, registry, tc,
				game.Player1, game.PermanentTarget(target.ObjectID), true, tc.manaToPayWithBuyback())
			movePermanentToZone(g, target, zone.Exile)
			card, ok := g.GetCardInstance(target.CardInstanceID)
			if !ok {
				t.Fatal("target card instance disappeared during exile")
			}
			reentered, ok := createCardPermanent(g, card, target.Controller, zone.Exile)
			if !ok || reentered.ObjectID == target.ObjectID ||
				reentered.CardInstanceID != target.CardInstanceID {
				t.Fatal("target did not return as a new permanent incarnation")
			}
			engine.resolveTopOfStack(g, &TurnLog{})
			_, newIncarnationLive := permanentByObjectID(g, reentered.ObjectID)
			if !spell.BuybackPaid ||
				!g.Players[game.Player1].Graveyard.Contains(sourceID) ||
				g.Players[game.Player1].Hand.Contains(sourceID) ||
				!newIncarnationLive {
				t.Fatal("old target identity resolved or an all-illegal paid spell used buyback")
			}
		})
	}
}

func (tc buybackInteractionCase) manaToPayWithBuyback() []mana.Color {
	colors := slices.Clone(tc.manaToPay)
	colors = append(colors, mana.C, mana.C, mana.C)
	return colors
}

func TestBuybackInteractionCapsizePermanentTypesTokensAndOwner(t *testing.T) {
	t.Parallel()
	for _, kind := range []types.Card{
		types.Artifact, types.Creature, types.Enchantment, types.Land,
		types.Planeswalker, types.Battle,
	} {
		t.Run(string(kind), func(t *testing.T) {
			registry, engine := buybackInteractionRuntime(t)
			g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
			target := addCombatPermanent(g, game.Player2, &game.CardDef{CardFace: game.CardFace{
				Name: "Capsize permanent", Types: []types.Card{kind},
			}})
			cardID, _ := castProductionBuyback(t, g, engine, registry, buybackInteractionCases()[2],
				game.Player1, game.PermanentTarget(target.ObjectID), true,
				buybackInteractionCases()[2].manaToPayWithBuyback())
			engine.resolveTopOfStack(g, &TurnLog{})
			if !g.Players[game.Player2].Hand.Contains(target.CardInstanceID) ||
				g.Players[game.Player1].Hand.Contains(target.CardInstanceID) ||
				!g.Players[game.Player1].Hand.Contains(cardID) {
				t.Fatalf("Capsize did not return the %s to its owner while buying back", kind)
			}
		})
	}

	t.Run("owner differs from controller", func(t *testing.T) {
		registry, engine := buybackInteractionRuntime(t)
		g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
		target := addOwnedControlledPermanent(g, game.Player2, game.Player3, &game.CardDef{CardFace: game.CardFace{
			Name: "Opponent-owned permanent", Types: []types.Card{types.Artifact},
		}})
		cardID, _ := castProductionBuyback(t, g, engine, registry, buybackInteractionCases()[2],
			game.Player1, game.PermanentTarget(target.ObjectID), true,
			buybackInteractionCases()[2].manaToPayWithBuyback())
		engine.resolveTopOfStack(g, &TurnLog{})
		if !g.Players[game.Player2].Hand.Contains(target.CardInstanceID) ||
			g.Players[game.Player1].Hand.Contains(target.CardInstanceID) ||
			g.Players[game.Player3].Hand.Contains(target.CardInstanceID) ||
			!g.Players[game.Player1].Hand.Contains(cardID) {
			t.Fatal("Capsize returned a controlled permanent to anyone other than its owner")
		}
	})

	t.Run("token ceases to exist after return", func(t *testing.T) {
		registry, engine := buybackInteractionRuntime(t)
		g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
		tokenDef := &game.CardDef{CardFace: game.CardFace{
			Name: "Created artifact token", Types: []types.Card{types.Artifact},
		}}
		addEffectSpellToStack(g, game.Player2, game.CreateToken{
			Amount: game.Fixed(1), Source: game.TokenDef(tokenDef),
		}, nil)
		engine.resolveTopOfStack(g, &TurnLog{})
		var token *game.Permanent
		for _, permanent := range g.Battlefield {
			if permanent.Token && permanent.TokenDef == tokenDef {
				token = permanent
				break
			}
		}
		if token == nil {
			t.Fatal("fixture failed to create the token through the Engine")
		}
		cardID, _ := castProductionBuyback(t, g, engine, registry, buybackInteractionCases()[2],
			game.Player1, game.PermanentTarget(token.ObjectID), true,
			buybackInteractionCases()[2].manaToPayWithBuyback())
		engine.resolveTopOfStack(g, &TurnLog{})
		if _, live := permanentByObjectID(g, token.ObjectID); live ||
			!g.Players[game.Player1].Hand.Contains(cardID) ||
			g.Players[game.Player2].Hand.Contains(token.CardInstanceID) {
			t.Fatal("returned token did not cease to exist or paid Capsize failed buyback")
		}
	})
}

func addOwnedControlledPermanent(g *game.Game, owner, controller game.PlayerID, def *game.CardDef) *game.Permanent {
	cardID := g.IDGen.Next()
	g.CardInstances[cardID] = &game.CardInstance{ID: cardID, Def: def, Owner: owner}
	permanent := &game.Permanent{
		ObjectID: g.IDGen.Next(), CardInstanceID: cardID, Owner: owner, Controller: controller,
	}
	g.Battlefield = append(g.Battlefield, permanent)
	return permanent
}

func TestBuybackInteractionCapsizeCommanderAndReplacement(t *testing.T) {
	t.Parallel()
	t.Run("commander owner chooses command zone instead", func(t *testing.T) {
		registry, engine := buybackInteractionRuntime(t)
		g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
		target := addCommanderPermanent(g, game.Player2)
		target.Controller = game.Player3
		cardID, _ := castProductionBuyback(t, g, engine, registry, buybackInteractionCases()[2],
			game.Player1, game.PermanentTarget(target.ObjectID), true,
			buybackInteractionCases()[2].manaToPayWithBuyback())
		choice := &authoredCopyChoiceAgent{choose: func(request game.ChoiceRequest) []int {
			if request.Kind == game.ChoiceMay {
				if request.Player != game.Player2 {
					t.Fatalf("commander replacement choice reached %v instead of owner", request.Player)
				}
				return []int{1}
			}
			return request.DefaultSelection
		}}
		engine.resolveTopOfStackWithChoices(g, [game.NumPlayers]PlayerAgent{game.Player2: choice}, &TurnLog{})
		if !g.Players[game.Player2].CommandZone.Contains(target.CardInstanceID) ||
			g.Players[game.Player2].Hand.Contains(target.CardInstanceID) ||
			!g.Players[game.Player1].Hand.Contains(cardID) {
			t.Fatal("Capsize bypassed the commander's owner's replacement choice or lost buyback")
		}
	})

	t.Run("hand destination is redirected to exile", func(t *testing.T) {
		registry, engine := buybackInteractionRuntime(t)
		g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
		target := addOwnedControlledPermanent(g, game.Player2, game.Player3, &game.CardDef{CardFace: game.CardFace{
			Name: "Redirected permanent", Types: []types.Card{types.Land},
		}})
		g.ReplacementEffects = append(g.ReplacementEffects, game.ReplacementEffect{
			ID: g.IDGen.Next(), MatchEvent: game.EventZoneChanged,
			MatchFromZone: true, FromZone: zone.Battlefield,
			MatchToZone: true, ToZone: zone.Hand,
			ReplaceToZone: zone.Exile, AffectedObjectID: target.ObjectID,
			Duration: game.DurationPermanent,
		})
		cardID, _ := castProductionBuyback(t, g, engine, registry, buybackInteractionCases()[2],
			game.Player1, game.PermanentTarget(target.ObjectID), true,
			buybackInteractionCases()[2].manaToPayWithBuyback())
		engine.resolveTopOfStack(g, &TurnLog{})
		if g.Players[game.Player2].Hand.Contains(target.CardInstanceID) ||
			!g.Players[game.Player2].Exile.Contains(target.CardInstanceID) ||
			!g.Players[game.Player1].Hand.Contains(cardID) {
			t.Fatal("Capsize bypassed the permanent's hand-destination replacement or failed buyback")
		}
	})
}

func TestBuybackInteractionCapsizeSimulatorAndClone(t *testing.T) {
	t.Parallel()
	registry, engine := buybackInteractionRuntime(t)
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	target := addCombatPermanent(g, game.Player2, &game.CardDef{CardFace: game.CardFace{
		Name: "Simulator target", Types: []types.Card{types.Artifact},
	}})
	def := buybackInteractionDef(t, registry, buybackInteractionCases()[2])
	cardID := addCardToHand(g, game.Player1, def)
	for _, color := range buybackInteractionCases()[2].manaToPayWithBuyback() {
		g.Players[game.Player1].ManaPool.Add(color, 1)
	}
	setMainPhasePriority(g, game.Player2)
	g.Turn.Phase, g.Turn.Step, g.Turn.PriorityPlayer =
		game.PhaseCombat, game.StepDeclareBlockers, game.Player1
	act := action.CastSpell(cardID, []game.Target{game.PermanentTarget(target.ObjectID)}, 0, nil)
	agents := buybackInteractionChoiceAgent(t, game.Player1, true)
	before := g.Clone()
	if !engine.applyActionWithChoices(g, game.Player1, act, agents, &TurnLog{}) {
		t.Fatal("live paid Capsize failed")
	}
	simulated, accepted := engine.Simulator().Apply(before, game.Player1, act, agents)
	if !accepted || simulated == nil {
		t.Fatal("Simulator rejected the exact paid Capsize action")
	}
	liveSpell, _ := g.Stack.Peek()
	simSpell, _ := simulated.Stack.Peek()
	if !liveSpell.BuybackPaid || !simSpell.BuybackPaid ||
		liveSpell.ManaSpentToCast != simSpell.ManaSpentToCast ||
		g.Players[game.Player1].ManaPool.Total() != simulated.Players[game.Player1].ManaPool.Total() ||
		before.Stack.Size() != 0 || !before.Players[game.Player1].Hand.Contains(cardID) {
		t.Fatal("Simulator payment diverged or changed its source state")
	}
	liveClone := g.Clone()
	engine.resolveTopOfStack(g, &TurnLog{})
	engine.resolveTopOfStack(liveClone, &TurnLog{})
	simulated = engine.Simulator().ResolvePriority(simulated, simPassPolicies())
	for name, result := range map[string]*game.Game{
		"live": g, "clone": liveClone, "Simulator": simulated,
	} {
		if !result.Players[game.Player2].Hand.Contains(target.CardInstanceID) ||
			!result.Players[game.Player1].Hand.Contains(cardID) ||
			result.Stack.Size() != 0 {
			t.Errorf("%s did not resolve target return and buyback", name)
		}
	}
	if !g.Players[game.Player2].Hand.Contains(target.CardInstanceID) ||
		!liveClone.Players[game.Player2].Hand.Contains(target.CardInstanceID) {
		t.Fatal("live and clone did not return the same card to its owner")
	}
}
