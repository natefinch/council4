package rules

import (
	"slices"
	"testing"

	"github.com/natefinch/council4/mtg/cards"
	"github.com/natefinch/council4/mtg/cards/authored"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
)

func choicesProgramEngine(t *testing.T) *Engine {
	t.Helper()
	return NewEngineWithPrograms(nil, cards.NewDefaultRegistry().Programs())
}

func TestAuthoredHiveMindCopiesForOtherPlayersNotItsController(t *testing.T) {
	t.Parallel()
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := choicesProgramEngine(t)
	mind := addCombatPermanent(g, game.Player1, authored.HiveMind.New())
	def := &game.CardDef{CardFace: game.CardFace{
		Name: "Triggering sorcery", Types: []types.Card{types.Sorcery},
	}}
	source := addImplementationSpellToStack(g, game.Player2, def, nil)
	original, _ := g.Stack.Peek()
	original.XValue = 7
	original.ChosenModes = []int{1}
	original.ResolvedAmounts = map[string]int{"paid": 2}
	emitEvent(g, game.Event{
		Kind: game.EventSpellCast, Controller: game.Player2, Player: game.Player2,
		CardID: source, StackObjectID: original.ID,
	})
	if !engine.putTriggeredAbilitiesOnStackWithChoices(g, [game.NumPlayers]PlayerAgent{}, &TurnLog{}) {
		t.Fatal("Hive Mind did not discover the real spell-cast trigger")
	}
	trigger, _ := g.Stack.Peek()
	if trigger.SourceID != mind.ObjectID {
		t.Fatal("wrong triggered source")
	}
	engine.resolveTopOfStack(g, &TurnLog{})
	seen := make(map[game.PlayerID]bool)
	for range 3 {
		copied, _ := g.Stack.Pop()
		if !copied.Copy || copied.Controller == game.Player2 || copied.XValue != 7 ||
			!copied.CopyOwner.Exists || copied.CopyOwner.Val != copied.Controller ||
			!slices.Equal(copied.ChosenModes, []int{1}) || copied.ResolvedAmounts["paid"] != 2 {
			t.Fatal("copy lost its copier, X, modes or captured receipts")
		}
		seen[copied.Controller] = true
		copied.ResolvedAmounts["paid"] = 999
	}
	if len(seen) != 3 || original.ResolvedAmounts["paid"] != 2 {
		t.Fatal("copy state was shared or an active player was omitted")
	}
	casts := 0
	for _, event := range g.Events {
		if event.Kind == game.EventSpellCast {
			casts++
		}
	}
	if casts != 1 {
		t.Fatal("copies emitted cast events")
	}
}

func TestAuthoredGildedDrakeExchangeAndIllegalTargetConsequence(t *testing.T) {
	t.Parallel()
	for _, illegal := range []bool{false, true} {
		t.Run(map[bool]string{false: "exchange", true: "illegal target"}[illegal], func(t *testing.T) {
			t.Parallel()
			g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
			engine := choicesProgramEngine(t)
			def := authored.GildedDrake.New()
			drake := addCombatPermanent(g, game.Player1, def)
			target := addCombatCreaturePermanent(g, game.Player2)
			if illegal {
				movePermanentToZone(g, target, zone.Exile)
			}
			g.Stack.Push(&game.StackObject{
				ID: g.IDGen.Next(), Kind: game.StackTriggeredAbility,
				SourceID: drake.ObjectID, SourceCardID: drake.CardInstanceID, Controller: game.Player1,
				InlineTrigger: &def.TriggeredAbilities[0],
				Targets:       []game.Target{game.PermanentTarget(target.ObjectID)}, TargetCounts: []int{1},
			})
			engine.resolveTopOfStack(g, &TurnLog{})
			if illegal {
				if !g.Players[game.Player1].Graveyard.Contains(drake.CardInstanceID) {
					t.Fatal("illegal-target exception did not resolve the sacrifice consequence")
				}
			} else if effectiveController(g, drake) != game.Player2 || effectiveController(g, target) != game.Player1 {
				t.Fatal("exchange did not change both controllers")
			}
		})
	}
}

func TestAuthoredGiftsUngivenFoundCardsAndOpponentPartition(t *testing.T) {
	t.Parallel()
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := choicesProgramEngine(t)
	var foundIDs []game.ObjectID
	for _, name := range []string{"One", "Two", "Three", "Four"} {
		foundIDs = append(foundIDs, addCardToLibrary(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: name}}))
	}
	addImplementationSpellToStack(g, game.Player1, authored.GiftsUngiven.New(), []game.Target{game.PlayerTarget(game.Player2)})
	log := TurnLog{}
	engine.resolveTopOfStackWithChoices(g, [game.NumPlayers]PlayerAgent{
		game.Player1: &choiceOnlyAgent{choices: [][]int{{0}, {0}, {0}, {0}}},
		game.Player2: &choiceOnlyAgent{choices: [][]int{{0, 2}}},
	}, &log)
	if !g.Players[game.Player1].Graveyard.Contains(foundIDs[1]) || !g.Players[game.Player1].Graveyard.Contains(foundIDs[3]) ||
		!g.Players[game.Player1].Hand.Contains(foundIDs[0]) || !g.Players[game.Player1].Hand.Contains(foundIDs[2]) {
		t.Fatal("opponent's exact partition was not applied")
	}
	if len(log.Choices) != 6 || log.Choices[4].Request.Player != game.Player2 || log.Choices[4].UsedFallback ||
		log.Choices[5].Request.Kind != game.ChoiceOrder || log.Choices[5].Request.Player != game.Player1 {
		t.Fatal("partition did not use the targeted opponent's recorded choice")
	}
}
