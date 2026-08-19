package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/components/tool/utils"
)

// 最大读取量
const maxReadBytes = 8 << 10 //8KB

type listInput struct{}

type listOutput struct {
	Files []string `json:"files"`
}

type readInput struct {
	Path string `json:"path" jsonschema:"description=相对 sandbox 的文件名，如 obs_hints.md,required=true"`
}

type readOutput struct {
	Path    string `json:"path"`
	Content string `json:"content"`
}

func sandboxRoot() (string, error) {
	candidates := []string{
		filepath.Join("example", "fileagent", "sandbox"), "sandbox"}

	for _, c := range candidates {
		st, err := os.Stat(c)
		if err == nil && st.IsDir() {
			abs, err := filepath.Abs(c)
			if err != nil {
				return "", err
			}
			return abs, nil
		}
	}
	return "", fmt.Errorf("sandbox directory not found")
}

// safeJoin函数用于将root和rel两个路径安全地合并成一个绝对路径。
// 入参：
//
//	root: 作为根目录的路径（字符串），比如 sandbox 的绝对路径
//	rel: 需要合并的相对路径（字符串），如 "obs_hints.md" 或 "subfolder/file.txt"
//
// 出参：
//
//	string: 合并后的绝对路径
//	error: 如果路径非法（例如 rel 含有 .. 跳出了 root），返回错误，否则为 nil
func safeJoin(root, rel string) (string, error) {
	rel = filepath.Clean(rel) // 确保 rel 没有冗余的 .. 或 .

	//1、rel 为 . 或 .. 或绝对路径，直接返回错误
	if rel == "." || strings.HasPrefix(rel, "..") || filepath.IsAbs(rel) {
		return "", fmt.Errorf("invalid path: %s", rel)
	}

	//2、合并路径
	full := filepath.Join(root, rel)
	absRoot, err := filepath.Abs(root)
	if err != nil {
		return "", err
	}

	//3、合并后的路径转换为绝对路径
	absFull, err := filepath.Abs(full)
	if err != nil {
		return "", err
	}

	sep := string(os.PathSeparator)
	if absFull != absRoot && !strings.HasPrefix(absFull, absRoot+sep) {
		return "", fmt.Errorf("path escapes sandbox: %q", rel)
	}
	return absFull, nil
}

func newFileTools() ([]tool.BaseTool, error) {

	listTool, err := utils.InferTool(
		"list_files",
		"列出 sandbox 目录下的文件名。回答「有哪些文件」前应先调用本工具。",
		func(_ context.Context, _ listInput) (listOutput, error) {
			root, err := sandboxRoot()
			if err != nil {
				return listOutput{}, err
			}
			entries, err := os.ReadDir(root)
			if err != nil {
				return listOutput{}, err
			}
			out := listOutput{Files: make([]string, 0, len(entries))}
			for _, e := range entries {
				if !e.IsDir() {
					out.Files = append(out.Files, e.Name())
				}
			}
			return out, nil
		},
	)
	if err != nil {
		return nil, err
	}

	readTool, err := utils.InferTool(
		"read_file",
		"读取 sandbox 内某个文本文件内容。总结某文件前应先调用本工具。",
		func(_ context.Context, in readInput) (readOutput, error) {
			root, err := sandboxRoot()
			if err != nil {
				return readOutput{}, err
			}
			full, err := safeJoin(root, in.Path)
			if err != nil {
				return readOutput{}, err
			}
			data, err := os.ReadFile(full)
			if err != nil {
				return readOutput{}, err
			}
			if len(data) > maxReadBytes {
				data = data[:maxReadBytes]
			}
			return readOutput{Path: in.Path, Content: string(data)}, nil
		},
	)
	if err != nil {
		return nil, err
	}

	return []tool.BaseTool{listTool, readTool}, nil
}
