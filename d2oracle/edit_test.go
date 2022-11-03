package d2oracle_test

import (
	"fmt"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"oss.terrastruct.com/xjson"

	"oss.terrastruct.com/diff"

	"oss.terrastruct.com/d2/d2compiler"
	"oss.terrastruct.com/d2/d2format"
	"oss.terrastruct.com/d2/d2graph"
	"oss.terrastruct.com/d2/d2oracle"
	"oss.terrastruct.com/d2/d2target"
	"oss.terrastruct.com/d2/lib/go2"
)

// TODO: make assertions less specific
// TODO: move n objects and n edges assertions as fields on test instead of as callback

func TestCreate(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name string
		text string
		key  string

		expKey     string
		expErr     string
		exp        string
		assertions func(t *testing.T, g *d2graph.Graph)
	}{
		{
			name: "base",
			text: ``,
			key:  `square`,

			expKey: `square`,
			exp: `square
`,
			assertions: func(t *testing.T, g *d2graph.Graph) {
				if len(g.Objects) != 1 {
					t.Fatalf("expected 1 objects: %#v", g.Objects)
				}
				if g.Objects[0].ID != "square" {
					t.Fatalf("expected g.Objects[0].ID to be square: %#v", g.Objects[0])
				}
				if g.Objects[0].Attributes.Label.MapKey.Value.Unbox() != nil {
					t.Fatalf("expected g.Objects[0].Attributes.Label.Node.Value.Unbox() == nil: %#v", g.Objects[0].Attributes.Label.MapKey.Value)
				}
				if d2format.Format(g.Objects[0].Attributes.Label.MapKey.Key) != "square" {
					t.Fatalf("expected g.Objects[0].Attributes.Label.Node.Key to be square: %#v", g.Objects[0].Attributes.Label.MapKey.Key)
				}
			},
		},
		{
			name: "gen_key_suffix",
			text: `"x "
`,
			key: `"x "`,

			expKey: `x  2`,
			exp: `"x "
x  2
`,
			assertions: func(t *testing.T, g *d2graph.Graph) {
				if len(g.Objects) != 2 {
					t.Fatalf("unexpected objects length: %#v", g.Objects)
				}
				if g.Objects[1].ID != `x  2` {
					t.Fatalf("bad object ID: %#v", g.Objects[1])
				}
			},
		},
		{
			name: "nested",
			text: ``,
			key:  `b.c.square`,

			expKey: `b.c.square`,
			exp: `b.c.square
`,
			assertions: func(t *testing.T, g *d2graph.Graph) {
				if len(g.Objects) != 3 {
					t.Fatalf("unexpected objects length: %#v", g.Objects)
				}
				if g.Objects[2].AbsID() != "b.c.square" {
					t.Fatalf("bad absolute ID: %#v", g.Objects[2].AbsID())
				}
				if d2format.Format(g.Objects[2].Attributes.Label.MapKey.Key) != "b.c.square" {
					t.Fatalf("bad mapkey: %#v", g.Objects[2].Attributes.Label.MapKey.Key)
				}
				if g.Objects[2].Attributes.Label.MapKey.Value.Unbox() != nil {
					t.Fatalf("expected nil mapkey value: %#v", g.Objects[2].Attributes.Label.MapKey.Value)
				}
			},
		},
		{
			name: "gen_key",
			text: `square`,
			key:  `square`,

			expKey: `square 2`,
			exp: `square
square 2
`,
			assertions: func(t *testing.T, g *d2graph.Graph) {
				if len(g.Objects) != 2 {
					t.Fatalf("expected 2 objects: %#v", g.Objects)
				}
				if g.Objects[1].ID != "square 2" {
					t.Fatalf("expected g.Objects[1].ID to be square 2: %#v", g.Objects[1])
				}
				if g.Objects[1].Attributes.Label.MapKey.Value.Unbox() != nil {
					t.Fatalf("expected g.Objects[1].Attributes.Label.Node.Value.Unbox() == nil: %#v", g.Objects[1].Attributes.Label.MapKey.Value)
				}
				if d2format.Format(g.Objects[1].Attributes.Label.MapKey.Key) != "square 2" {
					t.Fatalf("expected g.Objects[1].Attributes.Label.Node.Key to be square 2: %#v", g.Objects[1].Attributes.Label.MapKey.Key)
				}
			},
		},
		{
			name: "gen_key_nested",
			text: `x.y.z.square`,
			key:  `x.y.z.square`,

			expKey: `x.y.z.square 2`,
			exp: `x.y.z.square
x.y.z.square 2
`,
			assertions: func(t *testing.T, g *d2graph.Graph) {
				if len(g.Objects) != 5 {
					t.Fatalf("unexpected objects length: %#v", g.Objects)
				}
				if g.Objects[4].ID != "square 2" {
					t.Fatalf("unexpected object id: %#v", g.Objects[4])
				}
			},
		},
		{
			name: "scope",
			text: `x.y.z: {
}`,
			key: `x.y.z.square`,

			expKey: `x.y.z.square`,
			exp: `x.y.z: {
  square
}
`,
			assertions: func(t *testing.T, g *d2graph.Graph) {
				if len(g.Objects) != 4 {
					t.Fatalf("expected 4 objects: %#v", g.Objects)
				}
				if g.Objects[3].ID != "square" {
					t.Fatalf("expected g.Objects[3].ID to be square: %#v", g.Objects[3])
				}
				if g.Objects[3].Attributes.Label.MapKey.Value.Unbox() != nil {
					t.Fatalf("expected g.Objects[3].Attributes.Label.Node.Value.Unbox() == nil: %#v", g.Objects[3].Attributes.Label.MapKey.Value)
				}
				if d2format.Format(g.Objects[3].Attributes.Label.MapKey.Key) != "square" {
					t.Fatalf("expected g.Objects[3].Attributes.Label.Node.Key to be square: %#v", g.Objects[3].Attributes.Label.MapKey.Key)
				}
			},
		},
		{
			name: "gen_key_scope",
			text: `x.y.z: {
  square
}`,
			key: `x.y.z.square`,

			expKey: `x.y.z.square 2`,
			exp: `x.y.z: {
  square
  square 2
}
`,
			assertions: func(t *testing.T, g *d2graph.Graph) {
				if len(g.Objects) != 5 {
					t.Fatalf("expected 5 objects: %#v", g.Objects)
				}
				if g.Objects[4].ID != "square 2" {
					t.Fatalf("expected g.Objects[4].ID to be square 2: %#v", g.Objects[4])
				}
				if g.Objects[4].Attributes.Label.MapKey.Value.Unbox() != nil {
					t.Fatalf("expected g.Objects[4].Attributes.Label.Node.Value.Unbox() == nil: %#v", g.Objects[4].Attributes.Label.MapKey.Value)
				}
				if d2format.Format(g.Objects[4].Attributes.Label.MapKey.Key) != "square 2" {
					t.Fatalf("expected g.Objects[4].Attributes.Label.Node.Key to be square 2: %#v", g.Objects[4].Attributes.Label.MapKey.Key)
				}
			},
		},
		{
			name: "gen_key_n",
			text: `x.y.z: {
  square
  square 2
  square 3
  square 4
  square 5
  square 6
  square 7
  square 8
  square 9
  square 10
}`,
			key: `x.y.z.square`,

			expKey: `x.y.z.square 11`,
			exp: `x.y.z: {
  square
  square 2
  square 3
  square 4
  square 5
  square 6
  square 7
  square 8
  square 9
  square 10
  square 11
}
`,
			assertions: func(t *testing.T, g *d2graph.Graph) {
				if len(g.Objects) != 14 {
					t.Fatalf("expected 14 objects: %#v", g.Objects)
				}
				if g.Objects[13].ID != "square 11" {
					t.Fatalf("expected g.Objects[13].ID to be square 11: %#v", g.Objects[13])
				}
				if d2format.Format(g.Objects[13].Attributes.Label.MapKey.Key) != "square 11" {
					t.Fatalf("expected g.Objects[13].Attributes.Label.Node.Key to be square 11: %#v", g.Objects[13].Attributes.Label.MapKey.Key)
				}
			},
		},
		{
			name: "edge",
			text: ``,
			key:  `x -> y`,

			expKey: `(x -> y)[0]`,
			exp: `x -> y
`,
			assertions: func(t *testing.T, g *d2graph.Graph) {
				if len(g.Objects) != 2 {
					t.Fatalf("expected 2 objects: %#v", g.Objects)
				}
				if len(g.Edges) != 1 {
					t.Fatalf("expected 1 edge: %#v", g.Edges)
				}
				if g.Edges[0].Src.ID != "x" {
					t.Fatalf("expected g.Edges[0].Src.ID == x: %#v", g.Edges[0].Src.ID)
				}
				if g.Edges[0].Dst.ID != "y" {
					t.Fatalf("expected g.Edges[0].Dst.ID == y: %#v", g.Edges[0].Dst.ID)
				}
			},
		},
		{
			name: "edge_nested",
			text: ``,
			key:  `container.(x -> y)`,

			expKey: `container.(x -> y)[0]`,
			exp: `container.(x -> y)
`,
			assertions: func(t *testing.T, g *d2graph.Graph) {
				if len(g.Objects) != 3 {
					t.Fatalf("unexpected objects: %#v", g.Objects)
				}
				if len(g.Edges) != 1 {
					t.Fatalf("unexpected edges: %#v", g.Edges)
				}
			},
		},
		{
			name: "edge_scope",
			text: `container: {
}`,
			key: `container.(x -> y)`,

			expKey: `container.(x -> y)[0]`,
			exp: `container: {
  x -> y
}
`,
			assertions: func(t *testing.T, g *d2graph.Graph) {
				if len(g.Objects) != 3 {
					t.Fatalf("expected 3 objects: %#v", g.Objects)
				}
			},
		},
		{
			name: "edge_scope_flat",
			text: `container: {
}`,
			key: `container.x -> container.y`,

			expKey: `container.(x -> y)[0]`,
			exp: `container: {
  x -> y
}
`,
			assertions: func(t *testing.T, g *d2graph.Graph) {
				if len(g.Objects) != 3 {
					t.Fatalf("expected 3 objects: %#v", g.Objects)
				}
			},
		},
		{
			name: "edge_scope_nested",
			text: `x.y`,
			key:  `x.y.z -> x.y.q`,

			expKey: `x.y.(z -> q)[0]`,
			exp: `x.y: {
  z -> q
}
`,
			assertions: func(t *testing.T, g *d2graph.Graph) {
				if len(g.Objects) != 4 {
					t.Fatalf("unexpected objects: %#v", g.Objects)
				}
				if len(g.Edges) != 1 {
					t.Fatalf("unexpected edges: %#v", g.Edges)
				}
			},
		},
		{
			name: "edge_unique",
			text: `x -> y
hello.(x -> y)
hello.(x -> y)
`,
			key: `hello.(x -> y)`,

			expKey: `hello.(x -> y)[2]`,
			exp: `x -> y
hello.(x -> y)
hello.(x -> y)
hello.(x -> y)
`,
			assertions: func(t *testing.T, g *d2graph.Graph) {
				if len(g.Objects) != 5 {
					t.Fatalf("expected 5 objects: %#v", g.Objects)
				}
				if len(g.Edges) != 4 {
					t.Fatalf("expected 4 edges: %#v", g.Edges)
				}
			},
		},
		{
			name: "container",
			text: `b`,
			key:  `b.q`,

			expKey: `b.q`,
			exp: `b: {
  q
}
`,
			assertions: func(t *testing.T, g *d2graph.Graph) {
				if len(g.Objects) != 2 {
					t.Fatalf("expected 2 objects: %#v", g.Objects)
				}
			},
		},
		{
			name: "container_edge",
			text: `b`,
			key:  `b.x -> b.y`,

			expKey: `b.(x -> y)[0]`,
			exp: `b: {
  x -> y
}
`,
			assertions: func(t *testing.T, g *d2graph.Graph) {
				if len(g.Objects) != 3 {
					t.Fatalf("expected 3 objects: %#v", g.Objects)
				}
			},
		},
		{
			name: "container_edge_label",
			text: `b: zoom`,
			key:  `b.x -> b.y`,

			expKey: `b.(x -> y)[0]`,
			exp: `b: zoom {
  x -> y
}
`,
			assertions: func(t *testing.T, g *d2graph.Graph) {
				if len(g.Objects) != 3 {
					t.Fatalf("expected 3 objects: %#v", g.Objects)
				}
			},
		},
		{
			name: "make_scope_multiline",

			text: `rawr: {shape: circle}
`,
			key: `rawr.orange`,

			expKey: `rawr.orange`,
			exp: `rawr: {
  shape: circle
  orange
}
`,
		},
		{
			name: "make_scope_multiline_spacing_1",

			text: `before
rawr: {shape: circle}
after
`,
			key: `rawr.orange`,

			expKey: `rawr.orange`,
			exp: `before
rawr: {
  shape: circle
  orange
}
after
`,
		},
		{
			name: "make_scope_multiline_spacing_2",

			text: `before

rawr: {shape: circle}

after
`,
			key: `rawr.orange`,

			expKey: `rawr.orange`,
			exp: `before

rawr: {
  shape: circle
  orange
}

after
`,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			var newKey string
			et := editTest{
				text: tc.text,
				testFunc: func(g *d2graph.Graph) (*d2graph.Graph, error) {
					var err error
					g, newKey, err = d2oracle.Create(g, tc.key)
					return g, err
				},

				exp:    tc.exp,
				expErr: tc.expErr,
				assertions: func(t *testing.T, g *d2graph.Graph) {
					if newKey != tc.expKey {
						t.Fatalf("expected %q but got %q", tc.expKey, newKey)
					}
					if tc.assertions != nil {
						tc.assertions(t, g)
					}
				},
			}
			et.run(t)
		})
	}
}

