package core

import (
	"fmt"

	hftokenizer "github.com/sugarme/tokenizer/pretrained"
	onnxruntime "github.com/yalue/onnxruntime_go"
)

type Summarizer struct {
	encoderSession *onnxruntime.DynamicAdvancedSession
	decoderSession *onnxruntime.DynamicAdvancedSession
	tokenizer      tokenizerWrapper
	definition     summaryModelDefinition
	encoderInputs  []onnxruntime.InputOutputInfo
	encoderOutputs []onnxruntime.InputOutputInfo
	decoderInputs  []onnxruntime.InputOutputInfo
	decoderOutputs []onnxruntime.InputOutputInfo
}

func NewSummarizer(info SummaryModelInfo, options RuntimeOptions) (*Summarizer, error) {
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

	encoderInputs, encoderOutputs, err := onnxruntime.GetInputOutputInfoWithOptions(info.EncoderPath, sessionOptions)
	if err != nil {
		return nil, err
	}
	decoderInputs, decoderOutputs, err := onnxruntime.GetInputOutputInfoWithOptions(info.DecoderPath, sessionOptions)
	if err != nil {
		return nil, err
	}

	encoderSession, err := onnxruntime.NewDynamicAdvancedSession(
		info.EncoderPath,
		inputNames(encoderInputs),
		outputNames(encoderOutputs),
		sessionOptions,
	)
	if err != nil {
		return nil, err
	}

	decoderSession, err := onnxruntime.NewDynamicAdvancedSession(
		info.DecoderPath,
		inputNames(decoderInputs),
		outputNames(decoderOutputs),
		sessionOptions,
	)
	if err != nil {
		encoderSession.Destroy()
		return nil, err
	}

	return &Summarizer{
		encoderSession: encoderSession,
		decoderSession: decoderSession,
		tokenizer:      tokenizerWrapper{tokenizer: tokenizer},
		definition:     info.Definition,
		encoderInputs:  encoderInputs,
		encoderOutputs: encoderOutputs,
		decoderInputs:  decoderInputs,
		decoderOutputs: decoderOutputs,
	}, nil
}

func (s *Summarizer) Close() error {
	var err error

	if s.encoderSession != nil {
		err = s.encoderSession.Destroy()
		s.encoderSession = nil
	}
	if s.decoderSession != nil {
		if destroyErr := s.decoderSession.Destroy(); err == nil {
			err = destroyErr
		}
		s.decoderSession = nil
	}

	return err
}

func (s *Summarizer) Summarize(text string) (string, error) {
	normalized := NormalizeText(text)
	if normalized == "" {
		return "", fmt.Errorf("cannot summarize empty text")
	}

	encoding, err := s.tokenizer.EncodeSingle(s.definition.PromptPrefix+normalized, true)
	if err != nil {
		return "", err
	}

	ids, attentionMask, _ := truncateEncoding(encoding, s.definition.MaxInputTokens)
	encoderInputs, encoderCleanup, err := buildEncoderInputs(s.encoderInputs, ids, attentionMask)
	if err != nil {
		return "", err
	}
	defer encoderCleanup()

	encoderOutputs := make([]onnxruntime.Value, len(s.encoderOutputs))
	if err := s.encoderSession.Run(encoderInputs, encoderOutputs); err != nil {
		return "", err
	}

	encoderHidden, encoderIndex, err := firstFloatTensorWithIndex(encoderOutputs)
	if err != nil {
		destroyValues(encoderOutputs)
		return "", err
	}
	defer encoderHidden.Destroy()
	for i, value := range encoderOutputs {
		if i != encoderIndex && value != nil {
			value.Destroy()
		}
	}

	generated := []int{s.definition.DecoderStartTokenID}

	for range s.definition.MaxOutputTokens {
		decoderInputs, decoderCleanup, err := buildDecoderInputs(
			s.decoderInputs,
			generated,
			attentionMask,
			encoderHidden,
		)
		if err != nil {
			return "", err
		}

		decoderOutputs := make([]onnxruntime.Value, len(s.decoderOutputs))
		runErr := s.decoderSession.Run(decoderInputs, decoderOutputs)
		decoderCleanup()
		if runErr != nil {
			return "", runErr
		}

		logitsTensor, _, err := firstFloatTensorWithIndex(decoderOutputs)
		if err != nil {
			destroyValues(decoderOutputs)
			return "", err
		}

		nextToken, err := nextTokenFromLogits(logitsTensor.GetData(), len(generated))
		destroyValues(decoderOutputs)
		if err != nil {
			return "", err
		}

		if nextToken == s.definition.EOSTokenID {
			break
		}
		generated = append(generated, nextToken)
	}

	if len(generated) == 1 {
		return "", fmt.Errorf("summary model returned no tokens")
	}

	summary := NormalizeText(s.tokenizer.Decode(generated[1:], true))
	if summary == "" {
		return "", fmt.Errorf("summary model returned empty text")
	}

	return summary, nil
}

