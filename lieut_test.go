package lieut

import (
	"bytes"
	"context"
	"flag"
	"fmt"
	"io"
	"os"
	"runtime"
	"sort"
	"testing"
)

var testAppInfo = AppInfo{
	Name:    "test",
	Summary: "A test",
	Usage:   "testing",
	Version: "vTest",
}

var testNoOpExecutor = func(ctx context.Context, arguments []string, out io.Writer) error {
	return nil
}

func TestNewSingleCommandApp(t *testing.T) {
	flagSet := flag.NewFlagSet(testAppInfo.Name, flag.ExitOnError)

	app := NewSingleCommandApp(testAppInfo, testNoOpExecutor, flagSet, os.Stdout, os.Stderr)

	if app == nil {
		t.Error("NewSingleCommandApp returned nil")
	}
}

func TestNewSingleCommandApp_ZeroValues(t *testing.T) {
	app := NewSingleCommandApp(AppInfo{}, nil, nil, nil, nil)

	if app == nil {
		t.Fatal("NewSingleCommandApp returned nil")
	}

	if inferredName := inferAppName(); app.info.Name != inferredName {
		t.Errorf("NewSingleCommandApp with no given name gave %q name, wanted %q", app.info.Name, inferredName)
	}

	if app.info.Usage != DefaultCommandUsage {
		t.Errorf("NewSingleCommandApp with no given usage gave %q usage, wanted %q", app.info.Usage, DefaultCommandUsage)
	}

	if app.flags.Flags == nil {
		t.Errorf("NewSingleCommandApp with no given flags had nil flags")
	}
}

func TestNewMultiCommandApp(t *testing.T) {
	flagSet := flag.NewFlagSet(testAppInfo.Name, flag.ExitOnError)

	app := NewMultiCommandApp(testAppInfo, flagSet, os.Stdout, os.Stderr)

	if app == nil {
		t.Error("NewMultiCommandApp returned nil")
	}
}

func TestNewMultiCommandApp_ZeroValues(t *testing.T) {
	app := NewMultiCommandApp(AppInfo{}, nil, nil, nil)

	if app == nil {
		t.Fatal("NewMultiCommandApp returned nil")
	}

	if inferredName := inferAppName(); app.info.Name != inferredName {
		t.Errorf("NewMultiCommandApp with no given name gave %q name, wanted %q", app.info.Name, inferredName)
	}

	if app.info.Usage != DefaultParentCommandUsage {
		t.Errorf("NewMultiCommandApp with no given usage gave %q usage, wanted %q", app.info.Usage, DefaultParentCommandUsage)
	}

	if app.flags.Flags == nil {
		t.Errorf("NewMultiCommandApp with no given flags had nil flags")
	}
}

func TestMultiCommandApp_SetCommand(t *testing.T) {
	flagSet := flag.NewFlagSet(testAppInfo.Name, flag.ExitOnError)

	app := NewMultiCommandApp(testAppInfo, flagSet, os.Stdout, os.Stderr)

	for testName, testData := range map[string]struct {
		info    CommandInfo
		exec    Executor
		flags   Flags
		wantErr bool
	}{
		"all": {
			info: CommandInfo{
				Name:    "test",
				Summary: "testing",
				Usage:   "test testing test",
			},
			exec:  testNoOpExecutor,
			flags: flag.NewFlagSet(testAppInfo.Name, flag.ExitOnError),
		},
		"only info": {
			info: CommandInfo{
				Name:    "test",
				Summary: "testing",
				Usage:   "test testing test",
			},
		},
		"only exec": {
			exec: testNoOpExecutor,
		},
		"only flags": {
			flags: flag.NewFlagSet(testAppInfo.Name, flag.ExitOnError),
		},
		"zero values": {},
		"duplicate flags": {
			flags:   flagSet,
			wantErr: true,
		},
	} {
		t.Run(testName, func(t *testing.T) {
			err := app.SetCommand(testData.info, testData.exec, testData.flags)
			if err != nil && !testData.wantErr {
				t.Errorf("SetCommand returned error: %v", err)
			}
		})
	}
}

func TestMultiCommandApp_CommandNames(t *testing.T) {
	flagSet := flag.NewFlagSet(testAppInfo.Name, flag.ExitOnError)

	app := NewMultiCommandApp(testAppInfo, flagSet, os.Stdout, os.Stderr)

	if names := app.CommandNames(); len(names) > 0 {
		t.Errorf("CommandNames returned a non-empty slice %v", names)
	}

	app.SetCommand(CommandInfo{Name: "foo"}, nil, nil)
	app.SetCommand(CommandInfo{Name: "bar"}, nil, nil)

	names := app.CommandNames()
	sort.Strings(names)

	if names[0] != "bar" && names[1] != "foo" {
		t.Errorf("CommandNames returned an unexpected slice %v", names)
	}
}

