// Copyright © 2021 Jason Podgorny.
// License: https://creativecommons.org/licenses/by-nc-sa/4.0/

// The terrafunk command is intended to allow running terragrunt and terraform functions
// From a cli interface without the need for a full terragrunt or terraform executable
package main

// This allows you to quickly run a function and see the outputs
//
// By default the command will output json which can be piped into another
// utility like jq or powershell convertfrom-json.
//
// In addition you can get verbose output in order to see the
// cty Types that Terraform and Terragrunt use for data representation.
// This can be useful for data inspection at a deeper
// level when issues are encountered

// By default the command runs using the following expression:
// 		read_terragrunt_config("terragrunt.hcl")
//
// This function parses a terragrunt config, includes any configs
// that are set to be included, runs through all functions and
// gives an output that shows all configuration values, inputs, and
// locals that will be used for this terragrunt run.   This is great
// for troubleshooting, and you can use this function along with
// scratchpad type hcl files to play with functions quickly, and chain
// together several to allow more than just a single expression to be
// interpreted.

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"

	

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"

	"github.com/gruntwork-io/terragrunt/options"

	"github.com/JasonPodgorny/terrafunk/internal/config"

	ctyjson "github.com/zclconf/go-cty/cty/json"
)

func main() {

	// Set Up Logging
	infoLog := log.New(os.Stdout, "INFO\t", log.Ldate|log.Ltime)
	errorLog := log.New(os.Stderr, "ERROR\t", log.Ldate|log.Ltime)

	// Set Flag Arguments And Parse Inputs
	expression := flag.String("expression", "read_terragrunt_config(\"terragrunt.hcl\")", "Terraform Expression To Run")
	workdir := flag.String("workdir", ".", "Working Directory For Expression")
	verbose := flag.Bool("verbose", false, "Verbose Outputs")
	flag.Parse()

	// Get Leftover Arguments After Flag Parsing.
	extraArgs := flag.Args()

	// If There Are Extra Arguments Beyond Flags, Inputs Were Formatted Improperly
	// Print Usage/Defaults And Exit
	if len(extraArgs) > 0 {
		flag.PrintDefaults()
		os.Exit(1)
	}

	// Log Working Directory If Output Is Verbose
	if *verbose {
		infoLog.Printf("Workdir: \n%s\n\n", *workdir)
	}

	// Log Expression If Output Is Verbose
	if *verbose {
		infoLog.Printf("Expression: \n%s\n\n", *expression)
	}

	// Add Trailing Slash To Working Directory
	*workdir = *workdir + "\\"

	// Set Up Default Set Of Terragrunt Options
	// If Errors, Log And Exit
	terragruntOptions, err := options.NewTerragruntOptions(*workdir)
	if err != nil {
		errorLog.Fatalf("Get Options Had The Following Errors: %s", err)
	}

	// Parse Environment Variables And Add To Terragrunt Options
	terragruntOptions.Env = parseEnvironmentVariables(os.Environ())

	// Blank Extensions
	extensions := config.EvalContextExtensions{}

	// Generate HCL Eval Context
	terragruntEvalCtx, err := config.CreateTerragruntEvalContext(*workdir, terragruntOptions, extensions)
	if err != nil {
		errorLog.Printf("Create HCL Eval Context Had The Following Errors: %s", err)
	}

	// Parse HCL Expression
	expr, parseDiags := hclsyntax.ParseExpression([]byte(*expression), "", hcl.Pos{Line: 1, Column: 1, Byte: 0})

	// Get Value From Expression Using HCL Eval Context
	ctyValue, valDiags := expr.Value(terragruntEvalCtx)

	// See If There Were Output Diagnostics From Prior Routines
	diagCount := len(parseDiags) + len(valDiags)

	// If There Were Output Diagnostics Log These Errors And Exit
	if diagCount != 0 {
		errorLog.Printf("wrong number of diagnostics %d; want %d \n", diagCount, 0)
		for _, diag := range parseDiags {
			errorLog.Printf("ParseDiag - %s \n", diag.Error())
		}
		for _, diag := range valDiags {
			errorLog.Printf("ValDiag - %s \n", diag.Error())
		}
		os.Exit(1)
	}

	// Log Cty Value If Output Is Verbose
	if *verbose {
		infoLog.Printf("Got Value (Cty): \n%#v\n\n", ctyValue)
	}

	// Get Cty Type Of Value
	ctyType := ctyValue.Type()

	// Marshal Value Into Byte Slice With Json Formatting
	jsonBytes, err := ctyjson.Marshal(ctyValue, ctyType)
	if err != nil {
		errorLog.Printf("JSON Marshal Had The Following Errors: %s", err)
	}

	// Run Through Indent Function To Make Json Look Nice
	jsonPretty := &bytes.Buffer{}
	err = json.Indent(jsonPretty, jsonBytes, "", "  ")
	if err != nil {
		errorLog.Printf("JSON Indent Had The Following Errors: %s", err)
	}

	// If Verbose, Log Json Output Using Logger
	// If Not Verbose, Just Print Output - This Will Be Only Output
	// In These Cases And Can Be Piped Into Another Utility To
	// Parse Json Like jq or Powershell ConvertFrom-Json
	if *verbose {
		infoLog.Printf("Got String: \n%s", jsonPretty.String())
	} else {
		fmt.Printf("%s", jsonPretty.String())
	}

}

// Got This Function From Terragrunt Options. It Was Not Exported So Added To This Package
func parseEnvironmentVariables(environment []string) map[string]string {
	environmentMap := make(map[string]string)

	for i := 0; i < len(environment); i++ {
		variableSplit := strings.SplitN(environment[i], "=", 2)

		if len(variableSplit) == 2 {
			environmentMap[strings.TrimSpace(variableSplit[0])] = variableSplit[1]
		}
	}

	return environmentMap
}
