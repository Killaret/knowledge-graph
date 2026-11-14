package notehandler

import (
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"

	importer "knowledge-graph/internal/application/import"
	apicommon "knowledge-graph/internal/interfaces/api/common"
	"knowledge-graph/internal/interfaces/api/middleware"

	"github.com/gin-gonic/gin"
)

// maxDocumentFileSize caps the uploaded document. The contract (TZ §3) sets no
// limit; 8 MiB follows from client_max_body_size 10m in nginx.conf and
// nginx.personal.conf, leaving headroom for the request envelope.
const maxDocumentFileSize = 8 << 20

// ImportDocumentAvailability reports whether the Java source-text-handler
// listener is alive: GET /api/v1/import/document/availability.
func (h *Handler) ImportDocumentAvailability(c *gin.Context) {
	middleware.SetDBEntity(c, "notes")
	middleware.SetDBOperation(c, "import_document_availability")

	if _, exists := middleware.GetUserID(c); !exists {
		apicommon.Unauthorized(c)
		return
	}

	available := false
	if h.importSvc != nil {
		if ok, err := h.importSvc.DocumentAvailability(c.Request.Context()); err == nil {
			available = ok
		} else {
			log.Printf("[NoteHandler] document availability check failed: %v", err)
		}
	}
	apicommon.JSON(c, http.StatusOK, gin.H{"available": available})
}

// ImportDocument accepts a document for the Java source-text-handler:
// POST /api/v1/import/document?type=TEXT|URL|FILE[&filename=&content_type=]
// with the document as a raw application/octet-stream body. The body is raw
// (not multipart) because SvelteKit's CSRF check rejects form content types
// on every request whose Origin does not match the server origin — and the
// frontend container runs without ORIGIN, so a multipart post died with 403
// before the proxy ever saw it (SMART-CREATE-1 review blocker 1). With no
// live listener the route answers 503 and queues nothing.
func (h *Handler) ImportDocument(c *gin.Context) {
	middleware.SetDBEntity(c, "notes")
	middleware.SetDBOperation(c, "import_document_create")

	userID, exists := middleware.GetUserID(c)
	if !exists {
		apicommon.Unauthorized(c)
		return
	}

	if h.importSvc == nil {
		apicommon.InternalErrorWithMessage(c, "Import service is not configured")
		return
	}

	alive, err := h.importSvc.DocumentAvailability(c.Request.Context())
	if err != nil {
		apicommon.InternalErrorWithMessage(c, "Failed to check document handler availability")
		return
	}
	if !alive {
		apicommon.Error(c, http.StatusServiceUnavailable, "source_text_handler_unavailable",
			"Document handler is unavailable")
		return
	}

	jwt := extractRequestToken(c)
	if jwt == "" {
		apicommon.Unauthorized(c)
		return
	}

	input, ferr := parseDocumentRequest(c)
	if ferr != nil {
		apicommon.BadRequest(c, []apicommon.FieldError{
			apicommon.NewFieldError("request", apicommon.ReasonInvalidValue, ferr.Error()),
		})
		return
	}

	taskID, err := h.importSvc.SubmitDocument(c.Request.Context(), userID, jwt, input)
	if err != nil {
		if errors.Is(err, importer.ErrDocumentUnavailable) {
			apicommon.Error(c, http.StatusServiceUnavailable, "source_text_handler_unavailable",
				"Document handler is unavailable")
			return
		}
		log.Printf("[NoteHandler] document submit failed: %v", err)
		apicommon.InternalError(c)
		return
	}

	apicommon.JSON(c, http.StatusAccepted, gin.H{"task_id": taskID})
}

// extractRequestToken returns the bearer token Java will present to
// /import/batch — the same credential the user authenticated with.
func extractRequestToken(c *gin.Context) string {
	if header := c.GetHeader("Authorization"); header != "" {
		parts := strings.SplitN(header, " ", 2)
		if len(parts) == 2 && strings.EqualFold(parts[0], "Bearer") {
			return parts[1]
		}
	}
	if cookie, err := c.Cookie("access_token"); err == nil {
		return cookie
	}
	return ""
}

// parseDocumentRequest reads the raw-body upload into a validated
// DocumentInput. Metadata travels in the query string: type (TEXT, URL or
// FILE) for all kinds, filename and content_type for FILE. Form content
// types are refused — a multipart envelope would otherwise be read into the
// document body verbatim and pushed to the queue as junk.
func parseDocumentRequest(c *gin.Context) (importer.DocumentInput, error) {
	contentType := c.GetHeader("Content-Type")
	if strings.HasPrefix(contentType, "multipart/") ||
		strings.HasPrefix(contentType, "application/x-www-form-urlencoded") {
		return importer.DocumentInput{}, errors.New(
			"send the document as a raw application/octet-stream body with metadata in the query string")
	}

	docType := strings.ToUpper(strings.TrimSpace(c.Query("type")))
	switch docType {
	case "TEXT", "URL", "FILE":
	default:
		return importer.DocumentInput{}, errors.New("type must be one of TEXT, URL, FILE")
	}

	body, err := io.ReadAll(io.LimitReader(c.Request.Body, maxDocumentFileSize+1))
	if err != nil {
		return importer.DocumentInput{}, errors.New("failed to read request body")
	}
	if len(body) > maxDocumentFileSize {
		return importer.DocumentInput{}, fmt.Errorf("document is larger than %d MiB", maxDocumentFileSize>>20)
	}

	switch docType {
	case "TEXT":
		content := string(body)
		if strings.TrimSpace(content) == "" {
			return importer.DocumentInput{}, errors.New("content must not be empty")
		}
		return importer.DocumentInput{Type: "TEXT", Content: content}, nil

	case "URL":
		rawURL := strings.TrimSpace(string(body))
		if rawURL == "" {
			return importer.DocumentInput{}, errors.New("content must be a URL")
		}
		if !importer.IsAllowedURL(rawURL) {
			return importer.DocumentInput{}, errors.New("URL is not allowed or uses an unsupported scheme")
		}
		return importer.DocumentInput{Type: "URL", Content: rawURL}, nil

	default: // FILE
		if len(body) == 0 {
			return importer.DocumentInput{}, errors.New("file must not be empty")
		}
		filename := strings.TrimSpace(c.Query("filename"))
		if filename == "" {
			return importer.DocumentInput{}, errors.New("filename is required for FILE")
		}
		// §3 requires content_type for FILE; when the client cannot sniff one,
		// the lawful default is a generic binary stream.
		fileContentType := strings.TrimSpace(c.Query("content_type"))
		if fileContentType == "" {
			fileContentType = "application/octet-stream"
		}
		return importer.DocumentInput{
			Type:        "FILE",
			FileContent: body,
			ContentType: fileContentType,
			Filename:    filename,
		}, nil
	}
}
