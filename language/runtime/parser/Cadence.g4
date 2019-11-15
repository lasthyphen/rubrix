grammar Cadence;

// for handling optional semicolons between statement, see also `eos` rule

// NOTE: unusued builder variable, to avoid unused import error because
//    import will also be added to visitor code
@parser::header {
    import "strings"
    var _ = strings.Builder{}
}

@parser::members {
    // Returns true if on the current index of the parser's
    // token stream a token exists on the Hidden channel which
    // either is a line terminator, or is a multi line comment that
    // contains a line terminator.
    func (p *CadenceParser) lineTerminatorAhead() bool {
        // Get the token ahead of the current index.
        possibleIndexEosToken := p.GetCurrentToken().GetTokenIndex() - 1
        ahead := p.GetTokenStream().Get(possibleIndexEosToken)

        if ahead.GetChannel() != antlr.LexerHidden {
            // We're only interested in tokens on the HIDDEN channel.
            return true
        }

        if ahead.GetTokenType() == CadenceParserTerminator {
            // There is definitely a line terminator ahead.
            return true
        }

        if ahead.GetTokenType() == CadenceParserWS {
            // Get the token ahead of the current whitespaces.
            possibleIndexEosToken = p.GetCurrentToken().GetTokenIndex() - 2
            ahead = p.GetTokenStream().Get(possibleIndexEosToken)
        }

        // Get the token's text and type.
        text := ahead.GetText()
        _type := ahead.GetTokenType()

        // Check if the token is, or contains a line terminator.
        return (_type == CadenceParserBlockComment && (strings.Contains(text, "\r") || strings.Contains(text, "\n"))) ||
            (_type == CadenceParserTerminator)
    }

    func (p *CadenceParser) noWhitespace() bool {
    	index := p.GetCurrentToken().GetTokenIndex()
    	return p.GetTokenStream().Get(index-1).GetTokenType() != CadenceParserWS
    }
}

program
    : (declaration ';'?)* EOF
    ;

replInput
    : replElement* EOF
    ;

replElement
    : replDeclaration
    | replStatement
    ;

replStatement
    : statement eos
    ;

replDeclaration
    : declaration ';'?
    ;

declaration
    : compositeDeclaration
    | interfaceDeclaration
    | functionDeclaration[true]
    | variableDeclaration
    | importDeclaration
    | eventDeclaration
    ;

importDeclaration
    : Import (identifier (',' identifier)* From)? (stringLiteral | HexadecimalLiteral)
    ;

access
    : /* Not specified */
    | Pub
    | PubSet
    ;

compositeDeclaration
    : access compositeKind identifier conformances '{' members[true] '}'
    ;

conformances
    : (':' identifier (',' identifier)*)?
    ;

variableKind
    : Let
    | Var
    ;

field
    : access variableKind? identifier ':' typeAnnotation
    ;

interfaceDeclaration
    : access compositeKind Interface identifier '{' members[false] '}'
    ;

members[bool functionBlockRequired]
    : (member[functionBlockRequired] ';'?)*
    ;

member[bool functionBlockRequired]
    : field
    | specialFunctionDeclaration[functionBlockRequired]
    | functionDeclaration[functionBlockRequired]
    | interfaceDeclaration
    | compositeDeclaration
    ;

compositeKind
    : Struct
    | Resource
    | Contract
    ;

// specialFunctionDeclaration is the rule for special function declarations,
// i.e., those that don't require a `fun` keyword and don't have a return type,
// e.g. initializers (`init`) and destructors (`destroy`).
//
// NOTE: allow any identifier in parser, then check identifier is one of
// the valid identifiers in the semantic analysis to provide better error
//
specialFunctionDeclaration[bool functionBlockRequired]
    : identifier parameterList
      // only optional if parameter functionBlockRequired is false
      b=functionBlock? { !$functionBlockRequired || $ctx.b != nil }?
    ;

functionDeclaration[bool functionBlockRequired]
    : access Fun identifier parameterList (':' returnType=typeAnnotation)?
      // only optional if parameter functionBlockRequired is false
      b=functionBlock? { !$functionBlockRequired || $ctx.b != nil }?
    ;

eventDeclaration
    : Event identifier parameterList
    ;

parameterList
    : '(' (parameter (',' parameter)*)? ')'
    ;

parameter
    : (argumentLabel=identifier)? parameterName=identifier ':' typeAnnotation
    ;

typeAnnotation
    : Move? fullType
    ;

// NOTE: only allow reference or optionals – prevent ambiguous
// and not particular useful types like `&R?`
fullType
    : reference=Ampersand {p.noWhitespace()}? baseType
    | baseType ({p.noWhitespace()}? optionals+=Optional)*
    ;

baseType
    : nominalType
    | functionType
    | variableSizedType
    | constantSizedType
    | dictionaryType
    ;

nominalType
    : identifier
    ;

functionType
    : '('
        '(' (parameterTypes+=typeAnnotation (',' parameterTypes+=typeAnnotation)*)? ')'
        ':' returnType=typeAnnotation
      ')'
    ;

