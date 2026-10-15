package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

func TestContinuousDynamicDependencyGrantChainCombat(t *testing.T) {
	for _, reversed := range []bool{false, true} {
		name := "first strike source first"
		if reversed {
			name = "double strike source first"
		}
		t.Run(name, func(t *testing.T) {
			g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
			attacker := addCombatCreaturePermanentWithPower(g, game.Player1, 2)
			first := keywordDependencySource("First strike grant", game.KeywordNone, game.FirstStrike)
			second := keywordDependencySource("Conditional double strike grant", game.FirstStrike, game.DoubleStrike)
			if reversed {
				first, second = second, first
			}
			addCombatPermanent(g, game.Player1, first)
			addCombatPermanent(g, game.Player1, second)
			if !hasKeyword(g, attacker, game.DoubleStrike) {
				t.Error("dependent grant was filtered out before the ability layer")
			}
			g.Turn.ActivePlayer = game.Player1
			g.Turn.Phase = game.PhaseCombat
			g.Turn.Step = game.StepDeclareAttackers
			g.Combat = &game.CombatState{}
			policies := simPassPolicies()
			declare := action.DeclareAttackersAction{Attackers: []game.AttackDeclaration{
				{Attacker: attacker.ObjectID, Target: game.AttackTarget{Player: game.Player2}},
			}}
			resolved, ok := NewEngine(nil).Simulator().ResolveCombatWithAttackers(g, game.Player1, declare, policies)
			if !ok || resolved.Players[game.Player2].Life != 36 {
				t.Fatalf("Simulator combat = %v/%v, want legal attack dealing 4", ok, resolved.Players[game.Player2].Life)
			}
			policies[game.Player1] = simultaneousCombatPolicy{attacker: attacker.ObjectID, defender: game.Player2}
			live := g.Clone()
			NewEngine(nil).runCombatPhase(live, policies, &TurnLog{})
			if live.Players[game.Player2].Life != 36 || g.Players[game.Player2].Life != 40 {
				t.Fatal("live combat or original-state isolation disagreed")
			}
		})
	}
}

func TestContinuousDynamicDependencyRemovalInvalidatesCandidate(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	creature := addCombatCreaturePermanentWithPower(g, game.Player1, 2, game.FirstStrike)
	g.ContinuousEffects = append(g.ContinuousEffects,
		game.ContinuousEffect{Layer: game.LayerAbility, Timestamp: 1,
			Group: game.BattlefieldGroup(game.Selection{Keyword: game.FirstStrike}), AddKeywords: []game.Keyword{game.DoubleStrike}},
		game.ContinuousEffect{Layer: game.LayerAbility, Timestamp: 2, AffectedObjectID: creature.ObjectID,
			RemoveKeywords: []game.Keyword{game.FirstStrike}},
	)
	if hasKeyword(g, creature, game.FirstStrike) || hasKeyword(g, creature, game.DoubleStrike) {
		t.Fatal("dependency removal retained an effect that no longer applies")
	}
}

func TestContinuousDynamicDependencyCycleUsesTimestamp(t *testing.T) {
	for _, firstStrikeRemovedFirst := range []bool{false, true} {
		g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
		creature := addCombatCreaturePermanentWithPower(g, game.Player1, 2, game.FirstStrike, game.DoubleStrike)
		first := game.ContinuousEffect{Layer: game.LayerAbility, Timestamp: 1,
			Group: game.BattlefieldGroup(game.Selection{Keyword: game.FirstStrike}), RemoveKeywords: []game.Keyword{game.DoubleStrike}}
		second := game.ContinuousEffect{Layer: game.LayerAbility, Timestamp: 2,
			Group: game.BattlefieldGroup(game.Selection{Keyword: game.DoubleStrike}), RemoveKeywords: []game.Keyword{game.FirstStrike}}
		if firstStrikeRemovedFirst {
			first.Timestamp, second.Timestamp = second.Timestamp, first.Timestamp
		}
		g.ContinuousEffects = append(g.ContinuousEffects, first, second)
		if hasKeyword(g, creature, game.FirstStrike) == firstStrikeRemovedFirst ||
			hasKeyword(g, creature, game.DoubleStrike) != firstStrikeRemovedFirst {
			t.Fatal("dependency cycle failed timestamp fallback or applicability reevaluation")
		}
	}
}

