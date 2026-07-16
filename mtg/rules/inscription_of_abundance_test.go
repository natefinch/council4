package rules

import (
	"slices"
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

func inscriptionOfAbundanceDef() *game.CardDef {
	creature := func(controller game.ControllerRelation) game.TargetSpec {
		return game.TargetSpec{
			MinTargets: 1,
			MaxTargets: 1,
			Constraint: "creature",
			Selection: opt.Val(game.Selection{
				RequiredTypesAny: []types.Card{types.Creature},
				Controller:       controller,
			}),
		}
	}
	return &game.CardDef{CardFace: game.CardFace{
		Name:     "Inscription of Abundance",
		ManaCost: opt.Val(cost.Mana{cost.O(1), cost.G}),
		Types:    []types.Card{types.Instant},
		StaticAbilities: []game.StaticAbility{{
			KeywordAbilities: []game.KeywordAbility{game.KickerKeyword{Cost: cost.Mana{cost.O(2), cost.G}}},
		}},
		SpellAbility: opt.Val(game.AbilityContent{
			Modes: []game.Mode{
				{
					Targets: []game.TargetSpec{creature(game.ControllerAny)},
					Sequence: []game.Instruction{{Primitive: game.AddCounter{
						Object:      game.TargetPermanentReference(0),
						CounterKind: counter.PlusOnePlusOne,
						Amount:      game.Fixed(2),
					}}},
				},
				{
					Targets: []game.TargetSpec{{MinTargets: 1, MaxTargets: 1, Allow: game.TargetAllowPlayer}},
					Sequence: []game.Instruction{{Primitive: game.GainLife{
						Player: game.TargetPlayerReference(0),
						Amount: game.Dynamic(game.DynamicAmount{
							Kind:       game.DynamicAmountGreatestPowerInGroup,
							Multiplier: 1,
							Group: game.PlayerControlledGroup(
								game.TargetPlayerReference(0),
								game.Selection{RequiredTypesAny: []types.Card{types.Creature}},
							),
						}),
					}}},
				},
				{
					Targets: []game.TargetSpec{
						creature(game.ControllerYou),
						creature(game.ControllerNotYou),
					},
					Sequence: []game.Instruction{{Primitive: game.Fight{
						Object:        game.TargetPermanentReference(0),
						RelatedObject: game.TargetPermanentReference(1),
					}}},
				},
			},
			MinModes: 1,
			MaxModes: 1,
			ModeChoiceBonus: game.ModeChoiceBonus{
				Condition:    game.ModeChoiceConditionSpellKicked,
				ReplaceRange: true,
				MinModes:     0,
				MaxModes:     3,
			},
		}),
	}}
}

func TestInscriptionModalCardinalityTracksKicker(t *testing.T) {
	t.Parallel()
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	def := inscriptionOfAbundanceDef()

	plain := modeChoicesForSpellAtBranch(g, game.Player1, def, game.CastBranch{})
	if len(plain) != 3 {
		t.Fatalf("unkicked choices = %v, want three one-mode choices", plain)
	}
	for _, modes := range plain {
		if len(modes) != 1 {
			t.Fatalf("unkicked modes = %v, want exactly one", modes)
		}
	}

	kicked := modeChoicesForSpellAtBranch(g, game.Player1, def, game.CastBranch{Kicked: true})
	if len(kicked) != 8 || !slices.ContainsFunc(kicked, func(modes []int) bool { return len(modes) == 0 }) ||
		!slices.ContainsFunc(kicked, func(modes []int) bool { return slices.Equal(modes, []int{0, 1, 2}) }) {
		t.Fatalf("kicked choices = %v, want all eight subsets", kicked)
	}
	if modesValidForSpellAtBranch(g, game.Player1, def, nil, game.CastBranch{}) {
		t.Fatal("unkicked spell accepted zero modes")
	}
	if !modesValidForSpellAtBranch(g, game.Player1, def, nil, game.CastBranch{Kicked: true}) {
		t.Fatal("kicked spell rejected zero modes")
	}
	if modesValidForSpellAtBranch(g, game.Player1, def, []int{1, 1}, game.CastBranch{Kicked: true}) {
		t.Fatal("kicked spell accepted a duplicate mode")
	}
	if modesValidForSpellAtBranch(g, game.Player1, def, []int{0, 1}, game.CastBranch{}) {
		t.Fatal("unkicked spell accepted multiple modes")
	}
}

func TestInscriptionKickedCastActionsAnnounceSelectedModeTargets(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	spellID := addCardToHand(g, game.Player1, inscriptionOfAbundanceDef())
	addCombatPermanent(g, game.Player1, inscriptionCreature("Mine", 2, 2))
	addCombatPermanent(g, game.Player2, inscriptionCreature("Theirs", 3, 3))
	g.Players[game.Player1].ManaPool.Add(mana.G, 2)
	g.Players[game.Player1].ManaPool.Add(mana.C, 3)
	g.Turn.PriorityPlayer = game.Player1

	var sawUnkickedSingle, sawKickedZero, sawKickedAll bool
	var allModeCastTargets []game.Target
	for _, cast := range castActionsForCard(t, engine, g, game.Player1, spellID) {
		switch {
		case !cast.KickerPaid:
			if len(cast.ChosenModes) != 1 {
				t.Errorf("unkicked modes = %v, want exactly one", cast.ChosenModes)
			}
			sawUnkickedSingle = true
		case len(cast.ChosenModes) == 0:
			sawKickedZero = true
			if len(cast.Targets) != 0 {
				t.Errorf("zero-mode kicked targets = %v, want none", cast.Targets)
			}
		case slices.Equal(cast.ChosenModes, []int{0, 1, 2}):
			sawKickedAll = true
			if allModeCastTargets == nil {
				allModeCastTargets = append([]game.Target(nil), cast.Targets...)
			}
			if len(cast.Targets) != 4 {
				t.Errorf("all-mode kicked targets = %v, want four mode-local targets", cast.Targets)
			}
		default:
		}
	}
	if !sawUnkickedSingle || !sawKickedZero || !sawKickedAll {
		t.Fatalf("cast coverage: unkicked=%t kicked-zero=%t kicked-all=%t", sawUnkickedSingle, sawKickedZero, sawKickedAll)
	}
	if engine.canCastSpellWithKicker(g, game.Player1, spellID,
		[]game.Target{game.PlayerTarget(game.Player1), game.PlayerTarget(game.Player2)},
		0, []int{1, 1}, true) {
		t.Fatal("direct cast validation accepted a duplicate kicked mode")
	}
	if !engine.applyAction(g, game.Player1,
		action.CastKickedSpell(spellID, allModeCastTargets, 0, []int{0, 1, 2})) {
		t.Fatal("all-mode kicked cast action was rejected")
	}
	obj, ok := g.Stack.Peek()
	if !ok || !obj.KickerPaid ||
		!slices.Equal(obj.ChosenModes, []int{0, 1, 2}) ||
		!slices.Equal(obj.TargetCounts, []int{1, 1, 1, 1}) {
		t.Fatalf("stack object = %#v, want kicked all-mode choices and target segmentation", obj)
	}
}

func TestInscriptionAllModesResolveWithIndependentTargetSegments(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	addReplacementPermanent(t, g, game.Player1, anyCounterDoublingReplacementCardDef())
	counterTarget := addCombatPermanent(g, game.Player1, inscriptionCreature("Counter target", 1, 1))
	ownFighter := addCombatPermanent(g, game.Player1, inscriptionCreature("Own fighter", 3, 3))
	opponentFighter := addCombatPermanent(g, game.Player2, inscriptionCreature("Opponent fighter", 4, 4))
	def := inscriptionOfAbundanceDef()
	spellID := addCardToHand(g, game.Player1, def)
	g.Players[game.Player1].Hand.Remove(spellID)
	g.Stack.Push(&game.StackObject{
		ID:           g.IDGen.Next(),
		Kind:         game.StackSpell,
		SourceID:     spellID,
		SourceZone:   zone.Hand,
		Controller:   game.Player1,
		KickerPaid:   true,
		ChosenModes:  []int{0, 1, 2},
		Targets:      []game.Target{game.PermanentTarget(counterTarget.ObjectID), game.PlayerTarget(game.Player2), game.PermanentTarget(ownFighter.ObjectID), game.PermanentTarget(opponentFighter.ObjectID)},
		TargetCounts: []int{1, 1, 1, 1},
	})

	engine.resolveTopOfStack(g, &TurnLog{})

	if got := counterTarget.Counters.Get(counter.PlusOnePlusOne); got != 4 {
		t.Fatalf("counter target has %d counters, want 4 after replacement", got)
	}
	if got := g.Players[game.Player2].Life; got != 44 {
		t.Fatalf("target player life = %d, want 44 from greatest power 4", got)
	}
	if got := ownFighter.MarkedDamage; got != 4 {
		t.Fatalf("own fighter damage = %d, want 4", got)
	}
	if got := opponentFighter.MarkedDamage; got != 3 {
		t.Fatalf("opponent fighter damage = %d, want 3", got)
	}
}

func TestInscriptionGreatestPowerEmptyGroupIsZero(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	content := inscriptionOfAbundanceDef().SpellAbility.Val
	obj := &game.StackObject{
		Controller:   game.Player1,
		ChosenModes:  []int{1},
		Targets:      []game.Target{game.PlayerTarget(game.Player3)},
		TargetCounts: []int{1},
	}
	NewEngine(nil).resolveAbilityContentWithChoices(g, obj, content, [game.NumPlayers]PlayerAgent{}, &TurnLog{})
	if got := g.Players[game.Player3].Life; got != 40 {
		t.Fatalf("empty-group target life = %d, want 40", got)
	}
}

func TestInscriptionFightTargetControlRecheckedAtResolution(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	ownFighter := addCombatPermanent(g, game.Player1, inscriptionCreature("Changed control", 3, 3))
	opponentFighter := addCombatPermanent(g, game.Player2, inscriptionCreature("Opponent", 4, 4))
	spellID := addCardToHand(g, game.Player1, inscriptionOfAbundanceDef())
	g.Players[game.Player1].Hand.Remove(spellID)
	g.Stack.Push(&game.StackObject{
		ID:           g.IDGen.Next(),
		Kind:         game.StackSpell,
		SourceID:     spellID,
		SourceZone:   zone.Hand,
		Controller:   game.Player1,
		ChosenModes:  []int{2},
		Targets:      []game.Target{game.PermanentTarget(ownFighter.ObjectID), game.PermanentTarget(opponentFighter.ObjectID)},
		TargetCounts: []int{1, 1},
	})
	ownFighter.Controller = game.Player2

	engine.resolveTopOfStack(g, &TurnLog{})
	if ownFighter.MarkedDamage != 0 || opponentFighter.MarkedDamage != 0 {
		t.Fatalf("illegal fight dealt damage: own=%d opponent=%d", ownFighter.MarkedDamage, opponentFighter.MarkedDamage)
	}
}

func inscriptionCreature(name string, power, toughness int) *game.CardDef {
	return &game.CardDef{CardFace: game.CardFace{
		Name:      name,
		Types:     []types.Card{types.Creature},
		Power:     opt.Val(game.PT{Value: power}),
		Toughness: opt.Val(game.PT{Value: toughness}),
	}}
}
