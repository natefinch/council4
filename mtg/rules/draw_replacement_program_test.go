package rules

import (
	"slices"
	"testing"

	"github.com/natefinch/council4/mtg/cardprogram"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

func replacementDrawTwo(ctx cardprogram.Context) error {
	_, err := ctx.Execute(game.Draw{Player: game.EventPlayerReference(), Amount: game.Fixed(2)})
	return err
}

func replacementDoNothing(cardprogram.Context) error { return nil }

func TestDrawReplacementProgramDerivedDrawsInheritAppliedEffects(t *testing.T) {
	t.Parallel()
	ref := game.ProgramReference{ID: "test/replacement-draw-two", Version: "1"}
	catalog, err := cardprogram.NewCatalog([]cardprogram.Program{{Reference: ref, Resolve: replacementDrawTwo}})
	if err != nil {
		t.Fatal(err)
	}
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	abundanceLibrary(g, "A", "B", "C", "D", "Keep")
	e := NewEngineWithPrograms(nil, catalog)
	p := addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name: "Authored replacement", Types: []types.Card{types.Enchantment},
		ReplacementAbilities: []game.ReplacementAbility{{
			Replacement: game.ReplacementEffect{
				MatchEvent: game.EventCardDrawn, ControllerFilter: game.TriggerControllerYou,
				Program: ref, Duration: game.DurationPermanent,
			},
		}},
	}})
	registerPermanentReplacementEffects(g, p)
	drawDoublingPermanent(g, game.Player1, 2, false)
	a := &abundanceChoices{replacement: "doubler"}
	if !e.drawCards(g, game.Player1, 1, [game.NumPlayers]PlayerAgent{a}, &TurnLog{}) ||
		g.Players[game.Player1].Hand.Size() != 4 ||
		len(abundanceEvents(g, game.EventCardDrawn)) != 4 ||
		len(g.ReplacementDecisions) != 1 {
		t.Fatal("authored descendant draws reapplied themselves/ancestor or contaminated sibling lineage")
	}
	if _, exists := g.ChoiceContext(); exists {
		t.Fatal("pending draw lineage outlived the operation")
	}
}

func TestDrawReplacementSelfPrecedenceAndExpiredContext(t *testing.T) {
	t.Parallel()
	ref := game.ProgramReference{ID: "test/self-replacement", Version: "1"}
	var retained cardprogram.Context
	catalog, err := cardprogram.NewCatalog([]cardprogram.Program{{
		Reference: ref, Resolve: func(ctx cardprogram.Context) error {
			retained = ctx
			source := ctx.View().Source()
			if source.Controller != game.Player1 || !source.EventPlayer.Exists ||
				source.EventPlayer.Val != game.Player1 || source.CardID == 0 || source.ObjectID == 0 ||
				source.StackID != 0 {
				t.Fatal("replacement source/event metadata is not captured data")
			}
			return replacementDoNothing(ctx)
		},
	}})
	if err != nil {
		t.Fatal(err)
	}
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	abundanceLibrary(g, "Keep")
	e := NewEngineWithPrograms(nil, catalog)
	p := addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name: "Self replacement",
		ReplacementAbilities: []game.ReplacementAbility{{
			Replacement: game.ReplacementEffect{
				MatchEvent: game.EventCardDrawn, ControllerFilter: game.TriggerControllerYou,
				Program: ref, SelfReplacement: true, Duration: game.DurationPermanent,
			},
		}},
	}})
	registerPermanentReplacementEffects(g, p)
	drawDoublingPermanent(g, game.Player1, 2, false)
	e.drawCards(g, game.Player1, 1, [game.NumPlayers]PlayerAgent{&abundanceChoices{replacement: "doubler"}}, &TurnLog{})
	if retained == nil || g.Players[game.Player1].Library.Size() != 1 ||
		len(g.ReplacementDecisions) != 0 || g.FailedDraws[game.Player1] {
		t.Fatal("CR 616.1a self-replacement did not preempt an ordinary replacement")
	}
	if _, err := retained.Execute(game.Draw{Player: game.ControllerReference(), Amount: game.Fixed(1)}); err == nil {
		t.Fatal("a replacement context remained usable after its invocation")
	}
}

