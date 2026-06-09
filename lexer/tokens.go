package lexer

type Token struct {
	Kind  TokenKind
	Lexem string
	Line  int
}

type TokenKind int

const (
	// EOF TokenKind = iota
	IDENT       TokenKind = iota
	PLUS                  // +
	MINUS                 // -
	STAR                  // *
	SLASH                 // /
	OPEN_PAREN            // (
	CLOSE_PAREN           // )
	OPEN_BRACE            // {
	CLOSE_BRACE           // }
	EQ                    // =
	COLON                 // :
	COMMA                 // ,
	SEMICOLON             // ;
	DOT                   // .
	INT                   // 123
	FLOAT                 // 3.14
)
