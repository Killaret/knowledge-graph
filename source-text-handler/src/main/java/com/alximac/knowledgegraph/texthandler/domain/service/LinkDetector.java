package com.alximac.knowledgegraph.texthandler.domain.service;

import com.alximac.knowledgegraph.texthandler.domain.model.DocumentChunk;
import com.alximac.knowledgegraph.texthandler.domain.model.Link;

import java.util.List;

//находит связи между чанками. В MVP будет NoOpLinkDetector (не создаёт связей), позже можно подложить умный детектор.
public interface LinkDetector {

    List<Link> detectLinks(List<DocumentChunk> chunks);
}