func TestSet(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name  string
		text  string
		key   string
		tag   *string
		value *string

		expErr     string
		exp        string
		assertions func(t *testing.T, g *d2graph.Graph)
	}{
		{
			name: "base",
			text: ``,
			key:  `square`,

			exp: `square
`,
			assertions: func(t *testing.T, g *d2graph.Graph) {
				if len(g.Objects) != 1 {
					t.Fatalf("expected 1 objects: %#v", g.Objects)
				}
				if g.Objects[0].ID != "square" {
					t.Fatalf("expected g.Objects[0].ID to be square: %#v", g.Objects[0])
				}
				if g.Objects[0].Attributes.Label.MapKey.Value.Unbox() != nil {
					t.Fatalf("expected g.Objects[0].Attributes.Label.Node.Value.Unbox() == nil: %#v", g.Objects[0].Attributes.Label.MapKey.Value)
				}
				if d2format.Format(g.Objects[0].Attributes.Label.MapKey.Key) != "square" {
					t.Fatalf("expected g.Objects[0].Attributes.Label.Node.Key to be square: %#v", g.Objects[0].Attributes.Label.MapKey.Key)
				}
			},
		},
		{
			name:  "edge",
			text:  `x -> y: one`,
			key:   `(x -> y)[0]`,
			value: go2.Pointer(`two`),

			exp: `x -> y: two
`,
			assertions: func(t *testing.T, g *d2graph.Graph) {
				if len(g.Objects) != 2 {
					t.Fatalf("expected 2 objects: %#v", g.Objects)
				}
				if len(g.Edges) != 1 {
					t.Fatalf("expected 1 edge: %#v", g.Edges)
				}
				if g.Edges[0].Src.ID != "x" {
					t.Fatalf("expected g.Edges[0].Src.ID == x: %#v", g.Edges[0].Src.ID)
				}
				if g.Edges[0].Dst.ID != "y" {
					t.Fatalf("expected g.Edges[0].Dst.ID == y: %#v", g.Edges[0].Dst.ID)
				}
				if g.Edges[0].Attributes.Label.Value != "two" {
					t.Fatalf("expected g.Edges[0].Attributes.Label.Value == two: %#v", g.Edges[0].Attributes.Label.Value)
				}
			},
		},
		{
			name:  "shape",
			text:  `square`,
			key:   `square.shape`,
			value: go2.Pointer(`square`),

			exp: `square: {shape: square}
`,
			assertions: func(t *testing.T, g *d2graph.Graph) {
				if len(g.Objects) != 1 {
					t.Fatalf("expected 1 objects: %#v", g.Objects)
				}
				if g.Objects[0].ID != "square" {
					t.Fatalf("expected g.Objects[0].ID to be square: %#v", g.Objects[0])
				}
				if g.Objects[0].Attributes.Shape.Value != d2target.ShapeSquare {
					t.Fatalf("expected g.Objects[0].Attributes.Shape.Value == square: %#v", g.Objects[0].Attributes.Shape.Value)
				}
			},
		},
		{
			name:  "replace_shape",
			text:  `square.shape: square`,
			key:   `square.shape`,
			value: go2.Pointer(`circle`),

			exp: `square.shape: circle
`,
			assertions: func(t *testing.T, g *d2graph.Graph) {
				if len(g.Objects) != 1 {
					t.Fatalf("expected 1 objects: %#v", g.Objects)
				}
				if g.Objects[0].ID != "square" {
					t.Fatalf("expected g.Objects[0].ID to be square: %#v", g.Objects[0])
				}
				if g.Objects[0].Attributes.Shape.Value != d2target.ShapeCircle {
					t.Fatalf("expected g.Objects[0].Attributes.Shape.Value == circle: %#v", g.Objects[0].Attributes.Shape.Value)
				}
			},
		},
		{
			name: "new_style",
			text: `square
`,
			key:   `square.style.opacity`,
			value: go2.Pointer(`0.2`),
			exp: `square: {style.opacity: 0.2}
`,
			assertions: func(t *testing.T, g *d2graph.Graph) {
				if len(g.AST.Nodes) != 1 {
					t.Fatal(g.AST)
				}
				if len(g.Objects) != 1 {
					t.Fatalf("expected 1 object but got %#v", len(g.Objects))
				}
				f, err := strconv.ParseFloat(g.Objects[0].Attributes.Style.Opacity.Value, 64)
				if err != nil || f != 0.2 {
					t.Fatalf("expected g.Objects[0].Map.Nodes[0].MapKey.Value.Number.Value.Float64() == 0.2: %#v", f)
				}
			},
		},
		{
			name: "inline_style",
			text: `square: {style.opacity: 0.2}
`,
			key:   `square.style.fill`,
			value: go2.Pointer(`red`),
			exp: `square: {
  style.opacity: 0.2
  style.fill: red
}
`,
			assertions: func(t *testing.T, g *d2graph.Graph) {
				if len(g.AST.Nodes) != 1 {
					t.Fatal(g.AST)
				}
			},
		},
		{
			name: "expanded_map_style",
			text: `square: {
	style: {
    opacity: 0.1
  }
}
`,
			key:   `square.style.opacity`,
			value: go2.Pointer(`0.2`),
			exp: `square: {
  style: {
    opacity: 0.2
  }
}
`,
			assertions: func(t *testing.T, g *d2graph.Graph) {
				if len(g.AST.Nodes) != 1 {
					t.Fatal(g.AST)
				}
				if len(g.AST.Nodes[0].MapKey.Value.Map.Nodes) != 1 {
					t.Fatalf("expected 1 node within square but got %v", len(g.AST.Nodes[0].MapKey.Value.Map.Nodes))
				}
				f, err := strconv.ParseFloat(g.Objects[0].Attributes.Style.Opacity.Value, 64)
				if err != nil || f != 0.2 {
					t.Fatal(err, f)
				}
			},
		},
		{
			name: "replace_style",
			text: `square.style.opacity: 0.1
`,
			key:   `square.style.opacity`,
			value: go2.Pointer(`0.2`),
			exp: `square.style.opacity: 0.2
`,
			assertions: func(t *testing.T, g *d2graph.Graph) {
				if len(g.AST.Nodes) != 1 {
					t.Fatal(g.AST)
				}
				f, err := strconv.ParseFloat(g.Objects[0].Attributes.Style.Opacity.Value, 64)
				if err != nil || f != 0.2 {
					t.Fatal(err, f)
				}
			},
		},
		{
			name: "replace_style_edgecase",
			text: `square.style.fill: orange
`,
			key:   `square.style.opacity`,
			value: go2.Pointer(`0.2`),
			exp: `square.style.fill: orange
square.style.opacity: 0.2
`,
			assertions: func(t *testing.T, g *d2graph.Graph) {
				if len(g.AST.Nodes) != 2 {
					t.Fatal(g.AST)
				}
				f, err := strconv.ParseFloat(g.Objects[0].Attributes.Style.Opacity.Value, 64)
				if err != nil || f != 0.2 {
					t.Fatal(err, f)
				}
			},
		},
		{
			name: "label_unset",
			text: `square: "Always try to do things in chronological order; it's less confusing that way."
`,
			key:   `square.label`,
			value: nil,

			exp: `square
`,
			assertions: func(t *testing.T, g *d2graph.Graph) {
				if len(g.Objects) != 1 {
					t.Fatalf("expected 1 objects: %#v", g.Objects)
				}
				if g.Objects[0].ID != "square" {
					t.Fatalf("expected g.Objects[0].ID to be square: %#v", g.Objects[0])
				}
				if g.Objects[0].Attributes.Shape.Value == d2target.ShapeSquare {
					t.Fatalf("expected g.Objects[0].Attributes.Shape.Value == square: %#v", g.Objects[0].Attributes.Shape.Value)
				}
			},
		},
		{
			name:  "label",
			text:  `square`,
			key:   `square.label`,
			value: go2.Pointer(`Always try to do things in chronological order; it's less confusing that way.`),

			exp: `square: "Always try to do things in chronological order; it's less confusing that way."
`,
			assertions: func(t *testing.T, g *d2graph.Graph) {
				if len(g.Objects) != 1 {
					t.Fatalf("expected 1 objects: %#v", g.Objects)
				}
				if g.Objects[0].ID != "square" {
					t.Fatalf("expected g.Objects[0].ID to be square: %#v", g.Objects[0])
				}
				if g.Objects[0].Attributes.Shape.Value == d2target.ShapeSquare {
					t.Fatalf("expected g.Objects[0].Attributes.Shape.Value == square: %#v", g.Objects[0].Attributes.Shape.Value)
				}
			},
		},
		{
			name:  "label_replace",
			text:  `square: I am deeply CONCERNED and I want something GOOD for BREAKFAST!`,
			key:   `square`,
			value: go2.Pointer(`Always try to do things in chronological order; it's less confusing that way.`),

			exp: `square: "Always try to do things in chronological order; it's less confusing that way."
`,
			assertions: func(t *testing.T, g *d2graph.Graph) {
				if len(g.AST.Nodes) != 1 {
					t.Fatal(g.AST)
				}
				if len(g.Objects) != 1 {
					t.Fatal(g.Objects)
				}
				if g.Objects[0].ID != "square" {
					t.Fatal(g.Objects[0])
				}
				if g.Objects[0].Attributes.Label.Value == "I am deeply CONCERNED and I want something GOOD for BREAKFAST!" {
					t.Fatal(g.Objects[0].Attributes.Label.Value)
				}
			},
		},
		{
			name:  "map_key_missing",
			text:  `a -> b`,
			key:   `a`,
			value: go2.Pointer(`Never offend people with style when you can offend them with substance.`),

			exp: `a -> b
a: Never offend people with style when you can offend them with substance.
`,
			assertions: func(t *testing.T, g *d2graph.Graph) {
				if len(g.Objects) != 2 {
					t.Fatalf("expected 2 objects: %#v", g.Objects)
				}
				if len(g.Edges) != 1 {
					t.Fatalf("expected 1 edge: %#v", g.Edges)
				}
			},
		},
		{
			name: "nested_alex",
			text: `this: {
  label: do
  test -> here: asdf
}`,
			key: `this.here`,
			value: go2.Pointer(`How much of their influence on you is a result of your influence on them?
A conference is a gathering of important people who singly can do nothing`),

			exp: `this: {
  label: do
  test -> here: asdf
  here: "How much of their influence on you is a result of your influence on them?\nA conference is a gathering of important people who singly can do nothing"
}
`,
			assertions: func(t *testing.T, g *d2graph.Graph) {
				if len(g.Objects) != 3 {
					t.Fatalf("expected 3 objects: %#v", g.Objects)
				}
				if len(g.Edges) != 1 {
					t.Fatalf("expected 1 edge: %#v", g.Edges)
				}
			},
		},
		{
			name: "label_primary",
			text: `oreo: {
 q -> z
}`,
			key:   `oreo`,
			value: go2.Pointer(`QOTD: "It's been Monday all week today."`),

			exp: `oreo: 'QOTD: "It''s been Monday all week today."' {
  q -> z
}
`,
			assertions: func(t *testing.T, g *d2graph.Graph) {
				if len(g.Objects) != 3 {
					t.Fatalf("expected 3 objects: %#v", g.Objects)
				}
				if len(g.Edges) != 1 {
					t.Fatalf("expected 1 edge: %#v", g.Edges)
				}
			},
		},
		{
			name: "edge_index_nested",
			text: `oreo: {
 q -> z
}`,
			key:   `(oreo.q -> oreo.z)[0]`,
			value: go2.Pointer(`QOTD`),

			exp: `oreo: {
  q -> z: QOTD
}
`,
			assertions: func(t *testing.T, g *d2graph.Graph) {
				if len(g.Objects) != 3 {
					t.Fatalf("expected 3 objects: %#v", g.Objects)
				}
				if len(g.Edges) != 1 {
					t.Fatalf("expected 1 edge: %#v", g.Edges)
				}
			},
		},
		{
			name: "edge_index_case",
			text: `Square: {
  Square -> Square 2
}
z: {
  x -> y
}
`,
			key:   `Square.(Square -> Square 2)[0]`,
			value: go2.Pointer(`two`),

			exp: `Square: {
  Square -> Square 2: two
}
z: {
  x -> y
}
`,
			assertions: func(t *testing.T, g *d2graph.Graph) {
				if len(g.Objects) != 6 {
					t.Fatalf("expected 6 objects: %#v", g.Objects)
				}
				if len(g.Edges) != 2 {
					t.Fatalf("expected 2 edges: %#v", g.Edges)
				}
				if g.Edges[0].Attributes.Label.Value != "two" {
					t.Fatalf("expected g.Edges[0].Attributes.Label.Value == two: %#v", g.Edges[0].Attributes.Label.Value)
				}
			},
		},
		{
			name: "icon",
			text: `meow
			`,
			key:   `meow.icon`,
			value: go2.Pointer(`https://icons.terrastruct.com/essentials/087-menu.svg`),

			exp: `meow: {icon: https://icons.terrastruct.com/essentials/087-menu.svg}
`,
			assertions: func(t *testing.T, g *d2graph.Graph) {
				if len(g.Objects) != 1 {
					t.Fatal(g.Objects)
				}
				if g.Objects[0].Attributes.Icon.String() != "https://icons.terrastruct.com/essentials/087-menu.svg" {
					t.Fatal(g.Objects[0].Attributes.Icon.String())
				}
			},
		},
		{
			name: "edge_chain",
			text: `oreo: {
  q -> z -> p: wsup
}`,
			key: `(oreo.q -> oreo.z)[0]`,
			value: go2.Pointer(`QOTD:
  "It's been Monday all week today."`),

			exp: `oreo: {
  q -> z -> p: wsup
  (q -> z)[0]: "QOTD:\n  \"It's been Monday all week today.\""
}
`,
			assertions: func(t *testing.T, g *d2graph.Graph) {
				if len(g.Objects) != 4 {
					t.Fatalf("expected 4 objects: %#v", g.Objects)
				}
				if len(g.Edges) != 2 {
					t.Fatalf("expected 2 edges: %#v", g.Edges)
				}
			},
		},
		{
			name: "edge_nested_label_set",
			text: `oreo: {
  q -> z: wsup
}`,
			key:   `(oreo.q -> oreo.z)[0].label`,
			value: go2.Pointer(`yo`),

			exp: `oreo: {
  q -> z: yo
}
`,
			assertions: func(t *testing.T, g *d2graph.Graph) {
				if len(g.Objects) != 3 {
					t.Fatalf("expected 3 objects: %#v", g.Objects)
				}
				if len(g.Edges) != 1 {
					t.Fatalf("expected 1 edge: %#v", g.Edges)
				}
				if g.Edges[0].Src.ID != "q" {
					t.Fatal(g.Edges[0].Src.ID)
				}
			},
		},
		{
			name: "shape_nested_style_set",
			text: `x
`,
			key:   `x.style.opacity`,
			value: go2.Pointer(`0.4`),

			exp: `x: {style.opacity: 0.4}
`,
		},
		{
			name: "edge_nested_style_set",
			text: `oreo: {
  q -> z: wsup
}
`,
			key:   `(oreo.q -> oreo.z)[0].style.opacity`,
			value: go2.Pointer(`0.4`),

			exp: `oreo: {
  q -> z: wsup {style.opacity: 0.4}
}
`,
			assertions: func(t *testing.T, g *d2graph.Graph) {
				assert.Equal(t, 3, len(g.Objects))
				assert.Equal(t, 1, len(g.Edges))
				assert.Equal(t, "q", g.Edges[0].Src.ID)
				assert.Equal(t, "0.4", g.Edges[0].Attributes.Style.Opacity.Value)
			},
		},
		{
			name: "edge_chain_append_style",
			text: `x -> y -> z
`,
			key:   `(x -> y)[0].style.animated`,
			value: go2.Pointer(`true`),

			exp: `x -> y -> z
(x -> y)[0].style.animated: true
`,
		},
		{
			name: "edge_chain_existing_style",
			text: `x -> y -> z
(y -> z)[0].style.opacity: 0.4
`,
			key:   `(y -> z)[0].style.animated`,
			value: go2.Pointer(`true`),

			exp: `x -> y -> z
(y -> z)[0].style.opacity: 0.4
(y -> z)[0].style.animated: true
`,
		},
		{
			name: "edge_key_and_key",
			text: `a
a.b -> a.c
`,
			key:   `a.(b -> c)[0].style.animated`,
			value: go2.Pointer(`true`),

			exp: `a
a.b -> a.c: {style.animated: true}
`,
		},
		{
			name: "edge_label",
			text: `a -> b: "yo"
`,
			key:   `(a -> b)[0].style.animated`,
			value: go2.Pointer(`true`),

			exp: `a -> b: "yo" {style.animated: true}
`,
		},
		{
			name: "edge_append_style",
			text: `x -> y
`,
			key:   `(x -> y)[0].style.animated`,
			value: go2.Pointer(`true`),

			exp: `x -> y: {style.animated: true}
`,
		},
		{
			name: "edge_merge_style",
			text: `x -> y: {
	style: {
    opacity: 0.4
  }
}
`,
			key:   `(x -> y)[0].style.animated`,
			value: go2.Pointer(`true`),

			exp: `x -> y: {
  style: {
    opacity: 0.4
    animated: true
  }
}
`,
		},
		{
			name: "edge_chain_nested_set",
			text: `oreo: {
  q -> z -> p: wsup
}`,
			key:   `(oreo.q -> oreo.z)[0].style.opacity`,
			value: go2.Pointer(`0.4`),

			exp: `oreo: {
  q -> z -> p: wsup
  (q -> z)[0].style.opacity: 0.4
}
`,
			assertions: func(t *testing.T, g *d2graph.Graph) {
				if len(g.Objects) != 4 {
					t.Fatalf("expected 4 objects: %#v", g.Objects)
				}
				if len(g.Edges) != 2 {
					t.Fatalf("expected 2 edges: %#v", g.Edges)
				}
				if g.Edges[0].Src.ID != "q" {
					t.Fatal(g.Edges[0].Src.ID)
				}
				if g.Edges[0].Attributes.Style.Opacity.Value != "0.4" {
					t.Fatal(g.Edges[0].Attributes.Style.Opacity.Value)
				}
			},
		},
		{
			name: "block_string_oneline",

			text:  ``,
			key:   `x`,
			tag:   go2.Pointer("md"),
			value: go2.Pointer(`|||what's up|||`),

			exp: `x: ||||md |||what's up||| ||||
`,
		},
		{
			name: "block_string_multiline",

			text: ``,
			key:  `x`,
			tag:  go2.Pointer("md"),
			value: go2.Pointer(`# header
He has not acquired a fortune; the fortune has acquired him.
He has not acquired a fortune; the fortune has acquired him.`),

			exp: `x: |md
  # header
  He has not acquired a fortune; the fortune has acquired him.
  He has not acquired a fortune; the fortune has acquired him.
|
`,
		},
		// TODO: pass
		/*
			{
				name: "oneline_constraint",

				text: `My Table: {
					shape: sql_table
					column: int
				}
				`,
				key:   `My Table.column.constraint`,
				value: utils.Pointer("PK"),

				exp: `My Table: {
					shape: sql_table
					column: int {constraint: PK}
				}
				`,
			},
		*/
		// TODO: pass
		/*
					{
						name: "oneline_style",

						text: `foo: bar
			`,
						key:   `foo.style_fill`,
						value: utils.Pointer("red"),

						exp: `foo: bar {style_fill: red}
			`,
					},
		*/

		{
			name: "errors/bad_tag",

			text: `x.icon: hello
`,
			key: "x.icon",
			tag: go2.Pointer("one two"),
			value: go2.Pointer(`three
four
five
six
`),

			expErr: `failed to set "x.icon" to "one two" "\"three\\nfour\\nfive\\nsix\\n\"": spaces are not allowed in blockstring tags`,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			et := editTest{
				text: tc.text,
				testFunc: func(g *d2graph.Graph) (*d2graph.Graph, error) {
					return d2oracle.Set(g, tc.key, tc.tag, tc.value)
				},

				exp:        tc.exp,
				expErr:     tc.expErr,
				assertions: tc.assertions,
			}
			et.run(t)
		})
	}
}

