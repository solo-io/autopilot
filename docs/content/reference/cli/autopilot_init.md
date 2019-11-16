---
title: "autopilot init"
weight: 5
---
## autopilot init

Initialize a new project for the given top-level CRD

### Synopsis

The autopilot init command creates a project skeleton in the given directory. 
If the directory does not exist, it will be created. 


```
autopilot init <dir> --kind=<kind> --group=<apigroup> --verison=<apiversion> [--skip-gomod] [flags]
```

### Options

```
      --group string         API Group for the Top-Level CRD (default "example.io")
  -h, --help                 help for init
      --kind string          Kind (Camel-Cased Name) of Top-Level CRD (default "Example")
  -m, --module go mod init   Sets the name of the module for go mod init.Required if initializing outside your $GOPATH
  -s, --skip-gomod           skip generating go.mod for project
      --version string       API Version for the Top-Level CRD (default "v1")
```

### Options inherited from parent commands

```
  -v, --verbose   Enable verbose logging
```

### SEE ALSO

* [autopilot](../autopilot)	 - An SDK for building Service Mesh Operators with ease

