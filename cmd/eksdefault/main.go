package main

import (
	"bytes"
	"errors"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"text/tabwriter"

	"github.com/peterbueschel/eksdefault"
	"github.com/urfave/cli"
)

var (
	addFlags = []cli.Flag{
		cli.StringFlag{
			Name:  "user, u",
			Usage: "Name of the user to be added to the new context.",
		},
		cli.StringFlag{
			Name:  "namespace, n",
			Usage: "Namespace to be added to the new context.",
		},
		cli.StringFlag{
			Name:  "cluster, c",
			Usage: "Name of the cluster to be added to the new context.",
		},
		cli.StringFlag{
			Name:  "profile, p",
			Usage: "Name of the aws profile to be added to the new context.",
		},
	}
	output = ""

	missingNewContextName = errors.New(
		"the name of the new context is required and (optional) the user " +
			"or the namespace or the cluster or the profile",
	)
	duplicatedContextName = errors.New(
		"a context with the same name already exists in the kube config",
	)
)

// idToName is a helper function to add id vs. context name support. Means, you can
// either use the ID or the context name.
func idToName(str string, file *eksdefault.KubeConfig) (string, error) {
	if id, err := strconv.Atoi(str); err == nil {
		if len(file.Contexts) < id {
			return "", fmt.Errorf("The ID '%d' does not exists. Run 'eksdefault ls' to get the correct IDs", id)
		}
		return file.Contexts[id].Name, nil
	}
	return str, nil
}

func namedExists(name string, file *eksdefault.KubeConfig) bool {
	for _, n := range file.GetContextNames() {
		if n == name {
			return true
		}
	}
	return false
}

// printTabbed is a helper function and prints out a simple table.
func printTabbed(header []string, data [][]string) error {
	var b bytes.Buffer
	padding := len(header)
	w := tabwriter.NewWriter(&b, 0, 0, padding, ' ', 0)

	cols := strings.Join(header, "\t")
	fmt.Fprintln(w, cols)
	for _, row := range data {
		cols := strings.Join(row, "\t")
		fmt.Fprintln(w, cols)
	}
	if err := w.Flush(); err != nil {
		return err
	}
	output = b.String()
	return nil
}

// unsetDefaultContext empties the entry for the 'current-context' inside the kube config,
// which at the end disable a default context.
func unsetDefaultContext(file *eksdefault.KubeConfig) *cli.Command {
	return &cli.Command{
		Name:    "unset",
		Aliases: []string{"rm", "del"},
		Usage:   "'unset': Deletes the current-context entry by setting 'current-context: \"\" '.",
		Action: func(c *cli.Context) error {
			return file.UnSetDefault()
		},
	}
}

// getCurrentContext returns the current-context.
func getCurrentContext(file *eksdefault.KubeConfig) *cli.Command {
	return &cli.Command{
		Name:    "is",
		Aliases: []string{"current"},
		Usage:   "'is': Prints the current-context.",
		Action: func(c *cli.Context) error {
			if cc := file.CurrentContext; len(cc) > 0 {
				output = fmt.Sprintf("%v\n", file.CurrentContext)
				return nil
			}
			output = "no current-context set\n"
			return nil
		},
	}
}

// setDefaultContext adds or changes the current-context inside the kube config and also
// sets the default AWS profile inside the .aws/credentials file configured by the 'aws-profile'
// line inside the kube config.
func setDefaultContext(file *eksdefault.KubeConfig) *cli.Command {
	return &cli.Command{
		Name:    "set",
		Aliases: []string{"to", "use", "use-current"},
		Usage:   "'set <context>': Changes the current-context to the given context name.",
		Action: func(c *cli.Context) error {
			if c.NArg() < 1 {
				return fmt.Errorf(
					"the ID or name of an existing context is required",
				)
			}
			name, err := idToName(c.Args().First(), file)
			if err != nil {
				return err
			}
			if err := file.SetContextTo(name); err != nil {
				if err == eksdefault.NoProfilSet {
					return fmt.Errorf("%v.\nYou can run also 'eksdefault profile <aws profile> %s' to set an AWS profile for this context", err, c.Args().First())
				}
				return fmt.Errorf("%v", err)
			}
			return nil
		},
	}
}

// useProfile adds/changes AWS profiles to the current context.
func useProfile(file *eksdefault.KubeConfig) *cli.Command {
	return &cli.Command{
		Name:    "profile",
		Aliases: []string{"p", "use-profile", "pr"},
		Usage:   "'n <aws profile> [<context>]': Changes the aws profile of a given context. If no context was given the current one will be used.",
		Action: func(c *cli.Context) error {
			if c.NArg() < 1 {
				return fmt.Errorf(
					"the name of an existing aws profile is required and (optional the context)",
				)
			}
			if c.NArg() > 1 {
				name, err := idToName(c.Args().Get(1), file)
				if err != nil {
					return err
				}
				return file.AddProfileTo(name, c.Args().First())
			}
			return file.AddProfileTo(file.CurrentContext, c.Args().First())
		},
	}
}

