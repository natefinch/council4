package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/cards"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

type exactNumericPolicy struct {
	simPassPolicy

	announcement action.Action
	proposed     bool
}

func (p *exactNumericPolicy) ChooseAction(_ PlayerObservation, _ []action.Action) action.Action {
	if !p.proposed {
		p.proposed = true
		return p.announcement
	}
	return action.Pass()
}

func numericDamageSpell() *game.CardDef {
	return &game.CardDef{CardFace: game.CardFace{
		Name: "Exact X damage", Types: []types.Card{types.Instant},
		ManaCost: opt.Val(cost.Mana{cost.X, cost.R}),
		SpellAbility: opt.Val(game.Mode{
			Targets: []game.TargetSpec{{MinTargets: 1, MaxTargets: 1, Allow: game.TargetAllowPlayer}},
			Sequence: []game.Instruction{{Primitive: game.Damage{
				Amount:    game.Dynamic(game.DynamicAmount{Kind: game.DynamicAmountX}),
				Recipient: game.PlayerDamageRecipient(game.TargetPlayerReference(0)),
			}}},
		}.Ability()),
	}}
}

func TestExactNumericFreeAndPaidResolution(t *testing.T) {
	for _, free := range []bool{false, true} {
		for _, name := range []string{"multikicker", "additional X", "mana X"} {
			t.Run(name+map[bool]string{false: "/paid", true: "/free"}[free], func(t *testing.T) {
				g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
				engine := NewEngine(nil)
				def := numericDamageSpell()
				switch name {
				case "multikicker":
					def = multikickerCreature()
				case "additional X":
					def.ManaCost = opt.Val(cost.Mana{cost.R})
					def.AdditionalCosts = []cost.Additional{{Kind: cost.AdditionalPayLife, AmountFromX: true}}
				default:
				}
				cardID := addCardToHand(g, game.Player1, def)
				g.Players[game.Player1].ManaPool.Add(mana.C, 22)
				g.Players[game.Player1].ManaPool.Add(mana.R, 1)
				numberRequests := 0
				agent := &controlScript{choice: func(_ PlayerObservation, request game.ChoiceRequest) []int {
					if request.Kind == game.ChoiceNumber {
						numberRequests++
						if len(request.Options) != 0 || !request.NumberRange.Exists {
							t.Fatal("numeric announcement enumerated its values")
						}
						return []int{21}
					}
					for _, option := range request.Options {
						if len(option.Targets) == 1 && option.Targets[0] == game.PlayerTarget(game.Player2) {
							return []int{option.Index}
						}
					}
					return request.DefaultSelection
				}}
				agents := [game.NumPlayers]PlayerAgent{agent}
				log := &TurnLog{}
				accepted := false
				if free {
					accepted = engine.chooseAndCastFreeSpellFromSource(g, g.Players[game.Player1], game.Player1, cardID,
						zone.Hand, game.Selection{}, opt.V[int]{}, false, agents, log)
				} else {
					accepted = engine.castPaidSpellFromSource(g, g.Players[game.Player1], game.Player1, cardID, zone.Hand, agents, log)
				}
				if !accepted {
					t.Fatal("exact resolution-time announcement was rejected")
				}
				engine.resolveTopOfStackWithChoices(g, agents, log)
				if name == "multikicker" {
					permanent, present := findPermanentByCardID(g, cardID)
					if !present || permanent.Counters.Get(counter.PlusOnePlusOne) != 21 {
						t.Fatal("resolution-time multikicker count was lost")
					}
				} else {
					damage := 21
					if free && name == "mana X" {
						damage = 0
					}
					if g.Players[game.Player2].Life != 40-damage {
						t.Fatal("resolution-time X did not reach the actual effect")
					}
					if name == "additional X" && g.Players[game.Player1].Life != 19 {
						t.Fatal("the exact additional cost was not paid")
					}
				}
				wantNumbers := 1
				if free && name == "mana X" {
					wantNumbers = 0
				}
				if numberRequests != wantNumbers {
					t.Fatal("free mana-cost X was not fixed to zero or an additional numeric choice was omitted")
				}
			})
		}
	}
}

