package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/cards/authored"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

func productionPrivateTutorCases() []productionTutorCase {
	return []productionTutorCase{
		{entry: authored.DemonicTutor, mana: []mana.Color{mana.C, mana.B}},
		{entry: authored.DiabolicTutor, mana: []mana.Color{mana.C, mana.C, mana.B, mana.B}},
	}
}

func TestProductionPrivateTutorsMandatoryFindAndDisclosure(t *testing.T) {
	for _, tc := range productionPrivateTutorCases() {
		t.Run(tc.entry.Name, func(t *testing.T) {
			for _, scenario := range []string{"normal", "redirected", "declined mandatory", "empty"} {
				t.Run(scenario, func(t *testing.T) {
					g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
					engine := removalTokenEngine(t, tc.entry)
					var found game.ObjectID
					var version uint64
					if scenario != "empty" {
						found = addCardToLibrary(g, game.Player1, vanillaCreature("Secret find", 2, 2))
						card, _ := g.GetCardInstance(found)
						version = card.ZoneVersion
					}
					if scenario == "redirected" {
						g.ReplacementEffects = []game.ReplacementEffect{{
							ID: g.IDGen.Next(), MatchEvent: game.EventZoneChanged,
							MatchFromZone: true, FromZone: zone.Library, MatchToZone: true, ToZone: zone.Hand,
							ReplaceToZone: zone.Exile,
						}}
					}
					epoch := g.Players[game.Player1].Library.ReorderVersion()
					requests := 0
					agent := &controlScript{choice: func(obs PlayerObservation, request game.ChoiceRequest) []int {
						requests++
						if request.Kind != game.ChoiceSearch || request.MinChoices != 1 ||
							request.MaxChoices != 1 || len(obs.LibraryLookable(game.Player1)) != 1 ||
							request.Options[0].Label != "Secret find" ||
							len(NewObservation(g, game.Player2).LibraryLookable(game.Player1)) != 0 {
							t.Fatal("private mandatory search offered the wrong information or cardinality")
						}
						if scenario == "declined mandatory" {
							return nil
						}
						return request.DefaultSelection
					}}
					spell := addImplementationSpellToStack(g, game.Player1, tc.entry.New(), nil)
					log := &TurnLog{}
					engine.resolveTopOfStackWithChoices(g, [game.NumPlayers]PlayerAgent{agent}, log)
					wantRequests := 1
					if scenario == "empty" {
						wantRequests = 0
					} else {
						destination := g.Players[game.Player1].Hand
						if scenario == "redirected" {
							destination = g.Players[game.Player1].Exile
						}
						card, _ := g.GetCardInstance(found)
						if !destination.Contains(found) || card.ZoneVersion != version+1 {
							t.Fatal("private find did not arrive as exactly one new incarnation")
						}
						moves := 0
						for _, event := range g.Events {
							if event.Kind == game.EventZoneChanged && event.CardID == found {
								moves++
								wantZone := zone.Hand
								if scenario == "redirected" {
									wantZone = zone.Exile
								}
								if event.FromZone != zone.Library || event.ToZone != wantZone ||
									event.CardZoneVersion != card.ZoneVersion || event.SourceID != spell {
									t.Fatal("authoritative movement fact lost actual identity, source or destination")
								}
							}
						}
						if moves != 1 {
							t.Fatal("private placement invented or duplicated movement")
						}
					}
					if requests != wantRequests || len(searchPartitionReveals(g)) != 0 ||
						countEventsOfKind(g.Events, game.EventCardDrawn) != 0 ||
						g.Players[game.Player1].Library.ReorderVersion() != epoch+1 ||
						len(NewObservation(g, game.Player2).HandOf(game.Player1)) != 0 ||
						len(NewObservation(g, game.Player1).LibraryLookable(game.Player1)) != 0 {
						t.Fatal("private search leaked identities, drew its find or skipped the closing shuffle")
					}
					for _, decision := range log.Choices {
						for _, option := range decision.Request.Options {
							if option.Label == "Secret find" || option.Card.Exists || len(option.Targets) != 0 {
								t.Fatal("search log leaked a private candidate")
							}
						}
					}
					if scenario == "declined mandatory" && g.Execution.InvalidChoices != 1 {
						t.Fatal("declining a mandatory find did not record its explicit invalid-choice fallback")
					}
				})
			}
		})
	}
}

