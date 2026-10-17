package rules

import (
	"strings"
	"testing"

	"github.com/natefinch/council4/mtg/cardprogram"
	"github.com/natefinch/council4/mtg/cards"
	"github.com/natefinch/council4/mtg/cards/cardset"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

var rejectedManaProgram = game.ProgramReference{ID: "test/authored-mana", Version: "1"}

func manaProgramFixture(authored bool) *game.CardDef {
	mode := game.Mode{Sequence: []game.Instruction{{Primitive: game.AddMana{Amount: game.Fixed(1), ManaColor: mana.G}}}}
	if authored {
		mode = game.Mode{Program: rejectedManaProgram}
	}
	return &game.CardDef{CardFace: game.CardFace{
		Name: "Mana artifact", Types: []types.Card{types.Artifact},
		ManaAbilities: []game.ManaAbility{{AdditionalCosts: cost.Tap, Content: mode.Ability()}},
	}}
}

func manaAdmissionCatalog(t *testing.T) cardprogram.Catalog {
	t.Helper()
	catalog, err := cardprogram.NewCatalog([]cardprogram.Program{{
		Reference: rejectedManaProgram,
		Resolve: func(cardprogram.Context) error {
			t.Fatal("admission speculatively executed an opaque mana program")
			return nil
		},
	}})
	if err != nil {
		t.Fatal(err)
	}
	return catalog
}

func TestManaProgramAdmissionRejectsBeforeGameplay(t *testing.T) {
	for _, surface := range []string{"direct", "static grant", "resolving grant", "copy values", "token", "delayed grant", "copy rider"} {
		t.Run(surface, func(t *testing.T) {
			catalog := manaAdmissionCatalog(t)
			manaDef := manaProgramFixture(true)
			body := &manaDef.ManaAbilities[0]
			def := &game.CardDef{CardFace: game.CardFace{Name: "Introduces mana", Types: []types.Card{types.Artifact}}}
			effect := game.ContinuousEffect{Layer: game.LayerAbility, AffectedSource: true, AddAbilities: []game.Ability{body}}
			grant := game.ApplyContinuous{Object: opt.Val(game.SourcePermanentReference()), ContinuousEffects: []game.ContinuousEffect{effect}}
			switch surface {
			case "direct":
				def = manaDef
			case "static grant":
				def.StaticAbilities = []game.StaticAbility{{ContinuousEffects: []game.ContinuousEffect{effect}}}
			case "resolving grant":
				def.SpellAbility = opt.Val(game.Mode{Sequence: []game.Instruction{{Primitive: grant}}}.Ability())
			case "copy values":
				effect.Layer, effect.AddAbilities = game.LayerCopy, nil
				effect.CopyValues = opt.Val(game.CopyableValues{Name: "Opaque copied mana", Abilities: []game.Ability{body}})
				def.StaticAbilities = []game.StaticAbility{{ContinuousEffects: []game.ContinuousEffect{effect}}}
			case "token":
				def.SpellAbility = opt.Val(game.Mode{Sequence: []game.Instruction{{
					Primitive: game.CreateToken{Source: game.TokenDef(manaDef), Amount: game.Fixed(1)},
				}}}.Ability())
			case "delayed grant":
				def.SpellAbility = opt.Val(game.Mode{Sequence: []game.Instruction{{
					Primitive: game.CreateDelayedTrigger{Trigger: game.DelayedTriggerDef{
						Timing:  game.DelayedAtBeginningOfNextEndStep,
						Content: game.Mode{Sequence: []game.Instruction{{Primitive: grant}}}.Ability(),
					}},
				}}}.Ability())
			case "copy rider":
				replacement := game.EntersAsCopyReplacement("Copy an artifact.", &game.Selection{RequiredTypes: []types.Card{types.Artifact}},
					true, false, nil, false, nil, nil)
				replacement.Replacement.EntersAsCopyAddAbilities = []game.Ability{body}
				def.ReplacementAbilities = []game.ReplacementAbility{replacement}
			default:
				t.Fatalf("unknown surface %q", surface)
			}
			issues := game.ValidateCardDef(def)
			found := false
			for _, issue := range issues {
				found = found || strings.Contains(issue.Message, "program-backed mana abilities are unsupported")
			}
			if !found {
				t.Fatalf("game validator admitted opaque mana through %s: %+v", surface, issues)
			}
			if err := catalog.ValidateCard(def); err == nil || !strings.Contains(err.Error(), "program-backed mana abilities are unsupported") {
				t.Fatalf("catalog admitted %s or hid the support boundary: %v", surface, err)
			}
			var configs [game.NumPlayers]game.PlayerConfig
			configs[0].Deck = []*game.CardDef{def}
			if err := NewEngineWithPrograms(nil, catalog).ValidatePrograms(configs); err == nil ||
				!strings.Contains(err.Error(), "program-backed mana abilities are unsupported") {
				t.Fatalf("Engine preflight admitted %s: %v", surface, err)
			}
			registry, err := cards.NewProgramRegistry([]cardset.Entry{{
				Name: def.Name, OracleID: "test-mana-admission",
				New:      func() *game.CardDef { return def },
				Programs: []cardprogram.Program{{Reference: rejectedManaProgram, Resolve: func(cardprogram.Context) error { return nil }}},
			}})
			if err != nil {
				t.Fatal(err)
			}
			if _, err := registry.LookupChecked(def.Name); err == nil || !strings.Contains(err.Error(), "program-backed mana abilities are unsupported") {
				t.Fatalf("checked registry admitted %s: %v", surface, err)
			}
		})
	}
}

func TestManaProgramDeclarativeAdmissionMatchesImmediateExecution(t *testing.T) {
	def := manaProgramFixture(false)
	e := NewEngineWithPrograms(nil, manaAdmissionCatalog(t))
	var configs [game.NumPlayers]game.PlayerConfig
	configs[0].Deck = []*game.CardDef{def}
	if issues := game.ValidateCardDef(def); len(issues) != 0 {
		t.Fatal(issues)
	}
	if err := e.ValidatePrograms(configs); err != nil {
		t.Fatal(err)
	}
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	source := addCombatPermanent(g, game.Player1, def)
	if !e.applyAction(g, game.Player1, action.ActivateAbility(source.ObjectID, 0, nil, 0)) ||
		!source.Tapped || g.Players[game.Player1].ManaPool.Amount(mana.G) != 1 || g.Stack.Size() != 0 {
		t.Fatal("inspectable declarative mana no longer resolves immediately with its tap cost")
	}
}

func TestManaProgramRuntimeIntroductionsFailBeforeMutation(t *testing.T) {
	for _, authored := range []bool{false, true} {
		g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
		e := NewEngineWithPrograms(nil, manaAdmissionCatalog(t))
		def := manaProgramFixture(authored)
		source := addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
			Name: "Recipient artifact", Types: []types.Card{types.Artifact},
		}})
		ctx := &programContext{engine: e, game: g, active: true, log: &TurnLog{},
			obj: &game.StackObject{ID: g.IDGen.Next(), SourceID: source.ObjectID, Controller: game.Player1}}
		operation := game.ApplyContinuous{
			Object:            opt.Val(game.SourcePermanentReference()),
			ContinuousEffects: []game.ContinuousEffect{{Layer: game.LayerAbility, AddAbilities: []game.Ability{&def.ManaAbilities[0]}}},
		}

		_, grantErr := ctx.Execute(operation)
		_, tokenErr := ctx.CreateTokens(def, cardprogram.TokenRecipient{Player: game.Player1, Count: 1})
		if authored {
			if grantErr == nil || tokenErr == nil ||
				!strings.Contains(grantErr.Error(), "program-backed mana abilities are unsupported") ||
				!strings.Contains(tokenErr.Error(), "program-backed mana abilities are unsupported") ||
				len(g.ContinuousEffects) != 0 || len(g.Battlefield) != 1 {
				t.Fatalf("runtime opaque mana introduction did not fail before mutation: %v / %v", grantErr, tokenErr)
			}
		} else {
			if grantErr != nil || tokenErr != nil ||
				!e.applyAction(g, game.Player1, action.ActivateAbility(source.ObjectID, 0, nil, 0)) ||
				g.Players[game.Player1].ManaPool.Amount(mana.G) != 1 {
				t.Fatalf("declarative introduced mana was rejected: %v / %v", grantErr, tokenErr)
			}
		}
	}
}

