package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

func sakashimaTestReplacement() game.ReplacementAbility {
	return game.EntersAsCopyWithOtherAbilities(
		game.EntersAsCopyWithRetainedName(
			game.EntersAsCopyReplacement(
				"You may have Sakashima enter as a copy of another creature you control, except it has Sakashima's other abilities.",
				&game.Selection{
					RequiredTypes: []types.Card{types.Creature},
					Controller:    game.ControllerYou,
				},
				true, false, nil, false, nil, nil,
			),
		),
	)
}

func sakashimaTestDef() *game.CardDef {
	pt3 := opt.Val(game.PT{Value: 3})
	pt1 := opt.Val(game.PT{Value: 1})
	return &game.CardDef{CardFace: game.CardFace{
		Name:                 "Sakashima of a Thousand Faces",
		ManaCost:             opt.Val(cost.Mana{cost.O(3), cost.U}),
		Supertypes:           []types.Super{types.Legendary},
		Types:                []types.Card{types.Creature},
		Subtypes:             []types.Sub{types.Human, types.Rogue},
		Power:                pt3,
		Toughness:            pt1,
		ReplacementAbilities: []game.ReplacementAbility{sakashimaTestReplacement()},
		StaticAbilities: []game.StaticAbility{
			game.LegendRuleDoesNotApplyStaticBody,
			game.PartnerStaticBody,
		},
	}}
}

func TestSakashimaAcceptsOnlyControlledCreatureAndKeepsExceptions(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	opponentPT := opt.Val(game.PT{Value: 9})
	addCombatPermanent(g, game.Player2, &game.CardDef{CardFace: game.CardFace{
		Name:      "Opponent Giant",
		Types:     []types.Card{types.Creature},
		Power:     opponentPT,
		Toughness: opponentPT,
	}})
	targetPT := opt.Val(game.PT{Value: 4})
	addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:      "Controlled Dragon",
		Types:     []types.Card{types.Creature},
		Subtypes:  []types.Sub{types.Dragon},
		Power:     targetPT,
		Toughness: targetPT,
		StaticAbilities: []game.StaticAbility{{
			Text:             "Flying",
			KeywordAbilities: game.SimpleKeywords(game.Flying),
		}},
	}})
	sakashima := addCombatPermanent(g, game.Player1, sakashimaTestDef())
	agent := optionalMayAgent{accept: true}
	ctx := enterBattlefieldContext{
		engine: NewEngine(nil),
		agents: [game.NumPlayers]PlayerAgent{game.Player1: agent},
		log:    &TurnLog{},
	}

	replacement := sakashimaTestReplacement()
	applyEntersAsCopy(ctx, g, sakashima, &replacement.Replacement)

	if got := permanentEffectiveName(g, sakashima); got != "Sakashima of a Thousand Faces" {
		t.Fatalf("effective name = %q, want retained Sakashima name", got)
	}
	if got := effectivePower(g, sakashima); got != 4 {
		t.Fatalf("effective power = %d, want controlled creature's 4 (opponent's 9 is illegal)", got)
	}
	if !hasKeyword(g, sakashima, game.Flying) {
		t.Fatal("copy did not retain the chosen creature's Flying ability")
	}
	if !hasKeyword(g, sakashima, game.Partner) {
		t.Fatal("copy did not add Sakashima's printed Partner ability")
	}
	values := effectivePermanentValues(g, sakashima)
	for _, ability := range values.abilities {
		if copyAbility, ok := ability.(*game.ReplacementAbility); ok && copyAbility.Replacement.EntersAsCopy {
			t.Fatal("other-abilities exception recursively retained the copy replacement")
		}
	}
}

