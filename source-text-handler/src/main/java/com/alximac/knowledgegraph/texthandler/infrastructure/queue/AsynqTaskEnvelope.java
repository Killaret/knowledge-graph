package com.alximac.knowledgegraph.texthandler.infrastructure.queue;

import com.fasterxml.jackson.annotation.JsonProperty;

public record AsynqTaskEnvelope(
        @JsonProperty("ID") String id,
        @JsonProperty("Queue") String queue,
        @JsonProperty("Payload") String payload,
        @JsonProperty("Retry") int retry,
        @JsonProperty("State") int state
) {
}
