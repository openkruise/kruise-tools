/*
Copyright 2021 The Kruise Authors.
Copyright 2017 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/spf13/cobra/doc"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	cmdutil "k8s.io/kubectl/pkg/cmd/util"
)

var (
	defaultDocDir = "docs"
	yamlDir       = "yaml"
)

func NewCmdGenerateDocs(f cmdutil.Factory, streams genericclioptions.IOStreams) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "generate-docs",
		Short: "Generate documentation for kruise",
		Long:  "Generate documentation for kruise",
		Run: func(cmd *cobra.Command, args []string) {
			cmdutil.CheckErr(generateDocs(cmd))
		},
		Hidden: true,
	}

	cmd.Flags().String("directory", "", "The directory to write the generated docs to")
	return cmd

}

func generateDocs(cmd *cobra.Command) error {
	directory, err := cmd.Flags().GetString("directory")
	if err != nil {
		return err
	}
	if directory == "" {
		directory = defaultDocDir
	}
	// Remove the commands from the root command tree
	removeCmds(cmd.Root(), []string{"exec", "apply", "wait", "diff", "options", "help", "api-resources", "api-versions", "patch", "plugin", "scale", "replace", "options", "kustomize", "version", "config", "completion"})

	err = os.MkdirAll(directory, os.ModePerm)
	if err != nil {
		return err
	}
	err = os.MkdirAll(filepath.Join(directory, yamlDir), os.ModePerm)
	if err != nil {
		return err
	}
	err = doc.GenMarkdownTree(cmd.Root(), directory)
	if err != nil {
		return err
	}
	err = doc.GenYamlTree(cmd.Root(), filepath.Join(directory, yamlDir))
	if err != nil {
		return err
	}
	fmt.Println("documentation generated successfully")
	return nil
}

func removeCmds(rootCmd *cobra.Command, cmdsToRemove []string) {
	// Keep track of command names to detect duplicates
	encountered := make(map[string]bool)

	for _, cmdName := range cmdsToRemove {
		for _, cmd := range rootCmd.Commands() {
			if cmd.Name() == cmdName {
				rootCmd.RemoveCommand(cmd)
			}
		}
	}

	// Remove duplicates
	var uniqueCmds []*cobra.Command
	for _, cmd := range rootCmd.Commands() {
		if encountered[cmd.Name()] {
			continue
		}
		encountered[cmd.Name()] = true
		uniqueCmds = append(uniqueCmds, cmd)
	}
	rootCmd.ResetCommands()
	rootCmd.AddCommand(uniqueCmds...)
}
