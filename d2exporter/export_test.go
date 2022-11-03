package d2exporter_test

import (
	"context"
	"path/filepath"
	"strings"
	"testing"

	"cdr.dev/slog"

	"oss.terrastruct.com/diff"

	"github.com/stretchr/testify/assert"

	"oss.terrastruct.com/d2/d2compiler"
	"oss.terrastruct.com/d2/d2exporter"
	"oss.terrastruct.com/d2/d2layouts/d2dagrelayout"
	"oss.terrastruct.com/d2/d2renderers/textmeasure"
	"oss.terrastruct.com/d2/d2target"
	"oss.terrastruct.com/d2/d2themes/d2themescatalog"
	"oss.terrastruct.com/d2/lib/geo"
	"oss.terrastruct.com/d2/lib/log"
)

type testCase struct {
	name    string
	dsl     string
	themeID int64

	assertions func(t *testing.T, d *d2target.Diagram)
}

func TestExport(t *testing.T) {
	t.Parallel()

	t.Run("shape", testShape)
	t.Run("connection", testConnection)
	t.Run("label", testLabel)
	t.Run("theme", testTheme)
}

func testShape(t *testing.T) {
	tcs := []testCase{
		{
			name: "basic",
			dsl: `x
`,
		},
		{
			name: "synonyms",
			dsl: `x: {shape: circle}
y: {shape: square}
`,
		},
		{
			name: "text_color",
			dsl: `x: |md yo | { style.font-color: red }
`,
		},
		{
			name: "border-radius",
			dsl: `Square: "" { style.border-radius: 5 }
`,
		},
		{
			name: "image_dimensions",

			dsl: `hey: "" {
  icon: https://icons.terrastruct.com/essentials/004-picture.svg
  shape: image
	width: 200
	height: 230
}
`,
			assertions: func(t *testing.T, d *d2target.Diagram) {
				if d.Shapes[0].Width != 200 {
					t.Fatalf("expected width 200, got %v", d.Shapes[0].Width)
				}
				if d.Shapes[0].Height != 230 {
					t.Fatalf("expected height 230, got %v", d.Shapes[0].Height)
				}
			},
		},
	}

	runa(t, tcs)
}

func testConnection(t *testing.T) {
	tcs := []testCase{
		{
			name: "basic",
			dsl: `x -> y
`,
		},
		{
			name: "stroke-dash",
			dsl: `x -> y: { style.stroke-dash: 4 }
`,
		},
		{
			name: "arrowhead",
			dsl: `x -> y: {
  source-arrowhead: If you've done six impossible things before breakfast, why not round it
  target-arrowhead: {
    label: A man with one watch knows what time it is.
    shape: diamond
    style.filled: true
  }
}
`,
		},
		{
			// This is a regression test where a connection with stroke-dash of 0 on terrastruct flagship theme would have a diff color
			// than a connection without stroke dash
			themeID: d2themescatalog.FlagshipTerrastruct.ID,
			name:    "theme_stroke-dash",
			dsl: `x -> y: { style.stroke-dash: 0 }
x -> y
`,
		},
	}

	runa(t, tcs)
}

func testLabel(t *testing.T) {
	tcs := []testCase{
		{
			name: "basic_shape",
			dsl: `x: yo
`,
		},
		{
			name: "shape_font_color",
			dsl: `x: yo { style.font-color: red }
`,
		},
		{
			name: "connection_font_color",
			dsl: `x -> y: yo { style.font-color: red }
`,
		},
	}

	runa(t, tcs)
}

func testTheme(t *testing.T) {
	tcs := []testCase{
		{
			name:    "shape_without_bold",
			themeID: d2themescatalog.FlagshipTerrastruct.ID,
			dsl: `x: {
	style.bold: false
}
`,
		},
		{
			name:    "shape_with_italic",
			themeID: d2themescatalog.FlagshipTerrastruct.ID,
			dsl: `x: {
	style.italic: true
}
`,
		},
		{
			name: "connection_without_italic",
			dsl: `x -> y: asdf { style.italic: false }
`,
			themeID: d2themescatalog.FlagshipTerrastruct.ID,
		},
		{
			name: "connection_with_italic",
			dsl: `x -> y: asdf {
  style.italic: true
}
`,
			themeID: d2themescatalog.FlagshipTerrastruct.ID,
		},
		{
			name: "connection_with_bold",
			dsl: `x -> y: asdf {
  style.bold: true
}
`,
			themeID: d2themescatalog.FlagshipTerrastruct.ID,
		},
	}

	runa(t, tcs)
}

func runa(t *testing.T, tcs []testCase) {
	for _, tc := range tcs {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			run(t, tc)
		})
	}
}

func run(t *testing.T, tc testCase) {
	ctx := context.Background()
	ctx = log.WithTB(ctx, t, nil)
	ctx = log.Leveled(ctx, slog.LevelDebug)

	g, err := d2compiler.Compile("", strings.NewReader(tc.dsl), &d2compiler.CompileOptions{
		UTF16: true,
	})
	if err != nil {
		t.Fatal(err)
	}

	ruler, err := textmeasure.NewRuler()
	assert.Nil(t, err)

	err = g.SetDimensions(nil, ruler)
	assert.Nil(t, err)

	err = d2dagrelayout.Layout(ctx, g)
	if err != nil {
		t.Fatal(err)
	}

	got, err := d2exporter.Export(ctx, g, tc.themeID)
	if err != nil {
		t.Fatal(err)
	}

	if tc.assertions != nil {
		t.Run("assertions", func(t *testing.T) {
			tc.assertions(t, got)
		})
	}

	// This test is agnostic of layout changes
	for i := range got.Shapes {
		got.Shapes[i].Pos = d2target.Point{}
		got.Shapes[i].Width = 0
		got.Shapes[i].Height = 0
		got.Shapes[i].LabelWidth = 0
		got.Shapes[i].LabelHeight = 0
		got.Shapes[i].LabelPosition = ""
	}
	for i := range got.Connections {
		got.Connections[i].Route = []*geo.Point{}
		got.Connections[i].LabelWidth = 0
		got.Connections[i].LabelHeight = 0
		got.Connections[i].LabelPosition = ""
	}

	err = diff.Testdata(filepath.Join("..", "testdata", "d2exporter", t.Name()), got)
	if err != nil {
		t.Fatal(err)
	}
}
