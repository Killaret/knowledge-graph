package com.alximac.knowledgegraph.texthandler.application.port;

import com.alximac.knowledgegraph.texthandler.domain.model.ImportTask;

import java.util.function.Consumer;

public interface InboundQueuePort {

    void subscribe(Consumer<ImportTask> handler);//точка входа для непрерывного получения задач.
    // бесконечный цикл (или выделенный поток), который слушает очередь import:document через Lettuce .
}
