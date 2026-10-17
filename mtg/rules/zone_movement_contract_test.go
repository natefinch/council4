package rules

import (
	"strings"
	"testing"

	"github.com/natefinch/council4/mtg/cards"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
)

type zoneMovementChoices struct {
	command bool
	asked   int
}

func (*zoneMovementChoices) ChooseAction(PlayerObservation, []action.Action) action.Action {
	return action.Pass()
}

func (a *zoneMovementChoices) ChooseChoice(_ PlayerObservation, request game.ChoiceRequest) []int {
	if request.Kind == game.ChoiceMay && strings.Contains(request.Prompt, "commander") {
		a.asked++
		if a.command {
			return []int{1}
		}
		return []int{0}
	}
	if request.Kind == game.ChoiceMay {
		return []int{0}
	}
	return request.DefaultSelection
}

type libraryHandOperation struct {
	name      string
	primitive game.Primitive
	private   bool
	from      zone.Type
}

func libraryHandOperations() []libraryHandOperation {
	return []libraryHandOperation{
		{"draw", game.Draw{Player: game.ControllerReference(), Amount: game.Fixed(1)}, true, zone.Library},
		{"dig", game.Dig{Player: game.ControllerReference(), Look: game.Fixed(1), Take: game.Fixed(1)}, true, zone.Library},
		{"dig hand slot", game.Dig{
			Player: game.ControllerReference(), Look: game.Fixed(1), Take: game.Fixed(0),
			Slots: []game.DigSlot{{Count: game.Fixed(1), Destination: zone.Hand}},
		}, true, zone.Library},
		{"reveal until", game.RevealUntil{Player: game.ControllerReference(), Destination: zone.Hand}, false, zone.Library},
		{"reveal until random bottom", game.RevealUntil{
			Player: game.ControllerReference(), Destination: zone.Hand, MatchToDestinationRestRandomBottom: true,
		}, false, zone.Library},
		{"reveal until chosen bottom", game.RevealUntil{
			Player: game.ControllerReference(), Destination: zone.Hand, MatchToDestinationRestChosenBottom: true,
		}, false, zone.Library},
		{"reveal partition", game.RevealTopPartition{Player: game.ControllerReference(), Amount: game.Fixed(1)}, false, zone.Library},
		{"pile split", game.PileSplit{
			Player: game.ControllerReference(), Amount: game.Fixed(1), Kept: zone.Hand, Other: zone.Graveyard,
		}, false, zone.Library},
		{"reveal and drain", game.RevealToHandDrainManaValue{Amount: game.Fixed(1)}, false, zone.Library},
		{"consultation", game.IterativeLibraryProcess{
			Player: game.ControllerReference(), Stop: game.IterativeLibraryStopChosenName, ChooseName: true,
		}, false, zone.Library},
		{"discover", game.DiscoverCards{Amount: game.Fixed(100)}, false, zone.Exile},
	}
}

func TestLibraryHandOperationsOfferCommanderReplacement(t *testing.T) {
	for _, operation := range libraryHandOperations() {
		t.Run(operation.name, func(t *testing.T) {
			for _, command := range []bool{false, true} {
				name := "decline"
				if command {
					name = "accept"
				}
				t.Run(name, func(t *testing.T) {
					g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
					engine := NewEngine(nil)
					cardID := addCommanderCardToLibrary(g, game.Player1)
					addEffectSpellToStack(g, game.Player1, operation.primitive, nil)
					owner := &zoneMovementChoices{command: command}
					log := TurnLog{}
					engine.resolveTopOfStackWithChoices(g, [game.NumPlayers]PlayerAgent{game.Player1: owner}, &log)
					if owner.asked != 1 {
						t.Fatalf("owner choices = %d, want 1", owner.asked)
					}
					player := g.Players[game.Player1]
					if player.CommandZone.Contains(cardID) != command || player.Hand.Contains(cardID) == command {
						t.Fatal("placement did not honor the commander's owner")
					}
					draws := 0
					if operation.primitive.Kind() == game.PrimitiveDraw && !command {
						draws = 1
					}
					if got := countEventsOfKind(g.Events, game.EventCardDrawn); got != draws {
						t.Fatalf("actual draw events = %d, want %d", got, draws)
					}
					engine.applyStateBasedActions(g)
					if owner.asked != 1 {
						t.Fatal("hand replacement was incorrectly offered again as an SBA")
					}
					if operation.private {
						for _, event := range g.Events {
							if event.CardID == cardID && event.Kind == game.EventCardRevealed {
								t.Fatal("private hand placement revealed the commander")
							}
						}
					}
				})
			}
		})
	}
}

