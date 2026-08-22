package com.alximac.knowledgegraph.texthandler.domain.service;

import com.alximac.knowledgegraph.texthandler.domain.model.DocumentChunk;
import com.alximac.knowledgegraph.texthandler.domain.model.Link;

import java.util.List;

// Finds links between chunks. In MVP this is NoOpLinkDetector (creates no links),
// later can be replaced with a smart detector.
public interface LinkDetector {

    List<Link> detectLinks(List<DocumentChunk> chunks);
}
