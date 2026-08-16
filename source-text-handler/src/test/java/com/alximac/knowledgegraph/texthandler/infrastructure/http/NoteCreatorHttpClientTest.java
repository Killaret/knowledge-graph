package com.alximac.knowledgegraph.texthandler.infrastructure.http;

import com.alximac.knowledgegraph.texthandler.application.exception.RemoteServiceException;
import com.alximac.knowledgegraph.texthandler.domain.model.DocumentChunk;
import com.alximac.knowledgegraph.texthandler.domain.model.Link;
import com.fasterxml.jackson.databind.ObjectMapper;
import com.fasterxml.jackson.databind.json.JsonMapper;
import com.github.tomakehurst.wiremock.WireMockServer;
import com.github.tomakehurst.wiremock.core.WireMockConfiguration;
import org.junit.jupiter.api.*;
import java.util.Map;
import static com.github.tomakehurst.wiremock.client.WireMock.*;
import static org.assertj.core.api.Assertions.*;

class NoteCreatorHttpClientTest {

    private static WireMockServer wireMockServer;
    private NoteCreatorHttpClient client;
    private ObjectMapper objectMapper;

    @BeforeAll
    static void startWireMock() {
        wireMockServer = new WireMockServer(WireMockConfiguration.options().dynamicPort());
        wireMockServer.start();
    }

    @AfterAll
    static void stopWireMock() {
        if (wireMockServer != null) {
            wireMockServer.stop();
        }
    }

    @BeforeEach
    void setUp() {
        objectMapper = JsonMapper.builder().build();
        String baseUrl = "http://localhost:" + wireMockServer.port();
        // Создаём клиент без Retry/CB – мы тестируем именно его голую логику
        client = new NoteCreatorHttpClient(
                java.net.http.HttpClient.newHttpClient(),
                objectMapper,
                baseUrl
        );
        wireMockServer.resetAll();
    }

    @Test
    @DisplayName("Must create note successfully when backend returns 200")
    void mustCreateNoteSuccessfully() throws Exception {
        // Готовим WireMock: POST /notes возвращает 200 с телом {"id":"note-123"}
        wireMockServer.stubFor(post(urlEqualTo("/notes"))
                .willReturn(aResponse()
                        .withStatus(200)
                        .withHeader("Content-Type", "application/json")
                        .withBody("{\"id\":\"note-123\"}")));

        DocumentChunk chunk = new DocumentChunk("test text", 0, Map.of(), null);
        String noteId = client.createNote(chunk);

        assertThat(noteId).isEqualTo("note-123");
    }

    @Test
    @DisplayName("Must throw RemoteServiceException when backend returns 500")
    void mustThrowExceptionOn500() {
        wireMockServer.stubFor(post(urlEqualTo("/notes"))
                .willReturn(aResponse().withStatus(500)));

        DocumentChunk chunk = new DocumentChunk("test", 0, Map.of(), null);
        assertThatThrownBy(() -> client.createNote(chunk))
                .isInstanceOf(RemoteServiceException.class)
                .hasMessageContaining("HTTP 500");
    }

    @Test
    @DisplayName("Must send correct JSON in request body")
    void mustSendCorrectJson() throws Exception {
        wireMockServer.stubFor(post(urlEqualTo("/notes"))
                .willReturn(aResponse()
                        .withStatus(200)
                        .withHeader("Content-Type", "application/json")
                        .withBody("{\"id\":\"note-456\"}")));

        DocumentChunk chunk = new DocumentChunk("hello world", 2, Map.of("key", "value"), null);
        client.createNote(chunk);

        wireMockServer.verify(postRequestedFor(urlEqualTo("/notes"))
                .withRequestBody(containing("\"title\""))
                .withRequestBody(containing("\"content\":\"hello world\""))
                .withRequestBody(containing("\"metadata\":{\"key\":\"value\"}")));
    }

    @Test
    @DisplayName("Must throw RemoteServiceException when response JSON is invalid")
    void mustThrowExceptionOnInvalidResponseJson() {
        wireMockServer.stubFor(post(urlEqualTo("/notes"))
                .willReturn(aResponse()
                        .withStatus(200)
                        .withHeader("Content-Type", "application/json")
                        .withBody("not a json")));

        DocumentChunk chunk = new DocumentChunk("test", 0, Map.of(), null);
        assertThatThrownBy(() -> client.createNote(chunk))
                .isInstanceOf(RemoteServiceException.class)
                .hasMessageContaining("Failed to deserialize response");
    }

    @Test
    @DisplayName("Must extract note id from response")
    void mustExtractNoteIdFromResponse() throws Exception {
        wireMockServer.stubFor(post(urlEqualTo("/notes"))
                .willReturn(aResponse()
                        .withStatus(200)
                        .withHeader("Content-Type", "application/json")
                        .withBody("{\"id\":\"my-id\"}")));

        String id = client.createNote(new DocumentChunk("txt", 0, Map.of(), null));
        assertThat(id).isEqualTo("my-id");
    }

    @Test
    @DisplayName("Must create link successfully when backend returns 200")
    void mustCreateLinkSuccessfully() throws Exception {
        wireMockServer.stubFor(post(urlEqualTo("/links"))
                .willReturn(aResponse()
                        .withStatus(200)));

        Link link = new Link("src-1", "tgt-1", 0.8);
        assertThatCode(() -> client.createLink(link))
                .doesNotThrowAnyException();
    }

    @Test
    @DisplayName("Must throw RemoteServiceException when link creation fails with 500")
    void mustThrowExceptionOnLink500() {
        wireMockServer.stubFor(post(urlEqualTo("/links"))
                .willReturn(aResponse().withStatus(500)));

        Link link = new Link("src-1", "tgt-1", 0.8);
        assertThatThrownBy(() -> client.createLink(link))
                .isInstanceOf(RemoteServiceException.class)
                .hasMessageContaining("HTTP 500");
    }

    @Test
    @DisplayName("Must send correct JSON body for link creation")
    void mustSendCorrectJsonForLink() throws Exception {
        wireMockServer.stubFor(post(urlEqualTo("/links"))
                .willReturn(aResponse().withStatus(200)));

        Link link = new Link("src", "tgt", 0.75);
        client.createLink(link);

        wireMockServer.verify(postRequestedFor(urlEqualTo("/links"))
                .withRequestBody(containing("\"sourceNoteId\":\"src\""))
                .withRequestBody(containing("\"targetNoteId\":\"tgt\""))
                .withRequestBody(containing("\"weight\":0.75")));
    }
}