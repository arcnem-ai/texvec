package core

import (
	"fmt"

	"github.com/sugarme/tokenizer"
	hftokenizer "github.com/sugarme/tokenizer/pretrained"
	onnxruntime "github.com/yalue/onnxruntime_go"
)

type TextEmbedder struct {
	session     *onnxruntime.DynamicAdvancedSession
	tokenizer   tokenizerLike
	definition  embeddingModelDefinition
	inputInfos  []onnxruntime.InputOutputInfo
	outputInfos []onnxruntime.InputOutputInfo
}

type tokenizerLike interface {
	EncodeSingle(input string, addSpecialTokensOpt ...bool) (*EncodingAdapter, error)
	Decode(ids []int, skipSpecialTokens bool) string
}

type tokenizerWrapper struct {
	tokenizer *tokenizer.Tokenizer
}

type EncodingAdapter struct {
	IDs           []int
	AttentionMask []int
	TypeIDs       []int
}

func (t tokenizerWrapper) EncodeSingle(input string, addSpecialTokensOpt ...bool) (*EncodingAdapter, error) {
	encoding, err := t.tokenizer.EncodeSingle(input, addSpecialTokensOpt...)
	if err != nil {
		return nil, err
	}

	return &EncodingAdapter{
		IDs:           encoding.GetIds(),
		AttentionMask: encoding.GetAttentionMask(),
		TypeIDs:       encoding.GetTypeIds(),
	}, nil
}

func (t tokenizerWrapper) Decode(ids []int, skipSpecialTokens bool) string {
	return t.tokenizer.Decode(ids, skipSpecialTokens)
}

func NewTextEmbedder(info EmbeddingModelInfo, options RuntimeOptions) (*TextEmbedder, error) {
	if err := ensureONNXRuntime(); err != nil {
		return nil, err
	}

	tokenizer, err := hftokenizer.FromFile(info.TokenizerPath)
	if err != nil {
		return nil, fmt.Errorf("load tokenizer: %w", err)
	}

	sessionOptions, err := newSessionOptions(options)
	if err != nil {
		return nil, err
	}
	defer sessionOptions.Destroy()

	inputs, outputs, err := onnxruntime.GetInputOutputInfoWithOptions(info.ModelPath, sessionOptions)
	if err != nil {
		return nil, fmt.Errorf("read model input/output info: %w", err)
	}

	inputNames := make([]string, 0, len(inputs))
	for _, input := range inputs {
		inputNames = append(inputNames, input.Name)
	}

	outputNames := make([]string, 0, len(outputs))
	for _, output := range outputs {
		outputNames = append(outputNames, output.Name)
	}

	session, err := onnxruntime.NewDynamicAdvancedSession(info.ModelPath, inputNames, outputNames, sessionOptions)
	if err != nil {
		return nil, err
	}

	return &TextEmbedder{
		session:     session,
		tokenizer:   tokenizerWrapper{tokenizer: tokenizer},
		definition:  info.Definition,
		inputInfos:  inputs,
		outputInfos: outputs,
	}, nil
}

func (e *TextEmbedder) Close() error {
	if e.session == nil {
		return nil
	}

	err := e.session.Destroy()
	e.session = nil
	return err
}

func (e *TextEmbedder) EmbedDocument(text string) ([]float32, error) {
	return e.embed(text)
}

func (e *TextEmbedder) EmbedQuery(text string) ([]float32, error) {
	if e.definition.QueryPrefix != "" {
		text = e.definition.QueryPrefix + text
	}
	return e.embed(text)
}

func (e *TextEmbedder) embed(text string) ([]float32, error) {
	normalized := NormalizeText(text)
	if normalized == "" {
		return nil, fmt.Errorf("cannot embed empty text")
	}

	encoding, err := e.tokenizer.EncodeSingle(normalized, true)
	if err != nil {
		return nil, err
	}

	ids, attentionMask, typeIDs := truncateEncoding(encoding, e.definition.MaxTokens)
	inputs, cleanup, err := buildBERTInputs(e.inputInfos, ids, attentionMask, typeIDs)
	if err != nil {
		return nil, err
	}
	defer cleanup()

	outputs := make([]onnxruntime.Value, len(e.outputInfos))
	if err := e.session.Run(inputs, outputs); err != nil {
		return nil, err
	}
	defer destroyValues(outputs)

	outputTensor, err := firstFloatTensor(outputs)
	if err != nil {
		return nil, err
	}

	data := outputTensor.GetData()
	if len(data) == 0 {
		return nil, fmt.Errorf("embedding model returned empty output")
	}

	if len(data) == e.definition.EmbeddingDim {
		return L2Normalize(data), nil
	}

	seqLen := len(ids)
	if seqLen == 0 || len(data)%seqLen != 0 {
		return nil, fmt.Errorf("unexpected embedding output shape")
	}

	hiddenSize := len(data) / seqLen
	pooled := make([]float32, hiddenSize)
	var count float32

	for i := range seqLen {
		if attentionMask[i] == 0 {
			continue
		}
		count++
		rowStart := i * hiddenSize
		for j := range hiddenSize {
			pooled[j] += data[rowStart+j]
		}
	}

	if count == 0 {
		return nil, fmt.Errorf("attention mask removed all tokens")
	}

	for i := range pooled {
		pooled[i] /= count
	}

	return L2Normalize(pooled), nil
}
