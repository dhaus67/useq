package useq

import (
	"fmt"
	"go/ast"
	"go/token"
	"go/types"
	"slices"
	"strings"

	"github.com/golangci/plugin-module-register/register"
	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"
	"golang.org/x/tools/go/ast/astutil"
	"golang.org/x/tools/go/ast/inspector"
	"golang.org/x/tools/go/types/typeutil"
)

var _ register.LinterPlugin = (*UseqAnalyzer)(nil)

var (
	DefaultValidationSettings = map[string][]string{
		"fmt": {
			"Printf",
			"Sprintf",
			"Fprintf",
			"Errorf",
		},
		"github.com/pkg/errors": {
			"Errorf",
			"Wrapf",
		},
	}
)

func init() {
	register.Plugin("useq", New)
}

func New(settings any) (register.LinterPlugin, error) {
	s, err := register.DecodeSettings[Settings](settings)
	if err != nil {
		return nil, err
	}

	for pkg, funcs := range DefaultValidationSettings {
		s.Validate[pkg] = funcs
	}

	return &UseqAnalyzer{settings: s}, nil
}

type Settings struct {
	Validate map[string][]string `json:"validate"`
}

type UseqAnalyzer struct {
	settings Settings
}

func (u *UseqAnalyzer) BuildAnalyzers() ([]*analysis.Analyzer, error) {
	return []*analysis.Analyzer{{
		Name:     "useq",
		Doc:      "useq checks for preferring %q over %s as formatting argument when quotation is needed.",
		Requires: []*analysis.Analyzer{inspect.Analyzer},
		Run:      u.run,
	}}, nil
}

func (u *UseqAnalyzer) GetLoadMode() string {
	return register.LoadModeSyntax
}

func (u *UseqAnalyzer) run(pass *analysis.Pass) (interface{}, error) {
	result := pass.ResultOf[inspect.Analyzer].(*inspector.Inspector)

	nodeFilter := []ast.Node{
		(*ast.CallExpr)(nil),
	}

	preOrderFiltered(result, ast.IsGenerated, nodeFilter, func(n ast.Node) {
		u.checkCall(pass, n.(*ast.CallExpr))
	})

	return nil, nil
}

func (u *UseqAnalyzer) checkCall(pass *analysis.Pass, call *ast.CallExpr) {
	fn := astutil.Unparen(call.Fun)

	// Ignore type conversions and builtins.
	if pass.TypesInfo.Types[fn].IsType() || pass.TypesInfo.Types[fn].IsBuiltin() {
		return
	}

	namedFn, _ := typeutil.Callee(pass.TypesInfo, call).(*types.Func)
	if namedFn == nil {
		return
	}

	if u.isWrongFormattingCall(namedFn, call) {
		pass.Reportf(call.Fun.Pos(), "use %%q instead of \"%%s\" for formatting strings with quotations")
	}
}

func (u *UseqAnalyzer) isWrongFormattingCall(fn *types.Func, call *ast.CallExpr) bool {
	sig := fn.Type().(*types.Signature)

	// Check if the function signature is variadic and has the correct number of arguments (i.e. we have a formatting signature).
	if sig == nil || !sig.Variadic() || sig.Params().Len() < 2 {
		return false
	}
	methodsToVerify := u.settings.Validate[fn.Pkg().Path()]
	if methodsToVerify == nil {
		return false
	}

	name := getFunctionName(fn, sig)
	if !slices.Contains(methodsToVerify, name) {
		return false
	}

	// In fmt style APIs, the last argument will be the variadic argument, and the second to last argument will be the format string.
	formatIndex := sig.Params().Len() - 2
	formatArg, ok := call.Args[formatIndex].(*ast.BasicLit)
	if !ok || formatArg.Kind != token.STRING {
		return false
	}

	formatStr := strings.Trim(formatArg.Value, "\"")
	return strings.Contains(formatStr, `\"%s\"`)
}

// preOrderFiltered calls inspector.Preorder but filters out files that do not pass the filter function.
func preOrderFiltered(inspector *inspector.Inspector, filter func(*ast.File) bool, nodeFilter []ast.Node, do func(n ast.Node)) {
	var types []ast.Node
	types = append(types, nodeFilter...)
	types = append(types, (*ast.File)(nil))
	var skip bool
	inspector.Preorder(types, func(n ast.Node) {
		if f, ok := n.(*ast.File); ok {
			skip = filter(f)
			// Store the result, but skip the file since we do not evaluate files, only call expressions.
			return
		}
		if !skip {
			do(n)
		}
	})
}

// getFunctionName gets the function name with the receiver type if it exists.
func getFunctionName(fn *types.Func, sig *types.Signature) string {
	name := fn.Name()

	if sig.Recv() == nil {
		return name
	}
	recvType := sig.Recv().Type()
	qf := types.RelativeTo(fn.Pkg())
	if pointerType, _ := recvType.(*types.Pointer); pointerType != nil {
		name = fmt.Sprintf("(*%s).%s", types.TypeString(pointerType.Elem(), qf), name)
	} else {
		name = fmt.Sprintf("%s.%s", types.TypeString(recvType, qf), name)
	}
	return name
}
