<div align="center">
  <h1>
    <img src="./docs/assets/logo.svg" alt="D2" />
  </h1>
  <p>A modern DSL that turns text into diagrams.</p>

[Language docs](https://d2lang.com) | [Cheat sheet](./docs/assets/cheat_sheet.pdf)

[![ci](https://github.com/terrastruct/d2/actions/workflows/ci.yml/badge.svg)](https://github.com/terrastruct/d2/actions/workflows/ci.yml)
[![release](https://img.shields.io/github/v/release/terrastruct/d2)](https://github.com/terrastruct/d2/releases)
[![discord](https://img.shields.io/discord/1039184639652265985?label=discord)](https://discord.gg/NF6X8K4eDq)
[![twitter](https://img.shields.io/twitter/follow/terrastruct?style=social)](https://twitter.com/terrastruct)
[![license](https://img.shields.io/github/license/terrastruct/d2?color=9cf)](./LICENSE.txt)

<img src="./docs/assets/cli.gif" alt="D2 CLI" />

</div>

# Table of Contents

<!-- toc -->

- [Quickstart (CLI)](#quickstart-cli)
  * [MacOS](#macos)
  * [Linux/Windows](#linuxwindows)
- [Quickstart (library)](#quickstart-library)
- [Themes](#themes)
- [Fonts](#fonts)
- [Export file types](#export-file-types)
- [Language tooling](#language-tooling)
- [Layout engine](#layout-engine)
- [Comparison](#comparison)
- [Contributing](#contributing)
- [License](#license)
- [Dependencies](#dependencies)
- [Related](#related)
  * [VSCode extension](#vscode-extension)
  * [Vim extension](#vim-extension)
  * [Misc](#misc)

<!-- tocstop -->

## Quickstart (CLI)

The most convenient way to use D2 is to just run it as a CLI executable to
produce SVGs from `.d2` files.

```sh
go install oss.terrastruct.com/d2

echo 'x -> y -> z' > in.d2
d2 --watch in.d2 out.svg
```

A browser window will open with `out.svg` and live-reload on changes to `in.d2`.

### MacOS

Homebrew package coming soon.

### Linux/Windows

We have precompiled binaries on the [releases](https://github.com/terrastruct/d2/releases)
page. D2 will be added to OS-respective package managers soon.


## Quickstart (library)

In addition to being a runnable CLI tool, D2 can also be used to produce diagrams from
Go programs.

```go
import (
	"github.com/terrastruct/d2/d2compiler"
	"github.com/terrastruct/d2/d2exporter"
	"github.com/terrastruct/d2/d2layouts/d2dagrelayout"
	"github.com/terrastruct/d2/d2renderers/textmeasure"
	"github.com/terrastruct/d2/d2themes/d2themescatalog"
)

func main() {
  graph, err := d2compiler.Compile("", strings.NewReader("x -> y"), &d2compiler.CompileOptions{ UTF16: true })
  ruler, err := textmeasure.NewRuler()
  err = graph.SetDimensions(nil, ruler)
  err = d2dagrelayout.Layout(ctx, graph)
  diagram, err := d2exporter.Export(ctx, graph, d2themescatalog.NeutralDefault)
  ioutil.WriteFile(filepath.Join("out.svg"), d2svg.Render(*diagram), 0600)
}
```

D2 is built to be hackable -- the language has an API built on top of it to make edits
programmatically.

```go
import (
  "github.com/terrastruct/d2/d2oracle"
  "github.com/terrastruct/d2/d2format"
)

// ...modifying the diagram `x -> y` from above
// Create a shape with the ID, "meow"
graph, err = d2oracle.Create(graph, "meow")
// Style the shape green
graph, err = d2oracle.Set(graph, "meow.style.fill", "green")
// Create a shape with the ID, "cat"
graph, err = d2oracle.Create(graph, "cat")
// Move the shape "meow" inside the container "cat"
graph, err = d2oracle.Move(graph, "meow", "cat.meow")
// Prints formatted D2 code
println(d2format.Format(graph.AST))
```

This makes it easy to build functionality on top of D2. Terrastruct uses the above API to
implement editing of D2 from mouse actions in a visual interface.

## Themes

D2 includes a variety of official themes to style your diagrams beautifully right out of
the box. See [./d2themes](./d2themes) to browse the available themes and make or
contribute your own creation.

## Fonts

D2 ships with "Source Sans Pro" as the font in renders. If you wish to use a different
one, please see [./d2renderers/d2fonts](./d2renderers/d2fonts).

## Export file types

D2 currently supports SVG exports. More coming soon.

## Language tooling

D2 is designed with language tooling in mind. D2's parser can parse multiple errors from a
broken program, has an autoformatter, syntax highlighting, and we have plans for LSP's and
more. Good language tooling is necessary for creating and maintaining large diagrams.

The extensions for VSCode and Vim can be found in the [Related](#related) section.

## Plugins

D2 is designed to be extensible and composable. The plugin system allows you to
change out layout engines and customize the rendering pipeline. Plugins can either be
bundled with the build or separately installed as a standalone binary.

**Layout engines**:

- [dagre](https://github.com/dagrejs/dagre) (default, bundled): A fast, directed graph
  layout engine that produces layered/hierarchical layouts. Based on Graphviz's DOT
  algorithm.
- [ELK](https://github.com/kieler/elkjs) (bundled): A directed graph layout engine
  particularly suited for node-link diagrams with an inherent direction and ports.
- [TALA](https://github.com/terrastruct/TALA) (binary): Novel layout engine designed
  specifically for software architecture diagrams.

D2 intends to integrate with a variety of layout engines, e.g. `dot`, as well as
single-purpose layout types like sequence diagrams. You can choose whichever layout engine
you like and works best for the diagram you're making.

## Comparison

For a comparison against other popular text-to-diagram tools, see
[https://text-to-diagram.com](https://text-to-diagram.com).

## Contributing

Contributions are welcome! See [./docs/CONTRIBUTING.md](./docs/CONTRIBUTING.md).

## License

Copyright © 2022 Terrastruct, Inc. Open-source licensed under the Mozilla Public License
2.0.

## Related

### VSCode extension

[https://github.com/terrastruct/d2-vscode](https://github.com/terrastruct/d2-vscode)

### Vim extension

[https://github.com/terrastruct/d2-vim](https://github.com/terrastruct/d2-vim)

### Misc

- [https://github.com/terrastruct/d2-docs](https://github.com/terrastruct/d2-docs)
- [https://github.com/terrastruct/text-to-diagram-com](https://github.com/terrastruct/text-to-diagram-com)

## FAQ

- Does D2 collect telemetry?
  - No, D2 does not use an internet connection after installation, except to check for
    version updates from Github periodically.
- Does D2 need a browser to run?
  - No, D2 can run entirely server-side.
- I have a question or need help.
  - The best way to get help is to open an Issue, so that it's searchable by others in the
    future. If you prefer synchronous or just want to chat, you can pop into the help
    channel of the [D2 Discord](https://discord.gg/NF6X8K4eDq) as well.
- I have a feature request or proposal.
  - D2 uses Github Issues for everything. Just add a "discussion" label to your Issue.
- I have a private inquiry.
  - Please reach out at [hi@d2lang.com](hi@d2lang.com).