func TestSingleCommandApp_PrintVersion(t *testing.T) {
	for testName, testData := range map[string]struct {
		version string
		want    string
	}{
		"specified": {
			version: "vTest",
			want:    fmt.Sprintf("%s vTest (%s/%s)\n", testAppInfo.Name, runtime.GOOS, runtime.GOARCH),
		},
		"no version string": {
			version: "",
			want:    fmt.Sprintf("%s (%s/%s)\n", testAppInfo.Name, runtime.GOOS, runtime.GOARCH),
		},
	} {
		t.Run(testName, func(t *testing.T) {
			appInfo := testAppInfo
			appInfo.Version = testData.version
			var buf bytes.Buffer

			flagSet := flag.NewFlagSet(testAppInfo.Name, flag.ExitOnError)
			app := NewSingleCommandApp(appInfo, testNoOpExecutor, flagSet, &buf, &buf)

			app.PrintVersion()

			got := buf.String()

			if got != testData.want {
				t.Errorf("app.PrintVersion gave %q, want %q", got, testData.want)
			}
		})
	}
}

func TestMultiCommandApp_PrintVersion(t *testing.T) {
	for testName, testData := range map[string]struct {
		version string
		want    string
	}{
		"specified": {
			version: "vTest",
			want:    fmt.Sprintf("%s vTest (%s/%s)\n", testAppInfo.Name, runtime.GOOS, runtime.GOARCH),
		},
		"no version string": {
			version: "",
			want:    fmt.Sprintf("%s (%s/%s)\n", testAppInfo.Name, runtime.GOOS, runtime.GOARCH),
		},
	} {
		t.Run(testName, func(t *testing.T) {
			appInfo := testAppInfo
			appInfo.Version = testData.version
			var buf bytes.Buffer

			flagSet := flag.NewFlagSet(testAppInfo.Name, flag.ExitOnError)
			app := NewMultiCommandApp(appInfo, flagSet, &buf, &buf)

			app.PrintVersion()

			got := buf.String()

			if got != testData.want {
				t.Errorf("app.PrintVersion gave %q, want %q", got, testData.want)
			}
		})
	}
}

func TestSingleCommandApp_PrintUsage(t *testing.T) {
	for testName, testData := range map[string]struct {
		usage string
		want  string
	}{
		"specified": {
			usage: "testing [test]",
			want:  fmt.Sprintf("Usage: %s testing [test]\n", testAppInfo.Name),
		},
		"no usage string": {
			usage: "",
			want:  fmt.Sprintf("Usage: %s %s\n", testAppInfo.Name, DefaultCommandUsage),
		},
	} {
		t.Run(testName, func(t *testing.T) {
			appInfo := testAppInfo
			appInfo.Usage = testData.usage
			var buf bytes.Buffer

			flagSet := flag.NewFlagSet(testAppInfo.Name, flag.ExitOnError)
			app := NewSingleCommandApp(appInfo, testNoOpExecutor, flagSet, &buf, &buf)

			app.PrintUsage()

			got := buf.String()

			if got != testData.want {
				t.Errorf("app.PrintUsage gave %q, want %q", got, testData.want)
			}
		})
	}
}