variableSizedType
    : '[' fullType ']'
    ;

constantSizedType
    : '[' fullType ';' size=DecimalLiteral ']'
    ;

dictionaryType
    : '{' keyType=fullType ':' valueType=fullType '}'
    ;

block
    : '{' statements '}'
    ;

functionBlock
    : '{' preConditions? postConditions? statements '}'
    ;

preConditions
    : Pre '{' conditions '}'
    ;

postConditions
    : Post '{' conditions '}'
    ;

conditions
    : (condition eos)*
    ;

condition
    : test=expression (':' message=expression)?
    ;

statements
    : (statement eos)*
    ;

// NOTE: important to have expression last
statement
    : returnStatement
    | breakStatement
    | continueStatement
    | ifStatement
    | whileStatement
    | emitStatement
    // NOTE: allow all declarations, even structures, in parser,
    // then check identifier declaration is variable/constant or function
    // in semantic analysis to provide better error
    | declaration
    | assignment
    | swap
    | expression
    ;

// only parse the return value expression if it is
// on the same line. this prevents the return statement
// from greedily taking an expression from a potentialy
// following expression statement
//
returnStatement
    : Return ({!p.lineTerminatorAhead()}? expression)?
    ;

breakStatement
    : Break
    ;

continueStatement
    : Continue
    ;

ifStatement
    : If
      (testExpression=expression | testDeclaration=variableDeclaration)
      then=block
      (Else (ifStatement | alt=block))?
    ;

whileStatement
    : While expression block
    ;

emitStatement
    : Emit identifier invocation
    ;

// Variable declarations might be of the form `let|var <- x <- y`
//
variableDeclaration
    : variableKind identifier (':' typeAnnotation)?
      leftTransfer=transfer leftExpression=expression
      (rightTransfer=transfer rightExpression=expression)?
    ;

// NOTE: we allow any kind of transfer, i.e. moves, but ensure
//   that move is not used in the semantic analysis (as assignment
//   to resource type will cause a loss of the old value).
//   Being unrestritive here allows us to provide better error messages
//   in the semantic analysis.
assignment
    : identifier expressionAccess* transfer expression
    ;


// NOTE: we allow expressions on both sides when parsing, but ensure
//   that both sides are targets (identifier, member access, or index access)
//   in the semantic analysis. This allows us to provide better error messages
swap
    : left=expression '<->' right=expression
    ;

transfer
    : '='
    | Move
    ;

expression
    : conditionalExpression
    ;

conditionalExpression
    : <assoc=right> orExpression ('?' then=expression ':' alt=expression)?
    ;

orExpression
    : andExpression
    | orExpression '||' andExpression
    ;

andExpression
    : equalityExpression
    | andExpression '&&' equalityExpression
    ;

equalityExpression
    : relationalExpression
    | equalityExpression equalityOp relationalExpression
    ;

relationalExpression
    : nilCoalescingExpression
    | relationalExpression relationalOp nilCoalescingExpression
    ;

nilCoalescingExpression
    // NOTE: right associative
    : failableDowncastingExpression (NilCoalescing nilCoalescingExpression)?
    ;

failableDowncastingExpression
    : concatenatingExpression
    | failableDowncastingExpression FailableDowncasting typeAnnotation
    ;

concatenatingExpression
    : additiveExpression
    | concatenatingExpression Ampersand additiveExpression
    ;

additiveExpression
    : multiplicativeExpression
    | additiveExpression additiveOp multiplicativeExpression
    ;

multiplicativeExpression
    : unaryExpression
    | multiplicativeExpression multiplicativeOp unaryExpression
    ;

unaryExpression
    : primaryExpression
    // NOTE: allow multiple unary operators, but reject in visitor
    // to provide better error for invalid juxtaposition
    | unaryOp+ unaryExpression
    ;

primaryExpression
    : createExpression
    | destroyExpression
    | referenceExpression
    | composedExpression
    ;

composedExpression
    : primaryExpressionStart primaryExpressionSuffix*
    ;

primaryExpressionSuffix
    : expressionAccess
    | invocation
    ;

equalityOp
    : Equal
    | Unequal
    ;

Equal : '==' ;
Unequal : '!=' ;

relationalOp
    : Less
    | Greater
    | LessEqual
    | GreaterEqual
    ;

Less : '<' ;
Greater : '>' ;
LessEqual : '<=' ;
GreaterEqual : '>=' ;

additiveOp
    : Plus
    | Minus
    ;

Plus : '+' ;
Minus : '-' ;

multiplicativeOp
    : Mul
    | Div
    | Mod
    ;

Mul : '*' ;
Div : '/' ;
Mod : '%' ;

Ampersand : '&';

unaryOp
    : Minus
    | Negate
    | Move
    ;

Negate : '!' ;
Move : '<-' ;

Optional : '?' ;

NilCoalescing : WS '??';

Downcasting : 'as' ;

