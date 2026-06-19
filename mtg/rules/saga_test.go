package rules

import (
	"fmt"
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
)

func TestSagaEntersWithLoreCounterAndTriggersFirstChapter(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	card := addSagaCardInstance(g, game.Player1, []game.ChapterAbility{
		sagaDrawChapter(1),
		sagaDrawChapter(2),
	})
	addCardToLibrary(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Drawn"}})

	permanent, ok := createCardPermanentFace(g, card, game.Player1, zone.Stack, game.FaceFront)
	if !ok {
		t.Fatal("create Saga permanent failed")
	}
	if got := permanent.Counters.Get(counter.Lore); got != 1 {
		t.Fatalf("lore counters = %d, want 1", got)
	}
	if !engine.putTriggeredAbilitiesOnStack(g) {
		t.Fatal("first chapter was not put on the stack")
	}
	object, ok := g.Stack.Peek()
	if !ok || !object.SagaChapter || object.SourceID != permanent.ObjectID {
		t.Fatalf("stack object = %+v, want first Saga chapter", object)
	}

	engine.resolveTopOfStack(g, &TurnLog{})
	if got := g.Players[game.Player1].Hand.Size(); got != 1 {
		t.Fatalf("hand size = %d, want first chapter to draw", got)
	}
	if _, ok := permanentByObjectID(g, permanent.ObjectID); !ok {
		t.Fatal("Saga was sacrificed before its final chapter")
	}
}

func TestSagaCrossesMultipleChaptersAndSacrificesAfterTheyLeaveStack(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	permanent := addCombatPermanent(g, game.Player1, sagaCard([]game.ChapterAbility{
		sagaDrawChapter(1),
		sagaDrawChapter(2),
		sagaDrawChapter(3),
	}))
	permanent.Counters.Add(counter.Lore, 1)
	addCardToLibrary(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "First"}})
	addCardToLibrary(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Second"}})

	addCountersToPermanent(g, permanent, counter.Lore, 2)
	if !engine.putTriggeredAbilitiesOnStack(g) {
		t.Fatal("crossed chapters were not put on the stack")
	}
	if got := g.Stack.Size(); got != 2 {
		t.Fatalf("stack size = %d, want two crossed chapter abilities", got)
	}
	engine.applyStateBasedActions(g)
	if _, ok := permanentByObjectID(g, permanent.ObjectID); !ok {
		t.Fatal("Saga was sacrificed while chapter abilities remained on the stack")
	}

	engine.resolveTopOfStack(g, &TurnLog{})
	engine.applyStateBasedActions(g)
	if _, ok := permanentByObjectID(g, permanent.ObjectID); !ok {
		t.Fatal("Saga was sacrificed while one chapter ability remained on the stack")
	}
	engine.resolveTopOfStack(g, &TurnLog{})
	engine.applyStateBasedActions(g)
	if _, ok := permanentByObjectID(g, permanent.ObjectID); ok {
		t.Fatal("Saga remained after its final chapter ability left the stack")
	}
	if got := g.Players[game.Player1].Hand.Size(); got != 2 {
		t.Fatalf("hand size = %d, want both crossed chapters to resolve", got)
	}
}

func TestSagaSharedChapterAbilityTriggersOnceWhenSeveralNumbersAreCrossed(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	permanent := addCombatPermanent(g, game.Player1, sagaCard([]game.ChapterAbility{{
		Text:     "II, III — Draw a card.",
		Chapters: []int{2, 3},
		Content:  sagaDrawChapter(2).Content,
	}}))
	permanent.Counters.Add(counter.Lore, 1)

	addCountersToPermanent(g, permanent, counter.Lore, 2)
	if !engine.putTriggeredAbilitiesOnStack(g) {
		t.Fatal("shared chapter ability was not put on the stack")
	}
	if got := g.Stack.Size(); got != 1 {
		t.Fatalf("stack size = %d, want shared chapter ability once", got)
	}
}