func TestRename(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name string

		text    string
		key     string
		newName string

		expErr     string
		exp        string
		assertions func(t *testing.T, g *d2graph.Graph)
	}{
		{
			name: "flat",

			text: `nerve-gift-earther
`,
			key:     `nerve-gift-earther`,
			newName: `---`,

			exp: `"---"
`,
			assertions: func(t *testing.T, g *d2graph.Graph) {
				if len(g.Objects) != 1 {
					t.Fatalf("expected one object: %#v", g.Objects)
				}
				if g.Objects[0].ID != `"---"` {
					t.Fatalf("unexpected object id: %q", g.Objects[0].ID)
				}
			},
		},
		{
			name: "generated",

			text: `Square
`,
			key:     `Square`,
			newName: `Square`,

			exp: `Square
`,
		},
		{
			name: "near",

			text: `x: {
  near: y
}
y
`,
			key:     `y`,
			newName: `z`,

			exp: `x: {
  near: z
}
z
`,
		},
		{
			name: "conflict",

			text: `lalal
la
`,
			key:     `lalal`,
			newName: `la`,

			exp: `la 2
la
`,
		},
		{
			name: "conflict 2",

			text: `1.2.3: {
  4
  5
}
`,
			key:     "1.2.3.4",
			newName: "5",

			exp: `1.2.3: {
  5 2
  5
}
`,
		},
		{
			name: "conflict_with_dots",

			text: `"a.b"
y
`,
			key:     "y",
			newName: "a.b",

			exp: `"a.b"
"a.b 2"
`,
		},
		{
			name: "nested",

			text: `x.y.z.q.nerve-gift-earther
x.y.z.q: {
  nerve-gift-earther
}
`,
			key:     `x.y.z.q.nerve-gift-earther`,
			newName: `nerve-gift-jingler`,

			exp: `x.y.z.q.nerve-gift-jingler
x.y.z.q: {
  nerve-gift-jingler
}
`,
			assertions: func(t *testing.T, g *d2graph.Graph) {
				if len(g.Objects) != 5 {
					t.Fatalf("expected five objects: %#v", g.Objects)
				}
				if g.Objects[4].AbsID() != "x.y.z.q.nerve-gift-jingler" {
					t.Fatalf("unexpected object absolute id: %q", g.Objects[4].AbsID())
				}
			},
		},
		{
			name: "edges",

			text: `q.z -> p.k -> q.z -> l.a -> q.z
q: {
  q -> + -> z
  z: label
}
`,
			key:     `q.z`,
			newName: `%%%`,

			exp: `q.%%% -> p.k -> q.%%% -> l.a -> q.%%%
q: {
  q -> + -> %%%
  %%%: label
}
`,
			assertions: func(t *testing.T, g *d2graph.Graph) {
				if len(g.Objects) != 8 {
					t.Fatalf("expected eight objects: %#v", g.Objects)
				}
				if g.Objects[1].AbsID() != "q.%%%" {
					t.Fatalf("unexpected object absolute ID: %q", g.Objects[1].AbsID())
				}
			},
		},
		{
			name: "container",

			text: `ok.q.z -> p.k -> ok.q.z -> l.a -> ok.q.z
ok.q: {
  q -> + -> z
  z: label
}
ok: {
  q: {
    i
  }
}
(ok.q.z -> p.k)[0]: "furbling, v.:"
more.(ok.q.z -> p.k): "furbling, v.:"
`,
			key:     `ok.q`,
			newName: `<gosling>`,

			exp: `ok."<gosling>".z -> p.k -> ok."<gosling>".z -> l.a -> ok."<gosling>".z
ok."<gosling>": {
  q -> + -> z
  z: label
}
ok: {
  "<gosling>": {
    i
  }
}
(ok."<gosling>".z -> p.k)[0]: "furbling, v.:"
more.(ok.q.z -> p.k): "furbling, v.:"
`,
			assertions: func(t *testing.T, g *d2graph.Graph) {
				if len(g.Objects) != 16 {
					t.Fatalf("expected 16 objects: %#v", g.Objects)
				}
				if g.Objects[2].AbsID() != `ok."<gosling>".z` {
					t.Fatalf("unexpected object absolute ID: %q", g.Objects[1].AbsID())
				}
			},
		},
		{
			name: "complex_edge_1",

			text: `a.b.(x -> y).q.z
`,
			key:     "a.b",
			newName: "ooo",

			exp: `a.ooo.(x -> y).q.z
`,
			assertions: func(t *testing.T, g *d2graph.Graph) {
				if len(g.Objects) != 4 {
					t.Fatalf("expected 4 objects: %#v", g.Objects)
				}
				if len(g.Edges) != 1 {
					t.Fatalf("expected 1 edge: %#v", g.Edges)
				}
			},
		},
		{
			name: "complex_edge_2",

			text: `a.b.(x -> y).q.z
`,
			key:     "a.b.x",
			newName: "papa",

			exp: `a.b.(papa -> y).q.z
`,
			assertions: func(t *testing.T, g *d2graph.Graph) {
				if len(g.Objects) != 4 {
					t.Fatalf("expected 4 objects: %#v", g.Objects)
				}
				if len(g.Edges) != 1 {
					t.Fatalf("expected 1 edge: %#v", g.Edges)
				}
			},
		},
		/* TODO: handle edge keys
				{
					name: "complex_edge_3",

					text: `a.b.(x -> y).q.z
		`,
					key:     "a.b.(x -> y)[0].q",
					newName: "zoink",

					exp: `a.b.(x -> y).zoink.z
		`,
					assertions: func(t *testing.T, g *d2graph.Graph) {
						if len(g.Objects) != 4 {
							t.Fatalf("expected 4 objects: %#v", g.Objects)
						}
						if len(g.Edges) != 1 {
							t.Fatalf("expected 1 edge: %#v", g.Edges)
						}
					},
				},
		*/
		{
			name: "arrows",

			text: `x -> y
`,
			key:     "(x -> y)[0]",
			newName: "(x <- y)[0]",

			exp: `x <- y
`,
			assertions: func(t *testing.T, g *d2graph.Graph) {
				if len(g.Objects) != 2 {
					t.Fatalf("expected 2 objects: %#v", g.Objects)
				}
				if len(g.Edges) != 1 {
					t.Fatalf("expected 1 edge: %#v", g.Edges)
				}
				if !g.Edges[0].SrcArrow || g.Edges[0].DstArrow {
					t.Fatalf("expected src arrow and no dst arrow: %#v", g.Edges[0])
				}
			},
		},
		{
			name: "arrows_complex",

			text: `a.b.(x -- y).q.z
`,
			key:     "a.b.(x -- y)[0]",
			newName: "(x <-> y)[0]",

			exp: `a.b.(x <-> y).q.z
`,
			assertions: func(t *testing.T, g *d2graph.Graph) {
				if len(g.Objects) != 4 {
					t.Fatalf("expected 4 objects: %#v", g.Objects)
				}
				if len(g.Edges) != 1 {
					t.Fatalf("expected 1 edge: %#v", g.Edges)
				}
				if !g.Edges[0].SrcArrow || !g.Edges[0].DstArrow {
					t.Fatalf("expected src arrow and dst arrow: %#v", g.Edges[0])
				}
			},
		},
		{
			name: "arrows_chain",

			text: `x -> y -> z -> q
`,
			key:     "(x -> y)[0]",
			newName: "(x <-> y)[0]",

			exp: `x <-> y -> z -> q
`,
			assertions: func(t *testing.T, g *d2graph.Graph) {
				if len(g.Objects) != 4 {
					t.Fatalf("expected 4 objects: %#v", g.Objects)
				}
				if len(g.Edges) != 3 {
					t.Fatalf("expected 3 edges: %#v", g.Edges)
				}
				if !g.Edges[0].SrcArrow || !g.Edges[0].DstArrow {
					t.Fatalf("expected src arrow and dst arrow: %#v", g.Edges[0])
				}
			},
		},
		{
			name: "arrows_trim_common",

			text: `x.(x -> y -> z -> q)
`,
			key:     "(x.x -> x.y)[0]",
			newName: "(x.x <-> x.y)[0]",

			exp: `x.(x <-> y -> z -> q)
`,
			assertions: func(t *testing.T, g *d2graph.Graph) {
				if len(g.Objects) != 5 {
					t.Fatalf("expected 5 objects: %#v", g.Objects)
				}
				if len(g.Edges) != 3 {
					t.Fatalf("expected 3 edges: %#v", g.Edges)
				}
				if !g.Edges[0].SrcArrow || !g.Edges[0].DstArrow {
					t.Fatalf("expected src arrow and dst arrow: %#v", g.Edges[0])
				}
			},
		},
		{
			name: "arrows_trim_common_2",

			text: `x.x -> x.y -> x.z -> x.q)
`,
			key:     "(x.x -> x.y)[0]",
			newName: "(x.x <-> x.y)[0]",

			exp: `x.x <-> x.y -> x.z -> x.q)
`,
			assertions: func(t *testing.T, g *d2graph.Graph) {
				if len(g.Objects) != 5 {
					t.Fatalf("expected 5 objects: %#v", g.Objects)
				}
				if len(g.Edges) != 3 {
					t.Fatalf("expected 3 edges: %#v", g.Edges)
				}
				if !g.Edges[0].SrcArrow || !g.Edges[0].DstArrow {
					t.Fatalf("expected src arrow and dst arrow: %#v", g.Edges[0])
				}
			},
		},

		{
			name: "errors/empty_key",

			text: ``,
			key:  "",

			expErr: `failed to rename "" to "": empty map key: ""`,
		},
		{
			name: "errors/nonexistent",

			text:    ``,
			key:     "1.2.3.4",
			newName: "bic",

			expErr: `failed to rename "1.2.3.4" to "bic": key referenced by from does not exist`,
		},

		{
			name: "errors/reserved_keys",

			text: `x.icon: hello
`,
			key:     "x.icon",
			newName: "near",
			expErr:  `failed to rename "x.icon" to "near": cannot rename to reserved keyword: "near"`,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			et := editTest{
				text: tc.text,
				testFunc: func(g *d2graph.Graph) (*d2graph.Graph, error) {
					return d2oracle.Rename(g, tc.key, tc.newName)
				},

				exp:        tc.exp,
				expErr:     tc.expErr,
				assertions: tc.assertions,
			}
			et.run(t)
		})
	}
}