// useNamespace adds/changes the namespace entry inside the current context.
func useNamespace(file *eksdefault.KubeConfig) *cli.Command {
	return &cli.Command{
		Name:    "namespace",
		Aliases: []string{"n", "use-namespace", "ns"},
		Usage:   "'n <namespace> [<context>|<ID>]': Changes the namespace of a given context. If no context was given the current one will be used.",
		Action: func(c *cli.Context) error {
			if c.NArg() < 1 {
				return fmt.Errorf(
					"the name of a namespace is required and (optional the context or its ID)",
				)
			}
			if c.NArg() > 1 {
				name, err := idToName(c.Args().Get(1), file)
				if err != nil {
					return err
				}
				return file.AddNamespaceTo(name, c.Args().First())
			}
			return file.AddNamespaceTo(file.CurrentContext, c.Args().First())
		},
	}
}

// copyContext copies the current context into a new one. Flags for the namespace, user, profile and cluster
// let you customize the new copy.
func copyContext(file *eksdefault.KubeConfig) *cli.Command {
	overwrite := func(newer, current string) string {
		if len(newer) > 0 {
			return newer
		}
		return current
	}

	return &cli.Command{
		Name:    "copy",
		Aliases: []string{"cp", "copy-current"},
		Usage: "'copy <new context name> [-u <user name>|-n <namespace>|-c <cluster name>|-p <aws profile>]':" +
			" Copy the current context as a new context to the kube config file and overwrites the (if given) " +
			"the specific settings for user, cluster, namespace or profile.",
		Flags: addFlags,
		Action: func(c *cli.Context) error {
			if c.NArg() < 1 {
				return missingNewContextName
			}

			name := c.Args().First()
			if namedExists(name, file) {
				return duplicatedContextName
			}
			cc, _, err := file.GetContextBy(file.CurrentContext)
			if err != nil {
				return err
			}

			ctx := eksdefault.KubeContext{
				Name:       name,
				AWSprofile: overwrite(c.String("profile"), cc.AWSprofile),
				Context: &eksdefault.Context{
					User:      overwrite(c.String("user"), cc.Context.User),
					Namespace: overwrite(c.String("namespace"), cc.Context.Namespace),
					Cluster:   overwrite(c.String("cluster"), cc.Context.Cluster),
				},
			}
			file.Contexts = append(file.Contexts, ctx)
			return file.SaveContexts()
		},
	}
}

// addContext creates a new context customized by its flags like user or namespace.
func addContext(file *eksdefault.KubeConfig) *cli.Command {
	return &cli.Command{
		Name:    "new",
		Aliases: []string{"add", "new-context", "nc"},
		Usage: "'new <context name> [-u <user name>|-n <namespace>|-c <cluster name>|-p <aws profile>]':" +
			" Adds a new context to the kube config file.",
		Flags: addFlags,
		Action: func(c *cli.Context) error {
			if c.NArg() < 1 {
				return missingNewContextName
			}
			name := c.Args().First()
			if namedExists(name, file) {
				return duplicatedContextName
			}
			ctx := eksdefault.KubeContext{
				Name:       name,
				AWSprofile: c.String("profile"),
				Context: &eksdefault.Context{
					User:      c.String("user"),
					Namespace: c.String("namespace"),
					Cluster:   c.String("cluster"),
				},
			}
			file.Contexts = append(file.Contexts, ctx)
			return file.SaveContexts()
		},
	}
}

// getContexts prints the available contexts either as a list or as a table.
func getContexts(file *eksdefault.KubeConfig) *cli.Command {
	return &cli.Command{
		Name:    "list",
		Aliases: []string{"ls", "contexts"},
		Flags: []cli.Flag{cli.BoolFlag{
			Name:   "short, s",
			Usage:  "Prints only the context names as list.",
			EnvVar: "EKSDEFAULT_SHORT_INFO",
		}},
		Usage: "Returns all available contexts from the kube config file.",
		Action: func(c *cli.Context) error {
			tbl := [][]string{}
			if !c.Bool("short") {
				curr := file.CurrentContext
				for idx, c := range file.Contexts {
					p := c.AWSprofile
					cc := ""
					if c.Name == curr {
						cc = "*"
					}
					row := []string{
						fmt.Sprintf("%d", idx),
						cc,
						c.Name,
						p,
						c.Context.Cluster,
						c.Context.User,
						c.Context.Namespace,
					}
					tbl = append(tbl, row)
				}
				return printTabbed(
					[]string{
						"ID",
						"CURRENT",
						"KUBE CONTEXT",
						"AWS PROFILE",
						"CLUSTER",
						"USER",
						"NAMESPACE"},
					tbl,
				)
			}
			for _, c := range file.GetContextNames() {
				output += fmt.Sprintf("%v\n", c)
			}
			return nil
		},
	}
}

func runMain(args []string) (string, error) {
	output = ""
	file, err := eksdefault.GetConfigFile()
	if err != nil {
		return output, err
	}
	app := cli.NewApp()
	app.Commands = []cli.Command{
		*getCurrentContext(file),
		*copyContext(file),
		*addContext(file),
		*unsetDefaultContext(file),
		*useNamespace(file),
		*useProfile(file),
		*getContexts(file),
		*setDefaultContext(file),
	}
	return output, app.Run(args)
}

func main() {
	out, err := runMain(os.Args)
	if err != nil {
		log.Fatalf("[EKSDEFAULT][ERROR] %v.\n", err)
	}
	fmt.Printf("%+v", out)
}
