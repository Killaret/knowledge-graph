package com.alximac.knowledgegraph.texthandler.application.port;

import com.alximac.knowledgegraph.texthandler.domain.model.ImportResult;

public interface OutboundQueuePort {

    void publish (ImportResult result);//контракт на отправку финального результата во внешнюю очередь.
                                       // асинхронная очередь гарантирует доставку сама.
}