func TestSakashimaDeclinesCopyAndKeepsPrintedCharacteristics(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	pt := opt.Val(game.PT{Value: 4})
	addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name: "Controlled Dragon", Types: []types.Card{types.Creature}, Power: pt, Toughness: pt,
	}})
	sakashima := addCombatPermanent(g, game.Player1, sakashimaTestDef())
	agent := optionalMayAgent{accept: false}
	ctx := enterBattlefieldContext{
		engine: NewEngine(nil),
		agents: [game.NumPlayers]PlayerAgent{game.Player1: agent},
		log:    &TurnLog{},
	}

	replacement := sakashimaTestReplacement()
	applyEntersAsCopy(ctx, g, sakashima, &replacement.Replacement)

	if got := permanentEffectiveName(g, sakashima); got != "Sakashima of a Thousand Faces" {
		t.Fatalf("effective name = %q after decline", got)
	}
	if got := effectivePower(g, sakashima); got != 3 {
		t.Fatalf("effective power = %d after decline, want printed 3", got)
	}
	if !hasKeyword(g, sakashima, game.Partner) {
		t.Fatal("declined Sakashima lost its printed Partner ability")
	}
}

func TestSakashimaCopySnapshotPreservesCopyLayerForTokensAndCopies(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	targetPT := opt.Val(game.PT{Value: 6})
	addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name: "Controlled Wurm", ManaCost: opt.Val(cost.Mana{cost.O(5), cost.G}),
		Types: []types.Card{types.Creature}, Power: targetPT, Toughness: targetPT,
	}})
	sakashima := addCombatPermanent(g, game.Player1, sakashimaTestDef())
	replacement := sakashimaTestReplacement()
	applyEntersAsCopy(enterBattlefieldContext{}, g, sakashima, &replacement.Replacement)

	snapshot, ok := permanentCopyDef(g, sakashima)
	if !ok {
		t.Fatal("permanentCopyDef failed for copied Sakashima")
	}
	if snapshot.Name != "Sakashima of a Thousand Faces" || !snapshot.Power.Exists || snapshot.Power.Val.Value != 6 {
		t.Fatalf("copy snapshot = name %q power %#v, want retained name and copied 6", snapshot.Name, snapshot.Power)
	}
	if got := snapshot.ManaValue(); got != 6 {
		t.Fatalf("copy snapshot mana value = %d, want copied creature's 6", got)
	}
	if len(snapshot.ReplacementAbilities) != 0 {
		t.Fatalf("copy snapshot retained %d enters-as-copy abilities, want none", len(snapshot.ReplacementAbilities))
	}
	token := &game.Permanent{
		ObjectID:   g.IDGen.Next(),
		Owner:      game.Player1,
		Controller: game.Player1,
		Token:      true,
		TokenDef:   snapshot,
	}
	g.Battlefield = append(g.Battlefield, token)
	if !hasKeyword(g, token, game.Partner) {
		t.Fatal("token copy lost Sakashima's copiable Partner ability")
	}
	if !playerRuleEffectActive(g, game.Player1, game.RuleEffectLegendRuleDoesNotApply) {
		t.Fatal("token copy lost Sakashima's copiable legend-rule exemption")
	}
}

func legendRuleExemptionSourceDef(name string) *game.CardDef {
	return &game.CardDef{CardFace: game.CardFace{
		Name:            name,
		Types:           []types.Card{types.Artifact},
		StaticAbilities: []game.StaticAbility{game.LegendRuleDoesNotApplyStaticBody},
	}}
}

func legendaryTestDef(name string) *game.CardDef {
	return &game.CardDef{CardFace: game.CardFace{
		Name:       name,
		Supertypes: []types.Super{types.Legendary},
		Types:      []types.Card{types.Creature},
		Power:      opt.Val(game.PT{Value: 2}),
		Toughness:  opt.Val(game.PT{Value: 2}),
	}}
}

