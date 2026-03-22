package core

import (
	"fmt"

	onnxruntime "github.com/yalue/onnxruntime_go"
)

func truncateEncoding(encoding *EncodingAdapter, maxTokens int) ([]int, []int, []int) {
	ids := append([]int(nil), encoding.IDs...)
	attentionMask := append([]int(nil), encoding.AttentionMask...)
	typeIDs := append([]int(nil), encoding.TypeIDs...)

	if len(attentionMask) == 0 {
		attentionMask = make([]int, len(ids))
		for i := range attentionMask {
			attentionMask[i] = 1
		}
	}

	if len(typeIDs) == 0 {
		typeIDs = make([]int, len(ids))
	}

	if maxTokens > 0 && len(ids) > maxTokens {
		ids = ids[:maxTokens]
		attentionMask = attentionMask[:maxTokens]
		typeIDs = typeIDs[:maxTokens]
	}

	return ids, attentionMask, typeIDs
}

func buildBERTInputs(
	inputInfos []onnxruntime.InputOutputInfo,
	ids []int,
	attentionMask []int,
	typeIDs []int,
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
		case "token_type_ids":
			tensor, err := newIntTensor(typeIDs)
			if err != nil {
				cleanup()
				return nil, nil, err
			}
			values[i] = tensor
		default:
			cleanup()
			return nil, nil, fmt.Errorf("unsupported embedding model input %q", inputInfo.Name)
		}
	}

	return values, cleanup, nil
}

func newIntTensor(values []int) (onnxruntime.Value, error) {
	data := make([]int64, len(values))
	for i, value := range values {
		data[i] = int64(value)
	}

	return onnxruntime.NewTensor(onnxruntime.NewShape(1, int64(len(data))), data)
}

func firstFloatTensor(values []onnxruntime.Value) (*onnxruntime.Tensor[float32], error) {
	for _, value := range values {
		if value == nil {
			continue
		}
		if tensor, ok := value.(*onnxruntime.Tensor[float32]); ok {
			return tensor, nil
		}
	}

	return nil, fmt.Errorf("model returned no float32 tensor output")
}

func destroyValues(values []onnxruntime.Value) {
	for _, value := range values {
		if value != nil {
			value.Destroy()
		}
	}
}
