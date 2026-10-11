package rules

import (
	"slices"
	"testing"

	"github.com/natefinch/council4/cardgen"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

func compiledDamageAlternative(t *testing.T) *game.CardDef {
	t.Helper()
	return compileUnlessCard(t, cardgen.ScryfallCard{
		Name: "Damage Alternative", Layout: "normal", TypeLine: "Sorcery", ManaCost: "{2}{R}",
		OracleText: "Kicker {5}{R} (You may pay an additional {5}{R} as you cast this spell.)\n" +
			"Damage Alternative deals 5 damage to target creature.\nIf this spell was kicked, Damage Alternative deals 10 damage instead.",
	})
}

func TestCompiledDamageTargetSelection(t *testing.T) {
	for _, tc := range []struct {
		name, text string
		kicked     bool
		creature   bool
	}{
		{"Lightning Bolt", "Lightning Bolt deals 3 damage to any target.", false, false},
		{"Arc Lightning", "Arc Lightning deals 3 damage divided as you choose among one, two, or three targets.", false, false},
		{"Comet Storm", "Multikicker {1}\nChoose any target, then choose another target for each time this spell was kicked. Comet Storm deals X damage to each of them.", true, false},
		{"Creature Damage", "Creature Damage deals 3 damage to target creature.", false, true},
		{"Damage Alternative", "", false, true},
		{"Damage Alternative", "", true, true},
	} {
		t.Run(tc.name+map[bool]string{false: "/base", true: "/kicked"}[tc.kicked], func(t *testing.T) {
			var def *game.CardDef
			if tc.text == "" {
				def = compiledDamageAlternative(t)
			} else {
				def = compileUnlessCard(t, cardgen.ScryfallCard{
					Name: tc.name, Layout: "normal", TypeLine: "Instant", OracleText: tc.text,
				})
			}
			g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
			cases := []struct {
				name string
				kind types.Card
				want bool
			}{
				{"creature", types.Creature, true},
				{"planeswalker", types.Planeswalker, !tc.creature},
				{"battle", types.Battle, !tc.creature},
				{"land", types.Land, false},
				{"artifact-only", types.Artifact, false},
			}
			targets := make([]game.Target, len(cases))
			for i, candidate := range cases {
				permanent := addCombatPermanent(g, game.Player3, &game.CardDef{
					CardFace: game.CardFace{Name: candidate.name, Types: []types.Card{candidate.kind}},
				})
				targets[i] = game.PermanentTarget(permanent.ObjectID)
			}
			result := targetChoicesForSpell(g, game.Player2, def, nil, game.CastBranch{Kicked: tc.kicked})
			if result.kind != targetLegalChoicesFound {
				t.Fatalf("no legal choices: kind=%v err=%v", result.kind, result.err)
			}
			offered := func(target game.Target) bool {
				return slices.ContainsFunc(result.choices, func(choice []game.Target) bool {
					return slices.Contains(choice, target)
				})
			}
			for i, candidate := range cases {
				if got := offered(targets[i]); got != candidate.want {
					t.Errorf("offers %s=%v want=%v", candidate.name, got, candidate.want)
				}
			}
			if got := offered(game.PlayerTarget(game.Player3)); got == tc.creature {
				t.Errorf("offers player=%v want=%v", got, !tc.creature)
			}
		})
	}
}

func TestRepresentedDamageAlternativeAllocationAndTargetCounts(t *testing.T) {
	def := &game.CardDef{CardFace: game.CardFace{
		Name: "Represented divided branch", Types: []types.Card{types.Sorcery},
		SpellAbility: opt.Val(game.Mode{
			Targets: []game.TargetSpec{
				{MinTargets: 1, MaxTargets: 1, Allow: game.TargetAllowPermanent,
					Selection: opt.Val(game.Selection{RequiredTypesAny: []types.Card{types.Creature}}),
					Gate:      game.TargetGateSpellNotKicked},
				{MinTargets: 0, MaxTargets: 10, Allow: game.TargetAllowPermanent | game.TargetAllowPlayer,
					Selection: opt.Val(game.Selection{RequiredTypesAny: []types.Card{types.Creature, types.Planeswalker, types.Battle}}),
					Gate:      game.TargetGateSpellKicked},
			},
			Sequence: []game.Instruction{
				{Primitive: game.Damage{Amount: game.Fixed(5), Recipient: game.AnyTargetDamageRecipient(0)},
					Condition: opt.Val(game.EffectCondition{Condition: opt.Val(game.Condition{
						SpellWasKicked: true, Negate: true,
					})})},
				{Primitive: game.Damage{Amount: game.Fixed(10), Recipient: game.AnyTargetDamageRecipient(1), Divided: true},
					Condition: opt.Val(game.EffectCondition{Condition: opt.Val(game.Condition{SpellWasKicked: true})})},
			},
		}.Ability()),
	}}
	for _, tc := range []struct {
		name    string
		kicked  bool
		illegal bool
		want    int
	}{
		{"base", false, false, 5},
		{"kicked", true, false, 4},
		{"kicked illegal share", true, true, 4},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
			first := addCombatPermanent(g, game.Player3, vanillaCreature("First", 20, 20))
			second := addCombatPermanent(g, game.Player3, vanillaCreature("Second", 20, 20))
			source := addCardToHand(g, game.Player4, def)
			obj := &game.StackObject{
				ID: g.IDGen.Next(), Kind: game.StackSpell, SourceID: source,
				SourceCardID: source, Controller: game.Player2, KickerPaid: tc.kicked,
				Targets: []game.Target{game.PermanentTarget(first.ObjectID)}, TargetCounts: []int{1, 0},
			}

			if tc.kicked {
				obj.Targets = append(obj.Targets, game.PermanentTarget(second.ObjectID))
				obj.TargetCounts = []int{0, 2}
			}
			g.Stack.Push(obj)
			if tc.illegal {
				if _, ok := removePermanentFromBattlefield(g, second.ObjectID); !ok {
					t.Fatal("failed to invalidate target")
				}
			}
			agents := [game.NumPlayers]PlayerAgent{}
			agents[game.Player2] = &choiceOnlyAgent{choices: [][]int{{0, 0, 0, 0, 1, 1, 1, 1, 1, 1}}}
			log := TurnLog{}
			NewEngine(nil).resolveTopOfStackWithChoices(g, agents, &log)
			if first.MarkedDamage != tc.want {
				t.Fatalf("first damage=%d want=%d; branch or ordinal selection changed", first.MarkedDamage, tc.want)
			}
			if tc.kicked && !tc.illegal && second.MarkedDamage != 6 {
				t.Fatalf("second damage=%d want=6", second.MarkedDamage)
			}
			if tc.kicked && (len(log.Choices) != 1 || log.Choices[0].Request.Kind != game.ChoiceDamageAllocation) {
				t.Fatalf("allocation choices=%+v", log.Choices)
			}
			if !tc.kicked && len(log.Choices) != 0 {
				t.Fatalf("inactive divided branch offered allocation: %+v", log.Choices)
			}
		})
	}
}