func TestExactNumericXXCostModifiersAndAtomicRejection(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	setMainPhasePriority(g, game.Player1)
	def := numericDamageSpell()
	def.ManaCost = opt.Val(cost.Mana{cost.X, cost.X, cost.R, cost.R})
	card := addCardToHand(g, game.Player1, def)
	g.CostModifiers = []game.CostModifier{
		{Kind: game.CostModifierSpell, GenericIncrease: 1},
		{Kind: game.CostModifierSpell, GenericReduction: 40, MinimumGeneric: 3},
	}
	g.Players[game.Player1].ManaPool.Add(mana.C, 3)
	g.Players[game.Player1].ManaPool.Add(mana.R, 2)
	for _, x := range []int{-1, 22} {
		bad := action.CastSpell(card, []game.Target{game.PlayerTarget(game.Player2)}, x, nil)
		if engine.applyAction(g, game.Player1, bad) || !g.Players[game.Player1].Hand.Contains(card) ||
			g.Players[game.Player1].ManaPool.Total() != 5 || !g.Stack.IsEmpty() {
			t.Fatal("invalid or unpayable XX announcement partially changed resources")
		}
	}
	good := action.CastSpell(card, []game.Target{game.PlayerTarget(game.Player2)}, 21, nil)
	branch, accepted := engine.Simulator().Apply(g, game.Player1, good, simPassPolicies())
	if !accepted {
		t.Fatal("payable exact XX announcement rejected with modifiers and a minimum")
	}
	if branch.Players[game.Player1].ManaPool.Total() != 0 || g.Players[game.Player1].ManaPool.Total() != 5 {
		t.Fatalf("XX modified payment left branch/source mana=%d/%d, want 0/5",
			branch.Players[game.Player1].ManaPool.Total(), g.Players[game.Player1].ManaPool.Total())
	}
	branch = engine.Simulator().ResolvePriority(branch, simPassPolicies())
	if branch.Players[game.Player2].Life != 19 {
		t.Fatal("the exact XX announcement did not resolve with X=21")
	}
}

func TestExactNumericCastingZonesAndFaces(t *testing.T) {
	for _, name := range []string{"command", "graveyard", "exile", "library", "modal back", "prepared"} {
		t.Run(name, func(t *testing.T) {
			g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
			engine := NewEngine(nil)
			setMainPhasePriority(g, game.Player1)
			def := numericDamageSpell()
			from, face := zone.Hand, game.FaceFront
			if name == "modal back" || name == "prepared" {
				spell := def.CardFace
				def = vanillaCreature("Numeric face source", 2, 2)
				if name == "modal back" {
					def.Layout, def.Back = game.LayoutModalDFC, opt.Val(spell)
					face = game.FaceBack
				} else {
					def.Layout, def.Alternate = game.LayoutPrepare, opt.Val(spell)
					from, face = zone.Battlefield, game.FaceAlternate
				}
			}
			var card game.ObjectID
			if name == "prepared" {
				permanent := addCombatPermanent(g, game.Player1, def)
				permanent.Prepared = true
				card = permanent.CardInstanceID
			} else {
				card = addCardToHand(g, game.Player1, def)
				switch name {
				case "command":
					from = zone.Command
					g.CommanderIDs[card] = true
				case "graveyard":
					from = zone.Graveyard
				case "exile":
					from = zone.Exile
				case "library":
					from = zone.Library
				default:
				}
				if from != zone.Hand {
					g.Players[game.Player1].Hand.Remove(card)
					destination, _ := destinationZone(g, game.Player1, from)
					destination.Add(card)
				}
				switch from {
				case zone.Library:
					castSpellsFromLibraryTopPermanent(g, game.Player1, nil)
				case zone.Graveyard, zone.Exile:
					g.RuleEffects = append(g.RuleEffects, game.RuleEffect{
						Kind: game.RuleEffectCastFromZone, Controller: game.Player1, AffectedPlayer: game.PlayerYou,
						CastFromZone: from, AffectedCardID: card, Duration: game.DurationPermanent,
					})
				default:
				}
			}
			g.Players[game.Player1].ManaPool.Add(mana.C, 21)
			g.Players[game.Player1].ManaPool.Add(mana.R, 1)
			announcement := action.CastSpellFaceFromZone(card, from, face, []game.Target{game.PlayerTarget(game.Player2)}, 21, nil)
			branch, accepted := engine.Simulator().Apply(g, game.Player1, announcement, simPassPolicies())
			if !accepted {
				t.Fatal("exact announcement rejected from a supported zone or face")
			}
			branch = engine.Simulator().ResolvePriority(branch, simPassPolicies())
			if branch.Players[game.Player2].Life != 19 || g.Players[game.Player2].Life != 40 {
				t.Fatal("exact zone/face announcement resolved incorrectly or shared its branch")
			}
		})
	}
}

