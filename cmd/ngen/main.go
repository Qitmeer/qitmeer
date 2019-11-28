// Copyright 2017-2018 The qitmeer developers

/*
 	Generates Json marshaling methods for the qitmeer's go struct types.
	Inspired by github.com/fjl/gencodec & github.com/garslo/gogen
	TODO :
	1. remove fjl&garslo abstract and use AST directly
	2. ./ngen -dir ../../common/types -type Genesis -field-override genesisJSON -out ../../common/types/gen-genesis.go

*/

package main

import (
	"bytes"
	"flag"
	"fmt"
	"go/ast"
	"go/build"
	"go/importer"
	"go/printer"
	"go/token"
	"go/types"
	"golang.org/x/tools/go/buildutil"
	"golang.org/x/tools/go/loader"
	"golang.org/x/tools/imports"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"strconv"
	"strings"
	//"math/big"
)

// Configuration set most from command-line
type Config struct {
	Dir           string // input package directory
	Type          string // type to generate methods for
	FieldOverride string // name of struct type for field overrides
}

// marshalerType represents the intermediate struct type used during marshaling.
// This is the input data to all the Go code templates.
type marshalerType struct {
	name     string
	Fields   []*marshalerField
	fs       *token.FileSet
	orig     *types.Named
	override *types.Named
	scope    *fileScope
}

// loadOverrides sets field types of the intermediate marshaling type from
// matching fields of otyp.
func (mtyp *marshalerType) loadOverrides(otyp *types.Named) error {
	s := otyp.Underlying().(*types.Struct)
	for i := 0; i < s.NumFields(); i++ {
		of := s.Field(i)
		if of.Anonymous() || !of.Exported() {
			return fmt.Errorf("%v: field override type cannot have embedded or unexported fields", mtyp.fs.Position(of.Pos()))
		}
		f := mtyp.fieldByName(of.Name())
		if f == nil {
			// field not defined in original type, check if it maps to a suitable function and add it as an override
			if fun, retType := findFunction(mtyp.orig, of.Name(), of.Type()); fun != nil {
				f = &marshalerField{name: fun.Name(), origTyp: retType, typ: of.Type(), function: fun, tag: s.Tag(i)}
				mtyp.Fields = append(mtyp.Fields, f)
			} else {
				return fmt.Errorf("%v: no matching field or function for %s in original type %s", mtyp.fs.Position(of.Pos()), of.Name(), mtyp.name)
			}
		}
		if err := checkConvertible(of.Type(), f.origTyp); err != nil {
			return fmt.Errorf("%v: invalid field override: %v", mtyp.fs.Position(of.Pos()), err)
		}
		f.typ = of.Type()
	}
	mtyp.scope.addReferences(s)
	mtyp.override = otyp
	return nil
}

// findFunction returns a function with `name` that accepts no arguments
// and returns a single value that is convertible to the given to type.
func findFunction(typ *types.Named, name string, to types.Type) (*types.Func, types.Type) {
	for i := 0; i < typ.NumMethods(); i++ {
		fun := typ.Method(i)
		if fun.Name() != name || !fun.Exported() {
			continue
		}
		sign := fun.Type().(*types.Signature)
		if sign.Params().Len() != 0 || sign.Results().Len() != 1 {
			continue
		}
		if err := checkConvertible(sign.Results().At(0).Type(), to); err == nil {
			return fun, sign.Results().At(0).Type()
		}
	}
	return nil, nil
}

// checkConvertible determines whether values of type from can be converted to type to. It
// returns nil if convertible and a descriptive error otherwise.
// See package documentation for this definition of 'convertible'.
func checkConvertible(from, to types.Type) error {
	if types.ConvertibleTo(from, to) {
		return nil
	}
	// Slices.
	sfrom := underlyingSlice(from)
	sto := underlyingSlice(to)
	if sfrom != nil && sto != nil {
		if !types.ConvertibleTo(sfrom.Elem(), sto.Elem()) {
			return fmt.Errorf("slice element type %s is not convertible to %s", sfrom.Elem(), sto.Elem())
		}
		return nil
	}
	// Maps.
	mfrom := underlyingMap(from)
	mto := underlyingMap(to)
	if mfrom != nil && mto != nil {
		if !types.ConvertibleTo(mfrom.Key(), mto.Key()) {
			return fmt.Errorf("map key type %s is not convertible to %s", mfrom.Key(), mto.Key())
		}
		if !types.ConvertibleTo(mfrom.Elem(), mto.Elem()) {
			return fmt.Errorf("map element type %s is not convertible to %s", mfrom.Elem(), mto.Elem())
		}
		return nil
	}
	return fmt.Errorf("type %s is not convertible to %s", from, to)
}

func underlyingSlice(typ types.Type) *types.Slice {
	for {
		switch typ.(type) {
		case *types.Named:
			typ = typ.Underlying()
		case *types.Slice:
			return typ.(*types.Slice)
		default:
			return nil
		}
	}
}

func underlyingMap(typ types.Type) *types.Map {
	for {
		switch typ.(type) {
		case *types.Named:
			typ = typ.Underlying()
		case *types.Map:
			return typ.(*types.Map)
		default:
			return nil
		}
	}
}