func TestLibraryHandOperationsHonorGeneralReplacement(t *testing.T) {
	for _, operation := range libraryHandOperations() {
		t.Run(operation.name, func(t *testing.T) {
			g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
			engine := NewEngine(nil)
			cardID := addCardToLibrary(g, game.Player1, vanillaCreature("Redirected", 2, 2))
			g.ReplacementEffects = []game.ReplacementEffect{{
				ID: g.IDGen.Next(), MatchEvent: game.EventZoneChanged,
				MatchFromZone: true, FromZone: operation.from,
				MatchToZone: true, ToZone: zone.Hand, ReplaceToZone: zone.Exile,
			}}
			addEffectSpellToStack(g, game.Player1, operation.primitive, nil)
			engine.resolveTopOfStackWithChoices(g, [game.NumPlayers]PlayerAgent{
				game.Player1: &zoneMovementChoices{},
			}, &TurnLog{})
			if !g.Players[game.Player1].Exile.Contains(cardID) || g.Players[game.Player1].Hand.Contains(cardID) {
				t.Fatal("operation bypassed the shared replacement-aware move")
			}
			if g.FailedDraws[game.Player1] {
				t.Fatal("a replaced draw was treated as drawing from an empty library")
			}
			if countEventsOfKind(g.Events, game.EventCardDrawn) != 0 {
				t.Fatal("a replaced hand placement emitted an actual draw")
			}
		})
	}
}

func TestExploreUsesSharedHandPlacement(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	land := addCardToLibrary(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name: "Revealed land", Types: []types.Card{types.Land},
	}})
	explorer := addCombatCreaturePermanentWithPower(g, game.Player1, 2)
	g.ReplacementEffects = []game.ReplacementEffect{{
		ID: g.IDGen.Next(), MatchEvent: game.EventZoneChanged,
		MatchFromZone: true, FromZone: zone.Library,
		MatchToZone: true, ToZone: zone.Hand, ReplaceToZone: zone.Exile,
	}}
	addEffectSpellToStack(g, game.Player1, game.Explore{Creature: game.TargetPermanentReference(0)},
		[]game.Target{game.PermanentTarget(explorer.ObjectID)})
	engine.resolveTopOfStack(g, &TurnLog{})
	if !g.Players[game.Player1].Exile.Contains(land) || g.Players[game.Player1].Hand.Contains(land) {
		t.Fatal("explore bypassed the library-to-hand replacement")
	}
	if effectivePower(g, explorer) != 2 {
		t.Fatal("revealing a land incorrectly placed the nonland +1/+1 counter")
	}
}

