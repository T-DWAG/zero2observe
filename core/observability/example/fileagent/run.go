package main

import (
	"context"
	"fmt"
	"strings"

	"github.com/T-DWAG/zero2observe/collector"
	"github.com/T-DWAG/zero2observe/storage"
	"github.com/cloudwego/eino/adk"
	"github.com/cloudwego/eino/compose"
	"github.com/cloudwego/eino/schema"
)

func runOnce(ctx context.Context, store storage.Storage, cfg collector.Config, question string) (traceID, answer string, err error) {

	//1、添加观测回调
	ctx, handler, finish := collector.WithObsCallback(ctx, store, cfg)

	defer func() { finish(ctx, answer, err) }()

	//2、创建聊天模型
	cm, err := newChatModel(ctx)
	if err != nil {
		return "", "", err
	}

	//3、创建文件工具
	tools, err := newFileTools()
	if err != nil {
		return "", "", err
	}

	//4、创建聊天模型代理
	agent, err := adk.NewChatModelAgent(ctx, &adk.ChatModelAgentConfig{
		Name:        "file-agent",
		Description: "只能通过工具访问 sandbox 文件的助手",
		Instruction: `你是工作区文件助手。
	规则：
	1. 需要知道有哪些文件时，先调用 list_files。
	2. 需要文件内容或总结某文件时，先调用 read_file。
	3. 不要编造文件内容；工具没返回的内容不要假装读过。
	4. 用简洁中文回答。`,
		Model: cm,
		ToolsConfig: adk.ToolsConfig{
			ToolsNodeConfig: compose.ToolsNodeConfig{Tools: tools},
		},
	})
	if err != nil {
		return "", "", err
	}
	//5、创建运行器

	runner := adk.NewRunner(ctx, adk.RunnerConfig{Agent: agent, EnableStreaming: false})
	iter := runner.Query(ctx, question, adk.WithCallbacks(handler))
	answer, err = drainAgent(iter)
	traceID = collector.TraceIDFromCtx(ctx)
	return traceID, answer, err
}

// drainAgent 从迭代器中提取助手文本。
// 入参：
//
//	iter: 包含 AgentEvent 的迭代器
//
// 出参：
//
//	string: 助手文本
//	error: 如果迭代器为空或助手文本为空，返回错误，否则为 nil
func drainAgent(iter *adk.AsyncIterator[*adk.AgentEvent]) (string, error) {
	var last string
	for {
		ev, ok := iter.Next()
		if !ok {
			break
		}
		if ev == nil {
			continue
		}
		if ev.Err != nil {
			return last, ev.Err
		}
		if ev.Output == nil || ev.Output.MessageOutput == nil || ev.Output.MessageOutput.Message == nil {
			continue
		}
		msg := ev.Output.MessageOutput.Message
		if msg.Role == schema.Assistant && strings.TrimSpace(msg.Content) != "" {
			last = msg.Content
		}
	}
	if last == "" {
		return "", fmt.Errorf("no assistant text in agent events")
	}
	return last, nil
}