func (mtyp *marshalerType) fieldByName(name string) *marshalerField {
	for _, f := range mtyp.Fields {
		if f.name == name {
			return f
		}
	}
	return nil
}

// marshalerField represents a field of the intermediate marshaling type.
type marshalerField struct {
	name     string
	typ      types.Type
	origTyp  types.Type
	tag      string
	function *types.Func // map to a function instead of a field
}

// isRequired returns whether the field is required when decoding the given format.
func (mf *marshalerField) isRequired(format string) bool {
	rtag := reflect.StructTag(mf.tag)
	req := rtag.Get("required") == "true"
	// Fields with json:"-" must be treated as optional. This also works
	// for the other supported formats.
	return req && !strings.HasPrefix(rtag.Get(format), "-")
}

func (mf *marshalerField) hasMinimal(format string) (string, bool) {
	rtag := reflect.StructTag(mf.tag)
	min, ok := rtag.Lookup("min")
	// Fields with json:"-" must be treated as optional. This also works
	// for the other supported formats.
	return min, ok && !strings.HasPrefix(rtag.Get(format), "-")
}

// fileScope tracks imports and other names at file scope.
type fileScope struct {
	imports       []*types.Package
	importsByName map[string]*types.Package
	importNames   map[string]string
	otherNames    map[string]bool // non-package identifiers
	pkg           *types.Package
	imp           types.Importer
}

// qualify is a types.Qualifier that prepends the (possibly renamed) package name of
// imported types to a type name.
func (s *fileScope) qualify(pkg *types.Package) string {
	if pkg == s.pkg {
		return ""
	}
	return s.packageName(pkg.Path())
}
func (s *fileScope) packageName(path string) string {
	name, ok := s.importNames[path]
	if !ok {
		panic("BUG: missing package " + path)
	}
	return name
}

func (s *fileScope) addReferences(typ types.Type) {
	walkNamedTypes(typ, func(nt *types.Named) {
		pkg := nt.Obj().Pkg()
		if pkg == s.pkg {
			s.otherNames[nt.Obj().Name()] = true
		} else if pkg != nil {
			s.insertImport(nt.Obj().Pkg())
		}
	})
	s.rebuildImports()
}

// addImport loads a package and adds it to the import set.
func (s *fileScope) addImport(path string) {
	pkg, err := s.imp.Import(path)
	if err != nil {
		panic(fmt.Errorf("can't import %q: %v", path, err))
	}
	s.insertImport(pkg)
	s.rebuildImports()
}

// insertImport adds pkg to the list of known imports.
// This method should not be used directly because it doesn't
// rebuild the import name cache.
func (s *fileScope) insertImport(pkg *types.Package) {
	i := sort.Search(len(s.imports), func(i int) bool {
		return s.imports[i].Path() >= pkg.Path()
	})
	if i < len(s.imports) && s.imports[i] == pkg {
		return
	}
	s.imports = append(s.imports[:i], append([]*types.Package{pkg}, s.imports[i:]...)...)
}

// rebuildImports caches the names of imported packages.
func (s *fileScope) rebuildImports() {
	s.importNames = make(map[string]string)
	s.importsByName = make(map[string]*types.Package)
	for _, pkg := range s.imports {
		s.maybeRenameImport(pkg)
	}
}
func (s *fileScope) maybeRenameImport(pkg *types.Package) {
	name := pkg.Name()
	for i := 0; s.isNameTaken(name); i++ {
		name = pkg.Name()
		if i > 0 {
			name += strconv.Itoa(i - 1)
		}
	}
	s.importNames[pkg.Path()] = name
	s.importsByName[name] = pkg
}

// isNameTaken reports whether the given name is used by an import or other identifier.
func (s *fileScope) isNameTaken(name string) bool {
	return s.importsByName[name] != nil || s.otherNames[name] || types.Universe.Lookup(name) != nil
}

func (s *fileScope) writeImportDecl(w io.Writer) {
	fmt.Fprintln(w, "import (")
	for _, pkg := range s.imports {
		if s.importNames[pkg.Path()] != pkg.Name() {
			fmt.Fprintf(w, "\t%s %q\n", s.importNames[pkg.Path()], pkg.Path())
		} else {
			fmt.Fprintf(w, "\t%q\n", pkg.Path())
		}
	}
	fmt.Fprintln(w, ")")
}

// main logic to gen codde
func process(cfg *Config) (code []byte, err error) {
	pkg, err := loadPackage(cfg)
	if err != nil {
		return nil, err
	}
	typ, err := lookupStructType(pkg.Scope(), cfg.Type)
	if err != nil {
		return nil, fmt.Errorf("can't find (%s) in %q: %v", cfg.Type, pkg.Path(), err)
	}
	// Construct the marshaling type.
	mtyp := newMarshalerType(cfg, typ)
	if cfg.FieldOverride != "" {
		if otyp, err := lookupStructType(pkg.Scope(), cfg.FieldOverride); err != nil {
			return nil, fmt.Errorf("can't find field replacement type %s: %v", cfg.FieldOverride, err)
		} else if err := mtyp.loadOverrides(otyp); err != nil {
			return nil, err
		}
	}
	// Generate and format the output. Formatting uses goimports because it
	// removes unused imports.
	code, err = generate(mtyp, cfg)
	if err != nil {
		return nil, err
	}
	opt := &imports.Options{Comments: true, TabIndent: true, TabWidth: 8}
	code, err = imports.Process("", code, opt)
	if err != nil {
		panic(fmt.Errorf("BUG: can't gofmt generated code: %v", err))
	}
	return code, nil

}

