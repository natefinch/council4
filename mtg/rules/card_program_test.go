package rules

import (
	"reflect"
	"slices"
	"testing"

	"github.com/natefinch/council4/mtg/cardprogram"
	"github.com/natefinch/council4/mtg/cards"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

func TestAuthoredSpellMatchesLegacyLightningBolt(t *testing.T) {
	t.Parallel()
	original := cards.NewDefaultRegistry().Lookup("Lightning Bolt")
	if original == nil || !original.SpellAbility.Exists {
		t.Fatal("registered Lightning Bolt is missing")
	}
	originalMode := original.SpellAbility.Val.Modes[0]
	reference := game.ProgramReference{ID: "test/lightning-bolt", Version: "1"}
	catalog, err := cardprogram.NewCatalog([]cardprogram.Program{{
		Reference: reference,
		Resolve: func(ctx cardprogram.Context) error {
			_, err := ctx.Execute(originalMode.Sequence[0].Primitive)
			return err
		},
	}})
	if err != nil {
		t.Fatal(err)
	}
	authored := *original
	authored.SpellAbility = opt.Val(game.Mode{
		Targets: slices.Clone(originalMode.Targets), Program: reference,
	}.Ability())
	run := func(def *game.CardDef) (*game.Game, TurnLog) {
		t.Helper()
		g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
		engine := NewEngineWithPrograms(nil, catalog)
		addImplementationSpellToStack(g, game.Player1, def, []game.Target{game.PlayerTarget(game.Player2)})
		log := TurnLog{}
		engine.resolveTopOfStack(g, &log)
		return g, log
	}
	legacyGame, legacyLog := run(original)
	programGame, programLog := run(&authored)
	if !reflect.DeepEqual(legacyGame.Events, programGame.Events) ||
		!reflect.DeepEqual(legacyLog, programLog) ||
		legacyGame.Players[game.Player2].Life != programGame.Players[game.Player2].Life {
		t.Fatal("authored and declarative operations diverged")
	}
	if programGame.Players[game.Player2].Life != 37 {
		t.Fatal("authored program did not execute normal damage")
	}
	if original.SpellAbility.Val.Modes[0].Program != (game.ProgramReference{}) {
		t.Fatal("parity fixture changed the production card")
	}
}

func TestAuthoredBodiesUseTheSharedAbilityPath(t *testing.T) {
	t.Parallel()
	reference := game.ProgramReference{ID: "test/draw", Version: "1"}
	var retained cardprogram.Context
	catalog, err := cardprogram.NewCatalog([]cardprogram.Program{{
		Reference: reference,
		Resolve: func(ctx cardprogram.Context) error {
			retained = ctx
			if player, err := ctx.View().ActivePlayer(); err != nil || player != game.Player1 {
				t.Fatalf("active-player query = %v, %v", player, err)
			}
			_, err := ctx.Execute(game.Draw{Amount: game.Fixed(1), Player: game.ControllerReference()})
			return err
		},
	}})
	if err != nil {
		t.Fatal(err)
	}
	for _, kind := range []game.StackObjectKind{game.StackActivatedAbility, game.StackTriggeredAbility} {
		g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
		drawn := addCardToLibrary(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Drawn"}})
		obj := &game.StackObject{ID: g.IDGen.Next(), Kind: kind, Controller: game.Player1}
		body := game.Mode{Program: reference}.Ability()
		engine := NewEngineWithPrograms(nil, catalog)
		engine.resolveAbilityContentWithChoices(g, obj, body, [game.NumPlayers]PlayerAgent{}, &TurnLog{})
		if !g.Players[game.Player1].Hand.Contains(drawn) {
			t.Fatal("ability program did not draw")
		}
		if _, err := retained.Execute(game.Draw{Amount: game.Fixed(1), Player: game.ControllerReference()}); err == nil {
			t.Fatal("retained execution context remained active")
		}
		if _, err := retained.View().ActivePlayer(); err == nil {
			t.Fatal("retained active-player query remained active")
		}
	}
}

func TestProgramValidationPrecedesSetupAndSharesWithSearch(t *testing.T) {
	t.Parallel()
	reference := game.ProgramReference{ID: "test/installed", Version: "1"}
	def := &game.CardDef{CardFace: game.CardFace{
		Name: "Authored", Types: []types.Card{types.Sorcery},
		SpellAbility: opt.Val(game.Mode{Program: reference}.Ability()),
	}}
	var configs [game.NumPlayers]game.PlayerConfig
	configs[0].Deck = []*game.CardDef{def}
	if err := NewEngine(nil).ValidatePrograms(configs); err == nil {
		t.Fatal("missing program passed preflight")
	}
	catalog, err := cardprogram.NewCatalog([]cardprogram.Program{{
		Reference: reference, Resolve: func(cardprogram.Context) error { return nil },
	}})
	if err != nil {
		t.Fatal(err)
	}
	engine := NewEngineWithPrograms(nil, catalog)
	if err := engine.ValidatePrograms(configs); err != nil {
		t.Fatal(err)
	}
	g := engine.NewGame(configs)
	simulator := engine.newSearchContext(g, game.Player1).Simulator()
	if _, err := simulator.engine.programs.Lookup(reference); err != nil {
		t.Fatal("search engine lost the immutable catalog")
	}
	clone := g.Clone()
	if len(clone.Players[0].Library.All()) != len(g.Players[0].Library.All()) {
		t.Fatal("clone lost configured cards")
	}
	loaded, ok := clone.GetCardInstance(clone.Players[0].Library.All()[0])
	if !ok || loaded.Def.ProgramReferences()[0] != reference {
		t.Fatal("clone lost the versioned binding")
	}
}

func TestMalformedProgramContextsFailExplicitly(t *testing.T) {
	t.Parallel()
	var ctx *programContext
	if _, err := ctx.Execute(nil); err == nil {
		t.Fatal("nil context succeeded")
	}
	if _, err := (&programContext{}).Ask(game.ChoiceRequest{}); err == nil {
		t.Fatal("inactive context succeeded")
	}
}

func TestAuthoredChoiceAllowsRepeatedAllocations(t *testing.T) {
	t.Parallel()
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	ctx := &programContext{
		engine: NewEngine(nil), game: g, obj: &game.StackObject{Controller: game.Player1},
		log: &TurnLog{}, active: true,
	}
	selected, err := ctx.Ask(game.ChoiceRequest{
		Kind: game.ChoiceManaCombination, Player: game.Player1, MinChoices: 3, MaxChoices: 3,
		Options: []game.ChoiceOption{{Index: 0, Label: "Red"}, {Index: 1, Label: "Green"}},
	})
	if err != nil || len(selected) != 3 {
		t.Fatalf("valid repeated allocation failed: %v %v", selected, err)
	}
}