func TestNumericMultikickerPolicySamplesAreBounded(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	setMainPhasePriority(g, game.Player1)
	for range maxSampledMultikickCount - 2 {
		addCombatCreaturePermanentWithPower(g, game.Player2, 3)
	}
	def := cometStormDef()
	card := addCardToHand(g, game.Player1, def)
	g.Players[game.Player1].ManaPool.Add(mana.C, 40)
	g.Players[game.Player1].ManaPool.Add(mana.R, 1)

	// One X value exercises every sampled kicker count and the target-choice cap.
	g.BeginStaticSourceFrame()
	defer g.EndStaticSourceFrame()
	candidates := engine.appendMultikickedCastActions(g, game.Player1, nil, actionBuild, card, zone.Hand, game.FaceFront, def, 1, nil)
	if len(candidates) == 0 || len(candidates) > maxSampledMultikickCount*maxSampledNumericTargetChoices {
		t.Fatalf("sampled %d casts, want a nonempty bounded sample", len(candidates))
	}
	counts := make(map[int]int)
	for _, candidate := range candidates {
		cast, ok := candidate.CastSpellPayload()
		if !ok || cast.XValue != 1 || cast.KickerCount < 1 || cast.KickerCount > maxSampledMultikickCount ||
			len(cast.Targets) != cast.KickerCount+1 {
			t.Fatalf("invalid multikicker sample: %+v", cast)
		}
		counts[cast.KickerCount]++
		if counts[cast.KickerCount] > maxSampledNumericTargetChoices {
			t.Fatal("one kicker count exceeded the target-choice cap")
		}
	}
	if len(counts) != maxSampledMultikickCount || counts[1] != maxSampledNumericTargetChoices {
		t.Fatal("fixture did not exercise every kicker count and fill the target-choice cap")
	}
}

func TestNumericXPolicySamplesAreBounded(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	setMainPhasePriority(g, game.Player1)
	addCardToHand(g, game.Player1, numericDamageSpell())
	g.Players[game.Player1].ManaPool.Add(mana.C, maxSampledXValue+5)
	g.Players[game.Player1].ManaPool.Add(mana.R, 1)

	candidates := engine.Simulator().LegalActions(g, game.Player1)
	if len(candidates) > (maxSampledXValue+1)*maxSampledNumericTargetChoices+1 {
		t.Fatal("numeric policy candidates exceeded their concrete independent bound")
	}
	counts := make(map[int]int)
	for _, candidate := range candidates {
		cast, ok := candidate.CastSpellPayload()
		if !ok {
			continue
		}
		if cast.XValue < 0 || cast.XValue > maxSampledXValue {
			t.Fatalf("X sample exceeded the policy cap: %d", cast.XValue)
		}
		counts[cast.XValue]++
	}
	if len(counts) != maxSampledXValue+1 {
		t.Fatal("fixture did not exercise every sampled X value")
	}
}

func TestExactNumericMultikickerTargetCardinality(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	setMainPhasePriority(g, game.Player1)
	var targets []game.Target
	for range 30 {
		permanent := addCombatCreaturePermanentWithPower(g, game.Player2, 3)
		targets = append(targets, game.PermanentTarget(permanent.ObjectID))
	}
	targets = targets[len(targets)-22:]
	card := addCardToHand(g, game.Player1, cometStormDef())
	g.Players[game.Player1].ManaPool.Add(mana.C, 40)
	g.Players[game.Player1].ManaPool.Add(mana.R, 1)
	announcement := action.CastMultikickedSpellFaceFromZone(card, zone.Hand, game.FaceFront, targets, 1, nil, 21)
	branch, accepted := engine.Simulator().Apply(g, game.Player1, announcement, simPassPolicies())
	if !accepted {
		t.Fatal("actual count+one targets were constrained by the legacy MaxTargets sample hint")
	}
	branch = engine.Simulator().ResolvePriority(branch, simPassPolicies())
	for _, target := range targets {
		permanent, present := permanentByObjectID(branch, target.PermanentID)
		if !present || permanent.MarkedDamage != 1 {
			t.Fatal("exact multikicker target segmentation lost an actual recipient")
		}
	}
}

