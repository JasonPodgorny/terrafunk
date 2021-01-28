# terrafunk

## Overview
Terrafunk is a utility that allows you to execute terraform and terragrunt functions from the command line for quick tests of functions without the need to perform a full terraform or terragrunt run.

This can be especially useful to test terragrunt functions as they are not available to be tested in the `terraform console` sub command.

Another stand out capability in the terragrunt world is the ability to execute the read_terragrunt_config() function which allows fast parsing of terragrunt configs to view interpreted variables and inputs.

In addition you can use the `-verbose` flag to view the underlying cty types and values, which is what terraform and terragrunt use for their type system.   This can be useful to debug why a data format may struggle to covert to certain types, or show why types are unable to be compared.

# Usage

```
Usage of terrafunk:
  -expression string
        Terraform Expression To Run (default "read_terragrunt_config(\"terragrunt.hcl\")")
  -verbose
        Verbose Outputs
  -workdir string
        Working Directory For Expression (default ".")
```

# Default Behavior

By default the command will run the read_terragrunt_config() function in the current working directory.   This reads in the terragrunt configuration, reads in any configurations that are included, and runs through all functions.   It outputs all inputs, and readily available terragunt variables in json format for easy parsing with downstream tools like jq or Powershell ConvertFrom-Json.
 

## jq Example
```
PS C:\terragrunt_directory> .\terrafunk.exe | jq .inputs.commons.locationShort

{
  "centralus": "cus",
  "eastus2": "eu2",
  "northeurope": "neu",
  "westeurope": "weu"
}
```

# Custom Expression Examples

    Note - These Examples Used Powershell.   Double quotes inside of the expression need to be escaped much like the examples for state imports in terraform.   For Linux shells this isn't necessary.

## Terragrunt get_platform() function:

```
PS C:\temp> .\terrafunk.exe -expression 'get_platform()'
"windows"
```

## Terragrunt get_platform() function - (verbose - cty values):
```
PS C:\temp> .\terrafunk.exe -expression 'get_platform()' -verbose
INFO    2021/01/28 00:54:16 Workdir:
.

INFO    2021/01/28 00:54:16 Expression:
get_platform()

INFO    2021/01/28 00:54:16 Got Value (Cty):
cty.StringVal("windows")

INFO    2021/01/28 00:54:16 Got String:
"windows"
```

## Terraform set_union() function:
```
PS C:\temp> .\terrafunk.exe -expression 'setunion([\"terrafunk\", \"is\"], [\"awesome\"])'
[
  "awesome",
  "is",
  "terrafunk"
]
```

## Terraform set_union() with reverse() function:
```
PS C:\temp> .\terrafunk.exe -expression 'reverse(setunion([\"terrafunk\", \"is\"], [\"awesome\"]))'
[
  "terrafunk",
  "is",
  "awesome"
]
```

## Terraform set_union() with reverse() function - (verbose - cty values):
```
PS C:\temp> .\terrafunk.exe -expression 'reverse(setunion([\"terrafunk\", \"is\"], [\"awesome\"]))' -verbose
INFO    2021/01/28 00:57:42 Workdir:
.

INFO    2021/01/28 00:57:42 Expression:
reverse(setunion(["terrafunk", "is"], ["awesome"]))

INFO    2021/01/28 00:57:42 Got Value (Cty):
cty.ListVal([]cty.Value{cty.StringVal("terrafunk"), cty.StringVal("is"), cty.StringVal("awesome")})

INFO    2021/01/28 00:57:42 Got String:
[
  "terrafunk",
  "is",
  "awesome"
]
```

