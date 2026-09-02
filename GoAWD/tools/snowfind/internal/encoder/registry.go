package encoder

// Registry 编码器注册表
type Registry struct {
	encoders []Encoder
}

// NewRegistry 创建新的编码器注册表
func NewRegistry() *Registry {
	return &Registry{
		encoders: make([]Encoder, 0),
	}
}

// RegisterDefaultEncoders 注册所有默认编码器
func (r *Registry) RegisterDefaultEncoders() {
	r.Register(NewPlainTextEncoder())
	r.Register(NewHexEncoder())
	r.Register(NewHexXEncoder())
	r.Register(NewBase64Encoder())
	r.Register(NewURLEncoder())
	r.Register(NewASCIIEncoder())
	r.Register(NewASCIINoSpaceEncoder())
	r.Register(NewUnicodeEncoder())
	r.Register(NewUnicodeHTMLEntityEncoder())
	r.Register(NewBinaryEncoder())
	r.Register(NewBinaryNoSpaceEncoder())

	// 新增的编码器
	r.Register(NewROT13Encoder())
	r.Register(NewAtbashEncoder())
	r.Register(NewBase32Encoder())
	r.Register(NewBase58Encoder())
	r.Register(NewOctalEncoder())
	r.Register(NewReverseEncoder())
	r.Register(NewJSFuckEncoder())
	r.Register(NewMorseCodeEncoder())
}

// Register 注册编码器
func (r *Registry) Register(encoder Encoder) {
	r.encoders = append(r.encoders, encoder)
}

// GetEncoders 获取所有注册的编码器
func (r *Registry) GetEncoders() []Encoder {
	return r.encoders
}

// GetEncoder 根据名称获取编码器
func (r *Registry) GetEncoder(name string) Encoder {
	for _, encoder := range r.encoders {
		if encoder.Name() == name {
			return encoder
		}
	}
	return nil
}

// ListEncoders 列出所有编码器信息
func (r *Registry) ListEncoders() []string {
	var list []string
	for _, encoder := range r.encoders {
		list = append(list, encoder.Name()+": "+encoder.Description())
	}
	return list
}
