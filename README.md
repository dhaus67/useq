# useq

`useq` is a custom [Go analyzer](https://pkg.go.dev/golang.org/x/tools/go/analysis#hdr-Analyzer) that detects usages of `\"%s\"` in `fmt.Sprintf` calls and suggests using `%q` instead.

## Using it with `golangci-lint`

`useq` can be used with `golangci-lint` with the [module plugin system](https://golangci-lint.run/plugins/module-plugins).

Here's the `.custom-gcl.yml`:
```yaml
version: v1.63.0
plugins:
  - module: 'github.com/dhaus67/useq'
    import: 'github.com/dhaus67/useq/useq'
    version: v1.0.0
```

Here's a configuration of `useq` in `.golangci.yml`:
```yaml
linters-settings:
  custom:
    useq:
      type: "module"
      description: "Detects usages of `\"%s\"` in `fmt.Sprintf` calls and suggests using `%q` instead."
      settings:
        validate:
          # A list of packages and functions which accept string formatting to validate.
          # By default, `useq` validates fmt and github.com/pkg/errors functions.
          "github.com/custom/package":
            - "MyPrintf"

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

Currently, the standalone function does not support reading in a configuration file. It will only validate `fmt.Sprintf` and `github.com/pkg/errors` functions.
If you want more customizations, you can use `useq` with `golangci-lint` as described above.