func TestExactNumericActivationWithReductionAndAtomicFailure(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	setMainPhasePriority(g, game.Player1)
	source := addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name: "Exact X activator", Types: []types.Card{types.Creature},
		Power: opt.Val(game.PT{Value: 2}), Toughness: opt.Val(game.PT{Value: 2}),
		ActivatedAbilities: []game.ActivatedAbility{{
			ManaCost:        opt.Val(cost.Mana{cost.X}),
			AdditionalCosts: []cost.Additional{{Kind: cost.AdditionalTap}},
			Content:         numericDamageSpell().SpellAbility.Val,
		}},
	}})
	source.SummoningSick = false
	g.CostModifiers = []game.CostModifier{{Kind: game.CostModifierAbility, GenericReduction: 20, MinimumGeneric: 1}}
	g.Players[game.Player1].ManaPool.Add(mana.C, 1)
	bad := action.ActivateAbility(source.ObjectID, 0, []game.Target{game.PlayerTarget(game.Player2)}, 22)
	if engine.applyAction(g, game.Player1, bad) || source.Tapped || g.Players[game.Player1].ManaPool.Total() != 1 || !g.Stack.IsEmpty() {
		t.Fatal("unpayable exact activation partially paid its cost")
	}
	good := action.ActivateAbility(source.ObjectID, 0, []game.Target{game.PlayerTarget(game.Player2)}, 21)
	branch, accepted := engine.Simulator().Apply(g, game.Player1, good, simPassPolicies())
	if !accepted || source.Tapped || g.Players[game.Player1].ManaPool.Total() != 1 {
		t.Fatal("exact reduced activation was rejected or shared mutable state")
	}
	branch = engine.Simulator().ResolvePriority(branch, simPassPolicies())
	if branch.Players[game.Player2].Life != 19 || branch.Players[game.Player1].ManaPool.Total() != 0 {
		t.Fatal("exact activation did not preserve X while paying the reduced cost")
	}
}

func TestExactNumericChoiceProtocolAndUnpayableResolution(t *testing.T) {
	request := game.ChoiceRequest{
		Kind: game.ChoiceNumber, MinChoices: 1, MaxChoices: 1,
		NumberRange: opt.Val(game.NumberRange{Minimum: 0}), DefaultSelection: []int{0},
	}
	if !choiceSelectionValid(request, []int{1_000_000}) || choiceSelectionValid(request, []int{-1}) ||
		choiceSelectionValid(request, []int{1, 2}) || choiceSelectionValid(request, nil) {
		t.Fatal("numeric choices were treated as sampled option indices")
	}
	bounded := request
	bounded.NumberRange = opt.Val(game.NumberRange{Minimum: 2, Maximum: opt.Val(25)})
	if !choiceSelectionValid(bounded, []int{25}) || choiceSelectionValid(bounded, []int{26}) ||
		choiceSelectionValid(bounded, []int{1}) {
		t.Fatal("numeric choice ignored its exact domain")
	}
	for _, answer := range []int{-1, 21} {
		g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
		engine := NewEngine(nil)
		cardID := addCardToHand(g, game.Player1, multikickerCreature())
		card, _ := g.GetCardInstance(cardID)
		version := card.ZoneVersion
		agent := &controlScript{choice: func(_ PlayerObservation, choice game.ChoiceRequest) []int {
			if choice.Kind == game.ChoiceNumber {
				return []int{answer}
			}
			return choice.DefaultSelection
		}}
		accepted := engine.chooseAndCastFreeSpellFromSource(g, g.Players[game.Player1], game.Player1, cardID,
			zone.Hand, game.Selection{}, opt.V[int]{}, false, [game.NumPlayers]PlayerAgent{agent}, &TurnLog{})
		if g.Execution.InvalidChoices != 1 || g.Players[game.Player1].ManaPool.Total() != 0 {
			t.Fatal("invalid or unpayable numeric proposal lacked explicit diagnostics")
		}
		if answer == 21 {
			if accepted || !g.Stack.IsEmpty() || !g.Players[game.Player1].Hand.Contains(cardID) || card.ZoneVersion != version {
				t.Fatal("unpayable numeric proposal changed its source incarnation")
			}
		} else if !accepted || g.Execution.ChoiceFallbacks != 1 {
			t.Fatal("out-of-domain numeric answer did not use the recorded deterministic fallback")
		}
	}
}