func TestProductionPrivateTutorsControlAndReplacementChoices(t *testing.T) {
	for _, tc := range productionPrivateTutorCases() {
		t.Run(tc.entry.Name, func(t *testing.T) {
			for _, opposed := range []bool{false, true} {
				g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
				g.Turn.TurnNumber = 1
				g.PlayerControlEffects = []game.PlayerControlEffect{{
					ID: g.IDGen.Next(), Player: game.Player1, Controller: game.Player3, Turn: 1,
				}}
				decider := game.Player3
				destination := zone.Hand
				if opposed {
					addCombatPermanent(g, game.Player4, oppositionAgentDef())
					decider, destination = game.Player4, zone.Exile
				}
				found := addCardToLibrary(g, game.Player1, vanillaCreature("Controlled private find", 2, 2))
				for _, target := range []zone.Type{zone.Graveyard, zone.Command} {
					g.ReplacementEffects = append(g.ReplacementEffects, game.ReplacementEffect{
						ID: g.IDGen.Next(), MatchEvent: game.EventZoneChanged,
						MatchFromZone: true, FromZone: zone.Library, MatchToZone: true, ToZone: destination,
						ReplaceToZone: target,
					})
				}
				seenSearch, seenReplacement := false, false
				agent := &controlScript{choice: func(obs PlayerObservation, request game.ChoiceRequest) []int {
					if obs.Player != game.Player1 || obs.DecisionMaker != decider {
						t.Fatal("search and replacement choices did not preserve actor versus decision maker")
					}
					switch request.Kind {
					case game.ChoiceSearch:
						seenSearch = true
						if len(obs.LibraryLookable(game.Player1)) != 1 ||
							len(NewObservation(g, game.Player2).LibraryLookable(game.Player1)) != 0 {
							t.Fatal("controlled search visibility was not scoped")
						}
					case game.ChoiceReplacement:
						seenReplacement = true
						if request.Subject.Exists {
							t.Fatal("replacement request revealed the private find")
						}
						return []int{1}
					default:
					}
					return request.DefaultSelection
				}}
				policies := [game.NumPlayers]PlayerAgent{}
				policies[decider] = agent
				engine := removalTokenEngine(t, tc.entry)
				addImplementationSpellToStack(g, game.Player1, tc.entry.New(), nil)
				engine.resolveTopOfStackWithChoices(g, policies, &TurnLog{})
				if !seenSearch || !seenReplacement || !g.Players[game.Player1].CommandZone.Contains(found) ||
					len(g.SearchControlScopes) != 0 {
					t.Fatal("search bypassed its actual selected replacement")
				}
				if _, permission := playFromExilePermissionFor(g, found); permission {
					t.Fatal("Opposition Agent granted permission without an actual exile arrival")
				}
			}
		})
	}
}

func TestProductionPrivateTutorsSimulatorClone(t *testing.T) {
	for _, tc := range productionPrivateTutorCases() {
		t.Run(tc.entry.Name, func(t *testing.T) {
			g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
			engine := removalTokenEngine(t, tc.entry)
			found := addCardToLibrary(g, game.Player1, vanillaCreature("Simulator private find", 2, 2))
			addImplementationSpellToStack(g, game.Player1, tc.entry.New(), nil)
			branch := engine.Simulator().ResolvePriority(g, simPassPolicies())
			if !branch.Players[game.Player1].Hand.Contains(found) ||
				!g.Players[game.Player1].Library.Contains(found) || g.Stack.IsEmpty() {
				t.Fatal("Simulator changed the source game or failed private placement")
			}
		})
	}
}

