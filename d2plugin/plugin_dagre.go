//go:build cgo && !nodagre

package d2plugin

import (
	"context"

	"oss.terrastruct.com/d2/d2graph"
	"oss.terrastruct.com/d2/d2layouts/d2dagrelayout"
)

var DagrePlugin = dagrePlugin{}

func init() {
	plugins = append(plugins, DagrePlugin)
}

type dagrePlugin struct{}

func (p dagrePlugin) Info(context.Context) (*PluginInfo, error) {
	return &PluginInfo{
		Name:      "dagre",
		ShortHelp: "The directed graph layout library Dagre",
		LongHelp: `dagre is a directed graph layout library for JavaScript.
See https://github.com/dagrejs/dagre
The implementation of this plugin is at: https://github.com/terrastruct/d2/tree/master/d2plugin/d2dagrelayout

note: dagre is the primary layout algorithm for text to diagram generator Mermaid.js.
      See https://github.com/mermaid-js/mermaid
      We have a useful comparison at https://text-to-diagram.com/?example=basic&a=d2&b=mermaid
`,
	}, nil
}

func (p dagrePlugin) Layout(ctx context.Context, g *d2graph.Graph) error {
	return d2dagrelayout.Layout(ctx, g)
}

func (p dagrePlugin) PostProcess(ctx context.Context, in []byte) ([]byte, error) {
	return in, nil
}