func TestExactNumericControlledResolutionUsesActorResources(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	g.Turn.TurnNumber = 1
	g.PlayerControlEffects = []game.PlayerControlEffect{{
		ID: g.IDGen.Next(), Player: game.Player1, Controller: game.Player3, Turn: 1,
	}}
	def := numericDamageSpell()
	def.ManaCost = opt.Val(cost.Mana{cost.R})
	def.AdditionalCosts = []cost.Additional{{Kind: cost.AdditionalPayLife, AmountFromX: true}}
	cardID := addCardToHand(g, game.Player1, def)
	agent := &controlScript{choice: func(obs PlayerObservation, request game.ChoiceRequest) []int {
		if obs.Player != game.Player1 || obs.DecisionMaker != game.Player3 {
			t.Fatal("numeric announcement conflated actor and decision maker")
		}
		if request.Kind == game.ChoiceNumber {
			return []int{21}
		}
		return request.DefaultSelection
	}}
	agents := [game.NumPlayers]PlayerAgent{game.Player3: agent}
	engine := NewEngine(nil)
	if !engine.chooseAndCastFreeSpellFromSource(g, g.Players[game.Player1], game.Player1, cardID,
		zone.Hand, game.Selection{}, opt.V[int]{}, false, agents, &TurnLog{}) {
		t.Fatal("controlled numeric cast failed")
	}
	if g.Players[game.Player1].Life != 19 || g.Players[game.Player3].Life != 40 {
		t.Fatal("numeric costs were charged to the decision maker instead of the actor")
	}
}

func TestExactNumericReviewRegressions(t *testing.T) {
	t.Run("complete ability cost override", func(t *testing.T) {
		g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
		setMainPhasePriority(g, game.Player1)
		equipment := addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
			Name: "Printed equip three", Types: []types.Card{types.Artifact}, Subtypes: []types.Sub{types.Equipment},
			ActivatedAbilities: []game.ActivatedAbility{game.EquipActivatedAbility(cost.Mana{cost.O(3)})},
		}})
		creature := addCombatCreaturePermanentWithPower(g, game.Player1, 3)
		g.CostModifiers = []game.CostModifier{{
			Kind: game.CostModifierAbility, AbilityKeyword: game.Equip, SetManaCost: opt.Val(cost.Mana{}),
		}}
		if !NewEngine(nil).applyAction(g, game.Player1,
			action.ActivateAbility(equipment.ObjectID, 0, []game.Target{game.PermanentTarget(creature.ObjectID)}, 0)) {
			t.Fatal("full equip-cost setting was lost between validation and payment")
		}
	})
	t.Run("exact zero targets with constraint", func(t *testing.T) {
		g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
		setMainPhasePriority(g, game.Player1)
		def := &game.CardDef{CardFace: game.CardFace{
			Name: "By Force", Types: []types.Card{types.Sorcery}, ManaCost: opt.Val(cost.Mana{cost.X, cost.R}),
			SpellAbility: opt.Val(game.Mode{
				Targets: []game.TargetSpec{{
					CountEqualsX: true, MinTargets: 0, MaxTargets: 20,
					Constraint: "x target artifacts", Allow: game.TargetAllowPermanent,
				}},
				Sequence: []game.Instruction{{Primitive: game.Destroy{Object: game.AllTargetPermanentsReference(0)}}},
			}.Ability()),
		}}
		card := addCardToHand(g, game.Player1, def)
		g.Players[game.Player1].ManaPool.Add(mana.R, 1)
		if !NewEngine(nil).applyAction(g, game.Player1, action.CastSpell(card, nil, 0, nil)) {
			t.Fatal("a bound zero count was mistaken for an unspecified single target")
		}
	})
	t.Run("numeric fallback respects source prohibition", func(t *testing.T) {
		g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
		card := addCardToGraveyard(g, game.Player1, cometStormDef())
		g.RuleEffects = []game.RuleEffect{{
			Kind: game.RuleEffectCantCastFromZones, Controller: game.Player1,
			AffectedPlayer: game.PlayerYou, CantCastFromZones: []zone.Type{zone.Graveyard}, Duration: game.DurationPermanent,
		}}
		agent := &controlScript{choice: func(_ PlayerObservation, request game.ChoiceRequest) []int {
			return request.DefaultSelection
		}}
		if NewEngine(nil).chooseAndCastFreeSpellFromSource(g, g.Players[game.Player1], game.Player1, card,
			zone.Graveyard, game.Selection{}, opt.V[int]{}, false, [game.NumPlayers]PlayerAgent{agent}, &TurnLog{}) ||
			!g.Players[game.Player1].Graveyard.Contains(card) || !g.Stack.IsEmpty() {
			t.Fatal("numeric fallback bypassed an authoritative source-zone prohibition")
		}
	})
	t.Run("exact gift proposal requires a living opponent", func(t *testing.T) {
		g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
		setMainPhasePriority(g, game.Player1)
		def := giftSpellDef()
		def.Types, def.ManaCost = []types.Card{types.Instant}, opt.Val(cost.Mana{cost.O(0)})
		card := addCardToHand(g, game.Player1, def)
		plain, _ := action.CastSpell(card, nil, 0, nil).CastSpellPayload()
		plain.GiftPromised, plain.GiftRecipient = true, game.Player1
		if _, accepted := NewEngine(nil).Simulator().Apply(g, game.Player1, action.AnnounceSpell(plain), simPassPolicies()); accepted {
			t.Fatal("exact announcement promised a gift to its own caster")
		}
		g.Players[game.Player3].Eliminated = true
		plain.GiftRecipient = game.Player3
		if _, accepted := NewEngine(nil).Simulator().Apply(g, game.Player1, action.AnnounceSpell(plain), simPassPolicies()); accepted {
			t.Fatal("exact announcement promised a gift to a departed opponent")
		}
	})
}

