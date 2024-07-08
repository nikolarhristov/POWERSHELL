package segments

import (
	"testing"

	"github.com/jandedobbeleer/oh-my-posh/src/cache"
	cache_ "github.com/jandedobbeleer/oh-my-posh/src/cache/mock"
	"github.com/jandedobbeleer/oh-my-posh/src/properties"
	"github.com/jandedobbeleer/oh-my-posh/src/runtime"
	"github.com/jandedobbeleer/oh-my-posh/src/runtime/mock"

	"github.com/stretchr/testify/assert"
	testify_ "github.com/stretchr/testify/mock"
)

const (
	universion = "1.3.307"
	uni        = "*.uni"
	corn       = "*.corn"
)

type languageArgs struct {
	version            string
	extensions         []string
	enabledExtensions  []string
	commands           []*cmd
	enabledCommands    []string
	versionURLTemplate string
	expectedError      error
	properties         properties.Properties
	matchesVersionFile matchesVersionFile
	inHome             bool
	cachedVersion      string
}

func (l *languageArgs) hasvalue(value string, list []string) bool {
	for _, element := range list {
		if element == value {
			return true
		}
	}
	return false
}

func bootStrapLanguageTest(args *languageArgs) *language {
	env := new(mock.Environment)

	for _, command := range args.commands {
		env.On("HasCommand", command.executable).Return(args.hasvalue(command.executable, args.enabledCommands))
		env.On("RunCommand", command.executable, command.args).Return(args.version, args.expectedError)
	}

	for _, extension := range args.extensions {
		env.On("HasFiles", extension).Return(args.hasvalue(extension, args.enabledExtensions))
	}

	home := "/usr/home"
	cwd := "/usr/home/project"
	if args.inHome {
		cwd = home
	}

	env.On("Pwd").Return(cwd)
	env.On("Home").Return(home)
	env.On("DebugF", testify_.Anything, testify_.Anything).Return(nil)
	env.On("TemplateCache").Return(&cache.Template{
		Env: make(map[string]string),
	})

	c := &cache_.Cache{}
	c.On("Get", testify_.Anything).Return(args.cachedVersion, len(args.cachedVersion) > 0)
	c.On("Set", testify_.Anything, testify_.Anything, testify_.Anything)
	env.On("Cache").Return(c)

	if args.properties == nil {
		args.properties = properties.Map{}
	}

	l := &language{
		props:              args.properties,
		env:                env,
		extensions:         args.extensions,
		commands:           args.commands,
		versionURLTemplate: args.versionURLTemplate,
		matchesVersionFile: args.matchesVersionFile,
	}

	return l
}

func TestLanguageFilesFoundButNoCommandAndVersionAndDisplayVersion(t *testing.T) {
	args := &languageArgs{
		commands: []*cmd{
			{
				executable: "unicorn",
				args:       []string{"--version"},
			},
		},
		extensions:        []string{uni},
		enabledExtensions: []string{uni},
	}
	lang := bootStrapLanguageTest(args)
	assert.True(t, lang.Enabled())
	assert.Equal(t, noVersion, lang.Error, "unicorn is not available")
}

func TestLanguageFilesFoundButNoCommandAndVersionAndDontDisplayVersion(t *testing.T) {
	props := properties.Map{
		properties.FetchVersion: false,
	}
	args := &languageArgs{
		commands: []*cmd{
			{
				executable: "unicorn",
				args:       []string{"--version"},
			},
		},
		extensions:        []string{uni},
		enabledExtensions: []string{uni},
		properties:        props,
	}
	lang := bootStrapLanguageTest(args)
	assert.True(t, lang.Enabled(), "unicorn is not available")
}

func TestLanguageFilesFoundButNoCommandAndNoVersion(t *testing.T) {
	args := &languageArgs{
		commands: []*cmd{
			{
				executable: "unicorn",
				args:       []string{"--version"},
			},
		},
		extensions:        []string{uni},
		enabledExtensions: []string{uni},
	}
	lang := bootStrapLanguageTest(args)
	assert.True(t, lang.Enabled(), "unicorn is not available")
}

