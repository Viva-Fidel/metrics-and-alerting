package config

import "dario.cat/mergo"

func mergeOverlays[T any](layers ...*T) (*T, error) {
	merged := new(T)
	for _, layer := range layers {
		if layer == nil {
			continue
		}
		if err := mergo.Merge(merged, layer, mergo.WithOverride); err != nil {
			return nil, err
		}
	}

	return merged, nil
}