func TestMove(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		skip bool
		name string

		text   string
		key    string
		newKey string

		expErr     string
		exp        string
		assertions func(t *testing.T, g *d2graph.Graph)
	}{
		{
			name: "basic",

			text: `a
`,
			key:    `a`,
			newKey: `b`,

			exp: `b
`,
			assertions: func(t *testing.T, g *d2graph.Graph) {
				assert.Equal(t, len(g.Objects), 1)
				assert.Equal(t, g.Objects[0].ID, "b")
			},
		},
		{
			name: "basic_nested",

			text: `a: {
  b
}
`,
			key:    `a.b`,
			newKey: `a.c`,

			exp: `a: {
  c
}
`,
			assertions: func(t *testing.T, g *d2graph.Graph) {
				assert.Equal(t, len(g.Objects), 2)
				assert.Equal(t, g.Objects[1].ID, "c")
			},
		},
		{
			name: "rename_2",

			text: `a: {
  b 2
  y 2
}
b 2
x
`,
			key:    `a`,
			newKey: `x.a`,

			exp: `b
y 2

b 2
x: {
  a
}
`,
		},
		{
			name: "parentheses",

			text: `x -> y (z)
z: ""
`,
			key:    `"y (z)"`,
			newKey: `z.y (z)`,

			exp: `x -> z.y (z)
z: ""
`,
		},
		{
			name: "into_container_existing_map",

			text: `a: {
  b
}
c
`,
			key:    `c`,
			newKey: `a.c`,

			exp: `a: {
  b
  c
}
`,
			assertions: func(t *testing.T, g *d2graph.Graph) {
				assert.Equal(t, len(g.Objects), 3)
				assert.Equal(t, "a", g.Objects[0].ID)
				assert.Equal(t, 2, len(g.Objects[0].Children))
			},
		},
		{
			name: "into_container_with_flat_keys",

			text: `a
c: {
  style.opacity: 0.4
  style.fill: "#FFFFFF"
  style.stroke: "#FFFFFF"
}
`,
			key:    `c`,
			newKey: `a.c`,

			exp: `a: {
  c: {
    style.opacity: 0.4
    style.fill: "#FFFFFF"
    style.stroke: "#FFFFFF"
  }
}
`,
		},
		{
			name: "into_container_nonexisting_map",

			text: `a
c
`,
			key:    `c`,
			newKey: `a.c`,

			exp: `a: {
  c
}
`,
			assertions: func(t *testing.T, g *d2graph.Graph) {
				assert.Equal(t, len(g.Objects), 2)
				assert.Equal(t, "a", g.Objects[0].ID)
				assert.Equal(t, 1, len(g.Objects[0].Children))
			},
		},
		{
			name: "basic_out_of_container",

			text: `a: {
  b
}
`,
			key:    `a.b`,
			newKey: `b`,

			exp: `a
b
`,
			assertions: func(t *testing.T, g *d2graph.Graph) {
				assert.Equal(t, len(g.Objects), 2)
				assert.Equal(t, "a", g.Objects[0].ID)
				assert.Equal(t, 0, len(g.Objects[0].Children))
			},
		},
		{
			name: "partial_slice",

			text: `a: {
  b
}
a.b
`,
			key:    `a.b`,
			newKey: `b`,

			exp: `a
b
`,
		},
		{
			name: "partial_edge_slice",

			text: `a: {
  b
}
a.b -> c
`,
			key:    `a.b`,
			newKey: `b`,

			exp: `a
b -> c
b
`,
		},
		{
			name: "full_edge_slice",

			text: `a: {
	b: {
    c
  }
  b.c -> d
}
a.b.c -> a.d
`,
			key:    `a.b.c`,
			newKey: `c`,

			exp: `a: {
  b
  _.c -> d
}
c -> a.d
c
`,
		},
		{
			name: "full_slice",

			text: `a: {
	b: {
    c
  }
  b.c
}
a.b.c
`,
			key:    `a.b.c`,
			newKey: `c`,

			exp: `a: {
  b
}
c
`,
		},
		{
			name: "slice_style",

			text: `a: {
  b
}
a.b.icon: https://icons.terrastruct.com/essentials/142-target.svg
`,
			key:    `a.b`,
			newKey: `b`,

			exp: `a
b.icon: https://icons.terrastruct.com/essentials/142-target.svg
b
`,
		},
		{
			name: "between_containers",

			text: `a: {
  b
}
c
`,
			key:    `a.b`,
			newKey: `c.b`,

			exp: `a
c: {
  b
}
`,
			assertions: func(t *testing.T, g *d2graph.Graph) {
				assert.Equal(t, len(g.Objects), 3)
				assert.Equal(t, "a", g.Objects[0].ID)
				assert.Equal(t, 0, len(g.Objects[0].Children))
				assert.Equal(t, "c", g.Objects[1].ID)
				assert.Equal(t, 1, len(g.Objects[1].Children))
			},
		},
		{
			name: "hoist_container_children",

			text: `a: {
  b
  c
}
d
`,
			key:    `a`,
			newKey: `d.a`,

			exp: `b
c

d: {
  a
}
`,
		},
		{
			name: "middle_container",

			text: `x: {
  y: {
    z
  }
}
`,
			key:    `x.y`,
			newKey: `y`,

			exp: `x: {
  z
}
y
`,
		},
		{
			// a.b does not move from its scope, just extends path
			name: "extend_stationary_path",

			text: `a.b
a: {
	b
	c
}
`,
			key:    `a.b`,
			newKey: `a.c.b`,

			exp: `a.c.b
a: {
  c: {
    b
  }
}
`,
			assertions: func(t *testing.T, g *d2graph.Graph) {
				assert.Equal(t, len(g.Objects), 3)
			},
		},
		{
			name: "extend_map",

			text: `a.b: {
  e
}
a: {
	b
	c
}
`,
			key:    `a.b`,
			newKey: `a.c.b`,

			exp: `a: {
  e
}
a: {
  c: {
    b
  }
}
`,
		},
		{
			name: "into_container_with_flat_style",

			text: `x.style.border-radius: 5
y
`,
			key:    `y`,
			newKey: `x.y`,

			exp: `x: {
  style.border-radius: 5
  y
}
`,
		},
		{
			name: "flat_between_containers",

			text: `a.b
c
`,
			key:    `a.b`,
			newKey: `c.b`,

			exp: `a
c: {
  b
}
`,
			assertions: func(t *testing.T, g *d2graph.Graph) {
				assert.Equal(t, len(g.Objects), 3)
			},
		},
		{
			name: "flat_middle_container",

			text: `a.b.c
d
`,
			key:    `a.b`,
			newKey: `d.b`,

			exp: `a.c
d: {
  b
}
`,
			assertions: func(t *testing.T, g *d2graph.Graph) {
				assert.Equal(t, len(g.Objects), 4)
			},
		},
		{
			name: "flat_merge",

			text: `a.b
c.d: meow
`,
			key:    `a.b`,
			newKey: `c.b`,

			exp: `a
c: {
  d: meow
  b
}
`,
			assertions: func(t *testing.T, g *d2graph.Graph) {
				assert.Equal(t, len(g.Objects), 4)
			},
		},
		{
			name: "flat_reparent_with_value",
			text: `a.b: "yo"
`,
			key:    `a.b`,
			newKey: `b`,

			exp: `a
b: "yo"
`,
		},
		{
			name: "flat_reparent_with_map_value",
			text: `a.b: {
  shape: hexagon
}
`,
			key:    `a.b`,
			newKey: `b`,

			exp: `a
b: {
  shape: hexagon
}
`,
		},
		{
			name: "flat_reparent_with_mixed_map_value",
			text: `a.b: {
  # this is reserved
  shape: hexagon
  # this is not
  c
}
`,
			key:    `a.b`,
			newKey: `b`,

			exp: `a: {
  # this is not
  c
}
b: {
  # this is reserved
  shape: hexagon
}
`,
		},
		{
			name: "flat_style",

			text: `a.style.opacity: 0.4
a.style.fill: black
b
`,
			key:    `a`,
			newKey: `b.a`,

			exp: `b: {
  a.style.opacity: 0.4
  a.style.fill: black
}
`,
		},
		{
			name: "flat_nested_merge",

			text: `a.b.c.d.e
p.q.b.m.o
`,
			key:    `a.b.c`,
			newKey: `p.q.z`,

			exp: `a.b.d.e
p.q: {
  b.m.o
  z
}
`,
		},
		{
			// We open up only the most nested
			name: "flat_nested_merge_multiple_refs",

			text: `a: {
  b.c.d
  e.f
  e.g
}
a.b.c
a.b.c.q
`,
			key:    `a.e`,
			newKey: `a.b.c.e`,

			exp: `a: {
  b.c: {
    d
    e
  }
  f
  g
}
a.b.c
a.b.c.q
`,
		},
		{
			// TODO
			skip: true,
			// Choose to move to a reference that is less nested but has an existing map
			name: "less_nested_map",

			text: `a: {
  b: {
    c
  }
}
a.b.c: {
  d
}
e
`,
			key:    `e`,
			newKey: `a.b.c.e`,

			exp: `a: {
  b: {
    c
  }
}
a.b.c: {
  d
  e
}
`,
		},
		{
			name: "near",

			text: `x: {
  near: y
}
y
`,
			key:    `y`,
			newKey: `x.y`,

			exp: `x: {
  near: x.y
  y
}
`,
		},
		{
			name: "container_near",

			text: `x: {
  y: {
    near: x.a.b.z
  }
  a.b.z
}
y
`,
			key:    `x.a.b`,
			newKey: `y.a`,

			exp: `x: {
  y: {
    near: x.a.z
  }
  a.z
}
y: {
  a
}
`,
		},
		{
			name: "nhooyr_one",

			text: `a: {
  b.c
}
d
`,
			key:    `a.b`,
			newKey: `d.q`,

			exp: `a: {
  c
}
d: {
  q
}
`,
		},
		{
			name: "nhooyr_two",

			text: `a: {
  b.c -> meow
}
d: {
  x
}
`,
			key:    `a.b`,
			newKey: `d.b`,

			exp: `a: {
  c -> meow
}
d: {
  x
  b
}
`,
		},
		{
			name: "unique_name",

			text: `a: {
  b
}
a.b
c: {
  b
}
`,
			key:    `c.b`,
			newKey: `a.b`,

			exp: `a: {
  b
  b 2
}
a.b
c
`,
		},
		{
			name: "unique_name_with_references",

			text: `a: {
  b
}
d -> c.b
c: {
  b
}
`,
			key:    `c.b`,
			newKey: `a.b`,

			exp: `a: {
  b
  b 2
}
d -> a.b 2
c
`,
		},
		{
			name: "map_transplant",

			text: `a: {
  b
  style: {
    opacity: 0.4
  }
  c
  label: "yo"
}
d
`,
			key:    `a`,
			newKey: `d.a`,

			exp: `b

c

d: {
  a: {
    style: {
      opacity: 0.4
    }

    label: "yo"
  }
}
`,
		},
		{
			name: "map_with_label",

			text: `a: "yo" {
  c
}
d
`,
			key:    `a`,
			newKey: `d.a`,

			exp: `c

d: {
  a: "yo"
}
`,
		},
		{
			name: "underscore_merge",

			text: `a: {
	_.b: "yo"
}
b: "what"
c
`,
			key:    `b`,
			newKey: `c.b`,

			exp: `a

c: {
  b: "yo"
  b: "what"
}
`,
		},
		{
			name: "underscore_children",

			text: `a: {
  _.b
}
b
`,
			key:    `b`,
			newKey: `c`,

			exp: `a: {
  _.c
}
c
`,
		},
		{
			name: "underscore_transplant",

			text: `a: {
  b: {
    _.c
  }
}
`,
			key:    `a.c`,
			newKey: `c`,

			exp: `a: {
  b
}
c
`,
		},
		{
			name: "underscore_split",

			text: `a: {
  b: {
    _.c.f
  }
}
`,
			key:    `a.c`,
			newKey: `c`,

			exp: `a: {
  b: {
    _.f
  }
}
c
`,
		},
		{
			name: "underscore_edge_container_1",

			text: `a: {
  _.b -> c
}
`,
			key:    `b`,
			newKey: `a.b`,

			exp: `a: {
  b -> c
}
`,
		},
		{
			name: "underscore_edge_container_2",

			text: `a: {
  _.b -> c
}
`,
			key:    `b`,
			newKey: `a.c.b`,

			exp: `a: {
  c.b -> c
}
`,
		},
		{
			name: "underscore_edge_container_3",

			text: `a: {
  _.b -> c
}
`,
			key:    `b`,
			newKey: `d`,

			exp: `a: {
  _.d -> c
}
`,
		},
		{
			name: "underscore_edge_container_4",

			text: `a: {
  _.b -> c
}
`,
			key:    `b`,
			newKey: `a.f`,

			exp: `a: {
  f -> c
}
`,
		},
		{
			name: "underscore_edge_container_5",

			text: `a: {
  _.b -> _.c
}
`,
			key:    `b`,
			newKey: `c.b`,

			exp: `a: {
  _.c.b -> _.c
}
`,
		},
		{
			name: "underscore_edge_split",

			text: `a: {
  b: {
    _.c.f -> yo
  }
}
`,
			key:    `a.c`,
			newKey: `c`,

			exp: `a: {
  b: {
    _.f -> yo
  }
}
c
`,
		},
		{
			name: "underscore_split_out",

			text: `a: {
  b: {
    _.c.f
  }
  c: {
    e
  }
}
`,
			key:    `a.c.f`,
			newKey: `a.c.e.f`,

			exp: `a: {
  b: {
    _.c
  }
  c: {
    e: {
      f
    }
  }
}
`,
		},
		{
			name: "underscore_edge_children",

			text: `a: {
  _.b -> c
}
b
`,
			key:    `b`,
			newKey: `c`,

			exp: `a: {
  _.c -> c
}
c
`,
		},
		{
			name: "move_container_children",

			text: `b: {
  p
  q
}
a
d
`,
			key:    `b`,
			newKey: `d.b`,

			exp: `p
q

a
d: {
  b
}
`,
		},
		{
			name: "move_container_conflict_children",

			text: `x: {
  a
  b
}
a
d
`,
			key:    `x`,
			newKey: `d.x`,

			exp: `a 2
b

a
d: {
  x
}
`,
		},
		{
			name: "edge_conflict",

			text: `x.y.a -> x.y.b
y
`,
			key:    `x`,
			newKey: `y.x`,

			exp: `y 2.a -> y 2.b
y: {
  x
}
`,
		},
		{
			name: "edge_basic",

			text: `a -> b
`,
			key:    `a`,
			newKey: `c`,

			exp: `c -> b
`,
		},
		{
			name: "edge_nested_basic",

			text: `a: {
  b -> c
}
`,
			key:    `a.b`,
			newKey: `a.d`,

			exp: `a: {
  d -> c
}
`,
		},
		{
			name: "edge_into_container",

			text: `a: {
  d
}
b -> c
`,
			key:    `b`,
			newKey: `a.b`,

			exp: `a: {
  d
}
a.b -> c
`,
		},
		{
			name: "edge_out_of_container",

			text: `a: {
  b -> c
}
`,
			key:    `a.b`,
			newKey: `b`,

			exp: `a: {
  _.b -> c
}
`,
		},
		{
			name: "connected_nested",

			text: `x -> y.z
`,
			key:    `y.z`,
			newKey: `z`,

			exp: `x -> z
y
`,
		},
		{
			name: "chain_connected_nested",

			text: `y.z -> x -> y.z
`,
			key:    `y.z`,
			newKey: `z`,

			exp: `z -> x -> z
y
`,
		},
		{
			name: "chain_connected_nested_no_extra_create",

			text: `y.b -> x -> y.z
`,
			key:    `y.z`,
			newKey: `z`,

			exp: `y.b -> x -> z
`,
		},
		{
			name: "edge_across_containers",

			text: `a: {
  b -> c
}
d
`,
			key:    `a.b`,
			newKey: `d.b`,

			exp: `a: {
  _.d.b -> c
}
d
`,
		},
		{
			name: "move_out_of_edge",

			text: `a.b.c -> d.e.f
`,
			key:    `a.b`,
			newKey: `q`,

			exp: `a.c -> d.e.f
q
`,
		},
		{
			name: "move_out_of_nested_edge",

			text: `a.b.c -> d.e.f
`,
			key:    `a.b`,
			newKey: `d.e.q`,

			exp: `a.c -> d.e.f
d.e: {
  q
}
`,
		},
		{
			name: "append_multiple_styles",

			text: `a: {
  style: {
    opacity: 0.4
  }
}
a: {
  style: {
    fill: "red"
  }
}
d
`,
			key:    `a`,
			newKey: `d.a`,

			exp: `d: {
  a: {
    style: {
      opacity: 0.4
    }
  }
  a: {
    style: {
      fill: "red"
    }
  }
}
`,
		},
		{
			name: "move_into_key_with_value",

			text: `a: meow
b
`,
			key:    `b`,
			newKey: `a.b`,

			exp: `a: meow {
  b
}
`,
		},
		{
			name: "gnarly_1",

			text: `a.b.c -> d.e.f
b: meow {
	p: "eyy"
  q
  p.p -> q.q
}
b.p.x -> d
`,
			key:    `b`,
			newKey: `d.b`,

			exp: `a.b.c -> d.e.f
d: {
  b: meow
}
p: "eyy"
q
p.p -> q.q

p.x -> d
`,
		},
		{
			name: "reuse_map",

			text: `a: {
  b: {
    hey
  }
  b.yo
}
k
`,
			key:    `k`,
			newKey: `a.b.k`,

			exp: `a: {
  b: {
    hey
    k
  }
  b.yo
}
`,
		},
		{
			// TODO the heuristic for splitting open new maps should be only if the key has no existing maps and it also has either zero or one children. if it has two children or more then we should not be opening a map and just append the key at the most nested map.
			//       first loop over explicit references from first to last.
			//
			// explicit ref means its the leaf disregarding reserved fields.
			// implicit ref means there is a shape declared after the target element.
			//
			// then loop over the implicit references and only if there is no explicit ref do you need to add the implicit ref to the scope but only if appended == false (which would be set when looping through explicit refs).
			skip: true,
			name: "merge_nested_flat",

			text: `a: {
  b.c
  b.d
  b.e.g
}
k
`,
			key:    `k`,
			newKey: `a.b.k`,

			exp: `a: {
  b.c
  b.d
  b.e.g
  b.k
}
`,
		},
		{
			name: "merge_nested_maps",

			text: `a: {
  b.c
  b.d
  b.e.g
  b.d: {
    o
  }
}
k
`,
			key:    `k`,
			newKey: `a.b.k`,

			exp: `a: {
  b.c
  b.d
  b.e.g
  b: {
    d: {
      o
    }
    k
  }
}
`,
		},
		{
			name: "merge_reserved",

			text: `a: {
  b.c
	b.label: "yo"
	b.label: "hi"
  b.e.g
}
k
`,
			key:    `k`,
			newKey: `a.b.k`,

			exp: `a: {
  b.c
  b.label: "yo"
  b.label: "hi"
  b: {
    e.g
    k
  }
}
`,
		},
		{
			name: "multiple_nesting_levels",

			text: `a: {
	b: {
    c
    c.g
  }
  b.c.d
  x
}
a.b.c.f
`,
			key:    `a.x`,
			newKey: `a.b.c.x`,

			exp: `a: {
  b: {
    c
    c: {
      g
      x
    }
  }
  b.c.d
}
a.b.c.f
`,
		},
		{
			name: "edge_chain_basic",

			text: `a -> b -> c
`,
			key:    `a`,
			newKey: `d`,

			exp: `d -> b -> c
`,
		},
		{
			name: "edge_chain_into_container",

			text: `a -> b -> c
d
`,
			key:    `a`,
			newKey: `d.a`,

			exp: `d.a -> b -> c
d
`,
		},
		{
			name: "edge_chain_out_container",

			text: `a: {
  b -> c -> d
}
`,
			key:    `a.c`,
			newKey: `c`,

			exp: `a: {
  b -> _.c -> d
}
`,
		},
		{
			name: "edge_chain_circular",

			text: `a: {
  b -> c -> b
}
`,
			key:    `a.b`,
			newKey: `b`,

			exp: `a: {
  _.b -> c -> _.b
}
`,
		},
	}

	for _, tc := range testCases {
		if tc.skip {
			continue
		}
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			et := editTest{
				text: tc.text,
				testFunc: func(g *d2graph.Graph) (*d2graph.Graph, error) {
					objectsBefore := len(g.Objects)
					var err error
					g, err = d2oracle.Move(g, tc.key, tc.newKey)
					if err == nil {
						objectsAfter := len(g.Objects)
						if objectsBefore != objectsAfter {
							println(d2format.Format(g.AST))
							return nil, fmt.Errorf("move cannot destroy or create objects: found %d objects before and %d objects after", objectsBefore, objectsAfter)
						}
					}
					return g, err
				},

				exp:        tc.exp,
				expErr:     tc.expErr,
				assertions: tc.assertions,
			}
			et.run(t)
		})
	}
}

