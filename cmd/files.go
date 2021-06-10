/*
Copyright © 2021 NAME HERE <EMAIL ADDRESS>

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
	"errors"
	"io/ioutil"
	"log"
	"os"

	"github.com/databus23/helm-diff/v3/diff"
	"github.com/databus23/helm-diff/v3/manifest"
	"github.com/spf13/cobra"
)

type fileDiff struct {
	detailedExitCode   bool
	suppressedKinds    []string
	files              []string
	outputContext      int
	includeTests       bool
	showSecrets        bool
	output             string
	stripTrailingCR    bool
	normalizeManifests bool
	defaultNamespace   string
}

// filesCmd represents the files command
func filesCmd() *cobra.Command {
	/*
		var filesCmd = &cobra.Command{
			Use:   "files",
			Short: "Generate a diff based on two given files",
			Long:  ``,
			Run: func(cmd *cobra.Command, args []string) {
				fmt.Println("files called")
			},
		}
		return filesCmd*/
	diff := fileDiff{}
	filesCmd := &cobra.Command{
		Use:   "files",
		Short: "Show manifest differences",
		Long:  "Generate a diff based on two given files",
		//Alias root command to chart subcommand
		//Args: chartCommand.Args,
		RunE: func(cmd *cobra.Command, args []string) error {

			if len(args) < 2 {
				return errors.New("Too few arguments to Command \"files\".\nMinimum 2 arguments required: file1, file2")
			}
			diff.files = args[0:]

			if q, _ := cmd.Flags().GetBool("suppress-secrets"); q {
				diff.suppressedKinds = append(diff.suppressedKinds, "Secret")
			}
			return diff.differentiateFiles()

		},
	}
	filesCmd.Flags().BoolP("suppress-secrets", "q", false, "suppress secrets in the output")
	filesCmd.Flags().BoolVar(&diff.showSecrets, "show-secrets", false, "do not redact secret values in the output")
	filesCmd.Flags().BoolVar(&diff.detailedExitCode, "detailed-exitcode", false, "return a non-zero exit code when there are changes")
	filesCmd.Flags().StringArrayVar(&diff.suppressedKinds, "suppress", []string{}, "allows suppression of the values listed in the diff output")
	filesCmd.Flags().IntVarP(&diff.outputContext, "context", "C", -1, "output NUM lines of context around changes")
	filesCmd.Flags().BoolVar(&diff.includeTests, "include-tests", false, "enable the diffing of the helm test hooks")
	filesCmd.Flags().StringVar(&diff.output, "output", "diff", "Possible values: diff, simple, template. When set to \"template\", use the env var HELM_DIFF_TPL to specify the template.")
	filesCmd.Flags().BoolVar(&diff.stripTrailingCR, "strip-trailing-cr", false, "strip trailing carriage return on input")
	filesCmd.Flags().BoolVar(&diff.normalizeManifests, "normalize-manifests", false, "normalize manifests before running diff to exclude style differences from the output")
	filesCmd.Flags().StringVar(&diff.defaultNamespace, "default-namespace", "default", "Default namespace to use if no namespace specified in template")

	return filesCmd

}

func (fd *fileDiff) differentiateFiles() error {

	excludes := []string{helm3TestHook, helm2TestSuccessHook}
	if fd.includeTests {
		excludes = []string{}
	}

	content1, err := ioutil.ReadFile(fd.files[0])
	if err != nil {
		return err
	}

	content2, err := ioutil.ReadFile(fd.files[1])
	if err != nil {
		log.Fatal(err)
	}

	diff1 := manifest.Parse(string(content1), fd.defaultNamespace, fd.normalizeManifests, excludes...)
	diff2 := manifest.Parse(string(content2), fd.defaultNamespace, fd.normalizeManifests, excludes...)

	seenAnyChanges := diff.Manifests(diff1, diff2, excludes, fd.showSecrets, fd.outputContext, fd.output, fd.stripTrailingCR, os.Stdout)

	if fd.detailedExitCode && seenAnyChanges {
		return Error{
			error: errors.New("identified at least one change, exiting with non-zero exit code (detailed-exitcode parameter enabled)"),
			Code:  2,
		}
	}

	return nil
}
