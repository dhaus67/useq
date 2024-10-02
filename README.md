# useq

`useq` is a custom [Go analyzer](https://pkg.go.dev/golang.org/x/tools/go/analysis#hdr-Analyzer) that detects usages of `\"%s\"` in formatting arguments and suggests using `%q` instead.

## Using it with `golangci-lint`

_Note: [v1.1.0](https://github.com/dhaus67/useq/releases/tag/v1.1.0) can be used as module plugin, later version will be added as public linters in `golangci-lint`._

`useq` can be used with `golangci-lint` with the [module plugin system](https://golangci-lint.run/plugins/module-plugins).

Here's the `.custom-gcl.yml`:
```yaml
version: v1.60.1
plugins:
  - module: 'github.com/dhaus67/useq'
    import: 'github.com/dhaus67/useq'
    version: latest
```

Here's a configuration of `useq` in `.golangci.yml`:
```yaml
linters-settings:
  custom:
    useq:
      type: "module"
      description: "Detects usages of `\"%s\"` in `fmt.Sprintf` calls and suggests using `%q` instead."
      settings:
        functions:
          # A list of functions to validate. The function name needs to be the full qualified name (including potential pointers).
          # By default, `useq` validates fmt and github.com/pkg/errors functions.
          #- github.com/custom/package.MyPrintf              # exported package level function.
          #- (github.com/custom/package.MyStruct).MyPrintf   # exported method of a struct.
          #- (*github.com/custom/package.MyStruct).MyPrintf  # exported method of a struct with a pointer receiver.
linters:
  disable-all: true
  enable:
    - useq
```

## Using it standalone

`useq` can also be used as a standalone tool. Here's how to install it:
```bash
go install github.com/dhaus67/useq/cmd/useq@latest
```

You can define a config file with your custom functions you want to validate as well:
```json
{
  "functions": [
    "github.com/custom/package.MyPrintf",              # exported package level function.
    "(github.com/custom/package.MyStruct).MyPrintf",   # exported method of a struct.
    "(*github.com/custom/package.MyStruct).MyPrintf"   # exported method of a struct with a pointer receiver.
  ]
}
```

And pass it as a flag to the linter:
```sh
useq -config=path/to/config/file
```