func buildEncoderInputs(
	inputInfos []onnxruntime.InputOutputInfo,
	ids []int,
	attentionMask []int,
) ([]onnxruntime.Value, func(), error) {
	values := make([]onnxruntime.Value, len(inputInfos))
	cleanup := func() {
		destroyValues(values)
	}

	for i, inputInfo := range inputInfos {
		switch inputInfo.Name {
		case "input_ids":
			tensor, err := newIntTensor(ids)
			if err != nil {
				cleanup()
				return nil, nil, err
			}
			values[i] = tensor
		case "attention_mask":
			tensor, err := newIntTensor(attentionMask)
			if err != nil {
				cleanup()
				return nil, nil, err
			}
			values[i] = tensor
		default:
			cleanup()
			return nil, nil, fmt.Errorf("unsupported summary encoder input %q", inputInfo.Name)
		}
	}

	return values, cleanup, nil
}

func buildDecoderInputs(
	inputInfos []onnxruntime.InputOutputInfo,
	generated []int,
	attentionMask []int,
	encoderHidden onnxruntime.Value,
) ([]onnxruntime.Value, func(), error) {
	values := make([]onnxruntime.Value, len(inputInfos))
	cleanup := func() {
		for _, value := range values {
			if value != nil && value != encoderHidden {
				value.Destroy()
			}
		}
	}

	for i, inputInfo := range inputInfos {
		switch inputInfo.Name {
		case "input_ids":
			tensor, err := newIntTensor(generated)
			if err != nil {
				cleanup()
				return nil, nil, err
			}
			values[i] = tensor
		case "encoder_hidden_states":
			values[i] = encoderHidden
		case "encoder_attention_mask":
			tensor, err := newIntTensor(attentionMask)
			if err != nil {
				cleanup()
				return nil, nil, err
			}
			values[i] = tensor
		default:
			cleanup()
			return nil, nil, fmt.Errorf("unsupported summary decoder input %q", inputInfo.Name)
		}
	}

	return values, cleanup, nil
}

func inputNames(values []onnxruntime.InputOutputInfo) []string {
	names := make([]string, 0, len(values))
	for _, value := range values {
		names = append(names, value.Name)
	}
	return names
}

func outputNames(values []onnxruntime.InputOutputInfo) []string {
	names := make([]string, 0, len(values))
	for _, value := range values {
		names = append(names, value.Name)
	}
	return names
}

func firstFloatTensorWithIndex(values []onnxruntime.Value) (*onnxruntime.Tensor[float32], int, error) {
	for i, value := range values {
		if value == nil {
			continue
		}
		if tensor, ok := value.(*onnxruntime.Tensor[float32]); ok {
			return tensor, i, nil
		}
	}

	return nil, -1, fmt.Errorf("model returned no float32 tensor output")
}

func nextTokenFromLogits(logits []float32, sequenceLength int) (int, error) {
	if sequenceLength == 0 || len(logits)%sequenceLength != 0 {
		return 0, fmt.Errorf("unexpected decoder output shape")
	}

	vocabSize := len(logits) / sequenceLength
	start := (sequenceLength - 1) * vocabSize

	bestIndex := 0
	bestValue := logits[start]
	for i := 1; i < vocabSize; i++ {
		value := logits[start+i]
		if value > bestValue {
			bestValue = value
			bestIndex = i
		}
	}

	return bestIndex, nil
}