func TestManaProgramMalformedBindingsAreRejectedPrecisely(t *testing.T) {
	for _, ref := range []game.ProgramReference{
		{ID: "test/missing-version"}, {Version: "1"}, {ID: "test/bad id", Version: "1"},
	} {
		def := manaProgramFixture(true)
		def.ManaAbilities[0].Content.Modes[0].Program = ref
		issues := game.ValidateCardDef(def)
		if len(issues) == 0 || !strings.Contains(issues[0].Message, "program-backed mana abilities are unsupported") {
			t.Fatalf("malformed program-backed mana was not rejected precisely: %+v", issues)
		}
	}
}

func TestManaProgramBoundaryStillAllowsIntroducedNonManaPrograms(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	e := NewEngineWithPrograms(nil, manaAdmissionCatalog(t))
	source := addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name: "Grant recipient", Types: []types.Card{types.Artifact},
	}})
	body := &game.TriggeredAbility{
		Trigger: game.TriggerCondition{Type: game.TriggerWhenever, Pattern: game.TriggerPattern{Event: game.EventLifeGained}},
		Content: game.Mode{Program: rejectedManaProgram}.Ability(),
	}
	ctx := &programContext{engine: e, game: g, active: true, log: &TurnLog{},
		obj: &game.StackObject{ID: g.IDGen.Next(), SourceID: source.ObjectID, Controller: game.Player1}}
	if _, err := ctx.Execute(game.ApplyContinuous{
		Object:            opt.Val(game.SourcePermanentReference()),
		ContinuousEffects: []game.ContinuousEffect{{Layer: game.LayerAbility, AddAbilities: []game.Ability{body}}},
	}); err != nil {
		t.Fatalf("ordinary program-backed trigger grant was rejected: %v", err)
	}
	token := &game.CardDef{CardFace: game.CardFace{
		Name: "Ordinary program token", Types: []types.Card{types.Artifact}, TriggeredAbilities: []game.TriggeredAbility{*body},
	}}
	if _, err := ctx.CreateTokens(token, cardprogram.TokenRecipient{Player: game.Player1, Count: 1}); err != nil {
		t.Fatalf("ordinary program-backed token trigger was rejected: %v", err)
	}
}