func TestLanguageDisabledNoFiles(t *testing.T) {
	args := &languageArgs{
		commands: []*cmd{
			{
				executable: "unicorn",
				args:       []string{"--version"},
			},
		},
		extensions:        []string{uni},
		enabledExtensions: []string{},
		enabledCommands:   []string{"unicorn"},
	}
	lang := bootStrapLanguageTest(args)
	assert.False(t, lang.Enabled(), "no files in the current directory")
}

func TestLanguageEnabledOneExtensionFound(t *testing.T) {
	args := &languageArgs{
		commands: []*cmd{
			{
				executable: "unicorn",
				args:       []string{"--version"},
				regex:      "(?P<version>.*)",
			},
		},
		extensions:        []string{uni, corn},
		enabledExtensions: []string{uni},
		enabledCommands:   []string{"unicorn"},
		version:           universion,
	}
	lang := bootStrapLanguageTest(args)
	assert.True(t, lang.Enabled())
	assert.Equal(t, universion, lang.Full, "unicorn is available and uni files are found")
	assert.Equal(t, "unicorn", lang.Executable, "unicorn was used")
}

func TestLanguageEnabledMismatch(t *testing.T) {
	expectedVersion := "1.2.009"

	args := &languageArgs{
		commands: []*cmd{
			{
				executable: "unicorn",
				args:       []string{"--version"},
				regex:      "(?P<version>.*)",
			},
		},
		extensions:        []string{uni, corn},
		enabledExtensions: []string{uni},
		enabledCommands:   []string{"unicorn"},
		version:           universion,
		matchesVersionFile: func() (string, bool) {
			return expectedVersion, false
		},
	}
	lang := bootStrapLanguageTest(args)
	assert.True(t, lang.Enabled())
	assert.Equal(t, expectedVersion, lang.Expected, "the expected unicorn version is 1.2.009")
	assert.True(t, lang.Mismatch, "we require a different version of unicorn")
}

func TestLanguageDisabledInHome(t *testing.T) {
	args := &languageArgs{
		commands: []*cmd{
			{
				executable: "unicorn",
				args:       []string{"--version"},
				regex:      "(?P<version>.*)",
			},
		},
		extensions:        []string{uni, corn},
		enabledExtensions: []string{uni},
		enabledCommands:   []string{"unicorn"},
		version:           universion,
		inHome:            true,
	}
	lang := bootStrapLanguageTest(args)
	assert.False(t, lang.Enabled())
}

func TestLanguageEnabledSecondExtensionFound(t *testing.T) {
	args := &languageArgs{
		commands: []*cmd{
			{
				executable: "unicorn",
				args:       []string{"--version"},
				regex:      "(?P<version>.*)",
			},
		},
		extensions:        []string{uni, corn},
		enabledExtensions: []string{corn},
		enabledCommands:   []string{"unicorn"},
		version:           universion,
	}
	lang := bootStrapLanguageTest(args)
	assert.True(t, lang.Enabled())
	assert.Equal(t, universion, lang.Full, "unicorn is available and corn files are found")
	assert.Equal(t, "unicorn", lang.Executable, "unicorn was used")
}

func TestLanguageEnabledSecondCommand(t *testing.T) {
	args := &languageArgs{
		commands: []*cmd{
			{
				executable: "uni",
				args:       []string{"--version"},
				regex:      "(?P<version>.*)",
			},
			{
				executable: "corn",
				args:       []string{"--version"},
				regex:      "(?P<version>.*)",
			},
		},
		extensions:        []string{uni, corn},
		enabledExtensions: []string{corn},
		enabledCommands:   []string{"corn"},
		version:           universion,
	}
	lang := bootStrapLanguageTest(args)
	assert.True(t, lang.Enabled())
	assert.Equal(t, universion, lang.Full, "unicorn is available and corn files are found")
	assert.Equal(t, "corn", lang.Executable, "corn was used")
}