func generate(mtyp *marshalerType, cfg *Config) ([]byte, error) {
	w := new(bytes.Buffer)
	fmt.Fprint(w, "// Code generated by qitmeer/tools/ngen. DO NOT EDIT.\n\n")
	fmt.Fprintln(w, "package", mtyp.orig.Obj().Pkg().Name())
	fmt.Fprintln(w)
	mtyp.scope.writeImportDecl(w)
	fmt.Fprintln(w)
	if mtyp.override != nil {
		// write override struct
		name := types.TypeString(types.NewPointer(mtyp.override), mtyp.scope.qualify)
		fmt.Fprintf(w, "var _ = (%s)(nil)\n", name)
	}
	genMarshal := genMarshalJSON(mtyp)
	genUnmarshal := genUnmarshalJSON(mtyp)
	fmt.Fprintf(w, "// %s marshals as JSON", genMarshal.Name)
	fmt.Fprintln(w)
	writeFunction(w, mtyp.fs, genMarshal)
	fmt.Fprintln(w)
	fmt.Fprintln(w)
	fmt.Fprintf(w, "// %s unmarshals from JSON", genUnmarshal.Name)
	fmt.Fprintln(w)
	writeFunction(w, mtyp.fs, genUnmarshal)
	fmt.Fprintln(w)
	return w.Bytes(), nil
}

func writeFunction(w io.Writer, fs *token.FileSet, fn Function) {
	printer.Fprint(w, fs, fn.Declaration())
	fmt.Fprintln(w)
}

// genUnmarshalJSON generates the UnmarshalJSON method.
func genUnmarshalJSON(mtyp *marshalerType) Function {
	var (
		m        = newMarshalMethod(mtyp, true)
		recv     = m.receiver()
		input    = Name(m.scope.newIdent("input"))
		intertyp = m.intermediateType(m.scope.newIdent(m.mtyp.orig.Obj().Name()))
		dec      = Name(m.scope.newIdent("dec"))
		json     = Name(m.scope.parent.packageName("encoding/json"))
	)
	fn := Function{
		Receiver:    recv,
		Name:        "UnmarshalJSON",
		ReturnTypes: Types{{TypeName: "error"}},
		Parameters:  Types{{Name: input.Name, TypeName: "[]byte"}},
		Body: []Statement{
			declStmt{intertyp},
			Declare{Name: dec.Name, TypeName: intertyp.Name},
			errCheck(CallFunction{
				Func:   Dotted{Receiver: json, Name: "Unmarshal"},
				Params: []Expression{input, AddressOf{Value: dec}},
			}),
		},
	}
	fn.Body = append(fn.Body, m.unmarshalConversions(dec, Name(recv.Name), "json")...)
	fn.Body = append(fn.Body, Return{Values: []Expression{NIL}})
	return fn
}

// genMarshalJSON generates the MarshalJSON method.
func genMarshalJSON(mtyp *marshalerType) Function {
	var (
		m        = newMarshalMethod(mtyp, false)
		recv     = m.receiver()
		intertyp = m.intermediateType(m.scope.newIdent(m.mtyp.orig.Obj().Name()))
		enc      = Name(m.scope.newIdent("enc"))
		json     = Name(m.scope.parent.packageName("encoding/json"))
	)
	fn := Function{
		Receiver:    recv,
		Name:        "MarshalJSON",
		ReturnTypes: Types{{TypeName: "[]byte"}, {TypeName: "error"}},
		Body: []Statement{
			declStmt{intertyp},
			Declare{Name: enc.Name, TypeName: intertyp.Name},
		},
	}
	fn.Body = append(fn.Body, m.marshalConversions(Name(recv.Name), enc, "json")...)
	fn.Body = append(fn.Body, Return{Values: []Expression{
		CallFunction{
			Func:   Dotted{Receiver: json, Name: "Marshal"},
			Params: []Expression{AddressOf{Value: enc}},
		},
	}})
	return fn
}

var NIL = Name("nil")

type AddressOf struct {
	Value Expression
}

func (me AddressOf) Expression() ast.Expr {
	return &ast.UnaryExpr{
		X:  me.Value.Expression(),
		Op: token.AND,
	}
}

func errCheck(expr Expression) If {
	err := Name("err")
	return If{
		Init:      DeclareAndAssign{Lhs: err, Rhs: expr},
		Condition: NotEqual{Lhs: err, Rhs: NIL},
		Body:      []Statement{Return{Values: []Expression{err}}},
	}
}

