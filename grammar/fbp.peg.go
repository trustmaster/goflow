package flow

import (
	"fmt"
	"math"
	"sort"
	"strconv"
)

const end_symbol rune = 4

/* The rule types inferred from the grammar are below. */
type pegRule uint8

const (
	ruleUnknown pegRule = iota
	ruleGrammar
	ruleImport
	ruleDefinition
	ruleExpression
	ruleSequence
	rulePrefix
	ruleSuffix
	rulePrimary
	ruleIdentifier
	ruleIdentStart
	ruleIdentCont
	ruleLiteral
	ruleClass
	ruleRanges
	ruleDoubleRanges
	ruleRange
	ruleDoubleRange
	ruleChar
	ruleDoubleChar
	ruleEscape
	ruleLeftArrow
	ruleSlash
	ruleAnd
	ruleNot
	ruleQuestion
	ruleStar
	rulePlus
	ruleOpen
	ruleClose
	ruleDot
	ruleSpaceComment
	ruleSpacing
	ruleMustSpacing
	ruleComment
	ruleSpace
	ruleEndOfLine
	ruleEndOfFile
	ruleAction
	ruleBegin
	ruleEnd
	ruleruntimeGetruntime
	ruleruntimeRuntime
	ruleruntimePorts
	ruleruntimePacket
	rulegraphClear
	rulegraphAddnode
	rulegraphRemovenode
	rulegraphRenamenode
	rulegraphChangenode
	rulegraphAddedge
	rulegraphRemoveedge
	rulegraphChangeedge
	rulegraphAddinitial
	rulegraphRemoveinitial
	rulegraphAddinport
	rulegraphRemoveinport
	rulegraphRenameinport
	rulegraphAddoutport
	rulegraphRemoveoutport
	rulegraphRenameoutport
	rulegraphAddgroup
	rulegraphRemovegroup
	rulegraphRenamegroup
	rulegraphChangegroup
	rulecomponentList
	rulecomponentComponent
	rulecomponentGetsource
	rulecomponentSource
	rulenetworkStart
	rulenetworkGetstatus
	rulenetworkStop
	rulenetworkStarted
	rulenetworkStatus
	rulenetworkStopped
	rulenetworkDebug
	rulenetworkIcon
	rulenetworkOutput
	rulenetworkError
	rulenetworkProcesserror
	rulenetworkConnect
	rulenetworkBegingroup
	rulenetworkData
	rulenetworkEndgroup
	rulenetworkDisconnect
	rulenetworkEdges
	ruleAction0
	ruleAction1
	ruleAction2
	rulePegText
	ruleAction3
	ruleAction4
	ruleAction5
	ruleAction6
	ruleAction7
	ruleAction8
	ruleAction9
	ruleAction10
	ruleAction11
	ruleAction12
	ruleAction13
	ruleAction14
	ruleAction15
	ruleAction16
	ruleAction17
	ruleAction18
	ruleAction19
	ruleAction20
	ruleAction21
	ruleAction22
	ruleAction23
	ruleAction24
	ruleAction25
	ruleAction26
	ruleAction27
	ruleAction28
	ruleAction29
	ruleAction30
	ruleAction31
	ruleAction32
	ruleAction33
	ruleAction34
	ruleAction35
	ruleAction36
	ruleAction37
	ruleAction38
	ruleAction39
	ruleAction40
	ruleAction41
	ruleAction42
	ruleAction43
	ruleAction44
	ruleAction45
	ruleAction46
	ruleAction47
	ruleAction48
	ruleAction49
	ruleAction50
	ruleAction51
	ruleAction52
	ruleAction53
	ruleAction54
	ruleAction55
	ruleAction56
	ruleAction57
	ruleAction58
	ruleAction59
	ruleAction60
	ruleAction61
	ruleAction62
	ruleAction63
	ruleAction64
	ruleAction65
	ruleAction66
	ruleAction67
	ruleAction68
	ruleAction69
	ruleAction70
	ruleAction71
	ruleAction72
	ruleAction73
	ruleAction74
	ruleAction75
	ruleAction76
	ruleAction77
	ruleAction78
	ruleAction79
	ruleAction80
	ruleAction81
	ruleAction82
	ruleAction83
	ruleAction84
	ruleAction85
	ruleAction86
	ruleAction87
	ruleAction88
	ruleAction89
	ruleAction90
	ruleAction91
	ruleAction92

	rulePre_
	rule_In_
	rule_Suf
)

var rul3s = [...]string{
	"Unknown",
	"Grammar",
	"Import",
	"Definition",
	"Expression",
	"Sequence",
	"Prefix",
	"Suffix",
	"Primary",
	"Identifier",
	"IdentStart",
	"IdentCont",
	"Literal",
	"Class",
	"Ranges",
	"DoubleRanges",
	"Range",
	"DoubleRange",
	"Char",
	"DoubleChar",
	"Escape",
	"LeftArrow",
	"Slash",
	"And",
	"Not",
	"Question",
	"Star",
	"Plus",
	"Open",
	"Close",
	"Dot",
	"SpaceComment",
	"Spacing",
	"MustSpacing",
	"Comment",
	"Space",
	"EndOfLine",
	"EndOfFile",
	"Action",
	"Begin",
	"End",
	"runtimeGetruntime",
	"runtimeRuntime",
	"runtimePorts",
	"runtimePacket",
	"graphClear",
	"graphAddnode",
	"graphRemovenode",
	"graphRenamenode",
	"graphChangenode",
	"graphAddedge",
	"graphRemoveedge",
	"graphChangeedge",
	"graphAddinitial",
	"graphRemoveinitial",
	"graphAddinport",
	"graphRemoveinport",
	"graphRenameinport",
	"graphAddoutport",
	"graphRemoveoutport",
	"graphRenameoutport",
	"graphAddgroup",
	"graphRemovegroup",
	"graphRenamegroup",
	"graphChangegroup",
	"componentList",
	"componentComponent",
	"componentGetsource",
	"componentSource",
	"networkStart",
	"networkGetstatus",
	"networkStop",
	"networkStarted",
	"networkStatus",
	"networkStopped",
	"networkDebug",
	"networkIcon",
	"networkOutput",
	"networkError",
	"networkProcesserror",
	"networkConnect",
	"networkBegingroup",
	"networkData",
	"networkEndgroup",
	"networkDisconnect",
	"networkEdges",
	"Action0",
	"Action1",
	"Action2",
	"PegText",
	"Action3",
	"Action4",
	"Action5",
	"Action6",
	"Action7",
	"Action8",
	"Action9",
	"Action10",
	"Action11",
	"Action12",
	"Action13",
	"Action14",
	"Action15",
	"Action16",
	"Action17",
	"Action18",
	"Action19",
	"Action20",
	"Action21",
	"Action22",
	"Action23",
	"Action24",
	"Action25",
	"Action26",
	"Action27",
	"Action28",
	"Action29",
	"Action30",
	"Action31",
	"Action32",
	"Action33",
	"Action34",
	"Action35",
	"Action36",
	"Action37",
	"Action38",
	"Action39",
	"Action40",
	"Action41",
	"Action42",
	"Action43",
	"Action44",
	"Action45",
	"Action46",
	"Action47",
	"Action48",
	"Action49",
	"Action50",
	"Action51",
	"Action52",
	"Action53",
	"Action54",
	"Action55",
	"Action56",
	"Action57",
	"Action58",
	"Action59",
	"Action60",
	"Action61",
	"Action62",
	"Action63",
	"Action64",
	"Action65",
	"Action66",
	"Action67",
	"Action68",
	"Action69",
	"Action70",
	"Action71",
	"Action72",
	"Action73",
	"Action74",
	"Action75",
	"Action76",
	"Action77",
	"Action78",
	"Action79",
	"Action80",
	"Action81",
	"Action82",
	"Action83",
	"Action84",
	"Action85",
	"Action86",
	"Action87",
	"Action88",
	"Action89",
	"Action90",
	"Action91",
	"Action92",

	"Pre_",
	"_In_",
	"_Suf",
}

type tokenTree interface {
	Print()
	PrintSyntax()
	PrintSyntaxTree(buffer string)
	Add(rule pegRule, begin, end, next, depth int)
	Expand(index int) tokenTree
	Tokens() <-chan token32
	AST() *node32
	Error() []token32
	trim(length int)
}

type node32 struct {
	token32
	up, next *node32
}

func (node *node32) print(depth int, buffer string) {
	for node != nil {
		for c := 0; c < depth; c++ {
			fmt.Printf(" ")
		}
		fmt.Printf("\x1B[34m%v\x1B[m %v\n", rul3s[node.pegRule], strconv.Quote(buffer[node.begin:node.end]))
		if node.up != nil {
			node.up.print(depth+1, buffer)
		}
		node = node.next
	}
}

func (ast *node32) Print(buffer string) {
	ast.print(0, buffer)
}

type element struct {
	node *node32
	down *element
}

/* ${@} bit structure for abstract syntax tree */
type token16 struct {
	pegRule
	begin, end, next int16
}

func (t *token16) isZero() bool {
	return t.pegRule == ruleUnknown && t.begin == 0 && t.end == 0 && t.next == 0
}

func (t *token16) isParentOf(u token16) bool {
	return t.begin <= u.begin && t.end >= u.end && t.next > u.next
}

func (t *token16) getToken32() token32 {
	return token32{pegRule: t.pegRule, begin: int32(t.begin), end: int32(t.end), next: int32(t.next)}
}

func (t *token16) String() string {
	return fmt.Sprintf("\x1B[34m%v\x1B[m %v %v %v", rul3s[t.pegRule], t.begin, t.end, t.next)
}

type tokens16 struct {
	tree    []token16
	ordered [][]token16
}

func (t *tokens16) trim(length int) {
	t.tree = t.tree[0:length]
}

func (t *tokens16) Print() {
	for _, token := range t.tree {
		fmt.Println(token.String())
	}
}

func (t *tokens16) Order() [][]token16 {
	if t.ordered != nil {
		return t.ordered
	}

	depths := make([]int16, 1, math.MaxInt16)
	for i, token := range t.tree {
		if token.pegRule == ruleUnknown {
			t.tree = t.tree[:i]
			break
		}
		depth := int(token.next)
		if length := len(depths); depth >= length {
			depths = depths[:depth+1]
		}
		depths[depth]++
	}
	depths = append(depths, 0)

	ordered, pool := make([][]token16, len(depths)), make([]token16, len(t.tree)+len(depths))
	for i, depth := range depths {
		depth++
		ordered[i], pool, depths[i] = pool[:depth], pool[depth:], 0
	}

	for i, token := range t.tree {
		depth := token.next
		token.next = int16(i)
		ordered[depth][depths[depth]] = token
		depths[depth]++
	}
	t.ordered = ordered
	return ordered
}

type state16 struct {
	token16
	depths []int16
	leaf   bool
}

func (t *tokens16) AST() *node32 {
	tokens := t.Tokens()
	stack := &element{node: &node32{token32: <-tokens}}
	for token := range tokens {
		if token.begin == token.end {
			continue
		}
		node := &node32{token32: token}
		for stack != nil && stack.node.begin >= token.begin && stack.node.end <= token.end {
			stack.node.next = node.up
			node.up = stack.node
			stack = stack.down
		}
		stack = &element{node: node, down: stack}
	}
	return stack.node
}

func (t *tokens16) PreOrder() (<-chan state16, [][]token16) {
	s, ordered := make(chan state16, 6), t.Order()
	go func() {
		var states [8]state16
		for i, _ := range states {
			states[i].depths = make([]int16, len(ordered))
		}
		depths, state, depth := make([]int16, len(ordered)), 0, 1
		write := func(t token16, leaf bool) {
			S := states[state]
			state, S.pegRule, S.begin, S.end, S.next, S.leaf = (state+1)%8, t.pegRule, t.begin, t.end, int16(depth), leaf
			copy(S.depths, depths)
			s <- S
		}

		states[state].token16 = ordered[0][0]
		depths[0]++
		state++
		a, b := ordered[depth-1][depths[depth-1]-1], ordered[depth][depths[depth]]
	depthFirstSearch:
		for {
			for {
				if i := depths[depth]; i > 0 {
					if c, j := ordered[depth][i-1], depths[depth-1]; a.isParentOf(c) &&
						(j < 2 || !ordered[depth-1][j-2].isParentOf(c)) {
						if c.end != b.begin {
							write(token16{pegRule: rule_In_, begin: c.end, end: b.begin}, true)
						}
						break
					}
				}

				if a.begin < b.begin {
					write(token16{pegRule: rulePre_, begin: a.begin, end: b.begin}, true)
				}
				break
			}

			next := depth + 1
			if c := ordered[next][depths[next]]; c.pegRule != ruleUnknown && b.isParentOf(c) {
				write(b, false)
				depths[depth]++
				depth, a, b = next, b, c
				continue
			}

			write(b, true)
			depths[depth]++
			c, parent := ordered[depth][depths[depth]], true
			for {
				if c.pegRule != ruleUnknown && a.isParentOf(c) {
					b = c
					continue depthFirstSearch
				} else if parent && b.end != a.end {
					write(token16{pegRule: rule_Suf, begin: b.end, end: a.end}, true)
				}

				depth--
				if depth > 0 {
					a, b, c = ordered[depth-1][depths[depth-1]-1], a, ordered[depth][depths[depth]]
					parent = a.isParentOf(b)
					continue
				}

				break depthFirstSearch
			}
		}

		close(s)
	}()
	return s, ordered
}