func TestPrivateTutorCommanderChoiceDoesNotRevealFindInLog(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	found := addCommanderCardToLibrary(g, game.Player1)
	card, _ := g.GetCardInstance(found)
	engine := removalTokenEngine(t, authored.DemonicTutor)
	owner := &controlScript{choice: func(_ PlayerObservation, request game.ChoiceRequest) []int {
		if request.Kind == game.ChoiceMay {
			if !request.Subject.Exists || request.Subject.Val.CardID != found {
				t.Fatal("entitled owner lost the actual commander choice subject")
			}
			return []int{0}
		}
		return request.DefaultSelection
	}}
	addImplementationSpellToStack(g, game.Player1, authored.DemonicTutor.New(), nil)
	log := &TurnLog{}
	engine.resolveTopOfStackWithChoices(g, [game.NumPlayers]PlayerAgent{owner}, log)
	if !g.Players[game.Player1].Hand.Contains(found) {
		t.Fatal("declined commander replacement did not put the find into hand")
	}
	for _, decision := range log.Choices {
		if decision.Request.Subject.Exists {
			t.Fatal("private commander replacement leaked its identity into the log")
		}
		for _, option := range decision.Request.Options {
			if option.Card.Exists || option.Label == "Leave "+card.Def.Name+" in the destination zone" {
				t.Fatal("commander replacement log revealed the private find")
			}
		}
	}
}

func TestPrivateSearchPlacementReplacementBelongsToLibraryOwner(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	found := addCardToLibrary(g, game.Player2, selfLibraryReplacementCardDef())
	g.ReplacementEffects = []game.ReplacementEffect{{
		ID: g.IDGen.Next(), MatchEvent: game.EventZoneChanged,
		MatchToZone: true, ToZone: zone.Graveyard, ReplaceToZone: zone.Exile,
	}}
	def := &game.CardDef{CardFace: game.CardFace{
		Name: "Foreign library search", Types: []types.Card{types.Sorcery},
		SpellAbility: opt.Val(game.Mode{
			Targets: []game.TargetSpec{{MinTargets: 1, MaxTargets: 1, Allow: game.TargetAllowPlayer}},
			Sequence: []game.Instruction{{Primitive: game.Search{
				Player: game.TargetPlayerReference(0), Amount: game.Fixed(1),
				Spec: game.SearchSpec{SourceZone: zone.Library, Destination: zone.Graveyard},
			}}},
		}.Ability()),
	}}
	addImplementationSpellToStack(g, game.Player1, def, []game.Target{game.PlayerTarget(game.Player2)})
	ownerChose := false
	owner := &controlScript{choice: func(obs PlayerObservation, request game.ChoiceRequest) []int {
		if request.Kind == game.ChoiceReplacement {
			ownerChose = true
			if obs.Player != game.Player2 || obs.DecisionMaker != game.Player2 {
				t.Fatal("the resolving source was mistaken for the affected object")
			}
		}
		return request.DefaultSelection
	}}
	wrong := &controlScript{choice: func(_ PlayerObservation, request game.ChoiceRequest) []int {
		if request.Kind == game.ChoiceReplacement {
			t.Fatal("causative spell controller chose another library owner's replacement")
		}
		return request.DefaultSelection
	}}
	NewEngine(nil).runPriorityLoop(g, [game.NumPlayers]PlayerAgent{wrong, owner}, &TurnLog{})
	if !ownerChose || g.Players[game.Player2].Graveyard.Contains(found) {
		t.Fatal("owner did not choose the actual destination replacement")
	}
}

func TestPrivateSearchSameLibraryReplacementPreservesIncarnation(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	found := addCardToLibrary(g, game.Player1, selfLibraryReplacementCardDef())
	card, _ := g.GetCardInstance(found)
	version, epoch := card.ZoneVersion, g.Players[game.Player1].Library.ReorderVersion()
	addEffectSpellToStack(g, game.Player1, game.Search{
		Player: game.ControllerReference(), Amount: game.Fixed(1),
		Spec: game.SearchSpec{SourceZone: zone.Library, Destination: zone.Graveyard},
	}, nil)
	NewEngine(nil).resolveTopOfStack(g, &TurnLog{})
	if !g.Players[game.Player1].Library.Contains(found) || card.ZoneVersion != version {
		t.Fatal("a replacement shuffle within the source library invented a new incarnation")
	}
	if len(searchPartitionReveals(g)) != 1 || g.Players[game.Player1].Library.ReorderVersion() != epoch+2 {
		t.Fatalf("same-library replacement reveals=%d reorder delta=%d, want one reveal and two shuffles",
			len(searchPartitionReveals(g)), g.Players[game.Player1].Library.ReorderVersion()-epoch)
	}
	for _, event := range g.Events {
		if event.Kind == game.EventZoneChanged && event.CardID == found {
			t.Fatal("same-library shuffle emitted a fictitious zone change")
		}
	}
}
