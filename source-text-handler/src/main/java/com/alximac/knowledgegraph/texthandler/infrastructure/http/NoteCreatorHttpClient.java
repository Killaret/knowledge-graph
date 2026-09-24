package com.alximac.knowledgegraph.texthandler.infrastructure.http;

import com.alximac.knowledgegraph.texthandler.application.port.NoteCreatorPort;
import com.alximac.knowledgegraph.texthandler.application.exception.RemoteServiceException;
import com.alximac.knowledgegraph.texthandler.domain.model.DocumentChunk;
import com.alximac.knowledgegraph.texthandler.domain.model.Link;
import com.fasterxml.jackson.core.JsonProcessingException;
import com.fasterxml.jackson.databind.ObjectMapper;

import java.net.URI;
import java.net.http.HttpClient;
import java.net.http.HttpRequest;
import java.net.http.HttpResponse;
import java.io.IOException;
import java.time.Duration;

public class NoteCreatorHttpClient implements NoteCreatorPort {

    private final HttpClient httpClient;
    private final ObjectMapper objectMapper;
    private final String notesUrl;
    private final String linksUrl;

    public NoteCreatorHttpClient(HttpClient httpClient, ObjectMapper objectMapper,
                                 String baseUrl) {
        this.httpClient = httpClient;
        this.objectMapper = objectMapper;
        this.notesUrl = baseUrl + "/api/v1/notes";
        this.linksUrl = baseUrl + "/api/v1/links";
    }

    private static final long WAIT_DURATION = 5;

    @Override
    public String createNote(DocumentChunk chunk,String jwt) throws RemoteServiceException {
        CreateNoteRequest request = new CreateNoteRequest(
                generateTitle(chunk),
                chunk.text(),
                chunk.type(),
                chunk.metadata());

        String json = toJson(request);   // сериализация с обработкой ошибки
        HttpResponse<String> response = executePost(notesUrl, json,jwt);
        CreateNoteResponse noteResponse = fromJson(response.body(), CreateNoteResponse.class);
        return noteResponse.data().id();

    }

    @Override
    public void createLink(Link link,String jwt) throws RemoteServiceException {
        CreateLinkRequest request = new CreateLinkRequest(
                link.sourceNoteId(),
                link.targetNoteId(),
                link.linkType(),
                link.weight());

        String json = toJson(request);
        executePost(linksUrl, json,jwt);   // ответ не важен
    }

    private HttpResponse<String> executePost(String url, String json,String jwt) throws RemoteServiceException {
        try {
            HttpRequest.Builder builder = HttpRequest.newBuilder()
                    .uri(URI.create(url))
                    .header("Content-Type", "application/json")
                    .timeout(Duration.ofSeconds(WAIT_DURATION));

            if (jwt != null && !jwt.isBlank()) {
                builder.header("Authorization", "Bearer " + jwt);
            }

            HttpRequest httpRequest = builder
                    .POST(HttpRequest.BodyPublishers.ofString(json))
                    .build();


            HttpResponse<String> response = httpClient.send(httpRequest, HttpResponse.BodyHandlers.ofString());

            if (response.statusCode() != 200 && response.statusCode() != 201) {
                throw new RemoteServiceException(
                        "Request failed: HTTP " + response.statusCode() + " " + response.body());
            }
            return response;
        } catch (IOException | InterruptedException e) {
            throw new RemoteServiceException("HTTP request failed", e);
        }
    }

    private String toJson(Object obj) throws RemoteServiceException {
        try {
            return objectMapper.writeValueAsString(obj);
        } catch (JsonProcessingException e) {
            throw new RemoteServiceException("Failed to serialize request", e);
        }
    }

    private <T> T fromJson(String json, Class<T> type) throws RemoteServiceException {
        try {
            return objectMapper.readValue(json, type);
        } catch (JsonProcessingException e) {
            throw new RemoteServiceException("Failed to deserialize response", e);
        }
    }

    private String generateTitle(DocumentChunk chunk) {
        String text = chunk.text();
        // заголовок делается
        return text.length() > 100 ? text.substring(0, 100) : text;
    }
}