type Return struct {
	Values []Expression
}

func (me Return) Statement() ast.Stmt {
	ret := make([]ast.Expr, len(me.Values))
	for i, val := range me.Values {
		ret[i] = val.Expression()
	}
	return &ast.ReturnStmt{
		Results: ret,
	}
}

type CallFunction struct {
	Func   Expression
	Params []Expression
}

func (me CallFunction) Statement() ast.Stmt {
	return &ast.ExprStmt{
		X: me.Expression(),
	}
}

func (me CallFunction) Expression() ast.Expr {
	params := make([]ast.Expr, len(me.Params))
	for i, param := range me.Params {
		params[i] = param.Expression()
	}
	return &ast.CallExpr{
		Fun:  me.Func.Expression(),
		Args: params,
	}
}

type Dotted struct {
	Receiver Expression
	Name     string
}

func (me Dotted) Expression() ast.Expr {
	return &ast.SelectorExpr{
		X: me.Receiver.Expression(),
		Sel: &ast.Ident{
			Name: me.Name,
		},
	}
}

type If struct {
	Init      Statement
	Condition Expression
	Body      []Statement
}

func (me If) Statement() ast.Stmt {
	var (
		init ast.Stmt
	)
	if me.Init != nil {
		init = me.Init.Statement()
	}
	body := make([]ast.Stmt, len(me.Body))
	for j, stmt := range me.Body {
		body[j] = stmt.Statement()
	}
	return &ast.IfStmt{
		Init: init,
		Cond: me.Condition.Expression(),
		Body: &ast.BlockStmt{
			List: body,
		},
	}
}

type Equals struct {
	Lhs Expression
	Rhs Expression
}

func (me Equals) Expression() ast.Expr {
	return &ast.BinaryExpr{
		Op: token.EQL,
		X:  me.Lhs.Expression(),
		Y:  me.Rhs.Expression(),
	}
}

type NotEqual struct {
	Lhs Expression
	Rhs Expression
}

func (me NotEqual) Expression() ast.Expr {
	return &ast.BinaryExpr{
		Op: token.NEQ,
		X:  me.Lhs.Expression(),
		Y:  me.Rhs.Expression(),
	}
}

type LessThan struct {
	Lhs Expression
	Rhs Expression
}

func (me LessThan) Expression() ast.Expr {
	return &ast.BinaryExpr{
		Op: token.LSS,
		X:  me.Lhs.Expression(),
		Y:  me.Rhs.Expression(),
	}
}

type DeclareAndAssign struct {
	Lhs Expression
	Rhs Expression
}

func (me DeclareAndAssign) Statement() ast.Stmt {
	return &ast.AssignStmt{
		Tok: token.DEFINE,
		Lhs: []ast.Expr{me.Lhs.Expression()},
		Rhs: []ast.Expr{me.Rhs.Expression()},
	}
}

func newMarshalMethod(mtyp *marshalerType, isUnmarshal bool) *marshalMethod {
	s := newFuncScope(mtyp.scope)
	return &marshalMethod{
		mtyp:        mtyp,
		scope:       newFuncScope(mtyp.scope),
		isUnmarshal: isUnmarshal,
		iterKey:     Name(s.newIdent("k")),
		iterVal:     Name(s.newIdent("v")),
	}
}

type marshalMethod struct {
	mtyp        *marshalerType
	scope       *funcScope
	isUnmarshal bool
	// cached identifiers for map, slice conversions
	iterKey, iterVal Var
}

func (m *marshalMethod) marshalConversions(from, to Var, format string) (s []Statement) {
	for _, f := range m.mtyp.Fields {
		accessFrom := Dotted{Receiver: from, Name: f.name}
		accessTo := Dotted{Receiver: to, Name: f.name}
		if f.function != nil {
			s = append(s, m.convert(CallFunction{Func: accessFrom}, accessTo, f.origTyp, f.typ)...)
		} else {
			s = append(s, m.convert(accessFrom, accessTo, f.origTyp, f.typ)...)
		}
	}
	return s
}