func TestDelete(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name string

		text string
		key  string

		expErr     string
		exp        string
		assertions func(t *testing.T, g *d2graph.Graph)
	}{
		{
			name: "flat",

			text: `nerve-gift-earther
`,
			key: `nerve-gift-earther`,

			exp: ``,
			assertions: func(t *testing.T, g *d2graph.Graph) {
				if len(g.Objects) != 0 {
					t.Fatalf("expected zero objects: %#v", g.Objects)
				}
			},
		},
		{
			name: "edge_identical_child",

			text: `x.x.y.z -> x.y.b
`,
			key: `x`,

			exp: `x.y.z -> y.b
`,
		},
		{
			name: "edge_both_identical_childs",

			text: `x.x.y.z -> x.x.b
`,
			key: `x`,

			exp: `x.y.z -> x.b
`,
		},
		{
			name: "edge_conflict",

			text: `x.y.a -> x.y.b
y
`,
			key: `x`,

			exp: `y 2.a -> y 2.b
y
`,
		},
		{
			name: "underscore_remove",

			text: `x: {
  _.y
  _.a -> _.b
  _.c -> d
}
`,
			key: `x`,

			exp: `y
a -> b
c -> d
`,
		},
		{
			name: "underscore_no_conflict",

			text: `x: {
	y: {
    _._.z
  }
  z
}
`,
			key: `x.y`,

			exp: `x: {
  _.z

  z
}
`,
		},
		{
			name: "nested_underscore_update",

			text: `guitar: {
	books: {
    _._.pipe
  }
}
`,
			key: `guitar`,

			exp: `books: {
  _.pipe
}
`,
		},
		{
			name: "node_in_edge",

			text: `x -> y -> z -> q -> p
z.ok: {
  what's up
}
`,
			key: `z`,

			exp: `x -> y
q -> p
ok: {
  what's up
}
`,
			assertions: func(t *testing.T, g *d2graph.Graph) {
				if len(g.Objects) != 6 {
					t.Fatalf("expected 6 objects: %#v", g.Objects)
				}
				if len(g.Edges) != 2 {
					t.Fatalf("expected two edges: %#v", g.Edges)
				}
			},
		},
		{
			name: "node_in_edge_last",

			text: `x -> y -> z -> q -> a.b.p
a.b.p: {
  what's up
}
`,
			key: `a.b.p`,

			exp: `x -> y -> z -> q
a.b: {
  what's up
}
`,
			assertions: func(t *testing.T, g *d2graph.Graph) {
				if len(g.Objects) != 7 {
					t.Fatalf("expected 7 objects: %#v", g.Objects)
				}
				if len(g.Edges) != 3 {
					t.Fatalf("expected three edges: %#v", g.Edges)
				}
			},
		},
		{
			name: "children",

			text: `p: {
  what's up
  x -> y
}
`,
			key: `p`,

			exp: `what's up
x -> y
`,
			assertions: func(t *testing.T, g *d2graph.Graph) {
				if len(g.Objects) != 3 {
					t.Fatalf("expected 3 objects: %#v", g.Objects)
				}
				if len(g.Edges) != 1 {
					t.Fatalf("expected 1 edge: %#v", g.Edges)
				}
			},
		},
		{
			name: "hoist_children",

			text: `a: {
  b: {
    c
  }
}
`,
			key: `a.b`,

			exp: `a: {
  c
}
`,
		},
		{
			name: "hoist_edge_children",

			text: `a: {
  b
  c -> d
}
`,
			key: `a`,

			exp: `b
c -> d
`,
		},
		{
			name: "children_conflicts",

			text: `p: {
  x
}
x
`,
			key: `p`,

			exp: `x 2

x
`,
		},
		{
			name: "edge_map_style",

			text: `x -> y: { style.stroke: red }
`,
			key: `(x -> y)[0].style.stroke`,

			exp: `x -> y
`,
		},
		{
			// Just checks that removing an object removes the arrowhead field too
			name: "breakup_arrowhead",

			text: `x -> y: {
  target-arrowhead.shape: diamond
}
(x -> y)[0].source-arrowhead: {
  shape: diamond
}
`,
			key: `x`,

			exp: `y
`,
		},
		{
			name: "edge_key_style",

			text: `x -> y
(x -> y)[0].style.stroke: red
`,
			key: `(x -> y)[0].style.stroke`,

			exp: `x -> y
`,
		},
		{
			name: "nested_edge_key_style",

			text: `a: {
  x -> y
}
a.(x -> y)[0].style.stroke: red
`,
			key: `a.(x -> y)[0].style.stroke`,

			exp: `a: {
  x -> y
}
`,
		},
		{
			name: "multiple_flat_style",

			text: `x.style.opacity: 0.4
x.style.fill: red
`,
			key: `x.style.fill`,

			exp: `x.style.opacity: 0.4
`,
		},
		{
			name: "edge_flat_style",

			text: `A -> B
A.style.stroke-dash: 5
`,
			key: `A`,

			exp: `B
`,
		},
		{
			name: "flat_reserved",

			text: `A -> B
A.style.stroke-dash: 5
`,
			key: `A.style.stroke-dash`,

			exp: `A -> B
`,
		},
		{
			name: "singular_flat_style",

			text: `x.style.fill: red
`,
			key: `x.style.fill`,

			exp: `x
`,
		},
		{
			name: "nested_flat_style",

			text: `x: {
	style.fill: red
}
`,
			key: `x.style.fill`,

			exp: `x
`,
		},
		{
			name: "multiple_map_styles",

			text: `x: {
  style: {
    opacity: 0.4
    fill: red
  }
}
`,
			key: `x.style.fill`,

			exp: `x: {
  style: {
    opacity: 0.4
  }
}
`,
		},
		{
			name: "singular_map_style",

			text: `x: {
  style: {
    fill: red
  }
}
`,
			key: `x.style.fill`,

			exp: `x
`,
		},
		{
			name: "delete_near",

			text: `x: {
	near: y
}
y
`,
			key: `x.near`,

			exp: `x
y
`,
		},
		{
			name: "delete_tooltip",

			text: `x: {
	tooltip: yeah
}
`,
			key: `x.tooltip`,

			exp: `x
`,
		},
		{
			name: "delete_link",

			text: `x.link: https://google.com
`,
			key: `x.link`,

			exp: `x
`,
		},
		{
			name: "delete_icon",

			text: `y.x: {
  link: https://google.com
	icon: https://google.com/memes.jpeg
}
`,
			key: `y.x.icon`,

			exp: `y.x: {
  link: https://google.com
}
`,
		},
		{
			name: "delete_redundant_flat_near",

			text: `x

y
`,
			key: `x.near`,

			exp: `x

y
`,
		},
		{
			name: "delete_needed_flat_near",

			text: `x.near: y
y
`,
			key: `x.near`,

			exp: `x
y
`,
		},
		{
			name: "children_no_self_conflict",

			text: `x: {
  x
}
`,
			key: `x`,

			exp: `x
`,
		},
		{
			name: "near",

			text: `x: {
  near: y
}
y
`,
			key: `y`,

			exp: `x
`,
		},
		{
			name: "container_near",

			text: `x: {
  y: {
    near: x.z
  }
  z
	a: {
	  near: x.z
  }
}
`,
			key: `x`,

			exp: `y: {
  near: z
}
z
a: {
  near: z
}
`,
		},
		{
			name: "multi_near",

			text: `Starfish: {
  API
  Bluefish: {
    near: Starfish.API
  }
	Yo: {
    near: Blah
  }
}
Blah
`,
			key: `Starfish`,

			exp: `API
Bluefish: {
  near: API
}
Yo: {
  near: Blah
}

Blah
`,
		},
		{
			name: "children_nested_conflicts",

			text: `p: {
	x: {
    y
  }
}
x
`,
			key: `p`,

			exp: `x 2: {
  y
}

x
`,
		},
		{
			name: "children_referenced_conflicts",

			text: `p: {
	x
}
x

p.x: "hi"
`,
			key: `p`,

			exp: `x 2

x

x 2: "hi"
`,
		},
		{
			name: "children_flat_conflicts",

			text: `p.x
x

p.x: "hi"
`,
			key: `p`,

			exp: `x 2
x

x 2: "hi"
`,
		},
		{
			name: "children_edges_flat_conflicts",

			text: `p.x -> p.y -> p.z
x
z

p.x: "hi"
p.z: "ey"
`,
			key: `p`,

			exp: `x 2 -> y -> z 2
x
z

x 2: "hi"
z 2: "ey"
`,
		},
		{
			name: "children_nested_referenced_conflicts",

			text: `p: {
	x.y
}
x

p.x: "hi"
p.x.y: "hey"
`,
			key: `p`,

			exp: `x 2.y

x

x 2: "hi"
x 2.y: "hey"
`,
		},
		{
			name: "children_edge_conflicts",

			text: `p: {
	x -> y
}
x

p.x: "hi"
`,
			key: `p`,

			exp: `x 2 -> y

x

x 2: "hi"
`,
		},
		{
			name: "children_multiple_conflicts",

			text: `p: {
	x -> y
	x
	y
}
x
y

p.x: "hi"
`,
			key: `p`,

			exp: `x 2 -> y 2
x 2
y 2

x
y

x 2: "hi"
`,
		},
		{
			name: "multi_path_map_conflict",

			text: `x.y: {
  z
}
x: {
  z
}
`,
			key: `x.y`,

			exp: `x: {
  z 2
}
x: {
  z
}
`,
		},
		{
			name: "multi_path_map_no_conflict",

			text: `x.y: {
  z
}
x: {
  z
}
`,
			key: `x`,

			exp: `y: {
  z
}

z
`,
		},
		{
			name: "children_scope",

			text: `x.q: {
  p: {
    what's up
    x -> y
  }
}
`,
			key: `x.q.p`,

			exp: `x.q: {
  what's up
  x -> y
}
`,
			assertions: func(t *testing.T, g *d2graph.Graph) {
				if len(g.Objects) != 5 {
					t.Fatalf("expected 5 objects: %#v", g.Objects)
				}
				if len(g.Edges) != 1 {
					t.Fatalf("expected 1 edge: %#v", g.Edges)
				}
			},
		},
		{
			name: "children_order",

			text: `c: {
  before
  y: {
    congo
  }
  after
}
`,
			key: `c.y`,

			exp: `c: {
  before

  congo

  after
}
`,
			assertions: func(t *testing.T, g *d2graph.Graph) {
				if len(g.Objects) != 4 {
					t.Fatalf("expected 4 objects: %#v", g.Objects)
				}
			},
		},
		{
			name: "edge_first",

			text: `l.p.d: {x -> p -> y -> z}
`,
			key: `l.p.d.(x -> p)[0]`,

			exp: `l.p.d: {x; p -> y -> z}
`,
			assertions: func(t *testing.T, g *d2graph.Graph) {
				if len(g.Objects) != 7 {
					t.Fatalf("expected 7 objects: %#v", g.Objects)
				}
				if len(g.Edges) != 2 {
					t.Fatalf("unexpected edges: %#v", g.Objects)
				}
			},
		},
		{
			name: "multiple_flat_middle_container",

			text: `a.b.c
a.b.d
`,
			key: `a.b`,

			exp: `a.c
a.d
`,
		},
		{
			name: "edge_middle",

			text: `l.p.d: {x -> y -> z -> q -> p}
`,
			key: `l.p.d.(z -> q)[0]`,

			exp: `l.p.d: {x -> y -> z; q -> p}
`,
			assertions: func(t *testing.T, g *d2graph.Graph) {
				if len(g.Objects) != 8 {
					t.Fatalf("expected 8 objects: %#v", g.Objects)
				}
				if len(g.Edges) != 3 {
					t.Fatalf("expected three edges: %#v", g.Edges)
				}
			},
		},
		{
			name: "edge_last",

			text: `l.p.d: {x -> y -> z -> q -> p}
`,
			key: `l.p.d.(q -> p)[0]`,

			exp: `l.p.d: {x -> y -> z -> q; p}
`,
			assertions: func(t *testing.T, g *d2graph.Graph) {
				if len(g.Objects) != 8 {
					t.Fatalf("expected 8 objects: %#v", g.Objects)
				}
				if len(g.Edges) != 3 {
					t.Fatalf("expected three edges: %#v", g.Edges)
				}
			},
		},
		{
			name: "key_with_edges",

			text: `hello.meow -> hello.bark
`,
			key: `hello.(meow -> bark)[0]`,

			exp: `hello.meow
hello.bark
`,
			assertions: func(t *testing.T, g *d2graph.Graph) {
				if len(g.Objects) != 3 {
					t.Fatalf("expected three objects: %#v", g.Objects)
				}
				if len(g.Edges) != 0 {
					t.Fatalf("expected zero edges: %#v", g.Edges)
				}
			},
		},
		{
			name: "key_with_edges_2",

			text: `hello.meow -> hello.bark
`,
			key: `hello.meow`,

			exp: `hello.bark
`,
			assertions: func(t *testing.T, g *d2graph.Graph) {
				if len(g.Objects) != 2 {
					t.Fatalf("expected 2 objects: %#v", g.Objects)
				}
			},
		},
		{
			name: "key_with_edges_3",

			text: `hello.(meow -> bark)
`,
			key: `hello.meow`,

			exp: `hello.bark
`,
			assertions: func(t *testing.T, g *d2graph.Graph) {
				if len(g.Objects) != 2 {
					t.Fatalf("expected 2 objects: %#v", g.Objects)
				}
			},
		},
		{
			name: "key_with_edges_4",

			text: `hello.(meow -> bark)
`,
			key: `(hello.meow -> hello.bark)[0]`,

			exp: `hello.meow
hello.bark
`,
			assertions: func(t *testing.T, g *d2graph.Graph) {
				if len(g.Objects) != 3 {
					t.Fatalf("expected three objects: %#v", g.Objects)
				}
				if len(g.Edges) != 0 {
					t.Fatalf("expected zero edges: %#v", g.Edges)
				}
			},
		},
		{
			name: "nested",

			text: `a.b.c.d
`,
			key: `a.b`,

			exp: `a.c.d
`,
			assertions: func(t *testing.T, g *d2graph.Graph) {
				if len(g.Objects) != 3 {
					t.Fatalf("expected 3 objects: %#v", g.Objects)
				}
			},
		},
		{
			name: "nested_2",

			text: `a.b.c.d
`,
			key: `a.b.c.d`,

			exp: `a.b.c
`,
			assertions: func(t *testing.T, g *d2graph.Graph) {
				if len(g.Objects) != 3 {
					t.Fatalf("expected 3 objects: %#v", g.Objects)
				}
			},
		},
		{
			name: "order_1",

			text: `x -> p -> y -> z
`,
			key: `p`,

			exp: `x
y -> z
`,
			assertions: func(t *testing.T, g *d2graph.Graph) {
				if len(g.Objects) != 3 {
					t.Fatalf("expected 3 objects: %#v", g.Objects)
				}
			},
		},
		{
			name: "order_2",

			text: `p -> y -> z
`,
			key: `y`,

			exp: `p
z
`,
			assertions: func(t *testing.T, g *d2graph.Graph) {
				if len(g.Objects) != 2 {
					t.Fatalf("expected 2 objects: %#v", g.Objects)
				}
			},
		},
		{
			name: "order_3",

			text: `y -> p -> y -> z
`,
			key: `y`,

			exp: `p
z
`,
			assertions: func(t *testing.T, g *d2graph.Graph) {
				if len(g.Objects) != 2 {
					t.Fatalf("expected 2 objects: %#v", g.Objects)
				}
			},
		},
		{
			name: "order_4",

			text: `y -> p
`,
			key: `p`,

			exp: `y
`,
			assertions: func(t *testing.T, g *d2graph.Graph) {
				if len(g.Objects) != 1 {
					t.Fatalf("expected 1 object: %#v", g.Objects)
				}
			},
		},
		{
			name: "order_5",

			text: `x: {
  a -> b -> c
  q -> p
}
`,
			key: `x.a`,

			exp: `x: {
  b -> c
  q -> p
}
`,
			assertions: func(t *testing.T, g *d2graph.Graph) {
				if len(g.Objects) != 5 {
					t.Fatalf("expected 5 objects: %#v", g.Objects)
				}
			},
		},
		{
			name: "order_6",

			text: `x: {
  lol
}
x.p.q.z
`,
			key: `x.p.q.z`,

			exp: `x: {
  lol
}
x.p.q
`,
			assertions: func(t *testing.T, g *d2graph.Graph) {
				if len(g.Objects) != 4 {
					t.Fatalf("expected 4 objects: %#v", g.Objects)
				}
			},
		},
		{
			name: "order_7",

			text: `x: {
  lol
}
x.p.q.more
x.p.q.z
`,
			key: `x.p.q.z`,

			exp: `x: {
  lol
}
x.p.q.more
`,
			assertions: func(t *testing.T, g *d2graph.Graph) {
				if len(g.Objects) != 5 {
					t.Fatalf("expected 5 objects: %#v", g.Objects)
				}
			},
		},
		{
			name: "order_8",

			text: `x -> y
bark
y -> x
zebra
x -> q
kang
`,
			key: `x`,

			exp: `bark
y

zebra
q

kang
`,
			assertions: func(t *testing.T, g *d2graph.Graph) {
				if len(g.Objects) != 5 {
					t.Fatalf("expected 5 objects: %#v", g.Objects)
				}
			},
		},
		{
			name: "empty_map",

			text: `c: {
  y: {
    congo
  }
}
`,
			key: `c.y.congo`,

			exp: `c: {
  y
}
`,
			assertions: func(t *testing.T, g *d2graph.Graph) {
				if len(g.Objects) != 2 {
					t.Fatalf("expected 2 objects: %#v", g.Objects)
				}
			},
		},
		{
			name: "edge_common",

			text: `x.a -> x.y
`,
			key: "x",

			exp: `a -> y
`,
			assertions: func(t *testing.T, g *d2graph.Graph) {
				if len(g.Objects) != 2 {
					t.Fatalf("expected 2 objects: %#v", g.Objects)
				}
				if len(g.Edges) != 1 {
					t.Fatalf("unexpected edges: %#v", g.Edges)
				}
			},
		},
		{
			name: "edge_common_2",

			text: `x.(a -> y)
`,
			key: "x",

			exp: `a -> y
`,
			assertions: func(t *testing.T, g *d2graph.Graph) {
				if len(g.Objects) != 2 {
					t.Fatalf("expected 2 objects: %#v", g.Objects)
				}
				if len(g.Edges) != 1 {
					t.Fatalf("unexpected edges: %#v", g.Edges)
				}
			},
		},
		{
			name: "edge_common_3",

			text: `x.(a -> y)
`,
			key: "(x.a -> x.y)[0]",

			exp: `x.a
x.y
`,
			assertions: func(t *testing.T, g *d2graph.Graph) {
				if len(g.Objects) != 3 {
					t.Fatalf("expected 3 objects: %#v", g.Objects)
				}
				if len(g.Edges) != 0 {
					t.Fatalf("unexpected edges: %#v", g.Edges)
				}
			},
		},
		{
			name: "edge_common_4",

			text: `x.a -> x.y
`,
			key: "x.(a -> y)[0]",

			exp: `x.a
x.y
`,
			assertions: func(t *testing.T, g *d2graph.Graph) {
				if len(g.Objects) != 3 {
					t.Fatalf("expected 3 objects: %#v", g.Objects)
				}
				if len(g.Edges) != 0 {
					t.Fatalf("unexpected edges: %#v", g.Edges)
				}
			},
		},
		{
			name: "edge_decrement",

			text: `a -> b
a -> b
a -> b
a -> b
a -> b
(a -> b)[0]: zero
(a -> b)[1]: one
(a -> b)[2]: two
(a -> b)[3]: three
(a -> b)[4]: four
`,
			key: `(a -> b)[2]`,

			exp: `a -> b
a -> b

a -> b
a -> b
(a -> b)[0]: zero
(a -> b)[1]: one

(a -> b)[2]: three
(a -> b)[3]: four
`,
		},
		{
			name: "shape_class",
			text: `D2 Parser: {
  shape: class

  # Default visibility is + so no need to specify.
  +reader: io.RuneReader
  readerPos: d2ast.Position

  # Private field.
  -lookahead: "[]rune"

  # Protected field.
  # We have to escape the # to prevent the line from being parsed as a comment.
  \#lookaheadPos: d2ast.Position

  +peek(): (r rune, eof bool)
  rewind()
  commit()

  \#peekn(n int): (s string, eof bool)
}

"github.com/terrastruct/d2parser.git" -> D2 Parser
`,
			key: `D2 Parser`,

			exp: `"github.com/terrastruct/d2parser.git"
`,
		},
		// TODO: delete disks.id as it's redundant
		{
			name: "shape_sql_table",

			text: `cloud: {
  disks: {
    shape: sql_table
    id: int {constraint: primary_key}
  }
  blocks: {
    shape: sql_table
    id: int {constraint: primary_key}
    disk: int {constraint: foreign_key}
    blob: blob
  }
  blocks.disk -> disks.id

  AWS S3 Vancouver -> disks
}
`,
			key: "cloud.blocks",

			exp: `cloud: {
  disks: {
    shape: sql_table
    id: int {constraint: primary_key}
  }

  disks.id

  AWS S3 Vancouver -> disks
}
`,
		},
		{
			name: "nested_reserved",

			text: `x.y.z: {
  label: Sweet April showers do spring May flowers.
  icon: bingo
	near: x.y.jingle
  shape: parallelogram
  style: {
    stroke: red
  }
}
x.y.jingle
`,
			key: "x.y.z",

			exp: `x.y
x.y.jingle
`,
		},
		{
			name: "only_delete_obj_reserved",

			text: `A: {style.stroke: "#000e3d"}
B
A -> B: {style.stroke: "#2b50c2"}
`,
			key: `A.style.stroke`,
			exp: `A
B
A -> B: {style.stroke: "#2b50c2"}
`,
		},
		{
			name: "only_delete_edge_reserved",

			text: `A: {style.stroke: "#000e3d"}
B
A -> B: {style.stroke: "#2b50c2"}
`,
			key: `(A->B)[0].style.stroke`,
			exp: `A: {style.stroke: "#000e3d"}
B
A -> B
`,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			et := editTest{
				text: tc.text,
				testFunc: func(g *d2graph.Graph) (*d2graph.Graph, error) {
					return d2oracle.Delete(g, tc.key)
				},

				exp:        tc.exp,
				expErr:     tc.expErr,
				assertions: tc.assertions,
			}
			et.run(t)
		})
	}
}