func TestContinuousDynamicDependencyFramesAndClonesInvalidate(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	creature := addCombatCreaturePermanentWithPower(g, game.Player1, 2)
	first := addCombatPermanent(g, game.Player1, keywordDependencySource("First grant", game.KeywordNone, game.FirstStrike))
	addCombatPermanent(g, game.Player1, keywordDependencySource("Second grant", game.FirstStrike, game.DoubleStrike))
	g.BeginStaticSourceFrame()
	for range 3 {
		if !hasKeyword(g, creature, game.DoubleStrike) {
			t.Fatal("repeated framed query changed dependency results")
		}
	}
	g.EndStaticSourceFrame()
	branch := g.Clone()
	source, _ := permanentByObjectID(branch, first.ObjectID)
	if !movePermanentToZone(branch, source, zone.Graveyard) {
		t.Fatal("could not remove dependency source in clone")
	}
	branchedCreature, _ := permanentByObjectID(branch, creature.ObjectID)
	if hasKeyword(branch, branchedCreature, game.DoubleStrike) || !hasKeyword(g, creature, game.DoubleStrike) {
		t.Fatal("source mutation did not invalidate only its branch")
	}
}

func TestContinuousDynamicDependencyColorCandidates(t *testing.T) {
	for _, reversed := range []bool{false, true} {
		g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
		creature := addCombatCreaturePermanentWithPower(g, game.Player1, 2)
		first := &game.CardDef{CardFace: game.CardFace{
			Name: "Red grant", Types: []types.Card{types.Enchantment},
			StaticAbilities: []game.StaticAbility{{ContinuousEffects: []game.ContinuousEffect{{
				Layer: game.LayerColor, SetColors: []color.Color{color.Red},
				Group: game.BattlefieldGroup(game.Selection{RequiredTypes: []types.Card{types.Creature}, Controller: game.ControllerYou}),
			}}}},
		}}
		second := &game.CardDef{CardFace: game.CardFace{
			Name: "Blue grant", Types: []types.Card{types.Enchantment},
			StaticAbilities: []game.StaticAbility{{ContinuousEffects: []game.ContinuousEffect{{
				Layer: game.LayerColor, SetColors: []color.Color{color.Blue},
				Group: game.BattlefieldGroup(game.Selection{RequiredTypes: []types.Card{types.Creature},
					Controller: game.ControllerYou, ColorsAny: []color.Color{color.Red}}),
			}}}},
		}}
		if reversed {
			first, second = second, first
		}
		addCombatPermanent(g, game.Player1, first)
		addCombatPermanent(g, game.Player1, second)
		colors := permanentEffectiveColors(g, creature)
		if len(colors) != 1 || colors[0] != color.Blue {
			t.Fatalf("dependent color = %v, want blue in source order reversed=%v", colors, reversed)
		}
	}
}

func TestContinuousDynamicDependencyConditionsReadCurrentBoard(t *testing.T) {
	for _, negate := range []bool{false, true} {
		for _, reversed := range []bool{false, true} {
			g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
			creature := addCombatCreaturePermanentWithPower(g, game.Player1, 2)
			first := keywordDependencySource("First strike grant", game.KeywordNone, game.FirstStrike)
			second := keywordDependencySource("Conditioned grant", game.KeywordNone, game.DoubleStrike)
			second.StaticAbilities[0].Condition = opt.Val(game.Condition{
				ControlsMatching: opt.Val(game.SelectionCount{Selection: game.Selection{
					RequiredTypes: []types.Card{types.Creature}, Keyword: game.FirstStrike,
				}}),
				Negate: negate,
			})
			if reversed {
				first, second = second, first
			}
			addCombatPermanent(g, game.Player1, first)
			addCombatPermanent(g, game.Player1, second)
			if hasKeyword(g, creature, game.DoubleStrike) == negate {
				t.Fatalf("evolving condition wrong for negate=%v, reversed=%v", negate, reversed)
			}
		}
	}
}

func TestContinuousDynamicDependencyOrdinaryTypeCandidates(t *testing.T) {
	for _, reversed := range []bool{false, true} {
		g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
		creature := addCombatCreaturePermanentWithPower(g, game.Player1, 2)
		first := game.ContinuousEffect{
			Layer: game.LayerType, Timestamp: 1, AffectedObjectID: creature.ObjectID,
			AddTypes: []types.Card{types.Artifact},
		}
		second := game.ContinuousEffect{
			Layer: game.LayerType, Timestamp: 2,
			Group: game.BattlefieldGroup(game.Selection{AnyOf: []game.Selection{
				{RequiredTypes: []types.Card{types.Artifact}},
			}}),
			AddTypes: []types.Card{types.Enchantment},
		}
		if reversed {
			first.Timestamp, second.Timestamp = second.Timestamp, first.Timestamp
		}
		g.ContinuousEffects = append(g.ContinuousEffects, first, second)
		if !permanentHasType(g, creature, types.Enchantment) {
			t.Fatalf("ordinary type dependency lost its candidate, reversed=%v", reversed)
		}
	}
}

