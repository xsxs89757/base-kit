// basekit-migrate 把 base 模板里的 import 路径改写成 base-kit 的路径。
//
// 基底 v2.0.0 把框架层和系统管理模块搬到了 github.com/xsxs89757/base-kit，
// 下游 make sync-base 之后跑一次这个工具，业务代码里的 import 就自动改好了：
//
//	go run github.com/xsxs89757/base-kit/cmd/basekit-migrate@latest ./...
//
// 用 go/ast 改写而不是 sed：sed -i 在 BSD/GNU 下参数不同（dev.sh 还要支持
// Windows Git Bash），而且正则容易误伤字符串字面量和注释。
package main

import (
	"flag"
	"fmt"
	"go/ast"
	"go/format"
	"go/parser"
	"go/token"
	"io/fs"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

const kit = "github.com/xsxs89757/base-kit"

// 搬到 kit 的包。key 是模板里的旧路径，value 是 kit 里的新路径。
// 注意顺序无关：查表是精确匹配，不做前缀替换（base/internal/router 必须留在模板里）。
var importMap = map[string]string{
	"base/config":                   kit + "/config",
	"base/internal/dto":             kit + "/dto",
	"base/internal/dto/admin":       kit + "/dto/admin",
	"base/internal/validator":       kit + "/validator",
	"base/internal/validator/admin": kit + "/validator/admin",
	"base/internal/model":           kit + "/model",
	"base/internal/model/admin":     kit + "/model/admin",
	"base/internal/store":           kit + "/store",
	"base/internal/middleware":      kit + "/middleware",
	"base/internal/service/admin":   kit + "/service/admin",
	"base/internal/handler/admin":   kit + "/handler/admin",
}

func main() {
	dryRun := flag.Bool("n", false, "只列出会改动的文件，不写回")
	verbose := flag.Bool("v", false, "打印每处改写")
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, `用法: basekit-migrate [-n] [-v] [路径...]

把 base 模板的 import 路径改写为 base-kit 的路径（默认处理当前目录，递归）。
「./...」和目录名等价，都表示递归处理。

映射:
`)
		for _, k := range sortedKeys(importMap) {
			fmt.Fprintf(os.Stderr, "  %-32s -> %s\n", k, importMap[k])
		}
		fmt.Fprintf(os.Stderr, "\nbase/internal/router 不在映射里：路由挂载点仍留在模板中。\n")
	}
	flag.Parse()

	roots := flag.Args()
	if len(roots) == 0 {
		roots = []string{"."}
	}

	changed := 0
	for _, root := range roots {
		root = strings.TrimSuffix(strings.TrimSuffix(root, "/..."), "/")
		if root == "" {
			root = "."
		}
		n, err := walk(root, *dryRun, *verbose)
		if err != nil {
			fmt.Fprintf(os.Stderr, "错误: %v\n", err)
			os.Exit(1)
		}
		changed += n
	}

	switch {
	case changed == 0:
		fmt.Println("没有需要改写的 import")
	case *dryRun:
		fmt.Printf("%d 个文件需要改写（-n 未写回）\n", changed)
	default:
		fmt.Printf("已改写 %d 个文件，接着执行: cd server && go mod tidy\n", changed)
	}
}

func walk(root string, dryRun, verbose bool) (int, error) {
	changed := 0
	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			// vendor / 隐藏目录 / node_modules 一律跳过
			name := d.Name()
			if name != "." && name != ".." && (strings.HasPrefix(name, ".") || name == "vendor" || name == "node_modules") {
				return filepath.SkipDir
			}
			return nil
		}
		if !strings.HasSuffix(path, ".go") {
			return nil
		}
		ok, err := rewriteFile(path, dryRun, verbose)
		if err != nil {
			return err
		}
		if ok {
			changed++
		}
		return nil
	})
	return changed, err
}

func rewriteFile(path string, dryRun, verbose bool) (bool, error) {
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, path, nil, parser.ParseComments)
	if err != nil {
		return false, fmt.Errorf("解析 %s: %w", path, err)
	}

	modified := false
	for _, spec := range file.Imports {
		old, err := strconv.Unquote(spec.Path.Value)
		if err != nil {
			continue
		}
		new, ok := importMap[old]
		if !ok {
			continue
		}
		spec.Path.Value = strconv.Quote(new)
		modified = true
		if verbose {
			fmt.Printf("  %s: %s -> %s\n", path, old, new)
		}
	}
	if !modified {
		return false, nil
	}

	fmt.Println(path)
	if dryRun {
		return true, nil
	}

	// 重排 import 分组，避免留下 gofmt 不干净的文件
	ast.SortImports(fset, file)
	var buf strings.Builder
	if err := format.Node(&buf, fset, file); err != nil {
		return false, fmt.Errorf("格式化 %s: %w", path, err)
	}
	info, err := os.Stat(path)
	if err != nil {
		return false, err
	}
	if err := os.WriteFile(path, []byte(buf.String()), info.Mode().Perm()); err != nil {
		return false, err
	}
	return true, nil
}

func sortedKeys(m map[string]string) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	for i := range keys {
		for j := i + 1; j < len(keys); j++ {
			if keys[j] < keys[i] {
				keys[i], keys[j] = keys[j], keys[i]
			}
		}
	}
	return keys
}
