package com.alximac.knowledgegraph.texthandler.infrastructure.link;

import com.alximac.knowledgegraph.texthandler.domain.model.DocumentChunk;
import com.alximac.knowledgegraph.texthandler.domain.model.Link;
import com.alximac.knowledgegraph.texthandler.domain.service.LinkDetector;

import java.util.List;

public class NoOpLinkDetector implements LinkDetector {
    @Override
    public List<Link> detectLinks(List<DocumentChunk> chunks) {
        return List.of();
    }
}