func TestDrawReplacementMultipliersRecomputeFirstActualDrawAndConditions(t *testing.T) {
	t.Parallel()
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	e := NewEngine(nil)
	abundanceLibrary(g, "A", "B", "C", "D", "Keep")
	drawDoublingPermanent(g, game.Player1, 2, false)
	drawDoublingPermanent(g, game.Player1, 2, true)
	e.drawCardWithReplacements(g, game.Player1, [game.NumPlayers]PlayerAgent{}, &TurnLog{}, true)
	// The first ancestor doubles. Insight exempts the first actual descendant,
	// but doubles the second. The ancestor cannot reapply to either.
	if g.Players[game.Player1].Hand.Size() != 3 {
		t.Fatal("first actual draw exemption or CR 614.5 lineage was not recomputed")
	}
	events := slices.DeleteFunc(slices.Clone(g.Events), func(event game.Event) bool { return event.Kind != game.EventCardDrawn })
	if len(events) != 3 || !events[0].FirstInDrawStep || events[1].FirstInDrawStep || events[2].FirstInDrawStep {
		t.Fatal("draw-step event metadata did not track the first actual draw")
	}
	g = game.NewGame([game.NumPlayers]game.PlayerConfig{})
	abundanceLibrary(g, "A", "B", "C")
	p := addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name: "Conditional doubler",
		ReplacementAbilities: []game.ReplacementAbility{
			game.MaxSpeedDrawCardMultiplierReplacement("max speed", 2, false),
		},
	}})
	registerPermanentReplacementEffects(g, p)
	e.drawCards(g, game.Player1, 1, [game.NumPlayers]PlayerAgent{}, &TurnLog{})
	if g.Players[game.Player1].Hand.Size() != 1 {
		t.Fatal("inspectable replacement Condition was ignored")
	}
}

func TestAbundanceDrawDigDredgeAndEmptyWinOrdering(t *testing.T) {
	t.Parallel()
	t.Run("dig before doubler", func(t *testing.T) {
		t.Parallel()
		g, e, _ := abundanceGame(t)
		abundanceLibrary(g, "A", "B", "C", "D")
		drawDigPermanent(g, game.Player1, 3, 1, game.DigRemainderGraveyard)
		drawDoublingPermanent(g, game.Player1, 2, false)
		a := &abundanceChoices{replacement: "digger"}
		e.drawCards(g, game.Player1, 1, [game.NumPlayers]PlayerAgent{a}, &TurnLog{})
		if g.Players[game.Player1].Hand.Size() != 1 || g.Players[game.Player1].Graveyard.Size() != 2 ||
			g.Players[game.Player1].Library.Size() != 1 || len(abundanceEvents(g, game.EventCardDrawn)) != 0 {
			t.Fatal("legacy dig action did not consume the pending draw before doubling")
		}
	})
	t.Run("dredge chosen before Abundance", func(t *testing.T) {
		t.Parallel()
		g, e, _ := abundanceGame(t)
		abundanceLibrary(g, "A", "B", "C")
		dredger := addCardToGraveyard(g, game.Player1, dredgeCardDef(2))
		agents := [game.NumPlayers]PlayerAgent{&choiceOnlyAgent{choices: [][]int{{1}, {0}}}}
		e.drawCards(g, game.Player1, 1, agents, &TurnLog{})
		if !g.Players[game.Player1].Hand.Contains(dredger) ||
			g.Players[game.Player1].Graveyard.Size() != 2 ||
			len(abundanceEvents(g, game.EventCardRevealed)) != 0 ||
			len(abundanceEvents(g, game.EventCardDrawn)) != 0 {
			t.Fatal("dredge did not compete in the shared replacement loop")
		}
	})
	for _, first := range []string{"choose land or nonland", "win instead"} {
		t.Run("empty "+first, func(t *testing.T) {
			t.Parallel()
			g, e, _ := abundanceGame(t)
			p := addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
				Name: "Empty library winner",
				ReplacementAbilities: []game.ReplacementAbility{
					game.DrawFromEmptyLibraryWinReplacement("win instead"),
				},
			}})
			registerPermanentReplacementEffects(g, p)
			e.drawCards(g, game.Player1, 1, [game.NumPlayers]PlayerAgent{&abundanceChoices{replacement: first}}, &TurnLog{})
			if g.FailedDraws[game.Player1] || g.MarkedToLoseGame[game.Player2] != (first == "win instead") {
				t.Fatal("empty-library win bypassed ordering or Abundance's executed empty result")
			}
		})
	}
}

