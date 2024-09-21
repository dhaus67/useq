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

var (
	defaultFunctionsPerPackage = map[string][]string{
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

// New will create a new instance of the UseqAnalyzer.
func New(settings Settings) (*UseqAnalyzer, error) {
	u := &UseqAnalyzer{settings: settings}
	if err := u.Compile(); err != nil {
		return nil, err
	}
	return u, nil
}

// NewWithoutCompile will create a new instance of the UseqAnalyzer without compiling the settings.
// Note: the settings should be compiled before running the linter plugin.
func NewWithoutCompile(settings Settings) *UseqAnalyzer {
	return &UseqAnalyzer{settings: settings}
}

// Settings holds all the settings for the UseqAnalyzer.
type Settings struct {
	// The functions are fully qualified function names including the package (e.g. fmt.Printf).
	Functions []string `json:"functions"`
	// FunctionsPerPackage is a map of package names to the functions that should be checked.
	FunctionsPerPackage map[string][]string
}

// UseqAnalyzer is the main struct for the linter plugin.
type UseqAnalyzer struct {
	settings Settings
}

// Compile will compile the settings for the analyzer.
func (u *UseqAnalyzer) Compile() error {
	u.settings.FunctionsPerPackage = defaultFunctionsPerPackage
	for _, fn := range u.settings.Functions {
		lastDotIndex := strings.LastIndex(fn, ".")
		if lastDotIndex == -1 {
			return fmt.Errorf("invalid function name: %s", fn)
		}
		parts := []string{fn[:lastDotIndex], fn[lastDotIndex+1:]}
		if !slices.Contains(u.settings.FunctionsPerPackage[parts[0]], parts[1]) {
			u.settings.FunctionsPerPackage[parts[0]] = append(u.settings.FunctionsPerPackage[parts[0]], parts[1])
		}
	}
	return nil
}

// Analyzer returns the analyis.Analyzer that will be run.
func (u *UseqAnalyzer) Analyzer() *analysis.Analyzer {
	return &analysis.Analyzer{
		Name:     "useq",
		Doc:      "useq checks for preferring %q over %s as formatting argument when quotation is needed.",
		Requires: []*analysis.Analyzer{inspect.Analyzer},
		Run:      u.run,
	}
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
	methodsToVerify := u.settings.FunctionsPerPackage[fn.Pkg().Path()]
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
