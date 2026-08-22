package com.alximac.knowledgegraph.texthandler.domain.model;

// Processing parameters (chunk size, overlap, minimum length, retry count, link creation).
// Guarantees its own invariants.
public record ImportOptions(
        int chunkSize,
        int overlap,
        Integer minChunkLength,
        Integer maxRetries,
        Boolean createLinks) {

    public ImportOptions {
        if (maxRetries == null) maxRetries = 3;//default

        if (createLinks == null) createLinks = false;//default

        if (minChunkLength == null) minChunkLength = 50;


        if (chunkSize <= 0) throw new IllegalArgumentException("chunkSize is zero or negative " + chunkSize);

        if (overlap < 0) throw new IllegalArgumentException("overlap is less than zero " + overlap);

        if (overlap >= chunkSize) throw new IllegalArgumentException(
                "overlap is bigger or equal to chunkSize, got overlap " + overlap + " ,got chunkSize " + chunkSize);

        if (maxRetries < 0) throw new IllegalArgumentException("maxRetries is less than zero " + maxRetries);

        if (minChunkLength>chunkSize) throw new IllegalArgumentException(
                "minChunkLength can't be bigger than chunkSize," +
                        " got minChunkLength " + minChunkLength + " , got chunkSize " + chunkSize);

        if (minChunkLength < 50) throw new IllegalArgumentException("minimal chunk length should be more than 50, got  "
                + minChunkLength);


    }
}
