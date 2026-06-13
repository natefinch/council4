package compiler

import (
	"github.com/natefinch/council4/cardgen/oracle/parser"
	"github.com/natefinch/council4/cardgen/oracle/shared"
)

type pipelineContext struct {
	CardName         string
	InstantOrSorcery bool
	Planeswalker     bool
	Saga             bool
}

type cachedParserCard struct {
	Name       string       `json:"name"`
	OracleText string       `json:"oracle_text"`
	CardFaces  []cachedFace `json:"card_faces"`
	TypeLine   string       `json:"type_line"`
}

type cachedFace struct {
	Name       string `json:"name"`
	TypeLine   string `json:"type_line"`
	OracleText string `json:"oracle_text"`
}

func compileSource(source string, context pipelineContext) (Compilation, []shared.Diagnostic) {
	document, diagnostics := parser.Parse(source, parser.Context{
		InstantOrSorcery: context.InstantOrSorcery,
		Planeswalker:     context.Planeswalker,
		Saga:             context.Saga,
	})
	compilation, compilerDiagnostics := Compile(document, Context{CardName: context.CardName})
	return compilation, append(diagnostics, compilerDiagnostics...)
}