func TestLegendRuleExemptionAllowsControllerDuplicatesButNotOpponents(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	addPermanentForSBA(g, game.Player1, legendRuleExemptionSourceDef("Mirror One"))
	firstYours := addPermanentForSBA(g, game.Player1, legendaryTestDef("Shared Legend"))
	secondYours := addPermanentForSBA(g, game.Player1, legendaryTestDef("Shared Legend"))
	firstOpposing := addPermanentForSBA(g, game.Player2, legendaryTestDef("Opposing Legend"))
	secondOpposing := addPermanentForSBA(g, game.Player2, legendaryTestDef("Opposing Legend"))

	changed, deaths := checkLegendaryRuleStateBasedActions(g, newPassBatchID(g))

	if !changed || len(deaths) != 1 || deaths[0].Permanent != secondOpposing.ObjectID {
		t.Fatalf("deaths = %#v, want only opponent's newer duplicate", deaths)
	}
	if _, ok := permanentByObjectID(g, firstYours.ObjectID); !ok {
		t.Fatal("controller's first duplicate left despite legend-rule exemption")
	}
	if _, ok := permanentByObjectID(g, secondYours.ObjectID); !ok {
		t.Fatal("controller's second duplicate left despite legend-rule exemption")
	}
	if _, ok := permanentByObjectID(g, firstOpposing.ObjectID); !ok {
		t.Fatal("opponent's oldest legend should remain")
	}
}

func TestLegendRuleExemptionFollowsSourceController(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	source := addPermanentForSBA(g, game.Player1, legendRuleExemptionSourceDef("Control Mirror"))
	firstYours := addPermanentForSBA(g, game.Player1, legendaryTestDef("Your Legend"))
	secondYours := addPermanentForSBA(g, game.Player1, legendaryTestDef("Your Legend"))
	firstTheirs := addPermanentForSBA(g, game.Player2, legendaryTestDef("Their Legend"))
	secondTheirs := addPermanentForSBA(g, game.Player2, legendaryTestDef("Their Legend"))
	g.ContinuousEffects = append(g.ContinuousEffects, game.ContinuousEffect{
		ID:               g.IDGen.Next(),
		AffectedObjectID: source.ObjectID,
		Layer:            game.LayerControl,
		Duration:         game.DurationPermanent,
		NewController:    opt.Val(game.Player2),
	})

	_, deaths := checkLegendaryRuleStateBasedActions(g, newPassBatchID(g))

	if len(deaths) != 1 || deaths[0].Permanent != secondYours.ObjectID {
		t.Fatalf("deaths = %#v, want Player1's newer duplicate after control change", deaths)
	}
	if _, ok := permanentByObjectID(g, firstYours.ObjectID); !ok {
		t.Fatal("Player1's oldest legend should remain")
	}
	if _, ok := permanentByObjectID(g, firstTheirs.ObjectID); !ok {
		t.Fatal("Player2's first duplicate left despite gaining the exemption")
	}
	if _, ok := permanentByObjectID(g, secondTheirs.ObjectID); !ok {
		t.Fatal("Player2's second duplicate left despite gaining the exemption")
	}
}

func TestLegendRuleExemptionStacksAndEndsWhenLastSourceLeaves(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	firstSource := addPermanentForSBA(g, game.Player1, legendRuleExemptionSourceDef("Mirror One"))
	secondSource := addPermanentForSBA(g, game.Player1, legendRuleExemptionSourceDef("Mirror Two"))
	addPermanentForSBA(g, game.Player1, legendaryTestDef("Persistent Legend"))
	newer := addPermanentForSBA(g, game.Player1, legendaryTestDef("Persistent Legend"))

	if changed, deaths := checkLegendaryRuleStateBasedActions(g, newPassBatchID(g)); changed || len(deaths) != 0 {
		t.Fatalf("two active sources produced deaths = %#v", deaths)
	}
	if !movePermanentToZone(g, firstSource, zone.Graveyard) {
		t.Fatal("failed to remove first exemption source")
	}
	if changed, deaths := checkLegendaryRuleStateBasedActions(g, newPassBatchID(g)); changed || len(deaths) != 0 {
		t.Fatalf("one remaining source produced deaths = %#v", deaths)
	}
	if !movePermanentToZone(g, secondSource, zone.Graveyard) {
		t.Fatal("failed to remove last exemption source")
	}
	changed, deaths := checkLegendaryRuleStateBasedActions(g, newPassBatchID(g))
	if !changed || len(deaths) != 1 || deaths[0].Permanent != newer.ObjectID {
		t.Fatalf("deaths after last source left = %#v, want newer duplicate", deaths)
	}
}
