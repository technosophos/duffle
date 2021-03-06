package main

import (
	"errors"
	"fmt"
	"io"
	"path/filepath"

	"github.com/deislabs/duffle/pkg/duffle/home"
	"github.com/deislabs/duffle/pkg/packager"

	"github.com/spf13/cobra"
)

const exportDesc = `
Packages a bundle, invocation images, and all referenced images within a single
gzipped tarfile.

All images specified in the bundle metadata are saved as tar files in the artifacts/
directory along with an artifacts.json file which describes the contents of artifacts/.

By default, this command will use the name and version information of the bundle to create
a compressed archive file called <name>-<version>.tgz in the current directory. This
destination can be updated by specifying a file path to save the compressed bundle to using
the --output-file flag.

If you want to export only the bundle manifest without the invocation images and referenced 
images, use the --thin flag.
`

type exportCmd struct {
	dest     string
	path     string
	home     home.Home
	out      io.Writer
	thin     bool
	verbose  bool
	insecure bool
}

func newExportCmd(w io.Writer) *cobra.Command {
	export := &exportCmd{out: w}

	cmd := &cobra.Command{
		Use:   "export [PATH]",
		Short: "package CNAB bundle in gzipped tar file",
		Long:  exportDesc,
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 {
				return errors.New("this command requires the path to the bundle")
			}
			export.home = home.Home(homePath())
			export.path = args[0]

			return export.run()
		},
	}

	f := cmd.Flags()
	f.StringVarP(&export.dest, "output-file", "o", "", "Save exported bundle to file path")
	f.BoolVarP(&export.thin, "thin", "t", false, "Export only the bundle manifest")
	f.BoolVarP(&export.verbose, "verbose", "v", false, "Verbose output")
	f.BoolVarP(&export.insecure, "insecure", "k", false, "Do not verify the bundle (INSECURE)")

	return cmd
}

func (ex *exportCmd) run() error {
	source, err := filepath.Abs(ex.path)
	if err != nil {
		return err
	}

	l, err := getLoader(ex.insecure)
	if err != nil {
		return err
	}

	exp, err := packager.NewExporter(source, ex.dest, ex.home.Logs(), l, ex.thin, ex.insecure)
	if err != nil {
		return fmt.Errorf("Unable to set up exporter: %s", err)
	}
	if err := exp.Export(); err != nil {
		return err
	}
	if ex.verbose {
		fmt.Fprintf(ex.out, "Export logs: %s\n", exp.Logs)
	}
	return nil
}