func TestExactNumericFreeMultikickerTargetsAreCompact(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	var wanted []game.Target
	for range 30 {
		permanent := addCombatCreaturePermanentWithPower(g, game.Player2, 3)
		wanted = append(wanted, game.PermanentTarget(permanent.ObjectID))
	}
	wanted = wanted[len(wanted)-22:]
	cardID := addCardToHand(g, game.Player1, cometStormDef())
	g.Players[game.Player1].ManaPool.Add(mana.C, 21)
	targetRequest := false
	agent := &controlScript{choice: func(_ PlayerObservation, request game.ChoiceRequest) []int {
		if request.Kind == game.ChoiceNumber {
			return []int{21}
		}
		if request.Kind == game.ChoiceTarget {
			targetRequest = true
			if request.MinChoices != 22 || request.MaxChoices != 22 || len(request.Options) > 34 {
				t.Fatal("exact target selection materialized combinations rather than individual candidates")
			}
			var selected []int
			for _, option := range request.Options {
				for _, target := range wanted {
					if len(option.Targets) == 1 && option.Targets[0] == target {
						selected = append(selected, option.Index)
					}
				}
			}
			return selected
		}
		return request.DefaultSelection
	}}
	if !NewEngine(nil).chooseAndCastFreeSpellFromSource(g, g.Players[game.Player1], game.Player1, cardID,
		zone.Hand, game.Selection{}, opt.V[int]{}, false, [game.NumPlayers]PlayerAgent{agent}, &TurnLog{}) {
		t.Fatal("compact free multikicker declaration was rejected")
	}
	obj, present := g.Stack.Peek()
	if !present || obj.KickerCount != 21 || obj.XValue != 0 || len(obj.Targets) != 22 ||
		!targetRequest || g.Players[game.Player1].ManaPool.Total() != 0 {
		t.Fatal("exact free-cast count, target cardinality or cost was lost")
	}
}

