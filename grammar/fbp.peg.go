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
	ruleDot
	ruleAction0
	ruleAction1
	ruleAction2
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

	rulePre_
	rule_In_
	rule_Suf
)

var rul3s = [...]string{
	"Unknown",
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
	"Dot",
	"Action0",
	"Action1",
	"Action2",
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

type JLPeg struct {
	Buffer string
	buffer []rune
	rules  [92]func() bool
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
	p *JLPeg
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

func (p *JLPeg) PrintSyntaxTree() {
	p.tokenTree.PrintSyntaxTree(p.Buffer)
}

func (p *JLPeg) Highlighter() {
	p.tokenTree.PrintSyntax()
}

func (p *JLPeg) Execute() {
	buffer, begin, end := p.Buffer, 0, 0
	for token := range p.tokenTree.Tokens() {
		switch token.pegRule {
		case rulePegText:
			begin, end = int(token.begin), int(token.end)
		case ruleAction0:
			runtimeGetruntime()
		case ruleAction1:
			p.runtimeRuntime()
		case ruleAction2:
			p.runtimePorts()
		case ruleAction3:
			p.runtimePacket()
		case ruleAction4:
			p.graphClear()
		case ruleAction5:
			p.graphAddnode()
		case ruleAction6:
			p.graphRemovenode()
		case ruleAction7:
			p.graphRenamenode()
		case ruleAction8:
			p.graphChangenode()
		case ruleAction9:
			p.graphAddedge()
		case ruleAction10:
			p.graphRemoveedge()
		case ruleAction11:
			p.graphChangeedge()
		case ruleAction12:
			p.graphAddinitial()
		case ruleAction13:
			p.graphRemoveinitial()
		case ruleAction14:
			p.graphAddinport()
		case ruleAction15:
			p.graphRemoveinport()
		case ruleAction16:
			p.graphRenameinport()
		case ruleAction17:
			p.graphAddoutport()
		case ruleAction18:
			p.graphRemoveoutport()
		case ruleAction19:
			p.graphRenameoutport()
		case ruleAction20:
			p.graphAddgroup()
		case ruleAction21:
			p.graphRemovegroup()
		case ruleAction22:
			p.graphRenamegroup()
		case ruleAction23:
			p.graphChangegroup()
		case ruleAction24:
			p.componentList()
		case ruleAction25:
			p.componentComponent()
		case ruleAction26:
			p.componentGetsource()
		case ruleAction27:
			p.componentSource()
		case ruleAction28:
			p.networkStart()
		case ruleAction29:
			p.networkGetstatus()
		case ruleAction30:
			p.networkStop()
		case ruleAction31:
			p.networkStarted()
		case ruleAction32:
			p.networkStatus()
		case ruleAction33:
			p.networkStopped()
		case ruleAction34:
			p.networkDebug()
		case ruleAction35:
			p.networkIcon()
		case ruleAction36:
			p.networkOutput()
		case ruleAction37:
			p.networkError()
		case ruleAction38:
			p.networkProcesserror()
		case ruleAction39:
			p.networkConnect()
		case ruleAction40:
			p.networkBegingroup()
		case ruleAction41:
			p.networkData()
		case ruleAction42:
			p.networkEndgroup()
		case ruleAction43:
			p.networkDisconnect()
		case ruleAction44:
			p.networkEdges()

		}
	}
}

func (p *JLPeg) Init() {
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

	/*matchChar := func(c byte) bool {
		if buffer[position] == c {
			position++
			return true
		}
		return false
	}*/

	_rules = [...]func() bool{
		nil,
		/* 0 runtimeGetruntime <- <('r' 'u' 'n' 't' 'i' 'm' 'e' Dot ('g' 'e' 't' 'r' 'u' 'n' 't' 'i' 'm' 'e') Action0)> */
		func() bool {
			position0, tokenIndex0, depth0 := position, tokenIndex, depth
			{
				position1 := position
				depth++
				if buffer[position] != rune('r') {
					goto l0
				}
				position++
				if buffer[position] != rune('u') {
					goto l0
				}
				position++
				if buffer[position] != rune('n') {
					goto l0
				}
				position++
				if buffer[position] != rune('t') {
					goto l0
				}
				position++
				if buffer[position] != rune('i') {
					goto l0
				}
				position++
				if buffer[position] != rune('m') {
					goto l0
				}
				position++
				if buffer[position] != rune('e') {
					goto l0
				}
				position++
				if !_rules[ruleDot]() {
					goto l0
				}
				if buffer[position] != rune('g') {
					goto l0
				}
				position++
				if buffer[position] != rune('e') {
					goto l0
				}
				position++
				if buffer[position] != rune('t') {
					goto l0
				}
				position++
				if buffer[position] != rune('r') {
					goto l0
				}
				position++
				if buffer[position] != rune('u') {
					goto l0
				}
				position++
				if buffer[position] != rune('n') {
					goto l0
				}
				position++
				if buffer[position] != rune('t') {
					goto l0
				}
				position++
				if buffer[position] != rune('i') {
					goto l0
				}
				position++
				if buffer[position] != rune('m') {
					goto l0
				}
				position++
				if buffer[position] != rune('e') {
					goto l0
				}
				position++
				if !_rules[ruleAction0]() {
					goto l0
				}
				depth--
				add(ruleruntimeGetruntime, position1)
			}
			return true
		l0:
			position, tokenIndex, depth = position0, tokenIndex0, depth0
			return false
		},
		/* 1 runtimeRuntime <- <('r' 'u' 'n' 't' 'i' 'm' 'e' Dot ('r' 'u' 'n' 't' 'i' 'm' 'e') Action1)> */
		nil,
		/* 2 runtimePorts <- <('r' 'u' 'n' 't' 'i' 'm' 'e' Dot ('p' 'o' 'r' 't' 's') Action2)> */
		nil,
		/* 3 runtimePacket <- <('r' 'u' 'n' 't' 'i' 'm' 'e' Dot ('p' 'a' 'c' 'k' 'e' 't') Action3)> */
		nil,
		/* 4 graphClear <- <('g' 'r' 'a' 'p' 'h' Dot ('c' 'l' 'e' 'a' 'r') Action4)> */
		nil,
		/* 5 graphAddnode <- <('g' 'r' 'a' 'p' 'h' Dot ('a' 'd' 'd' 'n' 'o' 'd' 'e') Action5)> */
		nil,
		/* 6 graphRemovenode <- <('g' 'r' 'a' 'p' 'h' Dot ('r' 'e' 'm' 'o' 'v' 'e' 'n' 'o' 'd' 'e') Action6)> */
		nil,
		/* 7 graphRenamenode <- <('g' 'r' 'a' 'p' 'h' Dot ('r' 'e' 'n' 'a' 'm' 'e' 'n' 'o' 'd' 'e') Action7)> */
		nil,
		/* 8 graphChangenode <- <('g' 'r' 'a' 'p' 'h' Dot ('c' 'h' 'a' 'n' 'g' 'e' 'n' 'o' 'd' 'e') Action8)> */
		nil,
		/* 9 graphAddedge <- <('g' 'r' 'a' 'p' 'h' Dot ('a' 'd' 'd' 'e' 'd' 'g' 'e') Action9)> */
		nil,
		/* 10 graphRemoveedge <- <('g' 'r' 'a' 'p' 'h' Dot ('r' 'e' 'm' 'o' 'v' 'e' 'e' 'd' 'g' 'e') Action10)> */
		nil,
		/* 11 graphChangeedge <- <('g' 'r' 'a' 'p' 'h' Dot ('c' 'h' 'a' 'n' 'g' 'e' 'e' 'd' 'g' 'e') Action11)> */
		nil,
		/* 12 graphAddinitial <- <('g' 'r' 'a' 'p' 'h' Dot ('a' 'd' 'd' 'i' 'n' 'i' 't' 'i' 'a' 'l') Action12)> */
		nil,
		/* 13 graphRemoveinitial <- <('g' 'r' 'a' 'p' 'h' Dot ('r' 'e' 'm' 'o' 'v' 'e' 'i' 'n' 'i' 't' 'i' 'a' 'l') Action13)> */
		nil,
		/* 14 graphAddinport <- <('g' 'r' 'a' 'p' 'h' Dot ('a' 'd' 'd' 'i' 'n' 'p' 'o' 'r' 't') Action14)> */
		nil,
		/* 15 graphRemoveinport <- <('g' 'r' 'a' 'p' 'h' Dot ('r' 'e' 'm' 'o' 'v' 'e' 'i' 'n' 'p' 'o' 'r' 't') Action15)> */
		nil,
		/* 16 graphRenameinport <- <('g' 'r' 'a' 'p' 'h' Dot ('r' 'e' 'n' 'a' 'm' 'e' 'i' 'n' 'p' 'o' 'r' 't') Action16)> */
		nil,
		/* 17 graphAddoutport <- <('g' 'r' 'a' 'p' 'h' Dot ('a' 'd' 'd' 'o' 'u' 't' 'p' 'o' 'r' 't') Action17)> */
		nil,
		/* 18 graphRemoveoutport <- <('g' 'r' 'a' 'p' 'h' Dot ('r' 'e' 'm' 'o' 'v' 'e' 'o' 'u' 't' 'p' 'o' 'r' 't') Action18)> */
		nil,
		/* 19 graphRenameoutport <- <('g' 'r' 'a' 'p' 'h' Dot ('r' 'e' 'n' 'a' 'm' 'e' 'o' 'u' 't' 'p' 'o' 'r' 't') Action19)> */
		nil,
		/* 20 graphAddgroup <- <('g' 'r' 'a' 'p' 'h' Dot ('a' 'd' 'd' 'g' 'r' 'o' 'u' 'p') Action20)> */
		nil,
		/* 21 graphRemovegroup <- <('g' 'r' 'a' 'p' 'h' Dot ('r' 'e' 'm' 'o' 'v' 'e' 'g' 'r' 'o' 'u' 'p') Action21)> */
		nil,
		/* 22 graphRenamegroup <- <('g' 'r' 'a' 'p' 'h' Dot ('r' 'e' 'n' 'a' 'm' 'e' 'g' 'r' 'o' 'u' 'p') Action22)> */
		nil,
		/* 23 graphChangegroup <- <('g' 'r' 'a' 'p' 'h' Dot ('c' 'h' 'a' 'n' 'g' 'e' 'g' 'r' 'o' 'u' 'p') Action23)> */
		nil,
		/* 24 componentList <- <('c' 'o' 'm' 'p' 'o' 'n' 'e' 'n' 't' Dot ('l' 'i' 's' 't') Action24)> */
		nil,
		/* 25 componentComponent <- <('c' 'o' 'm' 'p' 'o' 'n' 'e' 'n' 't' Dot ('c' 'o' 'm' 'p' 'o' 'n' 'e' 'n' 't') Action25)> */
		nil,
		/* 26 componentGetsource <- <('c' 'o' 'm' 'p' 'o' 'n' 'e' 'n' 't' Dot ('g' 'e' 't' 's' 'o' 'u' 'r' 'c' 'e') Action26)> */
		nil,
		/* 27 componentSource <- <('c' 'o' 'm' 'p' 'o' 'n' 'e' 'n' 't' Dot ('s' 'o' 'u' 'r' 'c' 'e') Action27)> */
		nil,
		/* 28 networkStart <- <('n' 'e' 't' 'w' 'o' 'r' 'k' Dot ('s' 't' 'a' 'r' 't') Action28)> */
		nil,
		/* 29 networkGetstatus <- <('n' 'e' 't' 'w' 'o' 'r' 'k' Dot ('g' 'e' 't' 's' 't' 'a' 't' 'u' 's') Action29)> */
		nil,
		/* 30 networkStop <- <('n' 'e' 't' 'w' 'o' 'r' 'k' Dot ('s' 't' 'o' 'p') Action30)> */
		nil,
		/* 31 networkStarted <- <('n' 'e' 't' 'w' 'o' 'r' 'k' Dot ('s' 't' 'a' 'r' 't' 'e' 'd') Action31)> */
		nil,
		/* 32 networkStatus <- <('n' 'e' 't' 'w' 'o' 'r' 'k' Dot ('s' 't' 'a' 't' 'u' 's') Action32)> */
		nil,
		/* 33 networkStopped <- <('n' 'e' 't' 'w' 'o' 'r' 'k' Dot ('s' 't' 'o' 'p' 'p' 'e' 'd') Action33)> */
		nil,
		/* 34 networkDebug <- <('n' 'e' 't' 'w' 'o' 'r' 'k' Dot ('d' 'e' 'b' 'u' 'g') Action34)> */
		nil,
		/* 35 networkIcon <- <('n' 'e' 't' 'w' 'o' 'r' 'k' Dot ('i' 'c' 'o' 'n') Action35)> */
		nil,
		/* 36 networkOutput <- <('n' 'e' 't' 'w' 'o' 'r' 'k' Dot ('o' 'u' 't' 'p' 'u' 't') Action36)> */
		nil,
		/* 37 networkError <- <('n' 'e' 't' 'w' 'o' 'r' 'k' Dot ('e' 'r' 'r' 'o' 'r') Action37)> */
		nil,
		/* 38 networkProcesserror <- <('n' 'e' 't' 'w' 'o' 'r' 'k' Dot ('p' 'r' 'o' 'c' 'e' 's' 's' 'e' 'r' 'r' 'o' 'r') Action38)> */
		nil,
		/* 39 networkConnect <- <('n' 'e' 't' 'w' 'o' 'r' 'k' Dot ('c' 'o' 'n' 'n' 'e' 'c' 't') Action39)> */
		nil,
		/* 40 networkBegingroup <- <('n' 'e' 't' 'w' 'o' 'r' 'k' Dot ('b' 'e' 'g' 'i' 'n' 'g' 'r' 'o' 'u' 'p') Action40)> */
		nil,
		/* 41 networkData <- <('n' 'e' 't' 'w' 'o' 'r' 'k' Dot ('d' 'a' 't' 'a') Action41)> */
		nil,
		/* 42 networkEndgroup <- <('n' 'e' 't' 'w' 'o' 'r' 'k' Dot ('e' 'n' 'd' 'g' 'r' 'o' 'u' 'p') Action42)> */
		nil,
		/* 43 networkDisconnect <- <('n' 'e' 't' 'w' 'o' 'r' 'k' Dot ('d' 'i' 's' 'c' 'o' 'n' 'n' 'e' 'c' 't') Action43)> */
		nil,
		/* 44 networkEdges <- <('n' 'e' 't' 'w' 'o' 'r' 'k' Dot ('e' 'd' 'g' 'e' 's') Action44)> */
		nil,
		/* 46 Dot <- <> */
		func() bool {
			{
				position47 := position
				depth++
				depth--
				add(ruleDot, position47)
			}
			return true
		},
		/* 47 Action0 <- <{ runtimeGetruntime() }> */
		func() bool {
			{
				add(ruleAction0, position)
			}
			return true
		},
		/* 48 Action1 <- <{ p.runtimeRuntime() }> */
		nil,
		/* 49 Action2 <- <{ p.runtimePorts() }> */
		nil,
		/* 50 Action3 <- <{ p.runtimePacket() }> */
		nil,
		/* 51 Action4 <- <{ p.graphClear() }> */
		nil,
		/* 52 Action5 <- <{ p.graphAddnode() }> */
		nil,
		/* 53 Action6 <- <{ p.graphRemovenode() }> */
		nil,
		/* 54 Action7 <- <{ p.graphRenamenode() }> */
		nil,
		/* 55 Action8 <- <{ p.graphChangenode() }> */
		nil,
		/* 56 Action9 <- <{ p.graphAddedge() }> */
		nil,
		/* 57 Action10 <- <{ p.graphRemoveedge() }> */
		nil,
		/* 58 Action11 <- <{ p.graphChangeedge() }> */
		nil,
		/* 59 Action12 <- <{ p.graphAddinitial() }> */
		nil,
		/* 60 Action13 <- <{ p.graphRemoveinitial() }> */
		nil,
		/* 61 Action14 <- <{ p.graphAddinport() }> */
		nil,
		/* 62 Action15 <- <{ p.graphRemoveinport() }> */
		nil,
		/* 63 Action16 <- <{ p.graphRenameinport() }> */
		nil,
		/* 64 Action17 <- <{ p.graphAddoutport() }> */
		nil,
		/* 65 Action18 <- <{ p.graphRemoveoutport() }> */
		nil,
		/* 66 Action19 <- <{ p.graphRenameoutport() }> */
		nil,
		/* 67 Action20 <- <{ p.graphAddgroup() }> */
		nil,
		/* 68 Action21 <- <{ p.graphRemovegroup() }> */
		nil,
		/* 69 Action22 <- <{ p.graphRenamegroup() }> */
		nil,
		/* 70 Action23 <- <{ p.graphChangegroup() }> */
		nil,
		/* 71 Action24 <- <{ p.componentList() }> */
		nil,
		/* 72 Action25 <- <{ p.componentComponent() }> */
		nil,
		/* 73 Action26 <- <{ p.componentGetsource() }> */
		nil,
		/* 74 Action27 <- <{ p.componentSource() }> */
		nil,
		/* 75 Action28 <- <{ p.networkStart() }> */
		nil,
		/* 76 Action29 <- <{ p.networkGetstatus() }> */
		nil,
		/* 77 Action30 <- <{ p.networkStop() }> */
		nil,
		/* 78 Action31 <- <{ p.networkStarted() }> */
		nil,
		/* 79 Action32 <- <{ p.networkStatus() }> */
		nil,
		/* 80 Action33 <- <{ p.networkStopped() }> */
		nil,
		/* 81 Action34 <- <{ p.networkDebug() }> */
		nil,
		/* 82 Action35 <- <{ p.networkIcon() }> */
		nil,
		/* 83 Action36 <- <{ p.networkOutput() }> */
		nil,
		/* 84 Action37 <- <{ p.networkError() }> */
		nil,
		/* 85 Action38 <- <{ p.networkProcesserror() }> */
		nil,
		/* 86 Action39 <- <{ p.networkConnect() }> */
		nil,
		/* 87 Action40 <- <{ p.networkBegingroup() }> */
		nil,
		/* 88 Action41 <- <{ p.networkData() }> */
		nil,
		/* 89 Action42 <- <{ p.networkEndgroup() }> */
		nil,
		/* 90 Action43 <- <{ p.networkDisconnect() }> */
		nil,
		/* 91 Action44 <- <{ p.networkEdges() }> */
		nil,
	}
	p.rules = _rules
}
