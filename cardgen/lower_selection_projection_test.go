package cardgen

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/compare"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/id"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/mtg/rules"
	"github.com/natefinch/council4/opt"
)

func nestedQuantitySelection(player game.PlayerReference) game.Selection {
	group := game.PlayerControlledGroup(player, game.Selection{RequiredTypes: []types.Card{types.Creature}})
	return nestedGroupSelection(group)
}

func nestedGroupSelection(group game.GroupReference) game.Selection {
	return game.Selection{AnyOf: []game.Selection{{AnyOf: []game.Selection{{
		ManaValueDynamic: opt.Val(game.ManaValueDynamicBound{
			Kind: game.DynamicAmountCountSelector, Multiplier: 1, Group: &group,
		}),
	}}}}}
}

func assertNestedSelectionPlayer(t *testing.T, selection game.Selection, player game.PlayerReference) {
	t.Helper()
	bound := selection.AnyOf[0].AnyOf[0].ManaValueDynamic.Val
	got, ok := bound.Group.PlayerAnchor()
	if !ok || got != player {
		t.Fatalf("nested AnyOf group player=%+v, want %+v", got, player)
	}
}

func TestRecursiveSelectionTargetProjection(t *testing.T) {
	t.Parallel()
	player := game.TargetPlayerReference(0)
	selection := nestedQuantitySelection(player)
	if problems := selection.Validate(); len(problems) != 0 {
		t.Fatalf("selection is not a valid modeled operand: %v", problems)
	}
	projected, ok := transformSelectionTargetIndices(selection, rebaseTransform(2, 1))
	if !ok {
		t.Fatal("valid nested AnyOf selection refused")
	}
	assertNestedSelectionPlayer(t, projected, game.TargetPlayerReference(2))
	second, ok := transformSelectionTargetIndices(selection, rebaseTransform(3, 1))
	if !ok {
		t.Fatal("second independent projection refused")
	}
	assertNestedSelectionPlayer(t, second, game.TargetPlayerReference(3))
	assertNestedSelectionPlayer(t, projected, game.TargetPlayerReference(2))
	assertNestedSelectionPlayer(t, selection, player)
	*projected.AnyOf[0].AnyOf[0].ManaValueDynamic.Val.Group = game.PlayerControlledGroup(
		game.TargetPlayerReference(9), game.Selection{RequiredTypes: []types.Card{types.Creature}})
	assertNestedSelectionPlayer(t, selection, player)
	assertNestedSelectionPlayer(t, second, game.TargetPlayerReference(3))
	if _, ok := transformSelectionTargetIndices(selection, func(targetIndexKind, int) (int, bool) { return 0, false }); ok {
		t.Fatal("missing target map accepted a nested AnyOf group reference")
	}
}

func TestRecursiveSelectionQuantityAndConditionProjection(t *testing.T) {
	t.Parallel()
	player := game.TargetPlayerReference(0)
	selection := nestedQuantitySelection(player)
	for _, amount := range []game.DynamicAmount{
		{Kind: game.DynamicAmountCountCardsInZone, Player: &player, CardZone: zone.Hand, Selection: &selection, Multiplier: 1},
		{Kind: game.DynamicAmountCountSelector, Group: game.BattlefieldGroup(selection), Multiplier: 1},
		{Kind: game.DynamicAmountMaxOf, Operands: []game.DynamicAmount{{
			Kind: game.DynamicAmountCountCardsInZone, Player: &player, CardZone: zone.Hand, Selection: &selection, Multiplier: 1,
		}}},
	} {
		projected, ok := transformQuantity(game.Dynamic(amount), rebaseTransform(2, 1))
		if !ok {
			t.Fatal("valid recursive quantity selection refused")
		}
		got := projected.DynamicAmount().Val
		var rewritten game.Selection
		switch {
		case got.Selection != nil:
			rewritten = *got.Selection
		case len(got.Operands) != 0:
			rewritten = *got.Operands[0].Selection
		default:
			rewritten = got.Group.Selection()
		}
		assertNestedSelectionPlayer(t, rewritten, game.TargetPlayerReference(2))
		*rewritten.AnyOf[0].AnyOf[0].ManaValueDynamic.Val.Group = game.PlayerControlledGroup(
			game.TargetPlayerReference(9), game.Selection{RequiredTypes: []types.Card{types.Creature}})
		assertNestedSelectionPlayer(t, selection, player)
	}
	condition := game.EffectCondition{Condition: opt.Val(game.Condition{
		Object: opt.Val(game.TargetPermanentReference(0)), ObjectMatches: opt.Val(selection),
	})}
	projected, ok := transformEffectConditionTargetIndices(condition, rebaseTransform(2, 1))
	if !ok || projected.Condition.Val.Object.Val != game.TargetPermanentReference(2) {
		t.Fatal("condition lost its actual nested target subject")
	}
	assertNestedSelectionPlayer(t, projected.Condition.Val.ObjectMatches.Val, game.TargetPlayerReference(2))
	assertNestedSelectionPlayer(t, condition.Condition.Val.ObjectMatches.Val, player)
}

