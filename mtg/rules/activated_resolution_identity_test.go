package rules

import (
	"testing"

	"github.com/natefinch/council4/cardgen"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

func TestActivatedOrdinalDistinctAbilitiesAndSources(t *testing.T) {
	t.Parallel()
	text := "{1}: If this is the second time this ability has resolved this turn, draw a card."
	g, engine, source := ordinalFixture(t, text+"\n"+text)
	def, _ := permanentCardDef(g, source)
	other := addCombatPermanent(g, game.Player1, def)
	for _, activation := range []struct {
		source *game.Permanent
		index  int
		want   int
	}{
		{source, 0, 0}, {source, 1, 0}, {other, 0, 0},
		{source, 0, 1}, {source, 1, 2}, {other, 0, 3},
	} {
		activateOrdinal(t, g, engine, activation.source.ObjectID, activation.index, nil, nil)
		engine.resolveTopOfStack(g, &TurnLog{})
		if got := g.Players[game.Player1].Hand.Size(); got != activation.want {
			t.Fatalf("hand=%d, want %d; identical printed bodies must remain independent", got, activation.want)
		}
	}
}

func TestActivatedOrdinalTransformCapturesBody(t *testing.T) {
	t.Parallel()
	g, engine, source := ordinalFixture(t, "{1}: If this is the second time this ability has resolved this turn, draw a card.")
	def, _ := permanentCardDef(g, source)
	back := compileUnlessCard(t, cardgen.ScryfallCard{
		Name: "Ordinal Back", Layout: "normal", TypeLine: "Artifact",
		OracleText: "{1}: If this is the second time this ability has resolved this turn, you gain 2 life.",
	})
	def.Layout = game.LayoutTransform
	def.Back = opt.Val(back.CardFace)
	first := activateOrdinal(t, g, engine, source.ObjectID, 0, nil, nil)
	engine.resolveTopOfStack(g, &TurnLog{})
	activateOrdinal(t, g, engine, source.ObjectID, 0, nil, nil)
	if !transformPermanent(g, source) {
		t.Fatal("transform failed")
	}
	backObj := activateOrdinal(t, g, engine, source.ObjectID, 0, nil, nil)
	engine.resolveTopOfStack(g, &TurnLog{})
	engine.resolveTopOfStack(g, &TurnLog{})
	if g.Players[game.Player1].Hand.Size() != 1 || g.Players[game.Player1].Life != 40 ||
		first.ActivatedResolutionUse == backObj.ActivatedResolutionUse {
		t.Fatal("captured front body and same-index back body aliased")
	}
}

func TestActivatedOrdinalGrantedIdentitiesSurviveIndexShift(t *testing.T) {
	t.Parallel()
	g, engine, source := ordinalFixture(t, "{1}: If this is the second time this ability has resolved this turn, draw a card.")
	printed, _ := permanentCardDef(g, source)
	granted := printed.ActivatedAbilities[0]
	for range 2 {
		g.ContinuousEffects = append(g.ContinuousEffects, game.ContinuousEffect{
			ID: g.IDGen.Next(), AffectedObjectID: source.ObjectID,
			Layer: game.LayerAbility, Duration: game.DurationUntilEndOfTurn,
			CreatedTurn: g.Turn.TurnNumber, AddAbilities: []game.Ability{&granted},
		})
	}
	firstGrant := activateOrdinal(t, g, engine, source.ObjectID, 1, nil, nil)
	engine.resolveTopOfStack(g, &TurnLog{})
	secondGrant := activateOrdinal(t, g, engine, source.ObjectID, 2, nil, nil)
	engine.resolveTopOfStack(g, &TurnLog{})
	if firstGrant.ActivatedResolutionUse == secondGrant.ActivatedResolutionUse {
		t.Fatal("identical grant templates need distinct granting effect identities")
	}
	g.ContinuousEffects = g.ContinuousEffects[1:]
	shifted := activateOrdinal(t, g, engine, source.ObjectID, 1, nil, nil)
	engine.resolveTopOfStack(g, &TurnLog{})
	activateOrdinal(t, g, engine, source.ObjectID, 0, nil, nil)
	engine.resolveTopOfStack(g, &TurnLog{})
	if shifted.ActivatedResolutionUse != secondGrant.ActivatedResolutionUse ||
		g.Players[game.Player1].Hand.Size() != 1 {
		t.Fatal("removing a grant must not reset/reassign the surviving grant or increment the printed ability")
	}
}

func TestActivatedOrdinalMergedComponentIdentity(t *testing.T) {
	t.Parallel()
	g, engine, source := ordinalFixture(t, "{1}: If this is the second time this ability has resolved this turn, draw a card.")
	def, _ := permanentCardDef(g, source)
	component := addCardInstance(g, game.Player1, def)
	first := activateOrdinal(t, g, engine, source.ObjectID, 0, nil, nil)
	engine.resolveTopOfStack(g, &TurnLog{})
	source.MergedCards = []game.MergedCard{{CardInstanceID: component, Owner: game.Player1}}
	second := activateOrdinal(t, g, engine, source.ObjectID, 1, nil, nil)
	engine.resolveTopOfStack(g, &TurnLog{})
	if first.ActivatedResolutionUse == second.ActivatedResolutionUse {
		t.Fatal("same definition's separate merged components aliased")
	}
	// Replace the top with a card carrying two different abilities, shifting
	// both old components without changing their immutable origins.
	top := compileUnlessCard(t, cardgen.ScryfallCard{
		Name: "New Top", Layout: "normal", TypeLine: "Artifact",
		OracleText: "{1}: You gain 1 life.\n{1}: Draw a card.",
	})
	source.MergedCards = append([]game.MergedCard{{CardInstanceID: source.CardInstanceID, Owner: game.Player1}}, source.MergedCards...)
	source.CardInstanceID = addCardInstance(g, game.Player1, top)
	for _, activation := range []struct {
		index int
		use   opt.V[game.ActivatedAbilityResolutionUse]
		want  int
	}{{2, first.ActivatedResolutionUse, 1}, {3, second.ActivatedResolutionUse, 2}} {
		obj := activateOrdinal(t, g, engine, source.ObjectID, activation.index, nil, nil)
		engine.resolveTopOfStack(g, &TurnLog{})
		if obj.ActivatedResolutionUse != activation.use || g.Players[game.Player1].Hand.Size() != activation.want {
			t.Fatal("mutable effective index replaced captured component identity")
		}
	}
}

func TestActivatedOrdinalCardZoneIncarnationsAndCost(t *testing.T) {
	t.Parallel()
	g, engine, source := ordinalFixture(t, "{1}: If this is the second time this ability has resolved this turn, draw a card.")
	def, _ := permanentCardDef(g, source)
	body := &def.ActivatedAbilities[0]
	body.ZoneOfFunction = zone.Hand
	body.AdditionalCosts = []cost.Additional{{Kind: cost.AdditionalDiscard, Source: zone.Hand, Amount: 1}}
	if !movePermanentToZone(g, source, zone.Hand) {
		t.Fatal("move to hand failed")
	}
	card, _ := g.GetCardInstance(source.CardInstanceID)
	first := activateOrdinal(t, g, engine, card.ID, 0, nil, nil)
	if card.ZoneVersion == first.SourceZoneVersion || !g.Players[game.Player1].Graveyard.Contains(card.ID) {
		t.Fatal("discard cost must change zones after capturing source incarnation")
	}
	g.Stack.Push(game.NewStackObjectCopy(first, g.IDGen.Next()))
	engine.resolveTopOfStack(g, &TurnLog{})
	engine.resolveTopOfStack(g, &TurnLog{})
	if g.Players[game.Player1].Hand.Size() != 1 {
		t.Fatal("discarded source and copy must share the pre-cost incarnation")
	}
	if !moveCardBetweenZones(g, game.Player1, card.ID, zone.Graveyard, zone.Hand) {
		t.Fatal("return to hand failed")
	}
	second := activateOrdinal(t, g, engine, card.ID, 0, nil, nil)
	engine.resolveTopOfStack(g, &TurnLog{})
	if second.ActivatedResolutionUse == first.ActivatedResolutionUse ||
		g.Players[game.Player1].Hand.Size() != 1 {
		t.Fatal("card reentering hand must not reuse a departed card incarnation's tally")
	}
}

func TestActivatedOrdinalGraveyardCopies(t *testing.T) {
	t.Parallel()
	g, engine, source := ordinalFixture(t, "{1}: If this is the second time this ability has resolved this turn, draw a card.")
	def, _ := permanentCardDef(g, source)
	def.ActivatedAbilities[0].ZoneOfFunction = zone.Graveyard
	if !movePermanentToZone(g, source, zone.Graveyard) {
		t.Fatal("move to graveyard failed")
	}
	first := activateOrdinal(t, g, engine, source.CardInstanceID, 0, nil, nil)
	second := activateOrdinal(t, g, engine, source.CardInstanceID, 0, nil, nil)
	engine.resolveTopOfStack(g, &TurnLog{})
	engine.resolveTopOfStack(g, &TurnLog{})
	if first.ActivatedResolutionUse != second.ActivatedResolutionUse ||
		g.Players[game.Player1].Hand.Size() != 1 {
		t.Fatal("copied announcement bodies must retain the printed graveyard ability identity")
	}
}

func TestActivatedOrdinalMissingContextsFailClosed(t *testing.T) {
	t.Parallel()
	g, engine, source := ordinalFixture(t, "{1}: If this is the second time this ability has resolved this turn, draw a card.")
	obj := activateOrdinal(t, g, engine, source.ObjectID, 0, nil, nil)
	use := obj.ActivatedResolutionUse.Val
	g.ResolvedActivatedAbilitiesThisTurn[use] = 2
	g.ResolvedTriggeredAbilitiesThisTurn[game.TriggeredAbilityUse{SourceID: source.ObjectID, AbilityIndex: 0}] = 2
	for _, invalid := range []*game.StackObject{
		nil,
		{Kind: game.StackActivatedAbility, SourceID: source.ObjectID, AbilityIndex: 0},
		{Kind: game.StackTriggeredAbility, SourceID: source.ObjectID, AbilityIndex: 0},
		{Kind: game.StackSpell, SourceID: source.ObjectID, ResolutionOrdinalThisTurn: 2},
		{Kind: game.StackTriggeredAbility, AbilityIndex: 0, ResolutionOrdinalThisTurn: 2},
		{Kind: game.StackTriggeredAbility, SourceID: source.ObjectID, AbilityIndex: -1, ResolutionOrdinalThisTurn: 2},
	} {
		if sourceAbilityResolutionOrdinalMatches(g, conditionContext{obj: invalid}, 2) {
			t.Fatalf("invalid context admitted: %#v", invalid)
		}
	}
	obj.ActivatedResolutionUse = opt.V[game.ActivatedAbilityResolutionUse]{}
	obj, _ = g.Stack.Pop()
	if got := engine.resolveStackObject(g, obj, &TurnLog{}); got != "missing resolution identity" {
		t.Fatalf("missing captured identity resolution=%q", got)
	}
	if g.Players[game.Player1].Hand.Size() != 0 {
		t.Fatal("missing identity must not use another shell's tally")
	}
}

func TestActivatedAndTriggeredOrdinalSameIndexIsolation(t *testing.T) {
	t.Parallel()
	g, engine, source := ordinalFixture(t, "{1}: If this is the second time this ability has resolved this turn, draw a card.")
	def, _ := permanentCardDef(g, source)
	trigger := &game.TriggeredAbility{
		CountsResolutionsThisTurn: true, Content: def.ActivatedAbilities[0].Content,
	}
	for i := range 2 {
		activateOrdinal(t, g, engine, source.ObjectID, 0, nil, nil)
		engine.resolveTopOfStack(g, &TurnLog{})
		g.Stack.Push(&game.StackObject{
			ID: g.IDGen.Next(), Kind: game.StackTriggeredAbility,
			SourceID: source.ObjectID, SourceCardID: source.CardInstanceID,
			Controller: game.Player1, AbilityIndex: 0, InlineTrigger: trigger,
		})
		engine.resolveTopOfStack(g, &TurnLog{})
		want := 0
		if i == 1 {
			want = 2
		}
		if g.Players[game.Player1].Hand.Size() != want {
			t.Fatal("same source/index in separate shells aliased resolution counts")
		}
	}
}

func TestActivatedOrdinalManaRoutesRemainSeparate(t *testing.T) {
	t.Parallel()
	g, engine, source := ordinalFixture(t,
		"{1}: If this is the second time this ability has resolved this turn, draw a card.\n{T}: Add {G}.")
	def, _ := permanentCardDef(g, source)
	if !engine.applyAction(g, game.Player1, action.ActivateAbility(source.ObjectID, def.ManaAbilityIndex(0), nil, 0)) ||
		!g.Stack.IsEmpty() || len(g.ResolvedActivatedAbilitiesThisTurn) != 0 ||
		g.Players[game.Player1].ManaPool.Amount(mana.G) != 1 {
		t.Fatal("true mana ability must resolve immediately without counting another ability")
	}
	activateOrdinal(t, g, engine, source.ObjectID, 0, nil, nil)
	engine.resolveTopOfStack(g, &TurnLog{})
	if g.Players[game.Player1].Hand.Size() != 0 {
		t.Fatal("mana resolution incremented ordinary ability's ordinal")
	}
}

func TestInnerFlameIgniterActualThirdResolution(t *testing.T) {
	t.Parallel()
	def := compileUnlessCard(t, cardgen.ScryfallCard{
		Name: "Inner-Flame Igniter", Layout: "normal", TypeLine: "Creature - Elemental Warrior",
		Power: new("2"), Toughness: new("2"),
		OracleText: "{2}{R}: Creatures you control get +1/+0 until end of turn. If this is the third time this ability has resolved this turn, creatures you control gain first strike until end of turn.",
	})
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	source := addCombatPermanent(g, game.Player1, def)
	opponent := addCombatPermanent(g, game.Player2, def)
	g.Players[game.Player1].ManaPool.Add(mana.R, 9)
	engine := NewEngine(nil)
	for range 3 {
		activateOrdinal(t, g, engine, source.ObjectID, 0, nil, nil)
	}
	for i := range 3 {
		engine.resolveTopOfStack(g, &TurnLog{})
		if got := effectivePermanentValues(g, source).power; got != 3+i {
			t.Fatalf("resolution %d: power=%d, want %d", i+1, got, 3+i)
		}
		if hasKeyword(g, source, game.FirstStrike) != (i == 2) ||
			hasKeyword(g, opponent, game.FirstStrike) ||
			effectivePermanentValues(g, opponent).power != 2 {
			t.Fatal("actual third-resolution grant or controller isolation failed")
		}
	}
}