type editTest struct {
	text     string
	testFunc func(*d2graph.Graph) (*d2graph.Graph, error)

	exp        string
	expErr     string
	assertions func(*testing.T, *d2graph.Graph)
}

func (tc editTest) run(t *testing.T) {
	d2Path := fmt.Sprintf("d2/testdata/d2oracle/%v.d2", t.Name())
	g, err := d2compiler.Compile(d2Path, strings.NewReader(tc.text), nil)
	if err != nil {
		t.Fatal(err)
	}

	g, err = tc.testFunc(g)
	if tc.expErr != "" {
		if err == nil {
			t.Fatalf("expected error with: %q", tc.expErr)
		}
		ds, err := diff.Strings(tc.expErr, err.Error())
		if err != nil {
			t.Fatal(err)
		}
		if ds != "" {
			t.Fatalf("unexpected error: %s", ds)
		}
	} else if err != nil {
		t.Fatal(err)
	}

	if tc.expErr == "" {
		if tc.assertions != nil {
			t.Run("assertions", func(t *testing.T) {
				tc.assertions(t, g)
			})
		}

		newText := d2format.Format(g.AST)
		ds, err := diff.Strings(tc.exp, newText)
		if err != nil {
			t.Fatal(err)
		}
		if ds != "" {
			t.Fatalf("tc.exp != newText:\n%s", ds)
		}
	}

	got := struct {
		Graph *d2graph.Graph `json:"graph"`
		Err   string         `json:"err"`
	}{
		Graph: g,
		Err:   fmt.Sprintf("%#v", err),
	}

	err = diff.Testdata(filepath.Join("..", "testdata", "d2oracle", t.Name()), got)
	if err != nil {
		t.Fatal(err)
	}
}

