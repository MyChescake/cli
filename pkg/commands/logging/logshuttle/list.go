package logshuttle

import (
	"fmt"
	"io"

	"github.com/fastly/cli/pkg/cmd"
	"github.com/fastly/cli/pkg/commands/compute/manifest"
	"github.com/fastly/cli/pkg/config"
	"github.com/fastly/cli/pkg/errors"
	"github.com/fastly/cli/pkg/text"
	"github.com/fastly/go-fastly/v5/fastly"
)

// ListCommand calls the Fastly API to list Logshuttle logging endpoints.
type ListCommand struct {
	cmd.Base
	manifest       manifest.Data
	Input          fastly.ListLogshuttlesInput
	serviceVersion cmd.OptionalServiceVersion
}

// NewListCommand returns a usable command registered under the parent.
func NewListCommand(parent cmd.Registerer, globals *config.Data, data manifest.Data) *ListCommand {
	var c ListCommand
	c.Globals = globals
	c.manifest = data
	c.CmdClause = parent.Command("list", "List Logshuttle endpoints on a Fastly service version")
	c.RegisterServiceIDFlag(&c.manifest.Flag.ServiceID)
	c.RegisterServiceVersionFlag(cmd.ServiceVersionFlagOpts{
		Dst: &c.serviceVersion.Value,
	})
	return &c
}

// Exec invokes the application logic for the command.
func (c *ListCommand) Exec(in io.Reader, out io.Writer) error {
	serviceID, serviceVersion, err := cmd.ServiceDetails(cmd.ServiceDetailsOpts{
		AllowActiveLocked:  true,
		Client:             c.Globals.Client,
		Manifest:           c.manifest,
		Out:                out,
		ServiceVersionFlag: c.serviceVersion,
		VerboseMode:        c.Globals.Flag.Verbose,
	})
	if err != nil {
		c.Globals.ErrLog.AddWithContext(err, map[string]interface{}{
			"Service ID":      serviceID,
			"Service Version": errors.ServiceVersion(serviceVersion),
		})
		return err
	}

	c.Input.ServiceID = serviceID
	c.Input.ServiceVersion = serviceVersion.Number

	logshuttles, err := c.Globals.Client.ListLogshuttles(&c.Input)
	if err != nil {
		c.Globals.ErrLog.Add(err)
		return err
	}

	if !c.Globals.Verbose() {
		tw := text.NewTable(out)
		tw.AddHeader("SERVICE", "VERSION", "NAME")
		for _, logshuttle := range logshuttles {
			tw.AddLine(logshuttle.ServiceID, logshuttle.ServiceVersion, logshuttle.Name)
		}
		tw.Print()
		return nil
	}

	fmt.Fprintf(out, "Version: %d\n", c.Input.ServiceVersion)
	for i, logshuttle := range logshuttles {
		fmt.Fprintf(out, "\tLogshuttle %d/%d\n", i+1, len(logshuttles))
		fmt.Fprintf(out, "\t\tService ID: %s\n", logshuttle.ServiceID)
		fmt.Fprintf(out, "\t\tVersion: %d\n", logshuttle.ServiceVersion)
		fmt.Fprintf(out, "\t\tName: %s\n", logshuttle.Name)
		fmt.Fprintf(out, "\t\tURL: %s\n", logshuttle.URL)
		fmt.Fprintf(out, "\t\tToken: %s\n", logshuttle.Token)
		fmt.Fprintf(out, "\t\tFormat: %s\n", logshuttle.Format)
		fmt.Fprintf(out, "\t\tFormat version: %d\n", logshuttle.FormatVersion)
		fmt.Fprintf(out, "\t\tResponse condition: %s\n", logshuttle.ResponseCondition)
		fmt.Fprintf(out, "\t\tPlacement: %s\n", logshuttle.Placement)
	}
	fmt.Fprintln(out)

	return nil
}