func TestDividedTargetsRetainOnlyRepresentedDeferredRecipients(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	obj := &game.StackObject{
		Controller:   game.Player2,
		TargetCounts: []int{1, 6},
		Targets: []game.Target{
			game.PlayerTarget(game.Player1),
			game.PlayerTarget(game.Player3),
			game.DeferredTargetFrom(game.PermanentTarget(123)),
			game.DeferredTargetFrom(game.PlayerTarget(game.Player4)),
			game.DeferredTarget(),
			game.DeferredTargetFrom(game.CardTarget(456)),
			game.DeferredTargetFrom(game.StackObjectTarget(789)),
		},
	}
	resolver := newEffectResolver(NewEngine(nil), g, obj, [game.NumPlayers]PlayerAgent{}, &TurnLog{})
	targets := resolver.dividedTargets(1)
	if len(targets) != 3 || !targets[0].legal || targets[1].legal || targets[2].legal {
		t.Fatalf("represented recipient slots changed: %+v", targets)
	}
	if targets[1].target.Kind != game.TargetDeferred || targets[2].target.Kind != game.TargetDeferred {
		t.Fatal("deferred target identity was reconstructed")
	}
}

func TestDividedDamageNaturalRescreenRetainsCopiedIllegalSlot(t *testing.T) {
	def := &game.CardDef{CardFace: game.CardFace{
		Name: "Represented single division", Types: []types.Card{types.Instant},
		SpellAbility: opt.Val(game.Mode{
			Targets: []game.TargetSpec{{
				MinTargets: 1, MaxTargets: 10, Allow: game.TargetAllowPermanent | game.TargetAllowPlayer,
				Selection: opt.Val(game.Selection{RequiredTypesAny: []types.Card{types.Creature, types.Planeswalker, types.Battle}}),
			}},
			Sequence: []game.Instruction{{Primitive: dividedDamage(10)}},
		}.Ability()),
	}}
	for _, copied := range []bool{false, true} {
		g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
		first := addCombatPermanent(g, game.Player3, vanillaCreature("First", 20, 20))
		second := addCombatPermanent(g, game.Player3, vanillaCreature("Second", 20, 20))
		source := addCardToHand(g, game.Player4, def)
		obj := &game.StackObject{
			ID: g.IDGen.Next(), Kind: game.StackSpell, SourceID: source,
			SourceCardID: source, Controller: game.Player2, Copy: copied,
			Targets:      []game.Target{game.PermanentTarget(first.ObjectID), game.PermanentTarget(second.ObjectID)},
			TargetCounts: []int{2},
		}
		g.Stack.Push(obj)
		if _, ok := removePermanentFromBattlefield(g, second.ObjectID); !ok {
			t.Fatal("failed to invalidate target")
		}
		agents := [game.NumPlayers]PlayerAgent{}
		agents[game.Player2] = &choiceOnlyAgent{choices: [][]int{{0, 0, 0, 0, 1, 1, 1, 1, 1, 1}}}
		NewEngine(nil).resolveTopOfStackWithChoices(g, agents, &TurnLog{})
		if first.MarkedDamage != 4 {
			t.Fatalf("copied=%v surviving damage=%d want=4", copied, first.MarkedDamage)
		}
	}
}

func TestCompiledDamageAlternativeKickerAffordability(t *testing.T) {
	def := compiledDamageAlternative(t)
	for _, affordable := range []bool{false, true} {
		g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
		g.Turn.ActivePlayer = game.Player2
		g.Turn.Phase = game.PhasePrecombatMain
		g.Turn.Step = game.StepNone
		g.Turn.PriorityPlayer = game.Player2
		g.Players[game.Player2].ManaPool.Add(mana.R, 1)
		g.Players[game.Player2].ManaPool.Add(mana.C, 2)
		if affordable {
			g.Players[game.Player2].ManaPool.Add(mana.R, 1)
			g.Players[game.Player2].ManaPool.Add(mana.C, 5)
		}
		source := addCardToHand(g, game.Player2, def)
		addCombatPermanent(g, game.Player3, vanillaCreature("Recipient", 20, 20))
		casts := castActionsForCard(t, NewEngine(nil), g, game.Player2, source)
		base, kicked := false, false
		for _, cast := range casts {
			if cast.KickerPaid {
				kicked = true
			} else {
				base = true
			}
		}
		if !base || kicked != affordable {
			t.Fatalf("affordable=%v base=%v kicked=%v", affordable, base, kicked)
		}
	}
}