func (m *marshalMethod) unmarshalConversions(from, to Var, format string) (s []Statement) {
	for _, f := range m.mtyp.Fields {
		if f.function != nil {
			continue // fields generated from functions cannot be assigned
		}

		accessFrom := Dotted{Receiver: from, Name: f.name}
		accessTo := Dotted{Receiver: to, Name: f.name}
		typ := ensureNilCheckable(f.typ)
		if !f.isRequired(format) {
			s = append(s, If{
				Condition: NotEqual{Lhs: accessFrom, Rhs: NIL},
				Body:      m.convert(accessFrom, accessTo, typ, f.origTyp),
			})
		} else {
			err := fmt.Sprintf("missing required field '%s' for %s", f.encodedName(format), m.mtyp.name)
			errors := m.scope.parent.packageName("errors")
			s = append(s, If{
				Condition: Equals{Lhs: accessFrom, Rhs: NIL},
				Body: []Statement{
					Return{
						Values: []Expression{
							CallFunction{
								Func:   Dotted{Receiver: Name(errors), Name: "New"},
								Params: []Expression{stringLit{err}},
							},
						},
					},
				},
			})
			s = append(s, m.convert(accessFrom, accessTo, typ, f.origTyp)...)
			if v, ok := f.hasMinimal(format); ok {
				err := fmt.Sprintf("error field '%s' for %s, minimal is %s", f.encodedName(format), m.mtyp.name, v)
				if f.origTyp.String() == "*math/big.Int" { //need to handle bigInt
					s = append(s, If{
						Condition: Equals{
							Lhs: CallFunction{
								Func:   Dotted{Receiver: accessTo, Name: "Cmp"},
								Params: []Expression{Name("big.NewInt(" + v + ")")},
							},
							Rhs: Name("-1")},
						Body: []Statement{
							Return{
								Values: []Expression{
									CallFunction{
										Func:   Dotted{Receiver: Name(errors), Name: "New"},
										Params: []Expression{stringLit{err}},
									},
								},
							},
						},
					})
				} else {
					s = append(s, If{
						Condition: LessThan{Lhs: accessTo, Rhs: Name(v)},
						Body: []Statement{
							Return{
								Values: []Expression{
									CallFunction{
										Func:   Dotted{Receiver: Name(errors), Name: "New"},
										Params: []Expression{stringLit{err}},
									},
								},
							},
						},
					})
				}
			}
		}
	}
	return s
}

// encodedName returns the alternative field name assigned by the format's struct tag.
func (mf *marshalerField) encodedName(format string) string {
	val := reflect.StructTag(mf.tag).Get(format)
	if comma := strings.Index(val, ","); comma != -1 {
		val = val[:comma]
	}
	if val == "" || val == "-" {
		return uncapitalize(mf.name)
	}
	return val
}
func uncapitalize(s string) string {
	return strings.ToLower(s[:1]) + s[1:]
}

func (m *marshalMethod) convert(from, to Expression, fromtyp, totyp types.Type) (s []Statement) {
	// Remove pointer introduced by ensureNilCheckable during field building.
	if isPointer(fromtyp) && !isPointer(totyp) {
		from = Star{Value: from}
		fromtyp = fromtyp.(*types.Pointer).Elem()
	} else if !isPointer(fromtyp) && isPointer(totyp) {
		from = AddressOf{Value: from}
		fromtyp = types.NewPointer(fromtyp)
	}
	// Generate the conversion.
	qf := m.mtyp.scope.qualify
	switch {
	case types.ConvertibleTo(fromtyp, totyp):
		s = append(s, Assign{Lhs: to, Rhs: simpleConv(from, fromtyp, totyp, qf)})
	case underlyingSlice(fromtyp) != nil:
		s = append(s, m.loopConv(from, to, sliceKV(fromtyp), sliceKV(totyp))...)
	case underlyingMap(fromtyp) != nil:
		s = append(s, m.loopConv(from, to, mapKV(fromtyp), mapKV(totyp))...)
	default:
		invalidConv(fromtyp, totyp, qf)
	}
	return s
}

func simpleConv(from Expression, fromtyp, totyp types.Type, qf types.Qualifier) Expression {
	if types.AssignableTo(fromtyp, totyp) {
		return from
	}
	if !types.ConvertibleTo(fromtyp, totyp) {
		invalidConv(fromtyp, totyp, qf)
	}
	toname := types.TypeString(totyp, qf)
	if isPointer(totyp) {
		toname = "(" + toname + ")" // hack alert!
	}
	return CallFunction{Func: Name(toname), Params: []Expression{from}}
}
func invalidConv(from, to types.Type, qf types.Qualifier) {
	panic(fmt.Errorf("BUG: invalid conversion %s -> %s", types.TypeString(from, qf), types.TypeString(to, qf)))
}

func isPointer(typ types.Type) bool {
	_, ok := typ.(*types.Pointer)
	return ok
}

type kvType struct {
	Type      types.Type
	Key, Elem types.Type
}

var intType = types.Universe.Lookup("int").Type()

func mapKV(typ types.Type) kvType {
	maptyp := underlyingMap(typ)
	return kvType{typ, maptyp.Key(), maptyp.Elem()}
}

func sliceKV(typ types.Type) kvType {
	slicetyp := underlyingSlice(typ)
	return kvType{typ, intType, slicetyp.Elem()}
}

func (m *marshalMethod) loopConv(from, to Expression, fromTyp, toTyp kvType) (conv []Statement) {
	if hasSideEffects(from) {
		orig := from
		from = Name(m.scope.newIdent("tmp"))
		conv = []Statement{DeclareAndAssign{Lhs: from, Rhs: orig}}
	}
	// The actual conversion is a loop that assigns each element.
	inner := []Statement{
		Assign{Lhs: to, Rhs: makeExpr(toTyp.Type, from, m.scope.parent.qualify)},
		Range{
			Key:        m.iterKey,
			Value:      m.iterVal,
			RangeValue: from,
			Body: []Statement{Assign{
				Lhs: Index{Value: to, Index: simpleConv(m.iterKey, fromTyp.Key, toTyp.Key, m.scope.parent.qualify)},
				Rhs: simpleConv(m.iterVal, fromTyp.Elem, toTyp.Elem, m.scope.parent.qualify),
			}},
		},
	}
	// Preserve nil maps and slices when marshaling. This is not required for unmarshaling
	// methods because the field is already nil-checked earlier.
	if !m.isUnmarshal {
		inner = []Statement{If{
			Condition: NotEqual{Lhs: from, Rhs: NIL},
			Body:      inner,
		}}
	}
	return append(conv, inner...)
}

