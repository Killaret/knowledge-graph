package com.alximac.knowledgegraph.texthandler.domain.service;

import com.alximac.knowledgegraph.texthandler.domain.model.DocumentChunk;
import com.alximac.knowledgegraph.texthandler.domain.model.ImportOptions;
import java.util.List;

// Cuts text into DocumentChunk considering ImportOptions.
// Responsible for structural chunking, fallback, filtering by minChunkLength -> ImportOptions.

public interface ChunkingStrategy {

    List<DocumentChunk> chunk(String text, ImportOptions options);

}