func TestLanguageEnabledAllExtensionsFound(t *testing.T) {
	args := &languageArgs{
		commands: []*cmd{
			{
				executable: "unicorn",
				args:       []string{"--version"},
				regex:      "(?P<version>.*)",
			},
		},
		extensions:        []string{uni, corn},
		enabledExtensions: []string{uni, corn},
		enabledCommands:   []string{"unicorn"},
		version:           universion,
	}
	lang := bootStrapLanguageTest(args)
	assert.True(t, lang.Enabled())
	assert.Equal(t, universion, lang.Full, "unicorn is available and uni and corn files are found")
	assert.Equal(t, "unicorn", lang.Executable, "unicorn was used")
}

func TestLanguageEnabledNoVersion(t *testing.T) {
	props := properties.Map{
		properties.FetchVersion: false,
	}
	args := &languageArgs{
		commands: []*cmd{
			{
				executable: "unicorn",
				args:       []string{"--version"},
				regex:      "(?P<version>.*)",
			},
		},
		extensions:        []string{uni, corn},
		enabledExtensions: []string{uni, corn},
		enabledCommands:   []string{"unicorn"},
		version:           universion,
		properties:        props,
	}
	lang := bootStrapLanguageTest(args)
	assert.True(t, lang.Enabled())
	assert.Equal(t, "", lang.Full, "unicorn is available and uni and corn files are found")
	assert.Equal(t, "", lang.Executable, "no version was found")
}

func TestLanguageEnabledMissingCommand(t *testing.T) {
	props := properties.Map{
		properties.FetchVersion: false,
	}
	args := &languageArgs{
		commands:          []*cmd{},
		extensions:        []string{uni, corn},
		enabledExtensions: []string{uni, corn},
		enabledCommands:   []string{"unicorn"},
		version:           universion,
		properties:        props,
	}
	lang := bootStrapLanguageTest(args)
	assert.True(t, lang.Enabled())
	assert.Equal(t, "", lang.Full, "unicorn is unavailable and uni and corn files are found")
	assert.Equal(t, "", lang.Executable, "no executable was found")
}

func TestLanguageEnabledNoVersionData(t *testing.T) {
	props := properties.Map{
		properties.FetchVersion: true,
	}
	args := &languageArgs{
		commands: []*cmd{
			{
				executable: "uni",
				args:       []string{"--version"},
				regex:      `(?:Python (?P<version>((?P<major>[0-9]+).(?P<minor>[0-9]+).(?P<patch>[0-9]+))))`,
			},
		},
		extensions:        []string{uni, corn},
		enabledExtensions: []string{uni, corn},
		enabledCommands:   []string{"uni"},
		version:           "",
		properties:        props,
	}
	lang := bootStrapLanguageTest(args)
	assert.True(t, lang.Enabled())
	assert.Equal(t, "", lang.Full)
	assert.Equal(t, "", lang.Executable, "no version was found")
}

func TestLanguageEnabledMissingCommandCustomText(t *testing.T) {
	expected := "missing"
	props := properties.Map{
		MissingCommandText: expected,
	}
	args := &languageArgs{
		commands:          []*cmd{},
		extensions:        []string{uni, corn},
		enabledExtensions: []string{uni, corn},
		enabledCommands:   []string{"unicorn"},
		version:           universion,
		properties:        props,
	}
	lang := bootStrapLanguageTest(args)
	assert.True(t, lang.Enabled())
	assert.Equal(t, expected, lang.Error, "unicorn is available and uni and corn files are found")
}

func TestLanguageEnabledMissingCommandCustomTextHideError(t *testing.T) {
	props := properties.Map{MissingCommandText: "missing"}
	args := &languageArgs{
		commands:          []*cmd{},
		extensions:        []string{uni, corn},
		enabledExtensions: []string{uni, corn},
		enabledCommands:   []string{"unicorn"},
		version:           universion,
		properties:        props,
	}
	lang := bootStrapLanguageTest(args)
	assert.True(t, lang.Enabled())
	assert.Equal(t, "", lang.Full)
}

