package document

import (
	"context"

	"apigo.cc/go/file"
	"apigo.cc/go/jsmod"
)

func init() {
	jsmod.Register("document", map[string]any{
		"open": func(ctx context.Context, filename string, password ...string) (*jsDocument, error) {
			p, err := file.VerifyPathForSafeMode(ctx, filename)
			if err != nil {
				return nil, err
			}
			doc, err := Open(p, password...)
			if err != nil {
				return nil, err
			}
			return &jsDocument{ctx: ctx, d: doc}, nil
		},
		"create": func(ctx context.Context, ext string) (*jsDocument, error) {
			doc, err := Create(ext)
			if err != nil {
				return nil, err
			}
			return &jsDocument{ctx: ctx, d: doc}, nil
		},
		"newExcel": func(ctx context.Context) *jsDocument {
			return &jsDocument{ctx: ctx, d: NewExcel()}
		},
		"newGraph": func(ctx context.Context) *jsDocument {
			return &jsDocument{ctx: ctx, d: NewGraph()}
		},
	})
}

type jsDocument struct {
	ctx context.Context
	d   Document
}

func (j *jsDocument) ToJSON() string     { return j.d.ToJSON() }
func (j *jsDocument) ToMarkdown() string { return j.d.ToMarkdown() }

func (j *jsDocument) Save(filename ...string) error {
	var targetPath string
	if len(filename) > 0 {
		p, err := file.VerifyPathForSafeMode(j.ctx, filename[0])
		if err != nil {
			return err
		}
		targetPath = p
	}
	return j.d.Save(targetPath)
}

// Excel 增强方法 (如果底层是 Excel)
func (j *jsDocument) Get(sheetName string, start, end string) ([][]any, error) {
	if x, ok := j.d.(*Excel); ok {
		return x.Get(sheetName, start, end)
	}
	return nil, nil
}

func (j *jsDocument) GetData(sheetName string, start, end string) ([]map[string]any, error) {
	if x, ok := j.d.(*Excel); ok {
		return x.GetData(sheetName, start, end)
	}
	return nil, nil
}

func (j *jsDocument) Set(sheetName string, table [][]any, start, end string) error {
	if x, ok := j.d.(*Excel); ok {
		return x.Set(sheetName, table, start, end)
	}
	return nil
}

func (j *jsDocument) SetData(sheetName string, data []map[string]any, start, end string) error {
	if x, ok := j.d.(*Excel); ok {
		return x.SetData(sheetName, data, start, end)
	}
	return nil
}

func (j *jsDocument) Sheets() []string {
	if x, ok := j.d.(*Excel); ok {
		return x.Sheets()
	}
	return nil
}