// hasSideEffects returns whether an expression may have side effects.
func hasSideEffects(expr Expression) bool {
	switch expr := expr.(type) {
	case Var:
		return false
	case Dotted:
		return hasSideEffects(expr.Receiver)
	case Star:
		return hasSideEffects(expr.Value)
	case Index:
		return hasSideEffects(expr.Index) && hasSideEffects(expr.Value)
	default:
		return true
	}
}

func makeExpr(typ types.Type, lenfrom Expression, qf types.Qualifier) Expression {
	return CallFunction{Func: Name("make"), Params: []Expression{
		Name(types.TypeString(typ, qf)),
		CallFunction{Func: Name("len"), Params: []Expression{lenfrom}},
	}}
}

func (m *marshalMethod) receiver() Receiver {
	letter := strings.ToLower(m.mtyp.name[:1])
	r := Receiver{Name: m.scope.newIdent(letter), Type: Name(m.mtyp.name)}
	if m.isUnmarshal {
		r.Type = Star{Value: r.Type}
	}
	return r
}

func (m *marshalMethod) intermediateType(name string) Struct {
	s := Struct{Name: name}
	for _, f := range m.mtyp.Fields {
		if m.isUnmarshal && f.function != nil {
			continue // fields generated from functions cannot be assigned on unmarshal
		}
		typ := f.typ
		if m.isUnmarshal {
			typ = ensureNilCheckable(typ)
		}
		s.Fields = append(s.Fields, Field{
			Name:     f.name,
			TypeName: types.TypeString(typ, m.mtyp.scope.qualify),
			Tag:      f.tag,
		})
	}
	return s
}

func ensureNilCheckable(typ types.Type) types.Type {
	orig := typ
	named := false
	for {
		switch typ.(type) {
		case *types.Named:
			typ = typ.Underlying()
			named = true
		case *types.Slice, *types.Map:
			if named {
				// Named slices, maps, etc. are special because they can have a custom
				// decoder function that prevents the JSON null value. Wrap them with a
				// pointer to allow null always so required/optional works as expected.
				return types.NewPointer(orig)
			}
			return orig
		case *types.Pointer, *types.Interface:
			return orig
		default:
			return types.NewPointer(orig)
		}
	}
}

type declStmt struct {
	d Declaration
}

func (ds declStmt) Statement() ast.Stmt {
	return &ast.DeclStmt{Decl: ds.d.Declaration()}
}

type Declaration interface {
	Declaration() ast.Decl
}

type stringLit struct {
	V string
}

func (l stringLit) Expression() ast.Expr {
	return &ast.BasicLit{Kind: token.STRING, Value: strconv.Quote(l.V)}
}

type Declare struct {
	Name     string
	TypeName string
}

func (me Declare) Statement() ast.Stmt {
	return &ast.DeclStmt{
		Decl: &ast.GenDecl{
			Tok: token.VAR,
			Specs: []ast.Spec{
				&ast.ValueSpec{
					Names: []*ast.Ident{
						{
							Name: me.Name,
							Obj: &ast.Object{
								Kind: ast.Var,
								Name: me.Name,
							},
						},
					},
					Type: &ast.Ident{
						Name: me.TypeName,
					},
				},
			},
		},
	}
}

type Assign struct {
	Lhs Expression
	Rhs Expression
}

func (me Assign) Statement() ast.Stmt {
	return &ast.AssignStmt{
		Tok: token.ASSIGN,
		Lhs: []ast.Expr{me.Lhs.Expression()},
		Rhs: []ast.Expr{me.Rhs.Expression()},
	}
}

type Function struct {
	Receiver    Receiver
	Name        string
	ReturnTypes Types
	Parameters  Types
	Body        []Statement
}
type Functions []Function