func TestMoveIDDeltas(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name string

		text   string
		key    string
		newKey string

		exp    string
		expErr string
	}{
		{
			name: "rename",

			text: `x
`,
			key:    "x",
			newKey: "y",

			exp: `{
  "x": "y"
}`,
		},
		{
			name: "rename_identical",

			text: `Square
`,
			key:    "Square",
			newKey: "Square",

			exp: `{}`,
		},
		{
			name: "children_no_self_conflict",

			text: `x: {
  x
}
y
`,
			key:    `x`,
			newKey: `y.x`,

			exp: `{
  "x": "y.x",
  "x.x": "x"
}`,
		},
		{
			name: "into_container",

			text: `x
y
x -> z
`,
			key:    "x",
			newKey: "y.x",

			exp: `{
  "(x -> z)[0]": "(y.x -> z)[0]",
  "x": "y.x"
}`,
		},
		{
			name: "out_container",

			text: `x: {
  y
}
x.y -> z
`,
			key:    "x.y",
			newKey: "y",

			exp: `{
  "(x.y -> z)[0]": "(y -> z)[0]",
  "x.y": "y"
}`,
		},
		{
			name: "container_with_edge",

			text: `x {
  a
  b
  a -> b
}
y
`,
			key:    "x",
			newKey: "y.x",

			exp: `{
  "x": "y.x",
  "x.(a -> b)[0]": "(a -> b)[0]",
  "x.a": "a",
  "x.b": "b"
}`,
		},
		{
			name: "out_conflict",

			text: `x: {
  y
}
y
x.y -> z
`,
			key:    "x.y",
			newKey: "y",

			exp: `{
  "(x.y -> z)[0]": "(y 2 -> z)[0]",
  "x.y": "y 2"
}`,
		},
		{
			name: "into_conflict",

			text: `x: {
  y
}
y
x.y -> z
`,
			key:    "y",
			newKey: "x.y",

			exp: `{
  "y": "x.y 2"
}`,
		},
		{
			name: "move_container",

			text: `x: {
  a
  b
}
y
x.a -> x.b
x.a -> x.b
`,
			key:    "x",
			newKey: "y.x",

			exp: `{
  "x": "y.x",
  "x.(a -> b)[0]": "(a -> b)[0]",
  "x.(a -> b)[1]": "(a -> b)[1]",
  "x.a": "a",
  "x.b": "b"
}`,
		},
		{
			name: "conflicts",

			text: `x: {
  a
  b
}
a
y
x.a -> x.b
`,
			key:    "x",
			newKey: "y.x",

			exp: `{
  "x": "y.x",
  "x.(a -> b)[0]": "(a 2 -> b)[0]",
  "x.a": "a 2",
  "x.b": "b"
}`,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			d2Path := fmt.Sprintf("d2/testdata/d2oracle/%v.d2", t.Name())
			g, err := d2compiler.Compile(d2Path, strings.NewReader(tc.text), nil)
			if err != nil {
				t.Fatal(err)
			}

			deltas, err := d2oracle.MoveIDDeltas(g, tc.key, tc.newKey)
			if tc.expErr != "" {
				if err == nil {
					t.Fatalf("expected error with: %q", tc.expErr)
				}
				ds, err := diff.Strings(tc.expErr, err.Error())
				if err != nil {
					t.Fatal(err)
				}
				if ds != "" {
					t.Fatalf("unexpected error: %s", ds)
				}
			} else if err != nil {
				t.Fatal(err)
			}

			ds, err := diff.Strings(tc.exp, xjson.MarshalIndent(deltas))
			if err != nil {
				t.Fatal(err)
			}
			if ds != "" {
				t.Fatalf("unexpected deltas: %s", ds)
			}
		})
	}
}

