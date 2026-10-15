package rules

import (
	"fmt"
	"slices"
	"testing"

	"github.com/natefinch/council4/mtg/cards/authored"
	"github.com/natefinch/council4/mtg/cards/cardset"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// Current Oracle/rulings and official CR revision are recorded in production.json.
type productionCounterCase struct {
	entry    cardset.Entry
	generic  int
	blue     int
	creature bool
	noncreat bool
	ability  bool
}

func productionCounterCases() []productionCounterCase {
	return []productionCounterCase{
		{entry: authored.AnOfferYouCantRefuse, blue: 1, noncreat: true},
		{entry: authored.Counterspell, blue: 2},
		{entry: authored.Negate, generic: 1, blue: 1, noncreat: true},
		{entry: authored.EssenceScatter, generic: 1, blue: 1, creature: true},
		{entry: authored.Disallow, generic: 1, blue: 2, ability: true},
	}
}

func productionCounterSubject(g *game.Game, creature bool) *game.StackObject {
	def := &game.CardDef{CardFace: game.CardFace{Name: "Subject", Types: []types.Card{types.Instant}}}
	if creature {
		def = productionArtifactCreature()
	}
	addImplementationSpellToStack(g, game.Player2, def, nil)
	subject, _ := g.Stack.Peek()
	return subject
}

func TestProductionCountersCastAndKinds(t *testing.T) {
	t.Parallel()
	for _, tc := range productionCounterCases() {
		t.Run(tc.entry.Name, func(t *testing.T) {
			t.Parallel()
			for _, kind := range []game.StackObjectKind{game.StackSpell, game.StackActivatedAbility, game.StackTriggeredAbility} {
				for _, creature := range []bool{false, true} {
					name := "noncreature"
					if creature {
						name = "artifact creature"
					}
					t.Run(fmt.Sprintf("%d/%s", kind, name), func(t *testing.T) {
						g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
						engine := removalTokenEngine(t, tc.entry)
						subject := productionCounterSubject(g, creature)
						subject.Kind = kind
						spell := addCardToHand(g, game.Player1, tc.entry.New())
						g.Players[game.Player1].ManaPool.Add(mana.C, tc.generic)
						g.Players[game.Player1].ManaPool.Add(mana.U, tc.blue)
						setMainPhasePriority(g, game.Player1)
						act := action.CastSpell(spell, []game.Target{game.StackObjectTarget(subject.ID)}, 0, nil)
						want := kind != game.StackSpell && tc.ability || kind == game.StackSpell &&
							(!tc.creature || creature) && (!tc.noncreat || !creature)
						if cast := engine.applyAction(g, game.Player1, act); cast != want {
							t.Fatalf("target kind/type cast = %v, want %v", cast, want)
						}
						if !want {
							if !g.Players[game.Player1].Hand.Contains(spell) || g.Stack.Size() != 1 {
								t.Fatal("illegal cast changed zones")
							}
							return
						}
						engine.resolveTopOfStack(g, &TurnLog{})
						if !g.Stack.IsEmpty() || !g.Players[game.Player1].Graveyard.Contains(spell) {
							t.Fatal("countered object or counterspell remained")
						}
						if kind == game.StackSpell && !g.Players[game.Player2].Graveyard.Contains(subject.SourceID) {
							t.Fatal("countered spell missed owner's graveyard")
						}
					})
				}
			}
			t.Run("payment and target count", func(t *testing.T) {
				g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
				engine := removalTokenEngine(t, tc.entry)
				subject := productionCounterSubject(g, tc.creature)
				spell := addCardToHand(g, game.Player1, tc.entry.New())
				setMainPhasePriority(g, game.Player1)
				if engine.applyAction(g, game.Player1, action.CastSpell(spell, []game.Target{game.StackObjectTarget(subject.ID)}, 0, nil)) {
					t.Fatal("cast without printed cost")
				}
				g.Players[game.Player1].ManaPool.Add(mana.C, tc.generic)
				g.Players[game.Player1].ManaPool.Add(mana.U, tc.blue)
				if engine.applyAction(g, game.Player1, action.CastSpell(spell, nil, 0, nil)) ||
					engine.applyAction(g, game.Player1, action.CastSpell(spell, []game.Target{game.StackObjectTarget(subject.ID), game.StackObjectTarget(subject.ID)}, 0, nil)) {
					t.Fatal("incorrect target cardinality accepted")
				}
			})
		})
	}
}

func TestProductionCountersDepartedTargetAndClone(t *testing.T) {
	t.Parallel()
	for _, tc := range productionCounterCases() {
		t.Run(tc.entry.Name, func(t *testing.T) {
			t.Parallel()
			g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
			engine := removalTokenEngine(t, tc.entry)
			subject := productionCounterSubject(g, tc.creature)
			counter := addImplementationSpellToStack(g, game.Player1, tc.entry.New(), []game.Target{game.StackObjectTarget(subject.ID)})
			counterObject, _ := g.Stack.Peek()
			branch := g.Clone()
			engine.resolveTopOfStack(branch, &TurnLog{})
			if g.Stack.Size() != 2 || branch.Stack.Size() != 0 {
				t.Fatal("clone aliases stack or lost program execution")
			}
			if !counterTargetStackObject(g, &game.StackObject{Targets: []game.Target{game.StackObjectTarget(subject.ID)}}, 0, false, game.CounteredSpellGraveyard) {
				t.Fatal("failed to remove target")
			}
			// Recasting the same physical card has a new stack identity.
			moveCardBetweenZones(g, game.Player2, subject.SourceID, zone.Graveyard, zone.Hand)
			g.Players[game.Player2].Hand.Remove(subject.SourceID)
			newSubject := *subject
			newSubject.ID = g.IDGen.Next()
			g.Stack.RemoveByID(counterObject.ID)
			g.Stack.Push(&newSubject)
			g.Stack.Push(counterObject)
			engine.resolveTopOfStack(g, &TurnLog{})
			if g.Stack.Size() != 1 || namedTokenCount(g, game.Player2, "Treasure") != 0 ||
				!g.Players[game.Player1].Graveyard.Contains(counter) {
				t.Fatal("departed target gave payoff or followed physical card")
			}
		})
	}
}

func TestProductionCountersRetargetedCopy(t *testing.T) {
	t.Parallel()
	for _, tc := range productionCounterCases() {
		t.Run(tc.entry.Name, func(t *testing.T) {
			t.Parallel()
			g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
			engine := removalTokenEngine(t, tc.entry)
			first := productionCounterSubject(g, tc.creature)
			second := productionCounterSubject(g, tc.creature)
			addImplementationSpellToStack(g, game.Player1, tc.entry.New(), []game.Target{game.StackObjectTarget(first.ID)})
			original, _ := g.Stack.Peek()
			copySource := &game.StackObject{ID: g.IDGen.Next(), Controller: game.Player4, Targets: []game.Target{game.StackObjectTarget(original.ID)}}
			agents := [game.NumPlayers]PlayerAgent{game.Player4: authoredChoicePolicy{choose: func(_ PlayerObservation, request game.ChoiceRequest) []int {
				return chooseAuthoredTargets(t, request, game.StackObjectTarget(second.ID))
			}}}
			newEffectResolver(engine, g, copySource, agents, &TurnLog{}).resolveInstruction(&game.Instruction{Primitive: game.CopyStackObject{
				Object: game.TargetStackObjectReference(0), MayChooseNewTargets: true,
			}})
			engine.resolveTopOfStack(g, &TurnLog{})
			if _, ok := stackObjectByID(g, first.ID); !ok {
				t.Fatal("retargeting copy changed original")
			}
			if _, ok := stackObjectByID(g, second.ID); ok {
				t.Fatal("retargeted copy failed")
			}
			engine.resolveTopOfStack(g, &TurnLog{})
			if !g.Stack.IsEmpty() {
				t.Fatal("original did not counter its original target")
			}
		})
	}
}

func TestProductionOfferTokensAndCounterReplacements(t *testing.T) {
	t.Parallel()
	for _, scenario := range []string{"normal", "uncounterable", "exile", "controlled", "copy subject", "becomes creature"} {
		t.Run(scenario, func(t *testing.T) {
			t.Parallel()
			g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
			engine := removalTokenEngine(t, authored.AnOfferYouCantRefuse)
			subject := productionCounterSubject(g, false)
			if scenario == "uncounterable" {
				card, _ := g.GetCardInstance(subject.SourceID)
				card.Def.StaticAbilities = []game.StaticAbility{game.CantBeCounteredStaticBody}
			}
			if scenario == "exile" {
				g.ReplacementEffects = append(g.ReplacementEffects, game.ReplacementEffect{
					ID: g.IDGen.Next(), MatchEvent: game.EventZoneChanged, MatchFromZone: true, FromZone: zone.Stack,
					MatchToZone: true, ToZone: zone.Graveyard, ReplaceToZone: zone.Exile,
				})
			}
			if scenario == "controlled" {
				subject.Controller = game.Player3
			}
			if scenario == "copy subject" {
				subject.Copy = true
				subject.CopyOwner = opt.Val(game.Player2)
			}
			addImplementationSpellToStack(g, game.Player1, authored.AnOfferYouCantRefuse.New(), []game.Target{game.StackObjectTarget(subject.ID)})
			if scenario == "becomes creature" {
				card, _ := g.GetCardInstance(subject.SourceID)
				card.Def.Types = []types.Card{types.Artifact, types.Creature}
			}
			engine.resolveTopOfStack(g, &TurnLog{})
			want := 2
			controller := game.Player2
			if scenario == "controlled" {
				controller = game.Player3
			}
			if scenario == "becomes creature" {
				want = 0
			}
			if namedTokenCount(g, controller, "Treasure") != want {
				t.Fatal("legal/uncounterable target did not create exactly two Treasures for its controller")
			}
			_, remains := stackObjectByID(g, subject.ID)
			if remains != (scenario == "uncounterable" || scenario == "becomes creature") {
				t.Fatal("counterability/target revalidation failed")
			}
			if scenario == "exile" && !g.Players[game.Player2].Exile.Contains(subject.SourceID) {
				t.Fatal("counter destination replacement bypassed")
			}
		})
	}
}

func TestProductionOfferTreasuresPayAndReplace(t *testing.T) {
	t.Parallel()
	for _, scenario := range []string{"payment", "double and tapped", "loses artifact type", "Simulator loses artifact type", "automatic payment", "resolution payment"} {
		t.Run(scenario, func(t *testing.T) {
			t.Parallel()
			replaced := scenario == "double and tapped"
			g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
			engine := removalTokenEngine(t, authored.AnOfferYouCantRefuse)
			subject := productionCounterSubject(g, false)
			if replaced {
				addReplacementPermanent(t, g, game.Player2, tokenDoublingReplacementCardDef())
				addReplacementPermanent(t, g, game.Player2, &game.CardDef{CardFace: game.CardFace{
					Name: "Artifacts enter tapped", Types: []types.Card{types.Enchantment},
					ReplacementAbilities: []game.ReplacementAbility{game.EntersTappedGroupReplacement("Artifacts enter tapped.", game.TriggerControllerAny, types.Artifact)},
				}})
			}
			addImplementationSpellToStack(g, game.Player1, authored.AnOfferYouCantRefuse.New(), []game.Target{game.StackObjectTarget(subject.ID)})
			engine.resolveTopOfStack(g, &TurnLog{})
			count := 0
			var simultaneous game.ObjectID
			for _, permanent := range g.Battlefield {
				if !permanent.Token || permanent.TokenDef.Name != "Treasure" {
					continue
				}
				count++
				if permanent.Owner != game.Player2 || permanent.Controller != game.Player2 || permanent.Tapped != replaced ||
					!slices.Equal(permanent.TokenDef.Subtypes, []types.Sub{types.Treasure}) ||
					len(permanent.TokenDef.Colors) != 0 || len(permanent.TokenDef.ManaAbilities) != 1 {
					t.Fatal("Treasure characteristics/entry replacement incorrect")
				}
				for _, event := range g.Events {
					if event.Kind == game.EventPermanentEnteredBattlefield && event.PermanentID == permanent.ObjectID {
						if event.SimultaneousID == 0 || simultaneous != 0 && simultaneous != event.SimultaneousID {
							t.Fatal("Treasure creation was not simultaneous")
						}
						simultaneous = event.SimultaneousID
					}
				}
			}
			if count != 2 && !replaced || count != 4 && replaced {
				t.Fatal("token replacement quantity wrong")
			}
			if replaced {
				return
			}
			// Self-reference still means this object if it ceases to be an artifact.
			treasures := slices.Clone(g.Battlefield)
			setMainPhasePriority(g, game.Player2)
			agents := [game.NumPlayers]PlayerAgent{game.Player2: &choiceOnlyAgent{choices: [][]int{{1}, {1}}}}
			for index, treasure := range treasures {
				if scenario == "automatic payment" || scenario == "resolution payment" {
					break
				}
				if scenario == "loses artifact type" || scenario == "Simulator loses artifact type" {
					g.ContinuousEffects = append(g.ContinuousEffects, game.ContinuousEffect{
						ID: g.IDGen.Next(), AffectedObjectID: treasure.ObjectID, Layer: game.LayerType,
						SetTypes: []types.Card{types.Enchantment}, Duration: game.DurationPermanent,
					})
				}
				act := action.ActivateAbility(treasure.ObjectID, 0, nil, 0)
				if !containsAction(engine.legalActions(g, game.Player2), act) {
					t.Fatal("Treasure mana activation absent from legal actions")
				}
				if scenario == "Simulator loses artifact type" {
					branch, accepted := engine.Simulator().Apply(g, game.Player2, act, agents)
					if !accepted || namedTokenCount(g, game.Player2, "Treasure") != len(treasures)-index {
						t.Fatal("Simulator rejected mana activation or mutated its input")
					}
					g = branch
				} else if !engine.applyActionWithChoices(g, game.Player2, act, agents, &TurnLog{}) {
					t.Fatal("Treasure's tap/self-sacrifice mana ability was not legal or immediate after type change")
				}
				if !g.Stack.IsEmpty() || g.Players[game.Player2].ManaPool.Amount(mana.U) != index+1 {
					t.Fatal("manual/Simulator mana activation did not immediately choose blue")
				}
			}
			spellDef := &game.CardDef{CardFace: game.CardFace{
				Name: "Treasure-funded spell", Types: []types.Card{types.Instant}, ManaCost: opt.Val(cost.Mana{cost.U, cost.U}),
				SpellAbility: opt.Val(game.Mode{Sequence: []game.Instruction{{Primitive: game.GainLife{
					Player: game.ControllerReference(), Amount: game.Fixed(2),
				}}}}.Ability()),
			}}
			if scenario == "resolution payment" {
				spellDef.ManaCost = opt.Val(cost.Mana{cost.O(0)})
				spellDef.SpellAbility = opt.Val(game.Mode{Sequence: []game.Instruction{
					{Primitive: game.Pay{Payment: game.ResolutionPayment{ManaCost: opt.Val(cost.Mana{cost.U, cost.U})}}, PublishResult: "paid"},
					{Primitive: game.GainLife{Player: game.ControllerReference(), Amount: game.Fixed(2)},
						ResultGate: opt.Val(game.InstructionResultGate{Key: "paid", Succeeded: game.TriTrue})},
				}}.Ability())
			}
			spell := addCardToHand(g, game.Player2, spellDef)
			setMainPhasePriority(g, game.Player2)
			if cast := engine.applyAction(g, game.Player2, action.CastSpell(spell, nil, 0, nil)); !cast ||
				scenario != "resolution payment" && namedTokenCount(g, game.Player2, "Treasure") != 0 {
				t.Fatalf("actual Treasures did not pay colored mana by tap/sacrifice: cast=%v treasures=%d mana=%+v", cast, namedTokenCount(g, game.Player2, "Treasure"), g.Players[game.Player2].ManaPool)
			}
			engine.resolveTopOfStackWithChoices(g, agents, &TurnLog{})
			if g.Players[game.Player2].Life != 42 || namedTokenCount(g, game.Player2, "Treasure") != 0 {
				t.Fatal("Treasure-funded spell did not resolve")
			}
		})
	}
}

func TestProductionDisallowSourceDepartureAndManaAbility(t *testing.T) {
	t.Parallel()
	for _, activated := range []bool{false, true} {
		t.Run(map[bool]string{false: "triggered", true: "activated"}[activated], func(t *testing.T) {
			t.Parallel()
			g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
			engine := removalTokenEngine(t, authored.Disallow)
			content := game.Mode{Sequence: []game.Instruction{{Primitive: game.GainLife{
				Player: game.ControllerReference(), Amount: game.Fixed(9),
			}}}}.Ability()
			def := productionArtifactCreature()
			if activated {
				def.ActivatedAbilities = []game.ActivatedAbility{{
					AdditionalCosts: []cost.Additional{{Kind: cost.AdditionalSacrificeSource, Amount: 1}},
					Content:         content, ZoneOfFunction: zone.Battlefield,
				}}
			} else {
				def.TriggeredAbilities = []game.TriggeredAbility{{
					Trigger: game.TriggerCondition{Pattern: game.TriggerPattern{Event: game.EventPermanentEnteredBattlefield, Source: game.TriggerSourceSelf}},
					Content: content,
				}}
			}
			cardID := addCardToHand(g, game.Player2, def)
			card, _ := g.GetCardInstance(cardID)
			source, ok := createCardPermanent(g, card, game.Player2, zone.Hand)
			if !ok {
				t.Fatal("source entry failed")
			}
			setMainPhasePriority(g, game.Player2)
			if activated {
				if !engine.applyAction(g, game.Player2, action.ActivateAbility(source.ObjectID, 0, nil, 0)) {
					t.Fatal("actual sacrifice-source activation failed")
				}
			} else {
				if !engine.putTriggeredAbilitiesOnStack(g) {
					t.Fatal("actual entry did not trigger")
				}
				movePermanentToZone(g, source, zone.Graveyard)
			}
			target, _ := g.Stack.Peek()
			addImplementationSpellToStack(g, game.Player1, authored.Disallow.New(), []game.Target{game.StackObjectTarget(target.ID)})
			engine.resolveTopOfStack(g, &TurnLog{})
			if !g.Stack.IsEmpty() || g.Players[game.Player2].Life != 40 || !g.Players[game.Player2].Graveyard.Contains(cardID) {
				t.Fatal("Disallow depended on live source, moved source twice or let ability resolve")
			}
		})
	}
	t.Run("mana ability is not a stack target", func(t *testing.T) {
		t.Parallel()
		g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
		engine := removalTokenEngine(t, authored.Disallow)
		source := addCombatPermanent(g, game.Player2, &game.CardDef{CardFace: game.CardFace{
			Name: "Mana artifact", Types: []types.Card{types.Artifact}, ManaAbilities: []game.ManaAbility{game.TapManaAbility(mana.U)},
		}})
		setMainPhasePriority(g, game.Player2)
		if !engine.applyAction(g, game.Player2, action.ActivateAbility(source.ObjectID, 0, nil, 0)) ||
			!g.Stack.IsEmpty() || g.Players[game.Player2].ManaPool.Amount(mana.U) != 1 {
			t.Fatal("mana ability did not resolve immediately")
		}
		spell := addCardToHand(g, game.Player1, authored.Disallow.New())
		g.Players[game.Player1].ManaPool.Add(mana.U, 3)
		g.Turn.PriorityPlayer = game.Player1
		if engine.applyAction(g, game.Player1, action.CastSpell(spell, []game.Target{game.StackObjectTarget(source.ObjectID)}, 0, nil)) {
			t.Fatal("Disallow targeted mana ability/permanent")
		}
	})
}

func TestProductionDisallowDelayedTrigger(t *testing.T) {
	t.Parallel()
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := removalTokenEngine(t, authored.Disallow)
	source := &game.StackObject{ID: g.IDGen.Next(), Controller: game.Player2}
	newEffectResolver(engine, g, source, [game.NumPlayers]PlayerAgent{}, &TurnLog{}).resolveInstruction(&game.Instruction{Primitive: game.CreateDelayedTrigger{
		Trigger: game.DelayedTriggerDef{
			Timing: game.DelayedAtBeginningOfNextEndStep,
			Content: game.Mode{Sequence: []game.Instruction{{Primitive: game.GainLife{
				Player: game.ControllerReference(), Amount: game.Fixed(9),
			}}}}.Ability(),
		},
	}})
	emitBeginningOfStepEvent(g, game.StepEnd)
	if !engine.putTriggeredAbilitiesOnStack(g) || len(g.DelayedTriggers) != 0 {
		t.Fatal("delayed ability did not trigger exactly once")
	}
	target, _ := g.Stack.Peek()
	addImplementationSpellToStack(g, game.Player1, authored.Disallow.New(), []game.Target{game.StackObjectTarget(target.ID)})
	engine.resolveTopOfStack(g, &TurnLog{})
	emitBeginningOfStepEvent(g, game.StepEnd)
	engine.putTriggeredAbilitiesOnStack(g)
	if !g.Stack.IsEmpty() || g.Players[game.Player2].Life != 40 {
		t.Fatal("countered next-occurrence delayed ability triggered again")
	}
}