func (t *tokens16) PrintSyntax() {
	tokens, ordered := t.PreOrder()
	max := -1
	for token := range tokens {
		if !token.leaf {
			fmt.Printf("%v", token.begin)
			for i, leaf, depths := 0, int(token.next), token.depths; i < leaf; i++ {
				fmt.Printf(" \x1B[36m%v\x1B[m", rul3s[ordered[i][depths[i]-1].pegRule])
			}
			fmt.Printf(" \x1B[36m%v\x1B[m\n", rul3s[token.pegRule])
		} else if token.begin == token.end {
			fmt.Printf("%v", token.begin)
			for i, leaf, depths := 0, int(token.next), token.depths; i < leaf; i++ {
				fmt.Printf(" \x1B[31m%v\x1B[m", rul3s[ordered[i][depths[i]-1].pegRule])
			}
			fmt.Printf(" \x1B[31m%v\x1B[m\n", rul3s[token.pegRule])
		} else {
			for c, end := token.begin, token.end; c < end; c++ {
				if i := int(c); max+1 < i {
					for j := max; j < i; j++ {
						fmt.Printf("skip %v %v\n", j, token.String())
					}
					max = i
				} else if i := int(c); i <= max {
					for j := i; j <= max; j++ {
						fmt.Printf("dupe %v %v\n", j, token.String())
					}
				} else {
					max = int(c)
				}
				fmt.Printf("%v", c)
				for i, leaf, depths := 0, int(token.next), token.depths; i < leaf; i++ {
					fmt.Printf(" \x1B[34m%v\x1B[m", rul3s[ordered[i][depths[i]-1].pegRule])
				}
				fmt.Printf(" \x1B[34m%v\x1B[m\n", rul3s[token.pegRule])
			}
			fmt.Printf("\n")
		}
	}
}

func (t *tokens16) PrintSyntaxTree(buffer string) {
	tokens, _ := t.PreOrder()
	for token := range tokens {
		for c := 0; c < int(token.next); c++ {
			fmt.Printf(" ")
		}
		fmt.Printf("\x1B[34m%v\x1B[m %v\n", rul3s[token.pegRule], strconv.Quote(buffer[token.begin:token.end]))
	}
}

func (t *tokens16) Add(rule pegRule, begin, end, depth, index int) {
	t.tree[index] = token16{pegRule: rule, begin: int16(begin), end: int16(end), next: int16(depth)}
}

func (t *tokens16) Tokens() <-chan token32 {
	s := make(chan token32, 16)
	go func() {
		for _, v := range t.tree {
			s <- v.getToken32()
		}
		close(s)
	}()
	return s
}

func (t *tokens16) Error() []token32 {
	ordered := t.Order()
	length := len(ordered)
	tokens, length := make([]token32, length), length-1
	for i, _ := range tokens {
		o := ordered[length-i]
		if len(o) > 1 {
			tokens[i] = o[len(o)-2].getToken32()
		}
	}
	return tokens
}

/* ${@} bit structure for abstract syntax tree */
type token32 struct {
	pegRule
	begin, end, next int32
}

func (t *token32) isZero() bool {
	return t.pegRule == ruleUnknown && t.begin == 0 && t.end == 0 && t.next == 0
}

func (t *token32) isParentOf(u token32) bool {
	return t.begin <= u.begin && t.end >= u.end && t.next > u.next
}

func (t *token32) getToken32() token32 {
	return token32{pegRule: t.pegRule, begin: int32(t.begin), end: int32(t.end), next: int32(t.next)}
}

func (t *token32) String() string {
	return fmt.Sprintf("\x1B[34m%v\x1B[m %v %v %v", rul3s[t.pegRule], t.begin, t.end, t.next)
}

type tokens32 struct {
	tree    []token32
	ordered [][]token32
}

func (t *tokens32) trim(length int) {
	t.tree = t.tree[0:length]
}

func (t *tokens32) Print() {
	for _, token := range t.tree {
		fmt.Println(token.String())
	}
}

func (t *tokens32) Order() [][]token32 {
	if t.ordered != nil {
		return t.ordered
	}

	depths := make([]int32, 1, math.MaxInt16)
	for i, token := range t.tree {
		if token.pegRule == ruleUnknown {
			t.tree = t.tree[:i]
			break
		}
		depth := int(token.next)
		if length := len(depths); depth >= length {
			depths = depths[:depth+1]
		}
		depths[depth]++
	}
	depths = append(depths, 0)

	ordered, pool := make([][]token32, len(depths)), make([]token32, len(t.tree)+len(depths))
	for i, depth := range depths {
		depth++
		ordered[i], pool, depths[i] = pool[:depth], pool[depth:], 0
	}

	for i, token := range t.tree {
		depth := token.next
		token.next = int32(i)
		ordered[depth][depths[depth]] = token
		depths[depth]++
	}
	t.ordered = ordered
	return ordered
}

type state32 struct {
	token32
	depths []int32
	leaf   bool
}

func (t *tokens32) AST() *node32 {
	tokens := t.Tokens()
	stack := &element{node: &node32{token32: <-tokens}}
	for token := range tokens {
		if token.begin == token.end {
			continue
		}
		node := &node32{token32: token}
		for stack != nil && stack.node.begin >= token.begin && stack.node.end <= token.end {
			stack.node.next = node.up
			node.up = stack.node
			stack = stack.down
		}
		stack = &element{node: node, down: stack}
	}
	return stack.node
}

func (t *tokens32) PreOrder() (<-chan state32, [][]token32) {
	s, ordered := make(chan state32, 6), t.Order()
	go func() {
		var states [8]state32
		for i, _ := range states {
			states[i].depths = make([]int32, len(ordered))
		}
		depths, state, depth := make([]int32, len(ordered)), 0, 1
		write := func(t token32, leaf bool) {
			S := states[state]
			state, S.pegRule, S.begin, S.end, S.next, S.leaf = (state+1)%8, t.pegRule, t.begin, t.end, int32(depth), leaf
			copy(S.depths, depths)
			s <- S
		}

		states[state].token32 = ordered[0][0]
		depths[0]++
		state++
		a, b := ordered[depth-1][depths[depth-1]-1], ordered[depth][depths[depth]]
	depthFirstSearch:
		for {
			for {
				if i := depths[depth]; i > 0 {
					if c, j := ordered[depth][i-1], depths[depth-1]; a.isParentOf(c) &&
						(j < 2 || !ordered[depth-1][j-2].isParentOf(c)) {
						if c.end != b.begin {
							write(token32{pegRule: rule_In_, begin: c.end, end: b.begin}, true)
						}
						break
					}
				}

				if a.begin < b.begin {
					write(token32{pegRule: rulePre_, begin: a.begin, end: b.begin}, true)
				}
				break
			}

			next := depth + 1
			if c := ordered[next][depths[next]]; c.pegRule != ruleUnknown && b.isParentOf(c) {
				write(b, false)
				depths[depth]++
				depth, a, b = next, b, c
				continue
			}

			write(b, true)
			depths[depth]++
			c, parent := ordered[depth][depths[depth]], true
			for {
				if c.pegRule != ruleUnknown && a.isParentOf(c) {
					b = c
					continue depthFirstSearch
				} else if parent && b.end != a.end {
					write(token32{pegRule: rule_Suf, begin: b.end, end: a.end}, true)
				}

				depth--
				if depth > 0 {
					a, b, c = ordered[depth-1][depths[depth-1]-1], a, ordered[depth][depths[depth]]
					parent = a.isParentOf(b)
					continue
				}

				break depthFirstSearch
			}
		}

		close(s)
	}()
	return s, ordered
}

func (t *tokens32) PrintSyntax() {
	tokens, ordered := t.PreOrder()
	max := -1
	for token := range tokens {
		if !token.leaf {
			fmt.Printf("%v", token.begin)
			for i, leaf, depths := 0, int(token.next), token.depths; i < leaf; i++ {
				fmt.Printf(" \x1B[36m%v\x1B[m", rul3s[ordered[i][depths[i]-1].pegRule])
			}
			fmt.Printf(" \x1B[36m%v\x1B[m\n", rul3s[token.pegRule])
		} else if token.begin == token.end {
			fmt.Printf("%v", token.begin)
			for i, leaf, depths := 0, int(token.next), token.depths; i < leaf; i++ {
				fmt.Printf(" \x1B[31m%v\x1B[m", rul3s[ordered[i][depths[i]-1].pegRule])
			}
			fmt.Printf(" \x1B[31m%v\x1B[m\n", rul3s[token.pegRule])
		} else {
			for c, end := token.begin, token.end; c < end; c++ {
				if i := int(c); max+1 < i {
					for j := max; j < i; j++ {
						fmt.Printf("skip %v %v\n", j, token.String())
					}
					max = i
				} else if i := int(c); i <= max {
					for j := i; j <= max; j++ {
						fmt.Printf("dupe %v %v\n", j, token.String())
					}
				} else {
					max = int(c)
				}
				fmt.Printf("%v", c)
				for i, leaf, depths := 0, int(token.next), token.depths; i < leaf; i++ {
					fmt.Printf(" \x1B[34m%v\x1B[m", rul3s[ordered[i][depths[i]-1].pegRule])
				}
				fmt.Printf(" \x1B[34m%v\x1B[m\n", rul3s[token.pegRule])
			}
			fmt.Printf("\n")
		}
	}
}

func (t *tokens32) PrintSyntaxTree(buffer string) {
	tokens, _ := t.PreOrder()
	for token := range tokens {
		for c := 0; c < int(token.next); c++ {
			fmt.Printf(" ")
		}
		fmt.Printf("\x1B[34m%v\x1B[m %v\n", rul3s[token.pegRule], strconv.Quote(buffer[token.begin:token.end]))
	}
}

func (t *tokens32) Add(rule pegRule, begin, end, depth, index int) {
	t.tree[index] = token32{pegRule: rule, begin: int32(begin), end: int32(end), next: int32(depth)}
}

func (t *tokens32) Tokens() <-chan token32 {
	s := make(chan token32, 16)
	go func() {
		for _, v := range t.tree {
			s <- v.getToken32()
		}
		close(s)
	}()
	return s
}

func (t *tokens32) Error() []token32 {
	ordered := t.Order()
	length := len(ordered)
	tokens, length := make([]token32, length), length-1
	for i, _ := range tokens {
		o := ordered[length-i]
		if len(o) > 1 {
			tokens[i] = o[len(o)-2].getToken32()
		}
	}
	return tokens
}

func (t *tokens16) Expand(index int) tokenTree {
	tree := t.tree
	if index >= len(tree) {
		expanded := make([]token32, 2*len(tree))
		for i, v := range tree {
			expanded[i] = v.getToken32()
		}
		return &tokens32{tree: expanded}
	}
	return nil
}

func (t *tokens32) Expand(index int) tokenTree {
	tree := t.tree
	if index >= len(tree) {
		expanded := make([]token32, 2*len(tree))
		copy(expanded, tree)
		t.tree = expanded
	}
	return nil
}

type Peg struct {
	Runtime
	Graph

	Buffer string
	buffer []rune
	rules  [180]func() bool
	Parse  func(rule ...int) error
	Reset  func()
	tokenTree
}

type textPosition struct {
	line, symbol int
}

type textPositionMap map[int]textPosition

func translatePositions(buffer string, positions []int) textPositionMap {
	length, translations, j, line, symbol := len(positions), make(textPositionMap, len(positions)), 0, 1, 0
	sort.Ints(positions)

search:
	for i, c := range buffer[0:] {
		if c == '\n' {
			line, symbol = line+1, 0
		} else {
			symbol++
		}
		if i == positions[j] {
			translations[positions[j]] = textPosition{line, symbol}
			for j++; j < length; j++ {
				if i != positions[j] {
					continue search
				}
			}
			break search
		}
	}

	return translations
}

type parseError struct {
	p *Peg
}

func (e *parseError) Error() string {
	tokens, error := e.p.tokenTree.Error(), "\n"
	positions, p := make([]int, 2*len(tokens)), 0
	for _, token := range tokens {
		positions[p], p = int(token.begin), p+1
		positions[p], p = int(token.end), p+1
	}
	translations := translatePositions(e.p.Buffer, positions)
	for _, token := range tokens {
		begin, end := int(token.begin), int(token.end)
		error += fmt.Sprintf("parse error near \x1B[34m%v\x1B[m (line %v symbol %v - line %v symbol %v):\n%v\n",
			rul3s[token.pegRule],
			translations[begin].line, translations[begin].symbol,
			translations[end].line, translations[end].symbol,
			/*strconv.Quote(*/ e.p.Buffer[begin:end] /*)*/)
	}

	return error
}

func (p *Peg) PrintSyntaxTree() {
	p.tokenTree.PrintSyntaxTree(p.Buffer)
}

func (p *Peg) Highlighter() {
	p.tokenTree.PrintSyntax()
}

