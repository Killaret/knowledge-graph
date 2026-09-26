package main

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"knowledge-graph/internal/config"
	"knowledge-graph/internal/infrastructure/mongo"
	"knowledge-graph/internal/infrastructure/queue"
)

// NOTE-QUALITY-1 criterion 6: with nlp.quality.enabled unset the pipeline
// is not installed even when Mongo storage is present — no assessor, so no
// task is ever acted on and nothing reaches the quality_log collection.
func TestInstallQualityPipeline_FlagOff(t *testing.T) {
	w := queue.NewWorker(nil, nil, nil, nil, nil, nil, nil, nil, 0, nil, false, "")
	cfg := &config.Config{NLPQualityEnabled: false}
	assert.False(t, installQualityPipeline(w, cfg, &mongo.Client{}, nil, nil, nil, nil, nil))
}

func TestInstallQualityPipeline_NoMongo(t *testing.T) {
	w := queue.NewWorker(nil, nil, nil, nil, nil, nil, nil, nil, 0, nil, false, "")
	cfg := &config.Config{NLPQualityEnabled: true}
	assert.False(t, installQualityPipeline(w, cfg, nil, nil, nil, nil, nil, nil))
}
