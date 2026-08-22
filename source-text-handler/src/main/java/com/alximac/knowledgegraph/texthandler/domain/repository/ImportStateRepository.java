package com.alximac.knowledgegraph.texthandler.domain.repository;

import com.alximac.knowledgegraph.texthandler.domain.model.ImportResult;

public interface ImportStateRepository {// idempotency

     /*boolean isProcessed(String eventId);*/
     void markProcessed(String eventId, ImportResult result);

     boolean tryClaim(String eventId);

}
