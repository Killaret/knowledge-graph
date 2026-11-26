package com.alximac.knowledgegraph.texthandler.infrastructure.parser;

import com.alximac.knowledgegraph.texthandler.domain.model.ParsedDocument;
import com.alximac.knowledgegraph.texthandler.domain.service.DocumentParser;
import com.alximac.knowledgegraph.texthandler.domain.service.DocumentParserException;
import com.alximac.knowledgegraph.texthandler.infrastructure.validation.UrlValidator;

import java.io.IOException;
import java.io.InputStream;
import java.net.HttpURLConnection;
import java.net.MalformedURLException;
import java.net.URI;
import java.net.URL;
import java.nio.charset.StandardCharsets;
import java.util.Map;

public class TextDocumentParser implements DocumentParser {


    @Override
    public ParsedDocument parse(byte[] data, Map<String, Object> sourceMetadata) throws DocumentParserException {
        String text = new String(data, StandardCharsets.UTF_8);
        return new ParsedDocument(text, sourceMetadata != null ? sourceMetadata : Map.of());

    }

    @Override
    public ParsedDocument parseFromUrl(String url) throws DocumentParserException {
        UrlValidator.validate(url);

        URL parsedUrl;
        try {
            parsedUrl = URI.create(url).toURL();
        } catch (Exception e) {
            throw new DocumentParserException("Incorrect URL: " + url, e);
        }


        HttpURLConnection connection = null;

        try {
            connection = (HttpURLConnection) parsedUrl.openConnection();

            connection.setInstanceFollowRedirects(false);
            connection.setConnectTimeout(ParserConfigs.CONNECT_TIMEOUT);
            connection.setReadTimeout(ParserConfigs.READ_TIMEOUT);

            int statusCode = connection.getResponseCode();

            if (statusCode >= 300 && statusCode < 400) {
                throw new DocumentParserException("Redirects are not allowed. Status code: " + statusCode);
            }

            if (statusCode < 200 || statusCode >= 300) {
                throw new DocumentParserException("Failed to fetch URL. Status: " + statusCode);
            }
            try (InputStream stream = connection.getInputStream()) {
                String text = new String(stream.readAllBytes(), StandardCharsets.UTF_8);
                return new ParsedDocument(text, Map.of());
            }
        } catch (IOException | IllegalArgumentException e) {
            throw new DocumentParserException("Failed to read text from URL: " + url, e);
        } finally {
            if (connection != null) {
                connection.disconnect();
            }
        }
    }
}
