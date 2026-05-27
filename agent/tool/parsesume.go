package tool

import (
	"context"
	"errors"
	"os"
	"os/exec"
	"time"

	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/components/tool/utils"
)

type ParseInput struct {
	FilePath string `json:"file_path"`
}

type ParseOutput struct {
	Content  string         `json:"content"`
	MetaData map[string]any `json:"meta_data"`
	ErrMsg   string         `json:"err_msg"`
}

func Parse(ctx context.Context, input *ParseInput) (*ParseOutput, error) {
	output := &ParseOutput{
		MetaData: map[string]any{
			"parse_time": time.Now().Format("2006-01-02 15:04:05"),
			"method":     "pdftotext_cli",
			"file_path":  input.FilePath,
		},
	}

	if input.FilePath == "" {
		output.ErrMsg = "file path is empty"
		return output, errors.New("pdf_to_text: file_path is empty")
	}

	if _, err := os.Stat(input.FilePath); err != nil {
		if os.IsNotExist(err) {
			output.ErrMsg = "file path does not exist"
			return output, errors.New("pdf_to_text: file does not exist: " + input.FilePath)
		}
		output.ErrMsg = err.Error()
		return output, errors.New("pdf_to_text: failed to stat file: " + err.Error())
	}

	cmd := exec.CommandContext(ctx, "pdftotext", "-layout", input.FilePath, "-")
	out, err := cmd.Output()
	if err != nil {
		if errors.Is(ctx.Err(), context.DeadlineExceeded) {
			output.ErrMsg = "parse timeout"
			return output, errors.New("pdf_to_text: command timed out")
		}
		output.ErrMsg = err.Error()
		return output, errors.New("pdf_to_text: command failed: " + err.Error())
	}

	output.Content = string(out)
	return output, nil
}

func CreatParseResumeTool() tool.InvokableTool {
	tl, err := utils.InferTool(
		"pdf_to_text",
		"Convert a local PDF file to plain text. Only text-based PDFs are supported. Input requires a local PDF path in the file_path field.",
		Parse,
	)
	if err != nil {
		panic(err)
	}
	return tl
}