func TestSagaChapterResolvesAfterSourceLeavesBattlefield(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	permanent := addCombatPermanent(g, game.Player1, sagaCard([]game.ChapterAbility{
		sagaDrawChapter(1),
	}))
	addCardToLibrary(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Drawn"}})

	addCountersToPermanent(g, permanent, counter.Lore, 1)
	if !engine.putTriggeredAbilitiesOnStack(g) {
		t.Fatal("chapter ability was not put on the stack")
	}
	if !movePermanentToZone(g, permanent, zone.Graveyard) {
		t.Fatal("moving Saga to graveyard failed")
	}
	engine.resolveTopOfStack(g, &TurnLog{})

	if got := g.Players[game.Player1].Hand.Size(); got != 1 {
		t.Fatalf("hand size = %d, want chapter to resolve from inline source data", got)
	}
}

func TestAdvanceSagasAddsLoreOnlyToActivePlayersSagas(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	first := addCombatPermanent(g, game.Player1, sagaCard([]game.ChapterAbility{sagaDrawChapter(2)}))
	second := addCombatPermanent(g, game.Player2, sagaCard([]game.ChapterAbility{sagaDrawChapter(2)}))
	first.Counters.Add(counter.Lore, 1)
	second.Counters.Add(counter.Lore, 1)

	advanceSagas(g, game.Player1)

	if got := first.Counters.Get(counter.Lore); got != 2 {
		t.Fatalf("active player's lore counters = %d, want 2", got)
	}
	if got := second.Counters.Get(counter.Lore); got != 1 {
		t.Fatalf("nonactive player's lore counters = %d, want 1", got)
	}
}

func TestReadAheadChoosesEntryChapterAndSkipsEarlierChapters(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	card := addSagaCardInstance(g, game.Player2, []game.ChapterAbility{
		sagaNamedChapter(1),
		sagaNamedChapter(2),
		sagaNamedChapter(3),
		sagaNamedChapter(4),
	})
	card.Def.StaticAbilities = []game.StaticAbility{game.ReadAheadStaticBody}
	agent := &choiceOnlyAgent{choices: [][]int{{3}}}
	agents := [game.NumPlayers]PlayerAgent{game.Player2: agent}
	log := &TurnLog{}

	permanent, ok := createCardPermanentFaceWithChoices(engine, g, card, game.Player2, zone.Stack, game.FaceFront, agents, log)
	if !ok {
		t.Fatal("create Read ahead Saga failed")
	}
	if got := permanent.Counters.Get(counter.Lore); got != 3 {
		t.Fatalf("lore counters = %d, want 3", got)
	}
	if got := permanent.SagaEntryChapter; got != 3 {
		t.Fatalf("SagaEntryChapter = %d, want 3", got)
	}
	if len(log.Choices) != 1 {
		t.Fatalf("choices = %+v, want one Read ahead choice", log.Choices)
	}
	choice := log.Choices[0]
	if choice.Request.Player != game.Player2 ||
		choice.Request.Kind != game.ChoiceResolution ||
		len(choice.Request.Options) != 4 ||
		choice.Request.DefaultSelection[0] != 1 ||
		len(choice.Selected) != 1 ||
		choice.Selected[0] != 3 ||
		choice.UsedFallback {
		t.Fatalf("choice = %+v, want Player 2 choosing chapter 3", choice)
	}

	if !engine.putTriggeredAbilitiesOnStack(g) {
		t.Fatal("chosen chapter was not put on the stack")
	}
	assertSagaChapterOnTop(t, g, "Chapter 3")
	engine.resolveTopOfStack(g, &TurnLog{})

	advanceSagas(g, game.Player2)
	if !engine.putTriggeredAbilitiesOnStack(g) {
		t.Fatal("next chapter was not put on the stack")
	}
	assertSagaChapterOnTop(t, g, "Chapter 4")
}

func TestReadAheadFallbackChoosesFirstChapter(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	card := addSagaCardInstance(g, game.Player1, []game.ChapterAbility{
		sagaNamedChapter(1),
		sagaNamedChapter(2),
	})
	card.Def.StaticAbilities = []game.StaticAbility{game.ReadAheadStaticBody}
	log := &TurnLog{}

	permanent, ok := createCardPermanentFaceWithChoices(engine, g, card, game.Player1, zone.Stack, game.FaceFront, [game.NumPlayers]PlayerAgent{}, log)
	if !ok {
		t.Fatal("create Read ahead Saga failed")
	}
	if got := permanent.Counters.Get(counter.Lore); got != 1 {
		t.Fatalf("lore counters = %d, want 1", got)
	}
	if got := permanent.SagaEntryChapter; got != 1 {
		t.Fatalf("SagaEntryChapter = %d, want 1", got)
	}
	if len(log.Choices) != 1 || !log.Choices[0].UsedFallback || log.Choices[0].Selected[0] != 1 {
		t.Fatalf("choices = %+v, want chapter 1 fallback", log.Choices)
	}
	if !engine.putTriggeredAbilitiesOnStack(g) {
		t.Fatal("first chapter was not put on the stack")
	}
	assertSagaChapterOnTop(t, g, "Chapter 1")
}

func TestReadAheadSingleChapterWaitsForPendingTriggerBeforeSacrifice(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	card := addSagaCardInstance(g, game.Player1, []game.ChapterAbility{sagaNamedChapter(1)})
	card.Def.StaticAbilities = []game.StaticAbility{game.ReadAheadStaticBody}

	permanent, ok := createCardPermanentFace(g, card, game.Player1, zone.Stack, game.FaceFront)
	if !ok {
		t.Fatal("create Read ahead Saga failed")
	}
	if permanent.SagaEntryChapter != 1 {
		t.Fatalf("SagaEntryChapter = %d, want 1", permanent.SagaEntryChapter)
	}
	engine.applyStateBasedActions(g)
	if _, ok := permanentByObjectID(g, permanent.ObjectID); !ok {
		t.Fatal("Saga was sacrificed before its chapter trigger reached the stack")
	}
	if !engine.putTriggeredAbilitiesOnStack(g) {
		t.Fatal("final chapter was not put on the stack")
	}
	engine.applyStateBasedActions(g)
	if _, ok := permanentByObjectID(g, permanent.ObjectID); !ok {
		t.Fatal("Saga was sacrificed while its final chapter was on the stack")
	}
	engine.resolveTopOfStack(g, &TurnLog{})
	engine.applyStateBasedActions(g)
	if _, ok := permanentByObjectID(g, permanent.ObjectID); ok {
		t.Fatal("Saga remained after its final chapter left the stack")
	}
}

func TestReadAheadSkippedChaptersRemainSkippedAfterLoreCountersAreRemoved(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	card := addSagaCardInstance(g, game.Player1, []game.ChapterAbility{
		sagaNamedChapter(1),
		sagaNamedChapter(2),
		sagaNamedChapter(3),
	})
	card.Def.StaticAbilities = []game.StaticAbility{game.ReadAheadStaticBody}
	agent := &choiceOnlyAgent{choices: [][]int{{3}}}
	agents := [game.NumPlayers]PlayerAgent{game.Player1: agent}

	permanent, ok := createCardPermanentFaceWithChoices(engine, g, card, game.Player1, zone.Stack, game.FaceFront, agents, &TurnLog{})
	if !ok {
		t.Fatal("create Read ahead Saga failed")
	}
	if !engine.putTriggeredAbilitiesOnStack(g) {
		t.Fatal("chosen chapter was not put on the stack")
	}
	engine.resolveTopOfStack(g, &TurnLog{})

	permanent.Counters.Remove(counter.Lore, 3)
	addCountersToPermanent(g, permanent, counter.Lore, 2)
	if engine.putTriggeredAbilitiesOnStack(g) {
		t.Fatal("skipped chapters triggered after lore counters were removed")
	}
}

func TestReadAheadUsesEffectiveEntryCharacteristics(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	card := &game.CardInstance{
		ID:    g.IDGen.Next(),
		Owner: game.Player1,
		Def: &game.CardDef{CardFace: game.CardFace{
			Name:  "Copying Permanent",
			Types: []types.Card{types.Enchantment},
		}},
	}
	g.CardInstances[card.ID] = card
	ch1, ch2, ch3 := sagaNamedChapter(1), sagaNamedChapter(2), sagaNamedChapter(3)
	chapters := []game.Ability{
		&ch1,
		&ch2,
		&ch3,
	}
	continuous := []game.ContinuousEffect{
		{
			Layer:       game.LayerType,
			AddSubtypes: []types.Sub{types.Saga},
		},
		{
			Layer:        game.LayerAbility,
			AddKeywords:  []game.Keyword{game.ReadAhead},
			AddAbilities: chapters,
		},
	}
	agent := &choiceOnlyAgent{choices: [][]int{{3}}}
	agents := [game.NumPlayers]PlayerAgent{game.Player1: agent}

	permanent, ok := createCardPermanentFaceWithContinuous(engine, g, card, game.Player1, zone.Stack, game.FaceFront, continuous, agents, &TurnLog{})
	if !ok {
		t.Fatal("create effective Read ahead Saga failed")
	}
	if got := permanent.Counters.Get(counter.Lore); got != 3 {
		t.Fatalf("lore counters = %d, want 3", got)
	}
	if !engine.putTriggeredAbilitiesOnStack(g) {
		t.Fatal("effective chosen chapter was not put on the stack")
	}
	assertSagaChapterOnTop(t, g, "Chapter 3")
}

func TestReadAheadSagaTokenChoosesEntryChapter(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	token := sagaCard([]game.ChapterAbility{
		sagaNamedChapter(1),
		sagaNamedChapter(2),
		sagaNamedChapter(3),
	})
	token.StaticAbilities = []game.StaticAbility{game.ReadAheadStaticBody}
	agent := &choiceOnlyAgent{choices: [][]int{{3}}}
	agents := [game.NumPlayers]PlayerAgent{game.Player1: agent}

	permanent, ok := createTokenPermanentWithChoices(engine, g, game.Player1, token, agents, &TurnLog{})
	if !ok {
		t.Fatal("create Read ahead Saga token failed")
	}
	if got := permanent.Counters.Get(counter.Lore); got != 3 {
		t.Fatalf("lore counters = %d, want 3", got)
	}
	if !engine.putTriggeredAbilitiesOnStack(g) {
		t.Fatal("token's chosen chapter was not put on the stack")
	}
	assertSagaChapterOnTop(t, g, "Chapter 3")
}

func addSagaCardInstance(g *game.Game, owner game.PlayerID, chapters []game.ChapterAbility) *game.CardInstance {
	card := &game.CardInstance{
		ID:    g.IDGen.Next(),
		Owner: owner,
		Def:   sagaCard(chapters),
	}
	g.CardInstances[card.ID] = card
	return card
}

func sagaCard(chapters []game.ChapterAbility) *game.CardDef {
	return &game.CardDef{CardFace: game.CardFace{
		Name:             "Test Saga",
		Types:            []types.Card{types.Enchantment},
		Subtypes:         []types.Sub{types.Saga},
		ChapterAbilities: chapters,
	}}
}

func sagaDrawChapter(number int) game.ChapterAbility {
	return game.ChapterAbility{
		Text:     "Draw a card.",
		Chapters: []int{number},
		Content: game.Mode{Sequence: []game.Instruction{{
			Primitive: game.Draw{Amount: game.Fixed(1), Player: game.ControllerReference()},
		}}}.Ability(),
	}
}

func sagaNamedChapter(number int) game.ChapterAbility {
	chapter := sagaDrawChapter(number)
	chapter.Text = fmt.Sprintf("Chapter %d", number)
	return chapter
}

func assertSagaChapterOnTop(t *testing.T, g *game.Game, text string) {
	t.Helper()
	if got := g.Stack.Size(); got != 1 {
		t.Fatalf("stack size = %d, want one Saga chapter", got)
	}
	object, ok := g.Stack.Peek()
	if !ok || !object.SagaChapter || object.InlineTrigger == nil || object.InlineTrigger.Text != text {
		t.Fatalf("stack object = %+v, want Saga chapter %q", object, text)
	}
}