func TestExactNumericActivationTargetCountAboveSampleHint(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	setMainPhasePriority(g, game.Player1)
	var targets []game.Target
	for range 22 {
		permanent := addCombatCreaturePermanentWithPower(g, game.Player2, 3)
		targets = append(targets, game.PermanentTarget(permanent.ObjectID))
	}
	targets = targets[1:]
	source := addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name: "X targets ability", Types: []types.Card{types.Artifact},
		ActivatedAbilities: []game.ActivatedAbility{{
			ManaCost: opt.Val(cost.Mana{cost.X}),
			Content: game.Mode{
				Targets: []game.TargetSpec{{
					CountEqualsX: true, MinTargets: 0, MaxTargets: 20,
					Constraint: "x target permanents", Allow: game.TargetAllowPermanent,
				}},
				Sequence: []game.Instruction{{Primitive: game.Damage{
					Amount: game.Fixed(1), Recipient: game.AnyTargetDamageRecipient(0), EachTarget: true,
				}}},
			}.Ability(),
		}},
	}})
	g.Players[game.Player1].ManaPool.Add(mana.C, 21)
	engine := NewEngine(nil)
	branch, accepted := engine.Simulator().Apply(g, game.Player1,
		action.ActivateAbility(source.ObjectID, 0, targets, 21), simPassPolicies())
	if !accepted {
		t.Fatal("activation targets were limited by a policy cardinality hint")
	}
	branch = engine.Simulator().ResolvePriority(branch, simPassPolicies())
	for _, target := range targets {
		permanent, present := permanentByObjectID(branch, target.PermanentID)
		if !present || permanent.MarkedDamage != 1 {
			t.Fatal("activation target count did not survive actual stack resolution")
		}
	}
}

func TestExactNumericProductionDeathDeniedAboveLegacyWidth(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	setMainPhasePriority(g, game.Player1)
	var targets []game.Target
	for range 21 {
		card := addCardToGraveyard(g, game.Player1, vanillaCreature("Returned creature", 2, 2))
		targets = append(targets, currentCardTarget(t, g, card))
	}
	registry := cards.NewDefaultRegistry()
	def := registry.Lookup("Death Denied")
	card := addCardToHand(g, game.Player1, def)
	g.Players[game.Player1].ManaPool.Add(mana.C, 21)
	g.Players[game.Player1].ManaPool.Add(mana.B, 2)
	engine := NewEngineWithPrograms(nil, registry.Programs())
	branch, accepted := engine.Simulator().Apply(g, game.Player1, action.CastSpell(card, targets, 21, nil), simPassPolicies())
	if !accepted {
		t.Fatal("production Death Denied rejected the payable exact declaration")
	}
	branch = engine.Simulator().ResolvePriority(branch, simPassPolicies())
	if branch.Players[game.Player1].Hand.Size() != 21 {
		t.Fatalf("production Death Denied returned %d cards, want all 21 chosen targets", branch.Players[game.Player1].Hand.Size())
	}
}

func TestExactNumericAnnouncementsAbovePolicySamples(t *testing.T) {
	for _, name := range []string{"X", "multikicker"} {
		for _, simulator := range []bool{false, true} {
			t.Run(name+map[bool]string{false: "/live", true: "/Simulator"}[simulator], func(t *testing.T) {
				g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
				engine := NewEngine(nil)
				setMainPhasePriority(g, game.Player1)
				def := numericDamageSpell()
				if name == "multikicker" {
					def = multikickerCreature()
				}
				card := addCardToHand(g, game.Player1, def)
				g.Players[game.Player1].ManaPool.Add(mana.C, 22)
				g.Players[game.Player1].ManaPool.Add(mana.R, 1)
				announcement := action.CastSpell(card, []game.Target{game.PlayerTarget(game.Player2)}, 21, nil)
				if name == "multikicker" {
					announcement = action.CastMultikickedSpellFaceFromZone(card, zone.Hand, game.FaceFront, nil, 0, nil, 21)
				}
				policies := simPassPolicies()
				if simulator {
					branch, accepted := engine.Simulator().Apply(g, game.Player1, announcement, policies)
					if !accepted || !g.Players[game.Player1].Hand.Contains(card) {
						t.Fatal("Simulator rejected a payable exact announcement or changed its source")
					}
					g = engine.Simulator().ResolvePriority(branch, policies)
				} else {
					policies[game.Player1] = &exactNumericPolicy{announcement: announcement}
					engine.runPriorityLoop(g, policies, &TurnLog{})
				}
				if g.Execution.InvalidActions != 0 || g.Execution.ActionFallbacks != 0 {
					t.Fatal("an exact announcement was treated as an invalid sampled candidate")
				}
				if name == "X" {
					if g.Players[game.Player2].Life != 19 || !g.Players[game.Player1].Graveyard.Contains(card) {
						t.Fatal("the chosen X did not pay and resolve through ordinary damage")
					}
				} else {
					permanent, present := findPermanentByCardID(g, card)
					if !present || permanent.Counters.Get(counter.PlusOnePlusOne) != 21 {
						t.Fatal("the exact multikicker count did not reach the actual entry result")
					}
				}
			})
		}
	}
}