func TestRecursiveSelectionProjectionRefusesWholeTargetGroups(t *testing.T) {
	t.Parallel()
	group := game.PlayerGroupControlledGroup(game.TargetedPlayersReference(), game.Selection{})
	selection := game.Selection{AnyOf: []game.Selection{{AnyOf: []game.Selection{{
		ManaValueDynamic: opt.Val(game.ManaValueDynamicBound{Kind: game.DynamicAmountCountSelector, Group: &group}),
	}}}}}
	if problems := selection.Validate(); len(problems) != 0 {
		t.Fatalf("whole-group near miss should be structurally valid: %v", problems)
	}
	if _, ok := transformSelectionTargetIndices(selection, rebaseTransform(2, 1)); ok {
		t.Fatal("nested unscoped whole-target group was accepted")
	}
	if _, ok := transformQuantity(game.Dynamic(game.DynamicAmount{
		Kind: game.DynamicAmountCountSelector, Group: game.BattlefieldGroup(selection),
	}), rebaseTransform(2, 1)); ok {
		t.Fatal("a group selection hid its nested unscoped target group")
	}
}

func TestRecursiveSelectionProjectionPreservesFrozenPlayerDomains(t *testing.T) {
	t.Parallel()
	for _, player := range []game.PlayerReference{
		game.EventPlayerReference(), game.CapturedTargetControllerReference(1),
		game.ObjectControllerReference(game.SourcePermanentReference()),
		game.ObjectControllerReference(game.EventPermanentReference()),
		game.ObjectControllerReference(game.LinkedObjectReference("prior")),
	} {
		selection := nestedQuantitySelection(player)
		projected, ok := transformSelectionTargetIndices(selection, rebaseTransform(2, 1))
		if !ok {
			t.Fatal("frozen player domain refused")
		}
		assertNestedSelectionPlayer(t, projected, player)
		assertNestedSelectionPlayer(t, selection, player)
	}
}