func TestMultiCommandApp_PrintUsage(t *testing.T) {
	testCommandInfo := CommandInfo{
		Name:    "test",
		Summary: "testing",
	}

	for testName, testData := range map[string]struct {
		appUsage     string
		commandUsage string
		forCommand   string
		want         string
	}{
		"specified app and command usage, for command": {
			appUsage:     "testing [test]",
			commandUsage: "test [opts]",
			forCommand:   testCommandInfo.Name,
			want:         fmt.Sprintf("Usage: %s %s test [opts]\n", testAppInfo.Name, testCommandInfo.Name),
		},
		"specified app usage, no command usage, for command": {
			appUsage:     "testing [test]",
			commandUsage: "",
			forCommand:   testCommandInfo.Name,
			want:         fmt.Sprintf("Usage: %s %s %s\n", testAppInfo.Name, testCommandInfo.Name, DefaultCommandUsage),
		},
		"no app usage, specified command usage, for command": {
			appUsage:     "",
			commandUsage: "test [opts]",
			forCommand:   testCommandInfo.Name,
			want:         fmt.Sprintf("Usage: %s %s test [opts]\n", testAppInfo.Name, testCommandInfo.Name),
		},
		"no app or command usage, for command": {
			appUsage:     "",
			commandUsage: "",
			forCommand:   testCommandInfo.Name,
			want:         fmt.Sprintf("Usage: %s %s %s\n", testAppInfo.Name, testCommandInfo.Name, DefaultCommandUsage),
		},
		"specified app and command usage, no command": {
			appUsage:     "testing [test]",
			commandUsage: "test [opts]",
			forCommand:   "",
			want:         fmt.Sprintf("Usage: %s testing [test]\n", testAppInfo.Name),
		},
		"specified app usage, no command usage, no command": {
			appUsage:     "testing [test]",
			commandUsage: "",
			forCommand:   "",
			want:         fmt.Sprintf("Usage: %s testing [test]\n", testAppInfo.Name),
		},
		"no app usage, specified command usage, no command": {
			appUsage:     "",
			commandUsage: "test [opts]",
			forCommand:   "",
			want:         fmt.Sprintf("Usage: %s %s\n", testAppInfo.Name, DefaultParentCommandUsage),
		},
		"no app or command usage, no command": {
			appUsage:     "",
			commandUsage: "",
			forCommand:   "",
			want:         fmt.Sprintf("Usage: %s %s\n", testAppInfo.Name, DefaultParentCommandUsage),
		},
	} {
		t.Run(testName, func(t *testing.T) {
			appInfo := testAppInfo
			appInfo.Usage = testData.appUsage
			var buf bytes.Buffer

			flagSet := flag.NewFlagSet(testAppInfo.Name, flag.ExitOnError)
			app := NewMultiCommandApp(appInfo, flagSet, &buf, &buf)

			testCommandInfo.Usage = testData.commandUsage
			commandFlagSet := flag.NewFlagSet(testAppInfo.Name, flag.ExitOnError)
			err := app.SetCommand(testCommandInfo, testNoOpExecutor, commandFlagSet)
			if err != nil {
				t.Errorf("SetCommand returned error: %v", err)
			}

			app.PrintUsage(testData.forCommand)

			got := buf.String()

			if got != testData.want {
				t.Errorf("app.PrintUsage gave %q, want %q", got, testData.want)
			}
		})
	}
}

func TestSingleCommandApp_PrintHelp(t *testing.T) {
	wantFormat := `Usage: test testing

A test

Options:

  -help
    	Display the help message
  -testflag string
    	A test flag (default "testval")
  -version
    	Display the application version

test vTest (%s/%s)
`
	want := fmt.Sprintf(wantFormat, runtime.GOOS, runtime.GOARCH)

	var buf bytes.Buffer

	flagSet := flag.NewFlagSet(testAppInfo.Name, flag.ExitOnError)
	flagSet.String("testflag", "testval", "A test flag")

	app := NewSingleCommandApp(testAppInfo, testNoOpExecutor, flagSet, &buf, &buf)

	app.PrintHelp()

	got := buf.String()

	if got != want {
		t.Errorf("app.PrintHelp gave %q, want %q", got, want)
	}
}

func TestMultiCommandApp_PrintHelp(t *testing.T) {
	wantFormat := `Usage: test testing

A test

Commands:

	testcommand	A test command

Options:

  -help
    	Display the help message
  -testflag string
    	A test flag (default "testval")
  -version
    	Display the application version

test vTest (%s/%s)
`
	want := fmt.Sprintf(wantFormat, runtime.GOOS, runtime.GOARCH)

	var buf bytes.Buffer

	flagSet := flag.NewFlagSet(testAppInfo.Name, flag.ExitOnError)
	flagSet.String("testflag", "testval", "A test flag")

	app := NewMultiCommandApp(testAppInfo, flagSet, &buf, &buf)

	testCommandInfo := CommandInfo{
		Name:    "testcommand",
		Summary: "A test command",
		Usage:   "args here...",
	}

	commandFlagSet := flag.NewFlagSet(testCommandInfo.Name, flag.ExitOnError)
	commandFlagSet.Int("testcommandflag", 5, "A test command flag")

	err := app.SetCommand(testCommandInfo, testNoOpExecutor, commandFlagSet)
	if err != nil {
		t.Fatalf("SetCommand returned error: %v", err)
	}

	app.PrintHelp("")

	got := buf.String()

	if got != want {
		t.Errorf("app.PrintHelp gave %q, want %q", got, want)
	}
}