func (p *Peg) Execute() {
	buffer, begin, end := p.Buffer, 0, 0
	for token := range p.tokenTree.Tokens() {
		switch token.pegRule {
		case rulePegText:
			begin, end = int(token.begin), int(token.end)
		case ruleAction0:
			p.AddPackage(buffer[begin:end])
		case ruleAction1:
			p.AddPeg(buffer[begin:end])
		case ruleAction2:
			p.AddState(buffer[begin:end])
		case ruleAction3:
			p.AddImport(buffer[begin:end])
		case ruleAction4:
			p.AddRule(buffer[begin:end])
		case ruleAction5:
			p.AddExpression()
		case ruleAction6:
			p.AddAlternate()
		case ruleAction7:
			p.AddNil()
			p.AddAlternate()
		case ruleAction8:
			p.AddNil()
		case ruleAction9:
			p.AddSequence()
		case ruleAction10:
			p.AddPredicate(buffer[begin:end])
		case ruleAction11:
			p.AddPeekFor()
		case ruleAction12:
			p.AddPeekNot()
		case ruleAction13:
			p.AddQuery()
		case ruleAction14:
			p.AddStar()
		case ruleAction15:
			p.AddPlus()
		case ruleAction16:
			p.AddName(buffer[begin:end])
		case ruleAction17:
			p.AddDot()
		case ruleAction18:
			p.AddAction(buffer[begin:end])
		case ruleAction19:
			p.AddPush()
		case ruleAction20:
			p.AddSequence()
		case ruleAction21:
			p.AddSequence()
		case ruleAction22:
			p.AddPeekNot()
			p.AddDot()
			p.AddSequence()
		case ruleAction23:
			p.AddPeekNot()
			p.AddDot()
			p.AddSequence()
		case ruleAction24:
			p.AddAlternate()
		case ruleAction25:
			p.AddAlternate()
		case ruleAction26:
			p.AddRange()
		case ruleAction27:
			p.AddDoubleRange()
		case ruleAction28:
			p.AddCharacter(buffer[begin:end])
		case ruleAction29:
			p.AddDoubleCharacter(buffer[begin:end])
		case ruleAction30:
			p.AddCharacter(buffer[begin:end])
		case ruleAction31:
			p.AddCharacter("\a")
		case ruleAction32:
			p.AddCharacter("\b")
		case ruleAction33:
			p.AddCharacter("\x1B")
		case ruleAction34:
			p.AddCharacter("\f")
		case ruleAction35:
			p.AddCharacter("\n")
		case ruleAction36:
			p.AddCharacter("\r")
		case ruleAction37:
			p.AddCharacter("\t")
		case ruleAction38:
			p.AddCharacter("\v")
		case ruleAction39:
			p.AddCharacter("'")
		case ruleAction40:
			p.AddCharacter("\"")
		case ruleAction41:
			p.AddCharacter("[")
		case ruleAction42:
			p.AddCharacter("]")
		case ruleAction43:
			p.AddCharacter("-")
		case ruleAction44:
			p.AddHexaCharacter(buffer[begin:end])
		case ruleAction45:
			p.AddOctalCharacter(buffer[begin:end])
		case ruleAction46:
			p.AddOctalCharacter(buffer[begin:end])
		case ruleAction47:
			p.AddCharacter("\\")
		case ruleAction48:
			p.runtimeGetruntime()
		case ruleAction49:
			p.runtimeRuntime()
		case ruleAction50:
			p.runtimePorts()
		case ruleAction51:
			p.runtimePacket()
		case ruleAction52:
			p.graphClear()
		case ruleAction53:
			p.graphAddnode()
		case ruleAction54:
			p.graphRemovenode()
		case ruleAction55:
			p.graphRenamenode()
		case ruleAction56:
			p.graphChangenode()
		case ruleAction57:
			p.graphAddedge()
		case ruleAction58:
			p.graphRemoveedge()
		case ruleAction59:
			p.graphChangeedge()
		case ruleAction60:
			p.graphAddinitial()
		case ruleAction61:
			p.graphRemoveinitial()
		case ruleAction62:
			p.graphAddinport()
		case ruleAction63:
			p.graphRemoveinport()
		case ruleAction64:
			p.graphRenameinport()
		case ruleAction65:
			p.graphAddoutport()
		case ruleAction66:
			p.graphRemoveoutport()
		case ruleAction67:
			p.graphRenameoutport()
		case ruleAction68:
			p.graphAddgroup()
		case ruleAction69:
			p.graphRemovegroup()
		case ruleAction70:
			p.graphRenamegroup()
		case ruleAction71:
			p.graphChangegroup()
		case ruleAction72:
			p.componentList()
		case ruleAction73:
			p.componentComponent()
		case ruleAction74:
			p.componentGetsource()
		case ruleAction75:
			p.componentSource()
		case ruleAction76:
			p.networkStart()
		case ruleAction77:
			p.networkGetstatus()
		case ruleAction78:
			p.networkStop()
		case ruleAction79:
			p.networkStarted()
		case ruleAction80:
			p.networkStatus()
		case ruleAction81:
			p.networkStopped()
		case ruleAction82:
			p.networkDebug()
		case ruleAction83:
			p.networkIcon()
		case ruleAction84:
			p.networkOutput()
		case ruleAction85:
			p.networkError()
		case ruleAction86:
			p.networkProcesserror()
		case ruleAction87:
			p.networkConnect()
		case ruleAction88:
			p.networkBegingroup()
		case ruleAction89:
			p.networkData()
		case ruleAction90:
			p.networkEndgroup()
		case ruleAction91:
			p.networkDisconnect()
		case ruleAction92:
			p.networkEdges()

		}
	}
}