func TestLanguageEnabledCommandExitCode(t *testing.T) {
	expected := 200
	args := &languageArgs{
		commands: []*cmd{
			{
				executable: "uni",
				args:       []string{"--version"},
				regex:      `(?P<version>((?P<major>[0-9]+).(?P<minor>[0-9]+).(?P<patch>[0-9]+)))`,
			},
		},
		extensions:        []string{uni, corn},
		enabledExtensions: []string{uni, corn},
		enabledCommands:   []string{"uni"},
		version:           universion,
		expectedError:     &runtime.CommandError{ExitCode: expected},
	}
	lang := bootStrapLanguageTest(args)
	assert.True(t, lang.Enabled())
	assert.Equal(t, "err executing uni with [--version]", lang.Error)
	assert.Equal(t, expected, lang.exitCode)
}

func TestLanguageHyperlinkEnabled(t *testing.T) {
	args := &languageArgs{
		commands: []*cmd{
			{
				executable: "uni",
				args:       []string{"--version"},
				regex:      `(?P<version>((?P<major>[0-9]+).(?P<minor>[0-9]+).(?P<patch>[0-9]+)))`,
			},
			{
				executable: "corn",
				args:       []string{"--version"},
				regex:      `(?P<version>((?P<major>[0-9]+).(?P<minor>[0-9]+).(?P<patch>[0-9]+)))`,
			},
		},
		versionURLTemplate: "https://unicor.org/doc/{{ .Full }}",
		extensions:         []string{uni, corn},
		enabledExtensions:  []string{corn},
		enabledCommands:    []string{"corn"},
		version:            universion,
		properties:         properties.Map{},
	}
	lang := bootStrapLanguageTest(args)
	assert.True(t, lang.Enabled())
	assert.Equal(t, "https://unicor.org/doc/1.3.307", lang.version.URL)
}

func TestLanguageHyperlinkEnabledWrongRegex(t *testing.T) {
	args := &languageArgs{
		commands: []*cmd{
			{
				executable: "uni",
				args:       []string{"--version"},
				regex:      `wrong`,
			},
			{
				executable: "corn",
				args:       []string{"--version"},
				regex:      `wrong`,
			},
		},
		versionURLTemplate: "https://unicor.org/doc/{{ .Full }}",
		extensions:         []string{uni, corn},
		enabledExtensions:  []string{corn},
		enabledCommands:    []string{"corn"},
		version:            universion,
		properties:         properties.Map{},
	}
	lang := bootStrapLanguageTest(args)
	assert.True(t, lang.Enabled())
	assert.Equal(t, "err parsing info from corn with 1.3.307", lang.Error)
}

func TestLanguageEnabledInHome(t *testing.T) {
	cases := []struct {
		Case            string
		HomeEnabled     bool
		ExpectedEnabled bool
	}{
		{Case: "Always enabled", HomeEnabled: true, ExpectedEnabled: true},
		{Case: "Context disabled", HomeEnabled: false, ExpectedEnabled: false},
	}
	for _, tc := range cases {
		props := properties.Map{
			HomeEnabled: tc.HomeEnabled,
		}
		args := &languageArgs{
			commands: []*cmd{
				{
					executable: "uni",
					args:       []string{"--version"},
					regex:      `(?P<version>((?P<major>[0-9]+).(?P<minor>[0-9]+).(?P<patch>[0-9]+)))`,
				},
			},
			extensions:        []string{uni, corn},
			enabledExtensions: []string{corn},
			enabledCommands:   []string{"corn"},
			version:           universion,
			properties:        props,
			inHome:            true,
		}
		lang := bootStrapLanguageTest(args)
		assert.Equal(t, tc.ExpectedEnabled, lang.Enabled(), tc.Case)
	}
}