func (me Function) Declaration() ast.Decl {
	paramFields := make([]*ast.Field, len(me.Parameters))
	for j, param := range me.Parameters {
		var names []*ast.Ident
		if param.Name != "" {
			names = []*ast.Ident{
				{
					Name: param.Name,
					Obj: &ast.Object{
						Kind: ast.Var,
						Name: param.Name,
					},
				},
			}
		}
		paramFields[j] = &ast.Field{
			Names: names,
			Type: &ast.Ident{
				Name: param.TypeName,
			},
		}
	}
	returnFields := make([]*ast.Field, len(me.ReturnTypes))
	for j, ret := range me.ReturnTypes {
		var names []*ast.Ident
		if ret.Name != "" {
			names = []*ast.Ident{
				{
					Name: ret.Name,
					Obj: &ast.Object{
						Kind: ast.Var,
						Name: ret.Name,
					},
				},
			}
		}
		returnFields[j] = &ast.Field{
			Names: names,
			Type: &ast.Ident{
				Name: ret.TypeName,
			},
		}
	}
	stmts := make([]ast.Stmt, len(me.Body))
	for j, stmt := range me.Body {
		stmts[j] = stmt.Statement()
	}
	return &ast.FuncDecl{
		Recv: me.Receiver.Ast(),
		Name: &ast.Ident{
			Name: me.Name,
			Obj: &ast.Object{
				Kind: ast.Fun,
				Name: me.Name,
			},
		},
		Type: &ast.FuncType{
			Params: &ast.FieldList{
				List: paramFields,
			},
			Results: &ast.FieldList{
				List: returnFields,
			},
		},
		Body: &ast.BlockStmt{
			List: stmts,
		},
	}
}

type Struct struct {
	Name        string
	Fields      Fields
	Methods     Functions
	FieldValues map[string]Expression
}

func (me Struct) Declaration() ast.Decl {
	return &ast.GenDecl{
		Tok: token.TYPE,
		Specs: []ast.Spec{
			&ast.TypeSpec{
				Name: &ast.Ident{
					Name: me.Name,
					Obj: &ast.Object{
						Kind: ast.Typ,
						Name: me.Name,
					},
				},
				Type: &ast.StructType{
					Fields: me.Fields.Ast(),
				},
			},
		},
	}
}

func (me Struct) Expression() ast.Expr {
	elts := make([]ast.Expr, len(me.Fields))
	for i, field := range me.Fields {
		elts[i] = &ast.KeyValueExpr{
			Key: &ast.Ident{
				Name: field.Name,
			},
			Value: &ast.Ident{
				//Value: me.FieldValues[field.Name].Expression(),
			},
		}
	}
	return &ast.CompositeLit{
		Type: &ast.Ident{
			Name: me.Name,
		},
		Elts: elts,
	}
}

type Field struct {
	Name     string
	TypeName string
	Tag      string
}

func (me Field) Ast() *ast.Field {
	var tag *ast.BasicLit
	if me.Tag != "" {
		tag = &ast.BasicLit{
			Kind:  token.STRING,
			Value: "`" + me.Tag + "`",
		}
	}
	names := []*ast.Ident{}
	if me.Name != "" {
		names = []*ast.Ident{
			{
				Name: me.Name,
				Obj: &ast.Object{
					Kind: ast.Var,
					Name: me.Name,
				},
			},
		}
	}
	return &ast.Field{
		Names: names,
		Type: &ast.Ident{
			Name: me.TypeName,
		},
		Tag: tag,
	}
}

type Fields []Field

func (me Fields) Ast() *ast.FieldList {
	fields := make([]*ast.Field, len(me))
	for i, field := range me {
		fields[i] = field.Ast()
	}
	return &ast.FieldList{
		List: fields,
	}
}

type Type struct {
	Name        string // Optional, named type
	TypeName    string
	PackageName string // Optional
}

type Types []Type

type Statement interface {
	Statement() ast.Stmt
}

type Range struct {
	Key          Expression
	Value        Expression
	RangeValue   Expression
	Body         []Statement
	DoNotDeclare bool
}

func (me Range) Statement() ast.Stmt {
	body := make([]ast.Stmt, len(me.Body))
	for i, bodyPart := range me.Body {
		body[i] = bodyPart.Statement()
	}
	var (
		key   Expression = Var{"_"}
		value Expression = Var{"_"}
	)

	if me.Key != nil {
		key = me.Key
	}
	if me.Value != nil {
		value = me.Value
	}
	tok := token.DEFINE
	if me.DoNotDeclare || (me.Key == nil && me.Value == nil) {
		tok = token.ASSIGN
	}

	return &ast.RangeStmt{
		Key:   key.Expression(),
		Value: value.Expression(),
		X:     me.RangeValue.Expression(),
		Tok:   tok,
		Body: &ast.BlockStmt{
			List: body,
		},
	}
}

type Receiver struct {
	Name string
	Type Expression
}

func (me Receiver) Ast() *ast.FieldList {
	if me.Type == nil {
		return nil
	}
	return &ast.FieldList{
		List: []*ast.Field{
			{
				Names: []*ast.Ident{
					{
						Name: me.Name,
						Obj: &ast.Object{
							Kind: ast.Var,
							Name: me.Name,
						},
					},
				},
				Type: me.Type.Expression(),
			},
		},
	}
}

type Expression interface {
	Expression() ast.Expr
}

func Name(value string) Var {
	return Var{value}
}

type Var struct {
	Name string
}

func (me Var) Expression() ast.Expr {
	return &ast.Ident{
		Name: me.Name,
		Obj: &ast.Object{
			Kind: ast.Var,
			Name: me.Name,
		},
	}
}

type Star struct {
	Value Expression
}

func (me Star) Expression() ast.Expr {
	return &ast.StarExpr{
		X: me.Value.Expression(),
	}
}

type Index struct {
	Value, Index Expression
}

