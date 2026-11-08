package com.alximac.knowledgegraph.texthandler.application.port;

import com.alximac.knowledgegraph.texthandler.application.exception.RemoteServiceException;
import com.alximac.knowledgegraph.texthandler.domain.model.DocumentChunk;

public interface NoteCreatorPort {//вызовы к GO бэку
    String createNote(DocumentChunk chunk, String jwt) throws RemoteServiceException;//Принимает чанк (уже содержащий текст и метаданные),
    // отправляет POST /notes на Go-бэкенд.Возвращает noteId созданной заметки для метода  withNoteId.
    // Кидает RemoteServiceException, если Go недоступен,таймаут, 5xx.
    // Это checked, потому что вызывающий обязан решить: ретраить, пропустить чанк или пометить результат как PARTIAL.



}
