package rules

import (
	"slices"
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

type psychicPaperChoiceAgent struct {
	labels   []string
	requests int
}

func (*psychicPaperChoiceAgent) ChooseAction(PlayerObservation, []action.Action) action.Action {
	return action.Action{}
}

func (a *psychicPaperChoiceAgent) ChooseChoice(_ PlayerObservation, request game.ChoiceRequest) []int {
	a.requests++
	if len(a.labels) == 0 {
		return nil
	}
	label := a.labels[0]
	a.labels = a.labels[1:]
	for _, option := range request.Options {
		if option.Label == label {
			return []int{option.Index}
		}
	}
	return nil
}

func addPsychicPaper(g *game.Game, controller game.PlayerID) *game.Permanent {
	ward := game.WardStaticAbility(cost.Mana{cost.O(1)})
	return addCombatPermanent(g, controller, &game.CardDef{CardFace: game.CardFace{
		Name:     "Psychic Paper",
		Types:    []types.Card{types.Artifact},
		Subtypes: []types.Sub{types.Equipment},
		StaticAbilities: []game.StaticAbility{{
			ContinuousEffects: []game.ContinuousEffect{
				{
					Layer:        game.LayerAbility,
					Group:        game.AttachedObjectGroup(game.SourcePermanentReference()),
					AddAbilities: []game.Ability{&ward},
				},
				{
					Layer:                   game.LayerText,
					Group:                   game.AttachedObjectGroup(game.SourcePermanentReference()),
					SetNameFromSourceChoice: game.AttachmentCardNameChoiceKey,
				},
				{
					Layer:                      game.LayerType,
					Group:                      game.AttachedObjectGroup(game.SourcePermanentReference()),
					SetSubtypeFromSourceChoice: game.AttachmentSubtypeChoiceKey,
					SetSubtypeChoiceType:       types.Creature,
				},
			},
			RuleEffects: []game.RuleEffect{{
				Kind:             game.RuleEffectCantBeBlocked,
				AffectedAttached: true,
			}},
		}},
		ActivatedAbilities: []game.ActivatedAbility{
			game.EquipActivatedAbility(cost.Mana{cost.O(2)}),
		},
		ReplacementAbilities: []game.ReplacementAbility{
			game.AttachmentChoicesReplacement("", types.Creature, types.Creature),
		},
	}})
}

func TestPsychicPaperAttachmentChoicesPersistAndFollowAttachment(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	g.CardNameCatalog = map[types.Card][]string{
		types.Creature: {"Catalog Creature", "Other Catalog Creature"},
	}
	first := addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:      "Alpha Creature",
		Types:     []types.Card{types.Creature, types.Land},
		Subtypes:  []types.Sub{types.Golem, types.Forest},
		Power:     opt.Val(game.PT{Value: 2}),
		Toughness: opt.Val(game.PT{Value: 2}),
	}})
	second := addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:      "Beta Creature",
		Types:     []types.Card{types.Creature},
		Subtypes:  []types.Sub{types.Human},
		Power:     opt.Val(game.PT{Value: 2}),
		Toughness: opt.Val(game.PT{Value: 2}),
	}})
	paper := addPsychicPaper(g, game.Player1)
	agent := &psychicPaperChoiceAgent{labels: []string{"Catalog Creature", "Elf", "Other Catalog Creature", "Goblin"}}
	agents := [game.NumPlayers]PlayerAgent{game.Player1: agent}
	ctx := &replacementChoiceContext{engine: NewEngine(nil), agents: agents}

	if !attachPermanentWithChoices(g, paper, first, ctx) {
		t.Fatal("initial attachment failed")
	}
	assertPsychicPaperIdentity(t, g, first, "Catalog Creature", types.Elf)
	if !slices.Contains(effectivePermanentValues(g, first).subtypes, types.Forest) {
		t.Fatal("setting the creature subtype removed the land subtype")
	}
	if slices.Contains(effectivePermanentValues(g, first).subtypes, types.Golem) {
		t.Fatal("setting the creature subtype retained the prior creature subtype")
	}
	if !attachedCreatureHasGrantedWard(g, first) {
		t.Fatal("equipped creature lacks ward")
	}
	if !ruleEffectProhibitsBeingBlocked(g, first) {
		t.Fatal("equipped creature can be blocked")
	}

	paper.Controller = game.Player2
	if got := permanentEffectiveName(g, first); got != "Catalog Creature" {
		t.Fatalf("name after control change = %q", got)
	}
	if agent.requests != 2 {
		t.Fatalf("control change prompted choices: requests = %d, want 2", agent.requests)
	}
	paper.Controller = game.Player1

	if !attachPermanentWithChoices(g, paper, second, ctx) {
		t.Fatal("reattachment failed")
	}
	if got := permanentEffectiveName(g, first); got != "Alpha Creature" {
		t.Fatalf("detached first creature name = %q, want printed name", got)
	}
	if !permanentHasSubtype(g, first, types.Golem) || permanentHasSubtype(g, first, types.Elf) {
		t.Fatalf("detached first creature subtypes = %v", effectivePermanentValues(g, first).subtypes)
	}
	assertPsychicPaperIdentity(t, g, second, "Other Catalog Creature", types.Goblin)
	if agent.requests != 4 {
		t.Fatalf("reattachment requests = %d, want 4", agent.requests)
	}

	clone := g.Clone()
	clonedPaper, ok := permanentByObjectID(clone, paper.ObjectID)
	if !ok {
		t.Fatal("clone lacks Psychic Paper")
	}
	clonedSecond, ok := permanentByObjectID(clone, second.ObjectID)
	if !ok {
		t.Fatal("clone lacks equipped creature")
	}
	assertPsychicPaperIdentity(t, clone, clonedSecond, "Other Catalog Creature", types.Goblin)
	if clonedPaper.EntryChoices[game.AttachmentCardNameChoiceKey].CardName != "Other Catalog Creature" {
		t.Fatalf("cloned choices = %#v", clonedPaper.EntryChoices)
	}

	var view PermanentView
	found := false
	for _, candidate := range NewObservation(g, game.Player2).Battlefield() {
		if candidate.ObjectID == paper.ObjectID {
			view = candidate
			found = true
			break
		}
	}
	if !found {
		t.Fatal("observation lacks Psychic Paper")
	}
	if view.EntryChoices[game.AttachmentCardNameChoiceKey].CardName != "Other Catalog Creature" ||
		view.EntryChoices[game.AttachmentSubtypeChoiceKey].Subtype != types.Goblin {
		t.Fatalf("observed choices = %#v", view.EntryChoices)
	}
	delete(view.EntryChoices, game.AttachmentCardNameChoiceKey)
	if _, ok := paper.EntryChoices[game.AttachmentCardNameChoiceKey]; !ok {
		t.Fatal("mutating observation changed permanent choices")
	}

	detachPermanent(g, paper)
	if got := permanentEffectiveName(g, second); got != "Beta Creature" {
		t.Fatalf("detached second creature name = %q, want printed name", got)
	}
	if permanentHasSubtype(g, second, types.Goblin) || !permanentHasSubtype(g, second, types.Human) {
		t.Fatalf("detached second creature subtypes = %v", effectivePermanentValues(g, second).subtypes)
	}
	if paper.EntryChoices[game.AttachmentSubtypeChoiceKey].Kind != game.ResolutionChoiceSubtype {
		t.Fatalf("detachment erased choices = %#v", paper.EntryChoices)
	}
}

