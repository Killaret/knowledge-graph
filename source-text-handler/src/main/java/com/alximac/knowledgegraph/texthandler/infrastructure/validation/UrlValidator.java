package com.alximac.knowledgegraph.texthandler.infrastructure.validation;

import java.net.InetAddress;
import java.net.URI;
import java.net.URISyntaxException;
import java.net.UnknownHostException;
/**
 * Проверяет URL на SSRF-безопасность.
 * ВНИМАНИЕ: между validate() и реальным openStream() DNS может измениться (TOCTOU / DNS rebinding).
 * Полная защита требует подключения по заранее зарезолвленному IP с использованием низкоуровневого кода
 * Пока норм для проекта
 */

public class UrlValidator {
    private UrlValidator(){}

    public static void validate(String url) throws UrlValidationException {
        if (url == null || url.isBlank()) {
            throw new UrlValidationException("URL must not be null or empty");
        }

        URI uri;
        try {
            uri = new URI(url);
        } catch (URISyntaxException e) {
            throw new UrlValidationException("Invalid URL syntax: " + url, e);
        }

        // 1. Проверка схемы
        String scheme = uri.getScheme();
        if (scheme == null) {
            throw new UrlValidationException("URL scheme must not be null");
        }
        if (!scheme.equalsIgnoreCase("http") && !scheme.equalsIgnoreCase("https")) {
            throw new UrlValidationException("Only http and https schemes are allowed, got: " + scheme);
        }

        // 2. Проверка хоста
        String host = uri.getHost();
        if (host == null || host.isBlank()) {
            throw new UrlValidationException("Host must not be null or empty");
        }

        if (host.equalsIgnoreCase("localhost") || host.equals("127.0.0.1")) {
            throw new UrlValidationException("Host cannot be localhost or 127.0.0.1");
        }

        if (host.toLowerCase().endsWith(".local")) {
            throw new UrlValidationException("Host cannot end with .local");
        }

        // 3. Резолв IP и проверка диапазонов
        try {
            InetAddress[] addresses = InetAddress.getAllByName(host);
            for (InetAddress addr : addresses) {
                if (isPrivateIp(addr)) {
                    throw new UrlValidationException("Host resolves to private IP: " + addr.getHostAddress());
                }
            }
        } catch (UnknownHostException e) {
            throw new UrlValidationException("Failed to resolve host: " + host, e);
        }
    }

    private static boolean isPrivateIp(InetAddress addr) {
        return addr.isLoopbackAddress()      // 127.0.0.0/8, ::1
                || addr.isLinkLocalAddress()     // 169.254.0.0/16, fe80::/10
                || addr.isSiteLocalAddress()     // 10.0.0.0/8, 172.16.0.0/12, 192.168.0.0/16, fc00::/7
                || addr.isMulticastAddress()     // 224.0.0.0/4
                || addr.isAnyLocalAddress();     // 0.0.0.0, ::
    }
}
