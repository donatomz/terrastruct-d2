package d2chaos

import (
	"fmt"
	mathrand "math/rand"
	"strings"
	"time"

	"oss.terrastruct.com/d2/d2ast"
	"oss.terrastruct.com/d2/d2format"
	"oss.terrastruct.com/d2/d2graph"
	"oss.terrastruct.com/d2/d2oracle"
	"oss.terrastruct.com/d2/d2target"
	"oss.terrastruct.com/d2/lib/go2"
)

func GenDSL(maxi int) (_ string, err error) {
	gs := &dslGenState{
		rand:       mathrand.New(mathrand.NewSource(time.Now().UnixNano())),
		g:          d2graph.NewGraph(&d2ast.Map{}),
		nodeShapes: make(map[string]string),
	}
	err = gs.gen(maxi)
	if err != nil {
		return "", err
	}
	return d2format.Format(gs.g.AST), nil
}

type dslGenState struct {
	rand *mathrand.Rand
	g    *d2graph.Graph

	nodesArr   []string
	nodeShapes map[string]string
}

func (gs *dslGenState) gen(maxi int) error {
	maxi = gs.rand.Intn(maxi) + 1

	for i := 0; i < maxi; i++ {
		switch gs.roll(25, 75) {
		case 0:
			// 25% chance of creating a new node.
			err := gs.node()
			if err != nil {
				return err
			}
		case 1:
			// 75% chance of connecting two random nodes with a random label.
			err := gs.edge()
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (gs *dslGenState) genNode(containerID string) (string, error) {
	nodeID := gs.randStr(32, true)
	if containerID != "" {
		nodeID = containerID + "." + nodeID
	}
	var err error
	gs.g, nodeID, err = d2oracle.Create(gs.g, nodeID)
	if err != nil {
		return "", err
	}
	gs.nodesArr = append(gs.nodesArr, nodeID)
	gs.nodeShapes[nodeID] = "square"
	return nodeID, nil
}

func (gs *dslGenState) node() error {
	containerID := ""
	var err error
	if gs.roll(25, 75) == 1 {
		// 75% chance of creating this as a child under a container.
		containerID, err = gs.randContainer()
		if err != nil {
			return err
		}
	}

	nodeID, err := gs.genNode(containerID)
	if err != nil {
		return err
	}

	if gs.roll(25, 75) == 0 {
		// 25% chance of adding a label.
		gs.g, err = d2oracle.Set(gs.g, nodeID, nil, go2.Pointer(gs.randStr(256, false)))
		if err != nil {
			return err
		}
	}

	if gs.roll(25, 75) == 1 {
		// 75% chance of adding a shape.
		randShape := gs.randShape()
		gs.g, err = d2oracle.Set(gs.g, nodeID+".shape", nil, go2.Pointer(randShape))
		if err != nil {
			return err
		}
		gs.nodeShapes[nodeID] = randShape
	}

	return nil
}

func (gs *dslGenState) edge() error {
	var src string
	var dst string
	var err error
	for {
		src, err = gs.randNode()
		if err != nil {
			return err
		}
		dst, err = gs.randNode()
		if err != nil {
			return err
		}
		if src != dst {
			break
		}
		err = gs.node()
		if err != nil {
			return err
		}
	}

	srcArrow := "-"
	if gs.randBool() {
		srcArrow = "<"
	}
	dstArrow := "-"
	if gs.randBool() {
		dstArrow = ">"
		if srcArrow == "<" {
			dstArrow = "->"
		}
	}

	key := fmt.Sprintf("%s %s%s %s", src, srcArrow, dstArrow, dst)
	gs.g, key, err = d2oracle.Create(gs.g, key)
	if err != nil {
		return err
	}
	if gs.randBool() {
		gs.g, err = d2oracle.Set(gs.g, key, nil, go2.Pointer(gs.randStr(128, false)))
		if err != nil {
			return err
		}
	}
	return nil
}

func (gs *dslGenState) randContainer() (string, error) {
	containers := go2.Filter(gs.nodesArr, func(x string) bool {
		shape := gs.nodeShapes[x]
		return shape != "image" &&
			shape != "code" &&
			shape != "sql_table" &&
			shape != "text" &&
			shape != "class"
	})
	if len(containers) == 0 {
		return "", nil
	}
	return containers[gs.rand.Intn(len(containers))], nil
}

func (gs *dslGenState) randNode() (string, error) {
	if len(gs.nodesArr) == 0 {
		return gs.genNode("")
	}
	return gs.nodesArr[gs.rand.Intn(len(gs.nodesArr))], nil
}

func (gs *dslGenState) randBool() bool {
	return gs.rand.Intn(2) == 0
}

// TODO go back to using xrand.String, currently some incompatibility with
// stuffing these strings into a script for dagre
func randRune() rune {
	if mathrand.Int31n(100) == 0 {
		// Generate newline 1% of the time.
		return '\n'
	}
	return mathrand.Int31n(128) + 1
}

func String(n int, exclude []rune) string {
	var b strings.Builder
	for i := 0; i < n; i++ {
		r := randRune()
		excluded := false
		for _, xr := range exclude {
			if r == xr {
				excluded = true
				break
			}
		}
		if excluded {
			i--
			continue
		}
		b.WriteRune(r)
	}
	return b.String()
}

func (gs *dslGenState) randStr(n int, inKey bool) string {
	// Underscores have semantic meaning (parent)
	// Backticks are for opening and closing these strings
	// Curly braces can trigger templating
	// \\ triggers octal sequences
	s := String(gs.rand.Intn(n), []rune{
		rune('_'),
		rune('`'),
		rune('}'),
		rune('{'),
		rune('\\'),
	})
	as := d2ast.RawString(s, inKey)
	return d2format.Format(as)
}

func (gs *dslGenState) randShape() string {
	for {
		s := shapes[gs.rand.Intn(len(shapes))]
		if s != "image" {
			return s
		}
	}
}

func (gs *dslGenState) roll(probs ...int) int {
	max := 0
	for _, p := range probs {
		max += p
	}

	n := gs.rand.Intn(max)
	var acc int
	for i, p := range probs {
		if n >= acc && n < acc+p {
			return i
		}
		acc += p
	}

	panic("d2chaos: unreachable")
}

var shapes = d2target.Shapes