func TestPsychicPaperEquipResolutionUsesChoiceAgent(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	g.CardNameCatalog = map[types.Card][]string{
		types.Creature: {"Nondefault Creature"},
	}
	target := addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:      "Target Creature",
		Types:     []types.Card{types.Creature},
		Subtypes:  []types.Sub{types.Human},
		Power:     opt.Val(game.PT{Value: 2}),
		Toughness: opt.Val(game.PT{Value: 2}),
	}})
	paper := addPsychicPaper(g, game.Player1)
	g.Stack.Push(&game.StackObject{
		ID:           g.IDGen.Next(),
		Kind:         game.StackActivatedAbility,
		SourceID:     paper.ObjectID,
		SourceCardID: paper.CardInstanceID,
		AbilityIndex: 0,
		Controller:   game.Player1,
		Targets:      []game.Target{game.PermanentTarget(target.ObjectID)},
	})
	agent := &psychicPaperChoiceAgent{labels: []string{"Nondefault Creature", "Elf"}}
	agents := [game.NumPlayers]PlayerAgent{game.Player1: agent}

	NewEngine(nil).resolveTopOfStackWithChoices(g, agents, &TurnLog{})

	if !paper.AttachedTo.Exists || paper.AttachedTo.Val != target.ObjectID {
		t.Fatalf("Psychic Paper attached to = %v, want %v", paper.AttachedTo, target.ObjectID)
	}
	assertPsychicPaperIdentity(t, g, target, "Nondefault Creature", types.Elf)
	if agent.requests != 2 {
		t.Fatalf("choice requests = %d, want 2", agent.requests)
	}
}

func assertPsychicPaperIdentity(t *testing.T, g *game.Game, permanent *game.Permanent, name string, subtype types.Sub) {
	t.Helper()
	if got := permanentEffectiveName(g, permanent); got != name {
		t.Fatalf("effective name = %q, want %q", got, name)
	}
	if !permanentHasSubtype(g, permanent, subtype) {
		t.Fatalf("effective subtypes = %v, want %s", effectivePermanentValues(g, permanent).subtypes, subtype)
	}
}
