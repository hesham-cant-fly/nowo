package lexer

type scanner struct {
	input       string
	start       int
	current     int
	line        int
	lastNewline int
}

func Tokenize(input string) []Token {
	s := &scanner{input: input, line: 1, lastNewline: -1}
	var tokens []Token

	for !s.isAtEnd() {
		s.skipWhitespace()
		s.start = s.current

		if s.isAtEnd() {
			break
		}

		c := s.advance()

		switch c {
		case '+':
			if s.peek() == '+' {
				s.advance()
				tokens = append(tokens, s.makeToken(PLUS_PLUS))
			} else {
				tokens = append(tokens, s.makeToken(PLUS))
			}
		case '-':
			if s.peek() == '-' {
				for !s.isAtEnd() && s.peek() != '\n' {
					s.advance()
				}
			} else {
				tokens = append(tokens, s.makeToken(MINUS))
			}
		case '*':
			tokens = append(tokens, s.makeToken(STAR))
		case '/':
			tokens = append(tokens, s.makeToken(SLASH))
		case '(':
			tokens = append(tokens, s.makeToken(OPEN_PAREN))
		case ')':
			tokens = append(tokens, s.makeToken(CLOSE_PAREN))
		case '{':
			tokens = append(tokens, s.makeToken(OPEN_BRACE))
		case '}':
			tokens = append(tokens, s.makeToken(CLOSE_BRACE))
		case '?':
			tokens = append(tokens, s.makeToken(QUESTION_MARK))
		case '!':
			if s.peek() == '=' {
				s.advance()
				tokens = append(tokens, s.makeToken(BANG_EQ))
			} else {
				tokens = append(tokens, s.makeToken(EXCLAMATION_MARK))
			}
		case '=':
			if s.peek() == '=' {
				s.advance()
				tokens = append(tokens, s.makeToken(EQEQ))
			} else {
				tokens = append(tokens, s.makeToken(EQ))
			}
		case '<':
			if s.peek() == '=' {
				s.advance()
				tokens = append(tokens, s.makeToken(LESS_EQ))
			} else {
				tokens = append(tokens, s.makeToken(LESS))
			}
		case '>':
			if s.peek() == '=' {
				s.advance()
				tokens = append(tokens, s.makeToken(GREATER_EQ))
			} else {
				tokens = append(tokens, s.makeToken(GREATER))
			}
		case ':':
			if s.peek() == '=' {
				s.advance()
				tokens = append(tokens, s.makeToken(COLON_EQ))
			} else {
				tokens = append(tokens, s.makeToken(COLON))
			}
		case ',':
			tokens = append(tokens, s.makeToken(COMMA))
		case ';':
			tokens = append(tokens, s.makeToken(SEMICOLON))
		case '#':
			tokens = append(tokens, s.makeToken(HASHTAG))
		case '[':
			tokens = append(tokens, s.makeToken(OPEN_BRACKET))
		case ']':
			tokens = append(tokens, s.makeToken(CLOSE_BRACKET))
		case '.':
			tokens = append(tokens, s.makeToken(DOT))
		default:
			if isAlpha(c) {
				tokens = append(tokens, s.identifier())
			} else if isDigit(c) {
				tokens = append(tokens, s.number())
			}
		}
	}

	// tokens = append(tokens, Token{Kind: EOF, Line: s.line})
	return tokens
}

func (s *scanner) identifier() Token {
	for isAlphaNumeric(s.peek()) {
		s.advance()
	}
	return s.makeToken(IDENT)
}

func (s *scanner) number() Token {
	for isDigit(s.peek()) {
		s.advance()
	}

	if s.peek() == '.' && isDigit(s.peekNext()) {
		s.advance()
		for isDigit(s.peek()) {
			s.advance()
		}
		return s.makeToken(FLOAT)
	}

	return s.makeToken(INT)
}

func (s *scanner) skipWhitespace() {
	for {
		c := s.peek()
		switch c {
		case ' ', '\r', '\t':
			s.advance()
		case '\n':
			s.line++
			s.lastNewline = s.current
			s.advance()
		case 0:
			return
		default:
			return
		}
	}
}

func (s *scanner) peek() byte {
	if s.isAtEnd() {
		return 0
	}
	return s.input[s.current]
}

func (s *scanner) peekNext() byte {
	if s.current+1 >= len(s.input) {
		return 0
	}
	return s.input[s.current+1]
}

func (s *scanner) advance() byte {
	c := s.input[s.current]
	s.current++
	return c
}

func (s *scanner) isAtEnd() bool {
	return s.current >= len(s.input)
}

func (s *scanner) makeToken(kind TokenKind) Token {
	return Token{
		Kind:  kind,
		Lexem: s.input[s.start:s.current],
		Line:  s.line,
		Col:   s.start - s.lastNewline,
	}
}

func isAlpha(c byte) bool {
	return (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || c == '_'
}

func isDigit(c byte) bool {
	return c >= '0' && c <= '9'
}

func isAlphaNumeric(c byte) bool {
	return isAlpha(c) || (c >= '0' && c <= '9')
}