func TestRecursiveSelectionProjectedCountUsesControllerRelativeBound(t *testing.T) {
	t.Parallel()
	for _, name := range []string{"card selection", "group selection", "operand selection", "condition selection"} {
		for _, boundCount := range []int{0, 2} {
			caseName := name + "/zero bound"
			if boundCount == 2 {
				caseName = name + "/two bound"
			}
			t.Run(caseName, func(t *testing.T) {
				t.Parallel()
				player := game.TargetPlayerReference(0)
				selection := nestedGroupSelection(game.BattlefieldGroup(game.Selection{
					RequiredTypes: []types.Card{types.Creature}, Controller: game.ControllerYou,
				}))
				amount := game.DynamicAmount{
					Kind: game.DynamicAmountCountCardsInZone, Player: &player, CardZone: zone.Hand, Selection: &selection, Multiplier: 1,
				}
				if name == "group selection" {
					amount = game.DynamicAmount{Kind: game.DynamicAmountCountSelector,
						Group: game.PlayerControlledGroup(player, selection), Multiplier: 1}
				}
				if name == "operand selection" {
					amount = game.DynamicAmount{Kind: game.DynamicAmountMaxOf, Operands: []game.DynamicAmount{
						amount, {Kind: game.DynamicAmountConstant, Constant: 0},
					}}
				}
				instruction := game.Instruction{Primitive: game.Draw{Player: player, Amount: game.Dynamic(amount)}}
				if name == "condition selection" {
					instruction.Primitive = game.Draw{Player: player, Amount: game.Fixed(1)}
					instruction.Condition = opt.Val(game.EffectCondition{Condition: opt.Val(game.Condition{
						Aggregates: []game.AggregateComparison{{
							Aggregate: game.AggregateControllerHandSize, Op: compare.LessThan, ValueAmount: opt.Val(amount),
						}},
					})})
				}
				original := game.RepeatProcess{Times: game.Fixed(2), Body: game.Mode{Sequence: []game.Instruction{instruction}}.Ability()}
				projected, ok := rebaseTargetedPrimitive(original, 2, 1)
				if !ok {
					t.Fatal("nested selection count could not be projected")
				}
				g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
				g.Turn.ActivePlayer, g.Turn.PriorityPlayer = game.Player4, game.Player4
				g.Turn.Phase, g.Turn.Step = game.PhasePrecombatMain, game.StepNone
				add := func(owner game.PlayerID, def *game.CardDef) id.ID {
					cardID := g.IDGen.Next()
					g.CardInstances[cardID] = &game.CardInstance{ID: cardID, Owner: owner, Def: def}
					return cardID
				}
				creature := func(value int) *game.CardDef {
					return &game.CardDef{CardFace: game.CardFace{Name: "Counted creature",
						Types: []types.Card{types.Creature}, ManaCost: opt.Val(cost.Mana{cost.O(value)}),
						Power: opt.Val(game.PT{Value: 1}), Toughness: opt.Val(game.PT{Value: 1})}}
				}
				for _, value := range []int{1, 3} {
					g.Players[game.Player2].Hand.Add(add(game.Player2, creature(value)))
					cardID := add(game.Player2, creature(value))
					g.Battlefield = append(g.Battlefield, &game.Permanent{
						ObjectID: g.IDGen.Next(), CardInstanceID: cardID, Controller: game.Player2, Owner: game.Player2,
					})
				}
				for range boundCount {
					cardID := add(game.Player4, creature(3))
					g.Battlefield = append(g.Battlefield, &game.Permanent{
						ObjectID: g.IDGen.Next(), CardInstanceID: cardID, Controller: game.Player4, Owner: game.Player4,
					})
				}
				for range 6 {
					g.Players[game.Player2].Library.Add(add(game.Player2, creature(3)))
				}
				prefixCard := add(game.Player3, creature(1))
				g.Players[game.Player3].Graveyard.Add(prefixCard)
				def := &game.CardDef{CardFace: game.CardFace{Name: "Recursive selection count", Types: []types.Card{types.Sorcery},
					SpellAbility: opt.Val(game.Mode{
						Targets: []game.TargetSpec{
							{Allow: game.TargetAllowPlayer, MinTargets: 1, MaxTargets: 1},
							{Allow: game.TargetAllowCard, TargetZone: zone.Graveyard, MinTargets: 1, MaxTargets: 1},
							{Allow: game.TargetAllowPlayer, MinTargets: 1, MaxTargets: 1},
						},
						Sequence: []game.Instruction{{Primitive: projected}},
					}.Ability()),
				}}
				source := add(game.Player4, def)
				g.Stack.Push(&game.StackObject{ID: g.IDGen.Next(), Kind: game.StackSpell,
					SourceID: source, SourceCardID: source, Controller: game.Player4,
					Targets: []game.Target{game.PlayerTarget(game.Player1), game.CardTarget(prefixCard), game.PlayerTarget(game.Player2)},
				})
				var policies [game.NumPlayers]rules.PlayerAgent
				for i := range policies {
					policies[i] = repeatQuantityPassAgent{}
				}
				resolved := rules.NewEngine(nil).Simulator().ResolvePriority(g, policies)
				wantHand, wantLibrary := 2, 6
				if boundCount == 2 {
					wantHand, wantLibrary = 4, 4
				}
				if resolved.Players[game.Player2].Hand.Size() != wantHand || resolved.Players[game.Player2].Library.Size() != wantLibrary ||
					resolved.Players[game.Player1].Hand.Size() != 0 {
					t.Fatalf("nested controller bound=%d, target2 hand/library=%d/%d, want %d/%d",
						boundCount, resolved.Players[game.Player2].Hand.Size(), resolved.Players[game.Player2].Library.Size(), wantHand, wantLibrary)
				}
				projectedRepeat := captureTestPrimitive[game.RepeatProcess](t, projected)
				got := captureTestPrimitive[game.Draw](t, projectedRepeat.Body.Modes[0].Sequence[0].Primitive)
				originalDraw := captureTestPrimitive[game.Draw](t, original.Body.Modes[0].Sequence[0].Primitive)
				if got.Player != game.TargetPlayerReference(2) || originalDraw.Player != player {
					t.Fatal("real count action did not preserve projected/current and original actor slots")
				}
			})
		}
	}
}
