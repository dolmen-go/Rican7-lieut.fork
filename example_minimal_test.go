// Copyright © 2023 Trevor N. Suarez (Rican7)

package lieut_test

import (
	"context"
	"flag"
	"fmt"
	"io"
	"os"

	"github.com/Rican7/lieut"
)

func Example_minimal() {
	do := func(ctx context.Context, arguments []string, out io.Writer) error {
		_, err := fmt.Fprintln(out, arguments)

		return err
	}

	app := lieut.NewSingleCommandApp(
		lieut.AppInfo{Name: "example"},
		do,
		flag.CommandLine,
		os.Stdout,
		os.Stderr,
	)

	exitCode := app.Run(context.Background(), os.Args[1:])

	os.Exit(exitCode)
}