func TestMultiCommandApp_PrintHelp_Command(t *testing.T) {
	wantFormat := `Usage: test testcommand args here...

A test command

Options:

  -help
    	Display the help message
  -testcommandflag int
    	A test command flag (default 5)
  -testflag string
    	A test flag (default "testval")
  -version
    	Display the application version

test vTest (%s/%s)
`
	want := fmt.Sprintf(wantFormat, runtime.GOOS, runtime.GOARCH)

	var buf bytes.Buffer

	flagSet := flag.NewFlagSet(testAppInfo.Name, flag.ExitOnError)
	flagSet.String("testflag", "testval", "A test flag")

	app := NewMultiCommandApp(testAppInfo, flagSet, &buf, &buf)

	testCommandInfo := CommandInfo{
		Name:    "testcommand",
		Summary: "A test command",
		Usage:   "args here...",
	}

	commandFlagSet := flag.NewFlagSet(testCommandInfo.Name, flag.ExitOnError)
	commandFlagSet.Int("testcommandflag", 5, "A test command flag")

	err := app.SetCommand(testCommandInfo, testNoOpExecutor, commandFlagSet)
	if err != nil {
		t.Fatalf("SetCommand returned error: %v", err)
	}

	app.PrintHelp("testcommand")

	got := buf.String()

	if got != want {
		t.Errorf("app.PrintHelp gave %q, want %q", got, want)
	}
}

func TestSingleCommandApp_Run(t *testing.T) {
	var executorCapture struct {
		ctx       context.Context
		arguments []string
		out       io.Writer
	}

	executor := func(ctx context.Context, arguments []string, out io.Writer) error {
		executorCapture.ctx = ctx
		executorCapture.arguments = arguments
		executorCapture.out = out

		return nil
	}

	flagSet := flag.NewFlagSet(testAppInfo.Name, flag.ExitOnError)
	out := io.Discard

	app := NewSingleCommandApp(testAppInfo, executor, flagSet, out, out)

	ctxTestKey := struct{ k string }{k: "test-key-for-testing"}
	ctxTestVal := "test context val"
	ctx := context.WithValue(context.TODO(), ctxTestKey, ctxTestVal)
	args := []string{"testarg1", "testarg2"}
	wantedExitCode := 0

	initRan := false
	initFn := func() error {
		initRan = true
		return nil
	}

	app.OnInit(initFn)

	exitCode := app.Run(ctx, args)

	if exitCode != wantedExitCode {
		t.Errorf("app.Run gave %q, wanted %q", exitCode, wantedExitCode)
	}

	if !initRan {
		t.Error("app.Run didn't run init function")
	}

	if executorCapture.ctx.Value(ctxTestKey) != ctxTestVal {
		t.Errorf("app.Run executor gave %q, wanted %q", executorCapture.ctx.Value(ctxTestKey), ctxTestVal)
	}

	if executorCapture.arguments[0] != args[0] && executorCapture.arguments[1] != args[1] {
		t.Errorf("app.Run executor gave %q, wanted %q", executorCapture.arguments, args)
	}

	if executorCapture.out != out {
		t.Errorf("app.Run executor gave %q, wanted %q", executorCapture.out, out)
	}
}

func TestMultiCommandApp_Run(t *testing.T) {
	flagSet := flag.NewFlagSet(testAppInfo.Name, flag.ExitOnError)
	out := io.Discard

	app := NewMultiCommandApp(testAppInfo, flagSet, out, out)

	testCommandInfo := CommandInfo{Name: "testcommand"}

	var executorCapture struct {
		ctx       context.Context
		arguments []string
		out       io.Writer
	}
	executor := func(ctx context.Context, arguments []string, out io.Writer) error {
		executorCapture.ctx = ctx
		executorCapture.arguments = arguments
		executorCapture.out = out

		return nil
	}
	commandFlagSet := flag.NewFlagSet(testCommandInfo.Name, flag.ExitOnError)

	app.SetCommand(testCommandInfo, executor, commandFlagSet)

	ctxTestKey := struct{ k string }{k: "test-key-for-testing"}
	ctxTestVal := "test context val"
	ctx := context.WithValue(context.TODO(), ctxTestKey, ctxTestVal)
	args := []string{testCommandInfo.Name, "testarg1", "testarg2"}
	wantedExitCode := 0

	initRan := false
	initFn := func() error {
		initRan = true
		return nil
	}

	app.OnInit(initFn)

	exitCode := app.Run(ctx, args)

	if exitCode != wantedExitCode {
		t.Errorf("app.Run gave %q, wanted %q", exitCode, wantedExitCode)
	}

	if !initRan {
		t.Error("app.Run didn't run init function")
	}

	if executorCapture.ctx.Value(ctxTestKey) != ctxTestVal {
		t.Errorf("app.Run executor gave %q, wanted %q", executorCapture.ctx.Value(ctxTestKey), ctxTestVal)
	}

	if executorCapture.arguments[0] != args[1] && executorCapture.arguments[1] != args[2] {
		t.Errorf("app.Run executor gave %q, wanted %q", executorCapture.arguments, args)
	}

	if executorCapture.out != out {
		t.Errorf("app.Run executor gave %q, wanted %q", executorCapture.out, out)
	}
}
