package lexer

type Token struct {
	Kind  TokenKind
	Lexem string
	Line  int
	Col   int
}

type TokenKind int

const (
	// EOF TokenKind = iota
	IDENT            TokenKind = iota
	PLUS                       // +
	MINUS                      // -
	STAR                       // *
	SLASH                      // /
	OPEN_PAREN                 // (
	CLOSE_PAREN                // )
	OPEN_BRACE                 // {
	CLOSE_BRACE                // }
	QUESTION_MARK              // ?
	EXCLAMATION_MARK           // !
	EQ                         // =
	EQEQ                       // ==
	LESS                       // <
	GREATER                    // >
	LESS_EQ                    // <=
	GREATER_EQ                 // >=
	BANG_EQ                    // !=
	COLON                      // :
	COLON_EQ                   // :=
	COMMA                      // ,
	SEMICOLON                  // ;
	HASHTAG                    // #
	DOT                        // .
	OPEN_BRACKET               // [
	CLOSE_BRACKET              // ]
	PLUS_PLUS                  // ++
	INT                        // 123
	FLOAT                      // 3.14
)
