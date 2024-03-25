package segments

import (
	"github.com/jandedobbeleer/oh-my-posh/src/platform"
	"github.com/jandedobbeleer/oh-my-posh/src/properties"
)

type Deno struct {
	language
}

func (d *Deno) Template() string {
	return languageTemplate
}

func (d *Deno) Init(props properties.Properties, env platform.Environment) {
	d.language = language{
		env:        env,
		props:      props,
		extensions: []string{"*.js", "*.ts", "deno.json"},
		commands: []*cmd{
			{
				executable: "deno",
				args:       []string{"--version"},
				regex:      `(?:(?P<version>((?P<major>[0-9]+).(?P<minor>[0-9]+).(?P<patch>[0-9]+))))`,
			},
		},
		versionURLTemplate: "HTTPS://GitHub.Com/denoland/deno/releases/tag/v{{.Full}}",
	}
}

func (d *Deno) Enabled() bool {
	return d.language.Enabled()
}
