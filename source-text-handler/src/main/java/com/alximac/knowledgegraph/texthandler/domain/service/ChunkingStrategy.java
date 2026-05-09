package com.alximac.knowledgegraph.texthandler.domain.service;

import com.alximac.knowledgegraph.texthandler.domain.model.DocumentChunk;
import com.alximac.knowledgegraph.texthandler.domain.model.ImportOptions;
import java.util.List;

//нарезает текст на DocumentChunk с учётом ImportOptions.
// Отвечает за структурный чанкинг, fallback, фильтрацию по minChunkLength ->ImportOptions.

public interface ChunkingStrategy {

    List<DocumentChunk> chunk(String text, ImportOptions options);

}