func TestContinuousDynamicDependencyColorBegunRecipients(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	red := addCombatCreaturePermanentWithPower(g, game.Player1, 2)
	other := addCombatCreaturePermanentWithPower(g, game.Player1, 2)
	g.ContinuousEffects = append(g.ContinuousEffects, game.ContinuousEffect{
		Layer: game.LayerColor, Timestamp: 1, AffectedObjectID: red.ObjectID, SetColors: []color.Color{color.Red},
	})
	group := game.BattlefieldGroup(game.Selection{RequiredTypes: []types.Card{types.Creature}, ColorsAny: []color.Color{color.Red}})
	addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name: "Blue flying transformation", Types: []types.Card{types.Enchantment},
		StaticAbilities: []game.StaticAbility{{ContinuousEffects: []game.ContinuousEffect{
			{Layer: game.LayerColor, Group: group, SetColors: []color.Color{color.Blue}},
			{Layer: game.LayerAbility, Group: group, AddKeywords: []game.Keyword{game.Flying}},
		}}},
	}})
	if !hasKeyword(g, red, game.Flying) || hasKeyword(g, other, game.Flying) {
		t.Fatal("color-layer effect failed to retain exactly its begun recipients")
	}
	if colors := permanentEffectiveColors(g, red); len(colors) != 1 || colors[0] != color.Blue {
		t.Fatalf("begun recipient colors = %v, want blue", colors)
	}
}

func TestContinuousDynamicDependencyChosenColorCandidates(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	creature := addCombatCreaturePermanentWithPower(g, game.Player1, 2)
	source := addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name: "Chosen color becomes blue", Types: []types.Card{types.Enchantment},
		StaticAbilities: []game.StaticAbility{{ContinuousEffects: []game.ContinuousEffect{{
			Layer: game.LayerColor, SetColors: []color.Color{color.Blue},
			Group: game.BattlefieldGroup(game.Selection{
				RequiredTypes: []types.Card{types.Creature}, ColorChoice: game.ColorChoiceSourceEntry,
			}),
		}}}},
	}})
	source.EntryChoices = map[game.ChoiceKey]game.ResolutionChoiceResult{
		game.EntryColorChoiceKey: {Kind: game.ResolutionChoiceMana, Color: "R"},
	}
	g.ContinuousEffects = append(g.ContinuousEffects, game.ContinuousEffect{
		Layer: game.LayerColor, Timestamp: game.Timestamp(g.IDGen.Next()), AffectedObjectID: creature.ObjectID,
		SetColors: []color.Color{color.Red},
	})
	if colors := permanentEffectiveColors(g, creature); len(colors) != 1 || colors[0] != color.Blue {
		t.Fatalf("chosen-color dependent grant = %v, want blue", colors)
	}
}

func TestContinuousDynamicDependencyIndependentGrantsKeepTimestamp(t *testing.T) {
	for _, removeFirst := range []bool{false, true} {
		g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
		creature := addCombatCreaturePermanentWithPower(g, game.Player1, 2)
		add := game.ContinuousEffect{
			Layer: game.LayerAbility, Timestamp: 1, AffectedObjectID: creature.ObjectID,
			AddKeywords: []game.Keyword{game.Flying},
		}
		remove := game.ContinuousEffect{
			Layer: game.LayerAbility, Timestamp: 2, AffectedObjectID: creature.ObjectID,
			RemoveKeywords: []game.Keyword{game.Flying},
		}
		if removeFirst {
			add.Timestamp, remove.Timestamp = remove.Timestamp, add.Timestamp
		}
		g.ContinuousEffects = append(g.ContinuousEffects, add, remove)
		if hasKeyword(g, creature, game.Flying) != removeFirst {
			t.Fatal("independent keyword effects stopped following timestamp order")
		}
	}
}

func keywordDependencySource(name string, required, grant game.Keyword) *game.CardDef {
	return &game.CardDef{CardFace: game.CardFace{
		Name: name, Types: []types.Card{types.Enchantment},
		StaticAbilities: []game.StaticAbility{{ContinuousEffects: []game.ContinuousEffect{{
			Layer: game.LayerAbility,
			Group: game.BattlefieldGroup(game.Selection{
				RequiredTypes: []types.Card{types.Creature}, Controller: game.ControllerYou, Keyword: required,
			}),
			AddKeywords: []game.Keyword{grant},
		}}}},
	}}
}
