package document

import (
	"context"

	"apigo.cc/go/file"
	"apigo.cc/go/jsmod"
)

func init() {
	jsmod.Register("document", map[string]any{
		"Open":     jsOpen,
		"Create":   jsCreate,
		"NewExcel": jsNewExcel,
		"NewGraph": jsNewGraph,
	})
}

func jsOpen(ctx context.Context, filename string, password *string) (*jsDocument, error) {
	p, err := file.VerifyPathForSafeMode(ctx, filename)
	if err != nil {
		return nil, jsmod.MakeError(err)
	}
	var doc Document
	if password != nil {
		doc, err = Open(p, *password)
	} else {
		doc, err = Open(p)
	}
	if err != nil {
		return nil, jsmod.MakeError(err)
	}
	return &jsDocument{ctx: ctx, d: doc}, nil
}

func jsCreate(ctx context.Context, ext string) (*jsDocument, error) {
	doc, err := Create(ext)
	if err != nil {
		return nil, jsmod.MakeError(err)
	}
	return &jsDocument{ctx: ctx, d: doc}, nil
}

func jsNewExcel(ctx context.Context) *jsDocument {
	return &jsDocument{ctx: ctx, d: NewExcel()}
}

func jsNewGraph(ctx context.Context) *jsDocument {
	return &jsDocument{ctx: ctx, d: NewGraph()}
}

type jsDocument struct {
	ctx context.Context
	d   Document
}

func (j *jsDocument) ToJSON() string     { return j.d.ToJSON() }
func (j *jsDocument) ToMarkdown() string { return j.d.ToMarkdown() }

func (j *jsDocument) Save(filename *string) error {
	var targetPath string
	if filename != nil && *filename != "" {
		p, err := file.VerifyPathForSafeMode(j.ctx, *filename)
		if err != nil {
			return jsmod.MakeError(err)
		}
		targetPath = p
	}
	if err := j.d.Save(targetPath); err != nil {
		return jsmod.MakeError(err)
	}
	return nil
}

// Excel 增强方法 (如果底层是 Excel)
func (j *jsDocument) Get(sheetName string, start, end string) ([][]any, error) {
	if x, ok := j.d.(*Excel); ok {
		res, err := x.Get(sheetName, start, end)
		if err != nil {
			return nil, jsmod.MakeError(err)
		}
		return res, nil
	}
	return nil, nil
}

func (j *jsDocument) GetData(sheetName string, start, end string) ([]map[string]any, error) {
	if x, ok := j.d.(*Excel); ok {
		res, err := x.GetData(sheetName, start, end)
		if err != nil {
			return nil, jsmod.MakeError(err)
		}
		return res, nil
	}
	return nil, nil
}

func (j *jsDocument) Set(sheetName string, table [][]any, start, end string) error {
	if x, ok := j.d.(*Excel); ok {
		if err := x.Set(sheetName, table, start, end); err != nil {
			return jsmod.MakeError(err)
		}
	}
	return nil
}

func (j *jsDocument) SetData(sheetName string, data []map[string]any, start, end string) error {
	if x, ok := j.d.(*Excel); ok {
		if err := x.SetData(sheetName, data, start, end); err != nil {
			return jsmod.MakeError(err)
		}
	}
	return nil
}

func (j *jsDocument) Sheets() []string {
	if x, ok := j.d.(*Excel); ok {
		return x.Sheets()
	}
	return nil
}
