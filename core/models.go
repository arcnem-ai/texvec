package core

type modelAsset struct {
	FileName    string
	DownloadURL string
}

type embeddingModelDefinition struct {
	EmbeddingDim  int
	MaxTokens     int
	QueryPrefix   string
	ModelFile     string
	TokenizerFile string
	Assets        []modelAsset
}

type summaryModelDefinition struct {
	PromptPrefix        string
	MaxInputTokens      int
	MaxOutputTokens     int
	EncoderFile         string
	DecoderFile         string
	TokenizerFile       string
	DecoderStartTokenID int
	EOSTokenID          int
	Assets              []modelAsset
}

var embeddingModelRegistry = map[string]embeddingModelDefinition{
	"all-minilm-l6-v2": {
		EmbeddingDim:  384,
		MaxTokens:     256,
		ModelFile:     "model.onnx",
		TokenizerFile: "tokenizer.json",
		Assets: []modelAsset{
			{
				FileName:    "model.onnx",
				DownloadURL: "https://huggingface.co/Qdrant/all-MiniLM-L6-v2-onnx/resolve/main/model.onnx",
			},
			{
				FileName:    "tokenizer.json",
				DownloadURL: "https://huggingface.co/Qdrant/all-MiniLM-L6-v2-onnx/resolve/main/tokenizer.json",
			},
		},
	},
	"bge-small-en-v1.5": {
		EmbeddingDim:  384,
		MaxTokens:     256,
		QueryPrefix:   "Represent this sentence for searching relevant passages: ",
		ModelFile:     "model_optimized.onnx",
		TokenizerFile: "tokenizer.json",
		Assets: []modelAsset{
			{
				FileName:    "model_optimized.onnx",
				DownloadURL: "https://huggingface.co/Qdrant/bge-small-en-v1.5-onnx-Q/resolve/main/model_optimized.onnx",
			},
			{
				FileName:    "tokenizer.json",
				DownloadURL: "https://huggingface.co/Qdrant/bge-small-en-v1.5-onnx-Q/resolve/main/tokenizer.json",
			},
		},
	},
}

var summaryModelRegistry = map[string]summaryModelDefinition{
	"flan-t5-small": {
		PromptPrefix:        "summarize: ",
		MaxInputTokens:      512,
		MaxOutputTokens:     96,
		EncoderFile:         "encoder_model_int8.onnx",
		DecoderFile:         "decoder_model_quantized.onnx",
		TokenizerFile:       "tokenizer.json",
		DecoderStartTokenID: 0,
		EOSTokenID:          1,
		Assets: []modelAsset{
			{
				FileName:    "encoder_model_int8.onnx",
				DownloadURL: "https://huggingface.co/Xenova/flan-t5-small/resolve/main/onnx/encoder_model_int8.onnx",
			},
			{
				FileName:    "decoder_model_quantized.onnx",
				DownloadURL: "https://huggingface.co/Xenova/flan-t5-small/resolve/main/onnx/decoder_model_quantized.onnx",
			},
			{
				FileName:    "tokenizer.json",
				DownloadURL: "https://huggingface.co/Xenova/flan-t5-small/resolve/main/tokenizer.json",
			},
		},
	},
}

type EmbeddingModelInfo struct {
	ID            string
	ModelPath     string
	TokenizerPath string
	Definition    embeddingModelDefinition
}

type SummaryModelInfo struct {
	ID            string
	EncoderPath   string
	DecoderPath   string
	TokenizerPath string
	Definition    summaryModelDefinition
}