func TestAbundanceSourceActivityAndControl(t *testing.T) {
	t.Parallel()
	g, e, _ := abundanceGame(t)
	cards := abundanceLibrary(g, "A", "Land1", "Keep")
	a := &abundanceChoices{}
	agents := [game.NumPlayers]PlayerAgent{a}
	source := g.Battlefield[0]
	source.Controller = game.Player2
	e.drawCards(g, game.Player1, 1, agents, &TurnLog{})
	if !g.Players[game.Player1].Hand.Contains(cards[0]) || len(a.requests) != 0 {
		t.Fatal("Abundance did not follow current control")
	}
	addCardToLibrary(g, game.Player2, &game.CardDef{CardFace: game.CardFace{Name: "LandP2", Types: []types.Card{types.Land}}})
	e.drawCards(g, game.Player2, 1, [game.NumPlayers]PlayerAgent{game.Player2: a}, &TurnLog{})
	if g.Players[game.Player2].Hand.Size() != 1 || len(a.requests) != 2 ||
		len(abundanceEvents(g, game.EventCardDrawn)) != 1 {
		t.Fatal("new controller could not replace their draw")
	}
	source.Controller = game.Player1
	source.PhasedOut = true
	a.requests = nil
	e.drawCards(g, game.Player1, 1, agents, &TurnLog{})
	if !g.Players[game.Player1].Hand.Contains(cards[1]) || len(a.requests) != 0 {
		t.Fatal("phased-out source applied")
	}
	source.PhasedOut = false
	movePermanentToZone(g, source, zone.Graveyard)
	e.drawCards(g, game.Player1, 1, agents, &TurnLog{})
	if !g.Players[game.Player1].Hand.Contains(cards[2]) || len(a.requests) != 0 {
		t.Fatal("departed source applied")
	}
}

func TestDrawReplacementProgramsRejectMissingOrMalformedPendingBindings(t *testing.T) {
	t.Parallel()
	ref := game.ProgramReference{ID: "test/missing-replacement", Version: "1"}
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	g.ReplacementEffects = []game.ReplacementEffect{{
		ID: g.IDGen.Next(), Program: ref, MatchEvent: game.EventCardDrawn,
	}}
	defer func() {
		if recovered := recover(); recovered == nil {
			t.Fatal("Simulator admitted a pending replacement with no installed code")
		}
	}()
	NewEngine(nil).Simulator().ResolvePriority(g, [game.NumPlayers]PlayerAgent{})
}

func TestDrawReplacementProgramDynamicDeclarationPreflight(t *testing.T) {
	t.Parallel()
	replacement := game.ReplacementEffect{
		Program:    game.ProgramReference{ID: "test/action", Version: "1"},
		MatchEvent: game.EventCardDrawn, Optional: true,
	}
	if err := game.ValidatePrimitive(game.CreateReplacement{Replacement: &replacement}, nil); err != nil {
		t.Fatal(err)
	}
	replacement.DrawCardMultiplier = 2
	if err := game.ValidatePrimitive(game.CreateReplacement{Replacement: &replacement}, nil); err == nil {
		t.Fatal("authored Execute could schedule mixed legacy/program actions")
	}
	replacement.DrawCardMultiplier = 0
	replacement.MatchEvent = game.EventZoneChanged
	if err := game.ValidatePrimitive(game.CreateReplacement{Replacement: &replacement}, nil); err == nil {
		t.Fatal("unsupported authored replacement event was silently admitted")
	}
	replacement.MatchEvent = game.EventCardDrawn
	replacement.Condition = opt.Val(game.Condition{SourceClassLevelAtLeast: 2})
	if err := game.ValidatePrimitive(game.CreateReplacement{Replacement: &replacement}, nil); err != nil {
		t.Fatal("inspectable applicability was rejected")
	}
}
