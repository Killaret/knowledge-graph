package com.alximac.knowledgegraph.texthandler.domain.model;

import org.junit.jupiter.api.DisplayName;
import org.junit.jupiter.api.Test;
import static org.assertj.core.api.Assertions.*;

class LinkTest {

    @Test
    @DisplayName("Must create link with valid values")
    void mustCreateWithValidValues() {
        Link link = new Link("src-1", "tgt-1", 0.75);
        assertThat(link.sourceNoteId()).isEqualTo("src-1");
        assertThat(link.targetNoteId()).isEqualTo("tgt-1");
        assertThat(link.weight()).isEqualTo(0.75);
    }

    @Test
    @DisplayName("Must reject blank sourceNoteId")
    void mustRejectBlankSourceNoteId() {
        assertThatIllegalArgumentException()
                .isThrownBy(() -> new Link(null, "tgt", 1.0));
        assertThatIllegalArgumentException()
                .isThrownBy(() -> new Link("  ", "tgt", 1.0));
    }

    @Test
    @DisplayName("Must reject blank targetNoteId")
    void mustRejectBlankTargetNoteId() {
        assertThatIllegalArgumentException()
                .isThrownBy(() -> new Link("src", null, 1.0));
    }

    @Test
    @DisplayName("Must reject weight out of range")
    void mustRejectWeightOutOfRange() {
        assertThatIllegalArgumentException()
                .isThrownBy(() -> new Link("src", "tgt", -0.1));
        assertThatIllegalArgumentException()
                .isThrownBy(() -> new Link("src", "tgt", 1.1));
    }

    @Test
    @DisplayName("Must accept boundary weights")
    void mustAcceptBoundaryWeights() {
        assertThatCode(() -> new Link("src", "tgt", 0.0)).doesNotThrowAnyException();
        assertThatCode(() -> new Link("src", "tgt", 1.0)).doesNotThrowAnyException();
    }
}