func TestExiledCommanderLibraryReturnsUseSharedReplacement(t *testing.T) {
	const key game.LinkedKey = "movement-contract"
	operations := []struct {
		name      string
		primitive game.Primitive
	}{
		{"linked top", game.PutLinkedExiledCardsInLibrary{LinkedKey: key}},
		{"linked bottom", game.PutLinkedExiledCardsInLibrary{LinkedKey: key, Bottom: true}},
		{"linked random bottom", game.PutLinkedExiledCardsInLibrary{LinkedKey: key, Bottom: true, RandomOrder: true}},
		{"cost partition", game.PartitionExiledCostCards{ChosenToLibraryBottom: true}},
		{"unreturned linked cards", game.ReturnLinkedExiledCardsToBattlefield{
			LinkedKey: key, Amount: game.Fixed(0), RestToLibraryBottom: true,
		}},
	}
	for _, operation := range operations {
		t.Run(operation.name, func(t *testing.T) {
			for _, command := range []bool{false, true} {
				name := "decline"
				if command {
					name = "accept"
				}
				t.Run(name, func(t *testing.T) {
					g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
					engine := NewEngine(nil)
					commander := addCommanderPermanent(g, game.Player2)
					if !movePermanentToZone(g, commander, zone.Exile) {
						t.Fatal("could not prepare exiled commander")
					}
					engine.setReplacementChoiceContext(g, [game.NumPlayers]PlayerAgent{
						game.Player2: &zoneMovementChoices{},
					}, &TurnLog{})
					engine.applyStateBasedActions(g)
					g.ClearChoiceContext()
					if !g.Players[game.Player2].Exile.Contains(commander.CardInstanceID) {
						t.Fatal("setup did not decline the prior exile SBA")
					}
					source := addCombatPermanent(g, game.Player1, oringCardDef())
					obj := linkedSourceObject(source)
					obj.ExiledAsCostIDs = []game.ObjectID{commander.CardInstanceID}
					card, _ := g.GetCardInstance(commander.CardInstanceID)
					rememberLinkedObject(g, linkedObjectSourceKey(g, obj, string(key)), game.LinkedObjectRef{
						ObjectID: commander.ObjectID, CardID: card.ID,
						CardZoneVersion: card.ZoneVersion, CardZoneVersionSet: true,
					})
					owner := &zoneMovementChoices{command: command}
					controller := &zoneMovementChoices{command: !command}
					agents := [game.NumPlayers]PlayerAgent{game.Player1: controller, game.Player2: owner}
					log := TurnLog{}
					engine.setReplacementChoiceContext(g, agents, &log)
					engine.resolveInstructionWithChoices(g, obj, &game.Instruction{Primitive: operation.primitive}, agents, &log)
					g.ClearChoiceContext()
					player := g.Players[game.Player2]
					if owner.asked != 1 || controller.asked != 0 {
						t.Fatal("library return did not ask the owner rather than the effect controller")
					}
					if player.CommandZone.Contains(card.ID) != command || player.Library.Contains(card.ID) == command {
						t.Fatal("library return ignored the owner's replacement decision")
					}
				})
			}
		})
	}
}
func TestDigCardsUseSharedCommanderMovement(t *testing.T) {
	for _, name := range []string{"Impulse", "Anticipate", "Strategic Planning"} {
		t.Run(name, func(t *testing.T) {
			registry := cards.NewDefaultRegistry()
			def, err := registry.LookupChecked(name)
			if err != nil {
				t.Fatal(err)
			}
			g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
			engine := NewEngineWithPrograms(nil, registry.Programs())
			commander := addCommanderCardToLibrary(g, game.Player1)
			spell := addCardToHand(g, game.Player1, def)
			setMainPhasePriority(g, game.Player1)
			g.Players[game.Player1].ManaPool.Add(mana.C, 1)
			g.Players[game.Player1].ManaPool.Add(mana.U, 1)
			if !engine.applyAction(g, game.Player1, action.CastSpell(spell, nil, 0, nil)) {
				t.Fatal("real paid cast failed")
			}
			owner := &zoneMovementChoices{command: true}
			engine.resolveTopOfStackWithChoices(g, [game.NumPlayers]PlayerAgent{game.Player1: owner}, &TurnLog{})
			if owner.asked != 1 || !g.Players[game.Player1].CommandZone.Contains(commander) {
				t.Fatal("compiled card's shared Dig bypassed commander movement")
			}
		})
	}
}