func (me Index) Expression() ast.Expr {
	return &ast.IndexExpr{
		X:     me.Value.Expression(),
		Index: me.Index.Expression(),
	}
}

// funcScope tracks used identifiers in a function. It can create new identifiers that do
// not clash with the parent scope.
type funcScope struct {
	used   map[string]bool
	parent *fileScope
}

func newFuncScope(parent *fileScope) *funcScope {
	return &funcScope{make(map[string]bool), parent}
}

// newIdent creates a new identifier that doesn't clash with any name
// in the scope or its parent file scope.
func (s *funcScope) newIdent(base string) string {
	for i := 0; ; i++ {
		name := base
		if i > 0 {
			name += strconv.Itoa(i - 1)
		}
		if !s.parent.isNameTaken(name) && !s.used[name] {
			s.used[name] = true
			return name
		}
	}
}

func newMarshalerType(cfg *Config, typ *types.Named) *marshalerType {
	styp := typ.Underlying().(*types.Struct)
	mtyp := &marshalerType{name: typ.Obj().Name(), fs: token.NewFileSet(), orig: typ}
	mtyp.scope = newFileScope(cfg, typ.Obj().Pkg())
	mtyp.scope.addReferences(styp)

	// Add packages which are always needed.
	mtyp.scope.addImport("encoding/json")
	mtyp.scope.addImport("errors")

	for i := 0; i < styp.NumFields(); i++ {
		f := styp.Field(i)
		if !f.Exported() {
			continue
		}
		if f.Anonymous() {
			fmt.Fprintf(os.Stderr, "Warning: ignoring embedded field %s\n", f.Name())
			continue
		}

		mf := &marshalerField{
			name:    f.Name(),
			typ:     f.Type(),
			origTyp: f.Type(),
			tag:     styp.Tag(i),
		}

		mtyp.Fields = append(mtyp.Fields, mf)
	}
	return mtyp
}
func newFileScope(cfg *Config, pkg *types.Package) *fileScope {
	return &fileScope{otherNames: make(map[string]bool), pkg: pkg, imp: importer.Default()}
}

// walkNamedTypes runs the callback for all named types contained in the given type.
func walkNamedTypes(typ types.Type, callback func(*types.Named)) {
	switch typ := typ.(type) {
	case *types.Basic:
	case *types.Chan:
		walkNamedTypes(typ.Elem(), callback)
	case *types.Map:
		walkNamedTypes(typ.Key(), callback)
		walkNamedTypes(typ.Elem(), callback)
	case *types.Named:
		callback(typ)
	case *types.Pointer:
		walkNamedTypes(typ.Elem(), callback)
	case *types.Slice:
		walkNamedTypes(typ.Elem(), callback)
	case *types.Struct:
		for i := 0; i < typ.NumFields(); i++ {
			walkNamedTypes(typ.Field(i).Type(), callback)
		}
	case *types.Interface:
		if typ.NumMethods() > 0 {
			panic("BUG: can't walk non-empty interface")
		}
	default:
		panic(fmt.Errorf("BUG: can't walk %T", typ))
	}
}

func lookupStructType(scope *types.Scope, name string) (*types.Named, error) {
	obj := scope.Lookup(name)
	if obj == nil {
		return nil, fmt.Errorf("can't find type name (%s)", name)
	}
	typ, ok := obj.(*types.TypeName)
	if !ok {
		return nil, fmt.Errorf("%s not a type", name)
	}
	if _, ok = typ.Type().Underlying().(*types.Struct); !ok {
		return nil, fmt.Errorf(" %s not a struct type", name)
	}
	return typ.Type().(*types.Named), nil
}

// Find the import path of the package in the given directory.
func loadPackage(cfg *Config) (*types.Package, error) {
	// Find the import path of the package in the given directory.
	cwd, _ := os.Getwd()
	dir := filepath.Join(cfg.Dir, "*.go")
	pkg, err := buildutil.ContainingPackage(&build.Default, cwd, dir)
	if err != nil {
		return nil, err
	}
	nocheck := func(path string) bool { return false }
	lcfg := loader.Config{Fset: token.NewFileSet(), TypeCheckFuncBodies: nocheck}
	lcfg.ImportWithTests(pkg.ImportPath)
	prog, err := lcfg.Load()
	if err != nil {
		return nil, err
	}
	return prog.Package(pkg.ImportPath).Pkg, nil
}

func main() {
	var (
		pkgdir    = flag.String("dir", ".", "input package (default is .)")
		output    = flag.String("out", "-", "output file (default is stdout)")
		typename  = flag.String("type", "", "type to generate methods for")
		overrides = flag.String("field-override", "", "override type contains fields"+
			" which are excluded (default is `<typename>JSON`)")
	)
	flag.Parse()

	cfg := Config{Dir: *pkgdir, Type: *typename, FieldOverride: *overrides}
	code, err := process(&cfg)
	if err != nil {
		log.Fatal(err)
	}
	if *output == "-" {
		os.Stdout.Write(code)
	} else if err := ioutil.WriteFile(*output, code, 0644); err != nil {
		log.Fatal(err)
	}
}
