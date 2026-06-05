package hook

type Kind string

const (
	KindMarkdown Kind = "markdown"
	KindBuiltin  Kind = "builtin"
)

type RendererHookContext struct {
	Kind   Kind
	Name   string
	Params Params
	Output string
}

type RendererHookFunc func(ctx RendererHookContext) string

type Params map[string]ParamValue

func (p Params) Get(name string) ParamValue {
	return p[name]
}

func (p Params) Has(name string) bool {
	_, ok := p[name]
	return ok
}

func (p Params) String(name string) (string, bool) {
	return p.Get(name).String()
}

func (p Params) Array(name string) (Array, bool) {
	return p.Get(name).Array()
}

func (p Params) Record(name string) (Record, bool) {
	return p.Get(name).Record()
}

type ParamValue struct {
	value any
}

func NewString(value string) ParamValue {
	return ParamValue{value: value}
}

func NewArray(items []ParamValue) ParamValue {
	return ParamValue{value: Array(items)}
}

func NewRecord(fields map[string]ParamValue) ParamValue {
	return ParamValue{value: Record(fields)}
}

func (v ParamValue) String() (string, bool) {
	value, ok := v.value.(string)
	return value, ok
}

func (v ParamValue) Array() (Array, bool) {
	value, ok := v.value.(Array)
	return value, ok
}

func (v ParamValue) Record() (Record, bool) {
	value, ok := v.value.(Record)
	return value, ok
}

func (v ParamValue) Any() any {
	switch value := v.value.(type) {
	case string:
		return value
	case Array:
		items := make([]any, 0, len(value))
		for _, item := range value {
			items = append(items, item.Any())
		}
		return items
	case Record:
		fields := make(map[string]any, len(value))
		for key, field := range value {
			fields[key] = field.Any()
		}
		return fields
	default:
		return nil
	}
}

type Array []ParamValue

func (a Array) Get(index int) ParamValue {
	if index < 0 || index >= len(a) {
		return ParamValue{}
	}
	return a[index]
}

func (a Array) String(index int) (string, bool) {
	return a.Get(index).String()
}

func (a Array) Array(index int) (Array, bool) {
	return a.Get(index).Array()
}

func (a Array) Record(index int) (Record, bool) {
	return a.Get(index).Record()
}

type Record map[string]ParamValue

func (r Record) Get(name string) ParamValue {
	return r[name]
}

func (r Record) Has(name string) bool {
	_, ok := r[name]
	return ok
}

func (r Record) String(name string) (string, bool) {
	return r.Get(name).String()
}

func (r Record) Array(name string) (Array, bool) {
	return r.Get(name).Array()
}

func (r Record) Record(name string) (Record, bool) {
	return r.Get(name).Record()
}
