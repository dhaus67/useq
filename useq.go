package useq

import (
	"fmt"
	"go/ast"
	"go/token"
	"go/types"
	"slices"
	"strings"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"
	"golang.org/x/tools/go/ast/astutil"
	"golang.org/x/tools/go/ast/inspector"
	"golang.org/x/tools/go/types/typeutil"
)

type functionsPerPackage map[string][]string

var (
	defaultFunctionsPerPackage = functionsPerPackage{
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

// NewAnalyzer creates a new analyzer with the given settings.
func NewAnalyzer(settings Settings) (*analysis.Analyzer, error) {
	u := &UseqAnalyzer{settings: settings}
	u.Analyzer = &analysis.Analyzer{
		Name:     "useq",
		Doc:      "useq checks for preferring %q over %s as formatting argument when quotation is needed.",
		Requires: []*analysis.Analyzer{inspect.Analyzer},
		Run:      u.runAnalyzer,
	}
	if err := u.Compile(); err != nil {
		return nil, err
	}
	return u.Analyzer, nil
}

// Settings holds all the settings for the UseqAnalyzer.
type Settings struct {
	// The functions are fully qualified function names including the package (e.g. fmt.Printf).
	Functions []string `json:"functions"`
	// FunctionsPerPackage is a map of package names to the functions that should be checked.
	FunctionsPerPackage functionsPerPackage
}

// UseqAnalyzer is the main struct for the linter plugin.
type UseqAnalyzer struct {
	settings Settings
	Analyzer *analysis.Analyzer
}

// Compile will compile the settings for the analyzer.
func (u *UseqAnalyzer) Compile() error {
	compiledFunctions, err := u.compileSettings(u.settings)
	if err != nil {
		return err
	}
	u.settings.FunctionsPerPackage = compiledFunctions
	return nil
}

func (u *UseqAnalyzer) runAnalyzer(pass *analysis.Pass) (interface{}, error) {
	return run(pass, u.settings.FunctionsPerPackage)
}

func (u *UseqAnalyzer) compileSettings(settings Settings) (functionsPerPackage, error) {
	funcs := make(functionsPerPackage)
	for k, v := range defaultFunctionsPerPackage {
		funcs[k] = slices.Clone(v)
	}

	for _, fn := range settings.Functions {
		lastDotIndex := strings.LastIndex(fn, ".")
		if lastDotIndex == -1 {
			return nil, fmt.Errorf("invalid function name: %s", fn)
		}
		parts := []string{fn[:lastDotIndex], fn[lastDotIndex+1:]}
		if !slices.Contains(funcs[parts[0]], parts[1]) {
			funcs[parts[0]] = append(funcs[parts[0]], parts[1])
		}
	}
	return funcs, nil
}

func run(pass *analysis.Pass, funcs functionsPerPackage) (interface{}, error) {
	result := pass.ResultOf[inspect.Analyzer].(*inspector.Inspector)

	nodeFilter := []ast.Node{
		(*ast.CallExpr)(nil),
	}

	preOrderFiltered(result, ast.IsGenerated, nodeFilter, func(n ast.Node) {
		checkCall(pass, n.(*ast.CallExpr), funcs)
	})

	return nil, nil
}

func checkCall(pass *analysis.Pass, call *ast.CallExpr, funcs functionsPerPackage) {
	fn := astutil.Unparen(call.Fun)

	// Ignore type conversions and builtins.
	if pass.TypesInfo.Types[fn].IsType() || pass.TypesInfo.Types[fn].IsBuiltin() {
		return
	}

	namedFn, _ := typeutil.Callee(pass.TypesInfo, call).(*types.Func)
	if namedFn == nil {
		return
	}

	if wrong, fix := isWrongFormattingCall(namedFn, call, funcs); wrong {
		pass.Report(analysis.Diagnostic{
			Pos:            call.Fun.Pos(),
			End:            call.Fun.End(),
			Message:        "use %q instead of \"%s\" for formatting strings with quotations",
			SuggestedFixes: []analysis.SuggestedFix{fix},
		})
	}
}

func isWrongFormattingCall(fn *types.Func, call *ast.CallExpr, funcs functionsPerPackage) (bool, analysis.SuggestedFix) {
	sig := fn.Type().(*types.Signature)

	// Check if the function signature is variadic and has the correct number of arguments (i.e. we have a formatting signature).
	if sig == nil || !sig.Variadic() || sig.Params().Len() < 2 {
		return false, analysis.SuggestedFix{}
	}
	methodsToVerify := funcs[fn.Pkg().Path()]
	if methodsToVerify == nil {
		return false, analysis.SuggestedFix{}
	}

	name := getFunctionName(fn, sig)
	if !slices.Contains(methodsToVerify, name) {
		return false, analysis.SuggestedFix{}
	}

	// In fmt style APIs, the last argument will be the variadic argument, and the second to last argument will be the format string.
	formatIndex := sig.Params().Len() - 2
	formatArg, ok := call.Args[formatIndex].(*ast.BasicLit)
	if !ok || formatArg.Kind != token.STRING {
		return false, analysis.SuggestedFix{}
	}

	formatStr := strings.Trim(formatArg.Value, "\"")
	needsFix := strings.Contains(formatStr, `\"%s\"`)
	if !needsFix {
		return false, analysis.SuggestedFix{}
	}

	newFormatArgStr := strings.ReplaceAll(formatArg.Value, `\"%s\"`, `%q`)
	return true, analysis.SuggestedFix{
		Message: "replace \"%s\" with %q for formatting strings with quotations",
		TextEdits: []analysis.TextEdit{
			{
				Pos:     formatArg.Pos(),
				End:     formatArg.End(),
				NewText: []byte(newFormatArgStr),
			},
		},
	}
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