FailableDowncasting : 'as?' ;

primaryExpressionStart
    : identifierExpression
    | literalExpression
    | functionExpression
    | nestedExpression
    ;

createExpression
    : Create identifier invocation
    ;

destroyExpression
    : Destroy expression
    ;

referenceExpression
    : Ampersand expression Downcasting fullType
    ;

identifierExpression
    : identifier
    ;

literalExpression
    : literal
    ;

functionExpression
    : Fun parameterList (':' returnType=typeAnnotation)? functionBlock
    ;

nestedExpression
    : '(' expression ')'
    ;

expressionAccess
    : memberAccess
    | bracketExpression
    ;

memberAccess
    : '.' identifier
    ;

bracketExpression
    // storage is accessed using types
    : '[' (expression | fullType) ']'
    ;

invocation
    : '(' (argument (',' argument)*)? ')'
    ;

argument
    : (identifier ':')? expression
    ;

literal
    : integerLiteral
    | booleanLiteral
    | arrayLiteral
    | dictionaryLiteral
    | stringLiteral
    | nilLiteral
    ;

booleanLiteral
    : True
    | False
    ;

nilLiteral
    : Nil
    ;

stringLiteral
    : StringLiteral
    ;

integerLiteral
    : Minus? positiveIntegerLiteral
    ;

positiveIntegerLiteral
    : DecimalLiteral        # DecimalLiteral
    | BinaryLiteral         # BinaryLiteral
    | OctalLiteral          # OctalLiteral
    | HexadecimalLiteral    # HexadecimalLiteral
    | InvalidNumberLiteral  # InvalidNumberLiteral
    ;

arrayLiteral
    : '[' ( expression (',' expression)* )? ']'
    ;

dictionaryLiteral
    : '{' ( dictionaryEntry (',' dictionaryEntry)* )? '}'
    ;

dictionaryEntry
    : key=expression ':' value=expression
    ;

OpenParen: '(' ;
CloseParen: ')' ;

Transaction : 'transaction' ;

Struct : 'struct' ;
Resource : 'resource' ;
Contract : 'contract' ;

Interface : 'interface' ;

Fun : 'fun' ;

Event : 'event' ;
Emit : 'emit' ;

Pre : 'pre' ;
Post : 'post' ;

Pub : 'pub' ;
PubSet : 'pub(set)' ;

Return : 'return' ;

Break : 'break' ;
Continue : 'continue' ;

Let : 'let' ;
Var : 'var' ;

If : 'if' ;
Else : 'else' ;

While : 'while' ;

True : 'true' ;
False : 'false' ;

Nil : 'nil' ;

Import : 'import' ;
From : 'from' ;

Create : 'create' ;
Destroy : 'destroy' ;

identifier
    : Identifier
    | From
    | Create
    | Destroy
    ;

Identifier
    : IdentifierHead IdentifierCharacter*
    ;

fragment IdentifierHead
    : [a-zA-Z]
    |  '_'
    ;

fragment IdentifierCharacter
    : [0-9]
    | IdentifierHead
    ;


DecimalLiteral
    // NOTE: allows trailing underscores, but the parser checks underscores
    // only occur inside, to provide better syntax errors
    : [0-9] [0-9_]*
    ;


BinaryLiteral
    // NOTE: allows underscores anywhere after prefix, but the parser checks underscores
    // only occur inside, to provide better syntax errors
    : '0b' [01_]+
    ;


OctalLiteral
    // NOTE: allows underscores anywhere after prefix, but the parser checks underscores
    // only occur inside, to provide better syntax errors
    : '0o' [0-7_]+
    ;

HexadecimalLiteral
    // NOTE: allows underscores anywhere after prefix, but the parser checks underscores
    // only occur inside, to provide better syntax errors
    : '0x' [0-9a-fA-F_]+
    ;

// NOTE: invalid literal, to provide better syntax errors
InvalidNumberLiteral
    : '0' [a-zA-Z] [0-9a-zA-Z_]*
    ;

StringLiteral
    : '"' QuotedText* '"'
    ;

fragment QuotedText
    : EscapedCharacter
    | ~["\n\r\\]
    ;

fragment EscapedCharacter
    : '\\' [0\\tnr"']
    // NOTE: allow arbitrary length in parser, but check length in semantic analysis
    | '\\u' '{' HexadecimalDigit+ '}'
    ;

fragment HexadecimalDigit : [0-9a-fA-F] ;


WS
    : [ \t\u000B\u000C\u0000]+ -> channel(HIDDEN)
    ;

Terminator
    : [\r\n]+ -> channel(HIDDEN)
    ;

BlockComment
    : '/*' (BlockComment|.)*? '*/'	-> channel(HIDDEN) // nesting comments allowed
    ;

LineComment
    : '//' ~[\r\n]* -> channel(HIDDEN)
    ;

eos
    : ';'
    | EOF
    | {p.lineTerminatorAhead()}?
    | {p.GetTokenStream().LT(1).GetText() == "}"}?
    ;