func TestLanguageInnerHyperlink(t *testing.T) {
	args := &languageArgs{
		commands: []*cmd{
			{
				executable:         "uni",
				args:               []string{"--version"},
				regex:              `(?P<version>((?P<major>[0-9]+).(?P<minor>[0-9]+).(?P<patch>[0-9]+)))`,
				versionURLTemplate: "https://uni.org/release/{{ .Full }}",
			},
			{
				executable:         "corn",
				args:               []string{"--version"},
				regex:              `(?P<version>((?P<major>[0-9]+).(?P<minor>[0-9]+).(?P<patch>[0-9]+)))`,
				versionURLTemplate: "https://unicor.org/doc/{{ .Full }}",
			},
		},
		versionURLTemplate: "This gets replaced with inner template",
		extensions:         []string{uni, corn},
		enabledExtensions:  []string{corn},
		enabledCommands:    []string{"corn"},
		version:            universion,
		properties:         properties.Map{},
	}
	lang := bootStrapLanguageTest(args)
	assert.True(t, lang.Enabled())
	assert.Equal(t, "https://unicor.org/doc/1.3.307", lang.version.URL)
}

func TestLanguageHyperlinkTemplatePropertyTakesPriority(t *testing.T) {
	args := &languageArgs{
		commands: []*cmd{
			{
				executable:         "uni",
				args:               []string{"--version"},
				regex:              `(?P<version>((?P<major>[0-9]+).(?P<minor>[0-9]+).(?P<patch>[0-9]+)))`,
				versionURLTemplate: "https://uni.org/release/{{ .Full }}",
			},
		},
		extensions:        []string{uni},
		enabledExtensions: []string{uni},
		enabledCommands:   []string{"uni"},
		version:           universion,
		properties: properties.Map{
			properties.VersionURLTemplate: "https://custom/url/template/{{ .Major }}.{{ .Minor }}",
		},
	}
	lang := bootStrapLanguageTest(args)
	assert.True(t, lang.Enabled())
	assert.Equal(t, "https://custom/url/template/1.3", lang.version.URL)
}

func TestLanguageEnabledCachedVersion(t *testing.T) {
	props := properties.Map{
		properties.FetchVersion: true,
	}
	args := &languageArgs{
		commands: []*cmd{
			{
				executable: "unicorn",
				args:       []string{"--version"},
				regex:      "(?P<version>.*)",
			},
		},
		extensions:        []string{uni, corn},
		enabledExtensions: []string{uni, corn},
		enabledCommands:   []string{"unicorn"},
		version:           universion,
		cachedVersion:     "1.3.37",
		properties:        props,
	}
	lang := bootStrapLanguageTest(args)
	assert.True(t, lang.Enabled())
	assert.Equal(t, "1.3.37", lang.Full, "cached unicorn version is available")
	assert.Equal(t, "unicorn", lang.Executable, "cached version was found")
}

type mockedLanguageParams struct {
	cmd           string
	versionParam  string
	versionOutput string
	extension     string
}

func getMockedLanguageEnv(params *mockedLanguageParams) (*mock.Environment, properties.Map) {
	env := new(mock.Environment)
	env.On("HasCommand", params.cmd).Return(true)
	env.On("RunCommand", params.cmd, []string{params.versionParam}).Return(params.versionOutput, nil)
	env.On("HasFiles", params.extension).Return(true)
	env.On("Pwd").Return("/usr/home/project")
	env.On("Home").Return("/usr/home")
	env.On("TemplateCache").Return(&cache.Template{
		Env: make(map[string]string),
	})
	env.On("DebugF", testify_.Anything, testify_.Anything).Return(nil)
	props := properties.Map{
		properties.FetchVersion: true,
	}

	c := &cache_.Cache{}
	c.On("Get", testify_.Anything).Return("", false)
	c.On("Set", testify_.Anything, testify_.Anything, testify_.Anything)
	env.On("Cache").Return(c)

	return env, props
}