func TestDeleteIDDeltas(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name string

		text string
		key  string

		exp    string
		expErr string
	}{
		{
			name: "delete_node",

			text: `x.y.p -> x.y.q
x.y.z.w.e.p.l
x.y.z.1.2.3.4
x.y.3.4.5.6
x.y.3.4.6.7
x.y.3.4.6.7 -> x.y.3.4.5.6
x.y.z.w.e.p.l -> x.y.z.1.2.3.4
`,
			key: "x.y",

			exp: `{
  "x.y.(p -> q)[0]": "x.(p -> q)[0]",
  "x.y.3": "x.3",
  "x.y.3.4": "x.3.4",
  "x.y.3.4.(6.7 -> 5.6)[0]": "x.3.4.(6.7 -> 5.6)[0]",
  "x.y.3.4.5": "x.3.4.5",
  "x.y.3.4.5.6": "x.3.4.5.6",
  "x.y.3.4.6": "x.3.4.6",
  "x.y.3.4.6.7": "x.3.4.6.7",
  "x.y.p": "x.p",
  "x.y.q": "x.q",
  "x.y.z": "x.z",
  "x.y.z.(w.e.p.l -> 1.2.3.4)[0]": "x.z.(w.e.p.l -> 1.2.3.4)[0]",
  "x.y.z.1": "x.z.1",
  "x.y.z.1.2": "x.z.1.2",
  "x.y.z.1.2.3": "x.z.1.2.3",
  "x.y.z.1.2.3.4": "x.z.1.2.3.4",
  "x.y.z.w": "x.z.w",
  "x.y.z.w.e": "x.z.w.e",
  "x.y.z.w.e.p": "x.z.w.e.p",
  "x.y.z.w.e.p.l": "x.z.w.e.p.l"
}`,
		},
		{
			name: "children_no_self_conflict",

			text: `x: {
  x
}
`,
			key: `x`,

			exp: `{
  "x.x": "x"
}`,
		},
		{
			name: "delete_container_with_conflicts",

			text: `x {
  a
  b
}
a
b
c
x.a -> c
`,
			key: "x",

			exp: `{
  "(x.a -> c)[0]": "(a 2 -> c)[0]",
  "x.a": "a 2",
  "x.b": "b 2"
}`,
		},
		{
			name: "multiword",

			text: `Starfish: {
  API
}
Starfish.API
`,
			key: "Starfish",

			exp: `{
  "Starfish.API": "API"
}`,
		},
		{
			name: "delete_container_with_edge",

			text: `x {
  a
  b
  a -> b
}
`,
			key: "x",

			exp: `{
  "x.(a -> b)[0]": "(a -> b)[0]",
  "x.a": "a",
  "x.b": "b"
}`,
		},
		{
			name: "delete_edge_field",

			text: `a -> b
a -> b
`,
			key: "(a -> b)[0].style.opacity",

			exp: "null",
		},
		{
			name: "delete_edge",

			text: `x.y.z.w.e.p.l -> x.y.z.1.2.3.4
x.y.z.w.e.p.l -> x.y.z.1.2.3.4
x.y.z.w.e.p.l -> x.y.z.1.2.3.4
x.y.z.w.e.p.l -> x.y.z.1.2.3.4
x.y.z.w.e.p.l -> x.y.z.1.2.3.4
x.y.z.w.e.p.l -> x.y.z.1.2.3.4
x.y.z.w.e.p.l -> x.y.z.1.2.3.4
(x.y.z.w.e.p.l -> x.y.z.1.2.3.4)[0]: meow
(x.y.z.w.e.p.l -> x.y.z.1.2.3.4)[1]: meow
(x.y.z.w.e.p.l -> x.y.z.1.2.3.4)[2]: meow
(x.y.z.w.e.p.l -> x.y.z.1.2.3.4)[3]: meow
(x.y.z.w.e.p.l -> x.y.z.1.2.3.4)[4]: meow
(x.y.z.w.e.p.l -> x.y.z.1.2.3.4)[5]: meow
(x.y.z.w.e.p.l -> x.y.z.1.2.3.4)[6]: meow
`,
			key: "(x.y.z.w.e.p.l -> x.y.z.1.2.3.4)[1]",

			exp: `{
  "x.y.z.(w.e.p.l -> 1.2.3.4)[2]": "x.y.z.(w.e.p.l -> 1.2.3.4)[1]",
  "x.y.z.(w.e.p.l -> 1.2.3.4)[3]": "x.y.z.(w.e.p.l -> 1.2.3.4)[2]",
  "x.y.z.(w.e.p.l -> 1.2.3.4)[4]": "x.y.z.(w.e.p.l -> 1.2.3.4)[3]",
  "x.y.z.(w.e.p.l -> 1.2.3.4)[5]": "x.y.z.(w.e.p.l -> 1.2.3.4)[4]",
  "x.y.z.(w.e.p.l -> 1.2.3.4)[6]": "x.y.z.(w.e.p.l -> 1.2.3.4)[5]"
}`,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			d2Path := fmt.Sprintf("d2/testdata/d2oracle/%v.d2", t.Name())
			g, err := d2compiler.Compile(d2Path, strings.NewReader(tc.text), nil)
			if err != nil {
				t.Fatal(err)
			}

			deltas, err := d2oracle.DeleteIDDeltas(g, tc.key)
			if tc.expErr != "" {
				if err == nil {
					t.Fatalf("expected error with: %q", tc.expErr)
				}
				ds, err := diff.Strings(tc.expErr, err.Error())
				if err != nil {
					t.Fatal(err)
				}
				if ds != "" {
					t.Fatalf("unexpected error: %s", ds)
				}
			} else if err != nil {
				t.Fatal(err)
			}

			ds, err := diff.Strings(tc.exp, xjson.MarshalIndent(deltas))
			if err != nil {
				t.Fatal(err)
			}
			if ds != "" {
				t.Fatalf("unexpected deltas: %s", ds)
			}
		})
	}
}

func TestRenameIDDeltas(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name string

		text    string
		key     string
		newName string

		exp    string
		expErr string
	}{
		{
			name: "rename_node",

			text: `x.y.p -> x.y.q
x.y.z.w.e.p.l
x.y.z.1.2.3.4
x.y.3.4.5.6
x.y.3.4.6.7
x.y.3.4.6.7 -> x.y.3.4.5.6
x.y.z.w.e.p.l -> x.y.z.1.2.3.4
`,
			key:     "x.y",
			newName: "papa",

			exp: `{
  "x.y": "x.papa",
  "x.y.(p -> q)[0]": "x.papa.(p -> q)[0]",
  "x.y.3": "x.papa.3",
  "x.y.3.4": "x.papa.3.4",
  "x.y.3.4.(6.7 -> 5.6)[0]": "x.papa.3.4.(6.7 -> 5.6)[0]",
  "x.y.3.4.5": "x.papa.3.4.5",
  "x.y.3.4.5.6": "x.papa.3.4.5.6",
  "x.y.3.4.6": "x.papa.3.4.6",
  "x.y.3.4.6.7": "x.papa.3.4.6.7",
  "x.y.p": "x.papa.p",
  "x.y.q": "x.papa.q",
  "x.y.z": "x.papa.z",
  "x.y.z.(w.e.p.l -> 1.2.3.4)[0]": "x.papa.z.(w.e.p.l -> 1.2.3.4)[0]",
  "x.y.z.1": "x.papa.z.1",
  "x.y.z.1.2": "x.papa.z.1.2",
  "x.y.z.1.2.3": "x.papa.z.1.2.3",
  "x.y.z.1.2.3.4": "x.papa.z.1.2.3.4",
  "x.y.z.w": "x.papa.z.w",
  "x.y.z.w.e": "x.papa.z.w.e",
  "x.y.z.w.e.p": "x.papa.z.w.e.p",
  "x.y.z.w.e.p.l": "x.papa.z.w.e.p.l"
}`,
		},
		{
			name: "rename_conflict",

			text: `x
y
`,
			key:     "x",
			newName: "y",

			exp: `{
  "x": "y 2"
}`,
		},
		{
			name: "rename_conflict_with_dots",

			text: `"a.b"
y
`,
			key:     "y",
			newName: "a.b",

			exp: `{
  "y": "\"a.b 2\""
}`,
		},
		{
			name: "rename_identical",

			text: `Square
`,
			key:     "Square",
			newName: "Square",

			exp: `{}`,
		},
		{
			name: "rename_edge",

			text: `x.y.z.w.e.p.l -> x.y.z.1.2.3.4
x.y.z.w.e.p.l -> x.y.z.1.2.3.4
x.y.z.w.e.p.l -> x.y.z.1.2.3.4
x.y.z.w.e.p.l -> x.y.z.1.2.3.4
x.y.z.w.e.p.l -> x.y.z.1.2.3.4
x.y.z.w.e.p.l -> x.y.z.1.2.3.4
x.y.z.w.e.p.l -> x.y.z.1.2.3.4
(x.y.z.w.e.p.l -> x.y.z.1.2.3.4)[0]: meow
(x.y.z.w.e.p.l -> x.y.z.1.2.3.4)[1]: meow
(x.y.z.w.e.p.l -> x.y.z.1.2.3.4)[2]: meow
(x.y.z.w.e.p.l -> x.y.z.1.2.3.4)[3]: meow
(x.y.z.w.e.p.l -> x.y.z.1.2.3.4)[4]: meow
(x.y.z.w.e.p.l -> x.y.z.1.2.3.4)[5]: meow
(x.y.z.w.e.p.l -> x.y.z.1.2.3.4)[6]: meow
`,
			key:     "(x.y.z.w.e.p.l -> x.y.z.1.2.3.4)[1]",
			newName: "(x.y.z.w.e.p.l <-> x.y.z.1.2.3.4)[1]",

			exp: `{
  "x.y.z.(w.e.p.l -> 1.2.3.4)[1]": "x.y.z.(w.e.p.l <-> 1.2.3.4)[1]"
}`,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			d2Path := fmt.Sprintf("d2/testdata/d2oracle/%v.d2", t.Name())
			g, err := d2compiler.Compile(d2Path, strings.NewReader(tc.text), nil)
			if err != nil {
				t.Fatal(err)
			}

			deltas, err := d2oracle.RenameIDDeltas(g, tc.key, tc.newName)
			if tc.expErr != "" {
				if err == nil {
					t.Fatalf("expected error with: %q", tc.expErr)
				}
				ds, err := diff.Strings(tc.expErr, err.Error())
				if err != nil {
					t.Fatal(err)
				}
				if ds != "" {
					t.Fatalf("unexpected error: %s", ds)
				}
			} else if err != nil {
				t.Fatal(err)
			}

			ds, err := diff.Strings(tc.exp, xjson.MarshalIndent(deltas))
			if err != nil {
				t.Fatal(err)
			}
			if ds != "" {
				t.Fatalf("unexpected deltas: %s", ds)
			}
		})
	}
}
