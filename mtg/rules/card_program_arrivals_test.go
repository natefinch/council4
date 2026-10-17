package rules

import (
	"slices"
	"testing"

	"github.com/natefinch/council4/mtg/cardprogram"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

func resolveDestroyThenChooseArrival(ctx cardprogram.Context) error {
	result, err := ctx.Destroy(game.Destroy{Group: game.BattlefieldGroup(game.Selection{
		RequiredTypes: []types.Card{types.Creature},
	})})
	if err != nil {
		return err
	}
	var cards []cardprogram.FoundCard
	for _, member := range result.Members {
		if member.Destination != zone.Graveyard {
			continue
		}
		for _, card := range member.Cards {
			if slices.Contains(card.Info.Types, types.Creature) {
				cards = append(cards, card)
			}
		}
	}
	if len(cards) > 0 {
		options := make([]game.ChoiceOption, len(cards))
		for i, card := range cards {
			options[i] = game.ChoiceOption{Index: i, Label: card.Info.Name, Card: opt.Val(card.Info)}
		}
		selected, err := ctx.Ask(game.ChoiceRequest{
			Kind: game.ChoiceResolution, Player: ctx.View().Source().Controller,
			Prompt:  "Choose one creature card actually put into a graveyard this way.",
			Options: options, MinChoices: 1, MaxChoices: 1, DefaultSelection: []int{0},
		})
		if err != nil {
			return err
		}
		if _, err := ctx.MoveCards(cardprogram.CardPlacement{
			Cards: []cardprogram.FoundCard{cards[selected[0]]}, Destination: zone.Battlefield,
			Recipient: opt.Val(game.ControllerReference()),
			ContinuousEffects: []game.ContinuousEffect{
				{Layer: game.LayerColor, AddColors: []color.Color{color.Black}},
				{Layer: game.LayerType, AddSubtypes: []types.Sub{types.Zombie}},
			},
		}); err != nil {
			return err
		}
	}
	_, err = ctx.Execute(game.MoveResolvingSpell{Destination: zone.Exile})
	return err
}

type destructionArrivalChooser struct {
	name string
}

func (a destructionArrivalChooser) ChooseAction(PlayerObservation, []action.Action) action.Action {
	return action.Pass()
}

func (a destructionArrivalChooser) ChooseChoice(_ PlayerObservation, request game.ChoiceRequest) []int {
	if request.Kind == game.ChoiceResolution {
		for _, option := range request.Options {
			if option.Label == a.name {
				return []int{option.Index}
			}
		}
	}
	return request.DefaultSelection
}

func TestNecromanticSelectionSharedCompositionAndClone(t *testing.T) {
	reference := game.ProgramReference{ID: "fixture/destroy-choose-arrival", Version: "1"}
	catalog, err := cardprogram.NewCatalog([]cardprogram.Program{{
		Reference: reference, Resolve: resolveDestroyThenChooseArrival,
	}})
	if err != nil {
		t.Fatal(err)
	}
	engine := NewEngineWithPrograms(nil, catalog)
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	firstDef := vanillaCreature("First arrival", 2, 3)
	firstDef.Colors, firstDef.Subtypes = []color.Color{color.Green}, []types.Sub{types.Elf}
	first := addCombatPermanent(g, game.Player2, firstDef)
	second := addCombatPermanent(g, game.Player3, vanillaCreature("Second arrival", 2, 2))
	prior := addCardToGraveyard(g, game.Player2, vanillaCreature("Not destroyed this way", 3, 3))
	def := &game.CardDef{CardFace: game.CardFace{
		Name:  "Necromantic Selection shared-operation fixture",
		Types: []types.Card{types.Sorcery}, ManaCost: opt.Val(cost.Mana{cost.O(4), cost.B, cost.B, cost.B}),
		SpellAbility: opt.Val(game.Mode{Program: reference}.Ability()),
	}}
	spell := addCardToHand(g, game.Player1, def)
	setMainPhasePriority(g, game.Player1)
	g.Players[game.Player1].ManaPool.Add(mana.C, 4)
	g.Players[game.Player1].ManaPool.Add(mana.B, 3)
	if !engine.applyAction(g, game.Player1, action.CastSpell(spell, nil, 0, nil)) {
		t.Fatal("paid non-targeted composition did not cast")
	}
	branch := g.Clone()
	engine.resolveTopOfStackWithChoices(g, [game.NumPlayers]PlayerAgent{
		game.Player1: destructionArrivalChooser{name: "First arrival"},
	}, &TurnLog{})
	engine.resolveTopOfStackWithChoices(branch, [game.NumPlayers]PlayerAgent{
		game.Player1: destructionArrivalChooser{name: "Second arrival"},
	}, &TurnLog{})
	if permanentByCardID(g, first.CardInstanceID) == nil || permanentByCardID(g, second.CardInstanceID) != nil ||
		permanentByCardID(branch, second.CardInstanceID) == nil || permanentByCardID(branch, first.CardInstanceID) != nil {
		t.Fatal("choices did not stay restricted to actual arrivals and independent per clone")
	}
	if !g.Players[game.Player2].Graveyard.Contains(prior) ||
		!g.Players[game.Player1].Exile.Contains(spell) || !branch.Players[game.Player1].Exile.Contains(spell) {
		t.Fatal("composition moved an unrelated graveyard card or failed to exile its resolving spell")
	}
	arrived := permanentByCardID(g, first.CardInstanceID)
	if !permanentHasSubtype(g, arrived, types.Elf) || !permanentHasSubtype(g, arrived, types.Zombie) ||
		!slices.Contains(permanentEffectiveColors(g, arrived), color.Green) ||
		!slices.Contains(permanentEffectiveColors(g, arrived), color.Black) {
		t.Fatal("composition did not preserve original/additive characteristics")
	}
}

func arrivalContext(g *game.Game) *programContext {
	return &programContext{
		engine: NewEngine(nil), game: g, obj: &game.StackObject{Controller: game.Player1},
		active: true, log: &TurnLog{}, source: cardprogram.Source{Controller: game.Player1},
	}
}

func TestDestroyArrivalsComposeWithModifiedBattlefieldEntry(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	ctx := arrivalContext(g)
	prior := addCardToGraveyard(g, game.Player2, vanillaCreature("Preexisting graveyard card", 2, 2))
	creature := vanillaCreature("Destroyed green Elf", 2, 3)
	creature.Colors, creature.Subtypes = []color.Color{color.Green}, []types.Sub{types.Elf}
	first := addCombatPermanent(g, game.Player2, creature)
	second := addCombatPermanent(g, game.Player3, vanillaCreature("Destroyed colorless creature", 2, 2))
	protectedDef := vanillaCreature("Protected", 2, 2)
	protectedDef.StaticAbilities = []game.StaticAbility{game.IndestructibleStaticBody}
	protected := addCombatPermanent(g, game.Player4, protectedDef)
	artifact := addCombatPermanent(g, game.Player2, &game.CardDef{CardFace: game.CardFace{
		Name: "Not destroyed", Types: []types.Card{types.Artifact},
	}})
	destruction, err := ctx.Destroy(game.Destroy{Group: game.BattlefieldGroup(game.Selection{
		RequiredTypes: []types.Card{types.Creature},
	})})
	if err != nil {
		t.Fatal(err)
	}
	var found []cardprogram.FoundCard
	for _, member := range destruction.Members {
		if member.Destination == zone.Graveyard {
			found = append(found, member.Cards...)
		}
	}
	if len(found) != 2 || !g.Players[game.Player2].Graveyard.Contains(prior) {
		t.Fatalf("actual arrival handles = %d, want only the two newly destroyed creatures", len(found))
	}
	arrivals, err := ctx.MoveCards(cardprogram.CardPlacement{
		Cards: found, Destination: zone.Battlefield, Recipient: opt.Val(game.ControllerReference()),
		ContinuousEffects: []game.ContinuousEffect{
			{Layer: game.LayerColor, AddColors: []color.Color{color.Black}},
			{Layer: game.LayerType, AddSubtypes: []types.Sub{types.Zombie}},
		},
		EntryCounters: []game.CounterPlacement{{Kind: counter.Flying, Amount: 1}},
		EntryTapped:   true,
	})
	if err != nil || len(arrivals) != 2 {
		t.Fatalf("modified simultaneous placement = %+v, %v", arrivals, err)
	}
	for _, id := range []game.ObjectID{first.CardInstanceID, second.CardInstanceID} {
		permanent := permanentByCardID(g, id)
		if permanent == nil || effectiveController(g, permanent) != game.Player1 ||
			!permanentHasSubtype(g, permanent, types.Zombie) || !hasKeyword(g, permanent, game.Flying) || !permanent.Tapped {
			t.Fatal("actual returned creature lost recipient, entry counter, subtype or tappedness")
		}
		if !slices.Contains(permanentEffectiveColors(g, permanent), color.Black) {
			t.Fatal("black was not added")
		}
	}
	returned := permanentByCardID(g, first.CardInstanceID)
	if !permanentHasSubtype(g, returned, types.Elf) ||
		!slices.Contains(permanentEffectiveColors(g, returned), color.Green) {
		t.Fatal("additive entry effects erased the original characteristics")
	}
	if _, exists := permanentByObjectID(g, protected.ObjectID); !exists {
		t.Fatal("indestructible creature was destroyed")
	}
	if _, exists := permanentByObjectID(g, artifact.ObjectID); !exists {
		t.Fatal("nonmatching permanent was destroyed")
	}
	var batch game.ObjectID
	count := 0
	for _, event := range g.Events {
		if event.Kind != game.EventZoneChanged || event.FromZone != zone.Graveyard || event.ToZone != zone.Battlefield {
			continue
		}
		if event.SimultaneousID == 0 || batch != 0 && event.SimultaneousID != batch {
			t.Fatal("returns were not one simultaneous entry")
		}
		batch = event.SimultaneousID
		count++
	}
	if count != 2 {
		t.Fatal("expected two actual simultaneous returns")
	}
}

func TestDestroyArrivalHandlesRespectReplacementAndIncarnation(t *testing.T) {
	for _, destination := range []zone.Type{zone.Graveyard, zone.Exile, zone.Hand, zone.Library} {
		t.Run(destination.String(), func(t *testing.T) {
			g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
			ctx := arrivalContext(g)
			target := addCombatPermanent(g, game.Player2, vanillaCreature("Arrival", 2, 2))
			if destination != zone.Graveyard {
				g.ReplacementEffects = []game.ReplacementEffect{{
					ID: g.IDGen.Next(), MatchEvent: game.EventZoneChanged,
					MatchFromZone: true, FromZone: zone.Battlefield,
					MatchToZone: true, ToZone: zone.Graveyard, ReplaceToZone: destination,
				}}
			}
			result, err := ctx.Destroy(game.Destroy{Group: game.BattlefieldGroup(game.Selection{})})
			if err != nil || len(result.Members) != 1 {
				t.Fatalf("destroy result: %+v, %v", result, err)
			}
			member := result.Members[0]
			public := destination == zone.Graveyard || destination == zone.Exile
			if !public {
				if len(member.Cards) != 0 {
					t.Fatal("hidden-zone redirect minted public movement authority")
				}
				return
			}
			if len(member.Cards) != 1 || member.Cards[0].Identity.CardID != target.CardInstanceID {
				t.Fatal("actual public redirect did not mint an exact handle")
			}
			foreign := arrivalContext(g)
			if _, err := foreign.MoveCards(cardprogram.CardPlacement{Cards: member.Cards, Destination: zone.Hand}); err == nil {
				t.Fatal("another invocation accepted an unminted handle")
			}
			if !moveCardBetweenZones(g, game.Player2, target.CardInstanceID, destination, zone.Hand) ||
				!moveCardBetweenZones(g, game.Player2, target.CardInstanceID, zone.Hand, destination) {
				t.Fatal("could not change the arrival incarnation")
			}
			arrivals, err := ctx.MoveCards(cardprogram.CardPlacement{Cards: member.Cards, Destination: zone.Battlefield})
			if err != nil || len(arrivals) != 0 || permanentByCardID(g, target.CardInstanceID) != nil {
				t.Fatal("stale arrival followed a physical card through later movement")
			}
			ctx.active = false
			if _, err := ctx.MoveCards(cardprogram.CardPlacement{Cards: member.Cards, Destination: zone.Hand}); err == nil {
				t.Fatal("expired context retained arrival authority")
			}
		})
	}
}

func TestPublicArrivalReceiptsCanBeComposedAgain(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	ctx := arrivalContext(g)
	permanent := addCombatPermanent(g, game.Player2, vanillaCreature("Two placements", 2, 2))
	result, err := ctx.Destroy(game.Destroy{Group: game.BattlefieldGroup(game.Selection{})})
	if err != nil {
		t.Fatal(err)
	}
	exiled, err := ctx.MoveCards(cardprogram.CardPlacement{Cards: result.Members[0].Cards, Destination: zone.Exile})
	if err != nil || len(exiled) != 1 {
		t.Fatalf("first placement = %+v, %v", exiled, err)
	}
	returned, err := ctx.MoveCards(cardprogram.CardPlacement{Cards: exiled, Destination: zone.Battlefield})
	if err != nil || len(returned) != 1 || permanentByCardID(g, permanent.CardInstanceID) == nil {
		t.Fatalf("composed public arrival = %+v, %v", returned, err)
	}
}

func TestDestroyTokensAndCommanderCheckpointDoNotWidenArrivalAuthority(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	ctx := arrivalContext(g)
	token, ok := createTokenPermanentWithChoicesInBatch(ctx.engine, g, game.Player2,
		vanillaCreature("Temporary token", 2, 2), 0, false, [game.NumPlayers]PlayerAgent{}, &TurnLog{})
	if !ok {
		t.Fatal("token fixture failed")
	}
	commander := addCommanderPermanent(g, game.Player3)
	result, err := ctx.Destroy(game.Destroy{Group: game.BattlefieldGroup(game.Selection{})})
	if err != nil {
		t.Fatal(err)
	}
	var authority []cardprogram.FoundCard
	for _, member := range result.Members {
		if member.Before.ObjectID == token.ObjectID && len(member.Cards) != 0 {
			t.Fatal("departed token minted reanimation authority")
		}
		authority = append(authority, member.Cards...)
	}
	if len(authority) != 1 || authority[0].Identity.CardID != commander.CardInstanceID {
		t.Fatal("only the real commander graveyard arrival should be movable")
	}
	ctx.engine.applyStateBasedActions(g)
	if !g.Players[game.Player3].CommandZone.Contains(commander.CardInstanceID) {
		t.Fatal("normal commander SBA did not run")
	}
	arrivals, err := ctx.MoveCards(cardprogram.CardPlacement{Cards: authority, Destination: zone.Battlefield})
	if err != nil || len(arrivals) != 0 || permanentByCardID(g, commander.CardInstanceID) != nil {
		t.Fatal("an arrival handle followed the commander through its SBA move")
	}
}

func TestModifiedArrivalPlacementFailsBeforeMutation(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	ctx := arrivalContext(g)
	permanent := addCombatPermanent(g, game.Player2, vanillaCreature("Still in graveyard", 2, 2))
	result, err := ctx.Destroy(game.Destroy{Group: game.BattlefieldGroup(game.Selection{})})
	if err != nil {
		t.Fatal(err)
	}
	for _, placement := range []cardprogram.CardPlacement{
		{Cards: result.Members[0].Cards, Destination: zone.Hand, EntryTapped: true},
		{Cards: result.Members[0].Cards, Destination: zone.Battlefield,
			EntryCounters: []game.CounterPlacement{{Kind: counter.Poison, Amount: 1}}},
		{Cards: result.Members[0].Cards, Destination: zone.Battlefield,
			EntryCounters: []game.CounterPlacement{{Kind: counter.Flying, Amount: -1}}},
	} {
		if _, err := ctx.MoveCards(placement); err == nil {
			t.Fatal("invalid entry placement was silently accepted")
		}
		if !g.Players[game.Player2].Graveyard.Contains(permanent.CardInstanceID) ||
			permanentByCardID(g, permanent.CardInstanceID) != nil {
			t.Fatal("invalid placement partially moved the card")
		}
	}
}
