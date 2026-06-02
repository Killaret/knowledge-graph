package com.alximac.knowledgegraph.texthandler;

import com.alximac.knowledgegraph.texthandler.application.ImportDocumentHandler;
import com.alximac.knowledgegraph.texthandler.config.AppConfig;

import java.io.IOException;

public class Main {
    public static void main(String[] args) throws IOException {
        AppConfig appConfig = new AppConfig();
        appConfig.start();

    }
}