func (p *Peg) Init() {
	p.buffer = []rune(p.Buffer)
	if len(p.buffer) == 0 || p.buffer[len(p.buffer)-1] != end_symbol {
		p.buffer = append(p.buffer, end_symbol)
	}

	var tree tokenTree = &tokens16{tree: make([]token16, math.MaxInt16)}
	position, depth, tokenIndex, buffer, _rules := 0, 0, 0, p.buffer, p.rules

	p.Parse = func(rule ...int) error {
		r := 1
		if len(rule) > 0 {
			r = rule[0]
		}
		matches := p.rules[r]()
		p.tokenTree = tree
		if matches {
			p.tokenTree.trim(tokenIndex)
			return nil
		}
		return &parseError{p}
	}

	p.Reset = func() {
		position, tokenIndex, depth = 0, 0, 0
	}

	add := func(rule pegRule, begin int) {
		if t := tree.Expand(tokenIndex); t != nil {
			tree = t
		}
		tree.Add(rule, begin, position, depth, tokenIndex)
		tokenIndex++
	}

	matchDot := func() bool {
		if buffer[position] != end_symbol {
			position++
			return true
		}
		return false
	}

	/*matchChar := func(c byte) bool {
		if buffer[position] == c {
			position++
			return true
		}
		return false
	}*/

	/*matchRange := func(lower byte, upper byte) bool {
		if c := buffer[position]; c >= lower && c <= upper {
			position++
			return true
		}
		return false
	}*/

	_rules = [...]func() bool{
		nil,
		/* 0 Grammar <- <(Spacing ('p' 'a' 'c' 'k' 'a' 'g' 'e') MustSpacing Identifier Action0 Import* ('t' 'y' 'p' 'e') MustSpacing Identifier Action1 ('P' 'e' 'g') Spacing Action Action2 Definition+ EndOfFile)> */
		func() bool {
			position0, tokenIndex0, depth0 := position, tokenIndex, depth
			{
				position1 := position
				depth++
				if !_rules[ruleSpacing]() {
					goto l0
				}
				if buffer[position] != rune('p') {
					goto l0
				}
				position++
				if buffer[position] != rune('a') {
					goto l0
				}
				position++
				if buffer[position] != rune('c') {
					goto l0
				}
				position++
				if buffer[position] != rune('k') {
					goto l0
				}
				position++
				if buffer[position] != rune('a') {
					goto l0
				}
				position++
				if buffer[position] != rune('g') {
					goto l0
				}
				position++
				if buffer[position] != rune('e') {
					goto l0
				}
				position++
				if !_rules[ruleMustSpacing]() {
					goto l0
				}
				if !_rules[ruleIdentifier]() {
					goto l0
				}
				if !_rules[ruleAction0]() {
					goto l0
				}
			l2:
				{
					position3, tokenIndex3, depth3 := position, tokenIndex, depth
					if !_rules[ruleImport]() {
						goto l3
					}
					goto l2
				l3:
					position, tokenIndex, depth = position3, tokenIndex3, depth3
				}
				if buffer[position] != rune('t') {
					goto l0
				}
				position++
				if buffer[position] != rune('y') {
					goto l0
				}
				position++
				if buffer[position] != rune('p') {
					goto l0
				}
				position++
				if buffer[position] != rune('e') {
					goto l0
				}
				position++
				if !_rules[ruleMustSpacing]() {
					goto l0
				}
				if !_rules[ruleIdentifier]() {
					goto l0
				}
				if !_rules[ruleAction1]() {
					goto l0
				}
				if buffer[position] != rune('P') {
					goto l0
				}
				position++
				if buffer[position] != rune('e') {
					goto l0
				}
				position++
				if buffer[position] != rune('g') {
					goto l0
				}
				position++
				if !_rules[ruleSpacing]() {
					goto l0
				}
				if !_rules[ruleAction]() {
					goto l0
				}
				if !_rules[ruleAction2]() {
					goto l0
				}
				if !_rules[ruleDefinition]() {
					goto l0
				}
			l4:
				{
					position5, tokenIndex5, depth5 := position, tokenIndex, depth
					if !_rules[ruleDefinition]() {
						goto l5
					}
					goto l4
				l5:
					position, tokenIndex, depth = position5, tokenIndex5, depth5
				}
				if !_rules[ruleEndOfFile]() {
					goto l0
				}
				depth--
				add(ruleGrammar, position1)
			}
			return true
		l0:
			position, tokenIndex, depth = position0, tokenIndex0, depth0
			return false
		},
		/* 1 Import <- <('i' 'm' 'p' 'o' 'r' 't' Spacing '"' <([a-z] / [A-Z] / '_' / '/' / '.' / '-')+> '"' Spacing Action3)> */
		func() bool {
			position6, tokenIndex6, depth6 := position, tokenIndex, depth
			{
				position7 := position
				depth++
				if buffer[position] != rune('i') {
					goto l6
				}
				position++
				if buffer[position] != rune('m') {
					goto l6
				}
				position++
				if buffer[position] != rune('p') {
					goto l6
				}
				position++
				if buffer[position] != rune('o') {
					goto l6
				}
				position++
				if buffer[position] != rune('r') {
					goto l6
				}
				position++
				if buffer[position] != rune('t') {
					goto l6
				}
				position++
				if !_rules[ruleSpacing]() {
					goto l6
				}
				if buffer[position] != rune('"') {
					goto l6
				}
				position++
				{
					position8 := position
					depth++
					{
						position11, tokenIndex11, depth11 := position, tokenIndex, depth
						if c := buffer[position]; c < rune('a') || c > rune('z') {
							goto l12
						}
						position++
						goto l11
					l12:
						position, tokenIndex, depth = position11, tokenIndex11, depth11
						if c := buffer[position]; c < rune('A') || c > rune('Z') {
							goto l13
						}
						position++
						goto l11
					l13:
						position, tokenIndex, depth = position11, tokenIndex11, depth11
						if buffer[position] != rune('_') {
							goto l14
						}
						position++
						goto l11
					l14:
						position, tokenIndex, depth = position11, tokenIndex11, depth11
						if buffer[position] != rune('/') {
							goto l15
						}
						position++
						goto l11
					l15:
						position, tokenIndex, depth = position11, tokenIndex11, depth11
						if buffer[position] != rune('.') {
							goto l16
						}
						position++
						goto l11
					l16:
						position, tokenIndex, depth = position11, tokenIndex11, depth11
						if buffer[position] != rune('-') {
							goto l6
						}
						position++
					}
				l11:
				l9:
					{
						position10, tokenIndex10, depth10 := position, tokenIndex, depth
						{
							position17, tokenIndex17, depth17 := position, tokenIndex, depth
							if c := buffer[position]; c < rune('a') || c > rune('z') {
								goto l18
							}
							position++
							goto l17
						l18:
							position, tokenIndex, depth = position17, tokenIndex17, depth17
							if c := buffer[position]; c < rune('A') || c > rune('Z') {
								goto l19
							}
							position++
							goto l17
						l19:
							position, tokenIndex, depth = position17, tokenIndex17, depth17
							if buffer[position] != rune('_') {
								goto l20
							}
							position++
							goto l17
						l20:
							position, tokenIndex, depth = position17, tokenIndex17, depth17
							if buffer[position] != rune('/') {
								goto l21
							}
							position++
							goto l17
						l21:
							position, tokenIndex, depth = position17, tokenIndex17, depth17
							if buffer[position] != rune('.') {
								goto l22
							}
							position++
							goto l17
						l22:
							position, tokenIndex, depth = position17, tokenIndex17, depth17
							if buffer[position] != rune('-') {
								goto l10
							}
							position++
						}
					l17:
						goto l9
					l10:
						position, tokenIndex, depth = position10, tokenIndex10, depth10
					}
					depth--
					add(rulePegText, position8)
				}
				if buffer[position] != rune('"') {
					goto l6
				}
				position++
				if !_rules[ruleSpacing]() {
					goto l6
				}
				if !_rules[ruleAction3]() {
					goto l6
				}
				depth--
				add(ruleImport, position7)
			}
			return true
		l6:
			position, tokenIndex, depth = position6, tokenIndex6, depth6
			return false
		},
		/* 2 Definition <- <(Identifier Action4 LeftArrow Expression Action5 &((Identifier LeftArrow) / !.))> */
		func() bool {
			position23, tokenIndex23, depth23 := position, tokenIndex, depth
			{
				position24 := position
				depth++
				if !_rules[ruleIdentifier]() {
					goto l23
				}
				if !_rules[ruleAction4]() {
					goto l23
				}
				if !_rules[ruleLeftArrow]() {
					goto l23
				}
				if !_rules[ruleExpression]() {
					goto l23
				}
				if !_rules[ruleAction5]() {
					goto l23
				}
				{
					position25, tokenIndex25, depth25 := position, tokenIndex, depth
					{
						position26, tokenIndex26, depth26 := position, tokenIndex, depth
						if !_rules[ruleIdentifier]() {
							goto l27
						}
						if !_rules[ruleLeftArrow]() {
							goto l27
						}
						goto l26
					l27:
						position, tokenIndex, depth = position26, tokenIndex26, depth26
						{
							position28, tokenIndex28, depth28 := position, tokenIndex, depth
							if !matchDot() {
								goto l28
							}
							goto l23
						l28:
							position, tokenIndex, depth = position28, tokenIndex28, depth28
						}
					}
				l26:
					position, tokenIndex, depth = position25, tokenIndex25, depth25
				}
				depth--
				add(ruleDefinition, position24)
			}
			return true
		l23:
			position, tokenIndex, depth = position23, tokenIndex23, depth23
			return false
		},
		/* 3 Expression <- <((Sequence (Slash Sequence Action6)* (Slash Action7)?) / Action8)> */
		func() bool {
			position29, tokenIndex29, depth29 := position, tokenIndex, depth
			{
				position30 := position
				depth++
				{
					position31, tokenIndex31, depth31 := position, tokenIndex, depth
					if !_rules[ruleSequence]() {
						goto l32
					}
				l33:
					{
						position34, tokenIndex34, depth34 := position, tokenIndex, depth
						if !_rules[ruleSlash]() {
							goto l34
						}
						if !_rules[ruleSequence]() {
							goto l34
						}
						if !_rules[ruleAction6]() {
							goto l34
						}
						goto l33
					l34:
						position, tokenIndex, depth = position34, tokenIndex34, depth34
					}
					{
						position35, tokenIndex35, depth35 := position, tokenIndex, depth
						if !_rules[ruleSlash]() {
							goto l35
						}
						if !_rules[ruleAction7]() {
							goto l35
						}
						goto l36
					l35:
						position, tokenIndex, depth = position35, tokenIndex35, depth35
					}
				l36:
					goto l31
				l32:
					position, tokenIndex, depth = position31, tokenIndex31, depth31
					if !_rules[ruleAction8]() {
						goto l29
					}
				}
			l31:
				depth--
				add(ruleExpression, position30)
			}
			return true
		l29:
			position, tokenIndex, depth = position29, tokenIndex29, depth29
			return false
		},
		/* 4 Sequence <- <(Prefix (Prefix Action9)*)> */
		func() bool {
			position37, tokenIndex37, depth37 := position, tokenIndex, depth
			{
				position38 := position
				depth++
				if !_rules[rulePrefix]() {
					goto l37
				}
			l39:
				{
					position40, tokenIndex40, depth40 := position, tokenIndex, depth
					if !_rules[rulePrefix]() {
						goto l40
					}
					if !_rules[ruleAction9]() {
						goto l40
					}
					goto l39
				l40:
					position, tokenIndex, depth = position40, tokenIndex40, depth40
				}
				depth--
				add(ruleSequence, position38)
			}
			return true
		l37:
			position, tokenIndex, depth = position37, tokenIndex37, depth37
			return false
		},
		/* 5 Prefix <- <((And Action Action10) / (And Suffix Action11) / (Not Suffix Action12) / Suffix)> */
		func() bool {
			position41, tokenIndex41, depth41 := position, tokenIndex, depth
			{
				position42 := position
				depth++
				{
					position43, tokenIndex43, depth43 := position, tokenIndex, depth
					if !_rules[ruleAnd]() {
						goto l44
					}
					if !_rules[ruleAction]() {
						goto l44
					}
					if !_rules[ruleAction10]() {
						goto l44
					}
					goto l43
				l44:
					position, tokenIndex, depth = position43, tokenIndex43, depth43
					if !_rules[ruleAnd]() {
						goto l45
					}
					if !_rules[ruleSuffix]() {
						goto l45
					}
					if !_rules[ruleAction11]() {
						goto l45
					}
					goto l43
				l45:
					position, tokenIndex, depth = position43, tokenIndex43, depth43
					if !_rules[ruleNot]() {
						goto l46
					}
					if !_rules[ruleSuffix]() {
						goto l46
					}
					if !_rules[ruleAction12]() {
						goto l46
					}
					goto l43
				l46:
					position, tokenIndex, depth = position43, tokenIndex43, depth43
					if !_rules[ruleSuffix]() {
						goto l41
					}
				}
			l43:
				depth--
				add(rulePrefix, position42)
			}
			return true
		l41:
			position, tokenIndex, depth = position41, tokenIndex41, depth41
			return false
		},
		/* 6 Suffix <- <(Primary ((Question Action13) / (Star Action14) / (Plus Action15))?)> */
		func() bool {
			position47, tokenIndex47, depth47 := position, tokenIndex, depth
			{
				position48 := position
				depth++
				if !_rules[rulePrimary]() {
					goto l47
				}
				{
					position49, tokenIndex49, depth49 := position, tokenIndex, depth
					{
						position51, tokenIndex51, depth51 := position, tokenIndex, depth
						if !_rules[ruleQuestion]() {
							goto l52
						}
						if !_rules[ruleAction13]() {
							goto l52
						}
						goto l51
					l52:
						position, tokenIndex, depth = position51, tokenIndex51, depth51
						if !_rules[ruleStar]() {
							goto l53
						}
						if !_rules[ruleAction14]() {
							goto l53
						}
						goto l51
					l53:
						position, tokenIndex, depth = position51, tokenIndex51, depth51
						if !_rules[rulePlus]() {
							goto l49
						}
						if !_rules[ruleAction15]() {
							goto l49
						}
					}
				l51:
					goto l50
				l49:
					position, tokenIndex, depth = position49, tokenIndex49, depth49
				}
			l50:
				depth--
				add(ruleSuffix, position48)
			}
			return true
		l47:
			position, tokenIndex, depth = position47, tokenIndex47, depth47
			return false
		},
		/* 7 Primary <- <((Identifier !LeftArrow Action16) / (Open Expression Close) / Literal / Class / (Dot Action17) / (Action Action18) / (Begin Expression End Action19))> */
		func() bool {
			position54, tokenIndex54, depth54 := position, tokenIndex, depth
			{
				position55 := position
				depth++
				{
					position56, tokenIndex56, depth56 := position, tokenIndex, depth
					if !_rules[ruleIdentifier]() {
						goto l57
					}
					{
						position58, tokenIndex58, depth58 := position, tokenIndex, depth
						if !_rules[ruleLeftArrow]() {
							goto l58
						}
						goto l57
					l58:
						position, tokenIndex, depth = position58, tokenIndex58, depth58
					}
					if !_rules[ruleAction16]() {
						goto l57
					}
					goto l56
				l57:
					position, tokenIndex, depth = position56, tokenIndex56, depth56
					if !_rules[ruleOpen]() {
						goto l59
					}
					if !_rules[ruleExpression]() {
						goto l59
					}
					if !_rules[ruleClose]() {
						goto l59
					}
					goto l56
				l59:
					position, tokenIndex, depth = position56, tokenIndex56, depth56
					if !_rules[ruleLiteral]() {
						goto l60
					}
					goto l56
				l60:
					position, tokenIndex, depth = position56, tokenIndex56, depth56
					if !_rules[ruleClass]() {
						goto l61
					}
					goto l56
				l61:
					position, tokenIndex, depth = position56, tokenIndex56, depth56
					if !_rules[ruleDot]() {
						goto l62
					}
					if !_rules[ruleAction17]() {
						goto l62
					}
					goto l56
				l62:
					position, tokenIndex, depth = position56, tokenIndex56, depth56
					if !_rules[ruleAction]() {
						goto l63
					}
					if !_rules[ruleAction18]() {
						goto l63
					}
					goto l56
				l63:
					position, tokenIndex, depth = position56, tokenIndex56, depth56
					if !_rules[ruleBegin]() {
						goto l54
					}
					if !_rules[ruleExpression]() {
						goto l54
					}
					if !_rules[ruleEnd]() {
						goto l54
					}
					if !_rules[ruleAction19]() {
						goto l54
					}
				}
			l56:
				depth--
				add(rulePrimary, position55)
			}
			return true
		l54:
			position, tokenIndex, depth = position54, tokenIndex54, depth54
			return false
		},
		/* 8 Identifier <- <(<(IdentStart IdentCont*)> Spacing)> */
		func() bool {
			position64, tokenIndex64, depth64 := position, tokenIndex, depth
			{
				position65 := position
				depth++
				{
					position66 := position
					depth++
					if !_rules[ruleIdentStart]() {
						goto l64
					}
				l67:
					{
						position68, tokenIndex68, depth68 := position, tokenIndex, depth
						if !_rules[ruleIdentCont]() {
							goto l68
						}
						goto l67
					l68:
						position, tokenIndex, depth = position68, tokenIndex68, depth68
					}
					depth--
					add(rulePegText, position66)
				}
				if !_rules[ruleSpacing]() {
					goto l64
				}
				depth--
				add(ruleIdentifier, position65)
			}
			return true
		l64:
			position, tokenIndex, depth = position64, tokenIndex64, depth64
			return false
		},
		/* 9 IdentStart <- <([a-z] / [A-Z] / '_')> */
		func() bool {
			position69, tokenIndex69, depth69 := position, tokenIndex, depth
			{
				position70 := position
				depth++
				{
					position71, tokenIndex71, depth71 := position, tokenIndex, depth
					if c := buffer[position]; c < rune('a') || c > rune('z') {
						goto l72
					}
					position++
					goto l71
				l72:
					position, tokenIndex, depth = position71, tokenIndex71, depth71
					if c := buffer[position]; c < rune('A') || c > rune('Z') {
						goto l73
					}
					position++
					goto l71
				l73:
					position, tokenIndex, depth = position71, tokenIndex71, depth71
					if buffer[position] != rune('_') {
						goto l69
					}
					position++
				}
			l71:
				depth--
				add(ruleIdentStart, position70)
			}
			return true
		l69:
			position, tokenIndex, depth = position69, tokenIndex69, depth69
			return false
		},
		/* 10 IdentCont <- <(IdentStart / [0-9])> */
		func() bool {
			position74, tokenIndex74, depth74 := position, tokenIndex, depth
			{
				position75 := position
				depth++
				{
					position76, tokenIndex76, depth76 := position, tokenIndex, depth
					if !_rules[ruleIdentStart]() {
						goto l77
					}
					goto l76
				l77:
					position, tokenIndex, depth = position76, tokenIndex76, depth76
					if c := buffer[position]; c < rune('0') || c > rune('9') {
						goto l74
					}
					position++
				}
			l76:
				depth--
				add(ruleIdentCont, position75)
			}
			return true
		l74:
			position, tokenIndex, depth = position74, tokenIndex74, depth74
			return false
		},
		/* 11 Literal <- <(('\'' (!'\'' Char)? (!'\'' Char Action20)* '\'' Spacing) / ('"' (!'"' DoubleChar)? (!'"' DoubleChar Action21)* '"' Spacing))> */
		func() bool {
			position78, tokenIndex78, depth78 := position, tokenIndex, depth
			{
				position79 := position
				depth++
				{
					position80, tokenIndex80, depth80 := position, tokenIndex, depth
					if buffer[position] != rune('\'') {
						goto l81
					}
					position++
					{
						position82, tokenIndex82, depth82 := position, tokenIndex, depth
						{
							position84, tokenIndex84, depth84 := position, tokenIndex, depth
							if buffer[position] != rune('\'') {
								goto l84
							}
							position++
							goto l82
						l84:
							position, tokenIndex, depth = position84, tokenIndex84, depth84
						}
						if !_rules[ruleChar]() {
							goto l82
						}
						goto l83
					l82:
						position, tokenIndex, depth = position82, tokenIndex82, depth82
					}
				l83:
				l85:
					{
						position86, tokenIndex86, depth86 := position, tokenIndex, depth
						{
							position87, tokenIndex87, depth87 := position, tokenIndex, depth
							if buffer[position] != rune('\'') {
								goto l87
							}
							position++
							goto l86
						l87:
							position, tokenIndex, depth = position87, tokenIndex87, depth87
						}
						if !_rules[ruleChar]() {
							goto l86
						}
						if !_rules[ruleAction20]() {
							goto l86
						}
						goto l85
					l86:
						position, tokenIndex, depth = position86, tokenIndex86, depth86
					}
					if buffer[position] != rune('\'') {
						goto l81
					}
					position++
					if !_rules[ruleSpacing]() {
						goto l81
					}
					goto l80
				l81:
					position, tokenIndex, depth = position80, tokenIndex80, depth80
					if buffer[position] != rune('"') {
						goto l78
					}
					position++
					{
						position88, tokenIndex88, depth88 := position, tokenIndex, depth
						{
							position90, tokenIndex90, depth90 := position, tokenIndex, depth
							if buffer[position] != rune('"') {
								goto l90
							}
							position++
							goto l88
						l90:
							position, tokenIndex, depth = position90, tokenIndex90, depth90
						}
						if !_rules[ruleDoubleChar]() {
							goto l88
						}
						goto l89
					l88:
						position, tokenIndex, depth = position88, tokenIndex88, depth88
					}
				l89:
				l91:
					{
						position92, tokenIndex92, depth92 := position, tokenIndex, depth
						{
							position93, tokenIndex93, depth93 := position, tokenIndex, depth
							if buffer[position] != rune('"') {
								goto l93
							}
							position++
							goto l92
						l93:
							position, tokenIndex, depth = position93, tokenIndex93, depth93
						}
						if !_rules[ruleDoubleChar]() {
							goto l92
						}
						if !_rules[ruleAction21]() {
							goto l92
						}
						goto l91
					l92:
						position, tokenIndex, depth = position92, tokenIndex92, depth92
					}
					if buffer[position] != rune('"') {
						goto l78
					}
					position++
					if !_rules[ruleSpacing]() {
						goto l78
					}
				}
			l80:
				depth--
				add(ruleLiteral, position79)
			}
			return true
		l78:
			position, tokenIndex, depth = position78, tokenIndex78, depth78
			return false
		},
		/* 12 Class <- <((('[' '[' (('^' DoubleRanges Action22) / DoubleRanges)? (']' ']')) / ('[' (('^' Ranges Action23) / Ranges)? ']')) Spacing)> */
		func() bool {
			position94, tokenIndex94, depth94 := position, tokenIndex, depth
			{
				position95 := position
				depth++
				{
					position96, tokenIndex96, depth96 := position, tokenIndex, depth
					if buffer[position] != rune('[') {
						goto l97
					}
					position++
					if buffer[position] != rune('[') {
						goto l97
					}
					position++
					{
						position98, tokenIndex98, depth98 := position, tokenIndex, depth
						{
							position100, tokenIndex100, depth100 := position, tokenIndex, depth
							if buffer[position] != rune('^') {
								goto l101
							}
							position++
							if !_rules[ruleDoubleRanges]() {
								goto l101
							}
							if !_rules[ruleAction22]() {
								goto l101
							}
							goto l100
						l101:
							position, tokenIndex, depth = position100, tokenIndex100, depth100
							if !_rules[ruleDoubleRanges]() {
								goto l98
							}
						}
					l100:
						goto l99
					l98:
						position, tokenIndex, depth = position98, tokenIndex98, depth98
					}
				l99:
					if buffer[position] != rune(']') {
						goto l97
					}
					position++
					if buffer[position] != rune(']') {
						goto l97
					}
					position++
					goto l96
				l97:
					position, tokenIndex, depth = position96, tokenIndex96, depth96
					if buffer[position] != rune('[') {
						goto l94
					}
					position++
					{
						position102, tokenIndex102, depth102 := position, tokenIndex, depth
						{
							position104, tokenIndex104, depth104 := position, tokenIndex, depth
							if buffer[position] != rune('^') {
								goto l105
							}
							position++
							if !_rules[ruleRanges]() {
								goto l105
							}
							if !_rules[ruleAction23]() {
								goto l105
							}
							goto l104
						l105:
							position, tokenIndex, depth = position104, tokenIndex104, depth104
							if !_rules[ruleRanges]() {
								goto l102
							}
						}
					l104:
						goto l103
					l102:
						position, tokenIndex, depth = position102, tokenIndex102, depth102
					}
				l103:
					if buffer[position] != rune(']') {
						goto l94
					}
					position++
				}
			l96:
				if !_rules[ruleSpacing]() {
					goto l94
				}
				depth--
				add(ruleClass, position95)
			}
			return true
		l94:
			position, tokenIndex, depth = position94, tokenIndex94, depth94
			return false
		},
		/* 13 Ranges <- <(!']' Range (!']' Range Action24)*)> */
		func() bool {
			position106, tokenIndex106, depth106 := position, tokenIndex, depth
			{
				position107 := position
				depth++
				{
					position108, tokenIndex108, depth108 := position, tokenIndex, depth
					if buffer[position] != rune(']') {
						goto l108
					}
					position++
					goto l106
				l108:
					position, tokenIndex, depth = position108, tokenIndex108, depth108
				}
				if !_rules[ruleRange]() {
					goto l106
				}
			l109:
				{
					position110, tokenIndex110, depth110 := position, tokenIndex, depth
					{
						position111, tokenIndex111, depth111 := position, tokenIndex, depth
						if buffer[position] != rune(']') {
							goto l111
						}
						position++
						goto l110
					l111:
						position, tokenIndex, depth = position111, tokenIndex111, depth111
					}
					if !_rules[ruleRange]() {
						goto l110
					}
					if !_rules[ruleAction24]() {
						goto l110
					}
					goto l109
				l110:
					position, tokenIndex, depth = position110, tokenIndex110, depth110
				}
				depth--
				add(ruleRanges, position107)
			}
			return true
		l106:
			position, tokenIndex, depth = position106, tokenIndex106, depth106
			return false
		},
		/* 14 DoubleRanges <- <(!(']' ']') DoubleRange (!(']' ']') DoubleRange Action25)*)> */
		func() bool {
			position112, tokenIndex112, depth112 := position, tokenIndex, depth
			{
				position113 := position
				depth++
				{
					position114, tokenIndex114, depth114 := position, tokenIndex, depth
					if buffer[position] != rune(']') {
						goto l114
					}
					position++
					if buffer[position] != rune(']') {
						goto l114
					}
					position++
					goto l112
				l114:
					position, tokenIndex, depth = position114, tokenIndex114, depth114
				}
				if !_rules[ruleDoubleRange]() {
					goto l112
				}
			l115:
				{
					position116, tokenIndex116, depth116 := position, tokenIndex, depth
					{
						position117, tokenIndex117, depth117 := position, tokenIndex, depth
						if buffer[position] != rune(']') {
							goto l117
						}
						position++
						if buffer[position] != rune(']') {
							goto l117
						}
						position++
						goto l116
					l117:
						position, tokenIndex, depth = position117, tokenIndex117, depth117
					}
					if !_rules[ruleDoubleRange]() {
						goto l116
					}
					if !_rules[ruleAction25]() {
						goto l116
					}
					goto l115
				l116:
					position, tokenIndex, depth = position116, tokenIndex116, depth116
				}
				depth--
				add(ruleDoubleRanges, position113)
			}
			return true
		l112:
			position, tokenIndex, depth = position112, tokenIndex112, depth112
			return false
		},
		/* 15 Range <- <((Char '-' Char Action26) / Char)> */
		func() bool {
			position118, tokenIndex118, depth118 := position, tokenIndex, depth
			{
				position119 := position
				depth++
				{
					position120, tokenIndex120, depth120 := position, tokenIndex, depth
					if !_rules[ruleChar]() {
						goto l121
					}
					if buffer[position] != rune('-') {
						goto l121
					}
					position++
					if !_rules[ruleChar]() {
						goto l121
					}
					if !_rules[ruleAction26]() {
						goto l121
					}
					goto l120
				l121:
					position, tokenIndex, depth = position120, tokenIndex120, depth120
					if !_rules[ruleChar]() {
						goto l118
					}
				}
			l120:
				depth--
				add(ruleRange, position119)
			}
			return true
		l118:
			position, tokenIndex, depth = position118, tokenIndex118, depth118
			return false
		},
		/* 16 DoubleRange <- <((Char '-' Char Action27) / DoubleChar)> */
		func() bool {
			position122, tokenIndex122, depth122 := position, tokenIndex, depth
			{
				position123 := position
				depth++
				{
					position124, tokenIndex124, depth124 := position, tokenIndex, depth
					if !_rules[ruleChar]() {
						goto l125
					}
					if buffer[position] != rune('-') {
						goto l125
					}
					position++
					if !_rules[ruleChar]() {
						goto l125
					}
					if !_rules[ruleAction27]() {
						goto l125
					}
					goto l124
				l125:
					position, tokenIndex, depth = position124, tokenIndex124, depth124
					if !_rules[ruleDoubleChar]() {
						goto l122
					}
				}
			l124:
				depth--
				add(ruleDoubleRange, position123)
			}
			return true
		l122:
			position, tokenIndex, depth = position122, tokenIndex122, depth122
			return false
		},
		/* 17 Char <- <(Escape / (!'\\' <.> Action28))> */
		func() bool {
			position126, tokenIndex126, depth126 := position, tokenIndex, depth
			{
				position127 := position
				depth++
				{
					position128, tokenIndex128, depth128 := position, tokenIndex, depth
					if !_rules[ruleEscape]() {
						goto l129
					}
					goto l128
				l129:
					position, tokenIndex, depth = position128, tokenIndex128, depth128
					{
						position130, tokenIndex130, depth130 := position, tokenIndex, depth
						if buffer[position] != rune('\\') {
							goto l130
						}
						position++
						goto l126
					l130:
						position, tokenIndex, depth = position130, tokenIndex130, depth130
					}
					{
						position131 := position
						depth++
						if !matchDot() {
							goto l126
						}
						depth--
						add(rulePegText, position131)
					}
					if !_rules[ruleAction28]() {
						goto l126
					}
				}
			l128:
				depth--
				add(ruleChar, position127)
			}
			return true
		l126:
			position, tokenIndex, depth = position126, tokenIndex126, depth126
			return false
		},
		/* 18 DoubleChar <- <(Escape / (<([a-z] / [A-Z])> Action29) / (!'\\' <.> Action30))> */
		func() bool {
			position132, tokenIndex132, depth132 := position, tokenIndex, depth
			{
				position133 := position
				depth++
				{
					position134, tokenIndex134, depth134 := position, tokenIndex, depth
					if !_rules[ruleEscape]() {
						goto l135
					}
					goto l134
				l135:
					position, tokenIndex, depth = position134, tokenIndex134, depth134
					{
						position137 := position
						depth++
						{
							position138, tokenIndex138, depth138 := position, tokenIndex, depth
							if c := buffer[position]; c < rune('a') || c > rune('z') {
								goto l139
							}
							position++
							goto l138
						l139:
							position, tokenIndex, depth = position138, tokenIndex138, depth138
							if c := buffer[position]; c < rune('A') || c > rune('Z') {
								goto l136
							}
							position++
						}
					l138:
						depth--
						add(rulePegText, position137)
					}
					if !_rules[ruleAction29]() {
						goto l136
					}
					goto l134
				l136:
					position, tokenIndex, depth = position134, tokenIndex134, depth134
					{
						position140, tokenIndex140, depth140 := position, tokenIndex, depth
						if buffer[position] != rune('\\') {
							goto l140
						}
						position++
						goto l132
					l140:
						position, tokenIndex, depth = position140, tokenIndex140, depth140
					}
					{
						position141 := position
						depth++
						if !matchDot() {
							goto l132
						}
						depth--
						add(rulePegText, position141)
					}
					if !_rules[ruleAction30]() {
						goto l132
					}
				}
			l134:
				depth--
				add(ruleDoubleChar, position133)
			}
			return true
		l132:
			position, tokenIndex, depth = position132, tokenIndex132, depth132
			return false
		},
		/* 19 Escape <- <(('\\' ('a' / 'A') Action31) / ('\\' ('b' / 'B') Action32) / ('\\' ('e' / 'E') Action33) / ('\\' ('f' / 'F') Action34) / ('\\' ('n' / 'N') Action35) / ('\\' ('r' / 'R') Action36) / ('\\' ('t' / 'T') Action37) / ('\\' ('v' / 'V') Action38) / ('\\' '\'' Action39) / ('\\' '"' Action40) / ('\\' '[' Action41) / ('\\' ']' Action42) / ('\\' '-' Action43) / ('\\' ('0' ('x' / 'X')) <([0-9] / [a-f] / [A-F])+> Action44) / ('\\' <([0-3] [0-7] [0-7])> Action45) / ('\\' <([0-7] [0-7]?)> Action46) / ('\\' '\\' Action47))> */
		func() bool {
			position142, tokenIndex142, depth142 := position, tokenIndex, depth
			{
				position143 := position
				depth++
				{
					position144, tokenIndex144, depth144 := position, tokenIndex, depth
					if buffer[position] != rune('\\') {
						goto l145
					}
					position++
					{
						position146, tokenIndex146, depth146 := position, tokenIndex, depth
						if buffer[position] != rune('a') {
							goto l147
						}
						position++
						goto l146
					l147:
						position, tokenIndex, depth = position146, tokenIndex146, depth146
						if buffer[position] != rune('A') {
							goto l145
						}
						position++
					}
				l146:
					if !_rules[ruleAction31]() {
						goto l145
					}
					goto l144
				l145:
					position, tokenIndex, depth = position144, tokenIndex144, depth144
					if buffer[position] != rune('\\') {
						goto l148
					}
					position++
					{
						position149, tokenIndex149, depth149 := position, tokenIndex, depth
						if buffer[position] != rune('b') {
							goto l150
						}
						position++
						goto l149
					l150:
						position, tokenIndex, depth = position149, tokenIndex149, depth149
						if buffer[position] != rune('B') {
							goto l148
						}
						position++
					}
				l149:
					if !_rules[ruleAction32]() {
						goto l148
					}
					goto l144
				l148:
					position, tokenIndex, depth = position144, tokenIndex144, depth144
					if buffer[position] != rune('\\') {
						goto l151
					}
					position++
					{
						position152, tokenIndex152, depth152 := position, tokenIndex, depth
						if buffer[position] != rune('e') {
							goto l153
						}
						position++
						goto l152
					l153:
						position, tokenIndex, depth = position152, tokenIndex152, depth152
						if buffer[position] != rune('E') {
							goto l151
						}
						position++
					}
				l152:
					if !_rules[ruleAction33]() {
						goto l151
					}
					goto l144
				l151:
					position, tokenIndex, depth = position144, tokenIndex144, depth144
					if buffer[position] != rune('\\') {
						goto l154
					}
					position++
					{
						position155, tokenIndex155, depth155 := position, tokenIndex, depth
						if buffer[position] != rune('f') {
							goto l156
						}
						position++
						goto l155
					l156:
						position, tokenIndex, depth = position155, tokenIndex155, depth155
						if buffer[position] != rune('F') {
							goto l154
						}
						position++
					}
				l155:
					if !_rules[ruleAction34]() {
						goto l154
					}
					goto l144
				l154:
					position, tokenIndex, depth = position144, tokenIndex144, depth144
					if buffer[position] != rune('\\') {
						goto l157
					}
					position++
					{
						position158, tokenIndex158, depth158 := position, tokenIndex, depth
						if buffer[position] != rune('n') {
							goto l159
						}
						position++
						goto l158
					l159:
						position, tokenIndex, depth = position158, tokenIndex158, depth158
						if buffer[position] != rune('N') {
							goto l157
						}
						position++
					}
				l158:
					if !_rules[ruleAction35]() {
						goto l157
					}
					goto l144
				l157:
					position, tokenIndex, depth = position144, tokenIndex144, depth144
					if buffer[position] != rune('\\') {
						goto l160
					}
					position++
					{
						position161, tokenIndex161, depth161 := position, tokenIndex, depth
						if buffer[position] != rune('r') {
							goto l162
						}
						position++
						goto l161
					l162:
						position, tokenIndex, depth = position161, tokenIndex161, depth161
						if buffer[position] != rune('R') {
							goto l160
						}
						position++
					}
				l161:
					if !_rules[ruleAction36]() {
						goto l160
					}
					goto l144
				l160:
					position, tokenIndex, depth = position144, tokenIndex144, depth144
					if buffer[position] != rune('\\') {
						goto l163
					}
					position++
					{
						position164, tokenIndex164, depth164 := position, tokenIndex, depth
						if buffer[position] != rune('t') {
							goto l165
						}
						position++
						goto l164
					l165:
						position, tokenIndex, depth = position164, tokenIndex164, depth164
						if buffer[position] != rune('T') {
							goto l163
						}
						position++
					}
				l164:
					if !_rules[ruleAction37]() {
						goto l163
					}
					goto l144
				l163:
					position, tokenIndex, depth = position144, tokenIndex144, depth144
					if buffer[position] != rune('\\') {
						goto l166
					}
					position++
					{
						position167, tokenIndex167, depth167 := position, tokenIndex, depth
						if buffer[position] != rune('v') {
							goto l168
						}
						position++
						goto l167
					l168:
						position, tokenIndex, depth = position167, tokenIndex167, depth167
						if buffer[position] != rune('V') {
							goto l166
						}
						position++
					}
				l167:
					if !_rules[ruleAction38]() {
						goto l166
					}
					goto l144
				l166:
					position, tokenIndex, depth = position144, tokenIndex144, depth144
					if buffer[position] != rune('\\') {
						goto l169
					}
					position++
					if buffer[position] != rune('\'') {
						goto l169
					}
					position++
					if !_rules[ruleAction39]() {
						goto l169
					}
					goto l144
				l169:
					position, tokenIndex, depth = position144, tokenIndex144, depth144
					if buffer[position] != rune('\\') {
						goto l170
					}
					position++
					if buffer[position] != rune('"') {
						goto l170
					}
					position++
					if !_rules[ruleAction40]() {
						goto l170
					}
					goto l144
				l170:
					position, tokenIndex, depth = position144, tokenIndex144, depth144
					if buffer[position] != rune('\\') {
						goto l171
					}
					position++
					if buffer[position] != rune('[') {
						goto l171
					}
					position++
					if !_rules[ruleAction41]() {
						goto l171
					}
					goto l144
				l171:
					position, tokenIndex, depth = position144, tokenIndex144, depth144
					if buffer[position] != rune('\\') {
						goto l172
					}
					position++
					if buffer[position] != rune(']') {
						goto l172
					}
					position++
					if !_rules[ruleAction42]() {
						goto l172
					}
					goto l144
				l172:
					position, tokenIndex, depth = position144, tokenIndex144, depth144
					if buffer[position] != rune('\\') {
						goto l173
					}
					position++
					if buffer[position] != rune('-') {
						goto l173
					}
					position++
					if !_rules[ruleAction43]() {
						goto l173
					}
					goto l144
				l173:
					position, tokenIndex, depth = position144, tokenIndex144, depth144
					if buffer[position] != rune('\\') {
						goto l174
					}
					position++
					if buffer[position] != rune('0') {
						goto l174
					}
					position++
					{
						position175, tokenIndex175, depth175 := position, tokenIndex, depth
						if buffer[position] != rune('x') {
							goto l176
						}
						position++
						goto l175
					l176:
						position, tokenIndex, depth = position175, tokenIndex175, depth175
						if buffer[position] != rune('X') {
							goto l174
						}
						position++
					}
				l175:
					{
						position177 := position
						depth++
						{
							position180, tokenIndex180, depth180 := position, tokenIndex, depth
							if c := buffer[position]; c < rune('0') || c > rune('9') {
								goto l181
							}
							position++
							goto l180
						l181:
							position, tokenIndex, depth = position180, tokenIndex180, depth180
							if c := buffer[position]; c < rune('a') || c > rune('f') {
								goto l182
							}
							position++
							goto l180
						l182:
							position, tokenIndex, depth = position180, tokenIndex180, depth180
							if c := buffer[position]; c < rune('A') || c > rune('F') {
								goto l174
							}
							position++
						}
					l180:
					l178:
						{
							position179, tokenIndex179, depth179 := position, tokenIndex, depth
							{
								position183, tokenIndex183, depth183 := position, tokenIndex, depth
								if c := buffer[position]; c < rune('0') || c > rune('9') {
									goto l184
								}
								position++
								goto l183
							l184:
								position, tokenIndex, depth = position183, tokenIndex183, depth183
								if c := buffer[position]; c < rune('a') || c > rune('f') {
									goto l185
								}
								position++
								goto l183
							l185:
								position, tokenIndex, depth = position183, tokenIndex183, depth183
								if c := buffer[position]; c < rune('A') || c > rune('F') {
									goto l179
								}
								position++
							}
						l183:
							goto l178
						l179:
							position, tokenIndex, depth = position179, tokenIndex179, depth179
						}
						depth--
						add(rulePegText, position177)
					}
					if !_rules[ruleAction44]() {
						goto l174
					}
					goto l144
				l174:
					position, tokenIndex, depth = position144, tokenIndex144, depth144
					if buffer[position] != rune('\\') {
						goto l186
					}
					position++
					{
						position187 := position
						depth++
						if c := buffer[position]; c < rune('0') || c > rune('3') {
							goto l186
						}
						position++
						if c := buffer[position]; c < rune('0') || c > rune('7') {
							goto l186
						}
						position++
						if c := buffer[position]; c < rune('0') || c > rune('7') {
							goto l186
						}
						position++
						depth--
						add(rulePegText, position187)
					}
					if !_rules[ruleAction45]() {
						goto l186
					}
					goto l144
				l186:
					position, tokenIndex, depth = position144, tokenIndex144, depth144
					if buffer[position] != rune('\\') {
						goto l188
					}
					position++
					{
						position189 := position
						depth++
						if c := buffer[position]; c < rune('0') || c > rune('7') {
							goto l188
						}
						position++
						{
							position190, tokenIndex190, depth190 := position, tokenIndex, depth
							if c := buffer[position]; c < rune('0') || c > rune('7') {
								goto l190
							}
							position++
							goto l191
						l190:
							position, tokenIndex, depth = position190, tokenIndex190, depth190
						}
					l191:
						depth--
						add(rulePegText, position189)
					}
					if !_rules[ruleAction46]() {
						goto l188
					}
					goto l144
				l188:
					position, tokenIndex, depth = position144, tokenIndex144, depth144
					if buffer[position] != rune('\\') {
						goto l142
					}
					position++
					if buffer[position] != rune('\\') {
						goto l142
					}
					position++
					if !_rules[ruleAction47]() {
						goto l142
					}
				}
			l144:
				depth--
				add(ruleEscape, position143)
			}
			return true
		l142:
			position, tokenIndex, depth = position142, tokenIndex142, depth142
			return false
		},
		/* 20 LeftArrow <- <('<' '-' Spacing)> */
		func() bool {
			position192, tokenIndex192, depth192 := position, tokenIndex, depth
			{
				position193 := position
				depth++
				if buffer[position] != rune('<') {
					goto l192
				}
				position++
				if buffer[position] != rune('-') {
					goto l192
				}
				position++
				if !_rules[ruleSpacing]() {
					goto l192
				}
				depth--
				add(ruleLeftArrow, position193)
			}
			return true
		l192:
			position, tokenIndex, depth = position192, tokenIndex192, depth192
			return false
		},
		/* 21 Slash <- <('/' Spacing)> */
		func() bool {
			position194, tokenIndex194, depth194 := position, tokenIndex, depth
			{
				position195 := position
				depth++
				if buffer[position] != rune('/') {
					goto l194
				}
				position++
				if !_rules[ruleSpacing]() {
					goto l194
				}
				depth--
				add(ruleSlash, position195)
			}
			return true
		l194:
			position, tokenIndex, depth = position194, tokenIndex194, depth194
			return false
		},
		/* 22 And <- <('&' Spacing)> */
		func() bool {
			position196, tokenIndex196, depth196 := position, tokenIndex, depth
			{
				position197 := position
				depth++
				if buffer[position] != rune('&') {
					goto l196
				}
				position++
				if !_rules[ruleSpacing]() {
					goto l196
				}
				depth--
				add(ruleAnd, position197)
			}
			return true
		l196:
			position, tokenIndex, depth = position196, tokenIndex196, depth196
			return false
		},
		/* 23 Not <- <('!' Spacing)> */
		func() bool {
			position198, tokenIndex198, depth198 := position, tokenIndex, depth
			{
				position199 := position
				depth++
				if buffer[position] != rune('!') {
					goto l198
				}
				position++
				if !_rules[ruleSpacing]() {
					goto l198
				}
				depth--
				add(ruleNot, position199)
			}
			return true
		l198:
			position, tokenIndex, depth = position198, tokenIndex198, depth198
			return false
		},
		/* 24 Question <- <('?' Spacing)> */
		func() bool {
			position200, tokenIndex200, depth200 := position, tokenIndex, depth
			{
				position201 := position
				depth++
				if buffer[position] != rune('?') {
					goto l200
				}
				position++
				if !_rules[ruleSpacing]() {
					goto l200
				}
				depth--
				add(ruleQuestion, position201)
			}
			return true
		l200:
			position, tokenIndex, depth = position200, tokenIndex200, depth200
			return false
		},
		/* 25 Star <- <('*' Spacing)> */
		func() bool {
			position202, tokenIndex202, depth202 := position, tokenIndex, depth
			{
				position203 := position
				depth++
				if buffer[position] != rune('*') {
					goto l202
				}
				position++
				if !_rules[ruleSpacing]() {
					goto l202
				}
				depth--
				add(ruleStar, position203)
			}
			return true
		l202:
			position, tokenIndex, depth = position202, tokenIndex202, depth202
			return false
		},
		/* 26 Plus <- <('+' Spacing)> */
		func() bool {
			position204, tokenIndex204, depth204 := position, tokenIndex, depth
			{
				position205 := position
				depth++
				if buffer[position] != rune('+') {
					goto l204
				}
				position++
				if !_rules[ruleSpacing]() {
					goto l204
				}
				depth--
				add(rulePlus, position205)
			}
			return true
		l204:
			position, tokenIndex, depth = position204, tokenIndex204, depth204
			return false
		},
		/* 27 Open <- <('(' Spacing)> */
		func() bool {
			position206, tokenIndex206, depth206 := position, tokenIndex, depth
			{
				position207 := position
				depth++
				if buffer[position] != rune('(') {
					goto l206
				}
				position++
				if !_rules[ruleSpacing]() {
					goto l206
				}
				depth--
				add(ruleOpen, position207)
			}
			return true
		l206:
			position, tokenIndex, depth = position206, tokenIndex206, depth206
			return false
		},
		/* 28 Close <- <(')' Spacing)> */
		func() bool {
			position208, tokenIndex208, depth208 := position, tokenIndex, depth
			{
				position209 := position
				depth++
				if buffer[position] != rune(')') {
					goto l208
				}
				position++
				if !_rules[ruleSpacing]() {
					goto l208
				}
				depth--
				add(ruleClose, position209)
			}
			return true
		l208:
			position, tokenIndex, depth = position208, tokenIndex208, depth208
			return false
		},
		/* 29 Dot <- <('.' Spacing)> */
		func() bool {
			position210, tokenIndex210, depth210 := position, tokenIndex, depth
			{
				position211 := position
				depth++
				if buffer[position] != rune('.') {
					goto l210
				}
				position++
				if !_rules[ruleSpacing]() {
					goto l210
				}
				depth--
				add(ruleDot, position211)
			}
			return true
		l210:
			position, tokenIndex, depth = position210, tokenIndex210, depth210
			return false
		},
		/* 30 SpaceComment <- <(Space / Comment)> */
		func() bool {
			position212, tokenIndex212, depth212 := position, tokenIndex, depth
			{
				position213 := position
				depth++
				{
					position214, tokenIndex214, depth214 := position, tokenIndex, depth
					if !_rules[ruleSpace]() {
						goto l215
					}
					goto l214
				l215:
					position, tokenIndex, depth = position214, tokenIndex214, depth214
					if !_rules[ruleComment]() {
						goto l212
					}
				}
			l214:
				depth--
				add(ruleSpaceComment, position213)
			}
			return true
		l212:
			position, tokenIndex, depth = position212, tokenIndex212, depth212
			return false
		},
		/* 31 Spacing <- <SpaceComment*> */
		func() bool {
			{
				position217 := position
				depth++
			l218:
				{
					position219, tokenIndex219, depth219 := position, tokenIndex, depth
					if !_rules[ruleSpaceComment]() {
						goto l219
					}
					goto l218
				l219:
					position, tokenIndex, depth = position219, tokenIndex219, depth219
				}
				depth--
				add(ruleSpacing, position217)
			}
			return true
		},
		/* 32 MustSpacing <- <SpaceComment+> */
		func() bool {
			position220, tokenIndex220, depth220 := position, tokenIndex, depth
			{
				position221 := position
				depth++
				if !_rules[ruleSpaceComment]() {
					goto l220
				}
			l222:
				{
					position223, tokenIndex223, depth223 := position, tokenIndex, depth
					if !_rules[ruleSpaceComment]() {
						goto l223
					}
					goto l222
				l223:
					position, tokenIndex, depth = position223, tokenIndex223, depth223
				}
				depth--
				add(ruleMustSpacing, position221)
			}
			return true
		l220:
			position, tokenIndex, depth = position220, tokenIndex220, depth220
			return false
		},
		/* 33 Comment <- <('#' (!EndOfLine .)* EndOfLine)> */
		func() bool {
			position224, tokenIndex224, depth224 := position, tokenIndex, depth
			{
				position225 := position
				depth++
				if buffer[position] != rune('#') {
					goto l224
				}
				position++
			l226:
				{
					position227, tokenIndex227, depth227 := position, tokenIndex, depth
					{
						position228, tokenIndex228, depth228 := position, tokenIndex, depth
						if !_rules[ruleEndOfLine]() {
							goto l228
						}
						goto l227
					l228:
						position, tokenIndex, depth = position228, tokenIndex228, depth228
					}
					if !matchDot() {
						goto l227
					}
					goto l226
				l227:
					position, tokenIndex, depth = position227, tokenIndex227, depth227
				}
				if !_rules[ruleEndOfLine]() {
					goto l224
				}
				depth--
				add(ruleComment, position225)
			}
			return true
		l224:
			position, tokenIndex, depth = position224, tokenIndex224, depth224
			return false
		},
		/* 34 Space <- <(' ' / '\t' / EndOfLine)> */
		func() bool {
			position229, tokenIndex229, depth229 := position, tokenIndex, depth
			{
				position230 := position
				depth++
				{
					position231, tokenIndex231, depth231 := position, tokenIndex, depth
					if buffer[position] != rune(' ') {
						goto l232
					}
					position++
					goto l231
				l232:
					position, tokenIndex, depth = position231, tokenIndex231, depth231
					if buffer[position] != rune('\t') {
						goto l233
					}
					position++
					goto l231
				l233:
					position, tokenIndex, depth = position231, tokenIndex231, depth231
					if !_rules[ruleEndOfLine]() {
						goto l229
					}
				}
			l231:
				depth--
				add(ruleSpace, position230)
			}
			return true
		l229:
			position, tokenIndex, depth = position229, tokenIndex229, depth229
			return false
		},
		/* 35 EndOfLine <- <(('\r' '\n') / '\n' / '\r')> */
		func() bool {
			position234, tokenIndex234, depth234 := position, tokenIndex, depth
			{
				position235 := position
				depth++
				{
					position236, tokenIndex236, depth236 := position, tokenIndex, depth
					if buffer[position] != rune('\r') {
						goto l237
					}
					position++
					if buffer[position] != rune('\n') {
						goto l237
					}
					position++
					goto l236
				l237:
					position, tokenIndex, depth = position236, tokenIndex236, depth236
					if buffer[position] != rune('\n') {
						goto l238
					}
					position++
					goto l236
				l238:
					position, tokenIndex, depth = position236, tokenIndex236, depth236
					if buffer[position] != rune('\r') {
						goto l234
					}
					position++
				}
			l236:
				depth--
				add(ruleEndOfLine, position235)
			}
			return true
		l234:
			position, tokenIndex, depth = position234, tokenIndex234, depth234
			return false
		},
		/* 36 EndOfFile <- <!.> */
		func() bool {
			position239, tokenIndex239, depth239 := position, tokenIndex, depth
			{
				position240 := position
				depth++
				{
					position241, tokenIndex241, depth241 := position, tokenIndex, depth
					if !matchDot() {
						goto l241
					}
					goto l239
				l241:
					position, tokenIndex, depth = position241, tokenIndex241, depth241
				}
				depth--
				add(ruleEndOfFile, position240)
			}
			return true
		l239:
			position, tokenIndex, depth = position239, tokenIndex239, depth239
			return false
		},
		/* 37 Action <- <('{' <(!'}' .)*> '}' Spacing)> */
		func() bool {
			position242, tokenIndex242, depth242 := position, tokenIndex, depth
			{
				position243 := position
				depth++
				if buffer[position] != rune('{') {
					goto l242
				}
				position++
				{
					position244 := position
					depth++
				l245:
					{
						position246, tokenIndex246, depth246 := position, tokenIndex, depth
						{
							position247, tokenIndex247, depth247 := position, tokenIndex, depth
							if buffer[position] != rune('}') {
								goto l247
							}
							position++
							goto l246
						l247:
							position, tokenIndex, depth = position247, tokenIndex247, depth247
						}
						if !matchDot() {
							goto l246
						}
						goto l245
					l246:
						position, tokenIndex, depth = position246, tokenIndex246, depth246
					}
					depth--
					add(rulePegText, position244)
				}
				if buffer[position] != rune('}') {
					goto l242
				}
				position++
				if !_rules[ruleSpacing]() {
					goto l242
				}
				depth--
				add(ruleAction, position243)
			}
			return true
		l242:
			position, tokenIndex, depth = position242, tokenIndex242, depth242
			return false
		},
		/* 38 Begin <- <('<' Spacing)> */
		func() bool {
			position248, tokenIndex248, depth248 := position, tokenIndex, depth
			{
				position249 := position
				depth++
				if buffer[position] != rune('<') {
					goto l248
				}
				position++
				if !_rules[ruleSpacing]() {
					goto l248
				}
				depth--
				add(ruleBegin, position249)
			}
			return true
		l248:
			position, tokenIndex, depth = position248, tokenIndex248, depth248
			return false
		},
		/* 39 End <- <('>' Spacing)> */
		func() bool {
			position250, tokenIndex250, depth250 := position, tokenIndex, depth
			{
				position251 := position
				depth++
				if buffer[position] != rune('>') {
					goto l250
				}
				position++
				if !_rules[ruleSpacing]() {
					goto l250
				}
				depth--
				add(ruleEnd, position251)
			}
			return true
		l250:
			position, tokenIndex, depth = position250, tokenIndex250, depth250
			return false
		},
		/* 40 runtimeGetruntime <- <('r' 'u' 'n' 't' 'i' 'm' 'e' Dot ('g' 'e' 't' 'r' 'u' 'n' 't' 'i' 'm' 'e') Action48)> */
		nil,
		/* 41 runtimeRuntime <- <('r' 'u' 'n' 't' 'i' 'm' 'e' Dot ('r' 'u' 'n' 't' 'i' 'm' 'e') Action49)> */
		nil,
		/* 42 runtimePorts <- <('r' 'u' 'n' 't' 'i' 'm' 'e' Dot ('p' 'o' 'r' 't' 's') Action50)> */
		nil,
		/* 43 runtimePacket <- <('r' 'u' 'n' 't' 'i' 'm' 'e' Dot ('p' 'a' 'c' 'k' 'e' 't') Action51)> */
		nil,
		/* 44 graphClear <- <('g' 'r' 'a' 'p' 'h' Dot ('c' 'l' 'e' 'a' 'r') Action52)> */
		nil,
		/* 45 graphAddnode <- <('g' 'r' 'a' 'p' 'h' Dot ('a' 'd' 'd' 'n' 'o' 'd' 'e') Action53)> */
		nil,
		/* 46 graphRemovenode <- <('g' 'r' 'a' 'p' 'h' Dot ('r' 'e' 'm' 'o' 'v' 'e' 'n' 'o' 'd' 'e') Action54)> */
		nil,
		/* 47 graphRenamenode <- <('g' 'r' 'a' 'p' 'h' Dot ('r' 'e' 'n' 'a' 'm' 'e' 'n' 'o' 'd' 'e') Action55)> */
		nil,
		/* 48 graphChangenode <- <('g' 'r' 'a' 'p' 'h' Dot ('c' 'h' 'a' 'n' 'g' 'e' 'n' 'o' 'd' 'e') Action56)> */
		nil,
		/* 49 graphAddedge <- <('g' 'r' 'a' 'p' 'h' Dot ('a' 'd' 'd' 'e' 'd' 'g' 'e') Action57)> */
		nil,
		/* 50 graphRemoveedge <- <('g' 'r' 'a' 'p' 'h' Dot ('r' 'e' 'm' 'o' 'v' 'e' 'e' 'd' 'g' 'e') Action58)> */
		nil,
		/* 51 graphChangeedge <- <('g' 'r' 'a' 'p' 'h' Dot ('c' 'h' 'a' 'n' 'g' 'e' 'e' 'd' 'g' 'e') Action59)> */
		nil,
		/* 52 graphAddinitial <- <('g' 'r' 'a' 'p' 'h' Dot ('a' 'd' 'd' 'i' 'n' 'i' 't' 'i' 'a' 'l') Action60)> */
		nil,
		/* 53 graphRemoveinitial <- <('g' 'r' 'a' 'p' 'h' Dot ('r' 'e' 'm' 'o' 'v' 'e' 'i' 'n' 'i' 't' 'i' 'a' 'l') Action61)> */
		nil,
		/* 54 graphAddinport <- <('g' 'r' 'a' 'p' 'h' Dot ('a' 'd' 'd' 'i' 'n' 'p' 'o' 'r' 't') Action62)> */
		nil,
		/* 55 graphRemoveinport <- <('g' 'r' 'a' 'p' 'h' Dot ('r' 'e' 'm' 'o' 'v' 'e' 'i' 'n' 'p' 'o' 'r' 't') Action63)> */
		nil,
		/* 56 graphRenameinport <- <('g' 'r' 'a' 'p' 'h' Dot ('r' 'e' 'n' 'a' 'm' 'e' 'i' 'n' 'p' 'o' 'r' 't') Action64)> */
		nil,
		/* 57 graphAddoutport <- <('g' 'r' 'a' 'p' 'h' Dot ('a' 'd' 'd' 'o' 'u' 't' 'p' 'o' 'r' 't') Action65)> */
		nil,
		/* 58 graphRemoveoutport <- <('g' 'r' 'a' 'p' 'h' Dot ('r' 'e' 'm' 'o' 'v' 'e' 'o' 'u' 't' 'p' 'o' 'r' 't') Action66)> */
		nil,
		/* 59 graphRenameoutport <- <('g' 'r' 'a' 'p' 'h' Dot ('r' 'e' 'n' 'a' 'm' 'e' 'o' 'u' 't' 'p' 'o' 'r' 't') Action67)> */
		nil,
		/* 60 graphAddgroup <- <('g' 'r' 'a' 'p' 'h' Dot ('a' 'd' 'd' 'g' 'r' 'o' 'u' 'p') Action68)> */
		nil,
		/* 61 graphRemovegroup <- <('g' 'r' 'a' 'p' 'h' Dot ('r' 'e' 'm' 'o' 'v' 'e' 'g' 'r' 'o' 'u' 'p') Action69)> */
		nil,
		/* 62 graphRenamegroup <- <('g' 'r' 'a' 'p' 'h' Dot ('r' 'e' 'n' 'a' 'm' 'e' 'g' 'r' 'o' 'u' 'p') Action70)> */
		nil,
		/* 63 graphChangegroup <- <('g' 'r' 'a' 'p' 'h' Dot ('c' 'h' 'a' 'n' 'g' 'e' 'g' 'r' 'o' 'u' 'p') Action71)> */
		nil,
		/* 64 componentList <- <('c' 'o' 'm' 'p' 'o' 'n' 'e' 'n' 't' Dot ('l' 'i' 's' 't') Action72)> */
		nil,
		/* 65 componentComponent <- <('c' 'o' 'm' 'p' 'o' 'n' 'e' 'n' 't' Dot ('c' 'o' 'm' 'p' 'o' 'n' 'e' 'n' 't') Action73)> */
		nil,
		/* 66 componentGetsource <- <('c' 'o' 'm' 'p' 'o' 'n' 'e' 'n' 't' Dot ('g' 'e' 't' 's' 'o' 'u' 'r' 'c' 'e') Action74)> */
		nil,
		/* 67 componentSource <- <('c' 'o' 'm' 'p' 'o' 'n' 'e' 'n' 't' Dot ('s' 'o' 'u' 'r' 'c' 'e') Action75)> */
		nil,
		/* 68 networkStart <- <('n' 'e' 't' 'w' 'o' 'r' 'k' Dot ('s' 't' 'a' 'r' 't') Action76)> */
		nil,
		/* 69 networkGetstatus <- <('n' 'e' 't' 'w' 'o' 'r' 'k' Dot ('g' 'e' 't' 's' 't' 'a' 't' 'u' 's') Action77)> */
		nil,
		/* 70 networkStop <- <('n' 'e' 't' 'w' 'o' 'r' 'k' Dot ('s' 't' 'o' 'p') Action78)> */
		nil,
		/* 71 networkStarted <- <('n' 'e' 't' 'w' 'o' 'r' 'k' Dot ('s' 't' 'a' 'r' 't' 'e' 'd') Action79)> */
		nil,
		/* 72 networkStatus <- <('n' 'e' 't' 'w' 'o' 'r' 'k' Dot ('s' 't' 'a' 't' 'u' 's') Action80)> */
		nil,
		/* 73 networkStopped <- <('n' 'e' 't' 'w' 'o' 'r' 'k' Dot ('s' 't' 'o' 'p' 'p' 'e' 'd') Action81)> */
		nil,
		/* 74 networkDebug <- <('n' 'e' 't' 'w' 'o' 'r' 'k' Dot ('d' 'e' 'b' 'u' 'g') Action82)> */
		nil,
		/* 75 networkIcon <- <('n' 'e' 't' 'w' 'o' 'r' 'k' Dot ('i' 'c' 'o' 'n') Action83)> */
		nil,
		/* 76 networkOutput <- <('n' 'e' 't' 'w' 'o' 'r' 'k' Dot ('o' 'u' 't' 'p' 'u' 't') Action84)> */
		nil,
		/* 77 networkError <- <('n' 'e' 't' 'w' 'o' 'r' 'k' Dot ('e' 'r' 'r' 'o' 'r') Action85)> */
		nil,
		/* 78 networkProcesserror <- <('n' 'e' 't' 'w' 'o' 'r' 'k' Dot ('p' 'r' 'o' 'c' 'e' 's' 's' 'e' 'r' 'r' 'o' 'r') Action86)> */
		nil,
		/* 79 networkConnect <- <('n' 'e' 't' 'w' 'o' 'r' 'k' Dot ('c' 'o' 'n' 'n' 'e' 'c' 't') Action87)> */
		nil,
		/* 80 networkBegingroup <- <('n' 'e' 't' 'w' 'o' 'r' 'k' Dot ('b' 'e' 'g' 'i' 'n' 'g' 'r' 'o' 'u' 'p') Action88)> */
		nil,
		/* 81 networkData <- <('n' 'e' 't' 'w' 'o' 'r' 'k' Dot ('d' 'a' 't' 'a') Action89)> */
		nil,
		/* 82 networkEndgroup <- <('n' 'e' 't' 'w' 'o' 'r' 'k' Dot ('e' 'n' 'd' 'g' 'r' 'o' 'u' 'p') Action90)> */
		nil,
		/* 83 networkDisconnect <- <('n' 'e' 't' 'w' 'o' 'r' 'k' Dot ('d' 'i' 's' 'c' 'o' 'n' 'n' 'e' 'c' 't') Action91)> */
		nil,
		/* 84 networkEdges <- <('n' 'e' 't' 'w' 'o' 'r' 'k' Dot ('e' 'd' 'g' 'e' 's') Action92)> */
		nil,
		/* 86 Action0 <- <{ p.AddPackage(buffer[begin:end]) }> */
		func() bool {
			{
				add(ruleAction0, position)
			}
			return true
		},
		/* 87 Action1 <- <{ p.AddPeg(buffer[begin:end]) }> */
		func() bool {
			{
				add(ruleAction1, position)
			}
			return true
		},
		/* 88 Action2 <- <{ p.AddState(buffer[begin:end]) }> */
		func() bool {
			{
				add(ruleAction2, position)
			}
			return true
		},
		nil,
		/* 90 Action3 <- <{ p.AddImport(buffer[begin:end]) }> */
		func() bool {
			{
				add(ruleAction3, position)
			}
			return true
		},
		/* 91 Action4 <- <{ p.AddRule(buffer[begin:end]) }> */
		func() bool {
			{
				add(ruleAction4, position)
			}
			return true
		},
		/* 92 Action5 <- <{ p.AddExpression() }> */
		func() bool {
			{
				add(ruleAction5, position)
			}
			return true
		},
		/* 93 Action6 <- <{ p.AddAlternate() }> */
		func() bool {
			{
				add(ruleAction6, position)
			}
			return true
		},
		/* 94 Action7 <- <{ p.AddNil(); p.AddAlternate() }> */
		func() bool {
			{
				add(ruleAction7, position)
			}
			return true
		},
		/* 95 Action8 <- <{ p.AddNil() }> */
		func() bool {
			{
				add(ruleAction8, position)
			}
			return true
		},
		/* 96 Action9 <- <{ p.AddSequence() }> */
		func() bool {
			{
				add(ruleAction9, position)
			}
			return true
		},
		/* 97 Action10 <- <{ p.AddPredicate(buffer[begin:end]) }> */
		func() bool {
			{
				add(ruleAction10, position)
			}
			return true
		},
		/* 98 Action11 <- <{ p.AddPeekFor() }> */
		func() bool {
			{
				add(ruleAction11, position)
			}
			return true
		},
		/* 99 Action12 <- <{ p.AddPeekNot() }> */
		func() bool {
			{
				add(ruleAction12, position)
			}
			return true
		},
		/* 100 Action13 <- <{ p.AddQuery() }> */
		func() bool {
			{
				add(ruleAction13, position)
			}
			return true
		},
		/* 101 Action14 <- <{ p.AddStar() }> */
		func() bool {
			{
				add(ruleAction14, position)
			}
			return true
		},
		/* 102 Action15 <- <{ p.AddPlus() }> */
		func() bool {
			{
				add(ruleAction15, position)
			}
			return true
		},
		/* 103 Action16 <- <{ p.AddName(buffer[begin:end]) }> */
		func() bool {
			{
				add(ruleAction16, position)
			}
			return true
		},
		/* 104 Action17 <- <{ p.AddDot() }> */
		func() bool {
			{
				add(ruleAction17, position)
			}
			return true
		},
		/* 105 Action18 <- <{ p.AddAction(buffer[begin:end]) }> */
		func() bool {
			{
				add(ruleAction18, position)
			}
			return true
		},
		/* 106 Action19 <- <{ p.AddPush() }> */
		func() bool {
			{
				add(ruleAction19, position)
			}
			return true
		},
		/* 107 Action20 <- <{ p.AddSequence() }> */
		func() bool {
			{
				add(ruleAction20, position)
			}
			return true
		},
		/* 108 Action21 <- <{ p.AddSequence() }> */
		func() bool {
			{
				add(ruleAction21, position)
			}
			return true
		},
		/* 109 Action22 <- <{ p.AddPeekNot(); p.AddDot(); p.AddSequence() }> */
		func() bool {
			{
				add(ruleAction22, position)
			}
			return true
		},
		/* 110 Action23 <- <{ p.AddPeekNot(); p.AddDot(); p.AddSequence() }> */
		func() bool {
			{
				add(ruleAction23, position)
			}
			return true
		},
		/* 111 Action24 <- <{ p.AddAlternate() }> */
		func() bool {
			{
				add(ruleAction24, position)
			}
			return true
		},
		/* 112 Action25 <- <{ p.AddAlternate() }> */
		func() bool {
			{
				add(ruleAction25, position)
			}
			return true
		},
		/* 113 Action26 <- <{ p.AddRange() }> */
		func() bool {
			{
				add(ruleAction26, position)
			}
			return true
		},
		/* 114 Action27 <- <{ p.AddDoubleRange() }> */
		func() bool {
			{
				add(ruleAction27, position)
			}
			return true
		},
		/* 115 Action28 <- <{ p.AddCharacter(buffer[begin:end]) }> */
		func() bool {
			{
				add(ruleAction28, position)
			}
			return true
		},
		/* 116 Action29 <- <{ p.AddDoubleCharacter(buffer[begin:end]) }> */
		func() bool {
			{
				add(ruleAction29, position)
			}
			return true
		},
		/* 117 Action30 <- <{ p.AddCharacter(buffer[begin:end]) }> */
		func() bool {
			{
				add(ruleAction30, position)
			}
			return true
		},
		/* 118 Action31 <- <{ p.AddCharacter("\a") }> */
		func() bool {
			{
				add(ruleAction31, position)
			}
			return true
		},
		/* 119 Action32 <- <{ p.AddCharacter("\b") }> */
		func() bool {
			{
				add(ruleAction32, position)
			}
			return true
		},
		/* 120 Action33 <- <{ p.AddCharacter("\x1B") }> */
		func() bool {
			{
				add(ruleAction33, position)
			}
			return true
		},
		/* 121 Action34 <- <{ p.AddCharacter("\f") }> */
		func() bool {
			{
				add(ruleAction34, position)
			}
			return true
		},
		/* 122 Action35 <- <{ p.AddCharacter("\n") }> */
		func() bool {
			{
				add(ruleAction35, position)
			}
			return true
		},
		/* 123 Action36 <- <{ p.AddCharacter("\r") }> */
		func() bool {
			{
				add(ruleAction36, position)
			}
			return true
		},
		/* 124 Action37 <- <{ p.AddCharacter("\t") }> */
		func() bool {
			{
				add(ruleAction37, position)
			}
			return true
		},
		/* 125 Action38 <- <{ p.AddCharacter("\v") }> */
		func() bool {
			{
				add(ruleAction38, position)
			}
			return true
		},
		/* 126 Action39 <- <{ p.AddCharacter("'") }> */
		func() bool {
			{
				add(ruleAction39, position)
			}
			return true
		},
		/* 127 Action40 <- <{ p.AddCharacter("\"") }> */
		func() bool {
			{
				add(ruleAction40, position)
			}
			return true
		},
		/* 128 Action41 <- <{ p.AddCharacter("[") }> */
		func() bool {
			{
				add(ruleAction41, position)
			}
			return true
		},
		/* 129 Action42 <- <{ p.AddCharacter("]") }> */
		func() bool {
			{
				add(ruleAction42, position)
			}
			return true
		},
		/* 130 Action43 <- <{ p.AddCharacter("-") }> */
		func() bool {
			{
				add(ruleAction43, position)
			}
			return true
		},
		/* 131 Action44 <- <{ p.AddHexaCharacter(buffer[begin:end]) }> */
		func() bool {
			{
				add(ruleAction44, position)
			}
			return true
		},
		/* 132 Action45 <- <{ p.AddOctalCharacter(buffer[begin:end]) }> */
		func() bool {
			{
				add(ruleAction45, position)
			}
			return true
		},
		/* 133 Action46 <- <{ p.AddOctalCharacter(buffer[begin:end]) }> */
		func() bool {
			{
				add(ruleAction46, position)
			}
			return true
		},
		/* 134 Action47 <- <{ p.AddCharacter("\\") }> */
		func() bool {
			{
				add(ruleAction47, position)
			}
			return true
		},
		/* 135 Action48 <- <{ p.runtimeGetruntime() }> */
		nil,
		/* 136 Action49 <- <{ p.runtimeRuntime() }> */
		nil,
		/* 137 Action50 <- <{ p.runtimePorts() }> */
		nil,
		/* 138 Action51 <- <{ p.runtimePacket() }> */
		nil,
		/* 139 Action52 <- <{ p.graphClear() }> */
		nil,
		/* 140 Action53 <- <{ p.graphAddnode() }> */
		nil,
		/* 141 Action54 <- <{ p.graphRemovenode() }> */
		nil,
		/* 142 Action55 <- <{ p.graphRenamenode() }> */
		nil,
		/* 143 Action56 <- <{ p.graphChangenode() }> */
		nil,
		/* 144 Action57 <- <{ p.graphAddedge() }> */
		nil,
		/* 145 Action58 <- <{ p.graphRemoveedge() }> */
		nil,
		/* 146 Action59 <- <{ p.graphChangeedge() }> */
		nil,
		/* 147 Action60 <- <{ p.graphAddinitial() }> */
		nil,
		/* 148 Action61 <- <{ p.graphRemoveinitial() }> */
		nil,
		/* 149 Action62 <- <{ p.graphAddinport() }> */
		nil,
		/* 150 Action63 <- <{ p.graphRemoveinport() }> */
		nil,
		/* 151 Action64 <- <{ p.graphRenameinport() }> */
		nil,
		/* 152 Action65 <- <{ p.graphAddoutport() }> */
		nil,
		/* 153 Action66 <- <{ p.graphRemoveoutport() }> */
		nil,
		/* 154 Action67 <- <{ p.graphRenameoutport() }> */
		nil,
		/* 155 Action68 <- <{ p.graphAddgroup() }> */
		nil,
		/* 156 Action69 <- <{ p.graphRemovegroup() }> */
		nil,
		/* 157 Action70 <- <{ p.graphRenamegroup() }> */
		nil,
		/* 158 Action71 <- <{ p.graphChangegroup() }> */
		nil,
		/* 159 Action72 <- <{ p.componentList() }> */
		nil,
		/* 160 Action73 <- <{ p.componentComponent() }> */
		nil,
		/* 161 Action74 <- <{ p.componentGetsource() }> */
		nil,
		/* 162 Action75 <- <{ p.componentSource() }> */
		nil,
		/* 163 Action76 <- <{ p.networkStart() }> */
		nil,
		/* 164 Action77 <- <{ p.networkGetstatus() }> */
		nil,
		/* 165 Action78 <- <{ p.networkStop() }> */
		nil,
		/* 166 Action79 <- <{ p.networkStarted() }> */
		nil,
		/* 167 Action80 <- <{ p.networkStatus() }> */
		nil,
		/* 168 Action81 <- <{ p.networkStopped() }> */
		nil,
		/* 169 Action82 <- <{ p.networkDebug() }> */
		nil,
		/* 170 Action83 <- <{ p.networkIcon() }> */
		nil,
		/* 171 Action84 <- <{ p.networkOutput() }> */
		nil,
		/* 172 Action85 <- <{ p.networkError() }> */
		nil,
		/* 173 Action86 <- <{ p.networkProcesserror() }> */
		nil,
		/* 174 Action87 <- <{ p.networkConnect() }> */
		nil,
		/* 175 Action88 <- <{ p.networkBegingroup() }> */
		nil,
		/* 176 Action89 <- <{ p.networkData() }> */
		nil,
		/* 177 Action90 <- <{ p.networkEndgroup() }> */
		nil,
		/* 178 Action91 <- <{ p.networkDisconnect() }> */
		nil,
		/* 179 Action92 <- <{ p.networkEdges() }> */
		nil,
	}
	p.rules = _rules
}
