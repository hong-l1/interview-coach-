package ragkit

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/cloudwego/eino-ext/components/document/loader/file"
	"github.com/cloudwego/eino-ext/components/document/parser/docx"
	"github.com/cloudwego/eino-ext/components/document/parser/pdf"
	"github.com/cloudwego/eino/components/document"
	"github.com/cloudwego/eino/components/document/parser"
	"github.com/cloudwego/eino/schema"
)

// NewDocumentsLoader 加载单个文件（pdf/docx 走专用解析器，其余走 TextParser），
// 文件名作为文档 ID。文件不存在时 panic（沿用 rag 包行为）。
func NewDocumentsLoader(ctx context.Context, path string) ([]*schema.Document, error) {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		panic(fmt.Sprintf("文件不存在: %s, 错误: %v", path, err))
	}
	start := time.Now()
	pdfParser, err := pdf.NewPDFParser(ctx, &pdf.Config{})
	if err != nil {
		panic(fmt.Sprintf("PDF解析器创建失败: %v, 耗时: %v", err, time.Since(start)))
	}
	docxParser, err := docx.NewDocxParser(ctx, &docx.Config{})
	if err != nil {
		panic(fmt.Sprintf("DOCX解析器创建失败: %v, <UNK>: %v", err, time.Since(start)))
	}
	extParser, err := parser.NewExtParser(ctx, &parser.ExtParserConfig{
		Parsers: map[string]parser.Parser{
			".pdf":  pdfParser,
			".docx": docxParser,
		},
		FallbackParser: parser.TextParser{},
	})
	if err != nil {
		panic(fmt.Sprintf("扩展解析器创建失败: %v, 耗时: %v", err, time.Since(start)))
	}
	loader, err := file.NewFileLoader(ctx, &file.FileLoaderConfig{
		UseNameAsID: true,
		Parser:      extParser,
	})
	if err != nil {
		panic(fmt.Sprintf("文件加载器创建失败: %v, 耗时: %v", err, time.Since(start)))
	}
	fmt.Printf("文件加载器创建成功，耗时: %v\n", time.Since(start))
	ctxWithTimeout, cancel := context.WithTimeout(ctx, 60*time.Second)
	defer cancel()
	src := document.Source{URI: path}
	return loader.Load(ctxWithTimeout, src)
}
